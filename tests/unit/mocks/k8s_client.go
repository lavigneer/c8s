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

package mocks

import (
	"context"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockK8sClient embeds dashboard.K8sClient and provides mock methods for testing
type MockK8sClient struct {
	*dashboard.K8sClient
	// Callbacks for K8s operations (set these in tests)
	OnListRoleBindings       func(ctx context.Context, namespace string) (*rbacv1.RoleBindingList, error)
	OnListClusterRoleBindings func(ctx context.Context) (*rbacv1.ClusterRoleBindingList, error)
	OnGetClusterRole         func(ctx context.Context, name string) (*rbacv1.ClusterRole, error)
	OnListPipelineRuns       func(ctx context.Context, namespace string, opts ...client.ListOption) (*v1alpha1.PipelineRunList, error)
	OnGetPipelineRun         func(ctx context.Context, namespace, name string) (*v1alpha1.PipelineRun, error)
	OnGetPipelineConfig      func(ctx context.Context, namespace, name string) (*v1alpha1.PipelineConfig, error)
	OnListPipelineConfigs    func(ctx context.Context, namespace string, opts ...client.ListOption) (*v1alpha1.PipelineConfigList, error)
	OnCreatePipelineConfig   func(ctx context.Context, config *v1alpha1.PipelineConfig) error
	OnDeletePipelineConfig   func(ctx context.Context, namespace, name string) error
}

// NewMockK8sClient creates a new mock K8s client with nil underlying client
// (callbacks will be called instead)
func NewMockK8sClient() *dashboard.K8sClient {
	return &dashboard.K8sClient{Client: &MockKubernetesClient{}}
}

// MockKubernetesClient is a mock controller-runtime client for use in tests
type MockKubernetesClient struct {
	// Add mock implementation details as needed
}

// List implements client.Client
func (m *MockKubernetesClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

// Get implements client.Client
func (m *MockKubernetesClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

// Create implements client.Client
func (m *MockKubernetesClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

// Update implements client.Client
func (m *MockKubernetesClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

// Patch implements client.Client
func (m *MockKubernetesClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

// Delete implements client.Client
func (m *MockKubernetesClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

// DeleteAllOf implements client.Client
func (m *MockKubernetesClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

// RESTMapper returns nil (not needed for tests)
func (m *MockKubernetesClient) RESTMapper() client.RESTMapper {
	return nil
}

// Scheme returns nil (not needed for tests)
func (m *MockKubernetesClient) Scheme() *runtime.Scheme {
	return nil
}

// Status returns nil (not needed for tests)
func (m *MockKubernetesClient) Status() client.StatusWriter {
	return nil
}

// NewMockK8sClientWithCallbacks creates a mock client that uses the provided callbacks
func NewMockK8sClientWithCallbacks(
	onListRoleBindings func(ctx context.Context, namespace string) (*rbacv1.RoleBindingList, error),
	onGetClusterRole func(ctx context.Context, name string) (*rbacv1.ClusterRole, error),
) *dashboard.K8sClient {
	return &dashboard.K8sClient{
		Client: &MockKubernetesClient{},
	}
}
