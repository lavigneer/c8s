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

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/scheduler"
	"github.com/org/c8s/pkg/types"
)

// TestEmptySteps verifies that empty steps returns empty schedule
func TestEmptySteps(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)
	require.NotNil(t, dag)

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	assert.Empty(t, layers)
	assert.Equal(t, 0, dag.Size())
}

// TestSingleStepNoDependencies verifies single step with no dependencies
func TestSingleStepNoDependencies(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "build",
			Image: "golang:1.21",
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)
	require.NotNil(t, dag)
	assert.Equal(t, 1, dag.Size())

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 1)
	require.Len(t, layers[0], 1)
	assert.Equal(t, "build", layers[0][0])
}

// TestLinearDependencies verifies linear dependencies (A→B→C)
func TestLinearDependencies(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "test",
			Image: "golang:1.21",
		},
		{
			Name:      "build",
			Image:     "golang:1.21",
			DependsOn: []string{"test"},
		},
		{
			Name:      "deploy",
			Image:     "alpine:latest",
			DependsOn: []string{"build"},
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)
	assert.Equal(t, 3, dag.Size())

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 3)

	// Layer 0: test
	require.Len(t, layers[0], 1)
	assert.Equal(t, "test", layers[0][0])

	// Layer 1: build
	require.Len(t, layers[1], 1)
	assert.Equal(t, "build", layers[1][0])

	// Layer 2: deploy
	require.Len(t, layers[2], 1)
	assert.Equal(t, "deploy", layers[2][0])
}

// TestParallelSteps verifies parallel steps (A, B both depend on nothing)
func TestParallelSteps(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "lint",
			Image: "golangci/golangci-lint:latest",
		},
		{
			Name:  "test",
			Image: "golang:1.21",
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)
	assert.Equal(t, 2, dag.Size())

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 1)
	require.Len(t, layers[0], 2)

	// Both steps should be in the same layer (order doesn't matter)
	stepNames := map[string]bool{
		layers[0][0]: true,
		layers[0][1]: true,
	}
	assert.True(t, stepNames["lint"])
	assert.True(t, stepNames["test"])
}

// TestComplexDAGMixedParallelSequential verifies complex DAG with mixed parallel/sequential
// Topology:
//       lint    test
//         \    /   \
//          build    integration
//            \      /
//             deploy
func TestComplexDAGMixedParallelSequential(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "lint",
			Image: "golangci/golangci-lint:latest",
		},
		{
			Name:  "test",
			Image: "golang:1.21",
		},
		{
			Name:      "build",
			Image:     "golang:1.21",
			DependsOn: []string{"lint", "test"},
		},
		{
			Name:      "integration",
			Image:     "golang:1.21",
			DependsOn: []string{"test"},
		},
		{
			Name:      "deploy",
			Image:     "alpine:latest",
			DependsOn: []string{"build", "integration"},
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)
	assert.Equal(t, 5, dag.Size())

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 3)

	// Layer 0: lint and test (parallel)
	require.Len(t, layers[0], 2)
	layer0Names := map[string]bool{
		layers[0][0]: true,
		layers[0][1]: true,
	}
	assert.True(t, layer0Names["lint"])
	assert.True(t, layer0Names["test"])

	// Layer 1: build and integration (parallel, both depend on layer 0)
	require.Len(t, layers[1], 2)
	layer1Names := map[string]bool{
		layers[1][0]: true,
		layers[1][1]: true,
	}
	assert.True(t, layer1Names["build"])
	assert.True(t, layer1Names["integration"])

	// Layer 2: deploy (depends on layer 1)
	require.Len(t, layers[2], 1)
	assert.Equal(t, "deploy", layers[2][0])
}

// TestCircularDependencyDetection verifies circular dependency detection
func TestCircularDependencyDetection(t *testing.T) {
	tests := []struct {
		name  string
		steps []c8sv1alpha1.PipelineStep
	}{
		{
			name: "simple cycle A→B→A",
			steps: []c8sv1alpha1.PipelineStep{
				{
					Name:      "stepA",
					Image:     "alpine",
					DependsOn: []string{"stepB"},
				},
				{
					Name:      "stepB",
					Image:     "alpine",
					DependsOn: []string{"stepA"},
				},
			},
		},
		{
			name: "three-step cycle A→B→C→A",
			steps: []c8sv1alpha1.PipelineStep{
				{
					Name:      "stepA",
					Image:     "alpine",
					DependsOn: []string{"stepC"},
				},
				{
					Name:      "stepB",
					Image:     "alpine",
					DependsOn: []string{"stepA"},
				},
				{
					Name:      "stepC",
					Image:     "alpine",
					DependsOn: []string{"stepB"},
				},
			},
		},
		{
			name: "self-referencing step A→A",
			steps: []c8sv1alpha1.PipelineStep{
				{
					Name:      "stepA",
					Image:     "alpine",
					DependsOn: []string{"stepA"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dag, err := scheduler.BuildDAG(tt.steps)

			// Should either fail during DAG construction or during topological sort
			if err != nil {
				assert.ErrorIs(t, err, types.ErrInvalidDependencyGraph)
			} else {
				require.NotNil(t, dag)
				_, err = dag.TopologicalSort()
				assert.ErrorIs(t, err, types.ErrInvalidDependencyGraph)
			}
		})
	}
}

// TestNonExistentDependencyReference verifies error when dependency references non-existent step
func TestNonExistentDependencyReference(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "build",
			Image: "golang:1.21",
		},
		{
			Name:      "deploy",
			Image:     "alpine",
			DependsOn: []string{"build", "test"}, // "test" doesn't exist
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrStepNotFound)
	assert.Nil(t, dag)
	assert.Contains(t, err.Error(), "test")
}

// TestDuplicateStepNames verifies error on duplicate step names
func TestDuplicateStepNames(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "build",
			Image: "golang:1.21",
		},
		{
			Name:  "build", // duplicate name
			Image: "alpine",
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate step name")
	assert.Nil(t, dag)
}

// TestDAGGetters verifies DAG getter methods
func TestDAGGetters(t *testing.T) {
	steps := []c8sv1alpha1.PipelineStep{
		{
			Name:  "test",
			Image: "golang:1.21",
		},
		{
			Name:      "build",
			Image:     "golang:1.21",
			DependsOn: []string{"test"},
		},
		{
			Name:      "deploy",
			Image:     "alpine",
			DependsOn: []string{"build"},
		},
	}

	dag, err := scheduler.BuildDAG(steps)
	require.NoError(t, err)

	// Test GetStep
	step, exists := dag.GetStep("build")
	assert.True(t, exists)
	assert.Equal(t, "build", step.Name)

	step, exists = dag.GetStep("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, step)

	// Test GetDependencies
	deps := dag.GetDependencies("build")
	require.Len(t, deps, 1)
	assert.Equal(t, "test", deps[0])

	deps = dag.GetDependencies("test")
	assert.Empty(t, deps)

	// Test GetDependents
	dependents := dag.GetDependents("test")
	require.Len(t, dependents, 1)
	assert.Equal(t, "build", dependents[0])

	dependents = dag.GetDependents("deploy")
	assert.Empty(t, dependents)
}

// TestBuildSchedule verifies full schedule building
func TestBuildSchedule(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "lint",
					Image: "golangci/golangci-lint:latest",
				},
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"lint", "test"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)
	require.NotNil(t, schedule)

	// Verify layer count
	assert.Equal(t, 2, schedule.LayerCount())

	// Verify total steps
	assert.Equal(t, 3, schedule.TotalSteps())

	// Verify layers
	require.Len(t, schedule.Layers, 2)

	// Layer 0: lint and test
	assert.Len(t, schedule.Layers[0].Steps, 2)
	assert.Len(t, schedule.Layers[0].StepNames, 2)

	// Layer 1: build
	assert.Len(t, schedule.Layers[1].Steps, 1)
	assert.Equal(t, "build", schedule.Layers[1].StepNames[0])
}

// TestScheduleGetReadySteps verifies GetReadySteps functionality
func TestScheduleGetReadySteps(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"test"},
				},
				{
					Name:      "deploy",
					Image:     "alpine",
					DependsOn: []string{"build"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)

	// Initially, only "test" should be ready (no dependencies)
	completed := make(map[string]bool)
	ready := schedule.GetReadySteps(completed)
	require.Len(t, ready, 1)
	assert.Equal(t, "test", ready[0].Name)

	// After test completes, build should be ready
	completed["test"] = true
	ready = schedule.GetReadySteps(completed)
	require.Len(t, ready, 1)
	assert.Equal(t, "build", ready[0].Name)

	// After build completes, deploy should be ready
	completed["build"] = true
	ready = schedule.GetReadySteps(completed)
	require.Len(t, ready, 1)
	assert.Equal(t, "deploy", ready[0].Name)

	// After all complete, nothing is ready
	completed["deploy"] = true
	ready = schedule.GetReadySteps(completed)
	assert.Empty(t, ready)
}

// TestScheduleCanExecuteInParallel verifies parallel execution detection
func TestScheduleCanExecuteInParallel(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "lint",
					Image: "golangci/golangci-lint:latest",
				},
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"lint", "test"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)

	// lint and test should be able to run in parallel (same layer)
	assert.True(t, schedule.CanExecuteInParallel("lint", "test"))

	// lint and build should NOT be able to run in parallel (different layers)
	assert.False(t, schedule.CanExecuteInParallel("lint", "build"))
	assert.False(t, schedule.CanExecuteInParallel("test", "build"))

	// Non-existent steps should return false
	assert.False(t, schedule.CanExecuteInParallel("nonexistent", "test"))
}

// TestScheduleGetLayer verifies GetLayer functionality
func TestScheduleGetLayer(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"test"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)

	// test should be in layer 0
	assert.Equal(t, 0, schedule.GetLayer("test"))

	// build should be in layer 1
	assert.Equal(t, 1, schedule.GetLayer("build"))

	// non-existent step should return -1
	assert.Equal(t, -1, schedule.GetLayer("nonexistent"))
}

// TestScheduleGetStepDependencies verifies dependency retrieval
func TestScheduleGetStepDependencies(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "lint",
					Image: "golangci/golangci-lint:latest",
				},
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"lint", "test"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)

	// build should have two dependencies
	deps := schedule.GetStepDependencies("build")
	require.Len(t, deps, 2)
	depsMap := map[string]bool{deps[0]: true, deps[1]: true}
	assert.True(t, depsMap["lint"])
	assert.True(t, depsMap["test"])

	// lint should have no dependencies
	deps = schedule.GetStepDependencies("lint")
	assert.Empty(t, deps)
}

// TestScheduleGetStepDependents verifies dependent retrieval
func TestScheduleGetStepDependents(t *testing.T) {
	config := &c8sv1alpha1.PipelineConfig{
		Spec: c8sv1alpha1.PipelineConfigSpec{
			Steps: []c8sv1alpha1.PipelineStep{
				{
					Name:  "test",
					Image: "golang:1.21",
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					DependsOn: []string{"test"},
				},
				{
					Name:      "integration",
					Image:     "golang:1.21",
					DependsOn: []string{"test"},
				},
			},
		},
	}

	schedule, err := scheduler.BuildSchedule(config)
	require.NoError(t, err)

	// test should have two dependents
	dependents := schedule.GetStepDependents("test")
	require.Len(t, dependents, 2)
	dependentsMap := map[string]bool{dependents[0]: true, dependents[1]: true}
	assert.True(t, dependentsMap["build"])
	assert.True(t, dependentsMap["integration"])

	// build should have no dependents
	dependents = schedule.GetStepDependents("build")
	assert.Empty(t, dependents)
}
