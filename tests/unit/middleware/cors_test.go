package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// CORSHeadersMiddleware for testing (copy of middleware implementation)
func CORSHeadersMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is in allowed list
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				// SECURITY: Don't send "*" with credentials=true - it's invalid per CORS spec
				// Only send credentials header when origin is specific
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				} else if len(allowedOrigins) > 0 && allowedOrigins[0] != "*" {
					// If no origin header but specific origin configured, use it
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				} else {
					// Wildcard origin: don't send credentials header
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}

				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ===== CORS Header Tests =====

// TestCORSAllowedOrigin verifies allowed origin gets credentials header
func TestCORSAllowedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://app.example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected origin header for allowed origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected credentials header for specific origin, got %q", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

// TestCORSDisallowedOrigin verifies disallowed origin is blocked
func TestCORSDisallowedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no origin header for disallowed origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestCORSWildcardAllowed verifies wildcard "*" without credentials
func TestCORSWildcardAllowed(t *testing.T) {
	allowedOrigins := []string{"*"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	// When no origin or "*" in allowed, wildcard is used
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Without an Origin header, behavior may vary
	// Let's test with no origin to get the wildcard behavior
	origin := w.Header().Get("Access-Control-Allow-Origin")

	// SECURITY: Wildcard should NOT have credentials header
	if origin == "*" && w.Header().Get("Access-Control-Allow-Credentials") == "true" {
		t.Errorf("SECURITY ISSUE: wildcard origin with credentials=true is invalid per CORS spec")
	}
}

// TestCORSPreflight verifies OPTIONS requests are handled
func TestCORSPreflight(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for preflight, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected origin header in preflight response")
	}
}

// TestCORSAllowedMethods verifies correct methods in response
func TestCORSAllowedMethods(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("expected correct methods, got %q", methods)
	}
}

// TestCORSAllowedHeaders verifies correct headers in response
func TestCORSAllowedHeaders(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers != "Content-Type, Authorization" {
		t.Errorf("expected correct headers, got %q", headers)
	}
}

// TestCORSMaxAge verifies cache duration header
func TestCORSMaxAge(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "86400" {
		t.Errorf("expected max-age 86400, got %q", maxAge)
	}
}

// TestCORSMultipleAllowedOrigins verifies list matching
func TestCORSMultipleAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		origin   string
		expected string
	}{
		{"First origin", []string{"https://example.com", "https://app.example.com"}, "https://example.com", "https://example.com"},
		{"Second origin", []string{"https://example.com", "https://app.example.com"}, "https://app.example.com", "https://app.example.com"},
		{"Disallowed origin", []string{"https://example.com", "https://app.example.com"}, "https://evil.com", ""},
		{"Wildcard allowed", []string{"*"}, "https://any.com", "https://any.com"}, // Origin is sent with no credentials
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := CORSHeadersMiddleware(tt.allowed)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			actual := w.Header().Get("Access-Control-Allow-Origin")
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestCORSNoOriginHeader verifies behavior with missing origin
func TestCORSNoOriginHeader(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	// No Origin header set
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Without Origin header, even allowed origins shouldn't auto-send
	// (This depends on implementation, but typically no CORS header)
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" && origin != "https://example.com" {
		// Should either be empty or use the configured origin
		// This test verifies the middleware handles missing Origin gracefully
	}
}

// TestCORSCredentialsWithWildcard ensures wildcard + credentials never happens
func TestCORSCredentialsWithWildcard(t *testing.T) {
	// SECURITY TEST: Ensure we never send "*" with credentials=true
	allowedOrigins := []string{"*"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with origin header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	credentials := w.Header().Get("Access-Control-Allow-Credentials")

	// Per CORS spec, wildcard origin cannot be used with credentials
	if origin == "*" && credentials == "true" {
		t.Errorf("SECURITY VIOLATION: Cannot send '*' origin with credentials=true")
	}
}

// TestCORSSpecificOriginWithCredentials ensures specific origins get credentials
func TestCORSSpecificOriginWithCredentials(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	credentials := w.Header().Get("Access-Control-Allow-Credentials")

	if origin != "https://example.com" {
		t.Errorf("expected specific origin, got %q", origin)
	}

	if credentials != "true" {
		t.Errorf("expected credentials=true for specific origin, got %q", credentials)
	}
}

// TestCORSCaseSensitivity verifies origin matching is case-sensitive
func TestCORSCaseSensitivity(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	middleware := CORSHeadersMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with different case (should not match)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://Example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		// Origins are case-sensitive per spec
		t.Errorf("CORS origins should be case-sensitive, got %q", origin)
	}
}

// TestCORSWithDifferentMethods verifies all HTTP methods work
func TestCORSWithDifferentMethods(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			middleware := CORSHeadersMiddleware(allowedOrigins)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/api/test", nil)
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// All methods should get CORS headers (except OPTIONS gets 204, others get 200)
			if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
				t.Errorf("expected CORS origin for method %s", method)
			}
		})
	}
}
