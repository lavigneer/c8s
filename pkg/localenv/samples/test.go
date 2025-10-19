package samples

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PipelineTestResult contains the result of running a single pipeline test
type PipelineTestResult struct {
	Name       string
	Namespace  string
	Status     string // Running, Success, Failed, Timeout
	Duration   time.Duration
	StartTime  time.Time
	EndTime    time.Time
	ErrorMsg   string
	LogSummary string
}

// PipelineTestSummary contains aggregated test results
type PipelineTestSummary struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	TimeoutTests int
	Duration     time.Duration
	Results      []PipelineTestResult
	Message      string
}

// RunPipelineTests executes pipeline tests
func RunPipelineTests(ctx context.Context, namespace string, pipelineFilter string, timeout time.Duration) (*PipelineTestSummary, error) {
	summary := &PipelineTestSummary{
		Duration: 0,
	}

	if namespace == "" {
		namespace = "default"
	}

	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	startTime := time.Now()

	// List available PipelineConfigs
	configs, err := ListPipelineConfigs(namespace, pipelineFilter)
	if err != nil {
		summary.Message = fmt.Sprintf("Failed to list PipelineConfigs: %v", err)
		return summary, err
	}

	if len(configs) == 0 {
		summary.Message = "No PipelineConfigs found to test"
		return summary, nil
	}

	summary.TotalTests = len(configs)

	// Create PipelineRun for each config
	for _, config := range configs {
		result := PipelineTestResult{
			Name:      config,
			Namespace: namespace,
			StartTime: time.Now(),
		}

		// Create PipelineRun resource
		runName := fmt.Sprintf("%s-run-%d", config, time.Now().Unix())
		err := createPipelineRun(namespace, config, runName)
		if err != nil {
			result.Status = "Failed"
			result.ErrorMsg = fmt.Sprintf("Failed to create PipelineRun: %v", err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			summary.FailedTests++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Wait for completion with timeout
		status, err := waitForPipelineCompletion(namespace, runName, timeout)
		if err != nil {
			result.Status = "Timeout"
			result.ErrorMsg = fmt.Sprintf("Pipeline test timed out: %v", err)
			summary.TimeoutTests++
		} else if status == "Succeeded" {
			result.Status = "Success"
			summary.PassedTests++
		} else {
			result.Status = "Failed"
			result.ErrorMsg = fmt.Sprintf("Pipeline status: %s", status)
			summary.FailedTests++
		}

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		summary.Results = append(summary.Results, result)

		// Check for context cancellation
		select {
		case <-ctx.Done():
			summary.Message = "Test execution cancelled"
			return summary, ctx.Err()
		default:
		}
	}

	summary.Duration = time.Since(startTime)
	summary.Message = fmt.Sprintf("Tests completed: %d passed, %d failed, %d timeout", summary.PassedTests, summary.FailedTests, summary.TimeoutTests)
	return summary, nil
}

// ListPipelineConfigs lists available PipelineConfig resources
func ListPipelineConfigs(namespace string, filter string) ([]string, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelineconfigs", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list PipelineConfigs: %v\nOutput: %s", err, output)
	}

	names := strings.Fields(string(output))

	// Apply filter if specified
	if filter != "" {
		var filtered []string
		for _, name := range names {
			if strings.Contains(name, filter) {
				filtered = append(filtered, name)
			}
		}
		return filtered, nil
	}

	return names, nil
}

// createPipelineRun creates a PipelineRun resource for a given PipelineConfig
func createPipelineRun(namespace string, configName string, runName string) error {
	// Create PipelineRun manifest
	manifest := fmt.Sprintf(`apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: %s
  namespace: %s
spec:
  pipelineConfigRef:
    name: %s
  timeout: 10m
`, runName, namespace, configName)

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create PipelineRun: %v\nOutput: %s", err, output)
	}

	return nil
}

// waitForPipelineCompletion waits for a pipeline run to complete
func waitForPipelineCompletion(namespace string, runName string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	checkInterval := 2 * time.Second

	for {
		// Check pipeline status
		cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelinerun", runName, "-o", "jsonpath={.status.phase}")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Resource might not exist yet
			time.Sleep(checkInterval)
			if time.Since(startTime) > timeout {
				return "", fmt.Errorf("timeout waiting for PipelineRun to complete")
			}
			continue
		}

		status := strings.TrimSpace(string(output))
		if status == "" {
			status = "Unknown"
		}

		// Check if completed
		if status == "Succeeded" || status == "Failed" || status == "Error" {
			return status, nil
		}

		// Check timeout
		if time.Since(startTime) > timeout {
			return "", fmt.Errorf("timeout waiting for PipelineRun (current status: %s)", status)
		}

		time.Sleep(checkInterval)
	}
}

// GetPipelineRunLogs fetches logs for a pipeline run
func GetPipelineRunLogs(namespace string, runName string, tailLines int) (string, error) {
	// Find pods associated with the PipelineRun
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l",
		fmt.Sprintf("pipelinerun=%s", runName), "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	podNames := strings.Fields(string(output))
	if len(podNames) == 0 {
		return "No pods found for pipeline run", nil
	}

	var logs strings.Builder
	for _, podName := range podNames {
		// Get logs from pod
		logsCmd := exec.Command("kubectl", "-n", namespace, "logs", podName)
		if tailLines > 0 {
			logsCmd = exec.Command("kubectl", "-n", namespace, "logs", podName, fmt.Sprintf("--tail=%d", tailLines))
		}

		logOutput, err := logsCmd.CombinedOutput()
		if err != nil {
			logs.WriteString(fmt.Sprintf("Error getting logs from %s: %v\n", podName, err))
		} else {
			logs.WriteString(fmt.Sprintf("=== Logs from %s ===\n", podName))
			logs.Write(logOutput)
			logs.WriteString("\n")
		}
	}

	return logs.String(), nil
}

// GetPipelineRunStatus retrieves the status of a pipeline run
func GetPipelineRunStatus(namespace string, runName string) (map[string]interface{}, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelinerun", runName, "-o", "jsonpath={.status}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline run status: %v", err)
	}

	// Basic status information
	status := map[string]interface{}{
		"name":   runName,
		"output": string(output),
	}

	return status, nil
}
