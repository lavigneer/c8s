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

// NewMockK8sClient creates a new mock K8s client for testing
func NewMockK8sClient() *dashboard.K8sClient {
	return &dashboard.K8sClient{Client: nil}
}
