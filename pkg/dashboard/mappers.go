package dashboard

import (
	"github.com/org/c8s/pkg/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MapPipelineRunToDTO converts K8s PipelineRun to PipelineRunDTO
func MapPipelineRunToDTO(run *v1alpha1.PipelineRun) *PipelineRunDTO {
	if run == nil {
		return nil
	}

	dto := &PipelineRunDTO{
		ID:          run.Name,
		ProjectID:   run.Labels["project"],
		Name:        run.Name,
		Status:      string(run.Status.Phase),
		CommitSHA:   run.Spec.Commit,
		Branch:      run.Spec.Branch,
		Author:      run.Spec.TriggeredBy,
		AuthorEmail: "",
		TriggerSource: "webhook",
		TriggeredAt: run.CreationTimestamp.Time,
		StepCount:   len(run.Status.Steps),
	}

	// Count step statuses
	for _, step := range run.Status.Steps {
		switch step.Phase {
		case v1alpha1.StepPhaseSucceeded:
			dto.SuccessCount++
		case v1alpha1.StepPhaseFailed:
			dto.FailureCount++
		}
	}

	// Set timing information if available
	if !run.Status.StartTime.IsZero() {
		startTime := run.Status.StartTime.Time
		dto.StartedAt = &startTime
	}

	if !run.Status.CompletionTime.IsZero() {
		completionTime := run.Status.CompletionTime.Time
		dto.CompletedAt = &completionTime

		if dto.StartedAt != nil {
			duration := int64(completionTime.Sub(*dto.StartedAt).Seconds())
			dto.DurationSeconds = &duration
		}
	}

	// Count artifacts - sum all artifact URLs from steps
	for _, step := range run.Status.Steps {
		dto.ArtifactCount += len(step.ArtifactURLs)
	}

	return dto
}

// MapStepStatusToDTO converts K8s StepStatus to StepDTO
func MapStepStatusToDTO(step *v1alpha1.StepStatus) *StepDTO {
	if step == nil {
		return nil
	}

	dto := &StepDTO{
		ID:     step.Name,
		Name:   step.Name,
		Status: string(step.Phase),
		Image:  "",
	}

	// Copy commands - not available in StepStatus, set empty
	dto.Commands = []string{}
	dto.DependsOn = []string{}

	// Set timing information
	if step.StartTime != nil && !step.StartTime.IsZero() {
		startTime := step.StartTime.Time
		dto.StartedAt = &startTime
	}

	if step.CompletionTime != nil && !step.CompletionTime.IsZero() {
		completionTime := step.CompletionTime.Time
		dto.CompletedAt = &completionTime

		if dto.StartedAt != nil {
			duration := int64(completionTime.Sub(*dto.StartedAt).Seconds())
			dto.DurationSeconds = &duration
		}
	}

	// Copy exit code
	if step.ExitCode != nil {
		dto.ExitCode = step.ExitCode
	}

	// Set log URL
	if step.LogURL != "" {
		dto.LogURL = step.LogURL
	}

	return dto
}

// CalculateDuration returns duration in seconds between start and end
func CalculateDuration(start, end *metav1.Time) *int64 {
	if start == nil || end == nil || start.IsZero() || end.IsZero() {
		return nil
	}

	duration := int64(end.Time.Sub(start.Time).Seconds())
	return &duration
}

// MapProjectToDTO converts K8s PipelineConfig to ProjectDTO
func MapProjectToDTO(config *v1alpha1.PipelineConfig) *ProjectDTO {
	if config == nil {
		return nil
	}

	dto := &ProjectDTO{
		ID:        config.Name,
		Name:      config.Name,
		RepoURL:   config.Spec.Repository,
		Namespace: config.Namespace,
		CreatedAt: config.CreationTimestamp.Time,
		RunCount:  0, // TODO: Count actual PipelineRuns
	}

	return dto
}

// MapPipelineConfigToDTO converts K8s PipelineConfig to PipelineConfigDTO
func MapPipelineConfigToDTO(config *v1alpha1.PipelineConfig) *PipelineConfigDTO {
	if config == nil {
		return nil
	}

	dto := &PipelineConfigDTO{
		ID:        config.Name,
		Name:      config.Name,
		RepoURL:   config.Spec.Repository,
		CreatedAt: config.CreationTimestamp.Time,
		Timeout:   3600, // Default 1 hour
	}

	// Copy branches
	if config.Spec.Branches != nil {
		dto.Branches = append([]string{}, config.Spec.Branches...)
	} else {
		dto.Branches = []string{"*"}
	}

	return dto
}
