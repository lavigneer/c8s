# Local Testing Guide for C8S

This guide explains how to use the C8S local development environment to create, test, and debug the C8S operator locally without requiring cloud infrastructure.

## Prerequisites

- Docker (tested with Docker 27.3.1+)
- k3d (tested with v5.8.3+)
- kubectl (tested with v1.28.15+)
- Go 1.25+ (if building from source)

## Quick Start

### 1. Create a Local Cluster

```bash
c8s dev cluster create my-dev-cluster --wait
```

This creates a local Kubernetes cluster with:
- 1 server (control plane)
- 2 agents (worker nodes)
- Registry enabled for local image deployment
- Kubeconfig automatically configured and context switched

### 2. Deploy the Operator

```bash
c8s dev deploy operator --cluster my-dev-cluster
```

This command:
- Installs all required CRDs (PipelineConfig, PipelineRun, RepositoryConnection)
- Loads the operator image into the cluster
- Deploys the operator to the `c8s-system` namespace
- Waits for the operator to be ready

### 3. Deploy Sample Pipelines

```bash
c8s dev deploy samples --cluster my-dev-cluster
```

Available samples:
- `simple-build.yaml` - Single-step build pipeline
- `multi-step.yaml` - Multi-step pipeline with dependencies
- `matrix-build.yaml` - Matrix build with parallel execution

### 4. Run Pipeline Tests

```bash
c8s dev test run --cluster my-dev-cluster
```

This executes all sample pipelines and displays results:
```
======================================================================
Pipeline Test Results
======================================================================

Summary:
  Total Tests:    3
  Passed:         3 ✓
  Failed:         0 ✗
  Timeout:        0 ⏱
```

### 5. View Logs

```bash
c8s dev test logs --cluster my-dev-cluster --follow
```

Options:
- `--follow` - Stream logs in real-time
- `--tail 100` - Show last 100 lines
- `--pipeline simple-build` - View logs for specific pipeline

### 6. Stop and Restart Cluster

```bash
# Stop the cluster (preserves state)
c8s dev cluster stop my-dev-cluster

# Restart the cluster
c8s dev cluster start my-dev-cluster

# Check status
c8s dev cluster status my-dev-cluster
```

### 7. Clean Up

```bash
# Delete the cluster
c8s dev cluster delete my-dev-cluster

# Delete all C8S clusters
c8s dev cluster delete --all
```

## Common Workflows

### Development Iteration

```bash
# Create cluster once
c8s dev cluster create dev-env

# Deploy operator
c8s dev deploy operator --cluster dev-env

# Deploy custom image for testing
c8s dev deploy operator --cluster dev-env --image my-controller:v0.1.0

# Run tests
c8s dev test run --cluster dev-env

# View results
c8s dev test logs --cluster dev-env
```

### Testing Specific Pipeline

```bash
# Deploy only specific sample
c8s dev deploy samples --cluster dev-env --select simple-build

# Run only that pipeline
c8s dev test run --cluster dev-env --pipeline simple-build

# Stream logs
c8s dev test logs --cluster dev-env --pipeline simple-build --follow
```

### JSON Output for CI/CD

```bash
# Get test results in JSON
c8s dev test run --cluster dev-env --output json

# Example output:
# {
#   "totalTests": 3,
#   "passed": 2,
#   "failed": 1,
#   "timeout": 0,
#   "results": [...]
# }
```

### Direct kubectl Access

```bash
# Get cluster context name
kubectl config get-contexts

# Use context directly
kubectl --context k3d-my-dev-cluster get nodes

# View PipelineConfigs
kubectl get pipelineconfigs

# Describe specific pipeline
kubectl describe pipelineconfig simple-build
```

## Troubleshooting

### Cluster Creation Failed

**Problem**: `failed to create cluster`

**Solutions**:
1. Check Docker is running: `docker info`
2. Check disk space: `docker system df`
3. Try with different k3s version: `c8s dev cluster create my-cluster --k8s-version v1.27.0`

### Image Load Failed

**Problem**: `image not found: ghcr.io/org/c8s-controller:latest`

**Solutions**:
1. Build image locally: `docker build -t c8s-controller:local --target controller .`
2. Deploy with local image: `c8s dev deploy operator --cluster my-cluster --image c8s-controller:local`

### Operator Not Ready

**Problem**: Operator pod stuck in pending/crash loop

**Solutions**:
1. Check pod status: `kubectl get pods -n c8s-system`
2. View pod logs: `kubectl logs -n c8s-system -l app=c8s-controller`
3. Describe pod: `kubectl describe pod -n c8s-system -l app=c8s-controller`
4. Check events: `kubectl get events -n c8s-system`

### Kubeconfig Issues

**Problem**: `Unable to connect to the server`

**Solutions**:
1. Verify cluster is running: `k3d cluster list`
2. Update context: `c8s dev cluster status my-cluster`
3. Manually switch context: `kubectl config use-context k3d-my-cluster`

### Cleanup Stuck Resources

**Problem**: Cluster won't delete

**Solutions**:
```bash
# Force delete with k3d
k3d cluster delete my-cluster --force

# Clean up orphaned volumes
docker volume prune -f

# Clean up orphaned containers
docker container prune -f
```

## Performance Tips

1. **Resource Management**
   - Default: 2 CPUs, 3.8 GB memory per cluster
   - For larger workloads, allocate more resources to Docker/Rancher Desktop

2. **Image Caching**
   - Build images locally once, reuse across multiple test runs
   - Use `--image-pull-policy=IfNotPresent` for local testing

3. **Cluster Reuse**
   - Stop/start instead of delete/recreate when possible
   - Saves initialization time and preserves deployed resources

4. **Parallel Testing**
   - Run multiple clusters for parallel testing: `c8s dev cluster create test-1` + `c8s dev cluster create test-2`
   - Useful for testing compatibility across versions

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: macos-latest  # or ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Install k3d
        run: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

      - name: Build
        run: make build

      - name: Create test cluster
        run: ./bin/c8s dev cluster create ci-test --wait

      - name: Deploy operator
        run: ./bin/c8s dev deploy operator --cluster ci-test

      - name: Run tests
        run: ./bin/c8s dev test run --cluster ci-test --output json > results.json

      - name: Publish results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: results.json

      - name: Cleanup
        if: always()
        run: ./bin/c8s dev cluster delete ci-test --force
```

### GitLab CI Example

```yaml
stages:
  - build
  - test
  - cleanup

build:
  stage: build
  script:
    - make build

test:pipeline:
  stage: test
  services:
    - docker:dind
  before_script:
    - apk add --no-cache curl bash
    - curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
  script:
    - ./bin/c8s dev cluster create ci-$CI_JOB_ID --wait
    - ./bin/c8s dev deploy operator --cluster ci-$CI_JOB_ID
    - ./bin/c8s dev test run --cluster ci-$CI_JOB_ID --output json | tee results.json
  artifacts:
    paths:
      - results.json
  after_script:
    - ./bin/c8s dev cluster delete ci-$CI_JOB_ID --force

cleanup:
  stage: cleanup
  when: always
  script:
    - c8s dev cluster delete --all --force || true
```

## Advanced Usage

### Custom Cluster Configuration

Create a config file `cluster-config.yaml`:

```yaml
name: custom-cluster
kubernetesVersion: v1.28.15
servers: 1
agents: 3
registry:
  enabled: true
  hostPort: 5000
ports:
  - hostPort: 8080
    containerPort: 80
    protocol: TCP
```

Then deploy:

```bash
c8s dev cluster create custom --config cluster-config.yaml
```

### Custom Operator Image

```bash
# Build custom controller
docker build -t my-org/c8s-controller:v0.2.0 --target controller .

# Push to registry (optional)
docker push my-org/c8s-controller:v0.2.0

# Deploy to cluster
c8s dev deploy operator \
  --cluster dev \
  --image my-org/c8s-controller:v0.2.0 \
  --image-pull-policy Always
```

### Accessing Services

```bash
# Port forward to access services
kubectl port-forward -n c8s-system svc/api-server 8080:8080

# Access metrics
kubectl port-forward -n c8s-system svc/metrics 8081:8081
```

## More Information

- [Project README](../README.md) - Project overview
- [API Documentation](../pkg/apis/v1alpha1/README.md) - CRD specifications
- [Contributing Guide](../CONTRIBUTING.md) - Development guidelines

## Getting Help

If you encounter issues:

1. Check the troubleshooting section above
2. Review cluster logs: `k3d cluster logs my-cluster`
3. Run with verbose output: `c8s dev deploy operator --cluster my-cluster -v`
4. File an issue on GitHub with:
   - Environment details (OS, Docker version, k3d version)
   - Command that failed
   - Full error output with `--verbose` flag
