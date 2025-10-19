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

package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/parser"
)

// TestValidMinimalPipelineYAML verifies valid minimal pipeline YAML parses correctly
func TestValidMinimalPipelineYAML(t *testing.T) {
	yaml := `
version: v1alpha1
name: test-pipeline
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...
`

	spec, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, spec)

	assert.Len(t, spec.Steps, 1)
	assert.Equal(t, "test", spec.Steps[0].Name)
	assert.Equal(t, "golang:1.21", spec.Steps[0].Image)
	assert.Len(t, spec.Steps[0].Commands, 1)
	assert.Equal(t, "go test ./...", spec.Steps[0].Commands[0])

	// Check default timeout
	assert.Equal(t, "1h", spec.Timeout)
}

// TestValidComplexPipelineYAML verifies complex pipeline with all features
func TestValidComplexPipelineYAML(t *testing.T) {
	yaml := `
version: v1alpha1
name: complex-pipeline
timeout: 2h
steps:
  - name: lint
    image: golangci/golangci-lint:latest
    commands:
      - golangci-lint run
    timeout: 10m
    resources:
      cpu: 1000m
      memory: 2Gi
  - name: test
    image: golang:1.21
    commands:
      - go test -v ./...
      - go test -race ./...
  - name: build
    image: golang:1.21
    commands:
      - go build -o bin/app
    dependsOn:
      - lint
      - test
    resources:
      cpu: 2000m
      memory: 4Gi
matrix:
  dimensions:
    os:
      - ubuntu
      - alpine
    go_version:
      - "1.21"
      - "1.22"
  exclude:
    - os: alpine
      go_version: "1.21"
retryPolicy:
  maxRetries: 3
  backoffSeconds: 30
`

	spec, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify basic fields
	assert.Len(t, spec.Steps, 3)
	assert.Equal(t, "2h", spec.Timeout)

	// Verify lint step
	assert.Equal(t, "lint", spec.Steps[0].Name)
	assert.Equal(t, "10m", spec.Steps[0].Timeout)
	assert.Equal(t, "1000m", spec.Steps[0].Resources.CPU)
	assert.Equal(t, "2Gi", spec.Steps[0].Resources.Memory)

	// Verify test step
	assert.Equal(t, "test", spec.Steps[1].Name)
	assert.Len(t, spec.Steps[1].Commands, 2)

	// Verify build step with dependencies
	assert.Equal(t, "build", spec.Steps[2].Name)
	assert.Len(t, spec.Steps[2].DependsOn, 2)
	assert.Contains(t, spec.Steps[2].DependsOn, "lint")
	assert.Contains(t, spec.Steps[2].DependsOn, "test")

	// Verify matrix
	require.NotNil(t, spec.Matrix)
	assert.Len(t, spec.Matrix.Dimensions, 2)
	assert.Contains(t, spec.Matrix.Dimensions, "os")
	assert.Contains(t, spec.Matrix.Dimensions, "go_version")
	assert.Len(t, spec.Matrix.Exclude, 1)

	// Verify retry policy
	require.NotNil(t, spec.RetryPolicy)
	assert.Equal(t, 3, spec.RetryPolicy.MaxRetries)
	assert.Equal(t, 30, spec.RetryPolicy.BackoffSeconds)
}

// TestInvalidYAML verifies invalid YAML returns error
func TestInvalidYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "malformed YAML",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands
      - go test
`,
		},
		{
			name: "unclosed list",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parser.Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Nil(t, spec)
			assert.Contains(t, err.Error(), "failed to parse YAML")
		})
	}
}

// TestMissingRequiredFields verifies validation errors for missing required fields
func TestMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		errorMsg string
	}{
		{
			name: "missing version",
			yaml: `
name: test
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test
`,
			errorMsg: "version field is required",
		},
		{
			name: "missing steps",
			yaml: `
version: v1alpha1
name: test
steps: []
`,
			errorMsg: "at least one step is required",
		},
		{
			name: "missing step name",
			yaml: `
version: v1alpha1
name: test
steps:
  - image: golang:1.21
    commands:
      - go test
`,
			errorMsg: "name is required",
		},
		{
			name: "missing step image",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: test
    commands:
      - go test
`,
			errorMsg: "image is required",
		},
		{
			name: "missing step commands",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands: []
`,
			errorMsg: "at least one command is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parser.Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Nil(t, spec)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// TestInvalidStepNames verifies validation of step names with special characters
func TestInvalidStepNames(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Repository: "https://github.com/org/repo",
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:     "test step", // spaces not allowed
					Image:    "golang:1.21",
					Commands: []string{"go test"},
				},
			},
		},
	}

	err := parser.Validate(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain only alphanumeric characters")
}

// TestValidStepNames verifies valid step name patterns
func TestValidStepNames(t *testing.T) {
	validNames := []string{
		"test",
		"test-unit",
		"test_integration",
		"test123",
		"Test-Build_Deploy",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     name,
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
				},
			}

			err := parser.Validate(config)
			assert.NoError(t, err, "step name %s should be valid", name)
		})
	}
}

// TestValidResourceValues verifies valid Kubernetes resource quantity parsing
func TestValidResourceValues(t *testing.T) {
	yaml := `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test
    resources:
      cpu: 1000m
      memory: 2Gi
`

	spec, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, spec)

	config := &c8sv1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: *spec,
	}
	config.Spec.Repository = "https://github.com/org/repo"

	err = parser.Validate(config)
	assert.NoError(t, err)
}

// TestInvalidResourceValues verifies validation of invalid resource quantities
func TestInvalidResourceValues(t *testing.T) {
	tests := []struct {
		name     string
		cpu      string
		memory   string
		errorMsg string
	}{
		{
			name:     "invalid CPU format",
			cpu:      "invalid",
			memory:   "2Gi",
			errorMsg: "invalid CPU quantity",
		},
		{
			name:     "invalid memory format",
			cpu:      "1000m",
			memory:   "invalid",
			errorMsg: "invalid memory quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
							Resources: &c8sv1alpha1.ResourceRequirements{
								CPU:    tt.cpu,
								Memory: tt.memory,
							},
						},
					},
				},
			}

			err := parser.Validate(config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// TestInvalidTimeoutFormat verifies timeout format validation
func TestInvalidTimeoutFormat(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{
			name:    "invalid format - no unit",
			timeout: "123",
		},
		{
			name:    "invalid format - bad unit",
			timeout: "30x",
		},
		{
			name:    "invalid format - empty",
			timeout: "m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Timeout:    tt.timeout,
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
				},
			}

			err := parser.Validate(config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid duration format")
		})
	}
}

// TestValidTimeoutFormats verifies valid timeout formats
func TestValidTimeoutFormats(t *testing.T) {
	validTimeouts := []string{
		"30s",
		"5m",
		"2h",
		"1h30m",
		"90s",
	}

	for _, timeout := range validTimeouts {
		t.Run(timeout, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Timeout:    timeout,
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
				},
			}

			err := parser.Validate(config)
			assert.NoError(t, err, "timeout %s should be valid", timeout)
		})
	}
}

// TestDuplicateStepNames verifies duplicate step name detection
func TestDuplicateStepNames(t *testing.T) {
	yaml := `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test
  - name: test
    image: golang:1.22
    commands:
      - go test
`

	spec, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.Contains(t, err.Error(), "duplicate step name")
}

// TestNonExistentDependency verifies validation of non-existent dependencies
func TestNonExistentDependency(t *testing.T) {
	yaml := `
version: v1alpha1
name: test
steps:
  - name: build
    image: golang:1.21
    commands:
      - go build
    dependsOn:
      - test
`

	spec, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.Contains(t, err.Error(), "dependency test not found")
}

// TestCircularDependencyDetection verifies circular dependency detection
func TestCircularDependencyDetection(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "simple cycle A→B→A",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: stepA
    image: alpine
    commands:
      - echo A
    dependsOn:
      - stepB
  - name: stepB
    image: alpine
    commands:
      - echo B
    dependsOn:
      - stepA
`,
		},
		{
			name: "three-step cycle A→B→C→A",
			yaml: `
version: v1alpha1
name: test
steps:
  - name: stepA
    image: alpine
    commands:
      - echo A
    dependsOn:
      - stepC
  - name: stepB
    image: alpine
    commands:
      - echo B
    dependsOn:
      - stepA
  - name: stepC
    image: alpine
    commands:
      - echo C
    dependsOn:
      - stepB
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parser.Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Nil(t, spec)
			assert.Contains(t, err.Error(), "circular dependency")
		})
	}
}

// TestValidDependencies verifies valid dependency chains
func TestValidDependencies(t *testing.T) {
	yaml := `
version: v1alpha1
name: test
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test
  - name: build
    image: golang:1.21
    commands:
      - go build
    dependsOn:
      - test
  - name: deploy
    image: alpine
    commands:
      - echo deploy
    dependsOn:
      - build
`

	spec, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, spec)

	config := &c8sv1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: *spec,
	}
	config.Spec.Repository = "https://github.com/org/repo"

	err = parser.Validate(config)
	assert.NoError(t, err)
}

// TestRepositoryURLValidation verifies repository URL format validation
func TestRepositoryURLValidation(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		shouldErr bool
		errorMsg  string
	}{
		{
			name:      "valid HTTPS URL",
			repoURL:   "https://github.com/org/repo",
			shouldErr: false,
		},
		{
			name:      "valid HTTP URL",
			repoURL:   "http://gitlab.com/org/repo",
			shouldErr: false,
		},
		{
			name:      "valid git URL",
			repoURL:   "git://github.com/org/repo.git",
			shouldErr: false,
		},
		{
			name:      "valid SSH URL",
			repoURL:   "ssh://git@github.com/org/repo.git",
			shouldErr: false,
		},
		{
			name:      "empty URL",
			repoURL:   "",
			shouldErr: true,
			errorMsg:  "repository URL is required",
		},
		{
			name:      "invalid scheme",
			repoURL:   "ftp://github.com/org/repo",
			shouldErr: true,
			errorMsg:  "unsupported URL scheme",
		},
		{
			name:      "missing host",
			repoURL:   "https:///org/repo",
			shouldErr: true,
			errorMsg:  "must include a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: tt.repoURL,
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
				},
			}

			err := parser.Validate(config)
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMatrixValidation verifies matrix strategy validation
func TestMatrixValidation(t *testing.T) {
	tests := []struct {
		name      string
		matrix    *c8sv1alpha1.MatrixStrategy
		shouldErr bool
		errorMsg  string
	}{
		{
			name: "valid matrix",
			matrix: &c8sv1alpha1.MatrixStrategy{
				Dimensions: map[string][]string{
					"os":      {"ubuntu", "alpine"},
					"version": {"1.21", "1.22"},
				},
			},
			shouldErr: false,
		},
		{
			name: "empty dimensions",
			matrix: &c8sv1alpha1.MatrixStrategy{
				Dimensions: map[string][]string{},
			},
			shouldErr: true,
			errorMsg:  "at least one dimension is required",
		},
		{
			name: "dimension with no values",
			matrix: &c8sv1alpha1.MatrixStrategy{
				Dimensions: map[string][]string{
					"os": {},
				},
			},
			shouldErr: true,
			errorMsg:  "dimension must have at least one value",
		},
		{
			name: "exclusion references undefined dimension",
			matrix: &c8sv1alpha1.MatrixStrategy{
				Dimensions: map[string][]string{
					"os": {"ubuntu", "alpine"},
				},
				Exclude: []map[string]string{
					{"undefined": "value"},
				},
			},
			shouldErr: true,
			errorMsg:  "exclusion references undefined dimension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
					Matrix: tt.matrix,
				},
			}

			err := parser.Validate(config)
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRetryPolicyValidation verifies retry policy validation
func TestRetryPolicyValidation(t *testing.T) {
	tests := []struct {
		name      string
		policy    *c8sv1alpha1.RetryPolicy
		shouldErr bool
		errorMsg  string
	}{
		{
			name: "valid retry policy",
			policy: &c8sv1alpha1.RetryPolicy{
				MaxRetries:     3,
				BackoffSeconds: 30,
			},
			shouldErr: false,
		},
		{
			name: "negative max retries",
			policy: &c8sv1alpha1.RetryPolicy{
				MaxRetries:     -1,
				BackoffSeconds: 30,
			},
			shouldErr: true,
			errorMsg:  "must be non-negative",
		},
		{
			name: "too many retries",
			policy: &c8sv1alpha1.RetryPolicy{
				MaxRetries:     15,
				BackoffSeconds: 30,
			},
			shouldErr: true,
			errorMsg:  "maximum allowed retries is 10",
		},
		{
			name: "negative backoff",
			policy: &c8sv1alpha1.RetryPolicy{
				MaxRetries:     3,
				BackoffSeconds: -10,
			},
			shouldErr: true,
			errorMsg:  "must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &c8sv1alpha1.PipelineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Spec: c8sv1alpha1.PipelineConfigSpec{
					Repository: "https://github.com/org/repo",
					Steps: []c8sv1alpha1.PipelineStep{
						{
							Name:     "test",
							Image:    "golang:1.21",
							Commands: []string{"go test"},
						},
					},
					RetryPolicy: tt.policy,
				},
			}

			err := parser.Validate(config)
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSelfReferencingStep verifies detection of self-referencing steps
func TestSelfReferencingStep(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Repository: "https://github.com/org/repo",
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:      "test",
					Image:     "golang:1.21",
					Commands:  []string{"go test"},
					DependsOn: []string{"test"}, // self-reference
				},
			},
		},
	}

	err := parser.Validate(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "step cannot depend on itself")
}
