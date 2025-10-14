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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/org/c8s/pkg/apis/v1alpha1"
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
	namespace := extractNamespace(r)
	if namespace == "" {
		http.Error(w, "namespace is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listPipelineConfigs(w, r, namespace)
	case http.MethodPost:
		h.createPipelineConfig(w, r, namespace)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandlePipelineConfig handles get/update/delete operations on a specific PipelineConfig
func (h *PipelineConfigHandler) HandlePipelineConfig(w http.ResponseWriter, r *http.Request) {
	namespace := extractNamespace(r)
	name := extractResourceName(r)

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name are required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getPipelineConfig(w, r, namespace, name)
	case http.MethodPut, http.MethodPatch:
		h.updatePipelineConfig(w, r, namespace, name)
	case http.MethodDelete:
		h.deletePipelineConfig(w, r, namespace, name)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PipelineConfigHandler) listPipelineConfigs(w http.ResponseWriter, r *http.Request, namespace string) {
	var configs v1alpha1.PipelineConfigList
	if err := h.client.List(r.Context(), &configs, client.InNamespace(namespace)); err != nil {
		http.Error(w, fmt.Sprintf("failed to list pipeline configs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(configs); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *PipelineConfigHandler) getPipelineConfig(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var config v1alpha1.PipelineConfig
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := h.client.Get(r.Context(), key, &config); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline config not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get pipeline config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *PipelineConfigHandler) createPipelineConfig(w http.ResponseWriter, r *http.Request, namespace string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var config v1alpha1.PipelineConfig
	if err := json.Unmarshal(body, &config); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	config.Namespace = namespace

	if err := h.client.Create(r.Context(), &config); err != nil {
		http.Error(w, fmt.Sprintf("failed to create pipeline config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(config); err != nil {
		// Status already written, log error
		_, _ = fmt.Fprintf(w, `{"error": "failed to encode response: %v"}`, err)
	}
}

func (h *PipelineConfigHandler) updatePipelineConfig(w http.ResponseWriter, r *http.Request, namespace, name string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var config v1alpha1.PipelineConfig
	if err := json.Unmarshal(body, &config); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	config.Namespace = namespace
	config.Name = name

	var existingConfig v1alpha1.PipelineConfig
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := h.client.Get(r.Context(), key, &existingConfig); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline config not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get pipeline config: %v", err), http.StatusInternalServerError)
		return
	}

	existingConfig.Spec = config.Spec
	if err := h.client.Update(r.Context(), &existingConfig); err != nil {
		http.Error(w, fmt.Sprintf("failed to update pipeline config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(existingConfig); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *PipelineConfigHandler) deletePipelineConfig(w http.ResponseWriter, r *http.Request, namespace, name string) {
	config := &v1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}

	if err := h.client.Delete(r.Context(), config); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline config not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to delete pipeline config: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
