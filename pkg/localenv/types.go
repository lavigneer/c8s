package localenv

import "time"

// ClusterConfig represents the configuration for a local Kubernetes cluster
type ClusterConfig struct {
	Name              string          `json:"name" yaml:"name" validate:"required,cluster_name"`
	KubernetesVersion string          `json:"kubernetesVersion" yaml:"kubernetesVersion" validate:"required,k8s_version"`
	Nodes             []NodeConfig    `json:"nodes" yaml:"nodes" validate:"required,min=1,dive"`
	Ports             []PortMapping   `json:"ports,omitempty" yaml:"ports,omitempty" validate:"dive"`
	Registry          *RegistryConfig `json:"registry,omitempty" yaml:"registry,omitempty"`
	VolumeMounts      []VolumeMount   `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty" validate:"dive"`
	Options           ClusterOptions  `json:"options,omitempty" yaml:"options,omitempty"`
}

// NodeConfig represents node configuration within a cluster
type NodeConfig struct {
	Type      string          `json:"type" yaml:"type" validate:"required,oneof=server agent"`
	Count     int             `json:"count" yaml:"count" validate:"required,min=0"`
	Resources *ResourceLimits `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// ResourceLimits defines resource constraints for nodes
type ResourceLimits struct {
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
}

// PortMapping represents a port mapping from host to cluster
type PortMapping struct {
	HostPort      int    `json:"hostPort" yaml:"hostPort" validate:"required,min=1024,max=65535"`
	ContainerPort int    `json:"containerPort" yaml:"containerPort" validate:"required,min=1,max=65535"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty" validate:"omitempty,oneof=TCP UDP"`
	NodeFilter    string `json:"nodeFilter" yaml:"nodeFilter" validate:"required,node_filter"`
}

// RegistryConfig represents local container registry configuration
type RegistryConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Name        string `json:"name" yaml:"name" validate:"required_if=Enabled true,omitempty,hostname"`
	HostPort    int    `json:"hostPort" yaml:"hostPort" validate:"required_if=Enabled true,omitempty,min=1024,max=65535"`
	ProxyRemote string `json:"proxyRemote,omitempty" yaml:"proxyRemote,omitempty" validate:"omitempty,url"`
}

// VolumeMount represents a volume mount from host to cluster nodes
type VolumeMount struct {
	HostPath      string `json:"hostPath" yaml:"hostPath" validate:"required,absolute_path"`
	ContainerPath string `json:"containerPath" yaml:"containerPath" validate:"required,absolute_path"`
	NodeFilter    string `json:"nodeFilter" yaml:"nodeFilter" validate:"required,node_filter"`
}

// ClusterOptions contains advanced cluster configuration options
type ClusterOptions struct {
	WaitTimeout            string   `json:"waitTimeout,omitempty" yaml:"waitTimeout,omitempty" validate:"omitempty,duration"`
	UpdateDefaultKubeconfig bool     `json:"updateDefaultKubeconfig,omitempty" yaml:"updateDefaultKubeconfig,omitempty"`
	SwitchContext          bool     `json:"switchContext,omitempty" yaml:"switchContext,omitempty"`
	DisableLoadBalancer    bool     `json:"disableLoadBalancer,omitempty" yaml:"disableLoadBalancer,omitempty"`
	K3sArgs                []string `json:"k3sArgs,omitempty" yaml:"k3sArgs,omitempty"`
}

// EnvironmentConfig represents a complete local test environment (cluster + operator deployment)
type EnvironmentConfig struct {
	Cluster  ClusterConfig      `json:"cluster" yaml:"cluster" validate:"required"`
	Operator OperatorDeployment `json:"operator" yaml:"operator" validate:"required"`
	Samples  []SampleConfig     `json:"samples,omitempty" yaml:"samples,omitempty" validate:"dive"`
}

// OperatorDeployment represents C8S operator deployment configuration
type OperatorDeployment struct {
	Image           string `json:"image" yaml:"image" validate:"required"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty" validate:"omitempty,oneof=Always IfNotPresent Never"`
	CRDsPath        string `json:"crdsPath" yaml:"crdsPath" validate:"required"`
	ManifestsPath   string `json:"manifestsPath" yaml:"manifestsPath" validate:"required"`
	Namespace       string `json:"namespace,omitempty" yaml:"namespace,omitempty" validate:"omitempty,k8s_namespace"`
	Replicas        int    `json:"replicas,omitempty" yaml:"replicas,omitempty" validate:"omitempty,min=1"`
}

// SampleConfig represents a sample PipelineConfig for testing
type SampleConfig struct {
	Name        string `json:"name" yaml:"name" validate:"required"`
	Path        string `json:"path" yaml:"path" validate:"required"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// DefaultClusterConfig returns a cluster configuration with sensible defaults
func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Name:              "c8s-dev",
		KubernetesVersion: "v1.28.15",
		Nodes: []NodeConfig{
			{Type: "server", Count: 1},
			{Type: "agent", Count: 2},
		},
		Ports: []PortMapping{
			{
				HostPort:      8080,
				ContainerPort: 80,
				Protocol:      "TCP",
				NodeFilter:    "loadbalancer",
			},
		},
		Registry: &RegistryConfig{
			Enabled:  true,
			Name:     "registry.localhost",
			HostPort: 5000,
		},
		Options: ClusterOptions{
			WaitTimeout:            "60s",
			UpdateDefaultKubeconfig: true,
			SwitchContext:          true,
			DisableLoadBalancer:    false,
			K3sArgs:                []string{"--disable=traefik"},
		},
	}
}

// DefaultOperatorDeployment returns operator deployment configuration with defaults
func DefaultOperatorDeployment() OperatorDeployment {
	return OperatorDeployment{
		Image:           "c8s-operator:dev",
		ImagePullPolicy: "IfNotPresent",
		CRDsPath:        "config/crd/bases",
		ManifestsPath:   "config/manager",
		Namespace:       "c8s-system",
		Replicas:        1,
	}
}

// ParseDuration converts a duration string to time.Duration
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}
