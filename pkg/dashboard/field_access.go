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

package dashboard

// FieldAccessFilter provides field-level access control for DTOs
// Different roles can see different fields in API responses
// This implements principle of least privilege for data exposure

// FilterProjectDTOForRole filters ProjectDTO fields based on user's role
// Viewers see less detail than editors/admins
func FilterProjectDTOForRole(project *ProjectDTO, role Role) *ProjectDTO {
	if project == nil {
		return nil
	}

	// Create a filtered copy
	filtered := &ProjectDTO{
		ID:        project.ID,
		Name:      project.Name,
		RepoURL:   project.RepoURL,
		CreatedAt: project.CreatedAt,
		RunCount:  project.RunCount,
	}

	// Role-based field visibility
	switch role {
	case RoleViewer:
		// Viewers see basic info only
		filtered.Namespace = project.Namespace
		filtered.Description = project.Description
		// Hide webhook URL from viewers
		// filtered.WebhookURL = "" (already empty)

	case RoleEditor:
		// Editors see most fields except internal details
		filtered.Namespace = project.Namespace
		filtered.Description = project.Description
		filtered.WebhookURL = project.WebhookURL
		filtered.LastRunAt = project.LastRunAt

	case RoleAdmin:
		// Admins see all fields
		filtered.Namespace = project.Namespace
		filtered.Description = project.Description
		filtered.WebhookURL = project.WebhookURL
		filtered.LastRunAt = project.LastRunAt
	}

	return filtered
}

// FilterProjectDTOsForRole filters multiple ProjectDTOs based on role
func FilterProjectDTOsForRole(projects []*ProjectDTO, role Role) []*ProjectDTO {
	if projects == nil {
		return nil
	}

	filtered := make([]*ProjectDTO, 0, len(projects))
	for _, project := range projects {
		if project != nil {
			filtered = append(filtered, FilterProjectDTOForRole(project, role))
		}
	}
	return filtered
}

// FilterPipelineRunDTOForRole filters PipelineRunDTO fields based on role
func FilterPipelineRunDTOForRole(run *PipelineRunDTO, role Role) *PipelineRunDTO {
	if run == nil {
		return nil
	}

	// Create a filtered copy
	filtered := &PipelineRunDTO{
		ID:            run.ID,
		ProjectID:     run.ProjectID,
		Name:          run.Name,
		Status:        run.Status,
		CommitSHA:     run.CommitSHA,
		Branch:        run.Branch,
		Author:        run.Author,
		TriggeredAt:   run.TriggeredAt,
		StartedAt:     run.StartedAt,
		CompletedAt:   run.CompletedAt,
		DurationSeconds: run.DurationSeconds,
		StepCount:     run.StepCount,
		SuccessCount:  run.SuccessCount,
		FailureCount:  run.FailureCount,
		ArtifactCount: run.ArtifactCount,
	}

	// Role-based field visibility
	switch role {
	case RoleViewer:
		// Viewers see basic execution info
		filtered.Branch = run.Branch
		// Hide author email from viewers
		filtered.AuthorEmail = ""
		filtered.TriggerSource = run.TriggerSource

	case RoleEditor:
		// Editors see most fields
		filtered.Branch = run.Branch
		filtered.AuthorEmail = run.AuthorEmail
		filtered.TriggerSource = run.TriggerSource

	case RoleAdmin:
		// Admins see all fields
		filtered.Branch = run.Branch
		filtered.AuthorEmail = run.AuthorEmail
		filtered.TriggerSource = run.TriggerSource
	}

	return filtered
}

// FilterPipelineRunDTOsForRole filters multiple PipelineRunDTOs
func FilterPipelineRunDTOsForRole(runs []*PipelineRunDTO, role Role) []*PipelineRunDTO {
	if runs == nil {
		return nil
	}

	filtered := make([]*PipelineRunDTO, 0, len(runs))
	for _, run := range runs {
		if run != nil {
			filtered = append(filtered, FilterPipelineRunDTOForRole(run, role))
		}
	}
	return filtered
}

// FilterStepDTOForRole filters StepDTO fields based on role
func FilterStepDTOForRole(step *StepDTO, role Role) *StepDTO {
	if step == nil {
		return nil
	}

	// Create a filtered copy
	filtered := &StepDTO{
		ID:              step.ID,
		Name:            step.Name,
		Status:          step.Status,
		Image:           step.Image,
		Commands:        step.Commands,
		StartedAt:       step.StartedAt,
		CompletedAt:     step.CompletedAt,
		DurationSeconds: step.DurationSeconds,
		ExitCode:        step.ExitCode,
		DependsOn:       step.DependsOn,
		LogURL:          step.LogURL,
	}

	// Role-based resource visibility
	switch role {
	case RoleViewer:
		// Viewers see step execution details
		filtered.CPURequest = step.CPURequest
		filtered.MemoryRequest = step.MemoryRequest

	case RoleEditor:
		// Editors see all details
		filtered.CPURequest = step.CPURequest
		filtered.MemoryRequest = step.MemoryRequest

	case RoleAdmin:
		// Admins see all fields
		filtered.CPURequest = step.CPURequest
		filtered.MemoryRequest = step.MemoryRequest
	}

	return filtered
}

// FilterStepDTOsForRole filters multiple StepDTOs
func FilterStepDTOsForRole(steps []*StepDTO, role Role) []*StepDTO {
	if steps == nil {
		return nil
	}

	filtered := make([]*StepDTO, 0, len(steps))
	for _, step := range steps {
		if step != nil {
			filtered = append(filtered, FilterStepDTOForRole(step, role))
		}
	}
	return filtered
}

// FilterArtifactDTOForRole filters ArtifactDTO fields based on role
func FilterArtifactDTOForRole(artifact *ArtifactDTO, role Role) *ArtifactDTO {
	if artifact == nil {
		return nil
	}

	// Create a filtered copy
	filtered := &ArtifactDTO{
		ID:        artifact.ID,
		Name:      artifact.Name,
		Type:      artifact.Type,
		MimeType:  artifact.MimeType,
		SizeBytes: artifact.SizeBytes,
		CreatedAt: artifact.CreatedAt,
	}

	// Role-based URL visibility
	switch role {
	case RoleViewer:
		// Viewers cannot directly access artifact URLs
		filtered.URL = ""

	case RoleEditor:
		// Editors can access artifact URLs
		filtered.URL = artifact.URL

	case RoleAdmin:
		// Admins can access artifact URLs
		filtered.URL = artifact.URL
	}

	return filtered
}

// FilterArtifactDTOsForRole filters multiple ArtifactDTOs
func FilterArtifactDTOsForRole(artifacts []*ArtifactDTO, role Role) []*ArtifactDTO {
	if artifacts == nil {
		return nil
	}

	filtered := make([]*ArtifactDTO, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact != nil {
			filtered = append(filtered, FilterArtifactDTOForRole(artifact, role))
		}
	}
	return filtered
}

// FilterLogStreamDTOForRole filters LogStreamDTO fields based on role
// Log visibility depends on role - all roles can see logs but sensitive data might be redacted
func FilterLogStreamDTOForRole(log *LogStreamDTO, role Role) *LogStreamDTO {
	if log == nil {
		return nil
	}

	// Create a filtered copy
	filtered := &LogStreamDTO{
		StepID:    log.StepID,
		Content:   log.Content,
		Timestamp: log.Timestamp,
		IsError:   log.IsError,
	}

	// All roles can see logs (access control is at handler level)
	// Content filtering would happen here if needed for sensitive data redaction
	switch role {
	case RoleViewer:
		// Viewers see full logs
		filtered.Content = log.Content

	case RoleEditor:
		// Editors see full logs
		filtered.Content = log.Content

	case RoleAdmin:
		// Admins see full logs
		filtered.Content = log.Content
	}

	return filtered
}
