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

// PipelineRunPhase represents the current phase of a PipelineRun
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Cancelled
type PipelineRunPhase string

const (
	// PipelineRunPhasePending means the pipeline run is created but not yet started
	PipelineRunPhasePending PipelineRunPhase = "Pending"
	// PipelineRunPhaseRunning means at least one step is executing
	PipelineRunPhaseRunning PipelineRunPhase = "Running"
	// PipelineRunPhaseSucceeded means all steps completed successfully
	PipelineRunPhaseSucceeded PipelineRunPhase = "Succeeded"
	// PipelineRunPhaseFailed means at least one step failed
	PipelineRunPhaseFailed PipelineRunPhase = "Failed"
	// PipelineRunPhaseCancelled means the run was cancelled
	PipelineRunPhaseCancelled PipelineRunPhase = "Cancelled"
)

// StepPhase represents the current phase of a pipeline step
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
type StepPhase string

const (
	// StepPhasePending means the step is waiting to run
	StepPhasePending StepPhase = "Pending"
	// StepPhaseRunning means the step is currently executing
	StepPhaseRunning StepPhase = "Running"
	// StepPhaseSucceeded means the step completed successfully
	StepPhaseSucceeded StepPhase = "Succeeded"
	// StepPhaseFailed means the step failed
	StepPhaseFailed StepPhase = "Failed"
	// StepPhaseSkipped means the step was skipped due to conditions
	StepPhaseSkipped StepPhase = "Skipped"
)

// PipelineRunSpec defines the desired state of PipelineRun
type PipelineRunSpec struct {
	// PipelineConfigRef is the reference to the PipelineConfig name
	// +kubebuilder:validation:Required
	PipelineConfigRef string `json:"pipelineConfigRef"`

	// Commit is the Git commit SHA
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-f0-9]{7,40}$`
	Commit string `json:"commit"`

	// Branch is the branch name
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._/-]+$`
	// +optional
	Branch string `json:"branch,omitempty"`

	// TriggeredBy is the user or system that triggered the run
	// +kubebuilder:default="system"
	// +optional
	TriggeredBy string `json:"triggeredBy,omitempty"`

	// TriggeredAt is the time when the run was triggered
	// +optional
	TriggeredAt *metav1.Time `json:"triggeredAt,omitempty"`

	// MatrixIndex contains matrix variable values for this specific run
	// +optional
	MatrixIndex map[string]string `json:"matrixIndex,omitempty"`

	// CommitMessage is the git commit message
	// +optional
	CommitMessage string `json:"commitMessage,omitempty"`

	// Author is the commit author email
	// +optional
	Author string `json:"author,omitempty"`
}

// PipelineRunStatus defines the observed state of PipelineRun
type PipelineRunStatus struct {
	// Phase is the current phase of the pipeline run
	// +optional
	Phase PipelineRunPhase `json:"phase,omitempty"`

	// StartTime is when the run started executing
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the run completed (succeeded, failed, or cancelled)
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Steps contains the status of each step in the pipeline
	// +optional
	Steps []StepStatus `json:"steps,omitempty"`

	// Conditions represent the latest available observations of the run's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ResourceUsage tracks actual resource consumption for capacity planning
	// +optional
	ResourceUsage *ResourceUsage `json:"resourceUsage,omitempty"`
}

// ResourceUsage tracks resource consumption for a PipelineRun
type ResourceUsage struct {
	// CPU is the total CPU time consumed in core-seconds
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory is the average memory usage in bytes
	// +optional
	Memory string `json:"memory,omitempty"`

	// Duration is the total wall-clock duration in seconds
	// +optional
	Duration int64 `json:"duration,omitempty"`
}

// StepStatus represents the status of a single step in the pipeline
type StepStatus struct {
	// Name is the step identifier
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Phase is the current phase of the step
	// +kubebuilder:validation:Required
	Phase StepPhase `json:"phase"`

	// JobName is the Kubernetes Job name for this step
	// +optional
	JobName string `json:"jobName,omitempty"`

	// StartTime is when the step started executing
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the step completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// ExitCode is the exit code of the step's main container
	// +optional
	ExitCode *int32 `json:"exitCode,omitempty"`

	// LogURL is the object storage URL for the step's logs
	// +optional
	LogURL string `json:"logURL,omitempty"`

	// ArtifactURLs are the object storage URLs for step artifacts
	// +optional
	ArtifactURLs []string `json:"artifactURLs,omitempty"`

	// Message provides additional context about the step status
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pr
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Config",type=string,JSONPath=`.spec.pipelineConfigRef`
// +kubebuilder:printcolumn:name="Commit",type=string,JSONPath=`.spec.commit`,priority=1
// +kubebuilder:printcolumn:name="Branch",type=string,JSONPath=`.spec.branch`,priority=1
// +kubebuilder:printcolumn:name="Started",type=date,JSONPath=`.status.startTime`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PipelineRun is the Schema for the pipelineruns API
type PipelineRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineRunSpec   `json:"spec,omitempty"`
	Status PipelineRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineRunList contains a list of PipelineRun
type PipelineRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineRun{}, &PipelineRunList{})
}
