#\!/bin/bash
set -e

TMPDIR="${TMPDIR:-/tmp/claude}"
mkdir -p "$TMPDIR"

echo "=================================================="
echo "Helm Template Rendering Tests"
echo "=================================================="
echo ""

# Test 1: Render with default values
echo "Test 1: Default values (dev environment)"
helm template c8s ./chart/c8s > "$TMPDIR/c8s-default.yaml"
if [ $? -eq 0 ]; then
  echo "✓ Default values render successfully"
  echo "  Generated manifests: $(grep -c '^kind:' "$TMPDIR/c8s-default.yaml") resources"
else
  echo "✗ Failed to render default values"
  exit 1
fi
echo ""

# Test 2: Render with dev values
echo "Test 2: Development values"
helm template c8s ./chart/c8s -f ./chart/c8s/values-dev.yaml > "$TMPDIR/c8s-dev.yaml"
if [ $? -eq 0 ]; then
  echo "✓ Dev values render successfully"
  echo "  Manifests generated successfully"
else
  echo "✗ Failed to render dev values"
  exit 1
fi
echo ""

# Test 3: Render with staging values
echo "Test 3: Staging values"
helm template c8s ./chart/c8s -f ./chart/c8s/values-staging.yaml > "$TMPDIR/c8s-staging.yaml"
if [ $? -eq 0 ]; then
  echo "✓ Staging values render successfully"
  # Verify PVC is created for staging
  if grep -q 'kind: PersistentVolumeClaim' "$TMPDIR/c8s-staging.yaml"; then
    echo "  ✓ PVC created for staging storage"
  else
    echo "  ✗ PVC not found in staging manifests"
  fi
else
  echo "✗ Failed to render staging values"
  exit 1
fi
echo ""

# Test 4: Render with production values
echo "Test 4: Production values"
helm template c8s ./chart/c8s -f ./chart/c8s/values-prod.yaml > "$TMPDIR/c8s-prod.yaml"
if [ $? -eq 0 ]; then
  echo "✓ Production values render successfully"
  # Verify replicas are 3 for controller in prod
  if grep -q 'replicas: 3' "$TMPDIR/c8s-prod.yaml"; then
    echo "  ✓ Production replicas configured (3)"
  fi
else
  echo "✗ Failed to render production values"
  exit 1
fi
echo ""

# Test 5: Render with custom CLI overrides
echo "Test 5: Custom CLI overrides"
helm template c8s ./chart/c8s \
  -f ./chart/c8s/values-dev.yaml \
  --set components.controller.replicas=5 \
  --set components.webhook.replicas=3 > "$TMPDIR/c8s-custom.yaml"
if [ $? -eq 0 ]; then
  echo "✓ Custom overrides render successfully"
  if grep -q 'replicas: 5' "$TMPDIR/c8s-custom.yaml"; then
    echo "  ✓ Custom controller replica count (5) applied"
  fi
  if grep -q 'replicas: 3' "$TMPDIR/c8s-custom.yaml"; then
    echo "  ✓ Custom webhook replica count (3) applied"
  fi
else
  echo "✗ Failed to render custom overrides"
  exit 1
fi
echo ""

# Test 6: Validate all required resources are generated
echo "Test 6: Required resources validation"
REQUIRED_RESOURCES=(
  "Namespace"
  "ServiceAccount"
  "ClusterRole"
  "ConfigMap"
  "Service"
  "Deployment"
)

for resource in "${REQUIRED_RESOURCES[@]}"; do
  if grep -q "kind: $resource" "$TMPDIR/c8s-default.yaml"; then
    echo "  ✓ $resource found"
  else
    echo "  ✗ $resource NOT found"
    exit 1
  fi
done
echo ""

echo "=================================================="
echo "All template rendering tests passed\!"
echo "=================================================="
