# Quick Start: Local Test Environment

**Feature**: Local Test Environment Setup
**Audience**: C8S developers
**Time to Complete**: ~10 minutes
**Prerequisites**: Docker, kubectl

## Overview

This guide walks you through setting up a local Kubernetes test environment for developing and testing the C8S pipeline operator. By the end, you'll have a running cluster with the operator deployed and sample pipelines executing.

## Prerequisites

Before starting, ensure you have:

1. **Docker** (v20.10.5 or later)
   ```bash
   docker --version
   # Should show v20.10.5 or higher
   ```

2. **kubectl** (v1.28 or compatible)
   ```bash
   kubectl version --client
   ```

3. **C8S CLI** (built from source)
   ```bash
   make build-cli
   ./bin/c8s version
   ```

4. **System Resources**
   - Minimum: 8GB RAM, 20GB free disk space
   - Recommended: 16GB RAM, 50GB free disk space

## Step 1: Install k3d

The C8S local environment uses k3d to create lightweight Kubernetes clusters.

### macOS
```bash
brew install k3d
```

### Linux
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

### Verify Installation
```bash
k3d version
# Should show v5.8.x or later
```

## Step 2: Create a Local Cluster

Create your first local Kubernetes cluster:

```bash
c8s dev cluster create
```

This command:
- Creates a cluster named "c8s-dev"
- Deploys 1 server + 2 agent nodes
- Sets up a local container registry at `localhost:5000`
- Updates your kubeconfig and switches context
- Takes ~30 seconds to complete

**Expected output**:
```
Creating cluster 'c8s-dev'...
✓ Cluster 'c8s-dev' created successfully
✓ Kubeconfig updated: ~/.kube/config
✓ Current context: k3d-c8s-dev
✓ Registry available at: localhost:5000

Cluster Status:
  Name:        c8s-dev
  Nodes:       3 (1 server, 2 agents)
  K8s Version: v1.28.15+k3s1
  API Server:  https://0.0.0.0:6443
```

### Verify Cluster

```bash
kubectl get nodes
```

You should see 3 nodes in `Ready` status:
```
NAME                   STATUS   ROLES                  AGE   VERSION
k3d-c8s-dev-server-0   Ready    control-plane,master   1m    v1.28.15+k3s1
k3d-c8s-dev-agent-0    Ready    <none>                 1m    v1.28.15+k3s1
k3d-c8s-dev-agent-1    Ready    <none>                 1m    v1.28.15+k3s1
```

## Step 3: Build and Load Operator Image

Build the C8S operator container image:

```bash
# From the c8s repository root
docker build -t c8s-operator:dev .
```

Load the image into your local cluster:

```bash
k3d image import c8s-operator:dev -c c8s-dev
```

**Why?** The local cluster can't pull from your local Docker daemon. Loading the image makes it available inside the cluster.

## Step 4: Deploy the Operator

Deploy the C8S operator to your cluster:

```bash
c8s dev deploy operator
```

This command:
- Installs CRDs (PipelineConfig, PipelineRun, etc.)
- Creates the `c8s-system` namespace
- Deploys the operator with RBAC
- Waits for the operator to be ready
- Takes ~1 minute to complete

**Expected output**:
```
Deploying C8S operator to cluster 'c8s-dev'...

✓ Installing CRDs...
  - pipelineconfigs.c8s.io
  - pipelineruns.c8s.io

✓ Loading operator image to cluster...

✓ Deploying operator to namespace 'c8s-system'...
  - ServiceAccount: c8s-controller-manager
  - Role: c8s-leader-election-role
  - RoleBinding: c8s-leader-election-rolebinding
  - Deployment: c8s-controller-manager

✓ Waiting for operator to be ready...

Operator Status:
  Namespace:  c8s-system
  Replicas:   1/1 ready
  Image:      c8s-operator:dev
  Status:     Running
```

### Verify Operator

```bash
kubectl get pods -n c8s-system
```

You should see the operator pod running:
```
NAME                                     READY   STATUS    RESTARTS   AGE
c8s-controller-manager-xxxxxxxxxx-xxxxx  1/1     Running   0          1m
```

View operator logs:
```bash
kubectl logs -n c8s-system -l app=c8s-controller-manager --tail=50
```

## Step 5: Deploy Sample Pipelines

Deploy sample PipelineConfigs to test the operator:

```bash
c8s dev deploy samples
```

This command:
- Applies YAML manifests from `config/samples/`
- Deploys 3 sample pipelines:
  - `simple-build`: Single-step build
  - `multi-step`: Multi-step pipeline with dependencies
  - `matrix-build`: Parallel matrix builds

**Expected output**:
```
Deploying sample PipelineConfigs to cluster 'c8s-dev'...

✓ Applying samples:
  - simple-build.yaml
  - multi-step.yaml
  - matrix-build.yaml

✓ 3 PipelineConfigs deployed successfully
```

### View PipelineConfigs

```bash
kubectl get pipelineconfigs
```

You should see:
```
NAME           REPOSITORY                         BRANCHES   AGE
simple-build   https://github.com/example/repo    [main]     1m
multi-step     https://github.com/example/repo    [main]     1m
matrix-build   https://github.com/example/repo    [main]     1m
```

## Step 6: Run End-to-End Tests

Run automated tests to verify the entire pipeline lifecycle:

```bash
c8s dev test run
```

This command:
- Creates PipelineRun resources for each PipelineConfig
- Monitors job execution
- Validates successful completion
- Takes ~30-60 seconds

**Expected output**:
```
Running pipeline tests in cluster 'c8s-dev'...

Testing simple-build...
  ✓ Pipeline created
  ✓ Job started
  ✓ Step 1 completed (5s)
  ✓ Pipeline succeeded (8s)

Testing multi-step...
  ✓ Pipeline created
  ✓ Job started
  ✓ Step 1 completed (3s)
  ✓ Step 2 completed (4s)
  ✓ Pipeline succeeded (10s)

Testing matrix-build...
  ✓ Pipeline created
  ✓ 4 parallel jobs started
  ✓ All matrix jobs completed (12s)
  ✓ Pipeline succeeded (15s)

Test Results:
  Total:   3
  Passed:  3
  Failed:  0
  Duration: 33s

All tests passed ✓
```

### View Pipeline Logs

If you need to debug a pipeline:

```bash
c8s dev test logs --pipeline simple-build
```

Or follow logs in real-time:

```bash
c8s dev test logs --pipeline simple-build --follow
```

## Step 7: Iterative Development

Now that your environment is running, here's the typical development workflow:

### Make Code Changes

Edit operator code, add features, fix bugs, etc.

### Rebuild and Redeploy

```bash
# 1. Rebuild operator image
docker build -t c8s-operator:dev .

# 2. Load into cluster
k3d image import c8s-operator:dev -c c8s-dev

# 3. Restart operator deployment
kubectl rollout restart deployment/c8s-controller-manager -n c8s-system

# 4. Wait for rollout
kubectl rollout status deployment/c8s-controller-manager -n c8s-system

# 5. Rerun tests
c8s dev test run
```

**Time**: ~1-2 minutes total

### One-Line Rebuild Script

Create a helper script `scripts/dev-reload.sh`:

```bash
#!/bin/bash
set -e

echo "Rebuilding operator..."
docker build -t c8s-operator:dev .

echo "Loading into cluster..."
k3d image import c8s-operator:dev -c c8s-dev

echo "Restarting operator..."
kubectl rollout restart deployment/c8s-controller-manager -n c8s-system
kubectl rollout status deployment/c8s-controller-manager -n c8s-system

echo "Running tests..."
c8s dev test run
```

Make it executable:
```bash
chmod +x scripts/dev-reload.sh
```

Then iterate with:
```bash
./scripts/dev-reload.sh
```

## Step 8: Cleanup

When you're done testing, clean up the environment:

### Stop Cluster (Preserves State)

```bash
c8s dev cluster stop
```

Resume later with:
```bash
c8s dev cluster start
```

### Delete Cluster (Complete Cleanup)

```bash
c8s dev cluster delete
```

Or force delete without confirmation:
```bash
c8s dev cluster delete --force
```

## Troubleshooting

### Cluster Creation Fails

**Problem**: `Error: docker daemon not available`

**Solution**: Ensure Docker is running:
```bash
docker ps
# Should not show connection errors
```

---

**Problem**: `Error: cluster 'c8s-dev' already exists`

**Solution**: Delete existing cluster first:
```bash
c8s dev cluster delete --force
c8s dev cluster create
```

---

### Operator Not Starting

**Problem**: Operator pod in `CrashLoopBackOff`

**Solution**: Check operator logs:
```bash
kubectl logs -n c8s-system -l app=c8s-controller-manager
```

Common issues:
- Missing RBAC permissions (check ServiceAccount/Role/RoleBinding)
- Image pull errors (verify image was loaded with `k3d image import`)
- CRD installation failures (check CRDs with `kubectl get crds`)

---

### Pipeline Tests Failing

**Problem**: `Error: operator not deployed`

**Solution**: Deploy operator first:
```bash
c8s dev deploy operator
```

---

**Problem**: Pipeline job fails with image pull errors

**Solution**: Build and load required pipeline step images:
```bash
docker build -t my-step-image:dev ./path/to/dockerfile
k3d image import my-step-image:dev -c c8s-dev
```

Update PipelineConfig to use `imagePullPolicy: IfNotPresent` or `Never`.

---

### Port Conflicts

**Problem**: `Error: port 8080 already in use`

**Solution**: Either:
1. Stop the conflicting service
2. Create cluster with different ports:
   ```bash
   c8s dev cluster create --config custom-config.yaml
   ```

   Custom config with different ports:
   ```yaml
   # custom-config.yaml
   apiVersion: k3d.io/v1alpha5
   kind: Simple
   metadata:
     name: c8s-dev
   ports:
     - port: 9080:80
       nodeFilters:
         - loadbalancer
   ```

---

### Resources Exhausted

**Problem**: Cluster creation slow or nodes not ready

**Solution**: Increase Docker resource limits:
- Open Docker Desktop
- Go to Preferences > Resources
- Increase Memory to 8GB minimum
- Increase CPUs to 4 minimum
- Click "Apply & Restart"

## Advanced Usage

### Multiple Clusters

Create multiple clusters for different test scenarios:

```bash
# Cluster for feature development
c8s dev cluster create feature-test

# Cluster for integration tests
c8s dev cluster create integration-test

# List all clusters
c8s dev cluster list
```

Switch between clusters:
```bash
kubectl config use-context k3d-feature-test
kubectl config use-context k3d-integration-test
```

### Custom Cluster Configuration

Create a custom configuration file:

```yaml
# .c8s/my-cluster.yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: my-cluster
servers: 1
agents: 3  # More agents for testing distribution
image: rancher/k3s:v1.28.15-k3s1
ports:
  - port: 8080:80
    nodeFilters:
      - loadbalancer
registry:
  create:
    name: registry.localhost
    hostPort: 5001  # Different port if 5000 is taken
options:
  k3s:
    extraArgs:
      - arg: --disable=traefik
        nodeFilters:
          - server:*
```

Create cluster from config:
```bash
c8s dev cluster create --config .c8s/my-cluster.yaml
```

### CI/CD Integration

Run local tests in GitHub Actions:

```yaml
name: Test Operator
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install k3d
        run: |
          curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

      - name: Create cluster
        run: |
          c8s dev cluster create --wait

      - name: Build and deploy operator
        run: |
          docker build -t c8s-operator:test .
          k3d image import c8s-operator:test -c c8s-dev
          c8s dev deploy operator

      - name: Run tests
        run: |
          c8s dev deploy samples
          c8s dev test run
```

## Next Steps

Now that you have a working local environment:

1. **Explore the codebase**: `pkg/controllers/` contains reconciliation logic
2. **Read the architecture docs**: `docs/architecture.md`
3. **Modify a controller**: Make changes to `pkg/controllers/pipelineconfig_controller.go`
4. **Write tests**: Add tests to `tests/integration/`
5. **Test your changes**: Run `./scripts/dev-reload.sh`

For more details:
- **CLI Reference**: See `contracts/cli-commands.md`
- **Data Models**: See `data-model.md`
- **Development Guide**: See `docs/development.md`

## Useful Commands

```bash
# Cluster management
c8s dev cluster create              # Create cluster
c8s dev cluster status              # Check cluster status
c8s dev cluster list                # List all clusters
c8s dev cluster delete              # Delete cluster

# Operator management
c8s dev deploy operator             # Deploy operator
c8s dev deploy samples              # Deploy sample pipelines

# Testing
c8s dev test run                    # Run all tests
c8s dev test run --pipeline simple-build  # Test specific pipeline
c8s dev test logs --pipeline simple-build # View pipeline logs

# Kubernetes commands
kubectl get pipelineconfigs         # List pipeline configs
kubectl get pipelineruns            # List pipeline runs
kubectl get jobs                    # List pipeline jobs
kubectl logs -n c8s-system -l app=c8s-controller-manager  # Operator logs

# k3d commands
k3d cluster list                    # List k3d clusters
k3d image import IMAGE -c CLUSTER   # Load image to cluster
k3d cluster stop CLUSTER            # Stop cluster
k3d cluster start CLUSTER           # Start cluster
```

## Getting Help

If you run into issues:

1. **Check cluster status**: `c8s dev cluster status`
2. **Check operator logs**: `kubectl logs -n c8s-system -l app=c8s-controller-manager`
3. **Check pipeline status**: `kubectl describe pipelineconfig PIPELINE_NAME`
4. **Run with verbose output**: Add `--verbose` to any `c8s dev` command
5. **Ask the team**: Post in #c8s-dev Slack channel

---

**Congratulations!** You now have a fully functional local C8S test environment. Happy developing! 🚀
