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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/api/handlers"
	"github.com/org/c8s/pkg/api/middleware"
)

var (
	port            int
	kubeconfig      string
	enableDashboard bool
	enableCORS      bool
)

func init() {
	flag.IntVar(&port, "port", 8080, "API server port")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (leave empty for in-cluster config)")
	flag.BoolVar(&enableDashboard, "enable-dashboard", false, "Enable HTMX dashboard")
	flag.BoolVar(&enableCORS, "enable-cors", true, "Enable CORS middleware")
}

func main() {
	flag.Parse()

	// Setup logger
	opts := zap.Options{
		Development: true,
	}
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	ctx := log.IntoContext(context.Background(), logger)

	logger.Info("Starting C8S API Server",
		"port", port,
		"dashboard", enableDashboard,
		"cors", enableCORS,
	)

	// Create Kubernetes client config
	config, err := getKubeConfig(kubeconfig)
	if err != nil {
		logger.Error(err, "Failed to get kubeconfig")
		os.Exit(1)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create Kubernetes client")
		os.Exit(1)
	}

	// Create controller-runtime client for CRDs
	scheme := runtime.NewScheme()
	if err := c8sv1alpha1.AddToScheme(scheme); err != nil {
		logger.Error(err, "Failed to add c8s types to scheme")
		os.Exit(1)
	}

	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		logger.Error(err, "Failed to create controller-runtime client")
		os.Exit(1)
	}

	// Create HTTP router
	mux := http.NewServeMux()

	// Initialize handlers
	pipelineConfigHandler := handlers.NewPipelineConfigHandler(k8sClient)
	pipelineRunHandler := handlers.NewPipelineRunHandler(k8sClient)
	logsHandler := handlers.NewLogsHandler(clientset, k8sClient)

	// Register API routes
	// PipelineConfig endpoints
	mux.HandleFunc("/api/v1/namespaces/{namespace}/pipelineconfigs", pipelineConfigHandler.HandlePipelineConfigs)
	mux.HandleFunc("/api/v1/namespaces/{namespace}/pipelineconfigs/{name}", pipelineConfigHandler.HandlePipelineConfig)

	// PipelineRun endpoints
	mux.HandleFunc("/api/v1/namespaces/{namespace}/pipelineruns", pipelineRunHandler.HandlePipelineRuns)
	mux.HandleFunc("/api/v1/namespaces/{namespace}/pipelineruns/{name}", pipelineRunHandler.HandlePipelineRun)

	// Logs endpoints
	mux.HandleFunc("/api/v1/namespaces/{namespace}/pipelineruns/{name}/logs/{step}", logsHandler.HandleStepLogs)

	// Dashboard routes (if enabled)
	if enableDashboard {
		logger.Info("Dashboard enabled")
		dashboardHandler := handlers.NewDashboardHandler(k8sClient)
		mux.HandleFunc("/", dashboardHandler.ServeDashboard)
		mux.HandleFunc("/runs", dashboardHandler.ServeRuns)
		mux.HandleFunc("/logs/{namespace}/{name}/{step}", dashboardHandler.ServeLogs)
	}

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware
	var handler http.Handler = mux
	if enableCORS {
		handler = middleware.CORS(handler)
	}
	handler = middleware.Logging(handler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("API server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "API server failed")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Info("Shutting down API server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "Server shutdown failed")
		os.Exit(1)
	}

	logger.Info("API server stopped")
}

// getKubeConfig creates a Kubernetes client config
func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		// Use kubeconfig file
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	// Use in-cluster config
	return rest.InClusterConfig()
}
