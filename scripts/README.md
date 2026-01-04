# scripts/

Development and CI/CD utility scripts for C8S.

## Overview

Contains shell scripts for common development tasks, validation, and CI/CD operations.

## Scripts

### tilt-ci-local.sh

**Purpose**: Runs Tilt CI locally using a kind cluster for testing Tiltfile changes.

**Usage**:
```bash
# Run tilt ci with default settings
./scripts/tilt-ci-local.sh

# Set custom cluster name
CLUSTER_NAME=my-test tilt-ci-local.sh

# Set custom timeout (seconds)
TIMEOUT=900 tilt-ci-local.sh

# Clean up and recreate cluster
CLEANUP=true tilt-ci-local.sh

# Combine options
CLUSTER_NAME=test TIMEOUT=900 CLEANUP=true ./scripts/tilt-ci-local.sh
```

**Environment Variables**:

| Variable | Default | Description |
|----------|---------|-------------|
| `CLUSTER_NAME` | `c8s-ci` | Name of the kind cluster |
| `TIMEOUT` | `600` | Timeout in seconds for tilt ci |
| `CLEANUP` | `false` | Delete existing cluster before creating |

**What It Does**:
1. Checks for required tools (kind, tilt, Docker)
2. Creates a kind cluster with custom port mappings (9080/9443)
3. Runs `tilt ci` with specified timeout
4. On success: Shows cluster status and C8S namespace resources
5. On failure: Prints debugging info (pods, events, logs)

**Port Mappings**:
- **9080** → cluster:80 (HTTP)
- **9443** → cluster:443 (HTTPS)

*(Different from dev cluster's 8000/8080 to avoid conflicts)*

**Dependencies**:
- [kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [tilt](https://docs.tilt.dev/) - Local development orchestration
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Kubernetes CLI
- Docker - Container runtime

**Example Output** (Success):
```
=== Tilt CI Local Test ===
Creating kind cluster: c8s-ci
Created cluster
Note: CI cluster uses ports 9080/9443 to avoid conflicts with local dev (8000/8080)
Running tilt ci...
[Tilt CI output...]
Tilt CI succeeded!

=== Cluster Status ===
Cluster: c8s-ci
Nodes:
NAME                  STATUS   ROLE           AGE   VERSION
c8s-ci-control-plane  Ready    control-plane  2m    v1.27.0

=== C8S Namespace ===
[C8S resources...]
```

**Example Output** (Failure):
```
Tilt CI failed!

=== Debugging Info ===
--- Cluster Info ---
[cluster info...]

--- All Pods ---
[pod status...]

--- Recent Events ---
[last 20 events...]

--- Tilt Logs (last 50 lines) ---
[tilt logs...]
```

**When to Use**:
- Testing Tiltfile changes before pushing
- Debugging CI failures locally
- Validating Helm chart changes
- Testing full deployment workflow

**Cleanup**:
```bash
# Delete the CI cluster when done
kind delete cluster --name c8s-ci

# Or use the CLEANUP flag on next run
CLEANUP=true ./scripts/tilt-ci-local.sh
```

---

## Common Workflows

### Local CI Testing Before Push
```bash
# Test Tiltfile changes
./scripts/tilt-ci-local.sh

# If successful, push changes
git push
```

### CI/CD Integration

The **tilt-ci-local.sh** script is referenced in the `.github/workflows/tilt-ci.yml` implementation.

See [.github/workflows/README.md](../.github/workflows/README.md) for CI/CD details.

## Adding New Scripts

When adding new scripts to this directory:

1. **Make executable**: `chmod +x scripts/new-script.sh`
2. **Add shebang**: `#!/bin/bash` at the top
3. **Set error handling**: `set -e` to exit on error
4. **Document usage**: Include `--help` flag
5. **Update this README**: Add script documentation
6. **Test locally**: Verify script works before committing

**Script Template**:
```bash
#!/bin/bash
# Script description

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
  cat <<EOF
Usage: $0 [OPTIONS]

Description of what this script does.

OPTIONS:
    -h, --help    Show this help message

EXAMPLES:
    $0 --option value
EOF
  exit 1
}

# Main logic here
```

## Related

- [Makefile](../Makefile) - Build automation (many targets use these scripts)
- [Tiltfile](../Tiltfile) - Development workflow
- [.github/workflows/](../.github/workflows/) - CI/CD workflows
- [docs/development/](../docs/development/) - Development guides
