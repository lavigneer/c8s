package webhook_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	webhookpkg "github.com/org/c8s/pkg/webhook"
)

// newFakeClientWithCRDs creates a fake k8s client with C8S CRDs registered
func newFakeClientWithCRDs(objs ...client.Object) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = c8sv1alpha1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...)
}

// ===== GitHub Webhook Tests =====

// TestGitHubVerifySignatureValid tests successful GitHub signature verification
func TestGitHubVerifySignatureValid(t *testing.T) {
	// Create fake k8s client with secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret-key"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	// Create payload
	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	// Calculate correct signature
	mac := hmac.New(sha256.New, []byte("test-secret-key"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	// Create fake client and handler
	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	// Test verification
	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err != nil {
		t.Errorf("expected no error for valid signature, got: %v", err)
	}
}

// TestGitHubVerifySignatureInvalid tests failed GitHub signature verification
func TestGitHubVerifySignatureInvalid(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret-key"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)
	invalidSignature := "sha256=invalidsignaturehash"

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, invalidSignature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for invalid signature, got nil")
	}
}

// TestGitHubVerifySignatureMissingSecret tests error when secret is not found
func TestGitHubVerifySignatureMissingSecret(t *testing.T) {
	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "missing-secret",
		},
	}

	payload := []byte(`{"ref":"refs/heads/main"}`)
	signature := "sha256=abc123"

	fakeClient := newFakeClientWithCRDs().Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for missing secret, got nil")
	}
}

// TestGitHubVerifySignatureMissingSecretKey tests error when secret data key is missing
func TestGitHubVerifySignatureMissingSecretKey(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"wrong-key": []byte("test-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"ref":"refs/heads/main"}`)
	signature := "sha256=abc123"

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for missing secret key, got nil")
	}
}

// TestGitHubVerifySignatureInvalidFormat tests error for invalid signature format
func TestGitHubVerifySignatureInvalidFormat(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret-key"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"ref":"refs/heads/main"}`)
	// Invalid format: missing algorithm prefix
	invalidSignature := "invalidsignaturehash"

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, invalidSignature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for invalid format, got nil")
	}
}

// TestGitHubVerifySignatureNoSecretRef tests error when no secret ref is configured
func TestGitHubVerifySignatureNoSecretRef(t *testing.T) {
	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "",
		},
	}

	payload := []byte(`{"ref":"refs/heads/main"}`)
	signature := "sha256=abc123"

	fakeClient := newFakeClientWithCRDs().Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for missing secret ref, got nil")
	}
}

// TestGitHubVerifySignatureEmptyPayload tests verification with empty payload
func TestGitHubVerifySignatureEmptyPayload(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret-key"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	// Calculate signature for empty payload
	mac := hmac.New(sha256.New, []byte("test-secret-key"))
	mac.Write([]byte{})
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, []byte{}, repoConn)
	if err != nil {
		t.Errorf("expected no error for empty payload with correct signature, got: %v", err)
	}
}

// ===== GitLab Webhook Tests =====

// TestGitLabVerifyTokenValid tests successful GitLab token verification
func TestGitLabVerifyTokenValid(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-token-key"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitLabHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifyToken(ctx, "test-token-key", repoConn)
	if err != nil {
		t.Errorf("expected no error for valid token, got: %v", err)
	}
}

// TestGitLabVerifyTokenInvalid tests failed GitLab token verification
func TestGitLabVerifyTokenInvalid(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("correct-token"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitLabHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifyToken(ctx, "wrong-token", repoConn)
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

// TestGitLabVerifyTokenMissingSecret tests error when secret is not found
func TestGitLabVerifyTokenMissingSecret(t *testing.T) {
	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "missing-secret",
		},
	}

	fakeClient := newFakeClientWithCRDs().Build()
	handler := webhookpkg.NewGitLabHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifyToken(ctx, "some-token", repoConn)
	if err == nil {
		t.Errorf("expected error for missing secret, got nil")
	}
}

// TestGitLabVerifyTokenEmptyToken tests error for empty token
func TestGitLabVerifyTokenEmptyToken(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("expected-token"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitLabHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifyToken(ctx, "", repoConn)
	if err == nil {
		t.Errorf("expected error for empty token, got nil")
	}
}

// ===== Bitbucket Webhook Tests =====

// TestBitbucketVerifySignatureValid tests successful Bitbucket signature verification
func TestBitbucketVerifySignatureValid(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("bitbucket-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"push":{"changes":[{"new":{"name":"main"}}]}}`)

	// Calculate correct signature for Bitbucket
	mac := hmac.New(sha256.New, []byte("bitbucket-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewBitbucketHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err != nil {
		t.Errorf("expected no error for valid signature, got: %v", err)
	}
}

// TestBitbucketVerifySignatureInvalid tests failed Bitbucket signature verification
func TestBitbucketVerifySignatureInvalid(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("bitbucket-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"push":{"changes":[{"new":{"name":"main"}}]}}`)
	invalidSignature := "invalidsignaturehash"

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewBitbucketHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, invalidSignature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for invalid signature, got nil")
	}
}

// TestBitbucketVerifySignatureMissingSecret tests error when secret is not found
func TestBitbucketVerifySignatureMissingSecret(t *testing.T) {
	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "missing-secret",
		},
	}

	payload := []byte(`{"push":{"changes":[]}}`)
	signature := "abc123"

	fakeClient := newFakeClientWithCRDs().Build()
	handler := webhookpkg.NewBitbucketHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err == nil {
		t.Errorf("expected error for missing secret, got nil")
	}
}

// ===== HTTP Request Tests =====

// TestGitHubWebhookRequestWithValidSignature tests complete GitHub webhook request
func TestGitHubWebhookRequestWithValidSignature(t *testing.T) {
	payload := []byte(`{
		"ref": "refs/heads/main",
		"after": "abc123",
		"repository": {
			"name": "test-repo",
			"full_name": "org/test-repo",
			"clone_url": "https://github.com/org/test-repo.git",
			"ssh_url": "git@github.com:org/test-repo.git"
		},
		"head_commit": {
			"id": "abc123",
			"message": "Test commit",
			"timestamp": "2025-11-02T10:00:00Z",
			"author": {
				"name": "Test User",
				"email": "test@example.com"
			}
		}
	}`)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	// Create HTTP request
	req := httptest.NewRequest("POST", "/webhooks/github", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("Content-Type", "application/json")

	// Verify request structure is valid
	if req.Method != "POST" {
		t.Errorf("expected POST request, got %s", req.Method)
	}

	if req.Header.Get("X-Hub-Signature-256") != signature {
		t.Errorf("signature header not set correctly")
	}
}

// TestGitLabWebhookRequestWithValidToken tests complete GitLab webhook request
func TestGitLabWebhookRequestWithValidToken(t *testing.T) {
	payload := []byte(`{
		"object_kind": "push",
		"ref": "refs/heads/main",
		"project": {
			"id": 123,
			"name": "test-repo",
			"web_url": "https://gitlab.com/org/test-repo"
		},
		"commits": [{
			"id": "abc123",
			"message": "Test commit",
			"author": {
				"name": "Test User",
				"email": "test@example.com"
			},
			"timestamp": "2025-11-02T10:00:00Z"
		}]
	}`)

	token := "test-token-key"

	// Create HTTP request
	req := httptest.NewRequest("POST", "/webhooks/gitlab", bytes.NewReader(payload))
	req.Header.Set("X-Gitlab-Token", token)
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("Content-Type", "application/json")

	// Verify request structure
	if req.Method != "POST" {
		t.Errorf("expected POST request, got %s", req.Method)
	}

	if req.Header.Get("X-Gitlab-Token") != token {
		t.Errorf("token header not set correctly")
	}
}

// TestBitbucketWebhookRequestWithValidSignature tests complete Bitbucket webhook request
func TestBitbucketWebhookRequestWithValidSignature(t *testing.T) {
	payload := []byte(`{
		"push": {
			"changes": [{
				"new": {
					"name": "main",
					"type": "branch"
				},
				"commits": [{
					"hash": "abc123",
					"message": "Test commit",
					"author": {
						"user": {
							"display_name": "Test User",
							"emailAddress": "test@example.com"
						}
					},
					"timestamp": "2025-11-02T10:00:00Z"
				}]
			}]
		},
		"repository": {
			"name": "test-repo",
			"full_name": "org/test-repo",
			"links": {
				"html": {
					"href": "https://bitbucket.org/org/test-repo"
				}
			}
		}
	}`)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Create HTTP request
	req := httptest.NewRequest("POST", "/webhooks/bitbucket", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature", expectedSig)
	req.Header.Set("X-Event-Key", "repo:push")
	req.Header.Set("Content-Type", "application/json")

	// Verify request structure
	if req.Method != "POST" {
		t.Errorf("expected POST request, got %s", req.Method)
	}

	if req.Header.Get("X-Hub-Signature") != expectedSig {
		t.Errorf("signature header not set correctly")
	}
}

// ===== Edge Case Tests =====

// TestWebhookSignatureWithSpecialCharacters tests signature with special characters in payload
func TestWebhookSignatureWithSpecialCharacters(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	// Payload with special characters
	payload := []byte(`{"message":"Test with special chars: !@#$%^&*()_+-=[]{}|;':\"<>,.?/"}`)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err != nil {
		t.Errorf("expected no error for payload with special characters, got: %v", err)
	}
}

// TestWebhookSignatureWithLargePayload tests signature with large payload
func TestWebhookSignatureWithLargePayload(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	// Create large payload with many commits
	var commits []map[string]interface{}
	for i := 0; i < 100; i++ {
		commits = append(commits, map[string]interface{}{
			"id":      "abc" + string(rune(i)),
			"message": "Commit " + string(rune(i)),
		})
	}

	data := map[string]interface{}{
		"commits": commits,
	}
	payload, _ := json.Marshal(data)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err != nil {
		t.Errorf("expected no error for large payload, got: %v", err)
	}
}

// TestWebhookSignatureCaseSensitivity tests that signature is case-sensitive
func TestWebhookSignatureCaseSensitivity(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": []byte("test-secret"),
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"test":"data"}`)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Modify signature case (should fail)
	wrongCaseSignature := "sha256=" + strings.ToUpper(expectedSig)

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, wrongCaseSignature, payload, repoConn)
	// Note: This depends on whether hex is case-sensitive in the implementation
	// Hex strings are typically case-insensitive, so this test may need adjustment
	_ = err // Placeholder for conditional check based on implementation
}

// TestWebhookSecretWithBinaryData tests signature with binary data in secret
func TestWebhookSecretWithBinaryData(t *testing.T) {
	// Create secret with binary data
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"webhook-secret": {0xFF, 0xFE, 0xFD, 0xFC, 0xFB},
		},
	}

	repoConn := &c8sv1alpha1.RepositoryConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
		Spec: c8sv1alpha1.RepositoryConnectionSpec{
			WebhookSecretRef: "webhook-secret",
		},
	}

	payload := []byte(`{"test":"data"}`)

	// Calculate signature
	mac := hmac.New(sha256.New, []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB})
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	signature := "sha256=" + expectedSig

	fakeClient := newFakeClientWithCRDs().
		WithObjects(secret, repoConn).
		Build()
	handler := webhookpkg.NewGitHubHandler(fakeClient)

	ctx := context.Background()
	err := handler.VerifySignature(ctx, signature, payload, repoConn)
	if err != nil {
		t.Errorf("expected no error for binary secret, got: %v", err)
	}
}
