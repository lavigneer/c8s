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

// Package testutil provides utilities for integration testing with envtest
package testutil

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// TestEnvironment wraps envtest.Environment with helpers for testing
type TestEnvironment struct {
	testEnv *envtest.Environment
	Config  *rest.Config
	Client  client.Client
	ctx     context.Context
	cancel  context.CancelFunc
}

// SetupTestEnvironment creates and starts a test Kubernetes environment
// This should be called in TestMain or at the start of integration tests
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	// Get the root directory for CRD path
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join(root, "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false, // Will be false initially, set to true once CRDs exist
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err, "failed to start test environment")
	require.NotNil(t, cfg, "test environment config is nil")

	// Create client
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err, "failed to create client")

	return &TestEnvironment{
		testEnv: testEnv,
		Config:  cfg,
		Client:  k8sClient,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Stop tears down the test environment
func (te *TestEnvironment) Stop(t *testing.T) {
	t.Helper()

	te.cancel()
	err := te.testEnv.Stop()
	require.NoError(t, err, "failed to stop test environment")
}

// Context returns the test context
func (te *TestEnvironment) Context() context.Context {
	return te.ctx
}

// ContextWithTimeout returns a context with timeout for individual test operations
func (te *TestEnvironment) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(te.ctx, timeout)
}

// CreateNamespace creates a namespace for testing
func (te *TestEnvironment) CreateNamespace(t *testing.T, name string) {
	t.Helper()
	// Implementation will be added when we have the CRD types defined
	// For now, this is a placeholder
}

// CleanupNamespace deletes a test namespace
func (te *TestEnvironment) CleanupNamespace(t *testing.T, name string) {
	t.Helper()
	// Implementation will be added when we have the CRD types defined
	// For now, this is a placeholder
}
