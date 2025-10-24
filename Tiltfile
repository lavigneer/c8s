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

# Load Tilt libraries for utilities
# Note: ext://restart_process is optional - comment out if not available
# load('ext://restart_process', 'docker_build_with_restart')

# Configuration with defaults - use simple assignments instead of config API
with_samples = True
verbose_logs = False
k8s_namespace = 'c8s-system'

# Environment variables for builds
os.environ['CGO_ENABLED'] = '0'
os.environ['GOOS'] = 'linux'
os.environ['GOARCH'] = 'amd64'

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
# This applies RBAC, ServiceAccounts, and configuration to the cluster
k8s_yaml('deploy/install.yaml')

# ============================================================================
# Component Build Configuration
# ============================================================================

# Build all components using multi-stage Dockerfile with builder target
def build_component(component_name, port=None):
    """Configure Docker build for a C8S component"""
    dockerfile = 'Dockerfile'
    context = '.'
    target = component_name

    # Build context with hot-reload capability
    build_config = docker_build(
        ref='c8s-' + component_name + ':latest',
        context=context,
        dockerfile=dockerfile,
        target=target,
        only=[
            'cmd/' + component_name + '/',
            'pkg/',
            'go.mod',
            'go.sum',
            'Makefile',
            'PROJECT',
            'hack/',
        ],
        # Ignore files that shouldn't trigger rebuilds
        ignore=['.*', 'README*', 'specs/', 'docs/', '*.md', 'tests/', '.git/'],
        # For local development, enable live update when possible
        entrypoint=['/' + component_name]
    )

    return build_config

# ============================================================================
# Controller Component
# ============================================================================

docker_build(
    ref='c8s-controller:latest',
    context='.',
    dockerfile='Dockerfile',
    target='controller',
    only=[
        'cmd/controller/',
        'pkg/',
        'go.mod',
        'go.sum',
        'Makefile',
        'PROJECT',
        'hack/',
    ],
    ignore=['.*', 'README*', 'specs/', 'docs/', '*.md', 'tests/', '.git/'],
)

k8s_resource(
    'c8s-controller',
    port_forwards=['6060:6060'],  # Pprof debug port
    labels=['controller'],
    trigger_mode=TRIGGER_MODE_AUTO
)

# ============================================================================
# API Server Component
# ============================================================================

# Note: c8s-api-server deployment not yet in install.yaml, so docker_build skipped for now

# ============================================================================
# Webhook Component
# ============================================================================

docker_build(
    ref='c8s-webhook:latest',
    context='.',
    dockerfile='Dockerfile',
    target='webhook',
    only=[
        'cmd/webhook/',
        'pkg/',
        'go.mod',
        'go.sum',
        'Makefile',
        'PROJECT',
        'hack/',
    ],
    ignore=['.*', 'README*', 'specs/', 'docs/', '*.md', 'tests/', '.git/'],
)

# Note: c8s-webhook is deployed via k8s_yaml but not explicitly tracked as k8s_resource
# This is OK - it will be managed by Tilt via the manifests

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
    'kubectl cluster-info && echo "\\n=== C8S Components ===" && kubectl get pods -n ' + k8s_namespace + ' && echo "\\n=== Service Endpoints ===" && kubectl get svc -n ' + k8s_namespace,
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
╭─────────────────────────────────────────────────────────────────╮
│         C8S Local Development with Tilt                        │
│                                                                 │
│ Usage:                                                          │
│   tilt up              - Start development environment         │
│   tilt down            - Shut down development environment     │
│   tilt logs controller - View controller logs                  │
│                                                                 │
│ Web UI:                                                         │
│   Open http://localhost:10350 (Tilt dashboard)                │
│                                                                 │
│ Components:                                                     │
│   - Controller:   http://localhost:6060 (pprof)               │
│   - API Server:   http://localhost:8080                       │
│   - Webhook:      https://localhost:9443                      │
│                                                                 │
│ Workflow:                                                       │
│   1. Edit Go files in cmd/ or pkg/                            │
│   2. Tilt automatically detects changes and rebuilds          │
│   3. Components redeploy automatically                        │
│   4. View logs in Tilt dashboard                              │
│                                                                 │
│ Documentation:                                                  │
│   See docs/tilt-setup.md for detailed guide                   │
│                                                                 │
╰─────────────────────────────────────────────────────────────────╯
""")
