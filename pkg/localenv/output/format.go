package output

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Formatter handles output formatting for CLI commands
type Formatter struct {
	useColor bool
}

// NewFormatter creates a new output formatter
func NewFormatter(useColor bool) *Formatter {
	return &Formatter{
		useColor: useColor,
	}
}

// Color codes for terminal output
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

// Success prints a success message with checkmark
func (f *Formatter) Success(message string) {
	if f.useColor {
		fmt.Printf("%s✓%s %s\n", colorGreen, colorReset, message)
	} else {
		fmt.Printf("✓ %s\n", message)
	}
}

// Error prints an error message with cross
func (f *Formatter) Error(message string) {
	if f.useColor {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", colorRed, colorReset, message)
	} else {
		fmt.Fprintf(os.Stderr, "✗ %s\n", message)
	}
}

// Warning prints a warning message
func (f *Formatter) Warning(message string) {
	if f.useColor {
		fmt.Printf("%s⚠%s %s\n", colorYellow, colorReset, message)
	} else {
		fmt.Printf("⚠ %s\n", message)
	}
}

// Info prints an info message
func (f *Formatter) Info(message string) {
	if f.useColor {
		fmt.Printf("%sℹ%s %s\n", colorBlue, colorReset, message)
	} else {
		fmt.Printf("ℹ %s\n", message)
	}
}

// Table formats data as a table
func (f *Formatter) Table(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, header := range headers {
		fmt.Printf("%-*s", widths[i]+2, header)
	}
	fmt.Println()

	// Print separator
	for _, width := range widths {
		fmt.Print(strings.Repeat("-", width+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s", widths[i]+2, cell)
			}
		}
		fmt.Println()
	}
}

// ProgressBar displays a simple progress indicator
func (f *Formatter) ProgressBar(message string, current, total int) {
	percentage := (current * 100) / total
	barLength := 30
	filled := (current * barLength) / total

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLength-filled)

	if f.useColor {
		fmt.Printf("\r%s [%s] %d%% ", message, bar, percentage)
	} else {
		fmt.Printf("\r%s [%s] %d%% ", message, bar, percentage)
	}
}

// Duration formats a time duration for display
func Duration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FileSize formats a file size for display
func FileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

// StatusIndicator returns a status symbol
func StatusIndicator(success bool) string {
	if success {
		return "✓"
	}
	return "✗"
}

// Highlight highlights text with color
func (f *Formatter) Highlight(text string) string {
	if f.useColor {
		return fmt.Sprintf("%s%s%s", colorCyan, text, colorReset)
	}
	return text
}

// ListItems prints a bulleted list
func (f *Formatter) ListItems(items []string) {
	for _, item := range items {
		fmt.Printf("  • %s\n", item)
	}
}

// Section prints a section header
func (f *Formatter) Section(title string) {
	fmt.Println()
	if f.useColor {
		fmt.Printf("%s%s%s\n", colorMagenta, strings.ToUpper(title), colorReset)
	} else {
		fmt.Println(strings.ToUpper(title))
	}
	fmt.Println(strings.Repeat("=", len(title)))
}

// Separator prints a visual separator
func (f *Formatter) Separator() {
	fmt.Println(strings.Repeat("─", 60))
}

// KeyValue prints a key-value pair
func (f *Formatter) KeyValue(key, value string) {
	fmt.Printf("  %-20s: %s\n", key, value)
}
