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

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/dashboard"
)

// FetchPipelineRunsForUser retrieves all pipeline runs for a specific user namespace
// Returns a slice of PipelineRunDTO objects or an empty slice if none found
func FetchPipelineRunsForUser(ctx context.Context, namespace string) []*dashboard.PipelineRunDTO {
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
	var projects []*dashboard.ProjectDTO

	if k8sClient != nil {
		configs, err := k8sClient.ListPipelineConfigs(ctx, namespace)
		if err == nil && configs != nil {
			for _, config := range configs.Items {
				projects = append(projects, mapPipelineConfigToProjectDTO(&config, namespace))
			}
		}
	}

	return projects
}

// FetchPipelineRunByID retrieves a specific pipeline run from a user's namespace
// Returns nil if not found or error occurs
func FetchPipelineRunByID(ctx context.Context, namespace, runID string) *v1alpha1.PipelineRun {
	if k8sClient == nil {
		return nil
	}

	run, err := k8sClient.GetPipelineRun(ctx, namespace, runID)
	if err != nil {
		return nil
	}

	return run
}
