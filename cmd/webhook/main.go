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
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = c8sv1alpha1.AddToScheme(scheme)
}

func main() {
	var (
		port       int
		kubeconfig string
		logLevel   string
	)

	flag.IntVar(&port, "port", 8080, "Port to listen on for webhook requests")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (leave empty for in-cluster config)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Setup logging
	opts := zap.Options{
		Development: logLevel == "debug",
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Starting C8S Webhook Service",
		"port", port,
		"version", "v1alpha1",
	)

	// Setup Kubernetes client
	config, err := getKubeConfig(kubeconfig)
	if err != nil {
		setupLog.Error(err, "Failed to get kubeconfig")
		os.Exit(1)
	}

	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Failed to create Kubernetes client")
		os.Exit(1)
	}

	setupLog.Info("Successfully connected to Kubernetes API")

	// Create webhook handlers
	githubHandler := webhook.NewGitHubHandler(k8sClient)
	gitlabHandler := webhook.NewGitLabHandler(k8sClient)
	bitbucketHandler := webhook.NewBitbucketHandler(k8sClient)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Webhook endpoints
	mux.HandleFunc("/webhooks/github", githubHandler.Handle)
	mux.HandleFunc("/webhooks/gitlab", gitlabHandler.Handle)
	mux.HandleFunc("/webhooks/bitbucket", bitbucketHandler.Handle)

	// Health check endpoints
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)

	// Root endpoint
	mux.HandleFunc("/", handleRoot)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		setupLog.Info("Webhook service listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			setupLog.Error(err, "Failed to start HTTP server")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	setupLog.Info("Shutting down webhook service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		setupLog.Error(err, "Server forced to shutdown")
	}

	setupLog.Info("Webhook service stopped")
}

// getKubeConfig returns Kubernetes REST config
func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		// Use kubeconfig file
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Use in-cluster config
	return rest.InClusterConfig()
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReady handles readiness check requests
func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// handleRoot handles root path requests
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
  "service": "c8s-webhook",
  "version": "v1alpha1",
  "endpoints": {
    "github": "/webhooks/github",
    "gitlab": "/webhooks/gitlab",
    "bitbucket": "/webhooks/bitbucket",
    "health": "/health",
    "ready": "/ready"
  }
}`))
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := log.FromContext(r.Context())

		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)

		next.ServeHTTP(w, r)

		logger.Info("HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}
