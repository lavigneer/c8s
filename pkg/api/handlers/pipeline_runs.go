package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/auth"
	"github.com/org/c8s/pkg/dashboard"
	"github.com/org/c8s/pkg/api/responses"
	"github.com/org/c8s/pkg/pagination"
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
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		_ = responses.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse pagination parameters
	params := pagination.ParsePaginationParams(r.URL.Query())

	// Parse filter parameters
	filters := ParseFilters(r)

	// Fetch PipelineRuns from Kubernetes (user's namespace)
	dtos := FetchPipelineRunsForUser(r.Context(), user.Namespace)

	// Apply filters (client-side for now, should be K8s-side)
	dtos = filterPipelineRuns(dtos, filters)

	// Paginate results
	total := len(dtos)
	start, end := pagination.GetPaginationIndices(params.Page, params.PerPage)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pagedRuns := dtos[start:end]
	paginationMeta := pagination.CalculatePagination(total, params.Page, params.PerPage)

	// Check if HTMX request
	if dashboard.IsHTMXRequest(r) {
		// Return HTML fragment (pipeline rows only)
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"PipelineRuns": pagedRuns,
		}
		_ = dashboard.RenderTemplate(w, "pipeline_list_rows", data)
	} else {
		// Return JSON API response
		_ = responses.RespondSuccessWithMeta(w, http.StatusOK, pagedRuns, paginationMeta)
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
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "runId required")
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		_ = responses.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Fetch PipelineRun from Kubernetes (user's namespace)
	run := FetchPipelineRunByID(r.Context(), user.Namespace, runID)

	if run == nil {
		_ = responses.RespondNotFound(w, "run")
		return
	}

	// Convert to DTO
	dto := dashboard.MapPipelineRunToDTO(run)
	_ = responses.RespondSuccess(w, http.StatusOK, dto)
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

// GetPipelineStatusHandler handles GET /api/v1/pipelines/{name}/status
// Returns the status of a PipelineRun by name (for webhook integration)
// This is called by the GitHub Actions dog-fooding workflow to check pipeline status
func GetPipelineStatusHandler(w http.ResponseWriter, r *http.Request) {
	pipelineName := chi.URLParam(r, "name")
	if pipelineName == "" {
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "pipeline name required")
		return
	}

	// Fetch PipelineRun from Kubernetes (search in default namespace for webhook-created runs)
	run := FetchPipelineRunByID(r.Context(), "default", pipelineName)
	if run == nil {
		_ = responses.RespondNotFound(w, "pipeline")
		return
	}

	// Convert to DTO and extract status
	dto := dashboard.MapPipelineRunToDTO(run)

	// Return simple status response
	response := map[string]interface{}{
		"name":   pipelineName,
		"phase":  dto.Status,
		"status": dto.Status,
	}
	_ = responses.RespondSuccess(w, http.StatusOK, response)
}

// GetPipelineLogsHandler handles GET /api/v1/pipelines/{name}/logs
// Returns aggregated logs from all steps of a PipelineRun
func GetPipelineLogsHandler(w http.ResponseWriter, r *http.Request) {
	pipelineName := chi.URLParam(r, "name")
	if pipelineName == "" {
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "pipeline name required")
		return
	}

	// Fetch PipelineRun from Kubernetes (search in default namespace for webhook-created runs)
	run := FetchPipelineRunByID(r.Context(), "default", pipelineName)
	if run == nil {
		_ = responses.RespondNotFound(w, "pipeline")
		return
	}

	// Return pipeline logs response
	// TODO: Aggregate logs from all steps when detailed step logs are available
	response := map[string]interface{}{
		"name": pipelineName,
		"logs": []string{},
	}
	_ = responses.RespondSuccess(w, http.StatusOK, response)
}

// ListBranchesHandler returns unique branch names for a project
// GET /api/projects/{projectId}/branches
func ListBranchesHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		_ = responses.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		_ = responses.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
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

	_ = responses.RespondSuccess(w, http.StatusOK, branches)
}
