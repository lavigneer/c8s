#!/bin/bash
#
# E2E Test: Health Verification
# Tests the post-install hook health check functionality
#
# Usage: ./tests/e2e/health_test.sh [namespace]
# Example: ./tests/e2e/health_test.sh c8s-test

set -e

NAMESPACE="${1:-c8s-test-health}"
CHART_DIR="./chart/c8s"
RELEASE_NAME="test-health-$(date +%s)"
TEST_TIMEOUT=300

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=================================================="
echo "E2E Test: Health Verification"
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

# Test 1: Deploy with dev values and verify health check success (T065)
echo "=================================================="
echo "Test 1: Healthy Deployment"
echo "=================================================="
echo "Deploying C8S with dev values..."

# Create namespace
kubectl create namespace $NAMESPACE 2>/dev/null || true

# Deploy
helm install $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml \
  --set postInstallHook.timeout=120

echo ""
echo "Waiting for deployment to complete..."

# Wait for deployments to be ready
start_time=$(date +%s)
timeout=$((start_time + TEST_TIMEOUT))

while [ $(date +%s) -lt $timeout ]; do
  all_ready=true

  for deployment in api-server controller webhook frontend; do
    ready=$(kubectl get deployment c8s-$deployment -n $NAMESPACE \
      -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    desired=$(kubectl get deployment c8s-$deployment -n $NAMESPACE \
      -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

    if [ "$ready" != "$desired" ] || [ "$desired" == "0" ]; then
      all_ready=false
      break
    fi
  done

  if [ "$all_ready" = true ]; then
    echo -e "${GREEN}✓ All deployments are ready${NC}"
    break
  fi

  sleep 5
done

# Verify health check reported success
echo ""
echo "Verifying health check output..."
health_output=$(kubectl get pods -n $NAMESPACE \
  -l app.kubernetes.io/name=c8s,job-name=test-health-post-install \
  -o jsonpath='{.items[0].status.containerStatuses[0].state.terminated.message}' 2>/dev/null || echo "")

if echo "$health_output" | grep -q "All C8S components are ready"; then
  echo -e "${GREEN}✓ Health check reported success${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Health check did not report success${NC}"
  echo "Output: $health_output"
  ((TESTS_FAILED++))
fi

# Verify component status reported
if echo "$health_output" | grep -q "Ready:"; then
  echo -e "${GREEN}✓ Component status was reported${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Component status was not reported${NC}"
  ((TESTS_FAILED++))
fi

# Verify dashboard URL was displayed
if echo "$health_output" | grep -q "Dashboard"; then
  echo -e "${GREEN}✓ Dashboard URL was provided${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Dashboard URL was not provided${NC}"
  ((TESTS_FAILED++))
fi

# Test 2: Verify readiness probes disabled (T065)
echo ""
echo "=================================================="
echo "Test 2: Readiness Probes"
echo "=================================================="

for deployment in api-server controller webhook; do
  has_probe=$(kubectl get deployment c8s-$deployment -n $NAMESPACE \
    -o jsonpath='{.spec.template.spec.containers[0].readinessProbe}' 2>/dev/null)

  if [ -z "$has_probe" ] || [ "$has_probe" == "null" ]; then
    echo -e "${GREEN}✓ Readiness probe disabled for c8s-$deployment${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${YELLOW}⚠ Readiness probe is configured for c8s-$deployment${NC}"
  fi
done

# Test 3: Check service endpoints (T067)
echo ""
echo "=================================================="
echo "Test 3: Service Endpoints"
echo "=================================================="

for service in api-server controller webhook frontend; do
  endpoints=$(kubectl get endpoints c8s-$service -n $NAMESPACE \
    -o jsonpath='{.subsets[0].addresses[0].ip}' 2>/dev/null || echo "")

  if [ -n "$endpoints" ]; then
    echo -e "${GREEN}✓ Service c8s-$service has endpoints${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Service c8s-$service has no endpoints${NC}"
    ((TESTS_FAILED++))
  fi
done

# Summary
echo ""
echo "=================================================="
echo "Test Results"
echo "=================================================="
TESTS_PASSED=${TESTS_PASSED:-0}
TESTS_FAILED=${TESTS_FAILED:-0}
TOTAL=$((TESTS_PASSED + TESTS_FAILED))

echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
  echo -e "${RED}Failed: $TESTS_FAILED${NC}"
  exit 1
else
  echo "Failed: 0"
  echo -e "${GREEN}✓ All health verification tests passed!${NC}"
  exit 0
fi
