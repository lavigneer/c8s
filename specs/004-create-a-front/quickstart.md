# Quickstart: C8S Web Dashboard Development

**For**: Developers implementing the C8S web dashboard with HTMX + Go

---

## Prerequisites

- **Go 1.24+**: For building the API server with dashboard handlers
- **Git**: For cloning the repository
- **curl** or **Postman**: For testing API endpoints
- **Docker** (optional): For S3-compatible storage testing
- **k3d cluster**: For testing with actual Kubernetes integration

---

## Architecture Overview

The C8S dashboard is **not** a separate frontend application. Instead, it's integrated into the existing **API server** (`cmd/api-server`) with:

1. **Go HTTP handlers** that render HTML templates
2. **HTML templates** (Go's `html/template`) with HTMX enhancements
3. **Static assets** (CSS, JavaScript) served alongside templates
4. **Server-Sent Events (SSE)** for real-time log streaming and updates

**Key principle**: Server-driven HTML generation with HTMX for progressive enhancement. No heavy SPA framework (React, Vue) needed.

---

## Project Structure

```
cmd/api-server/
├── main.go                 # Existing API server entry point
├── handlers/
│   ├── dashboard.go        # NEW: Dashboard page handlers
│   ├── pipeline_runs.go    # NEW: Pipeline run endpoints
│   ├── logs.go             # NEW: Log streaming (SSE)
│   ├── artifacts.go        # NEW: Artifact listing/download
│   └── projects.go         # NEW: Project config endpoints
├── templates/              # NEW: HTML templates
│   ├── layout/
│   │   └── base.html       # Base HTML skeleton
│   ├── partials/
│   │   ├── nav.html        # Navigation
│   │   ├── step_status.html
│   │   └── log_viewer.html
│   └── pages/
│       ├── pipeline_list.html
│       ├── pipeline_detail.html
│       ├── logs.html
│       ├── artifacts.html
│       └── projects.html
└── static/                 # NEW: Static assets
    ├── css/dashboard.css
    ├── js/htmx.min.js
    └── img/

pkg/api/
├── dashboard.go            # NEW: Dashboard service logic
└── [existing packages...]
```

---

## Step 1: Set Up Development Environment

### Clone and Navigate

```bash
cd /Users/elavigne/workspace/c8s
git checkout 004-create-a-front
```

### Install Dependencies

The dashboard requires HTMX library (client-side) and potentially the `go-htmx` helper library:

```bash
# Download HTMX library (via npm or CDN)
# For development, you can use the CDN URL in templates:
# <script src="https://unpkg.com/htmx.org"></script>

# Or install locally via Go assets (optional):
go get github.com/donseba/go-htmx
go get github.com/r3labs/sse
```

---

## Step 2: Create Base HTML Template

Create `cmd/api-server/templates/layout/base.html`:

```html
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}C8S Dashboard{{end}}</title>
    <link rel="stylesheet" href="/static/css/dashboard.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        /* Minimal base styling */
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .navbar { background: #f5f5f5; padding: 10px 20px; border-bottom: 1px solid #ddd; }
        .sidebar { float: left; width: 250px; margin-right: 20px; }
        .main { margin-left: 270px; }
    </style>
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

---

## Step 3: Create Navigation Partial

Create `cmd/api-server/templates/partials/nav.html`:

```html
{{define "partials/nav"}}
<nav class="navbar">
    <div class="container">
        <h1><a href="/dashboard">C8S Dashboard</a></h1>
        <ul style="list-style: none; margin: 0; padding: 0;">
            <li><a href="/dashboard/projects">Projects</a></li>
            <li><a href="/dashboard">Pipelines</a></li>
            <li><a href="/user/profile">Profile</a></li>
            <li><a href="/logout">Logout</a></li>
        </ul>
    </div>
</nav>
{{end}}
```

---

## Step 4: Create Pipeline List Template

Create `cmd/api-server/templates/pages/pipeline_list.html`:

```html
{{define "content"}}
<h2>Pipeline Runs</h2>

<div class="filters" style="margin-bottom: 20px;">
    <input type="text"
           name="search"
           placeholder="Search by commit SHA..."
           hx-get="/api/projects/{{.ProjectId}}/runs"
           hx-target="#pipeline-list"
           hx-trigger="keyup changed delay:500ms"
           hx-include="[name='branch'],[name='status']">

    <select name="branch"
            hx-get="/api/projects/{{.ProjectId}}/runs"
            hx-target="#pipeline-list"
            hx-trigger="change"
            hx-include="[name='search'],[name='status']">
        <option value="">All Branches</option>
        {{range .Branches}}
            <option value="{{.}}">{{.}}</option>
        {{end}}
    </select>

    <select name="status"
            hx-get="/api/projects/{{.ProjectId}}/runs"
            hx-target="#pipeline-list"
            hx-trigger="change"
            hx-include="[name='search'],[name='branch']">
        <option value="">All Statuses</option>
        <option value="succeeded">Succeeded</option>
        <option value="failed">Failed</option>
        <option value="running">Running</option>
    </select>
</div>

<div id="pipeline-list"
     hx-ext="sse"
     sse-connect="/api/projects/{{.ProjectId}}/runs/updates"
     style="border: 1px solid #ddd; border-radius: 4px;">

    {{range .PipelineRuns}}
        <div class="pipeline-row"
             style="padding: 15px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center;">

            <div style="flex: 1;">
                <div style="font-weight: bold;">{{.Name}}</div>
                <div style="color: #666; font-size: 0.9em;">
                    {{.CommitSha | slice 0 7}} on {{.Branch}}
                    {{if eq .Status "running"}}
                        <span style="color: #ff9800;">⏳ Running</span>
                    {{else if eq .Status "succeeded"}}
                        <span style="color: #4caf50;">✓ Succeeded</span>
                    {{else if eq .Status "failed"}}
                        <span style="color: #f44336;">✗ Failed</span>
                    {{end}}
                </div>
            </div>

            <div style="text-align: right; color: #666; font-size: 0.9em;">
                <div>{{.TriggeredAt | formatTime}}</div>
                {{if .DurationSeconds}}
                    <div>{{.DurationSeconds}}s</div>
                {{end}}
            </div>

            <a href="/dashboard/runs/{{.Id}}"
               class="btn"
               style="margin-left: 20px; padding: 8px 16px; background: #2196f3; color: white; text-decoration: none; border-radius: 4px;">
                View Details
            </a>
        </div>
    {{end}}
</div>
{{end}}
```

---

## Step 5: Create Go Handler

Create `cmd/api-server/handlers/dashboard.go`:

```go
package handlers

import (
    "net/http"
    "text/template"

    "github.com/org/c8s/pkg/api"
)

var templates *template.Template

func init() {
    // Parse all templates at startup
    var err error
    templates, err = template.ParseGlob("cmd/api-server/templates/**/*.html")
    if err != nil {
        panic(err)
    }
}

// DashboardHandler renders the main dashboard page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    // Get user from context (assumes auth middleware set this)
    user, ok := r.Context().Value("user").(*api.User)
    if !ok {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        return
    }

    projectId := r.URL.Query().Get("projectId")
    if projectId == "" {
        // Redirect to first project
        projects, err := api.ListProjectsForUser(user.ID)
        if err != nil || len(projects) == 0 {
            http.Error(w, "No projects found", http.StatusNotFound)
            return
        }
        projectId = projects[0].ID
    }

    // Fetch data
    runs, err := api.ListPipelineRuns(projectId, 1, 20)
    if err != nil {
        http.Error(w, "Failed to fetch runs", http.StatusInternalServerError)
        return
    }

    // Check if HTMX partial request
    if r.Header.Get("HX-Request") == "true" {
        // Return fragment only
        templates.ExecuteTemplate(w, "content", map[string]interface{}{
            "ProjectId":    projectId,
            "PipelineRuns": runs,
        })
    } else {
        // Return full page
        templates.ExecuteTemplate(w, "base", map[string]interface{}{
            "ProjectId":    projectId,
            "PipelineRuns": runs,
        })
    }
}

// LogStreamHandler streams logs via Server-Sent Events
func LogStreamHandler(w http.ResponseWriter, r *http.Request) {
    runId := r.URL.Query().Get("runId")
    stepId := r.URL.Query().Get("stepId")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    // Subscribe to log channel
    logChan := api.SubscribeToLogs(runId, stepId)
    defer api.UnsubscribeFromLogs(runId, stepId, logChan)

    for {
        select {
        case <-r.Context().Done():
            return
        case logLine := <-logChan:
            if logLine == "" { // Completion signal
                fmt.Fprintf(w, "event: complete\ndata: {}\n\n")
                flusher.Flush()
                return
            }
            fmt.Fprintf(w, "event: log\ndata: %s\n\n", logLine)
            flusher.Flush()
        }
    }
}
```

---

## Step 6: Register Routes

In `cmd/api-server/main.go`, register the dashboard handlers:

```go
import (
    "github.com/org/c8s/cmd/api-server/handlers"
)

func main() {
    // ... existing setup ...

    router := chi.NewRouter()

    // Dashboard routes
    router.Get("/dashboard", handlers.DashboardHandler)
    router.Get("/dashboard/runs/{runId}", handlers.PipelineDetailHandler)
    router.Get("/api/projects/{projectId}/runs", handlers.ListPipelineRunsHandler)
    router.Get("/api/runs/{runId}/steps/{stepId}/logs", handlers.LogStreamHandler)

    // ... rest of routes ...

    http.ListenAndServe(":8080", router)
}
```

---

## Step 7: Test the Dashboard

### Start the API Server

```bash
go run cmd/api-server/main.go
```

### Access the Dashboard

Open browser to `http://localhost:8080/dashboard`

You should see:
- Pipeline list with real-time status updates
- Search/filter controls
- Real-time log streaming when viewing a running pipeline

---

## Step 8: Testing

### Unit Test Example

```go
// tests/integration/dashboard_test.go
package integration

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestDashboardHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/dashboard", nil)
    rr := httptest.NewRecorder()

    handlers.DashboardHandler(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    expected := `<h2>Pipeline Runs</h2>`
    if !strings.Contains(rr.Body.String(), expected) {
        t.Errorf("handler returned unexpected body")
    }
}
```

### E2E Test Example (Playwright)

```typescript
// tests/e2e/dashboard.spec.js
import { test, expect } from '@playwright/test';

test('pipeline list displays', async ({ page }) => {
  await page.goto('http://localhost:8080/dashboard');

  await expect(page.locator('h2:has-text("Pipeline Runs")')).toBeVisible();
  await expect(page.locator('.pipeline-row')).toHaveCount(20);
});

test('search filters pipelines', async ({ page }) => {
  await page.goto('http://localhost:8080/dashboard');

  await page.fill('input[name="search"]', 'abc123');
  await page.waitForTimeout(600); // Wait for debounce

  const rows = await page.locator('.pipeline-row').count();
  expect(rows).toBeLessThan(20); // Should be filtered
});
```

---

## Common Development Tasks

### Add a New Dashboard Page

1. Create template: `cmd/api-server/templates/pages/my_page.html`
2. Create handler: Add function to `handlers/my_handler.go`
3. Register route: Add to router in `main.go`
4. Test: Write unit + E2E tests

### Update Template Styling

1. Edit `cmd/api-server/static/css/dashboard.css`
2. Restart server (go reload will auto-reload templates)
3. Refresh browser

### Add Real-Time Updates

1. Use HTMX extensions: `hx-ext="sse"`
2. Connect to SSE endpoint: `sse-connect="/api/..."`
3. Define SSE handler in Go (see LogStreamHandler example)

---

## Debugging Tips

### HTMX Request Headers

HTMX adds these headers to requests - use them to differentiate partial vs full-page renders:

```go
if r.Header.Get("HX-Request") == "true" {
    // Partial request - return fragment only
} else {
    // Full page request
}
```

### Enable HTMX Debug Mode

Add to HTML template:

```html
<script>
    htmx.config.logAll = true;  // Log all events
    htmx.logAll();               // Enable verbose logging
</script>
```

### Monitor SSE Connection

In browser DevTools Network tab, look for `text/event-stream` requests. Click to see incoming events in real-time.

### Template Parse Errors

If templates fail to load:

```bash
# Check template syntax
go build ./cmd/api-server
```

Errors will indicate which template file has syntax issues.

---

## Next Steps

1. **Phase 1 MVP**: Implement P1 user stories (pipeline list + detail)
2. **Phase 2**: Add search/filtering (P2 user stories)
3. **Phase 3**: Add artifact management and project configuration
4. **Optimization**: Add caching, CDN, performance tuning

See `tasks.md` for implementation task breakdown and dependencies.

---

## References

- **HTMX Documentation**: https://htmx.org/docs/
- **Go html/template**: https://pkg.go.dev/html/template
- **Server-Sent Events**: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
- **C8S Architecture**: [../../../README.md](../../../README.md)

---

**Happy coding!** 🚀
