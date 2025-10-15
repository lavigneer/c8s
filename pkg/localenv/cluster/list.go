package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/org/c8s/pkg/localenv"
)

// ListOptions holds options for listing clusters
type ListOptions struct {
	All bool // Show all k3d clusters, not just c8s clusters
}

// ClusterListItem represents a cluster in the list
type ClusterListItem struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	NodeCount int    `json:"nodeCount"`
	Version   string `json:"version,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
}

// List lists clusters based on the provided options
func List(ctx context.Context, opts ListOptions) ([]ClusterListItem, error) {
	k3dClient := NewK3dClient()

	// Get all clusters from k3d
	clusters, err := k3dClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	var result []ClusterListItem

	for _, cluster := range clusters {
		// Filter by c8s prefix unless --all is specified
		if !opts.All && !isC8sCluster(cluster.Name) {
			continue
		}

		// Determine state
		state := determineClusterState(&cluster)

		// Calculate total node count
		nodeCount := cluster.Servers + cluster.Agents

		item := ClusterListItem{
			Name:      cluster.Name,
			State:     state,
			NodeCount: nodeCount,
		}

		// Try to get more detailed status if cluster is running
		if state == localenv.StateRunning {
			if status, err := GetStatus(ctx, cluster.Name); err == nil {
				if len(status.Nodes) > 0 && status.Nodes[0].Version != "" {
					item.Version = status.Nodes[0].Version
				}
				item.Uptime = status.CalculateUptime()
			}
		}

		result = append(result, item)
	}

	return result, nil
}

// isC8sCluster checks if a cluster name follows c8s naming convention
func isC8sCluster(name string) bool {
	// c8s clusters typically start with "c8s-"
	return strings.HasPrefix(name, "c8s-")
}
