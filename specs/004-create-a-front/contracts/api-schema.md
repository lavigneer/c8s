# API Schema: C8S Dashboard

**Date**: 2025-10-26
**Format**: RESTful JSON API
**Base URL**: `http://localhost:8080/api`
**Authentication**: Bearer token (OAuth2/JWT from existing C8S auth)

---

## Authentication

All endpoints require the `Authorization: Bearer {token}` header with a valid access token from the C8S authentication system.

```
GET /api/projects HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Host: localhost:8080
```

---

## Response Format

All responses follow a consistent structure:

**Success Response (2xx)**:
```json
{
  "success": true,
  "data": { /* entity or array */ },
  "meta": { "total": 10, "page": 1 }  // Only for paginated endpoints
}
```

**Error Response (4xx/5xx)**:
```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Pipeline run not found",
    "details": {}
  }
}
```

---

## Projects API

### List Projects

**Endpoint**: `GET /api/projects`

**Query Parameters**:
- `page`: integer (default: 1) - Page number for pagination
- `per_page`: integer (default: 20, max: 100) - Items per page
- `search`: string (optional) - Filter projects by name or repo URL substring

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
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
      "member_count": 3
    }
  ],
  "meta": {
    "total": 5,
    "page": 1,
    "per_page": 20
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid or missing token
- `403 Forbidden` - User not authorized to access projects

---

### Get Project Details

**Endpoint**: `GET /api/projects/{projectId}`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
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
        "username": "alice",
        "role": "admin",
        "added_at": "2025-10-01T12:00:00Z"
      },
      {
        "user_id": "user-789",
        "username": "bob",
        "role": "editor",
        "added_at": "2025-10-10T10:00:00Z"
      }
    ]
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access to this project
- `404 Not Found` - Project doesn't exist

---

### Create Project

**Endpoint**: `POST /api/projects`

**Headers**: `Content-Type: application/json`

**Request Body**:
```json
{
  "name": "my-new-project",
  "description": "My project description",
  "repository_url": "https://github.com/myorg/myrepo",
  "namespace": "c8s-my-project"
}
```

**Validation**:
- `name`: Required, 1-255 chars, unique
- `repository_url`: Required, valid Git URL
- `namespace`: Required, valid Kubernetes namespace format

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": "proj-999",
    "name": "my-new-project",
    "description": "My project description",
    "repository_url": "https://github.com/myorg/myrepo",
    "webhook_url": "https://c8s.example.com/webhook/github?projectId=proj-999",
    "namespace": "c8s-my-project",
    "created_at": "2025-10-26T16:00:00Z",
    "updated_at": "2025-10-26T16:00:00Z",
    "last_run_at": null,
    "owner_id": "user-456",
    "members": [
      {
        "user_id": "user-456",
        "role": "admin",
        "added_at": "2025-10-26T16:00:00Z"
      }
    ]
  }
}
```

**Error Responses**:
- `400 Bad Request` - Invalid input (validation error details in response)
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User not authorized to create projects
- `409 Conflict` - Project name already exists

---

### Delete Project

**Endpoint**: `DELETE /api/projects/{projectId}`

**Response** (204 No Content): Empty response

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User not authorized to delete project (owner only)
- `404 Not Found` - Project doesn't exist

---

## Pipeline Runs API

### List Pipeline Runs

**Endpoint**: `GET /api/projects/{projectId}/runs`

**Query Parameters**:
- `page`: integer (default: 1) - Page number
- `per_page`: integer (default: 20, max: 100) - Items per page
- `status`: enum(queued, running, succeeded, failed, cancelled, timeout) (optional) - Filter by status
- `branch`: string (optional) - Filter by branch name
- `search`: string (optional) - Search by commit SHA
- `sort`: enum(created_asc, created_desc) (default: created_desc) - Sort order
- `from_date`: ISO 8601 timestamp (optional) - Filter runs after date
- `to_date`: ISO 8601 timestamp (optional) - Filter runs before date

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
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
      "triggered_at": "2025-10-26T14:30:00Z",
      "started_at": "2025-10-26T14:30:15Z",
      "completed_at": "2025-10-26T14:35:45Z",
      "duration_seconds": 330,
      "error_message": null,
      "step_count": 5,
      "artifact_count": 2
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "per_page": 20
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access to project
- `404 Not Found` - Project doesn't exist

---

### Get Pipeline Run Detail

**Endpoint**: `GET /api/runs/{runId}`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
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
        "image": "golang:1.24",
        "commands": ["go test ./..."],
        "started_at": "2025-10-26T14:30:15Z",
        "completed_at": "2025-10-26T14:32:15Z",
        "duration_seconds": 120,
        "exit_code": 0,
        "error_output": null,
        "depends_on": [],
        "retry_count": 0,
        "max_retries": 3,
        "resource_usage": {
          "cpu_millicores": 800,
          "memory_bytes": 1073741824,
          "cpu_limit_millicores": 1000,
          "memory_limit_bytes": 2147483648
        }
      },
      {
        "id": "step-2",
        "name": "build",
        "status": "succeeded",
        "image": "golang:1.24",
        "commands": ["go build -o app"],
        "started_at": "2025-10-26T14:32:20Z",
        "completed_at": "2025-10-26T14:34:20Z",
        "duration_seconds": 120,
        "exit_code": 0,
        "error_output": null,
        "depends_on": ["test"],
        "retry_count": 0,
        "max_retries": 3,
        "resource_usage": {
          "cpu_millicores": 900,
          "memory_bytes": 1576341824,
          "cpu_limit_millicores": 1000,
          "memory_limit_bytes": 2147483648
        }
      }
    ],
    "artifacts": [
      {
        "id": "art-1",
        "name": "test-report.html",
        "type": "report",
        "mime_type": "text/html",
        "size_bytes": 245632,
        "url": "https://storage.example.com/artifacts/test-report.html?token=xyz",
        "created_at": "2025-10-26T14:32:16Z",
        "expires_at": "2025-11-26T14:32:16Z"
      }
    ]
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access to run
- `404 Not Found` - Run doesn't exist

---

## Logs API (Server-Sent Events)

### Stream Logs

**Endpoint**: `GET /api/runs/{runId}/steps/{stepId}/logs`

**Headers**: `Accept: text/event-stream`

**Query Parameters**:
- `follow`: boolean (default: true) - Follow live logs (stream until step completes)
- `lines`: integer (default: 100) - Number of previous lines to retrieve before following

**Response** (200 OK): Server-Sent Events stream

**SSE Message Format**:
```
event: log
data: {"timestamp":"2025-10-26T14:30:20.123Z","level":"info","message":"Running tests...","line_number":1,"source":"stdout"}

event: log
data: {"timestamp":"2025-10-26T14:30:25.456Z","level":"error","message":"Test failure: TestMain","line_number":2,"source":"stderr"}

event: complete
data: {"exit_code":1,"completed_at":"2025-10-26T14:32:15Z"}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access to run
- `404 Not Found` - Run or step doesn't exist
- `400 Bad Request` - Invalid parameters

**Client Implementation Example (HTMX)**:
```html
<div hx-ext="sse" sse-connect="/api/runs/c8s-abc123/steps/step-1/logs">
    <div id="log-container" sse-swap="log" hx-swap="beforeend">
        <div class="log-line">Waiting for logs...</div>
    </div>
</div>
```

---

### Get Logs Snapshot (Non-streaming)

**Endpoint**: `GET /api/runs/{runId}/steps/{stepId}/logs/snapshot`

**Query Parameters**:
- `lines`: integer (default: 100) - Number of lines to retrieve
- `offset`: integer (default: 0) - Line offset for pagination

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "step_id": "step-1",
    "status": "running",
    "lines": [
      {"line_number": 1, "timestamp": "2025-10-26T14:30:20Z", "level": "info", "message": "Starting test step", "source": "stdout"},
      {"line_number": 2, "timestamp": "2025-10-26T14:30:21Z", "level": "info", "message": "Running go test", "source": "stdout"},
      {"line_number": 3, "timestamp": "2025-10-26T14:30:25Z", "level": "error", "message": "Test failure", "source": "stderr"}
    ],
    "total_lines": 147,
    "is_complete": false
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access
- `404 Not Found` - Run or step doesn't exist

---

## Artifacts API

### List Artifacts

**Endpoint**: `GET /api/runs/{runId}/artifacts`

**Query Parameters**:
- `type`: enum(binary, report, documentation, log, other) (optional) - Filter by artifact type
- `step_id`: string (optional) - Filter artifacts from specific step

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
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
        "test_count": "1547",
        "passed": "1423",
        "failed": "124"
      }
    },
    {
      "id": "art-2",
      "pipeline_run_id": "c8s-abc123",
      "step_id": "step-2",
      "name": "app.tar.gz",
      "type": "binary",
      "mime_type": "application/gzip",
      "size_bytes": 52428800,
      "url": "https://storage.example.com/artifacts/app.tar.gz?token=abc",
      "created_at": "2025-10-26T14:34:21Z",
      "expires_at": "2025-11-26T14:34:21Z",
      "metadata": {}
    }
  ]
}
```

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access to run
- `404 Not Found` - Run doesn't exist

---

### Download Artifact

**Endpoint**: `GET /api/artifacts/{artifactId}/download`

**Response** (200 OK): Binary file content with appropriate `Content-Type` and `Content-Disposition` headers

**Error Responses**:
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - User doesn't have access
- `404 Not Found` - Artifact doesn't exist
- `410 Gone` - Artifact has expired/been deleted

---

## Real-Time Pipeline Updates (Server-Sent Events)

### Subscribe to Pipeline Updates

**Endpoint**: `GET /api/projects/{projectId}/runs/updates`

**Headers**: `Accept: text/event-stream`

**Query Parameters**:
- `status_filter`: comma-separated enum values (optional) - Only receive updates for specific statuses

**Response** (200 OK): Server-Sent Events stream

**SSE Message Format**:
```
event: run_created
data: {"id":"c8s-xyz789","status":"queued","branch":"feature/dashboard","commit_sha":"...","triggered_at":"2025-10-26T15:00:00Z"}

event: run_status_changed
data: {"id":"c8s-xyz789","old_status":"queued","new_status":"running","started_at":"2025-10-26T15:00:05Z"}

event: step_completed
data: {"run_id":"c8s-xyz789","step_id":"step-1","name":"test","status":"succeeded","duration_seconds":120,"exit_code":0}

event: run_completed
data: {"id":"c8s-xyz789","status":"succeeded","completed_at":"2025-10-26T15:05:00Z","duration_seconds":300}
```

**Use Case**: Keep the pipeline list up-to-date in real-time without polling.

**Client Implementation (HTMX)**:
```html
<div id="pipeline-list"
     hx-ext="sse"
     sse-connect="/api/projects/proj-123/runs/updates">
    <!-- Pipeline list items auto-update as SSE events arrive -->
</div>
```

---

## Webhook Management API

### Get Webhook Configuration

**Endpoint**: `GET /api/projects/{projectId}/webhook`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "project_id": "proj-123",
    "webhook_url": "https://c8s.example.com/webhook/github?projectId=proj-123",
    "webhook_secret": "wh_secret_abc123xyz..." (truncated for security),
    "created_at": "2025-10-01T12:00:00Z",
    "last_triggered_at": "2025-10-26T14:30:00Z",
    "git_platform": "github",
    "registration_instructions": "..."
  }
}
```

---

### Rotate Webhook Secret

**Endpoint**: `POST /api/projects/{projectId}/webhook/rotate-secret`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "webhook_url": "https://c8s.example.com/webhook/github?projectId=proj-123",
    "webhook_secret": "wh_secret_new_123xyz..." (new secret),
    "old_secret_expires_at": "2025-10-27T12:00:00Z"
  }
}
```

---

## Error Codes

Common error codes returned in error responses:

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `INVALID_REQUEST` | 400 | Request parameters invalid or malformed |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication token |
| `FORBIDDEN` | 403 | User doesn't have permission for this operation |
| `RESOURCE_NOT_FOUND` | 404 | Requested resource doesn't exist |
| `CONFLICT` | 409 | Request conflicts with existing resource (e.g., duplicate name) |
| `INTERNAL_ERROR` | 500 | Server-side error (rare) |
| `SERVICE_UNAVAILABLE` | 503 | Temporary service unavailability |

---

## Rate Limiting

- **Default limit**: 1000 requests per hour per user
- **Headers**:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Requests remaining this hour
  - `X-RateLimit-Reset`: UNIX timestamp when limit resets
- **Error Response** (429 Too Many Requests):
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "retry_after": 300
  }
}
```

---

## Pagination

All list endpoints support pagination:

**Request**:
```
GET /api/projects?page=2&per_page=50
```

**Response**:
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "total": 150,
    "page": 2,
    "per_page": 50,
    "total_pages": 3
  }
}
```

---

## API Versioning

Current API version: `v1`

All endpoints are prefixed with `/api/v1/`. If breaking changes are needed in the future, a new version (e.g., `/api/v2/`) will be introduced while maintaining backward compatibility with v1 for a deprecation period.

---

**API Schema Complete**. Ready for Quickstart guide and implementation tasks.
