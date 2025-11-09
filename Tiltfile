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
# Environment Variables (optional):
#   GHCR_REGISTRY=ghcr.io/anthropics  # GitHub Container Registry path
#   IMAGE_TAG=v0.1.0                   # Image tag (default: latest)

# Load extensions
load('ext://namespace', 'namespace_create')
load('ext://helm_resource', 'helm_resource')

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
# Create Namespace
# ============================================================================

namespace_create('c8s-system')

# ============================================================================
# Install CRDs
# ============================================================================

local_resource(
  name='crds',
  cmd='kubectl apply -f ./deploy/crds.yaml',
  trigger_mode=TRIGGER_MODE_OFF,
  labels=['setup'],
)

# ============================================================================
# Deploy Helm Chart
# ============================================================================

# Create values file with local image references
local_values = {
  'components': {
    'apiServer': {
      'image': {
        'registry': 'localhost:5000',  # Local Docker registry
        'repository': 'c8s-api-server',
        'tag': IMAGE_TAG,
      }
    },
    'controller': {
      'image': {
        'registry': 'localhost:5000',
        'repository': 'c8s-controller',
        'tag': IMAGE_TAG,
      }
    },
    'webhook': {
      'image': {
        'registry': 'localhost:5000',
        'repository': 'c8s-webhook',
        'tag': IMAGE_TAG,
      }
    },
  }
}

# Deploy C8S using Helm chart
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
  namespace='c8s-system'
)

# ============================================================================
# Watch Files
# ============================================================================

watch_file('./chart/c8s')
watch_file('./tilt/c8s-values.yaml')
watch_file('./cmd/')
watch_file('./pkg/')

# ============================================================================
# Print Instructions
# ============================================================================

local_resource(
  name='info',
  cmd='echo "C8S deployment info:"; kubectl get all -n c8s-system; echo ""; echo "Port forward to frontend:"; echo "kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80"',
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=['info'],
)

info_text = """
╔════════════════════════════════════════════════════════════════════════════╗
║                   C8S Tilt Development Environment                         ║
╚════════════════════════════════════════════════════════════════════════════╝

✅ Building images locally:
   - c8s-api-server
   - c8s-controller
   - c8s-webhook

📦 Deploying to namespace: c8s-system

🚀 Available commands:
   tilt up              - Start deployment
   tilt down            - Stop deployment
   tilt logs <service>  - View service logs
   tilt trigger c8s     - Force redeploy

📝 Images will be available at:
   {0}/c8s-api-server:{1}
   {0}/c8s-controller:{1}
   {0}/c8s-webhook:{1}

💡 To push to GHCR:
   docker tag c8s-api-server {0}/c8s-api-server:{1}
   docker push {0}/c8s-api-server:{1}

   Or create a version tag (GitHub Actions will auto-build and push):
   git tag v0.1.0 && git push --tags

🔗 Access dashboard:
   kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80
   Then visit http://localhost:3000

ℹ️  Repository: {2}
   Registry: {0}
""".format(GHCR_REGISTRY, IMAGE_TAG, GITHUB_REPO)

print(info_text)
