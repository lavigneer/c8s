package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/dashboard"
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

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// CORS headers should be handled by middleware, not set per-request
	// Removed hardcoded "Access-Control-Allow-Origin: *" as it's insecure and violates CORS spec with credentials

	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// TODO: Get actual log storage implementation
	// For now, use in-memory storage for testing - include demo logs for this run
	logStorage := dashboard.NewInMemoryLogStorageWithRun(runID)

	// Send initial connection message
	if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"Connected to log stream\"}\n\n"); err != nil {
		log.Printf("ERROR: Failed to send SSE connection message: %v", err)
		http.Error(w, "Failed to establish stream", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Create log channel
	logChan := make(chan string, 100)
	errChan := make(chan error, 1)

	// Start streaming logs in a goroutine
	go func() {
		if err := logStorage.StreamStepLogs(r.Context(), runID, stepID, logChan); err != nil {
			errChan <- err
		}
		close(logChan)
	}()

	// Stream logs to client
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return

		case err := <-errChan:
			// Error occurred while streaming
			if err != nil {
				errorResp := map[string]interface{}{
					"error": err.Error(),
				}
				data, err := json.Marshal(errorResp)
				if err != nil {
					log.Printf("ERROR: Failed to marshal error response: %v", err)
					return
				}
				if _, err := fmt.Fprintf(w, "event: error\ndata: %s\n\n", data); err != nil {
					log.Printf("ERROR: Failed to send error event: %v", err)
					return
				}
			}
			return

		case line, ok := <-logChan:
			if !ok {
				// All logs have been streamed
				if _, err := fmt.Fprintf(w, "event: complete\ndata: {\"message\":\"Log stream completed\"}\n\n"); err != nil {
					log.Printf("ERROR: Failed to send log complete event: %v", err)
					return
				}
				flusher.Flush()
				return
			}

			// Format log entry
			logEntry := map[string]interface{}{
				"line":      line,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			data, err := json.Marshal(logEntry)
			if err != nil {
				log.Printf("ERROR: Failed to marshal log entry: %v", err)
				return
			}

			// Send log event
			if _, err := fmt.Fprintf(w, "event: log\ndata: %s\n\n", data); err != nil {
				log.Printf("ERROR: Failed to send log event: %v", err)
				return
			}
			flusher.Flush()
		}
	}
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
	logStorage := dashboard.NewInMemoryLogStorageWithRun(runID)

	// Get logs
	reader, err := logStorage.GetStepLogs(r.Context(), runID, stepID)
	if err != nil {
		dashboard.RespondError(w, http.StatusNotFound, "LOGS_NOT_FOUND", "Logs not found for step")
		return
	}
	defer reader.Close()

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
			dashboard.RespondError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid lines parameter")
			return
		}
	}

	// TODO: Get actual log storage
	logStorage := dashboard.NewInMemoryLogStorageWithRun(runID)

	// Get log snapshot
	logLines, err := logStorage.GetLogSnapshot(r.Context(), runID, stepID, lines)
	if err != nil {
		dashboard.RespondError(w, http.StatusNotFound, "LOGS_NOT_FOUND", "Logs not found for step")
		return
	}

	// Return as JSON
	dashboard.RespondSuccess(w, http.StatusOK, map[string]interface{}{
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
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "runId required")
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		dashboard.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Fetch PipelineRun from Kubernetes (user's namespace)
	var run *dashboard.PipelineRunDTO
	if fetchedRun := FetchPipelineRunByID(r.Context(), user.Namespace, runID); fetchedRun != nil {
		run = dashboard.MapPipelineRunToDTO(fetchedRun)
	}

	if run == nil {
		dashboard.RespondNotFound(w, "run")
		return
	}

	// Return demo steps based on run data
	steps := []map[string]interface{}{
		{
			"id":             "step-1",
			"name":           "checkout",
			"status":         "Succeeded",
			"started_at":     "2025-10-27T04:30:10Z",
			"completed_at":   "2025-10-27T04:30:15Z",
			"duration_seconds": 5,
		},
		{
			"id":             "step-2",
			"name":           "build",
			"status":         "Succeeded",
			"started_at":     "2025-10-27T04:30:16Z",
			"completed_at":   "2025-10-27T04:30:25Z",
			"duration_seconds": 9,
		},
		{
			"id":             "step-3",
			"name":           "test",
			"status":         "Succeeded",
			"started_at":     "2025-10-27T04:30:26Z",
			"completed_at":   "2025-10-27T04:30:31Z",
			"duration_seconds": 5,
		},
	}

	dashboard.RespondSuccess(w, http.StatusOK, steps)
}
