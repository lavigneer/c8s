# Research: Local Test Environment Setup

**Date**: 2025-10-13
**Feature**: Local Test Environment for C8S Operator Development
**Researcher**: Technical Planning Phase

## Research Questions

1. Which local Kubernetes distribution (kind, minikube, k3d) best meets requirements for operator development?
2. What are the resource constraints, setup complexity, and compatibility considerations?
3. What best practices should be followed for local operator testing workflows?

## Findings

### Decision: Local Kubernetes Distribution

**Chosen**: **k3d (K3s in Docker)**

### Rationale

| Requirement | k3d | kind | minikube |
|------------|-----|------|----------|
| <3min cluster creation | ✅ 10-30s | ✅ 1-2min | ❌ 3-10min |
| <4GB RAM | ✅ <512MB | ⚠️ ~3GB | ❌ 4GB+ |
| <10GB disk | ✅ 5-8GB | ✅ ~10GB | ⚠️ 10-15GB |
| Docker-based | ✅ Native | ✅ Native | ⚠️ Issues |
| macOS support | ✅ Excellent | ✅ Excellent | ⚠️ Limited |
| Operator dev | ✅ Excellent | ✅ Excellent | ✅ Good |
| K8s v0.28 compat | ✅ Yes | ✅ Yes | ✅ Yes |

**Why k3d:**
- **Speed**: 10-30 second cluster creation enables rapid development iteration (far exceeds <3min requirement)
- **Resources**: Uses <512MB RAM per cluster, far under the 4GB limit
- **Docker integration**: Native Docker support with no macOS limitations
- **Configuration as code**: YAML config files enable team consistency and version control
- **CI/CD friendly**: Excellent GitHub Actions integration for automated testing
- **Operator compatibility**: Full CRD support, works seamlessly with controller-runtime v0.16.6
- **Team workflow**: Shareable cluster configs improve developer onboarding

### Alternatives Considered

**kind (Kubernetes IN Docker)**
- **Pros**: Runs vanilla Kubernetes, official CNCF project, excellent for CI/CD
- **Cons**: More resource-intensive than k3d (~3GB vs <512MB RAM), slower startup (1-2min vs 10-30s)
- **Why rejected**: k3d meets all requirements with better performance; kind remains viable fallback if vanilla K8s becomes critical

**minikube**
- **Pros**: Mature, official Kubernetes project, feature-rich with many addons
- **Cons**: Slowest startup (3-10min), highest resource usage (4GB+ RAM), Docker driver has ingress limitations on macOS
- **Why rejected**: Exceeds cluster creation time requirement, heaviest resource footprint, macOS Docker driver limitations

## Technical Specifications

### k3d Version and Compatibility

**Version**: k3d v5.8.x (latest stable as of Jan 2025)
**K8s Version**: v1.28.15-k3s1 (matches C8S operator K8s client-go v0.28.15 requirement)
**Docker Requirement**: v20.10.5+ (v24.0+ recommended)

### Installation

**macOS**:
```bash
brew install k3d kubectl
```

**Linux**:
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

**Verification**:
```bash
k3d version  # Should show v5.x.x
docker --version  # Requires v20.10.5+
```

### Cluster Configuration

**Recommended cluster config** (`k3d-dev-cluster.yaml`):

```yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: c8s-dev
servers: 1
agents: 2
image: rancher/k3s:v1.28.15-k3s1
ports:
  - port: 8080:80
    nodeFilters:
      - loadbalancer
  - port: 8443:443
    nodeFilters:
      - loadbalancer
options:
  k3d:
    wait: true
    timeout: "60s"
  k3s:
    extraArgs:
      - arg: --disable=traefik  # Disable Traefik if not needed
        nodeFilters:
          - server:*
  kubeconfig:
    updateDefaultKubeconfig: true
    switchCurrentContext: true
registries:
  create:
    name: registry.localhost
    host: "0.0.0.0"
    hostPort: "5000"
```

**Key design choices**:
- **1 server + 2 agents**: Minimal multi-node setup for testing pipeline distribution
- **Local registry**: Faster image loading during development (port 5000)
- **Traefik disabled**: C8S may not need ingress controller for pipeline testing
- **Automatic kubeconfig**: Switches context automatically on creation

### Operator Development Workflow

**Cluster lifecycle**:
```bash
# Create from config
k3d cluster create --config k3d-dev-cluster.yaml

# Verify
kubectl cluster-info
kubectl get nodes

# Delete
k3d cluster delete c8s-dev
```

**Image management**:
```bash
# Build operator
docker build -t c8s-operator:dev .

# Load into k3d
k3d image import c8s-operator:dev -c c8s-dev

# Or use local registry
docker tag c8s-operator:dev localhost:5000/c8s-operator:dev
docker push localhost:5000/c8s-operator:dev
```

**Operator deployment**:
```bash
# Install CRDs
kubectl apply -f config/crd/bases/

# Deploy operator
kubectl apply -f config/manager/manager.yaml

# Apply sample pipeline
kubectl apply -f config/samples/
```

### Best Practices

1. **Use envtest for unit tests**: Faster than real clusters
   - Already compatible with controller-runtime v0.16.6
   - No cluster startup overhead

2. **Use k3d for integration/e2e tests**: Real cluster validation
   - Full API compatibility
   - Tests operator reconciliation loops

3. **Image pull policy**: Set to `IfNotPresent` for local development
   ```yaml
   imagePullPolicy: IfNotPresent
   ```

4. **Configuration versioning**: Commit `k3d-dev-cluster.yaml` to repository
   - Ensures team consistency
   - Documents cluster requirements

5. **Docker resource allocation**: Ensure Docker Desktop has adequate resources
   - Recommended: 8GB RAM, 4 CPUs
   - Check: Docker Desktop > Preferences > Resources

### Known Gotchas

1. **Port conflicts**: Verify ports 8080, 8443, 6443 are available before cluster creation
2. **Traefik default**: K3s includes Traefik ingress by default (disable if using nginx-ingress)
3. **Context switching**: k3d automatically updates kubeconfig; verify with `kubectl config current-context`
4. **Cleanup**: Always use `k3d cluster delete` to properly remove resources
5. **Registry persistence**: Local registry is ephemeral; recreates with cluster

### CI/CD Integration

**GitHub Actions example**:

```yaml
name: Test Operator
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install k3d
        run: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

      - name: Create k3d cluster
        run: k3d cluster create --config k3d-dev-cluster.yaml --wait

      - name: Build and test operator
        run: |
          docker build -t c8s-operator:test .
          k3d image import c8s-operator:test
          kubectl apply -f config/crd/bases/
          kubectl apply -f config/manager/manager.yaml
          make test-e2e
```

## Implementation Notes

### Phase 0 Decisions Resolved

- ✅ **Local K8s distribution**: k3d v5.8.x with K3s v1.28.15-k3s1
- ✅ **Cluster configuration**: 1 server + 2 agents, local registry enabled
- ✅ **Target platforms**: macOS and Linux (Docker-based, no Windows for now)
- ✅ **Resource profile**: <512MB RAM per cluster, ~5-8GB disk
- ✅ **Performance targets**: 10-30s cluster creation (exceeds <3min requirement)

### Next Steps for Phase 1

1. Define data model for cluster/environment configuration schema
2. Design CLI command contracts (create, delete, deploy, test)
3. Specify sample PipelineConfig manifests for testing
4. Document quickstart guide for developers

## References

- k3d Documentation: https://k3d.io
- K3s Documentation: https://docs.k3s.io
- controller-runtime: https://github.com/kubernetes-sigs/controller-runtime
- kind (alternative): https://kind.sigs.k8s.io
- Kubernetes v1.28 Release Notes: https://kubernetes.io/blog/2023/08/15/kubernetes-v1-28-release/
