# C8S Deployment Manifests

**Status**: LEGACY - Helm chart is the recommended deployment method

## Current Status

This directory contains legacy Kubernetes manifests. **For all deployments, use the Helm chart instead**:

```bash
# Recommended: Use Helm chart
helm install c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml

# Or use Tilt for local development (preferred)
tilt up
```

## Active Manifests

These manifests are still actively used by specific workflows:

### `crds.yaml`
- **Status**: ACTIVE
- **Purpose**: Standalone CRD installation
- **Usage**: For users who want to install CRDs separately from Helm chart
- **Command**: `kubectl apply -f deploy/crds.yaml`
- **Note**: Also available via `make install-crds`

### `minio.yaml`
- **Status**: ACTIVE
- **Purpose**: Local S3-compatible storage for Tilt development
- **Usage**: Automatically deployed by Tiltfile for local development
- **Command**: `kubectl apply -f deploy/minio.yaml` (manual)
- **Note**: Only needed for local development, not production

## Deprecated Manifests

The following manifests are **deprecated** in favor of the Helm chart:

### `install.yaml`
- **Status**: DEPRECATED
- **Replacement**: `helm install c8s ./chart/c8s`
- **Note**: Contains full deployment, now maintained in Helm chart

### `namespace.yaml`
- **Status**: DEPRECATED
- **Replacement**: Helm chart creates namespace automatically (`--create-namespace`)

### `controller-deployment.yaml`, `controller-rbac.yaml`
- **Status**: DEPRECATED
- **Replacement**: Helm chart templates: `chart/c8s/templates/controller/`

### `webhook-deployment.yaml`, `webhook-service.yaml`, `webhook-ingress.yaml`, `webhook-rbac.yaml`
- **Status**: DEPRECATED
- **Replacement**: Helm chart templates: `chart/c8s/templates/webhook/`

## Migration Guide

If you're using manual `kubectl apply` commands, migrate to Helm:

### Before (Manual kubectl)
```bash
kubectl apply -f deploy/crds.yaml
kubectl apply -f deploy/install.yaml
kubectl apply -f deploy/webhook-deployment.yaml
kubectl apply -f deploy/webhook-service.yaml
```

### After (Helm chart)
```bash
# Install with default values
helm install c8s ./chart/c8s

# Or with development values
helm install c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml

# Upgrade existing deployment
helm upgrade c8s ./chart/c8s
```

## Why Helm?

Helm provides significant advantages over manual manifests:

1. **Single Source of Truth**: All configuration in one place
2. **Environment-Specific Values**: Use different values files for dev/staging/prod
3. **Template Flexibility**: Customize deployment without editing manifests
4. **Versioning**: Track chart versions and rollback easily
5. **Dependencies**: Manage dependencies like cert-manager
6. **Atomic Deploys**: All-or-nothing deployments reduce partial failures

## Local Development

For local development, **use Tilt** instead of manual manifests:

```bash
tilt up
```

Tilt automatically:
- Deploys the Helm chart with dev values
- Builds and syncs Docker images
- Provides live reload for code changes
- Sets up port forwarding and ngrok tunnels
- Deploys MinIO for local S3 storage

See [docs/development/tilt-setup.md](../docs/development/tilt-setup.md) for details.

## Future Plans

- ✅ Phase 1: Document deprecation (current)
- 📋 Phase 2: Remove Makefile `deploy`/`undeploy` targets
- 📋 Phase 3: Archive deprecated manifests to `deploy/legacy/`
- 📋 Phase 4: Keep only `crds.yaml` and `minio.yaml`

## Questions?

- **"I need manual manifests for CI/CD"**: Use `helm template` to generate manifests from the chart
- **"I don't want to use Helm"**: Consider kustomize with the Helm-generated manifests
- **"I need custom configuration"**: Customize Helm values files instead of editing manifests

See [chart/c8s/README.md](../chart/c8s/README.md) for Helm chart documentation.
