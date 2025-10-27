# Implementation Tasks: C8S Web Dashboard

**Feature**: Web Dashboard for C8S CI Workflows
**Branch**: `004-create-a-front`
**Generated**: 2025-10-26
**Tech Stack**: Go 1.24.0 backend (html/template), HTMX frontend, Server-Sent Events

---

## Summary

**Total Tasks**: 96+ (added 3 HTTPS/security + 3 RBAC + 1 real-time list + 1 keyboard shortcuts + 1 perf benchmarking + 4 caching + 5 artifact sanitization + 5 pipeline cancellation tasks)
**By Story**:
- Setup/Foundation: 21 tasks (T016-T018 for HTTPS/TLS/security headers; expanded T011 + T011a-T011c for Auth/RBAC)
- US1 (P1 - Pipeline History): 11 tasks (T037a for HTMX SSE subscription)
- US2 (P1 - Step Execution & Logs): 16 tasks (added T080d-T080e for pipeline cancellation)
- US3 (P2 - Search & Filter): 8 tasks
- US4 (P2 - Projects & Webhooks): 10 tasks
- US5 (P3 - Artifacts): 13 tasks (added T080a-T080c for artifact sanitization)
- Polish/Cross-cutting: 17 tasks (added T081a for keyboard shortcuts, T084a for perf benchmarking, T082-T082c for caching strategy)

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
├─ T016-T033 (Shared components, auth, API base)
│
Phase 3: US1 - Pipeline History [P1] ──┐
├─ T034-T043                            │
│                                       ├─ Can run in parallel
Phase 4: US2 - Logs & Execution [P1] ──┤
├─ T044-T054                            │
│                                       │
Phase 5: US3 - Search & Filter [P2] ───┤
├─ T055-T062 (depends on T034-T043)    │
│                                       │
Phase 6: US4 - Projects & Webhooks [P2]│
├─ T063-T072 (independent)             │
│                                       │
Phase 7: US5 - Artifacts [P3] ─────────┘
├─ T073-T080 (depends on T044-T054)
│
Phase 8: Polish & Cross-cutting
└─ T081-T085 (depends on all prior phases)
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
**Description**: Middleware to verify JWT/OAuth2 tokens and attach user context. Integrates with existing C8S auth system to validate bearer tokens per FR-010.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/auth_middleware.go`

**Functions**:
```go
package handlers

import (
    "context"
    "net/http"
    "github.com/org/c8s/pkg/auth"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware validates bearer token and attaches user to context
func AuthMiddleware(authClient *auth.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r.Header.Get("Authorization"))
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Validate token with existing C8S auth system
            user, err := authClient.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), userContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*User, bool)

// extractBearerToken extracts token from Authorization header
func extractBearerToken(authHeader string) string
```

**Acceptance**:
- Middleware validates token with auth client
- Unauthenticated requests return 401
- User object attached to context for downstream handlers

**Dependencies**: Assumes existing C8S auth.Client in pkg/auth
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/auth_middleware.go`

---

## T011a: Implement project-based access control [P]
**Description**: Create authorization layer to enforce per-project access control. Users should only see/access pipelines/artifacts for projects they have membership in per FR-010.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/authz_middleware.go` (new)

**Functions**:
```go
package handlers

import (
    "context"
    "net/http"
    "github.com/org/c8s/pkg/dashboard"
)

// ProjectAccessMiddleware enforces per-project authorization
func ProjectAccessMiddleware(projectSvc *dashboard.ProjectService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := GetUserFromContext(r.Context())
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            projectID := r.PathValue("projectID")  // Go 1.24 routing
            if projectID == "" {
                next.ServeHTTP(w, r)  // No project scope, proceed
                return
            }

            // Check if user has access to this project
            hasAccess, err := projectSvc.UserHasProjectAccess(r.Context(), user.ID, projectID)
            if err != nil || !hasAccess {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RoleBasedContextMiddleware attaches user role for project to context
func RoleBasedContextMiddleware(projectSvc *dashboard.ProjectService) func(http.Handler) http.Handler {
    // Extracts user role (admin/editor/viewer) for the project being accessed
    // Allows handlers to render different UI based on permission level
}
```

**Acceptance**:
- Users can only access projects they have membership in
- Forbidden response (403) for unauthorized access
- User role attached to context for role-based UI rendering

**Dependencies**: T011 (auth middleware prerequisite)
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/authz_middleware.go`

---

## T011b: Create project membership checking service [P]
**Description**: Implement service layer for checking user project membership and roles. Used by authorization middleware and handlers.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/project_access.go` (new)

**Functions**:
```go
package dashboard

import "context"

// ProjectAccessService checks user access to projects
type ProjectAccessService interface {
    // UserHasProjectAccess checks if user can access project
    UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error)

    // GetUserRoleForProject returns user's role in project (admin/editor/viewer)
    GetUserRoleForProject(ctx context.Context, userID, projectID string) (Role, error)

    // ListUserProjects returns all projects user has access to
    ListUserProjects(ctx context.Context, userID string) ([]*Project, error)
}

type Role string
const (
    RoleAdmin   Role = "admin"
    RoleEditor  Role = "editor"
    RoleViewer  Role = "viewer"
)
```

**Acceptance**:
- Service queries K8s for project membership
- Returns correct role for user-project pair
- Caches results for performance

**Dependencies**: T007 (K8s client)
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/project_access.go`

---

## T011c: Write authorization tests [P]
**Description**: Create unit and integration tests for auth/authz middleware and access control service.

**Files**:
- `/Users/elavigne/workspace/c8s/tests/unit/authz_test.go`
- `/Users/elavigne/workspace/c8s/tests/integration/authz_integration_test.go`

**Tests**:
```go
// Unit tests
func TestProjectAccessMiddleware_AllowedUser(t *testing.T)     // User with access succeeds
func TestProjectAccessMiddleware_DeniedUser(t *testing.T)      // User without access gets 403
func TestRoleBasedContext_AdminRole(t *testing.T)              // Admin role attached to context
func TestRoleBasedContext_ViewerRole(t *testing.T)             // Viewer role attached to context

// Integration tests
func TestUserCannotViewAnotherUsersProject(t *testing.T)       // Cross-project access denied
func TestAdminCanManageProject(t *testing.T)                   // Admin has full permissions
func TestViewerCannotDeleteProject(t *testing.T)               // Viewer restricted to read-only
```

**Acceptance**:
- All auth/authz tests pass
- Authorization layer correctly prevents unauthorized access
- Roles properly restrict UI/API capabilities

**Dependencies**: T011, T011a, T011b
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/authz_test.go`
- `/Users/elavigne/workspace/c8s/tests/integration/authz_integration_test.go`

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
        r.Get("/dashboard", handlers.DashboardHandler)          // T034
        r.Get("/dashboard/projects", handlers.ProjectsHandler)  // T063
        r.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler) // T035
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

## T016: Configure HTTPS/TLS for dashboard [P]
**Description**: Set up HTTPS/TLS termination with certificates for secure dashboard access per FR-014. Configure Go API server to serve HTTPS on port 8443 with certificate handling.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (modifications)

**Actions**:
1. Configure TLS listener in main.go:
   - Load certificates from environment variables or files
   - Support self-signed certs for development; production certs via Let's Encrypt or K8s-provided certs
   - Implement certificate reload without restart (hot-reload)

2. Update server configuration:
   - Create TLSConfig with strong cipher suites (TLS 1.2+)
   - Set secure defaults (no weak ciphers, require client verification if needed)

3. Document certificate setup:
   - How to generate self-signed certs for local development
   - How to configure with Let's Encrypt in production
   - Environment variables: `TLS_CERT_PATH`, `TLS_KEY_PATH`

**Acceptance**:
- HTTPS listener starts on port 8443
- HTTP requests redirect to HTTPS
- TLS certificate chain loads correctly
- Server responds with valid TLS handshake

**Dependencies**: T013
**Story**: Foundation
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

**Files Created**:
- `/Users/elavigne/workspace/c8s/docs/HTTPS_SETUP.md`

---

## T017: Implement HTTP security headers [P]
**Description**: Add security headers to all dashboard responses to prevent common attacks (XSS, clickjacking, etc.) per FR-014.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/middleware/security_headers.go` (new)

**Headers to add**:
```go
// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "SAMEORIGIN")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

**Acceptance**:
- All dashboard responses include security headers
- Integration tests verify header presence
- HSTS header enforces HTTPS for future requests
- CSP prevents inline script injection

**Dependencies**: T013
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/middleware/security_headers.go`

**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (register middleware)

---

## T018: Write HTTPS/TLS security tests [P]
**Description**: Create integration tests to verify HTTPS enforced, certificates valid, and security headers present.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/security_test.go` (new)

**Tests**:
```go
func TestHTTPSEnforced(t *testing.T) // HTTP redirects to HTTPS
func TestTLSHandshakeValid(t *testing.T) // TLS version 1.2+
func TestSecurityHeadersPresent(t *testing.T) // All security headers present
func TestCSPHeader(t *testing.T) // Content-Security-Policy properly formed
```

**Acceptance**:
- All TLS/security tests pass
- HTTP requests are rejected or redirected
- TLS version < 1.2 rejected

**Dependencies**: T016, T017
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/security_test.go`

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

## T022: Create pagination utility
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

## T023: Create time formatting utilities
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

## T024: Register template functions for formatting
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

**Dependencies**: T006, T023
**Story**: Foundation
**Files Modified**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/templates.go`

---

## T025: Create status badge partial template [P]
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

## T026: Create loading spinner partial template [P]
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

## T027: Create empty state partial template [P]
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

## T028: Write unit tests for mappers
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

## T029: Write unit tests for pagination
**Description**: Test pagination utility functions.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/pagination_test.go`

**Tests**:
```go
func TestParsePaginationParams(t *testing.T)
func TestPaginate(t *testing.T)
func TestPaginateEmptyList(t *testing.T)
```

**Acceptance**: Pagination tests pass, handles edge cases (empty lists, invalid pages).

**Dependencies**: T022
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/pagination_test.go`

---

## T030: Write unit tests for SSE broadcaster
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

## T031: Write unit tests for time utilities
**Description**: Test timestamp and duration formatting functions.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/time_utils_test.go`

**Tests**:
```go
func TestFormatTimestamp(t *testing.T)
func TestFormatDuration(t *testing.T)
func TestIsRecent(t *testing.T)
```

**Acceptance**: Time utility tests pass, covers various time ranges.

**Dependencies**: T023
**Story**: Foundation
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/time_utils_test.go`

---

## T032: Write integration test for auth middleware
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

## T033: Write integration test for static file serving
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

## T034: Create pipeline list page template
**Description**: HTML template for main dashboard page listing pipeline runs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

**Content**: (Reference quickstart.md example)
- Filter controls (search, branch, status)
- HTMX-enhanced table with SSE updates
- Pagination controls
- Empty state if no runs

**Acceptance**: Template renders with dummy data, HTMX attributes present.

**Dependencies**: T003, T025, T027
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

---

## T035: Implement ListPipelineRuns API handler
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

**Dependencies**: T007, T008, T009, T010, T022
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T036: Create DashboardHandler for main dashboard page
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

**Dependencies**: T034, T035
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T037: Implement SSE endpoint for pipeline status updates
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

**Dependencies**: T018, T035
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_sse.go`

---

## T037a: Implement HTMX SSE subscription for pipeline list [P]
**Description**: Add HTMX SSE integration to pipeline list page for real-time updates (per analysis A11). When new runs are triggered, they appear at the top of the list without page refresh per spec.md:US1 Acceptance Criteria #4.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html` (modifications)

**Implementation**:
```html
<!-- Add to pipeline_list.html base template -->
<div id="pipeline-list-container"
     hx-ext="sse"
     sse-connect="/api/projects/{{.ProjectID}}/runs/updates"
     sse-swap="run_status_changed:swap:#pipeline-list-container"
     hx-trigger="sse:run_status_changed from:body"
     hx-swap="innerHTML">

     <!-- Pipeline list rows inserted here -->
     <div id="pipeline-rows">
        {{range .Runs}}
          {{template "partials/pipeline_row" .}}
        {{end}}
     </div>
</div>

<script>
  // Handle new runs: fetch fresh list and insert at top
  document.body.addEventListener('htmx:sseMessage', function(event) {
    if (event.detail.event === 'run_status_changed') {
      var update = JSON.parse(event.detail.data);
      // For new runs (not in current list), fetch updated list and swap
      htmx.ajax('GET',
        '/api/projects/{{.ProjectID}}/runs',
        '#pipeline-rows'
      );
    }
  });
</script>
```

**Acceptance**:
- HTMX SSE extension active and subscribes to pipeline updates
- When new pipeline run created, list updates without page refresh
- New run appears at top of list (per US1 acceptance criteria #4)
- Existing runs update their status in real-time
- Reconnects automatically if connection drops

**Dependencies**: T037 (SSE endpoint), T034 (list template)
**Story**: US1
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

---

## T038: Create pipeline row partial template [P]
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

**Dependencies**: T001, T025
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/pipeline_row.html`

---

## T039: Write integration test for ListPipelineRuns handler
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

**Dependencies**: T035
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_list_test.go`

---

## T040: Write E2E test for pipeline list page
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

**Dependencies**: T034, T035, T036, T037
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_list.spec.ts`

---

## T041: Add CSS styling for pipeline list
**Description**: Style pipeline list table, filters, and status badges.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (update)

**Styles**:
- `.pipeline-row`: Grid layout with hover effect
- `.pipeline-meta`: Gray text, smaller font
- `.badge-*`: Status badge colors (green, red, orange)
- `.filters`: Flex layout for search/filter controls

**Acceptance**: Pipeline list visually appealing and responsive.

**Dependencies**: T005, T034
**Story**: US1
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T042: Implement Kubernetes watch for pipeline updates
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

**Dependencies**: T007, T018, T037
**Story**: US1
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/pipeline_watcher.go`

---

## T043: Register US1 routes in main.go
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

**Dependencies**: T035, T036, T037
**Story**: US1
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 4: US2 - Monitor Step-by-Step Execution and Logs [P1]

## T044: Create pipeline detail page template
**Description**: Template for detailed pipeline run view with steps and logs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

**Content**:
- Pipeline run metadata (commit, branch, status, duration)
- Step list with status, duration, dependencies
- Log viewer section (SSE-streamed logs)
- Step DAG visualization (optional)

**Acceptance**: Template renders with dummy data, log viewer HTMX attributes present.

**Dependencies**: T003, T025
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

---

## T045: Implement GetPipelineRun API handler
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

## T046: Create PipelineDetailHandler for detail page
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

**Dependencies**: T044, T045
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T047: Implement SSE log streaming endpoint
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

## T048: Create log viewer partial template [P]
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

**Dependencies**: T001, T047
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/log_viewer.html`

---

## T049: Create step status partial template [P]
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

**Dependencies**: T001, T025
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/step_status.html`

---

## T050: Write integration test for GetPipelineRun handler
**Description**: Test pipeline detail API endpoint.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/pipeline_detail_test.go`

**Tests**:
```go
func TestGetPipelineRun_ReturnsRun(t *testing.T)
func TestGetPipelineRun_ReturnsSteps(t *testing.T)
func TestGetPipelineRun_Returns404ForInvalidID(t *testing.T)
```

**Acceptance**: Detail endpoint tests pass, returns run with steps.

**Dependencies**: T045
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_detail_test.go`

---

## T051: Write integration test for log streaming endpoint
**Description**: Test SSE log streaming.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/log_stream_test.go`

**Tests**:
```go
func TestLogStream_StreamsLogs(t *testing.T)
func TestLogStream_ReturnsErrorForInvalidStep(t *testing.T)
func TestLogStream_CompletesWhenLogsDone(t *testing.T)
```

**Acceptance**: Log streaming tests pass, SSE events received correctly.

**Dependencies**: T047
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/log_stream_test.go`

---

## T052: Write E2E test for pipeline detail page
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

**Dependencies**: T044, T045, T046, T047
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_detail.spec.ts`

---

## T053: Add CSS styling for pipeline detail page
**Description**: Style step list, log viewer, and detail layout.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (update)

**Styles**:
- `.step-item`: Card layout with border, padding
- `.log-viewer`: Dark background, monospace font, auto-scroll
- `.log-line`: Individual log line styling
- `.step-details`: Collapsible section styling

**Acceptance**: Pipeline detail page visually organized and readable.

**Dependencies**: T005, T044, T048, T049
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css`

---

## T054: Register US2 routes in main.go
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

**Dependencies**: T045, T046, T047
**Story**: US2
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 5: US3 - Search and Filter Pipelines [P2]

## T055: Create filter panel partial template [P]
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

## T056: Update ListPipelineRuns handler with filter logic
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

**Dependencies**: T035, T055
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_runs.go`

---

## T057: Implement branch list API endpoint
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

## T058: Add filter panel to pipeline list template
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

**Dependencies**: T034, T055
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_list.html`

---

## T059: Write unit tests for filter parsing
**Description**: Test filter parameter parsing and validation.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/filters_test.go`

**Tests**:
```go
func TestParseFilters_ParsesAllParameters(t *testing.T)
func TestParseFilters_HandlesInvalidDates(t *testing.T)
func TestParseFilters_HandlesEmptyFilters(t *testing.T)
```

**Acceptance**: Filter parsing tests pass, handles invalid input.

**Dependencies**: T056
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/filters_test.go`

---

## T060: Write integration test for filtered pipeline list
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

**Dependencies**: T056
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_filters_test.go`

---

## T061: Write E2E test for search and filter
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

**Dependencies**: T055, T056, T058
**Story**: US3
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/pipeline_filters.spec.ts`

---

## T062: Register US3 routes in main.go
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

**Dependencies**: T057
**Story**: US3
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 6: US4 - Configure Projects and Webhooks [P2]

## T063: Create projects list page template
**Description**: Template for listing user's projects with webhook URLs.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/projects.html`

**Content**:
- List of projects with name, repo URL, last run time
- Webhook URL display with copy-to-clipboard button
- Create project button (opens modal/form)
- Delete project action

**Acceptance**: Template renders with project list and actions.

**Dependencies**: T003, T027
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/projects.html`

---

## T064: Implement ListProjects API handler
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

## T065: Implement CreateProject API handler
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

**Dependencies**: T007, T064
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

---

## T066: Implement GetWebhookConfig API handler
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

**Dependencies**: T064
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/projects.go`

---

## T067: Create ProjectsHandler for projects page
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

**Dependencies**: T063, T064
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/dashboard.go`

---

## T068: Create project creation form partial [P]
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

**Dependencies**: T001, T065
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/project_form.html`

---

## T069: Create webhook display partial [P]
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

## T070: Write integration test for projects API
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

**Dependencies**: T064, T065, T066
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/projects_test.go`

---

## T071: Write E2E test for project management
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

**Dependencies**: T063, T064, T065, T067
**Story**: US4
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/projects.spec.ts`

---

## T072: Register US4 routes in main.go
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

**Dependencies**: T064, T065, T066, T067
**Story**: US4
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

# Phase 7: US5 - View and Manage Artifacts [P3]

## T073: Create artifacts list partial template [P]
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

**Dependencies**: T001, T027
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/artifacts_list.html`

---

## T074: Implement ListArtifacts API handler
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

## T075: Implement DownloadArtifact handler
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

**Dependencies**: T074
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

---

## T076: Implement PreviewArtifact handler
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

**Dependencies**: T075
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go`

---

## T077: Add artifacts section to pipeline detail page
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

**Dependencies**: T044, T073
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html`

---

## T078: Write integration test for artifacts API
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

**Dependencies**: T074, T075, T076
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/artifacts_test.go`

---

## T079: Write E2E test for artifact management
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

**Dependencies**: T073, T074, T075, T077
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/artifacts.spec.ts`

---

## T080: Register US5 routes in main.go
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

**Dependencies**: T074, T075, T076
**Story**: US5
**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go`

---

## T080a: Implement artifact content sanitization (per A8) [P]
**Description**: Sanitize HTML/markdown artifact content to prevent XSS attacks per FR-009. Only allow safe HTML tags, remove scripts, style, and iframe elements.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/sanitizer.go` (new)

**Implementation**:
```go
package dashboard

import (
    "github.com/microcosm-cc/bluemonday"
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
)

type ArtifactSanitizer struct {
    htmlPolicy *bluemonday.Policy
}

// NewArtifactSanitizer creates sanitizer with safe HTML policy
func NewArtifactSanitizer() *ArtifactSanitizer

// SanitizeHTML removes dangerous tags/scripts from HTML content
func (as *ArtifactSanitizer) SanitizeHTML(html string) string

// RenderMarkdown converts markdown to safe HTML
func (as *ArtifactSanitizer) RenderMarkdown(markdown string) (string, error)

// AllowedTags defines which HTML tags are safe for artifact preview
var AllowedTags = []string{
    "p", "br", "span", "div",
    "h1", "h2", "h3", "h4", "h5", "h6",
    "ul", "ol", "li", "dl", "dt", "dd",
    "pre", "code", "blockquote",
    "strong", "em", "u", "s",
    "a", "img",
    "table", "thead", "tbody", "tfoot", "tr", "th", "td",
}
```

**Sanitization Rules**:
- Allow safe HTML tags (p, h1-h6, ul, ol, code, blockquote, table, etc.)
- Remove script, style, iframe, object, embed, form, input elements
- Remove event handlers (onclick, onload, etc.)
- Remove dangerous attributes (data-*, on*, etc.)
- For markdown: Convert to HTML first, then sanitize
- Whitelist img src to same-origin or data: URIs only

**Acceptance**:
- XSS injection attempts blocked
- Safe HTML tags preserved
- Markdown renders to safe HTML
- Links safe (no javascript: URIs)
- No inline scripts/styles execute

**Dependencies**: bluemonday, goldmark libraries
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/sanitizer.go`

**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/artifacts.go` (use sanitizer in PreviewArtifactHandler)

---

## T080b: Add Content-Security-Policy header for artifacts (per A8) [P]
**Description**: Set CSP headers specifically for artifact preview responses to prevent inline script injection.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/middleware/artifact_csp.go` (new)

**Implementation**:
```go
// ArtifactCSPMiddleware sets CSP headers for artifact preview responses
func ArtifactCSPMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only apply CSP to artifact preview endpoints
        if strings.HasPrefix(r.URL.Path, "/api/artifacts") &&
           strings.Contains(r.URL.Path, "preview") {

            // Strict CSP: no inline scripts, no external scripts
            w.Header().Set("Content-Security-Policy",
                "default-src 'none'; "+
                "script-src 'self'; "+
                "style-src 'unsafe-inline'; "+
                "img-src 'self' data:; "+
                "frame-ancestors 'none'")
        }
        next.ServeHTTP(w, r)
    })
}
```

**Acceptance**:
- CSP headers present on artifact preview responses
- No inline scripts allowed
- Only same-origin scripts allowed
- External resources blocked

**Dependencies**: T017 (security headers middleware)
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/middleware/artifact_csp.go`

---

## T080c: Write artifact sanitization tests (per A8) [P]
**Description**: Unit tests for sanitizer to verify XSS prevention and safe rendering.

**File**: `/Users/elavigne/workspace/c8s/tests/unit/sanitizer_test.go` (new)

**Tests**:
```go
func TestSanitizeHTML_RemovesScripts(t *testing.T)      // <script> tags removed
func TestSanitizeHTML_RemovesEventHandlers(t *testing.T) // onclick, etc removed
func TestSanitizeHTML_PreservesSafeHTML(t *testing.T)   // <p>, <h1>, etc preserved
func TestSanitizeHTML_BlocksDangerousLinks(t *testing.T) // javascript: URIs blocked
func TestRenderMarkdown_ProducesSafeHTML(t *testing.T)  // Markdown → safe HTML
func TestSanitizeHTML_ImgSrcWhitelisting(t *testing.T)  // Only safe image sources
```

**Acceptance**:
- All XSS vectors blocked
- Safe content preserved
- No regressions in rendering

**Dependencies**: T080a, T080b
**Story**: US5
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/sanitizer_test.go`

---

## T080d: Implement pipeline cancellation (per A9) [P]
**Description**: Add ability to cancel running pipelines per spec.md edge cases. Implement GET /api/runs/{runId}/cancel endpoint and "Cancel" button in pipeline detail view.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_control.go` (new)

**Implementation**:
```go
// CancelPipelineRunHandler stops a running pipeline
func CancelPipelineRunHandler(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runId")

    // Fetch pipeline run
    run, err := k8sClient.GetPipelineRun(r.Context(), runID)
    if err != nil {
        dashboard.WriteError(w, http.StatusNotFound, "RUN_NOT_FOUND", "")
        return
    }

    // Check authorization (user owns project)
    user, _ := handlers.GetUserFromContext(r.Context())
    if !canUserCancelRun(user, run) {
        dashboard.WriteError(w, http.StatusForbidden, "UNAUTHORIZED", "")
        return
    }

    // Cancel Kubernetes Job
    err = k8sClient.TerminateJob(r.Context(), run.Namespace, run.JobName)
    if err != nil {
        dashboard.WriteError(w, http.StatusInternalServerError, "CANCEL_FAILED", err.Error())
        return
    }

    // Update run status to "cancelled"
    run.Status = "cancelled"
    k8sClient.UpdatePipelineRun(r.Context(), run)

    // Return updated run to client
    dashboard.WriteSuccess(w, run, nil)
}

func canUserCancelRun(user *User, run *PipelineRun) bool {
    // Only allow cancellation by project admin/editor
    return user.ProjectRole(run.ProjectID) == "admin" ||
           user.ProjectRole(run.ProjectID) == "editor"
}
```

**Acceptance**:
- Cancel endpoint stops running pipelines
- Requires admin/editor permissions
- Pipeline status changes to "cancelled"
- Cannot cancel completed pipelines
- Cancel button hidden for non-running pipelines

**Dependencies**: T008 (PipelineRun model), T041 (detail page template)
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/handlers/pipeline_control.go`

**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html` (add Cancel button)
- `/Users/elavigne/workspace/c8s/cmd/api-server/main.go` (register cancel route)

---

## T080e: Write pipeline cancellation tests (per A9) [P]
**Description**: Test cancel endpoint authorization and cancellation logic.

**File**: `/Users/elavigne/workspace/c8s/tests/integration/pipeline_control_test.go` (new)

**Tests**:
```go
func TestCancelPipeline_Success(t *testing.T)           // Running pipeline cancels
func TestCancelPipeline_Unauthorized(t *testing.T)      // Viewer cannot cancel
func TestCancelPipeline_UpdatesStatus(t *testing.T)     // Status changed to "cancelled"
func TestCancelPipeline_CompletedError(t *testing.T)    // Cannot cancel completed run
func TestCancelPipeline_NotFound(t *testing.T)          // 404 for missing run
```

**Acceptance**:
- Cancellation works end-to-end
- Authorization enforced
- Edge cases handled

**Dependencies**: T080d
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/integration/pipeline_control_test.go`

---

# Phase 8: Polish & Cross-Cutting Concerns

## T081: Implement error handling middleware
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

## T081a: Implement keyboard shortcuts (per FR-013) [P]
**Description**: Add keyboard shortcut support per spec.md FR-013. Users can perform common actions (refresh, search, navigate, cancel) using keyboard combinations defined in spec.

**Files**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/keyboard_shortcuts.js` (new)
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/shortcuts_help_modal.html` (new)
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/keyboard_shortcuts.css` (new)

**Implementation**:
```javascript
// keyboard_shortcuts.js
class KeyboardShortcutManager {
  constructor() {
    this.shortcuts = {
      '?': () => this.showHelp(),
      'ctrl+k': () => this.focusSearch(),
      'ctrl+r': () => this.refreshPage(),
      'ctrl+l': () => this.jumpToLatestLog(),
      'escape': () => this.closeModal(),
      'ctrl+enter': () => this.submitForm(),
      'j': () => this.navigateDown(),  // Next run
      'k': () => this.navigateUp(),    // Previous run
      'x': () => this.cancelPipeline(),
      'd': () => this.downloadArtifact(),
      'v': () => this.viewArtifact(),
      'ctrl+slash': () => this.toggleFilterPanel(),
      'ctrl+s': () => this.saveFilterPreset(),
    };
  }

  register() {
    document.addEventListener('keydown', (e) => this.handleKeyPress(e));
  }

  handleKeyPress(event) {
    // Build key combination (e.g., "ctrl+k")
    const key = this.getKeyCombo(event);
    if (this.shortcuts[key]) {
      event.preventDefault();
      this.shortcuts[key]();
    }
  }

  // Implementation of each shortcut action...
}
```

**Shortcuts Implemented** (per spec.md keyboard shortcuts table):
- `?`: Show help modal with all available shortcuts
- `Ctrl/Cmd + K`: Focus search input field
- `Ctrl/Cmd + R`: Reload pipeline/project data
- `Ctrl/Cmd + L`: Scroll to latest log line
- `Esc`: Close currently open modal
- `Ctrl/Cmd + Enter`: Submit active form
- `J`/`K`: Navigate pipeline list (next/previous)
- `X`: Cancel selected running pipeline
- `D`/`V`: Download/view artifact
- `Ctrl/Cmd + /`: Toggle filter panel
- `Ctrl/Cmd + S`: Save filter preset

**Acceptance**:
- All 12+ keyboard shortcuts defined in spec.md work correctly
- Help modal displays context-aware shortcuts (only show relevant shortcuts per page)
- Shortcuts respect platform conventions (Cmd on Mac, Ctrl on Windows/Linux)
- No shortcut interference with text input fields
- Accessibility: All shortcuts have UI/mouse equivalents

**Dependencies**: T034, T041 (page templates)
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/js/keyboard_shortcuts.js`
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/shortcuts_help_modal.html`
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/keyboard_shortcuts.css`

**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/layout/base.html` (include keyboard_shortcuts.js)

---

## T082: Add caching layer for pipeline lists (per A5) [P]
**Description**: Implement in-memory cache for frequently accessed data per FR-015. Support TTL-based expiration and manual invalidation for real-time accuracy.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/cache.go` (new)

**Caching Strategy** (per A5):
```go
package dashboard

import (
    "context"
    "sync"
    "time"
)

type CacheEntry struct {
    Value      interface{}
    ExpiresAt  time.Time
    CreatedAt  time.Time
}

type CacheLayer struct {
    mu              sync.RWMutex
    entries         map[string]*CacheEntry
    defaultTTL      time.Duration
    cleanupInterval time.Duration
}

// NewCacheLayer creates cache with default TTL
func NewCacheLayer(defaultTTL, cleanupInterval time.Duration) *CacheLayer

// Get retrieves value from cache
func (c *CacheLayer) Get(key string) (interface{}, bool)

// Set stores value in cache with default TTL
func (c *CacheLayer) Set(key string, value interface{})

// SetWithTTL stores value with custom TTL
func (c *CacheLayer) SetWithTTL(key string, value interface{}, ttl time.Duration)

// Invalidate removes specific key from cache
func (c *CacheLayer) Invalidate(key string)

// InvalidatePattern removes all keys matching pattern (e.g., "project:*")
func (c *CacheLayer) InvalidatePattern(pattern string)

// StartCleanup runs background goroutine to remove expired entries
func (c *CacheLayer) StartCleanup(ctx context.Context)
```

**Cache Keys & TTL Configuration**:

| Data Type | Cache Key Pattern | Default TTL | Invalidation Trigger |
|-----------|------------------|-------------|----------------------|
| Pipeline list | `pipeline:list:{projectID}` | 5 seconds | New pipeline run created (SSE event) |
| Pipeline run details | `pipeline:run:{runID}` | 10 seconds | Run status changes (SSE event) |
| Project list | `project:list:{userID}` | 30 seconds | New project created |
| Project metadata | `project:{projectID}` | 60 seconds | Project updated |
| Log snapshot | `log:{runID}:{stepID}` | 2 seconds | New log lines streamed |
| User permissions | `user:perms:{userID}` | 300 seconds | User role changed |

**Acceptance**:
- Cache reduces queries to Kubernetes API
- TTL-based expiration prevents stale data
- SSE events trigger cache invalidation for real-time accuracy
- Background cleanup removes expired entries
- Cache hit/miss metrics available for monitoring

**Dependencies**: None
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/cache.go`

---

## T082a: Define cache key design document (per A5) [P]
**Description**: Document cache key naming conventions, TTL values, and invalidation strategy to ensure consistency and correctness.

**File**: `/Users/elavigne/workspace/c8s/docs/CACHE_STRATEGY.md` (new)

**Content**:
- Cache key naming convention: `entity:action:{id}` (e.g., `pipeline:list:proj-123`)
- TTL tiers: Real-time (2-5s), Short-lived (10-30s), Medium (60s), Long (5min+)
- Invalidation patterns: By entity ID, by user ID, by project
- Cache busting: Manual invalidation via SSE events, time-based expiration
- Monitoring: Cache hit/miss ratios, eviction rates, memory usage
- Testing: Cache behavior validation, TTL verification

**Acceptance**:
- Document defines all cache keys used in codebase
- TTL decisions justified (e.g., why pipeline list = 5s, not 30s)
- Invalidation strategy ensures stale-free behavior
- Guide for adding new cached data in future

**Dependencies**: T082
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/docs/CACHE_STRATEGY.md`

---

## T082b: Implement cache invalidation via SSE events (per A5) [P]
**Description**: Connect cache layer to SSE events so cache automatically invalidates when data changes. Ensures real-time accuracy without manual cache busting.

**File**: `/Users/elavigne/workspace/c8s/pkg/dashboard/cache_invalidation.go` (new)

**Implementation**:
```go
// CacheInvalidator listens to SSE events and invalidates related cache entries
type CacheInvalidator struct {
    cache *CacheLayer
    events chan SSEEvent
}

// OnPipelineStatusChanged invalidates pipeline list/detail cache
func (ci *CacheInvalidator) OnPipelineStatusChanged(event SSEEvent)

// OnNewPipelineRun invalidates pipeline list for project
func (ci *CacheInvalidator) OnNewPipelineRun(event SSEEvent)

// OnProjectUpdated invalidates project cache
func (ci *CacheInvalidator) OnProjectUpdated(event SSEEvent)
```

**Acceptance**:
- SSE events trigger cache invalidation
- Pipeline list cache cleared when new run created
- Pipeline detail cache cleared on status change
- Project cache cleared on project update
- No manual cache busting needed

**Dependencies**: T082, T018, T037
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/pkg/dashboard/cache_invalidation.go`

---

## T082c: Write cache layer tests (per A5) [P]
**Description**: Unit and integration tests for cache implementation, TTL behavior, and invalidation.

**Files**:
- `/Users/elavigne/workspace/c8s/tests/unit/cache_test.go` (new)
- `/Users/elavigne/workspace/c8s/tests/integration/cache_integration_test.go` (new)

**Tests**:
```go
// Unit tests
func TestCacheGet_Hits(t *testing.T)              // Cache returns stored value
func TestCacheGet_Misses(t *testing.T)            // Cache misses after TTL expires
func TestCacheSet_WithTTL(t *testing.T)           // Custom TTL respected
func TestCacheInvalidate_ByKey(t *testing.T)      // Specific key invalidated
func TestCacheInvalidate_ByPattern(t *testing.T)  // Pattern-based invalidation works
func TestCacheCleanup_RemovesExpired(t *testing.T) // Background cleanup runs

// Integration tests
func TestCacheInvalidation_OnSSEEvent(t *testing.T)      // SSE events invalidate cache
func TestCacheHitRate_Improves(t *testing.T)            // Cache hit rate metrics
func TestCacheMemory_Bounded(t *testing.T)              // Memory doesn't grow unbounded
```

**Acceptance**:
- Cache TTL behavior validated
- Invalidation strategy tested end-to-end
- Cache hit/miss metrics work
- Memory cleanup prevents leaks

**Dependencies**: T082, T082a, T082b
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/unit/cache_test.go`
- `/Users/elavigne/workspace/c8s/tests/integration/cache_integration_test.go`

---

## T083: Add loading states to HTMX requests
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

## T084: Add keyboard shortcuts
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

## T084a: Write performance benchmarking tests (per A3) [P]
**Description**: Create performance validation tests to verify all SC (success criteria) are met in defined test environment per spec.md Performance Test Environment section.

**Files**:
- `/Users/elavigne/workspace/c8s/tests/performance/benchmark_test.go` (new)
- `/Users/elavigne/workspace/c8s/tests/performance/load_test.go` (new)
- `/Users/elavigne/workspace/c8s/tests/performance/network_simulator.go` (new)
- `/Users/elavigne/workspace/c8s/docs/PERFORMANCE_TESTING.md` (new)

**Benchmarks to Implement**:
```go
// SC-001: Dashboard loads within 2 seconds
func BenchmarkDashboardLoad(b *testing.B)

// SC-002: New pipeline run appears within 5 seconds
func BenchmarkNewRunAppearance(b *testing.B)

// SC-003: Log streaming latency < 2 seconds
func BenchmarkLogStreamingLatency(b *testing.B)

// SC-004: Search/filter on 1000 runs in < 1 second
func BenchmarkSearchPerformance(b *testing.B)

// SC-005: 100+ concurrent users without degradation
func BenchmarkConcurrentUsers(b *testing.B)

// SC-006: Page loads in 3 seconds on standard internet
func BenchmarkPageLoadWithNetworkThrottling(b *testing.B)

// SC-011: Artifact download in < 30 seconds (500MB)
func BenchmarkArtifactDownload(b *testing.B)
```

**Performance Test Environment Setup**:
- Docker Compose or K8s manifest for test infrastructure
- MinIO for S3-compatible storage
- Sample data generation (1000 pipeline runs, logs, artifacts)
- Network simulation via tc (traffic control) for latency/bandwidth throttling
- Chrome DevTools Protocol integration for client-side measurements

**Acceptance**:
- All 7 performance benchmarks pass with target success criteria
- Test results documented with environment configuration
- Baseline metrics captured for regression detection
- Load testing shows no degradation at 100+ concurrent users
- Network-throttled tests validate SC-006 (standard internet conditions)

**Documentation**:
- PERFORMANCE_TESTING.md explains how to run benchmarks
- Includes steps to set up test environment
- Shows how to interpret results
- Provides troubleshooting guidance

**Dependencies**: T001-T085 (all implementation must complete first)
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/performance/benchmark_test.go`
- `/Users/elavigne/workspace/c8s/tests/performance/load_test.go`
- `/Users/elavigne/workspace/c8s/tests/performance/network_simulator.go`
- `/Users/elavigne/workspace/c8s/docs/PERFORMANCE_TESTING.md`

---

## T085a: Write mobile E2E tests (per A15) [P]
**Description**: End-to-end tests for mobile viewports to verify responsive design works correctly per FR-012 and SC-007. Test on iPhone SE (375px) and iPad (768px) viewports.

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/mobile.spec.ts` (new)

**Test Coverage**:
```typescript
// Mobile viewport tests (375px - iPhone SE)
test('mobile: pipeline list responsive', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('http://localhost:8080/dashboard');

    // Verify mobile layout loads
    await expect(page.locator('.pipeline-list')).toBeVisible();
    // Verify no horizontal scroll
    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(375);
});

test('mobile: navigation touch-friendly', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    // Verify touch targets are at least 44x44px
    const buttons = await page.locator('button').all();
    for (const btn of buttons) {
        const box = await btn.boundingBox();
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.height).toBeGreaterThanOrEqual(44);
    }
});

test('mobile: pipeline detail readable', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    // Font sizes should be readable (min 14px for body)
    const computed = await page.evaluate(() =>
        window.getComputedStyle(document.body).fontSize
    );
    expect(parseInt(computed)).toBeGreaterThanOrEqual(14);
});

test('mobile: logs scrollable', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    // Vertical scrolling should work for logs
    await page.goto('http://localhost:8080/dashboard/runs/123');
    const logViewer = page.locator('.log-viewer');
    await logViewer.evaluate(el => el.scrollTop = 100);
    const scrolled = await logViewer.evaluate(el => el.scrollTop);
    expect(scrolled).toBeGreaterThan(0);
});

// Tablet viewport tests (768px - iPad)
test('tablet: two-column layout', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    // Tablet can show side-by-side content
    await page.goto('http://localhost:8080/dashboard');
    // Verify layout optimizes for tablet
    const sidebar = page.locator('.sidebar');
    await expect(sidebar).toBeVisible();
});

// Touch interaction tests
test('mobile: dropdown menu works on touch', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('http://localhost:8080/dashboard');
    // Tap (not click) to open dropdown
    await page.locator('.menu-toggle').tap();
    await expect(page.locator('.dropdown-menu')).toBeVisible();
});

test('mobile: forms work with mobile keyboard', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    // Input fields should trigger mobile keyboard
    const input = page.locator('input[name="search"]');
    await input.focus();
    // Verify input visible above keyboard (viewport shift)
    const box = await input.boundingBox();
    expect(box.y).toBeLessThan(667 * 0.75); // In upper 75% of screen
});
```

**Viewport Sizes Tested**:
- **375px** (iPhone SE, per SC-007 requirement)
- **414px** (iPhone 12)
- **768px** (iPad mini)
- **1024px** (iPad)

**Accessibility for Mobile**:
- Touch targets minimum 44x44px (iOS HIG standard)
- No horizontal scroll on mobile viewports
- Form inputs shift into view above virtual keyboard
- Font sizes readable on small screens (≥14px)
- Tap gestures work (not just click)

**Acceptance**:
- All mobile viewports render correctly
- No layout shifts or overflow
- Touch interactions work
- Page loads within SC-006 (3 seconds) on mobile
- Forms usable on mobile keyboard
- Logs readable and scrollable

**Dependencies**: T034, T041, T073, T080 (all page templates)
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/mobile.spec.ts`

---

## T085b: Add accessibility audit and WCAG compliance (per A16) [P]
**Description**: Run Lighthouse accessibility audit and document WCAG 2.1 Level AA compliance per SC-009 (80+ Lighthouse score requirement).

**File**: `/Users/elavigne/workspace/c8s/tests/e2e/accessibility.spec.ts` (new)

**Implementation**:
```typescript
import { test, expect } from '@playwright/test';
import lighthouse from 'lighthouse';

test('lighthouse: accessibility score >= 80', async ({ page }) => {
    await page.goto('http://localhost:8080/dashboard');

    const options = {
        logLevel: 'info',
        output: 'json',
        onlyCategories: ['accessibility'],
    };

    const runnerResult = await lighthouse(page.url(), options);
    const accessibilityScore = runnerResult.lhr.categories.accessibility.score;

    // SC-009: 80+ Lighthouse score
    expect(accessibilityScore * 100).toBeGreaterThanOrEqual(80);

    // Log issues for remediation
    console.log('Accessibility Audit Results:');
    const auditResults = runnerResult.lhr.audits;
    Object.entries(auditResults).forEach(([key, audit]) => {
        if (audit.score < 1) {
            console.log(`⚠️  ${audit.title}: ${audit.description}`);
        }
    });
});

test('WCAG 2.1: text contrast ratio', async ({ page }) => {
    // Verify WCAG AA minimum contrast (4.5:1 normal, 3:1 large text)
    await page.goto('http://localhost:8080/dashboard');

    const textElements = await page.locator('body *:visible').all();
    for (const el of textElements) {
        const computed = await el.evaluate(e => window.getComputedStyle(e));
        // Should check contrast using color contrast analyzer library
        // e.g., computed.color vs computed.backgroundColor
    }
});

test('WCAG 2.1: keyboard navigation only', async ({ page }) => {
    // Test all functionality reachable via keyboard
    await page.goto('http://localhost:8080/dashboard');

    // Tab through page
    let focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    let tabCount = 0;

    while (focusedElement && tabCount < 50) {
        await page.keyboard.press('Tab');
        focusedElement = await page.evaluate(() => document.activeElement?.tagName);
        tabCount++;
    }

    expect(tabCount).toBeGreaterThan(5); // Many focusable elements
});

test('WCAG 2.1: form labels', async ({ page }) => {
    // All form inputs must have labels
    await page.goto('http://localhost:8080/dashboard/projects');

    const inputs = await page.locator('input').all();
    for (const input of inputs) {
        const id = await input.getAttribute('id');
        const label = page.locator(`label[for="${id}"]`);
        // Or aria-label attribute
        const ariaLabel = await input.getAttribute('aria-label');
        const hasLabel = await label.count() > 0 || ariaLabel;
        expect(hasLabel).toBeTruthy();
    }
});

test('WCAG 2.1: heading hierarchy', async ({ page }) => {
    // Headings must be in logical order
    await page.goto('http://localhost:8080/dashboard');

    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    let lastLevel = 0;

    for (const h of headings) {
        const level = parseInt(await h.evaluate(e => e.tagName[1]));
        // Skip only 1 level max (h1 → h2, h2 → h3, etc.)
        expect(level - lastLevel).toBeLessThanOrEqual(1);
        lastLevel = level;
    }
});

test('WCAG 2.1: alt text for images', async ({ page }) => {
    await page.goto('http://localhost:8080/dashboard');

    const images = await page.locator('img').all();
    for (const img of images) {
        const alt = await img.getAttribute('alt');
        const isDecorative = await img.getAttribute('role') === 'presentation';
        expect(alt || isDecorative).toBeTruthy();
    }
});

test('WCAG 2.1: color not only indicator', async ({ page }) => {
    // Cannot convey information using color alone
    await page.goto('http://localhost:8080/dashboard');

    // Status badges must have text labels, not just color
    const statusBadges = page.locator('[class*="status"]');
    for (const badge of await statusBadges.all()) {
        const text = await badge.textContent();
        expect(text?.trim().length).toBeGreaterThan(0);
    }
});
```

**WCAG 2.1 Level AA Compliance Checklist**:
- ✓ Perceived: Sufficient contrast (4.5:1), readable fonts
- ✓ Operable: Keyboard navigation, no keyboard traps, skip links
- ✓ Understandable: Labels, headings, link text, form instructions
- ✓ Robust: Valid HTML, proper ARIA attributes, semantic markup

**Acceptance**:
- Lighthouse accessibility score ≥80 (per SC-009)
- WCAG 2.1 Level AA compliance verified
- All form inputs have labels
- Keyboard navigation functional
- Color contrast meets standards
- Headings in logical order
- Alt text for all meaningful images

**Documentation**:
- `/Users/elavigne/workspace/c8s/docs/ACCESSIBILITY.md` - WCAG compliance guide
- Remediation items documented and linked to issues

**Dependencies**: T005, T034, T041, T080a (all page templates and styling)
**Story**: Polish
**Files Created**:
- `/Users/elavigne/workspace/c8s/tests/e2e/accessibility.spec.ts`
- `/Users/elavigne/workspace/c8s/docs/ACCESSIBILITY.md`

---

## T085c: Implement step dependency visualization (per A17) [P]
**Description**: Add visual representation of step dependencies on pipeline detail page. Show which steps block which other steps per spec.md:US2 Acceptance Criteria #5.

**File**: `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/step_dependency_graph.html` (new)

**Implementation**:
```html
{{define "step_dependency_graph"}}
<div class="step-dependency-container">
    <h3>Step Dependencies</h3>

    <!-- SVG-based dependency graph -->
    <svg class="dependency-graph" width="800" height="500">
        <!-- Nodes: steps -->
        {{range $i, $step := .Steps}}
        <g class="step-node" id="step-{{.ID}}">
            <rect class="step-box" x="{{mul $i 150}}" y="50"
                  width="120" height="60" rx="5"
                  fill="{{statusColor .Status}}" />
            <text class="step-name" x="{{add (mul $i 150) 60}}" y="85">
                {{.Name}}
            </text>
            <text class="step-status" x="{{add (mul $i 150) 60}}" y="105"
                  font-size="12">
                {{.Status}}
            </text>
        </g>
        {{end}}

        <!-- Edges: dependencies -->
        {{range $i, $step := .Steps}}
        {{range .DependsOn}}
        <line class="dependency-edge"
              x1="{{add (mul (indexOf .ID $.Steps) 150) 120}}" y1="80"
              x2="{{add (mul (indexOf $step.ID $.Steps) 150) 0}}" y2="80"
              stroke="#666" stroke-width="2" marker-end="url(#arrowhead)" />
        {{end}}
        {{end}}

        <!-- Arrow marker -->
        <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="10"
                    refX="9" refY="3" orient="auto">
                <polygon points="0 0, 10 3, 0 6" fill="#666" />
            </marker>
        </defs>
    </svg>

    <!-- Legend -->
    <div class="dependency-legend">
        <div class="legend-item">
            <span class="legend-color" style="background: #4CAF50"></span>
            <span>Succeeded</span>
        </div>
        <div class="legend-item">
            <span class="legend-color" style="background: #FF9800"></span>
            <span>Running</span>
        </div>
        <div class="legend-item">
            <span class="legend-color" style="background: #F44336"></span>
            <span>Failed</span>
        </div>
        <div class="legend-item">
            <span class="legend-color" style="background: #E0E0E0"></span>
            <span>Pending</span>
        </div>
    </div>

    <!-- Dependency Table (text alternative) -->
    <table class="dependency-table">
        <thead>
            <tr>
                <th>Step</th>
                <th>Status</th>
                <th>Depends On</th>
            </tr>
        </thead>
        <tbody>
        {{range .Steps}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Status}}</td>
                <td>
                    {{if .DependsOn}}
                        {{join (map .DependsOn (print "Name")) ", "}}
                    {{else}}
                        None (runs immediately)
                    {{end}}
                </td>
            </tr>
        {{end}}
        </tbody>
    </table>
</div>
{{end}}
```

**CSS Styling** (add to dashboard.css):
```css
.step-dependency-container {
    margin: 20px 0;
    padding: 20px;
    border: 1px solid #ddd;
    border-radius: 4px;
}

.dependency-graph {
    display: block;
    margin: 20px 0;
    border: 1px solid #eee;
    background: #f9f9f9;
}

.step-node {
    cursor: pointer;
    transition: opacity 0.2s;
}

.step-node:hover {
    opacity: 0.8;
}

.step-box {
    stroke: #333;
    stroke-width: 1;
}

.step-name {
    font-weight: bold;
    text-anchor: middle;
    fill: white;
}

.step-status {
    text-anchor: middle;
    fill: white;
}

.dependency-legend {
    display: flex;
    gap: 20px;
    margin-top: 20px;
    font-size: 14px;
}

.legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
}

.legend-color {
    width: 20px;
    height: 20px;
    border-radius: 3px;
}

.dependency-table {
    width: 100%;
    margin-top: 20px;
    border-collapse: collapse;
    font-size: 14px;
}

.dependency-table th, .dependency-table td {
    padding: 12px;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

.dependency-table th {
    background-color: #f5f5f5;
    font-weight: bold;
}
```

**Features**:
- Visual SVG graph showing step nodes and dependency edges
- Color-coded by status (green=succeeded, orange=running, red=failed, gray=pending)
- Interactive hover highlighting
- Text-based table alternative for accessibility
- Legend explaining colors
- Responsive (adapts to container width)

**Acceptance**:
- Dependencies visually displayed on detail page
- Relationship between steps clear (blocking/blocked-by)
- Accessible via keyboard and screen readers
- Works on mobile (table view as fallback)
- Performance: Renders 50+ steps without lag

**Dependencies**: T046 (pipeline detail template), T085b (accessibility)
**Story**: US2
**Files Created**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/partials/step_dependency_graph.html`

**Files Modified**:
- `/Users/elavigne/workspace/c8s/cmd/api-server/static/css/dashboard.css` (add dependency graph styles)
- `/Users/elavigne/workspace/c8s/cmd/api-server/templates/pages/pipeline_detail.html` (include dependency graph)

---

## T085: Write comprehensive E2E test suite
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
- SC-001: Dashboard loads in <2s ✓ (via caching T082)
- SC-003: Log latency <2s ✓ (via SSE T047)
- SC-004: Search <1s ✓ (via indexing T056)
- SC-005: 100+ concurrent users ✓ (via SSE broadcaster T018, caching T082)

**Deliverables**:
- Functional web dashboard integrated with C8S API server
- Real-time pipeline monitoring with SSE
- Search/filter capabilities
- Project and webhook management
- Artifact viewing and download
- Comprehensive test coverage (unit, integration, E2E)
