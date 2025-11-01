package dashboard

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

// NewInMemoryLogStorage creates a new in-memory log storage with demo data
func NewInMemoryLogStorage() *InMemoryLogStorage {
	storage := &InMemoryLogStorage{
		logs: make(map[string]string),
	}
	storage.populateDemoLogs()
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
func (s *InMemoryLogStorage) SetLog(runID, stepID string, content string) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	s.logs[key] = content
}

// GetStepLogs returns a reader for step logs
func (s *InMemoryLogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]

	// If logs not found for this specific run, generate demo logs
	if !ok {
		content = s.generateDemoLogsForStep(runID, stepID)
	}

	return io.NopCloser(strings.NewReader(content)), nil
}

// StreamStepLogs streams logs line-by-line to channel
func (s *InMemoryLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]

	// If logs not found for this specific run, generate demo logs
	if !ok {
		content = s.generateDemoLogsForStep(runID, stepID)
	}

	go func() {
		defer close(linesChan)
		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case linesChan <- scanner.Text():
			}
		}
	}()

	return nil
}

// GetLogSnapshot returns last N lines of logs
func (s *InMemoryLogStorage) GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]

	// If logs not found for this specific run, generate demo logs
	if !ok {
		content = s.generateDemoLogsForStep(runID, stepID)
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
func (s *InMemoryLogStorage) GetLogSize(ctx context.Context, runID, stepID string) (int64, error) {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]

	// If logs not found for this specific run, generate demo logs
	if !ok {
		content = s.generateDemoLogsForStep(runID, stepID)
	}

	return int64(len(content)), nil
}

// generateDemoLogsForStep generates demo logs for any step based on step name
func (s *InMemoryLogStorage) generateDemoLogsForStep(runID, stepID string) string {
	// Generate logs based on step name
	stepLogs := map[string]string{
		"step-1": fmt.Sprintf(`[2025-10-27T04:30:10Z] Step started: checkout
[2025-10-27T04:30:11Z] $ git clone https://github.com/example/repo.git
[2025-10-27T04:30:12Z] Cloning into 'repo'...
[2025-10-27T04:30:13Z] remote: Counting objects: 1000, done
[2025-10-27T04:30:14Z] Receiving objects: 100%% (1000/1000)
[2025-10-27T04:30:15Z] Step completed with status: Succeeded`),
		"step-2": fmt.Sprintf(`[2025-10-27T04:30:16Z] Step started: build
[2025-10-27T04:30:17Z] $ echo 'Starting build process'
[2025-10-27T04:30:17Z] Starting build process
[2025-10-27T04:30:18Z] $ go build -o bin/app ./cmd/api-server
[2025-10-27T04:30:19Z] go: downloading github.com/go-chi/chi/v5
[2025-10-27T04:30:20Z] go: downloading sigs.k8s.io/controller-runtime
[2025-10-27T04:30:22Z] Build completed successfully
[2025-10-27T04:30:23Z] Artifacts: bin/app (45.2 MB)
[2025-10-27T04:30:24Z] Step completed with status: Succeeded`),
		"step-3": fmt.Sprintf(`[2025-10-27T04:30:25Z] Step started: test
[2025-10-27T04:30:26Z] $ go test ./...
[2025-10-27T04:30:27Z] ok      github.com/org/c8s/cmd/api-server  0.234s
[2025-10-27T04:30:28Z] ok      github.com/org/c8s/pkg/dashboard   0.456s
[2025-10-27T04:30:29Z] ok      github.com/org/c8s/pkg/apis        0.123s
[2025-10-27T04:30:30Z] Coverage: 78.5%%
[2025-10-27T04:30:31Z] Step completed with status: Succeeded`),
	}

	// Check if we have a specific template for this step
	if logs, ok := stepLogs[stepID]; ok {
		return logs
	}

	// Default demo logs for unknown steps
	return fmt.Sprintf(`[2025-10-27T04:30:00Z] Step started: %s
[2025-10-27T04:30:01Z] $ echo 'Running step %s'
[2025-10-27T04:30:01Z] Running step %s
[2025-10-27T04:30:02Z] Processing...
[2025-10-27T04:30:03Z] Step completed with status: Succeeded`, stepID, stepID, stepID)
}

// NoOpLogStorage implements LogStorage with no-op methods (for testing without logs)
type NoOpLogStorage struct{}

// GetStepLogs returns empty log
func (s *NoOpLogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// StreamStepLogs closes channel immediately
func (s *NoOpLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	close(linesChan)
	return nil
}

// GetLogSnapshot returns empty snapshot
func (s *NoOpLogStorage) GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error) {
	return []string{}, nil
}

// GetLogSize returns 0
func (s *NoOpLogStorage) GetLogSize(ctx context.Context, runID, stepID string) (int64, error) {
	return 0, nil
}
