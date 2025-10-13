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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// GitHubHandler handles GitHub webhook events
type GitHubHandler struct {
	client client.Client
}

// NewGitHubHandler creates a new GitHub webhook handler
func NewGitHubHandler(c client.Client) *GitHubHandler {
	return &GitHubHandler{client: c}
}

// GitHubPushEvent represents a GitHub push webhook event
type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		SSHURL   string `json:"ssh_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		Author    struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
	} `json:"head_commit"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
}

// Handle processes GitHub webhook requests
func (h *GitHubHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithValues("provider", "github")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// Check GitHub event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
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
	var pushEvent GitHubPushEvent
	if err := json.Unmarshal(body, &pushEvent); err != nil {
		logger.Error(err, "Failed to parse push event JSON")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Extract branch name from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(pushEvent.Ref, "refs/heads/")

	logger.Info("Received GitHub push event",
		"repository", pushEvent.Repository.FullName,
		"branch", branch,
		"commit", pushEvent.After[:8],
	)

	// Find RepositoryConnection for this repository
	// Note: Using default namespace for now. In production, this would be configurable
	namespace := "default"
	repoConn, err := findRepositoryConnection(ctx, h.client, pushEvent.Repository.CloneURL, namespace)
	if err != nil {
		// Try SSH URL as well
		repoConn, err = findRepositoryConnection(ctx, h.client, pushEvent.Repository.SSHURL, namespace)
		if err != nil {
			logger.Info("No RepositoryConnection found for repository",
				"repository", pushEvent.Repository.FullName,
				"cloneURL", pushEvent.Repository.CloneURL,
				"sshURL", pushEvent.Repository.SSHURL,
			)
			writeErrorResponse(w, http.StatusNotFound,
				fmt.Sprintf("No configuration found for repository: %s", pushEvent.Repository.FullName))
			return
		}
	}

	// Verify webhook secret signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature != "" {
		if err := h.verifySignature(ctx, signature, body, repoConn); err != nil {
			logger.Error(err, "Webhook signature verification failed")
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid webhook signature")
			return
		}
		logger.Info("Webhook signature verified successfully")
	}

	// Parse timestamp
	timestamp, err := parseTimestamp(pushEvent.HeadCommit.Timestamp)
	if err != nil {
		logger.Error(err, "Failed to parse timestamp")
		timestamp = metav1.Now()
	}

	// Create normalized webhook event
	event := &WebhookEvent{
		Repository:    pushEvent.Repository.FullName,
		RepositoryURL: pushEvent.Repository.CloneURL,
		Commit:        pushEvent.After,
		Branch:        branch,
		Author:        pushEvent.HeadCommit.Author.Name,
		AuthorEmail:   pushEvent.HeadCommit.Author.Email,
		CommitMessage: pushEvent.HeadCommit.Message,
		Timestamp:     timestamp,
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

// verifySignature verifies the GitHub webhook HMAC signature
func (h *GitHubHandler) verifySignature(
	ctx context.Context,
	signature string,
	payload []byte,
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

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, webhookSecret)
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// GitHub signature format: sha256=<hex>
	signatureParts := strings.SplitN(signature, "=", 2)
	if len(signatureParts) != 2 || signatureParts[0] != "sha256" {
		return fmt.Errorf("invalid signature format")
	}

	receivedMAC := signatureParts[1]

	// Compare MACs
	if !hmac.Equal([]byte(expectedMAC), []byte(receivedMAC)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// parseTimestamp parses ISO 8601 timestamp from GitHub
func parseTimestamp(timestamp string) (metav1.Time, error) {
	// GitHub uses RFC3339 format
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return metav1.Time{}, err
	}
	return metav1.Time{Time: t}, nil
}
