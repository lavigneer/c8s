package localenv

import "time"

// ClusterStatus represents runtime status information for a cluster
type ClusterStatus struct {
	Name             string       `json:"name"`
	State            string       `json:"state"`
	Nodes            []NodeStatus `json:"nodes"`
	Kubeconfig       string       `json:"kubeconfig"`
	APIEndpoint      string       `json:"apiEndpoint"`
	RegistryEndpoint string       `json:"registryEndpoint,omitempty"`
	CreatedAt        *time.Time   `json:"createdAt,omitempty"`
	Uptime           string       `json:"uptime,omitempty"`
}

// NodeStatus represents runtime status for a single node
type NodeStatus struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// ClusterState constants
const (
	StateRunning  = "running"
	StateStopped  = "stopped"
	StateNotFound = "not_found"
	StateCreating = "creating"
	StateStarting = "starting"
	StateDeleting = "deleting"
)

// NodeState constants
const (
	NodeReady    = "Ready"
	NodeNotReady = "NotReady"
	NodeUnknown  = "Unknown"
)

// IsRunning returns true if the cluster is in running state
func (cs *ClusterStatus) IsRunning() bool {
	return cs.State == StateRunning
}

// IsReady returns true if all nodes are ready
func (cs *ClusterStatus) IsReady() bool {
	if !cs.IsRunning() {
		return false
	}
	for _, node := range cs.Nodes {
		if node.Status != NodeReady {
			return false
		}
	}
	return len(cs.Nodes) > 0
}

// CalculateUptime calculates uptime duration if CreatedAt is set
func (cs *ClusterStatus) CalculateUptime() string {
	if cs.CreatedAt == nil {
		return ""
	}
	duration := time.Since(*cs.CreatedAt)

	// Format duration in human-readable format
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return time.Duration(hours*int(time.Hour) + minutes*int(time.Minute)).String()
		}
		return time.Duration(hours * int(time.Hour)).String()
	}
	if minutes > 0 {
		return time.Duration(minutes * int(time.Minute)).String()
	}
	return "< 1m"
}
