package samples

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecutionStatus tracks the status of a pipeline execution
type ExecutionStatus struct {
	PipelineRun string
	Status      string // Running, Succeeded, Failed
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	StepStatus  []StepExecutionStatus
	ErrorMsg    string
}

// StepExecutionStatus tracks the status of individual pipeline steps
type StepExecutionStatus struct {
	Name       string
	Status     string // Running, Succeeded, Failed
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Retries    int
	ErrorMsg   string
}

// PipelineExecutionMonitor monitors pipeline execution
type PipelineExecutionMonitor struct {
	namespace string
	timeout   time.Duration
}

// NewPipelineExecutionMonitor creates a new execution monitor
func NewPipelineExecutionMonitor(namespace string, timeout time.Duration) *PipelineExecutionMonitor {
	return &PipelineExecutionMonitor{
		namespace: namespace,
		timeout:   timeout,
	}
}

// MonitorExecution monitors a pipeline run execution
func (m *PipelineExecutionMonitor) MonitorExecution(runName string) (*ExecutionStatus, error) {
	status := &ExecutionStatus{
		PipelineRun: runName,
		StartTime:   time.Now(),
	}

	if m.namespace == "" {
		m.namespace = "default"
	}

	startTime := time.Now()

	for {
		// Get pipeline run status
		cmd := exec.Command("kubectl", "-n", m.namespace, "get", "pipelinerun", runName, "-o", "jsonpath={.status.phase}")
		output, err := cmd.CombinedOutput()
		if err != nil {
			status.ErrorMsg = fmt.Sprintf("Failed to get status: %v", err)
			return status, err
		}

		phase := strings.TrimSpace(string(output))
		status.Status = phase

		// Check if completed
		if phase == "Succeeded" || phase == "Failed" || phase == "Error" {
			status.EndTime = time.Now()
			status.Duration = status.EndTime.Sub(status.StartTime)
			break
		}

		// Check timeout
		if time.Since(startTime) > m.timeout {
			status.Status = "Timeout"
			status.EndTime = time.Now()
			status.Duration = status.EndTime.Sub(status.StartTime)
			status.ErrorMsg = "Pipeline execution timed out"
			break
		}

		time.Sleep(2 * time.Second)
	}

	// Get step status
	stepStatus, err := m.getStepStatus(runName)
	if err == nil {
		status.StepStatus = stepStatus
	}

	return status, nil
}

// getStepStatus retrieves status of individual steps
func (m *PipelineExecutionMonitor) getStepStatus(runName string) ([]StepExecutionStatus, error) {
	// Get jobs associated with the pipeline run
	cmd := exec.Command("kubectl", "-n", m.namespace, "get", "jobs", "-l",
		fmt.Sprintf("pipelinerun=%s", runName), "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	jobNames := strings.Fields(string(output))
	var stepStatuses []StepExecutionStatus

	for _, jobName := range jobNames {
		// Get job status
		statusCmd := exec.Command("kubectl", "-n", m.namespace, "get", "job", jobName, "-o",
			"jsonpath={.status.succeeded,\",\",status.failed,\",\",status.startTime,\",\",status.completionTime}")
		statusOutput, err := statusCmd.CombinedOutput()
		if err != nil {
			continue
		}

		statusStr := string(statusOutput)
		parts := strings.Split(statusStr, ",")

		stepStatus := StepExecutionStatus{
			Name: jobName,
		}

		if len(parts) > 0 && parts[0] != "" {
			if parts[0] == "1" {
				stepStatus.Status = "Succeeded"
			} else if len(parts) > 1 && parts[1] != "" && parts[1] != "0" {
				stepStatus.Status = "Failed"
			} else {
				stepStatus.Status = "Running"
			}
		}

		stepStatuses = append(stepStatuses, stepStatus)
	}

	return stepStatuses, nil
}

// WatchPipelineEvents watches for pipeline execution events
func (m *PipelineExecutionMonitor) WatchPipelineEvents(runName string, eventChan chan string) error {
	// Watch for pod events
	cmd := exec.Command("kubectl", "-n", m.namespace, "get", "events", "--field-selector",
		fmt.Sprintf("involvedObject.name=%s", runName), "-w", "-o", "json")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read events line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		event := scanner.Text()
		eventChan <- event
	}

	cmd.Wait()
	return nil
}

// WatchPipelineEventsWithScanner watches for pipeline execution events using a scanner
func (m *PipelineExecutionMonitor) WatchPipelineEventsWithScanner(runName string, eventChan chan string) error {
	// Watch for pod events
	cmd := exec.Command("kubectl", "-n", m.namespace, "get", "events", "--field-selector",
		fmt.Sprintf("involvedObject.name=%s", runName), "-w", "-o", "json")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read events line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		event := scanner.Text()
		eventChan <- event
	}

	cmd.Wait()
	return nil
}

// DetectFailures analyzes execution status to detect failures
func (m *PipelineExecutionMonitor) DetectFailures(status *ExecutionStatus) []string {
	var failures []string

	if status.Status == "Failed" || status.Status == "Error" {
		failures = append(failures, fmt.Sprintf("Pipeline failed with status: %s", status.Status))
	}

	// Check step failures
	for _, step := range status.StepStatus {
		if step.Status == "Failed" {
			failures = append(failures, fmt.Sprintf("Step %s failed", step.Name))
			if step.ErrorMsg != "" {
				failures = append(failures, fmt.Sprintf("  Error: %s", step.ErrorMsg))
			}
		}
	}

	if status.ErrorMsg != "" {
		failures = append(failures, fmt.Sprintf("Execution error: %s", status.ErrorMsg))
	}

	return failures
}

// CalculateDurations calculates execution durations
func CalculateDurations(status *ExecutionStatus) map[string]time.Duration {
	durations := map[string]time.Duration{
		"total": status.Duration,
	}

	for i, step := range status.StepStatus {
		if step.Duration > 0 {
			durations[fmt.Sprintf("step_%d", i)] = step.Duration
		}
	}

	return durations
}
