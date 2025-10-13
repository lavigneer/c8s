package cli

import (
	"context"
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "c8s.io",
	Version:  "v1alpha1",
	Resource: "pipelineruns",
}

func runCommand(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	commit := fs.String("commit", "", "commit SHA to build (required)")
	branch := fs.String("branch", "", "branch name (required)")
	triggeredBy := fs.String("triggered-by", "manual", "who triggered this run")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("pipeline config name required")
	}

	configName := fs.Arg(0)

	if *commit == "" {
		return fmt.Errorf("--commit flag is required")
	}
	if *branch == "" {
		return fmt.Errorf("--branch flag is required")
	}

	// Create dynamic client for CRDs
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Generate unique name for PipelineRun
	runName := fmt.Sprintf("%s-%d", configName, time.Now().Unix())

	// Create PipelineRun object
	pipelineRun := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "c8s.io/v1alpha1",
			"kind":       "PipelineRun",
			"metadata": map[string]interface{}{
				"name":      runName,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"c8s.io/pipeline-config": configName,
					"c8s.io/branch":          *branch,
				},
			},
			"spec": map[string]interface{}{
				"pipelineConfigRef": map[string]interface{}{
					"name": configName,
				},
				"commit":      *commit,
				"branch":      *branch,
				"triggeredBy": *triggeredBy,
				"triggeredAt": time.Now().Format(time.RFC3339),
			},
		},
	}

	ctx := context.Background()

	// Create the PipelineRun
	result, err := dynamicClient.Resource(pipelineRunGVR).Namespace(namespace).Create(
		ctx,
		pipelineRun,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create PipelineRun: %w", err)
	}

	fmt.Printf("PipelineRun created: %s\n", result.GetName())
	fmt.Printf("\nTo view status:\n  c8s get runs %s\n\n", result.GetName())
	fmt.Printf("To stream logs:\n  c8s logs %s --step=<step-name> --follow\n", result.GetName())

	return nil
}
