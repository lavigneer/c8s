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

package handlers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/storage"
)

// LogsHandler handles log streaming API requests
type LogsHandler struct {
	clientset kubernetes.Interface
	client    client.Client
	storage   storage.StorageClient
}

// NewLogsHandler creates a new LogsHandler
func NewLogsHandler(clientset kubernetes.Interface, client client.Client, storage storage.StorageClient) *LogsHandler {
	return &LogsHandler{
		clientset: clientset,
		client:    client,
		storage:   storage,
	}
}

// HandleStepLogs handles log retrieval and streaming for a pipeline step
// GET /api/v1/namespaces/{ns}/pipelineruns/{name}/logs/{step}?follow=true
func (h *LogsHandler) HandleStepLogs(w http.ResponseWriter, r *http.Request) {
	namespace := extractNamespace(r)
	pipelineRunName := extractResourceName(r)
	stepName := extractStepName(r)

	if namespace == "" || pipelineRunName == "" || stepName == "" {
		http.Error(w, "namespace, pipelinerun name, and step name are required", http.StatusBadRequest)
		return
	}

	// Get PipelineRun to find the step status
	var run v1alpha1.PipelineRun
	key := client.ObjectKey{Namespace: namespace, Name: pipelineRunName}
	if err := h.client.Get(r.Context(), key, &run); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "pipeline run not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get pipeline run: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the step status
	var stepStatus *v1alpha1.StepStatus
	for i, step := range run.Status.Steps {
		if step.Name == stepName {
			stepStatus = &run.Status.Steps[i]
			break
		}
	}

	if stepStatus == nil {
		http.Error(w, fmt.Sprintf("step %s not found in pipeline run", stepName), http.StatusNotFound)
		return
	}

	// Check if we should follow logs (live streaming)
	follow := r.URL.Query().Get("follow") == "true"

	if follow && stepStatus.Phase != "Succeeded" && stepStatus.Phase != "Failed" {
		// Stream logs from running Pod
		h.streamLogsFromPod(w, r, namespace, stepStatus.JobName, stepName)
	} else {
		// Fetch completed logs from storage
		h.fetchLogsFromStorage(w, r, stepStatus.LogURL)
	}
}

func (h *LogsHandler) streamLogsFromPod(w http.ResponseWriter, r *http.Request, namespace, jobName, stepName string) {
	// Find the Pod created by the Job
	pods, err := h.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list pods: %v", err), http.StatusInternalServerError)
		return
	}

	if len(pods.Items) == 0 {
		http.Error(w, "no pods found for job", http.StatusNotFound)
		return
	}

	pod := pods.Items[0]

	// Stream logs from the Pod's main container
	req := h.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow:    true,
		Container: stepName, // Container name matches step name
	})

	stream, err := req.Stream(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to stream logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Set headers for streaming
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// Flush immediately to start streaming
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Copy logs to response
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			w.Write(line)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err != nil {
			if err != io.EOF {
				http.Error(w, fmt.Sprintf("error reading logs: %v", err), http.StatusInternalServerError)
			}
			break
		}
	}
}

func (h *LogsHandler) fetchLogsFromStorage(w http.ResponseWriter, r *http.Request, logURL string) {
	if logURL == "" {
		http.Error(w, "logs not yet available", http.StatusNotFound)
		return
	}

	// Extract key from logURL (assumes format: s3://bucket/key or just the key)
	key := logURL
	if len(logURL) > 5 && logURL[:5] == "s3://" {
		// Parse s3:// URL to extract key
		parts := logURL[5:] // Remove s3://
		// Find first slash to separate bucket from key
		slashIdx := 0
		for i, c := range parts {
			if c == '/' {
				slashIdx = i
				break
			}
		}
		if slashIdx > 0 && slashIdx < len(parts)-1 {
			key = parts[slashIdx+1:] // Everything after bucket name
		}
	}

	logsReader, err := h.storage.DownloadLog(r.Context(), key)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to download logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer logsReader.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Stream logs to response
	if _, err := io.Copy(w, logsReader); err != nil {
		// Can't change status code here, already sent headers
		// Log error but continue
		fmt.Fprintf(w, "\n\nError streaming logs: %v\n", err)
	}
}
