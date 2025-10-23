# Feature Specification: Local Kubernetes Development Tooling

**Feature Branch**: `003-implement-tilt-or`
**Created**: 2025-10-22
**Status**: Draft
**Input**: User description: "Implement tilt or something similar for local k8s development. This should improve iteration time and make deploying everything simpler"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Starts Working on C8S Components (Priority: P1)

A developer clones the C8S repository and wants to begin working on controller, API server, or webhook service code. They need to quickly spin up a local development environment without manually running multiple commands or managing complex local cluster setup.

**Why this priority**: This is the foundation for all local development work. Without a smooth initial setup, developers face friction before they can even begin contributing, leading to longer onboarding times and reduced productivity.

**Independent Test**: A new developer can complete initial environment setup with a single command and begin iterating on component code within minutes.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the repository, **When** developer runs the local dev tool setup command, **Then** a fully functional local Kubernetes cluster with C8S installed is created and ready for development
2. **Given** a running development environment, **When** developer modifies source code for any component, **Then** the changes are automatically detected and the affected component is rebuilt and redeployed
3. **Given** a development environment, **When** developer needs to inspect logs or debug issues, **Then** they can easily access container logs and component status through a unified interface

---

### User Story 2 - Developer Tests Pipeline Execution Locally (Priority: P1)

A developer writes or modifies pipeline definition files and wants to validate them against a real Kubernetes cluster without waiting for CI/CD or deploying to a staging environment. They need rapid feedback on whether their pipeline definitions work correctly.

**Why this priority**: Pipeline definition testing is a core workflow for contributors. Slow feedback loops delay validation and reduce iteration speed, which is exactly what this feature aims to solve.

**Independent Test**: A developer can apply a pipeline definition to their local cluster, observe execution, and verify results within seconds of making changes.

**Acceptance Scenarios**:

1. **Given** a modified pipeline YAML file, **When** developer applies it to the local cluster, **Then** the pipeline executes and produces observable results (logs, status updates) within seconds
2. **Given** a pipeline failure, **When** developer views the results, **Then** clear error messages and logs are available to diagnose the issue
3. **Given** multiple pipeline definitions, **When** developer applies them to the local cluster, **Then** they execute independently without interference

---

### User Story 3 - Developer Manages Sample Deployments (Priority: P2)

A developer needs to easily deploy sample pipelines and CRD instances to test the complete C8S system locally, including the API server, webhook service, and controller working together.

**Why this priority**: This enables testing integration scenarios and validating that all components work correctly together, but is secondary to core iteration on individual components.

**Independent Test**: Sample deployments can be created, updated, and removed through the development tooling without manual kubectl commands.

**Acceptance Scenarios**:

1. **Given** sample pipeline definitions, **When** developer deploys them, **Then** all samples are created and ready for testing
2. **Given** deployed samples, **When** developer cleans up, **Then** all sample resources are removed cleanly
3. **Given** multiple sample scenarios, **When** developer switches between them, **Then** the appropriate sample set is deployed without manual cleanup

---

### User Story 4 - Developer Observes Multi-Component Interactions (Priority: P2)

A developer needs insight into how different C8S components interact (controller, API server, webhook, CLI) in a real execution environment. They need to see logs and metrics across all components to understand behavior.

**Why this priority**: Observability supports debugging complex interactions between components, but core iteration can happen without this feature. It becomes valuable as features grow more complex.

**Independent Test**: A dashboard or log aggregation view shows all component activity in one place, allowing developers to trace request flows through the system.

**Acceptance Scenarios**:

1. **Given** a running development environment, **When** developer views a dashboard or log interface, **Then** logs from all C8S components (controller, API server, webhook) are visible and searchable
2. **Given** a pipeline execution, **When** developer traces the execution, **Then** they can see how the request flows through webhook → API server → controller → Jobs
3. **Given** multiple pipeline runs, **When** developer filters logs, **Then** they can isolate and analyze specific runs or components

---

### User Story 5 - Developer Manages Lifecycle of Local Cluster (Priority: P3)

A developer needs simple commands to create, update, and destroy local development clusters without memorizing kubectl or cluster management tool syntax.

**Why this priority**: This is a convenience feature that improves developer experience but isn't strictly necessary for iteration work. The existing `c8s dev cluster` commands partially address this.

**Independent Test**: A developer can manage cluster lifecycle through intuitive commands with clear feedback on cluster state.

**Acceptance Scenarios**:

1. **Given** no local cluster exists, **When** developer creates one, **Then** a cluster is created with all required components installed automatically
2. **Given** a local cluster, **When** developer views its status, **Then** they see clear information about cluster health and deployed components
3. **Given** a local cluster, **When** developer tears it down, **Then** all resources are cleaned up

---

### Edge Cases

- What happens if a developer's code has a compilation error? (Error should be displayed clearly with rebuild attempted when code is fixed)
- How does the system handle updates to manifests or CRD definitions? (Should re-apply manifests and restart affected components)
- What happens if a developer kills the dev tooling process? (Local cluster should remain functional, dev tooling can reconnect)
- How are resource constraints handled if developer's machine has limited CPU/memory? (Should provide guidance and allow configuration of resource limits)
- What happens when switching between different branches with different CRD definitions? (Should update CRDs and restart controller with new definitions)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a single command to initialize a complete local Kubernetes development environment
- **FR-002**: System MUST automatically detect source code changes in the project and rebuild affected components without manual intervention
- **FR-003**: System MUST re-deploy or restart components when code changes are detected
- **FR-004**: System MUST re-apply CRD and RBAC manifests when they are modified
- **FR-005**: System MUST provide unified log viewing across all C8S components (controller, API server, webhook, CLI)
- **FR-006**: System MUST support rapid iteration on pipeline definitions with immediate feedback
- **FR-007**: System MUST validate pipeline definitions before deployment using both syntax validation (YAML structure) and semantic validation against the PipelineConfig CRD schema (field types, constraints, required fields)
- **FR-008**: System MUST manage a local Kubernetes cluster lifecycle (create, start, stop, delete)
- **FR-009**: System MUST provide clear, actionable error messages when builds fail or deployments encounter issues
- **FR-010**: System MUST support deployment and cleanup of sample pipeline definitions and configurations
- **FR-011**: System MUST be idempotent - running setup multiple times should produce the same result without errors
- **FR-012**: System MUST provide a way to view component status and logs in real-time

### Key Entities

- **LocalCluster**: Represents a local Kubernetes cluster instance (name, status, created timestamp, component states)
- **BuildArtifact**: Represents built binaries and container images (component name, version, build timestamp, image digest)
- **DeploymentState**: Represents the current state of deployed components (component name, image hash, manifest version, last update timestamp)
- **ComponentLog**: Time-series log entries from C8S components with component name, timestamp, and log content

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can complete full local environment setup in under 5 minutes on a machine meeting stated hardware requirements
- **SC-002**: Code changes are detected and redeployed within 30 seconds of file save (excluding build time for compilation)
- **SC-003**: Build failures are reported to the developer within 10 seconds of code save
- **SC-004**: A developer can test a pipeline definition change from code edit to execution result within 2 minutes total
- **SC-005**: All C8S components and logs are accessible from a single interface (CLI tool, dashboard, or both) without requiring multiple terminal windows or manual log aggregation
- **SC-006**: 95% of local development sessions should run without crashes or manual intervention beyond code editing
- **SC-007**: New developers report reduced time-to-first-contribution by at least 50% compared to manual setup
- **SC-008**: Development environment remains stable over extended sessions (4+ hours) without memory leaks or performance degradation

## Clarifications

### Session 2025-10-22

- Q: Should validation include both syntax checking AND semantic validation against the PipelineConfig CRD schema? → A: Yes, validate both syntax AND semantics against CRDs for comprehensive error detection during development

## Assumptions

- The local development environment uses a lightweight local Kubernetes distribution (k3d or similar) as already defined in project documentation
- Developers have Docker installed and configured (stated requirement in README)
- The C8S codebase structure (cmd/, pkg/, config/, deploy/) remains relatively stable
- Developers are comfortable with command-line interfaces
- The primary use case is development and testing, not production-like performance testing at scale
