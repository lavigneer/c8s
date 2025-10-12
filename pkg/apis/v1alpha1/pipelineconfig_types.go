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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineConfigSpec defines the desired state of PipelineConfig
type PipelineConfigSpec struct {
	// Repository is the Git repository URL (https or ssh)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^(https?|git|ssh)://.*`
	Repository string `json:"repository"`

	// Branches are branch filters (glob patterns, default ["*"])
	// +kubebuilder:default={"*"}
	// +optional
	Branches []string `json:"branches,omitempty"`

	// Steps are the pipeline steps in execution order
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Steps []PipelineStep `json:"steps"`

	// Timeout is the pipeline-level timeout (e.g., "30m", "2h")
	// +kubebuilder:validation:Pattern=`^[0-9]+(s|m|h)$`
	// +kubebuilder:default="1h"
	// +optional
	Timeout string `json:"timeout,omitempty"`

	// Matrix strategy for parallel execution
	// +optional
	Matrix *MatrixStrategy `json:"matrix,omitempty"`

	// RetryPolicy defines retry behavior for failed steps
	// +optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

// PipelineStep defines a single step in the pipeline
type PipelineStep struct {
	// Name is the step identifier (must be unique)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name string `json:"name"`

	// Image is the container image for step execution
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Commands are shell commands to execute
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Commands []string `json:"commands"`

	// DependsOn are step names that must complete before this step
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`

	// Resources define CPU/memory requests and limits
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// Timeout is the step timeout (e.g., "30m", "2h")
	// +kubebuilder:validation:Pattern=`^[0-9]+(s|m|h)$`
	// +kubebuilder:default="30m"
	// +optional
	Timeout string `json:"timeout,omitempty"`

	// Artifacts are file patterns to upload to artifact storage
	// +optional
	Artifacts []string `json:"artifacts,omitempty"`

	// Secrets are secret references to inject as env vars
	// +optional
	Secrets []SecretReference `json:"secrets,omitempty"`

	// Conditional defines conditions for step execution
	// +optional
	Conditional *ConditionalExecution `json:"conditional,omitempty"`
}

// ResourceRequirements defines CPU and memory resource constraints
type ResourceRequirements struct {
	// CPU resource request/limit (e.g., "500m", "2")
	// +kubebuilder:validation:Pattern=`^[0-9]+m?$`
	// +kubebuilder:default="500m"
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory resource request/limit (e.g., "1Gi", "512Mi")
	// +kubebuilder:validation:Pattern=`^[0-9]+(Mi|Gi)$`
	// +kubebuilder:default="1Gi"
	// +optional
	Memory string `json:"memory,omitempty"`
}

// SecretReference defines how to inject a Kubernetes Secret into a step
type SecretReference struct {
	// SecretRef is the Kubernetes Secret name
	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`

	// Key is the key within the Secret
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// EnvVar is the environment variable name (defaults to key)
	// +optional
	EnvVar string `json:"envVar,omitempty"`
}

// ConditionalExecution defines conditions for step execution
type ConditionalExecution struct {
	// Branch pattern - execute only on matching branch
	// +optional
	Branch string `json:"branch,omitempty"`

	// OnSuccess - execute only if previous steps succeeded
	// +kubebuilder:default=true
	// +optional
	OnSuccess *bool `json:"onSuccess,omitempty"`
}

// MatrixStrategy defines matrix strategy for parallel execution
type MatrixStrategy struct {
	// Dimensions define matrix variables and their values
	// Example: {"os": ["ubuntu", "alpine"], "go_version": ["1.21", "1.22"]}
	// +kubebuilder:validation:Required
	Dimensions map[string][]string `json:"dimensions"`

	// Exclude specific combinations
	// +optional
	Exclude []map[string]string `json:"exclude,omitempty"`
}

// RetryPolicy defines retry behavior for failed steps
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=5
	// +kubebuilder:default=0
	// +optional
	MaxRetries int `json:"maxRetries,omitempty"`

	// BackoffSeconds is the delay between retries
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=60
	// +optional
	BackoffSeconds int `json:"backoffSeconds,omitempty"`
}

// PipelineConfigStatus defines the observed state of PipelineConfig
type PipelineConfigStatus struct {
	// LastRun is the timestamp of the last pipeline run
	// +optional
	LastRun *metav1.Time `json:"lastRun,omitempty"`

	// TotalRuns is the total number of pipeline runs
	// +optional
	TotalRuns int `json:"totalRuns,omitempty"`

	// SuccessRate is the percentage of successful runs
	// +optional
	SuccessRate float64 `json:"successRate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pc
// +kubebuilder:printcolumn:name="Repository",type=string,JSONPath=`.spec.repository`
// +kubebuilder:printcolumn:name="Steps",type=integer,JSONPath=`.spec.steps[*].name`
// +kubebuilder:printcolumn:name="Last Run",type=date,JSONPath=`.status.lastRun`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PipelineConfig is the Schema for the pipelineconfigs API
type PipelineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineConfigSpec   `json:"spec,omitempty"`
	Status PipelineConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineConfigList contains a list of PipelineConfig
type PipelineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineConfig{}, &PipelineConfigList{})
}
