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

package storage

import "errors"

// Storage configuration errors
var (
	// ErrMissingBucket indicates the bucket name is not configured
	ErrMissingBucket = errors.New("storage bucket name is required")

	// ErrMissingAccessKey indicates the access key ID is not configured
	ErrMissingAccessKey = errors.New("storage access key ID is required")

	// ErrMissingSecretKey indicates the secret access key is not configured
	ErrMissingSecretKey = errors.New("storage secret access key is required")

	// ErrObjectNotFound indicates the requested object doesn't exist
	ErrObjectNotFound = errors.New("object not found in storage")

	// ErrUploadFailed indicates the upload operation failed
	ErrUploadFailed = errors.New("failed to upload object to storage")

	// ErrDownloadFailed indicates the download operation failed
	ErrDownloadFailed = errors.New("failed to download object from storage")

	// ErrDeleteFailed indicates the delete operation failed
	ErrDeleteFailed = errors.New("failed to delete object from storage")

	// ErrListFailed indicates the list operation failed
	ErrListFailed = errors.New("failed to list objects in storage")

	// ErrSignedURLFailed indicates signed URL generation failed
	ErrSignedURLFailed = errors.New("failed to generate signed URL")
)
