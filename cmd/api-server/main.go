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
	"path/filepath"
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

	"github.com/org/c8s/pkg/api/handlers"
	"github.com/org/c8s/pkg/api/middleware"
	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/storage"
	"github.com/org/c8s/pkg/storage/s3"
)

var (
	port            int
	kubeconfig      string
	enableDashboard bool
	enableCORS      bool
	s3Bucket        string
	s3Region        string
	s3Endpoint      string
)

func init() {
	flag.IntVar(&port, "port", 8080, "API server port")
	flag.BoolVar(&enableDashboard, "enable-dashboard", false, "Enable HTMX dashboard")
	flag.BoolVar(&enableCORS, "enable-cors", true, "Enable CORS middleware")
	flag.StringVar(&s3Bucket, "s3-bucket", "", "S3 bucket for logs (env: C8S_S3_BUCKET)")
	flag.StringVar(&s3Region, "s3-region", "us-west-2", "S3 region (env: C8S_S3_REGION)")
	flag.StringVar(&s3Endpoint, "s3-endpoint", "", "S3 endpoint for MinIO/compatible storage (env: C8S_S3_ENDPOINT)")
}

func main() {
	flag.Parse()

	// Setup logger
	logger := zap.New(zap.UseDevMode(true))
	ctrl.SetLogger(logger)
	ctx := log.IntoContext(context.Background(), logger)

	logger.Info("Starting C8S API Server",
		"port", port,
		"dashboard", enableDashboard,
		"cors", enableCORS,
	)

	// Create Kubernetes client config
	// Use kubeconfig from env or default location if not in-cluster
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}
	config, err := getKubeConfig(kubeconfigPath)
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

	// Initialize storage client
	var storageClient storage.StorageClient
	if s3Bucket != "" {
		// Use env vars if flags not provided
		if s3Bucket == "" {
			s3Bucket = os.Getenv("C8S_S3_BUCKET")
		}
		if s3Region == "" {
			s3Region = os.Getenv("C8S_S3_REGION")
		}
		if s3Endpoint == "" {
			s3Endpoint = os.Getenv("C8S_S3_ENDPOINT")
		}

		storageConfig := &storage.Config{
			Bucket:          s3Bucket,
			Region:          s3Region,
			Endpoint:        s3Endpoint,
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			UsePathStyle:    s3Endpoint != "", // Use path-style for custom endpoints
		}

		storageClient, err = s3.NewClient(storageConfig)
		if err != nil {
			logger.Error(err, "Failed to create S3 storage client")
			os.Exit(1)
		}
		logger.Info("S3 storage client initialized", "bucket", s3Bucket)
	} else {
		logger.Info("No S3 bucket configured, log streaming from storage will be disabled")
	}

	// Create HTTP router
	mux := http.NewServeMux()

	// Initialize handlers
	pipelineConfigHandler := handlers.NewPipelineConfigHandler(k8sClient)
	pipelineRunHandler := handlers.NewPipelineRunHandler(k8sClient)
	logsHandler := handlers.NewLogsHandler(clientset, k8sClient, storageClient)

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
		dashboardHandler, err := handlers.NewDashboardHandler(k8sClient, "web/templates")
		if err != nil {
			logger.Error(err, "Failed to initialize dashboard handler")
			os.Exit(1)
		}
		mux.HandleFunc("/dashboard", dashboardHandler.ServeDashboard)
		mux.HandleFunc("/dashboard/runs", dashboardHandler.ServeRuns)
		mux.HandleFunc("/dashboard/logs", dashboardHandler.ServeLogs)
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
