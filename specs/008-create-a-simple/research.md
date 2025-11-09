# Research: C8S Kubernetes Deployment Strategy

**Date**: 2025-11-09
**Feature**: Deploy C8S Stack to Kubernetes (008-create-a-simple)
**Status**: Research Complete - All NEEDS CLARIFICATION Resolved

---

## Research Findings Summary

### Decision 1: Helm as Primary Deployment Mechanism

**Decision**: Use Helm 3.x as the primary deployment mechanism with a single Helm chart reusable across direct deployments and Tilt workflows

**Rationale**:
- Provides better user experience for deployment (`helm install` is simple and familiar)
- Single Helm chart can be reused in both direct user deployments AND in Tilt development workflows
- Eliminates duplication between plain YAML and Helm definitions
- Helm handles templating, values overrides, and environment presets (dev/staging/prod)
- Built-in upgrade/rollback capabilities via `helm upgrade` and `helm rollback`
- Helm hooks can be used for health verification post-install
- Packaging and distribution benefits for production users

**Alternatives Considered**:
1. **Plain YAML + kubectl**: Simpler initial implementation but doesn't address Tilt reusability or distribution
2. **Plain YAML + optional Helm**: Maintains simplicity but creates duplication and maintenance burden
3. **CLI wrapper around Helm**: Adds abstraction layer but unnecessary since Helm CLI is already user-friendly

**Implementation Approach**:
- Create `/chart/c8s/` directory with standard Helm chart structure
- `values.yaml` with defaults for all deployment parameters
- `values-dev.yaml`, `values-staging.yaml`, `values-prod.yaml` for environment presets
- Helm templates for all Kubernetes manifests (Deployment, Service, ConfigMap, etc.)
- Helm hooks for post-install health verification
- Chart is integrated into Tilt's Tiltfile for development workflows

**Added Dependency**: Helm 3.x must be installed (added to prerequisites)

---

### Decision 2: Cross-Distribution Compatibility Strategy

**Decision**: Stick to standard Kubernetes 1.24+ APIs with no distribution-specific features

**Rationale**:
- Modern Kubernetes is highly standardized across distributions
- C8S already uses CRDs (CustomResourceDefinitions) which are universally supported
- Core APIs (Deployments, Services, ConfigMaps, Secrets, RBAC) work identically across k3s, kind, EKS, GKE, AKS
- Existing C8S manifests already follow this pattern

**Compatibility Requirements**:
- Kubernetes 1.24+ (already specified in spec.md)
- Core APIs: Deployments, Services, ConfigMaps, Secrets, RBAC ✅
- CRDs: Stable and universally supported ✅
- Readiness/Liveness Probes: Stable across all distributions ✅

**Distribution-Specific Concerns Avoided**:
- NetworkPolicy: Not required (relies on pod-to-pod networking which is universal)
- PodSecurityPolicy: Deprecated in 1.21, removed in 1.25 (use Pod Security Admission if needed)
- Storage Classes: Documented as part of deployment customization (S3-compatible storage)
- Load Balancers: Service type LoadBalancer for cloud providers, can use NodePort for local

**Testing Strategy**: Validate on at least 4 distributions during implementation (k3s, kind, EKS, GKE)

---

### Decision 3: Health Check Implementation Pattern

**Decision**: Use built-in Kubernetes mechanisms (kubectl wait/rollout) + simple verification script

**Rationale**:
- C8S already has proper liveness and readiness probes defined
- `kubectl wait` is lightweight, available in all distributions, requires no external tools
- Standard Kubernetes pattern familiar to all operators
- Enables quick health verification without monitoring systems

**Implementation Pattern**:

```bash
# Wait for deployment readiness (uses built-in readiness probes)
kubectl wait --for=condition=available --timeout=300s deployment/c8s-controller -n c8s-system

# Alternative: Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=c8s-controller -n c8s-system --timeout=300s

# Verify complete stack
kubectl rollout status deployment/c8s-controller -n c8s-system --timeout=300s
```

**Health Check Components to Verify**:
1. API Server deployment available
2. Controller deployment available
3. Webhook deployment available
4. All pods in those deployments ready
5. Dashboard service accessible (port-forward if needed)

**Current State**: Existing manifests have excellent probe configuration - no changes needed to probes themselves

---

### Decision 4: Idempotency Strategy

**Decision**: Leverage Kubernetes' declarative model via `kubectl apply` (already idempotent)

**Rationale**:
- `kubectl apply` is inherently idempotent - uses three-way merge patch
- Compares last-applied, current, and desired state
- Safe to run multiple times without errors
- Won't fail if resources already exist (unlike `kubectl create`)
- Handles updates, creates, and no-op operations consistently

**Idempotent Patterns to Follow**:
- Always use `kubectl apply -f manifest.yaml` (not `create`)
- Define complete desired state in YAML
- Store manifests in Git (GitOps)
- Use descriptive labels and annotations
- Avoid imperative commands (`kubectl edit`, `patch`) in production

**ConfigMap/Secret Updates**: When updating these resources, existing pods won't restart automatically
- Option 1: Manually restart deployment: `kubectl rollout restart deployment/c8s-api-server`
- Option 2: Add annotation hash to deployment spec to auto-trigger restarts
- Document this behavior in quickstart

**Testing for Idempotency**: Run deployment twice and verify:
- Second run produces no changes
- No errors or warnings
- All components remain in ready state

---

### Decision 5: CLI Tool Implementation Approach

**Decision**: Single Go CLI tool that orchestrates `kubectl apply` and verification

**Rationale**:
- Aligns with C8S tech stack (Go 1.25.0)
- Leverages existing Kubernetes client libraries
- Provides user-friendly interface without requiring kubectl expertise
- Can encapsulate deployment logic, validation, and health checks
- Extensible for future lifecycle management (upgrade, uninstall)

**Core Responsibilities**:
1. Validate Kubernetes cluster prerequisites (version, RBAC, API availability)
2. Load and validate deployment configuration
3. Apply Kubernetes manifests in correct order
4. Verify component health and readiness
5. Provide clear feedback and troubleshooting guidance
6. Support environment-specific customization

**CLI Command Structure**:
- `c8s-deploy deploy` - Deploy C8S to cluster
- `c8s-deploy health` - Verify deployment health
- `c8s-deploy config` - Manage deployment configuration
- `c8s-deploy upgrade` - Upgrade C8S version
- `c8s-deploy uninstall` - Remove C8S from cluster

---

### Decision 6: Configuration Customization Mechanism

**Decision**: YAML configuration file + environment presets (dev/staging/prod)

**Rationale**:
- Minimal learning curve (YAML is already required for Kubernetes)
- Supports version control and GitOps workflows
- No external tools or dependencies
- Enables easy environment switching
- Future-ready for Kustomize if needed

**Configuration File Structure**:
```yaml
# c8s-config.yaml
deployment:
  namespace: c8s-system
  image:
    registry: docker.io
    tag: latest
  replicas:
    controller: 1
    webhook: 1
    api-server: 1
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
storage:
  type: s3-compatible  # or pvc, local
  endpoint: minio:9000
  bucketName: c8s-logs
```

**Environment Presets**:
- **dev**: Single replicas, minimal resources, local storage/MinIO
- **staging**: Single replicas, moderate resources, persistent storage
- **production**: Multiple replicas (HA), higher resources, external S3/persistent storage

---

## Resolved NEEDS CLARIFICATION Items

### Item 1: Deployment Tool Choice
**Question**: Should the solution use plain YAML + kubectl, Helm, Kustomize, or another approach?

**Resolution**: Helm 3.x as primary mechanism
- Single Helm chart reusable in both direct deployments and Tilt workflows
- Provides better user experience with familiar `helm install` workflow
- Eliminates duplication between plain YAML and Helm
- Built-in support for upgrades, templating, and environment presets
- Helm hooks enable post-install health verification
- Distribution and packaging benefits for production use

---

### Item 2: Cross-Distribution Support Mechanism
**Question**: How to ensure deployment works identically on k3s, kind, EKS, GKE, AKS without distribution-specific knowledge?

**Resolution**: Standard Kubernetes 1.24+ APIs + comprehensive testing
- Stick to core APIs (Deployments, Services, ConfigMaps, Secrets, RBAC)
- Avoid distribution-specific features (NetworkPolicy, PodSecurityPolicy)
- Validate on at least 4 distributions during implementation
- Document any distribution-specific workarounds clearly

---

### Item 3: Health Verification Pattern
**Question**: What is the best practice for verifying that all C8S components are healthy and ready?

**Resolution**: kubectl wait + custom verification script
- Use `kubectl wait --for=condition=available deployment/...` for each component
- Create simple bash/Go script that checks all components in order
- Leverages built-in readiness probes (already configured in manifests)
- No external monitoring tools required
- Provides clear pass/fail status and troubleshooting guidance

---

### Item 4: Configuration Customization Approach
**Question**: How should users customize deployments for different environments without editing YAML?

**Resolution**: YAML configuration file + command-line flags
- Configuration file approach (c8s-config.yaml) for version control
- Environment presets (dev/staging/prod) for quick setup
- CLI tool handles variable substitution and manifest generation
- Clear documentation on customizable parameters

---

## Recommendations for Implementation

### Phase 1 (MVP): Core Deployment
1. ✅ Create Go CLI tool structure (cmd/c8s-deploy/)
2. ✅ Implement deploy command with validation
3. ✅ Consolidate Kubernetes manifests into single install.yaml
4. ✅ Implement health check command using kubectl wait
5. ✅ Add clear error messages and troubleshooting guidance
6. ✅ Support basic configuration (namespace, image tags)

### Phase 2 (P2): Enhanced Features
1. Implement upgrade command with rollback support
2. Implement uninstall command with cleanup
3. Add environment preset support (dev/staging/prod)
4. Add configuration file support
5. Enhanced health checks with component-level details
6. Add progress indicators and verbose logging

### Phase 3 (P3): Production Ready
1. Add External Secrets Operator documentation (for secrets management)
2. Add GitOps documentation (Argo CD/Flux)
3. Add monitoring and observability guidance
4. Multi-cluster deployment support
5. HA configuration templates

---

## Assumptions Validated

- ✅ Kubernetes 1.24+ is available (standard across all distributions)
- ✅ kubectl is installed and configured on operator machines
- ✅ Container registry access is available (Docker Hub, ECR, GCR, etc.)
- ✅ S3-compatible object storage is available (or will be MinIO for local dev)
- ✅ Users have RBAC permissions to create namespace and resources
- ✅ Internet connectivity for pulling images (standard assumption)

---

## Next Steps

✅ **Research Complete** - All decisions documented with rationale and alternatives

**Proceeding to Phase 1**: Design artifacts
- data-model.md: Deployment configuration structure
- contracts/: CLI API and health check formats
- quickstart.md: Fast-track deployment guide
