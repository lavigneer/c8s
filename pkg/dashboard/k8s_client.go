package dashboard

import (
	"context"
	"fmt"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClient wraps the controller-runtime client for C8S-specific operations
type K8sClient struct {
	client.Client
}

// NewK8sClient creates a new K8s client wrapper
func NewK8sClient(c client.Client) *K8sClient {
	return &K8sClient{Client: c}
}

// ListPipelineRuns retrieves pipeline runs for a namespace
func (k *K8sClient) ListPipelineRuns(ctx context.Context, namespace string, opts ...client.ListOption) (*v1alpha1.PipelineRunList, error) {
	if k.Client == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	var list v1alpha1.PipelineRunList
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}
	listOpts = append(listOpts, opts...)

	if err := k.Client.List(ctx, &list, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	return &list, nil
}

// GetPipelineRun retrieves a single pipeline run by name
func (k *K8sClient) GetPipelineRun(ctx context.Context, namespace, name string) (*v1alpha1.PipelineRun, error) {
	if k.Client == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	var pr v1alpha1.PipelineRun
	key := client.ObjectKey{Namespace: namespace, Name: name}

	if err := k.Client.Get(ctx, key, &pr); err != nil {
		return nil, fmt.Errorf("failed to get pipeline run: %w", err)
	}

	return &pr, nil
}

// GetPipelineConfig retrieves a pipeline config by name
func (k *K8sClient) GetPipelineConfig(ctx context.Context, namespace, name string) (*v1alpha1.PipelineConfig, error) {
	if k.Client == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	var pc v1alpha1.PipelineConfig
	key := client.ObjectKey{Namespace: namespace, Name: name}

	if err := k.Client.Get(ctx, key, &pc); err != nil {
		return nil, fmt.Errorf("failed to get pipeline config: %w", err)
	}

	return &pc, nil
}

// ListPipelineConfigs retrieves pipeline configs for a namespace
func (k *K8sClient) ListPipelineConfigs(ctx context.Context, namespace string, opts ...client.ListOption) (*v1alpha1.PipelineConfigList, error) {
	if k.Client == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	var list v1alpha1.PipelineConfigList
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}
	listOpts = append(listOpts, opts...)

	if err := k.Client.List(ctx, &list, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list pipeline configs: %w", err)
	}

	return &list, nil
}

// CreatePipelineConfig creates a new pipeline config
func (k *K8sClient) CreatePipelineConfig(ctx context.Context, config *v1alpha1.PipelineConfig) error {
	if k.Client == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	if err := k.Client.Create(ctx, config); err != nil {
		return fmt.Errorf("failed to create pipeline config: %w", err)
	}

	return nil
}

// DeletePipelineConfig deletes a pipeline config
func (k *K8sClient) DeletePipelineConfig(ctx context.Context, namespace, name string) error {
	if k.Client == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	config := &v1alpha1.PipelineConfig{}
	config.SetName(name)
	config.SetNamespace(namespace)

	if err := k.Client.Delete(ctx, config); err != nil {
		return fmt.Errorf("failed to delete pipeline config: %w", err)
	}

	return nil
}

