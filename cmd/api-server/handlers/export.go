package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/org/c8s/pkg/dashboard"
)

// ExportFormat represents the export file format
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// ExportPipelineRunsHandler exports pipeline runs in specified format
// GET /api/exports/runs?format=json|csv
func ExportPipelineRunsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		dashboard.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse format parameter
	format := ExportFormat(r.URL.Query().Get("format"))
	if format != ExportFormatJSON && format != ExportFormatCSV {
		format = ExportFormatJSON // Default to JSON
	}

	// Parse filter parameters
	filters := ParseFilters(r)

	// Fetch PipelineRuns from Kubernetes (user's namespace)
	runs := FetchPipelineRunsForUser(r.Context(), user.Namespace)

	// Apply filters
	runs = filterPipelineRuns(runs, filters)

	// Export based on format
	switch format {
	case ExportFormatCSV:
		exportAsCSV(w, runs)
	case ExportFormatJSON:
		fallthrough
	default:
		exportAsJSON(w, runs)
	}
}

// exportAsJSON exports pipeline runs as JSON
func exportAsJSON(w http.ResponseWriter, runs []*dashboard.PipelineRunDTO) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=pipeline-runs-%s.json", time.Now().Format("2006-01-02")))
	w.WriteHeader(http.StatusOK)

	data := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"count":       len(runs),
		"runs":        runs,
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR: Failed to encode JSON export data: %v", err)
		return
	}
}

// exportAsCSV exports pipeline runs as CSV
func exportAsCSV(w http.ResponseWriter, runs []*dashboard.PipelineRunDTO) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=pipeline-runs-%s.csv", time.Now().Format("2006-01-02")))
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"ID", "Name", "Status", "Commit SHA", "Branch", "Author",
		"Triggered At", "Started At", "Completed At", "Duration (seconds)",
		"Step Count", "Success Count", "Failure Count", "Artifact Count",
	}
	writer.Write(headers)

	// Write data rows
	for _, run := range runs {
		startedAt := ""
		if run.StartedAt != nil {
			startedAt = run.StartedAt.Format(time.RFC3339)
		}

		completedAt := ""
		if run.CompletedAt != nil {
			completedAt = run.CompletedAt.Format(time.RFC3339)
		}

		duration := ""
		if run.DurationSeconds != nil {
			duration = fmt.Sprintf("%d", *run.DurationSeconds)
		}

		row := []string{
			run.ID,
			run.Name,
			run.Status,
			run.CommitSHA,
			run.Branch,
			run.Author,
			run.TriggeredAt.Format(time.RFC3339),
			startedAt,
			completedAt,
			duration,
			fmt.Sprintf("%d", run.StepCount),
			fmt.Sprintf("%d", run.SuccessCount),
			fmt.Sprintf("%d", run.FailureCount),
			fmt.Sprintf("%d", run.ArtifactCount),
		}
		writer.Write(row)
	}
}
