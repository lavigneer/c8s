package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/auth"
	"github.com/org/c8s/pkg/dashboard"
	"github.com/org/c8s/pkg/api/responses"
	"github.com/org/c8s/pkg/logstorage"
)

// LogStreamHandler streams step logs via Server-Sent Events
// GET /api/runs/{runId}/steps/{stepId}/logs
func LogStreamHandler(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")
	stepID := chi.URLParam(r, "stepId")

	if runID == "" || stepID == "" {
		http.Error(w, "runId and stepId required", http.StatusBadRequest)
		return
	}

	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set up SSE connection
	if !setupSSEConnection(w, flusher) {
		return
	}

	// Get user namespace for Kubernetes queries
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Use Kubernetes log storage to fetch real logs from Job Pods
	var logStorage logstorage.LogStorage
	if k8sClient != nil {
		logStorage = logstorage.NewKubernetesLogStorage(k8sClient, k8sClient.Clientset, user.Namespace, runID)
	} else {
		// Fallback to demo logs if K8s client not available
		logStorage = logstorage.NewInMemoryLogStorageWithRun(runID)
	}

	// Create channels for log streaming
	logChan := make(chan string, 100)
	errChan := make(chan error, 1)

	// Stream logs in a goroutine (StreamStepLogs blocks until all logs are sent)
	go func() {
		defer close(logChan) // Close channel when done streaming
		if err := logStorage.StreamStepLogs(r.Context(), runID, stepID, logChan); err != nil {
			log.Printf("Log streaming error for run %s step %s: %v", runID, stepID, err)
			errChan <- err
		} else {
			log.Printf("Log streaming completed successfully for run %s step %s", runID, stepID)
		}
	}()

	// Stream logs to client
	streamLogsToClient(w, r, flusher, logChan, errChan)
}

// setupSSEConnection initializes SSE headers and sends connection message
func setupSSEConnection(w http.ResponseWriter, flusher http.Flusher) bool {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// CORS headers should be handled by middleware, not set per-request

	// Write the response status and initial connection message
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"Connected to log stream\"}\n\n"); err != nil {
		log.Printf("ERROR: Failed to send SSE connection message: %v", err)
		return false
	}
	flusher.Flush()
	return true
}

// streamLogsToClient handles the main log streaming loop
func streamLogsToClient(w http.ResponseWriter, r *http.Request, flusher http.Flusher, logChan chan string, errChan chan error) {
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return

		case err := <-errChan:
			handleStreamError(w, err)
			return

		case line, ok := <-logChan:
			if !ok {
				// Channel closed, all logs sent
				sendCompleteEvent(w, flusher)
				return
			}
			if !sendLogEvent(w, flusher, line) {
				return
			}
		}
	}
}

// handleStreamError sends an error event to the client
func handleStreamError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	errorResp := map[string]interface{}{
		"error": err.Error(),
	}
	data, marshalErr := json.Marshal(errorResp)
	if marshalErr != nil {
		log.Printf("ERROR: Failed to marshal error response: %v", marshalErr)
		return
	}
	_, _ = fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
}

// sendCompleteEvent sends a completion event to the client
func sendCompleteEvent(w http.ResponseWriter, flusher http.Flusher) {
	if _, err := fmt.Fprintf(w, "event: complete\ndata: {\"message\":\"Log stream completed\"}\n\n"); err != nil {
		log.Printf("ERROR: Failed to send log complete event: %v", err)
		return
	}
	flusher.Flush()
}

// sendLogEvent sends a single log line to the client
func sendLogEvent(w http.ResponseWriter, flusher http.Flusher, line string) bool {
	logEntry := map[string]interface{}{
		"line":      line,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to marshal log entry: %v", err)
		return false
	}

	if _, err := fmt.Fprintf(w, "event: log\ndata: %s\n\n", data); err != nil {
		log.Printf("ERROR: Failed to send log event: %v", err)
		return false
	}
	flusher.Flush()
	return true
}

// GetLogsHandler returns logs as plain text
// GET /api/runs/{runId}/steps/{stepId}/logs/text
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")
	stepID := chi.URLParam(r, "stepId")

	if runID == "" || stepID == "" {
		http.Error(w, "runId and stepId required", http.StatusBadRequest)
		return
	}

	// TODO: Get actual log storage
	logStorage := logstorage.NewInMemoryLogStorageWithRun(runID)

	// Get logs
	reader, err := logStorage.GetStepLogs(r.Context(), runID, stepID)
	if err != nil {
		_ = responses.RespondError(w, http.StatusNotFound, "LOGS_NOT_FOUND", "Logs not found for step")
		return
	}
	defer func() { _ = reader.Close() }()

	// Return as plain text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Copy reader to response
	if _, err := fmt.Fprint(w, "Logs for "+stepID+"\n"); err != nil {
		log.Printf("ERROR: Failed to write log header: %v", err)
		return
	}

	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("ERROR: Failed to copy logs to response: %v", err)
		return
	}
}

// GetLogSnapshotHandler returns the last N lines of logs
// GET /api/runs/{runId}/steps/{stepId}/logs/snapshot?lines=100
func GetLogSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")
	stepID := chi.URLParam(r, "stepId")

	if runID == "" || stepID == "" {
		http.Error(w, "runId and stepId required", http.StatusBadRequest)
		return
	}

	// Parse lines parameter
	lines := 100
	if linesParam := r.URL.Query().Get("lines"); linesParam != "" {
		if _, err := fmt.Sscanf(linesParam, "%d", &lines); err != nil {
			_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid lines parameter")
			return
		}
	}

	// TODO: Get actual log storage
	logStorage := logstorage.NewInMemoryLogStorageWithRun(runID)

	// Get log snapshot
	logLines, err := logStorage.GetLogSnapshot(r.Context(), runID, stepID, lines)
	if err != nil {
		_ = responses.RespondError(w, http.StatusNotFound, "LOGS_NOT_FOUND", "Logs not found for step")
		return
	}

	// Return as JSON
	_ = responses.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"runId":  runID,
		"stepId": stepID,
		"lines":  logLines,
		"count":  len(logLines),
	})
}

// ListStepsHandler returns all steps for a pipeline run
// GET /api/runs/{runId}/steps
func ListStepsHandler(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	if runID == "" {
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "runId required")
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		_ = responses.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Fetch PipelineRun from Kubernetes (user's namespace)
	var run *dashboard.PipelineRunDTO
	if fetchedRun := FetchPipelineRunByID(r.Context(), user.Namespace, runID); fetchedRun != nil {
		run = dashboard.MapPipelineRunToDTO(fetchedRun)
	}

	if run == nil {
		_ = responses.RespondNotFound(w, "run")
		return
	}

	// Return demo steps based on run data
	steps := []map[string]interface{}{
		{
			"id":               "step-1",
			"name":             "checkout",
			"status":           "Succeeded",
			"started_at":       "2025-10-27T04:30:10Z",
			"completed_at":     "2025-10-27T04:30:15Z",
			"duration_seconds": 5,
		},
		{
			"id":               "step-2",
			"name":             "build",
			"status":           "Succeeded",
			"started_at":       "2025-10-27T04:30:16Z",
			"completed_at":     "2025-10-27T04:30:25Z",
			"duration_seconds": 9,
		},
		{
			"id":               "step-3",
			"name":             "test",
			"status":           "Succeeded",
			"started_at":       "2025-10-27T04:30:26Z",
			"completed_at":     "2025-10-27T04:30:31Z",
			"duration_seconds": 5,
		},
	}

	_ = responses.RespondSuccess(w, http.StatusOK, steps)
}

// GetStepOptionsHandler returns step options as HTML for HTMX
// GET /api/runs/{runId}/steps/options
func GetStepOptionsHandler(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	if runID == "" {
		http.Error(w, "runId required", http.StatusBadRequest)
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Fetch raw PipelineRun from Kubernetes to access step status
	pipelineRun := FetchPipelineRunByID(r.Context(), user.Namespace, runID)
	if pipelineRun == nil {
		http.Error(w, "Pipeline run not found", http.StatusNotFound)
		return
	}

	// Return HTML option elements for HTMX
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Default option
	_, _ = w.Write([]byte(`<option value="">-- Choose a step --</option>`))

	// If no steps, show message
	if len(pipelineRun.Status.Steps) == 0 {
		_, _ = w.Write([]byte(`<option disabled>No steps found</option>`))
		return
	}

	// Add step options from actual PipelineRun status
	for _, step := range pipelineRun.Status.Steps {
		html := fmt.Sprintf(`<option value="%s">%s (%s)</option>`,
			step.Name, step.Name, step.Phase)
		_, _ = w.Write([]byte(html))
	}
}
