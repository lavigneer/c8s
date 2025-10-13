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

// Package types contains common types, constants, and utilities for C8S
package types

const (
	// Label keys
	LabelPipelineConfig = "c8s.dev/pipeline-config"
	LabelPipelineRun    = "c8s.dev/pipeline-run"
	LabelStepName       = "c8s.dev/step-name"
	LabelCommit         = "c8s.dev/commit"
	LabelBranch         = "c8s.dev/branch"
	LabelManagedBy      = "app.kubernetes.io/managed-by"

	// Annotation keys
	AnnotationCommitMessage = "c8s.dev/commit-message"
	AnnotationAuthor        = "c8s.dev/author"
	AnnotationTriggeredBy   = "c8s.dev/triggered-by"
	AnnotationLogURL        = "c8s.dev/log-url"
	AnnotationArtifactURLs  = "c8s.dev/artifact-urls"

	// Finalizer names
	FinalizerPipelineRun = "c8s.dev/pipelinerun"
	FinalizerCleanupJobs = "c8s.dev/cleanup-jobs"
	FinalizerCleanupLogs = "c8s.dev/cleanup-logs"

	// Managed by value
	ManagedByC8S = "c8s"

	// Job configuration
	JobTTLSecondsAfterFinished = 3600 // 1 hour
	JobBackoffLimit            = 0    // No retries at Job level (handled by RetryPolicy)

	// Container names
	ContainerNameGitClone = "git-clone"
	ContainerNameStep     = "step"
	ContainerNameArtifact = "artifact-upload"

	// Volume names
	VolumeNameWorkspace = "workspace"
	VolumeNameSecrets   = "secrets"

	// Mount paths
	MountPathWorkspace = "/workspace"
	MountPathSecrets   = "/secrets"

	// Environment variables
	EnvCommitSHA    = "COMMIT_SHA"
	EnvBranch       = "BRANCH"
	EnvPipelineRun  = "PIPELINE_RUN"
	EnvStepName     = "STEP_NAME"
	EnvWorkspace    = "WORKSPACE"
	EnvC8SNamespace = "C8S_NAMESPACE"

	// Storage configuration
	StorageBucketEnv         = "C8S_STORAGE_BUCKET"
	StorageRegionEnv         = "C8S_STORAGE_REGION"
	StorageEndpointEnv       = "C8S_STORAGE_ENDPOINT"
	StorageAccessKeyEnv      = "AWS_ACCESS_KEY_ID"
	StorageSecretKeyEnv      = "AWS_SECRET_ACCESS_KEY"
	StorageLogPrefix         = "c8s-logs"
	StorageArtifactPrefix    = "c8s-artifacts"
	StorageURLExpirySeconds  = 3600 // 1 hour
)
