package dashboard

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/Masterminds/sprig/v3"
	"github.com/org/c8s/pkg/dashboard/helpers"
)

// Templates is the template set for rendering HTML pages.
var Templates *template.Template

// LoadTemplates parses all templates from templates/ directory with custom functions
func LoadTemplates(basePath string) error {
	// Start with Sprig's template functions (provides add, sub, mul, div, seq, etc.)
	funcMap := sprig.FuncMap()

	// Add custom template helpers
	customFuncs := template.FuncMap{
		"formatTime":       FormatTimestamp,
		"formatDuration":   FormatDuration,
		"formatBytes":      FormatBytes,
		"formatRelative":   FormatRelativeTime,
		"isRecent":         IsWithinLastHour,
		"isWithinLastHour": IsWithinLastHour,
		"isWithinLastDay":  IsWithinLastDay,
		"isWithinLastWeek": IsWithinLastWeek,
		"slice":            sliceString,
		"dict":             helpers.Dict,
		// Comparison functions (not provided by Sprig)
		"eq": eq,
		"ne": ne,
		"lt": lt,
		"le": le,
		"gt": gt,
		"ge": ge,
	}

	// Merge custom functions into the Sprig function map
	for name, fn := range customFuncs {
		funcMap[name] = fn
	}

	// Go's filepath.Glob doesn't support ** for recursive matching
	// Parse multiple patterns to handle all nested directories
	patterns := []string{
		filepath.Join(basePath, "templates", "*.html"),
		filepath.Join(basePath, "templates", "*", "*.html"),
		filepath.Join(basePath, "templates", "*", "*", "*.html"),
	}

	var err error
	tmpl := template.New("").Funcs(funcMap)

	for _, pattern := range patterns {
		tmpl, err = tmpl.ParseGlob(pattern)
		if err != nil && !isNoMatchErr(err) {
			return fmt.Errorf("failed to parse templates with pattern %s: %w", pattern, err)
		}
	}

	Templates = tmpl
	return nil
}

// isNoMatchErr checks if an error is a "no files matched" error
func isNoMatchErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "html/template: pattern matches no files"
}

// Unexported helper functions that can be used in templates
func sliceString(s string, start, end int) string {
	if start < 0 || end > len(s) || start > end {
		return s
	}
	return s[start:end]
}

func eq(a, b interface{}) bool {
	return a == b
}

func ne(a, b interface{}) bool {
	return a != b
}

func lt(a, b interface{}) bool {
	switch av := a.(type) {
	case int:
		return av < b.(int)
	case float64:
		return av < b.(float64)
	case string:
		return av < b.(string)
	}
	return false
}

func le(a, b interface{}) bool {
	return lt(a, b) || eq(a, b)
}

func gt(a, b interface{}) bool {
	switch av := a.(type) {
	case int:
		return av > b.(int)
	case float64:
		return av > b.(float64)
	case string:
		return av > b.(string)
	}
	return false
}

func ge(a, b interface{}) bool {
	return gt(a, b) || eq(a, b)
}

// IsHTMXRequest checks if request is from HTMX (partial render)
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if Templates == nil {
		return fmt.Errorf("templates not loaded")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Don't call WriteHeader here - let it be called implicitly when template writes
	return Templates.ExecuteTemplate(w, name, data)
}
