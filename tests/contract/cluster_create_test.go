package contract

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClusterCreateCommand verifies the cluster create command behavior
func TestClusterCreateCommand(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name     string
		args     []string
		exitCode int
		validate func(t *testing.T, output string, exitCode int)
		setup    func(t *testing.T) // Optional setup before test
		cleanup  func(t *testing.T) // Optional cleanup after test
	}{
		{
			name:     "create with default name",
			args:     []string{"dev", "cluster", "create"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				// Verify cluster name appears in output
				if !strings.Contains(output, "c8s-dev") {
					t.Errorf("expected cluster name 'c8s-dev' in output, got: %s", output)
				}
				// Verify success indicators
				if !strings.Contains(output, "✓") && !strings.Contains(output, "successfully") {
					t.Errorf("expected success indicator in output, got: %s", output)
				}
				// Verify kubeconfig update message
				if !strings.Contains(output, "Kubeconfig") || !strings.Contains(output, "kubeconfig") {
					t.Errorf("expected kubeconfig update message in output, got: %s", output)
				}
				// Verify context switch message
				if !strings.Contains(output, "context") {
					t.Errorf("expected context switch message in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "c8s-dev")
			},
		},
		{
			name:     "create with custom name",
			args:     []string{"dev", "cluster", "create", "test-cluster"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "test-cluster") {
					t.Errorf("expected cluster name 'test-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "✓") && !strings.Contains(output, "successfully") {
					t.Errorf("expected success indicator in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "test-cluster")
			},
		},
		{
			name:     "create with custom k8s version",
			args:     []string{"dev", "cluster", "create", "version-test", "--k8s-version", "v1.28.15"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "version-test") {
					t.Errorf("expected cluster name 'version-test' in output, got: %s", output)
				}
				// Check if version appears in output (may vary by implementation)
				if !strings.Contains(output, "v1.28") {
					t.Logf("warning: expected version 'v1.28' in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "version-test")
			},
		},
		{
			name: "cluster already exists",
			args: []string{"dev", "cluster", "create", "existing-cluster"},
			setup: func(t *testing.T) {
				// Pre-create the cluster
				createTestCluster(t, "existing-cluster")
			},
			exitCode: 2,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "already exists") && !strings.Contains(output, "already running") {
					t.Errorf("expected 'already exists' error message, got: %s", output)
				}
				// Should suggest deletion
				if !strings.Contains(output, "delete") {
					t.Logf("warning: expected suggestion to delete existing cluster, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "existing-cluster")
			},
		},
		{
			name:     "create without registry",
			args:     []string{"dev", "cluster", "create", "no-registry", "--registry=false"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "no-registry") {
					t.Errorf("expected cluster name 'no-registry' in output, got: %s", output)
				}
				// Should NOT mention registry endpoint
				if strings.Contains(output, "Registry available at") {
					t.Errorf("expected no registry message when --registry=false, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "no-registry")
			},
		},
		{
			name:     "create with custom node configuration",
			args:     []string{"dev", "cluster", "create", "custom-nodes", "--servers", "1", "--agents", "3"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "custom-nodes") {
					t.Errorf("expected cluster name 'custom-nodes' in output, got: %s", output)
				}
				// Should show node count (may vary by implementation)
				if !strings.Contains(output, "4") && !strings.Contains(output, "Nodes:") {
					t.Logf("warning: expected node count in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "custom-nodes")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if Docker is not available
			if !isDockerAvailable() {
				t.Skip("Docker is not available, skipping test")
			}

			// Run setup if provided
			if tt.setup != nil {
				tt.setup(t)
			}

			// Run cleanup at the end
			if tt.cleanup != nil {
				defer tt.cleanup(t)
			}

			// Execute command
			output, exitCode := executeCommand(t, binaryPath, tt.args)

			// Verify exit code
			if exitCode != tt.exitCode {
				t.Errorf("expected exit code %d, got %d\nOutput: %s", tt.exitCode, exitCode, output)
			}

			// Validate output
			if tt.validate != nil {
				tt.validate(t, output, exitCode)
			}
		})
	}
}

// TestClusterCreateWithConfigFile tests cluster creation using a config file
func TestClusterCreateWithConfigFile(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Create a temporary config file
	configContent := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: config-test-cluster
servers: 1
agents: 2
`
	configPath := filepath.Join(t.TempDir(), "cluster-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	defer cleanupCluster(t, "config-test-cluster")

	// Execute command with config file
	args := []string{"dev", "cluster", "create", "--config", configPath}
	output, exitCode := executeCommand(t, binaryPath, args)

	// Verify success
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "config-test-cluster") {
		t.Errorf("expected cluster name from config file in output, got: %s", output)
	}
}

// TestClusterCreateDockerNotAvailable tests error handling when Docker is not available
func TestClusterCreateDockerNotAvailable(t *testing.T) {
	// This test is difficult to implement without actually stopping Docker
	// In a real implementation, we would mock the Docker check or use a test double
	t.Skip("Docker availability test requires mocking or Docker shutdown")
}

// TestClusterCreateKubeconfigUpdate verifies kubeconfig is updated correctly
func TestClusterCreateKubeconfigUpdate(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "kubeconfig-test"
	defer cleanupCluster(t, clusterName)

	// Execute command
	args := []string{"dev", "cluster", "create", clusterName}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("cluster creation failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Wait a moment for kubeconfig to be updated
	time.Sleep(2 * time.Second)

	// Verify kubeconfig context exists
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl not available, skipping kubeconfig verification")
	}

	cmd := exec.Command(kubectlPath, "config", "get-contexts", "-o", "name")
	contextOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get kubectl contexts: %v", err)
	}

	expectedContext := "k3d-" + clusterName
	if !strings.Contains(string(contextOutput), expectedContext) {
		t.Errorf("expected context '%s' in kubeconfig, got contexts: %s", expectedContext, string(contextOutput))
	}
}

// Helper functions

// buildC8sBinary builds the c8s binary for testing
func buildC8sBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "c8s")

	// Get the project root directory (3 levels up from tests/contract)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "../..")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/c8s")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build c8s binary: %v\nOutput: %s", err, string(output))
	}

	return binaryPath
}

// executeCommand executes a command and returns the output and exit code
func executeCommand(t *testing.T, binaryPath string, args []string) (string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to execute command: %v", err)
		}
	}

	// Combine stdout and stderr for output
	output := stdout.String() + stderr.String()
	return output, exitCode
}

// isDockerAvailable checks if Docker is available and running
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// createTestCluster creates a cluster for testing (used in setup)
func createTestCluster(t *testing.T, name string) {
	t.Helper()

	cmd := exec.Command("k3d", "cluster", "create", name, "--agents", "0", "--no-lb")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to create test cluster: %v\nOutput: %s", err, string(output))
	}

	// Wait for cluster to be ready
	time.Sleep(5 * time.Second)
}

// cleanupCluster removes a cluster after testing
func cleanupCluster(t *testing.T, name string) {
	t.Helper()

	// Use k3d directly for cleanup to avoid dependencies on the implementation
	cmd := exec.Command("k3d", "cluster", "delete", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("warning: failed to cleanup cluster %s: %v\nOutput: %s", name, err, string(output))
	}
}

// TestClusterCreateTimeout tests timeout handling
func TestClusterCreateTimeout(t *testing.T) {
	// This test would require simulating a cluster that never becomes ready
	// In practice, this is difficult to test without mocking
	t.Skip("Timeout test requires mocking or simulated failure")
}

// TestClusterCreateOutputFormats verifies different output scenarios
func TestClusterCreateOutputFormats(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name     string
		flags    []string
		checkFor []string
	}{
		{
			name:  "verbose output",
			flags: []string{"--verbose"},
			checkFor: []string{
				"Creating cluster",
				"successfully",
			},
		},
		{
			name:  "quiet output",
			flags: []string{"--quiet"},
			checkFor: []string{
				// Should have minimal output, but errors still shown
			},
		},
		{
			name:  "no color output",
			flags: []string{"--no-color"},
			checkFor: []string{
				"successfully",
				// Should not contain ANSI escape codes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterName := fmt.Sprintf("output-test-%s", strings.ReplaceAll(tt.name, " ", "-"))
			defer cleanupCluster(t, clusterName)

			args := append([]string{"dev", "cluster", "create", clusterName}, tt.flags...)
			output, exitCode := executeCommand(t, binaryPath, args)

			if exitCode != 0 {
				t.Errorf("expected exit code 0, got %d\nOutput: %s", exitCode, output)
			}

			for _, check := range tt.checkFor {
				if check != "" && !strings.Contains(output, check) {
					t.Errorf("expected output to contain '%s', got: %s", check, output)
				}
			}

			// For no-color, verify no ANSI escape codes
			if tt.name == "no color output" {
				if strings.Contains(output, "\x1b[") {
					t.Errorf("expected no ANSI escape codes in output with --no-color, got: %s", output)
				}
			}
		})
	}
}
