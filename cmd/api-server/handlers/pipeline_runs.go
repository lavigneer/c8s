package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

// ListPipelineRunsHandler handles GET /api/projects/{projectId}/runs
// Returns paginated pipeline runs in JSON or HTML depending on request
func ListPipelineRunsHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	// Parse pagination parameters
	params := dashboard.ParsePaginationParams(r.URL.Query())

	// Get filter parameters
	status := r.URL.Query().Get("status")
	branch := r.URL.Query().Get("branch")
	search := r.URL.Query().Get("search")

	// TODO: Fetch PipelineRuns from Kubernetes
	// For now, return empty list
	runs := []*v1alpha1.PipelineRun{}

	// Transform to DTOs
	dtos := make([]*dashboard.PipelineRunDTO, len(runs))
	for i, run := range runs {
		dtos[i] = dashboard.MapPipelineRunToDTO(run)
	}

	// Apply filters (client-side for now, should be K8s-side)
	dtos = filterPipelineRuns(dtos, status, branch, search)

	// Paginate results
	total := len(dtos)
	start, end := dashboard.GetPaginationIndices(params.Page, params.PerPage)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pagedRuns := dtos[start:end]
	paginationMeta := dashboard.CalculatePagination(total, params.Page, params.PerPage)

	// Check if HTMX request
	if dashboard.IsHTMXRequest(r) {
		// Return HTML fragment (pipeline rows only)
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"PipelineRuns": pagedRuns,
		}
		dashboard.RenderTemplate(w, "pipeline_list_rows", data)
	} else {
		// Return JSON API response
		dashboard.RespondSuccessWithMeta(w, http.StatusOK, pagedRuns, paginationMeta)
	}
}

// filterPipelineRuns applies status, branch, and search filters to pipeline runs
func filterPipelineRuns(runs []*dashboard.PipelineRunDTO, status, branch, search string) []*dashboard.PipelineRunDTO {
	var filtered []*dashboard.PipelineRunDTO

	for _, run := range runs {
		// Status filter
		if status != "" && run.Status != status {
			continue
		}

		// Branch filter
		if branch != "" && !strings.EqualFold(run.Branch, branch) {
			continue
		}

		// Search filter (searches commit SHA, branch, author)
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(run.CommitSHA), searchLower) &&
				!strings.Contains(strings.ToLower(run.Branch), searchLower) &&
				!strings.Contains(strings.ToLower(run.Author), searchLower) {
				continue
			}
		}

		filtered = append(filtered, run)
	}

	return filtered
}

// GetPipelineRunHandler handles GET /api/projects/{projectId}/runs/{runId}
func GetPipelineRunHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	runID := chi.URLParam(r, "runId")

	if projectID == "" || runID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId and runId required")
		return
	}

	// TODO: Fetch PipelineRun from Kubernetes
	// For now, return not found
	dashboard.RespondNotFound(w, "run")
}

// FetchPipelineRuns queries Kubernetes for pipeline runs with optional filters
// This is a placeholder - actual implementation would use the K8s client
func FetchPipelineRuns(ctx context.Context, projectID, status, branch, search string) ([]*v1alpha1.PipelineRun, error) {
	// TODO: Implement actual K8s query using client-go
	// This would query PipelineRun resources from the cluster
	return []*v1alpha1.PipelineRun{}, nil
}
