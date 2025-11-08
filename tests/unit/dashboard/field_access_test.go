/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dashboard

import (
	"testing"
	"time"

	"github.com/org/c8s/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

// TestFilterProjectDTOViewerRole verifies viewers get limited project fields
func TestFilterProjectDTOViewerRole(t *testing.T) {
	now := time.Now()
	project := &dashboard.ProjectDTO{
		ID:          "proj-1",
		Name:        "My Project",
		Description: "A test project",
		RepoURL:     "https://github.com/example/repo",
		WebhookURL:  "https://c8s.example.com/webhooks/github/proj-1",
		Namespace:   "test-ns",
		CreatedAt:   now,
		RunCount:    5,
	}

	filtered := dashboard.FilterProjectDTOForRole(project, dashboard.RoleViewer)

	// Viewers should see basic fields
	assert.Equal(t, "proj-1", filtered.ID)
	assert.Equal(t, "My Project", filtered.Name)
	assert.Equal(t, "https://github.com/example/repo", filtered.RepoURL)
	assert.Equal(t, "test-ns", filtered.Namespace)
	assert.Equal(t, "A test project", filtered.Description)
	assert.Equal(t, 5, filtered.RunCount)

	// Viewers should NOT see webhook URL
	assert.Empty(t, filtered.WebhookURL)
}

// TestFilterProjectDTOEditorRole verifies editors get most project fields
func TestFilterProjectDTOEditorRole(t *testing.T) {
	now := time.Now()
	lastRun := now.Add(-24 * time.Hour)
	project := &dashboard.ProjectDTO{
		ID:          "proj-1",
		Name:        "My Project",
		Description: "A test project",
		RepoURL:     "https://github.com/example/repo",
		WebhookURL:  "https://c8s.example.com/webhooks/github/proj-1",
		Namespace:   "test-ns",
		CreatedAt:   now,
		LastRunAt:   &lastRun,
		RunCount:    5,
	}

	filtered := dashboard.FilterProjectDTOForRole(project, dashboard.RoleEditor)

	// Editors should see all public fields
	assert.Equal(t, "proj-1", filtered.ID)
	assert.Equal(t, "My Project", filtered.Name)
	assert.Equal(t, "A test project", filtered.Description)

	// Editors should see webhook URL
	assert.Equal(t, "https://c8s.example.com/webhooks/github/proj-1", filtered.WebhookURL)
	assert.NotNil(t, filtered.LastRunAt)
}

// TestFilterProjectDTOAdminRole verifies admins see all project fields
func TestFilterProjectDTOAdminRole(t *testing.T) {
	now := time.Now()
	lastRun := now.Add(-24 * time.Hour)
	project := &dashboard.ProjectDTO{
		ID:          "proj-1",
		Name:        "My Project",
		Description: "A test project",
		RepoURL:     "https://github.com/example/repo",
		WebhookURL:  "https://c8s.example.com/webhooks/github/proj-1",
		Namespace:   "test-ns",
		CreatedAt:   now,
		LastRunAt:   &lastRun,
		RunCount:    5,
	}

	filtered := dashboard.FilterProjectDTOForRole(project, dashboard.RoleAdmin)

	// Admins should see all fields
	assert.Equal(t, "proj-1", filtered.ID)
	assert.Equal(t, "My Project", filtered.Name)
	assert.Equal(t, "A test project", filtered.Description)
	assert.Equal(t, "https://c8s.example.com/webhooks/github/proj-1", filtered.WebhookURL)
	assert.NotNil(t, filtered.LastRunAt)
}

// TestFilterProjectDTOsForRole verifies multiple projects are filtered
func TestFilterProjectDTOsForRole(t *testing.T) {
	projects := []*dashboard.ProjectDTO{
		{
			ID:         "proj-1",
			Name:       "Project 1",
			WebhookURL: "https://example.com/1",
		},
		{
			ID:         "proj-2",
			Name:       "Project 2",
			WebhookURL: "https://example.com/2",
		},
	}

	filtered := dashboard.FilterProjectDTOsForRole(projects, dashboard.RoleViewer)

	assert.Len(t, filtered, 2)
	assert.Empty(t, filtered[0].WebhookURL)
	assert.Empty(t, filtered[1].WebhookURL)
}

// TestFilterPipelineRunDTOViewerRole verifies viewers cannot see author email
func TestFilterPipelineRunDTOViewerRole(t *testing.T) {
	now := time.Now()
	run := &dashboard.PipelineRunDTO{
		ID:          "run-1",
		ProjectID:   "proj-1",
		Name:        "Run 1",
		Status:      "Succeeded",
		Author:      "John Doe",
		AuthorEmail: "john@example.com",
		TriggeredAt: now,
	}

	filtered := dashboard.FilterPipelineRunDTOForRole(run, dashboard.RoleViewer)

	// Viewers should see author but not email
	assert.Equal(t, "John Doe", filtered.Author)
	assert.Empty(t, filtered.AuthorEmail)
}

// TestFilterPipelineRunDTOEditorRole verifies editors see author email
func TestFilterPipelineRunDTOEditorRole(t *testing.T) {
	now := time.Now()
	run := &dashboard.PipelineRunDTO{
		ID:          "run-1",
		ProjectID:   "proj-1",
		Name:        "Run 1",
		Status:      "Succeeded",
		Author:      "John Doe",
		AuthorEmail: "john@example.com",
		TriggeredAt: now,
	}

	filtered := dashboard.FilterPipelineRunDTOForRole(run, dashboard.RoleEditor)

	// Editors should see author email
	assert.Equal(t, "John Doe", filtered.Author)
	assert.Equal(t, "john@example.com", filtered.AuthorEmail)
}

// TestFilterArtifactDTOViewerRole verifies viewers cannot download artifacts
func TestFilterArtifactDTOViewerRole(t *testing.T) {
	now := time.Now()
	artifact := &dashboard.ArtifactDTO{
		ID:        "artifact-1",
		Name:      "report.html",
		Type:      "report",
		MimeType:  "text/html",
		SizeBytes: 12345,
		URL:       "https://storage.example.com/artifacts/artifact-1",
		CreatedAt: now,
	}

	filtered := dashboard.FilterArtifactDTOForRole(artifact, dashboard.RoleViewer)

	// Viewers should see metadata but not URL
	assert.Equal(t, "artifact-1", filtered.ID)
	assert.Equal(t, "report.html", filtered.Name)
	assert.Equal(t, int64(12345), filtered.SizeBytes)
	assert.Empty(t, filtered.URL)
}

// TestFilterArtifactDTOEditorRole verifies editors can download artifacts
func TestFilterArtifactDTOEditorRole(t *testing.T) {
	now := time.Now()
	artifact := &dashboard.ArtifactDTO{
		ID:        "artifact-1",
		Name:      "report.html",
		Type:      "report",
		MimeType:  "text/html",
		SizeBytes: 12345,
		URL:       "https://storage.example.com/artifacts/artifact-1",
		CreatedAt: now,
	}

	filtered := dashboard.FilterArtifactDTOForRole(artifact, dashboard.RoleEditor)

	// Editors should see URL
	assert.Equal(t, "artifact-1", filtered.ID)
	assert.Equal(t, "report.html", filtered.Name)
	assert.Equal(t, "https://storage.example.com/artifacts/artifact-1", filtered.URL)
}

// TestFilterArtifactDTOsForRole verifies multiple artifacts are filtered
func TestFilterArtifactDTOsForRole(t *testing.T) {
	artifacts := []*dashboard.ArtifactDTO{
		{
			ID:   "art-1",
			Name: "report1.html",
			URL:  "https://storage.example.com/art-1",
		},
		{
			ID:   "art-2",
			Name: "report2.html",
			URL:  "https://storage.example.com/art-2",
		},
	}

	// Viewer role filters out URLs
	filtered := dashboard.FilterArtifactDTOsForRole(artifacts, dashboard.RoleViewer)

	assert.Len(t, filtered, 2)
	assert.Empty(t, filtered[0].URL)
	assert.Empty(t, filtered[1].URL)

	// Editor role keeps URLs
	filtered = dashboard.FilterArtifactDTOsForRole(artifacts, dashboard.RoleEditor)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "https://storage.example.com/art-1", filtered[0].URL)
	assert.Equal(t, "https://storage.example.com/art-2", filtered[1].URL)
}

// TestFilterNilProjectDTO verifies nil input handling
func TestFilterNilProjectDTO(t *testing.T) {
	var project *dashboard.ProjectDTO
	filtered := dashboard.FilterProjectDTOForRole(project, dashboard.RoleViewer)
	assert.Nil(t, filtered)
}

// TestFilterNilProjectDTOs verifies nil slice handling
func TestFilterNilProjectDTOs(t *testing.T) {
	var projects []*dashboard.ProjectDTO
	filtered := dashboard.FilterProjectDTOsForRole(projects, dashboard.RoleViewer)
	assert.Nil(t, filtered)
}

// TestFilterEmptyProjectDTOs verifies empty slice handling
func TestFilterEmptyProjectDTOs(t *testing.T) {
	projects := []*dashboard.ProjectDTO{}
	filtered := dashboard.FilterProjectDTOsForRole(projects, dashboard.RoleViewer)
	assert.Empty(t, filtered)
}

// TestFilterStepDTOAllRoles verifies steps show same fields for all roles
func TestFilterStepDTOAllRoles(t *testing.T) {
	now := time.Now()
	step := &dashboard.StepDTO{
		ID:            "step-1",
		Name:          "Build",
		Status:        "Succeeded",
		Image:         "golang:1.21",
		Commands:      []string{"go build"},
		StartedAt:     &now,
		CPURequest:    "500m",
		MemoryRequest: "512Mi",
	}

	// All roles should see step details equally
	viewerFiltered := dashboard.FilterStepDTOForRole(step, dashboard.RoleViewer)
	editorFiltered := dashboard.FilterStepDTOForRole(step, dashboard.RoleEditor)
	adminFiltered := dashboard.FilterStepDTOForRole(step, dashboard.RoleAdmin)

	assert.Equal(t, viewerFiltered.CPURequest, editorFiltered.CPURequest)
	assert.Equal(t, editorFiltered.CPURequest, adminFiltered.CPURequest)
	assert.Equal(t, "500m", viewerFiltered.CPURequest)
}

// TestFilterLogStreamDTOAllRoles verifies logs are visible to all roles
func TestFilterLogStreamDTOAllRoles(t *testing.T) {
	now := time.Now()
	log := &dashboard.LogStreamDTO{
		StepID:    "step-1",
		Content:   "Build started",
		Timestamp: now,
		IsError:   false,
	}

	// All roles should see full log content
	viewerFiltered := dashboard.FilterLogStreamDTOForRole(log, dashboard.RoleViewer)
	editorFiltered := dashboard.FilterLogStreamDTOForRole(log, dashboard.RoleEditor)
	adminFiltered := dashboard.FilterLogStreamDTOForRole(log, dashboard.RoleAdmin)

	assert.Equal(t, "Build started", viewerFiltered.Content)
	assert.Equal(t, "Build started", editorFiltered.Content)
	assert.Equal(t, "Build started", adminFiltered.Content)
}
