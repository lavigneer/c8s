package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CleanupStatus contains information about cluster cleanup
type CleanupStatus struct {
	Success            bool
	OrphanedContainers []string
	OrphanedVolumes    []string
	OrphanedContexts   []string
	Message            string
}

// VerifyCleanupStatus checks for orphaned resources after cluster deletion
func VerifyCleanupStatus(clusterName string) (*CleanupStatus, error) {
	status := &CleanupStatus{}

	// Check for orphaned Docker containers
	containers, err := findOrphanedContainers(clusterName)
	if err == nil {
		status.OrphanedContainers = containers
	}

	// Check for orphaned Docker volumes
	volumes, err := findOrphanedVolumes(clusterName)
	if err == nil {
		status.OrphanedVolumes = volumes
	}

	// Check for orphaned kubeconfig contexts
	contexts, err := findOrphanedContexts(clusterName)
	if err == nil {
		status.OrphanedContexts = contexts
	}

	// Determine overall success
	if len(status.OrphanedContainers) == 0 && len(status.OrphanedVolumes) == 0 && len(status.OrphanedContexts) == 0 {
		status.Success = true
		status.Message = "Cluster cleanup verified - no orphaned resources"
	} else {
		status.Message = fmt.Sprintf("Found orphaned resources: %d containers, %d volumes, %d contexts",
			len(status.OrphanedContainers), len(status.OrphanedVolumes), len(status.OrphanedContexts))
	}

	return status, nil
}

// findOrphanedContainers finds Docker containers related to the cluster
func findOrphanedContainers(clusterName string) ([]string, error) {
	cmd := exec.Command("docker", "ps", "-a", "-q", "-f", fmt.Sprintf("label=io.k3d.cluster=%s", clusterName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	containers := strings.Fields(string(output))
	return containers, nil
}

// findOrphanedVolumes finds Docker volumes related to the cluster
func findOrphanedVolumes(clusterName string) ([]string, error) {
	cmd := exec.Command("docker", "volume", "ls", "-q", "-f", fmt.Sprintf("label=io.k3d.cluster=%s", clusterName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	volumes := strings.Fields(string(output))
	return volumes, nil
}

// findOrphanedContexts finds kubeconfig contexts related to the cluster
func findOrphanedContexts(clusterName string) ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	contexts := strings.Fields(string(output))
	var orphaned []string

	for _, ctx := range contexts {
		if strings.Contains(ctx, clusterName) {
			orphaned = append(orphaned, ctx)
		}
	}

	return orphaned, nil
}

// CleanupOrphanedResources removes orphaned resources
func CleanupOrphanedResources(clusterName string) error {
	// Remove Docker containers
	containers, _ := findOrphanedContainers(clusterName)
	for _, container := range containers {
		exec.Command("docker", "rm", "-f", container).Run()
	}

	// Remove Docker volumes
	volumes, _ := findOrphanedVolumes(clusterName)
	for _, volume := range volumes {
		exec.Command("docker", "volume", "rm", volume).Run()
	}

	// Remove kubeconfig contexts
	contexts, _ := findOrphanedContexts(clusterName)
	for _, context := range contexts {
		exec.Command("kubectl", "config", "delete-context", context).Run()
	}

	// Clean up kubeconfig file
	cleanupKubeconfig(clusterName)

	return nil
}

// cleanupKubeconfig removes cluster entries from kubeconfig
func cleanupKubeconfig(clusterName string) error {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	// Kubectl config delete-cluster and delete-user commands
	exec.Command("kubectl", "config", "delete-cluster", "k3d-"+clusterName).Run()
	exec.Command("kubectl", "config", "delete-user", "admin@k3d-"+clusterName).Run()

	return nil
}
