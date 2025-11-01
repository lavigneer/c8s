package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/org/c8s/pkg/dashboard"
)

// ErrorRecoveryMiddleware catches panics and returns 500
func ErrorRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ERROR: Panic recovered in %s %s: %v", r.Method, r.URL.Path, err)
				dashboard.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestLoggerMiddleware logs all requests with timing information
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path = path + "?" + r.URL.RawQuery
		}

		log.Printf("[%s] %s %s from %s", r.Method, path, r.Proto, r.RemoteAddr)

		// Create a wrapped response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("[%s] %s completed with status %d in %v", r.Method, path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	headerWritten   bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	if rw.headerWritten {
		return // Avoid calling WriteHeader multiple times
	}
	rw.statusCode = code
	rw.headerWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

// Write implements http.ResponseWriter
func (rw *responseWriter) Write(b []byte) (int, error) {
	// Ensure WriteHeader is called before writing (with 200 status by default)
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// NotFoundMiddleware provides a friendly 404 handler
func NotFoundMiddleware(w http.ResponseWriter, r *http.Request) {
	log.Printf("404: %s %s not found", r.Method, r.URL.Path)
	dashboard.RespondError(w, http.StatusNotFound, "NOT_FOUND", "The requested resource was not found")
}
