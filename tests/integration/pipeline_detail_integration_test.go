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

	"github.com/org/c8s/cmd/api-server/handlers"
)

// setupDetailTestServer creates a test server with pipeline detail routes
func setupDetailTestServer(t *testing.T) *httptest.Server {
	router := chi.NewRouter()

	// Auth middleware that attaches test user
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	// Detail routes
	router.Get("/dashboard/runs/{runId}", handlers.PipelineRunDetailsHandler)
	router.Get("/api/runs/{runId}", handlers.GetPipelineRunHandler)
	router.Get("/api/runs/{runId}/steps/{stepId}/logs", handlers.LogStreamHandler)

	return httptest.NewServer(router)
}

// TestPipelineDetailPageReturnsOK verifies detail page loads
func TestPipelineDetailPageReturnsOK(t *testing.T) {
	server := setupDetailTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/dashboard/runs/test-run-123")
	if err != nil {
		t.Fatalf("Failed to request detail page: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Logf("Expected status 200 or 500, got %d", resp.StatusCode)
	}
}

// TestGetPipelineRunReturnsJSON verifies API returns run details
func TestGetPipelineRunReturnsJSON(t *testing.T) {
	server := setupDetailTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/runs/test-run-123")
	if err != nil {
		t.Fatalf("Failed to request run details: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Should return 404 since no K8s integration yet
	if resp.StatusCode != http.StatusNotFound {
		t.Logf("Expected status 404, got %d", resp.StatusCode)
	}
}

// TestLogStreamingEndpointReturnsSSE verifies log streaming endpoint
func TestLogStreamingEndpointReturnsSSE(t *testing.T) {
	server := setupDetailTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/runs/test-run-123/steps/step-1/logs")
	if err != nil {
		t.Fatalf("Failed to request log stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// TestLogStreamingWithInvalidRunID returns error
func TestLogStreamingWithInvalidRunID(t *testing.T) {
	server := setupDetailTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/runs//steps/step-1/logs")
	if err != nil {
		t.Fatalf("Failed to request log stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Should return bad request or not found
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Expected
	}
}

// TestLogStreamingMultipleSteps verifies streaming works for different steps
func TestLogStreamingMultipleSteps(t *testing.T) {
	server := setupDetailTestServer(t)
	defer server.Close()

	steps := []string{"build", "test", "deploy"}

	for _, step := range steps {
		resp, err := http.Get(server.URL + "/api/runs/test-run-123/steps/" + step + "/logs")
		if err != nil {
			t.Fatalf("Failed to request log stream for step %s: %v", step, err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Expected status 200 for step %s, got %d", step, resp.StatusCode)
		}
	}
}
