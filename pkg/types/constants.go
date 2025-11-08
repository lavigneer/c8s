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
//nolint:revive // the package name is intentional for grouping C8S shared types
package types

const (
	// LabelPipelineConfig is the label key for pipeline configuration name.
	LabelPipelineConfig = "c8s.dev/pipeline-config"
	// LabelPipelineRun is the label key for pipeline run name.
	LabelPipelineRun = "c8s.dev/pipeline-run"
	// LabelStepName is the label key for step name.
	LabelStepName = "c8s.dev/step-name"
	// LabelCommit is the label key for commit SHA.
	LabelCommit = "c8s.dev/commit"
	// LabelBranch is the label key for branch name.
	LabelBranch = "c8s.dev/branch"
	// LabelManagedBy is the label key for managed-by identifier.
	LabelManagedBy = "app.kubernetes.io/managed-by"

	// AnnotationCommitMessage is the annotation key for commit message.
	AnnotationCommitMessage = "c8s.dev/commit-message"
	// AnnotationAuthor is the annotation key for commit author.
	AnnotationAuthor = "c8s.dev/author"
	// AnnotationTriggeredBy is the annotation key for trigger source.
	AnnotationTriggeredBy = "c8s.dev/triggered-by"
	// AnnotationLogURL is the annotation key for log URL.
	AnnotationLogURL = "c8s.dev/log-url"
	// AnnotationArtifactURLs is the annotation key for artifact URLs.
	AnnotationArtifactURLs = "c8s.dev/artifact-urls"

	// FinalizerPipelineRun is the finalizer for pipeline run cleanup.
	FinalizerPipelineRun = "c8s.dev/pipelinerun"
	// FinalizerCleanupJobs is the finalizer for job cleanup.
	FinalizerCleanupJobs = "c8s.dev/cleanup-jobs"
	// FinalizerCleanupLogs is the finalizer for log cleanup.
	FinalizerCleanupLogs = "c8s.dev/cleanup-logs"

	// ManagedByC8S is the value for the managed-by label.
	ManagedByC8S = "c8s"

	// JobTTLSecondsAfterFinished is the TTL for finished jobs (1 hour).
	JobTTLSecondsAfterFinished = 3600
	// JobBackoffLimit is the backoff limit for jobs (no retries at Job level).
	JobBackoffLimit = 0

	// ContainerNameGitClone is the container name for git clone step.
	ContainerNameGitClone = "git-clone"
	// ContainerNameStep is the container name for pipeline step.
	ContainerNameStep = "step"
	// ContainerNameArtifact is the container name for artifact upload.
	ContainerNameArtifact = "artifact-upload"

	// VolumeNameWorkspace is the volume name for workspace.
	VolumeNameWorkspace = "workspace"
	// VolumeNameSecrets is the volume name for secrets.
	VolumeNameSecrets = "secrets"

	// MountPathWorkspace is the mount path for workspace volume.
	MountPathWorkspace = "/workspace"
	// MountPathSecrets is the mount path for secrets volume.
	MountPathSecrets = "/secrets"

	// EnvCommitSHA is the environment variable for commit SHA.
	EnvCommitSHA = "COMMIT_SHA"
	// EnvBranch is the environment variable for branch name.
	EnvBranch = "BRANCH"
	// EnvPipelineRun is the environment variable for pipeline run name.
	EnvPipelineRun = "PIPELINE_RUN"
	// EnvStepName is the environment variable for step name.
	EnvStepName = "STEP_NAME"
	// EnvWorkspace is the environment variable for workspace path.
	EnvWorkspace = "WORKSPACE"
	// EnvC8SNamespace is the environment variable for C8S namespace.
	EnvC8SNamespace = "C8S_NAMESPACE"

	// StorageBucketEnv is the environment variable for storage bucket.
	StorageBucketEnv = "C8S_STORAGE_BUCKET"
	// StorageRegionEnv is the environment variable for storage region.
	StorageRegionEnv = "C8S_STORAGE_REGION"
	// StorageEndpointEnv is the environment variable for storage endpoint.
	StorageEndpointEnv = "C8S_STORAGE_ENDPOINT"
	// StorageAccessKeyEnv is the environment variable for storage access key.
	StorageAccessKeyEnv = "AWS_ACCESS_KEY_ID"
	// StorageSecretKeyEnv is the environment variable for storage secret key.
	StorageSecretKeyEnv = "AWS_SECRET_ACCESS_KEY"
	// StorageLogPrefix is the prefix for log storage paths.
	StorageLogPrefix = "c8s-logs"
	// StorageArtifactPrefix is the prefix for artifact storage paths.
	StorageArtifactPrefix = "c8s-artifacts"
	// StorageURLExpirySeconds is the expiry time for storage URLs (1 hour).
	StorageURLExpirySeconds = 3600
)
