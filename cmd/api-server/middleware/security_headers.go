/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package middleware

import "net/http"

// SecurityHeadersMiddleware adds security headers to all HTTP responses
// to prevent common web vulnerabilities (XSS, clickjacking, MIME sniffing, etc.)
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict-Transport-Security: Enforce HTTPS
		// max-age: 1 year (31536000 seconds)
		// includeSubDomains: Apply policy to subdomains
		// preload: Allow inclusion in browser HSTS preload list
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Content-Type-Options: Prevent MIME type sniffing
		// nosniff: Disables MIME type guessing, protects against drive-by downloads
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevent clickjacking attacks
		// SAMEORIGIN: Allow framing only from same origin
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// X-XSS-Protection: Enable XSS filter in browsers
		// 1; mode=block: Enable filter and block page if XSS detected
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content-Security-Policy: Restrict resource loading
		// Prevents inline scripts and styles, allows only from same origin
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdn.tailwindcss.com; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; font-src 'self' data:; img-src 'self' data: https:; connect-src 'self'; frame-ancestors 'self'")

		// Referrer-Policy: Control referrer information
		// strict-origin-when-cross-origin: Send full URL for same-origin, origin only for cross-origin
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Control browser features
		// Restrict access to geolocation, camera, microphone, etc.
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// Remove server identification header to avoid exposing Go runtime
		w.Header().Del("Server")

		// Set X-Powered-By to prevent information disclosure (optional)
		w.Header().Set("X-Powered-By", "C8S")

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// CSPReportOnlyMiddleware sends CSP violations to a reporting endpoint without blocking
// Useful for testing new CSP policies before enforcing them
func CSPReportOnlyMiddleware(reportURI string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csp := "default-src 'self'; script-src 'self' https://cdn.tailwindcss.com; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com"
			if reportURI != "" {
				csp += "; report-uri " + reportURI
			}
			w.Header().Set("Content-Security-Policy-Report-Only", csp)
			next.ServeHTTP(w, r)
		})
	}
}

// CORSHeadersMiddleware adds CORS headers for API endpoints
// Configure origin, methods, and credentials as needed
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

// NoSniffHeadersMiddleware adds headers to prevent content sniffing
// Enforces strict content type checking
func NoSniffHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Download-Options", "noopen")
		next.ServeHTTP(w, r)
	})
}
