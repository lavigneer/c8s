package cluster

import (
	"fmt"
	"os"
	"time"
)

// Logger provides structured logging for cluster operations
type Logger struct {
	verbose bool
	quiet   bool
}

// NewLogger creates a new logger
func NewLogger(verbose, quiet bool) *Logger {
	return &Logger{
		verbose: verbose,
		quiet:   quiet,
	}
}

// Debug logs debug-level messages (only shown in verbose mode)
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.verbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] DEBUG: %s\n", timestamp, fmt.Sprintf(format, args...))
}

// Info logs informational messages
func (l *Logger) Info(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	fmt.Printf(format+"\n", args...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", fmt.Sprintf(format, args...))
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, args...))
}

// Success logs success messages
func (l *Logger) Success(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	fmt.Printf("✓ %s\n", fmt.Sprintf(format, args...))
}

// Step logs a step in a multi-step process
func (l *Logger) Step(step int, total int, description string) {
	if l.quiet {
		return
	}
	fmt.Printf("[%d/%d] %s\n", step, total, description)
}
