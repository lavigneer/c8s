package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/org/c8s/pkg/metrics"
)

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)

		// Normalize path to avoid cardinality explosion
		path := normalizePath(r.URL.Path)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			path,
			statusCode,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			path,
		).Observe(duration)
	})
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// WriteHeader captures the status code
func (m *metricsResponseWriter) WriteHeader(code int) {
	if m.headerWritten {
		return
	}
	m.statusCode = code
	m.headerWritten = true
	m.ResponseWriter.WriteHeader(code)
}

// Write ensures WriteHeader is called
func (m *metricsResponseWriter) Write(b []byte) (int, error) {
	if !m.headerWritten {
		m.WriteHeader(http.StatusOK)
	}
	return m.ResponseWriter.Write(b)
}

// normalizePath reduces path cardinality for metrics
// Converts paths like /api/runs/123/logs to /api/runs/:id/logs
func normalizePath(path string) string {
	// Common API patterns
	patterns := map[string]string{
		"/api/runs/":      "/api/runs/:id",
		"/api/projects/":  "/api/projects/:id",
		"/api/artifacts/": "/api/artifacts/:id",
	}

	for prefix, normalized := range patterns {
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			// Check if this looks like an ID path
			remaining := path[len(prefix):]
			if len(remaining) > 0 && (remaining[0] >= '0' && remaining[0] <= '9' || remaining[0] >= 'a' && remaining[0] <= 'z') {
				// Find next slash or end
				for i, c := range remaining {
					if c == '/' {
						return normalized + remaining[i:]
					}
				}
				return normalized
			}
		}
	}

	return path
}
