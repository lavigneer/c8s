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

# Increase Tilt's apply timeout for long-running operations like cert-manager
update_settings(k8s_upsert_timeout_secs=600)

# ============================================================================
# Note: CRDs are managed by the Helm chart, not deployed separately
# ============================================================================

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

# Build API Server image
docker_build(
  ref='c8s-api-server',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='api-server',
)

# Build Controller image
docker_build(
  ref='c8s-controller',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='controller',
)

# Build Webhook image
docker_build(
  ref='c8s-webhook',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='webhook',
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
# Deploy C8S
# ============================================================================

namespace_create('c8s-system')

helm_resource(
  name='c8s',
  chart='./chart/c8s',
  flags=[
    '-f', './chart/c8s/values-dev.yaml',
    '-f', './tilt/c8s-values.yaml',
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
  resource_deps=['cert-manager'],
)

# ============================================================================
# Deploy Sample RepositoryConnection for dog-fooding
# ============================================================================
# This enables the webhook to trigger pipelines when code is pushed to GitHub
# Depends on the helm chart deployment to ensure C8S is ready first

k8s_yaml('./config/samples/repositoryconnection-c8s.yaml', allow_duplicates=True, resource_deps=['c8s'])

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

# nginx reverse proxy for consolidating API and webhook under one ngrok tunnel
local_resource(
  name='c8s-nginx-proxy',
  serve_cmd='docker run --rm --network host -v $(pwd)/tilt/nginx.conf:/etc/nginx/nginx.conf:ro nginx:latest',
  allow_parallel=True,
  resource_deps=['c8s-api-server-port-forward', 'c8s-webhook-port-forward'],
)

# ngrok tunnel for C8S services (optional)
# Routes through nginx reverse proxy on port 8888
# nginx handles path-based routing to API server and webhook

local_resource(
  name='c8s-ngrok-tunnel',
  serve_cmd='ngrok start --all --config=./tilt/ngrok-config.yml --config=$HOME/.config/ngrok/ngrok.yml --log=stdout',
  allow_parallel=True,
  resource_deps=['c8s-nginx-proxy'],
)


# ============================================================================
# Watch Files
# ============================================================================

watch_file('./chart/c8s')
watch_file('./tilt/c8s-values.yaml')
watch_file('./cmd/')
watch_file('./pkg/')
