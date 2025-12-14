# C8S Architecture

Complete architectural overview of the C8S continuous integration system.

## System Overview

C8S is a **Kubernetes-native CI system** that runs pipeline steps as isolated container Jobs, leveraging Kubernetes primitives for orchestration, scheduling, and state management.

### Core Principles

1. **Kubernetes-Native**: Everything is a Kubernetes resource (CRDs, Jobs, Pods)
2. **Declarative**: Pipelines defined in YAML, stored as CRDs
3. **Isolated Execution**: Each step runs in its own Job/Pod
4. **Git-Triggered**: Webhooks from GitHub/GitLab/Bitbucket trigger pipelines
5. **Observable**: Logs streamed to S3, metrics exposed via Prometheus
6. **Secure**: Secrets managed via Kubernetes Secrets, logs masked

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          Git Providers                          │
│              (GitHub, GitLab, Bitbucket)                        │
└────────────────────────┬────────────────────────────────────────┘
                         │ Webhook (push, PR)
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      C8S Webhook Service                        │
│  - Validates webhook signature                                  │
│  - Parses git event (commit, branch, author)                   │
│  - Creates PipelineRun CRD                                     │
└────────────────────────┬────────────────────────────────────────┘
                         │ Creates
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                   PipelineRun CRD (Custom Resource)             │
│  - Stores: pipeline config, commit SHA, branch, status         │
│  - Watched by: Controller                                      │
└────────────────────────┬────────────────────────────────────────┘
                         │ Watches
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                       C8S Controller                            │
│  - Reconciles PipelineRun resources                            │
│  - Resolves step dependencies (DAG)                            │
│  - Creates Kubernetes Jobs (one per step)                      │
│  - Updates PipelineRun status                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │ Creates
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Jobs & Pods                       │
│  - Each step = one Job                                         │
│  - Job creates Pod with specified image                        │
│  - Pod executes commands                                       │
│  - Logs streamed to stdout/stderr                              │
│  - Artifacts uploaded to S3                                    │
└────────────────────────┬────────────────────────────────────────┘
                         │ Logs & Artifacts
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    S3-Compatible Storage                        │
│  - MinIO (local dev) or AWS S3 (production)                   │
│  - Stores: logs, build artifacts                              │
│  - Accessed by: API server, dashboard                         │
└─────────────────────────────────────────────────────────────────┘
                         ▲
                         │ Fetches logs/artifacts
                         │
┌─────────────────────────────────────────────────────────────────┐
│                       C8S API Server                            │
│  - REST API for pipeline management                            │
│  - Serves dashboard UI (HTMX templates)                        │
│  - Streams logs via SSE                                        │
│  - Provides artifact download                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Users (CLI, Browser, CI/CD)                   │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Webhook Service

**Purpose**: Receives git webhook events and creates PipelineRun resources

**Technology**: Go HTTP server

**Responsibilities**:
- Validates webhook signatures (HMAC-SHA256)
- Parses webhook payload (GitHub/GitLab/Bitbucket formats)
- Looks up PipelineConfig CRD for repository
- Creates PipelineRun CRD with commit metadata
- Returns HTTP 200 to git provider

**Flow**:
```
1. Git provider sends POST to /api/v1/webhook
2. Webhook service validates signature
3. Parses JSON payload → extract repo, branch, commit SHA, author
4. Queries Kubernetes for PipelineConfig matching repo
5. Creates PipelineRun CRD with:
   - Pipeline config reference
   - Commit metadata (SHA, branch, author, message)
   - Timestamp
6. Returns 200 OK (or 404 if no PipelineConfig found)
```

**Code**: `cmd/webhook/`, `pkg/webhook/`

### 2. Controller

**Purpose**: Kubernetes controller that reconciles PipelineRun resources

**Technology**: Kubebuilder controller-runtime

**Responsibilities**:
- Watch PipelineRun CRDs for create/update/delete events
- Resolve step dependencies (build DAG)
- Create Kubernetes Jobs for each step
- Monitor Job status and update PipelineRun status
- Handle retries, timeouts, failures
- Clean up completed Jobs

**Reconciliation Loop**:
```
1. Receive PipelineRun event (created/updated)
2. Check current phase:
   - Pending → Start execution
   - Running → Monitor jobs
   - Succeeded/Failed → Cleanup
3. For Pending:
   - Parse pipeline config
   - Build dependency graph (DAG)
   - Create Jobs for steps with no dependencies
4. For Running:
   - Query Job statuses
   - Update PipelineRun status
   - Create Jobs for steps whose dependencies completed
5. For Succeeded/Failed:
   - Clean up Jobs (optional)
   - Update final status
6. Requeue if needed (active pipelines)
```

**Code**: `cmd/controller/`, `pkg/controller/`

### 3. API Server

**Purpose**: REST API and web dashboard for C8S

**Technology**: Go HTTP server with chi router, HTMX templates

**Responsibilities**:
- CRUD operations on PipelineConfig, PipelineRun
- Project management
- Log retrieval from S3
- Artifact management (upload/download)
- Server-Sent Events (SSE) for real-time log streaming
- Dashboard UI (HTMX + Tailwind CSS)
- Authentication (JWT)

**Endpoints**:
```
GET  /                           → Dashboard home
GET  /api/v1/pipelines           → List pipelines
POST /api/v1/pipelines           → Create pipeline
GET  /api/v1/pipelines/:id       → Get pipeline details
GET  /api/v1/runs                → List pipeline runs
GET  /api/v1/runs/:id            → Get run details
GET  /api/v1/runs/:id/logs       → Get logs for run
GET  /api/v1/runs/:id/logs/stream → SSE log stream
GET  /api/v1/artifacts/:id       → Download artifact
POST /api/v1/webhook             → Webhook endpoint (handled by webhook service)
```

**Code**: `cmd/api-server/`, `pkg/api/`, `pkg/dashboard/`

### 4. Storage Layer

**Purpose**: Persist logs and artifacts to S3-compatible storage

**Technology**: AWS S3 SDK, MinIO (local dev)

**Responsibilities**:
- Upload logs from completed Jobs
- Upload artifacts from pipeline steps
- Provide presigned URLs for downloads
- Implement retention policies
- Support multiple backends (S3, MinIO, GCS)

**Storage Structure**:
```
s3://bucket/
├── logs/
│   └── {pipeline-run-id}/
│       └── {step-name}.log
└── artifacts/
    └── {pipeline-run-id}/
        └── {artifact-name}
```

**Code**: `pkg/storage/`, `pkg/logstorage/`

### 5. CLI Tool

**Purpose**: Command-line interface for C8S operations

**Technology**: Go CLI with cobra

**Commands**:
```
c8s run <pipeline>               # Trigger pipeline run
c8s logs <run-id> [--follow]     # View logs
c8s status <run-id>              # Get run status
c8s list [pipelines|runs]        # List resources
c8s create pipeline <file>       # Create pipeline from YAML
c8s delete <resource> <id>       # Delete resource
```

**Code**: `cmd/c8s/`, `pkg/cli/`

## Data Model

### Custom Resource Definitions (CRDs)

#### PipelineConfig

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: my-app-pipeline
spec:
  repository: https://github.com/org/my-app
  branches: [main, develop]
  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test ./...
      resources:
        cpu: 1000m
        memory: 2Gi
    - name: build
      image: golang:1.25
      commands:
        - go build -o app
      dependsOn: [test]
      artifacts:
        - app
```

**Purpose**: Defines how to run CI for a repository

**Stored**: Kubernetes etcd (via API server)

#### PipelineRun

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: my-app-pipeline-abc123
spec:
  pipelineConfigRef: my-app-pipeline
  commit:
    sha: abc123def456
    branch: main
    author: john@example.com
    message: "Fix bug in handler"
  triggeredBy: webhook
status:
  phase: Running
  steps:
    - name: test
      status: Succeeded
      startTime: 2025-01-15T10:00:00Z
      completionTime: 2025-01-15T10:02:30Z
    - name: build
      status: Running
      startTime: 2025-01-15T10:02:31Z
```

**Purpose**: Represents one execution of a pipeline

**Created by**: Webhook service or manual trigger

## Request Flow Examples

### Example 1: GitHub Push Triggers Pipeline

```
1. Developer pushes to GitHub (main branch)
   ↓
2. GitHub sends webhook POST to https://c8s.example.com/api/v1/webhook
   Body: { repository: {...}, ref: "refs/heads/main", commits: [...] }
   ↓
3. Webhook service:
   - Validates HMAC signature
   - Parses payload
   - Finds PipelineConfig for repository
   - Creates PipelineRun:
     apiVersion: c8s.dev/v1alpha1
     kind: PipelineRun
     metadata:
       name: my-app-pipeline-1234567890
     spec:
       pipelineConfigRef: my-app-pipeline
       commit:
         sha: abc123
         branch: main
   ↓
4. Controller detects new PipelineRun:
   - Reads PipelineConfig
   - Parses steps and dependencies
   - Creates DAG: test → build → deploy
   - Creates Job for "test" step (no dependencies)
   ↓
5. Kubernetes schedules Job:
   - Creates Pod with golang:1.25 image
   - Mounts secrets if needed
   - Runs commands: go test ./...
   - Streams logs to stdout
   ↓
6. Controller monitors Job:
   - Job succeeds → Updates PipelineRun status
   - Creates Job for "build" step (dependency satisfied)
   ↓
7. Process repeats for remaining steps
   ↓
8. All steps complete:
   - Controller updates PipelineRun phase: Succeeded
   - Logs uploaded to S3
   - Artifacts uploaded to S3
   ↓
9. User views results:
   - Via dashboard: https://c8s.example.com/pipelines/1234567890
   - Via CLI: c8s logs 1234567890
   - Logs streamed from S3
```

### Example 2: Developer Views Real-Time Logs

```
1. User opens dashboard: https://c8s.example.com/runs/1234567890
   ↓
2. Frontend JavaScript connects to SSE:
   GET /api/v1/runs/1234567890/logs/stream
   ↓
3. API server:
   - Opens SSE connection
   - Subscribes to log events for run 1234567890
   - Streams existing logs from S3
   - Continues streaming new logs as they arrive
   ↓
4. Controller writes logs to S3:
   - Watches Pod logs (kubectl logs --follow)
   - Uploads to S3 in chunks
   - Notifies subscribers (via in-memory pubsub)
   ↓
5. API server receives notification:
   - Fetches new log chunk from S3
   - Sends SSE event to browser:
     data: {"step": "build", "line": "Compiling...\n"}
   ↓
6. Browser receives SSE event:
   - Appends log line to viewer
   - Auto-scrolls to bottom
   - Continues until connection closed
```

## Deployment Architecture

### Local Development (Tilt)

```
┌─────────────────────────────────────┐
│         Developer Laptop            │
│  ┌───────────────────────────────┐ │
│  │  Tilt                         │ │
│  │  ├─ Watches files             │ │
│  │  ├─ Compiles Go binaries      │ │
│  │  ├─ Builds Docker images      │ │
│  │  └─ Deploys to kind cluster   │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  kind Cluster                 │ │
│  │  ├─ c8s-controller (Pod)      │ │
│  │  ├─ c8s-api-server (Pod)      │ │
│  │  ├─ c8s-webhook (Pod)         │ │
│  │  ├─ minio (Pod)               │ │
│  │  └─ Pipeline Job Pods         │ │
│  └───────────────────────────────┘ │
└─────────────────────────────────────┘
```

### Production (Kubernetes Cluster)

```
┌─────────────────────────────────────────────────────┐
│                 Kubernetes Cluster                  │
│  ┌────────────────────────────────────────────────┐│
│  │  Namespace: c8s-system                         ││
│  │  ├─ Deployment: c8s-controller (3 replicas)   ││
│  │  ├─ Deployment: c8s-api-server (3 replicas)   ││
│  │  ├─ Deployment: c8s-webhook (2 replicas)      ││
│  │  ├─ Service: c8s-api-server (LoadBalancer)    ││
│  │  ├─ Service: c8s-webhook (LoadBalancer)       ││
│  │  ├─ Ingress: TLS termination                  ││
│  │  └─ HPA: Auto-scaling                         ││
│  └────────────────────────────────────────────────┘│
│  ┌────────────────────────────────────────────────┐│
│  │  Namespace: c8s-pipelines                     ││
│  │  ├─ PipelineConfig CRDs                       ││
│  │  ├─ PipelineRun CRDs                          ││
│  │  └─ Job Pods (pipeline execution)             ││
│  └────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────┘
             │                      │
             ▼                      ▼
    ┌──────────────┐      ┌──────────────┐
    │   AWS S3     │      │  Prometheus  │
    │  (Logs/      │      │  (Metrics)   │
    │  Artifacts)  │      │              │
    └──────────────┘      └──────────────┘
```

## Security Architecture

### Authentication & Authorization

1. **API Authentication**: JWT tokens
2. **Webhook Authentication**: HMAC signatures
3. **Inter-component**: Kubernetes ServiceAccounts
4. **Secret Management**: Kubernetes Secrets
5. **Log Masking**: Automatic secret redaction

### Network Security

1. **TLS**: All external traffic encrypted (webhook, API)
2. **Network Policies**: Pod-to-pod communication restricted
3. **RBAC**: Least-privilege access to Kubernetes resources
4. **Ingress**: Rate limiting, DDoS protection

### Secret Injection

```yaml
# PipelineConfig with secret
steps:
  - name: deploy
    image: ubuntu:22.04
    commands:
      - ./deploy.sh --token=$API_TOKEN
    secrets:
      - secretRef: deploy-credentials
        key: API_TOKEN
        envVar: API_TOKEN
```

Controller injects secret as environment variable into Job Pod.

## Scalability

### Horizontal Scaling

- **Controller**: Multiple replicas with leader election
- **API Server**: Stateless, scale to N replicas
- **Webhook**: Stateless, scale to N replicas

### Performance Optimizations

1. **Controller**: Caching, work queues, parallel reconciliation
2. **API Server**: Connection pooling, response caching
3. **Storage**: S3 multipart uploads, CDN for artifacts
4. **Database**: Kubernetes etcd (CRD storage)

### Resource Limits

```yaml
# Example resource limits
resources:
  controller:
    requests: { cpu: 500m, memory: 512Mi }
    limits: { cpu: 2000m, memory: 2Gi }
  api-server:
    requests: { cpu: 250m, memory: 256Mi }
    limits: { cpu: 1000m, memory: 1Gi }
  pipeline-jobs:
    requests: { cpu: 100m, memory: 128Mi }
    limits: { cpu: 4000m, memory: 8Gi }  # User-configurable
```

## Observability

### Metrics (Prometheus)

```
# Controller metrics
c8s_pipeline_runs_total{status="succeeded|failed|running"}
c8s_pipeline_duration_seconds{pipeline="name"}
c8s_controller_reconcile_duration_seconds
c8s_jobs_created_total

# API Server metrics
c8s_api_requests_total{endpoint="/api/v1/runs", status="200"}
c8s_api_request_duration_seconds{endpoint="/api/v1/runs"}
c8s_sse_connections_active
```

### Logging

- **Structured Logging**: JSON format (logr)
- **Log Levels**: Debug, Info, Error
- **Correlation IDs**: Trace requests across components
- **Log Aggregation**: Sent to S3 or CloudWatch

### Tracing

Future: OpenTelemetry integration for distributed tracing

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| Framework | Kubebuilder (controller), chi (HTTP router) |
| Frontend | HTMX + Tailwind CSS |
| Storage | S3-compatible (MinIO, AWS S3) |
| Database | Kubernetes etcd (via CRDs) |
| Deployment | Helm 3.x |
| Local Dev | Tilt + kind |
| Testing | Go test, Playwright, axe-core |
| CI/CD | GitHub Actions |

## Design Decisions

### Why Kubernetes-Native?

- **Leverage existing infrastructure**: No separate worker pools
- **Declarative**: Pipeline state stored as CRDs
- **Scalable**: Kubernetes handles scheduling, resource limits
- **Portable**: Works on any Kubernetes cluster

### Why CRDs over Database?

- **Native Kubernetes storage**: etcd is already HA
- **kubectl integration**: Standard tools work
- **Watches**: Efficient event-driven updates
- **RBAC**: Kubernetes-native access control

### Why Jobs over Pods?

- **Automatic retries**: Job controller handles failures
- **Completion tracking**: Built-in status
- **Parallelism**: Job can create multiple pods
- **Cleanup**: Job manages pod lifecycle

### Why S3 for Logs?

- **Scalable**: Handle TB of logs
- **Durable**: Multi-region replication
- **Cost-effective**: Cheap storage
- **Standard**: S3-compatible APIs everywhere

### Why HTMX over React?

- **Simplicity**: Server-side rendering, less JavaScript
- **Performance**: Smaller bundle size, faster loads
- **Maintainability**: Go developers can work on frontend
- **Progressive enhancement**: Works without JavaScript

## Future Architecture

### Planned Improvements

1. **Multi-tenancy**: Namespace isolation per organization
2. **Plugin system**: Custom step types
3. **Distributed caching**: Faster builds
4. **Matrix builds**: Native support for parallel variants
5. **ARM64 support**: Multi-architecture builds
6. **Queue system**: Limit concurrent pipelines

## References

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Kubernetes CRD Documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [HTMX Documentation](https://htmx.org/)

---

For implementation details, see:
- [Development Guide](../development/development.md)
- [API Documentation](./api-reference.md)
- [CRD Specification](../../specs/001-build-a-continuous/data-model.md)
