# C8S Helm Chart - Testing Matrix

Comprehensive testing documentation and results for the C8S Helm chart across multiple Kubernetes distributions and configurations.

## Test Execution Environment

- **Helm Version**: 3.x+
- **Kubernetes Version**: 1.24+
- **Chart Version**: 0.1.0
- **Test Date**: [Document when tests were run]
- **Test Duration**: [Document how long tests took]

---

## Test Categories

### 1. Lint & Validation Tests

**Purpose**: Verify chart structure and syntax

**Tests**:
- [ ] `helm lint ./chart/c8s` passes without errors
- [ ] `helm lint ./chart/c8s` passes without critical warnings
- [ ] Chart.yaml valid and complete
- [ ] values.yaml valid YAML syntax
- [ ] All templates valid YAML
- [ ] No hardcoded values in templates (except defaults)

**Command**:
```bash
helm lint ./chart/c8s
```

**Expected Result**: `1 chart(s) linted, 0 chart(s) failed`

---

### 2. Template Rendering Tests

**Purpose**: Verify templates render correctly with different values

**Test Scenarios**:
- [ ] Render with values-dev.yaml
- [ ] Render with values-staging.yaml
- [ ] Render with values-prod.yaml
- [ ] Render with CLI overrides (`--set`)
- [ ] Render with multiple values files

**Command**:
```bash
helm template c8s ./chart/c8s -f values-dev.yaml
```

**Expected Result**: Valid Kubernetes manifests output without errors

---

### 3. Functional Tests - Development Environment (T098c - Docker Desktop)

**Setup**:
```bash
# Assuming Docker Desktop Kubernetes is enabled
kubectl cluster-info
```

**Tests**:
- [ ] Install with dev values completes successfully
- [ ] All 4 deployments reach Ready state within 5 minutes
- [ ] Health check hook reports success
- [ ] Dashboard accessible via port-forward
- [ ] All pods are Running with Ready containers
- [ ] Services have endpoints assigned
- [ ] ConfigMaps and Secrets created correctly

**Commands**:
```bash
helm install c8s ./chart/c8s -f values-dev.yaml -n c8s-system --create-namespace
kubectl rollout status deployment/c8s-api-server -n c8s-system --timeout=300s
kubectl get all -n c8s-system
```

**Test Results**:
- [ ] Status: PASS/FAIL
- [ ] Time to ready: _____ seconds
- [ ] Issues encountered: _____

---

### 4. Functional Tests - Local Kubernetes (T098a - k3s)

**Prerequisites**:
```bash
# Install k3s
curl -sfL https://get.k3s.io | sh -

# Configure kubectl
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

**Tests**:
- [ ] Install with staging values completes successfully
- [ ] All deployments reach Ready state
- [ ] PVC creation and binding works
- [ ] Health check reports success
- [ ] Rolling update during upgrade works
- [ ] Resources cleaned up on uninstall

**Commands**:
```bash
helm install c8s ./chart/c8s -f values-staging.yaml -n c8s-system --create-namespace
kubectl get pvc -n c8s-system
```

**Test Results**:
- [ ] Status: PASS/FAIL
- [ ] PVC binding time: _____ seconds
- [ ] Issues encountered: _____

---

### 5. Functional Tests - Local Kubernetes (T098b - kind)

**Prerequisites**:
```bash
# Install kind
go install sigs.k8s.io/kind@latest

# Create cluster
kind create cluster
```

**Tests**:
- [ ] Install with prod values completes successfully
- [ ] HA setup works (3 replicas)
- [ ] LoadBalancer service type works with kind
- [ ] Rolling update maintains availability
- [ ] Helm upgrade preserves custom values
- [ ] Helm rollback restores previous version

**Commands**:
```bash
kind create cluster
helm install c8s ./chart/c8s -f values-prod.yaml -n c8s-system --create-namespace
kubectl get service -n c8s-system
```

**Test Results**:
- [ ] Status: PASS/FAIL
- [ ] Cluster setup time: _____ seconds
- [ ] Issues encountered: _____

---

### 6. Functional Tests - Cloud (T098d - EKS/GKE)

**Cloud Platform**: _____ (EKS/GKE/AKS)

**Prerequisites**:
```bash
# Configure cloud cluster access
kubectl cluster-info
```

**Tests**:
- [ ] Install with cloud-optimized values
- [ ] S3/Cloud storage configuration works
- [ ] LoadBalancer gets external IP
- [ ] Multi-zone resilience (if available)
- [ ] Cloud-native monitoring integration

**Commands**:
```bash
helm install c8s ./chart/c8s \
  -f values-prod.yaml \
  --set storage.s3.endpoint=s3.amazonaws.com \
  --set storage.s3.bucket=... \
  -n c8s-system --create-namespace

kubectl get svc -n c8s-system
# Wait for external IP to be assigned
```

**Test Results**:
- [ ] Status: PASS/FAIL
- [ ] External IP assigned: _____ seconds
- [ ] Storage connectivity: PASS/FAIL
- [ ] Issues encountered: _____

---

### 7. Health Verification Tests

**Purpose**: Verify post-install health checks work correctly

**Tests**:
- [ ] Health hook runs automatically after install
- [ ] All components reported as Ready
- [ ] Dashboard URL displayed
- [ ] Success message with exit code 0
- [ ] Component replica counts reported
- [ ] Readiness status verified

**Command**:
```bash
bash tests/e2e/health_test.sh c8s-health-test
```

**Expected Result**: All health checks pass

---

### 8. Lifecycle Management Tests

**Purpose**: Verify upgrade, downgrade, and uninstall work correctly

**Tests**:
- [ ] Upgrade with new values completes successfully
- [ ] Custom values preserved after upgrade
- [ ] Rolling update zero-downtime
- [ ] Rollback returns to previous version
- [ ] Rollback preserves data (PVCs)
- [ ] Clean uninstall removes C8S resources
- [ ] PVCs preserved after uninstall
- [ ] Release history available for rollback

**Command**:
```bash
bash tests/e2e/lifecycle_test.sh c8s-lifecycle-test
```

**Expected Result**: All lifecycle tests pass

---

### 9. Integration Tests (All User Stories)

**Purpose**: Verify all 4 user stories work together

**User Stories Tested**:
- [ ] US1: Deploy C8S with single command
- [ ] US2: Customize deployment configuration
- [ ] US3: Verify deployment health
- [ ] US4: Manage stack lifecycle

**Command**:
```bash
bash tests/e2e/integration_test.sh c8s-integration
```

**Expected Result**: All integration tests pass

---

### 10. Configuration Tests

**Purpose**: Verify configuration customization works

**Tests**:
- [ ] Log level override (debug, info, warn, error)
- [ ] Component replica override
- [ ] Resource request/limit override
- [ ] Image registry override
- [ ] Image tag override
- [ ] Storage type override
- [ ] S3 credential override

**Commands**:
```bash
helm install c8s ./chart/c8s \
  --set environment.logLevel=debug \
  --set components.controller.replicas=3 \
  --set storage.type=s3-compatible \
  --set storage.s3.endpoint=s3.amazonaws.com

# Verify values applied
kubectl get deployment -n c8s-system -o yaml
```

**Test Results**:
- [ ] All overrides applied correctly
- [ ] Environment variables set
- [ ] Replicas match configuration

---

### 11. Security Tests

**Purpose**: Verify security configuration

**Tests**:
- [ ] RBAC created with minimal permissions
- [ ] Service accounts created correctly
- [ ] ClusterRole has required permissions only
- [ ] No cluster-admin role used
- [ ] Secrets not logged in output
- [ ] Pod security context applied

**Commands**:
```bash
kubectl get clusterrole c8s-controller -o yaml
kubectl get clusterrolebinding c8s-controller -o yaml
kubectl get serviceaccount -n c8s-system
```

**Test Results**:
- [ ] RBAC properly configured
- [ ] No overly permissive rules
- [ ] Security baseline met

---

### 12. Performance Tests

**Purpose**: Verify deployment performance characteristics

**Metrics to Capture**:
- [ ] Time to full deployment: _____ seconds
- [ ] Time per component: api-server _____, controller _____, webhook _____, frontend _____
- [ ] Health check execution time: _____ seconds
- [ ] Memory usage per component: _____
- [ ] CPU usage during deployment: _____

**Commands**:
```bash
# Record start time
start=$(date +%s)

# Deploy
helm install c8s ./chart/c8s -f values-dev.yaml

# Wait for deployment
kubectl wait --for=condition=Available --timeout=300s deployment/c8s-api-server -n c8s-system

# Record end time
end=$(date +%s)
echo "Total time: $((end - start)) seconds"

# Check resource usage
kubectl top pods -n c8s-system
```

**Test Results**:
- [ ] Deployment time within acceptable range (<5 minutes)
- [ ] Resource usage within limits
- [ ] No performance regressions

---

## Test Execution Checklist

### Pre-Test Setup
- [ ] Kubernetes cluster available and accessible
- [ ] kubectl configured correctly
- [ ] Helm 3.x installed
- [ ] Chart files available in `./chart/c8s`
- [ ] Test scripts available in `./tests/e2e/`
- [ ] Required tools installed (docker, docker-compose, etc.)

### During Testing
- [ ] Document all test results
- [ ] Capture timestamps
- [ ] Note any failures and their causes
- [ ] Take screenshots/logs of issues
- [ ] Verify cleanup after each test

### Post-Test
- [ ] Compile all results
- [ ] Analyze failures
- [ ] Document recommendations
- [ ] Archive logs and evidence
- [ ] Create issues for any failures

---

## Test Summary Template

```
╔════════════════════════════════════════════════════════╗
║          C8S Helm Chart Test Summary                  ║
╚════════════════════════════════════════════════════════╝

Kubernetes Distribution: _____________
Kubernetes Version: _____________
Helm Version: _____________
Chart Version: 0.1.0
Test Date: _____________
Tester: _____________

Test Results:
  Lint & Validation:       PASS / FAIL
  Template Rendering:      PASS / FAIL
  Functional (Dev):        PASS / FAIL
  Functional (k3s):        PASS / FAIL
  Functional (kind):       PASS / FAIL
  Functional (Cloud):      PASS / FAIL
  Health Verification:     PASS / FAIL
  Lifecycle Management:    PASS / FAIL
  Integration Tests:       PASS / FAIL
  Configuration Tests:     PASS / FAIL
  Security Tests:          PASS / FAIL
  Performance Tests:       PASS / FAIL

Overall Result: ✅ PASS / ❌ FAIL

Issues Found:
  [Document any failures]

Recommendations:
  [Document any improvements needed]

Sign-Off:
  Tester: ________________  Date: ________
  Reviewer: ________________  Date: ________
```

---

## Continuous Testing

### GitHub Actions Integration

The chart includes CI/CD workflows in `.github/workflows/` that automatically:
- [ ] Run helm lint on every push
- [ ] Template validation on every PR
- [ ] Automated testing on pull requests
- [ ] Test report generation

### Local Testing

Run all tests locally:
```bash
# Quick validation
helm lint ./chart/c8s
helm template c8s ./chart/c8s -f values-dev.yaml

# Full E2E tests
bash tests/e2e/health_test.sh
bash tests/e2e/lifecycle_test.sh
bash tests/e2e/integration_test.sh
```

---

## Known Issues & Limitations

- [ ] Readiness probes require `/readyz` endpoint (currently disabled in dev)
- [ ] Liveness probes require `/livez` endpoint (currently disabled in dev)
- [ ] Storage class selection requires pre-existing storage class
- [ ] S3 credentials must be managed separately (not in cluster secrets by default)
- [ ] Cloud load balancer IP assignment may take 1-5 minutes

---

## References

- [Helm Testing Documentation](https://helm.sh/docs/helm/helm_test/)
- [Kubernetes Testing Best Practices](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [Chart Repository Standards](https://github.com/helm/charts#helm-charts)
