#!/bin/bash
set -e

NAMESPACE="${C8S_NAMESPACE:-c8s-system}"
RELEASE_NAME="${C8S_RELEASE_NAME:-c8s}"
TIMEOUT="${C8S_TIMEOUT:-300}"
VALUES_FILE="${C8S_VALUES_FILE:-./chart/c8s/values-dev.yaml}"

echo "=================================================="
echo "C8S Helm Chart Deployment Test"
echo "=================================================="
echo "Release: $RELEASE_NAME"
echo "Namespace: $NAMESPACE"
echo "Timeout: ${TIMEOUT}s"
echo "Values: $VALUES_FILE"
echo ""

# Check if helm is installed
if ! command -v helm &> /dev/null; then
  echo "Error: helm is not installed"
  exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
  echo "Error: kubectl is not installed"
  exit 1
fi

# Create namespace if it doesn't exist
echo "Creating namespace $NAMESPACE..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Install the Helm chart
echo "Installing Helm chart..."
helm install $RELEASE_NAME ./chart/c8s -f $VALUES_FILE -n $NAMESPACE

# Wait for deployment to be ready
echo ""
echo "Waiting for deployments to be ready (timeout: ${TIMEOUT}s)..."

kubectl wait --for=condition=available --timeout=${TIMEOUT}s \
  deployment/c8s-api-server \
  deployment/c8s-controller \
  deployment/c8s-webhook \
  deployment/c8s-frontend \
  -n $NAMESPACE || {
  echo "✗ Deployment failed to become ready"
  echo ""
  echo "Debugging information:"
  kubectl describe deployment -n $NAMESPACE
  kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=c8s --tail=20
  exit 1
}

echo "✓ All deployments are ready!"
echo ""

# Run health check
echo "Running health verification..."
helm get hooks $RELEASE_NAME -n $NAMESPACE

# Cleanup
echo ""
echo "Cleaning up..."
helm uninstall $RELEASE_NAME -n $NAMESPACE
kubectl delete namespace $NAMESPACE

echo ""
echo "✓ Deployment test passed!"
