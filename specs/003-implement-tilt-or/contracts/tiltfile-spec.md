# Tiltfile Configuration Specification

**Status**: API Contract
**Version**: 1.0
**Date**: 2025-10-22

## Overview

This document defines the configuration interface, expected behavior, and customization points for the C8S Tiltfile used in local Kubernetes development.

## Configuration Interface

### Command-Line Flags

Tilt configuration is specified via command-line flags when starting Tilt:

```bash
tilt up -- --flag=value
```

#### Flag: `with_samples` (boolean)

**Type**: Boolean
**Default**: `true`
**Purpose**: Controls whether sample pipeline definitions are deployed during startup

**Valid Values**:
- `true`: Deploy sample pipelines from `config/samples/`
- `false`: Skip sample deployment

**Example Usage**:
```bash
tilt up -- --with_samples=false
```

**Implementation**:
- Checked in Tiltfile via `config['with_samples']`
- Controls conditional `k8s_yaml()` call for samples
- Can be toggled via Tilt UI (Settings → Update Config)

#### Flag: `verbose_logs` (boolean)

**Type**: Boolean
**Default**: `false`
**Purpose**: Enables verbose logging output from all components

**Valid Values**:
- `true`: Enable debug-level logs
- `false`: Standard info-level logs

**Example Usage**:
```bash
tilt up -- --verbose_logs=true
```

**Implementation**:
- Sets log level for controller, api-server, webhook
- Passes `--log-level=debug` to component deployments
- Increases log volume but aids in troubleshooting

#### Flag: `k8s_namespace` (string)

**Type**: String
**Default**: `"c8s-system"`
**Purpose**: Kubernetes namespace where C8S components are deployed

**Valid Values**:
- Any valid Kubernetes namespace name (3-63 lowercase alphanumeric and hyphens)
- Examples: "c8s-system", "c8s-dev", "integration"

**Example Usage**:
```bash
tilt up -- --k8s_namespace=c8s-dev
```

**Implementation**:
- Creates namespace if it doesn't exist (idempotent)
- All component deployments use this namespace
- Must be consistent with `deploy/` manifest files

#### Flag: `image_registry` (string)

**Type**: String
**Default**: `"c8s-dev"`
**Purpose**: Docker image registry prefix for locally built images

**Valid Values**:
- Registry name without trailing slash
- Examples: "c8s-dev", "localhost:5000", "myregistry.azurecr.io"

**Example Usage**:
```bash
tilt up -- --image_registry=localhost:5000
```

**Implementation**:
- Prefixes all component images: `{registry}/c8s-controller`, etc.
- k3d's built-in registry used by default: "registry:5000"
- Custom registry requires appropriate Docker credentials

## File Structure

### Required Files

```
/Users/elavigne/workspace/c8s/
├── Tiltfile                    # Main configuration (must exist)
├── Dockerfile                  # Multi-stage build for components
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build targets
├── deploy/
│   ├── crds.yaml               # PipelineConfig CRD definitions
│   ├── rbac.yaml               # RBAC configuration
│   └── install.yaml            # Component manifests
└── config/
    └── samples/                # Sample pipeline YAML files
        ├── simple-pipeline.yaml
        └── ...
```

### Optional Files

```
├── tilt/
│   └── extensions/             # Custom Tilt extensions
├── docs/
│   └── tilt-setup.md          # Development guide
└── .tiltrc                      # Local Tilt settings
```

## Tiltfile API Contract

### Cluster Configuration

**Required Behavior**:
1. Cluster name: `c8s-dev` (hardcoded for consistency)
2. Auto-create if doesn't exist: Yes
3. Registry creation: k3d built-in registry on `localhost:5000`
4. Port forwarding: 8080:80 for ingress (loadbalancer)

**Testing**:
```bash
k3d cluster list | grep c8s-dev  # Should exist after tilt up
```

### Image Builds

**Build Process for Each Component**:
1. Dockerfile target matching component name (controller, api-server, webhook)
2. Watch directories: `cmd/{component}/`, `pkg/`, `go.mod`, `go.sum`
3. Ignore patterns: `.*`, `*.md`, `specs/`, `docs/`, `tests/`, `.git/`
4. Image naming: `{registry}/c8s-{component}`

**Caching**:
- Docker layer caching enabled
- Only rebuild when watched files change
- Build artifacts cached between runs

**Testing**:
```bash
# After code change, verify rebuild
docker images | grep c8s-
kubectl get pods -n c8s-system  # Verify pod restarted
```

### Kubernetes Deployment

**Components to Deploy**:
1. **controller**: Deployment watching PipelineRun CRDs
2. **api-server**: Deployment serving REST API
3. **webhook**: Deployment for Git webhook reception

**Port Forwarding**:
```
controller (debug):    localhost:6060 → pod:6060    (pprof)
api-server:            localhost:8080 → pod:8080    (HTTP)
webhook:               localhost:9443 → pod:9443    (HTTPS)
```

**Readiness Checks**:
- Pods must reach "Running" state
- Liveness/readiness probes defined in manifests
- Tilt marks resource "ready" when all replicas ready

**Testing**:
```bash
kubectl get pods -n c8s-system -w  # Watch pod startup
curl localhost:8080/health        # Test API server
```

### Resource Limits

**Default Limits** (if not overridden):
- Controller: 500m CPU, 512Mi memory
- API Server: 500m CPU, 512Mi memory
- Webhook: 500m CPU, 256Mi memory

**Override in Deployment**:
- Edit `deploy/install.yaml` to change resource requests/limits
- Or pass config flag (future enhancement)

**Validation**:
```bash
kubectl get pods -n c8s-system -o wide
# Check CPU/Memory allocation
```

## Log Aggregation

### Component Logs Available

**Access Points**:

1. **Tilt Dashboard** (http://localhost:10350):
   - Click component name to view logs
   - Real-time streaming
   - Search functionality

2. **kubectl**:
   ```bash
   kubectl logs -f deployment/c8s-controller -n c8s-system
   kubectl logs -f deployment/c8s-api-server -n c8s-system
   kubectl logs -f deployment/c8s-webhook -n c8s-system
   ```

3. **Tilt CLI**:
   ```bash
   tilt logs controller
   tilt logs api-server
   tilt logs webhook
   ```

### Log Format

Each component logs in structured format:
```
{timestamp} {log_level} {component} {message} [fields...]
```

Example:
```
2025-10-22T14:30:45.123Z INFO controller reconciling pipelinerun=my-pipeline
2025-10-22T14:30:45.234Z ERROR webhook failed to validate request error=invalid_schema
```

## Manual Triggers

### Available Resources

**Local Resources** (manual trigger buttons in Tilt UI):

1. **k3d_create_cluster**
   - Action: Creates k3d cluster
   - Trigger: Manual (automatic if cluster missing)
   - Idempotent: Yes

2. **namespace_setup**
   - Action: Creates Kubernetes namespace
   - Trigger: Manual
   - Idempotent: Yes

3. **install_crds**
   - Action: Apply CRDs from deploy/crds.yaml
   - Trigger: Auto when manifest changes
   - Idempotent: Yes

4. **install_rbac**
   - Action: Apply RBAC from deploy/rbac.yaml
   - Trigger: Auto when manifest changes
   - Idempotent: Yes

5. **pipeline_validator**
   - Action: Placeholder for pipeline validation
   - Trigger: Manual
   - Purpose: Validation framework extension point

6. **cluster_status**
   - Action: Show cluster info and component status
   - Trigger: Manual
   - Output: kubectl cluster-info, pod list, service list

## Customization Points

### Modifying Component Builds

To customize how a component is built:

1. Edit `Dockerfile` (applies to all components)
2. Edit `cmd/{component}/main.go` or related code
3. Tilt automatically detects and rebuilds
4. Pod automatically restarts with new image

### Adding New Components

To add a new component to Tilt:

1. Create `cmd/{new-component}/` directory
2. Add Dockerfile target for new component
3. Add k8s_resource() call in Tiltfile:
   ```python
   docker_build(
       ref=cfg['image_registry'] + '/c8s-{new-component}',
       context='.',
       dockerfile='Dockerfile',
       target='{new-component}',
       only=['cmd/{new-component}/', 'pkg/', 'go.mod', 'go.sum', ...],
   )

   k8s_resource(
       'c8s-{new-component}',
       port_forwards=['PORT:PORT'],
       labels=['components'],
   )
   ```

### Adding New Sample Pipelines

To add sample pipelines:

1. Create YAML file in `config/samples/{name}.yaml`
2. Reference in Tiltfile conditional:
   ```python
   if cfg['with_samples']:
       k8s_yaml(['config/samples/{name}.yaml'])
       k8s_resource('{name}', labels=['samples'], trigger_mode=TRIGGER_MODE_MANUAL)
   ```

### Extending with Tilt Extensions

Tilt allows loading external functionality:

```python
# Load Tilt extension (if needed)
load('ext://restart_process', 'docker_build_with_restart')
```

Current extensions used:
- `ext://restart_process`: For component restart configuration

## Success Criteria

### Tiltfile Must Support

1. ✅ Single `tilt up` command starts complete environment
2. ✅ Code changes detected within 2 seconds
3. ✅ Rebuild and redeploy within 30 seconds of file save
4. ✅ Unified logs in Tilt dashboard for all components
5. ✅ Component-specific resource monitoring (CPU, memory)
6. ✅ Manual triggers for validation and status checks
7. ✅ Configuration via command-line flags
8. ✅ Idempotent setup (running twice = same result)
9. ✅ Graceful error messages for missing dependencies
10. ✅ Cluster cleanup on `tilt down`

### Testing Checklist

- [ ] `tilt up` creates cluster from scratch in < 5 minutes
- [ ] File modification triggers rebuild within 30 seconds
- [ ] Dashboard shows logs from all three components
- [ ] Manual triggers work (CRD, RBAC reapplication)
- [ ] `tilt down` cleans up all resources
- [ ] Configuration flags are recognized and applied
- [ ] Cluster survives `tilt down` → `tilt up` cycle
- [ ] Multiple developers can run independently

## API Stability

**Current Version**: 1.0

**Backward Compatibility**:
- Flag additions: Non-breaking (new flags have defaults)
- Flag removals: Breaking (document deprecation path)
- Behavior changes: Breaking (document clearly)

**Future Enhancements**:
- Config file support (tilt.yaml)
- More granular resource limits per component
- Custom validation hooks
- Performance profiling integration

## References

- **Tilt Documentation**: https://docs.tilt.dev
- **Tiltfile Syntax**: https://docs.tilt.dev/tutorial
- **k3d Documentation**: https://k3d.io
- **Kubernetes API**: https://kubernetes.io/docs/reference/generated/kubernetes-api/
- **Go Build Options**: https://golang.org/cmd/go/

---

**Related Documents**:
- [Quick Start Guide](../docs/tilt-setup.md)
- [Data Model](../data-model.md)
- [Feature Specification](../spec.md)
