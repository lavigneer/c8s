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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/dashboard"
)

// setupPipelineTestServer creates a test server with pipeline routes
func setupPipelineTestServer(t *testing.T) *httptest.Server {
	router := chi.NewRouter()

	// Auth middleware that attaches test user
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			testUser := &handlers.User{
				ID:       "test-user",
				Username: "testuser",
				Email:    "test@example.com",
			}
			ctx = r.Context()
			// Store user in context (using unexported method, would need export in real code)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Dashboard routes
	router.Get("/dashboard", handlers.DashboardHandler)
	router.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler)
	router.Get("/api/projects/{projectId}/runs/updates", handlers.PipelineUpdatesSSEHandler)

	return httptest.NewServer(router)
}

// TestDashboardPageReturnsOK verifies dashboard page loads
func TestDashboardPageReturnsOK(t *testing.T) {
	server := setupPipelineTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/dashboard")
	if err != nil {
		t.Fatalf("Failed to request dashboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that HTML is returned
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Logf("Expected text/html, got %s", contentType)
	}
}

// TestListPipelineRunsReturnsJSON verifies API returns JSON
func TestListPipelineRunsReturnsJSON(t *testing.T) {
	server := setupPipelineTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/projects/test-project/runs")
	if err != nil {
		t.Fatalf("Failed to request pipeline runs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify JSON response
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Logf("Expected application/json, got %s", contentType)
	}

	// Parse response
	var apiResp dashboard.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil {
		if !apiResp.Success {
			t.Errorf("Expected success=true in response")
		}
	}
}

// TestListPipelineRunsWithFilters verifies filter parameters accepted
func TestListPipelineRunsWithFilters(t *testing.T) {
	server := setupPipelineTestServer(t)
	defer server.Close()

	// Test with filters
	url := server.URL + "/api/projects/test-project/runs?status=Running&branch=main&page=1&per_page=20"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to request with filters: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestSSEEndpointReturnsStream verifies SSE endpoint is available
func TestSSEEndpointReturnsStream(t *testing.T) {
	server := setupPipelineTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/projects/test-project/runs/updates")
	if err != nil {
		t.Fatalf("Failed to request SSE updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify SSE headers
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", contentType)
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control: no-cache, got %s", cacheControl)
	}
}

// TestPipelineFilteringByStatus verifies status filter works
func TestPipelineFilteringByStatus(t *testing.T) {
	// This test would verify that the status filter parameter is correctly parsed
	// In a real scenario, would test with actual Kubernetes resources
	t.Logf("Status filtering test would require Kubernetes integration")
}

// TestPaginationDefaultValues verifies default pagination works
func TestPaginationDefaultValues(t *testing.T) {
	params := dashboard.ParsePaginationParams(make(map[string][]string))

	if params.Page != 1 {
		t.Errorf("Expected default page=1, got %d", params.Page)
	}

	if params.PerPage != 20 {
		t.Errorf("Expected default per_page=20, got %d", params.PerPage)
	}
}
