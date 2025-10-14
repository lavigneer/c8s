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
	namespace := extractNamespace(r)
	if namespace == "" {
		http.Error(w, "namespace is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listPipelineRuns(w, r, namespace)
	case http.MethodPost:
		h.createPipelineRun(w, r, namespace)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandlePipelineRun handles get/update/delete operations on a specific PipelineRun
func (h *PipelineRunHandler) HandlePipelineRun(w http.ResponseWriter, r *http.Request) {
	namespace := extractNamespace(r)
	name := extractResourceName(r)

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name are required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getPipelineRun(w, r, namespace, name)
	case http.MethodDelete:
		h.deletePipelineRun(w, r, namespace, name)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PipelineRunHandler) listPipelineRuns(w http.ResponseWriter, r *http.Request, namespace string) {
	var runs v1alpha1.PipelineRunList
	listOpts := []client.ListOption{client.InNamespace(namespace)}

	// Support filtering by phase
	if phase := r.URL.Query().Get("phase"); phase != "" {
		listOpts = append(listOpts, client.MatchingFields{"status.phase": phase})
	}

	// Support filtering by config
	if config := r.URL.Query().Get("config"); config != "" {
		listOpts = append(listOpts, client.MatchingLabels{"c8s.dev/pipeline-config": config})
	}

	if err := h.client.List(r.Context(), &runs, listOpts...); err != nil {
		http.Error(w, fmt.Sprintf("failed to list pipeline runs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(runs); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *PipelineRunHandler) getPipelineRun(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var run v1alpha1.PipelineRun
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := h.client.Get(r.Context(), key, &run); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline run not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get pipeline run: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(run); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *PipelineRunHandler) createPipelineRun(w http.ResponseWriter, r *http.Request, namespace string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var run v1alpha1.PipelineRun
	if err := json.Unmarshal(body, &run); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	run.Namespace = namespace

	if err := h.client.Create(r.Context(), &run); err != nil {
		http.Error(w, fmt.Sprintf("failed to create pipeline run: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(run); err != nil {
		_, _ = fmt.Fprintf(w, `{"error": "failed to encode response: %v"}`, err)
	}
}

func (h *PipelineRunHandler) deletePipelineRun(w http.ResponseWriter, r *http.Request, namespace, name string) {
	run := &v1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}

	if err := h.client.Delete(r.Context(), run); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline run not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to delete pipeline run: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
