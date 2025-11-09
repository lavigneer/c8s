#!/bin/bash
# Cert-Manager Integration Tests for C8S Webhook TLS Certificates
# Tests automatic certificate provisioning and caBundle injection

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CERT_MANAGER_NS="cert-manager"
C8S_NS="c8s-system"
CHART_DIR="./chart/c8s"
TESTS_PASSED=0
TESTS_FAILED=0

echo "╔════════════════════════════════════════════════════════════╗"
echo "║          C8S Cert-Manager Integration Tests                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Test 1: Check if cert-manager is installed
test_cert_manager_installed() {
  echo "Test 1: Verify cert-manager is installed..."

  if kubectl get ns "$CERT_MANAGER_NS" >/dev/null 2>&1; then
    if kubectl get pod -n "$CERT_MANAGER_NS" -l app=cert-manager >/dev/null 2>&1; then
      echo -e "${GREEN}✓ Cert-manager is installed${NC}"
      ((TESTS_PASSED++))
      return 0
    fi
  fi

  echo -e "${YELLOW}⚠ Cert-manager not found. Installing...${NC}"

  # Install cert-manager
  helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
  helm repo update 2>/dev/null || true
  helm install cert-manager jetstack/cert-manager \
    --namespace "$CERT_MANAGER_NS" \
    --create-namespace \
    --set installCRDs=true \
    --wait \
    --timeout 300s

  echo -e "${GREEN}✓ Cert-manager installed${NC}"
  ((TESTS_PASSED++))
}

# Test 2: Deploy C8S with cert-manager
test_c8s_deployment_with_certmanager() {
  echo ""
  echo "Test 2: Deploy C8S with cert-manager values..."

  # Clean up previous deployment if exists
  helm uninstall c8s -n "$C8S_NS" 2>/dev/null || true
  kubectl delete ns "$C8S_NS" 2>/dev/null || true
  sleep 2

  if helm install c8s "$CHART_DIR" \
    -n "$C8S_NS" \
    --create-namespace \
    -f "$CHART_DIR/values-certmanager.yaml" \
    --wait \
    --timeout 300s; then
    echo -e "${GREEN}✓ C8S deployed with cert-manager${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ C8S deployment failed${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 3: Verify Certificate CRD is created
test_certificate_created() {
  echo ""
  echo "Test 3: Verify Certificate resource is created..."

  if kubectl get certificate c8s-webhook-tls -n "$C8S_NS" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Certificate resource created${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Certificate resource not found${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 4: Verify Certificate is ready
test_certificate_ready() {
  echo ""
  echo "Test 4: Wait for certificate to be ready..."

  # Wait for certificate to be ready
  if kubectl wait --for=condition=ready certificate c8s-webhook-tls \
    -n "$C8S_NS" --timeout=120s 2>/dev/null; then
    echo -e "${GREEN}✓ Certificate is ready${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Certificate is not ready${NC}"

    # Show diagnostic info
    echo ""
    echo "Certificate status:"
    kubectl describe certificate c8s-webhook-tls -n "$C8S_NS"
    echo ""
    echo "Cert-manager logs:"
    kubectl logs -n "$CERT_MANAGER_NS" -l app=cert-manager --tail=20

    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 5: Verify TLS Secret is created
test_tls_secret_created() {
  echo ""
  echo "Test 5: Verify TLS secret is created..."

  if kubectl get secret c8s-webhook-tls -n "$C8S_NS" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ TLS secret created${NC}"
    ((TESTS_PASSED++))

    # Show secret details
    echo "  Secret contains:"
    kubectl get secret c8s-webhook-tls -n "$C8S_NS" -o jsonpath='{.data}' | grep -o '"[^"]*"' | head -3
  else
    echo -e "${RED}✗ TLS secret not found${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 6: Verify ValidatingWebhookConfiguration has annotation
test_webhook_config_annotation() {
  echo ""
  echo "Test 6: Verify ValidatingWebhookConfiguration has cert-manager annotation..."

  annotation=$(kubectl get validatingwebhookconfigurations c8s-validating-webhook \
    -o jsonpath='{.metadata.annotations.cert-manager\.io/inject-ca-from}' 2>/dev/null || echo "")

  if [ -n "$annotation" ]; then
    echo -e "${GREEN}✓ Annotation found: $annotation${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Annotation not found${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 7: Verify caBundle is injected
test_cabundle_injected() {
  echo ""
  echo "Test 7: Verify caBundle is injected into ValidatingWebhookConfiguration..."

  # Wait a bit for caBundle to be injected
  sleep 5

  cabundle=$(kubectl get validatingwebhookconfigurations c8s-validating-webhook \
    -o jsonpath='{.webhooks[0].clientConfig.caBundle}' 2>/dev/null || echo "")

  if [ -n "$cabundle" ] && [ "$cabundle" != "\"\"" ]; then
    echo -e "${GREEN}✓ caBundle is populated${NC}"
    echo "  First 50 chars: ${cabundle:0:50}..."
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ caBundle is empty or not found${NC}"

    # Show diagnostic info
    echo ""
    echo "ValidatingWebhookConfiguration:"
    kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep -A 5 caBundle

    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 8: Verify Issuer is created
test_issuer_created() {
  echo ""
  echo "Test 8: Verify Issuer resource is created..."

  if kubectl get issuer c8s-webhook-selfsigned -n "$C8S_NS" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Issuer resource created${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Issuer resource not found${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 9: Verify webhook pod can read certificate
test_webhook_certificate_mounted() {
  echo ""
  echo "Test 9: Verify webhook pod has certificate mounted..."

  # Get a webhook pod
  webhook_pod=$(kubectl get pod -n "$C8S_NS" -l app.kubernetes.io/name=c8s,app.kubernetes.io/component=webhook \
    -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

  if [ -z "$webhook_pod" ]; then
    echo -e "${RED}✗ No webhook pod found${NC}"
    ((TESTS_FAILED++))
    return 1
  fi

  # Check if certificate files exist in pod
  if kubectl exec -n "$C8S_NS" "$webhook_pod" -- test -f /etc/webhook/certs/tls.crt >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Certificate mounted in webhook pod: $webhook_pod${NC}"
    ((TESTS_PASSED++))

    # Show certificate info
    cert_subject=$(kubectl exec -n "$C8S_NS" "$webhook_pod" -- \
      openssl x509 -in /etc/webhook/certs/tls.crt -noout -subject 2>/dev/null || echo "N/A")
    echo "  Certificate subject: $cert_subject"
  else
    echo -e "${RED}✗ Certificate not mounted in webhook pod${NC}"
    ((TESTS_FAILED++))
    return 1
  fi
}

# Test 10: Verify webhook pod is running
test_webhook_pod_running() {
  echo ""
  echo "Test 10: Verify webhook pod is running..."

  if kubectl rollout status deployment/c8s-webhook -n "$C8S_NS" --timeout=120s >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Webhook deployment is running${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Webhook deployment is not ready${NC}"

    # Show diagnostic info
    echo ""
    echo "Webhook pod status:"
    kubectl describe pod -n "$C8S_NS" -l app.kubernetes.io/component=webhook

    ((TESTS_FAILED++))
    return 1
  fi
}

# Run all tests
run_tests() {
  test_cert_manager_installed
  test_c8s_deployment_with_certmanager
  test_certificate_created
  test_certificate_ready
  test_tls_secret_created
  test_webhook_config_annotation
  test_cabundle_injected
  test_issuer_created
  test_webhook_certificate_mounted
  test_webhook_pod_running
}

# Print summary
print_summary() {
  echo ""
  echo "╔════════════════════════════════════════════════════════════╗"
  echo "║                      Test Summary                          ║"
  echo "╚════════════════════════════════════════════════════════════╝"
  echo ""
  echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
  echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
  echo -e "Total Tests:  $((TESTS_PASSED + TESTS_FAILED))"
  echo ""

  if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Cert-manager integration is working correctly."
    echo "Webhook TLS certificates are automatically managed."
    return 0
  else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Check the diagnostic information above for details."
    return 1
  fi
}

# Cleanup
cleanup() {
  echo ""
  echo "Cleaning up test deployment..."
  helm uninstall c8s -n "$C8S_NS" 2>/dev/null || true
  kubectl delete ns "$C8S_NS" 2>/dev/null || true
}

# Main execution
main() {
  trap cleanup EXIT

  run_tests
  print_summary
}

main
