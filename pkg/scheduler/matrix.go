package scheduler

import (
	"fmt"
	"strings"

	"github.com/org/c8s/pkg/apis/v1alpha1"
)

// ExpandMatrix generates all combinations of matrix dimensions
// Returns a slice of maps where each map represents one matrix combination
func ExpandMatrix(matrix *v1alpha1.MatrixStrategy) ([]map[string]string, error) {
	if matrix == nil {
		return []map[string]string{{}}, nil
	}

	if len(matrix.Dimensions) == 0 {
		return []map[string]string{{}}, nil
	}

	// Validate that all dimensions have at least one value
	for key, values := range matrix.Dimensions {
		if len(values) == 0 {
			return nil, fmt.Errorf("matrix dimension %s has no values", key)
		}
	}

	// Generate all combinations recursively
	combinations := generateCombinations(matrix.Dimensions)

	// Filter out excluded combinations
	filtered := filterExclusions(combinations, matrix.Exclude)

	if len(filtered) == 0 {
		return nil, fmt.Errorf("all matrix combinations are excluded")
	}

	return filtered, nil
}

// generateCombinations recursively generates all combinations of dimension values
func generateCombinations(dimensions map[string][]string) []map[string]string {
	if len(dimensions) == 0 {
		return []map[string]string{{}}
	}

	// Convert map to slice for consistent ordering
	type dimPair struct {
		key    string
		values []string
	}
	var dims []dimPair
	for k, v := range dimensions {
		dims = append(dims, dimPair{k, v})
	}

	// Base case: one dimension
	if len(dims) == 1 {
		var result []map[string]string
		for _, value := range dims[0].values {
			result = append(result, map[string]string{
				dims[0].key: value,
			})
		}
		return result
	}

	// Recursive case: split into first dimension and rest
	firstDim := dims[0]
	restDims := make(map[string][]string)
	for _, dim := range dims[1:] {
		restDims[dim.key] = dim.values
	}

	restCombinations := generateCombinations(restDims)

	var result []map[string]string
	for _, value := range firstDim.values {
		for _, restCombo := range restCombinations {
			combo := make(map[string]string)
			combo[firstDim.key] = value
			for k, v := range restCombo {
				combo[k] = v
			}
			result = append(result, combo)
		}
	}

	return result
}

// filterExclusions removes combinations that match exclusion rules
func filterExclusions(combinations []map[string]string, exclusions []map[string]string) []map[string]string {
	if len(exclusions) == 0 {
		return combinations
	}

	var filtered []map[string]string
	for _, combo := range combinations {
		if !isExcluded(combo, exclusions) {
			filtered = append(filtered, combo)
		}
	}

	return filtered
}

// isExcluded checks if a combination matches any exclusion rule
func isExcluded(combo map[string]string, exclusions []map[string]string) bool {
	for _, exclusion := range exclusions {
		if matchesExclusion(combo, exclusion) {
			return true
		}
	}
	return false
}

// matchesExclusion checks if a combination matches a specific exclusion rule
// All keys in the exclusion must match the combination
func matchesExclusion(combo map[string]string, exclusion map[string]string) bool {
	for key, value := range exclusion {
		comboValue, exists := combo[key]
		if !exists || comboValue != value {
			return false
		}
	}
	return true
}

// SubstituteMatrixVariables replaces matrix variable placeholders in a string
// Placeholders have format: ${{matrix.key}}
func SubstituteMatrixVariables(template string, matrixVars map[string]string) string {
	result := template
	for key, value := range matrixVars {
		placeholder := fmt.Sprintf("${{matrix.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)

		// Also support ${matrix.key} format
		placeholder2 := fmt.Sprintf("${matrix.%s}", key)
		result = strings.ReplaceAll(result, placeholder2, value)
	}
	return result
}

// ApplyMatrixToStep creates a new step with matrix variables substituted
func ApplyMatrixToStep(step v1alpha1.PipelineStep, matrixVars map[string]string) v1alpha1.PipelineStep {
	newStep := step

	// Substitute in image
	newStep.Image = SubstituteMatrixVariables(step.Image, matrixVars)

	// Substitute in commands
	newStep.Commands = make([]string, len(step.Commands))
	for i, cmd := range step.Commands {
		newStep.Commands[i] = SubstituteMatrixVariables(cmd, matrixVars)
	}

	// Substitute in step name if it contains matrix variables
	newStep.Name = SubstituteMatrixVariables(step.Name, matrixVars)

	// Substitute in environment variables (if we add that feature)
	// This would allow matrix variables in env values

	return newStep
}

// GenerateMatrixRunName generates a unique name for a matrix run
func GenerateMatrixRunName(baseName string, matrixIndex int, matrixVars map[string]string) string {
	// Create a descriptive suffix from matrix variables
	var parts []string
	for k, v := range matrixVars {
		// Sanitize values for use in names (Kubernetes names must be DNS-1123 compliant)
		sanitized := strings.ToLower(v)
		sanitized = strings.ReplaceAll(sanitized, ":", "-")
		sanitized = strings.ReplaceAll(sanitized, ".", "-")
		sanitized = strings.ReplaceAll(sanitized, "/", "-")
		parts = append(parts, fmt.Sprintf("%s-%s", k, sanitized))
	}

	// Sort for consistency
	// We don't actually need to sort since map iteration is random,
	// but the index will keep things unique anyway

	suffix := strings.Join(parts, "-")
	if len(suffix) > 40 {
		suffix = suffix[:40]
	}

	return fmt.Sprintf("%s-matrix-%d-%s", baseName, matrixIndex, suffix)
}

// MatrixToLabels converts matrix variables to Kubernetes labels
func MatrixToLabels(matrixVars map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range matrixVars {
		// Sanitize for label values
		sanitized := strings.ToLower(v)
		sanitized = strings.ReplaceAll(sanitized, ":", "-")
		sanitized = strings.ReplaceAll(sanitized, ".", "-")
		sanitized = strings.ReplaceAll(sanitized, "/", "-")

		// Truncate if too long (label values max 63 chars)
		if len(sanitized) > 63 {
			sanitized = sanitized[:63]
		}

		labelKey := fmt.Sprintf("c8s.dev/matrix-%s", k)
		labels[labelKey] = sanitized
	}
	return labels
}
