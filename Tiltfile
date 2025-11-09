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

GHCR_REGISTRY = os.getenv('GHCR_REGISTRY', 'ghcr.io/anthropics')
IMAGE_TAG = os.getenv('IMAGE_TAG', 'latest')

# Build context for all images
BUILD_DIR = '.'

# ============================================================================
# Build Images
# ============================================================================

print("Building C8S images...")
print(f"Registry: {GHCR_REGISTRY}")
print(f"Tag: {IMAGE_TAG}")

# Build API Server image
docker_build(
  ref='c8s-api-server',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='api-server',
  only=[
    'cmd/api-server/',
    'cmd/webhook/',
    'cmd/controller/',
    'pkg/',
    'hack/',
    'go.mod',
    'go.sum',
    'Makefile',
    'PROJECT',
  ],
)

# Build Controller image
docker_build(
  ref='c8s-controller',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='controller',
  only=[
    'cmd/controller/',
    'pkg/',
    'hack/',
    'go.mod',
    'go.sum',
    'Makefile',
    'PROJECT',
  ],
)

# Build Webhook image
docker_build(
  ref='c8s-webhook',
  context=BUILD_DIR,
  dockerfile='./Dockerfile',
  target='webhook',
  only=[
    'cmd/webhook/',
    'pkg/',
    'hack/',
    'go.mod',
    'go.sum',
    'Makefile',
    'PROJECT',
  ],
)

# Build Frontend image (if Dockerfile exists for it)
# For now, we use the chart's default image reference
# You can add a separate build when the frontend Dockerfile is available

# ============================================================================
# Create Namespace
# ============================================================================

namespace_create('c8s-system')

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
  ],
  image_keys=[
    ('components.apiServer.image.registry', 'components.apiServer.image.repository', 'components.apiServer.image.tag'),
    ('components.controller.image.registry', 'components.controller.image.repository', 'components.controller.image.tag'),
    ('components.webhook.image.registry', 'components.webhook.image.repository', 'components.webhook.image.tag'),
  ],
  image_refs=[
    'c8s-api-server',
    'c8s-controller',
    'c8s-webhook',
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

print("""
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
   """
) + GHCR_REGISTRY + """/c8s-api-server:""" + IMAGE_TAG + """
   """ + GHCR_REGISTRY + """/c8s-controller:""" + IMAGE_TAG + """
   """ + GHCR_REGISTRY + """/c8s-webhook:""" + IMAGE_TAG + """

💡 To push to GHCR:
   docker tag c8s-api-server """ + GHCR_REGISTRY + """/c8s-api-server:""" + IMAGE_TAG + """
   docker push """ + GHCR_REGISTRY + """/c8s-api-server:""" + IMAGE_TAG + """

   # Or use Tilt's built-in push (configure in Tiltfile custom_build)

🔗 Access dashboard:
   kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80
   Then visit http://localhost:3000
""")
