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

package cli

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SecretCommand handles secret-related CLI operations
type SecretCommand struct {
	client    kubernetes.Interface
	namespace string
}

// NewSecretCommand creates a new SecretCommand
func NewSecretCommand(client kubernetes.Interface, namespace string) *SecretCommand {
	return &SecretCommand{
		client:    client,
		namespace: namespace,
	}
}

// Create creates a new secret from literal key-value pairs
// Example: CreateFromLiteral("my-secret", map[string]string{"API_KEY": "secret123", "PASSWORD": "pass456"})
func (sc *SecretCommand) CreateFromLiteral(ctx context.Context, name string, data map[string]string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	if len(data) == 0 {
		return fmt.Errorf("secret data cannot be empty")
	}

	// Convert string data to byte data
	secretData := make(map[string][]byte, len(data))
	for key, value := range data {
		secretData[key] = []byte(value)
	}

	// Create the secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sc.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "c8s",
			},
			Annotations: map[string]string{
				"c8s.dev/description": "Secret managed by C8S CLI",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	// Create the secret
	_, err := sc.client.CoreV1().Secrets(sc.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return fmt.Errorf("secret '%s' already exists in namespace '%s'. Use 'c8s secret update' to update it", name, sc.namespace)
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// Update updates an existing secret with new key-value pairs
func (sc *SecretCommand) Update(ctx context.Context, name string, data map[string]string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	// Get the existing secret
	secret, err := sc.client.CoreV1().Secrets(sc.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("secret '%s' not found in namespace '%s'. Use 'c8s secret create' to create it", name, sc.namespace)
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Update the secret data
	for key, value := range data {
		secret.Data[key] = []byte(value)
	}

	// Update the secret
	_, err = sc.client.CoreV1().Secrets(sc.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// Delete deletes a secret
func (sc *SecretCommand) Delete(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	err := sc.client.CoreV1().Secrets(sc.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("secret '%s' not found in namespace '%s'", name, sc.namespace)
		}
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// Get retrieves a secret and returns it (without revealing sensitive data)
func (sc *SecretCommand) Get(ctx context.Context, name string) (*SecretInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("secret name cannot be empty")
	}

	secret, err := sc.client.CoreV1().Secrets(sc.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("secret '%s' not found in namespace '%s'", name, sc.namespace)
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Extract keys without revealing values
	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}

	return &SecretInfo{
		Name:      secret.Name,
		Namespace: secret.Namespace,
		Keys:      keys,
		CreatedAt: secret.CreationTimestamp.Time,
	}, nil
}

// List lists all secrets in the namespace
func (sc *SecretCommand) List(ctx context.Context) ([]SecretInfo, error) {
	secrets, err := sc.client.CoreV1().Secrets(sc.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/managed-by=c8s",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	result := make([]SecretInfo, 0, len(secrets.Items))
	for _, secret := range secrets.Items {
		keys := make([]string, 0, len(secret.Data))
		for key := range secret.Data {
			keys = append(keys, key)
		}

		result = append(result, SecretInfo{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Keys:      keys,
			CreatedAt: secret.CreationTimestamp.Time,
		})
	}

	return result, nil
}

// ParseLiteralArgs parses command-line arguments in the format KEY=VALUE
// Example: ["API_KEY=secret123", "PASSWORD=pass456"]
func ParseLiteralArgs(args []string) (map[string]string, error) {
	result := make(map[string]string, len(args))

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid literal format '%s', expected KEY=VALUE", arg)
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1] // Don't trim value to preserve spaces

		if key == "" {
			return nil, fmt.Errorf("key cannot be empty in '%s'", arg)
		}

		result[key] = value
	}

	return result, nil
}

// ValidateSecretName validates a Kubernetes secret name
func ValidateSecretName(name string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	if len(name) > 253 {
		return fmt.Errorf("secret name cannot be longer than 253 characters")
	}

	// Basic validation for Kubernetes names
	// Full validation would use k8s.io/apimachinery/pkg/util/validation
	if strings.Contains(name, " ") {
		return fmt.Errorf("secret name cannot contain spaces")
	}

	return nil
}

// PrintUsageExample prints helpful usage examples
func PrintUsageExample() string {
	return `
Examples:
  # Create a secret with one key-value pair
  c8s secret create my-secret --from-literal=API_KEY=secret123

  # Create a secret with multiple key-value pairs
  c8s secret create my-secret --from-literal=API_KEY=secret123 --from-literal=PASSWORD=pass456

  # Update an existing secret
  c8s secret update my-secret --from-literal=API_KEY=newsecret

  # List all C8S-managed secrets
  c8s secret list

  # Get details of a specific secret (without revealing values)
  c8s secret get my-secret

  # Delete a secret
  c8s secret delete my-secret

  # Use the secret in a pipeline step (in .c8s.yaml):
  steps:
    - name: deploy
      image: alpine:latest
      commands:
        - echo "Deploying with API key"
      secrets:
        - secretName: my-secret
          key: API_KEY
          envVar: API_KEY

Note: Secret values are automatically masked in logs to prevent exposure.
`
}
