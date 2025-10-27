package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/org/c8s/pkg/dashboard"
)

// ProjectAccessMiddleware enforces per-project authorization
// Checks if user has access to the project specified in route parameters
func ProjectAccessMiddleware(accessSvc dashboard.ProjectAccessService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract projectID from URL path parameter
			projectID := r.PathValue("projectID")
			if projectID == "" {
				// No project scope required, proceed
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has access to this project
			hasAccess, err := accessSvc.UserHasProjectAccess(r.Context(), user.ID, projectID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error checking access: %v", err), http.StatusInternalServerError)
				return
			}

			if !hasAccess {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RoleBasedContextMiddleware attaches user role for project to context
// Allows handlers to render different UI based on permission level
func RoleBasedContextMiddleware(accessSvc dashboard.ProjectAccessService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			projectID := r.PathValue("projectID")
			if projectID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Get user role for this project
			role, err := accessSvc.GetUserRoleForProject(r.Context(), user.ID, projectID)
			if err != nil {
				// Log error but proceed - UI will show limited options
				next.ServeHTTP(w, r)
				return
			}

			// Attach role to context for downstream handlers
			ctx := context.WithValue(r.Context(), "userRole", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserRoleFromContext extracts user role from request context
func GetUserRoleFromContext(ctx context.Context) (dashboard.Role, bool) {
	role, ok := ctx.Value("userRole").(dashboard.Role)
	return role, ok
}
