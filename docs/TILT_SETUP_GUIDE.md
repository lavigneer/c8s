# C8S Tilt Development Setup Guide

This guide explains how to set up and run the C8S development environment using Tilt with the HTMX-based dashboard.

## Prerequisites

- **Tilt**: https://docs.tilt.dev/install.html
- **k3d**: https://k3d.io/#installation
- **kubectl**: https://kubernetes.io/docs/tasks/tools/
- **Docker**: https://docs.docker.com/get-docker/

## Quick Start

### 1. Create Development Cluster (One-time)

```bash
k3d cluster create c8s-dev \
  --registry-create=registry:5000 \
  -p "8080:80@loadbalancer" \
  --servers 1 \
  --agents 2
```

This creates a local Kubernetes cluster with:
- Local registry at `localhost:5000`
- Port 8080 forwarding for load balancer

### 2. Start Tilt

```bash
tilt up
```

This will:
1. Create the `c8s-system` namespace
2. Install CRDs and RBAC resources
3. Build all components (controller, webhook, api-server)
4. Deploy all services to the cluster
5. Show you the Tilt dashboard

### 3. Access the Services

- **Tilt Dashboard**: http://localhost:10350
- **C8S Dashboard**: http://localhost:8080
- **Controller Metrics**: http://localhost:6060

## Component Details

### Controller
- **Purpose**: Manages C8S Kubernetes resources (PipelineConfigs, PipelineRuns)
- **Access**: Metrics at http://localhost:6060/debug/pprof
- **Logs**: View in Tilt dashboard or `tilt logs controller`

### API Server (HTMX Dashboard)
- **Purpose**: Web UI for pipeline management
- **Access**: http://localhost:8080
- **Routes**:
  - `/dashboard` - Main pipeline list
  - `/dashboard/projects` - Project management
  - `/login` - Authentication page
- **Logs**: View in Tilt dashboard or `tilt logs c8s-api-server`

### Webhook
- **Purpose**: Receives GitHub/GitLab webhook events
- **Access**: https://localhost:9443
- **Logs**: View in Tilt dashboard or `tilt logs c8s-webhook`

## Development Workflow

### Hot Reload

The Tilt setup supports automatic rebuilding and redeployment:

1. **Go Code Changes**
   - Edit any `.go` file in `cmd/` or `pkg/`
   - Tilt automatically rebuilds the binary
   - Container automatically restarts
   - Logs update in the Tilt dashboard

2. **Template Changes**
   - Edit any `.html` file in `cmd/api-server/templates/`
   - Changes are copied into the running container
   - Refresh browser to see updates

3. **Static Assets**
   - Edit CSS in `cmd/api-server/static/css/`
   - Edit JavaScript in `cmd/api-server/static/js/`
   - Clear browser cache or use Cmd+Shift+R (Ctrl+Shift+R) to reload

### Common Development Tasks

#### View Logs
```bash
# View specific service logs
tilt logs c8s-api-server
tilt logs controller
tilt logs c8s-webhook

# Stream logs in real-time
tilt logs c8s-api-server -f
```

#### Restart a Service
```bash
# In Tilt dashboard: Click the service and click "Restart"
# Or via CLI: tilt trigger <service-name>
tilt trigger c8s-api-server
```

#### Check Service Status
```bash
kubectl get pods -n c8s-system
kubectl get deployments -n c8s-system
kubectl describe pod <pod-name> -n c8s-system
```

#### Access Pod Directly
```bash
# Get a shell in the api-server pod
kubectl exec -it <pod-name> -n c8s-system -- /bin/sh

# View pod logs
kubectl logs <pod-name> -n c8s-system
```

## Troubleshooting

### Dashboard Not Loading

1. Check service status:
   ```bash
   kubectl get svc c8s-api-server -n c8s-system
   ```

2. Check pod logs:
   ```bash
   kubectl logs -l app=c8s-api-server -n c8s-system
   ```

3. Verify port forwarding:
   ```bash
   # In Tilt dashboard, check port_forwards shows 8080:8080
   ```

### Rebuilds Not Triggering

1. Verify Tilt is watching files:
   ```bash
   tilt status
   ```

2. Check for syntax errors in Go code:
   ```bash
   go build ./cmd/api-server
   ```

3. Force a rebuild:
   ```bash
   tilt trigger c8s-api-server
   ```

### Out of Disk Space

If you run out of space with Docker:
```bash
docker system prune -a
docker volume prune
```

## Configuration

### Disable Sample Pipelines

Edit `Tiltfile` and set:
```python
with_samples = False
```

### Increase Verbosity

Edit `Tiltfile` and set:
```python
verbose_logs = True
```

### Change Namespace

Edit `Tiltfile` and set:
```python
k8s_namespace = 'your-namespace'
```

## Production Deployment

When ready to deploy to production:

1. Build and push images:
   ```bash
   docker build --target api-server -t registry/c8s-api-server:v1.0.0 .
   docker push registry/c8s-api-server:v1.0.0
   ```

2. Update deployment manifests to use your image tags

3. Deploy to production cluster:
   ```bash
   kubectl apply -f deploy/ -n c8s-system
   ```

## Cleanup

### Stop Tilt
```bash
tilt down
```

### Delete Cluster
```bash
k3d cluster delete c8s-dev
```

### Remove Local Images
```bash
docker image rm c8s-controller:latest c8s-webhook:latest c8s-api-server:latest
```

## Frontend Development Specific

For developing the HTMX dashboard:

### Edit API Routes
File: `cmd/api-server/handlers/*.go`

### Edit Templates
Directory: `cmd/api-server/templates/`

### Edit Styles
File: `cmd/api-server/static/css/dashboard.css`

### Edit JavaScript
File: `cmd/api-server/static/js/keyboard_shortcuts.js`

### Test Changes Locally
```bash
# Get shell in running container
kubectl exec -it $(kubectl get pods -l app=c8s-api-server -n c8s-system -o jsonpath='{.items[0].metadata.name}') -n c8s-system -- /bin/sh

# Inside container, check templates are mounted
ls -la /app/templates/
```

## Performance Tips

1. **Reduce Docker Build Time**:
   - Tilt caches Docker layers
   - Only modified files trigger rebuilds
   - Ignore unnecessary files in Tiltfile `ignore` list

2. **Speed Up Kubernetes Operations**:
   - Local k3d cluster is much faster than cloud
   - Single controller and webhook for dev is sufficient
   - Monitor resource usage: `kubectl top nodes`

3. **Efficient Logging**:
   - Use Tilt dashboard for real-time logs
   - Filter by service name in logs view
   - Avoid watching too many services at once

## Getting Help

- **Tilt Documentation**: https://docs.tilt.dev/
- **k3d Documentation**: https://k3d.io/
- **C8S Dashboard Docs**: See `DASHBOARD_README.md`
- **Kubernetes Basics**: https://kubernetes.io/docs/

---

Happy developing! 🚀
