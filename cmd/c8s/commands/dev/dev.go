package dev

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags for dev commands
	verbose  bool
	quiet    bool
	noColor  bool
)

// NewDevCommand creates the dev command with subcommands
func NewDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Development environment management for local testing",
		Long: `Manage local Kubernetes test environments for C8S operator development.

The dev command provides tools to create local clusters, deploy the operator,
and run end-to-end pipeline tests without requiring cloud infrastructure.`,
		Example: `  # Create a local cluster
  c8s dev cluster create

  # Deploy operator to local cluster
  c8s dev deploy operator

  # Run pipeline tests
  c8s dev test run`,
	}

	// Add global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add subcommands
	cmd.AddCommand(newClusterCommand())
	cmd.AddCommand(newDeployCommand())
	cmd.AddCommand(newTestCommand())

	return cmd
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// IsQuiet returns whether quiet mode is enabled
func IsQuiet() bool {
	return quiet
}

// IsColorDisabled returns whether color output is disabled
func IsColorDisabled() bool {
	return noColor
}
