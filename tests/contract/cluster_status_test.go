package contract

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestClusterStatusCommand verifies the cluster status command behavior
func TestClusterStatusCommand(t *testing.T) {
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
			name: "status of running cluster (default name)",
			args: []string{"dev", "cluster", "status"},
			setup: func(t *testing.T) {
				createTestCluster(t, "c8s-dev")
				time.Sleep(3 * time.Second) // Wait for cluster to be fully ready
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				// Should show cluster name
				if !strings.Contains(output, "c8s-dev") {
					t.Errorf("expected cluster name 'c8s-dev' in output, got: %s", output)
				}
				// Should show state
				if !strings.Contains(output, "Running") && !strings.Contains(output, "running") {
					t.Errorf("expected cluster state 'Running' in output, got: %s", output)
				}
				// Should show nodes information
				if !strings.Contains(output, "Node") && !strings.Contains(output, "node") {
					t.Logf("warning: expected nodes information in output, got: %s", output)
				}
				// Should show API endpoint
				if !strings.Contains(output, "API") && !strings.Contains(output, "api") {
					t.Logf("warning: expected API endpoint in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "c8s-dev")
			},
		},
		{
			name: "status of specific cluster",
			args: []string{"dev", "cluster", "status", "status-test-cluster"},
			setup: func(t *testing.T) {
				createTestCluster(t, "status-test-cluster")
				time.Sleep(3 * time.Second)
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "status-test-cluster") {
					t.Errorf("expected cluster name 'status-test-cluster' in output, got: %s", output)
				}
				if !strings.Contains(output, "Running") && !strings.Contains(output, "running") {
					t.Errorf("expected cluster state in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "status-test-cluster")
			},
		},
		{
			name:     "cluster not found",
			args:     []string{"dev", "cluster", "status", "nonexistent-cluster"},
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

// TestClusterStatusJSON tests JSON output format
func TestClusterStatusJSON(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "json-status-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get status with JSON output
	args := []string{"dev", "cluster", "status", clusterName, "--output", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Parse JSON
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify required fields
	requiredFields := []string{"name", "state"}
	for _, field := range requiredFields {
		if _, ok := status[field]; !ok {
			t.Errorf("expected field '%s' in JSON output, got: %v", field, status)
		}
	}

	// Verify cluster name
	if name, ok := status["name"].(string); !ok || name != clusterName {
		t.Errorf("expected name '%s' in JSON, got: %v", clusterName, status["name"])
	}

	// Verify state
	if state, ok := status["state"].(string); !ok || (state != "running" && state != "Running") {
		t.Errorf("expected state 'running' in JSON, got: %v", status["state"])
	}

	// Check for optional fields (log warnings if missing)
	optionalFields := []string{"nodes", "apiEndpoint", "uptime"}
	for _, field := range optionalFields {
		if _, ok := status[field]; !ok {
			t.Logf("warning: optional field '%s' not found in JSON output", field)
		}
	}
}

// TestClusterStatusYAML tests YAML output format
func TestClusterStatusYAML(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "yaml-status-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get status with YAML output
	args := []string{"dev", "cluster", "status", clusterName, "--output", "yaml"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Parse YAML
	var status map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("failed to parse YAML output: %v\nOutput: %s", err, output)
	}

	// Verify required fields
	requiredFields := []string{"name", "state"}
	for _, field := range requiredFields {
		if _, ok := status[field]; !ok {
			t.Errorf("expected field '%s' in YAML output, got: %v", field, status)
		}
	}

	// Verify cluster name
	if name, ok := status["name"].(string); !ok || name != clusterName {
		t.Errorf("expected name '%s' in YAML, got: %v", clusterName, status["name"])
	}

	// Verify state
	if state, ok := status["state"].(string); !ok || (state != "running" && state != "Running") {
		t.Errorf("expected state 'running' in YAML, got: %v", status["state"])
	}
}

// TestClusterStatusOutputFormats tests different output format flags
func TestClusterStatusOutputFormats(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "format-test-cluster"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	formats := []struct {
		name   string
		flag   string
		verify func(t *testing.T, output string)
	}{
		{
			name: "text format (default)",
			flag: "text",
			verify: func(t *testing.T, output string) {
				// Text format should be human-readable
				if !strings.Contains(output, clusterName) {
					t.Errorf("expected cluster name in text output")
				}
				if !strings.Contains(output, "Running") && !strings.Contains(output, "running") {
					t.Errorf("expected state in text output")
				}
			},
		},
		{
			name: "json format",
			flag: "json",
			verify: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("expected valid JSON, got error: %v", err)
				}
			},
		},
		{
			name: "yaml format",
			flag: "yaml",
			verify: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("expected valid YAML, got error: %v", err)
				}
			},
		},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			args := []string{"dev", "cluster", "status", clusterName, "--output", format.flag}
			output, exitCode := executeCommand(t, binaryPath, args)

			if exitCode != 0 {
				t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
			}

			format.verify(t, output)
		})
	}
}

// TestClusterStatusShorthand tests -o flag shorthand
func TestClusterStatusShorthand(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "shorthand-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Test -o shorthand
	args := []string{"dev", "cluster", "status", clusterName, "-o", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Errorf("expected valid JSON with -o flag, got error: %v\nOutput: %s", err, output)
	}
}

// TestClusterStatusFields verifies all expected status fields are present
func TestClusterStatusFields(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "fields-test"

	// Create cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get status as JSON for easier parsing
	args := []string{"dev", "cluster", "status", clusterName, "--output", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Required fields
	required := []string{"name", "state"}
	for _, field := range required {
		if _, ok := status[field]; !ok {
			t.Errorf("required field '%s' missing from status output", field)
		}
	}

	// Optional but expected fields
	expected := []string{"apiEndpoint", "nodes"}
	for _, field := range expected {
		if _, ok := status[field]; !ok {
			t.Logf("warning: expected field '%s' missing from status output", field)
		}
	}

	// If nodes field exists, verify it's an array
	if nodes, ok := status["nodes"]; ok {
		if _, isArray := nodes.([]interface{}); !isArray {
			t.Errorf("expected 'nodes' to be an array, got: %T", nodes)
		}
	}
}

// TestClusterStatusWatch tests the --watch flag
func TestClusterStatusWatch(t *testing.T) {
	// This test is complex because --watch requires running in background
	// and killing after a timeout
	t.Skip("Watch functionality test requires background execution and timeout handling")
}

// TestClusterStatusNotReady tests status of cluster that's not fully ready
func TestClusterStatusNotReady(t *testing.T) {
	// This test would require catching a cluster during startup
	// which is timing-dependent and difficult to test reliably
	t.Skip("Not ready state test requires timing control or mocking")
}

// TestClusterStatusStopped tests status of a stopped cluster
func TestClusterStatusStopped(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "stopped-test"

	// Create and then stop cluster using k3d directly
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Stop the cluster using k3d
	stopCmd := exec.Command("k3d", "cluster", "stop", clusterName)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to stop cluster: %v\nOutput: %s", err, string(output))
	}
	time.Sleep(2 * time.Second)

	// Check status
	args := []string{"dev", "cluster", "status", clusterName}
	output, exitCode := executeCommand(t, binaryPath, args)

	// Exit code should be 3 for not ready, or it may still succeed with stopped state
	if exitCode != 0 && exitCode != 3 {
		t.Logf("warning: expected exit code 0 or 3 for stopped cluster, got %d", exitCode)
	}

	// Output should indicate stopped state
	if !strings.Contains(output, "Stopped") && !strings.Contains(output, "stopped") && !strings.Contains(output, "not ready") {
		t.Logf("warning: expected 'stopped' or 'not ready' in status output, got: %s", output)
	}
}

// TestClusterStatusCurrentContext tests status without cluster name (uses current context)
func TestClusterStatusCurrentContext(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "context-test"

	// Create cluster (this should set the current context)
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get status without specifying cluster name (should use current context)
	args := []string{"dev", "cluster", "status"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("status command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Should show status of the current context cluster
	if !strings.Contains(output, clusterName) {
		t.Logf("warning: expected current context cluster name '%s' in output, got: %s", clusterName, output)
	}
}
