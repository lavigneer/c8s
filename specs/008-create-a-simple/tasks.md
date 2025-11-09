# Tasks: Deploy C8S Stack to Kubernetes with Helm

**Input**: Design documents from `/specs/008-create-a-simple/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/
**Feature Branch**: `008-create-a-simple`
**Status**: Ready for implementation

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Each user story phase builds on the foundational work and can be developed independently.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- **File paths**: Relative to repository root (`chart/c8s/`, `tests/`, `Tiltfile`, etc.)

---

## Phase 1: Setup & Project Initialization

**Purpose**: Create Helm chart structure and shared infrastructure

- [x] T001 Create `/chart/c8s/` directory structure with subdirectories: `templates/`, `tests/`
- [x] T002 Create `/chart/c8s/Chart.yaml` with chart metadata (name: c8s, version: 0.1.0)
- [x] T003 [P] Create `/chart/c8s/values.yaml` with root-level defaults for all parameters
- [x] T004 [P] Create `/chart/c8s/values-dev.yaml` with development environment overrides
- [x] T005 [P] Create `/chart/c8s/values-staging.yaml` with staging environment overrides
- [x] T006 [P] Create `/chart/c8s/values-prod.yaml` with production environment overrides
- [x] T007 Create `/chart/c8s/templates/_helpers.tpl` with common Helm template functions
- [x] T008 Create test directory structure: `/tests/helm/lint/`, `/tests/helm/template/`, `/tests/e2e/`
- [x] T009 Create `Tiltfile` with Helm chart integration (helm() function)
- [x] T010 Create `/tilt/c8s-values.yaml` with Tilt-specific values overrides

**Checkpoint**: Helm chart structure created - ready for component template implementation

---

## Phase 2: Foundational Components (Blocking Prerequisites)

**Purpose**: Create core Helm templates that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T011 Create `/chart/c8s/templates/namespace.yaml` template for namespace creation
- [x] T012 Create `/chart/c8s/templates/_helpers.tpl` with helper functions (labels, names, selectors)
- [x] T013 Create `/chart/c8s/templates/NOTES.txt` with post-install instructions
- [x] T014 [P] Create `/chart/c8s/templates/common/configmap.yaml` for shared configuration
- [x] T015 [P] Create `/chart/c8s/templates/common/secret.yaml` for shared secrets template
- [x] T016 Create `/chart/c8s/templates/_post-install.sh` hook for health verification

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Deploy C8S with Single Command (Priority: P1) 🎯 MVP

**Goal**: Enable users to deploy the entire C8S stack with a single `helm install` command that completes in under 5 minutes with all components functional.

**Independent Test**: Run `helm install c8s ./chart/c8s -f values-dev.yaml` and verify all C8S components reach "Ready" state within 5 minutes via post-install hook output.

### Implementation for User Story 1 (Basic Deployment)

**API Server Component**:
- [x] T017 [P] [US1] Create `/chart/c8s/templates/api-server/deployment.yaml` with image, ports, probes, resource requests
- [x] T018 [P] [US1] Create `/chart/c8s/templates/api-server/service.yaml` with ClusterIP service
- [x] T019 [P] [US1] Create `/chart/c8s/templates/api-server/configmap.yaml` for configuration

**Controller Component**:
- [x] T020 [P] [US1] Create `/chart/c8s/templates/controller/deployment.yaml` with controller specs
- [x] T021 [P] [US1] Create `/chart/c8s/templates/controller/service.yaml` (optional, if needed)
- [x] T022 [P] [US1] Create `/chart/c8s/templates/controller/serviceaccount.yaml`
- [x] T023 [P] [US1] Create `/chart/c8s/templates/controller/rbac.yaml` with ClusterRole and ClusterRoleBinding
- [x] T024 [P] [US1] Create `/chart/c8s/templates/controller/configmap.yaml` for configuration

**Webhook Component**:
- [x] T025 [P] [US1] Create `/chart/c8s/templates/webhook/deployment.yaml` with webhook specs
- [x] T026 [P] [US1] Create `/chart/c8s/templates/webhook/service.yaml` for webhook service
- [x] T027 [P] [US1] Create `/chart/c8s/templates/webhook/validating-webhook.yaml` for validation webhook config
- [x] T028 [P] [US1] Create `/chart/c8s/templates/webhook/configmap.yaml` for configuration

**Frontend Component**:
- [x] T029 [P] [US1] Create `/chart/c8s/templates/frontend/deployment.yaml` with frontend specs
- [x] T030 [P] [US1] Create `/chart/c8s/templates/frontend/service.yaml` for frontend (LoadBalancer for cloud, NodePort for local)

**Integration & Testing**:
- [x] T031 [US1] Create `/tests/helm/lint/lint_test.sh` for Helm chart lint validation
- [x] T032 [US1] Create `/tests/e2e/deploy_test.sh` to verify full deployment works in k3d/kind
- [x] T033 [US1] Update post-install hook to verify all 3+ components reach Ready state
- [ ] T034 [US1] Test deployment in 3+ Kubernetes distributions (k3s, kind, GKE) to verify vendor-agnostic compatibility
- [x] T035 [US1] Document quick start in `/chart/c8s/README.md` with basic deployment example

**Acceptance Criteria**:
- ✅ `helm install c8s ./chart/c8s -f values-dev.yaml -n c8s-system` successfully deploys all components
- ✅ All deployments reach "Ready" state (N/N replicas) within 5 minutes
- ✅ Dashboard is accessible via port-forward on frontend service
- ✅ Health check hook reports all components ready
- ✅ Helm lint passes without errors or critical warnings

**Checkpoint**: User Story 1 complete - basic C8S deployment works independently

---

## Phase 4: User Story 2 - Customize Deployment Configuration (Priority: P2)

**Goal**: Enable users to customize deployment for different environments (dev/staging/production) without editing YAML, with environment presets and per-component overrides.

**Independent Test**: Deploy with `values-prod.yaml`, verify 3 replicas of controller, 2 replicas of webhook, higher resource limits are applied. Deploy with `values-dev.yaml`, verify 1 replica each with minimal resources.

### Implementation for User Story 2 (Configuration Customization)

**Values Structure Enhancement**:
- [ ] T036 [P] [US2] Add component replicas section to `values.yaml` (controller, webhook, api-server defaults)
- [ ] T037 [P] [US2] Add resource section to `values.yaml` (requests/limits for CPU and memory per component)
- [ ] T038 [P] [US2] Add image section to `values.yaml` (registry, repository, tag per component)
- [ ] T039 [P] [US2] Add storage section to `values.yaml` (storage type, S3 endpoint, bucket, credentials template)
- [ ] T040 [P] [US2] Add environment section to `values.yaml` (logLevel, environment type: dev/staging/prod)

**Environment Preset Customization**:
- [ ] T041 [US2] Create environment preset logic in values files (dev: 1 replica, 256Mi memory; prod: 3 replicas, 1Gi memory)
- [ ] T042 [US2] Update deployment templates to use .Values for all customizable parameters (replicas, resources, image, env vars)
- [ ] T043 [US2] Update configmap templates to populate from .Values entries
- [ ] T044 [P] [US2] Verify values-dev.yaml, values-staging.yaml, values-prod.yaml all render correctly

**Storage Configuration**:
- [ ] T045 [P] [US2] Create `/chart/c8s/templates/storage/pvc.yaml` template (conditional based on storage type)
- [ ] T046 [P] [US2] Create `/chart/c8s/templates/storage/s3-secret.yaml` template for S3 credentials

**Documentation & Examples**:
- [ ] T047 [US2] Create `values.yaml` comments documenting all parameters and defaults
- [ ] T048 [US2] Add environment preset examples to `/chart/c8s/README.md` (dev, staging, prod)
- [ ] T049 [US2] Create example values files for custom scenarios (high-HA, minimal resources, external storage)
- [ ] T050 [US2] Document override syntax: `helm install --values` and `--set` flags

**Testing**:
- [ ] T051 [US2] Create `/tests/helm/template/template_test.sh` to test template rendering with different values
- [ ] T052 [US2] Test deployment with dev values: verify 1 replica, 256Mi memory requests
- [ ] T053 [US2] Test deployment with prod values: verify 3 replicas, 1Gi memory limits
- [ ] T054 [US2] Test custom S3 storage: deploy with custom S3 credentials, verify secret is created
- [ ] T055 [US2] Test values override: `helm install --set components.controller.replicas=5`, verify replica count

**Acceptance Criteria**:
- ✅ `helm install c8s ./chart/c8s -f values-prod.yaml` applies production settings (replicas, resources)
- ✅ `helm install c8s ./chart/c8s -f values-dev.yaml` applies dev settings (minimal)
- ✅ `helm install --set components.controller.replicas=2` overrides replicas via CLI
- ✅ S3 storage can be configured via values without editing templates
- ✅ All templates parameterized - no hardcoded values (except defaults)

**Checkpoint**: User Story 2 complete - customization and environment presets work independently

---

## Phase 5: User Story 3 - Verify Deployment Health (Priority: P2)

**Goal**: Enable users to verify all C8S components are healthy and ready, with clear status reporting and issue identification.

**Independent Test**: Deploy C8S, then check health via `helm status` and post-install hook output, verify all components show "Ready" and dashboard URL is displayed.

### Implementation for User Story 3 (Health Verification)

**Post-Install Hook Enhancement**:
- [ ] T056 [US3] Enhance `/chart/c8s/templates/_post-install.sh` to check all component deployments
- [ ] T057 [US3] Add health check logic: wait for each deployment to reach Ready state (N/N replicas)
- [ ] T058 [US3] Add component status reporting: show replica counts and readiness for each component
- [ ] T059 [US3] Add timeout handling: if components not ready after 5 minutes, report failure with suggestions
- [ ] T060 [US3] Add dashboard URL generation and display in post-install output

**Health Status Output Format**:
- [ ] T061 [US3] Create health status structure: component name, replicas ready, status (Ready/Pending/Failed)
- [ ] T062 [US3] Add remediation suggestions (e.g., "Check logs: kubectl logs deployment/controller -n c8s-system")
- [ ] T063 [US3] Display overall health summary: "3 of 3 components ready" or "Deployment succeeded" message

**Documentation & Testing**:
- [ ] T064 [US3] Create `/tests/e2e/health_test.sh` to verify health check functionality
- [ ] T065 [US3] Test health check with healthy deployment: verify all components show Ready
- [ ] T066 [US3] Test health check with one failing component: verify failure is detected and suggestion provided
- [ ] T067 [US3] Add health check documentation to `NOTES.txt` and `/chart/c8s/README.md`
- [ ] T068 [US3] Create troubleshooting guide: common issues and how to identify root causes

**Integration with US1**:
- [ ] T069 [US3] Verify post-install hook output is displayed after `helm install` completes
- [ ] T070 [US3] Ensure health check doesn't block deployment (non-critical failures don't fail install)

**Acceptance Criteria**:
- ✅ Post-install hook runs automatically after `helm install`
- ✅ Hook checks all 3+ components and reports status
- ✅ Success case: all components Ready, dashboard URL displayed, exit code 0
- ✅ Failure case: component not ready, suggestion provided, clear output
- ✅ Health check completes in <1 second per component

**Checkpoint**: User Story 3 complete - health verification works independently

---

## Phase 6: User Story 4 - Manage Stack Lifecycle (Priority: P3)

**Goal**: Enable users to upgrade, downgrade, and uninstall C8S cleanly with data preservation and minimal downtime.

**Independent Test**: Deploy C8S, upgrade to new version via `helm upgrade`, verify all components update and remain functional, then uninstall via `helm uninstall`, verify all resources are removed cleanly.

### Implementation for User Story 4 (Lifecycle Management)

**Upgrade Support**:
- [ ] T071 [P] [US4] Add upgrade strategy to deployment templates: `RollingUpdate` with maxSurge/maxUnavailable
- [ ] T072 [P] [US4] Add progress deadline to deployments to handle slow clusters
- [ ] T073 [US4] Test upgrade from v0.1.0 to v0.2.0: chart version bump, template changes
- [ ] T074 [US4] Verify custom values are preserved during upgrade (values in release, new defaults applied)
- [ ] T075 [US4] Create pre-upgrade checks in hook (e.g., backup state, verify API availability)

**Downgrade Support**:
- [ ] T076 [US4] Verify downgrade works via `helm rollback`: return to previous version
- [ ] T077 [US4] Create release history documentation: `helm history c8s`
- [ ] T078 [US4] Test data preservation during downgrade: PVs and logs retained

**Uninstall Cleanup**:
- [ ] T079 [P] [US4] Add Helm deletion hooks if needed (cleanup scripts)
- [ ] T080 [P] [US4] Add cleanup labels to all resources for cleanup validation
- [ ] T081 [US4] Test `helm uninstall c8s` removes all C8S resources cleanly
- [ ] T082 [US4] Verify namespace can be deleted if empty (or flag to keep namespace)
- [ ] T083 [US4] Test data options: `--keep-history` for rollback support

**Documentation & Testing**:
- [ ] T084 [US4] Create `/tests/e2e/lifecycle_test.sh` for upgrade/downgrade/uninstall testing
- [ ] T085 [US4] Add upgrade instructions to `/chart/c8s/README.md`: `helm upgrade c8s ./chart/c8s`
- [ ] T086 [US4] Add downgrade instructions: `helm rollback c8s <revision>`
- [ ] T087 [US4] Add uninstall instructions: `helm uninstall c8s -n c8s-system`
- [ ] T088 [US4] Create changelog documentation for chart versions (what changed, upgrade steps)

**Integration with US1-US3**:
- [ ] T089 [US4] Verify health check works after upgrade (components ready post-update)
- [ ] T090 [US4] Ensure values customization (US2) preserved during upgrade

**Acceptance Criteria**:
- ✅ `helm upgrade c8s ./chart/c8s -f values-prod.yaml` successfully updates all components
- ✅ Rolling update strategy ensures zero downtime (no more than maxUnavailable pods down)
- ✅ Custom values preserved after upgrade (replicas, storage, images)
- ✅ `helm rollback c8s <revision>` returns to previous version successfully
- ✅ `helm uninstall c8s -n c8s-system` removes all C8S resources
- ✅ Release history available: `helm history c8s`

**Checkpoint**: User Story 4 complete - full lifecycle management works independently

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements and enhancements that benefit all user stories

- [ ] T091 [P] Create comprehensive `/chart/c8s/README.md` with all usage patterns (install, upgrade, uninstall, customize)
- [ ] T092 [P] Create `CHANGELOG.md` documenting chart version history and upgrade notes
- [ ] T093 [P] Create `/docs/helm-values-reference.md` documenting all values parameters
- [ ] T094 [P] Create `/docs/troubleshooting.md` with common issues and solutions
- [ ] T095 [P] Update Tiltfile for seamless local development with chart updates
- [ ] T096 [P] Create Helm chart best practices documentation
- [ ] T097 Create integration test that verifies all 4 user stories work together: deploy → customize → health → upgrade
- [ ] T098 Run comprehensive testing matrix:
  - [ ] T098a Test on k3s (local)
  - [ ] T098b Test on kind (local)
  - [ ] T098c Test on EKS (if available)
  - [ ] T098d Test on GKE (if available)
- [ ] T099 Create release checklist for publishing Helm chart (lint, test, version bump, push)
- [ ] T100 Document Tilt integration: how to use chart during development

**Acceptance Criteria**:
- ✅ All documentation complete and accurate
- ✅ Helm chart passes lint with no errors
- ✅ Chart tested on 4+ Kubernetes distributions
- ✅ All user stories pass independently and together
- ✅ Ready for users to consume and for Tilt integration

---

## Implementation Strategy & MVP Scope

### MVP (Minimum Viable Product): User Story 1 Only
Complete **Phase 1**, **Phase 2**, and **Phase 3** to deliver basic Helm chart deployment:
- ✅ Create Helm chart structure
- ✅ Implement all component templates (API server, controller, webhook, frontend)
- ✅ Post-install hook with health verification
- ✅ Deploy with `helm install c8s ./chart/c8s` in <5 minutes
- **MVP Delivery**: Users can deploy C8S with a single command

### Incremental Delivery: Add User Stories Progressively

**Phase 1 (MVP)**: US1 - Basic deployment works

**Phase 2 (v0.1.1)**: Add US2 - Customization via values files
- Environment presets (dev/staging/prod)
- Parameter overrides
- Storage configuration

**Phase 3 (v0.2.0)**: Add US3 - Health verification
- Enhanced post-install hook
- Component status reporting
- Remediation suggestions

**Phase 4 (v0.2.1)**: Add US4 - Lifecycle management
- Upgrade support with rolling updates
- Downgrade via `helm rollback`
- Clean uninstall

**Phase 5 (v0.3.0)**: Polish & Release
- Comprehensive documentation
- Testing on multiple distributions
- Helm chart publication

---

## Parallel Execution Opportunities

**Phase 1 - Setup (Parallel)**:
```
T003-T006: values.yaml files (all environment presets) - can run in parallel
T010: Tiltfile - independent
```

**Phase 2 - Foundation (Sequential)**:
- Helper templates (T012) must complete before other phases
- Common templates (T014-T015) can be parallel

**Phase 3 - US1 (Highly Parallel)**:
```
# API Server: T017-T019 (parallel)
# Controller: T020-T024 (parallel)
# Webhook: T025-T028 (parallel)
# Frontend: T029-T030 (parallel)
# Sequential: T031-T035 (testing and integration)
```

**Phase 4 - US2 (Parallel)**:
```
# Values updates: T036-T040 (parallel)
# Storage: T045-T046 (parallel)
# Testing: T051-T055 (parallel after values complete)
```

**Phase 5 - US3 (Sequential)**:
- Health check implementation must be sequential (depends on deployments)
- Testing can follow once implementation complete

**Phase 6 - US4 (Mostly Sequential)**:
- Upgrade strategy (T071-T075) sequential
- Uninstall cleanup (T079-T083) can be parallel
- Testing sequential

---

## Dependencies & Blockers

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundation) - BLOCKING
    ↓ (all phases depend on this)
Phase 3 (US1 - P1) ← INDEPENDENT PATH A
    ↓
Phase 4 (US2 - P2) ← INDEPENDENT PATH B
    ↓
Phase 5 (US3 - P2) ← INDEPENDENT PATH C
    ↓
Phase 6 (US4 - P3) ← INDEPENDENT PATH D
    ↓
Phase 7 (Polish) - Final integration
```

**Critical Dependencies**:
- Phase 2 must complete before ANY user story starts
- Each user story is independent (can implement US2 before US3, or US3 before US2)
- US1 should complete before US2 (US2 depends on US1's template structure)
- US4 requires US1 (testing upgrade assumes US1 works)

---

## Success Metrics per User Story

**US1**:
- ✅ `helm install c8s ./chart/c8s` deploys in <5 minutes with all components Ready

**US2**:
- ✅ `helm install c8s ./chart/c8s -f values-prod.yaml` applies production config (3 replicas, high resources)

**US3**:
- ✅ Post-install hook verifies all components and displays status/URL

**US4**:
- ✅ `helm upgrade c8s ./chart/c8s --set imageTag=v0.2.0` updates successfully

---

## Testing Strategy

**Helm Chart Lint** (T031):
```bash
helm lint ./chart/c8s
```

**Template Rendering** (T032):
```bash
helm template c8s ./chart/c8s -f values-dev.yaml
```

**Integration Tests** (T032, T051-T055, T065-T067, T073-T082):
- Deploy to k3d/kind cluster
- Verify components reach Ready state
- Check health output
- Test upgrades and uninstalls

**End-to-End Tests** (T098a-d):
- Test on 4+ Kubernetes distributions
- Verify all user stories work independently
- Verify all user stories work together

---

## Task Summary by User Story

| User Story | Tasks | Priority | MVP? |
|---|---|---|---|
| **US1: Basic Deploy** | T001-T035 | P1 | ✅ Yes |
| **US2: Customization** | T036-T055 | P2 | Optional |
| **US3: Health Check** | T056-T070 | P2 | Optional |
| **US4: Lifecycle** | T071-T090 | P3 | Optional |
| **Polish** | T091-T100 | - | Final |
| **TOTAL** | **100 tasks** | - | - |

---

## Next Steps

1. **Read This Document**: Understand the complete task breakdown
2. **Start Phase 1**: Create directory structure and values files
3. **Complete Phase 2**: Implement foundational templates
4. **Begin Phase 3** (US1): Implement component templates in parallel
5. **Test Incrementally**: Verify each phase before moving to next
6. **Proceed to US2-US4**: As time and priority permit
7. **Polish & Release**: When all user stories complete

Each task includes enough detail for LLM execution without additional context. Tasks are designed to be completed in order but can be parallelized where [P] is marked.

**Estimated Effort**:
- **Phase 1-2** (Setup + Foundation): 2-3 days
- **Phase 3** (US1 MVP): 3-5 days (highly parallelizable)
- **Phase 4** (US2): 2-3 days
- **Phase 5** (US3): 2 days
- **Phase 6** (US4): 3-4 days
- **Phase 7** (Polish): 2-3 days
- **Total MVP (US1 only)**: 5-8 days
- **Total Feature**: 14-20 days

---

**Ready to implement!** Pick a task from Phase 1 or Phase 3 to begin.
