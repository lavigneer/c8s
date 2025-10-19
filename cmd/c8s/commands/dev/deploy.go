package dev

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/org/c8s/pkg/localenv/cluster"
	"github.com/org/c8s/pkg/localenv/deploy"
	"github.com/org/c8s/pkg/localenv/samples"
)

// newDeployCommand creates the deploy subcommand
func newDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy C8S operator and samples to a cluster",
		Long: `Deploy the C8S operator and sample PipelineConfigs to a local Kubernetes cluster.

This command manages all aspects of deploying the C8S system:
- Installing CRDs
- Loading operator images
- Deploying the operator
- Deploying sample PipelineConfigs

Use 'c8s dev deploy operator' to deploy the operator and 'c8s dev deploy samples' to deploy samples.`,
	}

	cmd.AddCommand(newDeployOperatorCommand())
	cmd.AddCommand(newDeploySamplesCommand())

	return cmd
}

// newDeployOperatorCommand creates the deploy operator subcommand
func newDeployOperatorCommand() *cobra.Command {
	var (
		clusterName              string
		deployOperatorImage      string
		deployOperatorImagePolicy string
		deployOperatorNamespace  string
		deployOperatorCRDsPath   string
		deployOperatorManifests  string
		wait                     bool
		timeout                  int
	)

	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Deploy C8S operator to a cluster",
		Long: `Deploy the C8S operator to a local Kubernetes cluster.

This command:
1. Installs all required CRDs
2. Loads the operator image into the cluster
3. Deploys the operator and its configuration
4. Waits for the operator to be ready (if --wait is set)

Example:
  c8s dev deploy operator --cluster c8s-dev --namespace c8s-system
  c8s dev deploy operator --image ghcr.io/custom/controller:v0.1.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "Deploying operator to cluster %q\n", clusterName)
			}

			// Create kubectl client
			kubectlClient := cluster.NewKubectlClient()

			// Install CRDs
			if verbose {
				fmt.Fprintf(os.Stderr, "Installing CRDs from %s...\n", deployOperatorCRDsPath)
			}
			crdStatus, err := deploy.InstallCRDs(ctx, kubectlClient, deployOperatorCRDsPath)
			if err != nil {
				return fmt.Errorf("failed to install CRDs: %w", err)
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "CRDs installed: %s\n", strings.Join(crdStatus.CRDsInstalled, ", "))
			}

			// Load image to cluster
			if verbose {
				fmt.Fprintf(os.Stderr, "Loading image to cluster...\n")
			}
			imageStatus, err := deploy.LoadImageToCluster(clusterName, deployOperatorImage)
			if err != nil {
				return fmt.Errorf("failed to load image: %w", err)
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "Image loaded: %s\n", imageStatus.ImageName)
			}

			// Deploy operator
			if verbose {
				fmt.Fprintf(os.Stderr, "Deploying operator to namespace %s...\n", deployOperatorNamespace)
			}
			opStatus, err := deploy.DeployOperator(
				ctx,
				kubectlClient,
				clusterName,
				deployOperatorNamespace,
				deployOperatorManifests,
				deployOperatorImage,
				deployOperatorImagePolicy,
			)
			if err != nil {
				return fmt.Errorf("failed to deploy operator: %w", err)
			}

			// Display success message
			fmt.Printf("✓ Operator deployed successfully\n")
			fmt.Printf("  Namespace: %s\n", opStatus.Namespace)
			fmt.Printf("  Deployment: %s\n", opStatus.DeploymentName)
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  1. Deploy sample pipelines: c8s dev deploy samples --cluster %s\n", clusterName)
			fmt.Printf("  2. Run tests: c8s dev test run --cluster %s\n", clusterName)

			return nil
		},
	}

	// Operator flags
	cmd.Flags().StringVar(&clusterName, "cluster", "c8s-dev",
		"Name of the cluster to deploy to")
	cmd.Flags().StringVar(&deployOperatorImage, "image", "",
		"Docker image for the operator (default: ghcr.io/org/c8s-controller:latest)")
	cmd.Flags().StringVar(&deployOperatorImagePolicy, "image-pull-policy", "IfNotPresent",
		"Image pull policy (Always, IfNotPresent, Never)")
	cmd.Flags().StringVar(&deployOperatorNamespace, "namespace", "c8s-system",
		"Kubernetes namespace to deploy operator into")
	cmd.Flags().StringVar(&deployOperatorCRDsPath, "crds-path", "config/crd/bases",
		"Path to CRD manifests")
	cmd.Flags().StringVar(&deployOperatorManifests, "manifests-path", "config/manager",
		"Path to operator manifests")
	cmd.Flags().BoolVar(&wait, "wait", true,
		"Wait for operator deployment to be ready")
	cmd.Flags().IntVar(&timeout, "timeout", 300,
		"Timeout in seconds for operator to become ready")

	return cmd
}

// newDeploySamplesCommand creates the deploy samples subcommand
func newDeploySamplesCommand() *cobra.Command {
	var (
		clusterName           string
		deploySamplesNamespace string
		deploySamplesSamplesPath string
		deploySamplesSelect    string
	)

	cmd := &cobra.Command{
		Use:   "samples",
		Short: "Deploy sample PipelineConfigs to a cluster",
		Long: `Deploy sample PipelineConfigs to a Kubernetes cluster.

This command deploys example PipelineConfig resources that demonstrate
the capabilities of the C8S operator.

The operator must be deployed before deploying samples.

Example:
  c8s dev deploy samples --cluster c8s-dev
  c8s dev deploy samples --select simple-build --namespace custom-ns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				fmt.Fprintf(os.Stderr, "Deploying samples to cluster %q\n", clusterName)
			}

			// Deploy samples
			samplesStatus, err := samples.DeploySamples(
				deploySamplesNamespace,
				deploySamplesSamplesPath,
				deploySamplesSelect,
			)
			if err != nil {
				return fmt.Errorf("failed to deploy samples: %w", err)
			}

			if len(samplesStatus.SamplesDeployed) == 0 {
				fmt.Printf("No samples deployed (no matching files found)\n")
				return nil
			}

			// Display success message
			fmt.Printf("✓ Samples deployed successfully\n")
			fmt.Printf("  Namespace: %s\n", samplesStatus.Namespace)
			fmt.Printf("  Samples deployed:\n")
			for _, sample := range samplesStatus.SamplesDeployed {
				fmt.Printf("    - %s\n", sample)
			}

			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  1. View samples: kubectl get pipelineconfigs -n %s\n", deploySamplesNamespace)
			fmt.Printf("  2. Run tests: c8s dev test run --cluster %s\n", clusterName)

			return nil
		},
	}

	// Samples flags
	cmd.Flags().StringVar(&clusterName, "cluster", "c8s-dev",
		"Name of the cluster to deploy to")
	cmd.Flags().StringVar(&deploySamplesNamespace, "namespace", "default",
		"Kubernetes namespace to deploy samples into")
	cmd.Flags().StringVar(&deploySamplesSamplesPath, "samples-path", "config/samples",
		"Path to sample manifests")
	cmd.Flags().StringVar(&deploySamplesSelect, "select", "",
		"Filter samples by name (e.g., 'simple' to deploy simple-build.yaml)")

	return cmd
}
