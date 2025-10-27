package handlers

import (
	"net/http"

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

	// TODO: Fetch initial pipeline runs from Kubernetes
	// For now, use empty list
	var runs []*dashboard.PipelineRunDTO

	// Parse pagination parameters
	params := dashboard.ParsePaginationParams(r.URL.Query())

	// Calculate pagination metadata
	total := len(runs)
	paginationMeta := dashboard.CalculatePagination(total, params.Page, params.PerPage)

	// Prepare data for template
	data := map[string]interface{}{
		"User":         user,
		"ProjectID":    projectID,
		"PipelineRuns": runs,
		"Pagination":   paginationMeta,
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
	var projects []*dashboard.ProjectDTO
	if k8sClient != nil {
		configs, err := k8sClient.ListPipelineConfigs(r.Context(), user.Namespace)
		if err == nil && configs != nil {
			projects = make([]*dashboard.ProjectDTO, len(configs.Items))
			for i, config := range configs.Items {
				projects[i] = mapPipelineConfigToProjectDTO(&config, user.Namespace)
			}
		}
	}

	data := map[string]interface{}{
		"User":     user,
		"Projects": projects,
	}

	if err := dashboard.RenderTemplate(w, "projects", data); err != nil {
		http.Error(w, "Failed to render projects page", http.StatusInternalServerError)
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

	// TODO: Fetch PipelineRun details from Kubernetes
	var run *dashboard.PipelineRunDTO

	data := map[string]interface{}{
		"User": user,
		"Run":  run,
	}

	if err := dashboard.RenderTemplate(w, "pipeline_detail", data); err != nil {
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
