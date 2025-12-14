# Tilt configuration for C8S development
#
# This Tiltfile integrates the C8S Helm chart with local Docker image building
# and deploys to your local Kubernetes cluster
#
# Configuration:
#   GHCR_REGISTRY: GitHub Container Registry (ghcr.io)
#   DOCKER_BUILDKIT: Enabled for faster builds
#   TILT_DOCKER_BUILD: Uses local Docker daemon
#
# Usage:
#   tilt up              # Start Tilt with C8S deployment
#   tilt down            # Stop Tilt and remove resources
#   tilt logs c8s        # View logs for C8S
#   tilt trigger c8s     # Force redeploy
#   tilt logs controller # View controller logs
#
# ngrok Integration:
#   ngrok CLI is used to tunnel both the API server and webhook for external access
#   Separate ngrok instances are launched for each service
#   Port mapping: API Server 8000, Webhook 8080
#   The tunnel URLs are displayed in Tilt logs
#   Requirements:
#     - ngrok installed: https://ngrok.com/download
#     - ngrok authtoken configured: ngrok config add-authtoken <TOKEN>
#
# Setup for GitHub Actions dog-fooding:
#   1. Run: tilt up
#   2. Get ngrok URLs from Tilt logs (c8s-api-server-ngrok and c8s-webhook-ngrok)
#      Example: https://abc1234.ngrok.io
#   3. Add GitHub repository secrets:
#      - C8S_API_URL: <api-server-ngrok-url>
#      - C8S_WEBHOOK_URL: <webhook-ngrok-url>
#   4. Push changes - GitHub Actions will trigger your local C8S pipeline
#
# Environment Variables (optional):
#   GHCR_REGISTRY=ghcr.io/lavigneer  # GitHub Container Registry path
#   IMAGE_TAG=v0.1.0                   # Image tag (default: latest)
#   NGROK_AUTHTOKEN=<token>            # ngrok authentication token (if not in ~/.config/ngrok/ngrok.yml)

# Load extensions
load('ext://namespace', 'namespace_create')
load('ext://helm_resource', 'helm_resource')
load('ext://restart_process', 'docker_build_with_restart', 'custom_build_with_restart')

# Increase Tilt's apply timeout for long-running operations like cert-manager
update_settings(k8s_upsert_timeout_secs=600)

# ============================================================================
# Install CRDs (must be before Helm chart deployment)
# ============================================================================
# CRDs are deployed directly from config/crd/bases/ before the Helm chart
# This ensures they're available when the chart resources are created

k8s_yaml('./config/crd/bases/c8s.dev_pipelineconfigs.yaml')
k8s_yaml('./config/crd/bases/c8s.dev_pipelineruns.yaml')
k8s_yaml('./config/crd/bases/c8s.dev_repositoryconnections.yaml')

# ============================================================================
# Configuration
# ============================================================================

# Get the repository owner/name from git remote
# This will be used to construct the GHCR registry path: ghcr.io/{owner}/{repo}
def get_github_repo():
  # First check environment variable (set in GitHub Actions)
  repo = os.getenv('GITHUB_REPOSITORY', '')
  if repo:
    return repo

  # Try to get from git remote (local development)
  result = local('git config --get remote.origin.url', quiet=True)
  if result:
    url = str(result).strip()
    # Extract owner/repo from git@github.com:owner/repo.git or https://github.com/owner/repo
    if 'github.com' in url:
      if url.startswith('git@'):
        # git@github.com:owner/repo.git
        parts = url.replace('git@github.com:', '').replace('.git', '').split('/')
      else:
        # https://github.com/owner/repo
        parts = url.rstrip('/').split('/')
      if len(parts) >= 2:
        return parts[-2] + "/" + parts[-1]

  # If we can't detect, fail with a helpful error message
  fail("""
  ❌ Could not detect GitHub repository from git remote.

  Please ensure your git remote is configured correctly:
    git remote -v

  Expected format:
    - git@github.com:username/repo.git (SSH)
    - https://github.com/username/repo (HTTPS)

  Or set the GITHUB_REPOSITORY environment variable:
    export GITHUB_REPOSITORY=username/repo
    tilt up
  """)

GITHUB_REPO = get_github_repo()
GHCR_REGISTRY = "ghcr.io/" + GITHUB_REPO.lower()
IMAGE_TAG = os.getenv('IMAGE_TAG', 'latest')

# Build context for all images
BUILD_DIR = '.'

# ============================================================================
# Build Images
# ============================================================================

print("=" * 80)
print("Building C8S images...")
print("Repository: " + GITHUB_REPO)
print("Registry: " + GHCR_REGISTRY)
print("Tag: " + IMAGE_TAG)
print("=" * 80)

# ============================================================================
# Compile Go Binaries (local resources that watch for source changes)
# ============================================================================

# Compile API Server binary for Linux arm64
local_resource(
  name='compile-api-server',
  cmd='mkdir -p bin && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/api-server ./cmd/api-server',
  deps=[
    'cmd/api-server/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
  ],
)

# Compile Controller binary for Linux arm64
local_resource(
  name='compile-controller',
  cmd='mkdir -p bin && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/controller ./cmd/controller',
  deps=[
    'cmd/controller/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
  ],
)

# Compile Webhook binary for Linux arm64
local_resource(
  name='compile-webhook',
  cmd='mkdir -p bin && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/webhook ./cmd/webhook',
  deps=[
    'cmd/webhook/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
  ],
)

# ============================================================================
# Build Container Images (depend on compiled binaries)
# ============================================================================

# Build API Server with auto-restart
# Initial Docker build packages the pre-compiled binary, then live_sync updates it
docker_build_with_restart(
  ref='c8s-api-server',
  context='.',
  dockerfile='./Dockerfile',
  target='api-server',
  entrypoint=['/app/api-server', '-base-dir', '/app'],
  only=[
    'bin/api-server',
    'cmd/api-server/templates',
    'cmd/api-server/static',
    'Dockerfile',
    'cmd/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
    'hack/',
    'PROJECT',
  ],
  live_update=[
    sync('bin/api-server', '/app/api-server'),
    sync('cmd/api-server/templates', '/app/templates'),
    sync('cmd/api-server/static', '/app/static'),
  ],
)

# Build Controller with auto-restart
# Initial Docker build packages the pre-compiled binary, then live_sync updates it
docker_build_with_restart(
  ref='c8s-controller',
  context='.',
  dockerfile='./Dockerfile',
  target='controller',
  entrypoint=['/app/controller'],
  only=[
    'bin/controller',
    'Dockerfile',
    'cmd/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
    'hack/',
    'PROJECT',
  ],
  live_update=[
    sync('bin/controller', '/app/controller'),
  ],
)

# Build Webhook with auto-restart
# Initial Docker build packages the pre-compiled binary, then live_sync updates it
docker_build_with_restart(
  ref='c8s-webhook',
  context='.',
  dockerfile='./Dockerfile',
  target='webhook',
  entrypoint=['/app/webhook'],
  only=[
    'bin/webhook',
    'Dockerfile',
    'cmd/',
    'pkg/',
    'go.mod',
    'go.sum',
    'Makefile',
    'hack/',
    'PROJECT',
  ],
  live_update=[
    sync('bin/webhook', '/app/webhook'),
  ],
)

# Build Frontend image (if Dockerfile exists for it)
# For now, we use the chart's default image reference
# You can add a separate build when the frontend Dockerfile is available

# ============================================================================
# Install Cert-Manager (Separate Helm Release)
# ============================================================================

namespace_create('cert-manager')

helm_resource(
  name='cert-manager',
  chart='oci://quay.io/jetstack/charts/cert-manager',
  namespace='cert-manager',
  flags=[
    '--set', 'installCRDs=true',
    '--set', 'global.leaderElection.namespace=cert-manager',
    '--wait',
    '--timeout', '5m',
  ],
  labels=['infrastructure'],
)

# ============================================================================
# Install MinIO (S3-Compatible Object Storage for Log Persistence)
# ============================================================================

namespace_create('c8s-system')

# Deploy MinIO using the manifest (simpler than Helm, no registry auth needed)
local_resource(
  name='minio',
  cmd='kubectl apply -f ./deploy/minio.yaml',
  resource_deps=[],
  labels=['infrastructure'],
)

# ============================================================================
# Deploy C8S
# ============================================================================

helm_resource(
  name='c8s',
  chart='./chart/c8s',
  flags=[
    '-f', './chart/c8s/values-dev.yaml',
    '-f', './tilt/config/c8s-values.yaml',
    '--create-namespace',
  ],
  image_deps=[
    'c8s-api-server',
    'c8s-controller',
    'c8s-webhook',
  ],
  image_keys=[
    ('components.apiServer.image.registry', 'components.apiServer.image.repository', 'components.apiServer.image.tag'),
    ('components.controller.image.registry', 'components.controller.image.repository', 'components.controller.image.tag'),
    ('components.webhook.image.registry', 'components.webhook.image.repository', 'components.webhook.image.tag'),
  ],
  namespace='c8s-system',
  resource_deps=['cert-manager', 'minio'],
)

# ============================================================================
# Deploy Sample RepositoryConnection for dog-fooding
# ============================================================================
# This enables the webhook to trigger pipelines when code is pushed to GitHub
# Deployed via local_resource that depends on the helm chart

local_resource(
  name='deploy-pipelineconfig',
  cmd='kubectl apply -f ./config/samples/pipelineconfig-c8s.yaml',
  resource_deps=['c8s'],
)

local_resource(
  name='deploy-repository-connection',
  cmd='kubectl apply -f ./config/samples/repositoryconnection-c8s.yaml',
  resource_deps=['c8s', 'deploy-pipelineconfig'],
)

# Port-forwards for ngrok integration
# These create port-forward tunnels and expose ngrok buttons in the Tilt UI
local_resource(
  name='c8s-api-server-port-forward',
  serve_cmd='kubectl port-forward -n c8s-system svc/c8s-api-server 8000:8080',
  allow_parallel=True,
  resource_deps=['c8s'],
)

local_resource(
  name='c8s-controller-port-forward',
  serve_cmd='kubectl port-forward -n c8s-system svc/c8s-controller 8081:8081',
  allow_parallel=True,
  resource_deps=['c8s'],
)

local_resource(
  name='c8s-webhook-port-forward',
  serve_cmd='kubectl port-forward -n c8s-system svc/c8s-webhook 8080:8080',
  allow_parallel=True,
  resource_deps=['c8s'],
)

# MinIO port-forwards for S3 access
local_resource(
  name='minio-api-port-forward',
  serve_cmd='kubectl port-forward -n c8s-system svc/minio 9000:9000',
  allow_parallel=True,
  resource_deps=['minio'],
)

local_resource(
  name='minio-console-port-forward',
  serve_cmd='kubectl port-forward -n c8s-system svc/minio 9001:9001',
  allow_parallel=True,
  resource_deps=['minio'],
)

# nginx reverse proxy for consolidating API and webhook under one ngrok tunnel
local_resource(
  name='c8s-nginx-proxy',
  serve_cmd='docker run --rm --network host -v $(pwd)/tilt/config/nginx.conf:/etc/nginx/nginx.conf:ro nginx:latest',
  allow_parallel=True,
  resource_deps=['c8s-api-server-port-forward', 'c8s-webhook-port-forward'],
)

# ngrok tunnel for C8S services (optional)
# Routes through nginx reverse proxy on port 8888
# nginx handles path-based routing to API server and webhook

local_resource(
  name='c8s-ngrok-tunnel',
  serve_cmd='ngrok start --all --config=./tilt/config/ngrok-config.yml --config=$HOME/.config/ngrok/ngrok.yml --log=stdout',
  allow_parallel=True,
  resource_deps=['c8s-nginx-proxy'],
)


# ============================================================================
# Watch Files
# ============================================================================

watch_file('./chart/c8s')
watch_file('./tilt/config/c8s-values.yaml')
watch_file('./cmd/')
watch_file('./pkg/')
