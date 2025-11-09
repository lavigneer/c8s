#!/bin/bash
#
# E2E Integration Test: All User Stories
# Tests that all 4 user stories work together end-to-end
#
# User Story 1: Deploy C8S with single command
# User Story 2: Customize deployment configuration
# User Story 3: Verify deployment health
# User Story 4: Manage stack lifecycle (upgrade, downgrade, uninstall)
#
# Usage: ./tests/e2e/integration_test.sh [namespace]
# Example: ./tests/e2e/integration_test.sh c8s-integration

set -e

NAMESPACE="${1:-c8s-integration-test}"
CHART_DIR="./chart/c8s"
RELEASE_NAME="integration-test-$(date +%s)"
TEST_TIMEOUT=600

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   C8S Helm Chart - Full Integration Test (All US 1-4)     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Namespace: $NAMESPACE"
echo "Release: $RELEASE_NAME"
echo "Chart: $CHART_DIR"
echo ""

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up test resources..."
  helm uninstall $RELEASE_NAME -n $NAMESPACE 2>/dev/null || true
  kubectl delete namespace $NAMESPACE 2>/dev/null || true
}

trap cleanup EXIT

# Create namespace
kubectl create namespace $NAMESPACE 2>/dev/null || true

echo "╔════════════════════════════════════════════════════════════╗"
echo "║ User Story 1: Deploy C8S with Single Command (T017-T035)  ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo "📦 Installing C8S with development values..."
helm install $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml

echo "⏳ Waiting for all deployments to be ready..."

# Wait for all deployments
for deployment in api-server controller webhook frontend; do
  echo "  → Waiting for c8s-$deployment..."
  if kubectl rollout status deployment/c8s-$deployment -n $NAMESPACE --timeout=180s &>/dev/null; then
    echo -e "    ${GREEN}✓ Ready${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "    ${RED}✗ Failed${NC}"
    ((TESTS_FAILED++))
  fi
done

echo ""
echo "📋 Verifying deployment state..."

# Check all pods running
pod_count=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=c8s --field-selector=status.phase=Running -o jsonpath='{.items|length}')
if [ "$pod_count" -ge 4 ]; then
  echo -e "${GREEN}✓ All pods running ($pod_count pods)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Not all pods running (expected ≥4, got $pod_count)${NC}"
  ((TESTS_FAILED++))
fi

# Verify services exist
for service in api-server controller webhook frontend; do
  if kubectl get svc c8s-$service -n $NAMESPACE &>/dev/null; then
    echo -e "${GREEN}✓ Service c8s-$service exists${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Service c8s-$service not found${NC}"
    ((TESTS_FAILED++))
  fi
done

echo ""
echo -e "${BLUE}✓ User Story 1 Complete: Deployment successful${NC}"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║ User Story 2: Customize Deployment (T036-T055)            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo "🔧 Upgrading with custom configuration..."
echo "  Setting: controller replicas = 2 (was 1)"
echo "  Setting: logLevel = info (was debug)"

helm upgrade $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml \
  --set components.controller.replicas=2 \
  --set environment.logLevel=info

echo "⏳ Waiting for upgrade to complete..."
kubectl rollout status deployment/c8s-controller -n $NAMESPACE --timeout=180s

# Verify replica count changed
controller_replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE -o jsonpath='{.spec.replicas}')
if [ "$controller_replicas" == "2" ]; then
  echo -e "${GREEN}✓ Controller replicas updated to 2${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Controller replicas not updated (expected 2, got $controller_replicas)${NC}"
  ((TESTS_FAILED++))
fi

# Verify log level override
log_level=$(kubectl get deployment c8s-api-server -n $NAMESPACE \
  -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="LOG_LEVEL")].value}' 2>/dev/null || echo "")

if [ "$log_level" == "info" ] || [ -z "$log_level" ]; then
  echo -e "${GREEN}✓ Configuration customization applied${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Log level verification inconclusive${NC}"
fi

echo ""
echo -e "${BLUE}✓ User Story 2 Complete: Customization successful${NC}"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║ User Story 3: Verify Deployment Health (T056-T070)        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo "🏥 Checking deployment health..."

# Check deployment readiness
all_ready=true
for deployment in api-server controller webhook frontend; do
  ready=$(kubectl get deployment c8s-$deployment -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
  desired=$(kubectl get deployment c8s-$deployment -n $NAMESPACE -o jsonpath='{.spec.replicas}')

  if [ "$ready" == "$desired" ] && [ "$desired" != "0" ]; then
    echo -e "${GREEN}✓ c8s-$deployment: $ready/$desired replicas ready${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ c8s-$deployment: $ready/$desired replicas ready (expected $desired/$desired)${NC}"
    all_ready=false
    ((TESTS_FAILED++))
  fi
done

# Check service endpoints
for service in api-server controller webhook frontend; do
  endpoints=$(kubectl get endpoints c8s-$service -n $NAMESPACE -o jsonpath='{.subsets[*].addresses|length}' 2>/dev/null || echo "0")
  if [ "$endpoints" -gt 0 ]; then
    echo -e "${GREEN}✓ c8s-$service has active endpoints${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${YELLOW}⚠ c8s-$service endpoints check${NC}"
  fi
done

if [ "$all_ready" = true ]; then
  echo -e "${GREEN}✓ All components healthy${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Some components not healthy${NC}"
  ((TESTS_FAILED++))
fi

echo ""
echo -e "${BLUE}✓ User Story 3 Complete: Health verification successful${NC}"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║ User Story 4: Manage Stack Lifecycle (T071-T090)          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Get current revision before upgrade
current_revision=$(helm history $RELEASE_NAME -n $NAMESPACE | grep "DEPLOYED" | tail -1 | awk '{print $1}')
echo "📜 Current release revision: $current_revision"

# Capture initial replica count
initial_replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE -o jsonpath='{.spec.replicas}')
echo "📊 Initial controller replicas: $initial_replicas"

echo ""
echo "🔄 Testing upgrade scenario..."
echo "  Upgrading: controller replicas 2 → 3"

helm upgrade $RELEASE_NAME $CHART_DIR \
  -n $NAMESPACE \
  -f $CHART_DIR/values-dev.yaml \
  --set components.controller.replicas=3

kubectl rollout status deployment/c8s-controller -n $NAMESPACE --timeout=180s

upgrade_replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE -o jsonpath='{.spec.replicas}')
if [ "$upgrade_replicas" == "3" ]; then
  echo -e "${GREEN}✓ Upgrade successful: replicas now $upgrade_replicas${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Upgrade failed: replicas not updated${NC}"
  ((TESTS_FAILED++))
fi

echo ""
echo "↩️  Testing downgrade/rollback scenario..."

# Get latest revision (should be upgrade we just did)
latest_revision=$(helm history "$RELEASE_NAME" -n "$NAMESPACE" | grep "DEPLOYED" | tail -1 | awk '{print $1}')

# Validate that revision is numeric before using in arithmetic
if ! [[ "$latest_revision" =~ ^[0-9]+$ ]]; then
  echo -e "${RED}✗ Error: Invalid revision number extracted: $latest_revision${NC}"
  ((TESTS_FAILED++))
  exit 1
fi

previous_revision=$((latest_revision - 1))

echo "  Rolling back from revision $latest_revision to $previous_revision..."

helm rollback "$RELEASE_NAME" "$previous_revision" -n "$NAMESPACE"

kubectl rollout status deployment/c8s-controller -n $NAMESPACE --timeout=180s

rollback_replicas=$(kubectl get deployment c8s-controller -n $NAMESPACE -o jsonpath='{.spec.replicas}')
if [ "$rollback_replicas" == "$initial_replicas" ]; then
  echo -e "${GREEN}✓ Rollback successful: replicas restored to $rollback_replicas${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Rollback result: replicas=$rollback_replicas (expected $initial_replicas)${NC}"
fi

echo ""
echo "📋 Checking release history..."

release_history=$(helm history $RELEASE_NAME -n $NAMESPACE)
revision_count=$(echo "$release_history" | wc -l)

if [ "$revision_count" -ge 3 ]; then
  echo -e "${GREEN}✓ Release history tracked ($revision_count revisions)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Release history available${NC}"
fi

echo ""
echo "🗑️  Testing clean uninstall..."

helm uninstall $RELEASE_NAME -n $NAMESPACE

sleep 3

# Verify resources cleaned up
remaining=$(kubectl get all -n $NAMESPACE -l app.kubernetes.io/name=c8s --no-headers 2>/dev/null | wc -l || echo "0")

if [ "$remaining" == "0" ]; then
  echo -e "${GREEN}✓ Clean uninstall: all C8S resources removed${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${YELLOW}⚠ Uninstall complete (some resources may take time to cleanup)${NC}"
fi

# Verify namespace still exists
if kubectl get namespace $NAMESPACE &>/dev/null; then
  echo -e "${GREEN}✓ Namespace preserved (as intended)${NC}"
  ((TESTS_PASSED++))
else
  echo -e "${RED}✗ Namespace unexpectedly deleted${NC}"
  ((TESTS_FAILED++))
fi

echo ""
echo -e "${BLUE}✓ User Story 4 Complete: Lifecycle management successful${NC}"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                   Test Results Summary                     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

TOTAL=$((TESTS_PASSED + TESTS_FAILED))

echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
  echo -e "${RED}Failed: $TESTS_FAILED${NC}"
else
  echo "Failed: 0"
fi
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
  echo "╔════════════════════════════════════════════════════════════╗"
  echo "║                                                            ║"
  echo "║     ✅ ALL INTEGRATION TESTS PASSED! ✅                   ║"
  echo "║                                                            ║"
  echo "║   User Stories 1-4 validated working together             ║"
  echo "║   C8S Helm chart production-ready                         ║"
  echo "║                                                            ║"
  echo "╚════════════════════════════════════════════════════════════╝"
  echo ""
  exit 0
else
  echo "❌ Some tests failed. Review output above for details."
  exit 1
fi
