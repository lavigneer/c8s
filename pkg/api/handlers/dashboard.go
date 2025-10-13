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

// DashboardHandler handles HTMX dashboard requests
type DashboardHandler struct {
	client client.Client
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(client client.Client) *DashboardHandler {
	return &DashboardHandler{
		client: client,
	}
}

// ServeDashboard serves the main dashboard page
func (h *DashboardHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T054-T058
	http.Error(w, "Dashboard not implemented", http.StatusNotImplemented)
}

// ServeRuns serves the pipeline runs page
func (h *DashboardHandler) ServeRuns(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T054-T058
	http.Error(w, "Dashboard not implemented", http.StatusNotImplemented)
}

// ServeLogs serves the logs viewer page
func (h *DashboardHandler) ServeLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in T054-T058
	http.Error(w, "Dashboard not implemented", http.StatusNotImplemented)
}
