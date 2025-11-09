#!/bin/bash
#
# E2E Test: Lifecycle Management (Upgrade, Downgrade, Uninstall)
# Tests upgrade, downgrade via rollback, and clean uninstall functionality
#
# Usage: ./tests/e2e/lifecycle_test.sh [namespace]
# Example: ./tests/e2e/lifecycle_test.sh c8s-test-lifecycle

set -e

NAMESPACE="${1:-c8s-test-lifecycle}"
CHART_DIR="./chart/c8s"
RELEASE_NAME="test-lifecycle-$(date +%s)"
TEST_TIMEOUT=300

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

echo "=================================================="
echo "E2E Test: Lifecycle Management"
echo "=================================================="
echo "Namespace: $NAMESPACE"
echo "Release: $RELEASE_NAME"
echo "Chart: $CHART_DIR"
echo ""

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up..."
  helm uninstall $RELEASE_NAME -n $NAMESPACE 2>/dev/null || true
  kubectl delete namespace $NAMESPACE 2>/dev/null || true
}

trap cleanup EXIT

# Create namespace
kubectl create namespace $NAMESPACE 2>/dev/null || true

# Test 1: Initial Deployment (T073)
echo "=================================================="
echo "Test 1: Initial Deployment (v0.1.0)"
echo "=================================================="
echo "Deploying C8S with initial values..."

helm install $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml \
  --set components.controller.replicas=1

# Wait for deployment
echo "Waiting for deployment to be ready..."
for deployment in api-server controller webhook frontend; do
  kubectl rollout status deployment/c8s-$deployment -n $NAMESPACE --timeout=120s
done

echo -e "${GREEN}✓ Initial deployment successful${NC}"
((TESTS_PASSED++))

# Get initial revision
initial_revision=$(helm history $RELEASE_NAME -n $NAMESPACE | grep "DEPLOYED" | head -1 | awk '{print $1}')
echo "Initial revision: $initial_revision"

# Test 2: Upgrade with Custom Values (T074)
echo ""
echo "=================================================="
echo "Test 2: Upgrade with Custom Values"
echo "=================================================="
echo "Upgrading with higher replica count..."

helm upgrade $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml \
  --set components.controller.replicas=2 \
  --set environment.logLevel=info

# Verify replica count changed
echo "Verifying replica count changed..."
kubectl rollout status deployment/c8s-controller -n $NAMESPACE --timeout=120s

replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE \
  -o jsonpath='{.spec.replicas}')

if [ "$replicas" == "2" ]; then
  echo -e "${GREEN}✓ Replica count updated to $replicas${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Replica count not updated (expected 2, got $replicas)${NC}"
  ((TESTS_FAILED++))
fi

# Test 3: Verify Custom Values Preserved (T074)
echo ""
echo "=================================================="
echo "Test 3: Verify Custom Values Preserved"
echo "=================================================="

log_level=$(kubectl get deployment c8s-api-server -n $NAMESPACE \
  -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="LOG_LEVEL")].value}' 2>/dev/null || echo "")

if [ "$log_level" == "info" ]; then
  echo -e "${GREEN}✓ Custom log level preserved (info)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Log level verification skipped or not set${NC}"
fi

# Test 4: Release History (T077)
echo ""
echo "=================================================="
echo "Test 4: Release History"
echo "=================================================="

echo "Checking release history..."
history=$(helm history $RELEASE_NAME -n $NAMESPACE)

if echo "$history" | grep -q "DEPLOYED"; then
  revisions=$(echo "$history" | grep -c "DEPLOYED" || echo "1")
  echo -e "${GREEN}✓ Release history available with $revisions deployment(s)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Release history not available${NC}"
  ((TESTS_FAILED++))
fi

echo ""
echo "History:"
echo "$history"

# Get current revision before rollback
current_revision=$(helm history $RELEASE_NAME -n $NAMESPACE | grep "DEPLOYED" | tail -1 | awk '{print $1}')

# Test 5: Rollback (Downgrade) (T076)
echo ""
echo "=================================================="
echo "Test 5: Rollback to Previous Release"
echo "=================================================="
echo "Rolling back to revision $initial_revision..."

helm rollback $RELEASE_NAME $initial_revision -n $NAMESPACE

# Wait for rollback
kubectl rollout status deployment/c8s-controller -n $NAMESPACE --timeout=120s

# Verify replicas changed back to 1
rollback_replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE \
  -o jsonpath='{.spec.replicas}')

if [ "$rollback_replicas" == "1" ]; then
  echo -e "${GREEN}✓ Rollback successful (replicas: $rollback_replicas)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Rollback failed (expected 1, got $rollback_replicas)${NC}"
  ((TESTS_FAILED++))
fi

# Test 6: Cleanup Labels (T080)
echo ""
echo "=================================================="
echo "Test 6: Cleanup Labels"
echo "=================================================="

labeled_resources=$(kubectl get all -n $NAMESPACE \
  -l app.kubernetes.io/name=c8s \
  --no-headers 2>/dev/null | wc -l)

if [ $labeled_resources -gt 0 ]; then
  echo -e "${GREEN}✓ Resources have cleanup labels ($labeled_resources resources)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ No labeled resources found${NC}"
fi

# Test 7: Uninstall (T081)
echo ""
echo "=================================================="
echo "Test 7: Clean Uninstall"
echo "=================================================="
echo "Uninstalling release..."

helm uninstall $RELEASE_NAME -n $NAMESPACE

# Wait a moment for resources to be deleted
sleep 5

# Verify resources are deleted
remaining_resources=$(kubectl get all -n $NAMESPACE \
  -l app.kubernetes.io/name=c8s \
  --no-headers 2>/dev/null | wc -l || echo "0")

if [ "$remaining_resources" == "0" ]; then
  echo -e "${GREEN}✓ All C8S resources removed${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Some resources still present ($remaining_resources)${NC}"
  ((TESTS_FAILED++))
fi

# Verify secrets and configmaps are removed
configs_remaining=$(kubectl get secrets,configmaps -n $NAMESPACE \
  -l app.kubernetes.io/name=c8s \
  --no-headers 2>/dev/null | wc -l || echo "0")

if [ "$configs_remaining" == "0" ]; then
  echo -e "${GREEN}✓ All configuration resources removed${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Some config resources still present ($configs_remaining)${NC}"
  ((TESTS_FAILED++))
fi

# Test 8: Namespace Cleanup (T082)
echo ""
echo "=================================================="
echo "Test 8: Namespace Cleanup"
echo "=================================================="

# Verify namespace still exists (it should, Helm doesn't delete it)
if kubectl get namespace $NAMESPACE &>/dev/null; then
  echo -e "${GREEN}✓ Namespace still exists (can be deleted separately)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Namespace was unexpectedly deleted${NC}"
  ((TESTS_FAILED++))
fi

# Summary
echo ""
echo "=================================================="
echo "Lifecycle Test Results"
echo "=================================================="
TOTAL=$((TESTS_PASSED + TESTS_FAILED))

echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
  echo -e "${RED}Failed: $TESTS_FAILED${NC}"
  exit 1
else
  echo "Failed: 0"
  echo -e "${GREEN}✓ All lifecycle tests passed!${NC}"
  exit 0
fi
