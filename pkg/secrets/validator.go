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

package secrets

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/org/c8s/pkg/apis/v1alpha1"
)

// Validator validates secret references in pipeline configurations
type Validator struct {
	client kubernetes.Interface
}

// NewValidator creates a new secret validator
func NewValidator(client kubernetes.Interface) *Validator {
	return &Validator{
		client: client,
	}
}

// ValidateSecretReferences validates that all secret references in a pipeline step exist and are accessible
func (v *Validator) ValidateSecretReferences(ctx context.Context, step *v1alpha1.PipelineStep, namespace string) error {
	var validationErrors []string

	for _, secretRef := range step.Secrets {
		// Validate secret name is not empty
		if secretRef.SecretRef == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("step '%s': secret name cannot be empty", step.Name))
			continue
		}

		// Validate key is not empty
		if secretRef.Key == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("step '%s': secret key cannot be empty for secret '%s'", step.Name, secretRef.SecretRef))
			continue
		}

		// Note: EnvVar is optional - if empty, it defaults to the key name

		// Check if secret exists
		secret, err := v.client.CoreV1().Secrets(namespace).Get(ctx, secretRef.SecretRef, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				validationErrors = append(validationErrors, fmt.Sprintf("step '%s': secret '%s' not found in namespace '%s'", step.Name, secretRef.SecretRef, namespace))
			} else {
				validationErrors = append(validationErrors, fmt.Sprintf("step '%s': failed to access secret '%s': %v", step.Name, secretRef.SecretRef, err))
			}
			continue
		}

		// Check if the specified key exists in the secret
		if _, exists := secret.Data[secretRef.Key]; !exists {
			availableKeys := make([]string, 0, len(secret.Data))
			for key := range secret.Data {
				availableKeys = append(availableKeys, key)
			}
			validationErrors = append(validationErrors, fmt.Sprintf("step '%s': key '%s' not found in secret '%s'. Available keys: %s",
				step.Name, secretRef.Key, secretRef.SecretRef, strings.Join(availableKeys, ", ")))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("secret validation failed:\n  - %s", strings.Join(validationErrors, "\n  - "))
	}

	return nil
}

// ValidatePipelineConfig validates all secret references in a pipeline configuration
func (v *Validator) ValidatePipelineConfig(ctx context.Context, config *v1alpha1.PipelineConfig) error {
	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	var allErrors []string

	for i := range config.Spec.Steps {
		step := &config.Spec.Steps[i]
		if err := v.ValidateSecretReferences(ctx, step, namespace); err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("pipeline config validation failed:\n%s", strings.Join(allErrors, "\n"))
	}

	return nil
}

// ValidateSecret validates a Kubernetes Secret for use in C8S pipelines
func (v *Validator) ValidateSecret(secret *corev1.Secret) error {
	if secret == nil {
		return fmt.Errorf("secret cannot be nil")
	}

	if secret.Name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	if len(secret.Data) == 0 && len(secret.StringData) == 0 {
		return fmt.Errorf("secret '%s' has no data", secret.Name)
	}

	// Warn about large secret values (> 1MB) as they might cause issues
	for key, value := range secret.Data {
		if len(value) > 1024*1024 {
			return fmt.Errorf("secret '%s' key '%s' is larger than 1MB, which may cause performance issues", secret.Name, key)
		}
	}

	return nil
}

// GetMissingSecrets returns a list of secret references that don't exist
func (v *Validator) GetMissingSecrets(ctx context.Context, config *v1alpha1.PipelineConfig) ([]string, error) {
	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	missingSecrets := make(map[string]bool)

	for i := range config.Spec.Steps {
		step := &config.Spec.Steps[i]
		for _, secretRef := range step.Secrets {
			_, err := v.client.CoreV1().Secrets(namespace).Get(ctx, secretRef.SecretRef, metav1.GetOptions{})
			if err != nil && errors.IsNotFound(err) {
				missingSecrets[secretRef.SecretRef] = true
			}
		}
	}

	result := make([]string, 0, len(missingSecrets))
	for secretName := range missingSecrets {
		result = append(result, secretName)
	}

	return result, nil
}

// CheckSecretAccess verifies that the controller has permission to read a secret
func (v *Validator) CheckSecretAccess(ctx context.Context, secretName, namespace string) error {
	_, err := v.client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("secret '%s' not found in namespace '%s'", secretName, namespace)
		}
		if errors.IsForbidden(err) {
			return fmt.Errorf("access denied to secret '%s' in namespace '%s': insufficient permissions", secretName, namespace)
		}
		return fmt.Errorf("failed to access secret '%s': %w", secretName, err)
	}
	return nil
}
