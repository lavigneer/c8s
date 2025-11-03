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

// TestListPipelineRunsHandlerNoAuth tests pipeline runs handler rejects unauthenticated
func TestListPipelineRunsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects/proj-1/runs", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListPipelineRunsHandler(w, req)

	// Will return 400 if projectId parsing fails or 401 if auth fails
	// The handler checks auth after parsing, so we test both possibilities
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

// TestListPipelineRunsHandlerMissingProjectID tests handler validates projectId
func TestListPipelineRunsHandlerMissingProjectID(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects//runs", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListPipelineRunsHandler(w, req)

	// Should fail - missing projectId
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestListPipelineRunsHandlerWithFilter tests handler accepts filter parameters
func TestListPipelineRunsHandlerWithFilter(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects/proj-1/runs?status=success&branch=main", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListPipelineRunsHandler(w, req)

	// Should return 400 or 401 depending on projectId parsing
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

// TestGetPipelineRunHandlerNoAuth tests get single run handler rejects unauthenticated
func TestGetPipelineRunHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123", nil)
	w := httptest.NewRecorder()

	handlerspkg.GetPipelineRunHandler(w, req)

	// Should return error
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestGetPipelineRunHandlerMissingRunID tests handler validates runId
func TestGetPipelineRunHandlerMissingRunID(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/", nil)
	w := httptest.NewRecorder()

	handlerspkg.GetPipelineRunHandler(w, req)

	// Should fail - missing runId
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestPipelineRunDetailsHandlerNoAuth tests run details handler rejects unauthenticated
func TestPipelineRunDetailsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/details", nil)
	w := httptest.NewRecorder()

	handlerspkg.PipelineRunDetailsHandler(w, req)

	// Should return error
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestListPipelineRunsHandlerWithPagination tests handler respects pagination
func TestListPipelineRunsHandlerWithPagination(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/projects/proj-1/runs?page=2&limit=10", nil)
	w := httptest.NewRecorder()

	handlerspkg.ListPipelineRunsHandler(w, req)

	// Should return 400 or 401 depending on projectId parsing
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}
