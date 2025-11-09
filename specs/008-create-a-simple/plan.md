# Implementation Plan: Deploy C8S Stack to Kubernetes with Helm

**Branch**: `008-create-a-simple` | **Date**: 2025-11-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-create-a-simple/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a production-ready Helm 3.x chart that deploys the entire C8S stack (API server, controller, webhook, frontend, and dependencies) to Kubernetes in under 5 minutes. The chart must be reusable across direct user deployments and Tilt development workflows, with environment-specific customization via values files (dev/staging/production). Enable users to deploy, verify health, upgrade, and uninstall C8S cleanly using standard `helm install/upgrade/uninstall` commands with clear feedback and error handling.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Helm 3.x (chart templating), Kubernetes 1.24+, YAML manifests
**Primary Dependencies**: Helm 3.x CLI, kubectl CLI, existing C8S components (API server, controller, webhook, frontend)
**Storage**: Kubernetes persistent volumes, S3-compatible object storage (MinIO or AWS S3)
**Testing**: Helm chart lint, Kubernetes integration tests (k3d/kind), E2E deployment tests, Tilt integration validation
**Target Platform**: Kubernetes 1.24+ (k3s, kind, EKS, GKE, AKS), compatible with Tilt development framework
**Project Type**: Helm chart + Kubernetes manifests (single chart source of truth)
**Performance Goals**: Deploy complete stack in under 5 minutes via `helm install`, health check <1 second, idempotent via `helm upgrade`
**Constraints**: Vendor-agnostic (all distributions), reusable in Tilt and direct deployments, single Helm chart (no YAML duplication), values-driven customization
**Scale/Scope**: Single namespace deployments, 3 environment presets (dev/staging/prod), 6+ components monitored

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Specification-First Development ✅
- **Status**: PASS - Complete specification with user scenarios, functional requirements, success criteria, and assumptions defined in spec.md
- **Clarifications**: Helm adoption clarified with explicit decision documented

### User Story-Driven Architecture ✅
- **Status**: PASS - 4 prioritized, independently testable user stories:
  - P1: Deploy C8S with single command (MVP core - `helm install`)
  - P2: Customize Deployment Configuration (values files per environment)
  - P2: Verify Deployment Health (post-install hooks and status)
  - P3: Manage Stack Lifecycle (helm upgrade/uninstall)
- Each story delivers standalone value and can be tested/deployed independently

### Constitution Gates ✅
- **Status**: PASS - Architectural simplicity maintained:
  - Single Helm chart (no unnecessary abstraction)
  - Leverages existing C8S components (no duplication)
  - Standard Helm patterns (values-driven, environment presets)
  - Reusable across direct deployments and Tilt
  - Values files in Git for GitOps compatibility

### Test Independence ✅
- **Status**: PASS - Testing strategy supports independent validation:
  - Helm chart lint tests (structure and syntax)
  - Deployment smoke tests per user story
  - Health check tests (independent of deployment flow)
  - E2E tests for upgrade/downgrade/uninstall
  - Tilt integration tests

### Documentation as Artifact ✅
- **Status**: PASS - Implementation plan will generate:
  - research.md (Helm best practices and decisions)
  - data-model.md (Helm values structure and entities)
  - contracts/ (Helm install/upgrade CLI API, values schema)
  - quickstart.md (5-minute Helm deployment guide)
  - Chart README (Helm-specific documentation)

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
chart/c8s/                          # Standard Helm 3.x chart
├── Chart.yaml                      # Chart metadata, version, dependencies
├── values.yaml                     # Default values for all parameters
├── values-dev.yaml                 # Development environment overrides
├── values-staging.yaml             # Staging environment overrides
├── values-prod.yaml                # Production environment overrides
├── templates/                      # Kubernetes manifest templates
│   ├── namespace.yaml              # Namespace creation
│   ├── _helpers.tpl                # Template helper functions
│   ├── NOTES.txt                   # Post-install deployment notes
│   │
│   ├── api-server/
│   │   ├── deployment.yaml         # API server deployment
│   │   ├── service.yaml            # API server service
│   │   ├── configmap.yaml          # API server configuration
│   │   └── hpa.yaml                # Horizontal Pod Autoscaler [optional]
│   │
│   ├── controller/
│   │   ├── deployment.yaml         # Controller deployment
│   │   ├── service.yaml            # Controller service
│   │   ├── serviceaccount.yaml     # Service account
│   │   ├── rbac.yaml               # Cluster role and binding
│   │   └── configmap.yaml          # Controller configuration
│   │
│   ├── webhook/
│   │   ├── deployment.yaml         # Webhook deployment
│   │   ├── service.yaml            # Webhook service
│   │   ├── validating-webhook.yaml # Validating webhook config
│   │   └── configmap.yaml          # Webhook configuration
│   │
│   ├── frontend/
│   │   ├── deployment.yaml         # Frontend deployment
│   │   ├── service.yaml            # Frontend service
│   │   └── configmap.yaml          # Frontend configuration (if needed)
│   │
│   ├── storage/
│   │   ├── pvc.yaml                # PersistentVolumeClaim [conditional]
│   │   └── s3-secret.yaml          # S3 credentials secret [conditional]
│   │
│   ├── common/
│   │   ├── configmap.yaml          # Shared configuration
│   │   └── secret.yaml             # Shared secrets template
│   │
│   └── _post-install.sh            # Post-install health verification hook
│
├── tests/                          # Helm chart tests
│   ├── unit/                       # Unit tests for templates
│   │   ├── api-server_test.yaml
│   │   ├── controller_test.yaml
│   │   └── webhook_test.yaml
│   └── integration/                # Integration tests
│       └── deployment_test.sh      # K8s integration test script
│
└── README.md                       # Chart documentation

Tiltfile                           # Tilt configuration for local dev
tilt/
├── c8s-values.yaml                # Tilt-specific values overrides
└── c8s-local.yaml                 # Tilt local development config

tests/
├── helm/                          # Helm-specific tests
│   ├── lint/                      # Chart lint validation
│   │   └── lint_test.sh
│   └── template/                  # Template rendering tests
│       └── template_test.sh
└── e2e/                           # End-to-end deployment tests
    ├── deploy_test.sh             # Test direct Helm deployment
    └── tilt_integration_test.sh    # Test Tilt integration
```

**Structure Decision**: Standard Helm chart structure with `/chart/c8s/` as single source of truth for all Kubernetes manifests. Environment-specific customization via separate values files (values-dev.yaml, values-staging.yaml, values-prod.yaml). Chart templates organized by component with shared helpers. Post-install health verification via Helm hook. Chart is directly reusable in Tiltfile and by users via `helm install c8s ./chart/c8s`. This eliminates manifest duplication and ensures consistency across all deployment modes.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

**Status**: No violations identified. Constitution Check passed all gates. No complexity tracking entries required.

---

## Design Artifacts Generated

This plan generates the following design artifacts in Phase 1:

### research.md
- Helm best practices and patterns
- Helm vs. alternatives analysis (updated from previous clarification)
- Helm-specific decisions (values-driven design, hooks, environment presets)
- Tilt integration patterns

### data-model.md
- Helm values structure (root-level and component-level)
- Environment preset definitions (dev/staging/prod defaults)
- Post-install hook specification
- Validation rules for values

### contracts/
- **helm-commands.md**: `helm install`, `helm upgrade`, `helm uninstall`, `helm values` CLI API
- **values-schema.json**: JSON Schema for values.yaml with validation
- **hooks-specification.md**: Post-install hook contract and output format

### quickstart.md
- 5-minute Helm deployment guide (`helm install c8s ./chart/c8s`)
- Environment-specific deployment examples
- Tilt integration quick start
- Troubleshooting for Helm-specific issues

### Chart README
- Chart description and usage
- Prerequisites (Helm 3.x, Kubernetes 1.24+)
- Installation instructions
- Values documentation
- Examples for different scenarios
