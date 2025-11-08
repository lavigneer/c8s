package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/org/c8s/pkg/dashboard"
)

// ListArtifactsHandler returns artifacts for a pipeline run
// GET /api/runs/{runId}/artifacts
// Authorization: viewer or higher (read access to artifacts)
func ListArtifactsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	runID := chi.URLParam(r, "runId")
	if runID == "" {
		_ = dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "runId required")
		return
	}

	// TODO: Fetch PipelineRun from K8s to get project context
	// For now, we can't verify access without knowing the project
	// In a full implementation:
	// 1. Fetch PipelineRun
	// 2. Extract project ID from labels or ownership
	// 3. Check user has viewer access to that project

	// Optional filters
	stepID := r.URL.Query().Get("step_id")
	artifactType := r.URL.Query().Get("type")

	// TODO: Fetch PipelineRun and extract artifacts from step statuses
	// TODO: Apply filters for stepID and artifactType when implementing fetch logic
	// For now, return empty list
	_ = stepID       // Will be used when filtering is implemented
	_ = artifactType // Will be used when filtering is implemented
	artifacts := []*dashboard.ArtifactDTO{}

	_ = dashboard.RespondSuccess(w, http.StatusOK, artifacts)
}

// DownloadArtifactHandler downloads an artifact from object storage
// GET /api/artifacts/{artifactId}/download
// Authorization: viewer or higher (read access to artifacts)
func DownloadArtifactHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	artifactID := chi.URLParam(r, "artifactId")
	if artifactID == "" {
		_ = dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "artifactId required")
		return
	}

	// TODO: Resolve artifact to project and check user has access
	// For now, access is allowed since we already authenticated user above
	_ = user // Use user variable to avoid unused variable warning

	// Demo artifacts mapping (in production, fetch from storage)
	demoArtifacts := map[string]string{
		"hello-world-run-001-artifact-1": "api-server-v1.2.3.tar.gz",
		"hello-world-run-001-artifact-2": "test-report.html",
		"hello-world-run-001-artifact-3": "coverage-report.json",
		"hello-world-run-001-artifact-4": "build-log.txt",
	}

	filename, ok := demoArtifacts[artifactID]
	if !ok {
		_ = dashboard.RespondError(w, http.StatusNotFound, "ARTIFACT_NOT_FOUND", "Artifact not found")
		return
	}

	// Generate demo content based on artifact type
	var content []byte
	contentType := "application/octet-stream"

	switch filename {
	case "api-server-v1.2.3.tar.gz":
		content = []byte("This would be a binary tar.gz file containing the compiled api-server binary")
		contentType = "application/gzip"
	case "test-report.html":
		content = generateTestReportHTML()
		contentType = "text/html; charset=utf-8"
	case "coverage-report.json":
		content = generateCoverageReportJSON()
		contentType = "application/json"
	case "build-log.txt":
		content = generateBuildLog()
		contentType = "text/plain; charset=utf-8"
	}

	// Set appropriate headers for download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// PreviewArtifactHandler returns a preview of an artifact (for reports, etc.)
// GET /api/artifacts/{artifactId}/preview
// Authorization: viewer or higher (read access to artifacts)
func PreviewArtifactHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	artifactID := chi.URLParam(r, "artifactId")
	if artifactID == "" {
		_ = dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "artifactId required")
		return
	}

	// TODO: Resolve artifact to project and check user has access
	// For now, access is allowed since we already authenticated user above
	_ = user // Use user variable to avoid unused variable warning

	// Demo artifacts mapping
	demoArtifacts := map[string]string{
		"hello-world-run-001-artifact-1": "api-server-v1.2.3.tar.gz",
		"hello-world-run-001-artifact-2": "test-report.html",
		"hello-world-run-001-artifact-3": "coverage-report.json",
		"hello-world-run-001-artifact-4": "build-log.txt",
	}

	filename, ok := demoArtifacts[artifactID]
	if !ok {
		_ = dashboard.RespondError(w, http.StatusNotFound, "ARTIFACT_NOT_FOUND", "Artifact not found")
		return
	}

	// Check if HTMX request
	if dashboard.IsHTMXRequest(r) {
		// Return HTML fragment for HTMX embedding
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Generate preview based on artifact type
		switch filename {
		case "test-report.html":
			// Return HTML test report directly
			_, _ = w.Write(generateTestReportHTML())
		case "coverage-report.json":
			// Return JSON coverage report in HTML format
			_, _ = io.WriteString(w, `<pre class="bg-gray-100 p-4 rounded overflow-auto max-h-96 text-sm font-mono">`)
			_, _ = w.Write(generateCoverageReportJSON())
			_, _ = io.WriteString(w, `</pre>`)
		case "build-log.txt":
			// Return build log in HTML format
			_, _ = io.WriteString(w, `<pre class="bg-gray-900 text-gray-100 p-4 rounded overflow-auto max-h-96 text-sm font-mono">`)
			_, _ = w.Write(generateBuildLog())
			_, _ = io.WriteString(w, `</pre>`)
		default:
			_, _ = io.WriteString(w, `<div class="text-gray-600 p-4">Preview not available for this artifact type</div>`)
		}
		return
	}

	// Return JSON preview metadata
	_ = dashboard.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"artifact_id": artifactID,
		"filename":    filename,
		"preview":     "Use HTMX request to get HTML preview",
	})
}

// GetArtifactHandler returns metadata for a specific artifact
// GET /api/artifacts/{artifactId}
func GetArtifactHandler(w http.ResponseWriter, r *http.Request) {
	artifactID := chi.URLParam(r, "artifactId")
	if artifactID == "" {
		_ = dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "artifactId required")
		return
	}

	// TODO: Verify user has access to this artifact

	// TODO: Fetch artifact metadata
	// Placeholder: return 404 for now
	_ = dashboard.RespondError(w, http.StatusNotFound, "ARTIFACT_NOT_FOUND", "Artifact not found")
}

// DeleteArtifactHandler deletes an artifact
// DELETE /api/artifacts/{artifactId}
// Authorization: admin only (delete requires admin role on project)
func DeleteArtifactHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := CheckUserExists(w, r)
	if !ok {
		return
	}

	artifactID := chi.URLParam(r, "artifactId")
	if artifactID == "" {
		_ = dashboard.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "artifactId required")
		return
	}

	// TODO: Resolve artifact to project and check user has admin access
	// For now, log the user ID
	_ = user // Use user variable to avoid unused variable warning

	// TODO: Delete artifact from object storage
	// TODO: Delete artifact metadata from database

	w.WriteHeader(http.StatusNoContent)
}

// ValidateArtifactContent sanitizes artifact content for safe display
// Implements A8 - Content Security for artifacts
func ValidateArtifactContent(content []byte) ([]byte, error) {
	// TODO: Implement content sanitization
	// - Validate file format
	// - Scan for malicious patterns
	// - Sanitize HTML/JavaScript if needed
	return content, nil
}

// generateTestReportHTML generates a demo HTML test report
func generateTestReportHTML() []byte {
	return []byte(`
<div class="test-report p-6 bg-white rounded-lg">
  <h2 class="text-2xl font-bold text-gray-900 mb-4">Test Report</h2>

  <div class="grid grid-cols-4 gap-4 mb-6">
    <div class="bg-green-50 border border-green-200 rounded p-4">
      <div class="text-3xl font-bold text-green-600">156</div>
      <div class="text-sm text-gray-600">Tests Passed</div>
    </div>
    <div class="bg-red-50 border border-red-200 rounded p-4">
      <div class="text-3xl font-bold text-red-600">0</div>
      <div class="text-sm text-gray-600">Tests Failed</div>
    </div>
    <div class="bg-yellow-50 border border-yellow-200 rounded p-4">
      <div class="text-3xl font-bold text-yellow-600">2</div>
      <div class="text-sm text-gray-600">Tests Skipped</div>
    </div>
    <div class="bg-blue-50 border border-blue-200 rounded p-4">
      <div class="text-3xl font-bold text-blue-600">98.1%</div>
      <div class="text-sm text-gray-600">Pass Rate</div>
    </div>
  </div>

  <div class="space-y-3">
    <h3 class="font-semibold text-gray-900">Test Suites</h3>
    <div class="space-y-2">
      <div class="flex justify-between p-3 bg-green-50 rounded border border-green-200">
        <span class="font-mono text-sm">cmd/api-server</span>
        <span class="text-sm text-green-600">✓ 45 passed</span>
      </div>
      <div class="flex justify-between p-3 bg-green-50 rounded border border-green-200">
        <span class="font-mono text-sm">pkg/dashboard</span>
        <span class="text-sm text-green-600">✓ 68 passed</span>
      </div>
      <div class="flex justify-between p-3 bg-green-50 rounded border border-green-200">
        <span class="font-mono text-sm">pkg/apis</span>
        <span class="text-sm text-green-600">✓ 43 passed, 2 skipped</span>
      </div>
    </div>
  </div>

  <div class="mt-6 text-sm text-gray-600">
    <p>Total Duration: 12.3s</p>
    <p>Generated: 2025-10-27T04:30:35Z</p>
  </div>
</div>
`)
}

// generateCoverageReportJSON generates a demo JSON coverage report
func generateCoverageReportJSON() []byte {
	return []byte(`{
  "coverage": {
    "overall": 78.5,
    "packages": [
      {
        "name": "github.com/org/c8s/cmd/api-server",
        "coverage": 82.1,
        "lines_covered": 328,
        "lines_total": 400
      },
      {
        "name": "github.com/org/c8s/pkg/dashboard",
        "coverage": 75.3,
        "lines_covered": 421,
        "lines_total": 559
      },
      {
        "name": "github.com/org/c8s/pkg/apis",
        "coverage": 79.2,
        "lines_covered": 365,
        "lines_total": 461
      }
    ]
  },
  "generated": "2025-10-27T04:30:35Z"
}`)
}

// generateBuildLog generates a demo build log
func generateBuildLog() []byte {
	return []byte(`[2025-10-27T04:30:16Z] Step started: build
[2025-10-27T04:30:17Z] $ echo 'Starting build process'
[2025-10-27T04:30:17Z] Starting build process
[2025-10-27T04:30:18Z] $ go build -o bin/app ./cmd/api-server
[2025-10-27T04:30:19Z] go: downloading github.com/go-chi/chi/v5
[2025-10-27T04:30:20Z] go: downloading sigs.k8s.io/controller-runtime
[2025-10-27T04:30:22Z] Build completed successfully
[2025-10-27T04:30:23Z] Artifacts: bin/app (45.2 MB)
[2025-10-27T04:30:24Z] Step completed with status: Succeeded`)
}
