# Technical Research: Kubernetes-Native CI System

**Feature**: 001-build-a-continuous
**Date**: 2025-10-12
**Status**: Complete

## Overview

This document resolves all "NEEDS CLARIFICATION" items from the Technical Context and establishes the technical foundation for implementation. Research focuses on selecting technologies and patterns that align with Kubernetes-native paradigms while maintaining simplicity and avoiding premature abstraction.

---

## Decision 1: Primary Implementation Language

**Decision**: Go 1.25

**Rationale**:
- **Kubernetes ecosystem standard**: Kubernetes itself, kubectl, Helm, and most K8s operators are written in Go
- **First-class client library**: `client-go` is the canonical Kubernetes client with comprehensive API coverage
- **Strong concurrency primitives**: Goroutines and channels are ideal for managing concurrent pipeline executions and log streaming
- **Static binaries**: Produces single-binary CLI tools with no runtime dependencies, simplifying deployment
- **Performance**: Compiled language with excellent performance characteristics for API services and controllers
- **Community alignment**: Extensive resources, patterns, and examples for building Kubernetes-native applications in Go
- **Go 1.25 improvements**: Enhanced standard library features, improved routing patterns in `net/http`, better performance

**Alternatives Considered**:
- **Rust**: Excellent performance and safety but smaller Kubernetes ecosystem, steeper learning curve, and less mature client libraries (kube-rs)
- **Python**: Easier learning curve but slower performance, requires runtime dependencies, and the `kubernetes` client library is less mature than client-go
- **TypeScript/Node.js**: Good for web UIs but poor fit for Kubernetes controllers and system services; JavaScript K8s clients lack maturity

**Impact**: All core components (controller, API server, CLI) will be written in Go 1.25. UI dashboard (if built) may use different technology.

---

## Decision 2: API Framework

**Decision**: REST API using standard `net/http` with `http.ServeMux` (Go 1.22+ enhanced routing)

**Rationale**:
- **Simplicity**: Standard library HTTP server is production-ready and well-understood
- **Zero external dependencies**: Go 1.22+ added enhanced routing patterns to `http.ServeMux` (method-specific routes, path parameters)
- **Kubernetes alignment**: K8s API server uses HTTP/REST patterns; following same conventions
- **Minimal abstraction**: Using stdlib avoids unnecessary third-party dependencies
- **Flexibility**: Easy to add gRPC later if needed for internal service-to-service communication
- **Client generation**: Can generate OpenAPI specs for client SDKs
- **Go 1.25 enhancements**: Further improvements to HTTP routing and request handling

**Alternatives Considered**:
- **gorilla/mux**: Previously standard choice but Go 1.22+ stdlib now provides equivalent routing capabilities
- **GraphQL**: Adds complexity for minimal benefit; REST is sufficient for CRUD operations on pipelines/runs
- **gRPC**: Better for service mesh scenarios but adds complexity; REST is more accessible for external clients
- **Heavy frameworks (Gin, Echo, Fiber)**: Unnecessary abstractions that hide standard library patterns

**Impact**: API endpoints follow RESTful conventions using only stdlib. OpenAPI 3.0 specification will be generated for contracts. No external HTTP routing dependencies needed.

---

## Decision 3: Dashboard UI

**Decision**: Optional web dashboard using HTMX + Tailwind CSS with server-side Go templates (separate from core system)

**Rationale**:
- **Separation of concerns**: UI is P2 feature (User Story 2) and should not block P1 core functionality
- **HTMX simplicity**: Hypermedia-driven approach with zero build step, no JavaScript framework needed
- **Server-side rendering**: Go html/template stdlib generates HTML, reducing client-side complexity
- **Minimal JavaScript**: HTMX is a single ~14KB library, drastically smaller than SPA frameworks
- **Tailwind simplicity**: Utility-first CSS reduces custom styling complexity
- **No build pipeline**: No npm, webpack, or bundler - just serve static HTMX library and Go-generated HTML
- **Progressive enhancement**: Works without JavaScript, degrades gracefully
- **Independence**: Dashboard is a simple HTTP handler in API server, can be disabled via flag

**Alternatives Considered**:
- **Svelte/React/Vue**: Require build pipelines (npm, bundlers), more complex state management, heavier bundles
- **Pure server-side HTML**: Works but lacks interactivity; HTMX adds AJAX/WebSocket without complexity
- **No dashboard**: CLI-only approach is viable but User Story 2 explicitly mentions "dashboard or CLI tool"

**Impact**: Dashboard is served directly from Go API server using html/template + HTMX. No separate build/deployment needed. Can be toggled on/off with `--enable-dashboard` flag. Zero JavaScript build tooling required.

---

## Decision 4: State Storage for Pipeline Metadata

**Decision**: Kubernetes Custom Resource Definitions (CRDs) backed by etcd

**Rationale**:
- **Kubernetes-native**: CRDs leverage K8s existing storage (etcd), RBAC, and API server
- **Declarative**: Pipelines, PipelineRuns, and Secrets are Kubernetes resources managed via kubectl
- **No external database**: Eliminates PostgreSQL/MySQL dependency, reducing operational complexity
- **Watch API**: K8s watch mechanism provides efficient change notifications for controllers
- **Versioning**: CRD versioning supports schema evolution
- **Consistency**: Strong consistency guarantees from etcd

**Alternatives Considered**:
- **PostgreSQL**: Traditional choice but adds external dependency, connection pooling complexity, and schema migration burden
- **SQLite**: Simpler than Postgres but doesn't support distributed deployments or multi-replica controllers
- **MongoDB/NoSQL**: Adds operational complexity and doesn't align with K8s paradigms

**CRD Structure**:
```yaml
# PipelineConfig CRD - defines pipeline structure
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: my-app-pipeline
  namespace: default
spec:
  repository: https://github.com/org/repo
  steps: [...]

# PipelineRun CRD - represents execution instance
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: my-app-pipeline-abc123
  namespace: default
spec:
  pipelineConfigRef: my-app-pipeline
  commit: abc123def456
status:
  phase: Running
  steps: [...]
```

**Impact**: Controller uses `controller-runtime` framework to watch CRDs and reconcile desired state. No separate database deployment needed.

---

## Decision 5: Log and Artifact Storage

**Decision**: Two-tier approach - Object Storage (S3/GCS/MinIO) for persistence + in-memory buffer for streaming

**Rationale**:
- **Scalability**: Object storage handles unlimited artifacts and logs without capacity planning
- **Cost-effective**: S3-compatible storage is cheaper than persistent volumes for large data
- **Real-time streaming**: In-memory circular buffers provide <500ms latency for live log tailing
- **Standard interface**: S3 API is universal (AWS S3, GCS, Azure Blob, MinIO, Ceph)
- **Retention policies**: Object storage supports lifecycle rules for 30-day log retention

**Architecture**:
1. **During execution**: Logs streamed to in-memory buffer (last 10MB per step) + written to object storage
2. **Live tailing**: Clients connect to API server WebSocket, receive buffered logs in real-time
3. **Historical logs**: After completion, clients fetch from object storage via signed URLs

**Alternatives Considered**:
- **Persistent Volumes only**: Expensive at scale, complex capacity management, slower than object storage
- **Database storage**: Poor fit for large binary artifacts and log data; expensive for high write volume
- **Elasticsearch**: Over-engineered for simple log storage; adds significant operational burden

**Impact**: Requires object storage configuration (S3 bucket + credentials). Logs available real-time during execution and persisted for 30 days.

---

## Decision 6: Testing Framework

**Decision**: Go standard testing (`testing` package) + Testify assertions + Kubernetes envtest

**Rationale**:
- **Native**: Go's built-in testing package is simple and sufficient
- **Testify**: Adds readable assertions (`assert.Equal`) without heavy framework
- **envtest**: Official Kubernetes testing tool that runs real API server for integration tests
- **Table-driven tests**: Go idiom for comprehensive test coverage with minimal code

**Test Structure**:
```
tests/
├── contract/          # API contract tests (OpenAPI validation)
│   ├── pipeline_config_test.go
│   └── pipeline_run_test.go
├── integration/       # End-to-end with envtest K8s
│   ├── pipeline_execution_test.go
│   └── secret_injection_test.go
└── unit/             # Isolated component tests
    ├── parser_test.go
    └── scheduler_test.go
```

**Alternatives Considered**:
- **Ginkgo/Gomega**: BDD-style but adds complexity; standard testing is sufficient
- **go-mockgen**: Mocking is often unnecessary with interface-based design

**Impact**: All tests use standard Go tooling. CI runs `go test ./...` for all test types.

---

## Decision 7: Webhook Receiver Pattern

**Decision**: Dedicated webhook receiver service with signature verification

**Rationale**:
- **Security**: Webhooks from GitHub/GitLab/Bitbucket must verify signatures (HMAC-SHA256)
- **Reliability**: Webhook receiver acknowledges receipt immediately, queues work asynchronously
- **Standard patterns**: GitHub, GitLab provide webhook libraries for Go

**Architecture**:
1. Webhook hits `/webhooks/:provider` endpoint
2. Signature verified against configured secret
3. Event parsed and PipelineRun CRD created
4. Controller watches PipelineRun and executes pipeline

**Alternatives Considered**:
- **Polling repositories**: Inefficient and slow; webhooks are push-based and instant
- **Embedded in controller**: Violates separation of concerns; webhook receiver should be stateless

**Impact**: Webhook service is stateless and horizontally scalable. Webhook secrets stored in Kubernetes Secrets.

---

## Decision 8: Pipeline Configuration Format

**Decision**: YAML-based declarative format (`.c8s.yaml` in repository root)

**Rationale**:
- **Kubernetes alignment**: YAML is K8s ecosystem standard (Helm, Kustomize, CRDs)
- **Human-readable**: Developers are familiar with YAML from Docker Compose, GitHub Actions, GitLab CI
- **Validation**: JSON Schema can validate pipeline configs before execution
- **Tooling**: Extensive YAML libraries in Go (`gopkg.in/yaml.v3`)

**Example Format**:
```yaml
version: v1alpha1
name: my-app-ci
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...
    resources:
      cpu: 1000m
      memory: 2Gi
  - name: build
    image: golang:1.21
    commands:
      - go build -o app
    dependsOn: [test]
    artifacts:
      - app
```

**Alternatives Considered**:
- **JSON**: Less human-readable, no comments
- **HCL (Terraform)**: Unfamiliar to most developers
- **Custom DSL**: Adds learning curve and maintenance burden

**Impact**: Pipeline parser validates YAML against JSON schema. CLI can lint configs before commit.

---

## Decision 9: Workload Execution Pattern

**Decision**: Kubernetes Jobs (one Job per step) with Pod templates

**Rationale**:
- **Native primitive**: Jobs provide automatic retry, completion tracking, and cleanup
- **Isolation**: Each step runs in isolated Pod with dedicated resources
- **Restart policy**: Jobs support `OnFailure` restart for transient errors
- **TTL cleanup**: `ttlSecondsAfterFinished` automatically cleans up completed Jobs
- **Resource limits**: Pod specs enforce CPU/memory limits per step

**Architecture**:
1. Controller creates Job for each pipeline step
2. Job spec includes init containers for workspace setup (git clone)
3. Main container runs user commands
4. Artifacts uploaded to object storage in sidecar container

**Alternatives Considered**:
- **Pods directly**: No retry or completion semantics; must implement manually
- **Argo Workflows**: Heavy dependency that duplicates our requirements; adds CRD complexity
- **Tekton**: Similar to Argo; designed for CI/CD but we want lighter-weight solution

**Impact**: Each step = one Kubernetes Job. Controller watches Job status to update PipelineRun status.

---

## Decision 10: Secret Management

**Decision**: Kubernetes Secrets with optional external provider integration (Vault/AWS Secrets Manager)

**Rationale**:
- **Native**: Kubernetes Secrets API provides basic secret storage and injection
- **RBAC**: K8s RBAC controls who can read/write secrets
- **Injection**: Secrets mounted as env vars or files in Job Pods
- **Masking**: Custom log filtering in controller masks secret values before storing logs
- **Extensibility**: Can integrate external providers via Secret Store CSI driver

**Secret Flow**:
1. Admin creates Secret via kubectl or API
2. PipelineConfig references secret by name: `secretRef: database-credentials`
3. Controller injects secret as env var in Job Pod
4. Log collector redacts any values matching secret patterns

**Alternatives Considered**:
- **Vault required**: Adds operational complexity; K8s Secrets sufficient for v1
- **No masking**: Violates SC-004 (secrets never exposed in logs)

**Impact**: Secrets are Kubernetes native resources. Log masking implemented in controller before persisting logs.

---

## Decision 11: Concurrency and Queueing

**Decision**: Controller work queue with rate limiting (controller-runtime's workqueue)

**Rationale**:
- **Proven pattern**: controller-runtime provides production-ready work queue
- **Rate limiting**: Exponential backoff for transient errors
- **Priority**: Can prioritize PipelineRuns based on labels (e.g., `priority: high`)
- **Parallelism**: Configurable number of worker goroutines

**Architecture**:
```
[CRD Watch] → [Work Queue] → [N Worker Goroutines] → [Reconcile]
                   ↓
            [Rate Limiter]
```

**Alternatives Considered**:
- **Redis queue**: External dependency for minimal benefit
- **Direct goroutines**: Lacks rate limiting and retry logic

**Impact**: Controller can process 100+ concurrent PipelineRuns (SC-002) by tuning worker count.

---

## Decision 12: Resource Quotas and Limits

**Decision**: Kubernetes ResourceQuotas + Custom admission webhook for pipeline-level limits

**Rationale**:
- **Native quotas**: K8s ResourceQuota objects limit total CPU/memory per namespace
- **Namespace isolation**: Each team gets dedicated namespace with quota
- **Admission control**: Validating webhook rejects PipelineRun if it would exceed team quota
- **Fair scheduling**: Kubernetes scheduler handles Pod placement based on available resources

**Quota Structure**:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: team-a
spec:
  hard:
    requests.cpu: "100"
    requests.memory: 200Gi
    pods: "50"
```

**Alternatives Considered**:
- **Manual tracking**: Error-prone and racy; K8s quotas are atomic
- **External quota system**: Duplicates K8s functionality

**Impact**: Teams isolated by namespace. Admission webhook prevents quota violations at PipelineRun creation time.

---

## Summary of Technical Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Language** | Go 1.25 | Kubernetes ecosystem standard, client-go, enhanced stdlib |
| **State Storage** | Kubernetes CRDs (etcd) | Native, declarative, no external DB |
| **API Framework** | net/http + http.ServeMux | Simple, stdlib only (Go 1.22+ enhanced routing) |
| **Log/Artifact Storage** | S3-compatible object storage | Scalable, cost-effective, standard interface |
| **Log Streaming** | In-memory buffer + WebSocket (stdlib) | Real-time (<500ms), minimal dependencies |
| **Workload Execution** | Kubernetes Jobs | Native primitive, isolation, retry |
| **Secret Management** | Kubernetes Secrets + masking | Native with log redaction |
| **Webhook Handling** | Dedicated service with signature verification | Security, reliability |
| **Pipeline Config** | YAML (`.c8s.yaml`) | K8s ecosystem standard, human-readable |
| **Testing** | Go testing + Testify + envtest | Native, simple, K8s integration testing |
| **Dashboard (optional)** | HTMX + Tailwind CSS + html/template | Zero build step, server-side rendering, ~14KB JS |
| **Queueing** | controller-runtime workqueue | Proven, rate-limited, concurrent |
| **Resource Quotas** | K8s ResourceQuota + admission webhook | Native, atomic, fair scheduling |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐       ┌──────────────┐      ┌─────────────┐ │
│  │   Webhook    │       │  Controller  │      │  API Server │ │
│  │   Receiver   │──────▶│   (Operator) │◀─────│   (REST)    │ │
│  └──────────────┘       └──────────────┘      └─────────────┘ │
│         │                       │                      │        │
│         │ Creates               │ Watches              │        │
│         ▼                       ▼                      ▼        │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │              Kubernetes API (etcd-backed CRDs)           │ │
│  │  - PipelineConfig    - PipelineRun    - Secrets          │ │
│  └──────────────────────────────────────────────────────────┘ │
│                              │                                  │
│                              │ Reconciles to                    │
│                              ▼                                  │
│         ┌──────────────────────────────────────┐              │
│         │      Kubernetes Jobs (per step)      │              │
│         │  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐ │              │
│         │  │ Pod │  │ Pod │  │ Pod │  │ Pod │ │              │
│         │  └─────┘  └─────┘  └─────┘  └─────┘ │              │
│         └──────────────────────────────────────┘              │
│                              │                                  │
└──────────────────────────────┼──────────────────────────────────┘
                               │ Logs/Artifacts
                               ▼
                    ┌─────────────────────┐
                    │  Object Storage     │
                    │  (S3/GCS/MinIO)     │
                    └─────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                         External Systems                         │
├─────────────────────────────────────────────────────────────────┤
│  GitHub/GitLab ──webhooks──▶ Webhook Receiver                   │
│  Developers ────REST/CLI───▶ API Server                         │
│  Developers ────WebSocket──▶ API Server (log streaming)         │
└─────────────────────────────────────────────────────────────────┘
```

---

## Open Questions / Future Research

1. **Cluster Autoscaler integration**: How to signal scaling needs to Cluster Autoscaler? (Answer: Cluster Autoscaler watches pending Pods automatically)
2. **Multi-cluster support**: Out of scope for v1, but how would federation work? (Future: KubeFed or custom aggregation layer)
3. **Windows containers**: Out of scope for v1, but would require different base images and runtime classes
4. **Artifact caching**: How to cache dependencies (npm_modules, .m2/repository) between runs? (Consider PVC caching or registry-based layer caching)
5. **Pipeline visualization**: How to render DAG of steps in dashboard? (Consider D3.js or Cytoscape.js)

---

## Next Steps

With all technical decisions resolved, proceed to **Phase 1: Design & Contracts**:

1. Generate `data-model.md` with CRD schemas for PipelineConfig, PipelineRun, etc.
2. Create `contracts/` directory with OpenAPI spec for REST API
3. Create `quickstart.md` for developer onboarding
4. Update agent context with selected technologies (Go, CRDs, S3, etc.)
5. Re-evaluate Constitution Check to ensure design maintains simplicity
