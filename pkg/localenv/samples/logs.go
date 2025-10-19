package samples

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LogFetcher handles fetching and streaming logs from pipeline runs
type LogFetcher struct {
	namespace string
	follow    bool
	tailLines int
}

// NewLogFetcher creates a new log fetcher
func NewLogFetcher(namespace string, follow bool, tailLines int) *LogFetcher {
	return &LogFetcher{
		namespace: namespace,
		follow:    follow,
		tailLines: tailLines,
	}
}

// FetchPipelineLogs fetches logs from pipeline runs
func (lf *LogFetcher) FetchPipelineLogs(pipelineFilter string) (string, error) {
	if lf.namespace == "" {
		lf.namespace = "default"
	}

	// List pipeline runs
	runs, err := listPipelineRuns(lf.namespace, pipelineFilter)
	if err != nil {
		return "", fmt.Errorf("failed to list PipelineRuns: %v", err)
	}

	if len(runs) == 0 {
		return "No PipelineRuns found", nil
	}

	var output strings.Builder

	for _, run := range runs {
		output.WriteString(fmt.Sprintf("\n=== Pipeline Run: %s ===\n", run))

		// Get pods for this run
		pods, err := getPipelineRunPods(lf.namespace, run)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error getting pods: %v\n", err))
			continue
		}

		// Get logs from each pod
		for _, pod := range pods {
			output.WriteString(fmt.Sprintf("\n--- Pod: %s ---\n", pod))

			if lf.follow {
				// Stream logs
				logs, err := streamPodLogs(lf.namespace, pod, lf.tailLines)
				if err != nil {
					output.WriteString(fmt.Sprintf("Error streaming logs: %v\n", err))
				} else {
					output.WriteString(logs)
				}
			} else {
				// Get logs once
				logs, err := getPodLogs(lf.namespace, pod, lf.tailLines)
				if err != nil {
					output.WriteString(fmt.Sprintf("Error getting logs: %v\n", err))
				} else {
					output.WriteString(logs)
				}
			}
		}
	}

	return output.String(), nil
}

// StreamLogs streams logs from a pipeline run
func (lf *LogFetcher) StreamLogs(pipelineFilter string, outputChan chan string) error {
	if lf.namespace == "" {
		lf.namespace = "default"
	}

	// List pipeline runs
	runs, err := listPipelineRuns(lf.namespace, pipelineFilter)
	if err != nil {
		return fmt.Errorf("failed to list PipelineRuns: %v", err)
	}

	if len(runs) == 0 {
		outputChan <- "No PipelineRuns found"
		return nil
	}

	// Stream logs from each pod
	for _, run := range runs {
		outputChan <- fmt.Sprintf("\n=== Pipeline Run: %s ===\n", run)

		pods, err := getPipelineRunPods(lf.namespace, run)
		if err != nil {
			outputChan <- fmt.Sprintf("Error getting pods: %v\n", err)
			continue
		}

		for _, pod := range pods {
			outputChan <- fmt.Sprintf("\n--- Pod: %s ---\n", pod)

			// Stream logs for this pod
			cmd := exec.Command("kubectl", "-n", lf.namespace, "logs", pod, "-f")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				outputChan <- fmt.Sprintf("Error creating log stream: %v\n", err)
				continue
			}

			if err := cmd.Start(); err != nil {
				outputChan <- fmt.Sprintf("Error starting log stream: %v\n", err)
				continue
			}

			// Read and stream logs line by line
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				outputChan <- scanner.Text() + "\n"
			}

			cmd.Wait()
		}
	}

	return nil
}

// Helper functions

// listPipelineRuns lists pipeline runs matching a filter
func listPipelineRuns(namespace string, filter string) ([]string, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pipelineruns", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
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

// getPipelineRunPods gets pods associated with a pipeline run
func getPipelineRunPods(namespace string, runName string) ([]string, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l",
		fmt.Sprintf("pipelinerun=%s", runName), "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(output)), nil
}

// getPodLogs fetches logs from a pod
func getPodLogs(namespace string, podName string, tailLines int) (string, error) {
	var cmd *exec.Cmd

	if tailLines > 0 {
		cmd = exec.Command("kubectl", "-n", namespace, "logs", podName, fmt.Sprintf("--tail=%d", tailLines))
	} else {
		cmd = exec.Command("kubectl", "-n", namespace, "logs", podName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v\nOutput: %s", err, output)
	}

	return string(output), nil
}

// streamPodLogs streams logs from a pod
func streamPodLogs(namespace string, podName string, tailLines int) (string, error) {
	var cmd *exec.Cmd

	if tailLines > 0 {
		cmd = exec.Command("kubectl", "-n", namespace, "logs", podName, fmt.Sprintf("--tail=%d", tailLines))
	} else {
		cmd = exec.Command("kubectl", "-n", namespace, "logs", podName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to stream logs: %v", err)
	}

	return string(output), nil
}

// TailLogs continuously tails logs from running pods
func TailLogs(namespace string, pipelineFilter string, tailLines int) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get current logs
		runs, err := listPipelineRuns(namespace, pipelineFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing pipeline runs: %v\n", err)
			continue
		}

		for _, run := range runs {
			pods, err := getPipelineRunPods(namespace, run)
			if err != nil {
				continue
			}

			for _, pod := range pods {
				logs, err := getPodLogs(namespace, pod, tailLines)
				if err != nil {
					continue
				}

				// Output new logs
				if logs != "" {
					fmt.Println(logs)
				}
			}
		}
	}

	return nil
}
