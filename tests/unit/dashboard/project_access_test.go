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

package dashboard

import (
	"testing"

	"github.com/org/c8s/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

// TestRoleLevel verifies role hierarchy levels
func TestRoleLevel(t *testing.T) {
	tests := []struct {
		role     dashboard.Role
		expected int
	}{
		{dashboard.RoleAdmin, 3},
		{dashboard.RoleEditor, 2},
		{dashboard.RoleViewer, 1},
		{dashboard.Role("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.Level())
		})
	}
}

// TestRoleHierarchyComparison verifies role hierarchy works correctly
func TestRoleHierarchyComparison(t *testing.T) {
	tests := []struct {
		name            string
		actualRole      dashboard.Role
		requiredRole    dashboard.Role
		expectedAllowed bool
	}{
		{"Admin >= Admin", dashboard.RoleAdmin, dashboard.RoleAdmin, true},
		{"Admin >= Editor", dashboard.RoleAdmin, dashboard.RoleEditor, true},
		{"Admin >= Viewer", dashboard.RoleAdmin, dashboard.RoleViewer, true},
		{"Editor >= Editor", dashboard.RoleEditor, dashboard.RoleEditor, true},
		{"Editor >= Viewer", dashboard.RoleEditor, dashboard.RoleViewer, true},
		{"Editor < Admin", dashboard.RoleEditor, dashboard.RoleAdmin, false},
		{"Viewer >= Viewer", dashboard.RoleViewer, dashboard.RoleViewer, true},
		{"Viewer < Editor", dashboard.RoleViewer, dashboard.RoleEditor, false},
		{"Viewer < Admin", dashboard.RoleViewer, dashboard.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLevel := tt.actualRole.Level()
			requiredLevel := tt.requiredRole.Level()
			result := actualLevel >= requiredLevel
			assert.Equal(t, tt.expectedAllowed, result)
		})
	}
}

// TestProjectAccessServiceCreation verifies service can be created
func TestProjectAccessServiceCreation(t *testing.T) {
	// Create a K8s client (minimal, no underlying client needed for this test)
	k8sClient := &dashboard.K8sClient{}

	// Create the service
	service := dashboard.NewProjectAccessService(k8sClient)

	// Verify service was created (type assertion check)
	assert.NotNil(t, service)
	_, ok := service.(dashboard.ProjectAccessService)
	assert.True(t, ok, "service should implement ProjectAccessService interface")
}

// TestRoleStringValues verifies role string constants are correct
func TestRoleStringValues(t *testing.T) {
	assert.Equal(t, dashboard.Role("admin"), dashboard.RoleAdmin)
	assert.Equal(t, dashboard.Role("editor"), dashboard.RoleEditor)
	assert.Equal(t, dashboard.Role("viewer"), dashboard.RoleViewer)
}
