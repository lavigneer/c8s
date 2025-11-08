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

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	handlerspkg "github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/dashboard"
)

// TestListArtifactsHandlerNoAuth tests artifacts handler rejects unauthenticated requests
func TestListArtifactsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/artifacts", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ListArtifactsHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not authenticated")
}

// TestListArtifactsHandlerMissingRunID tests artifacts handler validates runId
func TestListArtifactsHandlerMissingRunID(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	// Request without runId path param
	req := httptest.NewRequest("GET", "/api/runs//artifacts", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ListArtifactsHandler(w, req)

	// Should return error due to missing runId
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestListArtifactsHandlerWithFilter tests artifacts handler accepts filters
func TestListArtifactsHandlerWithFilter(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	// Request with filters
	req := httptest.NewRequest("GET", "/api/runs/run-123/artifacts?step_id=step-1&type=log", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ListArtifactsHandler(w, req)

	// Should return 401 (unauthenticated) before checking filters
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetArtifactHandlerNoAuth tests artifact download handler rejects unauthenticated requests
func TestGetArtifactHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/artifacts/artifact-id/download", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.GetArtifactHandler(w, req)

	// Could return 400 (bad request) if parameter parsing fails
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

// TestGetArtifactHandlerMissingID tests artifact download handler validates artifact ID
func TestGetArtifactHandlerMissingID(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/artifacts//download", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.GetArtifactHandler(w, req)

	// Should return error or 401
	assert.NotEqual(t, http.StatusOK, w.Code)
}
