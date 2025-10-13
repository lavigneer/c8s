/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/types"
)

// StatusUpdater handles updating PipelineRun status based on Job statuses
type StatusUpdater struct {
	client client.Client
}

// NewStatusUpdater creates a new StatusUpdater
func NewStatusUpdater(c client.Client) *StatusUpdater {
	return &StatusUpdater{client: c}
}

// UpdatePipelineRunStatus updates the PipelineRun status based on Job statuses
func (su *StatusUpdater) UpdatePipelineRunStatus(
	ctx context.Context,
	pipelineRun *c8sv1alpha1.PipelineRun,
	jobs map[string]*batchv1.Job,
) error {
	// Initialize status if needed
	if pipelineRun.Status.Phase == "" {
		pipelineRun.Status.Phase = c8sv1alpha1.PipelineRunPhasePending
	}

	// Update step statuses from jobs
	stepStatusMap := make(map[string]*c8sv1alpha1.StepStatus)
	for i := range pipelineRun.Status.Steps {
		stepStatusMap[pipelineRun.Status.Steps[i].Name] = &pipelineRun.Status.Steps[i]
	}

	// Track overall status
	var (
		totalSteps      int
		pendingSteps    int
		runningSteps    int
		succeededSteps  int
		failedSteps     int
		hasStarted      bool
	)

	// Update status for each job
	for stepName, job := range jobs {
		status, exists := stepStatusMap[stepName]
		if !exists {
			// Create new status entry
			status = &c8sv1alpha1.StepStatus{
				Name:    stepName,
				Phase:   c8sv1alpha1.StepPhasePending,
				JobName: job.Name,
			}
			stepStatusMap[stepName] = status
			pipelineRun.Status.Steps = append(pipelineRun.Status.Steps, *status)
		}

		// Update from job
		su.updateStepStatusFromJob(status, job)

		// Count by phase
		totalSteps++
		switch status.Phase {
		case c8sv1alpha1.StepPhasePending:
			pendingSteps++
		case c8sv1alpha1.StepPhaseRunning:
			runningSteps++
			hasStarted = true
		case c8sv1alpha1.StepPhaseSucceeded:
			succeededSteps++
			hasStarted = true
		case c8sv1alpha1.StepPhaseFailed:
			failedSteps++
			hasStarted = true
		}
	}

	// Update overall phase
	newPhase := su.calculateOverallPhase(
		pipelineRun.Status.Phase,
		totalSteps,
		pendingSteps,
		runningSteps,
		succeededSteps,
		failedSteps,
	)

	// Handle phase transitions
	if newPhase != pipelineRun.Status.Phase {
		su.handlePhaseTransition(pipelineRun, pipelineRun.Status.Phase, newPhase)
	}

	pipelineRun.Status.Phase = newPhase

	// Update timestamps
	if hasStarted && pipelineRun.Status.StartTime == nil {
		now := metav1.Now()
		pipelineRun.Status.StartTime = &now
	}

	if su.isTerminalPhase(newPhase) && pipelineRun.Status.CompletionTime == nil {
		now := metav1.Now()
		pipelineRun.Status.CompletionTime = &now
	}

	// Update status subresource
	return su.client.Status().Update(ctx, pipelineRun)
}

// updateStepStatusFromJob updates a step status from a Job
func (su *StatusUpdater) updateStepStatusFromJob(status *c8sv1alpha1.StepStatus, job *batchv1.Job) {
	// Update phase
	status.Phase = GetJobStatus(job)
	status.JobName = job.Name

	// Update timestamps
	if job.Status.StartTime != nil && status.StartTime == nil {
		status.StartTime = job.Status.StartTime
	}

	if job.Status.CompletionTime != nil && status.CompletionTime == nil {
		status.CompletionTime = job.Status.CompletionTime
	}

	// Update exit code
	status.ExitCode = GetJobExitCode(job)

	// Update message based on conditions
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			status.Message = condition.Message
			break
		}
		if condition.Type == batchv1.JobComplete && condition.Status == "True" {
			status.Message = "Step completed successfully"
			break
		}
	}

	// TODO: Add log URL in Phase 4 (User Story 2 - Observability)
	// TODO: Add artifact URLs in Phase 4 (User Story 2 - Observability)
}

// calculateOverallPhase determines the overall pipeline phase based on step statuses
func (su *StatusUpdater) calculateOverallPhase(
	currentPhase c8sv1alpha1.PipelineRunPhase,
	totalSteps, pendingSteps, runningSteps, succeededSteps, failedSteps int,
) c8sv1alpha1.PipelineRunPhase {
	// If already in terminal state, don't change
	if su.isTerminalPhase(currentPhase) {
		return currentPhase
	}

	// If any step failed, pipeline fails
	if failedSteps > 0 {
		return c8sv1alpha1.PipelineRunPhaseFailed
	}

	// If all steps succeeded, pipeline succeeds
	if succeededSteps == totalSteps && totalSteps > 0 {
		return c8sv1alpha1.PipelineRunPhaseSucceeded
	}

	// If any step is running, pipeline is running
	if runningSteps > 0 {
		return c8sv1alpha1.PipelineRunPhaseRunning
	}

	// If we have steps but none are running/succeeded/failed, still pending
	if pendingSteps == totalSteps {
		return c8sv1alpha1.PipelineRunPhasePending
	}

	// Mixed state, consider running
	return c8sv1alpha1.PipelineRunPhaseRunning
}

// handlePhaseTransition handles actions when phase changes
func (su *StatusUpdater) handlePhaseTransition(
	pipelineRun *c8sv1alpha1.PipelineRun,
	oldPhase, newPhase c8sv1alpha1.PipelineRunPhase,
) {
	// Add conditions for phase transitions
	now := metav1.Now()

	switch newPhase {
	case c8sv1alpha1.PipelineRunPhaseRunning:
		if oldPhase == c8sv1alpha1.PipelineRunPhasePending {
			su.setCondition(pipelineRun, metav1.Condition{
				Type:               types.ConditionTypeJobsCreated,
				Status:             metav1.ConditionTrue,
				Reason:             types.ReasonJobsCreated,
				Message:            "Pipeline execution started",
				LastTransitionTime: now,
			})
		}

	case c8sv1alpha1.PipelineRunPhaseSucceeded:
		su.setCondition(pipelineRun, metav1.Condition{
			Type:               types.ConditionTypeStepsCompleted,
			Status:             metav1.ConditionTrue,
			Reason:             types.ReasonStepSucceeded,
			Message:            "All steps completed successfully",
			LastTransitionTime: now,
		})

	case c8sv1alpha1.PipelineRunPhaseFailed:
		su.setCondition(pipelineRun, metav1.Condition{
			Type:               types.ConditionTypeStepsCompleted,
			Status:             metav1.ConditionFalse,
			Reason:             types.ReasonStepFailed,
			Message:            "One or more steps failed",
			LastTransitionTime: now,
		})
	}
}

// setCondition sets or updates a condition in the PipelineRun status
func (su *StatusUpdater) setCondition(pipelineRun *c8sv1alpha1.PipelineRun, condition metav1.Condition) {
	// Find existing condition
	for i, existingCondition := range pipelineRun.Status.Conditions {
		if existingCondition.Type == condition.Type {
			// Update existing condition
			pipelineRun.Status.Conditions[i] = condition
			return
		}
	}

	// Add new condition
	pipelineRun.Status.Conditions = append(pipelineRun.Status.Conditions, condition)
}

// isTerminalPhase returns true if the phase is terminal (no further transitions)
func (su *StatusUpdater) isTerminalPhase(phase c8sv1alpha1.PipelineRunPhase) bool {
	return phase == c8sv1alpha1.PipelineRunPhaseSucceeded ||
		phase == c8sv1alpha1.PipelineRunPhaseFailed ||
		phase == c8sv1alpha1.PipelineRunPhaseCancelled
}

// GetStepStatus returns the status for a specific step
func GetStepStatus(pipelineRun *c8sv1alpha1.PipelineRun, stepName string) *c8sv1alpha1.StepStatus {
	for i := range pipelineRun.Status.Steps {
		if pipelineRun.Status.Steps[i].Name == stepName {
			return &pipelineRun.Status.Steps[i]
		}
	}
	return nil
}

// GetCompletedSteps returns a map of completed step names
func GetCompletedSteps(pipelineRun *c8sv1alpha1.PipelineRun) map[string]bool {
	completed := make(map[string]bool)
	for _, step := range pipelineRun.Status.Steps {
		if step.Phase == c8sv1alpha1.StepPhaseSucceeded {
			completed[step.Name] = true
		}
	}
	return completed
}

// IsStepReady returns true if a step is ready to execute (dependencies satisfied)
func IsStepReady(stepName string, dependencies []string, completedSteps map[string]bool) bool {
	for _, dep := range dependencies {
		if !completedSteps[dep] {
			return false
		}
	}
	return true
}
