package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/org/c8s/pkg/api/responses"
)

// ServeStatic serves static files from the specified directory
// Usage: router.Handle("/static/*", ServeStatic("cmd/api-server/static"))
func ServeStatic(staticDir string) http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
}

// StaticWithCacheControl serves static files with appropriate cache control headers
func StaticWithCacheControl(staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip /static/ prefix from URL path
		cleanPath := strings.TrimPrefix(r.URL.Path, "/static/")
		// Get file path
		filePath := filepath.Join(staticDir, cleanPath)

		// Determine cache duration based on file type
		ext := filepath.Ext(filePath)
		switch ext {
		case ".html":
			// HTML files - no caching (always fetch latest)
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		case ".css", ".js":
			// CSS and JS - cache for 1 year (these have hashes/versions in production)
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp":
			// Images - cache for 30 days
			w.Header().Set("Cache-Control", "public, max-age=2592000")
		default:
			// Default - cache for 1 day
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}

		// Set MIME type
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".woff", ".woff2":
			w.Header().Set("Content-Type", "font/woff2")
		}

		// Serve the file
		http.ServeFile(w, r, filePath)
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	// For API requests, return JSON error
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","message":"Resource not found"}}`))
		return
	}
	// For HTML requests, render error page
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>404 Not Found</title></head><body><h1>404</h1><p>Page not found</p></body></html>`))
}

// RespondSuccess is a convenience wrapper for dashboard response helper
func RespondSuccess(w http.ResponseWriter, statusCode int, data interface{}) error {
	return responses.RespondSuccess(w, statusCode, data)
}
