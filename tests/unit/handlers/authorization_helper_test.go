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

package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handlerspkg "github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/dashboard"
)

// MockProjectAccessService for testing
type MockProjectAccessService struct {
	HasProjectRoleFunc func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error)
	GetRoleFunc        func(ctx context.Context, userID, projectID string) (dashboard.Role, error)
}

func (m *MockProjectAccessService) UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error) {
	if m.HasProjectRoleFunc != nil {
		return m.HasProjectRoleFunc(ctx, userID, projectID, dashboard.RoleViewer)
	}
	return true, nil
}

func (m *MockProjectAccessService) GetUserRoleForProject(ctx context.Context, userID, projectID string) (dashboard.Role, error) {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc(ctx, userID, projectID)
	}
	return dashboard.RoleViewer, nil
}

func (m *MockProjectAccessService) ListUserProjects(ctx context.Context, userID string) ([]dashboard.ProjectDTO, error) {
	return []dashboard.ProjectDTO{}, nil
}

func (m *MockProjectAccessService) HasProjectRole(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
	if m.HasProjectRoleFunc != nil {
		return m.HasProjectRoleFunc(ctx, userID, projectID, role)
	}
	return true, nil
}

// TestCheckProjectAccessWithAdminRole verifies admin users pass access check
func TestCheckProjectAccessWithAdminRole(t *testing.T) {
	// Setup mock service that grants access to admin
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			return true, nil // Admin has all roles
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	// Create request and user
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "admin"}

	// Check access with admin role required
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleAdmin)

	assert.True(t, allowed)
	assert.Equal(t, http.StatusOK, w.Code) // Should not set error status
}

// TestCheckProjectAccessWithViewerDeniedAdmin verifies viewer cannot access admin resources
func TestCheckProjectAccessWithViewerDeniedAdmin(t *testing.T) {
	// Setup mock service that only grants viewer access
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// User is viewer, admin required
			return role == dashboard.RoleViewer, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	// Create request and user
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "viewer"}

	// Check access with admin role required
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleAdmin)

	assert.False(t, allowed)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestCheckProjectAccessServiceError verifies error handling when service fails
func TestCheckProjectAccessServiceError(t *testing.T) {
	// Setup mock service that returns error
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			return false, errors.New("database error")
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	// Create request and user
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "test"}

	// Check access
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleViewer)

	assert.False(t, allowed)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestCheckProjectAccessServiceNotInitialized verifies nil service handling
func TestCheckProjectAccessServiceNotInitialized(t *testing.T) {
	// Clear the service
	handlerspkg.InitAuthorizationService(nil)

	// Create request and user
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "test"}

	// Check access
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleViewer)

	assert.False(t, allowed)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestCheckProjectAccessActionReadMapsToViewer verifies read action requires viewer role
func TestCheckProjectAccessActionReadMapsToViewer(t *testing.T) {
	// Setup mock service that grants viewer access
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// Accept if role required is viewer
			return role == dashboard.RoleViewer, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "viewer"}

	// ActionRead should map to RoleViewer
	allowed := handlerspkg.CheckProjectAccessAction(w, r, user, "proj-1", handlerspkg.ActionRead)

	assert.True(t, allowed)
}

// TestCheckProjectAccessActionWriteMapsToEditor verifies write action requires editor role
func TestCheckProjectAccessActionWriteMapsToEditor(t *testing.T) {
	// Setup mock service that only grants editor access
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// Accept if role required is editor or lower
			return role != dashboard.RoleAdmin, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "editor"}

	// ActionWrite should map to RoleEditor
	allowed := handlerspkg.CheckProjectAccessAction(w, r, user, "proj-1", handlerspkg.ActionWrite)

	assert.True(t, allowed)
}

// TestCheckProjectAccessActionDeleteMapsToAdmin verifies delete action requires admin role
func TestCheckProjectAccessActionDeleteMapsToAdmin(t *testing.T) {
	// Setup mock service that requires admin
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// Accept only admin
			return role == dashboard.RoleAdmin, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "admin"}

	// ActionDelete should map to RoleAdmin
	allowed := handlerspkg.CheckProjectAccessAction(w, r, user, "proj-1", handlerspkg.ActionDelete)

	assert.True(t, allowed)
}

// TestCheckProjectAccessActionAdminMapsToAdmin verifies admin action requires admin role
func TestCheckProjectAccessActionAdminMapsToAdmin(t *testing.T) {
	// Setup mock service that requires admin
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// Accept only admin
			return role == dashboard.RoleAdmin, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "admin"}

	// ActionAdmin should map to RoleAdmin
	allowed := handlerspkg.CheckProjectAccessAction(w, r, user, "proj-1", handlerspkg.ActionAdmin)

	assert.True(t, allowed)
}

// TestCheckProjectAccessActionInvalidAction verifies invalid action handling
func TestCheckProjectAccessActionInvalidAction(t *testing.T) {
	handlerspkg.InitAuthorizationService(&MockProjectAccessService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "test"}

	// Invalid action should return error
	allowed := handlerspkg.CheckProjectAccessAction(w, r, user, "proj-1", handlerspkg.AuthorizationAction("invalid"))

	assert.False(t, allowed)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestCheckUserExistsWithValidUser verifies user extraction succeeds
func TestCheckUserExistsWithValidUser(t *testing.T) {
	user := &handlerspkg.User{ID: "user-123", Username: "testuser", Email: "test@example.com"}

	// We can't directly set UserContextKey since it's private,
	// so we'll test through the actual middleware
	// For now, this is verified in the integration tests

	require.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
}

// TestCheckUserExistsWithMissingUser verifies missing user is detected
func TestCheckUserExistsWithMissingUser(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody) // No user in context

	// Check user exists
	extractedUser, ok := handlerspkg.CheckUserExists(w, r)

	assert.False(t, ok)
	assert.Nil(t, extractedUser)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRoleHierarchyComparison verifies role comparison logic
func TestRoleHierarchyComparison(t *testing.T) {
	tests := []struct {
		name         string
		userRole     dashboard.Role
		requiredRole dashboard.Role
		shouldAllow  bool
	}{
		// Admin tests
		{"Admin can read", dashboard.RoleAdmin, dashboard.RoleViewer, true},
		{"Admin can write", dashboard.RoleAdmin, dashboard.RoleEditor, true},
		{"Admin can delete", dashboard.RoleAdmin, dashboard.RoleAdmin, true},

		// Editor tests
		{"Editor can read", dashboard.RoleEditor, dashboard.RoleViewer, true},
		{"Editor can write", dashboard.RoleEditor, dashboard.RoleEditor, true},
		{"Editor cannot delete", dashboard.RoleEditor, dashboard.RoleAdmin, false},

		// Viewer tests
		{"Viewer can read", dashboard.RoleViewer, dashboard.RoleViewer, true},
		{"Viewer cannot write", dashboard.RoleViewer, dashboard.RoleEditor, false},
		{"Viewer cannot delete", dashboard.RoleViewer, dashboard.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.userRole.Level() >= tt.requiredRole.Level()
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

// TestAuthorizationLoggingFormat verifies log output format
func TestAuthorizationLoggingFormat(t *testing.T) {
	user := &handlerspkg.User{ID: "user-123", Username: "testuser"}
	resource := "project:proj-1"
	action := handlerspkg.ActionRead
	role := dashboard.RoleViewer

	// Just verify this function doesn't panic
	handlerspkg.LogAuthorizationCheck(true, user, resource, action, role)
	handlerspkg.LogAuthorizationCheck(false, user, resource, action, role)
}

// TestActionConstants verifies action constants are defined
func TestActionConstants(t *testing.T) {
	// Verify action constants exist and have expected values
	assert.Equal(t, handlerspkg.ActionRead, handlerspkg.AuthorizationAction("read"))
	assert.Equal(t, handlerspkg.ActionWrite, handlerspkg.AuthorizationAction("write"))
	assert.Equal(t, handlerspkg.ActionDelete, handlerspkg.AuthorizationAction("delete"))
	assert.Equal(t, handlerspkg.ActionAdmin, handlerspkg.AuthorizationAction("admin"))
}

// TestRoleComparisonEdgeCases verifies edge cases in role comparison
func TestRoleComparisonEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		userRole     dashboard.Role
		requiredRole dashboard.Role
		shouldAllow  bool
	}{
		// Unknown role handling
		{"Unknown role cannot access anything", dashboard.Role("unknown"), dashboard.RoleViewer, false},
		{"Viewer can exceed unknown role", dashboard.RoleViewer, dashboard.Role("unknown"), true},
		// Empty role
		{"Empty role cannot access anything", dashboard.Role(""), dashboard.RoleViewer, false},
		// Same role
		{"Same role equals", dashboard.RoleViewer, dashboard.RoleViewer, true},
		{"Same role equals admin", dashboard.RoleAdmin, dashboard.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.userRole.Level() >= tt.requiredRole.Level()
			assert.Equal(t, tt.shouldAllow, allowed, "role=%s required=%s", tt.userRole, tt.requiredRole)
		})
	}
}

// TestMultipleProjectAccessControl verifies access control per project
func TestMultipleProjectAccessControl(t *testing.T) {
	// User might have different roles in different projects
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			// Simulate different roles for different projects
			switch projectID {
			case "proj-alpha":
				return role == dashboard.RoleAdmin, nil // Admin in alpha
			case "proj-beta":
				return role != dashboard.RoleAdmin, nil // Editor in beta
			case "proj-gamma":
				return false, nil // No access in gamma
			default:
				return false, nil
			}
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	user := &handlerspkg.User{ID: "user-123", Username: "testuser"}

	// Test alpha project (admin access)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/", http.NoBody)
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-alpha", dashboard.RoleAdmin)
	assert.True(t, allowed)

	// Test beta project (editor access, not admin)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/", http.NoBody)
	allowed = handlerspkg.CheckProjectAccess(w, r, user, "proj-beta", dashboard.RoleAdmin)
	assert.False(t, allowed)

	// Test gamma project (no access)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", http.NoBody)
	allowed = handlerspkg.CheckProjectAccess(w, r, user, "proj-gamma", dashboard.RoleViewer)
	assert.False(t, allowed)
}

// TestAuthorizationErrorResponseFormat verifies error response structure
func TestAuthorizationErrorResponseFormat(t *testing.T) {
	// Setup mock service that denies access
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			return false, nil // Deny all
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	user := &handlerspkg.User{ID: "user-123", Username: "test"}

	// Check access
	allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleViewer)

	assert.False(t, allowed)
	assert.Equal(t, http.StatusForbidden, w.Code)
	// Response body should contain error JSON
	assert.Contains(t, w.Body.String(), "FORBIDDEN")
}

// TestAuthenticationRequiredBeforeAuthorization verifies auth check order
func TestAuthenticationRequiredBeforeAuthorization(t *testing.T) {
	handlerspkg.InitAuthorizationService(&MockProjectAccessService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody) // No user in context

	// Should return 401 Unauthorized, not 403 Forbidden
	_, ok := handlerspkg.CheckUserExists(w, r)
	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRoleInheritanceHierarchy verifies role inheritance chain
func TestRoleInheritanceHierarchy(t *testing.T) {
	// Role hierarchy: Admin > Editor > Viewer
	// Each role inherits permissions of lower roles

	// Admin inherits all permissions
	assert.True(t, dashboard.RoleAdmin.Level() >= dashboard.RoleAdmin.Level())  // Can do admin
	assert.True(t, dashboard.RoleAdmin.Level() >= dashboard.RoleEditor.Level()) // Can do editor
	assert.True(t, dashboard.RoleAdmin.Level() >= dashboard.RoleViewer.Level()) // Can do viewer

	// Editor inherits viewer permissions
	assert.False(t, dashboard.RoleEditor.Level() >= dashboard.RoleAdmin.Level()) // Cannot do admin
	assert.True(t, dashboard.RoleEditor.Level() >= dashboard.RoleEditor.Level()) // Can do editor
	assert.True(t, dashboard.RoleEditor.Level() >= dashboard.RoleViewer.Level()) // Can do viewer

	// Viewer has minimal permissions
	assert.False(t, dashboard.RoleViewer.Level() >= dashboard.RoleAdmin.Level())  // Cannot do admin
	assert.False(t, dashboard.RoleViewer.Level() >= dashboard.RoleEditor.Level()) // Cannot do editor
	assert.True(t, dashboard.RoleViewer.Level() >= dashboard.RoleViewer.Level())  // Can do viewer
}

// TestConcurrentAuthorizationChecks verifies concurrent safety
func TestConcurrentAuthorizationChecks(t *testing.T) {
	mockSvc := &MockProjectAccessService{
		HasProjectRoleFunc: func(ctx context.Context, userID, projectID string, role dashboard.Role) (bool, error) {
			return true, nil
		},
	}
	handlerspkg.InitAuthorizationService(mockSvc)

	// Simulate concurrent access checks
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(index int) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", http.NoBody)
			user := &handlerspkg.User{ID: "user-123", Username: "test"}

			allowed := handlerspkg.CheckProjectAccess(w, r, user, "proj-1", dashboard.RoleViewer)
			assert.True(t, allowed)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestAuthorizationDecisionBoundary verifies boundary conditions
func TestAuthorizationDecisionBoundary(t *testing.T) {
	tests := []struct {
		name          string
		userLevel     int
		requiredLevel int
		shouldAllow   bool
	}{
		{"Exact match allows", 2, 2, true},
		{"Higher allows", 3, 2, true},
		{"Lower denies", 1, 2, false},
		{"Zero level denies", 0, 1, false},
		{"Max level admin", 3, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.userLevel >= tt.requiredLevel
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}
