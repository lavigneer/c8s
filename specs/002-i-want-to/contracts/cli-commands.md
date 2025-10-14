# CLI Command Contracts

**Feature**: Local Test Environment Setup
**Date**: 2025-10-13
**Purpose**: Define command-line interface contracts for local environment management

## Overview

This document specifies the CLI commands for managing local Kubernetes test environments. All commands are subcommands under `c8s dev` and follow standard CLI conventions.

## Command Hierarchy

```
c8s dev
├── cluster
│   ├── create
│   ├── delete
│   ├── start
│   ├── stop
│   ├── status
│   └── list
├── deploy
│   ├── operator
│   └── samples
└── test
    ├── run
    └── logs
```

---

## Cluster Management Commands

### `c8s dev cluster create`

Create a new local Kubernetes cluster.

**Synopsis**:
```bash
c8s dev cluster create [NAME] [flags]
```

**Arguments**:
- `NAME` (optional): Cluster name (default: "c8s-dev")

**Flags**:
- `--config, -c` (string): Path to cluster config file
- `--k8s-version` (string): Kubernetes version (default: "v1.28.15")
- `--servers` (int): Number of server nodes (default: 1)
- `--agents` (int): Number of agent nodes (default: 2)
- `--registry` (bool): Enable local registry (default: true)
- `--registry-port` (int): Registry host port (default: 5000)
- `--timeout` (duration): Creation timeout (default: "3m")
- `--wait` (bool): Wait for cluster to be ready (default: true)

**Examples**:
```bash
# Create cluster with defaults
c8s dev cluster create

# Create cluster with custom name
c8s dev cluster create my-test-cluster

# Create from config file
c8s dev cluster create --config k3d-dev-cluster.yaml

# Create with custom configuration
c8s dev cluster create test --k8s-version v1.28.15 --agents 3

# Create without registry
c8s dev cluster create --registry=false
```

**Exit Codes**:
- `0`: Success
- `1`: General error (invalid flags, cluster creation failed)
- `2`: Cluster already exists
- `3`: Timeout waiting for cluster ready
- `4`: Docker not available

**Output** (success):
```
Creating cluster 'c8s-dev'...
✓ Cluster 'c8s-dev' created successfully
✓ Kubeconfig updated: ~/.kube/config
✓ Current context: k3d-c8s-dev
✓ Registry available at: localhost:5000

Cluster Status:
  Name:       c8s-dev
  Nodes:      3 (1 server, 2 agents)
  K8s Version: v1.28.15+k3s1
  API Server: https://0.0.0.0:6443

Next steps:
  Deploy operator: c8s dev deploy operator
  Check status:    c8s dev cluster status
```

**Output** (error - already exists):
```
Error: cluster 'c8s-dev' already exists
Run 'c8s dev cluster delete c8s-dev' to remove it first
```

**Contract Tests**:
- ✅ Creates cluster with default name when no arguments provided
- ✅ Creates cluster with custom name when NAME argument provided
- ✅ Respects --config flag and loads configuration from file
- ✅ Returns exit code 2 if cluster already exists
- ✅ Returns exit code 4 if Docker is not running
- ✅ Updates kubeconfig and switches context when successful
- ✅ Creates local registry when --registry=true
- ✅ Times out after --timeout duration if cluster not ready

---

### `c8s dev cluster delete`

Delete a local Kubernetes cluster.

**Synopsis**:
```bash
c8s dev cluster delete [NAME] [flags]
```

**Arguments**:
- `NAME` (optional): Cluster name (default: "c8s-dev")

**Flags**:
- `--force, -f` (bool): Force deletion without confirmation
- `--all` (bool): Delete all c8s clusters

**Examples**:
```bash
# Delete default cluster (with confirmation)
c8s dev cluster delete

# Delete specific cluster
c8s dev cluster delete my-test-cluster

# Force deletion without prompt
c8s dev cluster delete --force

# Delete all clusters
c8s dev cluster delete --all
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found
- `130`: User cancelled (SIGINT during confirmation)

**Output** (success with confirmation):
```
Warning: This will delete cluster 'c8s-dev' and all its data
Are you sure? (yes/no): yes

Deleting cluster 'c8s-dev'...
✓ Cluster 'c8s-dev' deleted successfully
✓ Kubeconfig context removed
```

**Output** (cluster not found):
```
Error: cluster 'c8s-dev' not found
Run 'c8s dev cluster list' to see available clusters
```

**Contract Tests**:
- ✅ Prompts for confirmation before deleting
- ✅ Skips confirmation with --force flag
- ✅ Returns exit code 2 if cluster not found
- ✅ Removes kubeconfig context after deletion
- ✅ Deletes all c8s clusters with --all flag
- ✅ Cleans up Docker containers and volumes

---

### `c8s dev cluster status`

Show status of a local cluster.

**Synopsis**:
```bash
c8s dev cluster status [NAME] [flags]
```

**Arguments**:
- `NAME` (optional): Cluster name (default: current context)

**Flags**:
- `--output, -o` (string): Output format (text|json|yaml) (default: "text")
- `--watch, -w` (bool): Watch for status changes

**Examples**:
```bash
# Show status of current cluster
c8s dev cluster status

# Show status of specific cluster
c8s dev cluster status my-test-cluster

# Output as JSON
c8s dev cluster status --output json

# Watch status updates
c8s dev cluster status --watch
```

**Exit Codes**:
- `0`: Success (cluster running)
- `1`: General error
- `2`: Cluster not found
- `3`: Cluster not ready

**Output** (text format):
```
Cluster Status: c8s-dev

State:     Running
Uptime:    2h 15m
API:       https://0.0.0.0:6443
Registry:  localhost:5000

Nodes:
  NAME                   ROLE     STATUS   VERSION
  k3d-c8s-dev-server-0   server   Ready    v1.28.15+k3s1
  k3d-c8s-dev-agent-0    agent    Ready    v1.28.15+k3s1
  k3d-c8s-dev-agent-1    agent    Ready    v1.28.15+k3s1

Resources:
  Memory:    ~500 MB
  Disk:      ~6 GB
```

**Output** (JSON format):
```json
{
  "name": "c8s-dev",
  "state": "running",
  "uptime": "2h15m",
  "apiEndpoint": "https://0.0.0.0:6443",
  "registryEndpoint": "localhost:5000",
  "nodes": [
    {
      "name": "k3d-c8s-dev-server-0",
      "role": "server",
      "status": "Ready",
      "version": "v1.28.15+k3s1"
    }
  ],
  "resources": {
    "memory": "500MB",
    "disk": "6GB"
  }
}
```

**Contract Tests**:
- ✅ Shows status of cluster from current kubeconfig context when no NAME provided
- ✅ Shows status of specified cluster when NAME provided
- ✅ Returns exit code 2 if cluster not found
- ✅ Outputs valid JSON when --output=json
- ✅ Outputs valid YAML when --output=yaml
- ✅ Watches for status changes with --watch flag

---

### `c8s dev cluster list`

List all local clusters.

**Synopsis**:
```bash
c8s dev cluster list [flags]
```

**Flags**:
- `--output, -o` (string): Output format (text|json|yaml) (default: "text")
- `--all` (bool): Show all k3d clusters (not just c8s clusters)

**Examples**:
```bash
# List c8s clusters
c8s dev cluster list

# List all k3d clusters
c8s dev cluster list --all

# Output as JSON
c8s dev cluster list --output json
```

**Exit Codes**:
- `0`: Success

**Output** (text format):
```
NAME           STATE     NODES   VERSION        UPTIME
c8s-dev        Running   3       v1.28.15       2h 15m
c8s-test       Stopped   3       v1.28.15       -
```

**Output** (JSON format):
```json
{
  "clusters": [
    {
      "name": "c8s-dev",
      "state": "running",
      "nodeCount": 3,
      "version": "v1.28.15+k3s1",
      "uptime": "2h15m"
    },
    {
      "name": "c8s-test",
      "state": "stopped",
      "nodeCount": 3,
      "version": "v1.28.15+k3s1",
      "uptime": null
    }
  ]
}
```

**Contract Tests**:
- ✅ Lists only c8s clusters by default
- ✅ Lists all k3d clusters with --all flag
- ✅ Outputs valid JSON when --output=json
- ✅ Returns empty list when no clusters exist

---

### `c8s dev cluster start`

Start a stopped cluster.

**Synopsis**:
```bash
c8s dev cluster start [NAME] [flags]
```

**Arguments**:
- `NAME` (optional): Cluster name (default: "c8s-dev")

**Flags**:
- `--wait` (bool): Wait for cluster to be ready (default: true)
- `--timeout` (duration): Start timeout (default: "2m")

**Examples**:
```bash
# Start default cluster
c8s dev cluster start

# Start specific cluster
c8s dev cluster start my-test-cluster
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found
- `3`: Timeout waiting for ready

**Output**:
```
Starting cluster 'c8s-dev'...
✓ Cluster 'c8s-dev' started successfully
```

**Contract Tests**:
- ✅ Starts stopped cluster
- ✅ Returns exit code 2 if cluster not found
- ✅ Waits for cluster ready when --wait=true
- ✅ Times out after --timeout duration

---

### `c8s dev cluster stop`

Stop a running cluster (preserves state).

**Synopsis**:
```bash
c8s dev cluster stop [NAME] [flags]
```

**Arguments**:
- `NAME` (optional): Cluster name (default: "c8s-dev")

**Examples**:
```bash
# Stop default cluster
c8s dev cluster stop

# Stop specific cluster
c8s dev cluster stop my-test-cluster
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found

**Output**:
```
Stopping cluster 'c8s-dev'...
✓ Cluster 'c8s-dev' stopped successfully
```

**Contract Tests**:
- ✅ Stops running cluster
- ✅ Preserves cluster state (can be restarted)
- ✅ Returns exit code 2 if cluster not found

---

## Deployment Commands

### `c8s dev deploy operator`

Deploy C8S operator to the cluster.

**Synopsis**:
```bash
c8s dev deploy operator [flags]
```

**Flags**:
- `--cluster` (string): Target cluster name (default: current context)
- `--image` (string): Operator image (default: "c8s-operator:dev")
- `--image-pull-policy` (string): Image pull policy (default: "IfNotPresent")
- `--namespace` (string): Kubernetes namespace (default: "c8s-system")
- `--crds-path` (string): Path to CRD manifests (default: "config/crd/bases")
- `--manifests-path` (string): Path to operator manifests (default: "config/manager")
- `--wait` (bool): Wait for operator to be ready (default: true)
- `--timeout` (duration): Deployment timeout (default: "5m")

**Examples**:
```bash
# Deploy with defaults
c8s dev deploy operator

# Deploy custom image
c8s dev deploy operator --image c8s-operator:v1.0.0

# Deploy to specific cluster
c8s dev deploy operator --cluster my-test-cluster

# Deploy without waiting
c8s dev deploy operator --wait=false
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found
- `3`: Timeout waiting for ready
- `4`: CRD installation failed
- `5`: Operator deployment failed

**Output**:
```
Deploying C8S operator to cluster 'c8s-dev'...

✓ Installing CRDs...
  - pipelineconfigs.c8s.io
  - pipelineruns.c8s.io

✓ Loading operator image to cluster...

✓ Deploying operator to namespace 'c8s-system'...
  - ServiceAccount: c8s-controller-manager
  - Role: c8s-leader-election-role
  - RoleBinding: c8s-leader-election-rolebinding
  - Deployment: c8s-controller-manager

✓ Waiting for operator to be ready...

Operator Status:
  Namespace:  c8s-system
  Replicas:   1/1 ready
  Image:      c8s-operator:dev
  Status:     Running

Next steps:
  Deploy samples: c8s dev deploy samples
  View logs:      c8s dev test logs
```

**Contract Tests**:
- ✅ Installs CRDs before deploying operator
- ✅ Loads operator image into cluster
- ✅ Creates namespace if it doesn't exist
- ✅ Deploys operator resources (ServiceAccount, RBAC, Deployment)
- ✅ Waits for operator pod to be ready when --wait=true
- ✅ Returns exit code 4 if CRD installation fails
- ✅ Returns exit code 5 if operator deployment fails
- ✅ Uses current kubeconfig context if --cluster not specified

---

### `c8s dev deploy samples`

Deploy sample PipelineConfigs to the cluster.

**Synopsis**:
```bash
c8s dev deploy samples [flags]
```

**Flags**:
- `--cluster` (string): Target cluster name (default: current context)
- `--samples-path` (string): Path to samples directory (default: "config/samples")
- `--namespace` (string): Kubernetes namespace (default: "default")
- `--select` (string): Select specific samples (comma-separated)

**Examples**:
```bash
# Deploy all samples
c8s dev deploy samples

# Deploy specific samples
c8s dev deploy samples --select simple-build,multi-step

# Deploy to specific namespace
c8s dev deploy samples --namespace test-pipelines
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found
- `3`: Samples path not found
- `4`: Invalid sample manifests

**Output**:
```
Deploying sample PipelineConfigs to cluster 'c8s-dev'...

✓ Applying samples:
  - simple-build.yaml
  - multi-step.yaml
  - matrix-build.yaml

✓ 3 PipelineConfigs deployed successfully

Next steps:
  Monitor pipelines: kubectl get pipelineconfigs -n default
  View logs:         c8s dev test logs
```

**Contract Tests**:
- ✅ Deploys all YAML files in samples directory
- ✅ Filters samples with --select flag
- ✅ Returns exit code 3 if samples path doesn't exist
- ✅ Returns exit code 4 if YAML manifests are invalid
- ✅ Creates namespace if it doesn't exist

---

## Testing Commands

### `c8s dev test run`

Run end-to-end pipeline tests.

**Synopsis**:
```bash
c8s dev test run [flags]
```

**Flags**:
- `--cluster` (string): Target cluster name (default: current context)
- `--pipeline` (string): Specific pipeline to test (default: all)
- `--namespace` (string): Kubernetes namespace (default: "default")
- `--timeout` (duration): Test timeout (default: "10m")
- `--watch` (bool): Watch test progress in real-time

**Examples**:
```bash
# Run all pipeline tests
c8s dev test run

# Test specific pipeline
c8s dev test run --pipeline simple-build

# Watch test execution
c8s dev test run --watch
```

**Exit Codes**:
- `0`: All tests passed
- `1`: General error
- `2`: Cluster not found
- `3`: Operator not deployed
- `4`: One or more tests failed
- `5`: Test timeout

**Output**:
```
Running pipeline tests in cluster 'c8s-dev'...

Testing simple-build...
  ✓ Pipeline created
  ✓ Job started
  ✓ Step 1 completed (5s)
  ✓ Pipeline succeeded (8s)

Testing multi-step...
  ✓ Pipeline created
  ✓ Job started
  ✓ Step 1 completed (3s)
  ✓ Step 2 completed (4s)
  ✓ Pipeline succeeded (10s)

Testing matrix-build...
  ✓ Pipeline created
  ✓ 4 parallel jobs started
  ✓ All matrix jobs completed (12s)
  ✓ Pipeline succeeded (15s)

Test Results:
  Total:   3
  Passed:  3
  Failed:  0
  Duration: 33s

All tests passed ✓
```

**Output** (with failures):
```
Testing simple-build...
  ✓ Pipeline created
  ✓ Job started
  ✗ Step 1 failed: container image not found
  ✗ Pipeline failed (5s)

Test Results:
  Total:   3
  Passed:  2
  Failed:  1
  Duration: 28s

Tests failed ✗

Failed Tests:
  - simple-build: container image not found

View logs: c8s dev test logs --pipeline simple-build
```

**Contract Tests**:
- ✅ Tests all PipelineConfigs in namespace
- ✅ Filters by --pipeline flag when specified
- ✅ Returns exit code 3 if operator not deployed
- ✅ Returns exit code 4 if any tests fail
- ✅ Returns exit code 5 on timeout
- ✅ Shows real-time progress with --watch flag

---

### `c8s dev test logs`

View logs from pipeline executions.

**Synopsis**:
```bash
c8s dev test logs [flags]
```

**Flags**:
- `--cluster` (string): Target cluster name (default: current context)
- `--pipeline` (string): Pipeline name (required if multiple exist)
- `--namespace` (string): Kubernetes namespace (default: "default")
- `--follow, -f` (bool): Follow logs in real-time
- `--tail` (int): Number of lines to show from end (default: 100)

**Examples**:
```bash
# View logs for all pipelines
c8s dev test logs

# View logs for specific pipeline
c8s dev test logs --pipeline simple-build

# Follow logs in real-time
c8s dev test logs --pipeline simple-build --follow

# Show last 50 lines
c8s dev test logs --pipeline simple-build --tail 50
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Cluster not found
- `3`: Pipeline not found

**Output**:
```
Showing logs for pipeline 'simple-build' in cluster 'c8s-dev'...

==> Step: checkout
2025-10-13T10:30:00Z Cloning repository...
2025-10-13T10:30:02Z Repository cloned successfully

==> Step: build
2025-10-13T10:30:05Z Building application...
2025-10-13T10:30:10Z Build completed successfully

==> Pipeline Status: Succeeded
```

**Contract Tests**:
- ✅ Shows logs from all pipeline steps
- ✅ Filters by --pipeline flag when specified
- ✅ Follows logs in real-time with --follow flag
- ✅ Returns exit code 3 if pipeline not found
- ✅ Tails logs according to --tail value

---

## Global Flags

All commands inherit these global flags:

- `--help, -h`: Show help for command
- `--verbose, -v`: Enable verbose output
- `--quiet, -q`: Suppress non-error output
- `--no-color`: Disable colored output

## Environment Variables

- `C8S_DEV_CLUSTER`: Default cluster name (overrides "c8s-dev")
- `C8S_DEV_CONFIG`: Default cluster config file path
- `KUBECONFIG`: Kubernetes config file (standard K8s variable)
- `DOCKER_HOST`: Docker daemon socket (standard Docker variable)

## Error Handling

All commands follow these conventions:

1. **Exit codes**: Use meaningful exit codes (documented per command)
2. **Error messages**: Clear, actionable error messages with suggestions
3. **Validation**: Validate inputs before executing operations
4. **Idempotency**: Commands should be safe to retry (where possible)
5. **Cleanup**: Properly clean up resources on failure

## Testing Strategy

### Contract Tests

Each command must have contract tests that verify:

1. **Command parsing**: Flags and arguments parsed correctly
2. **Exit codes**: Correct exit code for each scenario
3. **Output format**: Output matches documented format
4. **Error handling**: Proper error messages and cleanup
5. **Integration**: Correct interaction with k3d/kubectl

### Test Organization

```
tests/
└── contract/
    ├── dev_commands_test.go         # Main test suite
    ├── cluster_commands_test.go     # Cluster command tests
    ├── deploy_commands_test.go      # Deploy command tests
    └── test_commands_test.go        # Test command tests
```

### Example Contract Test

```go
func TestClusterCreateCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        exitCode int
        validate func(output string) error
    }{
        {
            name:     "create with default name",
            args:     []string{"dev", "cluster", "create"},
            exitCode: 0,
            validate: func(output string) error {
                if !strings.Contains(output, "c8s-dev") {
                    return fmt.Errorf("expected cluster name 'c8s-dev' in output")
                }
                return nil
            },
        },
        {
            name:     "create with custom name",
            args:     []string{"dev", "cluster", "create", "test-cluster"},
            exitCode: 0,
            validate: func(output string) error {
                if !strings.Contains(output, "test-cluster") {
                    return fmt.Errorf("expected cluster name 'test-cluster' in output")
                }
                return nil
            },
        },
        {
            name:     "cluster already exists",
            args:     []string{"dev", "cluster", "create", "existing"},
            exitCode: 2,
            validate: func(output string) error {
                if !strings.Contains(output, "already exists") {
                    return fmt.Errorf("expected 'already exists' error message")
                }
                return nil
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute command
            output, exitCode := executeCommand(tt.args)

            // Verify exit code
            if exitCode != tt.exitCode {
                t.Errorf("expected exit code %d, got %d", tt.exitCode, exitCode)
            }

            // Validate output
            if err := tt.validate(output); err != nil {
                t.Error(err)
            }
        })
    }
}
```
