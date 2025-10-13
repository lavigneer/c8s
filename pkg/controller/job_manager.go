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
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/types"
)

// JobManager handles creation and management of Kubernetes Jobs for pipeline steps
type JobManager struct {
	// Repository URL from PipelineConfig
	repository string
}

// NewJobManager creates a new JobManager
func NewJobManager(repository string) *JobManager {
	return &JobManager{
		repository: repository,
	}
}

// CreateJobForStep creates a Kubernetes Job for a pipeline step
func (jm *JobManager) CreateJobForStep(
	step *c8sv1alpha1.PipelineStep,
	pipelineRun *c8sv1alpha1.PipelineRun,
	pipelineConfig *c8sv1alpha1.PipelineConfig,
) (*batchv1.Job, error) {
	jobName := fmt.Sprintf("%s-%s", pipelineRun.Name, step.Name)

	// Parse timeout
	timeout, err := parseTimeout(step.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout for step %s: %w", step.Name, err)
	}

	// Build job spec
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: pipelineRun.Namespace,
			Labels: map[string]string{
				types.LabelPipelineConfig: pipelineRun.Spec.PipelineConfigRef,
				types.LabelPipelineRun:    pipelineRun.Name,
				types.LabelStepName:       step.Name,
				types.LabelCommit:         pipelineRun.Spec.Commit,
				types.LabelBranch:         pipelineRun.Spec.Branch,
				types.LabelManagedBy:      types.ManagedByC8S,
			},
			Annotations: map[string]string{
				types.AnnotationCommitMessage: pipelineRun.Spec.CommitMessage,
				types.AnnotationAuthor:        pipelineRun.Spec.Author,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(pipelineRun, c8sv1alpha1.GroupVersion.WithKind("PipelineRun")),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(types.JobBackoffLimit),
			TTLSecondsAfterFinished: int32Ptr(types.JobTTLSecondsAfterFinished),
			ActiveDeadlineSeconds:   &timeout,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						types.LabelPipelineRun: pipelineRun.Name,
						types.LabelStepName:    step.Name,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						jm.buildGitCloneContainer(pipelineRun),
					},
					Containers: []corev1.Container{
						jm.buildStepContainer(step, pipelineRun),
					},
					Volumes: []corev1.Volume{
						{
							Name: types.VolumeNameWorkspace,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	return job, nil
}

// buildGitCloneContainer creates the init container for git clone
// Uses environment variables to prevent command injection
func (jm *JobManager) buildGitCloneContainer(pipelineRun *c8sv1alpha1.PipelineRun) corev1.Container {
	// Use a shell script with environment variables instead of string interpolation
	// This prevents command injection even if branch/commit values contain special characters
	cloneScript := `set -e
echo "Cloning repository: $REPO_URL"
echo "Branch: $BRANCH"
echo "Commit: $COMMIT"
git clone --depth=1 --single-branch --branch "$BRANCH" "$REPO_URL" "$WORKSPACE"
cd "$WORKSPACE"
git checkout "$COMMIT"
echo "Repository cloned successfully"`

	return corev1.Container{
		Name:  types.ContainerNameGitClone,
		Image: "alpine/git:latest",
		Command: []string{
			"/bin/sh",
			"-c",
			cloneScript,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      types.VolumeNameWorkspace,
				MountPath: types.MountPathWorkspace,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "REPO_URL",
				Value: jm.repository,
			},
			{
				Name:  "BRANCH",
				Value: pipelineRun.Spec.Branch,
			},
			{
				Name:  "COMMIT",
				Value: pipelineRun.Spec.Commit,
			},
			{
				Name:  "WORKSPACE",
				Value: types.MountPathWorkspace,
			},
			{
				Name:  types.EnvCommitSHA,
				Value: pipelineRun.Spec.Commit,
			},
			{
				Name:  types.EnvBranch,
				Value: pipelineRun.Spec.Branch,
			},
		},
	}
}

// buildStepContainer creates the main container for the step
func (jm *JobManager) buildStepContainer(
	step *c8sv1alpha1.PipelineStep,
	pipelineRun *c8sv1alpha1.PipelineRun,
) corev1.Container {
	// Build command script
	commandScript := strings.Join(step.Commands, "\n")

	container := corev1.Container{
		Name:       types.ContainerNameStep,
		Image:      step.Image,
		WorkingDir: types.MountPathWorkspace,
		Command: []string{
			"/bin/sh",
			"-c",
			commandScript,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      types.VolumeNameWorkspace,
				MountPath: types.MountPathWorkspace,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  types.EnvCommitSHA,
				Value: pipelineRun.Spec.Commit,
			},
			{
				Name:  types.EnvBranch,
				Value: pipelineRun.Spec.Branch,
			},
			{
				Name:  types.EnvPipelineRun,
				Value: pipelineRun.Name,
			},
			{
				Name:  types.EnvStepName,
				Value: step.Name,
			},
			{
				Name:  types.EnvWorkspace,
				Value: types.MountPathWorkspace,
			},
			{
				Name:  types.EnvC8SNamespace,
				Value: pipelineRun.Namespace,
			},
		},
	}

	// Add resource requirements if specified
	if step.Resources != nil {
		container.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{},
			Limits:   corev1.ResourceList{},
		}

		if step.Resources.CPU != "" {
			if qty, err := resource.ParseQuantity(step.Resources.CPU); err == nil {
				container.Resources.Requests[corev1.ResourceCPU] = qty
				container.Resources.Limits[corev1.ResourceCPU] = qty
			}
		}

		if step.Resources.Memory != "" {
			if qty, err := resource.ParseQuantity(step.Resources.Memory); err == nil {
				container.Resources.Requests[corev1.ResourceMemory] = qty
				container.Resources.Limits[corev1.ResourceMemory] = qty
			}
		}
	}

	// Add secret injection (User Story 3)
	for _, secret := range step.Secrets {
		// If EnvVar is not specified, use the key name as the environment variable name
		envVarName := secret.EnvVar
		if envVarName == "" {
			envVarName = secret.Key
		}

		container.Env = append(container.Env, corev1.EnvVar{
			Name: envVarName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret.SecretRef,
					},
					Key: secret.Key,
				},
			},
		})
	}

	// TODO: Add artifact upload sidecar in Phase 4 (User Story 2)

	return container
}

// parseTimeout converts timeout string (e.g., "30m", "2h") to seconds
func parseTimeout(timeoutStr string) (int64, error) {
	if timeoutStr == "" {
		timeoutStr = "30m"
	}

	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, err
	}

	return int64(duration.Seconds()), nil
}

// int32Ptr returns a pointer to an int32
func int32Ptr(i int32) *int32 {
	return &i
}

// GetJobForStep constructs the expected Job name for a pipeline step
func GetJobForStep(pipelineRunName, stepName string) string {
	return fmt.Sprintf("%s-%s", pipelineRunName, stepName)
}

// IsJobOwnedByPipelineRun checks if a Job is owned by a PipelineRun
func IsJobOwnedByPipelineRun(job *batchv1.Job, pipelineRunName string) bool {
	for _, owner := range job.OwnerReferences {
		if owner.Kind == "PipelineRun" && owner.Name == pipelineRunName {
			return true
		}
	}
	return false
}

// GetJobStatus extracts the status from a Kubernetes Job
func GetJobStatus(job *batchv1.Job) c8sv1alpha1.StepPhase {
	// Check for completion
	if job.Status.Succeeded > 0 {
		return c8sv1alpha1.StepPhaseSucceeded
	}

	// Check for failure
	if job.Status.Failed > 0 {
		return c8sv1alpha1.StepPhaseFailed
	}

	// Check for active
	if job.Status.Active > 0 {
		return c8sv1alpha1.StepPhaseRunning
	}

	// Otherwise pending
	return c8sv1alpha1.StepPhasePending
}

// GetJobExitCode extracts the exit code from a completed Job
// Returns nil if Job is not complete or exit code cannot be determined
func GetJobExitCode(job *batchv1.Job) *int32 {
	// Exit code extraction would require looking at Pod status
	// This is a simplified version; full implementation in Phase 4
	if job.Status.Succeeded > 0 {
		exitCode := int32(0)
		return &exitCode
	}
	if job.Status.Failed > 0 {
		exitCode := int32(1)
		return &exitCode
	}
	return nil
}
