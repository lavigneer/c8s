# Deployment and Cluster Management

Deploy C8S components to Kubernetes clusters.

## Running in Devbox (Recommended)
All commands should be run in devbox to ensure consistent environments:
```bash
devbox run make deploy          # Deploy controller and webhook to cluster
devbox run make undeploy        # Remove controller and webhook from cluster
devbox run make install-crds    # Install CRDs to cluster
devbox run make uninstall-crds  # Uninstall CRDs from cluster
```

Or enter devbox shell once:
```bash
devbox shell
make deploy
make install-crds
# etc...
```

## Quick Commands
```bash
make deploy          # Deploy controller and webhook to cluster
make undeploy        # Remove controller and webhook from cluster
make install-crds    # Install CRDs to cluster
make uninstall-crds  # Uninstall CRDs from cluster
```

## Docker Image Management
```bash
make docker-build           # Build all Docker images
make docker-build-controller # Build controller image only
make docker-build-webhook   # Build webhook image only
make docker-push            # Push images to registry
```

## Deployment Process
1. Generate manifests: `make manifests`
2. Build Docker images: `make docker-build`
3. Push to registry: `make docker-push` (if using remote registry)
4. Deploy to cluster: `make deploy`

## What Gets Deployed
- **CRDs**: Custom resource definitions used by C8S
- **Controller**: Main reconciliation loop
- **Webhook**: Validation and mutation webhooks
- **Namespace**: c8s-system (created automatically)

## Configuration
- **Registry**: Defaults to ghcr.io/org, override with `DOCKER_REGISTRY=your-registry`
- **Version**: Uses git tags, falls back to git describe or dirty indicator

## Prerequisites
- Kubernetes cluster with kubeconfig configured
- Docker for building images
- kubectl for applying manifests

## Notes
- Images are tagged with both version and "latest"
- All resources deployed to c8s-system namespace
- Webhook requires proper mTLS configuration in cluster
- Use `make undeploy` to clean up before redeploying
