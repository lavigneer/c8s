package dev

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/org/c8s/pkg/localenv"
	"github.com/org/c8s/pkg/localenv/cluster"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newClusterCommand creates the cluster subcommand
func newClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage local Kubernetes clusters",
		Long: `Manage local Kubernetes test clusters for C8S operator development.

Create, delete, start, stop, and inspect local k3d clusters.`,
		Example: `  # Create a new cluster
  c8s dev cluster create

  # Delete a cluster
  c8s dev cluster delete my-cluster

  # Check cluster status
  c8s dev cluster status`,
	}

	// Add subcommands
	cmd.AddCommand(newClusterCreateCommand())
	cmd.AddCommand(newClusterDeleteCommand())
	cmd.AddCommand(newClusterStatusCommand())
	cmd.AddCommand(newClusterListCommand())
	cmd.AddCommand(newClusterStartCommand())
	cmd.AddCommand(newClusterStopCommand())

	return cmd
}

// newClusterCreateCommand creates the cluster create subcommand
func newClusterCreateCommand() *cobra.Command {
	var (
		configPath   string
		k8sVersion   string
		servers      int
		agents       int
		registry     bool
		registryPort int
		timeout      string
		wait         bool
	)

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a new local Kubernetes cluster",
		Long: `Create a new local Kubernetes cluster using k3d.

The cluster will be configured with the specified number of server and agent nodes,
and optionally with a local container registry.`,
		Example: `  # Create cluster with defaults
  c8s dev cluster create

  # Create cluster with custom name
  c8s dev cluster create my-test-cluster

  # Create from config file
  c8s dev cluster create --config k3d-config.yaml

  # Create with custom configuration
  c8s dev cluster create test --k8s-version v1.28.15 --agents 3`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Determine cluster name
			name := "c8s-dev"
			if len(args) > 0 {
				name = args[0]
			}

			// Parse timeout
			timeoutDuration, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			// Build cluster config
			var config *localenv.ClusterConfig
			if configPath != "" {
				// Load from file
				config, err = loadClusterConfigFromFile(configPath)
				if err != nil {
					return err
				}
			} else {
				// Build from flags
				config = buildClusterConfigFromFlags(name, k8sVersion, servers, agents, registry, registryPort)
			}

			// Create options
			opts := cluster.CreateOptions{
				Config:  config,
				Wait:    wait,
				Timeout: timeoutDuration,
			}

			// Create cluster
			if IsVerbose() {
				printInfo("[DEBUG] Creating cluster with config:")
				printInfo("[DEBUG]   Name: %s", config.Name)
				printInfo("[DEBUG]   K8s Version: %s", config.KubernetesVersion)
				printInfo("[DEBUG]   Servers: %d, Agents: %d", servers, agents)
				printInfo("[DEBUG]   Registry: %v", registry)
				printInfo("[DEBUG]   Timeout: %s", timeout)
			}

			printInfo("Creating cluster '%s'...", config.Name)
			status, err := cluster.Create(ctx, opts)
			if err != nil {
				// Enhance error with suggestions
				enhancedErr := cluster.EnhanceError(err, "create")

				if cluster.IsClusterAlreadyExistsError(err) {
					printError("Cluster '%s' already exists", config.Name)
					printInfo("Run 'c8s dev cluster delete %s' to remove it first", config.Name)
					return exitWithCode(2)
				}
				if cluster.IsDockerNotAvailableError(err) {
					printError("Docker is not available")
					printInfo("Please ensure Docker is installed and running")
					printInfo("Verify with: docker info")
					return exitWithCode(4)
				}
				if cluster.IsTimeoutError(err) {
					printError("Cluster creation timed out")
					printInfo("The cluster may still be starting. Check status with: c8s dev cluster status %s", config.Name)
					printInfo("Or try again with a longer timeout: --timeout 5m")
					return exitWithCode(3)
				}
				printError("Failed to create cluster: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Cluster created successfully")
				printInfo("[DEBUG] Status: %+v", status)
			}

			// Display success message
			printSuccess("Cluster '%s' created successfully", status.Name)
			printSuccess("Kubeconfig updated: ~/.kube/config")
			printSuccess("Current context: k3d-%s", status.Name)
			if status.RegistryEndpoint != "" {
				printSuccess("Registry available at: %s", status.RegistryEndpoint)
			}

			// Display cluster status
			fmt.Println()
			fmt.Println("Cluster Status:")
			fmt.Printf("  Name:        %s\n", status.Name)
			fmt.Printf("  Nodes:       %d\n", len(status.Nodes))
			if status.APIEndpoint != "" {
				fmt.Printf("  API Server:  %s\n", status.APIEndpoint)
			}

			// Display next steps
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  Deploy operator: c8s dev deploy operator")
			fmt.Println("  Check status:    c8s dev cluster status")

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to cluster config file")
	cmd.Flags().StringVar(&k8sVersion, "k8s-version", "v1.28.15", "Kubernetes version")
	cmd.Flags().IntVar(&servers, "servers", 1, "Number of server nodes")
	cmd.Flags().IntVar(&agents, "agents", 2, "Number of agent nodes")
	cmd.Flags().BoolVar(&registry, "registry", true, "Enable local registry")
	cmd.Flags().IntVar(&registryPort, "registry-port", 5000, "Registry host port")
	cmd.Flags().StringVar(&timeout, "timeout", "3m", "Creation timeout")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for cluster to be ready")

	return cmd
}

// buildClusterConfigFromFlags builds a ClusterConfig from command flags
func buildClusterConfigFromFlags(name, k8sVersion string, servers, agents int, registry bool, registryPort int) *localenv.ClusterConfig {
	config := &localenv.ClusterConfig{
		Name:              name,
		KubernetesVersion: k8sVersion,
		Nodes: []localenv.NodeConfig{
			{Type: "server", Count: servers},
			{Type: "agent", Count: agents},
		},
		Options: localenv.ClusterOptions{
			WaitTimeout:             "60s",
			UpdateDefaultKubeconfig: true,
			SwitchContext:           true,
			K3sArgs:                 []string{"--disable=traefik"},
		},
	}

	if registry {
		config.Registry = &localenv.RegistryConfig{
			Enabled:  true,
			Name:     "registry.localhost",
			HostPort: registryPort,
		}
	}

	return config
}

// loadClusterConfigFromFile loads cluster configuration from a file
func loadClusterConfigFromFile(path string) (*localenv.ClusterConfig, error) {
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

// Output formatting helpers

func printSuccess(format string, args ...interface{}) {
	if IsQuiet() {
		return
	}
	prefix := "✓"
	if IsColorDisabled() {
		fmt.Printf("%s "+format+"\n", append([]interface{}{prefix}, args...)...)
	} else {
		fmt.Printf("\033[32m%s\033[0m "+format+"\n", append([]interface{}{prefix}, args...)...)
	}
}

func printError(format string, args ...interface{}) {
	prefix := "✗"
	if IsColorDisabled() {
		fmt.Fprintf(os.Stderr, "%s "+format+"\n", append([]interface{}{prefix}, args...)...)
	} else {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m "+format+"\n", append([]interface{}{prefix}, args...)...)
	}
}

func printInfo(format string, args ...interface{}) {
	if IsQuiet() {
		return
	}
	fmt.Printf(format+"\n", args...)
}

func printWarning(format string, args ...interface{}) {
	if IsQuiet() {
		return
	}
	prefix := "⚠"
	if IsColorDisabled() {
		fmt.Printf("%s "+format+"\n", append([]interface{}{prefix}, args...)...)
	} else {
		fmt.Printf("\033[33m%s\033[0m "+format+"\n", append([]interface{}{prefix}, args...)...)
	}
}

// exitWithCode is a helper to return an error that will cause a specific exit code
func exitWithCode(code int) error {
	os.Exit(code)
	return nil
}

// formatJSON formats data as JSON
func formatJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// formatYAML formats data as YAML
func formatYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(v)
}

// formatTable formats a simple table
func formatTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	headerParts := make([]string, len(headers))
	for i, h := range headers {
		headerParts[i] = padRight(h, widths[i])
	}
	fmt.Println(strings.Join(headerParts, "   "))

	// Print rows
	for _, row := range rows {
		rowParts := make([]string, len(row))
		for i, cell := range row {
			if i < len(widths) {
				rowParts[i] = padRight(cell, widths[i])
			} else {
				rowParts[i] = cell
			}
		}
		fmt.Println(strings.Join(rowParts, "   "))
	}
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// newClusterDeleteCommand creates the cluster delete subcommand
func newClusterDeleteCommand() *cobra.Command {
	var (
		force bool
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "delete [NAME]",
		Short: "Delete a local Kubernetes cluster",
		Long: `Delete a local Kubernetes cluster and clean up all resources.

This will remove the cluster, Docker containers, and kubeconfig context.`,
		Example: `  # Delete default cluster (with confirmation)
  c8s dev cluster delete

  # Delete specific cluster
  c8s dev cluster delete my-test-cluster

  # Force deletion without prompt
  c8s dev cluster delete --force

  # Delete all clusters
  c8s dev cluster delete --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if all {
				// Delete all c8s clusters
				deleted, err := cluster.DeleteAll(ctx)
				if err != nil {
					printError("Failed to delete all clusters: %v", err)
					return exitWithCode(1)
				}

				if len(deleted) == 0 {
					printInfo("No c8s clusters found to delete")
					return nil
				}

				for _, name := range deleted {
					printSuccess("Cluster '%s' deleted successfully", name)
				}
				printSuccess("Kubeconfig contexts removed")
				return nil
			}

			// Determine cluster name
			name := "c8s-dev"
			if len(args) > 0 {
				name = args[0]
			}

			// Confirm deletion unless --force
			if !force {
				fmt.Printf("Warning: This will delete cluster '%s' and all its data\n", name)
				fmt.Printf("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "yes" && strings.ToLower(response) != "y" {
					printInfo("Deletion cancelled")
					return exitWithCode(130)
				}
			}

			// Delete cluster
			if IsVerbose() {
				printInfo("[DEBUG] Deleting cluster: %s (force=%v)", name, force)
			}

			printInfo("Deleting cluster '%s'...", name)
			err := cluster.Delete(ctx, cluster.DeleteOptions{
				Name:  name,
				Force: force,
			})
			if err != nil {
				enhancedErr := cluster.EnhanceError(err, "delete")

				if cluster.IsClusterNotFoundError(err) {
					printError("Cluster '%s' not found", name)
					printInfo("Run 'c8s dev cluster list' to see available clusters")
					return exitWithCode(2)
				}
				printError("Failed to delete cluster: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Cluster deleted successfully")
			}

			printSuccess("Cluster '%s' deleted successfully", name)
			printSuccess("Kubeconfig context removed")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")
	cmd.Flags().BoolVar(&all, "all", false, "Delete all c8s clusters")

	return cmd
}

// newClusterStatusCommand creates the cluster status subcommand
func newClusterStatusCommand() *cobra.Command {
	var (
		output string
		watch  bool
	)

	cmd := &cobra.Command{
		Use:   "status [NAME]",
		Short: "Show status of a local cluster",
		Long: `Display the current status of a local Kubernetes cluster.

Shows cluster state, nodes, API endpoint, and other details.`,
		Example: `  # Show status of current cluster
  c8s dev cluster status

  # Show status of specific cluster
  c8s dev cluster status my-test-cluster

  # Output as JSON
  c8s dev cluster status --output json

  # Watch status updates
  c8s dev cluster status --watch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Determine cluster name
			name := "c8s-dev"
			if len(args) > 0 {
				name = args[0]
			}

			// Get status
			if IsVerbose() {
				printInfo("[DEBUG] Getting status for cluster: %s", name)
			}

			status, err := cluster.GetStatusWithUptime(ctx, name)
			if err != nil {
				enhancedErr := cluster.EnhanceError(err, "status")

				if cluster.IsClusterNotFoundError(err) {
					printError("Cluster '%s' not found", name)
					printInfo("List available clusters with: c8s dev cluster list")
					return exitWithCode(2)
				}
				printError("Failed to get cluster status: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Retrieved status: state=%s, nodes=%d", status.State, len(status.Nodes))
			}

			// Format output
			switch output {
			case "json":
				return formatJSON(status)
			case "yaml":
				return formatYAML(status)
			default:
				// Text format
				fmt.Printf("Cluster Status: %s\n\n", status.Name)
				fmt.Printf("State:     %s\n", status.State)
				if status.Uptime != "" {
					fmt.Printf("Uptime:    %s\n", status.Uptime)
				}
				if status.APIEndpoint != "" {
					fmt.Printf("API:       %s\n", status.APIEndpoint)
				}
				if status.RegistryEndpoint != "" {
					fmt.Printf("Registry:  %s\n", status.RegistryEndpoint)
				}

				if len(status.Nodes) > 0 {
					fmt.Println("\nNodes:")
					for _, node := range status.Nodes {
						fmt.Printf("  %s   %s   %s   %s\n",
							padRight(node.Name, 25),
							padRight(node.Role, 6),
							padRight(node.Status, 8),
							node.Version)
					}
				}
			}

			if !status.IsRunning() {
				return exitWithCode(3)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text|json|yaml)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for status changes")

	return cmd
}

// newClusterListCommand creates the cluster list subcommand
func newClusterListCommand() *cobra.Command {
	var (
		output string
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all local clusters",
		Long: `List all local Kubernetes clusters.

By default, shows only c8s clusters. Use --all to show all k3d clusters.`,
		Example: `  # List c8s clusters
  c8s dev cluster list

  # List all k3d clusters
  c8s dev cluster list --all

  # Output as JSON
  c8s dev cluster list --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// List clusters
			if IsVerbose() {
				printInfo("[DEBUG] Listing clusters (all=%v)", all)
			}

			clusters, err := cluster.List(ctx, cluster.ListOptions{
				All: all,
			})
			if err != nil {
				enhancedErr := cluster.EnhanceError(err, "list")
				printError("Failed to list clusters: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Found %d clusters", len(clusters))
			}

			// Format output
			switch output {
			case "json":
				result := map[string]interface{}{
					"clusters": clusters,
				}
				return formatJSON(result)
			case "yaml":
				result := map[string]interface{}{
					"clusters": clusters,
				}
				return formatYAML(result)
			default:
				// Text format (table)
				if len(clusters) == 0 {
					printInfo("No clusters found")
					return nil
				}

				headers := []string{"NAME", "STATE", "NODES", "VERSION", "UPTIME"}
				rows := make([][]string, len(clusters))
				for i, c := range clusters {
					rows[i] = []string{
						c.Name,
						c.State,
						fmt.Sprintf("%d", c.NodeCount),
						c.Version,
						c.Uptime,
					}
				}
				formatTable(headers, rows)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text|json|yaml)")
	cmd.Flags().BoolVar(&all, "all", false, "Show all k3d clusters (not just c8s clusters)")

	return cmd
}

// newClusterStartCommand creates the cluster start subcommand
func newClusterStartCommand() *cobra.Command {
	var (
		wait    bool
		timeout string
	)

	cmd := &cobra.Command{
		Use:   "start [NAME]",
		Short: "Start a stopped cluster",
		Long: `Start a stopped local Kubernetes cluster.

The cluster state and data will be preserved.`,
		Example: `  # Start default cluster
  c8s dev cluster start

  # Start specific cluster
  c8s dev cluster start my-test-cluster`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Determine cluster name
			name := "c8s-dev"
			if len(args) > 0 {
				name = args[0]
			}

			// Parse timeout
			timeoutDuration, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			// Start cluster
			if IsVerbose() {
				printInfo("[DEBUG] Starting cluster: %s (wait=%v, timeout=%s)", name, wait, timeout)
			}

			printInfo("Starting cluster '%s'...", name)
			err = cluster.Start(ctx, cluster.StartOptions{
				Name:    name,
				Wait:    wait,
				Timeout: timeoutDuration,
			})
			if err != nil {
				enhancedErr := cluster.EnhanceError(err, "start")

				if cluster.IsClusterNotFoundError(err) {
					printError("Cluster '%s' not found", name)
					printInfo("List available clusters with: c8s dev cluster list")
					return exitWithCode(2)
				}
				if cluster.IsTimeoutError(err) {
					printError("Cluster start timed out")
					printInfo("The cluster may still be starting. Check status with: c8s dev cluster status %s", name)
					return exitWithCode(3)
				}
				printError("Failed to start cluster: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Cluster started successfully")
			}

			printSuccess("Cluster '%s' started successfully", name)

			return nil
		},
	}

	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for cluster to be ready")
	cmd.Flags().StringVar(&timeout, "timeout", "2m", "Start timeout")

	return cmd
}

// newClusterStopCommand creates the cluster stop subcommand
func newClusterStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [NAME]",
		Short: "Stop a running cluster (preserves state)",
		Long: `Stop a running local Kubernetes cluster.

The cluster state and data will be preserved and can be restarted later.`,
		Example: `  # Stop default cluster
  c8s dev cluster stop

  # Stop specific cluster
  c8s dev cluster stop my-test-cluster`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Determine cluster name
			name := "c8s-dev"
			if len(args) > 0 {
				name = args[0]
			}

			// Stop cluster
			if IsVerbose() {
				printInfo("[DEBUG] Stopping cluster: %s", name)
			}

			printInfo("Stopping cluster '%s'...", name)
			err := cluster.Stop(ctx, cluster.StopOptions{
				Name: name,
			})
			if err != nil {
				enhancedErr := cluster.EnhanceError(err, "stop")

				if cluster.IsClusterNotFoundError(err) {
					printError("Cluster '%s' not found", name)
					printInfo("List available clusters with: c8s dev cluster list")
					return exitWithCode(2)
				}
				printError("Failed to stop cluster: %v", enhancedErr)
				return exitWithCode(1)
			}

			if IsVerbose() {
				printInfo("[DEBUG] Cluster stopped successfully")
			}

			printSuccess("Cluster '%s' stopped successfully", name)

			return nil
		},
	}

	return cmd
}
