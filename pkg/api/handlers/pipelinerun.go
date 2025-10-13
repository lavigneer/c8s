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

package handlers

import (
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PipelineRunHandler handles PipelineRun API requests
type PipelineRunHandler struct {
	client client.Client
}

// NewPipelineRunHandler creates a new PipelineRunHandler
func NewPipelineRunHandler(client client.Client) *PipelineRunHandler {
	return &PipelineRunHandler{
		client: client,
	}
}

// HandlePipelineRuns handles list/create operations on PipelineRuns
func (h *PipelineRunHandler) HandlePipelineRuns(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T050
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandlePipelineRun handles get/update/delete operations on a specific PipelineRun
func (h *PipelineRunHandler) HandlePipelineRun(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T050
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
