# Tiltfile for C8S Local Kubernetes Development
# Provides automatic code change detection, component rebuilding, and unified logging
# for local development of the C8S continuous integration system.

# Tilt version check
min_tilt_version('0.33.0')

# Load Tilt libraries for utilities
load('ext://restart_process', 'docker_build_with_restart')

# Configuration
config.define_bool('with_samples', True, 'Deploy sample pipelines')
config.define_bool('verbose_logs', False, 'Enable verbose logging for all components')
config.define_string('k8s_namespace', 'c8s-system', 'Kubernetes namespace for C8S components')
config.define_string('image_registry', 'c8s-dev', 'Docker image registry/prefix for local development')

cfg = config.parse()

# Environment variables for builds
os.environ['DOCKER_REGISTRY'] = cfg['image_registry']
os.environ['CGO_ENABLED'] = '0'
os.environ['GOOS'] = 'linux'
os.environ['GOARCH'] = 'amd64'

# ============================================================================
# Cluster Configuration
# ============================================================================

# Configure k3d cluster
cluster_name = 'c8s-dev'

# Check if cluster exists, create if not
result = local('k3d cluster list | grep -q ' + cluster_name, quiet=True, echo_off=True)
if result != 0:
    # Cluster doesn't exist, create it with Tilt
    print('Creating k3d cluster: ' + cluster_name)
    local_resource(
        'k3d_create_cluster',
        'k3d cluster create ' + cluster_name + ' --registry-create=registry:5000 -p "8080:80@loadbalancer" --servers 1 --agents 2',
        trigger_mode=TRIGGER_MODE_MANUAL,
        env={
            'PATH': os.environ['PATH'],
        }
    )
else:
    print('Using existing k3d cluster: ' + cluster_name)

# Set kubeconfig context
os.environ['KUBECONFIG'] = os.path.expanduser('~/.k3d/' + cluster_name + '/kubeconfig.yaml')

# Set default namespace
default_registry(cfg['image_registry'])
allow_k8s_contexts(cluster_name)
k8s_context(cluster_name)

# ============================================================================
# Namespace Setup
# ============================================================================

# Ensure namespace exists
k8s_resource('namespace', objects=[''])
local_resource(
    'namespace_setup',
    'kubectl create namespace ' + cfg['k8s_namespace'] + ' --dry-run=client -o yaml | kubectl apply -f -',
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

# Apply manifests (includes RBAC, ServiceAccounts, and configuration)
local_resource(
    'install_manifests',
    'kubectl apply -f deploy/install.yaml',
    deps=['deploy/install.yaml'],
    trigger_mode=TRIGGER_MODE_AUTO
)

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
        ref=cfg['image_registry'] + '/c8s-' + component_name,
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
    ref=cfg['image_registry'] + '/c8s-controller',
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
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True
)

# ============================================================================
# API Server Component
# ============================================================================

docker_build(
    ref=cfg['image_registry'] + '/c8s-api-server',
    context='.',
    dockerfile='Dockerfile',
    target='api-server',
    only=[
        'cmd/api-server/',
        'pkg/',
        'web/',
        'go.mod',
        'go.sum',
        'Makefile',
        'PROJECT',
        'hack/',
    ],
    ignore=['.*', 'README*', 'specs/', 'docs/', '*.md', 'tests/', '.git/'],
)

k8s_resource(
    'c8s-api-server',
    port_forwards=['8080:8080'],  # API server port
    labels=['api-server'],
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True
)

# ============================================================================
# Webhook Component
# ============================================================================

docker_build(
    ref=cfg['image_registry'] + '/c8s-webhook',
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

k8s_resource(
    'c8s-webhook',
    port_forwards=['9443:9443'],  # Webhook HTTPS port
    labels=['webhook'],
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True
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
    '''kubectl cluster-info &&
       echo "\\n=== C8S Components ===" &&
       kubectl get pods -n ''' + cfg['k8s_namespace'] + ''' &&
       echo "\\n=== Service Endpoints ===" &&
       kubectl get svc -n ''' + cfg['k8s_namespace'] + ''',
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['status'],
    allow_parallel=True
)

# ============================================================================
# Sample Pipelines (Optional)
# ============================================================================

if cfg['with_samples']:
    k8s_yaml(['config/samples/simple-build.yaml'])
    k8s_resource('simple-build', labels=['samples'], trigger_mode=TRIGGER_MODE_MANUAL)

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
