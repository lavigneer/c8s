# Tilt configuration for C8S development
#
# This Tiltfile integrates the C8S Helm chart for local development
#
# Usage:
#   tilt up              # Start Tilt with C8S deployment
#   tilt down            # Stop Tilt and remove resources
#   tilt logs <service>  # View logs for a service
#   tilt trigger c8s     # Manually trigger C8S update
#

# Load Helm resource extension
load('ext://helm_resource', 'helm_resource')

# Configure kubectl context
allow_k8s_context('kind-kind')
allow_k8s_context('docker-desktop')
allow_k8s_context('minikube')

# Deploy C8S using Helm chart
helm_resource(
  name='c8s',
  chart_dir='./chart/c8s',
  flags=[
    '-f', './chart/c8s/values-dev.yaml',
    '-f', './tilt/c8s-values.yaml',
  ],
  namespace='c8s-system',
  labels=['backend']
)

# Watch Helm chart files for changes
watch_file('./chart/c8s')
watch_file('./tilt')

# Set default port forwards (optional)
k8s_resource('c8s-frontend', port_forwards='3000:80')
