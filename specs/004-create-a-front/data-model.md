# Data Model: Web Dashboard for C8S CI Workflows

**Date**: 2025-10-26
**Feature**: Web Dashboard for C8S CI Workflows

This document defines the domain entities, their relationships, validation rules, and state transitions used in the C8S web dashboard.

---

## Core Entities

### 1. Project

Represents a connected Git repository integrated with C8S for CI/CD.

**Fields**:
- `id`: string (UUID) - Unique identifier
- `name`: string - Human-readable project name (e.g., "c8s-core")
- `description`: string (optional) - Project description
- `repository_url`: string - Git repository URL (e.g., "https://github.com/org/repo")
- `webhook_url`: string (read-only) - C8S webhook URL for GitHub/GitLab/Bitbucket registration
- `webhook_secret`: string (secret) - HMAC secret for webhook signature verification
- `namespace`: string - Kubernetes namespace where pipelines run
- `created_at`: timestamp - Project creation time
- `updated_at`: timestamp - Last modification time
- `last_run_at`: timestamp (nullable) - Last pipeline execution time
- `owner_id`: string - User ID of project owner
- `members`: array of Member - Team members with access

**Member subentity**:
- `user_id`: string - User identifier
- `role`: enum(admin, editor, viewer) - Access level
- `added_at`: timestamp - When user was added to project

**Validation Rules**:
- `name`: Required, 1-255 characters, unique within workspace
- `repository_url`: Required, valid Git URL format
- `namespace`: Required, valid Kubernetes namespace format (lowercase, alphanumeric + hyphens)
- `owner_id`: Required, references valid User
- At least one member required (owner)

**State Transitions**:
```
Created → Active (when first pipeline succeeds)
      ↓
   Active → Disabled (user-initiated)
      ↓
   Disabled → Active (user-initiated re-enable)
```

**Relationships**:
- 1:Many → PipelineRun (one project has many pipeline runs)
- Many:Many → User (through Member, projects and users)

**API Representation**:
```json
{
  "id": "proj-123",
  "name": "c8s-core",
  "description": "C8S CI/CD controller",
  "repository_url": "https://github.com/org/c8s",
  "webhook_url": "https://c8s.example.com/webhook/github?projectId=proj-123",
  "namespace": "c8s-project-default",
  "created_at": "2025-10-01T12:00:00Z",
  "updated_at": "2025-10-26T15:30:00Z",
  "last_run_at": "2025-10-26T14:45:00Z",
  "owner_id": "user-456",
  "members": [
    {
      "user_id": "user-456",
      "role": "admin",
      "added_at": "2025-10-01T12:00:00Z"
    }
  ]
}
```

---

### 2. PipelineRun

Represents a single execution of a pipeline triggered by a commit, tag, or manual trigger.

**Fields**:
- `id`: string (UUID) - Unique identifier (e.g., "c8s-xyz123")
- `project_id`: string - Reference to parent Project
- `name`: string - Pipeline name from configuration (e.g., "test-build-deploy")
- `status`: enum(queued, running, succeeded, failed, cancelled, timeout) - Current execution status
- `commit_sha`: string - Git commit hash that triggered pipeline
- `branch`: string - Git branch name (e.g., "main", "feature/dashboard")
- `tag`: string (nullable) - Git tag if triggered by tag push
- `author`: string - Git commit author name
- `author_email`: string - Git commit author email
- `trigger_source`: enum(webhook, api, manual, scheduled) - How pipeline was triggered
- `webhook_id`: string (nullable) - Reference to triggering webhook event
- `triggered_at`: timestamp - When pipeline was triggered
- `started_at`: timestamp (nullable) - When first step started
- `completed_at`: timestamp (nullable) - When last step completed
- `duration_seconds`: integer (nullable) - Total execution time in seconds
- `steps`: array of PipelineStep - Steps within this run
- `artifacts`: array of Artifact - Generated outputs
- `error_message`: string (nullable) - Failure reason if status is failed/timeout

**Validation Rules**:
- `id`: Required, unique
- `project_id`: Required, references valid Project
- `commit_sha`: Required, valid Git SHA-1 hash format
- `branch`: Required, non-empty string
- `author`: Required
- `status`: Must be one of allowed enum values
- `trigger_source`: Required, must be one of allowed enum values
- Duration calculated as: `completed_at - started_at` (only when both present)

**State Transitions**:
```
           ┌─→ Succeeded (all steps succeeded)
Queued ──→ Running ──→ ┼─→ Failed (any step failed)
           │           └─→ Timeout (execution exceeded limit)
           └───────────────→ Cancelled (user-initiated)
```

**Relationships**:
- Many:1 → Project (many runs per project)
- 1:Many → PipelineStep (one run has many steps)
- 1:Many → Artifact (run generates artifacts)
- Many:1 → User (run triggered by user/webhook)

**API Representation**:
```json
{
  "id": "c8s-abc123",
  "project_id": "proj-123",
  "name": "test-build-deploy",
  "status": "succeeded",
  "commit_sha": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "branch": "main",
  "tag": null,
  "author": "Alice Developer",
  "author_email": "alice@example.com",
  "trigger_source": "webhook",
  "webhook_id": "wh-456",
  "triggered_at": "2025-10-26T14:30:00Z",
  "started_at": "2025-10-26T14:30:15Z",
  "completed_at": "2025-10-26T14:35:45Z",
  "duration_seconds": 330,
  "error_message": null,
  "steps": [
    {
      "id": "step-1",
      "name": "test",
      "status": "succeeded",
      "duration_seconds": 120
    }
  ],
  "artifacts": [
    {
      "id": "art-1",
      "name": "test-report.html",
      "type": "test-report"
    }
  ]
}
```

---

### 3. PipelineStep

Represents a single step (task) within a pipeline run.

**Fields**:
- `id`: string (UUID) - Unique identifier
- `pipeline_run_id`: string - Reference to parent PipelineRun
- `name`: string - Step name from configuration (e.g., "test", "build", "deploy")
- `status`: enum(pending, running, succeeded, failed, skipped, timeout) - Current status
- `image`: string - Container image used for execution (e.g., "golang:1.24")
- `commands`: array of string - Commands executed in step
- `started_at`: timestamp (nullable) - When step execution started
- `completed_at`: timestamp (nullable) - When step execution completed
- `duration_seconds`: integer (nullable) - Execution time
- `exit_code`: integer (nullable) - Process exit code (0 for success, non-zero for failure)
- `log_url`: string - URL to fetch logs for this step
- `resource_usage`: ResourceUsage - CPU and memory consumed
- `depends_on`: array of string - Step names this step depends on (for DAG)
- `error_output`: string (nullable) - Stderr from failed step
- `conditional`: Conditional (nullable) - Execution conditions (branch filters, etc.)
- `retry_count`: integer - Number of times this step was retried
- `max_retries`: integer - Maximum allowed retries

**ResourceUsage subentity**:
- `cpu_millicores`: integer - CPU used in millicores
- `memory_bytes`: integer - Memory used in bytes
- `cpu_limit_millicores`: integer - CPU limit in millicores
- `memory_limit_bytes`: integer - Memory limit in bytes

**Conditional subentity** (optional):
- `branch_pattern`: string (regex) - Only run if branch matches pattern
- `skip_on_tag`: boolean - Skip step if triggered by tag
- `on_success`: boolean - Only run if previous steps succeeded
- `on_failure`: boolean - Only run if previous steps failed

**Validation Rules**:
- `id`: Required, unique per pipeline run
- `pipeline_run_id`: Required, references valid PipelineRun
- `name`: Required, 1-255 characters
- `image`: Required, valid container image format
- `status`: One of allowed enum values
- `depends_on`: References must be valid step names in same pipeline
- `exit_code`: When present, must be >= 0
- `retry_count`: Must be <= `max_retries`

**State Transitions**:
```
           ┌─→ Succeeded (exit_code = 0)
Pending ──→ Running ──→ ┼─→ Failed (exit_code != 0, no retry or max_retries exceeded)
           │            └─→ Skipped (conditional not met)
           └───────────────────→ Timeout

Failed ──→ Pending (retry) ──→ ...
```

**Relationships**:
- Many:1 → PipelineRun (many steps per run)
- 1:Many → LogEntry (one step generates many log lines, managed via LogURL)
- 1:Many → Artifact (step can generate artifacts)

**API Representation**:
```json
{
  "id": "step-1",
  "pipeline_run_id": "c8s-abc123",
  "name": "test",
  "status": "succeeded",
  "image": "golang:1.24",
  "commands": [
    "go test ./..."
  ],
  "started_at": "2025-10-26T14:30:15Z",
  "completed_at": "2025-10-26T14:32:15Z",
  "duration_seconds": 120,
  "exit_code": 0,
  "log_url": "/api/runs/c8s-abc123/steps/step-1/logs",
  "resource_usage": {
    "cpu_millicores": 800,
    "memory_bytes": 1073741824,
    "cpu_limit_millicores": 1000,
    "memory_limit_bytes": 2147483648
  },
  "depends_on": [],
  "error_output": null,
  "conditional": null,
  "retry_count": 0,
  "max_retries": 3
}
```

---

### 4. Artifact

Represents a file or output generated by a pipeline step.

**Fields**:
- `id`: string (UUID) - Unique identifier
- `pipeline_run_id`: string - Reference to PipelineRun
- `step_id`: string - Reference to PipelineStep that generated artifact
- `name`: string - Artifact filename (e.g., "test-report.html", "build.jar")
- `type`: enum(binary, report, documentation, log, other) - Artifact classification
- `mime_type`: string - MIME type (e.g., "text/html", "application/octet-stream")
- `size_bytes`: integer - File size in bytes
- `url`: string - Download URL for artifact
- `created_at`: timestamp - When artifact was created
- `expires_at`: timestamp (nullable) - When artifact will be deleted (if applicable)
- `metadata`: map[string]string - Custom metadata (e.g., build target, platform)

**Validation Rules**:
- `id`: Required, unique
- `pipeline_run_id`: Required, references valid PipelineRun
- `step_id`: Required, references valid PipelineStep
- `name`: Required, 1-512 characters
- `type`: One of allowed enum values
- `mime_type`: Valid MIME type format
- `size_bytes`: >= 0
- `url`: Valid HTTP(S) URL

**Relationships**:
- Many:1 → PipelineRun (run can have many artifacts from different steps)
- Many:1 → PipelineStep (step can generate multiple artifacts)

**API Representation**:
```json
{
  "id": "art-1",
  "pipeline_run_id": "c8s-abc123",
  "step_id": "step-1",
  "name": "test-report.html",
  "type": "report",
  "mime_type": "text/html",
  "size_bytes": 245632,
  "url": "https://storage.example.com/artifacts/test-report.html?token=xyz",
  "created_at": "2025-10-26T14:32:16Z",
  "expires_at": "2025-11-26T14:32:16Z",
  "metadata": {
    "test_framework": "pytest",
    "coverage_percent": "87.5"
  }
}
```

---

### 5. User

Represents a dashboard user with access to projects and role-based permissions.

**Fields**:
- `id`: string (UUID) - Unique identifier
- `username`: string - Unique username for login
- `email`: string - User email address
- `full_name`: string - User's full name
- `avatar_url`: string (nullable) - Avatar image URL
- `created_at`: timestamp - Account creation time
- `last_login_at`: timestamp (nullable) - Last authentication time
- `roles`: array of string - Global roles (admin, user, viewer)
- `project_memberships`: array of ProjectMembership - Per-project access

**ProjectMembership subentity**:
- `project_id`: string - Project reference
- `role`: enum(admin, editor, viewer) - Project-specific role
- `added_at`: timestamp - When user joined project

**Validation Rules**:
- `id`: Required, unique
- `username`: Required, 3-32 characters, alphanumeric + underscore, unique
- `email`: Required, valid email format, unique
- `full_name`: 1-255 characters
- `roles`: Array of valid role enum values, at least one role required

**Relationships**:
- 1:Many → ProjectMembership (user can be member of many projects)
- 1:Many → PipelineRun (via webhook user reference)

**API Representation**:
```json
{
  "id": "user-456",
  "username": "alice",
  "email": "alice@example.com",
  "full_name": "Alice Developer",
  "avatar_url": "https://avatars.example.com/alice.jpg",
  "created_at": "2025-09-01T08:00:00Z",
  "last_login_at": "2025-10-26T14:00:00Z",
  "roles": ["user"],
  "project_memberships": [
    {
      "project_id": "proj-123",
      "role": "admin",
      "added_at": "2025-10-01T12:00:00Z"
    }
  ]
}
```

---

### 6. LogEntry

Represents a single line of output from a pipeline step execution.

**Note**: LogEntry is not explicitly queried by the dashboard UI. Instead, logs are streamed via the `/api/runs/{runId}/steps/{stepId}/logs` endpoint using Server-Sent Events. LogEntry structure is documented here for completeness.

**Fields**:
- `id`: string (UUID) - Unique identifier
- `pipeline_run_id`: string - Reference to PipelineRun
- `step_id`: string - Reference to PipelineStep
- `line_number`: integer - Sequence number (for ordering)
- `timestamp`: timestamp - When log line was generated
- `level`: enum(debug, info, warning, error) - Log level
- `message`: string - Log message content
- `source`: enum(stdout, stderr) - Output stream source

**Validation Rules**:
- `step_id`: Required, references valid PipelineStep
- `line_number`: >= 1, unique per step
- `message`: Non-empty string
- `level`: One of allowed enum values
- `source`: One of allowed enum values

**Storage Strategy**: Logs stored in S3-compatible object storage, not relational database, for scalability.

**API Representation** (SSE format):
```
data: {"timestamp":"2025-10-26T14:30:20Z","level":"info","message":"Running tests...","source":"stdout"}

data: {"timestamp":"2025-10-26T14:30:25Z","level":"error","message":"Test failure: TestPipeline","source":"stderr"}
```

---

## Entity Relationship Diagram

```
┌─────────────┐
│   Project   │
│  (workspace)│
└──────┬──────┘
       │
       ├─ 1:Many ──→ PipelineRun
       │
       └─ Many:Many ──→ User (via ProjectMembership)


┌─────────────────┐
│  PipelineRun    │
│ (one execution) │
└────────┬────────┘
         │
         ├─ 1:Many ──→ PipelineStep
         │
         ├─ 1:Many ──→ Artifact
         │
         └─ Many:1 ──→ Project


┌──────────────────┐
│  PipelineStep    │
│  (one task)      │
└────────┬─────────┘
         │
         ├─ 1:Many ──→ LogEntry (via log streaming endpoint)
         │
         ├─ 1:Many ──→ Artifact
         │
         └─ Many:1 ──→ PipelineRun


┌──────────────┐
│   Artifact   │
│  (output)    │
└──────┬───────┘
       │
       ├─ Many:1 ──→ PipelineRun
       │
       └─ Many:1 ──→ PipelineStep


┌─────────────┐
│    User     │
│  (dashboard)│
└──────┬──────┘
       │
       └─ 1:Many ──→ ProjectMembership (pivot to Project)
```

---

## Key Design Decisions

### 1. Status Enums as Strings (Not Codes)
**Decision**: Use human-readable enum strings (e.g., "succeeded", "failed") instead of numeric codes.

**Rationale**: Makes logs, debugging, and UI templates more readable. Minimal JSON size overhead. Easier for humans to understand at a glance.

**Example**: `"status": "succeeded"` instead of `"status": 1`

---

### 2. Timestamps in ISO 8601 Format
**Decision**: All timestamps use ISO 8601 format with timezone (e.g., "2025-10-26T14:30:00Z").

**Rationale**: Language-agnostic, sortable, unambiguous timezone handling, human-readable in logs.

---

### 3. Artifacts Reference PipelineRun AND PipelineStep
**Decision**: Artifacts include both `pipeline_run_id` and `step_id`.

**Rationale**:
- Allows querying all artifacts for a run or specific step
- Enables dashboard to group artifacts by step for clear UI organization
- Supports filtering artifacts by step status (failed step artifacts highlighted)

---

### 4. Logs Via Streaming API (Not Database Entities)
**Decision**: Logs are streamed via Server-Sent Events endpoint, not stored as queryable entities.

**Rationale**:
- Logs are append-only, high-volume, short-lived data
- Relational database not optimized for append-heavy workloads
- S3 object storage scales infinitely, costs less than database storage
- Streaming SSE directly from S3 avoids memory overhead
- Simplifies retention policies (delete old log files, not database rows)

---

### 5. ResourceUsage as Nested Object
**Decision**: CPU/memory metrics nested within PipelineStep instead of separate entity.

**Rationale**:
- Avoids excessive table joins
- Resource usage is always queried together with step status
- Simplified data model without sacrificing queryability

---

### 6. Conditional Execution as Optional Nested Object
**Decision**: Conditional step execution rules nested within PipelineStep.

**Rationale**:
- Conditions are specific to individual step, not shared across multiple steps
- Most steps have no conditions; optional field keeps API payloads small
- Avoids unnecessary joins for simple workflows

---

## Data Validation Summary

| Entity | Key Constraints | Database Type | Indices |
|--------|-----------------|---------------|---------|
| Project | name, repo_url unique; ns format; owner required | PostgreSQL | project_name, project_owner, namespace |
| PipelineRun | commit_sha format; status enum; project_id FK | PostgreSQL | run_project, run_status, run_branch, run_created |
| PipelineStep | step_name, depends_on validation; status enum | PostgreSQL | step_run, step_status, step_name |
| Artifact | mime_type format; size >= 0; step_id FK | PostgreSQL | artifact_run, artifact_step, artifact_type |
| User | username/email unique; role enum | PostgreSQL | user_email, user_username |
| LogEntry | line_number unique per step; level enum | S3 Object Storage | (N/A - streaming API) |

---

## API Integration Points

The following entities are exposed through REST API endpoints (detailed in `contracts/api-schema.md`):

- **GET /api/projects** - List projects user has access to
- **GET /api/projects/{projectId}** - Fetch project details
- **POST /api/projects** - Create project (admin only)
- **GET /api/projects/{projectId}/runs** - List pipeline runs for project
- **GET /api/runs/{runId}** - Fetch pipeline run detail
- **GET /api/runs/{runId}/steps/{stepId}/logs** - Stream logs (SSE)
- **GET /api/runs/{runId}/artifacts** - List artifacts for run
- **POST /api/projects/{projectId}/webhook** - Register/rotate webhook

---

## Data Consistency and Integrity

### Transactional Boundaries
- **Pipeline Run Creation**: Atomic - create run + initialize steps in single transaction
- **Step Completion**: Atomic - update step status + calculate run status in single transaction
- **Artifact Registration**: Atomic with storage - verify file exists in S3 before creating database record

### Foreign Key Constraints
- Deleting a Project cascades to delete all associated PipelineRuns
- Deleting a PipelineRun cascades to delete all associated Steps and LogEntry references
- Deleting a User cascades to remove ProjectMembership but not project itself

### Eventual Consistency
- LogEntry records may be written to S3 with eventual consistency (AWS standard)
- Pipeline status updates propagate to dashboard via SSE (real-time)
- Project membership changes reflected immediately on user's next request

---

## Migration Path from Core C8S

**Existing C8S CRD entities that map to this model**:

| C8S CRD | Dashboard Entity | Notes |
|---------|-----------------|-------|
| Project (custom CR) | Project | Already exists in C8S, dashboard reuses |
| PipelineRun (custom CR) | PipelineRun | Maps directly with status mapping |
| PipelineConfig (custom CR) | PipelineRun.name (from config) | Configuration stored in CRD, run inherits |
| Job (Kubernetes) | PipelineStep | Derived from Job resources created for each step |

The dashboard reads from existing Kubernetes resources and C8S database, avoiding duplication.

---

**Data model complete**. Ready for API contract specification (Phase 1b).
