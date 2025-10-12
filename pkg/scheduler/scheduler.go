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

package scheduler

import (
	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// Schedule represents an execution plan for pipeline steps
type Schedule struct {
	// Layers contains ordered layers of steps
	// Steps within a layer can execute in parallel
	// Layers must execute sequentially
	Layers []Layer

	// DAG is the underlying dependency graph
	DAG *DAG
}

// Layer represents a set of steps that can execute in parallel
type Layer struct {
	// Steps in this layer (all have their dependencies satisfied)
	Steps []*c8sv1alpha1.PipelineStep

	// StepNames for quick lookup
	StepNames []string
}

// Schedule creates an execution schedule from a PipelineConfig
func Schedule(config *c8sv1alpha1.PipelineConfig) (*Schedule, error) {
	// Build DAG from steps
	dag, err := BuildDAG(config.Spec.Steps)
	if err != nil {
		return nil, err
	}

	// Get topological ordering in layers
	layerNames, err := dag.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Convert layer names to layer structs with step definitions
	var layers []Layer
	for _, stepNames := range layerNames {
		layer := Layer{
			StepNames: stepNames,
			Steps:     make([]*c8sv1alpha1.PipelineStep, 0, len(stepNames)),
		}

		for _, name := range stepNames {
			if step, exists := dag.GetStep(name); exists {
				layer.Steps = append(layer.Steps, step)
			}
		}

		layers = append(layers, layer)
	}

	return &Schedule{
		Layers: layers,
		DAG:    dag,
	}, nil
}

// GetReadySteps returns steps that are ready to execute given completed steps
// A step is ready if all its dependencies have completed successfully
func (s *Schedule) GetReadySteps(completedSteps map[string]bool) []*c8sv1alpha1.PipelineStep {
	var ready []*c8sv1alpha1.PipelineStep

	for _, layer := range s.Layers {
		for _, step := range layer.Steps {
			// Skip if already completed
			if completedSteps[step.Name] {
				continue
			}

			// Check if all dependencies are completed
			allDepsCompleted := true
			for _, dep := range step.DependsOn {
				if !completedSteps[dep] {
					allDepsCompleted = false
					break
				}
			}

			if allDepsCompleted {
				ready = append(ready, step)
			}
		}
	}

	return ready
}

// GetLayer returns the layer index for a given step
// Returns -1 if step not found
func (s *Schedule) GetLayer(stepName string) int {
	for i, layer := range s.Layers {
		for _, name := range layer.StepNames {
			if name == stepName {
				return i
			}
		}
	}
	return -1
}

// TotalSteps returns the total number of steps in the schedule
func (s *Schedule) TotalSteps() int {
	return s.DAG.Size()
}

// LayerCount returns the number of execution layers
func (s *Schedule) LayerCount() int {
	return len(s.Layers)
}

// CanExecuteInParallel returns true if two steps can execute in parallel
// (i.e., they are in the same layer)
func (s *Schedule) CanExecuteInParallel(step1, step2 string) bool {
	layer1 := s.GetLayer(step1)
	layer2 := s.GetLayer(step2)
	return layer1 == layer2 && layer1 != -1
}

// GetStepDependencies returns the direct dependencies for a step
func (s *Schedule) GetStepDependencies(stepName string) []string {
	return s.DAG.GetDependencies(stepName)
}

// GetStepDependents returns the steps that depend on this step
func (s *Schedule) GetStepDependents(stepName string) []string {
	return s.DAG.GetDependents(stepName)
}
