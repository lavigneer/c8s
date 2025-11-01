package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/dashboard"
)

// Global broadcaster for pipeline updates per project
var (
	pipelineBroadcasters = make(map[string]*dashboard.SSEBroadcaster)
	broadcasterMutex     sync.RWMutex
)

// PipelineUpdatesSSEHandler streams real-time pipeline status updates via Server-Sent Events
// GET /api/projects/{projectId}/runs/updates
func PipelineUpdatesSSEHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		http.Error(w, "projectId required", http.StatusBadRequest)
		return
	}

	// Get or create broadcaster for this project
	broadcaster := getOrCreateBroadcaster(projectID)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Write 200 status before checking for flusher
	// This prevents issues with wrapped response writers
	w.WriteHeader(http.StatusOK)

	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		// If flushing isn't supported, we can't do SSE properly
		// This shouldn't happen with standard http servers
		return
	}

	// Subscribe to updates
	updateChan := broadcaster.Subscribe()
	defer broadcaster.Unsubscribe(updateChan)

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"Connected to pipeline updates\"}\n\n")
	flusher.Flush()

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return

		case event, ok := <-updateChan:
			if !ok {
				// Channel closed, broadcaster shut down
				return
			}

			// Write SSE event
			fmt.Fprintf(w, "%s", event.String())
			flusher.Flush()
		}
	}
}

// BroadcastPipelineUpdate sends a pipeline status update to all subscribers
func BroadcastPipelineUpdate(projectID string, run *dashboard.PipelineRunDTO) {
	broadcaster := getOrCreateBroadcaster(projectID)

	// Create SSE event
	data, _ := json.Marshal(run)
	event := dashboard.NewEventBuilder().
		WithID(run.ID).
		WithEvent("run_status_changed").
		WithData(string(data)).
		Build()

	broadcaster.BroadcastAsync(event)
}

// getOrCreateBroadcaster gets or creates a broadcaster for a project
func getOrCreateBroadcaster(projectID string) *dashboard.SSEBroadcaster {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()

	if broadcaster, exists := pipelineBroadcasters[projectID]; exists {
		return broadcaster
	}

	// Create new broadcaster
	broadcaster := dashboard.NewSSEBroadcaster()
	pipelineBroadcasters[projectID] = broadcaster

	return broadcaster
}

// CloseBroadcaster closes a broadcaster when project is deleted
func CloseBroadcaster(projectID string) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()

	if broadcaster, exists := pipelineBroadcasters[projectID]; exists {
		broadcaster.Close()
		delete(pipelineBroadcasters, projectID)
	}
}

// GetBroadcasterStats returns statistics about connected clients
func GetBroadcasterStats(projectID string) map[string]int {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()

	stats := make(map[string]int)
	for id, broadcaster := range pipelineBroadcasters {
		stats[id] = broadcaster.ClientCount()
	}

	return stats
}
