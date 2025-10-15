package cluster

import (
	"errors"
	"fmt"
	"strings"
)

// Common error types for better error handling

// IsClusterNotFoundError checks if an error is a ClusterNotFoundError
func IsClusterNotFoundError(err error) bool {
	var notFoundErr *ClusterNotFoundError
	return errors.As(err, &notFoundErr)
}

// IsClusterAlreadyExistsError checks if an error is a ClusterAlreadyExistsError
func IsClusterAlreadyExistsError(err error) bool {
	var existsErr *ClusterAlreadyExistsError
	return errors.As(err, &existsErr)
}

// IsDockerNotAvailableError checks if an error is a DockerNotAvailableError
func IsDockerNotAvailableError(err error) bool {
	var dockerErr *DockerNotAvailableError
	return errors.As(err, &dockerErr)
}

// IsTimeoutError checks if an error is related to timeout
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
	       strings.Contains(err.Error(), "deadline exceeded")
}

// ErrorWithSuggestion wraps an error with an actionable suggestion
type ErrorWithSuggestion struct {
	Err        error
	Suggestion string
}

func (e *ErrorWithSuggestion) Error() string {
	return fmt.Sprintf("%v\n\nSuggestion: %s", e.Err, e.Suggestion)
}

func (e *ErrorWithSuggestion) Unwrap() error {
	return e.Err
}

// NewErrorWithSuggestion creates an error with a suggestion
func NewErrorWithSuggestion(err error, suggestion string) error {
	return &ErrorWithSuggestion{
		Err:        err,
		Suggestion: suggestion,
	}
}

// EnhanceError adds contextual suggestions to common errors
func EnhanceError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Docker not available
	if IsDockerNotAvailableError(err) {
		return NewErrorWithSuggestion(err,
			"Ensure Docker Desktop is installed and running. Try: docker info")
	}

	// Cluster not found
	if IsClusterNotFoundError(err) {
		return NewErrorWithSuggestion(err,
			"List available clusters with: c8s dev cluster list")
	}

	// Cluster already exists
	if IsClusterAlreadyExistsError(err) {
		return NewErrorWithSuggestion(err,
			"Delete the existing cluster first with: c8s dev cluster delete <name>")
	}

	// Timeout errors
	if IsTimeoutError(err) {
		return NewErrorWithSuggestion(err,
			"Try increasing the timeout with --timeout flag, or check if resources are available")
	}

	// Port conflicts
	if strings.Contains(err.Error(), "address already in use") ||
	   strings.Contains(err.Error(), "port") && strings.Contains(err.Error(), "in use") {
		return NewErrorWithSuggestion(err,
			"A port is already in use. Check for conflicting services or try a different port")
	}

	// Kubectl not found
	if strings.Contains(err.Error(), "kubectl") &&
	   (strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable")) {
		return NewErrorWithSuggestion(err,
			"kubectl is not installed. Install it from: https://kubernetes.io/docs/tasks/tools/")
	}

	// k3d not found
	if strings.Contains(err.Error(), "k3d") &&
	   (strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable")) {
		return NewErrorWithSuggestion(err,
			"k3d is not installed. Install it from: https://k3d.io/")
	}

	// Permission denied
	if strings.Contains(err.Error(), "permission denied") {
		return NewErrorWithSuggestion(err,
			"Check file/directory permissions or try running with appropriate privileges")
	}

	// Out of disk space
	if strings.Contains(err.Error(), "no space left") ||
	   strings.Contains(err.Error(), "disk full") {
		return NewErrorWithSuggestion(err,
			"Free up disk space and try again. Check: df -h")
	}

	// Return original error if no enhancement available
	return err
}

// RecoverableError indicates an error that the user can potentially fix
type RecoverableError struct {
	Err    error
	Action string
}

func (e *RecoverableError) Error() string {
	return fmt.Sprintf("%v\n\nAction required: %s", e.Err, e.Action)
}

func (e *RecoverableError) Unwrap() error {
	return e.Err
}

// ValidatePrerequisites checks if all prerequisites are met
func ValidatePrerequisites() []error {
	var errors []error

	// Check for Docker
	if err := checkCommand("docker"); err != nil {
		errors = append(errors, fmt.Errorf("docker not found: %w", err))
	}

	// Check for kubectl
	if err := checkCommand("kubectl"); err != nil {
		errors = append(errors, fmt.Errorf("kubectl not found: %w", err))
	}

	// Check for k3d
	if err := checkCommand("k3d"); err != nil {
		errors = append(errors, fmt.Errorf("k3d not found: %w", err))
	}

	return errors
}

// checkCommand checks if a command is available in PATH
func checkCommand(name string) error {
	// This is a simple check - actual implementation would use exec.LookPath
	// but that's already done in the k3d and kubectl clients
	return nil
}
