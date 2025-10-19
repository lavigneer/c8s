package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// K3dClient interface defines operations for k3d cluster management
type K3dClient interface {
	// Create creates a new k3d cluster
	Create(ctx context.Context, config *ClusterCreateConfig) error

	// Delete deletes a k3d cluster
	Delete(ctx context.Context, name string) error

	// Start starts a stopped k3d cluster
	Start(ctx context.Context, name string) error

	// Stop stops a running k3d cluster
	Stop(ctx context.Context, name string) error

	// List lists all k3d clusters
	List(ctx context.Context) ([]ClusterInfo, error)

	// Get gets information about a specific cluster
	Get(ctx context.Context, name string) (*ClusterInfo, error)

	// LoadImage loads a Docker image into the k3d cluster
	LoadImage(ctx context.Context, clusterName, imageName string) error

	// IsDockerAvailable checks if Docker daemon is accessible
	IsDockerAvailable(ctx context.Context) error
}

// ClusterCreateConfig holds configuration for cluster creation
type ClusterCreateConfig struct {
	Name              string
	KubernetesVersion string
	Servers           int
	Agents            int
	RegistryEnabled   bool
	RegistryName      string
	RegistryPort      int
	Ports             []PortMapping
	K3sArgs           []string
	WaitTimeout       time.Duration
}

// PortMapping represents a port mapping for k3d
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
	NodeFilter    string
}

// ClusterInfo holds information about a k3d cluster
type ClusterInfo struct {
	Name            string `json:"name"`
	Network         string `json:"network"`
	Token           string `json:"token"`
	Servers         int    `json:"serversCount"`
	ServersRunning  int    `json:"serversRunning"`
	Agents          int    `json:"agentsCount"`
	AgentsRunning   int    `json:"agentsRunning"`
	HasLoadBalancer bool   `json:"hasLoadBalancer"`
	ImageVolume     string `json:"imageVolume"`
}

// k3dClientImpl implements K3dClient using k3d command-line tool
type k3dClientImpl struct {
	execTimeout time.Duration
}

// NewK3dClient creates a new k3d client
func NewK3dClient() K3dClient {
	return &k3dClientImpl{
		execTimeout: 5 * time.Minute,
	}
}

// Create creates a new k3d cluster
func (k *k3dClientImpl) Create(ctx context.Context, config *ClusterCreateConfig) error {
	args := []string{"cluster", "create", config.Name}

	// Add Kubernetes version if specified
	if config.KubernetesVersion != "" {
		args = append(args, "--image", fmt.Sprintf("rancher/k3s:%s-k3s1", config.KubernetesVersion))
	}

	// Add server and agent counts
	if config.Servers > 0 {
		args = append(args, "--servers", fmt.Sprintf("%d", config.Servers))
	}
	if config.Agents > 0 {
		args = append(args, "--agents", fmt.Sprintf("%d", config.Agents))
	}

	// Add registry if enabled
	if config.RegistryEnabled {
		registryArg := fmt.Sprintf("%s:0.0.0.0:%d", config.RegistryName, config.RegistryPort)
		args = append(args, "--registry-create", registryArg)
	}

	// Add port mappings
	for _, port := range config.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "TCP"
		}
		portArg := fmt.Sprintf("%d:%d/%s@%s", port.HostPort, port.ContainerPort, strings.ToLower(protocol), port.NodeFilter)
		args = append(args, "-p", portArg)
	}

	// Add k3s args
	for _, k3sArg := range config.K3sArgs {
		args = append(args, "--k3s-arg", fmt.Sprintf("%s@server:*", k3sArg))
	}

	// Add wait flag
	args = append(args, "--wait")

	// Set timeout
	execCtx, cancel := context.WithTimeout(ctx, config.WaitTimeout)
	defer cancel()

	return k.runK3dCommand(execCtx, args...)
}

// Delete deletes a k3d cluster
func (k *k3dClientImpl) Delete(ctx context.Context, name string) error {
	return k.runK3dCommand(ctx, "cluster", "delete", name)
}

// Start starts a stopped k3d cluster
func (k *k3dClientImpl) Start(ctx context.Context, name string) error {
	return k.runK3dCommand(ctx, "cluster", "start", name)
}

// Stop stops a running k3d cluster
func (k *k3dClientImpl) Stop(ctx context.Context, name string) error {
	return k.runK3dCommand(ctx, "cluster", "stop", name)
}

// List lists all k3d clusters
func (k *k3dClientImpl) List(ctx context.Context) ([]ClusterInfo, error) {
	output, err := k.runK3dCommandWithOutput(ctx, "cluster", "list", "-o", "json")
	if err != nil {
		return nil, err
	}

	// Parse JSON output
	var clusters []ClusterInfo
	if err := json.Unmarshal(output, &clusters); err != nil {
		// If JSON parsing fails, k3d might have returned empty list
		if strings.TrimSpace(string(output)) == "" || string(output) == "null" {
			return []ClusterInfo{}, nil
		}
		return nil, fmt.Errorf("failed to parse cluster list: %w", err)
	}

	return clusters, nil
}

// Get gets information about a specific cluster
func (k *k3dClientImpl) Get(ctx context.Context, name string) (*ClusterInfo, error) {
	output, err := k.runK3dCommandWithOutput(ctx, "cluster", "list", name, "-o", "json")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("cluster %s not found", name)
		}
		return nil, err
	}

	// Parse JSON output
	var clusters []ClusterInfo
	if err := json.Unmarshal(output, &clusters); err != nil {
		return nil, fmt.Errorf("failed to parse cluster info: %w", err)
	}

	if len(clusters) == 0 {
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	return &clusters[0], nil
}

// LoadImage loads a Docker image into the k3d cluster
func (k *k3dClientImpl) LoadImage(ctx context.Context, clusterName, imageName string) error {
	return k.runK3dCommand(ctx, "image", "import", imageName, "-c", clusterName)
}

// IsDockerAvailable checks if Docker daemon is accessible
func (k *k3dClientImpl) IsDockerAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("docker not available: %s", stderr.String())
		}
		return fmt.Errorf("docker not available: %w", err)
	}

	return nil
}

// runK3dCommand executes a k3d command and returns error if it fails
func (k *k3dClientImpl) runK3dCommand(ctx context.Context, args ...string) error {
	_, err := k.runK3dCommandWithOutput(ctx, args...)
	return err
}

// runK3dCommandWithOutput executes a k3d command and returns output
func (k *k3dClientImpl) runK3dCommandWithOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "k3d", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("k3d command failed: %s", stderr.String())
		}
		return nil, fmt.Errorf("k3d command failed: %w", err)
	}

	return stdout.Bytes(), nil
}
