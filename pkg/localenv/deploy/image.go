package deploy

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ImageLoadStatus contains information about image loading
type ImageLoadStatus struct {
	Success   bool
	ImageName string
	Message   string
	Timestamp time.Time
}

// LoadImageToCluster loads a Docker image into the k3d cluster
func LoadImageToCluster(clusterName string, imageName string) (*ImageLoadStatus, error) {
	status := &ImageLoadStatus{
		ImageName: imageName,
		Timestamp: time.Now(),
	}

	// Use default image if not specified
	if imageName == "" {
		imageName = "ghcr.io/org/c8s-controller:latest"
	}

	// Check if image exists locally
	checkCmd := exec.Command("docker", "image", "inspect", imageName)
	if err := checkCmd.Run(); err != nil {
		// Try to pull the image
		pullCmd := exec.Command("docker", "pull", imageName)
		pullCmd.Stdout = nil
		pullCmd.Stderr = nil
		if err := pullCmd.Run(); err != nil {
			status.Message = fmt.Sprintf("Image not found locally and pull failed: %s", imageName)
			return status, fmt.Errorf("image not found: %s", imageName)
		}
	}

	// Load image into k3d cluster
	loadCmd := exec.Command("k3d", "image", "import", imageName, "-c", clusterName)
	output, err := loadCmd.CombinedOutput()
	if err != nil {
		status.Message = fmt.Sprintf("Failed to import image to cluster: %v\nOutput: %s", err, output)
		return status, err
	}

	status.Success = true
	status.Message = fmt.Sprintf("Successfully loaded image %s to cluster %s", imageName, clusterName)
	return status, nil
}

// VerifyImageInCluster checks if an image is available in the cluster
func VerifyImageInCluster(clusterName string, imageName string) error {
	// Query cluster for the image by trying to use it in a test pod
	cmd := exec.Command("kubectl", "--context", "k3d-"+clusterName,
		"run", "test-image-check", "--image="+imageName, "--rm", "-i",
		"--restart=Never", "--", "echo", "ok")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("image verification failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// LoadImagesFromRegistry handles loading images from a registry
func LoadImagesFromRegistry(clusterName string, registryName string, images []string) error {
	if registryName == "" {
		registryName = "docker.io"
	}

	for _, image := range images {
		// Ensure full image path includes registry
		if !strings.Contains(image, "/") {
			image = registryName + "/library/" + image
		} else if !strings.Contains(strings.Split(image, "/")[0], ".") {
			// No registry specified (no dots in first part), prepend default registry
			image = registryName + "/" + image
		}

		_, err := LoadImageToCluster(clusterName, image)
		if err != nil {
			return err
		}
	}

	return nil
}
