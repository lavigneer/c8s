# Tilt Development Workflow Guide

Complete guide to using Tilt for C8S local development with live reload.

## What is Tilt?

Tilt is a local Kubernetes development tool that provides:
- **Live Reload**: Code changes sync to pods in ~5 seconds
- **Smart Rebuilds**: Only rebuilds what changed
- **Unified UI**: All logs, builds, and status in one place
- **Port Forwarding**: Automatic local access to services
- **Resource Management**: Easy start/stop of individual components

## Quick Start

```bash
tilt up    # Start development environment
# Press spacebar to open Tilt UI in browser
```

That's it! Tilt handles everything else.

## What Happens When You Run `tilt up`

```
1. Tilt reads Tiltfile (like a Makefile for Kubernetes)
   ├─ Detects your GitHub repository
   ├─ Finds kind cluster or creates one
   └─ Loads configuration

2. Builds Components (parallel)
   ├─ Compiles controller binary locally (Go)
   ├─ Compiles webhook binary locally (Go)
   ├─ Compiles api-server binary locally (Go)
   ├─ Builds Docker images (fast with live_update)
   └─ Pushes images to kind cluster

3. Deploys Kubernetes Resources
   ├─ Installs cert-manager (for TLS certificates)
   ├─ Deploys MinIO (local S3 storage)
   ├─ Deploys Helm chart (C8S components)
   └─ Creates sample PipelineConfig

4. Sets Up Developer Tools
   ├─ Port forwards (API server, webhook, MinIO)
   ├─ nginx reverse proxy (optional, for dog-fooding)
   ├─ ngrok tunnel (optional, for GitHub webhooks)
   └─ Watches source files for changes

5. Opens Tilt UI
   └─ Browser opens to http://localhost:10350
```

**Expected startup time**: 2-3 minutes (first run), 1 minute (subsequent runs)

## Tilt UI Overview

### Main Dashboard

```
┌─────────────────────────────────────────────────────────┐
│  Tilt - C8S Development                                │
├─────────────────────────────────────────────────────────┤
│  Resources              Status      Update  Build Time  │
│  ────────────────────────────────────────────────────── │
│  ✓ cert-manager        Running     Auto    1m 23s      │
│  ✓ minio               Running     Auto    45s         │
│  ✓ c8s-controller      Running     Auto    32s         │
│  ✓ c8s-api-server      Running     Auto    28s         │
│  ✓ c8s-webhook         Running     Auto    30s         │
│  ✓ deploy-pipelineconfig Running  Manual   2s          │
│  ────────────────────────────────────────────────────── │
│  [All resources healthy]                                │
└─────────────────────────────────────────────────────────┘
```

### Resource Details

Click any resource to see:
- **Overview**: Current status, pod info, endpoints
- **Logs**: Real-time log streaming
- **Build History**: Previous builds and their status
- **K8s Resources**: Pods, services, deployments

## Making Code Changes

### Scenario 1: Edit Go Code (Hot Reload)

1. **Edit any Go file**:
   ```bash
   vim pkg/controller/reconciler.go
   # Add: log.Info("Processing pipeline run", "name", req.Name)
   ```

2. **Save the file**

3. **Watch Tilt UI**:
   ```
   c8s-controller
   ├─ Detected change: pkg/controller/reconciler.go
   ├─ Building locally... (15s)
   ├─ Syncing binary to pod... (2s)
   ├─ Restarting pod... (3s)
   └─ ✓ Ready (total: 20s)
   ```

4. **View logs** (in Tilt UI or terminal):
   ```bash
   tilt logs c8s-controller
   # See your new log line appear
   ```

### Scenario 2: Edit Templates/Static Files (Instant Reload)

1. **Edit template**:
   ```bash
   vim cmd/api-server/templates/dashboard.html
   # Make changes
   ```

2. **Save the file**

3. **Changes sync instantly** (~2 seconds):
   ```
   c8s-api-server
   ├─ Detected change: cmd/api-server/templates/dashboard.html
   ├─ Syncing file to pod... (1s)
   └─ ✓ Ready (no restart needed)
   ```

4. **Refresh browser** - see changes immediately

### Scenario 3: Edit CRDs or Helm Chart

1. **Edit CRD**:
   ```bash
   vim config/crd/bases/c8s.dev_pipelineruns.yaml
   ```

2. **Save and regenerate**:
   ```bash
   make manifests  # Regenerate CRDs
   ```

3. **Tilt auto-deploys**:
   ```
   c8s (Helm chart)
   ├─ Detected change: config/crd/bases/
   ├─ Updating Helm release... (5s)
   └─ ✓ Ready
   ```

## Tilt Commands

### Basic Operations

```bash
# Start Tilt
tilt up

# Start in background (no UI)
tilt up --hud=false

# Stop Tilt (keeps cluster)
ctrl+c  # or tilt down

# Stop Tilt and delete resources
tilt down

# View status
tilt status

# Get resource info
tilt get resources
```

### Viewing Logs

```bash
# All logs (mixed)
tilt logs

# Specific resource
tilt logs c8s-controller
tilt logs c8s-api-server

# Follow logs
tilt logs c8s-controller -f

# Last N lines
tilt logs c8s-controller --tail 100
```

### Triggering Rebuilds

```bash
# Trigger specific resource rebuild
tilt trigger c8s-controller

# Trigger all resources
tilt trigger --all

# Restart pod without rebuild
kubectl rollout restart deployment/c8s-controller -n c8s-system
```

## Resource Dependency Graph

Understanding the startup order:

```
cert-manager (Helm dependency)
    │
    ├─→ c8s (Helm chart)
    │       ├─→ c8s-controller
    │       ├─→ c8s-api-server
    │       └─→ c8s-webhook
    │
minio (S3 storage)
    │
    └─→ c8s (Helm chart)

deploy-pipelineconfig
    └─→ Depends on: c8s (Helm chart)

Port Forwards
    ├─→ c8s-api-server-port-forward
    ├─→ c8s-webhook-port-forward
    └─→ minio-port-forward

Optional Dog-fooding
    ├─→ c8s-nginx-proxy (depends on port forwards)
    └─→ c8s-ngrok-tunnel (depends on nginx)
```

## Debugging with Tilt

### Check Resource Status

```bash
# In Tilt UI
Click resource → Overview tab → View pod status

# In terminal
kubectl get pods -n c8s-system
kubectl describe pod <pod-name> -n c8s-system
```

### View Build Errors

```bash
# In Tilt UI
Click resource → Build History → Click failed build

# Check compile errors
tilt logs <resource> | grep error
```

### Exec into Pod

```bash
# Find pod name
kubectl get pods -n c8s-system

# Exec into pod
kubectl exec -it <pod-name> -n c8s-system -- sh

# Inside pod
ps aux          # Check running processes
ls /app        # Check binary exists
cat /app/templates/dashboard.html  # Check synced files
```

### Reset Everything

```bash
# Stop Tilt
tilt down

# Delete cluster
kind delete cluster --name c8s

# Restart fresh
tilt up  # Tilt will create new cluster
```

## Advanced Features

### Live Update Configuration

Tilt uses `live_update` to sync changes without full rebuilds:

**For Go binaries**:
1. Compile locally
2. Sync binary to pod
3. Restart process

**For templates/static files**:
1. Sync files to pod
2. No restart needed

**Configuration** (in Tiltfile):
```python
docker_build_with_restart(
  ref='c8s-api-server',
  context='.',
  dockerfile='Dockerfile.tilt',
  target='api-server',
  live_update=[
    sync('./bin/api-server', '/app/bin/api-server'),        # Sync binary
    sync('./cmd/api-server/templates', '/app/templates'),  # Sync templates
    sync('./cmd/api-server/static', '/app/static'),        # Sync static
  ],
)
```

### Port Forwarding

Tilt automatically port forwards these services:

| Service | Local Port | Pod Port | URL |
|---------|------------|----------|-----|
| API Server | 8000 | 8000 | http://localhost:8000 |
| Webhook | 8080 | 8080 | http://localhost:8080 |
| MinIO Console | 9001 | 9001 | http://localhost:9001 |
| MinIO API | 9000 | 9000 | http://localhost:9000 |

### Dog-fooding Workflow (Optional)

To use C8S to test itself via GitHub webhooks:

1. **Set up ngrok** (one-time):
   ```bash
   ngrok config add-authtoken <your-token>
   ```

2. **Start Tilt with ngrok**:
   ```bash
   tilt up  # ngrok automatically starts if configured
   ```

3. **Get ngrok URL**:
   ```bash
   tilt logs c8s-ngrok-tunnel | grep "started tunnel"
   # Example: https://abc123.ngrok.io
   ```

4. **Configure GitHub webhook**:
   - GitHub Repo → Settings → Webhooks → Add webhook
   - URL: `https://abc123.ngrok.io/api/v1/webhook`
   - Secret: (from webhook secret)
   - Events: Push, Pull Request

Now GitHub events trigger C8S pipelines locally!

See [c8s-dogfooding.md](./c8s-dogfooding.md) for details.

## Tilt Configuration Files

- **Tiltfile**: Main configuration (like Makefile)
- **tilt/config/c8s-values.yaml**: Helm values overrides
- **tilt/config/nginx.conf**: nginx reverse proxy config
- **tilt/config/ngrok-config.yml**: ngrok tunnel config
- **.tilt-ci.yml**: Tilt CI configuration

## Performance Tips

1. **Use live_update**: Don't disable it - 5s vs 30s rebuilds
2. **Filter logs**: Use search in Tilt UI to find relevant logs
3. **Disable unused resources**: Comment out in Tiltfile if not needed
4. **Increase Docker resources**: 4 CPU, 8GB RAM recommended
5. **Use local registry**: Faster image pulls (Tilt does this automatically with kind)

## Common Issues

### "Tilt won't start"
```bash
# Check Docker is running
docker ps

# Check ports are free
lsof -i :10350  # Tilt UI port
lsof -i :8000   # API server port

# Delete .tilt directory
rm -rf .tilt
```

### "Resources stuck in 'Pending'"
```bash
# Check cluster has resources
kubectl top nodes
kubectl describe pod <pod-name> -n c8s-system

# Check image pull
kubectl get events -n c8s-system --sort-by='.lastTimestamp'
```

### "Live update not working"
```bash
# Check file is being watched
tilt get uiresources

# Manually trigger rebuild
tilt trigger <resource>

# Check sync configuration in Tiltfile
```

### "Changes not appearing"
```bash
# Hard browser refresh
Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)

# Check file was synced
tilt logs <resource> | grep sync

# Exec into pod and verify
kubectl exec -it <pod> -n c8s-system -- ls /app/templates
```

## Best Practices

1. **Always use Tilt UI**: Don't just watch terminal - UI shows full picture
2. **Check logs first**: Most issues show up in logs
3. **One resource at a time**: Fix one resource before investigating another
4. **Commit before experimenting**: Easy to reset if Tilt state gets weird
5. **Use tilt down**: Clean shutdown prevents resource leaks

## Resources

- [Tiltfile API Reference](https://docs.tilt.dev/api.html)
- [Live Update Guide](https://docs.tilt.dev/live_update_tutorial.html)
- [Tilt Best Practices](https://docs.tilt.dev/best_practices.html)
- [Debugging Guide](https://docs.tilt.dev/debug_faq.html)

---

**Ready to code?** Run `tilt up` and start making changes!
