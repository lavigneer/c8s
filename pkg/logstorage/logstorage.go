// Package logstorage provides log storage interfaces and implementations for pipeline runs.
package logstorage

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// LogStorage defines interface for retrieving logs from object storage
type LogStorage interface {
	// GetStepLogs returns a reader for step logs
	GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error)

	// StreamStepLogs streams logs line-by-line to channel
	// Returns error if logs cannot be read
	StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error

	// GetLogSnapshot returns last N lines of logs
	GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error)

	// GetLogSize returns size of log file in bytes
	GetLogSize(ctx context.Context, runID, stepID string) (int64, error)
}

// InMemoryLogStorage implements LogStorage using in-memory storage (for testing)
type InMemoryLogStorage struct {
	logs map[string]string
}

// NewInMemoryLogStorage creates a new in-memory log storage with demo data for all known runs
func NewInMemoryLogStorage() *InMemoryLogStorage {
	storage := &InMemoryLogStorage{
		logs: make(map[string]string),
	}
	storage.populateDemoLogs()
	return storage
}

// NewInMemoryLogStorageWithRun creates a new in-memory log storage with demo data for a specific run
func NewInMemoryLogStorageWithRun(runID string) *InMemoryLogStorage {
	storage := &InMemoryLogStorage{
		logs: make(map[string]string),
	}
	storage.populateDemoLogs()
	storage.PopulateDemoLogsForRun(runID)
	return storage
}

// populateDemoLogs adds sample logs for testing
func (s *InMemoryLogStorage) populateDemoLogs() {
	demoLogs := map[string]string{
		"hello-world-run-001/step-1": `[2025-10-27T04:30:10Z] Step started: checkout
[2025-10-27T04:30:11Z] $ git clone https://github.com/example/repo.git
[2025-10-27T04:30:12Z] Cloning into 'repo'...
[2025-10-27T04:30:13Z] remote: Counting objects: 1000, done
[2025-10-27T04:30:14Z] Receiving objects: 100% (1000/1000)
[2025-10-27T04:30:15Z] Step completed with status: Succeeded`,
		"hello-world-run-001/step-2": `[2025-10-27T04:30:16Z] Step started: build
[2025-10-27T04:30:17Z] $ echo 'Starting build process'
[2025-10-27T04:30:17Z] Starting build process
[2025-10-27T04:30:18Z] $ go build -o bin/app ./cmd/api-server
[2025-10-27T04:30:19Z] go: downloading github.com/go-chi/chi/v5
[2025-10-27T04:30:20Z] go: downloading sigs.k8s.io/controller-runtime
[2025-10-27T04:30:22Z] Build completed successfully
[2025-10-27T04:30:23Z] Artifacts: bin/app (45.2 MB)
[2025-10-27T04:30:24Z] Step completed with status: Succeeded`,
		"hello-world-run-001/step-3": `[2025-10-27T04:30:25Z] Step started: test
[2025-10-27T04:30:26Z] $ go test ./...
[2025-10-27T04:30:27Z] ok      github.com/org/c8s/cmd/api-server  0.234s
[2025-10-27T04:30:28Z] ok      github.com/org/c8s/pkg/dashboard   0.456s
[2025-10-27T04:30:29Z] ok      github.com/org/c8s/pkg/apis        0.123s
[2025-10-27T04:30:30Z] Coverage: 78.5%
[2025-10-27T04:30:31Z] Step completed with status: Succeeded`,
	}

	s.logs = demoLogs
}

// SetLog sets log content for a step (for testing)
func (s *InMemoryLogStorage) SetLog(runID, stepID, content string) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	s.logs[key] = content
}

// GetStepLogs returns a reader for step logs
func (s *InMemoryLogStorage) GetStepLogs(_ context.Context, runID, stepID string) (io.ReadCloser, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]
	if !ok {
		return nil, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

// StreamStepLogs streams logs line-by-line to channel
// Blocks until all logs are sent or context is canceled
func (s *InMemoryLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]
	if !ok {
		return fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case linesChan <- scanner.Text():
		}
	}

	return nil
}

// GetLogSnapshot returns last N lines of logs
func (s *InMemoryLogStorage) GetLogSnapshot(_ context.Context, runID, stepID string, lines int) ([]string, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]
	if !ok {
		return nil, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}

	var result []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	// Return last N lines
	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return result, nil
}

// GetLogSize returns size of log content in bytes
func (s *InMemoryLogStorage) GetLogSize(_ context.Context, runID, stepID string) (int64, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]
	if !ok {
		return 0, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}
	return int64(len(content)), nil
}

// PopulateDemoLogsForRun adds demo logs for a specific run ID
func (s *InMemoryLogStorage) PopulateDemoLogsForRun(runID string) {
	stepLogs := map[string]string{
		"step-1": `[2025-10-27T04:30:10Z] Step started: checkout
[2025-10-27T04:30:11Z] $ git clone https://github.com/example/repo.git
[2025-10-27T04:30:12Z] Cloning into 'repo'...
[2025-10-27T04:30:13Z] remote: Counting objects: 1000, done
[2025-10-27T04:30:14Z] Receiving objects: 100% (1000/1000)
[2025-10-27T04:30:15Z] Step completed with status: Succeeded`,
		"step-2": `[2025-10-27T04:30:16Z] Step started: build
[2025-10-27T04:30:17Z] $ echo 'Starting build process'
[2025-10-27T04:30:17Z] Starting build process
[2025-10-27T04:30:18Z] $ go build -o bin/app ./cmd/api-server
[2025-10-27T04:30:19Z] go: downloading github.com/go-chi/chi/v5
[2025-10-27T04:30:20Z] go: downloading sigs.k8s.io/controller-runtime
[2025-10-27T04:30:22Z] Build completed successfully
[2025-10-27T04:30:23Z] Artifacts: bin/app (45.2 MB)
[2025-10-27T04:30:24Z] Step completed with status: Succeeded`,
		"step-3": `[2025-10-27T04:30:25Z] Step started: test
[2025-10-27T04:30:26Z] $ go test ./...
[2025-10-27T04:30:27Z] ok      github.com/org/c8s/cmd/api-server  0.234s
[2025-10-27T04:30:28Z] ok      github.com/org/c8s/pkg/dashboard   0.456s
[2025-10-27T04:30:29Z] ok      github.com/org/c8s/pkg/apis        0.123s
[2025-10-27T04:30:30Z] Coverage: 78.5%
[2025-10-27T04:30:31Z] Step completed with status: Succeeded`,
	}

	for stepID, content := range stepLogs {
		key := fmt.Sprintf("%s/%s", runID, stepID)
		s.logs[key] = content
	}
}

// NoOpLogStorage implements LogStorage with no-op methods (for testing without logs)
type NoOpLogStorage struct{}

// GetStepLogs returns empty log
func (s *NoOpLogStorage) GetStepLogs(_ context.Context, _, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// StreamStepLogs closes channel immediately
func (s *NoOpLogStorage) StreamStepLogs(_ context.Context, _, _ string, linesChan chan<- string) error {
	close(linesChan)
	return nil
}

// GetLogSnapshot returns empty snapshot
func (s *NoOpLogStorage) GetLogSnapshot(_ context.Context, _, _ string, _ int) ([]string, error) {
	return []string{}, nil
}

// GetLogSize returns 0
func (s *NoOpLogStorage) GetLogSize(_ context.Context, _, _ string) (int64, error) {
	return 0, nil
}
