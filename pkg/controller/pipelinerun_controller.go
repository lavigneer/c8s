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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/scheduler"
	ctypes "github.com/org/c8s/pkg/types"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	LogCollector  *LogCollector
}

// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PipelineRun instance
	pipelineRun := &c8sv1alpha1.PipelineRun{}
	if err := r.Get(ctx, req.NamespacedName, pipelineRun); err != nil {
		// PipelineRun not found, ignore since object must have been deleted
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling PipelineRun",
		"name", pipelineRun.Name,
		"namespace", pipelineRun.Namespace,
		"phase", pipelineRun.Status.Phase,
	)

	// Handle deletion with finalizer
	if !pipelineRun.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, pipelineRun)
	}

	// Add finalizer if not present
	if !containsString(pipelineRun.Finalizers, ctypes.FinalizerPipelineRun) {
		logger.Info("Adding finalizer to PipelineRun")
		pipelineRun.Finalizers = append(pipelineRun.Finalizers, ctypes.FinalizerPipelineRun)
		if err := r.Update(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Skip reconciliation if pipeline is in terminal state
	if r.isTerminalPhase(pipelineRun.Status.Phase) {
		logger.Info("PipelineRun in terminal phase, skipping reconciliation",
			"phase", pipelineRun.Status.Phase,
		)
		return ctrl.Result{}, nil
	}

	// Step 1: Fetch referenced PipelineConfig
	pipelineConfig := &c8sv1alpha1.PipelineConfig{}
	configKey := types.NamespacedName{
		Name:      pipelineRun.Spec.PipelineConfigRef,
		Namespace: pipelineRun.Namespace,
	}
	if err := r.Get(ctx, configKey, pipelineConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "PipelineConfig not found",
				"config", pipelineRun.Spec.PipelineConfigRef,
			)
			// Update status to Failed
			pipelineRun.Status.Phase = c8sv1alpha1.PipelineRunPhaseFailed
			if updateErr := r.Status().Update(ctx, pipelineRun); updateErr != nil {
				logger.Error(updateErr, "Failed to update PipelineRun status")
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Step 2: Initialize status if needed
	if pipelineRun.Status.Phase == "" {
		logger.Info("Initializing PipelineRun status")
		pipelineRun.Status.Phase = c8sv1alpha1.PipelineRunPhasePending
		if err := r.Status().Update(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to initialize PipelineRun status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Step 3: Build execution schedule using DAG scheduler
	schedule, err := scheduler.BuildSchedule(pipelineConfig)
	if err != nil {
		logger.Error(err, "Failed to build execution schedule")
		pipelineRun.Status.Phase = c8sv1alpha1.PipelineRunPhaseFailed
		if updateErr := r.Status().Update(ctx, pipelineRun); updateErr != nil {
			logger.Error(updateErr, "Failed to update PipelineRun status")
		}
		return ctrl.Result{}, nil
	}

	logger.Info("Built execution schedule",
		"totalSteps", schedule.TotalSteps(),
		"layers", schedule.LayerCount(),
	)

	// Step 4: Get completed steps to determine which steps are ready
	completedSteps := GetCompletedSteps(pipelineRun)
	logger.Info("Completed steps", "count", len(completedSteps))

	// Step 5: Create Jobs for steps that are ready to execute
	jobManager := NewJobManager(pipelineConfig.Spec.Repository)
	readySteps := schedule.GetReadySteps(completedSteps)

	for _, step := range readySteps {
		// Check if Job already exists
		jobName := GetJobForStep(pipelineRun.Name, step.Name)
		existingJob := &batchv1.Job{}
		jobKey := types.NamespacedName{
			Name:      jobName,
			Namespace: pipelineRun.Namespace,
		}

		err := r.Get(ctx, jobKey, existingJob)
		if err == nil {
			// Job already exists, skip creation
			logger.Info("Job already exists", "step", step.Name, "job", jobName)
			continue
		}

		if !apierrors.IsNotFound(err) {
			// Real error occurred
			logger.Error(err, "Failed to check if Job exists", "step", step.Name)
			continue
		}

		// Job doesn't exist, create it
		logger.Info("Creating Job for step", "step", step.Name)
		job, err := jobManager.CreateJobForStep(step, pipelineRun, pipelineConfig)
		if err != nil {
			logger.Error(err, "Failed to create Job spec", "step", step.Name)
			continue
		}

		if err := r.Create(ctx, job); err != nil {
			logger.Error(err, "Failed to create Job", "step", step.Name, "job", job.Name)
			continue
		}

		logger.Info("Successfully created Job", "step", step.Name, "job", job.Name)
	}

	// Step 6: List all Jobs owned by this PipelineRun
	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList,
		client.InNamespace(pipelineRun.Namespace),
		client.MatchingLabels{
			ctypes.LabelPipelineRun: pipelineRun.Name,
		},
	); err != nil {
		logger.Error(err, "Failed to list Jobs")
		return ctrl.Result{}, err
	}

	// Build map of jobs by step name
	jobsByStep := make(map[string]*batchv1.Job)
	for i := range jobList.Items {
		job := &jobList.Items[i]
		if stepName, ok := job.Labels[ctypes.LabelStepName]; ok {
			jobsByStep[stepName] = job
		}
	}

	logger.Info("Found Jobs for PipelineRun",
		"totalJobs", len(jobsByStep),
		"expectedSteps", schedule.TotalSteps(),
	)

	// Step 7: Update PipelineRun status based on Job statuses
	statusUpdater := NewStatusUpdater(r.Client)
	if err := statusUpdater.UpdatePipelineRunStatus(ctx, pipelineRun, jobsByStep, schedule.TotalSteps()); err != nil {
		logger.Error(err, "Failed to update PipelineRun status")
		return ctrl.Result{}, err
	}

	// Step 7.5: Collect and upload logs for completed Jobs
	if r.LogCollector != nil {
		if err := r.collectLogsForCompletedJobs(ctx, pipelineRun, pipelineConfig, jobsByStep); err != nil {
			logger.Error(err, "Failed to collect logs for completed jobs")
			// Continue even if log collection fails - don't block pipeline progress
		}
	}

	// Step 8: Requeue if not in terminal state
	if !r.isTerminalPhase(pipelineRun.Status.Phase) {
		logger.Info("PipelineRun still running, requeuing")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	logger.Info("PipelineRun reconciliation complete",
		"finalPhase", pipelineRun.Status.Phase,
	)

	return ctrl.Result{}, nil
}

// collectLogsForCompletedJobs collects logs from completed Job Pods and uploads them
func (r *PipelineRunReconciler) collectLogsForCompletedJobs(ctx context.Context, pipelineRun *c8sv1alpha1.PipelineRun, pipelineConfig *c8sv1alpha1.PipelineConfig, jobsByStep map[string]*batchv1.Job) error {
	logger := log.FromContext(ctx)
	statusUpdated := false

	// Get the current status to check which steps have completed but don't have logs yet
	for i := range pipelineRun.Status.Steps {
		step := &pipelineRun.Status.Steps[i]

		// Skip if logs already collected
		if step.LogURL != "" {
			continue
		}

		// Skip if step hasn't completed yet
		if step.Phase != c8sv1alpha1.StepPhaseSucceeded && step.Phase != c8sv1alpha1.StepPhaseFailed {
			continue
		}

		// Find the Job for this step
		job, ok := jobsByStep[step.Name]
		if !ok {
			logger.Info("Job not found for step", "step", step.Name)
			continue
		}

		// Find the Pod created by this Job
		podList := &corev1.PodList{}
		if err := r.List(ctx, podList,
			client.InNamespace(pipelineRun.Namespace),
			client.MatchingLabels{
				"job-name": job.Name,
			},
		); err != nil {
			logger.Error(err, "Failed to list Pods for Job", "job", job.Name)
			continue
		}

		if len(podList.Items) == 0 {
			logger.Info("No Pod found for Job yet", "job", job.Name, "step", step.Name)
			continue
		}

		// Get the first Pod (Jobs typically create one Pod)
		pod := &podList.Items[0]

		// Skip if Pod is not in a state where we can collect logs
		if pod.Status.Phase != corev1.PodSucceeded &&
		   pod.Status.Phase != corev1.PodFailed &&
		   pod.Status.Phase != corev1.PodRunning {
			logger.Info("Pod not ready for log collection", "pod", pod.Name, "phase", pod.Status.Phase)
			continue
		}

		// Collect and upload logs (with secret masking)
		logger.Info("Collecting logs for step", "step", step.Name, "pod", pod.Name)
		logURL, err := r.LogCollector.CollectAndUpload(ctx, pod, pipelineRun, step.Name, pipelineConfig)
		if err != nil {
			logger.Error(err, "Failed to collect and upload logs", "step", step.Name)
			// Continue to next step - don't block on log collection failures
			continue
		}

		// Update step status with log URL
		pipelineRun.Status.Steps[i].LogURL = logURL
		statusUpdated = true
		logger.Info("Updated step with log URL", "step", step.Name, "url", logURL)
	}

	// Update PipelineRun status if any logs were collected
	if statusUpdated {
		if err := r.Status().Update(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to update PipelineRun status with log URLs")
			return err
		}
	}

	return nil
}

// isTerminalPhase returns true if the phase is terminal (no further transitions)
func (r *PipelineRunReconciler) isTerminalPhase(phase c8sv1alpha1.PipelineRunPhase) bool {
	return phase == c8sv1alpha1.PipelineRunPhaseSucceeded ||
		phase == c8sv1alpha1.PipelineRunPhaseFailed ||
		phase == c8sv1alpha1.PipelineRunPhaseCancelled
}

// handleDeletion handles cleanup when a PipelineRun is being deleted
func (r *PipelineRunReconciler) handleDeletion(ctx context.Context, pipelineRun *c8sv1alpha1.PipelineRun) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if containsString(pipelineRun.Finalizers, ctypes.FinalizerPipelineRun) {
		logger.Info("Cleaning up resources for PipelineRun")

		// Delete all Jobs owned by this PipelineRun
		jobList := &batchv1.JobList{}
		if err := r.List(ctx, jobList,
			client.InNamespace(pipelineRun.Namespace),
			client.MatchingLabels{
				ctypes.LabelPipelineRun: pipelineRun.Name,
			},
		); err != nil {
			logger.Error(err, "Failed to list Jobs for cleanup")
			return ctrl.Result{}, err
		}

		// Delete each Job
		for i := range jobList.Items {
			job := &jobList.Items[i]
			logger.Info("Deleting Job", "job", job.Name)
			if err := r.Delete(ctx, job, client.PropagationPolicy("Background")); err != nil {
				if !apierrors.IsNotFound(err) {
					logger.Error(err, "Failed to delete Job", "job", job.Name)
					return ctrl.Result{}, err
				}
			}
		}

		// Wait a moment for Jobs to be deleted (they may have finalizers too)
		// In a real implementation, you might want to check if all Jobs are gone
		// before removing the finalizer
		logger.Info("Jobs cleanup initiated", "count", len(jobList.Items))

		// Remove finalizer
		pipelineRun.Finalizers = removeString(pipelineRun.Finalizers, ctypes.FinalizerPipelineRun)
		if err := r.Update(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}

		logger.Info("Finalizer removed, PipelineRun will be deleted")
	}

	return ctrl.Result{}, nil
}

// containsString checks if a slice contains a string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&c8sv1alpha1.PipelineRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
