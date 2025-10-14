package health

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// HealthStatus represents overall health check status
type HealthStatus struct {
	Healthy bool          `json:"healthy"`
	Checks  []CheckResult `json:"checks"`
}

// CheckResult represents a single health check result
type CheckResult struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// Checker provides health check operations
type Checker struct {
	timeout time.Duration
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		timeout: 10 * time.Second,
	}
}

// CheckDocker checks if Docker daemon is available and running
func (c *Checker) CheckDocker(ctx context.Context) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:    "Docker",
			Healthy: false,
			Message: fmt.Sprintf("Docker daemon not available: %v", err),
		}
	}

	return CheckResult{
		Name:    "Docker",
		Healthy: true,
		Message: "Docker daemon is running",
	}
}

// CheckKubectl checks if kubectl is installed and accessible
func (c *Checker) CheckKubectl(ctx context.Context) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "version", "--client", "--output=json")
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:    "kubectl",
			Healthy: false,
			Message: fmt.Sprintf("kubectl not available: %v", err),
		}
	}

	return CheckResult{
		Name:    "kubectl",
		Healthy: true,
		Message: "kubectl is installed",
	}
}

// CheckClusterReady checks if a cluster is ready (nodes are Ready, API is accessible)
func (c *Checker) CheckClusterReady(ctx context.Context) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Get nodes as JSON
	cmd := exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:    "Cluster",
			Healthy: false,
			Message: fmt.Sprintf("Cannot access cluster API: %v", err),
		}
	}

	// Parse JSON to check node status
	var nodesResponse struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &nodesResponse); err != nil {
		return CheckResult{
			Name:    "Cluster",
			Healthy: false,
			Message: fmt.Sprintf("Cannot parse cluster status: %v", err),
		}
	}

	if len(nodesResponse.Items) == 0 {
		return CheckResult{
			Name:    "Cluster",
			Healthy: false,
			Message: "No nodes found in cluster",
		}
	}

	// Check if all nodes are Ready
	notReadyNodes := []string{}
	for _, node := range nodesResponse.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				isReady = true
				break
			}
		}
		if !isReady {
			notReadyNodes = append(notReadyNodes, node.Metadata.Name)
		}
	}

	if len(notReadyNodes) > 0 {
		return CheckResult{
			Name:    "Cluster",
			Healthy: false,
			Message: fmt.Sprintf("Nodes not ready: %s", strings.Join(notReadyNodes, ", ")),
		}
	}

	return CheckResult{
		Name:    "Cluster",
		Healthy: true,
		Message: fmt.Sprintf("Cluster is ready (%d nodes)", len(nodesResponse.Items)),
	}
}

// CheckCRDRegistered checks if a CRD is registered in the cluster
func (c *Checker) CheckCRDRegistered(ctx context.Context, crdName string) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "get", "crd", crdName)
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:    fmt.Sprintf("CRD:%s", crdName),
			Healthy: false,
			Message: fmt.Sprintf("CRD %s not registered", crdName),
		}
	}

	return CheckResult{
		Name:    fmt.Sprintf("CRD:%s", crdName),
		Healthy: true,
		Message: fmt.Sprintf("CRD %s is registered", crdName),
	}
}

// CheckPodStatus checks if a pod is running and ready
func (c *Checker) CheckPodStatus(ctx context.Context, namespace, labelSelector string) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := []string{"get", "pods", "-n", namespace, "-l", labelSelector, "-o", "json"}
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:    fmt.Sprintf("Pod:%s", labelSelector),
			Healthy: false,
			Message: fmt.Sprintf("Cannot get pod status: %v", err),
		}
	}

	// Parse JSON to check pod status
	var podsResponse struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Phase      string `json:"phase"`
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podsResponse); err != nil {
		return CheckResult{
			Name:    fmt.Sprintf("Pod:%s", labelSelector),
			Healthy: false,
			Message: fmt.Sprintf("Cannot parse pod status: %v", err),
		}
	}

	if len(podsResponse.Items) == 0 {
		return CheckResult{
			Name:    fmt.Sprintf("Pod:%s", labelSelector),
			Healthy: false,
			Message: "No pods found",
		}
	}

	// Check if all pods are Running and Ready
	notReadyPods := []string{}
	for _, pod := range podsResponse.Items {
		if pod.Status.Phase != "Running" {
			notReadyPods = append(notReadyPods, fmt.Sprintf("%s (phase: %s)", pod.Metadata.Name, pod.Status.Phase))
			continue
		}

		isReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				isReady = true
				break
			}
		}
		if !isReady {
			notReadyPods = append(notReadyPods, fmt.Sprintf("%s (not ready)", pod.Metadata.Name))
		}
	}

	if len(notReadyPods) > 0 {
		return CheckResult{
			Name:    fmt.Sprintf("Pod:%s", labelSelector),
			Healthy: false,
			Message: fmt.Sprintf("Pods not ready: %s", strings.Join(notReadyPods, ", ")),
		}
	}

	return CheckResult{
		Name:    fmt.Sprintf("Pod:%s", labelSelector),
		Healthy: true,
		Message: fmt.Sprintf("%d pod(s) running and ready", len(podsResponse.Items)),
	}
}

// CheckAll runs all basic health checks
func (c *Checker) CheckAll(ctx context.Context) *HealthStatus {
	checks := []CheckResult{
		c.CheckDocker(ctx),
		c.CheckKubectl(ctx),
	}

	// Overall health is true only if all checks pass
	healthy := true
	for _, check := range checks {
		if !check.Healthy {
			healthy = false
			break
		}
	}

	return &HealthStatus{
		Healthy: healthy,
		Checks:  checks,
	}
}
