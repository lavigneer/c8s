# Data Model: Local Kubernetes Development Tooling with Tilt

**Status**: Design Document
**Version**: 1.0
**Date**: 2025-10-22

## Overview

This document defines the data structures, state management, and configuration models used by Tilt for C8S local development. Since Tilt handles most state management internally, this model focuses on the developer-facing configuration and the state that must be tracked within Kubernetes.

## Core Entities

### 1. LocalCluster

**Purpose**: Represents a local Kubernetes cluster instance

**Attributes**:
- `name` (string): Cluster name, e.g., "c8s-dev"
- `status` (enum): `creating`, `running`, `stopped`, `error`
- `created_timestamp` (timestamp): When cluster was created
- `last_accessed` (timestamp): When cluster was last used
- `k8s_version` (string): Kubernetes version, e.g., "v1.28.0"
- `nodes` (int): Number of worker nodes
- `total_memory_mb` (int): Total available memory
- `total_cpu_cores` (int): Total available CPU cores
- `namespace` (string): Default namespace for C8S components, default "c8s-system"

**State Transitions**:
```
[Created] → [Running] → [Stopped] → [Deleted]
    ↓
[Error] → [Running] (retry)
```

**Source of Truth**: k3d cluster metadata (queried via `k3d cluster list`, `kubectl cluster-info`)

**Validation Rules**:
- `name`: 3-63 alphanumeric characters, lowercase
- `nodes`: Minimum 1 for development
- `status`: Must be valid transition

### 2. ComponentBuild

**Purpose**: Tracks Docker image builds for C8S components

**Attributes**:
- `component_name` (string): One of "controller", "api-server", "webhook"
- `image_ref` (string): Docker image reference, e.g., "c8s-dev/c8s-controller:abc123"
- `image_digest` (string): SHA256 digest of built image
- `build_timestamp` (timestamp): When image was built
- `build_duration_seconds` (int): How long build took
- `status` (enum): `building`, `success`, `failed`
- `error_message` (string): If failed, what went wrong
- `source_files_hash` (string): Hash of source code that produced this build

**Relationships**:
- ComponentBuild → DeploymentState (latest successful build deploys)

**Source of Truth**: Docker daemon and Tilt's internal build cache

**Validation Rules**:
- `component_name`: Must match known component
- `image_digest`: Valid SHA256 hash format
- `build_duration_seconds`: Non-negative integer

### 3. DeploymentState

**Purpose**: Tracks the current deployment status of each component in the cluster

**Attributes**:
- `component_name` (string): "controller", "api-server", "webhook"
- `deployment_name` (string): Kubernetes Deployment resource name
- `pod_name` (string): Current running pod name
- `image_digest` (string): Digest of currently deployed image
- `manifest_version` (string): Hash of manifest that deployed this version
- `last_update_timestamp` (timestamp): When component was last deployed
- `status` (enum): `deploying`, `ready`, `error`, `crashing`
- `ready_replicas` (int): Number of ready pods
- `desired_replicas` (int): Desired number of pods
- `error_message` (string): If error, what caused it
- `resource_usage` (object):
  - `cpu_millicores` (int): Current CPU usage
  - `memory_mb` (int): Current memory usage
  - `cpu_limit_millicores` (int): CPU limit
  - `memory_limit_mb` (int): Memory limit

**Relationships**:
- DeploymentState → ComponentBuild (current running build)
- DeploymentState → ComponentLog (logs for this component)

**Source of Truth**: Kubernetes API (Deployment, Pod, metrics)

**Validation Rules**:
- `ready_replicas` ≤ `desired_replicas`
- `resource_usage` values non-negative
- `status` must be valid Deployment state

### 4. ComponentLog

**Purpose**: Time-series log entries from C8S component pods

**Attributes**:
- `id` (string): Unique log entry identifier
- `component_name` (string): Source component
- `pod_name` (string): Source pod
- `timestamp` (timestamp): When log entry was produced
- `log_level` (enum): `debug`, `info`, `warn`, `error`
- `message` (string): Log message
- `fields` (map<string,string>): Structured fields (e.g., request_id, operation)
- `source_file` (string): File that produced log
- `line_number` (int): Line number in source file

**Relationships**:
- ComponentLog → DeploymentState (logs from component's pod)

**Source of Truth**: Kubernetes pod logs (queried via `kubectl logs`)

**Lifecycle**:
- Created: When component produces log entry
- Stored: In Kubernetes pod logs (ephemeral or persistent depending on setup)
- Accessed: Via Tilt dashboard, `kubectl logs`, log streaming

**Validation Rules**:
- `timestamp`: Valid RFC3339 format
- `log_level`: One of defined levels
- `message`: Non-empty string

### 5. PipelineDefinition

**Purpose**: Developer-authored pipeline YAML files with validation metadata

**Attributes**:
- `file_path` (string): Path to YAML file relative to repo root
- `name` (string): Pipeline name from YAML (must match CRD spec)
- `version` (string): API version from YAML (v1alpha1)
- `yaml_content` (string): Full YAML content
- `syntax_valid` (boolean): Whether YAML syntax is valid
- `schema_valid` (boolean): Whether passes CRD schema validation
- `validation_errors` (array):
  - `field` (string): Path to invalid field (e.g., "spec.steps[0].image")
  - `error` (string): Validation error message
  - `expected_type` (string): Expected type if type mismatch
- `last_validated_timestamp` (timestamp): When last validated
- `last_validation_error` (string): Error message if validation failed

**Relationships**:
- PipelineDefinition → PipelineRun (deployed instance)

**Source of Truth**: Developer's YAML file + Tilt validation results

**Validation Rules** (CRD Schema):
- `name`: 3-63 alphanumeric lowercase with hyphens
- `version`: Must be "v1alpha1"
- `spec.steps`: Array with at least 1 step
- `spec.steps[*].image`: Non-empty Docker image reference
- `spec.steps[*].commands`: Array with at least 1 command
- `spec.steps[*].resources.cpu`: Valid CPU quantity (e.g., "100m", "1")
- `spec.steps[*].resources.memory`: Valid memory quantity (e.g., "128Mi", "1Gi")

### 6. PipelineRun

**Purpose**: Runtime instance of an executed pipeline

**Attributes**:
- `name` (string): Auto-generated run identifier
- `pipeline_name` (string): Name of pipeline that was run
- `status` (enum): `pending`, `running`, `success`, `failed`, `error`
- `created_timestamp` (timestamp): When run started
- `completed_timestamp` (timestamp): When run finished (null if running)
- `duration_seconds` (int): Time from start to completion
- `step_results` (array):
  - `step_name` (string): Name of pipeline step
  - `status` (enum): `pending`, `running`, `success`, `failed`
  - `start_time` (timestamp): When step started
  - `end_time` (timestamp): When step completed
  - `exit_code` (int): Exit code of step command
  - `logs` (string): Step output logs

**Relationships**:
- PipelineRun → PipelineDefinition (which pipeline was executed)
- PipelineRun → ComponentLog (logs from controller/webhook processing)

**Source of Truth**: Kubernetes PipelineRun CRD + Job/Pod status

**State Transitions**:
```
[Pending] → [Running] → [Success]
    ↓           ↓
  [Error]   [Failed]
```

## Configuration Models

### 7. TiltConfiguration

**Purpose**: Configuration for Tilt behavior and development environment

**Attributes**:
- `with_samples` (boolean): Deploy sample pipelines on startup, default=true
- `verbose_logs` (boolean): Enable verbose logging, default=false
- `k8s_namespace` (string): Kubernetes namespace for components, default="c8s-system"
- `image_registry` (string): Docker image registry prefix, default="c8s-dev"
- `resource_limits` (object):
  - `controller` (object): {cpu: "500m", memory: "512Mi"}
  - `api_server` (object): {cpu: "500m", memory: "512Mi"}
  - `webhook` (object): {cpu: "500m", memory: "256Mi"}
- `watch_patterns` (array): File patterns to watch for changes
- `ignore_patterns` (array): File patterns to ignore when watching

**Source**: Tilt command-line flags or local configuration

**Defaults**:
```yaml
with_samples: true
verbose_logs: false
k8s_namespace: "c8s-system"
image_registry: "c8s-dev"
watch_patterns:
  - "cmd/**/*.go"
  - "pkg/**/*.go"
  - "go.mod"
  - "go.sum"
ignore_patterns:
  - ".*"
  - "*.md"
  - "specs/"
  - "tests/"
```

## State Management

### Lifecycle Flow

```
Developer Action → File Change Detected → Build Triggered
    ↓
Build Component (Docker) → Image Created → Image Deployment
    ↓
Kubernetes Reconciliation → Pod Start → Component Running
    ↓
Logs Streamed → Developer Views Results → Iteration
```

### State Consistency

**Eventual Consistency Model**:
- File changes detected within 1-2 seconds (fsnotify)
- Build starts within 5 seconds of file change
- Image built and available within 10-30 seconds
- Pod restart within 5-10 seconds of image update
- Logs appear in dashboard within 1-2 seconds

**State Sources** (priority):
1. Kubernetes API (source of truth for cluster state)
2. Docker daemon (source of truth for image state)
3. Local filesystem (source of truth for YAML definitions)
4. Tilt daemon (caching and acceleration)

## Data Validation

### Component Build Validation

- Source files unchanged → Skip rebuild
- Only dependency files changed (go.mod, go.sum) → Rebuild one component
- Multiple files changed → Rebuild affected components

### Pipeline Definition Validation

- **Syntax**: YAML parseable
- **Schema**: Matches PipelineConfig CRD with these constraints:
  - Required fields: `version`, `name`, `steps`
  - Step constraints: image, commands required; resources optional
  - No unknown fields allowed
  - Proper type checking (strings, integers, arrays)

### Deployment Validation

- Only run deployment after successful build
- Only start pod if manifest is valid
- Verify pod readiness before marking deployment complete

## Error Handling

### Build Errors

Captured in `ComponentBuild.error_message`:
- Go syntax errors with file:line:column location
- Docker build failures with layer information
- Dependency resolution failures

### Validation Errors

Captured in `PipelineDefinition.validation_errors`:
- Each error includes field path, expected type, actual value
- Multiple errors reported in single validation run

### Deployment Errors

Captured in `DeploymentState.error_message`:
- Pod crash reasons (exit code, last message)
- Image pull failures
- Resource constraint violations

## References

- **PipelineConfig CRD**: `pkg/apis/v1alpha1/pipelineconfig_types.go`
- **Kubernetes API**: https://kubernetes.io/docs/reference/generated/kubernetes-api/
- **Tilt Architecture**: https://docs.tilt.dev/tutorial
- **k3d Documentation**: https://k3d.io/v5.8/usage/

---

**Related Documents**:
- [Tiltfile Configuration](../../Tiltfile)
- [Quick Start Guide](tilt-setup.md)
- [Feature Specification](spec.md)
- [Implementation Plan](plan.md)
