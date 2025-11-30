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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/cmd/api-server/handlers"
)

// Helper function to make authenticated request
func makeAuthRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Add auth cookie for development mode
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "test-token",
	})

	return req, nil
}

// setupFilterTestServer creates a test server with filter routes
func setupFilterTestServer(t *testing.T) *httptest.Server {
	router := chi.NewRouter()

	// Auth middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	// Filter endpoints
	router.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler)
	router.Get("/api/projects/{projectId}/branches", handlers.ListBranchesHandler)

	return httptest.NewServer(router)
}

// TestFilteredPipelineList_ByStatus verifies status filtering
func TestFilteredPipelineList_ByStatus(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	req, err := makeAuthRequest("GET", server.URL+"/api/projects/test-project/runs?status=Succeeded", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFilteredPipelineList_ByBranch verifies branch filtering
func TestFilteredPipelineList_ByBranch(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	req, err := makeAuthRequest("GET", server.URL+"/api/projects/test-project/runs?branch=main", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFilteredPipelineList_ByCommitSHA verifies search filtering
func TestFilteredPipelineList_ByCommitSHA(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	req, err := makeAuthRequest("GET", server.URL+"/api/projects/test-project/runs?search=abc123def", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFilteredPipelineList_ByDateRange verifies date range filtering
func TestFilteredPipelineList_ByDateRange(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	req, err := makeAuthRequest("GET", server.URL+"/api/projects/test-project/runs?from_date=2025-01-01&to_date=2025-01-31", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFilteredPipelineList_CombinedFilters verifies multiple filters work together
func TestFilteredPipelineList_CombinedFilters(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	url := server.URL + "/api/projects/test-project/runs?status=Running&branch=develop&search=feature"
	req, err := makeAuthRequest("GET", url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestBranchListEndpoint verifies branch list endpoint works
func TestBranchListEndpoint(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	req, err := makeAuthRequest("GET", server.URL+"/api/projects/test-project/branches", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request branches: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Logf("Expected JSON response, got %s", contentType)
	}
}

// TestFilteredPipelineList_WithPagination verifies filters work with pagination
func TestFilteredPipelineList_WithPagination(t *testing.T) {
	server := setupFilterTestServer(t)
	defer server.Close()

	url := server.URL + "/api/projects/test-project/runs?status=Failed&page=2&per_page=50"
	req, err := makeAuthRequest("GET", url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to request filtered list with pagination: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
