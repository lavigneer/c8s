package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

// PipelineFilters holds filter parameters for pipeline runs
type PipelineFilters struct {
	Status   string
	Branch   string
	Search   string
	FromDate time.Time
	ToDate   time.Time
}

// ListPipelineRunsHandler handles GET /api/projects/{projectId}/runs
// Returns paginated pipeline runs in JSON or HTML depending on request
func ListPipelineRunsHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		dashboard.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse pagination parameters
	params := dashboard.ParsePaginationParams(r.URL.Query())

	// Parse filter parameters
	filters := ParseFilters(r)

	// Fetch PipelineRuns from Kubernetes (user's namespace)
	dtos := FetchPipelineRunsForUser(r.Context(), user.Namespace)

	// Apply filters (client-side for now, should be K8s-side)
	dtos = filterPipelineRuns(dtos, filters)

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

// filterPipelineRuns applies status, branch, search, and date range filters to pipeline runs
func filterPipelineRuns(runs []*dashboard.PipelineRunDTO, filters PipelineFilters) []*dashboard.PipelineRunDTO {
	var filtered []*dashboard.PipelineRunDTO

	for _, run := range runs {
		// Status filter
		if filters.Status != "" && run.Status != filters.Status {
			continue
		}

		// Branch filter
		if filters.Branch != "" && !strings.EqualFold(run.Branch, filters.Branch) {
			continue
		}

		// Search filter (searches commit SHA, branch, author)
		if filters.Search != "" {
			searchLower := strings.ToLower(filters.Search)
			if !strings.Contains(strings.ToLower(run.CommitSHA), searchLower) &&
				!strings.Contains(strings.ToLower(run.Branch), searchLower) &&
				!strings.Contains(strings.ToLower(run.Author), searchLower) {
				continue
			}
		}

		// Date range filter
		if !filters.FromDate.IsZero() && run.TriggeredAt.Before(filters.FromDate) {
			continue
		}

		if !filters.ToDate.IsZero() && run.TriggeredAt.After(filters.ToDate) {
			continue
		}

		filtered = append(filtered, run)
	}

	return filtered
}

// GetPipelineRunHandler handles GET /api/runs/{runId}
// Returns full pipeline run with step details
func GetPipelineRunHandler(w http.ResponseWriter, r *http.Request) {
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
	run := FetchPipelineRunByID(r.Context(), user.Namespace, runID)

	if run == nil {
		dashboard.RespondNotFound(w, "run")
		return
	}

	// Convert to DTO
	dto := dashboard.MapPipelineRunToDTO(run)
	dashboard.RespondSuccess(w, http.StatusOK, dto)
}

// FetchPipelineRuns queries Kubernetes for pipeline runs with optional filters
// This is a placeholder - actual implementation would use the K8s client
func FetchPipelineRuns(ctx context.Context, projectID, status, branch, search string) ([]*v1alpha1.PipelineRun, error) {
	// TODO: Implement actual K8s query using client-go
	// This would query PipelineRun resources from the cluster
	return []*v1alpha1.PipelineRun{}, nil
}

// ParseFilters extracts filter parameters from query string
func ParseFilters(r *http.Request) PipelineFilters {
	filters := PipelineFilters{
		Status: r.URL.Query().Get("status"),
		Branch: r.URL.Query().Get("branch"),
		Search: r.URL.Query().Get("search"),
	}

	// Parse date filters
	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filters.FromDate = fromDate
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			// Set to end of day
			filters.ToDate = toDate.Add(24*time.Hour - time.Second)
		}
	}

	return filters
}

// ListBranchesHandler returns unique branch names for a project
// GET /api/projects/{projectId}/branches
func ListBranchesHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		dashboard.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Fetch PipelineRuns from Kubernetes (user's namespace)
	dtos := FetchPipelineRunsForUser(r.Context(), user.Namespace)

	// Extract unique branch names
	branchMap := make(map[string]bool)
	var branches []string

	for _, dto := range dtos {
		if !branchMap[dto.Branch] && dto.Branch != "" {
			branches = append(branches, dto.Branch)
			branchMap[dto.Branch] = true
		}
	}

	dashboard.RespondSuccess(w, http.StatusOK, branches)
}
