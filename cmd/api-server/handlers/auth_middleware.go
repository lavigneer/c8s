package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/org/c8s/cmd/api-server/auth"
)

type contextKey string

const userContextKey contextKey = "user"

// User represents an authenticated user
// Extracted from JWT claims in the authorization middleware
type User struct {
	ID        string
	Username  string
	Email     string
	Namespace string
	Roles     []string
}

// Global JWT validator - initialized at startup
var jwtValidator *auth.Validator
var jwtNoOpValidator *auth.NoOpValidator

// InitAuthValidator initializes the JWT validator with configuration
// Should be called once at application startup before serving requests
func InitAuthValidator(config *auth.Config) error {
	var err error
	jwtValidator, err = auth.NewValidator(config)
	if err != nil {
		return err
	}
	return nil
}

// UseNoOpValidator enables development-mode validator that accepts any token
// WARNING: Should never be used in production!
func UseNoOpValidator() {
	jwtNoOpValidator = auth.NewNoOpValidator()
}

// ResetAuthValidators resets both validators to nil (for testing purposes)
// This is useful for tests that use global validator state and need cleanup
func ResetAuthValidators() {
	jwtValidator = nil
	jwtNoOpValidator = nil
}

// AuthMiddleware validates JWT bearer token and attaches user to context
// Supports token extraction from:
// 1. Authorization header (Bearer <token>)
// 2. auth_token cookie (for development/UI)
//
// For HTML page requests without auth, redirects to login.
// For API requests without auth, returns 401 Unauthorized.
//
// Token validation includes:
// - Signature verification
// - Expiration checking
// - Issuer and audience validation
// - Required claims enforcement
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or cookie
		token := extractBearerToken(r.Header.Get("Authorization"))

		// For local dev, also check for auth cookie
		if token == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				token = cookie.Value
			}
		}

		// If no token found, handle based on request type
		if token == "" {
			// Check if this is an HTML page request or API request
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

		// Validate token and extract user information
		var user *User

		// Use NoOp validator if enabled (development only)
		if jwtNoOpValidator != nil {
			authUser, validErr := jwtNoOpValidator.ValidateTokenAndGetUser(token)
			if validErr != nil {
				log.Printf("NoOp validator error: %v", validErr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user = &User{
				ID:        authUser.ID,
				Username:  authUser.Username,
				Email:     authUser.Email,
				Namespace: authUser.Namespace,
				Roles:     authUser.Roles,
			}
		} else if jwtValidator != nil {
			// Use real JWT validator
			authUser, validErr := jwtValidator.ValidateTokenAndGetUser(token)
			if validErr != nil {
				log.Printf("Token validation failed: %v", validErr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user = &User{
				ID:        authUser.ID,
				Username:  authUser.Username,
				Email:     authUser.Email,
				Namespace: authUser.Namespace,
				Roles:     authUser.Roles,
			}
		} else {
			// No validator configured - should not happen if properly initialized
			log.Printf("WARNING: JWT validator not initialized")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Attach user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't require authentication
// Useful for public endpoints that benefit from auth but work without it.
// If valid auth is provided, the user is attached to context.
// If no auth or invalid auth, the request continues without user context.
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		token := extractBearerToken(r.Header.Get("Authorization"))

		// Also check auth cookie if no header token
		if token == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				token = cookie.Value
			}
		}

		// If no token found, continue without auth
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token, but continue if validation fails
		var user *auth.User
		if jwtNoOpValidator != nil {
			// Development mode
			authUser, err := jwtNoOpValidator.ValidateTokenAndGetUser(token)
			if err != nil {
				// Validation failed, continue without auth
				next.ServeHTTP(w, r)
				return
			}
			user = authUser
		} else if jwtValidator != nil {
			// Production mode
			authUser, err := jwtValidator.ValidateTokenAndGetUser(token)
			if err != nil {
				// Validation failed, continue without auth
				next.ServeHTTP(w, r)
				return
			}
			user = authUser
		} else {
			// No validator configured, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Token validated successfully - attach user to context
		handlerUser := &User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Namespace: user.Namespace,
			Roles:     user.Roles,
		}
		ctx := context.WithValue(r.Context(), userContextKey, handlerUser)
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
