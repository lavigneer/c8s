package cluster

import (
	"context"
	"fmt"
)

// DeleteOptions holds options for cluster deletion
type DeleteOptions struct {
	Name  string
	Force bool // Skip confirmation
}

// Delete deletes a local Kubernetes cluster
func Delete(ctx context.Context, opts DeleteOptions) error {
	k3dClient := NewK3dClient()

	// Check if cluster exists
	_, err := k3dClient.Get(ctx, opts.Name)
	if err != nil {
		return &ClusterNotFoundError{Name: opts.Name}
	}

	// Delete the cluster using k3d
	if err := k3dClient.Delete(ctx, opts.Name); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	// Note: k3d automatically removes the kubeconfig context when deleting the cluster
	// Docker containers and volumes are also cleaned up automatically by k3d

	return nil
}

// DeleteAll deletes all c8s clusters
func DeleteAll(ctx context.Context) ([]string, error) {
	k3dClient := NewK3dClient()

	// List all clusters
	clusters, err := k3dClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Filter for c8s clusters (names starting with "c8s-")
	var deleted []string
	var errors []error

	for _, cluster := range clusters {
		// Only delete c8s-prefixed clusters
		if len(cluster.Name) >= 4 && cluster.Name[:4] == "c8s-" {
			if err := k3dClient.Delete(ctx, cluster.Name); err != nil {
				errors = append(errors, fmt.Errorf("failed to delete cluster '%s': %w", cluster.Name, err))
			} else {
				deleted = append(deleted, cluster.Name)
			}
		}
	}

	if len(errors) > 0 {
		return deleted, fmt.Errorf("encountered %d errors during deletion: %v", len(errors), errors)
	}

	return deleted, nil
}

// VerifyCleanup verifies that cluster resources have been cleaned up
func VerifyCleanup(ctx context.Context, clusterName string) error {
	k3dClient := NewK3dClient()

	// Check if cluster still exists
	_, err := k3dClient.Get(ctx, clusterName)
	if err == nil {
		return fmt.Errorf("cluster '%s' still exists after deletion", clusterName)
	}

	// If cluster is not found, cleanup is verified
	return nil
}
