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

// GitProvider represents the version control provider type
// +kubebuilder:validation:Enum=github;gitlab;bitbucket
type GitProvider string

const (
	// GitProviderGitHub represents GitHub
	GitProviderGitHub GitProvider = "github"
	// GitProviderGitLab represents GitLab
	GitProviderGitLab GitProvider = "gitlab"
	// GitProviderBitbucket represents Bitbucket
	GitProviderBitbucket GitProvider = "bitbucket"
)

// RepositoryConnectionSpec defines the desired state of RepositoryConnection
type RepositoryConnectionSpec struct {
	// Repository is the Git repository URL
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^(https?|git|ssh)://.*`
	Repository string `json:"repository"`

	// Provider is the version control provider
	// +kubebuilder:validation:Required
	Provider GitProvider `json:"provider"`

	// WebhookSecretRef is the Kubernetes Secret containing webhook signature secret
	// The secret must contain a "webhook-secret" key
	// +optional
	WebhookSecretRef string `json:"webhookSecretRef,omitempty"`

	// AuthSecretRef is the Kubernetes Secret containing Git credentials
	// For HTTPS: must contain "username" and "password" keys
	// For SSH: must contain "ssh-key" key
	// +optional
	AuthSecretRef string `json:"authSecretRef,omitempty"`

	// PipelineConfigRef is the default PipelineConfig to use for this repository
	// +optional
	PipelineConfigRef string `json:"pipelineConfigRef,omitempty"`

	// Events are the webhook event types to handle (push, pull_request, etc.)
	// +optional
	Events []string `json:"events,omitempty"`

	// Branches are branch patterns to filter webhook events
	// +kubebuilder:validation:items:Pattern=`^[a-zA-Z0-9._/*-]+$`
	// +optional
	Branches []string `json:"branches,omitempty"`

	// Tags are tag patterns to filter webhook events
	// +optional
	Tags []string `json:"tags,omitempty"`
}

// RepositoryConnectionStatus defines the observed state of RepositoryConnection
type RepositoryConnectionStatus struct {
	// WebhookURL is the webhook endpoint URL to configure in the provider
	// +optional
	WebhookURL string `json:"webhookURL,omitempty"`

	// WebhookRegistered indicates whether the webhook is successfully registered
	// +optional
	WebhookRegistered bool `json:"webhookRegistered,omitempty"`

	// LastEvent contains information about the last received webhook event
	// +optional
	LastEvent *WebhookEvent `json:"lastEvent,omitempty"`

	// Conditions represent the latest available observations of the connection's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// WebhookEvent represents information about a received webhook event
type WebhookEvent struct {
	// Type is the event type (push, pull_request, etc.)
	// +optional
	Type string `json:"type,omitempty"`

	// Commit is the commit SHA from the event
	// +optional
	Commit string `json:"commit,omitempty"`

	// Branch is the branch name from the event
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._/-]+$`
	// +optional
	Branch string `json:"branch,omitempty"`

	// Timestamp is when the event was received
	// +optional
	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rc
// +kubebuilder:printcolumn:name="Repository",type=string,JSONPath=`.spec.repository`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Registered",type=boolean,JSONPath=`.status.webhookRegistered`
// +kubebuilder:printcolumn:name="Last Event",type=date,JSONPath=`.status.lastEvent.timestamp`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RepositoryConnection is the Schema for the repositoryconnections API
type RepositoryConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositoryConnectionSpec   `json:"spec,omitempty"`
	Status RepositoryConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepositoryConnectionList contains a list of RepositoryConnection
type RepositoryConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepositoryConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RepositoryConnection{}, &RepositoryConnectionList{})
}
