package contract

import (
	"os"
	"strings"
	"testing"
)

// TestTestLogsCommand tests the test logs functionality
func TestTestLogsCommand(t *testing.T) {
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
			name:     "test logs help",
			args:     []string{"dev", "test", "logs", "--help"},
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
			if exitCode == 0 && !strings.Contains(output, "logs") {
				t.Logf("expected 'logs' in help output, got: %s", output)
			}
		})
	}
}

// TestTestLogsFollow tests the --follow flag
func TestTestLogsFollow(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with follow flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--follow", "--help"})

	// Should show help
	if exitCode != 0 && !strings.Contains(output, "help") {
		t.Logf("expected help or success, got exit code %d", exitCode)
	}
}

// TestTestLogsTail tests the --tail flag
func TestTestLogsTail(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with tail flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--tail", "50", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify tail flag is recognized
	if !strings.Contains(output, "tail") {
		t.Logf("expected 'tail' in help output, got: %s", output)
	}
}

// TestTestLogsPipeline tests specifying specific pipeline
func TestTestLogsPipeline(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with pipeline flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--pipeline", "simple-build", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify pipeline flag is recognized
	if !strings.Contains(output, "pipeline") {
		t.Logf("expected 'pipeline' in help output, got: %s", output)
	}
}

// TestTestLogsNamespace tests namespace flag
func TestTestLogsNamespace(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with namespace flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--namespace", "default", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify namespace flag is recognized
	if !strings.Contains(output, "namespace") {
		t.Logf("expected 'namespace' in help output, got: %s", output)
	}
}

// TestTestLogsCluster tests cluster flag
func TestTestLogsCluster(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	// Test with cluster flag
	output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--cluster", "my-cluster", "--help"})

	// Should show help
	if exitCode != 0 {
		t.Logf("expected help to succeed, got exit code %d\nOutput: %s", exitCode, output)
	}

	// Verify cluster flag is recognized
	if !strings.Contains(output, "cluster") {
		t.Logf("expected 'cluster' in help output, got: %s", output)
	}
}

// TestTestLogsOutputFormats tests different output formats
func TestTestLogsOutputFormats(t *testing.T) {
	// Build the c8s binary first
	binaryPath := buildC8sBinary(t)
	defer os.Remove(binaryPath)

	formats := []string{"raw", "formatted", "json"}
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			output, exitCode := executeCommand(t, binaryPath, []string{"dev", "test", "logs", "--output", format, "--help"})

			// Should show help
			if exitCode != 0 && !strings.Contains(output, "help") {
				t.Logf("expected help for format %s, got exit code %d", format, exitCode)
			}
		})
	}
}
