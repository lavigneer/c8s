# Tilt Development Environment

Use Tilt for local Kubernetes development with hot reload.

## Running in Devbox (Recommended)
All commands should be run in devbox to ensure consistent environments:
```bash
devbox run make tilt-up        # Start Tilt (creates k3d cluster if needed)
devbox run make tilt-down      # Stop Tilt
devbox run make tilt-logs      # View Tilt logs
devbox run make tilt-status    # Check Tilt status
```

Or enter devbox shell once:
```bash
devbox shell
make tilt-up
make tilt-logs
# etc...
```

## Quick Start
```bash
make tilt-up        # Start Tilt (creates k3d cluster if needed)
make tilt-down      # Stop Tilt
make tilt-logs      # View Tilt logs
make tilt-status    # Check Tilt status
```

## What Tilt Does
- Creates and manages a local Kubernetes cluster (k3d) for development
- Automatically rebuilds and reloads code changes to running containers
- Provides a web UI at localhost:10350 for monitoring and logs
- Manages local image building and pushing to local registry

## Development Workflow
1. `make tilt-up` - Starts development environment and creates cluster if needed
2. Edit code - Changes are automatically detected by Tilt
3. Tilt rebuilds and reloads containers automatically
4. Check localhost:10350 for build status, logs, and resource state
5. `make tilt-down` - Stop when done

## Alternative: Manual Development (without Tilt)
```bash
make run-controller  # Run controller locally (requires kubeconfig)
make run-webhook     # Run webhook server locally
```

## CI/CD Testing
```bash
make tilt-ci-local   # Run tilt CI locally with kind cluster
make tilt-ci-clean   # Clean up kind cluster
```

## Cluster Management
```bash
make check-deps          # Verify Docker, kubectl, k3d installed
make clean-clusters      # Delete all c8s test clusters
```

## Troubleshooting
- **Dependencies not installed**: Run `make check-deps`
- **Cluster issues**: Check `make tilt-logs` or `make tilt-status`
- **Need fresh cluster**: Run `make clean-clusters` then `make tilt-up`
- **Manual logs**: `make tilt-logs` shows comprehensive output

## Notes
- Requires: Docker, kubectl, k3d, and Tilt
- Web UI is the primary interface for monitoring
- Hot reload makes iterative development fast
- Great for testing controller behavior in a real cluster
