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

package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user"

// Middleware handles HTTP authentication using dependency injection
type Middleware struct {
	Validator ValidatorInterface
}

// NewMiddleware creates a new auth middleware with the given validator
func NewMiddleware(validator ValidatorInterface) *Middleware {
	return &Middleware{
		Validator: validator,
	}
}

// Handler validates JWT bearer token and attaches user to context
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
func (m *Middleware) Handler(next http.Handler) http.Handler {
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
		user, err := m.Validator.ValidateTokenAndGetUser(token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Attach user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalHandler is like Handler but doesn't require authentication
// Useful for public endpoints that benefit from auth but work without it.
// If valid auth is provided, the user is attached to context.
// If no auth or invalid auth, the request continues without user context.
func (m *Middleware) OptionalHandler(next http.Handler) http.Handler {
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
		user, err := m.Validator.ValidateTokenAndGetUser(token)
		if err != nil {
			// Validation failed, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Token validated successfully - attach user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireHandler ensures user is authenticated (must be used after Handler)
func (m *Middleware) RequireHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
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
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
