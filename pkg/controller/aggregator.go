package controller

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// AggregatedResult represents aggregated results from matrix executions
type AggregatedResult struct {
	TotalRuns       int
	SucceededRuns   int
	FailedRuns      int
	PendingRuns     int
	RunningRuns     int
	CancelledRuns   int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	LogURLs         []string
	MatrixRuns      []*MatrixRunSummary
}

// MatrixRunSummary provides a summary of a single matrix run
type MatrixRunSummary struct {
	Name        string
	Phase       c8sv1alpha1.PipelineRunPhase
	MatrixVars  map[string]string
	StartTime   *metav1.Time
	CompletionTime *metav1.Time
	Duration    time.Duration
	StepCount   int
	FailedSteps int
}

// AggregateMatrixResults aggregates results from multiple matrix PipelineRuns
func AggregateMatrixResults(ctx context.Context, c client.Client, namespace, parentID string) (*AggregatedResult, error) {
	// List all matrix runs
	matrixRuns, err := ListMatrixRuns(ctx, c, namespace, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list matrix runs: %w", err)
	}

	if len(matrixRuns) == 0 {
		return nil, fmt.Errorf("no matrix runs found for parent: %s", parentID)
	}

	result := &AggregatedResult{
		TotalRuns:  len(matrixRuns),
		MatrixRuns: make([]*MatrixRunSummary, 0, len(matrixRuns)),
	}

	var totalDurationSeconds int64

	for _, run := range matrixRuns {
		// Count by phase
		switch run.Status.Phase {
		case c8sv1alpha1.PipelineRunPhaseSucceeded:
			result.SucceededRuns++
		case c8sv1alpha1.PipelineRunPhaseFailed:
			result.FailedRuns++
		case c8sv1alpha1.PipelineRunPhasePending:
			result.PendingRuns++
		case c8sv1alpha1.PipelineRunPhaseRunning:
			result.RunningRuns++
		case c8sv1alpha1.PipelineRunPhaseCancelled:
			result.CancelledRuns++
		}

		// Calculate duration
		var duration time.Duration
		if run.Status.StartTime != nil && run.Status.CompletionTime != nil {
			duration = run.Status.CompletionTime.Sub(run.Status.StartTime.Time)
			totalDurationSeconds += int64(duration.Seconds())
		}

		// Count failed steps
		failedSteps := 0
		for _, step := range run.Status.Steps {
			if step.Phase == c8sv1alpha1.StepPhaseFailed {
				failedSteps++
			}
		}

		// Collect log URLs
		for _, step := range run.Status.Steps {
			if step.LogURL != "" {
				result.LogURLs = append(result.LogURLs, step.LogURL)
			}
		}

		// Create summary
		summary := &MatrixRunSummary{
			Name:           run.Name,
			Phase:          run.Status.Phase,
			MatrixVars:     run.Spec.MatrixIndex,
			StartTime:      run.Status.StartTime,
			CompletionTime: run.Status.CompletionTime,
			Duration:       duration,
			StepCount:      len(run.Status.Steps),
			FailedSteps:    failedSteps,
		}
		result.MatrixRuns = append(result.MatrixRuns, summary)
	}

	// Calculate average duration
	if result.SucceededRuns > 0 {
		result.TotalDuration = time.Duration(totalDurationSeconds) * time.Second
		result.AverageDuration = time.Duration(totalDurationSeconds/int64(result.SucceededRuns)) * time.Second
	}

	return result, nil
}

// FormatAggregatedResult formats the aggregated result as a human-readable string
func FormatAggregatedResult(result *AggregatedResult) string {
	successRate := 0.0
	if result.TotalRuns > 0 {
		successRate = float64(result.SucceededRuns) / float64(result.TotalRuns) * 100
	}

	return fmt.Sprintf(
		"Matrix Results: %d/%d succeeded (%.1f%%), %d failed, %d running, %d pending, avg duration: %s",
		result.SucceededRuns,
		result.TotalRuns,
		successRate,
		result.FailedRuns,
		result.RunningRuns,
		result.PendingRuns,
		result.AverageDuration,
	)
}

// IsMatrixComplete checks if all matrix runs have completed
func IsMatrixComplete(result *AggregatedResult) bool {
	return result.PendingRuns == 0 && result.RunningRuns == 0
}

// GetMatrixSuccessRate calculates the success rate of matrix runs
func GetMatrixSuccessRate(result *AggregatedResult) float64 {
	if result.TotalRuns == 0 {
		return 0.0
	}
	return float64(result.SucceededRuns) / float64(result.TotalRuns) * 100
}

// GetFailedMatrixRuns returns summaries of failed matrix runs
func GetFailedMatrixRuns(result *AggregatedResult) []*MatrixRunSummary {
	var failed []*MatrixRunSummary
	for _, run := range result.MatrixRuns {
		if run.Phase == c8sv1alpha1.PipelineRunPhaseFailed {
			failed = append(failed, run)
		}
	}
	return failed
}

// GetFastestMatrixRun returns the matrix run with shortest duration
func GetFastestMatrixRun(result *AggregatedResult) *MatrixRunSummary {
	if len(result.MatrixRuns) == 0 {
		return nil
	}

	var fastest *MatrixRunSummary
	for _, run := range result.MatrixRuns {
		if run.Duration == 0 {
			continue
		}
		if fastest == nil || run.Duration < fastest.Duration {
			fastest = run
		}
	}
	return fastest
}

// GetSlowestMatrixRun returns the matrix run with longest duration
func GetSlowestMatrixRun(result *AggregatedResult) *MatrixRunSummary {
	if len(result.MatrixRuns) == 0 {
		return nil
	}

	var slowest *MatrixRunSummary
	for _, run := range result.MatrixRuns {
		if run.Duration == 0 {
			continue
		}
		if slowest == nil || run.Duration > slowest.Duration {
			slowest = run
		}
	}
	return slowest
}

// UpdateMatrixParentStatus updates the parent PipelineRun status with aggregated results
func UpdateMatrixParentStatus(ctx context.Context, c client.Client, parent *c8sv1alpha1.PipelineRun, result *AggregatedResult) error {
	// Update status message with aggregated results
	parent.Status.Phase = getOverallPhase(result)

	// Set completion time if all runs are complete
	if IsMatrixComplete(result) {
		now := metav1.Now()
		parent.Status.CompletionTime = &now
	}

	// Store aggregated duration
	if result.AverageDuration > 0 {
		parent.Status.ResourceUsage = &c8sv1alpha1.ResourceUsage{
			Duration: int64(result.AverageDuration.Seconds()),
		}
	}

	// Update the parent status
	return c.Status().Update(ctx, parent)
}

// getOverallPhase determines the overall phase based on matrix run statuses
func getOverallPhase(result *AggregatedResult) c8sv1alpha1.PipelineRunPhase {
	// If any failed, overall is failed
	if result.FailedRuns > 0 {
		return c8sv1alpha1.PipelineRunPhaseFailed
	}

	// If any running, overall is running
	if result.RunningRuns > 0 {
		return c8sv1alpha1.PipelineRunPhaseRunning
	}

	// If any pending, overall is pending
	if result.PendingRuns > 0 {
		return c8sv1alpha1.PipelineRunPhasePending
	}

	// If all succeeded, overall is succeeded
	if result.SucceededRuns == result.TotalRuns {
		return c8sv1alpha1.PipelineRunPhaseSucceeded
	}

	// Default to pending
	return c8sv1alpha1.PipelineRunPhasePending
}
