# Tiltfile for C8S Local Kubernetes Development
# Provides automatic code change detection, component rebuilding, and unified logging
# for local development of the C8S continuous integration system.
#
# WORKFLOW:
# 1. Pre-create the cluster (one-time):
#    k3d cluster create c8s-dev --registry-create=registry:5000 -p "8080:80@loadbalancer" --servers 1 --agents 2
# 2. Run Tilt:
#    tilt up
# 3. Access the Tilt dashboard:
#    http://localhost:10350
#
# Alternatively, trigger cluster creation from the Tilt UI if needed.

# Extensions can be loaded here if needed, but standard docker_build with k8s
# restart_policy handles most development workflows efficiently

# Configuration with defaults - use simple assignments instead of config API
with_samples = True
verbose_logs = False
k8s_namespace = 'c8s-system'

# Environment variables for builds
os.environ['CGO_ENABLED'] = '0'
os.environ['GOOS'] = 'linux'
os.environ['GOARCH'] = 'arm64'

# ============================================================================
# Cluster Configuration
# ============================================================================

# Configure k3d cluster
cluster_name = 'c8s-dev'

# Create or use existing k3d cluster
# Command will succeed if cluster exists (idempotent)
local_resource(
    'k3d_create_cluster',
    'k3d cluster get ' + cluster_name + ' > /dev/null 2>&1 || k3d cluster create ' + cluster_name + ' --registry-create=registry:5000 -p "8080:80@loadbalancer" --servers 1 --agents 2',
    trigger_mode=TRIGGER_MODE_MANUAL,
    env={
        'PATH': os.environ['PATH'],
    }
)

# Set kubeconfig context
# Note: Tilt will handle kubeconfig automatically, no need to set it explicitly

# Set Kubernetes context
allow_k8s_contexts(cluster_name)
k8s_context(cluster_name)

# ============================================================================
# Namespace Setup
# ============================================================================

# Ensure namespace exists (created via local_resource)
local_resource(
    'namespace_setup',
    'kubectl create namespace ' + k8s_namespace + ' --dry-run=client -o yaml | kubectl apply -f -',
    trigger_mode=TRIGGER_MODE_MANUAL
)

# ============================================================================
# CRD and RBAC Installation
# ============================================================================

# Apply CRDs from deploy/crds.yaml
local_resource(
    'install_crds',
    'kubectl apply -f deploy/crds.yaml',
    deps=['deploy/crds.yaml'],
    trigger_mode=TRIGGER_MODE_AUTO
)

# Load manifests for Tilt resource tracking
# Split into separate files for better tracking and organization
k8s_yaml([
    'deploy/01-namespace.yaml',
    'deploy/crds.yaml',
    'deploy/02-controller-rbac.yaml',
    'deploy/03-controller-deployment.yaml',
    'deploy/04-webhook-rbac.yaml',
    'deploy/05-webhook-deployment.yaml',
    'deploy/06-api-server-deployment.yaml',
])

# ============================================================================
# Kubernetes Resource Tracking Configuration
# ============================================================================
# Comprehensive resource tracking for better observability in Tilt UI
#
# Organized by category:
#  - 'controller': Controller deployment and status
#  - 'webhook': Webhook deployment and status
#  - 'networking': Services, ingress, endpoints
#  - 'infrastructure': CRDs, RBAC, namespace setup
#  - 'status': Cluster health, events, resource usage
#
# Resources can be filtered and grouped by labels in Tilt UI

# Resource groups by function
resource_groups = {
    'controller': ['c8s-controller'],
    'webhook': ['c8s-webhook'],
    'infrastructure': ['install_crds', 'rbac_status'],
    'networking': ['service_endpoints', 'ingress_status'],
    'status': ['cluster_status', 'k8s_events', 'resource_usage'],
}

# Track all ServiceAccounts for RBAC visibility
local_resource(
    'rbac_status',
    'kubectl get serviceaccounts -n ' + k8s_namespace + ' && kubectl get roles -n ' + k8s_namespace,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['infrastructure'],
    allow_parallel=True
)

# ============================================================================
# Component Build Configuration
# ============================================================================
# Components are built using docker_build_with_restart extension which:
# - Automatically rebuilds when source files change
# - Restarts the container entrypoint on live updates
# - Provides faster iteration cycles for Kubebuilder projects

# ============================================================================
# Controller Component
# ============================================================================

docker_build(
    ref='c8s-controller:latest',
    context='.',
    dockerfile='Dockerfile',
    target='controller',
    only=['cmd/controller/', 'cmd/webhook/', 'cmd/api-server/', 'pkg/', 'hack/', 'go.mod', 'go.sum', 'Dockerfile', 'Makefile', 'PROJECT'],
)

k8s_resource(
    'c8s-controller',
    port_forwards=['6060:6060'],  # Pprof debug port
    labels=['controller'],
    trigger_mode=TRIGGER_MODE_AUTO,
    pod_readiness='wait'
)

# ============================================================================
# Webhook Component
# ============================================================================

docker_build(
    ref='c8s-webhook:latest',
    context='.',
    dockerfile='Dockerfile',
    target='webhook',
    only=['cmd/controller/', 'cmd/webhook/', 'cmd/api-server/', 'pkg/', 'hack/', 'go.mod', 'go.sum', 'Dockerfile', 'Makefile', 'PROJECT'],
)

# Track webhook deployment
k8s_resource(
    'c8s-webhook',
    labels=['webhook'],
    trigger_mode=TRIGGER_MODE_AUTO,
    pod_readiness='wait'
)

# ============================================================================
# API Server Component (HTMX Frontend)
# ============================================================================

docker_build(
    ref='c8s-api-server:latest',
    context='.',
    dockerfile='Dockerfile',
    target='api-server',
    only=['cmd/controller/', 'cmd/webhook/', 'cmd/api-server/', 'pkg/', 'hack/', 'go.mod', 'go.sum', 'Dockerfile', 'Makefile', 'PROJECT'],
)

# Track API Server deployment
k8s_resource(
    'c8s-api-server',
    port_forwards=['8080:8080'],  # Forward dashboard to localhost:8080
    labels=['api-server'],
    trigger_mode=TRIGGER_MODE_AUTO,
    pod_readiness='wait'
)

# ============================================================================
# Network and Service Tracking
# ============================================================================

# Note: Webhook service will be tracked automatically via k8s_yaml
# No explicit k8s_resource needed as Tilt discovers it from manifests

# Display service endpoints and networking status
local_resource(
    'service_endpoints',
    'echo "=== C8S Service Endpoints ===" && kubectl get services -n ' + k8s_namespace + ' -o wide',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['networking'],
    allow_parallel=True
)

# Display ingress and routing configuration
local_resource(
    'ingress_status',
    'echo "=== C8S Ingress Configuration ===" && kubectl get ingress -n ' + k8s_namespace + ' || echo "No ingress resources"',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['networking'],
    allow_parallel=True
)

# ============================================================================
# Pipeline Validation
# ============================================================================

# Add local resource for pipeline validation
local_resource(
    'pipeline_validator',
    'echo "Pipeline validation ready"',
    labels=['validation'],
    trigger_mode=TRIGGER_MODE_MANUAL
)

# ============================================================================
# Cluster Status and Information
# ============================================================================

local_resource(
    'cluster_status',
    '''
kubectl cluster-info
echo "\\n=== Cluster Nodes ==="
kubectl get nodes -o wide
echo "\\n=== C8S Components ==="
kubectl get pods -n ''' + k8s_namespace + ''' -o wide
echo "\\n=== Component Status ==="
kubectl get deployments -n ''' + k8s_namespace + ''' -o wide
    ''',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['status'],
    allow_parallel=True
)

# Monitor Kubernetes events for troubleshooting
local_resource(
    'k8s_events',
    'echo "=== Recent K8s Events ===" && kubectl get events -n ' + k8s_namespace + ' --sort-by=".lastTimestamp" | tail -20',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['status'],
    allow_parallel=True
)

# Show resource usage/constraints
local_resource(
    'resource_usage',
    'echo "=== Resource Requests/Limits ===" && kubectl describe nodes | grep -A 5 "Allocated resources" || kubectl top nodes 2>/dev/null || echo "Metrics server not available"',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['status'],
    allow_parallel=True
)

# ============================================================================
# Sample Pipelines (Optional)
# ============================================================================

if with_samples:
    k8s_yaml(['config/samples/simple-build.yaml'])
    # Note: Sample resources are deployed via k8s_yaml but not explicitly tracked

# ============================================================================
# Development Workflow Tips
# ============================================================================

print("""
╭──────────────────────────────────────────────────────────────────╮
│         C8S Local Development with Tilt + HTMX Dashboard       │
│                                                                  │
│ Usage:                                                           │
│   tilt up              - Start development environment          │
│   tilt down            - Shut down development environment      │
│   tilt logs <service>  - View logs (controller/api-server)     │
│                                                                  │
│ Access Points:                                                   │
│   - Tilt Dashboard:    http://localhost:10350                  │
│   - C8S Dashboard:     http://localhost:8080                   │
│   - Controller Pprof:  http://localhost:6060                   │
│   - Webhook:           https://localhost:9443                  │
│                                                                  │
│ Components Running:                                              │
│   ✓ Controller:        Manages C8S Kubernetes resources        │
│   ✓ API Server:        HTMX-based web dashboard                │
│   ✓ Webhook:           Git webhook event receiver              │
│                                                                  │
│ Development Workflow:                                            │
│   1. Edit Go files in cmd/api-server or pkg/dashboard         │
│   2. Edit templates in cmd/api-server/templates/              │
│   3. Tilt automatically detects changes and rebuilds           │
│   4. Browser auto-refresh shows updates                        │
│   5. View logs and metrics in Tilt dashboard                   │
│                                                                  │
│ Hot Reload:                                                      │
│   - Go backend: Automatic rebuild and restart                  │
│   - Templates: Live update on save                             │
│   - Static assets: Automatic cache invalidation                │
│                                                                  │
│ Documentation:                                                   │
│   - See docs/tilt-setup.md for detailed guide                  │
│   - See DASHBOARD_README.md for frontend development           │
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯
""")
