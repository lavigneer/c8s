# Implementation Tasks: Kubernetes-Native CI System

**Feature**: 001-build-a-continuous
**Branch**: `001-build-a-continuous`
**Generated**: 2025-10-12
**Status**: Ready for Implementation

## Overview

This document provides dependency-ordered implementation tasks for building the C8S Kubernetes-native CI system. Tasks are organized by user story to enable independent, incremental delivery.

**Total Tasks**: 89 tasks across 7 phases
**Parallel Opportunities**: 45 parallelizable tasks marked with [P]
**Estimated MVP**: Phase 1-3 (Setup + Foundational + User Story 1) = ~42 tasks

---

## Task Organization

- **Phase 1**: Setup & Project Initialization (12 tasks)
- **Phase 2**: Foundational Infrastructure (10 tasks) - **BLOCKING** for all user stories
- **Phase 3**: User Story 1 (P1) - Basic Pipeline Execution (22 tasks) - **MVP**
- **Phase 4**: User Story 2 (P2) - Observability (15 tasks)
- **Phase 5**: User Story 3 (P2) - Secret Management (12 tasks)
- **Phase 6**: User Story 4 (P3) - Resource Optimization (10 tasks)
- **Phase 7**: User Story 5 (P3) - Parallel Execution (8 tasks)

**Legend**:
- **[P]** = Parallelizable (can be done simultaneously with other [P] tasks in same phase)
- **[US1]**, **[US2]**, etc. = User Story tags
- **Dependencies**: Listed before each phase

---

## Phase 1: Setup & Project Initialization

**Goal**: Bootstrap Go project with Kubernetes tooling and establish development environment

**Prerequisites**: None (starting from scratch)

**Tasks**:

### T001: Initialize Go module and project structure
- Create `go.mod` with module name `github.com/org/c8s` and Go 1.25
- Create directory structure: `cmd/`, `pkg/`, `tests/`, `config/`, `web/`, `deploy/`
- Create subdirectories per plan.md structure
- Initialize git repository if not already done
- **File**: `go.mod`, directory structure
- **No dependencies**

### T002: [P] Add core Kubernetes dependencies
- Add `k8s.io/client-go@v0.28+` to go.mod
- Add `sigs.k8s.io/controller-runtime@v0.16+` to go.mod
- Add `k8s.io/apimachinery@v0.28+` to go.mod
- Add `k8s.io/api@v0.28+` to go.mod
- Run `go mod tidy` and `go mod download`
- **File**: `go.mod`, `go.sum`
- **Depends on**: T001

### T003: [P] Add utility dependencies
- Add `gopkg.in/yaml.v3` for YAML parsing
- Add `github.com/stretchr/testify@v1.8+` for testing assertions
- Add AWS SDK Go v2 for S3 client (`github.com/aws/aws-sdk-go-v2/service/s3`)
- Run `go mod tidy`
- **File**: `go.mod`, `go.sum`
- **Depends on**: T001

### T004: [P] Create Makefile for common operations
- Add targets: `build`, `test`, `lint`, `generate`, `install-crds`, `deploy`
- Add target for running controller locally: `run-controller`
- Add target for code generation: `generate-crds`
- Add target for building CLI: `build-cli`
- **File**: `Makefile`
- **Depends on**: T001

### T005: [P] Setup kubebuilder project markers
- Add `//+kubebuilder` markers for CRD generation in empty marker file
- Create `PROJECT` file with kubebuilder metadata
- Create `.gitignore` with Go and Kubernetes patterns
- **File**: `PROJECT`, `.gitignore`, `hack/boilerplate.go.txt`
- **Depends on**: T001

### T006: [P] Create README.md with project overview
- Document: What is C8S, architecture diagram, quick start link
- Document: Prerequisites (K8s 1.24+, kubectl, Go 1.25)
- Document: Development setup instructions
- Link to quickstart.md in specs/ directory
- **File**: `README.md`
- **Depends on**: T001

### T007: [P] Setup CI/CD workflow skeleton
- Create `.github/workflows/ci.yml` for GitHub Actions
- Add jobs: test, lint, build
- Add integration test job with kind cluster
- Configure to run on PR and main branch
- **File**: `.github/workflows/ci.yml`
- **Depends on**: T001

### T008: [P] Create Docker multi-stage Dockerfile
- Stage 1: Build Go binaries (controller, api-server, webhook, cli)
- Stage 2: Runtime image with binaries
- Use distroless or alpine base image
- **File**: `Dockerfile`
- **Depends on**: T001

### T009: [P] Setup golangci-lint configuration
- Create `.golangci.yml` with linters: govet, staticcheck, errcheck, gosimple
- Configure exclusions for generated code
- Add to Makefile lint target
- **File**: `.golangci.yml`
- **Depends on**: T001, T004

### T010: [P] Create config/samples directory with example CRDs
- Create sample PipelineConfig YAML
- Create sample PipelineRun YAML
- Create sample RepositoryConnection YAML
- These serve as examples and test fixtures
- **File**: `config/samples/pipelineconfig.yaml`, `config/samples/pipelinerun.yaml`, `config/samples/repositoryconnection.yaml`
- **Depends on**: T001

### T011: [P] Setup envtest for integration testing
- Add envtest setup script in `hack/setup-envtest.sh`
- Configure envtest to download K8s binaries (1.24+)
- Add helper functions in `tests/testutil/envtest.go`
- **File**: `hack/setup-envtest.sh`, `tests/testutil/envtest.go`
- **Depends on**: T002, T003

### T012: Document development workflow
- Create `docs/development.md` with:
  - How to run tests locally
  - How to run controller locally against kind cluster
  - How to debug with delve
  - Code generation workflow
- **File**: `docs/development.md`
- **Depends on**: T001

**✓ Checkpoint**: Project structure established, dependencies installed, development tooling ready

---

## Phase 2: Foundational Infrastructure

**Goal**: Implement core CRD types and controller framework that ALL user stories depend on

**Prerequisites**: Phase 1 complete

**CRITICAL**: These tasks MUST complete before any user story implementation can begin.

**Tasks**:

### T013: Define PipelineConfig CRD type (pkg/apis/v1alpha1/)
- Create `pkg/apis/v1alpha1/pipelineconfig_types.go`
- Define `PipelineConfig` struct with kubebuilder markers per data-model.md
- Define `PipelineConfigSpec` with: repository, branches, steps, timeout, matrix
- Define `PipelineStep` struct with: name, image, commands, dependsOn, resources, timeout, artifacts, secrets, conditional
- Define `PipelineConfigStatus` with: lastRun, totalRuns, successRate
- Add kubebuilder validation markers for required fields, patterns, defaults
- **File**: `pkg/apis/v1alpha1/pipelineconfig_types.go`
- **Depends on**: T002

### T014: Define PipelineRun CRD type (pkg/apis/v1alpha1/)
- Create `pkg/apis/v1alpha1/pipelinerun_types.go`
- Define `PipelineRun` struct with kubebuilder markers per data-model.md
- Define `PipelineRunSpec` with: pipelineConfigRef, commit, branch, triggeredBy, triggeredAt, matrixIndex
- Define `PipelineRunStatus` with: phase (enum), startTime, completionTime, steps[]
- Define `StepStatus` struct with: name, phase, jobName, startTime, completionTime, exitCode, logURL, artifactURLs
- Add kubebuilder validation and printer columns
- **File**: `pkg/apis/v1alpha1/pipelinerun_types.go`
- **Depends on**: T002

### T015: Define RepositoryConnection CRD type (pkg/apis/v1alpha1/)
- Create `pkg/apis/v1alpha1/repositoryconnection_types.go`
- Define `RepositoryConnection` struct with kubebuilder markers per data-model.md
- Define `RepositoryConnectionSpec` with: repository, provider (enum), webhookSecretRef, authSecretRef, pipelineConfigRef
- Define `RepositoryConnectionStatus` with: webhookURL, webhookRegistered, lastEvent
- Add validation markers
- **File**: `pkg/apis/v1alpha1/repositoryconnection_types.go`
- **Depends on**: T002

### T016: Generate CRD manifests and DeepCopy methods
- Create `pkg/apis/v1alpha1/groupversion_info.go` with scheme registration
- Run `controller-gen` to generate CRD YAML files in `config/crd/`
- Run `controller-gen` to generate `zz_generated.deepcopy.go` with DeepCopy methods
- Verify generated CRDs match data-model.md schemas
- **File**: `config/crd/bases/`, `pkg/apis/v1alpha1/zz_generated.deepcopy.go`
- **Depends on**: T013, T014, T015

### T017: Create controller-runtime manager setup
- Create `cmd/controller/main.go` entry point
- Initialize controller-runtime manager with scheme
- Register CRD types (PipelineConfig, PipelineRun, RepositoryConnection)
- Setup signal handling for graceful shutdown
- Add flags for: kubeconfig, metrics-addr, leader-election
- **File**: `cmd/controller/main.go`
- **Depends on**: T016

### T018: Implement PipelineRun controller scaffold
- Create `pkg/controller/pipelinerun_controller.go`
- Implement `Reconcile()` method scaffold (empty for now)
- Register controller with manager
- Setup watch on PipelineRun resources
- Setup watch on Jobs (owned by PipelineRun)
- **File**: `pkg/controller/pipelinerun_controller.go`
- **Depends on**: T016, T017

### T019: [P] Create common types and constants package
- Create `pkg/types/constants.go` with: label keys, annotation keys, finalizer names
- Create `pkg/types/conditions.go` with condition types (JobsCreated, StepsCompleted, etc.)
- Define common error types
- **File**: `pkg/types/constants.go`, `pkg/types/conditions.go`, `pkg/types/errors.go`
- **Depends on**: T001

### T020: [P] Implement S3 storage client interface
- Create `pkg/storage/interface.go` with `StorageClient` interface
- Define methods: `UploadLog(ctx, key, content)`, `DownloadLog(ctx, key)`, `GenerateSignedURL(ctx, key, expiry)`
- Create `pkg/storage/s3/client.go` implementing interface using AWS SDK
- Add configuration struct for bucket, region, credentials
- **File**: `pkg/storage/interface.go`, `pkg/storage/s3/client.go`
- **Depends on**: T003

### T021: [P] Implement pipeline YAML parser
- Create `pkg/parser/parser.go` with `Parse(yamlContent []byte) (*PipelineConfig, error)`
- Use `gopkg.in/yaml.v3` to unmarshal YAML
- Validate against JSON schema from contracts/pipeline-config-schema.json
- Implement error handling with clear messages
- **File**: `pkg/parser/parser.go`
- **Depends on**: T003, T013

### T022: [P] Setup RBAC manifests for controller
- Create `config/rbac/role.yaml` with permissions:
  - PipelineConfigs, PipelineRuns, RepositoryConnections: get, list, watch, create, update, patch
  - Jobs, Pods, Secrets, ConfigMaps: get, list, watch, create, delete
  - Events: create, patch
- Create `config/rbac/role_binding.yaml`
- Create `config/rbac/service_account.yaml`
- **File**: `config/rbac/role.yaml`, `config/rbac/role_binding.yaml`, `config/rbac/service_account.yaml`
- **Depends on**: T016

**✓ Checkpoint**: CRDs defined, controller scaffold created, core infrastructure ready for user story implementation

---

## Phase 3: User Story 1 (P1) - Basic Pipeline Execution [MVP]

**Goal**: Implement core CI functionality - detect commits, execute pipeline steps in isolated Jobs, report results

**User Story**: A developer commits code, CI system automatically runs tests in containers, reports success/failure

**Independent Test**: Create sample repository with `.c8s.yaml`, push commit, verify Jobs created and PipelineRun status updated

**Prerequisites**: Phase 2 complete

**Tasks**:

### T023: [US1] Implement DAG scheduler for step dependencies
- Create `pkg/scheduler/dag.go` with DAG (directed acyclic graph) implementation
- Implement `BuildDAG(steps []PipelineStep) (*DAG, error)` to construct dependency graph
- Implement `TopologicalSort() ([][]string, error)` to return execution order in layers
- Implement cycle detection to reject circular dependencies
- Return error if `dependsOn` references non-existent steps
- **File**: `pkg/scheduler/dag.go`
- **Depends on**: T013

### T024: [US1] Implement step scheduler
- Create `pkg/scheduler/scheduler.go`
- Implement `Schedule(pipelineConfig *PipelineConfig) (*Schedule, error)`
- Use DAG to determine which steps can run in parallel (same layer)
- Return `Schedule` struct with ordered layers of steps
- **File**: `pkg/scheduler/scheduler.go`
- **Depends on**: T023

### T025: [US1] Implement Job creation from PipelineStep
- Create `pkg/controller/job_manager.go`
- Implement `CreateJobForStep(step *PipelineStep, pipelineRun *PipelineRun) (*batchv1.Job, error)`
- Generate Job with:
  - Init container for git clone (using commit SHA from PipelineRun)
  - Main container with step.image, step.commands
  - Resource requests/limits from step.resources
  - Owner reference to PipelineRun (for garbage collection)
  - Labels: pipeline-run, step-name
  - TTL: `ttlSecondsAfterFinished: 3600`
- **File**: `pkg/controller/job_manager.go`
- **Depends on**: T013, T014, T018

### T026: [US1] Implement PipelineRun reconciliation loop core logic
- In `pkg/controller/pipelinerun_controller.go`, implement `Reconcile()`:
  - Fetch PipelineRun from API
  - If phase is terminal (Succeeded/Failed/Cancelled), skip reconciliation
  - Fetch referenced PipelineConfig
  - If PipelineRun.status.phase is empty, set to "Pending"
  - Call scheduler to get execution order
  - Create Jobs for steps that are ready to execute (dependencies satisfied)
  - Update PipelineRun status with step statuses from Jobs
  - Handle phase transitions: Pending → Running → Succeeded/Failed
- **File**: `pkg/controller/pipelinerun_controller.go`
- **Depends on**: T018, T024, T025

### T027: [US1] Implement status updater for PipelineRun
- Create `pkg/controller/status_updater.go`
- Implement `UpdatePipelineRunStatus(ctx, pipelineRun, jobs) error`
- For each Job, extract status (Pending/Running/Succeeded/Failed) and update PipelineRun.status.steps[]
- Update PipelineRun.status.phase based on overall step statuses
- Update timestamps (startTime, completionTime)
- Handle status conflicts and retries
- **File**: `pkg/controller/status_updater.go`
- **Depends on**: T014, T018

### T028: [US1] [P] Implement pipeline config validator
- Create `pkg/parser/validator.go`
- Implement `Validate(config *PipelineConfig) error`
- Validate: repository URL format, step names unique, no circular dependencies
- Validate: timeout values parse correctly, resource values valid K8s quantities
- Return structured validation errors
- **File**: `pkg/parser/validator.go`
- **Depends on**: T021

### T029: [US1] [P] Implement webhook receiver service entry point
- Create `cmd/webhook/main.go` entry point
- Initialize HTTP server using `net/http`
- Setup routes using `http.ServeMux`:
  - `POST /webhooks/github`
  - `POST /webhooks/gitlab`
  - `POST /webhooks/bitbucket`
- Add flags for: port, kubeconfig
- Initialize Kubernetes client for creating PipelineRuns
- **File**: `cmd/webhook/main.go`
- **Depends on**: T002

### T030: [US1] Implement GitHub webhook handler
- Create `pkg/webhook/github.go`
- Implement `HandleGitHubWebhook(w http.ResponseWriter, r *http.Request)`
- Verify HMAC signature using X-Hub-Signature-256 header
- Parse push event JSON payload
- Extract: repository URL, commit SHA, branch, author
- Look up RepositoryConnection by repository URL
- Validate webhook secret matches RepositoryConnection.spec.webhookSecretRef
- Create PipelineRun CRD with appropriate spec fields
- Return 200 OK immediately (async processing)
- **File**: `pkg/webhook/github.go`
- **Depends on**: T015, T029

### T031: [US1] [P] Implement GitLab webhook handler
- Create `pkg/webhook/gitlab.go`
- Implement `HandleGitLabWebhook(w http.ResponseWriter, r *http.Request)`
- Verify X-Gitlab-Token header
- Parse push event JSON payload
- Extract repository, commit, branch information
- Create PipelineRun CRD
- **File**: `pkg/webhook/gitlab.go`
- **Depends on**: T015, T029

### T032: [US1] [P] Implement Bitbucket webhook handler
- Create `pkg/webhook/bitbucket.go`
- Implement `HandleBitbucketWebhook(w http.ResponseWriter, r *http.Request)`
- Verify HMAC signature
- Parse push event JSON payload
- Create PipelineRun CRD
- **File**: `pkg/webhook/bitbucket.go`
- **Depends on**: T015, T029

### T033: [US1] [P] Create deployment manifests for controller
- Create `config/manager/deployment.yaml` for controller
- Configure: replicas: 1, image, resource requests/limits
- Mount ServiceAccount from RBAC
- Add liveness/readiness probes
- **File**: `config/manager/deployment.yaml`
- **Depends on**: T022

### T034: [US1] [P] Create deployment manifests for webhook service
- Create `deploy/webhook-deployment.yaml`
- Configure: replicas: 2 (for HA), image, resource limits
- Create `deploy/webhook-service.yaml` (ClusterIP or LoadBalancer)
- Create `deploy/webhook-ingress.yaml` (optional, for external access)
- **File**: `deploy/webhook-deployment.yaml`, `deploy/webhook-service.yaml`, `deploy/webhook-ingress.yaml`
- **Depends on**: T029

### T035: [US1] [P] Implement CLI tool entry point and run command
- Create `cmd/c8s/main.go` with CLI framework (cobra or stdlib flags)
- Implement `run` command: `c8s run <pipeline-config-name> --commit=<sha> --branch=<name>`
- Command creates PipelineRun CRD using kubectl client
- Print created PipelineRun name
- **File**: `cmd/c8s/main.go`, `pkg/cli/run.go`
- **Depends on**: T002, T014

### T036: [US1] [P] Implement CLI `get` command
- Implement `c8s get runs` command to list PipelineRuns
- Implement `c8s get runs <name>` to show specific PipelineRun details
- Format output as table with columns: Name, Config, Commit, Phase, Age
- Add `--namespace` flag
- **File**: `pkg/cli/get.go`
- **Depends on**: T035

### T037: [US1] [P] Implement CLI `validate` command
- Implement `c8s validate <pipeline-yaml-file>` command
- Parse YAML file using pkg/parser
- Run validator checks
- Print validation errors or "Valid!" message
- Exit with code 0 (valid) or 1 (invalid)
- **File**: `pkg/cli/validate.go`
- **Depends on**: T028, T035

### T038: [US1] Implement git clone init container logic
- Create helper function in `pkg/controller/job_manager.go`
- Generate init container spec with:
  - Image: `alpine/git:latest` or similar
  - Command: `git clone <repo> --branch <branch> --single-branch /workspace && cd /workspace && git checkout <commit>`
  - Volume mount: `/workspace` (emptyDir shared with main container)
  - Handle authentication if RepositoryConnection.spec.authSecretRef is set
- **File**: `pkg/controller/job_manager.go` (update)
- **Depends on**: T025

### T039: [US1] Implement workspace volume sharing between init and main container
- Update `CreateJobForStep()` to add emptyDir volume named "workspace"
- Mount to both init container (git clone) and main container (at /workspace)
- Set working directory of main container to `/workspace`
- **File**: `pkg/controller/job_manager.go` (update)
- **Depends on**: T038

### T040: [US1] Implement timeout enforcement at Job level
- Update `CreateJobForStep()` to set `activeDeadlineSeconds` on Job based on step.timeout
- Parse timeout string ("30m", "2h") to seconds
- Controller watches for Job timeout and updates PipelineRun status to Failed
- **File**: `pkg/controller/job_manager.go` (update)
- **Depends on**: T025

### T041: [US1] Implement resource cleanup finalizer
- Add finalizer to PipelineRun when created
- In reconcile loop, if PipelineRun is being deleted:
  - Delete all owned Jobs
  - Wait for Jobs to terminate
  - Remove finalizer
- This ensures Jobs are cleaned up when PipelineRun is deleted
- **File**: `pkg/controller/pipelinerun_controller.go` (update)
- **Depends on**: T026

### T042: [US1] [P] Write unit tests for DAG scheduler
- Create `tests/unit/scheduler_test.go`
- Test cases:
  - Empty steps returns empty schedule
  - Single step with no dependencies
  - Linear dependencies (A→B→C)
  - Parallel steps (A, B both depend on nothing)
  - Complex DAG with mixed parallel/sequential
  - Circular dependency detection (should error)
  - Non-existent dependency reference (should error)
- Use testify assertions
- **File**: `tests/unit/scheduler_test.go`
- **Depends on**: T023, T024, T003

### T043: [US1] [P] Write unit tests for pipeline parser
- Create `tests/unit/parser_test.go`
- Test cases:
  - Valid minimal pipeline YAML parses correctly
  - Invalid YAML returns error
  - Missing required fields (name, steps) returns validation error
  - Invalid step names (spaces, special chars) return error
  - Valid resource values parse correctly
  - Invalid timeout format returns error
- **File**: `tests/unit/parser_test.go`
- **Depends on**: T021, T028, T003

### T044: [US1] Write integration test for basic pipeline execution
- Create `tests/integration/pipeline_execution_test.go`
- Use envtest to start fake K8s API server
- Create PipelineConfig CRD with simple 2-step pipeline (test → build)
- Create PipelineRun CRD
- Run controller reconcile loop
- Assert:
  - Jobs created for both steps
  - Step 2 Job created only after Step 1 Job succeeds (simulate success)
  - PipelineRun status progresses: Pending → Running → Succeeded
  - Step statuses updated correctly
- **File**: `tests/integration/pipeline_execution_test.go`
- **Depends on**: T026, T027, T011

**✓ Checkpoint User Story 1**: Basic CI pipeline working end-to-end. Commits trigger pipelines, steps execute in Jobs, results reported. **MVP COMPLETE!**

---

## Phase 4: User Story 2 (P2) - Observability

**Goal**: Add monitoring, logging, and dashboard capabilities so developers can track pipeline execution

**User Story**: Developer views real-time logs, tracks pipeline progress, reviews execution history via CLI and dashboard

**Independent Test**: Run pipeline, stream logs via CLI, check dashboard shows live status

**Prerequisites**: Phase 3 complete (User Story 1 working)

**Tasks**:

### T045: [US2] Implement log collection from Job Pods
- Create `pkg/controller/log_collector.go`
- Implement `CollectLogs(ctx, pod) ([]byte, error)` using K8s client `GetLogs()`
- Stream Pod logs as they're written
- Buffer last 10MB in memory for real-time streaming (circular buffer)
- **File**: `pkg/controller/log_collector.go`
- **Depends on**: T002

### T046: [US2] Implement log uploader to S3
- In `pkg/controller/log_collector.go`, add `UploadLogsToStorage(ctx, pipelineRun, step, logs)`
- Use S3 storage client to upload logs
- Key format: `{namespace}/{pipelinerun-name}/{step-name}.log`
- Update PipelineRun.status.steps[].logURL with S3 URL or signed URL
- Handle upload failures gracefully (retry)
- **File**: `pkg/controller/log_collector.go` (update)
- **Depends on**: T020, T045

### T047: [US2] Integrate log collection into reconcile loop
- Update `pkg/controller/pipelinerun_controller.go` reconcile logic:
  - When Job completes, collect logs from Pod
  - Upload to S3
  - Update PipelineRun status with logURL
- Handle Pod not found (Job cleanup race condition)
- **File**: `pkg/controller/pipelinerun_controller.go` (update)
- **Depends on**: T026, T046

### T048: [US2] Implement API server entry point
- Create `cmd/api-server/main.go` entry point
- Initialize HTTP server using `net/http` and `http.ServeMux`
- Setup routes per OpenAPI spec:
  - `GET /api/v1/namespaces/{ns}/pipelineconfigs`
  - `GET /api/v1/namespaces/{ns}/pipelineruns`
  - `GET /api/v1/namespaces/{ns}/pipelineruns/{name}/logs/{step}`
- Initialize Kubernetes client
- Add flags for: port, kubeconfig, enable-dashboard
- **File**: `cmd/api-server/main.go`
- **Depends on**: T002

### T049: [US2] Implement PipelineConfig API handlers
- Create `pkg/api/handlers/pipelineconfig.go`
- Implement handlers:
  - `ListPipelineConfigs(w, r)` - list PipelineConfigs in namespace
  - `GetPipelineConfig(w, r)` - get single PipelineConfig
  - `CreatePipelineConfig(w, r)` - create from JSON body
  - `UpdatePipelineConfig(w, r)` - update existing
  - `DeletePipelineConfig(w, r)` - delete by name
- Use K8s client to interact with CRDs
- Return JSON responses matching OpenAPI schemas
- **File**: `pkg/api/handlers/pipelineconfig.go`
- **Depends on**: T013, T048

### T050: [US2] Implement PipelineRun API handlers
- Create `pkg/api/handlers/pipelinerun.go`
- Implement handlers:
  - `ListPipelineRuns(w, r)` - list with optional filters (phase, config)
  - `GetPipelineRun(w, r)` - get single PipelineRun with full status
  - `CreatePipelineRun(w, r)` - manual trigger
  - `DeletePipelineRun(w, r)` - cancel running pipeline
- **File**: `pkg/api/handlers/pipelinerun.go`
- **Depends on**: T014, T048

### T051: [US2] Implement log streaming API handler
- Create `pkg/api/handlers/logs.go`
- Implement `GetStepLogs(w, r)`:
  - Extract namespace, pipelinerun name, step name from path
  - If `?follow=true` query param, upgrade to WebSocket and stream logs
  - If no follow, fetch from S3 and return as plain text
  - For WebSocket: connect to in-memory log buffer and stream real-time logs
- Use stdlib `net/http` WebSocket support (Upgrader)
- **File**: `pkg/api/handlers/logs.go`
- **Depends on**: T045, T048

### T052: [US2] [P] Implement in-memory circular buffer for log streaming
- Create `pkg/storage/buffer.go`
- Implement `CircularBuffer` struct with fixed size (10MB per step)
- Implement `Write(data []byte)`, `Read() []byte`, `Subscribe() <-chan []byte`
- Thread-safe with mutex
- Controller writes to buffer as logs are collected
- API handlers subscribe to buffer for real-time streaming
- **File**: `pkg/storage/buffer.go`
- **Depends on**: T001

### T053: [US2] [P] Implement middleware for CORS and auth
- Create `pkg/api/middleware/cors.go` with CORS middleware
- Create `pkg/api/middleware/auth.go` with JWT/OIDC validation (optional, can be no-op initially)
- Apply middleware to API server routes
- **File**: `pkg/api/middleware/cors.go`, `pkg/api/middleware/auth.go`
- **Depends on**: T048

### T054: [US2] [P] Create HTMX dashboard layout template
- Create `web/templates/layout.html` with base HTML structure
- Include: `<head>` with HTMX and Tailwind CSS from CDN
- Define `<nav>` with links to Pipelines, Runs, Settings
- Define `<main>` content area
- Add WebSocket setup for live updates
- **File**: `web/templates/layout.html`
- **Depends on**: T001

### T055: [US2] [P] Create HTMX dashboard pipelines page
- Create `web/templates/pipelines.html` extending layout
- Show table of PipelineConfigs with: name, repository, last run, success rate
- Add HTMX attributes for auto-refresh every 5s: `hx-get="/api/v1/namespaces/default/pipelineconfigs" hx-trigger="every 5s"`
- Add link to trigger manual run
- **File**: `web/templates/pipelines.html`
- **Depends on**: T054

### T056: [US2] [P] Create HTMX dashboard runs page
- Create `web/templates/runs.html` extending layout
- Show table of PipelineRuns with: name, config, commit, phase, age
- Add HTMX attributes for auto-refresh
- Add click handler to view logs
- Phase badge with colors: green (Succeeded), red (Failed), yellow (Running), gray (Pending)
- **File**: `web/templates/runs.html`
- **Depends on**: T054

### T057: [US2] [P] Create HTMX dashboard logs page
- Create `web/templates/logs.html` extending layout
- Show step-by-step execution status
- Add log viewer with live streaming via HTMX WebSocket extension
- Add download logs button
- Show step resource usage if available
- **File**: `web/templates/logs.html`
- **Depends on**: T054

### T058: [US2] Implement dashboard HTTP handlers
- Create `pkg/api/handlers/dashboard.go`
- Implement `ServeDashboard(w, r)` - render pipelines page
- Implement `ServeRuns(w, r)` - render runs page
- Implement `ServeLogs(w, r)` - render logs page
- Use Go `html/template` to render HTMX templates
- Only enabled if `--enable-dashboard` flag is true
- **File**: `pkg/api/handlers/dashboard.go`
- **Depends on**: T054, T055, T056, T057

### T059: [US2] Implement CLI `logs` command
- Create `pkg/cli/logs.go`
- Implement `c8s logs <pipelinerun-name> --step=<step-name> --follow`
- If `--follow`, open WebSocket to API server and stream logs to stdout
- If no `--follow`, fetch logs from API and print
- Add `--tail=N` flag to limit lines
- **File**: `pkg/cli/logs.go`
- **Depends on**: T035, T051

**✓ Checkpoint User Story 2**: Observability complete. Logs streaming, dashboard showing live status, CLI providing rich inspection tools.

---

## Phase 5: User Story 3 (P2) - Secret Management

**Goal**: Securely inject secrets into pipelines and mask values in logs

**User Story**: Developer references secrets in pipeline config, secrets injected into containers, values never appear in logs

**Independent Test**: Create pipeline with secret reference, run pipeline, verify secret value in container, verify masked in logs

**Prerequisites**: Phase 3 complete (Phase 4 optional but recommended)

**Tasks**:

### T060: [US3] Implement secret injection into Job Pods
- Update `pkg/controller/job_manager.go` `CreateJobForStep()`
- For each secret reference in step.secrets:
  - Fetch Secret from K8s API
  - Validate Secret exists and has required key
  - Add environment variable to Pod spec: `env: [{name: step.secrets[].envVar, valueFrom: {secretKeyRef: ...}}]`
  - Kubernetes handles actual injection (no manual reading of secret values)
- **File**: `pkg/controller/job_manager.go` (update)
- **Depends on**: T025

### T061: [US3] Implement secret value masking in logs
- Create `pkg/secrets/masker.go`
- Implement `MaskSecrets(logs []byte, secrets map[string]string) []byte`
- For each secret value, replace with `***REDACTED***` in logs
- Use regex to catch secret values even if partially printed
- Handle case-insensitive matching
- **File**: `pkg/secrets/masker.go`
- **Depends on**: T001

### T062: [US3] Integrate secret masking into log collection
- Update `pkg/controller/log_collector.go`:
  - Before uploading logs to S3, fetch all secret values from referenced Secrets
  - Apply masking using `MaskSecrets()`
  - Upload masked logs
  - Ensure in-memory buffer also stores masked logs
- Never persist unmasked secret values anywhere
- **File**: `pkg/controller/log_collector.go` (update)
- **Depends on**: T046, T061

### T063: [US3] [P] Implement secret validator
- Create `pkg/secrets/validator.go`
- Implement `ValidateSecretReferences(step *PipelineStep, namespace string) error`
- Check that all referenced Secrets exist in namespace
- Check that all referenced keys exist in Secret
- Return clear error messages for missing secrets/keys
- **File**: `pkg/secrets/validator.go`
- **Depends on**: T013

### T064: [US3] Integrate secret validation into PipelineConfig admission webhook
- Create admission webhook service (can be embedded in controller or separate)
- Implement ValidatingWebhookConfiguration for PipelineConfig
- On PipelineConfig create/update, validate all secret references
- Reject if any secret reference is invalid
- **File**: `pkg/webhook/admission.go`, `config/webhook/manifests.yaml`
- **Depends on**: T063

### T065: [US3] [P] Add RBAC permissions for reading Secrets
- Update `config/rbac/role.yaml`:
  - Add `secrets: [get, list]` permissions for controller ServiceAccount
  - Ensure controller can read Secrets in same namespace as PipelineConfig
- Do NOT grant create/update/delete on Secrets (read-only)
- **File**: `config/rbac/role.yaml` (update)
- **Depends on**: T022

### T066: [US3] [P] Implement CLI command to create secrets
- Create `pkg/cli/secret.go`
- Implement `c8s secret create <name> --from-literal=KEY=VALUE`
- Wrapper around `kubectl create secret`
- Add helpful messages about referencing secrets in pipelines
- **File**: `pkg/cli/secret.go`
- **Depends on**: T035

### T067: [US3] [P] Document secret management in quickstart
- Update `specs/001-build-a-continuous/quickstart.md`:
  - Add section on creating secrets
  - Add example pipeline with secret reference
  - Document masking behavior
  - Add troubleshooting for missing secrets
- **File**: `specs/001-build-a-continuous/quickstart.md` (update)
- **Depends on**: Quickstart already exists from planning phase

### T068: [US3] [P] Write unit tests for secret masking
- Create `tests/unit/masker_test.go`
- Test cases:
  - Single secret value masked
  - Multiple secret values masked
  - Partial secret values masked (e.g., "My password is: <secret>")
  - Secrets with special characters
  - Empty logs return empty
  - No secrets returns original logs
- **File**: `tests/unit/masker_test.go`
- **Depends on**: T061, T003

### T069: [US3] Write integration test for secret injection
- Create `tests/integration/secret_injection_test.go`
- Create Secret in test namespace
- Create PipelineConfig referencing secret
- Create PipelineRun
- Simulate Job execution and log collection
- Assert:
  - Job Pod has environment variable with secret value
  - Logs uploaded to S3 have masked values
  - PipelineRun status does not contain secret values
- **File**: `tests/integration/secret_injection_test.go`
- **Depends on**: T060, T062, T011

### T070: [US3] [P] Add secret reference examples to config/samples
- Create `config/samples/secret.yaml` with example Secret
- Update `config/samples/pipelineconfig.yaml` to reference secret
- Add comments explaining secret injection
- **File**: `config/samples/secret.yaml`, `config/samples/pipelineconfig.yaml` (update)
- **Depends on**: T010

### T071: [US3] Implement audit logging for secret access
- Add audit log entry when controller reads Secret for injection
- Log: timestamp, Secret name, PipelineRun name, user/SA
- Use structured logging (e.g., logr)
- DO NOT log secret values, only metadata
- **File**: `pkg/controller/audit.go`
- **Depends on**: T060

**✓ Checkpoint User Story 3**: Secret management complete. Secrets injected securely, values masked in all logs, audit trail maintained.

---

## Phase 6: User Story 4 (P3) - Resource Optimization

**Goal**: Implement resource quotas, admission control, and autoscaling integration

**User Story**: Ops team sets quotas per team, system enforces limits, autoscaling handles capacity

**Independent Test**: Set ResourceQuota, create PipelineRun exceeding quota, verify rejection. Create many PipelineRuns, verify autoscaling triggers.

**Prerequisites**: Phase 3 complete

**Tasks**:

### T072: [US4] Implement admission webhook for quota validation
- Create `pkg/webhook/quota_admission.go`
- Implement ValidatingWebhook for PipelineRun creation
- On PipelineRun create, sum all step resource requests
- Query namespace ResourceQuota
- Reject if PipelineRun would exceed quota
- Return clear error message: "Would exceed CPU quota: 120/100 cores requested"
- **File**: `pkg/webhook/quota_admission.go`
- **Depends on**: T014

### T073: [US4] Deploy admission webhook service
- Create `deploy/admission-webhook-deployment.yaml`
- Create TLS certificate for webhook (can use cert-manager or self-signed)
- Create ValidatingWebhookConfiguration manifest
- Configure webhook to intercept PipelineRun create operations
- **File**: `deploy/admission-webhook-deployment.yaml`, `config/webhook/validating-webhook.yaml`
- **Depends on**: T072

### T074: [US4] [P] Implement resource requirement aggregation
- Create `pkg/controller/resources.go`
- Implement `CalculateTotalResources(steps []PipelineStep) (cpu, memory string)`
- Sum all step CPU and memory requests
- Handle cases where resources not specified (use defaults)
- **File**: `pkg/controller/resources.go`
- **Depends on**: T013

### T075: [US4] [P] Add ResourceQuota examples to config/samples
- Create `config/samples/resourcequota.yaml`
- Example: team quota with 100 CPU, 200Gi memory, 50 pods
- Add documentation on setting up per-team namespaces
- **File**: `config/samples/resourcequota.yaml`
- **Depends on**: T010

### T076: [US4] [P] Document autoscaling integration
- Create `docs/autoscaling.md`
- Explain how Cluster Autoscaler watches pending Pods
- Document setting up Cluster Autoscaler
- Document node affinity/taints for CI workloads
- Document scaling behavior (scale up/down timing)
- **File**: `docs/autoscaling.md`
- **Depends on**: Documentation structure from T012

### T077: [US4] [P] Implement priority labels for PipelineRuns
- Update PipelineRun CRD to support label: `c8s.dev/priority: high|medium|low`
- Document that Kubernetes scheduler will prioritize high-priority Pods when resources are constrained
- Add examples of setting priority
- **File**: `pkg/apis/v1alpha1/pipelinerun_types.go` (update), docs
- **Depends on**: T014

### T078: [US4] [P] Add resource usage metrics to PipelineRun status
- Update `PipelineRunStatus` to include: `resourceUsage: {cpu, memory, duration}`
- Calculate from Job Pod metrics
- Use for reporting and capacity planning
- **File**: `pkg/apis/v1alpha1/pipelinerun_types.go` (update), `pkg/controller/status_updater.go` (update)
- **Depends on**: T027

### T079: [US4] [P] Write integration test for quota enforcement
- Create `tests/integration/quota_test.go`
- Create namespace with ResourceQuota (10 CPU)
- Create PipelineRun with total 5 CPU - should succeed
- Create PipelineRun with total 15 CPU - should be rejected by admission webhook
- Assert rejection error message is clear
- **File**: `tests/integration/quota_test.go`
- **Depends on**: T072, T011

### T080: [US4] [P] Implement graceful handling of insufficient cluster capacity
- When Job cannot be scheduled (pending due to insufficient resources):
  - Update PipelineRun status with condition: "InsufficientResources"
  - Add event to PipelineRun: "Waiting for cluster capacity"
  - Continue reconciling (don't fail permanently)
  - Once capacity available, Job will schedule automatically
- **File**: `pkg/controller/pipelinerun_controller.go` (update)
- **Depends on**: T026

### T081: [US4] [P] Add Prometheus metrics for resource usage
- Instrument controller with Prometheus metrics:
  - `c8s_pipelineruns_total{phase,namespace}`
  - `c8s_pipelineruns_duration_seconds{namespace}`
  - `c8s_pipeline_step_resource_usage_cpu_cores{step,namespace}`
  - `c8s_pipeline_step_resource_usage_memory_bytes{step,namespace}`
- Expose metrics endpoint on controller: `/metrics`
- **File**: `pkg/metrics/metrics.go`
- **Depends on**: Add `github.com/prometheus/client_golang` dependency

**✓ Checkpoint User Story 4**: Resource optimization complete. Quotas enforced, autoscaling integrated, capacity managed efficiently.

---

## Phase 7: User Story 5 (P3) - Parallel Execution

**Goal**: Enable parallel steps and matrix strategies for faster pipeline execution

**User Story**: Developer defines parallel steps or matrix, system executes simultaneously, aggregates results

**Independent Test**: Create pipeline with 3 parallel steps, verify all run simultaneously. Create matrix pipeline (2×2), verify 4 PipelineRuns created.

**Prerequisites**: Phase 3 complete

**Tasks**:

### T082: [US5] Enhance scheduler to identify parallel steps
- Update `pkg/scheduler/scheduler.go`:
  - Identify steps with no dependencies (layer 0) - all can run in parallel
  - Identify steps with same dependencies (same layer) - can run in parallel
  - Return schedule with layers, each layer runs in parallel
- **File**: `pkg/scheduler/scheduler.go` (update)
- **Depends on**: T024

### T083: [US5] Update Job creation to launch parallel Jobs simultaneously
- Update `pkg/controller/pipelinerun_controller.go`:
  - For each layer in schedule, create Jobs for all steps in parallel (don't wait)
  - Track which layer is currently executing
  - Move to next layer only when all Jobs in current layer complete
- **File**: `pkg/controller/pipelinerun_controller.go` (update)
- **Depends on**: T082

### T084: [US5] Implement matrix strategy expansion
- Create `pkg/scheduler/matrix.go`
- Implement `ExpandMatrix(matrix *MatrixStrategy) ([]map[string]string, error)`
- Generate all combinations of matrix dimensions
- Example: {os: [ubuntu, alpine], go: [1.21, 1.22]} → 4 combinations
- Handle `exclude` to filter out specific combinations
- **File**: `pkg/scheduler/matrix.go`
- **Depends on**: T013

### T085: [US5] Implement PipelineRun creation for matrix executions
- When PipelineConfig has matrix strategy, create multiple PipelineRuns:
  - One PipelineRun per matrix combination
  - Set `matrixIndex` field in PipelineRun.spec
  - Substitute matrix variables in step commands/images (e.g., `image: golang:${{matrix.go_version}}`)
  - Add label: `c8s.dev/matrix-parent: <original-trigger-id>`
- **File**: `pkg/webhook/github.go` (update) or new `pkg/controller/matrix_controller.go`
- **Depends on**: T084

### T086: [US5] Implement result aggregation for matrix executions
- Create `pkg/controller/aggregator.go`
- Implement `AggregateMatrixResults(matrixRuns []*PipelineRun) *AggregatedResult`
- Aggregate: total succeeded, total failed, duration, logs
- Store aggregated result in parent PipelineRun or separate CRD
- **File**: `pkg/controller/aggregator.go`
- **Depends on**: T085

### T087: [US5] [P] Update dashboard to show matrix executions
- Update `web/templates/runs.html`:
  - Show matrix runs grouped by parent
  - Show matrix combination parameters (os=ubuntu, go=1.21)
  - Show aggregated status (e.g., "3/4 succeeded")
  - Allow expanding to see individual matrix run logs
- **File**: `web/templates/runs.html` (update)
- **Depends on**: T056, T086

### T088: [US5] [P] Write integration test for parallel execution
- Create `tests/integration/parallel_test.go`
- Create PipelineConfig with 3 steps, all no dependencies (parallel)
- Create PipelineRun
- Assert:
  - All 3 Jobs created simultaneously (within same reconcile loop)
  - Jobs run concurrently (check timestamps)
  - PipelineRun completes when all Jobs finish
- **File**: `tests/integration/parallel_test.go`
- **Depends on**: T083, T011

### T089: [US5] [P] Write integration test for matrix execution
- Create `tests/integration/matrix_test.go`
- Create PipelineConfig with matrix: {os: [ubuntu, alpine], version: [1.21, 1.22]}
- Trigger pipeline (via webhook or manual)
- Assert:
  - 4 PipelineRuns created (2×2)
  - Each has correct matrixIndex
  - All run in parallel
  - Aggregated result shows overall status
- **File**: `tests/integration/matrix_test.go`
- **Depends on**: T085, T086, T011

**✓ Checkpoint User Story 5**: Parallel execution complete. Pipelines run faster with parallel steps and matrix strategies.

---

## Implementation Strategy

### MVP Scope (Phases 1-3)
**Recommended first delivery**: Complete Phases 1, 2, and 3 (Tasks T001-T044)
- Sets up project and foundational CRDs
- Delivers core CI functionality (User Story 1)
- Provides immediate value: automated testing on commits
- Estimated: ~42 tasks

### Incremental Delivery
After MVP, add user stories incrementally:
1. **Phase 4** (US2): Adds observability - critical for production use
2. **Phase 5** (US3): Adds secret management - required for integration tests
3. **Phase 6** (US4): Adds resource optimization - important at scale
4. **Phase 7** (US5): Adds parallel execution - performance optimization

### Parallel Development Opportunities
Within each phase, tasks marked **[P]** can be developed in parallel by different team members:

**Phase 1 Example** (Setup):
- Team Member A: T001, T002, T005, T007 (core Go setup)
- Team Member B: T004, T009, T011 (tooling and testing)
- Team Member C: T006, T010, T012 (documentation)

**Phase 3 Example** (US1 - Basic Pipeline):
- Team Member A: T023-T024 (scheduler)
- Team Member B: T025-T027 (controller core)
- Team Member C: T029-T032 (webhooks)
- Team Member D: T035-T037 (CLI)
- Team Member E: T042-T043 (unit tests)

---

## Dependencies Summary

```
Phase 1 (Setup)
  ↓
Phase 2 (Foundational) ← BLOCKING for all user stories
  ↓
Phase 3 (US1 - Basic Pipeline) ← MVP
  ↓
Phase 4 (US2 - Observability) ← Independent, can be done in parallel with Phase 5
  ↓
Phase 5 (US3 - Secret Management) ← Independent, can be done in parallel with Phase 4
  ↓
Phase 6 (US4 - Resource Optimization) ← Independent
  ↓
Phase 7 (US5 - Parallel Execution) ← Extends Phase 3 scheduler
```

**Critical Path**: T001 → T002 → T013-T016 → T017-T018 → T023-T027 → T044

---

## Testing Strategy

### Test Organization
- **Unit tests** (`tests/unit/`): Test individual components in isolation
- **Contract tests** (`tests/contract/`): Validate API contracts match OpenAPI spec
- **Integration tests** (`tests/integration/`): Test end-to-end flows with envtest

### Test-First Approach (Recommended)
For each user story:
1. Write integration test defining expected behavior (will fail initially)
2. Implement tasks to make test pass
3. Add unit tests for complex components
4. Run full test suite to ensure no regressions

### Test Execution
```bash
# Run all tests
make test

# Run specific test suite
go test ./tests/unit/...
go test ./tests/integration/...

# Run with coverage
go test -cover ./...
```

---

## Success Criteria Validation

After completing each user story phase, validate against spec.md success criteria:

**User Story 1 (P1)**:
- ✓ SC-001: Pipeline results within 10 minutes
- ✓ SC-012: Config changes take effect immediately
- ✓ SC-009: New pipeline defined and executed within 30 minutes

**User Story 2 (P2)**:
- ✓ SC-003: Logs available for real-time streaming + 30-day retention
- ✓ SC-005: Failed steps debuggable within 5 minutes
- ✓ SC-011: Complete audit trail maintained

**User Story 3 (P2)**:
- ✓ SC-004: Secrets never exposed in logs

**User Story 4 (P3)**:
- ✓ SC-002: 100 concurrent runs without degradation
- ✓ SC-006: Autoscale response within 2 minutes
- ✓ SC-007: Scale down within 10 minutes idle
- ✓ SC-008: 95% success rate excluding infra failures

**User Story 5 (P3)**:
- ✓ SC-010: Parallel execution reduces duration by 60%

---

## Next Steps

1. **Start with Phase 1**: Initialize project structure and tooling
2. **Complete Phase 2**: Implement CRDs and controller foundation
3. **Build MVP (Phase 3)**: User Story 1 - basic pipeline execution
4. **Test MVP end-to-end**: Run integration tests, deploy to test cluster
5. **Iterate on remaining user stories**: Add observability, secrets, optimization, parallelization

**Ready to begin implementation!** 🚀
