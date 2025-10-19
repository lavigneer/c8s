package cluster

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Workload represents an active workload in the cluster
type Workload struct {
	Name       string
	Type       string // PipelineRun, Job, Pod
	Namespace  string
	Status     string
	Age        time.Duration
	CreatedAt  time.Time
}

// WorkloadStatus contains information about active workloads
type WorkloadStatus struct {
	HasActiveWorkloads bool
	Workloads          []Workload
	Message            string
}

// DetectActiveWorkloads checks for running PipelineRuns, Jobs, and Pods
func DetectActiveWorkloads(namespace string) (*WorkloadStatus, error) {
	status := &WorkloadStatus{
		Workloads: []Workload{},
	}

	if namespace == "" {
		namespace = "default"
	}

	// Check for PipelineRuns
	pipelineRuns, err := getPipelineRuns(namespace)
	if err == nil {
		status.Workloads = append(status.Workloads, pipelineRuns...)
	}

	// Check for active Jobs
	jobs, err := getActiveJobs(namespace)
	if err == nil {
		status.Workloads = append(status.Workloads, jobs...)
	}

	// Check for running Pods
	pods, err := getRunningPods(namespace)
	if err == nil {
		status.Workloads = append(status.Workloads, pods...)
	}

	// Determine if there are active workloads
	status.HasActiveWorkloads = len(status.Workloads) > 0

	if status.HasActiveWorkloads {
		status.Message = fmt.Sprintf("Found %d active workload(s)", len(status.Workloads))
	} else {
		status.Message = "No active workloads detected"
	}

	return status, nil
}

// getPipelineRuns lists active PipelineRun resources
func getPipelineRuns(namespace string) ([]Workload, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelineruns", "-o",
		"jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.phase}{\"\\t\"}{.metadata.creationTimestamp}{\"\\n\"}{end}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var workloads []Workload
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			// Skip completed runs
			if strings.Contains(parts[1], "Succeeded") || strings.Contains(parts[1], "Failed") {
				continue
			}

			workload := Workload{
				Name:      parts[0],
				Type:      "PipelineRun",
				Namespace: namespace,
				Status:    parts[1],
			}

			if len(parts) >= 3 {
				workload.CreatedAt = parseTime(parts[2])
				workload.Age = time.Since(workload.CreatedAt)
			}

			workloads = append(workloads, workload)
		}
	}

	return workloads, nil
}

// getActiveJobs lists active Job resources
func getActiveJobs(namespace string) ([]Workload, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "jobs", "-o",
		"jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.succeeded}{\"\\t\"}{.status.failed}{\"\\t\"}{.metadata.creationTimestamp}{\"\\n\"}{end}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var workloads []Workload
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			// Skip completed jobs
			if parts[1] != "" || parts[2] != "" {
				continue
			}

			workload := Workload{
				Name:      parts[0],
				Type:      "Job",
				Namespace: namespace,
				Status:    "Running",
			}

			if len(parts) >= 4 {
				workload.CreatedAt = parseTime(parts[3])
				workload.Age = time.Since(workload.CreatedAt)
			}

			workloads = append(workloads, workload)
		}
	}

	return workloads, nil
}

// getRunningPods lists running Pod resources
func getRunningPods(namespace string) ([]Workload, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-o",
		"jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.phase}{\"\\t\"}{.metadata.creationTimestamp}{\"\\n\"}{end}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var workloads []Workload
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			// Only include running pods
			if parts[1] != "Running" && parts[1] != "Pending" {
				continue
			}

			workload := Workload{
				Name:      parts[0],
				Type:      "Pod",
				Namespace: namespace,
				Status:    parts[1],
			}

			if len(parts) >= 3 {
				workload.CreatedAt = parseTime(parts[2])
				workload.Age = time.Since(workload.CreatedAt)
			}

			workloads = append(workloads, workload)
		}
	}

	return workloads, nil
}

// parseTime parses Kubernetes timestamp format
func parseTime(timeStr string) time.Time {
	// Try standard format first
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t
	}

	// Return zero time if parsing fails
	return time.Time{}
}

// WaitForWorkloadsCompletion waits for all workloads to complete
func WaitForWorkloadsCompletion(namespace string, timeout time.Duration) error {
	startTime := time.Now()
	checkInterval := 5 * time.Second

	for {
		status, err := DetectActiveWorkloads(namespace)
		if err != nil {
			return err
		}

		if !status.HasActiveWorkloads {
			return nil
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout waiting for workloads to complete after %v", timeout)
		}

		time.Sleep(checkInterval)
	}
}
