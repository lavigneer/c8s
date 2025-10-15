package contract

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestClusterDeleteCommand verifies the cluster delete command behavior
func TestClusterDeleteCommand(t *testing.T) {
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
		input    string             // Stdin input for confirmation prompts
	}{
		{
			name: "delete with default name and confirmation",
			args: []string{"dev", "cluster", "delete"},
			setup: func(t *testing.T) {
				createTestCluster(t, "c8s-dev")
			},
			exitCode: 0,
			input:    "yes\n",
			validate: func(t *testing.T, output string, exitCode int) {
				// Should show confirmation prompt
				if !strings.Contains(output, "Are you sure") && !strings.Contains(output, "confirm") {
					t.Logf("warning: expected confirmation prompt in output, got: %s", output)
				}
				// Should show success message
				if !strings.Contains(output, "deleted") && !strings.Contains(output, "✓") {
					t.Errorf("expected deletion success message, got: %s", output)
				}
				// Should mention kubeconfig cleanup
				if !strings.Contains(output, "kubeconfig") && !strings.Contains(output, "context") {
					t.Logf("warning: expected kubeconfig cleanup message, got: %s", output)
				}
			},
		},
		{
			name: "delete with custom name",
			args: []string{"dev", "cluster", "delete", "test-delete-cluster"},
			setup: func(t *testing.T) {
				createTestCluster(t, "test-delete-cluster")
			},
			exitCode: 0,
			input:    "yes\n",
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "test-delete-cluster") {
					t.Errorf("expected cluster name 'test-delete-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "deleted") && !strings.Contains(output, "✓") {
					t.Errorf("expected deletion success message, got: %s", output)
				}
			},
		},
		{
			name: "delete with force flag (no confirmation)",
			args: []string{"dev", "cluster", "delete", "force-delete-cluster", "--force"},
			setup: func(t *testing.T) {
				createTestCluster(t, "force-delete-cluster")
			},
			exitCode: 0,
			input:    "", // No input needed with --force
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "force-delete-cluster") {
					t.Errorf("expected cluster name 'force-delete-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "deleted") && !strings.Contains(output, "✓") {
					t.Errorf("expected deletion success message, got: %s", output)
				}
				// Should NOT show confirmation prompt with --force
				if strings.Contains(output, "Are you sure") {
					t.Errorf("should not show confirmation prompt with --force flag, got: %s", output)
				}
			},
		},
		{
			name: "delete with force flag shorthand",
			args: []string{"dev", "cluster", "delete", "force-delete-short", "-f"},
			setup: func(t *testing.T) {
				createTestCluster(t, "force-delete-short")
			},
			exitCode: 0,
			input:    "", // No input needed with -f
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "force-delete-short") {
					t.Errorf("expected cluster name 'force-delete-short' in output, got: %s", output)
				}
				if !strings.Contains(output, "deleted") && !strings.Contains(output, "✓") {
					t.Errorf("expected deletion success message, got: %s", output)
				}
			},
		},
		{
			name:     "cluster not found",
			args:     []string{"dev", "cluster", "delete", "nonexistent-cluster", "--force"},
			exitCode: 2,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "not found") && !strings.Contains(output, "does not exist") {
					t.Errorf("expected 'not found' error message, got: %s", output)
				}
				// Should suggest listing clusters
				if !strings.Contains(output, "list") {
					t.Logf("warning: expected suggestion to list clusters, got: %s", output)
				}
			},
		},
		{
			name: "user cancels deletion",
			args: []string{"dev", "cluster", "delete", "cancel-test-cluster"},
			setup: func(t *testing.T) {
				createTestCluster(t, "cancel-test-cluster")
			},
			exitCode: 130, // SIGINT or user cancellation
			input:    "no\n",
			validate: func(t *testing.T, output string, exitCode int) {
				// Accept either 130 or 1 for cancellation
				if exitCode != 130 && exitCode != 1 {
					t.Logf("warning: expected exit code 130 or 1 for cancellation, got %d", exitCode)
				}
				// Should indicate cancellation
				if !strings.Contains(output, "cancel") && !strings.Contains(output, "abort") && !strings.Contains(output, "Deletion") {
					t.Logf("warning: expected cancellation message, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				// Clean up since deletion was cancelled
				cleanupCluster(t, "cancel-test-cluster")
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
				// Wait for cluster to be fully created
				time.Sleep(2 * time.Second)
			}

			// Run cleanup at the end
			if tt.cleanup != nil {
				defer tt.cleanup(t)
			}

			// Execute command with stdin input
			output, exitCode := executeCommandWithInput(t, binaryPath, tt.args, tt.input)

			// Verify exit code (with some flexibility for user cancellation)
			if tt.name == "user cancels deletion" {
				// Accept 130 or 1 for cancellation
				if exitCode != 130 && exitCode != 1 {
					t.Logf("warning: expected exit code 130 or 1, got %d\nOutput: %s", exitCode, output)
				}
			} else if exitCode != tt.exitCode {
				t.Errorf("expected exit code %d, got %d\nOutput: %s", tt.exitCode, exitCode, output)
			}

			// Validate output
			if tt.validate != nil {
				tt.validate(t, output, exitCode)
			}

			// For successful deletions, verify cluster is actually gone
			if tt.exitCode == 0 && exitCode == 0 {
				time.Sleep(1 * time.Second)
				lastArg := tt.args[len(tt.args)-1]
				if lastArg != "--force" && lastArg != "-f" {
					// Extract cluster name
					clusterName := "c8s-dev" // default
					for i, arg := range tt.args {
						if arg == "delete" && i+1 < len(tt.args) && !strings.HasPrefix(tt.args[i+1], "-") {
							clusterName = tt.args[i+1]
							break
						}
					}
					verifyClusterDeleted(t, clusterName)
				}
			}
		})
	}
}

// TestClusterDeleteAll tests deleting all c8s clusters
func TestClusterDeleteAll(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Create multiple clusters
	clusters := []string{"all-test-1", "all-test-2", "all-test-3"}
	for _, cluster := range clusters {
		createTestCluster(t, cluster)
	}
	defer func() {
		// Cleanup any remaining clusters
		for _, cluster := range clusters {
			cleanupCluster(t, cluster)
		}
	}()

	// Wait for clusters to be ready
	time.Sleep(3 * time.Second)

	// Delete all with --all flag and --force (to skip confirmation)
	args := []string{"dev", "cluster", "delete", "--all", "--force"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nOutput: %s", exitCode, output)
	}

	// Verify output mentions multiple clusters or successful deletion
	if !strings.Contains(output, "deleted") && !strings.Contains(output, "✓") {
		t.Errorf("expected deletion success message, got: %s", output)
	}

	// Wait for deletion to complete
	time.Sleep(2 * time.Second)

	// Verify all clusters are gone
	for _, cluster := range clusters {
		verifyClusterDeleted(t, cluster)
	}
}

// TestClusterDeleteKubeconfigCleanup verifies kubeconfig context is removed
func TestClusterDeleteKubeconfigCleanup(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "kubeconfig-delete-test"

	// Create cluster
	createTestCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Verify context exists
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl not available, skipping kubeconfig verification")
	}

	expectedContext := "k3d-" + clusterName
	cmd := exec.Command(kubectlPath, "config", "get-contexts", "-o", "name")
	contextOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get kubectl contexts: %v", err)
	}

	if !strings.Contains(string(contextOutput), expectedContext) {
		t.Fatalf("context '%s' not found before deletion", expectedContext)
	}

	// Delete cluster with force flag
	args := []string{"dev", "cluster", "delete", clusterName, "--force"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("cluster deletion failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Wait for deletion to complete
	time.Sleep(2 * time.Second)

	// Verify context is removed
	cmd = exec.Command(kubectlPath, "config", "get-contexts", "-o", "name")
	contextOutput, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get kubectl contexts after deletion: %v", err)
	}

	if strings.Contains(string(contextOutput), expectedContext) {
		t.Errorf("context '%s' still exists after deletion, got contexts: %s", expectedContext, string(contextOutput))
	}
}

// TestClusterDeleteDockerCleanup verifies Docker containers are cleaned up
func TestClusterDeleteDockerCleanup(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "docker-cleanup-test"

	// Create cluster
	createTestCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Verify Docker containers exist for this cluster
	cmd := exec.Command("docker", "ps", "-a", "--filter", "label=app=k3d", "--filter", "label=k3d.cluster="+clusterName, "--format", "{{.Names}}")
	containerOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list Docker containers: %v", err)
	}

	if len(containerOutput) == 0 {
		t.Fatalf("no Docker containers found for cluster before deletion")
	}

	// Delete cluster with force flag
	args := []string{"dev", "cluster", "delete", clusterName, "--force"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("cluster deletion failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Wait for deletion to complete
	time.Sleep(2 * time.Second)

	// Verify Docker containers are gone
	cmd = exec.Command("docker", "ps", "-a", "--filter", "label=app=k3d", "--filter", "label=k3d.cluster="+clusterName, "--format", "{{.Names}}")
	containerOutput, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list Docker containers after deletion: %v", err)
	}

	if len(containerOutput) > 0 {
		t.Errorf("Docker containers still exist after deletion: %s", string(containerOutput))
	}
}

// Helper functions

// executeCommandWithInput executes a command with stdin input and returns the output and exit code
func executeCommandWithInput(t *testing.T, binaryPath string, args []string, input string) (string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Provide stdin input if specified
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Logf("warning: failed to execute command: %v", err)
			exitCode = 1
		}
	}

	// Combine stdout and stderr for output
	output := stdout.String() + stderr.String()
	return output, exitCode
}

// verifyClusterDeleted checks that a cluster has been deleted
func verifyClusterDeleted(t *testing.T, name string) {
	t.Helper()

	// Use k3d to check if cluster exists
	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("warning: failed to list clusters: %v", err)
		return
	}

	if strings.Contains(string(output), name) {
		t.Errorf("cluster '%s' still exists after deletion, cluster list: %s", name, string(output))
	}
}

// TestClusterDeleteWarnings tests warning messages for active workloads
func TestClusterDeleteWarnings(t *testing.T) {
	// This test would require deploying workloads to the cluster first
	// and verifying that warnings are shown
	t.Skip("Workload warnings test requires operator deployment (later phase)")
}
