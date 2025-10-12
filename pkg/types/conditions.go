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

package types

// Condition types for PipelineRun status
const (
	// ConditionTypeReady indicates the PipelineRun is ready to execute
	ConditionTypeReady = "Ready"

	// ConditionTypePipelineConfigResolved indicates the PipelineConfig was found and loaded
	ConditionTypePipelineConfigResolved = "PipelineConfigResolved"

	// ConditionTypeJobsCreated indicates all Jobs have been created
	ConditionTypeJobsCreated = "JobsCreated"

	// ConditionTypeStepsCompleted indicates all steps have completed (success or failure)
	ConditionTypeStepsCompleted = "StepsCompleted"

	// ConditionTypeLogsUploaded indicates logs have been uploaded to object storage
	ConditionTypeLogsUploaded = "LogsUploaded"

	// ConditionTypeArtifactsUploaded indicates artifacts have been uploaded
	ConditionTypeArtifactsUploaded = "ArtifactsUploaded"
)

// Condition reasons for PipelineRun status
const (
	// ReasonPipelineConfigNotFound indicates the referenced PipelineConfig doesn't exist
	ReasonPipelineConfigNotFound = "PipelineConfigNotFound"

	// ReasonPipelineConfigResolved indicates the PipelineConfig was found
	ReasonPipelineConfigResolved = "PipelineConfigResolved"

	// ReasonJobCreationFailed indicates a Job could not be created
	ReasonJobCreationFailed = "JobCreationFailed"

	// ReasonJobsCreated indicates all Jobs were created successfully
	ReasonJobsCreated = "JobsCreated"

	// ReasonStepRunning indicates at least one step is running
	ReasonStepRunning = "StepRunning"

	// ReasonStepFailed indicates a step failed
	ReasonStepFailed = "StepFailed"

	// ReasonStepSucceeded indicates all steps succeeded
	ReasonStepSucceeded = "StepSucceeded"

	// ReasonStepSkipped indicates a step was skipped due to conditions
	ReasonStepSkipped = "StepSkipped"

	// ReasonTimeout indicates the pipeline timed out
	ReasonTimeout = "Timeout"

	// ReasonCancelled indicates the pipeline was cancelled by user
	ReasonCancelled = "Cancelled"

	// ReasonStorageError indicates an error uploading to object storage
	ReasonStorageError = "StorageError"

	// ReasonLogsUploaded indicates logs were uploaded successfully
	ReasonLogsUploaded = "LogsUploaded"

	// ReasonArtifactsUploaded indicates artifacts were uploaded successfully
	ReasonArtifactsUploaded = "ArtifactsUploaded"

	// ReasonResourceQuotaExceeded indicates namespace quota was exceeded
	ReasonResourceQuotaExceeded = "ResourceQuotaExceeded"

	// ReasonSecretNotFound indicates a referenced Secret doesn't exist
	ReasonSecretNotFound = "SecretNotFound"
)

// Condition types for RepositoryConnection status
const (
	// ConditionTypeWebhookRegistered indicates the webhook is registered with the provider
	ConditionTypeWebhookRegistered = "WebhookRegistered"

	// ConditionTypeAuthenticated indicates credentials are valid
	ConditionTypeAuthenticated = "Authenticated"
)

// Condition reasons for RepositoryConnection status
const (
	// ReasonWebhookRegistered indicates webhook registration succeeded
	ReasonWebhookRegistered = "WebhookRegistered"

	// ReasonWebhookRegistrationFailed indicates webhook registration failed
	ReasonWebhookRegistrationFailed = "WebhookRegistrationFailed"

	// ReasonAuthenticationSucceeded indicates authentication succeeded
	ReasonAuthenticationSucceeded = "AuthenticationSucceeded"

	// ReasonAuthenticationFailed indicates authentication failed
	ReasonAuthenticationFailed = "AuthenticationFailed"

	// ReasonInvalidCredentials indicates credentials are invalid
	ReasonInvalidCredentials = "InvalidCredentials"
)
