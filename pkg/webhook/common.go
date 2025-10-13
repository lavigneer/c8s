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

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// WebhookEvent represents a normalized webhook event from any provider
type WebhookEvent struct {
	Repository    string
	RepositoryURL string
	Commit        string
	Branch        string
	Author        string
	AuthorEmail   string
	CommitMessage string
	Timestamp     metav1.Time
}

// Handler is the common interface for webhook handlers
type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// createPipelineRun creates a PipelineRun CRD from a webhook event
func createPipelineRun(
	ctx context.Context,
	k8sClient client.Client,
	event *WebhookEvent,
	repoConn *c8sv1alpha1.RepositoryConnection,
) error {
	logger := log.FromContext(ctx)

	// Generate PipelineRun name
	runName := fmt.Sprintf("%s-%s", repoConn.Name, event.Commit[:8])

	// Create PipelineRun
	pipelineRun := &c8sv1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runName,
			Namespace: repoConn.Namespace,
			Labels: map[string]string{
				"c8s.dev/pipeline-config": repoConn.Spec.PipelineConfigRef,
				"c8s.dev/repository":      repoConn.Name,
				"c8s.dev/branch":          event.Branch,
				"c8s.dev/commit":          event.Commit[:8],
			},
			Annotations: map[string]string{
				"c8s.dev/repository-url": event.RepositoryURL,
				"c8s.dev/author":         event.Author,
				"c8s.dev/author-email":   event.AuthorEmail,
			},
		},
		Spec: c8sv1alpha1.PipelineRunSpec{
			PipelineConfigRef: repoConn.Spec.PipelineConfigRef,
			Commit:            event.Commit,
			Branch:            event.Branch,
			TriggeredBy:       event.Author,
			TriggeredAt:       &event.Timestamp,
			CommitMessage:     event.CommitMessage,
			Author:            event.Author,
		},
	}

	// Check if PipelineRun already exists (idempotent)
	existing := &c8sv1alpha1.PipelineRun{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      runName,
		Namespace: repoConn.Namespace,
	}, existing)

	if err == nil {
		logger.Info("PipelineRun already exists, skipping creation",
			"name", runName,
			"namespace", repoConn.Namespace,
		)
		return nil
	}

	// Create new PipelineRun
	if err := k8sClient.Create(ctx, pipelineRun); err != nil {
		return fmt.Errorf("failed to create PipelineRun: %w", err)
	}

	logger.Info("Created PipelineRun",
		"name", runName,
		"namespace", repoConn.Namespace,
		"commit", event.Commit[:8],
		"branch", event.Branch,
	)

	return nil
}

// findRepositoryConnection finds a RepositoryConnection by repository URL
func findRepositoryConnection(
	ctx context.Context,
	k8sClient client.Client,
	repositoryURL string,
	namespace string,
) (*c8sv1alpha1.RepositoryConnection, error) {
	// List all RepositoryConnections in namespace
	repoConnList := &c8sv1alpha1.RepositoryConnectionList{}
	if err := k8sClient.List(ctx, repoConnList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list RepositoryConnections: %w", err)
	}

	// Find matching repository
	for i := range repoConnList.Items {
		conn := &repoConnList.Items[i]
		if conn.Spec.Repository == repositoryURL {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("no RepositoryConnection found for repository: %s", repositoryURL)
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]string{
		"error": message,
	})
}

// writeSuccessResponse writes a success response
func writeSuccessResponse(w http.ResponseWriter, message string) {
	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": message,
	})
}
