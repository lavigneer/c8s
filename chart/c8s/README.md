# C8S Helm Chart

A production-ready Helm 3.x chart for deploying the C8S continuous integration stack to Kubernetes.

## Overview

This Helm chart provides an easy way to deploy the complete C8S stack (API server, controller, webhook, and frontend) to any Kubernetes cluster (k3s, kind, EKS, GKE, AKS) with a single command.

**Features**:
- ✅ Single-command deployment: `helm install c8s ./chart/c8s`
- ✅ Environment presets: dev, staging, production
- ✅ Customizable via values files and CLI flags
- ✅ Health verification via post-install hook
- ✅ Rolling updates and zero-downtime deployments
- ✅ Cross-distribution compatible (Kubernetes 1.24+)
- ✅ S3-compatible and PVC storage support
- ✅ RBAC and security-first design

## Prerequisites

- **Kubernetes 1.24+** (any distribution: k3s, kind, EKS, GKE, AKS)
- **Helm 3.x** (install from https://helm.sh/docs/intro/install/)
- **kubectl** configured to access your cluster
- Adequate cluster resources (see Requirements section below)

## Quick Start

### 1. Install on Development Cluster

```bash
# Clone the repository
git clone https://github.com/anthropics/c8s.git
cd c8s

# Install the chart with development values
helm install c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml -n c8s-system --create-namespace

# Watch the deployment
kubectl rollout status deployment/c8s-api-server -n c8s-system
kubectl rollout status deployment/c8s-controller -n c8s-system
kubectl rollout status deployment/c8s-webhook -n c8s-system
kubectl rollout status deployment/c8s-frontend -n c8s-system

# Access the dashboard
kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80
# Open http://localhost:3000 in your browser
```

### 2. Install on Production Cluster

```bash
# Install with production values
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-prod.yaml \
  -f ./chart/c8s/values.yaml \
  -n c8s-system \
  --create-namespace

# Verify deployment
kubectl get deployment -n c8s-system
```

### 3. Install with Custom Configuration

```bash
# Override specific values
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-dev.yaml \
  --set components.controller.replicas=3 \
  --set storage.type=s3-compatible \
  --set storage.s3.endpoint=s3.amazonaws.com \
  -n c8s-system \
  --create-namespace
```

## Configuration

### Environment Presets

The chart includes environment-specific value files:

**Development** (`values-dev.yaml`):
- Single replicas
- Minimal resource requests/limits
- Local storage
- Debug logging

**Staging** (`values-staging.yaml`):
- Single replicas
- Moderate resources
- PersistentVolumeClaim storage
- Info-level logging

**Production** (`values-prod.yaml`):
- Multiple replicas (HA)
- High resource requests/limits
- S3-compatible object storage
- Warn-level logging

### Customizable Parameters

#### Global Settings
```yaml
global:
  namespace: c8s-system  # Kubernetes namespace
```

#### Component Configuration
Each component (apiServer, controller, webhook, frontend) supports:
- `replicas`: Number of pod replicas
- `image`: Container image configuration
- `port`: Port configuration
- `resources`: CPU/memory requests and limits
- `livenessProbe`: Kubernetes liveness probe
- `readinessProbe`: Kubernetes readiness probe

#### Storage Configuration
```yaml
storage:
  type: s3-compatible | pvc | local
  s3:
    endpoint: minio.c8s-system.svc.cluster.local:9000
    bucket: c8s-logs
    region: us-east-1
    accessKey: <your-key>
    secretKey: <your-secret>
```

### Full Values Reference

See `values.yaml` for the complete list of customizable parameters with documentation.

## Deployment Examples

### Minimal Development Setup
```bash
helm install c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml -n c8s-system --create-namespace
```

### Production with S3 Storage
```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-prod.yaml \
  --set storage.s3.accessKey=AKIA... \
  --set storage.s3.secretKey=... \
  -n c8s-system \
  --create-namespace
```

### Custom Resource Configuration
```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-staging.yaml \
  --set components.controller.replicas=5 \
  --set components.controller.resources.limits.memory=4Gi \
  -n c8s-system \
  --create-namespace
```

## Lifecycle Management

### Upgrading

#### To Upgrade to a New Version
```bash
# Upgrade the release with new version
helm upgrade c8s ./chart/c8s \
  -f ./chart/c8s/values-prod.yaml \
  -n c8s-system

# Verify the upgrade is rolling out
kubectl rollout status deployment/c8s-api-server -n c8s-system
kubectl rollout status deployment/c8s-controller -n c8s-system
kubectl rollout status deployment/c8s-webhook -n c8s-system
kubectl rollout status deployment/c8s-frontend -n c8s-system
```

**Zero-Downtime Upgrades**: The chart uses RollingUpdate strategy with:
- `maxSurge: 1` - One extra pod allowed during update
- `maxUnavailable: 0` - No pods taken down (zero downtime)
- `progressDeadlineSeconds: 600` - 10 minute timeout for each component

#### Upgrading with New Configuration
```bash
# Upgrade and change settings at the same time
helm upgrade c8s ./chart/c8s \
  -f ./chart/c8s/values-prod.yaml \
  --set components.controller.replicas=5 \
  --set environment.logLevel=debug \
  -n c8s-system

# Custom values are automatically preserved during upgrade
```

### Release History & Rollback

#### View Release History
```bash
# See all releases for this deployment
helm history c8s -n c8s-system

# Output shows:
# REVISION  STATUS      CHART          APP VERSION  DATE                 DESCRIPTION
# 1         SUPERSEDED  c8s-0.1.0      0.1.0       ...                 Install complete
# 2         DEPLOYED    c8s-0.1.1      0.1.1       ...                 Upgrade complete
```

#### Rollback to Previous Version
```bash
# Rollback to previous release (revision 1)
helm rollback c8s -n c8s-system

# Or rollback to specific revision
helm rollback c8s 1 -n c8s-system

# Verify rollback completed
kubectl rollout status deployment/c8s-api-server -n c8s-system
```

**Note**: Rollback restores the Helm values from the previous release. Data (PVCs, logs) is preserved.

### Uninstalling

#### Clean Uninstall
```bash
# Uninstall the release
helm uninstall c8s -n c8s-system

# This removes:
# ✓ All C8S deployments, pods, and services
# ✓ ConfigMaps and secrets
# ✓ RBAC rules and service accounts
#
# This keeps (for data safety):
# - PersistentVolumeClaims (PVCs)
# - PersistentVolumes (PVs)
# - The c8s-system namespace
```

#### Full Cleanup with Data Deletion
```bash
# Delete persistent volumes
kubectl delete pvc -n c8s-system -l app.kubernetes.io/name=c8s

# Delete the namespace
kubectl delete namespace c8s-system

# This completely removes all C8S resources and data
```

#### Preserve Release History for Rollback
```bash
# Uninstall but keep history for rollback
helm uninstall c8s -n c8s-system --keep-history

# Later, restore from history instead of reinstalling
helm upgrade --install c8s ./chart/c8s -n c8s-system
```

### Upgrade Strategy

The chart uses a **RollingUpdate** strategy for deployments:

1. **Zero Downtime**: New replicas start before old ones terminate
2. **Gradual Update**: Only one extra pod runs during update
3. **Health Checks**: Each component waits for readiness before considering deployment ready
4. **Automatic Rollback**: If upgrade fails, previous version is automatically kept

**Upgrade Flow**:
```
Before:  [Pod1] [Pod2] [Pod3]  (all v0.1.0)
Step 1:  [Pod1] [Pod2] [Pod3] [NewPod]  (new pod with v0.1.1 starting)
Step 2:  [NewPod] [Pod2] [Pod3]  (Pod1 removed, NewPod ready)
Step 3:  [NewPod] [NewPod] [Pod3]  (Pod2 updated)
After:   [NewPod] [NewPod] [NewPod]  (all v0.1.1)
```

## Troubleshooting

### Check Deployment Status
```bash
# List all resources
kubectl get all -n c8s-system

# Check pod status
kubectl get pods -n c8s-system

# View pod logs
kubectl logs deployment/c8s-controller -n c8s-system --tail=50
```

### Common Issues

#### Pods Stuck in Pending
```bash
# Check resource availability
kubectl describe node

# Check events
kubectl describe pod <pod-name> -n c8s-system
```

#### Image Pull Errors
```bash
# Verify image exists and is accessible
kubectl describe pod <pod-name> -n c8s-system

# Update image in values file if needed
helm upgrade c8s ./chart/c8s \
  --set components.controller.image.tag=v0.2.0 \
  -n c8s-system
```

#### RBAC Permissions Error
```bash
# Verify service account permissions
kubectl get clusterrole c8s-controller -o yaml
kubectl get clusterrolebinding c8s-controller -o yaml
```

## Testing

### Lint the Chart
```bash
helm lint ./chart/c8s
```

### Template Validation
```bash
helm template c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml
```

### Dry-Run Install
```bash
helm install c8s ./chart/c8s \
  -f ./chart/c8s/values-dev.yaml \
  -n c8s-system \
  --create-namespace \
  --dry-run=client
```

### Full Integration Test
```bash
# Run the test script
./tests/e2e/deploy_test.sh
```

## Storage Options

### Local Storage (Development)
No additional configuration needed. Logs and artifacts are stored in pod ephemeral storage.

### PersistentVolumeClaim (Staging)
```yaml
storage:
  type: pvc
  pvc:
    enabled: true
    size: 50Gi
    storageClass: ""  # Leave empty for default
```

### S3-Compatible (Production)
```yaml
storage:
  type: s3-compatible
  s3:
    enabled: true
    endpoint: s3.amazonaws.com
    bucket: c8s-logs
    region: us-west-2
    useSSL: true
    accessKey: <your-key>
    secretKey: <your-secret>
```

## Integrating with Tilt

This Helm chart is integrated with Tilt for local development:

```bash
# Install Tilt (https://tilt.dev)
# Run Tilt with the chart
tilt up

# View logs
tilt logs

# Stop Tilt
tilt down
```

The `Tiltfile` uses this Helm chart with `values-dev.yaml` and Tilt-specific overrides.

## Architecture

### Components
- **API Server**: REST API for pipeline management
- **Controller**: Watches for pipeline changes and executes them
- **Webhook**: Validates pipeline configuration changes
- **Frontend**: Web UI for pipeline management

### Deployment Model
- Deployments with rolling update strategy (zero-downtime)
- StatefulSets (if needed for controllers)
- Service discovery via Kubernetes Services
- ConfigMaps for configuration
- Secrets for credentials (S3 keys, etc.)

## Health Checks

The chart includes automatic health verification via the post-install hook:

```bash
# The hook runs after installation to verify all components are ready
# Check the hook output
kubectl get pod -n c8s-system -l job-name=c8s-post-install-hook
```

## Security Considerations

1. **RBAC**: The chart includes a ClusterRole with minimal required permissions
2. **Secrets**: S3 credentials are stored as Kubernetes Secrets
3. **Network**: All services are ClusterIP by default (use NodePort or LoadBalancer for external access)
4. **Images**: Specify image tags explicitly (avoid `latest`)
5. **Resource Limits**: Always set resource requests and limits

## Contributing

To contribute improvements to this Helm chart, please submit a pull request to the [C8S repository](https://github.com/anthropics/c8s).

## License

This Helm chart is part of the C8S project and is licensed under the same license as the main C8S repository.

## Support

For issues, questions, or contributions, please visit:
- [GitHub Issues](https://github.com/anthropics/c8s/issues)
- [C8S Documentation](https://github.com/anthropics/c8s)
