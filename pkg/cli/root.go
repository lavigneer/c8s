package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig string
	namespace  string
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
)

func init() {
	// Setup flags
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", "default", "kubernetes namespace")
}

// Execute is the entry point for the CLI
func Execute() error {
	flag.Parse()

	// Get subcommand
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("no command specified. Available commands: run, get, validate")
	}

	command := args[0]
	commandArgs := args[1:]

	// Initialize Kubernetes client
	if command != "validate" {
		if err := initKubeClient(); err != nil {
			return fmt.Errorf("failed to initialize kubernetes client: %w", err)
		}
	}

	// Route to subcommands
	switch command {
	case "run":
		return runCommand(commandArgs)
	case "get":
		return getCommand(commandArgs)
	case "validate":
		return validateCommand(commandArgs)
	case "logs":
		return logsCommand(commandArgs)
	default:
		return fmt.Errorf("unknown command: %s. Available commands: run, get, validate, logs", command)
	}
}

func initKubeClient() error {
	var err error

	// Use in-cluster config if running in a pod
	restConfig, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to build config: %w", err)
		}
	}

	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `c8s - Kubernetes-native CI system

Usage:
  c8s run <pipeline-config-name> --commit=<sha> --branch=<name>
  c8s get runs [<name>]
  c8s get configs [<name>]
  c8s validate <pipeline-yaml-file>
  c8s logs <pipelinerun-name> --step=<step-name> [--follow]

Flags:
  --kubeconfig string   Path to kubeconfig file (default: $HOME/.kube/config)
  --namespace string    Kubernetes namespace (default: "default")

Examples:
  # Run a pipeline manually
  c8s run my-pipeline --commit=abc123 --branch=main

  # List all pipeline runs
  c8s get runs

  # Get details of a specific run
  c8s get runs my-run-12345

  # Validate a pipeline configuration
  c8s validate .c8s.yaml

  # Stream logs from a pipeline step
  c8s logs my-run-12345 --step=test --follow
`)
}
