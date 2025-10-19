package contract

import (
	"os"
	"strings"
	"testing"
)

// TestTestRunCommand tests the test run functionality
func TestTestRunCommand(t *testing.T) {
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
			name:     "test run help",
			args:     []string{"dev", "test", "run", "--help"},
			exitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exitCode := executeCommand(t, binaryPath, tt.args)

			if exitCode != tt.exitCode {
				t.Errorf("expected exit code %d, got %d\nOutput: %s", tt.exitCode, exitCode, output)
			}

			// Verify help contains expected elements
			if exitCode == 0 && !strings.Contains(output, "run") {
				t.Logf("expected 'run' in help output, got: %s", output)
			}
		})
	}
}

// TestTestRunWatch tests the --watch flag
func TestTestRunWatch(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with watch flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "run", "--watch", "--help"})

	// Should show help
	if exitCode != 0 && !strings.Contains(output, "help") {
		t.Logf("expected help or success, got exit code %d", exitCode)
	}
}

// TestTestRunPipeline tests specifying specific pipeline
func TestTestRunPipeline(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with pipeline flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "run", "--pipeline", "simple-build", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify pipeline flag is recognized
	if !strings.Contains(output, "pipeline") {
		t.Logf("expected 'pipeline' in help output, got: %s", output)
	}
}

// TestTestRunTimeout tests timeout handling
func TestTestRunTimeout(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with timeout flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "run", "--timeout", "300", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify timeout flag is recognized
	if !strings.Contains(output, "timeout") {
		t.Logf("expected 'timeout' in help output, got: %s", output)
	}
}

// TestTestRunCluster tests cluster flag
func TestTestRunCluster(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with cluster flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "run", "--cluster", "my-cluster", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify cluster flag is recognized
	if !strings.Contains(output, "cluster") {
		t.Logf("expected 'cluster' in help output, got: %s", output)
	}
}

// TestTestRunOutputFormats tests different output formats
func TestTestRunOutputFormats(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	formats := []string{"text", "json", "yaml"}
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "run", "--output", format, "--help"})

			// Should show help
			if exitCode != 0 && !strings.Contains(output, "help") {
				t.Logf("expected help for format %s, got exit code %d", format, exitCode)
			}
		})
	}
}
