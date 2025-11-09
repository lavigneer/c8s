# Tilt configuration for C8S development
#
# This Tiltfile integrates the C8S Helm chart for local development
#
# Usage:
#   tilt up              # Start Tilt with C8S deployment
#   tilt down            # Stop Tilt and remove resources
#   tilt logs c8s        # View logs for C8S
#   tilt trigger c8s     # Manually trigger C8S update
#

# Load Helm resource extension for Helm chart support
load('ext://helm_resource', 'helm_resource')

# Deploy C8S using Helm chart from ./chart/c8s
# Uses development values with Tilt overrides for faster iteration
helm_resource(
  name='c8s',
  chart='./chart/c8s',
  flags=[
    '-f', './chart/c8s/values-dev.yaml',
    '-f', './tilt/c8s-values.yaml',
    '--create-namespace',
  ],
  namespace='c8s-system'
)

# Watch Helm chart files for automatic redeployment on changes
watch_file('./chart/c8s')
watch_file('./tilt/c8s-values.yaml')
