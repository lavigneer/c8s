# Implementation Plan: Local Test Environment Setup

**Branch**: `002-i-want-to` | **Date**: 2025-10-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-i-want-to/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Provide developers with a streamlined way to create, configure, and test the C8S pipeline operator in a local Kubernetes environment. This enables rapid iteration without cloud infrastructure, reducing feedback loops from minutes to seconds. The solution will support cluster creation, operator deployment, sample pipeline execution, and complete environment teardown.

## Technical Context

**Language/Version**: Go 1.25.0
**Primary Dependencies**: Kubernetes client-go v0.28.15, controller-runtime v0.16.6, k3d v5.8.x (chosen after research)
**Storage**: N/A (cluster state managed by k3d/Docker)
**Testing**: Go testing framework, integration tests with envtest
**Target Platform**: macOS and Linux developer workstations (Docker-based local clusters)
**Project Type**: Single project with CLI tooling additions
**Performance Goals**: Cluster creation <3min (achieves 10-30s), operator deployment + test run <10min, iteration cycle <2min
**Constraints**: <4GB RAM usage (<512MB achieved with k3d), <10GB disk space (~5-8GB with k3d), must work with existing C8S operator codebase
**Scale/Scope**: Single-node local clusters (1 server + 2 agents), support for 3-5 sample pipeline configurations, CLI commands for lifecycle management

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Specification-First Development ✅
- **Status**: PASS
- **Evidence**: Complete specification exists at `spec.md` with user stories, requirements, and success criteria before implementation planning

### User Story-Driven Architecture ✅
- **Status**: PASS
- **Evidence**: Four prioritized user stories (P1-P4) defined, each independently testable:
  - P1: Local Cluster Creation (foundation)
  - P2: Pipeline Operator Deployment (enables testing)
  - P3: End-to-End Pipeline Test (integration validation)
  - P4: Test Environment Teardown (cleanup)

### Constitution Gates ✅
- **Status**: PASS
- **Evidence**: This Constitution Check performed before Phase 0; will re-evaluate after Phase 1 design

### Test Independence ✅
- **Status**: PASS
- **Evidence**: Each user story includes independent test criteria; contract tests will validate CLI commands, integration tests will validate cluster operations

### Documentation as Artifact ✅
- **Status**: PASS
- **Evidence**: Following structured workflow: spec.md → plan.md (this file) → research.md → data-model.md → contracts/ → tasks.md

### Architectural Simplicity ✅
- **Status**: PASS
- **Post-Design Assessment**: Design leverages existing tools (k3d, kubectl) and follows standard CLI patterns. New code is isolated in `pkg/localenv/` and `cmd/c8s/commands/dev/` packages. No premature abstraction detected. Complexity is justified by significant developer productivity gains (enabling rapid local testing without cloud infrastructure).

## Project Structure

### Documentation (this feature)

```
specs/002-i-want-to/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (local K8s tool evaluation)
├── data-model.md        # Phase 1 output (cluster/environment config schema)
├── quickstart.md        # Phase 1 output (getting started guide)
├── contracts/           # Phase 1 output (CLI command contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/
├── c8s/                 # Existing CLI (will extend with new commands)
│   ├── main.go
│   └── commands/
│       ├── dev/         # NEW: Development environment commands
│       │   ├── cluster.go      # cluster create/delete/status commands
│       │   ├── deploy.go       # operator deployment commands
│       │   └── test.go         # test execution commands
│       └── ... (existing commands)

pkg/
├── localenv/            # NEW: Local environment management
│   ├── cluster/         # Cluster lifecycle (kind/minikube/k3d wrapper)
│   ├── deploy/          # Operator deployment logic
│   ├── samples/         # Sample PipelineConfig generators
│   └── health/          # Health checks and validation
├── apis/                # Existing CRDs
├── controllers/         # Existing controllers
└── ... (existing packages)

tests/
├── contract/
│   └── dev_commands_test.go    # NEW: CLI command contract tests
├── integration/
│   └── localenv_test.go        # NEW: Local environment integration tests
└── unit/
    └── localenv/               # NEW: Unit tests for local env package

config/
└── samples/             # NEW: Sample PipelineConfigs for local testing
    ├── simple-build.yaml
    ├── multi-step.yaml
    └── matrix-build.yaml

docs/
└── local-testing.md     # NEW: Local testing guide (links to quickstart.md)
```

**Structure Decision**: Single project with CLI extension. The existing `cmd/c8s` CLI will be enhanced with a new `dev` command group for local environment management. Core logic will reside in `pkg/localenv/`, following the existing package organization pattern. This maintains architectural consistency while adding developer tooling capabilities.

## Complexity Tracking

*No violations identified. All gates pass.*
