# C8S Helm Chart - Values Reference

Complete documentation of all customizable parameters in the C8S Helm chart.

## Quick Navigation

- [Global Settings](#global-settings)
- [Environment Configuration](#environment-configuration)
- [Component Configuration](#component-configuration)
- [Storage Configuration](#storage-configuration)
- [Usage Examples](#usage-examples)

---

## Global Settings

### `global.namespace`

- **Type**: `string`
- **Default**: `c8s-system`
- **Description**: Kubernetes namespace where C8S will deploy
- **Example**: `helm install c8s ./chart/c8s --set global.namespace=production`

---

## Environment Configuration

### `environment.type`

- **Type**: `string`
- **Default**: `dev`
- **Valid Values**: `dev` | `staging` | `production`
- **Description**: Environment type affecting defaults for replicas, resources, storage

**Effects on Defaults**:
- `dev`: 1 replica, 50m CPU/128Mi memory, local storage, debug logging
- `staging`: 1 replica, 250m CPU/512Mi memory, PVC storage, info logging
- `production`: 3 replicas (HA), 500m CPU/512Mi memory, S3 storage, warn logging

### `environment.logLevel`

- **Type**: `string`
- **Default**: `info`
- **Valid Values**: `debug` | `info` | `warn` | `error`
- **Description**: Log level for all components

---

## Component Configuration

All components (apiServer, controller, webhook, frontend) support these parameters:

### Replicas

`components.{COMPONENT}.replicas`

- **Type**: `integer`
- **Default**: 1
- **Valid Range**: 1-20
- **Example**: `--set components.controller.replicas=3`

### Image Configuration

#### Registry
`components.{COMPONENT}.image.registry`
- **Default**: `localhost:5000` (dev), `ghcr.io` (prod)
- **Example**: `--set components.controller.image.registry=docker.io`

#### Repository
`components.{COMPONENT}.image.repository`
- **Default**: `{owner}/{repo}/{component}`
- **Example**: `--set components.controller.image.repository=myorg/c8s-controller`

#### Tag
`components.{COMPONENT}.image.tag`
- **Default**: `latest`
- **Example**: `--set components.controller.image.tag=v0.2.0`

### Resources

#### CPU Request
`components.{COMPONENT}.resources.requests.cpu`
- **Format**: Kubernetes notation (e.g., `500m` = 0.5 CPU)
- **Default**: `50m` (dev), `250m` (staging), `500m` (prod)

#### Memory Request
`components.{COMPONENT}.resources.requests.memory`
- **Format**: Kubernetes notation (e.g., `512Mi`, `1Gi`)
- **Default**: `128Mi` (dev), `512Mi` (staging), `512Mi` (prod)

#### CPU Limit
`components.{COMPONENT}.resources.limits.cpu`
- **Default**: `250m` (dev), `500m` (staging), `2000m` (prod)

#### Memory Limit
`components.{COMPONENT}.resources.limits.memory`
- **Default**: `256Mi` (dev), `1Gi` (staging), `2Gi` (prod)

---

## Storage Configuration

### `storage.type`

- **Type**: `string`
- **Default**: `local` (dev), `pvc` (staging), `s3-compatible` (prod)
- **Valid Values**: `local` | `pvc` | `s3-compatible`

**Behavior**:
- `local`: Ephemeral (lost on pod restart)
- `pvc`: PersistentVolumeClaim
- `s3-compatible`: AWS S3, MinIO, or compatible service

### S3 Configuration

`storage.s3.*` parameters:

- **endpoint**: S3 host (default: `minio.c8s-system.svc.cluster.local:9000`)
- **bucket**: Bucket name (default: `c8s-logs`)
- **region**: AWS region (default: `us-east-1`)
- **accessKey**: AWS Access Key ID
- **secretKey**: AWS Secret Access Key

Example:
```bash
helm install c8s ./chart/c8s \
  --set storage.s3.endpoint=s3.amazonaws.com \
  --set storage.s3.accessKey=$AWS_KEY \
  --set storage.s3.secretKey=$AWS_SECRET
```

### PVC Configuration

`storage.pvc.*` parameters:

- **size**: Volume size (default: `10Gi`)
  - Example: `--set storage.pvc.size=50Gi`

- **storageClass**: Storage class name (default: empty, uses default)
  - Example: `--set storage.pvc.storageClass=fast-ssd`

---

## Usage Examples

### Development Setup

```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-dev.yaml \
  -n c8s-system --create-namespace
```

### Production Setup

```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-prod.yaml \
  --set storage.s3.accessKey=$AWS_KEY \
  --set storage.s3.secretKey=$AWS_SECRET \
  -n c8s-system --create-namespace
```

### Custom Configuration

```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-staging.yaml \
  --set components.controller.replicas=2 \
  --set environment.logLevel=debug \
  --set storage.type=pvc \
  --set storage.pvc.size=50Gi \
  -n c8s-system --create-namespace
```

---

For complete documentation, see [values.yaml](../chart/c8s/values.yaml) with inline comments.
