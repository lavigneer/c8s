# Tasks: Local Kubernetes Development Tooling with Tilt

**Input**: Design documents from `/specs/003-implement-tilt-or/`
**Prerequisites**: plan.md, spec.md
**Status**: Ready for implementation

**Organization**: Tasks organized by user story priority (P1, P2, P3) to enable independent implementation and incremental delivery.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story ID (US1, US2, US3, US4, US5)
- File paths are absolute from repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for Tilt integration

- [ ] T001 Create Tiltfile at repository root (`/Users/elavigne/workspace/c8s/Tiltfile`)
- [ ] T002 [P] Create `docs/tilt-setup.md` with installation and usage guide
- [ ] T003 [P] Create `specs/003-implement-tilt-or/data-model.md` documenting Tilt state and configuration
- [ ] T004 [P] Create `specs/003-implement-tilt-or/contracts/tiltfile-spec.md` defining Tiltfile configuration contract
- [ ] T005 Create `.gitignore` entries for Tilt local files (`.tilt/`, `tilt_modules/`, local overrides)

**Checkpoint**: Tilt structure and documentation foundation in place

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core Tiltfile configuration that MUST be complete before any user story iteration

**⚠️ CRITICAL**: No effective local development can happen until this phase is complete

- [ ] T006 Implement base Tiltfile configuration with load/import statements and version checks
- [ ] T007 [P] Create Docker build configuration for `controller` image in Tiltfile (references existing `Dockerfile` and `cmd/controller/`)
- [ ] T008 [P] Create Docker build configuration for `api-server` image in Tiltfile (references existing `Dockerfile` and `cmd/api-server/`)
- [ ] T009 [P] Create Docker build configuration for `webhook` image in Tiltfile (references existing `Dockerfile` and `cmd/webhook/`)
- [ ] T010 Configure k3d cluster creation in Tiltfile (or document manual k3d setup)
- [ ] T011 [P] Configure Kubernetes resource deployment in Tiltfile (CRDs, RBAC, component manifests from `deploy/`)
- [ ] T012 Implement environment validation in Tiltfile (check Docker, k3d, kubectl availability)
- [ ] T013 Configure resource limits and constraints for local development (CPU, memory per component)
- [ ] T014 Implement error handling and helpful messages for setup failures

**Checkpoint**: Foundation ready - all three C8S components buildable and deployable via Tilt

---

## Phase 3: User Story 1 - Developer Starts Working on C8S Components (Priority: P1) 🎯 MVP

**Goal**: Developer can run a single command to set up local development environment with auto-rebuild on code changes

**Independent Test**:
1. Run `tilt up` with no prior cluster
2. Verify cluster is created and all components deploy within 5 minutes
3. Modify a Go source file in `cmd/controller/` and verify auto-rebuild and redeploy within 30 seconds
4. View logs in Tilt UI showing rebuild activity

### Implementation for User Story 1

- [ ] T015 [US1] Implement file watching configuration in Tiltfile for Go source changes (`cmd/controller/`, `cmd/api-server/`, `cmd/webhook/`, `pkg/`)
- [ ] T016 [P] [US1] Configure hot-reload/restart for `controller` component on code changes
- [ ] T017 [P] [US1] Configure hot-reload/restart for `api-server` component on code changes
- [ ] T018 [P] [US1] Configure hot-reload/restart for `webhook` component on code changes
- [ ] T019 [US1] Implement build caching optimization in Tiltfile to minimize rebuild time
- [ ] T020 [US1] Configure Tilt to watch manifest files (`deploy/crds.yaml`, `deploy/rbac.yaml`, `deploy/install.yaml`) and re-apply on changes
- [ ] T021 [US1] Implement startup and initialization order (CRDs → RBAC → Components)
- [ ] T022 [US1] Test full setup workflow: `tilt up` → modifications → rebuild → redeploy (manual testing)
- [ ] T023 [US1] Document single-command setup in `docs/tilt-setup.md` (getting started section)

**Checkpoint**: User Story 1 complete - single `tilt up` command provides fully functional development environment with hot reload for all components

---

## Phase 4: User Story 2 - Developer Tests Pipeline Execution Locally (Priority: P1)

**Goal**: Developer can write pipeline definitions locally, validate them, apply to cluster, and get immediate feedback on execution

**Independent Test**:
1. Create/modify a simple pipeline YAML file
2. Use Tilt interface to validate the pipeline definition
3. Apply pipeline to local cluster via Tilt
4. Verify pipeline executes and produces visible logs/status in Tilt UI
5. Modify pipeline and repeat validation immediately

### Implementation for User Story 2

- [ ] T024 [US2] Create pipeline validation helper in Tiltfile (YAML syntax validation via `yq` or similar)
- [ ] T025 [US2] Implement CRD schema validation for PipelineConfig in Tiltfile (validate against `pkg/apis/v1alpha1/pipelineconfig_types.go`)
- [ ] T026 [US2] Configure file watching for pipeline definition files (default: `*.c8s.yaml` in repo root and examples)
- [ ] T027 [US2] Create Tilt resource or button for validating and applying pipeline definitions
- [ ] T028 [US2] Implement detailed validation error reporting with field paths and constraints
- [ ] T029 [US2] Add sample pipeline definitions to `config/samples/` directory (if not exists)
- [ ] T030 [US2] Configure Tilt to stream PipelineRun logs to Tilt UI for visibility
- [ ] T031 [US2] Document pipeline testing workflow in `docs/tilt-setup.md` (pipeline development section)
- [ ] T032 [US2] Test validation catches common errors (missing required fields, type mismatches, invalid references) - manual verification

**Checkpoint**: User Story 2 complete - developers can test pipeline definitions with validation feedback within seconds

---

## Phase 5: User Story 3 - Developer Manages Sample Deployments (Priority: P2)

**Goal**: Developer can easily deploy and manage sample pipelines to test multi-component integration

**Independent Test**:
1. Deploy sample pipelines from `config/samples/` using Tilt
2. Verify all samples are created in cluster
3. Use Tilt to switch between different sample sets (e.g., minimal vs. full examples)
4. Clean up samples with single command

### Implementation for User Story 3

- [ ] T033 [US3] Create Tilt local configuration file support for sample selection (e.g., `tilt_config_samples.json`)
- [ ] T034 [US3] Implement Tilt resource for deploying sample pipelines from `config/samples/`
- [ ] T035 [US3] Add Tilt button/resource to list available sample sets
- [ ] T036 [US3] Implement cleanup function to remove deployed samples (uses `kubectl delete`)
- [ ] T037 [US3] Add ability to switch between sample scenarios without manual cleanup
- [ ] T038 [US3] Create sample pipeline files if missing: `config/samples/simple-pipeline.yaml`, `config/samples/matrix-pipeline.yaml`
- [ ] T039 [US3] Document sample management in `docs/tilt-setup.md` (samples section)
- [ ] T040 [US3] Test sample deployment and cleanup workflows - manual verification

**Checkpoint**: User Story 3 complete - sample management integrated into Tilt workflow

---

## Phase 6: User Story 4 - Developer Observes Multi-Component Interactions (Priority: P2)

**Goal**: Developer can see unified logs from all C8S components in one place to debug interactions

**Independent Test**:
1. Trigger a pipeline execution via API server
2. View logs from webhook (receiving request) → API server (processing) → controller (creating job) in unified view
3. Filter logs by component name
4. Search logs for specific error messages

### Implementation for User Story 4

- [ ] T041 [US4] Configure Tilt dashboard resource groups by component (controller, api-server, webhook, test-jobs)
- [ ] T042 [P] [US4] Configure log aggregation for each component in Tilt UI:
  - [ ] T042a [US4] Controller pod logs aggregation
  - [ ] T042b [US4] API server pod logs aggregation
  - [ ] T042c [US4] Webhook pod logs aggregation
- [ ] T043 [US4] Add pod log filtering configuration (by component, by time range, by log level)
- [ ] T044 [US4] Create Tilt button to toggle verbose logging mode
- [ ] T045 [US4] Document log viewing and debugging workflow in `docs/tilt-setup.md` (debugging section)
- [ ] T046 [US4] Verify multi-component log visibility in Tilt UI - manual testing

**Checkpoint**: User Story 4 complete - unified logging enables debugging of component interactions

---

## Phase 7: User Story 5 - Developer Manages Lifecycle of Local Cluster (Priority: P3)

**Goal**: Developer can easily manage local cluster creation, status checking, and cleanup

**Independent Test**:
1. Use Tilt to check/create cluster
2. Verify cluster health and component status in Tilt UI
3. Destroy cluster cleanly with command

### Implementation for User Story 5

- [ ] T047 [US5] Implement Tilt button for cluster creation (wraps k3d cluster create)
- [ ] T048 [US5] Implement Tilt button for cluster deletion (wraps k3d cluster delete)
- [ ] T049 [US5] Add cluster status display in Tilt UI (cluster name, k3d version, node count)
- [ ] T050 [US5] Display component deployment status (controller ready, api-server ready, webhook ready)
- [ ] T051 [US5] Create `tilt down` hook documentation (cleanup behavior)
- [ ] T052 [US5] Document cluster management in `docs/tilt-setup.md` (cluster lifecycle section)
- [ ] T053 [US5] Verify cluster lifecycle commands work correctly - manual testing

**Checkpoint**: User Story 5 complete - full cluster lifecycle management integrated

---

## Phase 8: Documentation & Edge Cases

**Purpose**: Handle edge cases and document complete development workflow

- [ ] T054 [P] Handle build failures gracefully: display in Tilt UI with actionable error messages (e.g., "syntax error in cmd/controller/main.go:123")
- [ ] T055 [P] Handle CRD schema changes: detect manifest updates and rebuild components automatically
- [ ] T056 [P] Handle branch switching: document process for updating CRDs when switching branches with incompatible definitions
- [ ] T057 [P] Handle resource constraints: add configuration for low-resource machines (< 8 GB RAM) with reduced replicas
- [ ] T058 [P] Handle dev environment persistence: document that cluster persists when Tilt is killed and reconnection procedure
- [ ] T059 Validate success criteria against Tilt behavior:
  - [ ] T059a SC-001: Setup in < 5 minutes - test on minimum spec machine
  - [ ] T059b SC-002: Change detection < 30 seconds - time rebuild cycles
  - [ ] T059c SC-003: Build failure reporting < 10 seconds - measure error display latency
  - [ ] T059d SC-004: Pipeline test < 2 minutes - test e2e workflow
  - [ ] T059e SC-005: Unified logs in single interface - verify Tilt UI completeness
  - [ ] T059f SC-006: 95% crash-free sessions - run extended stability tests
  - [ ] T059g SC-007: 50% faster onboarding - conduct developer feedback session
  - [ ] T059h SC-008: 4+ hour stability - run extended session test
- [ ] T060 [P] Create troubleshooting section in `docs/tilt-setup.md` for common issues
- [ ] T061 [P] Create video walkthrough script for quick-start guide
- [ ] T062 Update main `README.md` to reference Tilt setup as primary development method
- [ ] T063 Update `CLAUDE.md` project guidelines to mention Tilt as standard dev tool
- [ ] T064 Create PR instructions for new developers using Tilt workflow

**Checkpoint**: All user stories functional with comprehensive documentation

---

## Phase 9: Polish & Integration

**Purpose**: Final refinements and cross-cutting improvements

- [ ] T065 [P] Code review of Tiltfile for best practices and clarity
- [ ] T066 [P] Test Tiltfile on Linux and macOS (if available)
- [ ] T067 [P] Performance tuning: optimize build caching and rebuild speed
- [ ] T068 Verify idempotence: running `tilt up` multiple times produces consistent results
- [ ] T069 Test cleanup: verify `tilt down` removes all C8S resources cleanly
- [ ] T070 Create integration test runner that uses Tilt to validate setup works
- [ ] T071 [P] Update existing developer documentation to reference Tilt
- [ ] T072 Run full workflow end-to-end: setup → modify code → rebuild → test pipeline → cleanup
- [ ] T073 Validate against success criteria and close feature

**Checkpoint**: Feature complete and validated

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Phase 1 (Setup)**: No dependencies - start immediately
2. **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
3. **Phase 3 (US1 - P1)**: Depends on Phase 2 - MVP core functionality
4. **Phase 4 (US2 - P1)**: Depends on Phase 2, can start when Phase 3 begun (independent feature)
5. **Phase 5 (US3 - P2)**: Depends on Phase 2, can start after Phase 4 or in parallel with it
6. **Phase 6 (US4 - P2)**: Depends on Phase 2, can run parallel with US3 or US2
7. **Phase 7 (US5 - P3)**: Depends on Phase 2, can run parallel with other stories or last
8. **Phase 8 (Documentation & Edge Cases)**: Depends on all stories, finalizes implementation
9. **Phase 9 (Polish & Integration)**: Depends on Phase 8, final validation

### Within-Phase Dependencies

**Phase 1**: No dependencies - all can run parallel with [P]

**Phase 2**:
- T006 (base Tiltfile) must run before T007-T009 (component builds)
- T010 (k3d config) can run parallel with T007-T009
- T011 (resource deployment) depends on T006
- T012-T014 can run after T006

**Phase 3 (US1)**:
- T015 (file watching config) must be before T016-T018 (component rebuild)
- T021 (startup order) must be before final testing T022
- Can proceed when Phase 2 complete

**Phase 4 (US2)**:
- T024 (validation helper) before T025 (schema validation)
- T025 before T028 (error reporting)
- T026 (file watching) parallel with T024-T025
- Can start immediately after T024 framework is in place

**Phase 5-7 (US3-5)**:
- No internal dependencies, can start after Phase 2

### Parallel Execution Examples

**Setup (Phase 1)**:
```
T001 (Tiltfile creation)
T002, T003, T004, T005 [P] (documentation & contracts)
```

**Foundational (Phase 2)**:
```
T006 (base Tiltfile)
├─ T007, T008, T009 [P] (Docker configs for all components)
├─ T010 [P] (k3d setup)
├─ T011 (resource deployment)
├─ T012, T013, T014 [P] (validation, constraints, errors)
```

**User Stories in Parallel** (after Phase 2):
```
Developer A: Phase 3 (US1 - hot reload) - Critical for MVP
Developer B: Phase 4 (US2 - validation) - Can start when Phase 3 begun
Developer C: Phase 5 (US3 - samples) - Can start when Phase 2 complete
```

---

## Implementation Strategy

### MVP First (Recommended)

1. **Complete Phase 1**: Setup infrastructure (2-4 hours)
2. **Complete Phase 2**: Foundational Tiltfile configuration (4-6 hours)
3. **Complete Phase 3**: User Story 1 - Hot reload development (4-6 hours)
4. **STOP and VALIDATE**:
   - Test `tilt up` end-to-end
   - Verify hot reload works for all components
   - Get developer feedback
5. **Deploy MVP**: Create PR with Phase 1-3 tasks complete

**MVP Completion Time**: ~1-2 weeks for experienced developer

### Incremental Delivery

After MVP (Phase 1-3):

1. **Phase 4 (US2)**: Add pipeline validation (2-3 days)
2. **Phase 5 (US3)**: Add sample management (1-2 days)
3. **Phase 6 (US4)**: Add unified logging (1-2 days)
4. **Phase 7 (US5)**: Add cluster lifecycle management (1 day)
5. **Phase 8-9**: Documentation and polish (2-3 days)

Each phase can be delivered as a separate PR/feature branch.

### Sequential Team Strategy (Single Developer)

1. Phase 1: Setup (1-2 days)
2. Phase 2: Foundational (2-3 days)
3. Phase 3: US1 (2-3 days)
4. Phase 4: US2 (1-2 days)
5. Phase 5: US3 (1 day)
6. Phase 6: US4 (1 day)
7. Phase 7: US5 (1 day)
8. Phase 8-9: Polish (1-2 days)

**Total Estimated Time**: 2-3 weeks

---

## Success Criteria Validation

Each user story addresses specific success criteria from spec.md:

- **US1 (Hot Reload)** → SC-001, SC-002, SC-003
- **US2 (Pipeline Validation)** → SC-004, SC-007
- **US3 (Sample Management)** → SC-005
- **US4 (Unified Logging)** → SC-005, SC-012
- **US5 (Cluster Lifecycle)** → SC-006, SC-001
- **Phase 8 (Testing)** → All SC criteria validated

---

## Notes

- [P] tasks can run in parallel (different files, no code dependencies)
- [Story] label enables traceability and independent completion tracking
- Each phase should be independently testable and deployable
- Commit after each task or logical group (as requested)
- Update `CLAUDE.md` after feature completion
- Tilt handles file watching natively - no custom implementation needed
- Dashboard/UI provided by Tilt - no custom tooling required
- All paths assume repository root at `/Users/elavigne/workspace/c8s`
