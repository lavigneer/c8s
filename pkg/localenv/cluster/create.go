package cluster

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/org/c8s/pkg/localenv"
	"gopkg.in/yaml.v3"
)

// CreateOptions holds options for cluster creation
type CreateOptions struct {
	ConfigPath string
	Config     *localenv.ClusterConfig
	Wait       bool
	Timeout    time.Duration
}

// Create creates a new local Kubernetes cluster
func Create(ctx context.Context, opts CreateOptions) (*localenv.ClusterStatus, error) {
	// Step 1: Load or use provided configuration
	config, err := loadOrDefaultConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Step 2: Validate cluster configuration
	if err := localenv.ValidateClusterConfig(config); err != nil {
		return nil, fmt.Errorf("invalid cluster configuration: %w", err)
	}

	// Step 3: Check Docker availability
	k3dClient := NewK3dClient()
	if err := k3dClient.IsDockerAvailable(ctx); err != nil {
		return nil, &DockerNotAvailableError{Err: err}
	}

	// Step 4: Check if cluster already exists
	existing, err := k3dClient.Get(ctx, config.Name)
	if err == nil && existing != nil {
		return nil, &ClusterAlreadyExistsError{Name: config.Name}
	}

	// Step 5: Convert to k3d create config
	k3dConfig := convertToK3dConfig(config, opts.Timeout)

	// Step 6: Create the cluster
	if err := k3dClient.Create(ctx, k3dConfig); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Step 7: Wait for cluster to be ready (if requested)
	if opts.Wait {
		if err := waitForClusterReady(ctx, config.Name, opts.Timeout); err != nil {
			// Attempt to clean up the partially created cluster
			_ = k3dClient.Delete(ctx, config.Name)
			return nil, fmt.Errorf("cluster creation timeout: %w", err)
		}
	}

	// Step 8: Get cluster status
	status, err := GetStatus(ctx, config.Name)
	if err != nil {
		return nil, fmt.Errorf("cluster created but failed to get status: %w", err)
	}

	return status, nil
}

// loadOrDefaultConfig loads configuration from file or uses provided/default config
func loadOrDefaultConfig(opts CreateOptions) (*localenv.ClusterConfig, error) {
	// If config is provided directly, use it
	if opts.Config != nil {
		return opts.Config, nil
	}

	// If config path is provided, load from file
	if opts.ConfigPath != "" {
		config, err := loadConfigFromFile(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Otherwise, use default config
	defaultConfig := localenv.DefaultClusterConfig()
	return &defaultConfig, nil
}

// loadConfigFromFile loads cluster configuration from a YAML file
func loadConfigFromFile(path string) (*localenv.ClusterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config localenv.ClusterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// convertToK3dConfig converts ClusterConfig to k3d ClusterCreateConfig
func convertToK3dConfig(config *localenv.ClusterConfig, timeout time.Duration) *ClusterCreateConfig {
	k3dConfig := &ClusterCreateConfig{
		Name:              config.Name,
		KubernetesVersion: config.KubernetesVersion,
		K3sArgs:           config.Options.K3sArgs,
		WaitTimeout:       timeout,
	}

	// Count servers and agents
	for _, node := range config.Nodes {
		switch node.Type {
		case "server":
			k3dConfig.Servers = node.Count
		case "agent":
			k3dConfig.Agents = node.Count
		}
	}

	// Configure registry
	if config.Registry != nil && config.Registry.Enabled {
		k3dConfig.RegistryEnabled = true
		k3dConfig.RegistryName = config.Registry.Name
		k3dConfig.RegistryPort = config.Registry.HostPort
	}

	// Convert port mappings
	for _, port := range config.Ports {
		k3dConfig.Ports = append(k3dConfig.Ports, PortMapping{
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
			NodeFilter:    port.NodeFilter,
		})
	}

	return k3dConfig
}

// waitForClusterReady waits for the cluster to become ready
func waitForClusterReady(ctx context.Context, clusterName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for cluster to be ready")
			}

			// Check cluster status
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

// Custom error types

// ClusterAlreadyExistsError is returned when attempting to create a cluster that already exists
type ClusterAlreadyExistsError struct {
	Name string
}

func (e *ClusterAlreadyExistsError) Error() string {
	return fmt.Sprintf("cluster '%s' already exists", e.Name)
}

// DockerNotAvailableError is returned when Docker is not available
type DockerNotAvailableError struct {
	Err error
}

func (e *DockerNotAvailableError) Error() string {
	return fmt.Sprintf("Docker is not available: %v", e.Err)
}

// ClusterNotReadyError is returned when cluster is not ready within timeout
type ClusterNotReadyError struct {
	Name string
}

func (e *ClusterNotReadyError) Error() string {
	return fmt.Sprintf("cluster '%s' is not ready", e.Name)
}
