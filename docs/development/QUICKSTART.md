# Development Quick Start

Get C8S development environment up and running in **5 minutes**.

## Prerequisites Checklist

Before you begin, ensure you have:

- [ ] **Docker** (27.3.1+) - [Install Docker](https://docs.docker.com/get-docker/)
- [ ] **Git** - For cloning the repository
- [ ] **10 GB free disk space** - For Docker images and dependencies

**That's it!** Devbox will handle all other tools (Go, kubectl, Tilt, etc.)

## Quick Setup (5 Minutes)

### Step 1: Install Devbox (1 minute)

```bash
curl -fsSL https://get.jetify.com/devbox | bash
```

### Step 2: Clone Repository (1 minute)

```bash
git clone https://github.com/lavigneer/c8s.git
cd c8s
```

### Step 3: Enter Development Environment (1 minute)

```bash
devbox shell
```

This automatically installs and configures:
- Go 1.25
- kubectl, kind, Tilt
- golangci-lint
- Node.js (for E2E tests)
- All other development tools

### Step 4: Start Development Environment (2 minutes)

```bash
make dev  # or: tilt up
```

**What happens next:**
1. Tilt creates a local Kubernetes cluster (kind)
2. Builds Docker images for all components
3. Deploys C8S with Helm chart
4. Sets up live reload for code changes
5. Exposes services on localhost ports

### Step 5: Verify It's Working

Open http://localhost:10350 in your browser to see the Tilt UI.

You should see:
- ✅ c8s-controller (green)
- ✅ c8s-api-server (green)
- ✅ c8s-webhook (green)
- ✅ All other resources healthy

## Access Points

Once Tilt is running:

| Service | URL | Purpose |
|---------|-----|---------|
| **Tilt UI** | http://localhost:10350 | Monitor resources, view logs |
| **C8S API Server** | http://localhost:8000 | REST API for pipeline management |
| **C8S Dashboard** | http://localhost:8000 | Web UI for pipelines |
| **MinIO Console** | http://localhost:9001 | Local S3 storage (user: minioadmin, pass: minioadmin) |

## Your First Code Change

1. **Edit a file**:
   ```bash
   vim pkg/controller/reconciler.go
   # Make any change (e.g., add a log line)
   ```

2. **Watch Tilt rebuild automatically**:
   - Open Tilt UI: http://localhost:10350
   - See "c8s-controller" rebuild (~30 seconds)
   - New code is live!

3. **View logs**:
   ```bash
   tilt logs c8s-controller  # In Tilt UI or terminal
   ```

## Common Commands

```bash
# Development
make dev          # Start Tilt development environment
make build        # Build all binaries locally
make test         # Run unit + integration tests
make lint         # Run golangci-lint

# Testing
make test-unit           # Unit tests only
make test-integration    # Integration tests (requires envtest)
make test-e2e            # E2E tests with Playwright
make test-all            # All tests

# Code Generation
make generate     # Generate DeepCopy methods
make manifests    # Generate CRD manifests

# Tilt Operations
tilt up           # Start Tilt
tilt down         # Stop Tilt and cleanup
tilt logs <resource>  # View logs for specific resource
ctrl+c            # Stop Tilt (in terminal)
```

## Troubleshooting

### "devbox: command not found"
Devbox wasn't installed or isn't in PATH:
```bash
# Re-run installation
curl -fsSL https://get.jetify.com/devbox | bash

# Restart your terminal
```

### "Docker is not running"
Start Docker Desktop or the Docker daemon:
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### "Tilt fails to start"
Check Docker has enough resources:
- Docker Desktop → Preferences → Resources
- Recommended: 4 CPU, 8GB RAM

### "Port already in use"
Another process is using required ports:
```bash
# Find what's using port 8000
lsof -i :8000

# Kill the process or use different port
```

### "kubectl: command not found" inside devbox
Exit and re-enter devbox shell:
```bash
exit
devbox shell
```

## Next Steps

Now that your environment is running:

1. **Read the codebase**: Start with [pkg/README.md](../../pkg/README.md)
2. **Understand the architecture**: See [Development Guide](./development.md)
3. **Write tests**: See [Local Testing Guide](./local-testing.md)
4. **Make changes**: See [Contributing Guidelines](../CONTRIBUTING.md)
5. **Understand Tilt**: See [Tilt Setup Guide](./tilt-setup.md)

## Development Workflow

```
┌─────────────────────────────────────────┐
│  1. Edit code in your editor           │
│     (VSCode, Vim, etc.)                 │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  2. Tilt detects change automatically   │
│     • Compiles Go binary                │
│     • Syncs to Docker container         │
│     • Restarts affected pods            │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  3. View results in Tilt UI             │
│     • Build logs                        │
│     • Application logs                  │
│     • Resource status                   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  4. Test changes                        │
│     • Use Dashboard (localhost:8000)    │
│     • Make API requests                 │
│     • Run E2E tests                     │
└─────────────────────────────────────────┘
```

## Tips for Productivity

- **Use Tilt UI tabs**: Each resource has logs, build history, and pod info
- **Filter logs**: Click "Search" in Tilt UI to filter logs
- **Trigger manual rebuild**: Click circular arrow icon next to resource
- **Watch test output**: Run `make test` in separate terminal while coding
- **Use live_update**: Changes sync in ~5 seconds vs ~30 second full rebuild

## Getting Help

- **Documentation**: Check [docs/](../)
- **Troubleshooting**: See [Troubleshooting Guide](../guides/troubleshooting.md)
- **Code structure**: See [pkg/README.md](../../pkg/README.md)
- **Tilt issues**: See [Tilt Setup Guide](./tilt-setup.md)
- **Devbox issues**: See [devbox-README.md](../../devbox-README.md)

## Clean Up

When you're done developing:

```bash
# Stop Tilt
tilt down

# Exit devbox shell
exit

# Optional: Delete the kind cluster
kind delete cluster --name c8s
```

---

**Ready to contribute?** See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.
