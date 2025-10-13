package cli

import (
	"flag"
	"fmt"
	"os"

	v1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/parser"
)

func validateCommand(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("pipeline YAML file required")
	}

	filePath := fs.Arg(0)

	// Read file
	yamlContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	spec, err := parser.Parse(yamlContent)
	if err != nil {
		fmt.Printf("❌ Invalid pipeline configuration\n\n")
		fmt.Printf("Parse error: %v\n", err)
		return fmt.Errorf("validation failed")
	}

	// Create a full PipelineConfig for validation
	config := &v1alpha1.PipelineConfig{
		Spec: *spec,
	}

	// Validate configuration
	if err := parser.Validate(config); err != nil {
		fmt.Printf("❌ Invalid pipeline configuration\n\n")
		fmt.Printf("Validation error: %v\n", err)
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("✅ Valid pipeline configuration\n\n")
	fmt.Printf("Repository: %s\n", spec.Repository)
	fmt.Printf("Steps: %d\n", len(spec.Steps))

	return nil
}
