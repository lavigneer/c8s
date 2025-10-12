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

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrPipelineConfigNotFound indicates the referenced PipelineConfig doesn't exist
	ErrPipelineConfigNotFound = errors.New("pipeline config not found")

	// ErrInvalidDependencyGraph indicates the step dependency graph has cycles
	ErrInvalidDependencyGraph = errors.New("invalid dependency graph: circular dependencies detected")

	// ErrStepNotFound indicates a referenced step doesn't exist
	ErrStepNotFound = errors.New("step not found")

	// ErrSecretNotFound indicates a referenced Secret doesn't exist
	ErrSecretNotFound = errors.New("secret not found")

	// ErrInvalidYAML indicates the pipeline YAML is malformed
	ErrInvalidYAML = errors.New("invalid pipeline YAML")

	// ErrResourceQuotaExceeded indicates namespace quota was exceeded
	ErrResourceQuotaExceeded = errors.New("resource quota exceeded")

	// ErrStorageUploadFailed indicates log/artifact upload failed
	ErrStorageUploadFailed = errors.New("storage upload failed")

	// ErrStorageDownloadFailed indicates log/artifact download failed
	ErrStorageDownloadFailed = errors.New("storage download failed")

	// ErrWebhookValidationFailed indicates webhook signature validation failed
	ErrWebhookValidationFailed = errors.New("webhook validation failed")

	// ErrInvalidWebhookPayload indicates the webhook payload is malformed
	ErrInvalidWebhookPayload = errors.New("invalid webhook payload")

	// ErrAuthenticationFailed indicates Git authentication failed
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")
)

// PipelineError wraps errors with pipeline run context
type PipelineError struct {
	PipelineRun string
	Step        string
	Err         error
}

func (e *PipelineError) Error() string {
	if e.Step != "" {
		return fmt.Sprintf("pipeline %s, step %s: %v", e.PipelineRun, e.Step, e.Err)
	}
	return fmt.Sprintf("pipeline %s: %v", e.PipelineRun, e.Err)
}

func (e *PipelineError) Unwrap() error {
	return e.Err
}

// NewPipelineError creates a new PipelineError
func NewPipelineError(pipelineRun, step string, err error) *PipelineError {
	return &PipelineError{
		PipelineRun: pipelineRun,
		Step:        step,
		Err:         err,
	}
}

// StorageError wraps storage operation errors
type StorageError struct {
	Operation string // "upload" or "download"
	Key       string
	Err       error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage %s failed for key %s: %v", e.Operation, e.Key, e.Err)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new StorageError
func NewStorageError(operation, key string, err error) *StorageError {
	return &StorageError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}
