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

The dev command provides a complete local development workflow:
- Create isolated k3d Kubernetes clusters for testing
- Deploy the C8S operator and sample pipelines
- Run end-to-end tests and view results
- Manage cluster lifecycle (stop, start, delete)

All operations handle kubeconfig management automatically.

Workflows:

  Create and test complete setup:
    $ c8s dev cluster create dev-env
    $ c8s dev deploy operator --cluster dev-env
    $ c8s dev deploy samples --cluster dev-env
    $ c8s dev test run --cluster dev-env

  Manage cluster state:
    $ c8s dev cluster stop dev-env    # Pause cluster
    $ c8s dev cluster start dev-env   # Resume cluster
    $ c8s dev cluster delete dev-env  # Clean up

  Development iteration:
    $ c8s dev test run --cluster dev-env --output json
    $ c8s dev test logs --cluster dev-env --follow`,
		Example: `  # Create and setup local environment
  c8s dev cluster create my-env
  c8s dev deploy operator --cluster my-env
  c8s dev deploy samples --cluster my-env

  # Run tests
  c8s dev test run --cluster my-env

  # View logs
  c8s dev test logs --cluster my-env --follow

  # Cleanup
  c8s dev cluster delete my-env`,
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
