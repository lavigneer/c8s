package controller

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/org/c8s/pkg/apis/v1alpha1"
)

// CalculateTotalResources sums all step resource requirements
// Returns CPU and memory as Kubernetes quantity strings
func CalculateTotalResources(steps []v1alpha1.PipelineStep) (cpu, memory string) {
	totalCPU := resource.MustParse("0")
	totalMemory := resource.MustParse("0")

	for _, step := range steps {
		// Add CPU
		if step.Resources.CPU != "" {
			stepCPU, err := resource.ParseQuantity(step.Resources.CPU)
			if err == nil {
				totalCPU.Add(stepCPU)
			}
		} else {
			// Default: 1 CPU core if not specified
			totalCPU.Add(resource.MustParse("1"))
		}

		// Add Memory
		if step.Resources.Memory != "" {
			stepMemory, err := resource.ParseQuantity(step.Resources.Memory)
			if err == nil {
				totalMemory.Add(stepMemory)
			}
		} else {
			// Default: 2Gi if not specified
			totalMemory.Add(resource.MustParse("2Gi"))
		}
	}

	return totalCPU.String(), totalMemory.String()
}

// GetDefaultResources returns default resource requirements
func GetDefaultResources() v1alpha1.ResourceRequirements {
	return v1alpha1.ResourceRequirements{
		CPU:    "1",
		Memory: "2Gi",
	}
}

// ParseResources parses resource strings into Kubernetes quantities
func ParseResources(cpu, memory string) (resource.Quantity, resource.Quantity, error) {
	cpuQuantity := resource.MustParse("0")
	memQuantity := resource.MustParse("0")

	if cpu != "" {
		parsed, err := resource.ParseQuantity(cpu)
		if err != nil {
			return cpuQuantity, memQuantity, err
		}
		cpuQuantity = parsed
	}

	if memory != "" {
		parsed, err := resource.ParseQuantity(memory)
		if err != nil {
			return cpuQuantity, memQuantity, err
		}
		memQuantity = parsed
	}

	return cpuQuantity, memQuantity, nil
}
