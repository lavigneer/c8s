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
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// GitLabHandler handles GitLab webhook events
type GitLabHandler struct {
	client client.Client
}

// NewGitLabHandler creates a new GitLab webhook handler
func NewGitLabHandler(c client.Client) *GitLabHandler {
	return &GitLabHandler{client: c}
}

// GitLabPushEvent represents a GitLab push webhook event
type GitLabPushEvent struct {
	ObjectKind string `json:"object_kind"`
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Project    struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		GitHTTPURL        string `json:"git_http_url"`
		GitSSHURL         string `json:"git_ssh_url"`
	} `json:"project"`
	Commits []struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// Handle processes GitLab webhook requests
func (h *GitLabHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithValues("provider", "gitlab")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// Check GitLab event type
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType != "Push Hook" {
		logger.Info("Ignoring non-push event", "eventType", eventType)
		writeSuccessResponse(w, fmt.Sprintf("Event type '%s' ignored", eventType))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(err, "Failed to read request body")
		writeErrorResponse(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Parse push event
	var pushEvent GitLabPushEvent
	if err := json.Unmarshal(body, &pushEvent); err != nil {
		logger.Error(err, "Failed to parse push event JSON")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Extract branch name from ref (refs/heads/main -> main)
	branch := pushEvent.Ref[len("refs/heads/"):]

	logger.Info("Received GitLab push event",
		"project", pushEvent.Project.PathWithNamespace,
		"branch", branch,
		"commit", pushEvent.After[:8],
	)

	// Find RepositoryConnection for this repository
	namespace := "default"
	repoConn, err := findRepositoryConnection(ctx, h.client, pushEvent.Project.GitHTTPURL, namespace)
	if err != nil {
		repoConn, err = findRepositoryConnection(ctx, h.client, pushEvent.Project.GitSSHURL, namespace)
		if err != nil {
			logger.Info("No RepositoryConnection found for project",
				"project", pushEvent.Project.PathWithNamespace,
			)
			writeErrorResponse(w, http.StatusNotFound,
				fmt.Sprintf("No configuration found for project: %s", pushEvent.Project.PathWithNamespace))
			return
		}
	}

	// Verify webhook token
	token := r.Header.Get("X-Gitlab-Token")
	if token != "" {
		if err := h.verifyToken(ctx, token, repoConn); err != nil {
			logger.Error(err, "Webhook token verification failed")
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid webhook token")
			return
		}
		logger.Info("Webhook token verified successfully")
	}

	// Get most recent commit
	commitMsg := ""
	commitAuthor := pushEvent.UserName
	commitEmail := pushEvent.UserEmail
	commitTimestamp := metav1.Now()

	if len(pushEvent.Commits) > 0 {
		lastCommit := pushEvent.Commits[len(pushEvent.Commits)-1]
		commitMsg = lastCommit.Message
		commitAuthor = lastCommit.Author.Name
		commitEmail = lastCommit.Author.Email
		if t, err := parseTimestamp(lastCommit.Timestamp); err == nil {
			commitTimestamp = t
		}
	}

	// Create normalized webhook event
	event := &WebhookEvent{
		Repository:    pushEvent.Project.PathWithNamespace,
		RepositoryURL: pushEvent.Project.GitHTTPURL,
		Commit:        pushEvent.After,
		Branch:        branch,
		Author:        commitAuthor,
		AuthorEmail:   commitEmail,
		CommitMessage: commitMsg,
		Timestamp:     commitTimestamp,
	}

	// Create PipelineRun
	if err := createPipelineRun(ctx, h.client, event, repoConn); err != nil {
		logger.Error(err, "Failed to create PipelineRun")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create pipeline run")
		return
	}

	// Return success
	writeSuccessResponse(w, "Pipeline run created successfully")
}

// verifyToken verifies the GitLab webhook token
func (h *GitLabHandler) verifyToken(
	ctx context.Context,
	token string,
	repoConn *c8sv1alpha1.RepositoryConnection,
) error {
	logger := log.FromContext(ctx)

	// Get webhook secret from Kubernetes Secret
	if repoConn.Spec.WebhookSecretRef == "" {
		return fmt.Errorf("no webhook secret configured for repository connection")
	}

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Name:      repoConn.Spec.WebhookSecretRef,
		Namespace: repoConn.Namespace,
	}

	if err := h.client.Get(ctx, secretKey, secret); err != nil {
		logger.Error(err, "Failed to get webhook secret", "secret", repoConn.Spec.WebhookSecretRef)
		return fmt.Errorf("failed to get webhook secret: %w", err)
	}

	// Get secret value (default key is "webhook-secret")
	webhookSecret, ok := secret.Data["webhook-secret"]
	if !ok {
		return fmt.Errorf("webhook secret key 'webhook-secret' not found")
	}

	// Compare tokens
	if token != string(webhookSecret) {
		return fmt.Errorf("token mismatch")
	}

	return nil
}
