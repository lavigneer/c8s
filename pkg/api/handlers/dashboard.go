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
	"html/template"
	"net/http"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DashboardHandler handles HTMX dashboard requests
type DashboardHandler struct {
	client    client.Client
	templates *template.Template
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(client client.Client, templateDir string) (*DashboardHandler, error) {
	// Parse all templates
	templates, err := template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		client:    client,
		templates: templates,
	}, nil
}

// DashboardData holds common data for all dashboard pages
type DashboardData struct {
	Title     string
	Active    string
	Namespace string
	// Page-specific data
	RunName string
}

// ServeDashboard serves the main dashboard page (pipelines list)
func (h *DashboardHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	data := DashboardData{
		Title:     "Pipelines",
		Active:    "pipelines",
		Namespace: getNamespace(r),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the pipelines content template
	if err := h.templates.ExecuteTemplate(w, "pipelines.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeRuns serves the pipeline runs page
func (h *DashboardHandler) ServeRuns(w http.ResponseWriter, r *http.Request) {
	data := DashboardData{
		Title:     "Runs",
		Active:    "runs",
		Namespace: getNamespace(r),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the runs content template
	if err := h.templates.ExecuteTemplate(w, "runs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeLogs serves the logs viewer page
func (h *DashboardHandler) ServeLogs(w http.ResponseWriter, r *http.Request) {
	runName := r.URL.Query().Get("run")
	if runName == "" {
		http.Error(w, "Missing 'run' query parameter", http.StatusBadRequest)
		return
	}

	data := DashboardData{
		Title:     "Logs",
		Active:    "runs",
		Namespace: getNamespace(r),
		RunName:   runName,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the logs content template
	if err := h.templates.ExecuteTemplate(w, "logs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getNamespace extracts namespace from query param or defaults to "default"
func getNamespace(r *http.Request) string {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		return "default"
	}
	return ns
}
