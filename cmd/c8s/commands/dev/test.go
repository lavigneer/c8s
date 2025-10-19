package dev

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/org/c8s/pkg/localenv/samples"
	"github.com/spf13/cobra"
)

// newTestCommand creates the test subcommand
func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run and monitor pipeline tests",
		Long: `Run and monitor end-to-end pipeline tests on a local cluster.

This command manages all aspects of pipeline testing:
- Running pipeline tests and collecting results
- Viewing and streaming pipeline logs
- Monitoring pipeline execution status
- Collecting metrics and debugging information

Use 'c8s dev test run' to execute tests and 'c8s dev test logs' to view results.`,
	}

	cmd.AddCommand(newTestRunCommand())
	cmd.AddCommand(newTestLogsCommand())

	return cmd
}

// newTestRunCommand creates the test run subcommand
func newTestRunCommand() *cobra.Command {
	var (
		clusterName    string
		pipelineFilter string
		namespace      string
		timeout        int
		watch          bool
		outputFormat   string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run pipeline tests",
		Long: `Execute pipeline tests on a cluster.

This command:
1. Lists all deployed PipelineConfigs
2. Creates PipelineRun resources for each
3. Monitors execution and collects results
4. Displays test summary and results

The operator must be deployed before running tests.

Example:
  c8s dev test run --cluster c8s-dev
  c8s dev test run --pipeline simple-build --watch
  c8s dev test run --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "Running pipeline tests on cluster %q\n", clusterName)
			}

			testTimeout := time.Duration(timeout) * time.Second

			// Run tests
			summary, err := samples.RunPipelineTests(ctx, namespace, pipelineFilter, testTimeout)
			if err != nil {
				return fmt.Errorf("failed to run pipeline tests: %w", err)
			}

			// Format and display results
			return displayTestResults(summary, outputFormat, watch)
		},
	}

	// Flags
	cmd.Flags().StringVar(&clusterName, "cluster", "c8s-dev",
		"Name of the cluster to run tests on")
	cmd.Flags().StringVar(&pipelineFilter, "pipeline", "",
		"Run only pipelines matching this name")
	cmd.Flags().StringVar(&namespace, "namespace", "default",
		"Kubernetes namespace containing pipelines")
	cmd.Flags().IntVar(&timeout, "timeout", 600,
		"Timeout in seconds for all tests")
	cmd.Flags().BoolVar(&watch, "watch", false,
		"Watch test progress in real-time")
	cmd.Flags().StringVar(&outputFormat, "output", "text",
		"Output format: text, json, yaml")

	return cmd
}

// newTestLogsCommand creates the test logs subcommand
func newTestLogsCommand() *cobra.Command {
	var (
		clusterName    string
		pipelineFilter string
		namespace      string
		follow         bool
		tail           int
		outputFormat   string
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View pipeline test logs",
		Long: `Display logs from pipeline execution.

This command:
1. Finds PipelineRun resources
2. Retrieves associated pod logs
3. Displays logs from all execution steps
4. Optionally streams logs in real-time

Example:
  c8s dev test logs --cluster c8s-dev
  c8s dev test logs --pipeline simple-build --follow
  c8s dev test logs --tail 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "Fetching pipeline logs from cluster %q\n", clusterName)
			}

			fetcher := samples.NewLogFetcher(namespace, follow, tail)

			if follow {
				// Stream logs
				outputChan := make(chan string, 10)
				go func() {
					fetcher.StreamLogs(pipelineFilter, outputChan)
					close(outputChan)
				}()

				for log := range outputChan {
					fmt.Print(log)
				}
			} else {
				// Get logs once
				logs, err := fetcher.FetchPipelineLogs(pipelineFilter)
				if err != nil {
					return fmt.Errorf("failed to fetch logs: %w", err)
				}

				fmt.Print(logs)
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&clusterName, "cluster", "c8s-dev",
		"Name of the cluster")
	cmd.Flags().StringVar(&pipelineFilter, "pipeline", "",
		"View logs for specific pipeline")
	cmd.Flags().StringVar(&namespace, "namespace", "default",
		"Kubernetes namespace")
	cmd.Flags().BoolVar(&follow, "follow", false,
		"Follow logs in real-time (-f)")
	cmd.Flags().IntVar(&tail, "tail", 0,
		"Show last N lines of logs (0 = all)")
	cmd.Flags().StringVar(&outputFormat, "output", "formatted",
		"Output format: raw, formatted, json")

	return cmd
}

// displayTestResults formats and displays test results
func displayTestResults(summary *samples.PipelineTestSummary, format string, watch bool) error {
	switch format {
	case "json":
		return displayTestResultsJSON(summary)
	case "yaml":
		return displayTestResultsYAML(summary)
	default:
		return displayTestResultsText(summary)
	}
}

// displayTestResultsText displays results in text format
func displayTestResultsText(summary *samples.PipelineTestSummary) error {
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("Pipeline Test Results\n")
	fmt.Printf("%s\n", strings.Repeat("=", 70))

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total Tests:    %d\n", summary.TotalTests)
	fmt.Printf("  Passed:         %d ✓\n", summary.PassedTests)
	fmt.Printf("  Failed:         %d ✗\n", summary.FailedTests)
	fmt.Printf("  Timeout:        %d ⏱\n", summary.TimeoutTests)
	fmt.Printf("  Total Duration: %v\n", summary.Duration)

	fmt.Printf("\nResults:\n")
	for i, result := range summary.Results {
		status := "?"
		switch result.Status {
		case "Success":
			status = "✓"
		case "Failed":
			status = "✗"
		case "Timeout":
			status = "⏱"
		}

		fmt.Printf("  [%d] %s %s (%v)\n", i+1, status, result.Name, result.Duration)
		if result.ErrorMsg != "" {
			fmt.Printf("      Error: %s\n", result.ErrorMsg)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))

	// Exit code based on results
	if summary.FailedTests > 0 || summary.TimeoutTests > 0 {
		return fmt.Errorf("tests failed: %d failed, %d timeout", summary.FailedTests, summary.TimeoutTests)
	}

	return nil
}

// displayTestResultsJSON displays results in JSON format
func displayTestResultsJSON(summary *samples.PipelineTestSummary) error {
	fmt.Printf(`{
  "totalTests": %d,
  "passed": %d,
  "failed": %d,
  "timeout": %d,
  "duration": "%v",
  "message": "%s",
  "results": [
`, summary.TotalTests, summary.PassedTests, summary.FailedTests,
		summary.TimeoutTests, summary.Duration, summary.Message)

	for i, result := range summary.Results {
		if i > 0 {
			fmt.Printf(",\n")
		}
		fmt.Printf(`    {
      "name": "%s",
      "namespace": "%s",
      "status": "%s",
      "duration": "%v",
      "error": "%s"
    }`, result.Name, result.Namespace, result.Status, result.Duration, result.ErrorMsg)
	}

	fmt.Printf("\n  ]\n}\n")
	return nil
}

// displayTestResultsYAML displays results in YAML format
func displayTestResultsYAML(summary *samples.PipelineTestSummary) error {
	fmt.Printf(`summary:
  totalTests: %d
  passed: %d
  failed: %d
  timeout: %d
  duration: %v
  message: "%s"
results:
`, summary.TotalTests, summary.PassedTests, summary.FailedTests,
		summary.TimeoutTests, summary.Duration, summary.Message)

	for _, result := range summary.Results {
		fmt.Printf(`  - name: %s
    namespace: %s
    status: %s
    duration: %v
    error: "%s"
`, result.Name, result.Namespace, result.Status, result.Duration, result.ErrorMsg)
	}

	return nil
}
