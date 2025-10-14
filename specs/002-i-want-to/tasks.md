# Tasks: Local Test Environment Setup

**Input**: Design documents from `/specs/002-i-want-to/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/
**Branch**: `002-i-want-to`

**Tests**: Contract tests are included to validate CLI command behavior

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- All paths are relative to repository root

## Path Conventions (from plan.md)
- **CLI commands**: `cmd/c8s/commands/dev/`
- **Core logic**: `pkg/localenv/`
- **Tests**: `tests/contract/`, `tests/integration/`, `tests/unit/`
- **Config**: `config/samples/`
- **Docs**: `docs/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for local environment tooling

- [x] T001 Create directory structure for local environment packages
  - Create `pkg/localenv/`
  - Create `pkg/localenv/cluster/`
  - Create `pkg/localenv/deploy/`
  - Create `pkg/localenv/samples/`
  - Create `pkg/localenv/health/`
  - Create `cmd/c8s/commands/dev/`
  - Create `tests/contract/`
  - Create `tests/integration/localenv/`
  - Create `tests/unit/localenv/`
  - Create `config/samples/`

- [x] T002 [P] Update go.mod with required dependencies
  - Add k3d Go client library (if available) or exec wrapper dependencies
  - Add cobra/pflag for CLI (should already exist)
  - Add validation libraries
  - Run `go mod tidy`

- [x] T003 [P] Create base CLI command structure in `cmd/c8s/commands/dev/dev.go`
  - Define `dev` root command using cobra
  - Add global flags (--verbose, --quiet, --no-color)
  - Setup command group structure (cluster, deploy, test subcommands)
  - Wire into existing c8s CLI main.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create configuration data structures in `pkg/localenv/types.go`
  - Define ClusterConfig struct (name, kubernetesVersion, nodes, ports, registry, options)
  - Define NodeConfig struct (type, count, resources)
  - Define PortMapping struct (hostPort, containerPort, protocol, nodeFilter)
  - Define RegistryConfig struct (enabled, name, hostPort, proxyRemote)
  - Define ClusterOptions struct (waitTimeout, updateDefaultKubeconfig, switchContext, k3sArgs)
  - Add JSON/YAML struct tags for serialization
  - Add validation tags

- [x] T005 [P] Create cluster status structures in `pkg/localenv/status.go`
  - Define ClusterStatus struct (name, state, nodes, kubeconfig, apiEndpoint, registryEndpoint, createdAt, uptime)
  - Define NodeStatus struct (name, role, status, version)
  - Add JSON struct tags

- [x] T006 [P] Implement configuration validation in `pkg/localenv/validation.go`
  - Validate cluster name pattern (lowercase alphanumeric + hyphens)
  - Validate port ranges (1024-65535 for host ports)
  - Validate Kubernetes version format
  - Validate node configuration (at least 1 server)
  - Validate no duplicate port mappings
  - Return clear error messages for each validation failure

- [x] T007 Create k3d wrapper interface in `pkg/localenv/cluster/k3d.go`
  - Define K3dClient interface (Create, Delete, Start, Stop, List, Get, LoadImage methods)
  - Implement k3dClientImpl using k3d command-line execution
  - Add error handling and output parsing
  - Add Docker availability check

- [ ] T008 [P] Implement kubectl wrapper in `pkg/localenv/cluster/kubectl.go`
  - Define KubectlClient interface (ApplyManifest, DeleteResource, GetResource, WaitForReady, GetLogs methods)
  - Implement using kubectl command-line execution
  - Add context switching logic
  - Add timeout handling

- [ ] T009 [P] Create health check utilities in `pkg/localenv/health/checks.go`
  - Check Docker daemon availability (docker info command)
  - Check kubectl installation
  - Check cluster readiness (nodes Ready, API accessible)
  - Check CRD registration
  - Check pod status
  - Return structured health status

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Local Cluster Creation (Priority: P1) 🎯 MVP

**Goal**: Enable developers to create, manage, and delete local Kubernetes clusters using simple CLI commands

**Independent Test**: Run `c8s dev cluster create`, verify cluster is accessible with kubectl, deploy a test pod, then delete the cluster

### Contract Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Contract test for `cluster create` command in `tests/contract/cluster_create_test.go`
  - Test default cluster creation (no arguments)
  - Test custom cluster name creation
  - Test cluster creation with --config flag
  - Test error: cluster already exists (exit code 2)
  - Test error: Docker not running (exit code 4)
  - Verify kubeconfig is updated
  - Verify context is switched
  - Verify output format matches spec

- [ ] T011 [P] [US1] Contract test for `cluster delete` command in `tests/contract/cluster_delete_test.go`
  - Test default cluster deletion
  - Test deletion with custom name
  - Test deletion with --force flag (no confirmation)
  - Test error: cluster not found (exit code 2)
  - Verify confirmation prompt behavior
  - Verify kubeconfig context removal
  - Verify Docker cleanup

- [ ] T012 [P] [US1] Contract test for `cluster status` command in `tests/contract/cluster_status_test.go`
  - Test status of running cluster
  - Test status with --output=json
  - Test status with --output=yaml
  - Test error: cluster not found
  - Verify output format matches spec

- [ ] T013 [P] [US1] Contract test for `cluster list` command in `tests/contract/cluster_list_test.go`
  - Test listing c8s clusters
  - Test listing all clusters with --all flag
  - Test output in JSON format
  - Test empty list when no clusters exist

- [ ] T014 [P] [US1] Contract test for `cluster start/stop` commands in `tests/contract/cluster_lifecycle_test.go`
  - Test stopping running cluster
  - Test starting stopped cluster
  - Test error: cluster not found
  - Verify state persistence

### Implementation for User Story 1

- [ ] T015 [US1] Implement cluster create logic in `pkg/localenv/cluster/create.go`
  - Load configuration from file or use defaults
  - Validate cluster configuration
  - Check if cluster already exists
  - Check Docker availability
  - Generate k3d cluster config YAML
  - Execute k3d cluster create command
  - Wait for cluster ready (with timeout)
  - Update kubeconfig and switch context
  - Return ClusterStatus

- [ ] T016 [US1] Implement cluster delete logic in `pkg/localenv/cluster/delete.go`
  - Check if cluster exists
  - Prompt for confirmation (unless --force)
  - Execute k3d cluster delete command
  - Remove kubeconfig context
  - Verify Docker cleanup
  - Return success/error status

- [ ] T017 [P] [US1] Implement cluster status logic in `pkg/localenv/cluster/status.go`
  - Query k3d for cluster info
  - Query kubectl for node status
  - Calculate uptime
  - Format output (text, JSON, YAML)
  - Handle cluster not found

- [ ] T018 [P] [US1] Implement cluster list logic in `pkg/localenv/cluster/list.go`
  - Query k3d for all clusters (or filter for c8s clusters)
  - Get status for each cluster
  - Format as table or JSON/YAML
  - Handle empty list

- [ ] T019 [P] [US1] Implement cluster start/stop logic in `pkg/localenv/cluster/lifecycle.go`
  - Start: execute k3d cluster start, wait for ready
  - Stop: execute k3d cluster stop, wait for stopped
  - Handle errors and timeouts

- [ ] T020 [US1] Implement `cluster create` command in `cmd/c8s/commands/dev/cluster.go`
  - Define cluster subcommand with cobra
  - Add create subcommand with flags (--config, --k8s-version, --servers, --agents, --registry, --timeout, --wait)
  - Parse flags and validate inputs
  - Call pkg/localenv/cluster/create.go logic
  - Format and display output
  - Set exit code based on result

- [ ] T021 [US1] Implement `cluster delete` command in `cmd/c8s/commands/dev/cluster.go`
  - Add delete subcommand with flags (--force, --all)
  - Handle confirmation prompt
  - Call pkg/localenv/cluster/delete.go logic
  - Format and display output
  - Set exit code

- [ ] T022 [P] [US1] Implement `cluster status` command in `cmd/c8s/commands/dev/cluster.go`
  - Add status subcommand with flags (--output, --watch)
  - Call pkg/localenv/cluster/status.go logic
  - Format output based on --output flag
  - Handle --watch flag with polling
  - Set exit code

- [ ] T023 [P] [US1] Implement `cluster list` command in `cmd/c8s/commands/dev/cluster.go`
  - Add list subcommand with flags (--output, --all)
  - Call pkg/localenv/cluster/list.go logic
  - Format output
  - Set exit code

- [ ] T024 [P] [US1] Implement `cluster start/stop` commands in `cmd/c8s/commands/dev/cluster.go`
  - Add start subcommand with flags (--wait, --timeout)
  - Add stop subcommand
  - Call pkg/localenv/cluster/lifecycle.go logic
  - Format output
  - Set exit code

- [ ] T025 [US1] Add error handling and logging for cluster commands
  - Standardize error messages with actionable suggestions
  - Add verbose logging for debugging
  - Add colored output support (with --no-color override)
  - Handle common errors gracefully (Docker not running, port conflicts, etc.)

**Checkpoint**: At this point, User Story 1 should be fully functional - developers can create, manage, and delete local clusters independently

---

## Phase 4: User Story 2 - Pipeline Operator Deployment (Priority: P2)

**Goal**: Enable developers to deploy the C8S operator to their local cluster with a single command

**Independent Test**: Create a cluster with US1 commands, run `c8s dev deploy operator`, verify CRDs are registered, operator pod is running, and sample PipelineConfig can be applied

### Contract Tests for User Story 2

- [ ] T026 [P] [US2] Contract test for `deploy operator` command in `tests/contract/deploy_operator_test.go`
  - Test operator deployment with defaults
  - Test deployment with custom image
  - Test deployment to specific cluster
  - Test deployment with custom namespace
  - Test error: cluster not found (exit code 2)
  - Test error: CRD installation fails (exit code 4)
  - Test error: operator deployment fails (exit code 5)
  - Verify CRDs are installed
  - Verify operator pod is running
  - Verify output format matches spec

- [ ] T027 [P] [US2] Contract test for `deploy samples` command in `tests/contract/deploy_samples_test.go`
  - Test deploying all samples
  - Test deploying specific samples with --select
  - Test deploying to custom namespace
  - Test error: samples path not found (exit code 3)
  - Test error: invalid YAML manifests (exit code 4)
  - Verify PipelineConfigs are created

### Implementation for User Story 2

- [ ] T028 [US2] Implement CRD installation logic in `pkg/localenv/deploy/crds.go`
  - Locate CRD manifests (from --crds-path or default config/crd/bases)
  - Apply CRDs using kubectl apply -f
  - Wait for CRDs to be registered
  - Verify CRD installation with kubectl get crds
  - Return success/error with CRD list

- [ ] T029 [US2] Implement image loading logic in `pkg/localenv/deploy/image.go`
  - Check if image exists locally (docker images)
  - Load image into k3d cluster (k3d image import)
  - Handle errors (image not found, import failed)
  - Support custom image names and tags

- [ ] T030 [US2] Implement operator deployment logic in `pkg/localenv/deploy/operator.go`
  - Create namespace if it doesn't exist
  - Locate operator manifests (from --manifests-path or default config/manager)
  - Apply operator manifests using kubectl apply
  - Wait for operator deployment to be ready
  - Check operator pod status
  - Fetch and display operator logs on error
  - Return OperatorStatus

- [ ] T031 [P] [US2] Implement sample deployment logic in `pkg/localenv/samples/deploy.go`
  - Locate sample manifests directory
  - Filter samples based on --select flag
  - Validate YAML manifests
  - Apply samples using kubectl apply
  - Create namespace if it doesn't exist
  - Return list of deployed samples

- [ ] T032 [US2] Implement `deploy operator` command in `cmd/c8s/commands/dev/deploy.go`
  - Define deploy subcommand with cobra
  - Add operator subcommand with flags (--cluster, --image, --image-pull-policy, --namespace, --crds-path, --manifests-path, --wait, --timeout)
  - Validate inputs
  - Call CRD installation (pkg/localenv/deploy/crds.go)
  - Call image loading (pkg/localenv/deploy/image.go)
  - Call operator deployment (pkg/localenv/deploy/operator.go)
  - Format and display progress with checkmarks
  - Display next steps
  - Set exit code based on result

- [ ] T033 [P] [US2] Implement `deploy samples` command in `cmd/c8s/commands/dev/deploy.go`
  - Add samples subcommand with flags (--cluster, --samples-path, --namespace, --select)
  - Validate inputs
  - Call sample deployment logic (pkg/localenv/samples/deploy.go)
  - Format and display output
  - Display next steps
  - Set exit code

- [ ] T034 [US2] Create default operator deployment manifests in `config/manager/manager.yaml` (if not exists)
  - ServiceAccount for controller
  - Role and RoleBinding for RBAC
  - Deployment for operator
  - Use imagePullPolicy: IfNotPresent for local development
  - Add resource limits/requests
  - Add liveness/readiness probes

- [ ] T035 [US2] Add error handling and logging for deploy commands
  - Handle CRD installation failures with specific error messages
  - Handle image loading failures (image not found, etc.)
  - Handle operator deployment failures (pod crash loop, etc.)
  - Provide kubectl commands for manual debugging
  - Add verbose logging for each deployment step

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - clusters can be created and operators can be deployed

---

## Phase 5: User Story 3 - End-to-End Pipeline Test (Priority: P3)

**Goal**: Enable developers to run automated end-to-end tests that validate pipeline execution

**Independent Test**: Deploy operator (US2), run `c8s dev test run`, verify all sample pipelines execute successfully and results are displayed

### Contract Tests for User Story 3

- [ ] T036 [P] [US3] Contract test for `test run` command in `tests/contract/test_run_test.go`
  - Test running all pipeline tests
  - Test running specific pipeline with --pipeline flag
  - Test with --watch flag for real-time progress
  - Test error: cluster not found (exit code 2)
  - Test error: operator not deployed (exit code 3)
  - Test error: tests failed (exit code 4)
  - Test error: timeout (exit code 5)
  - Verify test results output format

- [ ] T037 [P] [US3] Contract test for `test logs` command in `tests/contract/test_logs_test.go`
  - Test viewing logs for all pipelines
  - Test viewing logs for specific pipeline
  - Test --follow flag behavior
  - Test --tail flag behavior
  - Test error: pipeline not found (exit code 3)
  - Verify log output format

### Implementation for User Story 3

- [ ] T038 [US3] Create sample PipelineConfigs in `config/samples/simple-build.yaml`
  - Define simple single-step build pipeline
  - Use lightweight test image (e.g., alpine)
  - Add simple commands that complete quickly
  - Include repository, branch, and step configuration

- [ ] T039 [P] [US3] Create sample PipelineConfig in `config/samples/multi-step.yaml`
  - Define multi-step pipeline with dependencies
  - Use dependsOn to sequence steps
  - Use lightweight images
  - Test step dependency execution order

- [ ] T040 [P] [US3] Create sample PipelineConfig in `config/samples/matrix-build.yaml`
  - Define matrix build with parallel execution
  - Use matrix dimensions (e.g., OS, version)
  - Use lightweight images
  - Test parallel job execution

- [ ] T041 [US3] Implement pipeline test runner in `pkg/localenv/samples/test.go`
  - List PipelineConfigs in namespace
  - Filter by --pipeline flag if specified
  - For each pipeline:
    - Create PipelineRun resource
    - Monitor job creation and status
    - Wait for completion with timeout
    - Collect results (success/failure, duration)
  - Return aggregated test results

- [ ] T042 [P] [US3] Implement pipeline log fetcher in `pkg/localenv/samples/logs.go`
  - Find PipelineRun resources
  - Get associated Job/Pod resources
  - Fetch logs from all containers in pipeline pods
  - Support --follow flag with streaming
  - Support --tail flag for last N lines
  - Format logs by step

- [ ] T043 [US3] Implement `test run` command in `cmd/c8s/commands/dev/test.go`
  - Define test subcommand with cobra
  - Add run subcommand with flags (--cluster, --pipeline, --namespace, --timeout, --watch)
  - Validate operator is deployed before running tests
  - Call pipeline test runner (pkg/localenv/samples/test.go)
  - Display real-time progress if --watch
  - Display summary of test results
  - Set exit code based on results (0 if all pass, 4 if any fail)

- [ ] T044 [P] [US3] Implement `test logs` command in `cmd/c8s/commands/dev/test.go`
  - Add logs subcommand with flags (--cluster, --pipeline, --namespace, --follow, --tail)
  - Call pipeline log fetcher (pkg/localenv/samples/logs.go)
  - Display logs with formatting
  - Handle --follow for streaming logs
  - Set exit code

- [ ] T045 [US3] Add pipeline execution monitoring in `pkg/localenv/samples/monitor.go`
  - Watch PipelineRun status changes
  - Watch Job status changes
  - Watch Pod events
  - Detect failures and extract error messages
  - Calculate execution duration per step
  - Return structured execution status

- [ ] T046 [US3] Add error handling for test commands
  - Detect operator not deployed (check for CRDs and operator pod)
  - Handle pipeline creation failures
  - Handle test timeouts gracefully
  - Provide actionable error messages
  - Suggest debugging commands (kubectl describe, kubectl logs)

**Checkpoint**: All user stories 1-3 should now be independently functional - complete pipeline testing workflow from cluster creation to test execution

---

## Phase 6: User Story 4 - Test Environment Teardown (Priority: P4)

**Goal**: Provide convenient commands for managing cluster lifecycle (stop/start) and complete cleanup

**Independent Test**: Create cluster, deploy operator, run tests, then stop cluster, verify it can be restarted, then delete cluster and verify complete cleanup

### Contract Tests for User Story 4

**Note**: Some tests overlap with US1, but we're testing them in context of complete lifecycle

- [ ] T047 [P] [US4] Integration test for full lifecycle in `tests/integration/localenv/lifecycle_test.go`
  - Create cluster
  - Deploy operator
  - Deploy samples
  - Run tests
  - Stop cluster
  - Verify cluster stopped
  - Start cluster
  - Verify operator still works
  - Delete cluster
  - Verify complete cleanup (no Docker containers, no kubeconfig context)

### Implementation for User Story 4

- [ ] T048 [US4] Add cleanup verification in `pkg/localenv/cluster/cleanup.go`
  - Check for orphaned Docker containers
  - Check for orphaned Docker volumes
  - Check for orphaned kubeconfig contexts
  - Check for orphaned processes
  - Return cleanup status report

- [ ] T049 [US4] Add warning detection for active workloads in `pkg/localenv/cluster/workloads.go`
  - Check for running PipelineRuns
  - Check for pending Jobs
  - Check for active Pods
  - Return list of active workloads
  - Used by delete command to warn user

- [ ] T050 [US4] Enhance delete command with workload warnings
  - Call workload detection before deletion
  - Display warning if workloads are active
  - Offer force cleanup or wait option
  - Proceed with deletion based on user choice
  - Run cleanup verification after deletion

- [ ] T051 [P] [US4] Implement state persistence in `pkg/localenv/cluster/state.go`
  - Save cluster metadata to ~/.c8s/state/clusters.json
  - Track cluster creation time, configuration hash, k3d version
  - Update state on cluster operations
  - Clean up state entries on cluster deletion
  - Used for cluster list and status commands

- [ ] T052 [US4] Add environment recreation support
  - Create environment config file (.c8s/environment.yaml) with cluster + operator config
  - Command to export current environment to file
  - Command to recreate environment from file
  - Useful for team sharing and consistent setups

**Checkpoint**: Complete local test environment workflow is now functional - create, deploy, test, stop, start, delete with full cleanup

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T053 [P] Create local testing guide in `docs/local-testing.md`
  - Link to quickstart.md from specs/
  - Add troubleshooting section
  - Add common workflow examples
  - Add CI/CD integration examples
  - Document all CLI commands with examples

- [ ] T054 [P] Add unit tests for configuration validation in `tests/unit/localenv/validation_test.go`
  - Test all validation functions
  - Test edge cases (invalid names, port conflicts, etc.)
  - Test error message clarity

- [ ] T055 [P] Add unit tests for k3d wrapper in `tests/unit/localenv/k3d_test.go`
  - Mock k3d command execution
  - Test command generation
  - Test output parsing
  - Test error handling

- [ ] T056 [P] Add unit tests for kubectl wrapper in `tests/unit/localenv/kubectl_test.go`
  - Mock kubectl command execution
  - Test manifest application
  - Test resource querying
  - Test context switching

- [ ] T057 [P] Add integration test for cluster creation in `tests/integration/localenv/cluster_test.go`
  - Test full cluster lifecycle (create, status, stop, start, delete)
  - Test with various configurations
  - Test error scenarios
  - Requires Docker to be running

- [ ] T058 [P] Add integration test for operator deployment in `tests/integration/localenv/deploy_test.go`
  - Test CRD installation
  - Test image loading
  - Test operator deployment
  - Test sample deployment
  - Requires cluster to exist

- [ ] T059 [P] Add performance optimizations
  - Cache Docker availability checks
  - Parallelize CRD installations where possible
  - Optimize kubectl wait operations
  - Add progress indicators for long operations

- [ ] T060 [P] Add output formatting utilities in `pkg/localenv/output/format.go`
  - Table formatting for list commands
  - JSON/YAML formatting
  - Colored output with checkmarks/crosses
  - Progress spinners
  - Respect --no-color flag

- [ ] T061 [P] Add environment variable support
  - C8S_DEV_CLUSTER for default cluster name
  - C8S_DEV_CONFIG for default config file
  - Document environment variables in CLI help
  - Add precedence rules (env vars < flags)

- [ ] T062 Add CLI help and usage documentation
  - Add detailed command help text
  - Add examples for each command
  - Add common flags documentation
  - Test help output format

- [ ] T063 [P] Run quickstart.md validation
  - Follow quickstart guide step by step
  - Verify all commands work as documented
  - Update quickstart if needed
  - Add quickstart to CI pipeline

- [ ] T064 [P] Add Makefile targets for local development
  - `make dev-cluster-create`: Create test cluster
  - `make dev-cluster-delete`: Delete test cluster
  - `make dev-deploy`: Deploy operator to local cluster
  - `make dev-test`: Run end-to-end tests
  - `make dev-reload`: Quick iteration (rebuild + redeploy + test)

- [ ] T065 Update main project README.md
  - Add section on local testing
  - Link to docs/local-testing.md
  - Add prerequisites (Docker, k3d)
  - Add quick start example

- [ ] T066 [P] Add default cluster configuration in `.c8s/cluster-defaults.yaml`
  - Create template configuration file
  - Document all available options
  - Include comments explaining each field
  - Used as fallback if --config not specified

- [ ] T067 Final code cleanup and refactoring
  - Remove any debug code
  - Consolidate duplicate logic
  - Ensure consistent error handling
  - Add package documentation comments
  - Run go fmt and golangci-lint

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-6)**: All depend on Foundational phase completion
  - US1 (Cluster Creation): Can start after Foundational - No dependencies on other stories
  - US2 (Operator Deployment): Can start after Foundational - No dependencies on US1 for implementation, but requires US1 cluster commands to test
  - US3 (Pipeline Testing): Can start after Foundational - Requires US2 for testing
  - US4 (Teardown): Can start after Foundational - Extends US1 cluster commands
- **Polish (Phase 7)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Fully independent
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent implementation, requires US1 for testing
- **User Story 3 (P3)**: Requires US1 + US2 for testing (cluster must exist, operator must be deployed)
- **User Story 4 (P4)**: Extends US1 - Adds cleanup and lifecycle features to existing cluster commands

### Within Each User Story

- Contract tests MUST be written and FAIL before implementation
- Implementation tasks follow logical dependency order
- Commands depend on their corresponding pkg/ logic being implemented
- Error handling comes after core functionality

### Parallel Opportunities

- **Phase 1 (Setup)**: All 3 tasks can run in parallel (T001, T002, T003)
- **Phase 2 (Foundational)**: Tasks T005, T006, T008, T009 can run in parallel after T004 completes
- **User Story 1 (Cluster)**:
  - All contract tests (T010-T014) can run in parallel
  - Implementation: T017, T018, T019 can run in parallel after T015-T016
  - Commands: T022, T023, T024 can run in parallel after T020-T021
- **User Story 2 (Deploy)**:
  - Contract tests T026, T027 can run in parallel
  - Implementation: T031, T033, T034 can run in parallel after T028-T030
- **User Story 3 (Test)**:
  - Contract tests T036, T037 can run in parallel
  - Sample configs T038, T039, T040 can run in parallel
  - Implementation: T042, T044 can run in parallel after T041, T043
- **Phase 7 (Polish)**: Most tasks (T053-T061, T063-T066) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all contract tests for User Story 1 together:
Task: "Contract test for cluster create command" (tests/contract/cluster_create_test.go)
Task: "Contract test for cluster delete command" (tests/contract/cluster_delete_test.go)
Task: "Contract test for cluster status command" (tests/contract/cluster_status_test.go)
Task: "Contract test for cluster list command" (tests/contract/cluster_list_test.go)
Task: "Contract test for cluster start/stop commands" (tests/contract/cluster_lifecycle_test.go)

# After T015-T016 complete, launch these together:
Task: "Implement cluster status logic" (pkg/localenv/cluster/status.go)
Task: "Implement cluster list logic" (pkg/localenv/cluster/list.go)
Task: "Implement cluster start/stop logic" (pkg/localenv/cluster/lifecycle.go)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T009) - CRITICAL
3. Complete Phase 3: User Story 1 (T010-T025)
4. **STOP and VALIDATE**: Test cluster create/delete/status/list commands independently
5. Deploy/demo if ready

**MVP Deliverable**: Developers can create local k8s clusters with `c8s dev cluster create`, manage them, and delete them. This is immediately useful even without operator deployment.

### Incremental Delivery

1. **Foundation** (Phases 1-2): Project structure + core types ready
2. **MVP** (Phase 3): Cluster management → Deploy/Demo ✅
3. **Phase 4**: + Operator deployment → Deploy/Demo ✅
4. **Phase 5**: + Pipeline testing → Deploy/Demo ✅
5. **Phase 6**: + Lifecycle management → Deploy/Demo ✅
6. **Phase 7**: Polish and optimize

Each phase adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T009)
2. Once Foundational is done, parallelize:
   - **Developer A**: User Story 1 (T010-T025) - Cluster management
   - **Developer B**: User Story 2 (T026-T035) - Operator deployment
   - **Developer C**: User Story 3 (T036-T046) - Pipeline testing
   - **Developer D**: User Story 4 (T047-T052) - Lifecycle/cleanup
3. Stories complete and integrate independently
4. Team collaborates on Phase 7 polish tasks

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Write contract tests first, verify they fail, then implement
- Commit after each task or logical group of tasks
- Stop at checkpoints to validate story independently
- US3 depends on US1+US2 at test time but can be implemented in parallel
- US4 extends US1 by adding cleanup and lifecycle features

---

## Summary

**Total Tasks**: 67 tasks
- Phase 1 (Setup): 3 tasks
- Phase 2 (Foundational): 6 tasks
- Phase 3 (US1 - Cluster): 16 tasks (5 tests + 11 implementation)
- Phase 4 (US2 - Deploy): 10 tasks (2 tests + 8 implementation)
- Phase 5 (US3 - Testing): 11 tasks (2 tests + 9 implementation)
- Phase 6 (US4 - Teardown): 5 tasks (1 test + 4 implementation)
- Phase 7 (Polish): 15 tasks

**Parallel Opportunities**: ~35 tasks can run in parallel at various points

**MVP Scope**: Phases 1-3 (25 tasks) = Cluster management foundation

**Independent Test Criteria**:
- **US1**: Create cluster → kubectl get nodes shows nodes → Delete cluster → No Docker containers remain
- **US2**: Deploy operator → kubectl get crds shows C8S CRDs → Operator pod running → Apply sample PipelineConfig succeeds
- **US3**: Run test → All sample pipelines execute → Test results displayed → View logs shows pipeline output
- **US4**: Stop cluster → Cluster stopped but data persists → Start cluster → Operator still functional → Delete with cleanup verification

**Suggested Implementation Order**:
1. MVP: Phases 1-3 (cluster management)
2. Phase 4 (operator deployment)
3. Phase 5 (pipeline testing)
4. Phase 6 (lifecycle/cleanup)
5. Phase 7 (polish)
