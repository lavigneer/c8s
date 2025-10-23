package samples

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SampleDeploymentStatus contains information about sample deployment
type SampleDeploymentStatus struct {
	Success         bool
	SamplesDeployed []string
	Namespace       string
	Message         string
	Timestamp       time.Time
}

// DeploySamples deploys sample PipelineConfigs to the cluster
func DeploySamples(
	namespace string,
	samplesPath string,
	selectFilter string,
) (*SampleDeploymentStatus, error) {
	status := &SampleDeploymentStatus{
		Namespace: namespace,
		Timestamp: time.Now(),
	}

	if namespace == "" {
		namespace = "default"
		status.Namespace = namespace
	}

	// Create namespace if it doesn't exist
	createNSCmd := exec.Command("kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	createNSPipeCmd := exec.Command("kubectl", "apply", "-f", "-")
	pipe, err := createNSCmd.StdoutPipe()
	if err != nil {
		return status, fmt.Errorf("failed to create pipe: %w", err)
	}
	createNSPipeCmd.Stdin = pipe
	if err := createNSCmd.Start(); err != nil {
		return status, fmt.Errorf("failed to start namespace creation: %w", err)
	}
	if err := createNSPipeCmd.Run(); err != nil {
		// Ignore error if namespace already exists
		_ = createNSCmd.Wait()
	} else {
		_ = createNSCmd.Wait()
	}

	// Resolve samples path
	if samplesPath == "" {
		samplesPath = "config/samples"
	}

	// Check if path is relative and convert to absolute
	if !filepath.IsAbs(samplesPath) {
		cwd, err := os.Getwd()
		if err != nil {
			status.Message = fmt.Sprintf("Failed to get working directory: %v", err)
			return status, err
		}
		samplesPath = filepath.Join(cwd, samplesPath)
	}

	// Verify path exists
	if _, err := os.Stat(samplesPath); err != nil {
		status.Message = fmt.Sprintf("Samples path not found: %s", samplesPath)
		return status, err
	}

	// Parse comma-separated select filters
	var selectFilters []string
	if selectFilter != "" {
		selectFilters = strings.Split(selectFilter, ",")
		for i := range selectFilters {
			selectFilters[i] = strings.TrimSpace(selectFilters[i])
		}
	}

	// Find all YAML files in the directory
	var sampleFiles []string
	walkErr := filepath.Walk(samplesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			filename := filepath.Base(path)

			// Apply select filter if specified
			if len(selectFilters) > 0 {
				matches := false
				for _, filter := range selectFilters {
					if strings.Contains(filename, filter) {
						matches = true
						break
					}
				}
				if !matches {
					return nil
				}
			}

			sampleFiles = append(sampleFiles, path)
		}
		return nil
	})

	if walkErr != nil {
		status.Message = fmt.Sprintf("Failed to scan samples directory: %v", walkErr)
		return status, walkErr
	}

	if len(sampleFiles) == 0 {
		status.Message = fmt.Sprintf("No sample files found in %s", samplesPath)
		if selectFilter != "" {
			status.Message = fmt.Sprintf("No sample files matching filter %q found in %s", selectFilter, samplesPath)
		}
		return status, nil // Not an error - just no samples to deploy
	}

	// Validate YAML manifests before applying
	for _, sampleFile := range sampleFiles {
		if err := validateManifest(sampleFile); err != nil {
			status.Message = fmt.Sprintf("Invalid YAML in %s: %v", sampleFile, err)
			return status, err
		}
	}

	// Apply each sample file
	for _, sampleFile := range sampleFiles {
		err := applySampleManifest(sampleFile, namespace)
		if err != nil {
			status.Message = fmt.Sprintf("Failed to deploy sample from %s: %v", sampleFile, err)
			return status, err
		}
		status.SamplesDeployed = append(status.SamplesDeployed, filepath.Base(sampleFile))
	}

	status.Success = true
	status.Message = fmt.Sprintf("Successfully deployed %d sample(s) to namespace %s", len(status.SamplesDeployed), namespace)
	return status, nil
}

// applySampleManifest applies a sample manifest to the cluster
func applySampleManifest(manifestPath string, namespace string) error {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	manifestStr := string(content)

	// Ensure namespace is set in the manifest
	lines := strings.Split(manifestStr, "\n")
	var result []string
	hasNamespace := false

	for _, line := range lines {
		if strings.Contains(line, "namespace:") {
			result = append(result, fmt.Sprintf("  namespace: %s", namespace))
			hasNamespace = true
		} else if strings.Contains(line, "kind:") && strings.Contains(line, "PipelineConfig") {
			result = append(result, line)
			// Add namespace after kind if not already present
			if !hasNamespace {
				result = append(result, fmt.Sprintf("  namespace: %s", namespace))
				hasNamespace = true
			}
		} else {
			result = append(result, line)
		}
	}

	manifestStr = strings.Join(result, "\n")

	// Apply using kubectl
	cmd := exec.Command("kubectl", "apply", "-n", namespace, "-f", "-")
	cmd.Stdin = strings.NewReader(manifestStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// validateManifest checks if a YAML manifest is valid
func validateManifest(manifestPath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", manifestPath, "--dry-run=client")
	_, err := cmd.CombinedOutput()
	return err
}

// ListDeployedSamples returns the list of deployed sample PipelineConfigs
func ListDeployedSamples(namespace string) ([]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelineconfigs", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(output))
	return names, nil
}
