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

package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

var (
	// Valid step name pattern: alphanumeric, dashes, underscores
	stepNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// Validate performs comprehensive validation on a PipelineConfig
func Validate(config *c8sv1alpha1.PipelineConfig) error {
	errors := &ValidationErrors{}

	// Validate repository URL format
	if err := validateRepositoryURL(config.Spec.Repository); err != nil {
		errors.Add("spec.repository", err.Error())
	}

	// Validate step names are unique and valid
	stepNames := make(map[string]bool)
	for i, step := range config.Spec.Steps {
		stepPrefix := fmt.Sprintf("spec.steps[%d]", i)

		// Validate step name format
		if !stepNamePattern.MatchString(step.Name) {
			errors.Add(fmt.Sprintf("%s.name", stepPrefix),
				"must contain only alphanumeric characters, dashes, and underscores")
		}

		// Check for duplicate names
		if stepNames[step.Name] {
			errors.Add(fmt.Sprintf("%s.name", stepPrefix),
				fmt.Sprintf("duplicate step name: %s", step.Name))
		}
		stepNames[step.Name] = true

		// Validate step-specific fields
		if err := validateStep(&step, stepNames, stepPrefix); err != nil {
			errors.Merge(err)
		}
	}

	// Validate no circular dependencies
	if err := validateNoCycles(config.Spec.Steps); err != nil {
		errors.Add("spec.steps", err.Error())
	}

	// Validate timeout format
	if config.Spec.Timeout != "" {
		if _, err := time.ParseDuration(config.Spec.Timeout); err != nil {
			errors.Add("spec.timeout",
				fmt.Sprintf("invalid duration format: %v", err))
		}
	}

	// Validate matrix strategy if present
	if config.Spec.Matrix != nil {
		if err := validateMatrix(config.Spec.Matrix); err != nil {
			errors.Merge(err)
		}
	}

	// Validate retry policy if present
	if config.Spec.RetryPolicy != nil {
		if err := validateRetryPolicy(config.Spec.RetryPolicy); err != nil {
			errors.Merge(err)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// validateRepositoryURL validates the repository URL format
func validateRepositoryURL(repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("repository URL is required")
	}

	// Parse URL
	u, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "git" && u.Scheme != "ssh" {
		return fmt.Errorf("unsupported URL scheme: %s (expected http, https, git, or ssh)", u.Scheme)
	}

	// Check host
	if u.Host == "" {
		return fmt.Errorf("repository URL must include a host")
	}

	return nil
}

// validateStep validates a single pipeline step
func validateStep(step *c8sv1alpha1.PipelineStep, existingSteps map[string]bool, prefix string) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate image format
	if step.Image == "" {
		errors.Add(fmt.Sprintf("%s.image", prefix), "image is required")
	}

	// Validate commands
	if len(step.Commands) == 0 {
		errors.Add(fmt.Sprintf("%s.commands", prefix), "at least one command is required")
	}

	// Validate dependencies reference existing steps
	for j, dep := range step.DependsOn {
		if dep == step.Name {
			errors.Add(fmt.Sprintf("%s.dependsOn[%d]", prefix, j),
				"step cannot depend on itself")
		}
		// Note: We check if dependency exists, but it might be defined later
		// The circular dependency check will catch invalid references
	}

	// Validate timeout format if specified
	if step.Timeout != "" {
		if _, err := time.ParseDuration(step.Timeout); err != nil {
			errors.Add(fmt.Sprintf("%s.timeout", prefix),
				fmt.Sprintf("invalid duration format: %v", err))
		}
	}

	// Validate resource values are valid Kubernetes quantities
	if step.Resources != nil {
		if step.Resources.CPU != "" {
			if _, err := resource.ParseQuantity(step.Resources.CPU); err != nil {
				errors.Add(fmt.Sprintf("%s.resources.cpu", prefix),
					fmt.Sprintf("invalid CPU quantity: %v", err))
			}
		}
		if step.Resources.Memory != "" {
			if _, err := resource.ParseQuantity(step.Resources.Memory); err != nil {
				errors.Add(fmt.Sprintf("%s.resources.memory", prefix),
					fmt.Sprintf("invalid memory quantity: %v", err))
			}
		}
	}

	// Validate conditional execution branch pattern if present
	if step.Conditional != nil && step.Conditional.Branch != "" {
		if _, err := regexp.Compile(step.Conditional.Branch); err != nil {
			errors.Add(fmt.Sprintf("%s.conditional.branch", prefix),
				fmt.Sprintf("invalid regex pattern: %v", err))
		}
	}

	return errors
}

// validateNoCycles checks for circular dependencies using DFS
func validateNoCycles(steps []c8sv1alpha1.PipelineStep) error {
	// Build adjacency list
	graph := make(map[string][]string)
	allSteps := make(map[string]bool)

	for _, step := range steps {
		graph[step.Name] = step.DependsOn
		allSteps[step.Name] = true
	}

	// Validate all dependencies exist
	for _, step := range steps {
		for _, dep := range step.DependsOn {
			if !allSteps[dep] {
				return fmt.Errorf("step %s depends on non-existent step: %s", step.Name, dep)
			}
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var dfs func(string, []string) error
	dfs = func(step string, path []string) error {
		visited[step] = true
		recursionStack[step] = true
		path = append(path, step)

		for _, dep := range graph[step] {
			if !visited[dep] {
				if err := dfs(dep, path); err != nil {
					return err
				}
			} else if recursionStack[dep] {
				// Found a cycle
				cycleStart := 0
				for i, s := range path {
					if s == dep {
						cycleStart = i
						break
					}
				}
				cyclePath := append(path[cycleStart:], dep)
				return fmt.Errorf("circular dependency detected: %s", strings.Join(cyclePath, " -> "))
			}
		}

		recursionStack[step] = false
		return nil
	}

	for _, step := range steps {
		if !visited[step.Name] {
			if err := dfs(step.Name, []string{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateMatrix validates matrix strategy configuration
func validateMatrix(matrix *c8sv1alpha1.MatrixStrategy) *ValidationErrors {
	errors := &ValidationErrors{}

	if len(matrix.Dimensions) == 0 {
		errors.Add("spec.matrix.dimensions", "at least one dimension is required")
		return errors
	}

	for dimName, values := range matrix.Dimensions {
		if len(values) == 0 {
			errors.Add(fmt.Sprintf("spec.matrix.dimensions.%s", dimName),
				"dimension must have at least one value")
		}
	}

	// Validate exclusion patterns reference valid dimensions
	for i, exclusion := range matrix.Exclude {
		for key := range exclusion {
			if _, exists := matrix.Dimensions[key]; !exists {
				errors.Add(fmt.Sprintf("spec.matrix.exclude[%d]", i),
					fmt.Sprintf("exclusion references undefined dimension: %s", key))
			}
		}
	}

	return errors
}

// validateRetryPolicy validates retry policy configuration
func validateRetryPolicy(policy *c8sv1alpha1.RetryPolicy) *ValidationErrors {
	errors := &ValidationErrors{}

	if policy.MaxRetries < 0 {
		errors.Add("spec.retryPolicy.maxRetries", "must be non-negative")
	}

	if policy.MaxRetries > 10 {
		errors.Add("spec.retryPolicy.maxRetries", "maximum allowed retries is 10")
	}

	if policy.BackoffSeconds < 0 {
		errors.Add("spec.retryPolicy.backoffSeconds", "must be non-negative")
	}

	return errors
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []*ValidationError
}

func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, &ValidationError{
		Field:   field,
		Message: message,
	})
}

func (ve *ValidationErrors) Merge(other *ValidationErrors) {
	if other != nil {
		ve.Errors = append(ve.Errors, other.Errors...)
	}
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(messages, "\n  - "))
}
