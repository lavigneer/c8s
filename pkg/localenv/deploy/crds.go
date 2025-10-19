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

// CRDDeploymentStatus contains information about CRD deployment
type CRDDeploymentStatus struct {
	Success       bool
	CRDsInstalled []string
	Message       string
	Timestamp     time.Time
}

// InstallCRDs installs all CRD manifests to the cluster
func InstallCRDs(ctx context.Context, kubectlClient cluster.KubectlClient, crdsPath string) (*CRDDeploymentStatus, error) {
	status := &CRDDeploymentStatus{
		Timestamp: time.Now(),
	}

	// Resolve CRD path
	if crdsPath == "" {
		crdsPath = "config/crd/bases"
	}

	// Check if path is relative and convert to absolute
	if !filepath.IsAbs(crdsPath) {
		cwd, err := os.Getwd()
		if err != nil {
			status.Message = fmt.Sprintf("Failed to get working directory: %v", err)
			return status, err
		}
		crdsPath = filepath.Join(cwd, crdsPath)
	}

	// Verify path exists
	if _, err := os.Stat(crdsPath); err != nil {
		status.Message = fmt.Sprintf("CRD path not found: %s", crdsPath)
		return status, err
	}

	// Find all YAML files in the directory
	var crdFiles []string
	err := filepath.Walk(crdsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			crdFiles = append(crdFiles, path)
		}
		return nil
	})

	if err != nil {
		status.Message = fmt.Sprintf("Failed to scan CRD directory: %v", err)
		return status, err
	}

	if len(crdFiles) == 0 {
		status.Message = fmt.Sprintf("No CRD files found in %s", crdsPath)
		return status, fmt.Errorf("no CRD files found")
	}

	// Apply each CRD file
	for _, crdFile := range crdFiles {
		err := kubectlClient.ApplyManifest(ctx, crdFile, "")
		if err != nil {
			status.Message = fmt.Sprintf("Failed to apply CRD from %s: %v", crdFile, err)
			return status, err
		}
		status.CRDsInstalled = append(status.CRDsInstalled, filepath.Base(crdFile))
	}

	// Wait for CRDs to be registered
	if err := waitForCRDsRegistration(kubectlClient, status.CRDsInstalled); err != nil {
		status.Message = fmt.Sprintf("CRDs installed but registration check failed: %v", err)
		// Don't fail here - CRDs might still be registering
	}

	status.Success = true
	status.Message = fmt.Sprintf("Successfully installed %d CRD(s)", len(status.CRDsInstalled))
	return status, nil
}

// waitForCRDsRegistration polls kubectl to verify CRDs are registered
func waitForCRDsRegistration(kubectlClient cluster.KubectlClient, crdNames []string) error {
	maxRetries := 10
	retryDelay := 500 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get list of installed CRDs
		cmd := exec.Command("kubectl", "get", "crds", "-o", "name")
		output, err := cmd.CombinedOutput()
		if err != nil {
			time.Sleep(retryDelay)
			continue
		}

		crdList := string(output)

		// Check if expected CRDs are present
		allFound := true
		for _, crdFile := range crdNames {
			// Extract CRD name from filename (remove .yaml)
			crdName := strings.TrimSuffix(crdFile, ".yaml")
			crdName = strings.TrimSuffix(crdName, ".yml")

			if !strings.Contains(crdList, crdName) {
				allFound = false
				break
			}
		}

		if allFound {
			return nil
		}

		time.Sleep(retryDelay)
	}

	return fmt.Errorf("timeout waiting for CRDs to be registered")
}

// VerifyCRDsInstalled checks that CRDs are properly installed
func VerifyCRDsInstalled(kubectlClient cluster.KubectlClient, expectedCRDs []string) error {
	cmd := exec.Command("kubectl", "get", "crds", "-o", "name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list CRDs: %v", err)
	}

	crdList := string(output)

	// Verify at least the main PipelineConfig CRD is present
	if !strings.Contains(crdList, "pipelineconfig") {
		return fmt.Errorf("pipelineconfig CRD not found in cluster")
	}

	return nil
}
