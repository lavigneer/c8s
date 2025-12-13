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

package logstorage

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"

	"sigs.k8s.io/controller-runtime/pkg/client"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/storage"
	"k8s.io/client-go/kubernetes"
)

// HybridLogStorage provides smart fallback: tries S3 first, falls back to K8s pods
// This allows seamless experience for:
// - Running jobs: Fetches live logs from pods (LogURL not yet set)
// - Completed jobs: Fetches persisted logs from S3 (faster, more reliable)
// - Mixed scenarios: Uses S3 when available, pods when not
type HybridLogStorage struct {
	k8sClient       client.Client
	clientset       kubernetes.Interface
	storageClient   storage.Client
	namespace       string
	runID           string
	s3Storage       *S3LogStorage
	k8sStorage      *KubernetesLogStorage
}

// NewHybridLogStorage creates a new hybrid log storage with fallback
func NewHybridLogStorage(
	k8sClient client.Client,
	clientset kubernetes.Interface,
	storageClient storage.Client,
	namespace, runID string,
) *HybridLogStorage {
	return &HybridLogStorage{
		k8sClient:      k8sClient,
		clientset:      clientset,
		storageClient:  storageClient,
		namespace:      namespace,
		runID:          runID,
		s3Storage:      NewS3LogStorage(storageClient, namespace),
		k8sStorage:     NewKubernetesLogStorage(k8sClient, clientset, namespace, runID),
	}
}

// GetStepLogs gets logs using smart fallback: S3 (if available) then K8s pods
func (h *HybridLogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	// Fetch the PipelineRun to check if LogURL is populated
	pipelineRun := &c8sv1alpha1.PipelineRun{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{Namespace: h.namespace, Name: runID}, pipelineRun); err != nil {
		log.Printf("Failed to fetch PipelineRun for hybrid storage: %v", err)
		// Fall back to K8s pods if we can't get PipelineRun
		return h.k8sStorage.GetStepLogs(ctx, runID, stepID)
	}

	// Find the step status
	var stepStatus *c8sv1alpha1.StepStatus
	for i := range pipelineRun.Status.Steps {
		if pipelineRun.Status.Steps[i].Name == stepID {
			stepStatus = &pipelineRun.Status.Steps[i]
			break
		}
	}

	// If LogURL is populated, try S3 first
	if stepStatus != nil && stepStatus.LogURL != "" {
		logStream, err := h.s3Storage.GetStepLogs(ctx, runID, stepID)
		if err == nil {
			log.Printf("Retrieved logs from S3 for step %s", stepID)
			return logStream, nil
		}
		// If S3 fails, log the error but continue to fallback
		log.Printf("Failed to get logs from S3, falling back to K8s pods: %v", err)
	}

	// Fall back to Kubernetes pods (for running jobs or if S3 unavailable)
	logStream, err := h.k8sStorage.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs from both S3 and Kubernetes: %w", err)
	}

	return logStream, nil
}

// StreamStepLogs streams logs using hybrid fallback
func (h *HybridLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	logsReader, err := h.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return err
	}
	defer logsReader.Close()

	// Stream logs line by line
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case linesChan <- scanner.Text():
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}

// GetLogSnapshot returns the last N lines using hybrid fallback
func (h *HybridLogStorage) GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error) {
	logsReader, err := h.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return nil, err
	}
	defer logsReader.Close()

	// Reuse snapshot logic from S3Storage
	return h.s3Storage.GetLogSnapshot(ctx, runID, stepID, lines)
}

// GetLogSize returns the log size using hybrid fallback
func (h *HybridLogStorage) GetLogSize(ctx context.Context, runID, stepID string) (int64, error) {
	logsReader, err := h.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return 0, err
	}
	defer logsReader.Close()

	// Reuse size logic from S3Storage
	return h.s3Storage.GetLogSize(ctx, runID, stepID)
}
