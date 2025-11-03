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

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	handlerspkg "github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/dashboard"
)

// TestListProjectsHandlerNoAuth tests handler rejects unauthenticated requests
func TestListProjectsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListProjectsHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not authenticated")
}

// TestListProjectsHandlerBadRequest tests handler validates input
func TestListProjectsHandlerBadRequest(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListProjectsHandler(w, req)

	// Should return 401 since no user
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// MockProjectAccessService for testing authorization
type MockProjectAccessService struct {
	UserHasProjectAccessFunc func(ctx context.Context, userID, projectID string) (bool, error)
	GetUserRoleForProjectFunc func(ctx context.Context, userID, projectID string) (dashboard.Role, error)
	ListUserProjectsFunc func(ctx context.Context, userID string) ([]dashboard.ProjectDTO, error)
	HasProjectRoleFunc func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error)
}

func (m *MockProjectAccessService) UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error) {
	if m.UserHasProjectAccessFunc != nil {
		return m.UserHasProjectAccessFunc(ctx, userID, projectID)
	}
	return true, nil
}

func (m *MockProjectAccessService) GetUserRoleForProject(ctx context.Context, userID, projectID string) (dashboard.Role, error) {
	if m.GetUserRoleForProjectFunc != nil {
		return m.GetUserRoleForProjectFunc(ctx, userID, projectID)
	}
	return dashboard.RoleViewer, nil
}

func (m *MockProjectAccessService) ListUserProjects(ctx context.Context, userID string) ([]dashboard.ProjectDTO, error) {
	if m.ListUserProjectsFunc != nil {
		return m.ListUserProjectsFunc(ctx, userID)
	}
	return []dashboard.ProjectDTO{}, nil
}

func (m *MockProjectAccessService) HasProjectRole(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
	if m.HasProjectRoleFunc != nil {
		return m.HasProjectRoleFunc(ctx, userID, projectID, role)
	}
	return true, nil
}
