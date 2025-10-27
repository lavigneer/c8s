# Tilt CI Setup

This document describes how to test the Tiltfile using `tilt ci` locally and in CI/CD pipelines.

## What is Tilt CI?

`tilt ci` is Tilt's mode for continuous integration. It:
- Executes the Tiltfile in a non-interactive way
- Builds all Docker images
- Deploys all Kubernetes resources
- Waits for all services to be healthy
- Exits with status code 0 on success, non-zero on failure
- Provides detailed logging for debugging

See https://docs.tilt.dev/ci.html for more information.

## Prerequisites

For local testing, you need:

1. **Docker** - Container runtime
2. **kind** - Kubernetes in Docker (for local testing)
3. **kubectl** - Kubernetes CLI
4. **Tilt** - The Tilt development tool

### Installation

```bash
# Install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install tilt
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

# Verify installations
kind version
tilt version
kubectl version --client
```

## Running Tilt CI Locally

### Quick Start

```bash
# Run tilt ci with a kind cluster
make tilt-ci-local

# Or manually
bash scripts/tilt-ci-local.sh
```

The script will:
1. Check dependencies (kind, tilt, docker)
2. Create a `c8s-ci` kind cluster
3. Run `tilt ci` with a 600-second timeout
4. Display cluster status and debug info on failure

### Custom Configuration

```bash
# Use a different cluster name
CLUSTER_NAME=my-test-cluster timeout=900 bash scripts/tilt-ci-local.sh

# Increase timeout to 15 minutes
TIMEOUT=900 make tilt-ci-local
```

### Cleanup

```bash
# Clean up the kind cluster
make tilt-ci-clean

# Or manually
kind delete cluster --name c8s-ci
```

## Tilt CI Configuration

### `.tilt-ci.yml`

The `.tilt-ci.yml` file configures how Tilt behaves in CI:

```yaml
kubeContext: kind-c8s       # Kubernetes context to use
env:
  CGO_ENABLED: "0"          # Build with CGO disabled
  GOOS: linux               # Target OS
  GOARCH: amd64             # Target architecture
update_mode: auto           # Auto-update mode
```

### Tiltfile Adjustments for CI

The Tiltfile automatically detects CI mode and adjusts behavior:

```python
# Tilt sets environment variables in CI mode
# You can check for CI and adjust resource configuration
if os.environ.get('TILT_CI') == '1':
    # CI-specific configuration
    pass
```

## GitHub Actions CI

The `.github/workflows/tilt-ci.yml` workflow:

1. **Checks out code** - Using `actions/checkout`
2. **Sets up Docker** - With Docker Buildx for better layer caching
3. **Creates Kind cluster** - Custom configuration with port mappings
4. **Installs Go 1.24** - Matches project requirements
5. **Installs Tilt** - Latest stable version
6. **Runs `tilt ci`** - With 10-minute timeout
7. **Collects debug info** - On failure (cluster status, logs, events)
8. **Uploads logs** - As artifacts for investigation

### Trigger Conditions

The workflow runs on:
- Push to `main` and `develop` branches
- Pull requests to `main` and `develop` branches

### Workflow Status

Check the status in your GitHub repository:
- **Actions** tab → **Tilt CI** → Select a run

Failed runs will have artifact logs available for download.

## Troubleshooting

### "kind is not installed"

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### "tilt is not installed"

```bash
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
```

### "Docker is not running"

Start Docker daemon:
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker

# Windows
docker desktop or docker wsl2
```

### Tilt CI Timeout

If the timeout is too short, increase it:
```bash
TIMEOUT=1200 make tilt-ci-local  # 20 minutes
```

### Network Issues in CI

The GitHub Actions workflow provides custom kind cluster networking:
- Allows connections to localhost:8080 (HTTP)
- Allows connections to localhost:8443 (HTTPS)

If services aren't accessible, check the Kind cluster config in `.github/workflows/tilt-ci.yml`.

### Memory or Resource Issues

Kind cluster might need more resources. Check Docker's resource allocation:
- macOS: Docker Desktop → Preferences → Resources
- Linux: Adjust docker daemon configuration
- Windows: WSL2 resource limits

Increase the number of agents in Kind config if needed:
```yaml
nodes:
- role: control-plane
- role: worker  # Add worker nodes
- role: worker
```

## Tiltfile Best Practices for CI

1. **Use explicit resource names** - Don't rely on auto-discovery
2. **Set pod_readiness='wait'** - CI waits for pods to be ready
3. **Use health checks** - livenessProbe and readinessProbe
4. **Avoid port forwarding** - Not needed in CI
5. **Use meaningful labels** - For resource grouping and filtering
6. **Set reasonable timeouts** - Avoid hanging indefinitely

## Debugging Failed CI Runs

When `tilt ci` fails locally:

1. **Check the output** - Tilt prints detailed error messages
2. **Review pod logs** - `kubectl logs <pod> -n <namespace>`
3. **Check events** - `kubectl get events -A --sort-by='.lastTimestamp'`
4. **Inspect resources** - `kubectl describe <resource> -n <namespace>`
5. **View Tilt logs** - `tail -f ~/.tilt/tilt.log`

## Integration with Development

You can run `tilt ci` before pushing changes:

```bash
# Test your Tiltfile changes locally
make tilt-ci-local

# If successful, commit and push
git add .
git commit -m "Update Tiltfile"
git push
```

This catches issues before they reach the main CI pipeline.

## Performance Tuning

### Build Caching

Tilt caches Docker layers by default. To improve CI performance:
1. Use shared Docker layer caching
2. Order Dockerfile commands from least to most frequently changed
3. Use `.dockerignore` to exclude unnecessary files

### Resource Limits

Set appropriate resource requests/limits in Kubernetes manifests:
```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

This helps Kind cluster with limited resources.

## Next Steps

1. **Run locally first**: `make tilt-ci-local`
2. **Fix any issues** based on debug output
3. **Push to GitHub** - CI will automatically test the Tiltfile
4. **Monitor CI runs** - Check the Actions tab for status

## References

- [Tilt CI Documentation](https://docs.tilt.dev/ci.html)
- [Kind Project](https://kind.sigs.k8s.io/)
- [GitHub Actions for Kubernetes](https://github.com/marketplace?query=kubernetes)
