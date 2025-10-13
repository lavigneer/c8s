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

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/secrets"
)

// AdmissionWebhook handles admission webhook requests for PipelineConfig validation
type AdmissionWebhook struct {
	client    kubernetes.Interface
	validator *secrets.Validator
}

// NewAdmissionWebhook creates a new admission webhook handler
func NewAdmissionWebhook(client kubernetes.Interface) *AdmissionWebhook {
	return &AdmissionWebhook{
		client:    client,
		validator: secrets.NewValidator(client),
	}
}

// HandleValidation handles ValidatingAdmissionWebhook requests for PipelineConfig
func (aw *AdmissionWebhook) HandleValidation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	// Read the admission review request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(err, "failed to read request body")
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the admission review
	admissionReview := &admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, admissionReview); err != nil {
		logger.Error(err, "failed to unmarshal admission review")
		http.Error(w, "failed to parse admission review", http.StatusBadRequest)
		return
	}

	// Validate the request
	if admissionReview.Request == nil {
		logger.Error(fmt.Errorf("admission review request is nil"), "invalid admission review")
		http.Error(w, "invalid admission review: request is nil", http.StatusBadRequest)
		return
	}

	// Process the admission request
	response := aw.validatePipelineConfig(ctx, admissionReview.Request)

	// Build the admission review response
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	// Marshal and send the response
	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		logger.Error(err, "failed to marshal admission response")
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

// validatePipelineConfig validates a PipelineConfig for secret references
func (aw *AdmissionWebhook) validatePipelineConfig(ctx context.Context, req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	logger := log.FromContext(ctx)

	// Only validate CREATE and UPDATE operations
	if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Message: fmt.Sprintf("operation %s is not validated", req.Operation),
			},
		}
	}

	// Parse the PipelineConfig
	pipelineConfig := &v1alpha1.PipelineConfig{}
	if err := json.Unmarshal(req.Object.Raw, pipelineConfig); err != nil {
		logger.Error(err, "failed to unmarshal PipelineConfig")
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("failed to parse PipelineConfig: %v", err),
				Code:    http.StatusBadRequest,
			},
		}
	}

	// Set namespace from request if not set in object
	if pipelineConfig.Namespace == "" {
		pipelineConfig.Namespace = req.Namespace
	}

	logger.Info("validating PipelineConfig", "name", pipelineConfig.Name, "namespace", pipelineConfig.Namespace)

	// Validate secret references
	if err := aw.validator.ValidatePipelineConfig(ctx, pipelineConfig); err != nil {
		logger.Info("PipelineConfig validation failed", "name", pipelineConfig.Name, "error", err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("PipelineConfig validation failed: %v", err),
				Code:    http.StatusUnprocessableEntity,
			},
		}
	}

	logger.Info("PipelineConfig validation passed", "name", pipelineConfig.Name)

	// Validation passed
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Result: &metav1.Status{
			Status:  "Success",
			Message: "PipelineConfig is valid",
		},
	}
}

// HandleHealth handles health check requests
func (aw *AdmissionWebhook) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleReady handles readiness check requests
func (aw *AdmissionWebhook) HandleReady(w http.ResponseWriter, r *http.Request) {
	// Check if we can connect to the API server
	ctx := r.Context()
	_, err := aw.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		http.Error(w, "not ready: cannot connect to API server", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}
