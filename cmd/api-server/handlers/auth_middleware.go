package handlers

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user"

// User represents an authenticated user
type User struct {
	ID        string
	Username  string
	Email     string
	Namespace string
	Roles     []string
}

// AuthMiddleware validates bearer token and attaches user to context
// This is a basic implementation that can be extended with proper JWT/OAuth2 validation
// For HTML page requests without auth, redirects to login
// For API requests without auth, returns 401 Unauthorized
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get token from Authorization header first
		token := extractBearerToken(r.Header.Get("Authorization"))

		// For local dev, also check for auth cookie
		if token == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			// If no token, check if this is an HTML page request or API request
			acceptHeader := r.Header.Get("Accept")
			isHTMLRequest := strings.Contains(acceptHeader, "text/html") || acceptHeader == ""

			if isHTMLRequest {
				// For HTML requests, redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// For API requests, return 401
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Integrate with C8S auth system to validate token
		// For now, we assume token is valid if present
		// In production, validate against JWT or OAuth2 provider

		user := &User{
			ID:        "user-id", // Extract from token
			Username:  "user",    // Extract from token
			Email:     "",        // Extract from token
			Namespace: "default", // TODO: Extract from token or config
			Roles:     []string{},
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't require authentication
// Useful for public endpoints that benefit from auth but work without it
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractBearerToken(authHeader)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: Validate token with C8S auth system
		user := &User{
			ID:        "user-id",
			Username:  "user",
			Email:     "",
			Namespace: "default", // TODO: Extract from token or config
			Roles:     []string{},
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// extractBearerToken extracts token from Authorization header
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

// RequireAuth is a middleware that ensures user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
