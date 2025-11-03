package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

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
	runs := FetchPipelineRunsForUser(r.Context(), user.Namespace)

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
	projects := FetchPipelineConfigsForUser(r.Context(), user.Namespace)

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

	// Fetch PipelineRun details from Kubernetes
	var run *dashboard.PipelineRunDTO
	if fetchedRun := FetchPipelineRunByID(r.Context(), user.Namespace, runID); fetchedRun != nil {
		run = dashboard.MapPipelineRunToDTO(fetchedRun)
	}

	if run == nil {
		http.Error(w, "Pipeline run not found", http.StatusNotFound)
		return
	}

	// Add demo artifacts for successful runs
	artifacts := generateDemoArtifacts(run.ID)

	data := map[string]interface{}{
		"User":        user,
		"PipelineRun": run,
		"Artifacts":   artifacts,
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

	// Use HX-Redirect header for HTMX-based requests
	// This tells HTMX to follow the redirect instead of processing the response as content
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Use HX-Redirect header for HTMX-based requests
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusOK)
	} else {
		// Use HTTP redirect for regular requests
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// generateDemoArtifacts creates demo artifacts for a pipeline run
func generateDemoArtifacts(runID string) []*dashboard.ArtifactDTO {
	artifacts := []*dashboard.ArtifactDTO{
		{
			ID:        "hello-world-run-001-artifact-1",
			Name:      "api-server-v1.2.3.tar.gz",
			Type:      "binary",
			MimeType:  "application/gzip",
			SizeBytes: 45200000, // 45.2 MB
			URL:       "/api/artifacts/hello-world-run-001-artifact-1/download",
			CreatedAt: time.Now(),
		},
		{
			ID:        "hello-world-run-001-artifact-2",
			Name:      "test-report.html",
			Type:      "report",
			MimeType:  "text/html",
			SizeBytes: 512000, // 512 KB
			URL:       "/api/artifacts/hello-world-run-001-artifact-2/download",
			CreatedAt: time.Now(),
		},
		{
			ID:        "hello-world-run-001-artifact-3",
			Name:      "coverage-report.json",
			Type:      "report",
			MimeType:  "application/json",
			SizeBytes: 256000, // 256 KB
			URL:       "/api/artifacts/hello-world-run-001-artifact-3/download",
			CreatedAt: time.Now(),
		},
		{
			ID:        "hello-world-run-001-artifact-4",
			Name:      "build-log.txt",
			Type:      "log",
			MimeType:  "text/plain",
			SizeBytes: 128000, // 128 KB
			URL:       "/api/artifacts/hello-world-run-001-artifact-4/download",
			CreatedAt: time.Now(),
		},
	}

	// For non-demo runs, use runID-based artifact IDs
	if runID != "hello-world-run-001" {
		for i, artifact := range artifacts {
			artifact.ID = runID + "-artifact-" + strconv.Itoa(i+1)
			artifact.URL = "/api/artifacts/" + artifact.ID + "/download"
		}
	}

	return artifacts
}
