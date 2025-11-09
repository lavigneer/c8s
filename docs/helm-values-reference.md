# C8S Helm Chart - Values Reference

Complete reference for all configurable parameters in the C8S Helm chart.

## Quick Reference

### Environment Presets

Use pre-built values files for common scenarios:

```bash
# Development (minimal resources, single replica)
helm install c8s ./chart/c8s -f values-dev.yaml

# Staging (moderate resources, persistent storage)
helm install c8s ./chart/c8s -f values-staging.yaml

# Production (high availability, S3 storage)
helm install c8s ./chart/c8s -f values-prod.yaml
```

### Configuration Override Methods

1. **Using Values Files** (recommended for complex configurations)
   ```bash
   helm install c8s ./chart/c8s -f custom-values.yaml
   ```

2. **Using --set Flags** (good for simple overrides)
   ```bash
   helm install c8s ./chart/c8s \
     --set components.controller.replicas=3 \
     --set environment.logLevel=debug
   ```

3. **Combining Multiple Files and Overrides**
   ```bash
   helm install c8s ./chart/c8s \
     -f values-dev.yaml \
     -f custom-overrides.yaml \
     --set components.webhook.replicas=2
   ```

## Configuration Parameters

### Global Settings

#### `global.namespace`
- **Type**: string
- **Default**: `c8s-system`
- **Description**: Kubernetes namespace where C8S will be deployed
- **Valid values**: Kubernetes namespace names (lowercase, alphanumeric, hyphens, 1-63 chars)
- **Example**: `--set global.namespace=my-c8s`

### Environment Settings

#### `environment.type`
- **Type**: string (enum)
- **Default**: `dev`
- **Description**: Deployment environment type (determines resource presets)
- **Valid values**: `dev`, `staging`, `production`
- **Example**: `--set environment.type=production`

#### `environment.logLevel`
- **Type**: string (enum)
- **Default**: `info`
- **Description**: Log level for all components
- **Valid values**: `debug`, `info`, `warn`, `error`
- **Example**: `--set environment.logLevel=debug`

### Component Configuration

Each component (apiServer, controller, webhook, frontend) has the following configuration:

#### `components.<component>.enabled`
- **Type**: boolean
- **Default**: `true`
- **Description**: Enable/disable this component
- **Example**: `--set components.webhook.enabled=false`

#### `components.<component>.image.registry`
- **Type**: string
- **Default**: `docker.io`
- **Description**: Container registry hostname
- **Valid values**: Any valid container registry (docker.io, gcr.io, ecr.aws, quay.io, etc.)
- **Example**: `--set components.controller.image.registry=gcr.io`

#### `components.<component>.image.repository`
- **Type**: string
- **Default**: `anthropics/c8s-<component>`
- **Description**: Image repository path
- **Example**: `--set components.apiServer.image.repository=myorg/c8s-api-server`

#### `components.<component>.image.tag`
- **Type**: string
- **Default**: `latest`
- **Description**: Image tag/version
- **Recommendation**: Use specific versions in production (never use 'latest')
- **Examples**: `v0.1.0`, `v0.2.0-rc1`

#### `components.<component>.image.pullPolicy`
- **Type**: string (enum)
- **Default**: `IfNotPresent`
- **Description**: Kubernetes image pull policy
- **Valid values**: `Always`, `IfNotPresent`, `Never`
- **Example**: `--set components.controller.image.pullPolicy=Always`

#### `components.<component>.replicas`
- **Type**: integer
- **Default**: `1`
- **Description**: Number of pod replicas
- **Valid values**: 1-5
- **Environment Defaults**:
  - Development: 1
  - Staging: 1
  - Production: 2-3 (depending on component)
- **Example**: `--set components.controller.replicas=3`

#### `components.<component>.port.containerPort`
- **Type**: integer
- **Default**: Component-specific (8080, 8081, 8443, 3000)
- **Description**: Port that the container listens on
- **Valid values**: 1-65535
- **Example**: `--set components.apiServer.port.containerPort=9000`

#### `components.<component>.port.servicePort`
- **Type**: integer
- **Default**: Component-specific
- **Description**: Port exposed by the Kubernetes Service
- **Valid values**: 1-65535
- **Example**: `--set components.apiServer.port.servicePort=8080`

#### `components.<component>.resources.requests.cpu`
- **Type**: string
- **Default**: Component-specific (50m-500m)
- **Description**: Minimum CPU guaranteed to the pod
- **Format**: Kubernetes CPU quantity (e.g., "100m", "1", "2000m")
- **Example**: `--set components.controller.resources.requests.cpu=250m`

#### `components.<component>.resources.requests.memory`
- **Type**: string
- **Default**: Component-specific (128Mi-512Mi)
- **Description**: Minimum memory guaranteed to the pod
- **Format**: Kubernetes memory quantity (e.g., "256Mi", "512Mi", "1Gi")
- **Example**: `--set components.controller.resources.requests.memory=512Mi`

#### `components.<component>.resources.limits.cpu`
- **Type**: string
- **Default**: Component-specific (250m-2000m)
- **Description**: Maximum CPU the pod can use
- **Format**: Kubernetes CPU quantity
- **Recommendation**: Set to 2-4x the request value
- **Example**: `--set components.controller.resources.limits.cpu=1000m`

#### `components.<component>.resources.limits.memory`
- **Type**: string
- **Default**: Component-specific (256Mi-2Gi)
- **Description**: Maximum memory the pod can use
- **Format**: Kubernetes memory quantity
- **Recommendation**: Set to 1.5-2x the request value
- **Example**: `--set components.controller.resources.limits.memory=1Gi`

### Storage Configuration

#### `storage.type`
- **Type**: string (enum)
- **Default**: `local`
- **Description**: Type of storage backend
- **Valid values**: `local`, `pvc`, `s3-compatible`
- **Use cases**:
  - `local`: Development (ephemeral)
  - `pvc`: Staging (persistent, requires storage class)
  - `s3-compatible`: Production (MinIO, AWS S3, etc.)
- **Example**: `--set storage.type=s3-compatible`

#### S3-Compatible Storage

##### `storage.s3.enabled`
- **Type**: boolean
- **Default**: `false`
- **Description**: Enable S3 storage
- **Example**: `--set storage.s3.enabled=true`

##### `storage.s3.endpoint`
- **Type**: string
- **Default**: `minio.c8s-system.svc.cluster.local:9000`
- **Description**: S3 endpoint URL
- **Examples**:
  - MinIO in-cluster: `minio.c8s-system.svc.cluster.local:9000`
  - AWS S3: `s3.amazonaws.com` or `s3.us-west-2.amazonaws.com`
  - DigitalOcean Spaces: `nyc3.digitaloceanspaces.com`
- **Example**: `--set storage.s3.endpoint=s3.amazonaws.com`

##### `storage.s3.bucket`
- **Type**: string
- **Default**: `c8s-logs`
- **Description**: S3 bucket name
- **Rules**: Lowercase, 3-63 chars, alphanumeric + hyphens
- **Example**: `--set storage.s3.bucket=my-org-c8s-logs`

##### `storage.s3.region`
- **Type**: string
- **Default**: `us-east-1`
- **Description**: AWS region
- **Examples**: `us-west-2`, `eu-west-1`, `ap-southeast-1`
- **Example**: `--set storage.s3.region=us-west-2`

##### `storage.s3.useSSL`
- **Type**: boolean
- **Default**: `false` (dev), `true` (prod)
- **Description**: Use HTTPS for S3 connection
- **Recommendation**: Always `true` in production
- **Example**: `--set storage.s3.useSSL=true`

##### `storage.s3.insecureSkipVerify`
- **Type**: boolean
- **Default**: `false`
- **Description**: Skip SSL certificate verification
- **WARNING**: Do NOT use in production
- **Example**: `--set storage.s3.insecureSkipVerify=true`

##### `storage.s3.accessKey`
- **Type**: string
- **Default**: empty
- **Description**: AWS access key ID
- **Security**: Use environment variables or Kubernetes Secrets
- **Example**: `--set storage.s3.accessKey=AKIA...`

##### `storage.s3.secretKey`
- **Type**: string
- **Default**: empty
- **Description**: AWS secret access key
- **Security**: Use environment variables or Kubernetes Secrets
- **Example**: `--set storage.s3.secretKey=...`

#### PersistentVolumeClaim Storage

##### `storage.pvc.enabled`
- **Type**: boolean
- **Default**: `false`
- **Description**: Enable PVC storage
- **Example**: `--set storage.pvc.enabled=true`

##### `storage.pvc.storageClass`
- **Type**: string
- **Default**: empty (uses default)
- **Description**: Kubernetes StorageClass name
- **How to list**: `kubectl get storageclass`
- **Examples**: `fast`, `slow`, `ebs-gp2`, `pd-standard`
- **Example**: `--set storage.pvc.storageClass=fast`

##### `storage.pvc.size`
- **Type**: string
- **Default**: `10Gi`
- **Description**: PVC size
- **Format**: Kubernetes quantity (e.g., "10Gi", "100Gi", "1Ti")
- **Example**: `--set storage.pvc.size=100Gi`

##### `storage.pvc.accessMode`
- **Type**: string (enum)
- **Default**: `ReadWriteOnce`
- **Description**: PVC access mode
- **Valid values**:
  - `ReadWriteOnce`: Single node read/write (most common)
  - `ReadWriteMany`: Multiple nodes read/write
  - `ReadOnlyMany`: Multiple nodes read-only
- **Example**: `--set storage.pvc.accessMode=ReadWriteMany`

### RBAC Configuration

#### `serviceAccount.create`
- **Type**: boolean
- **Default**: `true`
- **Description**: Create service account for controller
- **Example**: `--set serviceAccount.create=true`

#### `serviceAccount.name`
- **Type**: string
- **Default**: `c8s-controller`
- **Description**: Service account name
- **Example**: `--set serviceAccount.name=c8s-sa`

#### `rbac.create`
- **Type**: boolean
- **Default**: `true`
- **Description**: Create ClusterRole and ClusterRoleBinding
- **Example**: `--set rbac.create=true`

### Post-Install Hook

#### `postInstallHook.enabled`
- **Type**: boolean
- **Default**: `true`
- **Description**: Run health check after installation
- **Example**: `--set postInstallHook.enabled=true`

#### `postInstallHook.timeout`
- **Type**: integer
- **Default**: `300`
- **Description**: Timeout in seconds for post-install health check
- **Example**: `--set postInstallHook.timeout=600`

## Configuration Examples

### Example 1: Development Setup
```bash
helm install c8s ./chart/c8s -f values-dev.yaml \
  --set environment.logLevel=debug \
  --set components.controller.replicas=1
```

**Result**: Single replica of each component, debug logging, minimal resources

### Example 2: Staging with S3
```bash
helm install c8s ./chart/c8s -f values-staging.yaml \
  --set storage.type=s3-compatible \
  --set storage.s3.enabled=true \
  --set storage.s3.endpoint=minio.staging.example.com:9000 \
  --set storage.s3.accessKey=AKIA... \
  --set storage.s3.secretKey=...
```

**Result**: Staging configuration with S3-compatible storage

### Example 3: Production HA
```bash
helm install c8s ./chart/c8s -f values-prod.yaml \
  --set components.controller.replicas=5 \
  --set components.webhook.replicas=3 \
  --set components.apiServer.replicas=3 \
  --set storage.s3.endpoint=s3.us-west-2.amazonaws.com \
  --set storage.s3.accessKey=AKIA... \
  --set storage.s3.secretKey=...
```

**Result**: Production HA setup with high replicas and AWS S3 storage

### Example 4: Custom Resource Limits
```bash
helm install c8s ./chart/c8s -f values-prod.yaml \
  --set components.controller.resources.requests.cpu=1000m \
  --set components.controller.resources.requests.memory=2Gi \
  --set components.controller.resources.limits.cpu=4000m \
  --set components.controller.resources.limits.memory=4Gi
```

**Result**: Custom resource requests and limits for controller

## Validating Your Configuration

Before installing, always validate:

```bash
# Check Helm chart syntax
helm lint ./chart/c8s

# Preview generated manifests
helm template c8s ./chart/c8s -f your-values.yaml

# Dry-run installation
helm install c8s ./chart/c8s -f your-values.yaml --dry-run --debug
```

## Common Configuration Patterns

### Pattern 1: Image Registry Mirror
If you have a private registry mirror:
```bash
helm install c8s ./chart/c8s \
  --set components.apiServer.image.registry=registry.corp.example.com \
  --set components.controller.image.registry=registry.corp.example.com \
  --set components.webhook.image.registry=registry.corp.example.com \
  --set components.frontend.image.registry=registry.corp.example.com
```

### Pattern 2: Custom Namespace
```bash
helm install c8s ./chart/c8s \
  --set global.namespace=my-cicd \
  -n my-cicd \
  --create-namespace
```

### Pattern 3: All S3 Configuration
Create a file `s3-values.yaml`:
```yaml
storage:
  type: s3-compatible
  s3:
    enabled: true
    endpoint: s3.amazonaws.com
    bucket: my-org-c8s
    region: us-west-2
    useSSL: true
    accessKey: AKIA...
    secretKey: ...
```

Then install:
```bash
helm install c8s ./chart/c8s -f values-prod.yaml -f s3-values.yaml
```

## Security Best Practices

1. **Never use 'latest' image tag in production** - always specify explicit versions
2. **Store S3 credentials in Kubernetes Secrets**, not values files
3. **Enable SSL for S3 connections** in production
4. **Set resource limits** to prevent resource exhaustion
5. **Use NetworkPolicy** to restrict pod-to-pod communication (if available)
6. **Review RBAC permissions** after installation

## Environment Preset Comparison

| Setting | Development | Staging | Production |
|---------|-------------|---------|------------|
| Replicas | 1 | 1 | 2-3 |
| CPU Requests | 50-100m | 250m | 500m |
| Memory Requests | 128-256Mi | 512Mi | 512Mi |
| Storage Type | local | pvc | s3-compatible |
| Log Level | debug | info | warn |

## Related Documentation

- [C8S Helm Chart README](../chart/c8s/README.md) - Installation and usage guide
- [Helm Official Documentation](https://helm.sh/docs/) - Helm reference
- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) - Resource requests/limits
