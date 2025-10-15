package contract

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestClusterStartStopCommands verifies the cluster start and stop command behavior
func TestClusterStartStopCommands(t *testing.T) {
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
			name: "stop running cluster (default name)",
			args: []string{"dev", "cluster", "stop"},
			setup: func(t *testing.T) {
				createTestCluster(t, "c8s-dev")
				time.Sleep(3 * time.Second)
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "stopped") && !strings.Contains(output, "✓") {
					t.Errorf("expected stop success message, got: %s", output)
				}
				if !strings.Contains(output, "c8s-dev") {
					t.Errorf("expected cluster name 'c8s-dev' in output, got: %s", output)
				}

				// Verify cluster is actually stopped
				time.Sleep(2 * time.Second)
				verifyClusterStopped(t, "c8s-dev")
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "c8s-dev")
			},
		},
		{
			name: "stop specific cluster",
			args: []string{"dev", "cluster", "stop", "stop-test-cluster"},
			setup: func(t *testing.T) {
				createTestCluster(t, "stop-test-cluster")
				time.Sleep(3 * time.Second)
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "stop-test-cluster") {
					t.Errorf("expected cluster name 'stop-test-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "stopped") && !strings.Contains(output, "✓") {
					t.Errorf("expected stop success message, got: %s", output)
				}

				// Verify cluster is stopped
				time.Sleep(2 * time.Second)
				verifyClusterStopped(t, "stop-test-cluster")
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "stop-test-cluster")
			},
		},
		{
			name:     "stop nonexistent cluster",
			args:     []string{"dev", "cluster", "stop", "nonexistent-cluster"},
			exitCode: 2,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "not found") && !strings.Contains(output, "does not exist") {
					t.Errorf("expected 'not found' error message, got: %s", output)
				}
			},
		},
		{
			name: "start stopped cluster (default name)",
			args: []string{"dev", "cluster", "start"},
			setup: func(t *testing.T) {
				createTestCluster(t, "c8s-dev")
				time.Sleep(3 * time.Second)
				// Stop it first
				stopCmd := exec.Command("k3d", "cluster", "stop", "c8s-dev")
				if output, err := stopCmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to stop cluster in setup: %v\nOutput: %s", err, string(output))
				}
				time.Sleep(2 * time.Second)
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "started") && !strings.Contains(output, "✓") {
					t.Errorf("expected start success message, got: %s", output)
				}
				if !strings.Contains(output, "c8s-dev") {
					t.Errorf("expected cluster name 'c8s-dev' in output, got: %s", output)
				}

				// Verify cluster is running
				time.Sleep(3 * time.Second)
				verifyClusterRunning(t, "c8s-dev")
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "c8s-dev")
			},
		},
		{
			name: "start specific stopped cluster",
			args: []string{"dev", "cluster", "start", "start-test-cluster"},
			setup: func(t *testing.T) {
				createTestCluster(t, "start-test-cluster")
				time.Sleep(3 * time.Second)
				// Stop it first
				stopCmd := exec.Command("k3d", "cluster", "stop", "start-test-cluster")
				if output, err := stopCmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to stop cluster in setup: %v\nOutput: %s", err, string(output))
				}
				time.Sleep(2 * time.Second)
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "start-test-cluster") {
					t.Errorf("expected cluster name 'start-test-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "started") && !strings.Contains(output, "✓") {
					t.Errorf("expected start success message, got: %s", output)
				}

				// Verify cluster is running
				time.Sleep(3 * time.Second)
				verifyClusterRunning(t, "start-test-cluster")
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "start-test-cluster")
			},
		},
		{
			name:     "start nonexistent cluster",
			args:     []string{"dev", "cluster", "start", "nonexistent-cluster"},
			exitCode: 2,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "not found") && !strings.Contains(output, "does not exist") {
					t.Errorf("expected 'not found' error message, got: %s", output)
				}
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

// TestClusterStartWait tests the --wait flag for start command
func TestClusterStartWait(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "wait-test-cluster"

	// Create and stop cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	stopCmd := exec.Command("k3d", "cluster", "stop", clusterName)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to stop cluster: %v\nOutput: %s", err, string(output))
	}
	time.Sleep(2 * time.Second)

	// Start with --wait flag
	args := []string{"dev", "cluster", "start", clusterName, "--wait"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("start command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Should wait for cluster to be ready before returning
	verifyClusterRunning(t, clusterName)
}

// TestClusterStartTimeout tests the --timeout flag
func TestClusterStartTimeout(t *testing.T) {
	// This test is difficult to implement without mocking
	// because we'd need a cluster that never becomes ready
	t.Skip("Timeout test requires mocking or simulated failure")
}

// TestClusterStopNoWait tests stopping without waiting
func TestClusterStopNoWait(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "no-wait-stop-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Stop cluster (default behavior may or may not wait)
	args := []string{"dev", "cluster", "stop", clusterName}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("stop command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Give it a moment to stop
	time.Sleep(2 * time.Second)

	// Verify cluster is stopped
	verifyClusterStopped(t, clusterName)
}

// TestClusterLifecycleFull tests complete start/stop lifecycle
func TestClusterLifecycleFull(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "lifecycle-full-test"

	// 1. Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Verify it's running
	verifyClusterRunning(t, clusterName)

	// 2. Stop cluster
	stopArgs := []string{"dev", "cluster", "stop", clusterName}
	stopOutput, stopExitCode := executeCommand(t, binaryPath, stopArgs)

	if stopExitCode != 0 {
		t.Fatalf("stop command failed with exit code %d\nOutput: %s", stopExitCode, stopOutput)
	}

	time.Sleep(2 * time.Second)
	verifyClusterStopped(t, clusterName)

	// 3. Start cluster again
	startArgs := []string{"dev", "cluster", "start", clusterName}
	startOutput, startExitCode := executeCommand(t, binaryPath, startArgs)

	if startExitCode != 0 {
		t.Fatalf("start command failed with exit code %d\nOutput: %s", startExitCode, startOutput)
	}

	time.Sleep(3 * time.Second)
	verifyClusterRunning(t, clusterName)

	// 4. Stop again
	stopArgs2 := []string{"dev", "cluster", "stop", clusterName}
	stopOutput2, stopExitCode2 := executeCommand(t, binaryPath, stopArgs2)

	if stopExitCode2 != 0 {
		t.Fatalf("second stop command failed with exit code %d\nOutput: %s", stopExitCode2, stopOutput2)
	}

	time.Sleep(2 * time.Second)
	verifyClusterStopped(t, clusterName)
}

// TestClusterStatePersistence tests that cluster state persists across stop/start
func TestClusterStatePersistence(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "persistence-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(5 * time.Second)

	// Deploy a test workload (simple pod)
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl not available, skipping persistence test")
	}

	// Switch to cluster context
	contextCmd := exec.Command(kubectlPath, "config", "use-context", "k3d-"+clusterName)
	if output, err := contextCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to switch context: %v\nOutput: %s", err, string(output))
	}

	// Create a test pod
	createPodCmd := exec.Command(kubectlPath, "run", "test-pod", "--image=nginx:alpine", "--restart=Never")
	if output, err := createPodCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test pod: %v\nOutput: %s", err, string(output))
	}

	// Wait for pod to be created
	time.Sleep(3 * time.Second)

	// Stop cluster
	stopArgs := []string{"dev", "cluster", "stop", clusterName}
	_, stopExitCode := executeCommand(t, binaryPath, stopArgs)
	if stopExitCode != 0 {
		t.Fatalf("stop command failed")
	}

	time.Sleep(2 * time.Second)

	// Start cluster
	startArgs := []string{"dev", "cluster", "start", clusterName}
	_, startExitCode := executeCommand(t, binaryPath, startArgs)
	if startExitCode != 0 {
		t.Fatalf("start command failed")
	}

	time.Sleep(5 * time.Second)

	// Verify pod still exists
	getPodCmd := exec.Command(kubectlPath, "get", "pod", "test-pod", "--context", "k3d-"+clusterName)
	output, err := getPodCmd.CombinedOutput()
	if err != nil {
		t.Errorf("pod should still exist after restart, got error: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "test-pod") {
		t.Errorf("expected test-pod to exist after restart, got: %s", string(output))
	}

	// Cleanup pod
	deletePodCmd := exec.Command(kubectlPath, "delete", "pod", "test-pod", "--context", "k3d-"+clusterName, "--force", "--grace-period=0")
	deletePodCmd.Run()
}

// TestClusterStartAlreadyRunning tests starting an already running cluster
func TestClusterStartAlreadyRunning(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "already-running-test"

	// Create cluster (will be running)
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Try to start already running cluster
	args := []string{"dev", "cluster", "start", clusterName}
	output, exitCode := executeCommand(t, binaryPath, args)

	// Should either succeed (idempotent) or report already running
	if exitCode != 0 {
		t.Logf("warning: starting already running cluster returned exit code %d, output: %s", exitCode, output)
	}

	// Cluster should still be running
	verifyClusterRunning(t, clusterName)
}

// TestClusterStopAlreadyStopped tests stopping an already stopped cluster
func TestClusterStopAlreadyStopped(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "already-stopped-test"

	// Create and stop cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	stopCmd := exec.Command("k3d", "cluster", "stop", clusterName)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to stop cluster: %v\nOutput: %s", err, string(output))
	}
	time.Sleep(2 * time.Second)

	// Try to stop already stopped cluster
	args := []string{"dev", "cluster", "stop", clusterName}
	output, exitCode := executeCommand(t, binaryPath, args)

	// Should either succeed (idempotent) or report already stopped
	if exitCode != 0 {
		t.Logf("warning: stopping already stopped cluster returned exit code %d, output: %s", exitCode, output)
	}
}

// Helper functions

// verifyClusterRunning checks that a cluster is in running state
func verifyClusterRunning(t *testing.T, name string) {
	t.Helper()

	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list clusters: %v\nOutput: %s", err, string(output))
	}

	// Parse output to find cluster and check state
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, name) {
			// Check if line indicates running state
			if !strings.Contains(line, "running") && !strings.Contains(line, "Running") {
				t.Errorf("cluster '%s' is not in running state: %s", name, line)
			}
			return
		}
	}

	t.Errorf("cluster '%s' not found in cluster list", name)
}

// verifyClusterStopped checks that a cluster is in stopped state
func verifyClusterStopped(t *testing.T, name string) {
	t.Helper()

	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list clusters: %v\nOutput: %s", err, string(output))
	}

	// Parse output to find cluster and check state
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, name) {
			// Check if line indicates stopped state
			if !strings.Contains(line, "stopped") && !strings.Contains(line, "Stopped") {
				t.Errorf("cluster '%s' is not in stopped state: %s", name, line)
			}
			return
		}
	}

	t.Errorf("cluster '%s' not found in cluster list", name)
}

// TestClusterStartStopGlobalFlags tests global flags with start/stop commands
func TestClusterStartStopGlobalFlags(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "global-flags-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "stop with verbose flag",
			args: []string{"dev", "cluster", "stop", clusterName, "--verbose"},
		},
		{
			name: "start with verbose flag",
			args: []string{"dev", "cluster", "start", clusterName, "-v"},
		},
		{
			name: "stop with no-color flag",
			args: []string{"dev", "cluster", "stop", clusterName, "--no-color"},
		},
		{
			name: "start with no-color flag",
			args: []string{"dev", "cluster", "start", clusterName, "--no-color"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exitCode := executeCommand(t, binaryPath, tt.args)

			if exitCode != 0 {
				t.Errorf("command failed with exit code %d\nOutput: %s", exitCode, output)
			}

			// For no-color, verify no ANSI escape codes
			if strings.Contains(tt.name, "no-color") {
				if strings.Contains(output, "\x1b[") {
					t.Errorf("expected no ANSI escape codes with --no-color, got: %s", output)
				}
			}
		})
	}
}
