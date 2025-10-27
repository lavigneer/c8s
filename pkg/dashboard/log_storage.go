package dashboard

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

// NewInMemoryLogStorage creates a new in-memory log storage
func NewInMemoryLogStorage() *InMemoryLogStorage {
	return &InMemoryLogStorage{
		logs: make(map[string]string),
	}
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
	if !ok {
		return nil, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}
	return io.NopCloser(io.Reader(bufio.NewReader(io.StringReader(content)))), nil
}

// StreamStepLogs streams logs line-by-line to channel
func (s *InMemoryLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	key := fmt.Sprintf("%s/%s", runID, stepID)
	content, ok := s.logs[key]
	if !ok {
		return fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}

	go func() {
		defer close(linesChan)
		scanner := bufio.NewScanner(io.StringReader(content))
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
	if !ok {
		return nil, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}

	var result []string
	scanner := bufio.NewScanner(io.StringReader(content))
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
	if !ok {
		return 0, fmt.Errorf("logs not found for %s/%s", runID, stepID)
	}
	return int64(len(content)), nil
}

// NoOpLogStorage implements LogStorage with no-op methods (for testing without logs)
type NoOpLogStorage struct{}

// GetStepLogs returns empty log
func (s *NoOpLogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	return io.NopCloser(io.Reader(bufio.NewReader(io.StringReader("")))), nil
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
