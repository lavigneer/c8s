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

// Package storage provides interfaces and implementations for object storage (S3)
package storage

import (
	"context"
	"io"
	"time"
)

// StorageClient defines the interface for object storage operations
type StorageClient interface {
	// UploadLog uploads log content to object storage
	// key format: "c8s-logs/{namespace}/{pipeline-run}/{step-name}.log"
	UploadLog(ctx context.Context, key string, content io.Reader) error

	// DownloadLog downloads log content from object storage
	DownloadLog(ctx context.Context, key string) (io.ReadCloser, error)

	// UploadArtifact uploads an artifact file to object storage
	// key format: "c8s-artifacts/{namespace}/{pipeline-run}/{step-name}/{filename}"
	UploadArtifact(ctx context.Context, key string, content io.Reader) error

	// DownloadArtifact downloads an artifact file from object storage
	DownloadArtifact(ctx context.Context, key string) (io.ReadCloser, error)

	// GenerateSignedURL generates a pre-signed URL for downloading a file
	// Useful for providing time-limited access to logs and artifacts
	GenerateSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// ListObjects lists objects with the given prefix
	// Used for listing all artifacts for a pipeline run
	ListObjects(ctx context.Context, prefix string) ([]string, error)

	// DeleteObject deletes an object from storage
	// Used for cleanup when pipeline run is deleted
	DeleteObject(ctx context.Context, key string) error

	// ObjectExists checks if an object exists in storage
	ObjectExists(ctx context.Context, key string) (bool, error)
}

// Config holds configuration for storage client
type Config struct {
	// Bucket is the S3 bucket name
	Bucket string

	// Region is the AWS region (e.g., "us-west-2")
	Region string

	// Endpoint is the S3 endpoint URL (optional, for S3-compatible services like MinIO)
	Endpoint string

	// AccessKeyID is the AWS access key ID
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string

	// UsePathStyle forces path-style URLs (required for MinIO)
	UsePathStyle bool
}

// Validate validates the storage configuration
func (c *Config) Validate() error {
	if c.Bucket == "" {
		return ErrMissingBucket
	}
	if c.AccessKeyID == "" {
		return ErrMissingAccessKey
	}
	if c.SecretAccessKey == "" {
		return ErrMissingSecretKey
	}
	return nil
}
