package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

// k8sClient is initialized by the main package
var k8sClient *dashboard.K8sClient

// InitK8sClient sets the k8s client for handlers
func InitK8sClient(client *dashboard.K8sClient) {
	k8sClient = client
}

// ListProjectsHandler returns projects for authenticated user
// GET /api/projects
// Authorization: viewer or higher (read access)
func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	// Query Kubernetes for PipelineConfigs (projects)
	configs, err := k8sClient.ListPipelineConfigs(r.Context(), user.Namespace)
	if err != nil {
		dashboard.RespondError(w, http.StatusInternalServerError, "FETCH_FAILED", fmt.Sprintf("Failed to fetch projects: %v", err))
		return
	}

	// Map to DTOs - filter by user's access
	dtos := make([]*dashboard.ProjectDTO, 0, len(configs.Items))
	for _, config := range configs.Items {
		// Check if user has read access to this project
		hasAccess, err := authzService.UserHasProjectAccess(r.Context(), user.ID, config.Name)
		if err != nil {
			// Log but continue (user might not have role binding)
			fmt.Printf("authorization check failed for project %s: %v\n", config.Name, err)
			continue
		}

		if hasAccess {
			// Get user's role for field-level filtering
			role, err := authzService.GetUserRoleForProject(r.Context(), user.ID, config.Name)
			if err != nil {
				// Default to viewer role if role lookup fails
				role = dashboard.RoleViewer
			}

			// Map and filter fields based on role
			dto := mapPipelineConfigToProjectDTO(&config, user.Namespace)
			dto = dashboard.FilterProjectDTOForRole(dto, role)
			dtos = append(dtos, dto)
		}
	}

	dashboard.RespondSuccess(w, http.StatusOK, dtos)
}

// CreateProjectHandler creates new project (PipelineConfig)
// POST /api/projects
// Authorization: editor or higher (write access required)
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		RepoURL     string `json:"repository_url"`
		Namespace   string `json:"namespace"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate inputs
	if err := validateProjectRequest(req.Name, req.RepoURL); err != nil {
		dashboard.RespondError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	// Use user's namespace if not provided
	namespace := req.Namespace
	if namespace == "" {
		namespace = user.Namespace
	}

	// Create PipelineConfig in Kubernetes
	config := &v1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.PipelineConfigSpec{
			Repository: req.RepoURL,
		},
	}

	if err := k8sClient.CreatePipelineConfig(r.Context(), config); err != nil {
		dashboard.RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", fmt.Sprintf("Failed to create project: %v", err))
		return
	}

	// Map to DTO with webhook URL
	dto := mapPipelineConfigToProjectDTO(config, namespace)
	dto.WebhookURL = generateWebhookURL(config.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// GetWebhookConfigHandler returns webhook configuration for project
// GET /api/projects/{projectId}/webhook
// Authorization: editor or higher (admin access for webhook config)
func GetWebhookConfigHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	// Check authorization: webhook config is admin-level access
	if !CheckProjectAccessAction(w, r, user, projectID, ActionAdmin) {
		return
	}

	// Fetch project
	config, err := k8sClient.GetPipelineConfig(r.Context(), user.Namespace, projectID)
	if err != nil {
		dashboard.RespondError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "Project not found")
		return
	}

	// Generate webhook URL
	webhookURL := generateWebhookURL(config.Name)

	response := map[string]interface{}{
		"project_id":   projectID,
		"webhook_url":  webhookURL,
		"git_platform": "github",
	}

	dashboard.RespondSuccess(w, http.StatusOK, response)
}

// DeleteProjectHandler deletes a project
// DELETE /api/projects/{projectId}
// Authorization: admin only (requires admin role)
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "projectId required")
		return
	}

	// Check authorization: delete requires admin role
	if !CheckProjectAccessAction(w, r, user, projectID, ActionDelete) {
		return
	}

	// Delete from Kubernetes
	if err := k8sClient.DeletePipelineConfig(r.Context(), user.Namespace, projectID); err != nil {
		dashboard.RespondError(w, http.StatusInternalServerError, "DELETE_FAILED", fmt.Sprintf("Failed to delete project: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

// validateProjectRequest validates project creation request
func validateProjectRequest(name, repoURL string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if repoURL == "" {
		return fmt.Errorf("repository URL is required")
	}
	return nil
}

// generateWebhookURL generates webhook URL for a project
func generateWebhookURL(projectName string) string {
	// This should be configurable based on the API server's host/port
	// For now, return a placeholder that can be configured
	return fmt.Sprintf("https://c8s.example.com/webhooks/github/%s", projectName)
}

// mapPipelineConfigToProjectDTO converts a PipelineConfig to ProjectDTO
func mapPipelineConfigToProjectDTO(config *v1alpha1.PipelineConfig, namespace string) *dashboard.ProjectDTO {
	dto := &dashboard.ProjectDTO{
		ID:        config.Name,
		Name:      config.Name,
		RepoURL:   config.Spec.Repository,
		Namespace: namespace,
		CreatedAt: config.CreationTimestamp.Time,
		RunCount:  0, // TODO: Calculate from PipelineRun count
	}

	// If description is in labels/annotations, extract it
	if desc, ok := config.Labels["description"]; ok {
		dto.Description = desc
	}

	return dto
}
