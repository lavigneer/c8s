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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// BitbucketHandler handles Bitbucket webhook events
type BitbucketHandler struct {
	client client.Client
}

// NewBitbucketHandler creates a new Bitbucket webhook handler
func NewBitbucketHandler(c client.Client) *BitbucketHandler {
	return &BitbucketHandler{client: c}
}

// BitbucketPushEvent represents a Bitbucket push webhook event
type BitbucketPushEvent struct {
	Push struct {
		Changes []struct {
			New struct {
				Type   string `json:"type"`
				Name   string `json:"name"`
				Target struct {
					Hash    string `json:"hash"`
					Message string `json:"message"`
					Date    string `json:"date"`
					Author  struct {
						User struct {
							DisplayName string `json:"display_name"`
							Email       string `json:"email_address"`
						} `json:"user"`
					} `json:"author"`
				} `json:"target"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Links    struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Clone []struct {
				Name string `json:"name"`
				Href string `json:"href"`
			} `json:"clone"`
		} `json:"links"`
	} `json:"repository"`
	Actor struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email_address"`
	} `json:"actor"`
}

// Handle processes Bitbucket webhook requests
func (h *BitbucketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithValues("provider", "bitbucket")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// Check Bitbucket event type
	eventType := r.Header.Get("X-Event-Key")
	if eventType != "repo:push" {
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
	var pushEvent BitbucketPushEvent
	if err := json.Unmarshal(body, &pushEvent); err != nil {
		logger.Error(err, "Failed to parse push event JSON")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Bitbucket can have multiple changes in one push
	if len(pushEvent.Push.Changes) == 0 {
		logger.Info("No changes in push event")
		writeSuccessResponse(w, "No changes to process")
		return
	}

	// Process the first change (most common case)
	change := pushEvent.Push.Changes[0]
	branch := change.New.Name
	commit := change.New.Target.Hash

	logger.Info("Received Bitbucket push event",
		"repository", pushEvent.Repository.FullName,
		"branch", branch,
		"commit", commit[:8],
	)

	// Get clone URL (prefer HTTPS)
	cloneURL := ""
	for _, link := range pushEvent.Repository.Links.Clone {
		if link.Name == "https" {
			cloneURL = link.Href
			break
		}
	}
	if cloneURL == "" && len(pushEvent.Repository.Links.Clone) > 0 {
		cloneURL = pushEvent.Repository.Links.Clone[0].Href
	}

	// Find RepositoryConnection for this repository
	namespace := "default"
	repoConn, err := findRepositoryConnection(ctx, h.client, cloneURL, namespace)
	if err != nil {
		logger.Info("No RepositoryConnection found for repository",
			"repository", pushEvent.Repository.FullName,
			"cloneURL", cloneURL,
		)
		writeErrorResponse(w, http.StatusNotFound,
			fmt.Sprintf("No configuration found for repository: %s", pushEvent.Repository.FullName))
		return
	}

	// Verify webhook signature
	signature := r.Header.Get("X-Hub-Signature")
	if signature != "" {
		if err := h.verifySignature(ctx, signature, body, repoConn); err != nil {
			logger.Error(err, "Webhook signature verification failed")
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid webhook signature")
			return
		}
		logger.Info("Webhook signature verified successfully")
	}

	// Parse timestamp
	timestamp := metav1.Now()
	if change.New.Target.Date != "" {
		if t, err := parseTimestamp(change.New.Target.Date); err == nil {
			timestamp = t
		}
	}

	// Create normalized webhook event
	event := &WebhookEvent{
		Repository:    pushEvent.Repository.FullName,
		RepositoryURL: cloneURL,
		Commit:        commit,
		Branch:        branch,
		Author:        change.New.Target.Author.User.DisplayName,
		AuthorEmail:   change.New.Target.Author.User.Email,
		CommitMessage: change.New.Target.Message,
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

// verifySignature verifies the Bitbucket webhook HMAC signature
func (h *BitbucketHandler) verifySignature(
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

	// Compare MACs (Bitbucket uses sha256=<hex> format)
	if signature != "sha256="+expectedMAC {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
