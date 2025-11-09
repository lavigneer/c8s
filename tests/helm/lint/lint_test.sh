#!/bin/bash
set -e

echo "Running Helm chart lint..."
helm lint ./chart/c8s

echo ""
echo "✓ Helm chart lint passed!"
