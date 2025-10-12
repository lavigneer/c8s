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

// Package scheduler provides DAG-based scheduling for pipeline steps
package scheduler

import (
	"fmt"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/types"
)

// DAG represents a directed acyclic graph of pipeline steps
type DAG struct {
	// nodes maps step names to their step definitions
	nodes map[string]*c8sv1alpha1.PipelineStep

	// edges maps step names to their dependencies (incoming edges)
	edges map[string][]string

	// reverseEdges maps step names to steps that depend on them (outgoing edges)
	reverseEdges map[string][]string
}

// BuildDAG constructs a DAG from pipeline steps
func BuildDAG(steps []c8sv1alpha1.PipelineStep) (*DAG, error) {
	dag := &DAG{
		nodes:        make(map[string]*c8sv1alpha1.PipelineStep),
		edges:        make(map[string][]string),
		reverseEdges: make(map[string][]string),
	}

	// First pass: register all nodes
	for i := range steps {
		step := &steps[i]
		if _, exists := dag.nodes[step.Name]; exists {
			return nil, fmt.Errorf("duplicate step name: %s", step.Name)
		}
		dag.nodes[step.Name] = step
		dag.edges[step.Name] = []string{}
		dag.reverseEdges[step.Name] = []string{}
	}

	// Second pass: build edges
	for _, step := range steps {
		for _, dep := range step.DependsOn {
			// Verify dependency exists
			if _, exists := dag.nodes[dep]; !exists {
				return nil, fmt.Errorf("%w: step %s depends on non-existent step %s",
					types.ErrStepNotFound, step.Name, dep)
			}

			// Add edge: step depends on dep
			dag.edges[step.Name] = append(dag.edges[step.Name], dep)

			// Add reverse edge: dep is depended on by step
			dag.reverseEdges[dep] = append(dag.reverseEdges[dep], step.Name)
		}
	}

	// Validate no cycles
	if err := dag.detectCycles(); err != nil {
		return nil, err
	}

	return dag, nil
}

// TopologicalSort returns steps grouped into execution layers
// Steps in the same layer can be executed in parallel
// Returns [][]string where each inner slice is a layer of step names
func (d *DAG) TopologicalSort() ([][]string, error) {
	// Track in-degree (number of dependencies) for each node
	inDegree := make(map[string]int)
	for node := range d.nodes {
		inDegree[node] = len(d.edges[node])
	}

	var layers [][]string
	remaining := len(d.nodes)

	for remaining > 0 {
		// Find all nodes with in-degree 0 (no unfulfilled dependencies)
		var currentLayer []string
		for node, degree := range inDegree {
			if degree == 0 {
				currentLayer = append(currentLayer, node)
			}
		}

		if len(currentLayer) == 0 {
			// No nodes ready but still have remaining nodes = cycle
			return nil, types.ErrInvalidDependencyGraph
		}

		// Add layer to result
		layers = append(layers, currentLayer)

		// Remove these nodes and update in-degrees
		for _, node := range currentLayer {
			delete(inDegree, node)
			remaining--

			// Decrease in-degree for dependent nodes
			for _, dependent := range d.reverseEdges[node] {
				if deg, exists := inDegree[dependent]; exists {
					inDegree[dependent] = deg - 1
				}
			}
		}
	}

	return layers, nil
}

// detectCycles checks for circular dependencies using DFS
func (d *DAG) detectCycles() error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var visit func(node string) error
	visit = func(node string) error {
		visited[node] = true
		recursionStack[node] = true

		for _, dep := range d.edges[node] {
			if !visited[dep] {
				if err := visit(dep); err != nil {
					return err
				}
			} else if recursionStack[dep] {
				return fmt.Errorf("%w: cycle detected involving steps %s and %s",
					types.ErrInvalidDependencyGraph, node, dep)
			}
		}

		recursionStack[node] = false
		return nil
	}

	for node := range d.nodes {
		if !visited[node] {
			if err := visit(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetStep returns the step definition for a given step name
func (d *DAG) GetStep(name string) (*c8sv1alpha1.PipelineStep, bool) {
	step, exists := d.nodes[name]
	return step, exists
}

// GetDependencies returns the list of direct dependencies for a step
func (d *DAG) GetDependencies(name string) []string {
	return d.edges[name]
}

// GetDependents returns the list of steps that depend on this step
func (d *DAG) GetDependents(name string) []string {
	return d.reverseEdges[name]
}

// Size returns the number of steps in the DAG
func (d *DAG) Size() int {
	return len(d.nodes)
}
