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

	"github.com/org/c8s/cmd/api-server/handlers"
	"github.com/org/c8s/pkg/apis/v1alpha1"
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

	// Create router
	router := chi.NewRouter()

	// Global middleware
	router.Use(handlers.ErrorRecoveryMiddleware)
	router.Use(handlers.RequestLoggerMiddleware)
	router.Use(middleware.RequestID)
	router.Use(SecurityHeadersMiddleware)

	// Static files (no auth required)
	router.Handle("/static/*", handlers.StaticWithCacheControl(*baseDir + "/static"))
	router.HandleFunc("/health", healthHandler)

	// Login route (no auth required)
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("GET /", redirectToLogin)

	// Dashboard routes (protected by auth)
	router.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware)

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
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"api-server"}`)
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
