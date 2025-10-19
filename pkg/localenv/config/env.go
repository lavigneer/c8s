package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnvConfig manages environment variable configuration for C8S
type EnvConfig struct {
	DefaultCluster  string
	DefaultConfig   string
	VerboseMode     bool
	QuietMode       bool
	NoColorOutput   bool
	DevTimeout      int
	ImagePullPolicy string
	Namespace       string
	RegistryEnabled bool
}

// LoadEnvConfig loads configuration from environment variables
// Environment variables follow the pattern: C8S_<SETTING>
func LoadEnvConfig() *EnvConfig {
	cfg := &EnvConfig{
		DefaultCluster:  "c8s-dev",
		DefaultConfig:   "",
		VerboseMode:     false,
		QuietMode:       false,
		NoColorOutput:   false,
		DevTimeout:      300,
		ImagePullPolicy: "IfNotPresent",
		Namespace:       "default",
		RegistryEnabled: true,
	}

	// Load default cluster name
	if val := os.Getenv("C8S_DEV_CLUSTER"); val != "" {
		cfg.DefaultCluster = val
	}

	// Load default config file path
	if val := os.Getenv("C8S_DEV_CONFIG"); val != "" {
		cfg.DefaultConfig = val
	}

	// Load verbose mode
	if val := os.Getenv("C8S_VERBOSE"); val != "" {
		cfg.VerboseMode = parseBool(val)
	}

	// Load quiet mode
	if val := os.Getenv("C8S_QUIET"); val != "" {
		cfg.QuietMode = parseBool(val)
	}

	// Load no-color mode
	if val := os.Getenv("C8S_NO_COLOR"); val != "" {
		cfg.NoColorOutput = parseBool(val)
	}

	// Load dev timeout
	if val := os.Getenv("C8S_DEV_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			cfg.DevTimeout = timeout
		}
	}

	// Load image pull policy
	if val := os.Getenv("C8S_IMAGE_PULL_POLICY"); val != "" {
		cfg.ImagePullPolicy = val
	}

	// Load default namespace
	if val := os.Getenv("C8S_NAMESPACE"); val != "" {
		cfg.Namespace = val
	}

	// Load registry enabled flag
	if val := os.Getenv("C8S_REGISTRY_ENABLED"); val != "" {
		cfg.RegistryEnabled = parseBool(val)
	}

	return cfg
}

// parseBool converts string to boolean
func parseBool(val string) bool {
	val = strings.ToLower(strings.TrimSpace(val))
	return val == "true" || val == "1" || val == "yes"
}

// GetClusterName returns the effective cluster name
// Priority: command-line flag > environment variable > default
func GetClusterName(flagValue string) string {
	if flagValue != "" && flagValue != "c8s-dev" {
		return flagValue
	}
	if val := os.Getenv("C8S_DEV_CLUSTER"); val != "" {
		return val
	}
	return "c8s-dev"
}

// GetNamespace returns the effective namespace
// Priority: command-line flag > environment variable > default
func GetNamespace(flagValue string) string {
	if flagValue != "" && flagValue != "default" {
		return flagValue
	}
	if val := os.Getenv("C8S_NAMESPACE"); val != "" {
		return val
	}
	return "default"
}

// GetConfigPath returns the effective config file path
// Priority: command-line flag > environment variable > default (~/.c8s/cluster.yaml)
func GetConfigPath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if val := os.Getenv("C8S_DEV_CONFIG"); val != "" {
		return val
	}
	home, _ := os.UserHomeDir()
	return home + "/.c8s/cluster.yaml"
}

// PrintEnvVarHelp prints documentation for available environment variables
func PrintEnvVarHelp() {
	help := `
C8S Environment Variables

Environment variables allow you to set default values for CLI options.
Priority: command-line flags > environment variables > defaults

Available Variables:

  C8S_DEV_CLUSTER <name>
    Default cluster name (default: c8s-dev)
    Example: export C8S_DEV_CLUSTER=dev-env

  C8S_DEV_CONFIG <path>
    Default path to cluster configuration file
    Example: export C8S_DEV_CONFIG=~/.c8s/config.yaml

  C8S_NAMESPACE <namespace>
    Default Kubernetes namespace (default: default)
    Example: export C8S_NAMESPACE=c8s-system

  C8S_VERBOSE <true|false>
    Enable verbose logging (default: false)
    Example: export C8S_VERBOSE=true

  C8S_QUIET <true|false>
    Suppress non-error output (default: false)
    Example: export C8S_QUIET=true

  C8S_NO_COLOR <true|false>
    Disable colored output (default: false)
    Example: export C8S_NO_COLOR=true

  C8S_DEV_TIMEOUT <seconds>
    Timeout for dev operations (default: 300)
    Example: export C8S_DEV_TIMEOUT=600

  C8S_IMAGE_PULL_POLICY <Always|IfNotPresent|Never>
    Default image pull policy (default: IfNotPresent)
    Example: export C8S_IMAGE_PULL_POLICY=Always

  C8S_REGISTRY_ENABLED <true|false>
    Enable registry in cluster (default: true)
    Example: export C8S_REGISTRY_ENABLED=true

Setup Example:

  # Add to ~/.bashrc or ~/.zshrc for persistent configuration
  export C8S_DEV_CLUSTER=my-dev-cluster
  export C8S_NAMESPACE=c8s-system
  export C8S_VERBOSE=true

Then use commands without repeating flags:

  c8s dev cluster create            # Uses C8S_DEV_CLUSTER
  c8s dev deploy operator           # Uses C8S_NAMESPACE
  c8s dev test run -v               # Already verbose from env
`

	fmt.Println(help)
}
