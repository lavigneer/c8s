package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/org/c8s/pkg/apis/v1alpha1"
)

// QuotaAdmissionWebhook validates PipelineRuns against namespace ResourceQuotas
type QuotaAdmissionWebhook struct {
	client    client.Client
	clientset *kubernetes.Clientset
}

// NewQuotaAdmissionWebhook creates a new quota admission webhook
func NewQuotaAdmissionWebhook(c client.Client, cs *kubernetes.Clientset) *QuotaAdmissionWebhook {
	return &QuotaAdmissionWebhook{
		client:    c,
		clientset: cs,
	}
}

// Handle processes admission requests for PipelineRuns
func (w *QuotaAdmissionWebhook) Handle(ctx context.Context, req admissionv1.AdmissionRequest) admissionv1.AdmissionResponse {
	// Only validate CREATE operations
	if req.Operation != admissionv1.Create {
		return admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Decode PipelineRun
	var pipelineRun v1alpha1.PipelineRun
	if err := json.Unmarshal(req.Object.Raw, &pipelineRun); err != nil {
		return admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to decode PipelineRun: %v", err),
			},
		}
	}

	// Fetch PipelineConfig to get step details
	var pipelineConfig v1alpha1.PipelineConfig
	configKey := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      pipelineRun.Spec.PipelineConfigRef,
	}
	if err := w.client.Get(ctx, configKey, &pipelineConfig); err != nil {
		return admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to fetch PipelineConfig %s: %v", pipelineRun.Spec.PipelineConfigRef, err),
			},
		}
	}

	// Calculate total resources required
	totalCPU, totalMemory := calculateTotalResources(pipelineConfig.Spec.Steps)

	// Fetch namespace ResourceQuota
	quotaList, err := w.clientset.CoreV1().ResourceQuotas(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// If quota check fails, allow by default (fail open)
		return admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// If no quotas defined, allow
	if len(quotaList.Items) == 0 {
		return admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Check against each quota
	for _, quota := range quotaList.Items {
		if err := checkQuota(quota, totalCPU, totalMemory); err != nil {
			return admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: err.Error(),
				},
			}
		}
	}

	return admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

// calculateTotalResources sums all step resource requirements
func calculateTotalResources(steps []v1alpha1.PipelineStep) (cpu, memory resource.Quantity) {
	cpu = resource.MustParse("0")
	memory = resource.MustParse("0")

	for _, step := range steps {
		if step.Resources.CPU != "" {
			stepCPU, err := resource.ParseQuantity(step.Resources.CPU)
			if err == nil {
				cpu.Add(stepCPU)
			}
		} else {
			// Default: 1 CPU if not specified
			cpu.Add(resource.MustParse("1"))
		}

		if step.Resources.Memory != "" {
			stepMemory, err := resource.ParseQuantity(step.Resources.Memory)
			if err == nil {
				memory.Add(stepMemory)
			}
		} else {
			// Default: 2Gi if not specified
			memory.Add(resource.MustParse("2Gi"))
		}
	}

	return cpu, memory
}

// checkQuota validates if the requested resources would exceed quota
func checkQuota(quota corev1.ResourceQuota, requestedCPU, requestedMemory resource.Quantity) error {
	// Check CPU quota
	if hardCPU, ok := quota.Status.Hard[corev1.ResourceRequestsCPU]; ok {
		usedCPU := quota.Status.Used[corev1.ResourceRequestsCPU]
		availableCPU := hardCPU.DeepCopy()
		availableCPU.Sub(usedCPU)

		if requestedCPU.Cmp(availableCPU) > 0 {
			totalRequested := requestedCPU.DeepCopy()
			totalRequested.Add(usedCPU)
			return fmt.Errorf("would exceed CPU quota: %s/%s cores requested (requested: %s, available: %s)",
				totalRequested.String(), hardCPU.String(), requestedCPU.String(), availableCPU.String())
		}
	}

	// Check Memory quota
	if hardMemory, ok := quota.Status.Hard[corev1.ResourceRequestsMemory]; ok {
		usedMemory := quota.Status.Used[corev1.ResourceRequestsMemory]
		availableMemory := hardMemory.DeepCopy()
		availableMemory.Sub(usedMemory)

		if requestedMemory.Cmp(availableMemory) > 0 {
			totalRequested := requestedMemory.DeepCopy()
			totalRequested.Add(usedMemory)
			return fmt.Errorf("would exceed memory quota: %s/%s requested (requested: %s, available: %s)",
				totalRequested.String(), hardMemory.String(), requestedMemory.String(), availableMemory.String())
		}
	}

	return nil
}

// ServeHTTP implements http.Handler
func (w *QuotaAdmissionWebhook) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview

	if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
		http.Error(rw, fmt.Sprintf("Failed to decode admission review: %v", err), http.StatusBadRequest)
		return
	}

	if admissionReview.Request == nil {
		http.Error(rw, "Missing admission request", http.StatusBadRequest)
		return
	}

	response := w.Handle(r.Context(), *admissionReview.Request)
	response.UID = admissionReview.Request.UID

	admissionReview.Response = &response

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(admissionReview)
}
