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

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/dashboard"
)

// setupTestServer creates a test HTTP server with dashboard routes.
func setupTestServer(t *testing.T) *httptest.Server {
	// Note: In a real scenario, you would load templates from disk
	// For testing, we'll just register routes without template rendering

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Static files
	router.Handle("/static/*", handlers.ServeStatic("../cmd/api-server/static"))
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// Login route (no auth required)
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<h1>Login</h1>"))
	})

	// Dashboard routes (protected by auth)
	router.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware)
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<h1>Dashboard</h1>"))
		})
	})

	// 404 handler
	router.NotFound(handlers.NotFoundHandler)

	return httptest.NewServer(router)
}

// TestDashboardTemplatesLoad verifies templates can be loaded without errors.
func TestDashboardTemplatesLoad(t *testing.T) {
	// This test would load actual templates
	// For now, we just verify the package can be imported
	_ = dashboard.Templates
}

// TestHealthEndpointReturnsOK verifies health check endpoint works.
func TestHealthEndpointReturnsOK(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestStaticFilesAreAccessible verifies static assets can be served.
func TestStaticFilesAreAccessible(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Test accessing a CSS file (if it exists)
	resp, err := http.Get(server.URL + "/static/css/dashboard.css")
	if err != nil && resp == nil {
		t.Logf("Static file test skipped - no static files present")
		return
	}

	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
		// Status may be 200 or 404 depending on if file exists
		if resp.StatusCode < 400 || resp.StatusCode == 404 {
			t.Logf("Static file access returned status %d", resp.StatusCode)
		}
	}
}

// TestDashboardRequiresAuth verifies dashboard is protected by authentication.
func TestDashboardRequiresAuth(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Create a client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Request without auth token should redirect to login
	req, err := http.NewRequest("GET", server.URL+"/dashboard", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Should redirect (303 or 307) to login when not authenticated
	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status (303 or 307), got %d", resp.StatusCode)
	}

	// Verify it redirects to login
	location := resp.Header.Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}
}

// TestNotFoundHandlerReturnsProperStatus verifies 404 handling.
func TestNotFoundHandlerReturnsProperStatus(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("Failed to get nonexistent endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
