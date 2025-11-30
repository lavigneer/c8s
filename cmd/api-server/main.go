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

// Package main implements the C8S API server with web dashboard and REST endpoints.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/org/c8s/pkg/api/handlers"
	apimiddleware "github.com/org/c8s/pkg/api/middleware"
	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/auth"
	"github.com/org/c8s/pkg/dashboard"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	var (
		port      = flag.String("port", ":8080", "Port to listen on")
		baseDir   = flag.String("base-dir", ".", "Base directory for templates and static files")
		tlsCert   = flag.String("tls-cert", os.Getenv("TLS_CERT_PATH"), "Path to TLS certificate")
		tlsKey    = flag.String("tls-key", os.Getenv("TLS_KEY_PATH"), "Path to TLS key")
		tlsPort   = flag.String("tls-port", ":8443", "TLS port to listen on")
		enableTLS = flag.Bool("enable-tls", false, "Enable HTTPS/TLS")
	)
	flag.Parse()

	// Load dashboard templates
	if err := dashboard.LoadTemplates(*baseDir); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Initialize authentication
	// Create validator based on configuration (NoOp for development)
	var validator auth.ValidatorInterface
	authConfig := auth.NewConfigFromEnv()

	if authConfig.Mode == auth.ModeNone {
		// Development mode - accepts any token
		validator = auth.NewNoOpValidator()
		log.Println("Auth validator initialized (development mode)")
	} else {
		// Production mode - validate JWT tokens
		v, err := auth.NewValidator(authConfig)
		if err != nil {
			log.Fatalf("Failed to create auth validator: %v", err)
		}
		validator = v
		log.Println("Auth validator initialized (JWT mode)")
	}

	// Create auth middleware with dependency injection
	authMiddleware := auth.NewMiddleware(validator)

	// Initialize Kubernetes client
	log.Println("Initializing Kubernetes client...")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes config: %v", err)
	}

	// Create a scheme and add C8S types
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		log.Fatalf("Failed to add v1alpha1 types to scheme: %v", err)
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	k8sClient := dashboard.NewK8sClient(c)
	handlers.InitK8sClient(k8sClient)
	log.Println("Kubernetes client initialized successfully")

	// Create rate limiter for public API endpoints
	// 100 requests per second per IP, burst of 200 (for general API protection)
	apiRateLimiter := apimiddleware.NewRateLimiter(100.0, 200)
	log.Println("API rate limiter initialized (100 rps, burst 200)")

	// Create router
	router := chi.NewRouter()

	// Global middleware
	router.Use(handlers.ErrorRecoveryMiddleware)
	router.Use(handlers.RequestLoggerMiddleware)
	router.Use(middleware.RequestID)
	router.Use(SecurityHeadersMiddleware)
	router.Use(RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB limit

	// Static files (no auth required)
	router.Handle("/static/*", handlers.StaticWithCacheControl(*baseDir+"/static"))
	router.HandleFunc("/health", healthHandler)

	// Login route (no auth required)
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("GET /logout", handlers.LogoutHandler)
	router.HandleFunc("GET /", redirectToLogin)

	// Public v1 API endpoints (no auth required - for external integrations like GitHub Actions)
	// Apply rate limiting to protect against abuse
	router.Group(func(r chi.Router) {
		r.Use(apiRateLimiter.Middleware)
		r.Get("/api/v1/pipelines/{name}/status", handlers.GetPipelineStatusHandler)
		r.Get("/api/v1/pipelines/{name}/logs", handlers.GetPipelineLogsHandler)
		r.Get("/v1/pipelines/{name}/status", handlers.GetPipelineStatusHandler)
		r.Get("/v1/pipelines/{name}/logs", handlers.GetPipelineLogsHandler)
	})

	// Dashboard routes (protected by auth)
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Handler)

		// Dashboard pages (US1)
		r.Get("/dashboard", handlers.DashboardHandler)
		r.Get("/dashboard/projects", handlers.ProjectsHandler)

		// Dashboard pages (US2)
		r.Get("/dashboard/runs/{runId}", handlers.PipelineRunDetailsHandler)

		// API endpoints - Projects (US4)
		r.Get("/api/projects", handlers.ListProjectsHandler)
		r.Post("/api/projects", handlers.CreateProjectHandler)
		r.Delete("/api/projects/{projectId}", handlers.DeleteProjectHandler)
		r.Get("/api/projects/{projectId}/webhook", handlers.GetWebhookConfigHandler)

		// API endpoints - Pipeline Runs (US1)
		r.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler)

		// API endpoints - Pipeline Branches (US3)
		r.Get("/api/projects/{projectId}/branches", handlers.ListBranchesHandler)

		// API endpoints - Pipeline Run Details (US2)
		r.Get("/api/runs/{runId}", handlers.GetPipelineRunHandler)

		// Log Streaming endpoints (US2)
		r.Get("/api/runs/{runId}/steps/{stepId}/logs", handlers.LogStreamHandler)
		r.Get("/api/runs/{runId}/steps/{stepId}/logs/text", handlers.GetLogsHandler)
		r.Get("/api/runs/{runId}/steps/{stepId}/logs/snapshot", handlers.GetLogSnapshotHandler)

		// SSE endpoints (US1)
		r.Get("/api/projects/{projectId}/runs/updates", handlers.PipelineUpdatesSSEHandler)

		// API endpoints - Artifacts (US5)
		r.Get("/api/runs/{runId}/artifacts", handlers.ListArtifactsHandler)
		r.Get("/api/artifacts/{artifactId}", handlers.GetArtifactHandler)
		r.Get("/api/artifacts/{artifactId}/download", handlers.DownloadArtifactHandler)
		r.Get("/api/artifacts/{artifactId}/preview", handlers.PreviewArtifactHandler)
		r.Delete("/api/artifacts/{artifactId}", handlers.DeleteArtifactHandler)

		// API endpoints - Export
		r.Get("/api/exports/runs", handlers.ExportPipelineRunsHandler)
	})

	// 404 handler
	router.NotFound(http.HandlerFunc(handlers.NotFoundMiddleware))

	// Start HTTP server
	log.Printf("Starting API server on %s", *port)
	go func() {
		if err := http.ListenAndServe(*port, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server if enabled
	if *enableTLS && *tlsCert != "" && *tlsKey != "" {
		log.Printf("Starting HTTPS server on %s", *tlsPort)
		if err := http.ListenAndServeTLS(*tlsPort, *tlsCert, *tlsKey, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS server error: %v", err)
		}
	}

	// Keep server running
	select {}
}

// healthHandler returns a simple health check response
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"healthy","service":"api-server"}`)
}

// redirectToLogin redirects root to login page
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimitMiddleware enforces maximum request body size
// This prevents DoS attacks via oversized payloads
func RequestSizeLimitMiddleware(maxBytes int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}
