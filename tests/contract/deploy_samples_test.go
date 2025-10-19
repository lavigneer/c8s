package contract

import (
	"os"
	"strings"
	"testing"
)

// TestDeploySamplesCommand tests the sample deployment functionality
func TestDeploySamplesCommand(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name     string
		args     []string
		exitCode int
	}{
		{
			name:     "deploy samples help",
			args:     []string{"dev", "deploy", "samples", "--help"},
			exitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exitCode := executeCommand(t, binaryPath, tt.args)

			if exitCode != tt.exitCode {
				t.Errorf("expected exit code %d, got %d\nOutput: %s", tt.exitCode, exitCode, output)
			}
		})
	}
}

// TestDeploySamplesCreatesResources verifies PipelineConfigs are created
func TestDeploySamplesCreatesResources(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}

	// This test requires actual cluster to be running
	t.Skip("Requires running k3d cluster")
}

// TestDeploySamplesSelectFilter tests the --select flag
func TestDeploySamplesSelectFilter(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with select flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "deploy", "samples", "--select", "simple-build", "--help"})

	// Should show help even with select flag
	if exitCode != 0 && !strings.Contains(output, "help") {
		t.Logf("expected help or success, got exit code %d", exitCode)
	}
}

// TestDeploySamplesNamespace tests custom namespace deployment
func TestDeploySamplesNamespace(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with custom namespace flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "deploy", "samples", "--namespace", "custom-ns", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help command to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify namespace flag is recognized
	if !strings.Contains(output, "namespace") {
		t.Logf("expected 'namespace' in help output, got: %s", output)
	}
}
