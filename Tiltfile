# Tilt configuration for C8S development

# Load the C8S Helm chart
load('ext://helm_resource', 'helm_resource', 'helm_repo')

# Add Helm chart
helm_resource(
  'c8s',
  './chart/c8s',
  flags=[
    '-f', './chart/c8s/values-dev.yaml',
    '-f', './tilt/c8s-values.yaml',
  ],
  namespace='c8s-system'
)

# Watch for changes to Helm chart files
watch_file('./chart/c8s')

# Set resource limits for Tilt
resources = config.analysis.resources
if resources:
  # Tilt development defaults
  for resource in resources:
    resource.set_resource_requests('150m', '256Mi')
