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
	"log"
	"net/http"

	"github.com/org/c8s/pkg/dashboard"
)

// Global authorization service - initialized by main package
var authzService dashboard.ProjectAccessService

// InitAuthorizationService initializes the authorization service
func InitAuthorizationService(service dashboard.ProjectAccessService) {
	authzService = service
}

// AuthorizationAction represents an action being performed
type AuthorizationAction string

const (
	ActionRead   AuthorizationAction = "read"
	ActionWrite  AuthorizationAction = "write"
	ActionDelete AuthorizationAction = "delete"
	ActionAdmin  AuthorizationAction = "admin"
)

// CheckProjectAccess verifies user has required role for project
// Returns true if allowed, sends error response and returns false otherwise
func CheckProjectAccess(w http.ResponseWriter, r *http.Request, user *User, projectID string, requiredRole dashboard.Role) bool {
	if authzService == nil {
		log.Printf("ERROR: Authorization service not initialized")
		dashboard.RespondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Authorization service not available")
		return false
	}

	// Check if user has required role
	hasRole, err := authzService.HasProjectRole(r.Context(), user.ID, projectID, requiredRole)
	if err != nil {
		// Log error but don't leak details to client
		log.Printf("ERROR: Failed to check authorization: user=%s project=%s role=%s err=%v",
			user.ID, projectID, requiredRole, err)
		dashboard.RespondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to verify permissions")
		return false
	}

	if !hasRole {
		// User doesn't have required role
		log.Printf("AUTHZ_DENIED: user=%s project=%s required_role=%s", user.ID, projectID, requiredRole)
		dashboard.RespondError(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action")
		return false
	}

	return true
}

// CheckProjectAccessAction verifies user can perform action on project
// Maps actions to required roles: read→viewer, write→editor, delete→admin
func CheckProjectAccessAction(w http.ResponseWriter, r *http.Request, user *User, projectID string, action AuthorizationAction) bool {
	var requiredRole dashboard.Role

	switch action {
	case ActionRead:
		requiredRole = dashboard.RoleViewer
	case ActionWrite:
		requiredRole = dashboard.RoleEditor
	case ActionDelete, ActionAdmin:
		requiredRole = dashboard.RoleAdmin
	default:
		dashboard.RespondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Invalid authorization action")
		return false
	}

	return CheckProjectAccess(w, r, user, projectID, requiredRole)
}

// CheckUserExists verifies user is in context
func CheckUserExists(w http.ResponseWriter, r *http.Request) (*User, bool) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		dashboard.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return nil, false
	}
	return user, true
}

// LogAuthorizationCheck logs authorization decisions for audit trail
func LogAuthorizationCheck(allowed bool, user *User, resource string, action AuthorizationAction, role dashboard.Role) {
	status := "DENIED"
	if allowed {
		status = "ALLOWED"
	}
	log.Printf("AUTHZ: status=%s user=%s resource=%s action=%s role=%s",
		status, user.ID, resource, action, role)
}
