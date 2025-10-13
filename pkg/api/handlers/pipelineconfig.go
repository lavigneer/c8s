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

// PipelineConfigHandler handles PipelineConfig API requests
type PipelineConfigHandler struct {
	client client.Client
}

// NewPipelineConfigHandler creates a new PipelineConfigHandler
func NewPipelineConfigHandler(client client.Client) *PipelineConfigHandler {
	return &PipelineConfigHandler{
		client: client,
	}
}

// HandlePipelineConfigs handles list/create operations on PipelineConfigs
func (h *PipelineConfigHandler) HandlePipelineConfigs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T049
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandlePipelineConfig handles get/update/delete operations on a specific PipelineConfig
func (h *PipelineConfigHandler) HandlePipelineConfig(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T049
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
