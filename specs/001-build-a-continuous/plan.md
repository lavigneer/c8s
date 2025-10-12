# Implementation Plan: Kubernetes-Native Continuous Integration System

**Branch**: `001-build-a-continuous` | **Date**: 2025-10-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-build-a-continuous/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a continuous integration system that leverages Kubernetes-native paradigms (Jobs, Pods, ConfigMaps, Secrets) to execute containerized pipeline workloads. The system automatically detects repository changes via webhooks, parses declarative pipeline configurations, schedules isolated container execution for each step, manages secrets securely, and provides observability through real-time logs and metrics. Core capabilities include sequential and parallel execution, artifact sharing between steps, resource quotas, autoscaling integration, and comprehensive audit trails.

## Technical Context

**Language/Version**: Go 1.25 (resolved: Kubernetes ecosystem standard, client-go library, enhanced stdlib)
**Primary Dependencies**:
  - `client-go` v0.28+ (Kubernetes client library)
  - `controller-runtime` v0.16+ (CRD controller framework)
  - `gopkg.in/yaml.v3` (YAML parsing for pipeline configs)
  - `net/http` stdlib (HTTP server and routing with Go 1.22+ enhanced ServeMux)
  - `html/template` stdlib (server-side HTML rendering for dashboard)
  - HTMX 2.0+ (~14KB JS library, CDN-hosted) + Tailwind CSS 3+ (optional dashboard, embedded in API server)
**Storage**:
  - Kubernetes CRDs backed by etcd (PipelineConfig, PipelineRun, RepositoryConnection state)
  - S3-compatible object storage (logs and artifacts: AWS S3, GCS, MinIO, Ceph)
  - In-memory circular buffers (real-time log streaming with <500ms latency)
**Testing**: Go standard `testing` package + `testify` v1.8+ + `envtest` (Kubernetes integration testing)
**Target Platform**: Kubernetes 1.24+ clusters (Linux containers only, no Windows support in v1)
**Project Type**: Backend service (controller + API server + webhook receiver) + CLI tool + optional dashboard UI
**Performance Goals**:
  - 100 concurrent pipeline runs without degradation (SC-002)
  - Pipeline results within 10 minutes for standard test suites (SC-001)
  - Log streaming with <500ms latency
  - Autoscale response within 2 minutes (SC-006)
**Constraints**:
  - Must use Kubernetes-native primitives (Jobs/Pods) for workload execution
  - Secret values never logged or exposed (SC-004)
  - 95% success rate excluding infrastructure failures (SC-008)
  - Real-time log access during execution (SC-003)
  - No external databases (use CRDs/etcd only)
**Scale/Scope**:
  - Support for hundreds of repositories per cluster
  - Thousands of pipeline runs per day
  - 30-day log retention (SC-003)
  - Multi-tenant with resource quotas per team

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Specification-First Development ✅

- ✅ Complete feature specification exists in `spec.md` with user scenarios, requirements, and success criteria
- ✅ 5 prioritized user stories (P1-P3) are defined and independently testable
- ✅ 30 functional requirements documented with clear acceptance criteria
- ✅ 12 measurable success criteria defined before implementation

**Status**: PASS - Following specification-first workflow

### II. User Story-Driven Architecture ✅

- ✅ User Story 1 (P1): Basic pipeline execution - standalone MVP delivering core CI value
- ✅ User Story 2 (P2): Observability - independent monitoring and debugging capability
- ✅ User Story 3 (P2): Secret management - independent security feature
- ✅ User Story 4 (P3): Resource optimization - independent scaling and cost management
- ✅ User Story 5 (P3): Parallel execution - independent performance optimization

Each story can be implemented, tested, and deployed independently while delivering incremental value.

**Status**: PASS - User stories are properly decomposed and prioritized

### III. Constitution Gates ✅

This is the initial Constitution Check for planning phase. Gates validated:

- ✅ Architectural approach leverages Kubernetes-native primitives (Jobs, Pods, CRDs) - simple and aligned with platform
- ✅ No premature abstractions - using standard Kubernetes patterns before custom solutions
- ✅ Complexity justifications documented in Complexity Tracking table when needed
- ✅ Technical decisions will be researched in Phase 0 with alternatives evaluated

**Status**: PASS - Pre-research gate cleared. Will re-evaluate after Phase 1 design.

### IV. Test Independence ✅

Test strategy per user story:

- ✅ P1 (Basic pipeline): Contract tests for pipeline config parsing, integration tests for Job creation, unit tests for state machine
- ✅ P2 (Observability): Contract tests for log streaming API, integration tests for metrics collection, unit tests for status aggregation
- ✅ P3 (Secret management): Contract tests for secret injection API, integration tests for masking behavior, unit tests for access control
- ✅ P4 (Resource optimization): Integration tests for autoscaling triggers, unit tests for quota enforcement
- ✅ P5 (Parallel execution): Contract tests for matrix config, integration tests for parallel Job creation, unit tests for result aggregation

Each test category can be written and executed independently per user story.

**Status**: PASS - Test independence achievable for all user stories

### V. Documentation as Artifact ✅

Design artifacts structure:

- ✅ `spec.md` - Feature specification (completed)
- ✅ `plan.md` - This implementation plan (in progress)
- ⏳ `research.md` - Phase 0 research findings (pending)
- ⏳ `data-model.md` - Entity definitions and relationships (pending Phase 1)
- ⏳ `contracts/` - API contracts for pipeline config and REST/GraphQL APIs (pending Phase 1)
- ⏳ `quickstart.md` - Developer onboarding guide (pending Phase 1)
- ⏳ `tasks.md` - Implementation tasks (pending `/speckit.tasks` command)

All artifacts will be version-controlled in `specs/001-build-a-continuous/` directory.

**Status**: PASS - Documentation workflow established

### Summary

**Overall Gate Status**: ✅ PASS - Proceed to Phase 0 Research

All constitution principles are satisfied. No violations requiring complexity justification. The architecture leverages Kubernetes-native paradigms, avoiding premature abstraction while enabling independent development of prioritized user stories.

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
cmd/
├── controller/           # Controller (operator) main entry point
│   └── main.go
├── api-server/          # REST API server main entry point
│   └── main.go
├── webhook/             # Webhook receiver service main entry point
│   └── main.go
└── c8s/                 # CLI tool main entry point
    └── main.go

pkg/
├── apis/                # CRD API definitions
│   └── v1alpha1/
│       ├── pipelineconfig_types.go
│       ├── pipelinerun_types.go
│       └── repositoryconnection_types.go
├── controller/          # Controller reconciliation logic
│   ├── pipelinerun_controller.go
│   ├── job_manager.go
│   └── status_updater.go
├── parser/              # Pipeline YAML parser and validator
│   ├── parser.go
│   └── validator.go
├── scheduler/           # Pipeline step scheduling and dependency resolution
│   ├── scheduler.go
│   └── dag.go
├── storage/             # Object storage client (S3)
│   ├── client.go
│   └── log_uploader.go
├── webhook/             # Webhook receiver logic
│   ├── github.go
│   ├── gitlab.go
│   └── bitbucket.go
├── api/                 # REST API handlers
│   ├── handlers/
│   │   ├── pipelineconfig.go
│   │   ├── pipelinerun.go
│   │   └── logs.go
│   └── middleware/
│       ├── auth.go
│       └── cors.go
├── cli/                 # CLI command implementations
│   ├── run.go
│   ├── logs.go
│   └── validate.go
└── secrets/             # Secret injection and log masking
    ├── injector.go
    └── masker.go

tests/
├── contract/            # API contract tests (OpenAPI validation)
│   ├── pipelineconfig_test.go
│   ├── pipelinerun_test.go
│   └── logs_api_test.go
├── integration/         # End-to-end tests with envtest
│   ├── pipeline_execution_test.go
│   ├── secret_injection_test.go
│   └── webhook_trigger_test.go
└── unit/               # Component unit tests
    ├── parser_test.go
    ├── scheduler_test.go
    └── dag_test.go

config/
├── crd/                # CRD YAML definitions
│   ├── pipelineconfig.yaml
│   ├── pipelinerun.yaml
│   └── repositoryconnection.yaml
├── rbac/               # RBAC manifests
│   └── role.yaml
└── manager/            # Controller deployment manifests
    └── deployment.yaml

web/                    # Optional HTMX dashboard (embedded in API server)
├── templates/          # Go html/template files
│   ├── layout.html
│   ├── pipelines.html
│   ├── runs.html
│   └── logs.html
└── static/            # Static assets
    ├── htmx.min.js    # HTMX library (~14KB)
    └── styles.css     # Tailwind-generated CSS

deploy/                 # Deployment resources
├── install.yaml        # Combined installation manifest
├── crds.yaml           # CRDs only
└── minio.yaml          # MinIO for local development
```

**Structure Decision**:

Selected **Option 1: Single project** structure adapted for Kubernetes operator pattern with Go. The project follows standard Go Kubernetes operator layout:

- **`cmd/`**: Multiple entry points for controller, API server, webhook service, and CLI
- **`pkg/`**: Shared libraries organized by domain (APIs, controller, parser, storage, etc.)
- **`tests/`**: Three-tier test structure (contract, integration, unit) per constitution
- **`config/`**: Kubernetes manifests for CRDs and deployments
- **`web/`**: HTMX dashboard templates and static assets (embedded in API server binary)
- **`deploy/`**: Installation artifacts for end users

This structure enables:
- Independent deployment of controller, API, and webhook services
- Shared code reuse through `pkg/` libraries
- Clear separation between CRD definitions (`pkg/apis`) and business logic
- Standard Go testing patterns with dedicated test directories
- Optional dashboard embedded in API server (toggled via `--enable-dashboard` flag)
- Zero JavaScript build tooling required (HTMX served from CDN or embedded)

## Post-Design Constitution Check

*Re-evaluation after Phase 1 design completion*

### I. Specification-First Development ✅

- ✅ All design artifacts generated from specification (research.md, data-model.md, contracts/, quickstart.md)
- ✅ Technical decisions documented with alternatives considered in research.md
- ✅ No implementation started before design completion

**Status**: PASS - Design phase completed per specification-first workflow

### II. User Story-Driven Architecture ✅

- ✅ Architecture enables independent implementation of each user story:
  - P1 (Basic pipeline): Controller + Job management + basic CRDs
  - P2 (Observability): API server + log streaming (independent service)
  - P2 (Secret management): Secret injector + log masker (independent package)
  - P3 (Resource optimization): Quota admission webhook (optional component)
  - P3 (Parallel execution): Scheduler enhancement (extends P1)
- ✅ Each component can be tested independently
- ✅ Incremental delivery possible (P1 MVP, then add P2/P3)

**Status**: PASS - Architecture supports independent user story implementation

### III. Constitution Gates ✅

**Architectural Simplicity Assessment**:

✅ **Leverages Kubernetes primitives**: Using native CRDs, Jobs, Pods, Secrets instead of custom abstractions
✅ **No premature abstractions**: Direct use of client-go and controller-runtime (standard libraries)
✅ **Minimal external dependencies**: No database, no message queue, no service mesh, no third-party HTTP routers
✅ **Standard library focus**: Go 1.25 stdlib for HTTP server/routing (net/http with enhanced ServeMux), WebSocket support
✅ **Standard patterns**: Standard Go K8s operator layout, REST API with stdlib only, standard testing

**Complexity Introduced** (justified):

1. **Three separate services** (controller, API server, webhook):
   - **Why needed**: Separation of concerns - controller watches CRDs, API serves REST, webhook receives GitHub events
   - **Simpler alternative rejected**: Single binary with all services → would violate K8s best practices for controllers (should not expose HTTP endpoints) and makes horizontal scaling harder

2. **In-memory log buffer + object storage**:
   - **Why needed**: Real-time streaming (<500ms latency) + persistent storage for 30 days
   - **Simpler alternative rejected**: Object storage only → cannot meet <500ms streaming requirement (S3 has 100-200ms+ latency)

3. **Custom admission webhook**:
   - **Why needed**: Enforce resource quotas at PipelineRun creation time (before Jobs created)
   - **Simpler alternative rejected**: Check quotas in controller → racy, Jobs might already be created

**Status**: PASS - All complexity justified with documented alternatives

### IV. Test Independence ✅

- ✅ Contract tests validate CRD schemas and API contracts (can run without cluster)
- ✅ Integration tests use envtest (isolated fake K8s API server per test)
- ✅ Unit tests for parser, scheduler, masker (no K8s dependencies)
- ✅ Each user story has dedicated test coverage (test/integration/*_test.go)

**Status**: PASS - Three-tier test structure enables independent execution

### V. Documentation as Artifact ✅

All design artifacts completed:

- ✅ `spec.md` - Feature specification
- ✅ `plan.md` - This implementation plan
- ✅ `research.md` - 12 technical decisions with rationales
- ✅ `data-model.md` - Complete CRD schemas and relationships
- ✅ `contracts/openapi.yaml` - REST API specification
- ✅ `contracts/pipeline-config-schema.json` - Pipeline YAML validation schema
- ✅ `quickstart.md` - Developer onboarding guide

**Status**: PASS - All Phase 1 artifacts complete and synchronized

### Summary

**Overall Gate Status**: ✅ PASS - Proceed to `/speckit.tasks` for task generation

- Architecture maintains simplicity by leveraging K8s primitives
- Three justified complexity points documented (3 services, dual storage, admission webhook)
- All complexity has clear rationale with rejected alternatives
- Design enables independent user story implementation
- Complete documentation artifacts ready for task generation

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Three separate services (controller, API, webhook) | Separation of concerns: controller shouldn't expose HTTP, webhook needs independent scaling | Single binary would violate K8s controller best practices and prevent independent horizontal scaling |
| Dual storage (in-memory + S3) for logs | Real-time streaming (<500ms) + 30-day persistence | S3-only has 100-200ms+ latency, cannot meet real-time streaming requirement |
| Custom admission webhook | Atomic quota enforcement at PipelineRun creation before Jobs created | Controller-based quota checking is racy - Jobs might already be created |
