package cluster

import (
	"context"
	"fmt"
	"time"
)

// StartOptions holds options for starting a cluster
type StartOptions struct {
	Name    string
	Wait    bool
	Timeout time.Duration
}

// StopOptions holds options for stopping a cluster
type StopOptions struct {
	Name string
}

// Start starts a stopped cluster
func Start(ctx context.Context, opts StartOptions) error {
	k3dClient := NewK3dClient()

	// Check if cluster exists
	_, err := k3dClient.Get(ctx, opts.Name)
	if err != nil {
		return &ClusterNotFoundError{Name: opts.Name}
	}

	// Start the cluster
	if err := k3dClient.Start(ctx, opts.Name); err != nil {
		return fmt.Errorf("failed to start cluster: %w", err)
	}

	// Wait for cluster to be ready if requested
	if opts.Wait {
		if err := WaitForReady(ctx, opts.Name, opts.Timeout); err != nil {
			return fmt.Errorf("cluster started but not ready: %w", err)
		}
	}

	return nil
}

// Stop stops a running cluster (preserves state)
func Stop(ctx context.Context, opts StopOptions) error {
	k3dClient := NewK3dClient()

	// Check if cluster exists
	_, err := k3dClient.Get(ctx, opts.Name)
	if err != nil {
		return &ClusterNotFoundError{Name: opts.Name}
	}

	// Stop the cluster
	if err := k3dClient.Stop(ctx, opts.Name); err != nil {
		return fmt.Errorf("failed to stop cluster: %w", err)
	}

	return nil
}

// Restart restarts a cluster (stop then start)
func Restart(ctx context.Context, name string, timeout time.Duration) error {
	// Stop the cluster
	if err := Stop(ctx, StopOptions{Name: name}); err != nil {
		return err
	}

	// Wait a moment for stop to complete
	time.Sleep(2 * time.Second)

	// Start the cluster
	if err := Start(ctx, StartOptions{
		Name:    name,
		Wait:    true,
		Timeout: timeout,
	}); err != nil {
		return err
	}

	return nil
}
