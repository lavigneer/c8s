package dashboard

import (
	"fmt"
	"html/template"
	"time"
)

// FormatTimestamp formats time.Time to relative ("2 hours ago") or absolute format
func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "—"
	}

	now := time.Now()
	diff := now.Sub(t)

	// Show relative time if within last week
	if diff < 7*24*time.Hour {
		return FormatRelativeTime(diff)
	}

	// Show absolute time for older dates
	return t.Format("Jan 2, 15:04")
}

// FormatRelativeTime formats a duration as relative time ("2 hours ago")
func FormatRelativeTime(duration time.Duration) string {
	if duration < 0 {
		duration = -duration
	}

	switch {
	case duration < time.Minute:
		seconds := int(duration.Seconds())
		if seconds == 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

// FormatDuration formats duration in seconds to readable format ("2m 30s", "1h 15m")
func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0s"
	}

	duration := time.Duration(seconds) * time.Second

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	if minutes > 0 {
		if secs == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}

	return fmt.Sprintf("%ds", secs)
}

// FormatBytes formats byte count to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsRecent returns true if timestamp is within the specified duration
func IsRecent(t time.Time, duration time.Duration) bool {
	if t.IsZero() {
		return false
	}
	return time.Since(t) < duration
}

// IsWithinLastHour returns true if timestamp is within last hour
func IsWithinLastHour(t time.Time) bool {
	return IsRecent(t, time.Hour)
}

// IsWithinLastDay returns true if timestamp is within last 24 hours
func IsWithinLastDay(t time.Time) bool {
	return IsRecent(t, 24*time.Hour)
}

// IsWithinLastWeek returns true if timestamp is within last week
func IsWithinLastWeek(t time.Time) bool {
	return IsRecent(t, 7*24*time.Hour)
}

// TemplateFuncMap returns a map of template functions for use in Go templates
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatTime":       FormatTimestamp,
		"formatDuration":   FormatDuration,
		"formatBytes":      FormatBytes,
		"formatRelative":   FormatRelativeTime,
		"isRecent":         IsWithinLastHour,
		"isWithinLastHour": IsWithinLastHour,
		"isWithinLastDay":  IsWithinLastDay,
		"isWithinLastWeek": IsWithinLastWeek,
		"slice":            sliceString,
		"eq":               eq,
		"ne":               ne,
		"lt":               lt,
		"le":               le,
		"gt":               gt,
		"ge":               ge,
		"add":              add,
		"sub":              sub,
		"mul":              mul,
		"div":              divInt,
	}
}

// add returns the sum of two integers
func add(a, b int) int {
	return a + b
}

// sub returns the difference of two integers
func sub(a, b int) int {
	return a - b
}

// mul returns the product of two integers
func mul(a, b int) int {
	return a * b
}

// divInt returns the quotient of two integers
func divInt(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}
