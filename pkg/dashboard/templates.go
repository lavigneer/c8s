package dashboard

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var Templates *template.Template

// LoadTemplates parses all templates from templates/ directory
func LoadTemplates(basePath string) error {
	pattern := filepath.Join(basePath, "templates/**/*.html")
	var err error
	Templates, err = template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	return nil
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
	return Templates.ExecuteTemplate(w, name, data)
}
