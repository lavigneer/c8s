# Data Model: Local Test Environment

**Feature**: Local Test Environment Setup
**Date**: 2025-10-13
**Purpose**: Define configuration schema for local Kubernetes clusters and C8S operator deployment

## Overview

This data model defines the configuration structures for managing local test environments. These entities are used by CLI commands to create, configure, and manage local Kubernetes clusters for C8S operator testing.

## Entities

### ClusterConfig

Represents the configuration for a local Kubernetes cluster.

**Attributes**:
- `name` (string, required): Unique identifier for the cluster (e.g., "c8s-dev", "c8s-test")
- `kubernetesVersion` (string, required): Kubernetes version to deploy (e.g., "v1.28.15")
- `nodes` (NodeConfig[], required): Node configuration (servers and agents)
- `ports` (PortMapping[], optional): Port mappings from host to cluster
- `registry` (RegistryConfig, optional): Local container registry configuration
- `volumeMounts` (VolumeMount[], optional): Host path volume mounts
- `options` (ClusterOptions, optional): Advanced cluster options

**Relationships**:
- Has many NodeConfig (1 server minimum, 0+ agents)
- Has one optional RegistryConfig
- Has many PortMapping
- Has many VolumeMount

**Validation Rules**:
- `name` must match pattern: `[a-z0-9-]+` (lowercase alphanumeric and hyphens)
- `name` must be unique (no duplicate clusters)
- `kubernetesVersion` must match available K3s versions
- Must have at least 1 server node
- Port numbers must be in range 1024-65535 (host side)

**State Transitions**:
- `not_created` → `creating` → `running` (on successful creation)
- `running` → `stopped` (on cluster stop)
- `stopped` → `starting` → `running` (on cluster start)
- `running` → `deleting` → `not_created` (on cluster delete)

**Example**:
```yaml
name: c8s-dev
kubernetesVersion: v1.28.15
nodes:
  - type: server
    count: 1
  - type: agent
    count: 2
ports:
  - hostPort: 8080
    containerPort: 80
    protocol: TCP
    nodeFilter: loadbalancer
registry:
  enabled: true
  name: registry.localhost
  hostPort: 5000
```

---

### NodeConfig

Represents node configuration within a cluster.

**Attributes**:
- `type` (string, required): Node type ("server" or "agent")
- `count` (int, required): Number of nodes of this type
- `resources` (ResourceLimits, optional): Resource constraints

**Validation Rules**:
- `type` must be either "server" or "agent"
- `count` must be >= 1 for servers
- `count` must be >= 0 for agents

**Example**:
```yaml
type: server
count: 1
resources:
  memory: "2Gi"
  cpu: "2"
```

---

### PortMapping

Represents a port mapping from host to cluster.

**Attributes**:
- `hostPort` (int, required): Port on host machine
- `containerPort` (int, required): Port inside cluster
- `protocol` (string, optional): Protocol (default: "TCP")
- `nodeFilter` (string, required): Target node(s) (e.g., "loadbalancer", "server:0", "agent:*")

**Validation Rules**:
- `hostPort` must be 1024-65535
- `containerPort` must be 1-65535
- `protocol` must be "TCP" or "UDP"
- `nodeFilter` must match pattern: `(loadbalancer|server|agent)(:\d+|\:\*)?`
- No duplicate hostPort values within same cluster

**Example**:
```yaml
hostPort: 8080
containerPort: 80
protocol: TCP
nodeFilter: loadbalancer
```

---

### RegistryConfig

Represents local container registry configuration.

**Attributes**:
- `enabled` (bool, required): Whether to create local registry
- `name` (string, required): Registry hostname
- `hostPort` (int, required): Port for registry on host
- `proxyRemote` (string, optional): Remote registry to proxy (e.g., "https://registry-1.docker.io")

**Validation Rules**:
- `name` must be valid hostname
- `hostPort` must be 1024-65535
- `hostPort` must not conflict with other PortMapping entries
- If `proxyRemote` is set, must be valid URL

**Example**:
```yaml
enabled: true
name: registry.localhost
hostPort: 5000
```

---

### VolumeMount

Represents a volume mount from host to cluster nodes.

**Attributes**:
- `hostPath` (string, required): Absolute path on host machine
- `containerPath` (string, required): Path inside container
- `nodeFilter` (string, required): Target node(s) (e.g., "all", "server:*", "agent:0")

**Validation Rules**:
- `hostPath` must be absolute path
- `containerPath` must be absolute path
- `nodeFilter` must match pattern: `(all|server|agent)(:\d+|\:\*)?`

**Example**:
```yaml
hostPath: /tmp/c8s-data
containerPath: /data
nodeFilter: agent:*
```

---

### ClusterOptions

Advanced cluster configuration options.

**Attributes**:
- `waitTimeout` (duration, optional): Max time to wait for cluster ready (default: "60s")
- `updateDefaultKubeconfig` (bool, optional): Update ~/.kube/config (default: true)
- `switchContext` (bool, optional): Switch to new cluster context (default: true)
- `disableLoadBalancer` (bool, optional): Disable integrated load balancer (default: false)
- `k3sArgs` (string[], optional): Additional arguments to pass to K3s

**Validation Rules**:
- `waitTimeout` must be valid duration string (e.g., "60s", "2m")
- `k3sArgs` must be valid K3s flags

**Example**:
```yaml
waitTimeout: 60s
updateDefaultKubeconfig: true
switchContext: true
k3sArgs:
  - "--disable=traefik"
```

---

### EnvironmentConfig

Represents a complete local test environment (cluster + operator deployment).

**Attributes**:
- `cluster` (ClusterConfig, required): Cluster configuration
- `operator` (OperatorDeployment, required): Operator deployment configuration
- `samples` (SampleConfig[], optional): Sample PipelineConfigs to deploy

**Relationships**:
- Has one ClusterConfig
- Has one OperatorDeployment
- Has many SampleConfig

**Example**:
```yaml
cluster:
  name: c8s-dev
  kubernetesVersion: v1.28.15
  nodes:
    - type: server
      count: 1
operator:
  image: c8s-operator:dev
  imagePullPolicy: IfNotPresent
  crdsPath: config/crd/bases
samples:
  - name: simple-build
    path: config/samples/simple-build.yaml
```

---

### OperatorDeployment

Represents C8S operator deployment configuration.

**Attributes**:
- `image` (string, required): Operator container image
- `imagePullPolicy` (string, optional): Pull policy (default: "IfNotPresent")
- `crdsPath` (string, required): Path to CRD manifests
- `manifestsPath` (string, required): Path to operator deployment manifests
- `namespace` (string, optional): Kubernetes namespace (default: "c8s-system")
- `replicas` (int, optional): Number of operator replicas (default: 1)

**Validation Rules**:
- `image` must be valid container image reference
- `imagePullPolicy` must be "Always", "IfNotPresent", or "Never"
- `crdsPath` must be valid directory path
- `manifestsPath` must be valid directory path
- `namespace` must match Kubernetes namespace pattern
- `replicas` must be >= 1

**Example**:
```yaml
image: c8s-operator:dev
imagePullPolicy: IfNotPresent
crdsPath: config/crd/bases
manifestsPath: config/manager
namespace: c8s-system
replicas: 1
```

---

### SampleConfig

Represents a sample PipelineConfig for testing.

**Attributes**:
- `name` (string, required): Sample identifier
- `path` (string, required): Path to YAML manifest
- `description` (string, optional): Human-readable description

**Validation Rules**:
- `name` must be unique within EnvironmentConfig
- `path` must point to existing YAML file
- YAML file must contain valid PipelineConfig

**Example**:
```yaml
name: simple-build
path: config/samples/simple-build.yaml
description: Basic single-step build pipeline
```

---

### ClusterStatus

Runtime status information for a cluster.

**Attributes**:
- `name` (string, required): Cluster name
- `state` (string, required): Current state ("running", "stopped", "not_found")
- `nodes` (NodeStatus[], required): Status of each node
- `kubeconfig` (string, required): Path to kubeconfig file
- `apiEndpoint` (string, required): Kubernetes API server endpoint
- `registryEndpoint` (string, optional): Local registry endpoint (if enabled)
- `createdAt` (timestamp, optional): Cluster creation time
- `uptime` (duration, optional): Time since cluster started

**Validation Rules**:
- `state` must be "running", "stopped", or "not_found"
- `apiEndpoint` must be valid URL

**Example**:
```json
{
  "name": "c8s-dev",
  "state": "running",
  "nodes": [
    {"name": "k3d-c8s-dev-server-0", "role": "server", "status": "Ready"},
    {"name": "k3d-c8s-dev-agent-0", "role": "agent", "status": "Ready"},
    {"name": "k3d-c8s-dev-agent-1", "role": "agent", "status": "Ready"}
  ],
  "kubeconfig": "/Users/dev/.kube/config",
  "apiEndpoint": "https://0.0.0.0:6443",
  "registryEndpoint": "localhost:5000",
  "createdAt": "2025-10-13T10:30:00Z",
  "uptime": "2h15m"
}
```

---

### NodeStatus

Runtime status for a single node.

**Attributes**:
- `name` (string, required): Node name
- `role` (string, required): Node role ("server" or "agent")
- `status` (string, required): Node status ("Ready", "NotReady", "Unknown")
- `version` (string, optional): Kubernetes version

**Example**:
```json
{
  "name": "k3d-c8s-dev-server-0",
  "role": "server",
  "status": "Ready",
  "version": "v1.28.15+k3s1"
}
```

---

## Configuration File Format

### Default Cluster Configuration

Location: `~/.c8s/cluster-defaults.yaml`

```yaml
apiVersion: c8s.io/v1alpha1
kind: ClusterDefaults
metadata:
  name: default
spec:
  kubernetesVersion: v1.28.15
  nodes:
    - type: server
      count: 1
    - type: agent
      count: 2
  ports:
    - hostPort: 8080
      containerPort: 80
      protocol: TCP
      nodeFilter: loadbalancer
  registry:
    enabled: true
    name: registry.localhost
    hostPort: 5000
  options:
    waitTimeout: 60s
    updateDefaultKubeconfig: true
    switchContext: true
    k3sArgs:
      - "--disable=traefik"
```

### Environment Configuration

Location: `.c8s/environment.yaml` (in project root)

```yaml
apiVersion: c8s.io/v1alpha1
kind: Environment
metadata:
  name: c8s-dev
spec:
  cluster:
    name: c8s-dev
    kubernetesVersion: v1.28.15
    nodes:
      - type: server
        count: 1
      - type: agent
        count: 2
    registry:
      enabled: true
      hostPort: 5000
  operator:
    image: c8s-operator:dev
    imagePullPolicy: IfNotPresent
    crdsPath: config/crd/bases
    manifestsPath: config/manager
    namespace: c8s-system
  samples:
    - name: simple-build
      path: config/samples/simple-build.yaml
      description: Single-step build pipeline
    - name: multi-step
      path: config/samples/multi-step.yaml
      description: Multi-step pipeline with dependencies
    - name: matrix-build
      path: config/samples/matrix-build.yaml
      description: Matrix build with parallel execution
```

## Storage

### Persistent State

Cluster state is managed by k3d/Docker:
- **Docker containers**: Represent cluster nodes
- **Docker volumes**: Store cluster data
- **Kubeconfig**: Stored in `~/.kube/config`

### C8S CLI State

Location: `~/.c8s/`

```
~/.c8s/
├── cluster-defaults.yaml    # Default cluster configuration
├── environments/            # Named environment configurations
│   ├── dev.yaml
│   └── test.yaml
└── state/                   # Runtime state tracking
    └── clusters.json        # Cluster status cache
```

### State File Format

`~/.c8s/state/clusters.json`:

```json
{
  "clusters": [
    {
      "name": "c8s-dev",
      "state": "running",
      "createdAt": "2025-10-13T10:30:00Z",
      "configHash": "abc123...",
      "k3dVersion": "v5.8.0"
    }
  ],
  "lastUpdated": "2025-10-13T12:45:00Z"
}
```

## API Compatibility

All configuration structures support serialization to:
- **YAML**: Primary format for user-facing configs
- **JSON**: Used for CLI state files and API responses
- **Go structs**: Internal representation using standard Go types + K8s meta types

## Validation

All entities are validated using:
1. **Schema validation**: JSON Schema or Go struct tags
2. **Business logic validation**: Custom validators for cross-field rules
3. **Kubernetes validation**: Leverage K8s validation libraries for names, labels, etc.

## Migration Strategy

If schema changes in future versions:
1. Use API versioning (`v1alpha1` → `v1alpha2`)
2. Provide conversion webhooks or CLI migration commands
3. Support reading older formats with automatic upgrade
4. Log warnings for deprecated fields
