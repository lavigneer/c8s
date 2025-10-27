# Implementation Tasks: C8S Web Dashboard

**Feature**: Web Dashboard for C8S CI Workflows
**Branch**: `004-create-a-front`
**Generated**: 2025-10-26
**Tech Stack**: Go 1.24.0 backend (html/template), HTMX frontend, Server-Sent Events

---

## Summary

**Total Tasks**: 67
**By Story**:
- Setup/Foundation: 15 tasks
- US1 (P1 - Pipeline History): 10 tasks
- US2 (P1 - Step Execution & Logs): 11 tasks
- US3 (P2 - Search & Filter): 8 tasks
- US4 (P2 - Projects & Webhooks): 10 tasks
- US5 (P3 - Artifacts): 8 tasks
- Polish/Cross-cutting: 5 tasks

**Parallel Opportunities**: Tasks marked with [P] can be executed in parallel within their phase.

**Estimated Timeline**:
- Phase 1 (Setup): 1-2 days
- Phase 2 (Foundation): 2-3 days
- Phase 3 (US1): 2-3 days
- Phase 4 (US2): 3-4 days
- Phase 5 (US3): 2 days
- Phase 6 (US4): 2-3 days
- Phase 7 (US5): 2 days
- Phase 8 (Polish): 1-2 days

**Total MVP (P1 stories only)**: ~8-12 days
**Full Implementation (P1 + P2 + P3)**: ~15-20 days

---

## Implementation Strategy

### MVP Scope (Phase 3-4)
The minimum viable product focuses on P1 stories:
1. **US1**: View pipeline history - Basic list with status
2. **US2**: Monitor execution & logs - Detail view with live streaming

This provides core value: developers can see pipeline status and debug failures.

### Phase 5-6 (P2 Stories)
Add productivity features:
3. **US3**: Search & filter - Find specific runs quickly
4. **US4**: Project configuration - Self-service project setup

### Phase 7 (P3 Stories)
Enhancement features:
5. **US5**: Artifact management - Download and view outputs

### Phase 8
Polish: Performance optimization, error handling, accessibility

---

## Dependency Diagram

```
Phase 1: Setup
├─ T001-T015 (Foundation infrastructure)
│
Phase 2: Foundation (Blocking for all stories)
├─ T016-T030 (Shared components, auth, API base)
│
Phase 3: US1 - Pipeline History [P1] ──┐
├─ T031-T040                            │
│                                       ├─ Can run in parallel
Phase 4: US2 - Logs & Execution [P1] ──┤
├─ T041-T051                            │
│                                       │
Phase 5: US3 - Search & Filter [P2] ───┤
├─ T052-T059 (depends on T031-T040)    │
│                                       │
Phase 6: US4 - Projects & Webhooks [P2]│
├─ T060-T069 (independent)             │
│                                       │
Phase 7: US5 - Artifacts [P3] ─────────┘
├─ T070-T077 (depends on T041-T051)
│
Phase 8: Polish & Cross-cutting
└─ T078-T082 (depends on all prior phases)
```

---

# Phase 1: Setup & Infrastructure (Foundation)

## T001: Create dashboard directory structure [P]
**Description**: Set up folder hierarchy for dashboard components within existing API server.

**Actions**:
```bash
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/handlers
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/templates/layout
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/templates/partials
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/templates/pages
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/static/css
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/static/js
mkdir -p /Users/elavigne/workspace/c8s/cmd/api-server/static/img
mkdir -p /Users/elavigne/workspace/c8s/pkg/dashboard
```

**Acceptance**: Directory structure exists and matches plan.md

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/` (directory)
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/` (directory tree)
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/` (directory tree)
- `/Users/elavigne/workspace/c8s/pkg/dashboard/` (directory)

---

## T002: Add HTMX library to static assets [P]
**Description**: Download HTMX library and SSE extension, place in static/js.

**Actions**:
1. Download HTMX 1.9.10+ from https://unpkg.com/htmx.org
2. Download SSE extension from https://unpkg.com/htmx.org/dist/ext/sse.js
3. Place in `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/`

**Acceptance**: HTMX files exist in static/js/ and are accessible.

**Dependencies**: T001
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/htmx.min.js`
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/sse.js`

---

## T003: Create base HTML template layout [P]
**Description**: Create base.html template with head, nav placeholder, and content block.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/layout/base.html`

**Content**:
```html
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}C8S Dashboard{{end}}</title>
    <link rel="stylesheet" href="/static/css/dashboard.css">
    <script src="/static/js/htmx.min.js"></script>
    <script src="/static/js/sse.js"></script>
</head>
<body>
    {{template "partials/nav" .}}
    <div class="container">
        {{block "content" .}}{{end}}
    </div>
</body>
</html>
{{end}}
```

**Acceptance**: Template parses without errors, includes HTMX scripts.

**Dependencies**: T001, T002
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/layout/base.html`

---

## T004: Create navigation partial template [P]
**Description**: Create nav.html partial with links to dashboard sections.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/nav.html`

**Content**:
```html
{{define "partials/nav"}}
<nav class="navbar">
    <div class="nav-container">
        <h1 class="nav-title"><a href="/dashboard">C8S Dashboard</a></h1>
        <ul class="nav-menu">
            <li><a href="/dashboard/projects">Projects</a></li>
            <li><a href="/dashboard">Pipelines</a></li>
            {{if .User}}
                <li><span>{{.User.Username}}</span></li>
                <li><a href="/logout">Logout</a></li>
            {{end}}
        </ul>
    </div>
</nav>
{{end}}
```

**Acceptance**: Navigation renders with project and pipeline links.

**Dependencies**: T001
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/nav.html`

---

## T005: Create base dashboard CSS [P]
**Description**: Create dashboard.css with basic styling for layout, nav, tables, and status badges.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

**Content**: Include styles for:
- Layout container (max-width: 1200px, centered)
- Navbar (background, padding, links)
- Status badges (green=succeeded, red=failed, orange=running)
- Tables (pipeline list styling)
- Buttons (primary, secondary, danger)
- Log viewer (monospace, dark background)

**Acceptance**: CSS loads and provides clean, functional UI styling.

**Dependencies**: T001
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T006: Create template loader utility [P]
**Description**: Create Go utility to parse templates once at startup with hot-reload support for development.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/templates.go`

**Functions**:
```go
package dashboard

import (
    "html/template"
    "path/filepath"
)

var Templates *template.Template

// LoadTemplates parses all templates from templates/ directory
func LoadTemplates(basePath string) error {
    pattern := filepath.Join(basePath, "templates/**/*.html")
    var err error
    Templates, err = template.ParseGlob(pattern)
    return err
}

// IsHTMXRequest checks if request is from HTMX (partial render)
func IsHTMXRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}
```

**Acceptance**: Templates load without errors, utility functions work.

**Dependencies**: T001, T003, T004
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/templates.go`

---

## T007: Create Kubernetes client wrapper [P]
**Description**: Create utility to interact with C8S CRDs (PipelineRun, PipelineConfig) via Kubernetes API.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/k8s_client.go`

**Functions**:
```go
package dashboard

import (
    "context"
    "github.com/org/c8s/pkg/apis/v1alpha1"
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sClient struct {
    client.Client
}

// ListPipelineRuns retrieves pipeline runs for a namespace
func (k *K8sClient) ListPipelineRuns(ctx context.Context, namespace string, opts ...client.ListOption) (*v1alpha1.PipelineRunList, error)

// GetPipelineRun retrieves a single pipeline run by name
func (k *K8sClient) GetPipelineRun(ctx context.Context, namespace, name string) (*v1alpha1.PipelineRun, error)

// GetPipelineConfig retrieves a pipeline config by name
func (k *K8sClient) GetPipelineConfig(ctx context.Context, namespace, name string) (*v1alpha1.PipelineConfig, error)
```

**Acceptance**: Client can list and get PipelineRuns from Kubernetes.

**Dependencies**: None (uses existing pkg/apis)
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/k8s_client.go`

---

## T008: Create DTO models for API responses [P]
**Description**: Define Go structs for dashboard API responses (differ from K8s CRDs).

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/models.go`

**Structs**:
```go
package dashboard

import "time"

// PipelineRunDTO represents a pipeline run for dashboard display
type PipelineRunDTO struct {
    ID              string    `json:"id"`
    ProjectID       string    `json:"project_id"`
    Name            string    `json:"name"`
    Status          string    `json:"status"` // queued, running, succeeded, failed
    CommitSHA       string    `json:"commit_sha"`
    Branch          string    `json:"branch"`
    Author          string    `json:"author"`
    AuthorEmail     string    `json:"author_email"`
    TriggerSource   string    `json:"trigger_source"`
    TriggeredAt     time.Time `json:"triggered_at"`
    StartedAt       *time.Time `json:"started_at,omitempty"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
    DurationSeconds *int64     `json:"duration_seconds,omitempty"`
    StepCount       int        `json:"step_count"`
    ArtifactCount   int        `json:"artifact_count"`
}

// StepDTO represents a pipeline step for dashboard display
type StepDTO struct {
    ID              string     `json:"id"`
    Name            string     `json:"name"`
    Status          string     `json:"status"`
    Image           string     `json:"image"`
    Commands        []string   `json:"commands"`
    StartedAt       *time.Time `json:"started_at,omitempty"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
    DurationSeconds *int64     `json:"duration_seconds,omitempty"`
    ExitCode        *int32     `json:"exit_code,omitempty"`
    DependsOn       []string   `json:"depends_on,omitempty"`
    LogURL          string     `json:"log_url,omitempty"`
}

// ProjectDTO represents a project for dashboard display
type ProjectDTO struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    RepoURL     string    `json:"repository_url"`
    WebhookURL  string    `json:"webhook_url"`
    Namespace   string    `json:"namespace"`
    CreatedAt   time.Time `json:"created_at"`
    LastRunAt   *time.Time `json:"last_run_at,omitempty"`
}

// ArtifactDTO represents an artifact for dashboard display
type ArtifactDTO struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Type      string    `json:"type"` // binary, report, documentation
    MimeType  string    `json:"mime_type"`
    SizeBytes int64     `json:"size_bytes"`
    URL       string    `json:"url"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Acceptance**: Models compile and can be serialized to JSON.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/models.go`

---

## T009: Create mapper functions (K8s CRD → DTO) [P]
**Description**: Transform Kubernetes CRD objects to dashboard DTOs.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/mappers.go`

**Functions**:
```go
package dashboard

import "github.com/org/c8s/pkg/apis/v1alpha1"

// MapPipelineRunToDTO converts K8s PipelineRun to PipelineRunDTO
func MapPipelineRunToDTO(run *v1alpha1.PipelineRun) *PipelineRunDTO

// MapStepStatusToDTO converts K8s StepStatus to StepDTO
func MapStepStatusToDTO(step *v1alpha1.StepStatus) *StepDTO

// CalculateDuration returns duration in seconds between start and completion
func CalculateDuration(start, end *metav1.Time) *int64
```

**Acceptance**: Mappers correctly transform CRD fields to DTO fields.

**Dependencies**: T008
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/mappers.go`

---

## T010: Create API response wrapper utilities
**Description**: Standardize API response format (success/error) for all endpoints.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/responses.go`

**Functions**:
```go
package dashboard

import (
    "encoding/json"
    "net/http"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Metadata   `json:"meta,omitempty"`
}

type APIError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Metadata struct {
    Total      int `json:"total,omitempty"`
    Page       int `json:"page,omitempty"`
    PerPage    int `json:"per_page,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

// WriteSuccess writes successful API response
func WriteSuccess(w http.ResponseWriter, data interface{}, meta *Metadata)

// WriteError writes error API response
func WriteError(w http.ResponseWriter, statusCode int, code, message string)
```

**Acceptance**: Response helpers consistently format JSON responses.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/responses.go`

---

## T011: Create authentication middleware
**Description**: Middleware to verify JWT/OAuth2 tokens and attach user context.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/auth_middleware.go`

**Functions**:
```go
package handlers

import (
    "context"
    "net/http"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware validates bearer token and attaches user to context
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // TODO: Validate token with existing C8S auth system
        user := &User{ID: "user-123", Username: "alice"}
        ctx := context.WithValue(r.Context(), userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*User, bool)
```

**Acceptance**: Middleware blocks unauthenticated requests and attaches user.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/auth_middleware.go`

---

## T012: Create static file server handler
**Description**: Serve static assets (CSS, JS, images) from /static route.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/static.go`

**Functions**:
```go
package handlers

import (
    "net/http"
    "path/filepath"
)

// ServeStatic serves static files from static/ directory
func ServeStatic(staticDir string) http.Handler {
    return http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
}
```

**Acceptance**: CSS and JS files accessible via /static/css/dashboard.css.

**Dependencies**: T001, T002, T005
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/static.go`

---

## T013: Update api-server main.go with dashboard routes
**Description**: Register dashboard routes in existing API server router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update existing)

**Actions**:
1. Import dashboard handlers package
2. Load templates at startup
3. Register static file handler
4. Register placeholder dashboard routes (will implement in later phases)
5. Apply auth middleware to dashboard routes

**Changes**:
```go
import (
    "github.com/org/c8s/cmd/api-server/handlers"
    "github.com/org/c8s/pkg/dashboard"
)

func main() {
    // ... existing setup ...

    // Load dashboard templates
    if err := dashboard.LoadTemplates("cmd/api-server"); err != nil {
        log.Fatal("Failed to load templates:", err)
    }

    router := chi.NewRouter()

    // Static files
    router.Handle("/static/*", handlers.ServeStatic("cmd/api-server/static"))

    // Dashboard routes (protected by auth)
    router.Group(func(r chi.Router) {
        r.Use(handlers.AuthMiddleware)
        r.Get("/dashboard", handlers.DashboardHandler)          // T031
        r.Get("/dashboard/projects", handlers.ProjectsHandler)  // T060
        r.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler) // T032
    })

    // ... existing routes ...

    http.ListenAndServe(":8080", router)
}
```

**Acceptance**: Server starts, static files accessible, auth middleware applied.

**Dependencies**: T001-T012
**Story**: Foundation
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (existing file, no new creation)

**Note**: Since cmd/api-server does not exist yet, may need to create it. If it exists, update accordingly.

---

## T014: Create error page template [P]
**Description**: Create generic error page template for 404, 500, etc.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/error.html`

**Content**:
```html
{{define "content"}}
<div class="error-container">
    <h1>{{.ErrorCode}}</h1>
    <p>{{.ErrorMessage}}</p>
    <a href="/dashboard" class="btn btn-primary">Back to Dashboard</a>
</div>
{{end}}
```

**Acceptance**: Error page renders with custom error code and message.

**Dependencies**: T001, T003
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/error.html`

---

## T015: Create integration test setup for dashboard [P]
**Description**: Set up test infrastructure for dashboard handlers using httptest.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/dashboard_test.go`

**Setup**:
```go
package integration

import (
    "net/http/httptest"
    "testing"
    "github.com/org/c8s/cmd/api-server/handlers"
    "github.com/org/c8s/pkg/dashboard"
)

// setupTestServer creates test HTTP server with dashboard routes
func setupTestServer(t *testing.T) *httptest.Server

// TestDashboardTemplatesLoad verifies templates parse without errors
func TestDashboardTemplatesLoad(t *testing.T)
```

**Acceptance**: Test setup compiles and can create test server.

**Dependencies**: T001-T013
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/dashboard_test.go`

---

# Phase 2: Foundation Components (Shared Infrastructure)

## T016: Create log storage interface
**Description**: Define interface for retrieving logs from S3/object storage.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/log_storage.go`

**Interface**:
```go
package dashboard

import "io"

type LogStorage interface {
    // GetStepLogs returns reader for step logs
    GetStepLogs(runID, stepID string) (io.ReadCloser, error)

    // StreamStepLogs streams logs line-by-line to channel
    StreamStepLogs(runID, stepID string, linesChan chan<- string) error

    // GetLogSnapshot returns last N lines of logs
    GetLogSnapshot(runID, stepID string, lines int) ([]string, error)
}

// S3LogStorage implements LogStorage using S3-compatible storage
type S3LogStorage struct {
    // ... S3 client fields ...
}
```

**Acceptance**: Interface compiles, S3 implementation skeleton exists.

**Dependencies**: None (uses existing pkg/storage/s3)
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/log_storage.go`

---

## T017: Implement S3LogStorage for log retrieval
**Description**: Connect LogStorage interface to existing S3 storage backend.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/log_storage.go` (continue)

**Implementation**:
```go
func NewS3LogStorage(s3Client *storage.S3Client) *S3LogStorage

func (s *S3LogStorage) GetStepLogs(runID, stepID string) (io.ReadCloser, error) {
    objectKey := fmt.Sprintf("logs/%s/%s.log", runID, stepID)
    return s.s3Client.GetObject(objectKey)
}

func (s *S3LogStorage) StreamStepLogs(runID, stepID string, linesChan chan<- string) error {
    reader, err := s.GetStepLogs(runID, stepID)
    if err != nil {
        return err
    }
    defer reader.Close()

    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        linesChan <- scanner.Text()
    }
    return scanner.Err()
}
```

**Acceptance**: S3LogStorage retrieves logs from object storage successfully.

**Dependencies**: T016
**Story**: Foundation
**Files Modified**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/log_storage.go`

---

## T018: Create SSE broadcaster utility
**Description**: Implement pub/sub mechanism for broadcasting SSE events to multiple clients.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/sse_broadcaster.go`

**Functions**:
```go
package dashboard

import "sync"

type SSEBroadcaster struct {
    clients   map[chan string]bool
    mutex     sync.RWMutex
}

func NewSSEBroadcaster() *SSEBroadcaster

// Subscribe adds client channel to receive broadcasts
func (b *SSEBroadcaster) Subscribe() chan string

// Unsubscribe removes client channel
func (b *SSEBroadcaster) Unsubscribe(ch chan string)

// Broadcast sends message to all subscribed clients
func (b *SSEBroadcaster) Broadcast(message string)
```

**Acceptance**: Multiple clients can subscribe and receive broadcast messages.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/sse_broadcaster.go`

---

## T019: Create pagination utility
**Description**: Utility functions for paginating lists (pipeline runs, artifacts).

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/pagination.go`

**Functions**:
```go
package dashboard

type PaginationParams struct {
    Page    int
    PerPage int
}

type PaginatedResult struct {
    Items      interface{}
    Total      int
    Page       int
    PerPage    int
    TotalPages int
}

// ParsePaginationParams extracts pagination from query params
func ParsePaginationParams(r *http.Request) PaginationParams

// Paginate applies pagination to a slice
func Paginate(items interface{}, params PaginationParams) *PaginatedResult
```

**Acceptance**: Pagination correctly slices arrays and returns metadata.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/pagination.go`

---

## T020: Create time formatting utilities
**Description**: Template functions for formatting timestamps and durations.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/time_utils.go`

**Functions**:
```go
package dashboard

import "time"

// FormatTimestamp formats time.Time to "2 hours ago" or "Jan 2, 15:04"
func FormatTimestamp(t time.Time) string

// FormatDuration formats duration in seconds to "2m 30s" or "1h 15m"
func FormatDuration(seconds int64) string

// IsRecent returns true if timestamp is within last hour
func IsRecent(t time.Time) bool
```

**Acceptance**: Time utilities format timestamps and durations correctly.

**Dependencies**: None
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/time_utils.go`

---

## T021: Register template functions for formatting
**Description**: Register custom template functions (formatTime, formatDuration) with html/template.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/templates.go` (update)

**Changes**:
```go
func LoadTemplates(basePath string) error {
    funcMap := template.FuncMap{
        "formatTime":     FormatTimestamp,
        "formatDuration": FormatDuration,
        "slice":          func(s string, start, end int) string { return s[start:end] },
        "eq":             func(a, b string) bool { return a == b },
    }

    pattern := filepath.Join(basePath, "templates/**/*.html")
    var err error
    Templates, err = template.New("").Funcs(funcMap).ParseGlob(pattern)
    return err
}
```

**Acceptance**: Templates can use custom functions (formatTime, eq, slice).

**Dependencies**: T006, T020
**Story**: Foundation
**Files Modified**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/templates.go`

---

## T022: Create status badge partial template [P]
**Description**: Reusable component for displaying status badges (running/succeeded/failed).

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/status_badge.html`

**Content**:
```html
{{define "partials/status_badge"}}
{{if eq .Status "running"}}
    <span class="badge badge-running">⏳ Running</span>
{{else if eq .Status "succeeded"}}
    <span class="badge badge-success">✓ Succeeded</span>
{{else if eq .Status "failed"}}
    <span class="badge badge-error">✗ Failed</span>
{{else if eq .Status "pending"}}
    <span class="badge badge-pending">⏸ Pending</span>
{{else if eq .Status "cancelled"}}
    <span class="badge badge-cancelled">⊘ Cancelled</span>
{{end}}
{{end}}
```

**Acceptance**: Badge partial renders appropriate status with icon and color.

**Dependencies**: T001, T005
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/status_badge.html`

---

## T023: Create loading spinner partial template [P]
**Description**: Reusable loading spinner for async HTMX requests.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/loading.html`

**Content**:
```html
{{define "partials/loading"}}
<div class="loading-spinner">
    <div class="spinner"></div>
    <span>{{if .Message}}{{.Message}}{{else}}Loading...{{end}}</span>
</div>
{{end}}
```

**Acceptance**: Loading spinner displays during HTMX requests.

**Dependencies**: T001, T005
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/loading.html`

---

## T024: Create empty state partial template [P]
**Description**: Reusable empty state component for when no data exists.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/empty_state.html`

**Content**:
```html
{{define "partials/empty_state"}}
<div class="empty-state">
    <div class="empty-icon">📭</div>
    <h3>{{.Title}}</h3>
    <p>{{.Message}}</p>
    {{if .ActionText}}
        <a href="{{.ActionURL}}" class="btn btn-primary">{{.ActionText}}</a>
    {{end}}
</div>
{{end}}
```

**Acceptance**: Empty state renders with custom title, message, and action.

**Dependencies**: T001, T005
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/empty_state.html`

---

## T025: Write unit tests for mappers
**Description**: Test K8s CRD → DTO transformation functions.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/dashboard_mappers_test.go`

**Tests**:
```go
func TestMapPipelineRunToDTO(t *testing.T)
func TestMapStepStatusToDTO(t *testing.T)
func TestCalculateDuration(t *testing.T)
```

**Acceptance**: All mapper tests pass, coverage >80%.

**Dependencies**: T009
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/dashboard_mappers_test.go`

---

## T026: Write unit tests for pagination
**Description**: Test pagination utility functions.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/pagination_test.go`

**Tests**:
```go
func TestParsePaginationParams(t *testing.T)
func TestPaginate(t *testing.T)
func TestPaginateEmptyList(t *testing.T)
```

**Acceptance**: Pagination tests pass, handles edge cases (empty lists, invalid pages).

**Dependencies**: T019
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/pagination_test.go`

---

## T027: Write unit tests for SSE broadcaster
**Description**: Test SSE pub/sub mechanism.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/sse_broadcaster_test.go`

**Tests**:
```go
func TestSSEBroadcaster_Subscribe(t *testing.T)
func TestSSEBroadcaster_Broadcast(t *testing.T)
func TestSSEBroadcaster_Unsubscribe(t *testing.T)
```

**Acceptance**: SSE broadcaster tests pass, no race conditions.

**Dependencies**: T018
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/sse_broadcaster_test.go`

---

## T028: Write unit tests for time utilities
**Description**: Test timestamp and duration formatting functions.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/time_utils_test.go`

**Tests**:
```go
func TestFormatTimestamp(t *testing.T)
func TestFormatDuration(t *testing.T)
func TestIsRecent(t *testing.T)
```

**Acceptance**: Time utility tests pass, covers various time ranges.

**Dependencies**: T020
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/time_utils_test.go`

---

## T029: Write integration test for auth middleware
**Description**: Test authentication middleware blocks unauthenticated requests.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/auth_test.go`

**Tests**:
```go
func TestAuthMiddleware_BlocksUnauthenticated(t *testing.T)
func TestAuthMiddleware_AllowsAuthenticated(t *testing.T)
func TestAuthMiddleware_AttachesUserToContext(t *testing.T)
```

**Acceptance**: Auth middleware tests pass, correctly blocks/allows requests.

**Dependencies**: T011
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/auth_test.go`

---

## T030: Write integration test for static file serving
**Description**: Test static assets are served correctly from /static route.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/static_files_test.go`

**Tests**:
```go
func TestStaticFiles_ServesCSS(t *testing.T)
func TestStaticFiles_ServesJS(t *testing.T)
func TestStaticFiles_Returns404ForMissingFile(t *testing.T)
```

**Acceptance**: Static file tests pass, files accessible via HTTP.

**Dependencies**: T012
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/static_files_test.go`

---

# Phase 3: US1 - View Pipeline History and Current Status [P1]

## T031: Create pipeline list page template
**Description**: HTML template for main dashboard page listing pipeline runs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

**Content**: (Reference quickstart.md example)
- Filter controls (search, branch, status)
- HTMX-enhanced table with SSE updates
- Pagination controls
- Empty state if no runs

**Acceptance**: Template renders with dummy data, HTMX attributes present.

**Dependencies**: T003, T022, T024
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

---

## T032: Implement ListPipelineRuns API handler
**Description**: Go handler for GET /api/projects/{projectId}/runs endpoint.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

**Function**:
```go
package handlers

// ListPipelineRunsHandler handles GET /api/projects/{projectId}/runs
func ListPipelineRunsHandler(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectId")

    // Parse query params (page, per_page, status, branch, search)
    params := dashboard.ParsePaginationParams(r)
    status := r.URL.Query().Get("status")
    branch := r.URL.Query().Get("branch")
    search := r.URL.Query().Get("search")

    // Query Kubernetes for PipelineRuns
    runs, err := fetchPipelineRuns(r.Context(), projectID, status, branch, search)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
        return
    }

    // Transform to DTOs
    dtos := make([]*dashboard.PipelineRunDTO, len(runs))
    for i, run := range runs {
        dtos[i] = dashboard.MapPipelineRunToDTO(run)
    }

    // Paginate
    result := dashboard.Paginate(dtos, params)

    // Check if HTMX request
    if dashboard.IsHTMXRequest(r) {
        // Return fragment (pipeline rows only)
        dashboard.Templates.ExecuteTemplate(w, "pipeline_list_rows", result)
    } else {
        // Return full page
        dashboard.WriteSuccess(w, result.Items, &dashboard.Metadata{
            Total:      result.Total,
            Page:       result.Page,
            PerPage:    result.PerPage,
            TotalPages: result.TotalPages,
        })
    }
}

// fetchPipelineRuns queries Kubernetes for pipeline runs with filters
func fetchPipelineRuns(ctx context.Context, projectID, status, branch, search string) ([]*v1alpha1.PipelineRun, error)
```

**Acceptance**: Endpoint returns paginated pipeline runs in JSON or HTML.

**Dependencies**: T007, T008, T009, T010, T019
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T033: Create DashboardHandler for main dashboard page
**Description**: Handler for GET /dashboard (renders full page with pipeline list).

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

**Function**:
```go
package handlers

// DashboardHandler renders main dashboard page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUserFromContext(r.Context())
    if !ok {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        return
    }

    // Get user's first project (or from query param)
    projectID := r.URL.Query().Get("projectId")
    if projectID == "" {
        projects, err := fetchUserProjects(r.Context(), user.ID)
        if err != nil || len(projects) == 0 {
            dashboard.Templates.ExecuteTemplate(w, "base", map[string]interface{}{
                "User":  user,
                "Error": "No projects found",
            })
            return
        }
        projectID = projects[0].ID
    }

    // Fetch initial pipeline runs (first page)
    runs, err := fetchPipelineRuns(r.Context(), projectID, "", "", "")
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
        return
    }

    dtos := make([]*dashboard.PipelineRunDTO, len(runs))
    for i, run := range runs {
        dtos[i] = dashboard.MapPipelineRunToDTO(run)
    }

    // Render full page
    dashboard.Templates.ExecuteTemplate(w, "base", map[string]interface{}{
        "User":         user,
        "ProjectID":    projectID,
        "PipelineRuns": dtos,
    })
}
```

**Acceptance**: /dashboard renders full page with pipeline list.

**Dependencies**: T031, T032
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T034: Implement SSE endpoint for pipeline status updates
**Description**: SSE endpoint GET /api/projects/{projectId}/runs/updates streams real-time updates.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_sse.go`

**Function**:
```go
package handlers

// PipelineUpdatesSSEHandler streams pipeline status updates
func PipelineUpdatesSSEHandler(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectId")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    // Subscribe to pipeline updates
    updateChan := subscribeToPipelineUpdates(projectID)
    defer unsubscribeFromPipelineUpdates(projectID, updateChan)

    for {
        select {
        case <-r.Context().Done():
            return
        case update := <-updateChan:
            // Format SSE message
            eventType := "run_status_changed"
            data, _ := json.Marshal(update)
            fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
            flusher.Flush()
        }
    }
}

// subscribeToPipelineUpdates watches Kubernetes for PipelineRun changes
func subscribeToPipelineUpdates(projectID string) chan *dashboard.PipelineRunDTO
```

**Acceptance**: SSE endpoint streams pipeline status changes in real-time.

**Dependencies**: T018, T032
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_sse.go`

---

## T035: Create pipeline row partial template [P]
**Description**: Reusable template for single pipeline row (used by HTMX swaps).

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/pipeline_row.html`

**Content**:
```html
{{define "partials/pipeline_row"}}
<div class="pipeline-row" id="pipeline-{{.ID}}">
    <div class="pipeline-info">
        <div class="pipeline-name">{{.Name}}</div>
        <div class="pipeline-meta">
            <span class="commit">{{.CommitSHA | slice 0 7}}</span>
            <span class="branch">{{.Branch}}</span>
            <span class="author">{{.Author}}</span>
        </div>
    </div>
    <div class="pipeline-status">
        {{template "partials/status_badge" .}}
    </div>
    <div class="pipeline-time">
        <span>{{.TriggeredAt | formatTime}}</span>
        {{if .DurationSeconds}}
            <span>{{.DurationSeconds | formatDuration}}</span>
        {{end}}
    </div>
    <div class="pipeline-actions">
        <a href="/dashboard/runs/{{.ID}}" class="btn btn-sm">View</a>
    </div>
</div>
{{end}}
```

**Acceptance**: Pipeline row renders with status, commit info, and actions.

**Dependencies**: T001, T022
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/pipeline_row.html`

---

## T036: Write integration test for ListPipelineRuns handler
**Description**: Test pipeline list API endpoint with various filters.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/pipeline_list_test.go`

**Tests**:
```go
func TestListPipelineRuns_ReturnsRuns(t *testing.T)
func TestListPipelineRuns_FiltersBy Status(t *testing.T)
func TestListPipelineRuns_FiltersByBranch(t *testing.T)
func TestListPipelineRuns_SearchByCommitSHA(t *testing.T)
func TestListPipelineRuns_Pagination(t *testing.T)
```

**Acceptance**: All list endpoint tests pass, filtering and pagination work.

**Dependencies**: T032
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_list_test.go`

---

## T037: Write E2E test for pipeline list page
**Description**: Playwright test verifying pipeline list displays and updates.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_list.spec.ts`

**Tests**:
```typescript
test('pipeline list displays runs', async ({ page }) => {})
test('pipeline list updates via SSE', async ({ page }) => {})
test('search filters pipeline list', async ({ page }) => {})
test('status filter updates list', async ({ page }) => {})
```

**Acceptance**: E2E tests pass, pipeline list interactive and updates in real-time.

**Dependencies**: T031, T032, T033, T034
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_list.spec.ts`

---

## T038: Add CSS styling for pipeline list
**Description**: Style pipeline list table, filters, and status badges.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (update)

**Styles**:
- `.pipeline-row`: Grid layout with hover effect
- `.pipeline-meta`: Gray text, smaller font
- `.badge-*`: Status badge colors (green, red, orange)
- `.filters`: Flex layout for search/filter controls

**Acceptance**: Pipeline list visually appealing and responsive.

**Dependencies**: T005, T031
**Story**: US1
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T039: Implement Kubernetes watch for pipeline updates
**Description**: Watch Kubernetes PipelineRun resources and broadcast changes via SSE.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/pipeline_watcher.go`

**Functions**:
```go
package dashboard

import (
    "context"
    "github.com/org/c8s/pkg/apis/v1alpha1"
    "k8s.io/apimachinery/pkg/watch"
)

type PipelineWatcher struct {
    k8sClient     *K8sClient
    broadcasters  map[string]*SSEBroadcaster // projectID -> broadcaster
}

// Start begins watching PipelineRuns and broadcasting updates
func (w *PipelineWatcher) Start(ctx context.Context, namespace string) error

// handleWatchEvent processes Kubernetes watch events (Added, Modified, Deleted)
func (w *PipelineWatcher) handleWatchEvent(event watch.Event)
```

**Acceptance**: Watcher detects PipelineRun changes and broadcasts to SSE clients.

**Dependencies**: T007, T018, T034
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/pipeline_watcher.go`

---

## T040: Register US1 routes in main.go
**Description**: Add all US1 routes to API server router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update)

**Routes**:
```go
router.Group(func(r chi.Router) {
    r.Use(handlers.AuthMiddleware)
    r.Get("/dashboard", handlers.DashboardHandler)
    r.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler)
    r.Get("/api/projects/{projectId}/runs/updates", handlers.PipelineUpdatesSSEHandler)
})
```

**Acceptance**: All US1 routes registered and accessible.

**Dependencies**: T032, T033, T034
**Story**: US1
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 4: US2 - Monitor Step-by-Step Execution and Logs [P1]

## T041: Create pipeline detail page template
**Description**: Template for detailed pipeline run view with steps and logs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

**Content**:
- Pipeline run metadata (commit, branch, status, duration)
- Step list with status, duration, dependencies
- Log viewer section (SSE-streamed logs)
- Step DAG visualization (optional)

**Acceptance**: Template renders with dummy data, log viewer HTMX attributes present.

**Dependencies**: T003, T022
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

---

## T042: Implement GetPipelineRun API handler
**Description**: Handler for GET /api/runs/{runId} returning full run details with steps.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go` (continue)

**Function**:
```go
// GetPipelineRunHandler handles GET /api/runs/{runId}
func GetPipelineRunHandler(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runId")

    // Fetch PipelineRun from Kubernetes
    run, err := k8sClient.GetPipelineRun(r.Context(), namespace, runID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "RUN_NOT_FOUND", "Pipeline run not found")
        return
    }

    // Map to DTO
    dto := dashboard.MapPipelineRunToDTO(run)

    // Map steps
    dto.Steps = make([]*dashboard.StepDTO, len(run.Status.Steps))
    for i, step := range run.Status.Steps {
        dto.Steps[i] = dashboard.MapStepStatusToDTO(&step)
    }

    dashboard.WriteSuccess(w, dto, nil)
}
```

**Acceptance**: Endpoint returns full pipeline run with step details.

**Dependencies**: T007, T008, T009
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T043: Create PipelineDetailHandler for detail page
**Description**: Handler for GET /dashboard/runs/{runId} rendering detail page.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go` (continue)

**Function**:
```go
// PipelineDetailHandler renders pipeline run detail page
func PipelineDetailHandler(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runId")

    // Fetch run details
    run, err := k8sClient.GetPipelineRun(r.Context(), namespace, runID)
    if err != nil {
        renderErrorPage(w, 404, "Pipeline run not found")
        return
    }

    dto := dashboard.MapPipelineRunToDTO(run)
    dto.Steps = make([]*dashboard.StepDTO, len(run.Status.Steps))
    for i, step := range run.Status.Steps {
        dto.Steps[i] = dashboard.MapStepStatusToDTO(&step)
    }

    // Render full page
    dashboard.Templates.ExecuteTemplate(w, "base", map[string]interface{}{
        "User":        GetUserFromContext(r.Context()),
        "PipelineRun": dto,
    })
}
```

**Acceptance**: /dashboard/runs/{runId} renders detail page with steps.

**Dependencies**: T041, T042
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T044: Implement SSE log streaming endpoint
**Description**: SSE endpoint GET /api/runs/{runId}/steps/{stepId}/logs streams live logs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/logs.go`

**Function**:
```go
package handlers

// LogStreamHandler streams logs via SSE
func LogStreamHandler(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runId")
    stepID := chi.URLParam(r, "stepId")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    // Get log storage
    logStorage := dashboard.NewS3LogStorage(s3Client)

    // Stream logs
    logChan := make(chan string, 100)
    errChan := make(chan error, 1)

    go func() {
        err := logStorage.StreamStepLogs(runID, stepID, logChan)
        if err != nil {
            errChan <- err
        }
        close(logChan)
    }()

    for {
        select {
        case <-r.Context().Done():
            return
        case err := <-errChan:
            fmt.Fprintf(w, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
            flusher.Flush()
            return
        case line, ok := <-logChan:
            if !ok {
                fmt.Fprintf(w, "event: complete\ndata: {}\n\n")
                flusher.Flush()
                return
            }
            logEntry := map[string]interface{}{
                "message":   line,
                "timestamp": time.Now().Format(time.RFC3339),
            }
            data, _ := json.Marshal(logEntry)
            fmt.Fprintf(w, "event: log\ndata: %s\n\n", data)
            flusher.Flush()
        }
    }
}
```

**Acceptance**: Logs stream in real-time via SSE to client.

**Dependencies**: T016, T017, T018
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/logs.go`

---

## T045: Create log viewer partial template [P]
**Description**: HTMX-enhanced log viewer component with SSE streaming.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/log_viewer.html`

**Content**:
```html
{{define "partials/log_viewer"}}
<div class="log-viewer"
     hx-ext="sse"
     sse-connect="/api/runs/{{.RunID}}/steps/{{.StepID}}/logs">

    <div class="log-header">
        <h3>Logs: {{.StepName}}</h3>
        <button class="btn btn-sm" onclick="clearLogs()">Clear</button>
    </div>

    <div class="log-container" id="log-container-{{.StepID}}" sse-swap="log" hx-swap="beforeend">
        <div class="log-line loading">Waiting for logs...</div>
    </div>
</div>

<script>
function clearLogs() {
    document.getElementById('log-container-{{.StepID}}').innerHTML = '';
}
</script>
{{end}}
```

**Acceptance**: Log viewer displays streaming logs with auto-scroll.

**Dependencies**: T001, T044
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/log_viewer.html`

---

## T046: Create step status partial template [P]
**Description**: Reusable component for displaying step status with details.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/step_status.html`

**Content**:
```html
{{define "partials/step_status"}}
<div class="step-item" id="step-{{.ID}}" data-status="{{.Status}}">
    <div class="step-header">
        <div class="step-name">
            {{template "partials/status_badge" .}}
            <span>{{.Name}}</span>
        </div>
        <div class="step-time">
            {{if .StartedAt}}
                Started: {{.StartedAt | formatTime}}
            {{end}}
            {{if .DurationSeconds}}
                ({{.DurationSeconds | formatDuration}})
            {{end}}
        </div>
    </div>

    <div class="step-details">
        <div class="step-image">Image: {{.Image}}</div>
        {{if .Commands}}
            <div class="step-commands">
                <strong>Commands:</strong>
                {{range .Commands}}
                    <code>{{.}}</code>
                {{end}}
            </div>
        {{end}}
        {{if .ExitCode}}
            <div class="step-exit-code">Exit Code: {{.ExitCode}}</div>
        {{end}}
    </div>

    <button class="btn btn-sm"
            hx-get="/api/runs/{{.PipelineRunID}}/steps/{{.ID}}/logs"
            hx-target="#log-section"
            hx-swap="innerHTML">
        View Logs
    </button>
</div>
{{end}}
```

**Acceptance**: Step status displays with expand/collapse and view logs action.

**Dependencies**: T001, T022
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/step_status.html`

---

## T047: Write integration test for GetPipelineRun handler
**Description**: Test pipeline detail API endpoint.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/pipeline_detail_test.go`

**Tests**:
```go
func TestGetPipelineRun_ReturnsRun(t *testing.T)
func TestGetPipelineRun_ReturnsSteps(t *testing.T)
func TestGetPipelineRun_Returns404ForInvalidID(t *testing.T)
```

**Acceptance**: Detail endpoint tests pass, returns run with steps.

**Dependencies**: T042
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_detail_test.go`

---

## T048: Write integration test for log streaming endpoint
**Description**: Test SSE log streaming.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/log_stream_test.go`

**Tests**:
```go
func TestLogStream_StreamsLogs(t *testing.T)
func TestLogStream_ReturnsErrorForInvalidStep(t *testing.T)
func TestLogStream_CompletesWhenLogsDone(t *testing.T)
```

**Acceptance**: Log streaming tests pass, SSE events received correctly.

**Dependencies**: T044
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/log_stream_test.go`

---

## T049: Write E2E test for pipeline detail page
**Description**: Playwright test for pipeline detail view with live logs.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_detail.spec.ts`

**Tests**:
```typescript
test('pipeline detail displays steps', async ({ page }) => {})
test('logs stream in real-time', async ({ page }) => {})
test('step status updates dynamically', async ({ page }) => {})
test('clicking step shows logs', async ({ page }) => {})
```

**Acceptance**: E2E tests pass, logs stream and steps update.

**Dependencies**: T041, T042, T043, T044
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_detail.spec.ts`

---

## T050: Add CSS styling for pipeline detail page
**Description**: Style step list, log viewer, and detail layout.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (update)

**Styles**:
- `.step-item`: Card layout with border, padding
- `.log-viewer`: Dark background, monospace font, auto-scroll
- `.log-line`: Individual log line styling
- `.step-details`: Collapsible section styling

**Acceptance**: Pipeline detail page visually organized and readable.

**Dependencies**: T005, T041, T045, T046
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T051: Register US2 routes in main.go
**Description**: Add all US2 routes to API server router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update)

**Routes**:
```go
router.Group(func(r chi.Router) {
    r.Use(handlers.AuthMiddleware)
    r.Get("/dashboard/runs/{runId}", handlers.PipelineDetailHandler)
    r.Get("/api/runs/{runId}", handlers.GetPipelineRunHandler)
    r.Get("/api/runs/{runId}/steps/{stepId}/logs", handlers.LogStreamHandler)
})
```

**Acceptance**: All US2 routes registered and accessible.

**Dependencies**: T042, T043, T044
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 5: US3 - Search and Filter Pipelines [P2]

## T052: Create filter panel partial template [P]
**Description**: Reusable filter controls for search, branch, status, date range.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/filter_panel.html`

**Content**:
```html
{{define "partials/filter_panel"}}
<div class="filter-panel">
    <div class="filter-group">
        <label>Search</label>
        <input type="text"
               name="search"
               placeholder="Commit SHA..."
               hx-get="/api/projects/{{.ProjectID}}/runs"
               hx-target="#pipeline-list"
               hx-trigger="keyup changed delay:500ms"
               hx-include=".filter-panel [name]">
    </div>

    <div class="filter-group">
        <label>Branch</label>
        <select name="branch"
                hx-get="/api/projects/{{.ProjectID}}/runs"
                hx-target="#pipeline-list"
                hx-trigger="change"
                hx-include=".filter-panel [name]">
            <option value="">All Branches</option>
            {{range .Branches}}
                <option value="{{.}}">{{.}}</option>
            {{end}}
        </select>
    </div>

    <div class="filter-group">
        <label>Status</label>
        <select name="status"
                hx-get="/api/projects/{{.ProjectID}}/runs"
                hx-target="#pipeline-list"
                hx-trigger="change"
                hx-include=".filter-panel [name]">
            <option value="">All Statuses</option>
            <option value="succeeded">Succeeded</option>
            <option value="failed">Failed</option>
            <option value="running">Running</option>
            <option value="pending">Pending</option>
        </select>
    </div>

    <div class="filter-group">
        <label>Date Range</label>
        <input type="date"
               name="from_date"
               hx-get="/api/projects/{{.ProjectID}}/runs"
               hx-target="#pipeline-list"
               hx-trigger="change"
               hx-include=".filter-panel [name]">
        <span>to</span>
        <input type="date"
               name="to_date"
               hx-get="/api/projects/{{.ProjectID}}/runs"
               hx-target="#pipeline-list"
               hx-trigger="change"
               hx-include=".filter-panel [name]">
    </div>

    <button class="btn btn-secondary" onclick="clearFilters()">Clear Filters</button>
</div>
{{end}}
```

**Acceptance**: Filter panel renders with all filter controls.

**Dependencies**: T001
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/filter_panel.html`

---

## T053: Update ListPipelineRuns handler with filter logic
**Description**: Enhance handler to support all filter parameters.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go` (update)

**Changes**:
```go
func ListPipelineRunsHandler(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectId")

    // Parse filters
    filters := parseFilters(r)

    // Query Kubernetes with filters
    runs, err := fetchPipelineRunsWithFilters(r.Context(), projectID, filters)
    // ... rest of implementation ...
}

type PipelineFilters struct {
    Status    string
    Branch    string
    Search    string
    FromDate  time.Time
    ToDate    time.Time
}

func parseFilters(r *http.Request) PipelineFilters {
    // Parse query parameters into filters struct
}

func fetchPipelineRunsWithFilters(ctx context.Context, projectID string, filters PipelineFilters) ([]*v1alpha1.PipelineRun, error) {
    // Query Kubernetes with label selectors and filters
}
```

**Acceptance**: Filter parameters correctly filter pipeline runs.

**Dependencies**: T032, T052
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T054: Implement branch list API endpoint
**Description**: Endpoint GET /api/projects/{projectId}/branches returns unique branches.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go` (continue)

**Function**:
```go
// ListBranchesHandler returns unique branch names for project
func ListBranchesHandler(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectId")

    // Query Kubernetes for unique branches
    runs, err := k8sClient.ListPipelineRuns(r.Context(), namespace)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
        return
    }

    // Extract unique branches
    branchSet := make(map[string]bool)
    for _, run := range runs.Items {
        if run.Spec.Branch != "" {
            branchSet[run.Spec.Branch] = true
        }
    }

    branches := make([]string, 0, len(branchSet))
    for branch := range branchSet {
        branches = append(branches, branch)
    }

    dashboard.WriteSuccess(w, branches, nil)
}
```

**Acceptance**: Endpoint returns list of unique branches for project.

**Dependencies**: T007
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T055: Add filter panel to pipeline list template
**Description**: Integrate filter panel into main pipeline list page.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html` (update)

**Changes**:
```html
{{define "content"}}
<h2>Pipeline Runs</h2>

<!-- Add filter panel -->
{{template "partials/filter_panel" .}}

<!-- Existing pipeline list -->
<div id="pipeline-list" ...>
    ...
</div>
{{end}}
```

**Acceptance**: Filter panel displays above pipeline list.

**Dependencies**: T031, T052
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

---

## T056: Write unit tests for filter parsing
**Description**: Test filter parameter parsing and validation.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/filters_test.go`

**Tests**:
```go
func TestParseFilters_ParsesAllParameters(t *testing.T)
func TestParseFilters_HandlesInvalidDates(t *testing.T)
func TestParseFilters_HandlesEmptyFilters(t *testing.T)
```

**Acceptance**: Filter parsing tests pass, handles invalid input.

**Dependencies**: T053
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/filters_test.go`

---

## T057: Write integration test for filtered pipeline list
**Description**: Test pipeline list endpoint with all filter combinations.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/pipeline_filters_test.go`

**Tests**:
```go
func TestFilteredPipelineList_ByStatus(t *testing.T)
func TestFilteredPipelineList_ByBranch(t *testing.T)
func TestFilteredPipelineList_ByCommitSHA(t *testing.T)
func TestFilteredPipelineList_ByDateRange(t *testing.T)
func TestFilteredPipelineList_CombinedFilters(t *testing.T)
```

**Acceptance**: All filter combinations return correct results.

**Dependencies**: T053
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_filters_test.go`

---

## T058: Write E2E test for search and filter
**Description**: Playwright test for interactive filtering.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_filters.spec.ts`

**Tests**:
```typescript
test('search filters by commit SHA', async ({ page }) => {})
test('branch filter updates list', async ({ page }) => {})
test('status filter updates list', async ({ page }) => {})
test('clear filters resets list', async ({ page }) => {})
test('combined filters work together', async ({ page }) => {})
```

**Acceptance**: E2E tests pass, filters interactive and responsive.

**Dependencies**: T052, T053, T055
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_filters.spec.ts`

---

## T059: Register US3 routes in main.go
**Description**: Add branch list endpoint to router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update)

**Routes**:
```go
router.Group(func(r chi.Router) {
    r.Use(handlers.AuthMiddleware)
    r.Get("/api/projects/{projectId}/branches", handlers.ListBranchesHandler)
})
```

**Acceptance**: Branch list endpoint accessible.

**Dependencies**: T054
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 6: US4 - Configure Projects and Webhooks [P2]

## T060: Create projects list page template
**Description**: Template for listing user's projects with webhook URLs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/projects.html`

**Content**:
- List of projects with name, repo URL, last run time
- Webhook URL display with copy-to-clipboard button
- Create project button (opens modal/form)
- Delete project action

**Acceptance**: Template renders with project list and actions.

**Dependencies**: T003, T024
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/projects.html`

---

## T061: Implement ListProjects API handler
**Description**: Handler for GET /api/projects returning user's projects.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

**Function**:
```go
package handlers

// ListProjectsHandler returns projects for authenticated user
func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUserFromContext(r.Context())
    if !ok {
        dashboard.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
        return
    }

    // Query Kubernetes for PipelineConfigs (projects)
    configs, err := k8sClient.ListPipelineConfigs(r.Context(), user.Namespace)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
        return
    }

    // Map to DTOs
    dtos := make([]*dashboard.ProjectDTO, len(configs.Items))
    for i, config := range configs.Items {
        dtos[i] = mapPipelineConfigToProjectDTO(&config)
    }

    dashboard.WriteSuccess(w, dtos, nil)
}
```

**Acceptance**: Endpoint returns user's projects.

**Dependencies**: T007, T008
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

---

## T062: Implement CreateProject API handler
**Description**: Handler for POST /api/projects creating new project.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go` (continue)

**Function**:
```go
// CreateProjectHandler creates new project (PipelineConfig)
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUserFromContext(r.Context())
    if !ok {
        dashboard.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
        return
    }

    var req struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        RepoURL     string `json:"repository_url"`
        Namespace   string `json:"namespace"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        dashboard.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
        return
    }

    // Validate inputs
    if err := validateProjectRequest(req); err != nil {
        dashboard.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
        return
    }

    // Create PipelineConfig in Kubernetes
    config := &v1alpha1.PipelineConfig{
        ObjectMeta: metav1.ObjectMeta{
            Name:      req.Name,
            Namespace: req.Namespace,
        },
        Spec: v1alpha1.PipelineConfigSpec{
            Repository: req.RepoURL,
            // ... other fields ...
        },
    }

    if err := k8sClient.Create(r.Context(), config); err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }

    // Generate webhook URL
    webhookURL := generateWebhookURL(req.Name)

    dto := mapPipelineConfigToProjectDTO(config)
    dto.WebhookURL = webhookURL

    dashboard.WriteSuccess(w, dto, nil)
}
```

**Acceptance**: Endpoint creates project and returns webhook URL.

**Dependencies**: T007, T061
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

---

## T063: Implement GetWebhookConfig API handler
**Description**: Handler for GET /api/projects/{projectId}/webhook returning webhook details.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go` (continue)

**Function**:
```go
// GetWebhookConfigHandler returns webhook configuration for project
func GetWebhookConfigHandler(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectId")

    // Fetch project
    config, err := k8sClient.GetPipelineConfig(r.Context(), namespace, projectID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "Project not found")
        return
    }

    // Generate webhook URL
    webhookURL := generateWebhookURL(config.Name)

    response := map[string]interface{}{
        "project_id":                projectID,
        "webhook_url":               webhookURL,
        "git_platform":              "github", // TODO: detect from repo URL
        "registration_instructions": getWebhookInstructions("github"),
    }

    dashboard.WriteSuccess(w, response, nil)
}
```

**Acceptance**: Endpoint returns webhook URL and registration instructions.

**Dependencies**: T061
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

---

## T064: Create ProjectsHandler for projects page
**Description**: Handler for GET /dashboard/projects rendering projects page.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go` (continue)

**Function**:
```go
// ProjectsHandler renders projects list page
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUserFromContext(r.Context())
    if !ok {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        return
    }

    // Fetch user's projects
    configs, err := k8sClient.ListPipelineConfigs(r.Context(), user.Namespace)
    if err != nil {
        renderErrorPage(w, 500, "Failed to fetch projects")
        return
    }

    dtos := make([]*dashboard.ProjectDTO, len(configs.Items))
    for i, config := range configs.Items {
        dtos[i] = mapPipelineConfigToProjectDTO(&config)
    }

    // Render full page
    dashboard.Templates.ExecuteTemplate(w, "base", map[string]interface{}{
        "User":     user,
        "Projects": dtos,
    })
}
```

**Acceptance**: /dashboard/projects renders projects list.

**Dependencies**: T060, T061
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T065: Create project creation form partial [P]
**Description**: Modal/form for creating new project.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/project_form.html`

**Content**:
```html
{{define "partials/project_form"}}
<div class="modal" id="create-project-modal">
    <div class="modal-content">
        <h2>Create New Project</h2>
        <form hx-post="/api/projects"
              hx-target="#projects-list"
              hx-swap="afterbegin"
              hx-on::after-request="closeModal()">

            <div class="form-group">
                <label>Project Name</label>
                <input type="text" name="name" required placeholder="my-project">
            </div>

            <div class="form-group">
                <label>Description</label>
                <textarea name="description" placeholder="Project description"></textarea>
            </div>

            <div class="form-group">
                <label>Repository URL</label>
                <input type="text" name="repository_url" required placeholder="https://github.com/org/repo">
            </div>

            <div class="form-group">
                <label>Namespace</label>
                <input type="text" name="namespace" required placeholder="c8s-my-project">
            </div>

            <div class="form-actions">
                <button type="submit" class="btn btn-primary">Create Project</button>
                <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
            </div>
        </form>
    </div>
</div>
{{end}}
```

**Acceptance**: Form submits via HTMX and creates project.

**Dependencies**: T001, T062
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/project_form.html`

---

## T066: Create webhook display partial [P]
**Description**: Component showing webhook URL with copy-to-clipboard.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/webhook_display.html`

**Content**:
```html
{{define "partials/webhook_display"}}
<div class="webhook-section">
    <h3>Webhook Configuration</h3>
    <div class="webhook-url-box">
        <code id="webhook-url">{{.WebhookURL}}</code>
        <button class="btn btn-sm" onclick="copyWebhookURL()">Copy</button>
    </div>

    <div class="webhook-instructions">
        <h4>Setup Instructions</h4>
        <ol>
            <li>Go to your repository settings on {{.GitPlatform}}</li>
            <li>Navigate to Webhooks</li>
            <li>Add new webhook with the URL above</li>
            <li>Select "Push events" and "Pull request events"</li>
            <li>Save webhook</li>
        </ol>
    </div>
</div>

<script>
function copyWebhookURL() {
    const url = document.getElementById('webhook-url').innerText;
    navigator.clipboard.writeText(url);
    alert('Webhook URL copied!');
}
</script>
{{end}}
```

**Acceptance**: Webhook URL displays with functional copy button.

**Dependencies**: T001
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/webhook_display.html`

---

## T067: Write integration test for projects API
**Description**: Test project CRUD operations.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/projects_test.go`

**Tests**:
```go
func TestListProjects_ReturnsProjects(t *testing.T)
func TestCreateProject_CreatesProject(t *testing.T)
func TestCreateProject_ValidatesInput(t *testing.T)
func TestGetWebhookConfig_ReturnsWebhookURL(t *testing.T)
```

**Acceptance**: Project API tests pass, CRUD operations work.

**Dependencies**: T061, T062, T063
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/projects_test.go`

---

## T068: Write E2E test for project management
**Description**: Playwright test for project creation and webhook display.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/projects.spec.ts`

**Tests**:
```typescript
test('projects page displays projects', async ({ page }) => {})
test('create project form opens and submits', async ({ page }) => {})
test('webhook URL displays and copies', async ({ page }) => {})
test('project appears in list after creation', async ({ page }) => {})
```

**Acceptance**: E2E tests pass, project management interactive.

**Dependencies**: T060, T061, T062, T064
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/projects.spec.ts`

---

## T069: Register US4 routes in main.go
**Description**: Add all US4 routes to router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update)

**Routes**:
```go
router.Group(func(r chi.Router) {
    r.Use(handlers.AuthMiddleware)
    r.Get("/dashboard/projects", handlers.ProjectsHandler)
    r.Get("/api/projects", handlers.ListProjectsHandler)
    r.Post("/api/projects", handlers.CreateProjectHandler)
    r.Get("/api/projects/{projectId}/webhook", handlers.GetWebhookConfigHandler)
})
```

**Acceptance**: All US4 routes registered and accessible.

**Dependencies**: T061, T062, T063, T064
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 7: US5 - View and Manage Artifacts [P3]

## T070: Create artifacts list partial template [P]
**Description**: Component displaying artifacts for a pipeline run.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/artifacts_list.html`

**Content**:
```html
{{define "partials/artifacts_list"}}
<div class="artifacts-section">
    <h3>Artifacts</h3>

    {{if .Artifacts}}
        <div class="artifacts-table">
            {{range .Artifacts}}
                <div class="artifact-row">
                    <div class="artifact-icon">📦</div>
                    <div class="artifact-info">
                        <div class="artifact-name">{{.Name}}</div>
                        <div class="artifact-meta">
                            {{.Type}} • {{.SizeBytes | formatBytes}} • {{.CreatedAt | formatTime}}
                        </div>
                    </div>
                    <div class="artifact-actions">
                        <a href="/api/artifacts/{{.ID}}/download" class="btn btn-sm" download>Download</a>
                        {{if eq .Type "report"}}
                            <button class="btn btn-sm" hx-get="/api/artifacts/{{.ID}}/preview" hx-target="#artifact-preview">Preview</button>
                        {{end}}
                    </div>
                </div>
            {{end}}
        </div>
    {{else}}
        {{template "partials/empty_state" (dict "Title" "No Artifacts" "Message" "This pipeline run did not produce any artifacts.")}}
    {{end}}
</div>
{{end}}
```

**Acceptance**: Artifacts list renders with download and preview actions.

**Dependencies**: T001, T024
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/artifacts_list.html`

---

## T071: Implement ListArtifacts API handler
**Description**: Handler for GET /api/runs/{runId}/artifacts returning artifact list.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

**Function**:
```go
package handlers

// ListArtifactsHandler returns artifacts for pipeline run
func ListArtifactsHandler(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runId")
    stepID := r.URL.Query().Get("step_id") // optional filter
    artifactType := r.URL.Query().Get("type") // optional filter

    // Fetch pipeline run
    run, err := k8sClient.GetPipelineRun(r.Context(), namespace, runID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "RUN_NOT_FOUND", "Pipeline run not found")
        return
    }

    // Extract artifacts from step statuses
    artifacts := extractArtifactsFromRun(run, stepID, artifactType)

    // Map to DTOs
    dtos := make([]*dashboard.ArtifactDTO, len(artifacts))
    for i, artifact := range artifacts {
        dtos[i] = mapArtifactToDTO(artifact)
    }

    dashboard.WriteSuccess(w, dtos, nil)
}

// extractArtifactsFromRun retrieves artifact URLs from step statuses
func extractArtifactsFromRun(run *v1alpha1.PipelineRun, stepFilter, typeFilter string) []Artifact
```

**Acceptance**: Endpoint returns artifacts with metadata.

**Dependencies**: T007, T008
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

---

## T072: Implement DownloadArtifact handler
**Description**: Handler for GET /api/artifacts/{artifactId}/download streaming artifact file.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go` (continue)

**Function**:
```go
// DownloadArtifactHandler streams artifact file for download
func DownloadArtifactHandler(w http.ResponseWriter, r *http.Request) {
    artifactID := chi.URLParam(r, "artifactId")

    // Fetch artifact metadata
    artifact, err := fetchArtifactMetadata(artifactID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "ARTIFACT_NOT_FOUND", "Artifact not found")
        return
    }

    // Stream from S3
    reader, err := s3Storage.GetObject(artifact.StorageKey)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "DOWNLOAD_FAILED", err.Error())
        return
    }
    defer reader.Close()

    // Set headers
    w.Header().Set("Content-Type", artifact.MimeType)
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", artifact.Name))
    w.Header().Set("Content-Length", fmt.Sprintf("%d", artifact.SizeBytes))

    // Stream to client
    io.Copy(w, reader)
}
```

**Acceptance**: Artifact downloads successfully with correct headers.

**Dependencies**: T071
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

---

## T073: Implement PreviewArtifact handler
**Description**: Handler for GET /api/artifacts/{artifactId}/preview rendering HTML/markdown inline.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go` (continue)

**Function**:
```go
// PreviewArtifactHandler renders artifact inline (HTML/markdown)
func PreviewArtifactHandler(w http.ResponseWriter, r *http.Request) {
    artifactID := chi.URLParam(r, "artifactId")

    // Fetch artifact metadata
    artifact, err := fetchArtifactMetadata(artifactID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "ARTIFACT_NOT_FOUND", "Artifact not found")
        return
    }

    // Check if artifact is previewable
    if !isPreviewable(artifact.MimeType) {
        dashboard.WriteError(w, http.StatusBadRequest, "NOT_PREVIEWABLE", "Artifact cannot be previewed")
        return
    }

    // Fetch content from S3
    reader, err := s3Storage.GetObject(artifact.StorageKey)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "PREVIEW_FAILED", err.Error())
        return
    }
    defer reader.Close()

    content, _ := io.ReadAll(reader)

    // Render preview template
    dashboard.Templates.ExecuteTemplate(w, "artifact_preview", map[string]interface{}{
        "Artifact": artifact,
        "Content":  string(content),
    })
}

func isPreviewable(mimeType string) bool {
    return mimeType == "text/html" || mimeType == "text/markdown" || mimeType == "text/plain"
}
```

**Acceptance**: HTML/markdown artifacts render inline in preview.

**Dependencies**: T072
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

---

## T074: Add artifacts section to pipeline detail page
**Description**: Integrate artifacts list into pipeline detail template.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html` (update)

**Changes**:
```html
{{define "content"}}
<!-- Existing pipeline run details -->
...

<!-- Add artifacts section -->
{{template "partials/artifacts_list" .}}

{{end}}
```

**Acceptance**: Artifacts section displays on pipeline detail page.

**Dependencies**: T041, T070
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

---

## T075: Write integration test for artifacts API
**Description**: Test artifact listing, download, and preview.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/artifacts_test.go`

**Tests**:
```go
func TestListArtifacts_ReturnsArtifacts(t *testing.T)
func TestDownloadArtifact_StreamsFile(t *testing.T)
func TestPreviewArtifact_RendersHTML(t *testing.T)
func TestListArtifacts_FiltersByStep(t *testing.T)
```

**Acceptance**: Artifact API tests pass, download and preview work.

**Dependencies**: T071, T072, T073
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/artifacts_test.go`

---

## T076: Write E2E test for artifact management
**Description**: Playwright test for artifact download and preview.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/artifacts.spec.ts`

**Tests**:
```typescript
test('artifacts list displays', async ({ page }) => {})
test('artifact downloads successfully', async ({ page }) => {})
test('HTML artifact previews inline', async ({ page }) => {})
test('empty state shows when no artifacts', async ({ page }) => {})
```

**Acceptance**: E2E tests pass, artifacts interactive.

**Dependencies**: T070, T071, T072, T074
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/artifacts.spec.ts`

---

## T077: Register US5 routes in main.go
**Description**: Add all US5 routes to router.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (update)

**Routes**:
```go
router.Group(func(r chi.Router) {
    r.Use(handlers.AuthMiddleware)
    r.Get("/api/runs/{runId}/artifacts", handlers.ListArtifactsHandler)
    r.Get("/api/artifacts/{artifactId}/download", handlers.DownloadArtifactHandler)
    r.Get("/api/artifacts/{artifactId}/preview", handlers.PreviewArtifactHandler)
})
```

**Acceptance**: All US5 routes registered and accessible.

**Dependencies**: T071, T072, T073
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 8: Polish & Cross-Cutting Concerns

## T078: Implement error handling middleware
**Description**: Centralized error handling with proper logging.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/error_middleware.go`

**Function**:
```go
package handlers

import (
    "log"
    "net/http"
)

// ErrorRecoveryMiddleware catches panics and returns 500
func ErrorRecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                dashboard.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}
```

**Acceptance**: Errors logged, panics caught and returned as 500.

**Dependencies**: T010
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/error_middleware.go`

---

## T079: Add caching layer for pipeline lists
**Description**: Implement Redis/in-memory cache for frequently accessed data.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/cache.go`

**Functions**:
```go
package dashboard

import (
    "context"
    "time"
)

type CacheLayer struct {
    cache map[string]interface{}
    ttl   time.Duration
}

// Get retrieves value from cache
func (c *CacheLayer) Get(key string) (interface{}, bool)

// Set stores value in cache with TTL
func (c *CacheLayer) Set(key string, value interface{}, ttl time.Duration)

// Invalidate removes value from cache
func (c *CacheLayer) Invalidate(key string)
```

**Acceptance**: Cache reduces database queries, improves performance.

**Dependencies**: None
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/cache.go`

---

## T080: Add loading states to HTMX requests
**Description**: Show loading indicators during async HTMX operations.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (update)

**Styles**:
```css
/* HTMX loading indicators */
.htmx-request .htmx-indicator {
    display: inline-block;
}

.htmx-indicator {
    display: none;
}

.htmx-swapping {
    opacity: 0.5;
    transition: opacity 200ms ease-in;
}

.spinner {
    border: 2px solid #f3f3f3;
    border-top: 2px solid #3498db;
    border-radius: 50%;
    width: 20px;
    height: 20px;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}
```

**Acceptance**: Loading indicators display during HTMX requests.

**Dependencies**: T005
**Story**: Polish
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T081: Add keyboard shortcuts
**Description**: Implement keyboard navigation (r=refresh, /=search focus, esc=close modal).

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/keyboard.js`

**Content**:
```javascript
// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // r = refresh current view
    if (e.key === 'r' && !isInputFocused()) {
        location.reload();
    }

    // / = focus search
    if (e.key === '/' && !isInputFocused()) {
        e.preventDefault();
        document.querySelector('input[name="search"]')?.focus();
    }

    // esc = close modal
    if (e.key === 'Escape') {
        document.querySelectorAll('.modal').forEach(m => m.style.display = 'none');
    }
});

function isInputFocused() {
    return document.activeElement.tagName === 'INPUT' ||
           document.activeElement.tagName === 'TEXTAREA';
}
```

**Acceptance**: Keyboard shortcuts work as specified.

**Dependencies**: None
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/keyboard.js`

---

## T082: Write comprehensive E2E test suite
**Description**: Master E2E test covering complete user workflow.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/full_workflow.spec.ts`

**Test**:
```typescript
test('complete dashboard workflow', async ({ page }) => {
    // 1. Login
    await page.goto('http://localhost:8080/login');
    await page.fill('input[name="username"]', 'testuser');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');

    // 2. View dashboard
    await expect(page.locator('h2:has-text("Pipeline Runs")')).toBeVisible();

    // 3. Filter pipelines
    await page.fill('input[name="search"]', 'abc123');
    await page.waitForTimeout(600);

    // 4. Click pipeline detail
    await page.click('.pipeline-row a:has-text("View")');
    await expect(page.locator('.step-item')).toHaveCount(3);

    // 5. View logs
    await page.click('button:has-text("View Logs")');
    await expect(page.locator('.log-line')).toHaveCountGreaterThan(0);

    // 6. Download artifact
    const downloadPromise = page.waitForEvent('download');
    await page.click('a:has-text("Download")');
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toContain('.html');

    // 7. Create new project
    await page.goto('http://localhost:8080/dashboard/projects');
    await page.click('button:has-text("Create Project")');
    await page.fill('input[name="name"]', 'test-project');
    await page.fill('input[name="repository_url"]', 'https://github.com/test/repo');
    await page.click('button[type="submit"]');
    await expect(page.locator('.project-row:has-text("test-project")')).toBeVisible();
});
```

**Acceptance**: Full workflow E2E test passes, covering all stories.

**Dependencies**: All prior tasks
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/full_workflow.spec.ts`

---

## End of Task List

**Next Steps**:
1. Review and approve task list
2. Begin Phase 1 implementation (T001-T015)
3. Implement in dependency order
4. Test continuously (unit → integration → E2E)
5. Deploy MVP (P1 stories) for feedback
6. Iterate on P2 and P3 stories

**Success Metrics** (from spec.md):
- SC-001: Dashboard loads in <2s ✓ (via caching T079)
- SC-003: Log latency <2s ✓ (via SSE T044)
- SC-004: Search <1s ✓ (via indexing T053)
- SC-005: 100+ concurrent users ✓ (via SSE broadcaster T018, caching T079)

**Deliverables**:
- Functional web dashboard integrated with C8S API server
- Real-time pipeline monitoring with SSE
- Search/filter capabilities
- Project and webhook management
- Artifact viewing and download
- Comprehensive test coverage (unit, integration, E2E)
