# C8S Image Building & Distribution Guide

This guide explains how to build C8S container images and distribute them via GitHub Container Registry (GHCR).

## Overview

C8S uses a multi-component architecture with three main services:
- **API Server**: REST API for pipeline management
- **Controller**: Watches and executes pipelines
- **Webhook**: Validates pipeline configurations
- **Frontend**: Web UI (optional, if separate Dockerfile exists)

## Local Development with Tilt

### Prerequisites

- Docker installed and running
- Kubernetes cluster (kind, Docker Desktop, or k3s)
- Helm 3.x installed
- Tilt installed: https://docs.tilt.dev/install.html

### Building Images Locally

Tilt automatically builds images from source code whenever files change:

```bash
# Start Tilt (builds images and deploys)
tilt up

# View logs
tilt logs c8s

# Force rebuild
tilt trigger c8s

# Stop deployment
tilt down
```

**What Tilt does**:
1. Builds Docker images from the Dockerfile using multi-stage builds
2. Tags images for local Kubernetes cluster (localhost:5000)
3. Deploys C8S Helm chart with these images
4. Watches source code changes and rebuilds automatically

### Environment Variables

Configure the build process with environment variables:

```bash
# Use custom registry (default: ghcr.io/lavigneer)
export GHCR_REGISTRY=ghcr.io/myorg

# Use custom image tag (default: latest)
export IMAGE_TAG=v0.1.0

# Start Tilt with custom settings
tilt up
```

### Dockerfile Structure

The `Dockerfile` uses multi-stage builds to create minimal images:

```dockerfile
# Stage 1: Builder
FROM golang:1.24-alpine AS builder
# Builds api-server, controller, and webhook binaries

# Stage 2: Component Images
FROM alpine:3.18 AS controller
FROM alpine:3.18 AS webhook
FROM alpine:3.18 AS api-server
```

Each target is a separate image:
- `docker build --target controller ...` → controller image
- `docker build --target webhook ...` → webhook image
- `docker build --target api-server ...` → api-server image

## GitHub Actions CI/CD

### Automatic Builds on Push

The GitHub Actions workflow `.github/workflows/build-and-push.yml` automatically:

1. **On push to main/develop**:
   - Builds all component images
   - Pushes to GHCR (if not a PR)
   - Tags with commit SHA and `latest`

2. **On version tags** (e.g., `v0.1.0`):
   - Builds all component images
   - Pushes to GHCR
   - Publishes Helm chart as OCI artifact
   - Creates GitHub Release with chart

### Workflow Triggers

```yaml
on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main, develop]
```

### Available Image Tags

The workflow creates multiple tags for flexibility:

```
ghcr.io/lavigneer/c8s-api-server:latest           # Latest on main
ghcr.io/lavigneer/c8s-api-server:main             # Latest on main branch
ghcr.io/lavigneer/c8s-api-server:develop         # Latest on develop branch
ghcr.io/lavigneer/c8s-api-server:<commit-sha>    # Specific commit
ghcr.io/lavigneer/c8s-api-server:v0.1.0          # Release version
```

### Permissions Required

The GitHub Actions workflow needs these permissions:

- `contents: read` - Read repository code
- `packages: write` - Write to GitHub Container Registry

These are configured in the workflow file and should work automatically.

## Publishing Images

### Option 1: Automatic with GitHub Actions

When you create a release tag, images are automatically built and published:

```bash
# Create a version tag
git tag v0.1.0
git push origin v0.1.0

# GitHub Actions will:
# 1. Build all component images
# 2. Push to ghcr.io/lavigneer/c8s-*:v0.1.0
# 3. Publish Helm chart as OCI artifact
# 4. Create GitHub Release
```

### Option 2: Manual from Local Build

After building locally with Tilt:

```bash
# Tag local images for GHCR
docker tag c8s-api-server ghcr.io/lavigneer/c8s-api-server:v0.1.0
docker tag c8s-controller ghcr.io/lavigneer/c8s-controller:v0.1.0
docker tag c8s-webhook ghcr.io/lavigneer/c8s-webhook:v0.1.0

# Authenticate with GitHub (requires token with package write permissions)
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Push images
docker push ghcr.io/lavigneer/c8s-api-server:v0.1.0
docker push ghcr.io/lavigneer/c8s-controller:v0.1.0
docker push ghcr.io/lavigneer/c8s-webhook:v0.1.0
```

### Option 3: Custom Registry

For private registries or air-gapped environments:

```bash
# Set custom registry
export GHCR_REGISTRY=registry.corp.example.com/c8s

# Start Tilt (will use custom registry in Helm chart)
tilt up

# Tag and push manually
docker tag c8s-api-server registry.corp.example.com/c8s/api-server:latest
docker push registry.corp.example.com/c8s/api-server:latest
```

## Using Published Images

### Update Helm Chart Values

To use published images from GHCR:

```bash
helm install c8s ./chart/c8s \
  --set components.apiServer.image.registry=ghcr.io \
  --set components.apiServer.image.repository=lavigneer/c8s-api-server \
  --set components.apiServer.image.tag=v0.1.0 \
  --set components.controller.image.registry=ghcr.io \
  --set components.controller.image.repository=lavigneer/c8s-controller \
  --set components.controller.image.tag=v0.1.0 \
  --set components.webhook.image.registry=ghcr.io \
  --set components.webhook.image.repository=lavigneer/c8s-webhook \
  --set components.webhook.image.tag=v0.1.0 \
  -n c8s-system --create-namespace
```

Or create a custom values file:

```yaml
# custom-registry-values.yaml
components:
  apiServer:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s-api-server
      tag: v0.1.0
  controller:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s-controller
      tag: v0.1.0
  webhook:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s-webhook
      tag: v0.1.0
```

Then install:
```bash
helm install c8s ./chart/c8s -f custom-registry-values.yaml
```

### Pull from Private Registry

If using a private registry, create an image pull secret:

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token> \
  --docker-email=<email> \
  -n c8s-system

# Then add to values.yaml or helm command
# imagePullSecrets: [name: ghcr-secret]
```

## Troubleshooting

### Images Not Built Locally

**Problem**: `tilt up` shows ImagePullBackOff errors

**Solution**: Ensure Docker is running and Dockerfile exists:
```bash
# Check Docker
docker ps

# Verify Dockerfile
ls -la Dockerfile Dockerfile.tilt

# Check Tilt logs
tilt logs
```

### GitHub Actions Build Failed

**Problem**: Workflow shows build failure

**Check**:
1. Dockerfile syntax: `docker build .`
2. Dockerfile targets exist: `docker build --target controller .`
3. Repository has required files: `go.mod`, `go.sum`, `cmd/`, `pkg/`

### Push to GHCR Failed

**Problem**: "Authentication required" error

**Solution**:
1. Use GitHub Personal Access Token (PAT) with `write:packages` scope
2. Create token: https://github.com/settings/tokens
3. Login: `echo $TOKEN | docker login ghcr.io -u <user> --password-stdin`

## Image Layer Caching

Both Tilt and GitHub Actions use Docker layer caching for faster builds:

**Local (Tilt)**:
```bash
# First build: slow (downloads base images, compiles Go)
tilt up

# Subsequent builds: fast (reuses layers, only rebuilds changed code)
tilt up
```

**CI/CD (GitHub Actions)**:
- Uses GitHub Actions Cache (type=gha)
- Automatically caches layers between runs
- Significantly faster on subsequent builds

## Security Considerations

### Image Signing

For production, consider signing images:

```bash
# Build and sign with cosign
docker build -t ghcr.io/lavigneer/c8s-api-server:v0.1.0 .
cosign sign ghcr.io/lavigneer/c8s-api-server:v0.1.0
```

### Registry Permissions

Set up GitHub Container Registry access controls:

1. Go to package settings: https://github.com/lavigneer/c8s/settings/packages
2. Configure visibility: Private or Public
3. Manage permissions for team members

### Base Image Security

Keep base images updated:

```dockerfile
# Current (secure)
FROM alpine:3.18

# Update periodically
FROM alpine:3.19  # when available
```

## Advanced: Publish to Multiple Registries

Modify GitHub Actions workflow to push to multiple registries:

```yaml
- name: Push to GHCR
  run: docker push ghcr.io/lavigneer/c8s-api-server:latest

- name: Push to Docker Hub
  run: docker push docker.io/lavigneer/c8s-api-server:latest

- name: Push to ECR
  run: docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/c8s-api-server:latest
```

## Quick Reference

```bash
# Local development
tilt up              # Build & deploy locally
tilt down            # Stop & cleanup

# Manual build
docker build --target api-server .
docker tag c8s-api-server ghcr.io/lavigneer/c8s-api-server:v0.1.0
docker push ghcr.io/lavigneer/c8s-api-server:v0.1.0

# CI/CD (GitHub)
git tag v0.1.0 && git push --tags  # Triggers workflow

# Use published images
helm install c8s ./chart/c8s \
  --set components.apiServer.image.tag=v0.1.0
```

## Related Documentation

- [Dockerfile Guide](../Dockerfile)
- [Helm Chart README](../chart/c8s/README.md)
- [GitHub Container Registry Docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Tilt Documentation](https://docs.tilt.dev/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
