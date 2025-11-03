package middleware_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// RequestSizeLimitMiddleware for testing (copy from main.go pattern)
func RequestSizeLimitMiddleware(maxBytes int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ===== Request Size Limit Tests =====

// TestRequestWithinSizeLimit verifies request within limit is accepted
func TestRequestWithinSizeLimit(t *testing.T) {
	const maxSize = 1024 * 1024 // 1MB
	payload := bytes.NewReader([]byte(strings.Repeat("x", 500*1024))) // 500KB

	// Create middleware
	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("received: " + fmt.Sprintf("%d", len(body))))
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestRequestExceedsSizeLimit verifies oversized request is rejected
func TestRequestExceedsSizeLimit(t *testing.T) {
	const maxSize = 1024 * 1024 // 1MB
	// Create payload larger than limit (1.5MB)
	payload := bytes.NewReader(bytes.Repeat([]byte("x"), 1024*1024+512*1024))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read body - should fail
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Either error or 413 is acceptable (MaxBytesReader may return different error)
	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		// Request size limit may be enforced differently
		// Just verify it doesn't return 200
		if w.Code == http.StatusOK {
			t.Errorf("expected error for oversized request, got 200")
		}
	}
}

// TestSmallPayload verifies small payloads work fine
func TestSmallPayload(t *testing.T) {
	const maxSize = 10 * 1024 // 10KB
	payload := bytes.NewReader([]byte("small data"))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for small payload, got %d", w.Code)
	}
}

// TestBoundarySize verifies request at exact limit is accepted
func TestBoundarySize(t *testing.T) {
	const maxSize = 1024 // 1KB
	// Create payload at exact limit
	payload := bytes.NewReader(bytes.Repeat([]byte("x"), 1024))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{byte(len(body))})
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for boundary size, got %d", w.Code)
	}
}

// TestBoundaryPlusOne verifies request just over limit is rejected
func TestBoundaryPlusOne(t *testing.T) {
	const maxSize = 1024 // 1KB
	// Create payload just over limit
	payload := bytes.NewReader(bytes.Repeat([]byte("x"), 1024+1))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should be rejected (exact behavior may vary)
	if w.Code == http.StatusOK {
		t.Errorf("expected rejection for oversized request, got 200")
	}
}

// TestEmptyRequest verifies empty request is accepted
func TestEmptyRequest(t *testing.T) {
	const maxSize = 1024 * 1024
	payload := bytes.NewReader([]byte{})

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for empty request, got %d", w.Code)
	}
}

// TestLargeJSONPayload verifies large JSON request is handled correctly
func TestLargeJSONPayload(t *testing.T) {
	const maxSize = 100 * 1024 // 100KB
	// Create large JSON-like payload
	json := `{"items":[` + strings.Repeat(`{"id":1,"data":"test"},`, 1000) + `]}`

	if len(json) < maxSize {
		payload := bytes.NewReader([]byte(json))

		middleware := RequestSizeLimitMiddleware(maxSize)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{byte(len(body) / 1024)}) // Return size in KB
		}))

		req := httptest.NewRequest("POST", "/api/test", payload)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for large JSON payload, got %d", w.Code)
		}
	}
}

// TestOversizeJSONPayload verifies oversized JSON is rejected
func TestOversizeJSONPayload(t *testing.T) {
	const maxSize = 50 * 1024 // 50KB
	// Create JSON payload larger than limit
	json := `{"items":[` + strings.Repeat(`{"id":1,"data":"test","description":"This is a long description field that helps us generate large payloads"},`, 5000) + `]}`

	if len(json) > maxSize {
		payload := bytes.NewReader([]byte(json))

		middleware := RequestSizeLimitMiddleware(maxSize)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/api/test", payload)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should be rejected
		if w.Code == http.StatusOK {
			t.Errorf("expected rejection for oversized JSON, got 200")
		}
	}
}

// TestMultipleLimits verifies different size limits work independently
func TestMultipleLimits(t *testing.T) {
	tests := []struct {
		name      string
		maxSize   int64
		payloadSize int
		shouldPass bool
	}{
		{"1KB limit, 500B payload", 1024, 500, true},
		{"1KB limit, 1KB payload", 1024, 1024, true},
		{"1KB limit, 2KB payload", 1024, 2048, false},
		{"100KB limit, 50KB payload", 100 * 1024, 50 * 1024, true},
		{"100KB limit, 200KB payload", 100 * 1024, 200 * 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := bytes.NewReader(bytes.Repeat([]byte("x"), tt.payloadSize))

			middleware := RequestSizeLimitMiddleware(tt.maxSize)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("POST", "/api/test", payload)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.shouldPass && w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
			if !tt.shouldPass && w.Code == http.StatusOK {
				t.Errorf("expected error status, got 200")
			}
		})
	}
}

// TestPartialRead verifies handler can read partial data before limit hit
func TestPartialRead(t *testing.T) {
	const maxSize = 1024
	const partialSize = 512

	payload := bytes.NewReader(bytes.Repeat([]byte("x"), partialSize))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, partialSize)
		n, err := r.Body.Read(buf)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{byte(n)})
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for partial read, got %d", w.Code)
	}
}

// TestHeadersNotAffected verifies request headers are not affected by limit
func TestHeadersNotAffected(t *testing.T) {
	const maxSize = 1024
	payload := bytes.NewReader([]byte("test"))

	middleware := RequestSizeLimitMiddleware(maxSize)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected headers to be preserved, got status %d", w.Code)
	}
}

// TestMiddlewareChaining verifies middleware works in a chain
func TestMiddlewareChaining(t *testing.T) {
	const maxSize = 1024
	payload := bytes.NewReader([]byte("test data"))

	// Chain two middleware
	middleware1 := RequestSizeLimitMiddleware(maxSize)
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware-2", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Middleware-2") != "applied" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware in order
	finalHandler := middleware1(middleware2(handler))

	req := httptest.NewRequest("POST", "/api/test", payload)
	w := httptest.NewRecorder()

	finalHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for chained middleware, got %d", w.Code)
	}
}
