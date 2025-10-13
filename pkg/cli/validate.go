package cli

import (
	"flag"
	"fmt"
	"os"

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
	config, err := parser.Parse(yamlContent)
	if err != nil {
		fmt.Printf("❌ Invalid pipeline configuration\n\n")
		fmt.Printf("Parse error: %v\n", err)
		return fmt.Errorf("validation failed")
	}

	// Validate configuration
	validator := parser.NewValidator()
	if err := validator.Validate(config); err != nil {
		fmt.Printf("❌ Invalid pipeline configuration\n\n")
		fmt.Printf("Validation error: %v\n", err)
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("✅ Valid pipeline configuration\n\n")
	fmt.Printf("Pipeline: %s\n", config.Name)
	fmt.Printf("Repository: %s\n", config.Spec.Repository)
	fmt.Printf("Steps: %d\n", len(config.Spec.Steps))

	return nil
}
