package handlers

import (
	"log"
	"net/http"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

// DashboardHandler renders the main dashboard page with pipeline list
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Get project ID from query param or use first available project
	projectID := r.URL.Query().Get("projectId")
	if projectID == "" {
		// TODO: Get user's first project from access service
		// For now, use a default value
		projectID = "default-project"
	}

	// Fetch initial pipeline runs from Kubernetes
	var runs []*dashboard.PipelineRunDTO
	if k8sClient != nil {
		// Query both user namespace and c8s-system for test data
		var pipelineRuns []*v1alpha1.PipelineRun

		if userRuns, err := k8sClient.ListPipelineRuns(r.Context(), user.Namespace); err == nil && userRuns != nil {
			for i := range userRuns.Items {
				pipelineRuns = append(pipelineRuns, &userRuns.Items[i])
			}
		}

		if sysRuns, err := k8sClient.ListPipelineRuns(r.Context(), "c8s-system"); err == nil && sysRuns != nil {
			for i := range sysRuns.Items {
				pipelineRuns = append(pipelineRuns, &sysRuns.Items[i])
			}
		}

		// Convert to DTOs
		runs = make([]*dashboard.PipelineRunDTO, len(pipelineRuns))
		for i, run := range pipelineRuns {
			runs[i] = dashboard.MapPipelineRunToDTO(run)
		}
	}

	// Parse filter parameters and apply filters
	filters := ParseFilters(r)
	runs = filterPipelineRuns(runs, filters)

	// Parse pagination parameters
	params := dashboard.ParsePaginationParams(r.URL.Query())

	// Calculate pagination metadata
	total := len(runs)
	paginationMeta := dashboard.CalculatePagination(total, params.Page, params.PerPage)

	// Get branches for filter
	var branches []string
	branchMap := make(map[string]bool)
	for _, run := range runs {
		if !branchMap[run.Branch] {
			branches = append(branches, run.Branch)
			branchMap[run.Branch] = true
		}
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":         user,
		"ProjectID":    projectID,
		"PipelineRuns": runs,
		"Pagination":   paginationMeta,
		"Branches":     branches,
	}

	// Render template
	if err := dashboard.RenderTemplate(w, "pipeline_list", data); err != nil {
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		return
	}
}

// ProjectsHandler renders the projects page
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Fetch user's projects from Kubernetes
	// For development, also check c8s-system namespace for test data
	var projects []*dashboard.ProjectDTO
	if k8sClient != nil {
		namespaces := []string{user.Namespace, "c8s-system"}
		projectMap := make(map[string]*dashboard.ProjectDTO)

		for _, ns := range namespaces {
			configs, err := k8sClient.ListPipelineConfigs(r.Context(), ns)
			if err == nil && configs != nil {
				for _, config := range configs.Items {
					projectMap[config.Name] = mapPipelineConfigToProjectDTO(&config, ns)
				}
			}
		}

		// Convert map to slice
		projects = make([]*dashboard.ProjectDTO, 0, len(projectMap))
		for _, p := range projectMap {
			projects = append(projects, p)
		}
	}

	data := map[string]interface{}{
		"User":     user,
		"Projects": projects,
	}

	if err := dashboard.RenderTemplate(w, "projects", data); err != nil {
		log.Printf("ERROR: Failed to render projects page: %v", err)
		// Don't call http.Error since RenderTemplate already wrote the header
		return
	}
}

// PipelineRunDetailsHandler renders pipeline run detail page
func PipelineRunDetailsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Extract runID from URL path
	runID := r.PathValue("runId")
	if runID == "" {
		http.Error(w, "Run ID required", http.StatusBadRequest)
		return
	}

	// Fetch PipelineRun details from Kubernetes (check both namespaces)
	var run *dashboard.PipelineRunDTO
	if k8sClient != nil {
		// Try default namespace first
		if fetchedRun, err := k8sClient.GetPipelineRun(r.Context(), "default", runID); err == nil {
			run = dashboard.MapPipelineRunToDTO(fetchedRun)
		}

		// Try c8s-system namespace if not found
		if run == nil {
			if fetchedRun, err := k8sClient.GetPipelineRun(r.Context(), "c8s-system", runID); err == nil {
				run = dashboard.MapPipelineRunToDTO(fetchedRun)
			}
		}
	}

	if run == nil {
		http.Error(w, "Pipeline run not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"User":        user,
		"PipelineRun": run,
	}

	if err := dashboard.RenderTemplate(w, "pipeline_detail", data); err != nil {
		log.Printf("ERROR: Failed to render pipeline details: %v", err)
		http.Error(w, "Failed to render pipeline details", http.StatusInternalServerError)
		return
	}
}

// LoginHandler renders login page or handles login POST
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Handle login submission
		handleLoginSubmit(w, r)
		return
	}

	// Render login page for GET requests
	data := map[string]interface{}{
		"Title": "Login",
	}

	if err := dashboard.RenderTemplate(w, "login", data); err != nil {
		http.Error(w, "Failed to render login page", http.StatusInternalServerError)
		return
	}
}

// handleLoginSubmit processes login form submission
func handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Basic validation (in production, validate against actual auth system)
	if username == "" || password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Create a demo token (in production, this would be a proper JWT)
	token := "demo_token_" + username

	// Set auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
