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

package handlers

import (
	"context"
	"time"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

const (
	// k8sQueryTimeout is the default timeout for Kubernetes queries
	k8sQueryTimeout = 30 * time.Second
)

// FetchPipelineRunsForUser retrieves all pipeline runs for a specific user namespace
// Returns a slice of PipelineRunDTO objects or an empty slice if none found
func FetchPipelineRunsForUser(ctx context.Context, namespace string) []*dashboard.PipelineRunDTO {
	// Add timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(ctx, k8sQueryTimeout)
	defer cancel()

	var runs []*v1alpha1.PipelineRun

	if k8sClient != nil {
		if userRuns, err := k8sClient.ListPipelineRuns(ctx, namespace); err == nil && userRuns != nil {
			for i := 0; i < len(userRuns.Items); i++ {
				runs = append(runs, &userRuns.Items[i])
			}
		}
	}

	// Convert to DTOs
	dtos := make([]*dashboard.PipelineRunDTO, len(runs))
	for i, run := range runs {
		dtos[i] = dashboard.MapPipelineRunToDTO(run)
	}

	return dtos
}

// FetchPipelineConfigsForUser retrieves all pipeline configs for a specific user namespace
// Returns a slice of ProjectDTO objects or an empty slice if none found
func FetchPipelineConfigsForUser(ctx context.Context, namespace string) []*dashboard.ProjectDTO {
	// Add timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(ctx, k8sQueryTimeout)
	defer cancel()

	var projects []*dashboard.ProjectDTO

	if k8sClient != nil {
		configs, err := k8sClient.ListPipelineConfigs(ctx, namespace)
		if err == nil && configs != nil {
			for i := range configs.Items {
				projects = append(projects, mapPipelineConfigToProjectDTO(&configs.Items[i], namespace))
			}
		}
	}

	return projects
}

// FetchPipelineRunByID retrieves a specific pipeline run from a user's namespace
// Returns nil if not found or error occurs
func FetchPipelineRunByID(ctx context.Context, namespace, runID string) *v1alpha1.PipelineRun {
	// Add timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(ctx, k8sQueryTimeout)
	defer cancel()

	if k8sClient == nil {
		return nil
	}

	run, err := k8sClient.GetPipelineRun(ctx, namespace, runID)
	if err != nil {
		return nil
	}

	return run
}

// GenerateActivityFeed creates activity entries from recent pipeline runs
// Returns the most recent activities in reverse chronological order (newest first)
func GenerateActivityFeed(runs []*dashboard.PipelineRunDTO, limit int) []*dashboard.ActivityDTO {
	activities := make([]*dashboard.ActivityDTO, 0)

	// Sort runs by triggered_at (newest first)
	// Create a copy to avoid modifying the original slice
	runsCopy := make([]*dashboard.PipelineRunDTO, len(runs))
	copy(runsCopy, runs)

	// Simple bubble sort for small datasets
	for i := 0; i < len(runsCopy)-1; i++ {
		for j := 0; j < len(runsCopy)-i-1; j++ {
			if runsCopy[j].TriggeredAt.Before(runsCopy[j+1].TriggeredAt) {
				runsCopy[j], runsCopy[j+1] = runsCopy[j+1], runsCopy[j]
			}
		}
	}

	// Convert runs to activities
	for _, run := range runsCopy {
		if len(activities) >= limit {
			break
		}

		activity := &dashboard.ActivityDTO{
			ID:        run.ID,
			User:      run.Author,
			Timestamp: run.TriggeredAt,
			ProjectID: run.ProjectID,
			RunID:     run.ID,
		}

		// Determine activity type and message based on status
		switch run.Status {
		case "Succeeded":
			activity.Type = "build"
			activity.Message = "Build completed successfully"
		case "Failed":
			activity.Type = "error"
			activity.Message = "Build failed"
		case "Running":
			activity.Type = "build"
			activity.Message = "Build in progress"
		case "Pending":
			activity.Type = "build"
			activity.Message = "Build queued"
		default:
			activity.Type = "commit"
			activity.Message = "Pipeline triggered"
		}

		// Append branch and commit info to message
		if run.CommitSHA != "" {
			shortSHA := run.CommitSHA
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}
			activity.Message += " on " + run.Branch + " (" + shortSHA + ")"
		}

		activities = append(activities, activity)
	}

	return activities
}
