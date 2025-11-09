# Data Model: C8S Deployment Configuration & Status

**Feature**: Deploy C8S Stack to Kubernetes (008-create-a-simple)
**Date**: 2025-11-09

---

## Core Entities

### 1. DeploymentConfig

Represents the user's desired deployment configuration for C8S stack.

**Fields**:
- `metadata` (object)
  - `name`: string - Human-readable deployment name (e.g., "production-c8s")
  - `version`: string - C8S version to deploy (e.g., "v0.1.0", "latest")
  - `createdAt`: timestamp - When configuration was created
  - `updatedAt`: timestamp - Last update timestamp

- `cluster` (object) - Kubernetes cluster target
  - `namespace`: string [1-63 chars, RFC 1123] - Kubernetes namespace (default: "c8s-system")
  - `kubeconfig`: string - Path to kubeconfig (default: ~/.kube/config)
  - `context`: string - Kubernetes context name (default: current context)

- `components` (object) - C8S component specifications
  - `controller` (ComponentConfig)
  - `webhook` (ComponentConfig)
  - `apiServer` (ComponentConfig)
  - `database` (ComponentConfig) [optional]

- `storage` (object) - Storage configuration
  - `type`: enum ["s3-compatible", "pvc", "local"] (default: "s3-compatible")
  - `s3Config` (S3StorageConfig) [if type="s3-compatible"]
  - `pvcConfig` (PVCStorageConfig) [if type="pvc"]

- `environment` (object) - Environment settings
  - `type`: enum ["dev", "staging", "production"] (default: "dev")
  - `logLevel`: enum ["debug", "info", "warn", "error"] (default: "info")
  - `replicas`: object - Override default replicas per environment
    - `controller`: int [1-5] (default: 1 for dev, 3 for prod)
    - `webhook`: int [1-5] (default: 1 for dev, 2 for prod)
    - `apiServer`: int [1-5] (default: 1 for dev, 2 for prod)

- `resources` (object) - Kubernetes resource limits/requests
  - `requests` (ResourceLimit)
  - `limits` (ResourceLimit)

- `authentication` (object)
  - `enabled`: boolean (default: true)
  - `type`: enum ["basic", "oauth2", "saml"] (default: "basic")
  - `secretRef`: string - Reference to Kubernetes secret containing auth credentials

- `tls` (object) [optional]
  - `enabled`: boolean (default: false)
  - `certPath`: string - Path to TLS certificate
  - `keyPath`: string - Path to TLS key
  - `secretRef`: string - Reference to Kubernetes secret containing cert/key

**Validation Rules**:
- `metadata.name` must be unique within cluster and namespace
- `metadata.version` must match semantic versioning (v\d+\.\d+\.\d+)
- `cluster.namespace` must be valid RFC 1123 subdomain (lowercase, hyphens, 1-63 chars)
- Component replicas must be between 1 and 5
- `environment.type` determines default resource limits and replica counts
- All component replicas must be specified if any customization is provided

---

### 2. ComponentConfig

Configuration for individual C8S components.

**Fields**:
- `enabled`: boolean (default: true) - Whether to deploy this component
- `image` (object)
  - `registry`: string - Container image registry (default: "docker.io")
  - `repository`: string - Image repository (e.g., "org/c8s-controller")
  - `tag`: string - Image tag/version (default: "latest")
  - `pullPolicy`: enum ["Always", "IfNotPresent", "Never"] (default: "IfNotPresent")

- `replicas`: int [1-5] - Number of pod replicas (overrides environment default)

- `port` (object)
  - `containerPort`: int [1-65535] - Port container listens on
  - `servicePort`: int [1-65535] - Port exposed by Kubernetes Service

- `resources` (object) [optional]
  - `requests` (ResourceLimit) - Min resources guaranteed
  - `limits` (ResourceLimit) - Max resources allowed

- `env` (array of object) [optional] - Additional environment variables
  - `name`: string - Variable name
  - `value`: string - Variable value
  - `valueFrom`: object [optional] - Value from ConfigMap/Secret/Field
    - `configMapKeyRef`: {name: string, key: string}
    - `secretKeyRef`: {name: string, key: string}
    - `fieldRef`: {fieldPath: string}

- `volumeMounts` (array) [optional]
  - `name`: string - Volume name (matches volumes[].name)
  - `mountPath`: string - Container path where volume mounts
  - `readOnly`: boolean (default: false)

- `livenessProbe` (Probe) [optional]
  - `httpGet` (object)
    - `path`: string (e.g., "/healthz")
    - `port`: int
  - `initialDelaySeconds`: int (default: 30)
  - `periodSeconds`: int (default: 10)
  - `timeoutSeconds`: int (default: 5)
  - `failureThreshold`: int (default: 3)

- `readinessProbe` (Probe) [optional]
  - `httpGet` (object)
    - `path`: string (e.g., "/readyz")
    - `port`: int
  - `initialDelaySeconds`: int (default: 10)
  - `periodSeconds`: int (default: 5)
  - `timeoutSeconds`: int (default: 3)
  - `failureThreshold`: int (default: 2)

**Validation Rules**:
- `image.registry` must be valid registry hostname
- `image.tag` must not be empty
- `port.containerPort` and `port.servicePort` must be unique within component
- `replicas` must be between 1 and 5
- Resource limits must be >= resource requests
- Probe paths must start with "/" and be valid HTTP paths
- No duplicate environment variable names

---

### 3. S3StorageConfig

Configuration for S3-compatible object storage.

**Fields**:
- `endpoint`: string - S3 endpoint URL (e.g., "minio.c8s-system.svc.cluster.local:9000")
- `bucket`: string - S3 bucket name for logs/artifacts
- `region`: string (default: "us-east-1") - S3 region
- `accessKey`: string [SecretRef] - AWS access key ID
- `secretKey`: string [SecretRef] - AWS secret access key
- `useSSL`: boolean (default: false) - Enable HTTPS for S3 connection
- `insecureSkipVerify`: boolean (default: false) - Skip SSL verification (dev only)

**Validation Rules**:
- `endpoint` must be valid hostname:port
- `bucket` must match S3 bucket naming rules (lowercase, 3-63 chars, alphanumeric + hyphens)
- `accessKey` and `secretKey` must not be empty
- `insecureSkipVerify` should only be true for development

---

### 4. PVCStorageConfig

Configuration for Kubernetes PersistentVolumeClaim storage.

**Fields**:
- `storageClass`: string - Kubernetes StorageClass name (default: system default)
- `size`: string - PVC size (e.g., "10Gi", "100Gi")
- `accessMode`: enum ["ReadWriteOnce", "ReadWriteMany", "ReadOnlyMany"] (default: "ReadWriteOnce")

**Validation Rules**:
- `size` must be valid Kubernetes quantity (number + unit)
- `storageClass` must exist in Kubernetes cluster (validated at deployment time)
- `accessMode` must match storage class capabilities

---

### 5. ResourceLimit

Kubernetes resource specification.

**Fields**:
- `cpu`: string - CPU request/limit (e.g., "100m", "1", "2000m")
- `memory`: string - Memory request/limit (e.g., "256Mi", "512Mi", "1Gi")

**Validation Rules**:
- Both fields must use valid Kubernetes quantity format
- Limits must be >= requests
- Reasonable defaults per environment:
  - **dev**: requests={cpu: "100m", memory: "256Mi"}, limits={cpu: "500m", memory: "512Mi"}
  - **staging**: requests={cpu: "250m", memory: "512Mi"}, limits={cpu: "1000m", memory: "1Gi"}
  - **production**: requests={cpu: "500m", memory: "512Mi"}, limits={cpu: "2000m", memory: "2Gi"}

---

### 6. DeploymentResult

Status and result of a deployment operation.

**Fields**:
- `status`: enum ["success", "failed", "partial"]
  - `success`: All components deployed and healthy
  - `failed`: Deployment did not proceed (validation failed)
  - `partial`: Some components deployed but not all ready

- `timestamp`: timestamp - When operation completed
- `duration`: duration - How long deployment took (milliseconds)

- `components` (array of ComponentResult)
  - `name`: string - Component name (e.g., "controller")
  - `status`: enum ["ready", "pending", "failed", "error"]
  - `replicas` (object)
    - `desired`: int
    - `ready`: int
    - `updated`: int
    - `available`: int
  - `message`: string - Status message
  - `conditions` (array)
    - `type`: string (e.g., "Available", "Progressing")
    - `status`: string (e.g., "True", "False")
    - `lastUpdateTime`: timestamp
    - `reason`: string
    - `message`: string

- `cluster` (object)
  - `namespace`: string - Kubernetes namespace
  - `version`: string - Kubernetes server version
  - `context`: string - Kubernetes context name
  - `node_count`: int - Number of nodes in cluster

- `endpoints` (object) [if successful]
  - `dashboard`: string - URL to access C8S dashboard
  - `api`: string - C8S API endpoint
  - `instructions`: string - Instructions for accessing dashboard

- `errors` (array of object) [if partial or failed]
  - `component`: string - Which component had error
  - `code`: string - Error code
  - `message`: string - Human-readable error message
  - `suggestion`: string - Suggested remediation

---

### 7. HealthCheckResult

Status of all deployed C8S components.

**Fields**:
- `timestamp`: timestamp - When health check was performed
- `overall_status`: enum ["healthy", "degraded", "unhealthy"]
  - `healthy`: All components ready and responding
  - `degraded`: Some components ready but issues detected
  - `unhealthy`: Critical components not ready

- `components` (array of ComponentHealth)
  - `name`: string - Component name
  - `status`: enum ["ready", "pending", "failed", "unknown"]
  - `replicas` (object)
    - `ready`: int
    - `total`: int
  - `last_transition`: timestamp - When status last changed
  - `message`: string - Status details
  - `probe_result` (object)
    - `type`: enum ["liveness", "readiness", "custom"]
    - `path`: string
    - `response_time_ms`: int
    - `last_check`: timestamp

- `dependencies` (array)
  - `name`: string (e.g., "storage", "rbac", "crd")
  - `status`: enum ["ready", "missing", "misconfigured"]
  - `message`: string

- `cluster_info` (object)
  - `version`: string - Kubernetes version
  - `node_count`: int
  - `available_cpu`: string
  - `available_memory`: string

---

## State Transitions

### DeploymentConfig Lifecycle

```
CREATED
  ↓ (validate)
VALIDATED
  ↓ (apply manifests)
DEPLOYING
  ↓ (all components ready)
DEPLOYED ✓
  ↓ (later: modify config)
UPDATED

Failure paths:
ANY → FAILED (validation/cluster issue)
DEPLOYING → PARTIAL (some components ready)
```

### ComponentResult Status Progression

```
[Initial] → pending → progressing → ready ✓

Or on failure:
[Initial] → pending → failed
       ↓
   error (diagnostic info collected)
```

---

## Integration with User Stories

### User Story 1: Deploy C8S with Single Command
- **Uses**: DeploymentConfig (with defaults), DeploymentResult
- **Output**: DeploymentResult with endpoints
- **Health verification**: Uses HealthCheckResult

### User Story 2: Customize Deployment Configuration
- **Uses**: DeploymentConfig (all fields), environment presets
- **Validation**: Enforces all validation rules
- **Output**: DeploymentResult showing applied customizations

### User Story 3: Verify Deployment Health
- **Uses**: HealthCheckResult
- **Queries**: ComponentResult status from all components
- **Output**: Overall health status with remediation suggestions

### User Story 4: Manage Stack Lifecycle
- **Uses**: DeploymentConfig for upgrade/downgrade
- **Tracks**: Previous DeploymentConfig versions
- **Output**: DeploymentResult showing update status

---

## Configuration Examples

### Minimal Config (Dev)
```yaml
metadata:
  name: local-c8s
  version: latest
cluster:
  namespace: c8s-system
environment:
  type: dev
storage:
  type: local
```

### Production Config
```yaml
metadata:
  name: production-c8s
  version: v0.2.0
cluster:
  namespace: c8s-system
  context: gke-prod
components:
  controller:
    replicas: 3
    resources:
      requests: {cpu: "500m", memory: "512Mi"}
      limits: {cpu: "2000m", memory: "2Gi"}
  webhook:
    replicas: 2
  apiServer:
    replicas: 2
environment:
  type: production
  logLevel: info
storage:
  type: s3-compatible
  endpoint: s3.amazonaws.com
  bucket: prod-c8s-logs
  region: us-west-2
authentication:
  type: oauth2
  secretRef: c8s-oauth-secret
tls:
  enabled: true
  secretRef: c8s-tls-cert
```

---

## API Contracts for Data Models

See `contracts/` directory for detailed API specifications:
- `contracts/cli-api.md` - CLI argument schema
- `contracts/config-schema.json` - JSON schema for configuration files
- `contracts/health-check-format.md` - Health check output format
