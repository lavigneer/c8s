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

// TestClusterListCommand verifies the cluster list command behavior
func TestClusterListCommand(t *testing.T) {
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
			name:     "list with no clusters",
			args:     []string{"dev", "cluster", "list"},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				// Should show empty list or "no clusters" message
				// This is acceptable as either format
				if strings.Contains(output, "error") || strings.Contains(output, "Error") {
					t.Errorf("should not show error for empty list, got: %s", output)
				}
			},
		},
		{
			name: "list with single cluster",
			args: []string{"dev", "cluster", "list"},
			setup: func(t *testing.T) {
				createTestCluster(t, "list-test-1")
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				if !strings.Contains(output, "list-test-1") {
					t.Errorf("expected cluster 'list-test-1' in output, got: %s", output)
				}
				// Should show state
				if !strings.Contains(output, "Running") && !strings.Contains(output, "running") {
					t.Logf("warning: expected cluster state in output, got: %s", output)
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "list-test-1")
			},
		},
		{
			name: "list with multiple clusters",
			args: []string{"dev", "cluster", "list"},
			setup: func(t *testing.T) {
				createTestCluster(t, "list-test-multi-1")
				createTestCluster(t, "list-test-multi-2")
				createTestCluster(t, "list-test-multi-3")
			},
			exitCode: 0,
			validate: func(t *testing.T, output string, exitCode int) {
				clusters := []string{"list-test-multi-1", "list-test-multi-2", "list-test-multi-3"}
				for _, cluster := range clusters {
					if !strings.Contains(output, cluster) {
						t.Errorf("expected cluster '%s' in output, got: %s", cluster, output)
					}
				}
			},
			cleanup: func(t *testing.T) {
				cleanupCluster(t, "list-test-multi-1")
				cleanupCluster(t, "list-test-multi-2")
				cleanupCluster(t, "list-test-multi-3")
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
				// Wait for clusters to be ready
				time.Sleep(3 * time.Second)
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

// TestClusterListJSON tests JSON output format
func TestClusterListJSON(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Create test clusters
	clusters := []string{"json-list-1", "json-list-2"}
	for _, cluster := range clusters {
		createTestCluster(t, cluster)
		defer cleanupCluster(t, cluster)
	}
	time.Sleep(3 * time.Second)

	// Get list with JSON output
	args := []string{"dev", "cluster", "list", "--output", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Should have a clusters array
	clustersData, ok := data["clusters"]
	if !ok {
		t.Fatalf("expected 'clusters' field in JSON output, got: %v", data)
	}

	clustersList, ok := clustersData.([]interface{})
	if !ok {
		t.Fatalf("expected 'clusters' to be an array, got: %T", clustersData)
	}

	// Should have at least our test clusters
	if len(clustersList) < len(clusters) {
		t.Errorf("expected at least %d clusters, got %d", len(clusters), len(clustersList))
	}

	// Verify each cluster has required fields
	for _, clusterData := range clustersList {
		cluster, ok := clusterData.(map[string]interface{})
		if !ok {
			t.Errorf("expected cluster to be an object, got: %T", clusterData)
			continue
		}

		// Check for name field
		if _, ok := cluster["name"]; !ok {
			t.Errorf("cluster missing 'name' field: %v", cluster)
		}

		// Check for state field
		if _, ok := cluster["state"]; !ok {
			t.Logf("warning: cluster missing 'state' field: %v", cluster)
		}
	}
}

// TestClusterListYAML tests YAML output format
func TestClusterListYAML(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Create test cluster
	clusterName := "yaml-list-test"
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get list with YAML output
	args := []string{"dev", "cluster", "list", "--output", "yaml"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Parse YAML
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to parse YAML output: %v\nOutput: %s", err, output)
	}

	// Should have a clusters array
	clustersData, ok := data["clusters"]
	if !ok {
		t.Fatalf("expected 'clusters' field in YAML output, got: %v", data)
	}

	clustersList, ok := clustersData.([]interface{})
	if !ok {
		t.Fatalf("expected 'clusters' to be an array, got: %T", clustersData)
	}

	// Should have at least our test cluster
	if len(clustersList) < 1 {
		t.Errorf("expected at least 1 cluster, got %d", len(clustersList))
	}
}

// TestClusterListAll tests --all flag to show all k3d clusters
func TestClusterListAll(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Create c8s cluster
	c8sCluster := "c8s-list-all-test"
	createTestCluster(t, c8sCluster)
	defer cleanupCluster(t, c8sCluster)

	// Create non-c8s cluster (without c8s prefix)
	nonC8sCluster := "other-cluster-test"
	createTestCluster(t, nonC8sCluster)
	defer cleanupCluster(t, nonC8sCluster)

	time.Sleep(3 * time.Second)

	// List without --all (should only show c8s clusters)
	args := []string{"dev", "cluster", "list"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Should show c8s cluster
	if !strings.Contains(output, c8sCluster) {
		t.Logf("warning: expected c8s cluster '%s' in default list, got: %s", c8sCluster, output)
	}

	// List with --all (should show all clusters)
	argsAll := []string{"dev", "cluster", "list", "--all"}
	outputAll, exitCodeAll := executeCommand(t, binaryPath, argsAll)

	if exitCodeAll != 0 {
		t.Fatalf("list --all command failed with exit code %d\nOutput: %s", exitCodeAll, outputAll)
	}

	// Should show both clusters
	if !strings.Contains(outputAll, c8sCluster) {
		t.Errorf("expected c8s cluster '%s' in --all list, got: %s", c8sCluster, outputAll)
	}
	if !strings.Contains(outputAll, nonC8sCluster) {
		t.Errorf("expected other cluster '%s' in --all list, got: %s", nonC8sCluster, outputAll)
	}
}

// TestClusterListEmpty tests listing when no clusters exist
func TestClusterListEmpty(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Ensure no test clusters exist
	// This test assumes a clean environment or proper cleanup from other tests

	args := []string{"dev", "cluster", "list"}
	output, exitCode := executeCommand(t, binaryPath, args)

	// Should succeed with exit code 0 even with no clusters
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for empty list, got %d\nOutput: %s", exitCode, output)
	}

	// Output should indicate no clusters or show empty list
	// Accept various formats: "No clusters", empty table, etc.
	if strings.Contains(output, "error") || strings.Contains(output, "Error") {
		t.Errorf("should not show error for empty list, got: %s", output)
	}
}

// TestClusterListOutputFormats tests different output format flags
func TestClusterListOutputFormats(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "format-list-test"

	// Create test cluster
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
				// Text format should be human-readable table
				if !strings.Contains(output, clusterName) {
					t.Errorf("expected cluster name in text output")
				}
				// Should have header or column labels
				if !strings.Contains(output, "NAME") && !strings.Contains(output, "Name") {
					t.Logf("warning: expected table header in text output, got: %s", output)
				}
			},
		},
		{
			name: "json format",
			flag: "json",
			verify: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("expected valid JSON, got error: %v\nOutput: %s", err, output)
				}
				// Should have clusters array
				if _, ok := data["clusters"]; !ok {
					t.Errorf("expected 'clusters' field in JSON output")
				}
			},
		},
		{
			name: "yaml format",
			flag: "yaml",
			verify: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("expected valid YAML, got error: %v\nOutput: %s", err, output)
				}
				// Should have clusters array
				if _, ok := data["clusters"]; !ok {
					t.Errorf("expected 'clusters' field in YAML output")
				}
			},
		},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			args := []string{"dev", "cluster", "list", "--output", format.flag}
			output, exitCode := executeCommand(t, binaryPath, args)

			if exitCode != 0 {
				t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
			}

			format.verify(t, output)
		})
	}
}

// TestClusterListShorthand tests -o flag shorthand
func TestClusterListShorthand(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "shorthand-list-test"

	// Create test cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Test -o shorthand
	args := []string{"dev", "cluster", "list", "-o", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Errorf("expected valid JSON with -o flag, got error: %v\nOutput: %s", err, output)
	}
}

// TestClusterListFields verifies all expected list fields are present
func TestClusterListFields(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "fields-list-test"

	// Create test cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get list as JSON for easier parsing
	args := []string{"dev", "cluster", "list", "--output", "json"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	clustersData, ok := data["clusters"].([]interface{})
	if !ok {
		t.Fatalf("expected 'clusters' array in output")
	}

	// Find our test cluster
	var testCluster map[string]interface{}
	for _, clusterData := range clustersData {
		cluster, ok := clusterData.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := cluster["name"].(string); ok && name == clusterName {
			testCluster = cluster
			break
		}
	}

	if testCluster == nil {
		t.Fatalf("test cluster '%s' not found in list output", clusterName)
	}

	// Required fields
	required := []string{"name", "state"}
	for _, field := range required {
		if _, ok := testCluster[field]; !ok {
			t.Errorf("required field '%s' missing from cluster in list output", field)
		}
	}

	// Optional but expected fields
	expected := []string{"nodeCount", "version", "uptime"}
	for _, field := range expected {
		if _, ok := testCluster[field]; !ok {
			t.Logf("warning: expected field '%s' missing from cluster in list output", field)
		}
	}
}

// TestClusterListTextFormat verifies text/table output formatting
func TestClusterListTextFormat(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	clusterName := "text-format-test"

	// Create test cluster
	createTestCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)
	time.Sleep(3 * time.Second)

	// Get list with default (text) output
	args := []string{"dev", "cluster", "list"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Should have table-like structure with headers
	expectedHeaders := []string{"NAME", "STATE", "NODES"}
	headerCount := 0
	for _, header := range expectedHeaders {
		if strings.Contains(output, header) || strings.Contains(output, strings.ToLower(header)) {
			headerCount++
		}
	}

	if headerCount < 2 {
		t.Logf("warning: expected at least 2 column headers in text output, found %d\nOutput: %s", headerCount, output)
	}

	// Should show cluster data
	if !strings.Contains(output, clusterName) {
		t.Errorf("expected cluster name in text output, got: %s", output)
	}
}

// TestClusterListMixedStates tests listing clusters in different states
func TestClusterListMixedStates(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	runningCluster := "mixed-running"
	stoppedCluster := "mixed-stopped"

	// Create two clusters
	createTestCluster(t, runningCluster)
	defer cleanupCluster(t, runningCluster)

	createTestCluster(t, stoppedCluster)
	defer cleanupCluster(t, stoppedCluster)

	time.Sleep(3 * time.Second)

	// Stop one cluster
	stopCmd := exec.Command("k3d", "cluster", "stop", stoppedCluster)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to stop cluster: %v\nOutput: %s", err, string(output))
	}
	time.Sleep(2 * time.Second)

	// List all clusters
	args := []string{"dev", "cluster", "list"}
	output, exitCode := executeCommand(t, binaryPath, args)

	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	// Should show both clusters
	if !strings.Contains(output, runningCluster) {
		t.Errorf("expected running cluster '%s' in output, got: %s", runningCluster, output)
	}
	if !strings.Contains(output, stoppedCluster) {
		t.Errorf("expected stopped cluster '%s' in output, got: %s", stoppedCluster, output)
	}

	// Should show different states (though exact wording may vary)
	if !strings.Contains(output, "running") && !strings.Contains(output, "Running") {
		t.Logf("warning: expected 'running' state in output, got: %s", output)
	}
	if !strings.Contains(output, "stopped") && !strings.Contains(output, "Stopped") {
		t.Logf("warning: expected 'stopped' state in output, got: %s", output)
	}
}
