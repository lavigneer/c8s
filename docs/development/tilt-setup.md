# C8S Local Development with Tilt

This guide explains how to use Tilt for local Kubernetes development on the C8S project.

## Prerequisites

Before setting up Tilt, ensure you have:

- **Go 1.25+**: `go version`
- **Docker**: `docker version` (Docker daemon running)
- **kubectl**: `kubectl version --client`
- **k3d 5.8.3+**: `k3d version`
- **Tilt 0.33.0+**: `tilt version`
- **4+ GB RAM** available for the k3d cluster

### Installation

**macOS (Homebrew)**:
```bash
brew install tilt
brew install k3d
brew install kubectl
```

**Linux (apt/yum)**:
```bash
# Tilt
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

# k3d
wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# kubectl
# Use your package manager or download from kubernetes.io
```

**Verify Installation**:
```bash
tilt version
k3d version
kubectl version --client
```

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/org/c8s.git
cd c8s
```

### 2. Start Development Environment

```bash
# From repository root, start Tilt
tilt up

# First time may take 2-5 minutes as it creates the cluster and builds images
```

Tilt will:
1. Create a local k3d cluster named `c8s-dev` (if not exists)
2. Install CRDs and RBAC configuration
3. Build Docker images for controller, api-server, and webhook
4. Deploy components to the cluster
5. Open Tilt dashboard at http://localhost:10350

### 3. Access the Development Environment

**Tilt Dashboard**: http://localhost:10350
- Real-time logs for all components
- Component status and build history
- Manual triggers for validation and other tasks

**Component Endpoints**:
- Controller (debug): http://localhost:6060/debug/pprof
- API Server: http://localhost:8080
- Webhook: https://localhost:9443

### 4. Edit and Iterate

```bash
# In your editor, modify any Go file in cmd/ or pkg/
# For example: cmd/controller/main.go

# Tilt automatically detects changes and:
# 1. Rebuilds the affected component
# 2. Updates the container image
# 3. Redeploys the pod

# Watch Tilt dashboard for rebuild progress
```

### 5. Stop Development

```bash
# Stop Tilt and clean up resources
tilt down

# This removes the k3d cluster and all deployed resources
```

## Development Workflows

### Hot Reload Development Cycle

1. **Edit code**: Modify a Go file
2. **Auto rebuild**: Tilt detects change and rebuilds (~10-30 seconds)
3. **Auto deploy**: Component pod is restarted with new image
4. **Verify**: Check logs in Tilt dashboard or via `kubectl logs`

**Example**: Modify controller logging
```bash
# Edit pkg/controller/controller.go
vim pkg/controller/controller.go

# Save file
# Watch Tilt dashboard - controller will rebuild and redeploy
# Click controller resource in Tilt UI to view logs
```

### Pipeline Definition Testing

1. **Create pipeline YAML**:
```yaml
# cat > test-pipeline.yaml <<EOF
version: v1alpha1
name: test-hello
steps:
  - name: hello
    image: alpine:latest
    commands:
      - echo "Hello from C8S!"
    resources:
      cpu: 100m
      memory: 128Mi
EOF
```

2. **Apply to cluster**:
```bash
kubectl apply -f test-pipeline.yaml -n c8s-system
```

3. **Monitor execution**:
```bash
# View pipeline status
kubectl get pipelinerun -n c8s-system

# View pipeline logs
kubectl logs -n c8s-system -l pipeline=test-hello --all-containers=true -f
```

4. **Validate and iterate**:
- Modify pipeline YAML
- Reapply with `kubectl apply -f`
- Check logs and results
- Repeat until pipeline works as expected

### Debugging Multi-Component Flows

When testing interactions between components (webhook → api-server → controller → job):

1. **Open Tilt Dashboard**: http://localhost:10350

2. **Trigger action** (e.g., create PipelineRun):
```bash
kubectl apply -f config/samples/simple-pipeline.yaml -n c8s-system
```

3. **View unified logs**:
- Webhook logs: Shows incoming webhook request
- API Server logs: Shows request processing
- Controller logs: Shows job creation and status updates

4. **Filter and search**:
- Use Tilt dashboard text search to find specific log entries
- Filter by component to isolate issues
- View error messages with full context

### Sample Pipeline Management

**Deploy samples**:
```bash
# Deploy sample pipelines from config/samples/
kubectl apply -f config/samples/ -n c8s-system
```

**List samples**:
```bash
kubectl get pipelines -n c8s-system
kubectl get pipelineruns -n c8s-system
```

**Delete samples**:
```bash
kubectl delete -f config/samples/ -n c8s-system
```

## Advanced Usage

### Configuration

The Tiltfile supports several configuration options via command-line flags:

```bash
# Enable sample pipelines deployment
tilt up -- --with_samples=true

# Use verbose logging
tilt up -- --verbose_logs=true

# Change Kubernetes namespace
tilt up -- --k8s_namespace=my-c8s

# Customize image registry prefix
tilt up -- --image_registry=mycustom
```

### Accessing Cluster

Use standard kubectl commands:

```bash
# Set context if Tilt didn't do it automatically
kubectl config use-context k3d-c8s-dev

# View all resources
kubectl get all -n c8s-system

# Get detailed resource info
kubectl describe pod <pod-name> -n c8s-system

# Execute commands in pod
kubectl exec -it <pod-name> -n c8s-system -- /bin/sh
```

### Viewing Logs

**Via Tilt Dashboard**: Click resource name to view logs

**Via kubectl**:
```bash
# Controller logs
kubectl logs -f deployment/c8s-controller -n c8s-system

# API Server logs
kubectl logs -f deployment/c8s-api-server -n c8s-system

# Webhook logs
kubectl logs -f deployment/c8s-webhook -n c8s-system

# Stream logs from all pods
kubectl logs -f -n c8s-system --all-containers=true
```

### Manual Trigger Tasks

Tilt dashboard provides manual trigger buttons for:

- **Cluster Status**: Run `kubectl cluster-info` and list pods/services
- **Pipeline Validator**: Test pipeline validation framework
- **CRD Installation**: Re-apply CRDs from deploy/crds.yaml
- **RBAC Installation**: Re-apply RBAC from deploy/rbac.yaml

### Resource Constraints

For machines with limited resources:

```bash
# Edit Tiltfile to reduce resource requests
# Look for k8s_resource definitions and modify resource requests
# Or use Tilt config:
tilt up -- --resource_limits='{"controller": "100m/256Mi", "api-server": "200m/512Mi"}'
```

## Troubleshooting

### Issue: "k3d cluster not found"

**Solution**: Tilt should create the cluster automatically. If not:
```bash
# Manually create cluster
k3d cluster create c8s-dev --registry-create=registry:5000 -p "8080:80@loadbalancer"

# Then start Tilt
tilt up
```

### Issue: "Cannot connect to Docker daemon"

**Solution**: Ensure Docker is running:
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### Issue: "Build fails with syntax error"

**Solution**:
- Fix the Go syntax error in the file shown in error message
- Save the file
- Tilt automatically rebuilds when syntax is fixed
- View logs in Tilt dashboard to confirm success

### Issue: "Pod stuck in CrashLoopBackOff"

**Solution**: Check logs for the failure reason:
```bash
kubectl logs <pod-name> -n c8s-system --previous
# Or in Tilt dashboard, click the resource to view logs
```

### Issue: "Port already in use (8080, 6060, 9443)"

**Solution**: Change Tilt configuration or kill process using port:
```bash
# Find and kill process on port 8080
lsof -i :8080 | grep -v COMMAND | awk '{print $2}' | xargs kill -9

# Or change port forwarding in Tiltfile
```

### Issue: "Out of memory or CPU resources"

**Solution**:
- Increase Docker/k3d resource allocation
- Reduce resource requests in Tiltfile
- Close other applications using resources
- Use smaller test pipelines

## Performance Tips

1. **Faster Rebuilds**:
   - Only modify files in `cmd/` or `pkg/` - they trigger rebuilds
   - Modifying `*.md` or specs files won't trigger builds

2. **Quicker Iteration**:
   - Keep Tilt running while making changes
   - Use Tilt logs to verify changes
   - Avoid stopping/starting Tilt between iterations

3. **Better Debugging**:
   - Use Tilt dashboard text search to find logs
   - Monitor multiple component logs simultaneously
   - Check error messages for specific failure details

4. **Effective Testing**:
   - Test one pipeline/feature at a time
   - Keep sample pipelines simple
   - Verify logs show expected behavior

## Common Commands

```bash
# Start development
tilt up

# Stop development
tilt down

# Restart a component
tilt trigger <resource-name>

# View specific logs
tilt logs controller
tilt logs api-server
tilt logs webhook

# Enter Tilt UI
tilt open

# Check status
tilt status

# Get help
tilt help
```

## Next Steps

- **For C8S Contributors**: See [CONTRIBUTING.md](../CONTRIBUTING.md)
- **For Pipeline Writing**: See [Pipeline Configuration](../README.md#pipeline-configuration-schema)
- **For API Details**: See API docs at http://localhost:8080/api/v1/docs (when running)
- **For More Tilt Features**: Visit https://docs.tilt.dev

## Additional Resources

- **Tilt Documentation**: https://docs.tilt.dev
- **k3d Documentation**: https://k3d.io
- **Kubernetes Documentation**: https://kubernetes.io/docs
- **Go Development**: https://golang.org/doc
- **C8S Repository**: https://github.com/org/c8s

---

**Questions or Issues?** Open an issue at https://github.com/org/c8s/issues
