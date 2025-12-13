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

package logstorage

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/org/c8s/pkg/storage"
)

// S3LogStorage fetches logs from S3-compatible object storage
type S3LogStorage struct {
	storageClient storage.Client
	namespace     string
}

// NewS3LogStorage creates a new S3LogStorage instance
func NewS3LogStorage(storageClient storage.Client, namespace string) *S3LogStorage {
	return &S3LogStorage{
		storageClient: storageClient,
		namespace:     namespace,
	}
}

// GetStepLogs downloads logs from S3 storage
func (s *S3LogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	if s.storageClient == nil {
		return nil, fmt.Errorf("storage client not initialized")
	}

	// Build S3 key: {namespace}/{runID}/{stepID}.log
	key := fmt.Sprintf("%s/%s/%s.log", s.namespace, runID, stepID)

	// Download from S3
	logStream, err := s.storageClient.DownloadLog(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to download logs from S3 (key: %s): %w", key, err)
	}

	return logStream, nil
}

// StreamStepLogs streams logs from S3 line by line
func (s *S3LogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	logsReader, err := s.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return err
	}
	defer logsReader.Close()

	// Stream logs line by line
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case linesChan <- scanner.Text():
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs from S3: %w", err)
	}

	return nil
}

// GetLogSnapshot returns the last N lines of logs from S3
func (s *S3LogStorage) GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error) {
	logsReader, err := s.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return nil, err
	}
	defer logsReader.Close()

	var result []string
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading logs from S3: %w", err)
	}

	// Return last N lines
	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return result, nil
}

// GetLogSize returns the size of logs in bytes from S3
func (s *S3LogStorage) GetLogSize(ctx context.Context, runID, stepID string) (int64, error) {
	logsReader, err := s.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return 0, err
	}
	defer logsReader.Close()

	// Read all logs and count bytes
	data, err := io.ReadAll(logsReader)
	if err != nil {
		return 0, fmt.Errorf("failed to read logs from S3: %w", err)
	}

	return int64(len(data)), nil
}
