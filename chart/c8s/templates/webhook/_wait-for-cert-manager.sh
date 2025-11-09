#!/bin/bash
# Wait for cert-manager CRDs to be installed
# This hook ensures cert-manager is ready before we create Certificate/Issuer resources

set -e

echo "Waiting for cert-manager CRDs to be available..."

# Wait up to 5 minutes for cert-manager to be ready
TIMEOUT=300
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $TIMEOUT ]; do
  # Check if Certificate CRD exists
  if kubectl get crd certificates.cert-manager.io >/dev/null 2>&1; then
    # Check if Issuer CRD exists
    if kubectl get crd issuers.cert-manager.io >/dev/null 2>&1; then
      echo "✓ cert-manager CRDs are available"
      exit 0
    fi
  fi

  echo "Waiting for cert-manager CRDs... ($ELAPSED/$TIMEOUT seconds)"
  sleep $INTERVAL
  ELAPSED=$((ELAPSED + INTERVAL))
done

echo "✗ Timeout waiting for cert-manager CRDs"
exit 1
