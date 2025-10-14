# Feature Specification: Local Test Environment Setup

**Feature Branch**: `002-i-want-to`
**Created**: 2025-10-13
**Status**: Draft
**Input**: User description: "I want to set up a way to launch a fully local test environment for this project. I'd like to spin up a k8s cluster locally and run a full test of a pipeline"

## User Scenarios & Testing

### User Story 1 - Local Cluster Creation (Priority: P1)

As a developer, I need to create a local Kubernetes cluster on my machine so that I can test the C8S pipeline operator without requiring external infrastructure or cloud costs.

**Why this priority**: This is the foundation for all local testing. Without a running cluster, no pipeline testing can occur. This delivers immediate value by enabling developers to work offline and iterate quickly.

**Independent Test**: Can be fully tested by running a single command that creates a cluster, verifying the cluster is accessible via kubectl, and confirming the cluster can run basic workloads (e.g., a simple pod deployment).

**Acceptance Scenarios**:

1. **Given** I have no local cluster running, **When** I execute the cluster creation command, **Then** a Kubernetes cluster is created and accessible via kubectl
2. **Given** a local cluster is created, **When** I run kubectl get nodes, **Then** I see at least one node in Ready status
3. **Given** a local cluster exists, **When** I stop and restart the cluster, **Then** the cluster state persists and workloads resume
4. **Given** I finish testing, **When** I execute the cluster deletion command, **Then** all cluster resources are cleaned up from my local machine

---

### User Story 2 - Pipeline Operator Deployment (Priority: P2)

As a developer, I need to deploy the C8S pipeline operator to my local cluster so that I can test the controller functionality with realistic Kubernetes resources.

**Why this priority**: Once a cluster exists, deploying the operator is essential to test the actual pipeline functionality. This enables testing of CRDs, controllers, and reconciliation logic.

**Independent Test**: Can be fully tested by deploying the operator to a local cluster, verifying all CRDs are registered, confirming the operator pod is running, and checking logs for successful startup.

**Acceptance Scenarios**:

1. **Given** a local cluster is running, **When** I deploy the operator using the provided deployment method, **Then** all CRDs are registered in the cluster
2. **Given** the operator is deployed, **When** I check the operator pod status, **Then** the pod is in Running state with no crash loops
3. **Given** the operator is running, **When** I view the operator logs, **Then** I see successful initialization messages
4. **Given** the operator is deployed, **When** I apply a sample PipelineConfig, **Then** the operator reconciles and creates corresponding pipeline resources

---

### User Story 3 - End-to-End Pipeline Test (Priority: P3)

As a developer, I need to run a complete pipeline test on my local cluster so that I can verify the full lifecycle from PipelineConfig creation to pipeline execution and result reporting.

**Why this priority**: This provides comprehensive validation of the entire system working together. While critical for quality assurance, it builds on the previous two stories and represents the full integration test.

**Independent Test**: Can be fully tested by creating a sample PipelineConfig with a simple build step, triggering execution, monitoring the pipeline job, and verifying successful completion with expected artifacts or logs.

**Acceptance Scenarios**:

1. **Given** the operator is running in my local cluster, **When** I create a PipelineConfig for a simple repository, **Then** the pipeline job is created and starts executing
2. **Given** a pipeline is executing, **When** I monitor the job status, **Then** I can see real-time progress and logs from pipeline steps
3. **Given** a pipeline completes successfully, **When** I check the PipelineConfig status, **Then** the status reflects success with execution time and step results
4. **Given** a pipeline encounters an error, **When** I inspect the failure, **Then** I see clear error messages and can identify which step failed
5. **Given** I want to rerun a pipeline, **When** I trigger a new execution, **Then** a new job is created without affecting previous run history

---

### User Story 4 - Test Environment Teardown (Priority: P4)

As a developer, I need to easily tear down and recreate my local test environment so that I can ensure clean state between test runs or free up system resources when not testing.

**Why this priority**: While important for maintainability and resource management, this is lower priority than the core testing functionality. Nice to have for convenience but not blocking for initial testing.

**Independent Test**: Can be fully tested by running a teardown command, verifying all cluster resources are removed, confirming no orphaned processes or containers remain, and successfully recreating the environment from scratch.

**Acceptance Scenarios**:

1. **Given** a local test environment is running, **When** I execute the teardown command, **Then** the cluster is deleted and all resources are cleaned up
2. **Given** the environment has been torn down, **When** I check for running processes, **Then** no cluster-related processes remain
3. **Given** I want to start fresh, **When** I recreate the environment after teardown, **Then** the new environment is identical to a first-time setup
4. **Given** I have pipeline runs in progress, **When** I attempt teardown, **Then** I receive a warning about active workloads and can choose to force cleanup or wait

---

### Edge Cases

- What happens when the local machine runs out of disk space during pipeline execution?
- How does the system handle incomplete teardown (e.g., process killed mid-cleanup)?
- What occurs if the local cluster fails to start due to port conflicts with other services?
- How does the system respond to creating a cluster when one already exists?
- What happens if the operator deployment fails due to insufficient cluster resources?
- How does the test environment handle Docker daemon being unavailable or restarted?
- What occurs if a pipeline test runs for longer than expected system timeout limits?

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a command to create a local Kubernetes cluster suitable for C8S operator testing
- **FR-002**: System MUST verify cluster health and readiness before allowing operator deployment
- **FR-003**: System MUST provide a command to deploy the C8S operator and its dependencies to the local cluster
- **FR-004**: System MUST register all C8S CRDs (PipelineConfig, PipelineRun, etc.) during operator deployment
- **FR-005**: System MUST provide sample PipelineConfig manifests for common test scenarios
- **FR-006**: System MUST support running end-to-end pipeline tests that exercise the full reconciliation loop
- **FR-007**: System MUST capture and display logs from pipeline executions for debugging
- **FR-008**: System MUST provide a command to completely tear down the local test environment
- **FR-009**: System MUST validate that required dependencies (Docker, kubectl, etc.) are installed before cluster creation
- **FR-010**: System MUST support recreating the environment multiple times on the same machine
- **FR-011**: System MUST provide documentation explaining how to use the local test environment
- **FR-012**: System MUST isolate the local test environment from other Kubernetes contexts to prevent accidental production deployments

### Key Entities

- **Local Cluster**: A Kubernetes cluster running on the developer's machine using a lightweight distribution (e.g., kind, minikube, k3d), configured with necessary resources (CPU, memory) to run the C8S operator and test pipelines
- **Test Environment**: The complete setup including the local cluster, deployed operator, registered CRDs, sample configurations, and any supporting services needed for pipeline execution
- **Sample Pipeline**: Pre-configured PipelineConfig examples that demonstrate common use cases (simple build, multi-step pipeline, matrix builds, secret injection) for validation testing

## Success Criteria

### Measurable Outcomes

- **SC-001**: Developers can create a local test cluster in under 3 minutes on standard development hardware
- **SC-002**: Developers can deploy the operator and run a complete pipeline test in under 10 minutes from initial setup
- **SC-003**: The local test environment successfully executes 100% of the sample pipeline configurations without cluster-related failures
- **SC-004**: Developers can iterate on code changes, redeploy the operator, and rerun tests in under 2 minutes
- **SC-005**: Environment teardown completes in under 1 minute and leaves no orphaned resources
- **SC-006**: 90% of developers can set up and run their first pipeline test without external assistance after reading provided documentation
- **SC-007**: The local test environment uses less than 4GB RAM and 10GB disk space under normal testing load

## Assumptions

- Developers have Docker installed and running on their local machines (standard requirement for Kubernetes local development)
- Developers have kubectl installed (standard Kubernetes tooling)
- Local machines have sufficient resources (minimum 8GB RAM, 20GB free disk space) for Kubernetes development
- The chosen local cluster solution (kind/minikube/k3d) supports the Kubernetes version required by C8S operator
- Network connectivity is available for pulling container images during initial setup
- Developers have basic familiarity with Kubernetes concepts (pods, deployments, CRDs)
- The local test environment will use a single-node cluster by default (sufficient for functional testing)
- Test pipelines will use small container images and short-running tasks to minimize resource usage

## Scope

### In Scope

- Local Kubernetes cluster creation and lifecycle management
- Deployment of C8S operator to local cluster
- Sample PipelineConfig manifests for testing
- End-to-end pipeline execution validation
- Log capture and debugging support
- Complete environment teardown and cleanup
- Documentation for local testing workflows
- Validation of required dependencies

### Out of Scope

- Multi-node cluster configurations (can use single-node for testing)
- Performance benchmarking or load testing (focus is functional validation)
- Integration with external Git repositories (can use local repositories or mocks)
- CI/CD integration for automated local testing
- Graphical user interfaces for cluster management
- Production-grade monitoring or observability tooling
- Cross-platform compatibility beyond macOS and Linux (Windows support can be added later if needed)
- Automated test generation or test case discovery

## Dependencies

- Docker or compatible container runtime
- kubectl command-line tool
- Local Kubernetes distribution (kind, minikube, or k3d) - choice to be determined during implementation
- C8S operator code and build artifacts
- Container registry access for pulling operator and test images

## Risks & Mitigations

- **Risk**: Different local cluster tools (kind vs minikube vs k3d) have varying capabilities and quirks
  - **Mitigation**: Select one tool as the officially supported option based on stability, performance, and ease of use; document the choice clearly

- **Risk**: Local machine resource constraints may prevent successful pipeline execution
  - **Mitigation**: Document minimum resource requirements; provide lightweight sample pipelines; include resource check during setup

- **Risk**: Docker daemon issues or restarts during testing may break cluster state
  - **Mitigation**: Include health checks and recovery procedures in documentation; implement validation before operations

- **Risk**: Port conflicts with other services may prevent cluster creation
  - **Mitigation**: Use configurable port ranges; detect and report conflicts with clear error messages

- **Risk**: Developers may accidentally deploy to wrong Kubernetes context
  - **Mitigation**: Implement context validation; use distinctive naming for local test clusters; provide clear warnings

## Open Questions

None - all reasonable defaults have been assumed based on standard Kubernetes local development practices.
