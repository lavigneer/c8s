package cluster

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// KubectlClient interface defines operations for kubectl interactions
type KubectlClient interface {
	// ApplyManifest applies a Kubernetes manifest from a file
	ApplyManifest(ctx context.Context, manifestPath string, namespace string) error

	// ApplyManifestFromString applies a Kubernetes manifest from a string
	ApplyManifestFromString(ctx context.Context, manifestYAML string, namespace string) error

	// DeleteResource deletes a Kubernetes resource
	DeleteResource(ctx context.Context, resourceType, name, namespace string) error

	// GetResource gets a Kubernetes resource as JSON
	GetResource(ctx context.Context, resourceType, name, namespace string) ([]byte, error)

	// WaitForReady waits for a resource to be ready
	WaitForReady(ctx context.Context, resourceType, name, namespace string, timeout time.Duration) error

	// GetLogs gets logs from a pod
	GetLogs(ctx context.Context, podName, namespace string, follow bool, tail int) ([]byte, error)

	// SetContext switches kubectl context
	SetContext(ctx context.Context, contextName string) error

	// GetCurrentContext gets the current kubectl context
	GetCurrentContext(ctx context.Context) (string, error)

	// GetNodes gets cluster nodes status for a specific cluster
	GetNodes(ctx context.Context, clusterName string) ([]KubeNode, error)
}

// kubectlClientImpl implements KubectlClient using kubectl command-line tool
type kubectlClientImpl struct {
	execTimeout time.Duration
}

// NewKubectlClient creates a new kubectl client
func NewKubectlClient() KubectlClient {
	return &kubectlClientImpl{
		execTimeout: 5 * time.Minute,
	}
}

// ApplyManifest applies a Kubernetes manifest from a file
func (k *kubectlClientImpl) ApplyManifest(ctx context.Context, manifestPath string, namespace string) error {
	args := []string{"apply", "-f", manifestPath}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return k.runKubectlCommand(ctx, args...)
}

// ApplyManifestFromString applies a Kubernetes manifest from a string
func (k *kubectlClientImpl) ApplyManifestFromString(ctx context.Context, manifestYAML string, namespace string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	if namespace != "" {
		cmd.Args = append(cmd.Args, "-n", namespace)
	}

	cmd.Stdin = strings.NewReader(manifestYAML)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("kubectl apply failed: %s", stderr.String())
		}
		return fmt.Errorf("kubectl apply failed: %w", err)
	}

	return nil
}

// DeleteResource deletes a Kubernetes resource
func (k *kubectlClientImpl) DeleteResource(ctx context.Context, resourceType, name, namespace string) error {
	args := []string{"delete", resourceType, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	// Ignore not found errors
	args = append(args, "--ignore-not-found")
	return k.runKubectlCommand(ctx, args...)
}

// GetResource gets a Kubernetes resource as JSON
func (k *kubectlClientImpl) GetResource(ctx context.Context, resourceType, name, namespace string) ([]byte, error) {
	args := []string{"get", resourceType, name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return k.runKubectlCommandWithOutput(ctx, args...)
}

// WaitForReady waits for a resource to be ready
func (k *kubectlClientImpl) WaitForReady(ctx context.Context, resourceType, name, namespace string, timeout time.Duration) error {
	args := []string{"wait", fmt.Sprintf("%s/%s", resourceType, name), "--for=condition=Ready"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, fmt.Sprintf("--timeout=%s", timeout))

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return k.runKubectlCommand(waitCtx, args...)
}

// GetLogs gets logs from a pod
func (k *kubectlClientImpl) GetLogs(ctx context.Context, podName, namespace string, follow bool, tail int) ([]byte, error) {
	args := []string{"logs", podName}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", tail))
	}
	return k.runKubectlCommandWithOutput(ctx, args...)
}

// SetContext switches kubectl context
func (k *kubectlClientImpl) SetContext(ctx context.Context, contextName string) error {
	return k.runKubectlCommand(ctx, "config", "use-context", contextName)
}

// GetCurrentContext gets the current kubectl context
func (k *kubectlClientImpl) GetCurrentContext(ctx context.Context) (string, error) {
	output, err := k.runKubectlCommandWithOutput(ctx, "config", "current-context")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// KubeNode represents a Kubernetes node
type KubeNode struct {
	Name    string
	Role    string
	Status  string
	Version string
}

// GetNodes gets cluster nodes status for a specific cluster
func (k *kubectlClientImpl) GetNodes(ctx context.Context, clusterName string) ([]KubeNode, error) {
	// Set context to the cluster
	contextName := fmt.Sprintf("k3d-%s", clusterName)
	args := []string{"get", "nodes", "--context", contextName, "-o", "wide", "--no-headers"}

	output, err := k.runKubectlCommandWithOutput(ctx, args...)
	if err != nil {
		return nil, err
	}

	return parseNodesOutput(string(output)), nil
}

// parseNodesOutput parses kubectl get nodes output
func parseNodesOutput(output string) []KubeNode {
	var nodes []KubeNode
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Extract role from labels/roles
		role := "agent"
		if strings.Contains(line, "control-plane") || strings.Contains(line, "master") {
			role = "server"
		}

		nodes = append(nodes, KubeNode{
			Name:    fields[0],
			Status:  fields[1],
			Role:    role,
			Version: fields[4],
		})
	}

	return nodes
}

// runKubectlCommand executes a kubectl command and returns error if it fails
func (k *kubectlClientImpl) runKubectlCommand(ctx context.Context, args ...string) error {
	_, err := k.runKubectlCommandWithOutput(ctx, args...)
	return err
}

// runKubectlCommandWithOutput executes a kubectl command and returns output
func (k *kubectlClientImpl) runKubectlCommandWithOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("kubectl command failed: %s", stderr.String())
		}
		return nil, fmt.Errorf("kubectl command failed: %w", err)
	}

	return stdout.Bytes(), nil
}
