package localenv

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// clusterNamePattern matches lowercase alphanumeric and hyphens
	clusterNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

	// k8sVersionPattern matches Kubernetes version format (v1.28.15, v1.29.0, etc.)
	k8sVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+.*$`)

	// nodeFilterPattern matches node filter format (loadbalancer, server:0, agent:*, etc.)
	nodeFilterPattern = regexp.MustCompile(`^(loadbalancer|server|agent)(:\d+|:\*)?$`)

	// k8sNamespacePattern matches valid Kubernetes namespace names
	k8sNamespacePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// Validator wraps go-playground/validator with custom validation rules
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator with custom rules registered
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation functions
	_ = v.RegisterValidation("cluster_name", validateClusterName)
	_ = v.RegisterValidation("k8s_version", validateK8sVersion)
	_ = v.RegisterValidation("node_filter", validateNodeFilter)
	_ = v.RegisterValidation("k8s_namespace", validateK8sNamespace)
	_ = v.RegisterValidation("absolute_path", validateAbsolutePath)
	_ = v.RegisterValidation("duration", validateDuration)

	return &Validator{validate: v}
}

// ValidateClusterConfig validates a ClusterConfig and returns detailed errors
func (v *Validator) ValidateClusterConfig(config *ClusterConfig) error {
	if err := v.validate.Struct(config); err != nil {
		return v.formatValidationErrors(err)
	}

	// Additional cross-field validations
	if err := v.validateNodes(config.Nodes); err != nil {
		return err
	}

	if err := v.validatePorts(config.Ports); err != nil {
		return err
	}

	return nil
}

// ValidateEnvironmentConfig validates an EnvironmentConfig
func (v *Validator) ValidateEnvironmentConfig(config *EnvironmentConfig) error {
	if err := v.validate.Struct(config); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// validateNodes ensures at least one server node exists
func (v *Validator) validateNodes(nodes []NodeConfig) error {
	serverCount := 0
	for _, node := range nodes {
		if node.Type == "server" {
			serverCount += node.Count
		}
	}

	if serverCount < 1 {
		return fmt.Errorf("cluster must have at least 1 server node")
	}

	return nil
}

// validatePorts checks for duplicate host ports
func (v *Validator) validatePorts(ports []PortMapping) error {
	seen := make(map[int]bool)
	for _, port := range ports {
		if seen[port.HostPort] {
			return fmt.Errorf("duplicate host port: %d", port.HostPort)
		}
		seen[port.HostPort] = true
	}
	return nil
}

// formatValidationErrors converts validator errors to user-friendly messages
func (v *Validator) formatValidationErrors(err error) error {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var messages []string
	for _, e := range validationErrs {
		messages = append(messages, formatFieldError(e))
	}

	return fmt.Errorf("validation failed:\n  - %s", strings.Join(messages, "\n  - "))
}

// formatFieldError formats a single field validation error
func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "cluster_name":
		return fmt.Sprintf("%s must contain only lowercase letters, numbers, and hyphens", field)
	case "k8s_version":
		return fmt.Sprintf("%s must be a valid Kubernetes version (e.g., v1.28.15)", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "node_filter":
		return fmt.Sprintf("%s must match pattern: (loadbalancer|server|agent)(:<number>|:*)", field)
	case "k8s_namespace":
		return fmt.Sprintf("%s must be a valid Kubernetes namespace name", field)
	case "absolute_path":
		return fmt.Sprintf("%s must be an absolute path", field)
	case "duration":
		return fmt.Sprintf("%s must be a valid duration (e.g., 60s, 5m)", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "hostname":
		return fmt.Sprintf("%s must be a valid hostname", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// Custom validation functions

func validateClusterName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return false
	}
	return clusterNamePattern.MatchString(name)
}

func validateK8sVersion(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	if version == "" {
		return false
	}
	return k8sVersionPattern.MatchString(version)
}

func validateNodeFilter(fl validator.FieldLevel) bool {
	filter := fl.Field().String()
	if filter == "" {
		return false
	}
	return nodeFilterPattern.MatchString(filter)
}

func validateK8sNamespace(fl validator.FieldLevel) bool {
	namespace := fl.Field().String()
	if namespace == "" {
		return false
	}
	if len(namespace) > 63 {
		return false
	}
	return k8sNamespacePattern.MatchString(namespace)
}

func validateAbsolutePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return false
	}
	// Simple check: absolute paths start with /
	return strings.HasPrefix(path, "/")
}

func validateDuration(fl validator.FieldLevel) bool {
	duration := fl.Field().String()
	if duration == "" {
		return true // Empty is valid (will use defaults)
	}
	_, err := ParseDuration(duration)
	return err == nil
}

// Package-level validation functions for convenience

var defaultValidator = NewValidator()

// ValidateClusterConfig is a convenience function that validates using the default validator
func ValidateClusterConfig(config *ClusterConfig) error {
	return defaultValidator.ValidateClusterConfig(config)
}

// ValidateEnvironmentConfig is a convenience function that validates using the default validator
func ValidateEnvironmentConfig(config *EnvironmentConfig) error {
	return defaultValidator.ValidateEnvironmentConfig(config)
}
