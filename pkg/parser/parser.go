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

// Package parser provides functions to parse and validate pipeline YAML files
package parser

import (
	"fmt"

	"gopkg.in/yaml.v3"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// PipelineYAML represents the structure of a .c8s.yaml file
type PipelineYAML struct {
	Version string                       `yaml:"version"`
	Name    string                       `yaml:"name"`
	Steps   []c8sv1alpha1.PipelineStep   `yaml:"steps"`
	Timeout string                       `yaml:"timeout,omitempty"`
	Matrix  *c8sv1alpha1.MatrixStrategy  `yaml:"matrix,omitempty"`
	Retry   *c8sv1alpha1.RetryPolicy     `yaml:"retryPolicy,omitempty"`
}

// Parse parses pipeline YAML content into a PipelineConfig spec
func Parse(yamlContent []byte) (*c8sv1alpha1.PipelineConfigSpec, error) {
	var pipeline PipelineYAML

	if err := yaml.Unmarshal(yamlContent, &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&pipeline); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	spec := &c8sv1alpha1.PipelineConfigSpec{
		Steps:       pipeline.Steps,
		Timeout:     pipeline.Timeout,
		Matrix:      pipeline.Matrix,
		RetryPolicy: pipeline.Retry,
	}

	// Set defaults
	if spec.Timeout == "" {
		spec.Timeout = "1h"
	}

	return spec, nil
}

// validate validates the pipeline structure
func validate(pipeline *PipelineYAML) error {
	// Check version
	if pipeline.Version == "" {
		return fmt.Errorf("version field is required")
	}
	if pipeline.Version != "v1alpha1" {
		return fmt.Errorf("unsupported version: %s (expected v1alpha1)", pipeline.Version)
	}

	// Check steps
	if len(pipeline.Steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}

	// Validate step names are unique
	stepNames := make(map[string]bool)
	for i, step := range pipeline.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}
		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true

		// Validate step fields
		if step.Image == "" {
			return fmt.Errorf("step %s: image is required", step.Name)
		}
		if len(step.Commands) == 0 {
			return fmt.Errorf("step %s: at least one command is required", step.Name)
		}

		// Validate dependencies reference existing steps
		for _, dep := range step.DependsOn {
			if !stepNames[dep] {
				return fmt.Errorf("step %s: dependency %s not found", step.Name, dep)
			}
		}
	}

	// Check for circular dependencies
	if err := checkCircularDependencies(pipeline.Steps); err != nil {
		return err
	}

	// Validate matrix if present
	if pipeline.Matrix != nil {
		if len(pipeline.Matrix.Dimensions) == 0 {
			return fmt.Errorf("matrix must have at least one dimension")
		}
		for dim, values := range pipeline.Matrix.Dimensions {
			if len(values) == 0 {
				return fmt.Errorf("matrix dimension %s must have at least one value", dim)
			}
		}
	}

	return nil
}

// checkCircularDependencies detects circular dependencies in the step graph
func checkCircularDependencies(steps []c8sv1alpha1.PipelineStep) error {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, step := range steps {
		graph[step.Name] = step.DependsOn
	}

	// Check each step for cycles using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(step string) bool {
		visited[step] = true
		recursionStack[step] = true

		for _, dep := range graph[step] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recursionStack[dep] {
				return true
			}
		}

		recursionStack[step] = false
		return false
	}

	for _, step := range steps {
		if !visited[step.Name] {
			if hasCycle(step.Name) {
				return fmt.Errorf("circular dependency detected involving step: %s", step.Name)
			}
		}
	}

	return nil
}

// ParseFile is a convenience function to parse a pipeline file
// This will be used by the CLI and webhook service
func ParseFile(filename string) (*c8sv1alpha1.PipelineConfigSpec, error) {
	// TODO: Implement file reading in Phase 3
	// For now, this is a placeholder
	return nil, fmt.Errorf("not implemented: use Parse() with file contents")
}
