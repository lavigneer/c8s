# C8S Tilt Quick Start

## One-Time Setup

```bash
# Create local Kubernetes cluster
k3d cluster create c8s-dev \
  --registry-create=registry:5000 \
  -p "8080:80@loadbalancer" \
  --servers 1 \
  --agents 2
```

## Start Development Environment

```bash
# Start all services (controller, webhook, api-server)
tilt up

# Open Tilt dashboard (automatically shown in browser)
# Or manually: http://localhost:10350
```

## Access Services

| Service | URL | Purpose |
|---------|-----|---------|
| **Dashboard** | http://localhost:8080 | HTMX web UI for pipeline management |
| **Tilt UI** | http://localhost:10350 | Monitor builds and logs |
| **Metrics** | http://localhost:6060 | Controller pprof metrics |

## Development Workflow

1. **Edit Go code** (`cmd/api-server/handlers/*.go`)
   - Tilt auto-rebuilds and restarts
   - Check Tilt UI for build status

2. **Edit templates** (`cmd/api-server/templates/**/*.html`)
   - Changes live-update in container
   - Refresh browser to see changes

3. **Edit styles** (`cmd/api-server/static/css/dashboard.css`)
   - Changes live-update in container
   - Ctrl+Shift+R (Cmd+Shift+R on Mac) to reload in browser

4. **View logs**
   - Use Tilt dashboard (recommended)
   - Or: `tilt logs c8s-api-server`

## Common Commands

```bash
# View specific service logs
tilt logs c8s-api-server
tilt logs controller
tilt logs c8s-webhook

# Restart a service
tilt trigger c8s-api-server

# Check pod status
kubectl get pods -n c8s-system

# Get shell in api-server pod
kubectl exec -it $(kubectl get pods -l app=c8s-api-server -n c8s-system -o jsonpath='{.items[0].metadata.name}') -n c8s-system -- /bin/sh

# Stop Tilt
tilt down

# Delete cluster
k3d cluster delete c8s-dev
```

## What's Running

- **c8s-controller**: Manages pipeline CRDs and execution
- **c8s-webhook**: Receives GitHub/GitLab webhook events
- **c8s-api-server**: HTMX dashboard (this is the new frontend)

## Next Steps

1. Run `tilt up`
2. Open http://localhost:8080
3. Start editing code in `cmd/api-server/`
4. Watch auto-reload in action!

For detailed documentation, see `docs/TILT_SETUP_GUIDE.md`
