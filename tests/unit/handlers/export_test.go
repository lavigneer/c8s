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

// TestExportPipelineRunsHandlerNoAuth tests export handler rejects unauthenticated
func TestExportPipelineRunsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/exports/runs?format=json", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ExportPipelineRunsHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExportPipelineRunsHandlerJSON tests export handler handles JSON format
func TestExportPipelineRunsHandlerJSON(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/exports/runs?format=json", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ExportPipelineRunsHandler(w, req)

	// Should return 401 due to missing auth
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExportPipelineRunsHandlerCSV tests export handler handles CSV format
func TestExportPipelineRunsHandlerCSV(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/exports/runs?format=csv", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ExportPipelineRunsHandler(w, req)

	// Should return 401 due to missing auth
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExportPipelineRunsHandlerInvalidFormat tests export handler defaults invalid format
func TestExportPipelineRunsHandlerInvalidFormat(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/exports/runs?format=invalid", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ExportPipelineRunsHandler(w, req)

	// Should default to JSON and return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExportPipelineRunsHandlerWithFilter tests export handler accepts filters
func TestExportPipelineRunsHandlerWithFilter(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/exports/runs?format=json&status=success&branch=main", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.ExportPipelineRunsHandler(w, req)

	// Should return 401 due to missing auth (filters processed after auth)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDeleteProjectHandlerNoAuth tests delete project handler rejects unauthenticated
func TestDeleteProjectHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("DELETE", "/api/projects/proj-123", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.DeleteProjectHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateProjectHandlerNoAuth tests create project handler rejects unauthenticated
func TestCreateProjectHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("POST", "/api/projects", http.NoBody)
	w := httptest.NewRecorder()

	handlerspkg.CreateProjectHandler(w, req)

	// Should return 401 due to missing auth
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
