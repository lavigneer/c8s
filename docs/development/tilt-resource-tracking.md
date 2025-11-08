# Tilt Resource Tracking Guide

This document explains how resources are organized and tracked in the C8S Tiltfile for better observability and debugging.

## Resource Discovery

Tilt automatically discovers resources from the `k8s_yaml()` manifests. All Deployments, Services, and other resources defined in `deploy/install.yaml` are automatically available for tracking via their resource name. When multiple resources share the same name (e.g., a Deployment and Service both named "webhook"), Tilt uses the first match found in the manifests.

For explicit resource tracking with `k8s_resource()`, we reference resources by their metadata name:
- `c8s-controller` - References the c8s-controller Deployment
- `c8s-webhook` - References the c8s-webhook Deployment

This approach keeps the configuration simple while allowing full observability.

## Resource Organization

Resources are organized by functional category with labels for easy filtering in the Tilt UI. Each resource is tracked using Tilt's object selector syntax for precise Kubernetes resource identification.

### Component Resources (Deployed Services)

#### Controller Deployment
- **Resource Name**: `c8s-controller`
- **Label**: `controller`
- **Port Forwards**: 6060 (pprof debug port)
- **Status**: Auto-tracked, waits for pod readiness
- **Docker Image**: `c8s-controller:latest`
- **Tracks**: Deployment, replicas, pod logs

#### Webhook Deployment
- **Resource Name**: `c8s-webhook`
- **Label**: `webhook`
- **Status**: Auto-tracked, waits for pod readiness
- **Docker Image**: `c8s-webhook:latest`
- **Tracks**: Deployment, replicas, pod logs

#### Webhook Service
- **Resource Name**: `c8s-webhook` (Service)
- **Label**: `networking`
- **Status**: Auto-discovered from manifests
- **Shows**: Cluster IP, ports, endpoints
- **Tracks**: Service endpoints and port mappings

### Networking Resources

#### Service Endpoints
- **Command**: `service_endpoints`
- **Label**: `networking`
- **Shows**: All services in c8s-system namespace with IP and port mappings
- **Trigger**: Manual (run on demand)

#### Ingress Status
- **Command**: `ingress_status`
- **Label**: `networking`
- **Shows**: Ingress configuration and routing rules
- **Trigger**: Manual (run on demand)

### Infrastructure Resources

#### RBAC Status
- **Command**: `rbac_status`
- **Label**: `infrastructure`
- **Shows**: ServiceAccounts and Roles in the cluster
- **Trigger**: Manual (run on demand)

#### CRD Installation
- **Command**: `install_crds`
- **Label**: `infrastructure`
- **Watches**: `deploy/crds.yaml`
- **Trigger**: Automatic (watches file changes)

#### Namespace Setup
- **Command**: `namespace_setup`
- **Label**: `infrastructure`
- **Creates**: `c8s-system` namespace (idempotent)
- **Trigger**: Manual (run once during setup)

### Status and Monitoring Resources

#### Cluster Status
- **Command**: `cluster_status`
- **Label**: `status`
- **Shows**:
  - Cluster info (API endpoints, etc.)
  - Cluster nodes (CPU, memory, status)
  - C8S component pods (running state)
  - Deployment status (replicas, ready count)
- **Trigger**: Manual (run on demand)

#### Kubernetes Events
- **Command**: `k8s_events`
- **Label**: `status`
- **Shows**: Recent cluster events (last 20) sorted by timestamp
- **Use Case**: Debugging pod failures, scheduling issues
- **Trigger**: Manual (run on demand)

#### Resource Usage
- **Command**: `resource_usage`
- **Label**: `status`
- **Shows**: Node resource allocation and limits
- **Fallback**: Uses `kubectl describe nodes` if metrics server unavailable
- **Trigger**: Manual (run on demand)

## Viewing Resources in Tilt UI

### By Label Filter
In the Tilt UI, you can filter resources by label:

1. Open http://localhost:10350 (Tilt dashboard)
2. Use the label filter at the top to view:
   - `controller` - Only controller deployment
   - `webhook` - Only webhook deployment
   - `networking` - Service and ingress information
   - `infrastructure` - CRDs, RBAC, namespace setup
   - `status` - Cluster health and monitoring

### Viewing Logs
Click on any resource to view its logs:
- **Deployments**: Shows container logs in real-time
- **Local Resources**: Shows command output and execution status

## Debugging Workflow

### 1. Check Component Status
```
Click "cluster_status" → Run
```
Verify all pods are running and ready replicas match desired.

### 2. Check Recent Events
```
Click "k8s_events" → Run
```
Look for warnings or errors explaining why components aren't starting.

### 3. Check Networking
```
Click "service_endpoints" → Run
```
Verify services have endpoints and IP addresses assigned.

### 4. Check RBAC
```
Click "rbac_status" → Run
```
Verify ServiceAccounts exist and are properly configured.

### 5. View Component Logs
```
Click "c8s-controller" or "c8s-webhook"
```
Scroll through logs to find errors or debug output.

## Resource Readiness

Resources are configured with `pod_readiness='wait'`, which means Tilt will:

1. Wait for the Deployment to be created
2. Wait for pods to reach "Ready" state
3. Only then mark the resource as "OK" in the UI

This prevents cascading failures where downstream resources try to start before dependencies are ready.

## Auto-Trigger Behavior

### Automatic (TRIGGER_MODE_AUTO)
- `install_crds`: Watches `deploy/crds.yaml`
- Component deployments: Watch docker builds and source code

### Manual (TRIGGER_MODE_MANUAL)
- Status commands: Run on demand for performance
- Infrastructure setup: Run once, don't repeat
- Network debugging: Run when needed

## Resource Groups

The Tiltfile defines logical groups for organizational purposes:

```python
resource_groups = {
    'controller': ['c8s-controller'],
    'webhook': ['c8s-webhook'],
    'infrastructure': ['namespace_setup', 'install_crds', 'rbac_status'],
    'networking': ['service_endpoints', 'ingress_status'],
    'status': ['cluster_status', 'k8s_events', 'resource_usage'],
}
```

These can be referenced for custom Tilt buttons or resource dependencies in future enhancements.

## Common Troubleshooting

### "CrashLoopBackOff" Status
1. Run `cluster_status` to confirm pod is in crash loop
2. Run `k8s_events` to see error messages
3. Click the pod resource to view container logs

### "Pending" Status
1. Run `k8s_events` - check for scheduling errors
2. Run `resource_usage` - verify node has available capacity
3. Check resource requests/limits in deployment

### Service Not Reachable
1. Run `service_endpoints` - verify service has endpoints
2. Click the pod - verify it's listening on the right port
3. Run `k8s_events` - check for networking issues

### Image Pull Errors
1. Check docker build status in Tilt UI
2. Verify image exists in local Docker: `docker images | grep c8s`
3. Check k3d registry: `k3d image list --cluster c8s-dev`

## Performance Tips

- Most status commands are manual triggers to avoid overhead
- Use `allow_parallel=True` for status commands so they don't block the UI
- Component resources use `TRIGGER_MODE_AUTO` for fast iteration
- Docker builds cache aggressively for quick rebuilds
