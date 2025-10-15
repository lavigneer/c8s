package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/org/c8s/pkg/localenv"
)

// GetStatus retrieves the status of a cluster
func GetStatus(ctx context.Context, clusterName string) (*localenv.ClusterStatus, error) {
	k3dClient := NewK3dClient()
	kubectlClient := NewKubectlClient()

	// Get cluster info from k3d
	clusterInfo, err := k3dClient.Get(ctx, clusterName)
	if err != nil {
		return nil, &ClusterNotFoundError{Name: clusterName}
	}

	// Determine cluster state
	state := determineClusterState(clusterInfo)

	status := &localenv.ClusterStatus{
		Name:  clusterName,
		State: state,
	}

	// If cluster is running, get detailed information
	if state == localenv.StateRunning {
		// Get node status from kubectl
		nodes, err := kubectlClient.GetNodes(ctx, clusterName)
		if err != nil {
			// Cluster might not be fully ready yet
			status.Nodes = []localenv.NodeStatus{}
		} else {
			status.Nodes = convertToNodeStatus(nodes)
		}

		// Get API endpoint
		status.APIEndpoint = fmt.Sprintf("https://0.0.0.0:6443")

		// Set registry endpoint if registry exists
		if clusterInfo.ImageVolume != "" {
			status.RegistryEndpoint = "localhost:5000"
		}

		// Get kubeconfig context
		status.Kubeconfig = fmt.Sprintf("k3d-%s", clusterName)
	}

	return status, nil
}

// GetStatusWithUptime retrieves status with uptime calculation
func GetStatusWithUptime(ctx context.Context, clusterName string) (*localenv.ClusterStatus, error) {
	status, err := GetStatus(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	// Calculate uptime if cluster is running
	if status.IsRunning() && status.CreatedAt != nil {
		status.Uptime = status.CalculateUptime()
	}

	return status, nil
}

// determineClusterState determines the cluster state from k3d info
func determineClusterState(info *ClusterInfo) string {
	// Check if all servers are running
	if info.ServersRunning == info.Servers && info.Servers > 0 {
		// Also check if agents are running (if any exist)
		if info.Agents == 0 || info.AgentsRunning == info.Agents {
			return localenv.StateRunning
		}
	}

	// If some nodes are running but not all, consider it starting
	if info.ServersRunning > 0 || info.AgentsRunning > 0 {
		return localenv.StateStarting
	}

	// Otherwise, it's stopped
	return localenv.StateStopped
}

// convertToNodeStatus converts kubectl node info to NodeStatus
func convertToNodeStatus(nodes []KubeNode) []localenv.NodeStatus {
	var status []localenv.NodeStatus
	for _, node := range nodes {
		status = append(status, localenv.NodeStatus{
			Name:    node.Name,
			Role:    node.Role,
			Status:  node.Status,
			Version: node.Version,
		})
	}
	return status
}

// ClusterNotFoundError is returned when a cluster cannot be found
type ClusterNotFoundError struct {
	Name string
}

func (e *ClusterNotFoundError) Error() string {
	return fmt.Sprintf("cluster '%s' not found", e.Name)
}

// WaitForReady waits for a cluster to become ready
func WaitForReady(ctx context.Context, clusterName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for cluster '%s' to be ready", clusterName)
			}

			status, err := GetStatus(ctx, clusterName)
			if err != nil {
				continue // Keep waiting
			}

			if status.IsReady() {
				return nil
			}
		}
	}
}
