package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/org/c8s/pkg/localenv/cluster"
)

// OperatorDeploymentStatus contains information about operator deployment
type OperatorDeploymentStatus struct {
	Success      bool
	Namespace    string
	Message      string
	DeploymentName string
	PodCount     int
	Timestamp    time.Time
}

// DeployOperator deploys the C8S operator to the cluster
func DeployOperator(
	ctx context.Context,
	kubectlClient cluster.KubectlClient,
	clusterName string,
	namespace string,
	manifestsPath string,
	imageName string,
	imagePullPolicy string,
) (*OperatorDeploymentStatus, error) {
	status := &OperatorDeploymentStatus{
		Namespace: namespace,
		Timestamp: time.Now(),
	}

	if namespace == "" {
		namespace = "c8s-system"
		status.Namespace = namespace
	}

	if imagePullPolicy == "" {
		imagePullPolicy = "IfNotPresent"
	}

	// Create namespace if it doesn't exist
	createNSCmd := exec.Command("kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	createNSPipeCmd := exec.Command("kubectl", "apply", "-f", "-")
	createNSPipeCmd.Stdin, _ = createNSCmd.StdoutPipe()
	if err := createNSPipeCmd.Run(); err != nil {
		// Namespace might already exist, continue
	}
	createNSCmd.Run()

	// Resolve manifests path
	if manifestsPath == "" {
		manifestsPath = "config/manager"
	}

	// Check if path is relative and convert to absolute
	if !filepath.IsAbs(manifestsPath) {
		cwd, err := os.Getwd()
		if err != nil {
			status.Message = fmt.Sprintf("Failed to get working directory: %v", err)
			return status, err
		}
		manifestsPath = filepath.Join(cwd, manifestsPath)
	}

	// Verify path exists
	if _, err := os.Stat(manifestsPath); err != nil {
		status.Message = fmt.Sprintf("Manifests path not found: %s", manifestsPath)
		return status, err
	}

	// Find all YAML files in the directory
	var manifestFiles []string
	err := filepath.Walk(manifestsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			manifestFiles = append(manifestFiles, path)
		}
		return nil
	})

	if err != nil {
		status.Message = fmt.Sprintf("Failed to scan manifests directory: %v", err)
		return status, err
	}

	if len(manifestFiles) == 0 {
		status.Message = fmt.Sprintf("No manifest files found in %s", manifestsPath)
		return status, fmt.Errorf("no manifest files found")
	}

	// Apply manifests with namespace and image overrides
	for _, manifestFile := range manifestFiles {
		err := applyManifestWithOverrides(manifestFile, namespace, imageName, imagePullPolicy)
		if err != nil {
			status.Message = fmt.Sprintf("Failed to apply manifest %s: %v", manifestFile, err)
			return status, err
		}
	}

	// Wait for operator deployment to be ready
	if err := waitForDeploymentReady(namespace, "c8s-controller", 5*time.Minute); err != nil {
		status.Message = fmt.Sprintf("Operator deployment did not become ready: %v", err)
		// Check operator logs for debugging
		logs, _ := getDeploymentLogs(namespace, "c8s-controller")
		status.Message = fmt.Sprintf("Operator deployment did not become ready: %v\nLogs:\n%s", err, logs)
		return status, err
	}

	// Get deployment info
	status.DeploymentName = "c8s-controller"
	status.Success = true
	status.Message = fmt.Sprintf("Successfully deployed operator to namespace %s", namespace)

	return status, nil
}

// applyManifestWithOverrides applies a manifest with image and namespace overrides
func applyManifestWithOverrides(manifestPath string, namespace string, imageName string, imagePullPolicy string) error {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	manifestStr := string(content)

	// Apply namespace override
	manifestStr = strings.ReplaceAll(manifestStr, "namespace: default", fmt.Sprintf("namespace: %s", namespace))
	manifestStr = strings.ReplaceAll(manifestStr, "NAMESPACE_PLACEHOLDER", namespace)

	// Apply image override if specified
	if imageName != "" {
		manifestStr = strings.ReplaceAll(manifestStr, "IMAGE_PLACEHOLDER", imageName)
		// Also try common image field patterns
		manifestStr = replaceImageInManifest(manifestStr, imageName)
	}

	// Apply image pull policy
	if imagePullPolicy != "" {
		manifestStr = strings.ReplaceAll(manifestStr, "imagePullPolicy: Always", fmt.Sprintf("imagePullPolicy: %s", imagePullPolicy))
		manifestStr = strings.ReplaceAll(manifestStr, "imagePullPolicy: IfNotPresent", fmt.Sprintf("imagePullPolicy: %s", imagePullPolicy))
	}

	// Apply the manifest using kubectl
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifestStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// replaceImageInManifest replaces container images in manifest YAML
func replaceImageInManifest(manifest string, imageName string) string {
	lines := strings.Split(manifest, "\n")
	var result []string

	for _, line := range lines {
		if strings.Contains(line, "image:") && strings.Contains(line, "controller") {
			// Replace the image line
			indent := len(line) - len(strings.TrimLeft(line, " "))
			result = append(result, fmt.Sprintf("%*simage: %s", indent, "", imageName))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// waitForDeploymentReady waits for a deployment to have all replicas ready
func waitForDeploymentReady(namespace string, deploymentName string, timeout time.Duration) error {
	startTime := time.Now()

	for time.Since(startTime) < timeout {
		cmd := exec.Command("kubectl", "-n", namespace, "get", "deployment", deploymentName, "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
		output, err := cmd.CombinedOutput()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		status := string(output)
		if status == "1/1" || status == "2/2" || strings.HasPrefix(status, strings.TrimSpace(status)) {
			// Parse the ratio
			parts := strings.Split(status, "/")
			if len(parts) == 2 && parts[0] == parts[1] && parts[0] != "0" {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("deployment %s did not become ready within %v", deploymentName, timeout)
}

// getDeploymentLogs fetches logs from operator pods for debugging
func getDeploymentLogs(namespace string, deploymentName string) (string, error) {
	// Get pods for the deployment
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", fmt.Sprintf("app=%s", deploymentName), "-o", "jsonpath={.items[0].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	podName := strings.TrimSpace(string(output))
	if podName == "" {
		return "No pods found", nil
	}

	// Get logs from the pod
	logsCmd := exec.Command("kubectl", "-n", namespace, "logs", podName, "--tail=50")
	logs, _ := logsCmd.CombinedOutput()

	return string(logs), nil
}
