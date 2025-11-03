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

// TestGetLogsHandlerNoAuth tests logs handler rejects unauthenticated requests
func TestGetLogsHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/logs", nil)
	w := httptest.NewRecorder()

	handlerspkg.GetLogsHandler(w, req)

	// Should fail due to missing auth
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestGetLogSnapshotHandlerNoAuth tests log snapshot handler rejects unauthenticated
func TestGetLogSnapshotHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/logs/snapshot", nil)
	w := httptest.NewRecorder()

	handlerspkg.GetLogSnapshotHandler(w, req)

	// Should return error due to missing auth
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestPreviewArtifactHandlerNoAuth tests artifact preview handler rejects unauthenticated
func TestPreviewArtifactHandlerNoAuth(t *testing.T) {
	handlerspkg.InitK8sClient(&dashboard.K8sClient{})

	req := httptest.NewRequest("GET", "/api/runs/run-123/artifacts/art-1/preview", nil)
	w := httptest.NewRecorder()

	handlerspkg.PreviewArtifactHandler(w, req)

	// Should return error
	assert.NotEqual(t, http.StatusOK, w.Code)
}
