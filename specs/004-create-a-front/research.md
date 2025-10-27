# Research: HTMX and Go Dashboard Implementation

**Date**: 2025-10-26
**Feature**: Web Dashboard for C8S CI Workflows
**Context**: Implementing a real-time dashboard for C8S using HTMX framework with Go server-side rendering

---

## 1. HTMX Real-Time Updates Architecture

### Decision
Use **Server-Sent Events (SSE)** for real-time log streaming and status updates, with HTMX's SSE extension for client-side handling.

### Rationale
- **Purpose-built for streaming**: SSE is specifically designed for server-to-client unidirectional data streams, perfect for log output and status updates
- **HTMX SSE extension**: Abstracts complexity with automatic reconnection, exponential backoff, and event parsing
- **HTTP-friendly**: Works through proxies and firewalls without custom protocols
- **Go's concurrency model**: Goroutines and channels naturally map to SSE connections (one goroutine per client)
- **Low latency**: Meets success criteria (SC-003) of <2 seconds latency for log updates

### Technical Implementation
- Set SSE headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
- Implement in Go using goroutines per connection with channel-based message broadcasting
- HTMX client-side: `<div hx-ext="sse" sse-connect="/api/logs/stream?runId=123">`
- Use redis pub/sub for distributed deployments across multiple API server instances

### Alternatives Considered

**WebSockets** - REJECTED
- Provides bidirectional communication when only server→client needed
- Higher protocol complexity without corresponding benefit
- Should only use if true bidirectional interaction required (not for CI logs)

**HTTP Polling** - REJECTED
- Polling every 5 seconds × 100 concurrent users = 50 req/s unnecessary load
- Artificial latency minimum equals polling interval
- Inefficient overhead (HTTP headers repeated per request)
- Acceptable only as fallback if SSE unavailable

**Hybrid Approach** - ACCEPTABLE ALTERNATIVE
- SSE for real-time logs and running step status
- HTMX polling with `hx-trigger="every 30s"` for less critical updates (pipeline list refresh)
- Balances server load with real-time requirements

### Recommendation for C8S
**Implement SSE for Phase 1 MVP.** Log streaming and running pipeline updates justify SSE complexity, while future optimizations can implement hybrid approach.

---

## 2. Go HTML Template Organization

### Decision
Use **Go's `html/template` with component-based architecture** (base layouts + partials + pages).

### Rationale
- **Standard library**: Mature, optimized, provides automatic XSS protection via context-aware escaping
- **No extra dependencies**: Aligns with Go philosophy of minimal dependencies
- **Component model**: Use `define` blocks for layouts, inheritance with `block`, and `ParseGlob` for modular templates
- **Production pattern**: Parse templates once at startup, not per-request (performance and error detection)
- **HTMX-friendly**: Excellent for returning HTML fragments (both full pages and HTMX-partial responses)

### Technical Implementation
```
templates/
├── layout/base.html         # Skeleton with blocks for title, content
├── partials/
│   ├── nav.html            # Reusable navigation
│   ├── step_status.html    # Step status indicator
│   └── log_viewer.html     # Log viewer component
└── pages/
    ├── pipeline_list.html
    ├── pipeline_detail.html
    └── projects.html
```

Handler pattern:
```go
// Detect HTMX request header
if r.Header.Get("HX-Request") == "true" {
    // Return fragment only (no HTML shell)
    templates.ExecuteTemplate(w, "partials/pipeline_detail", data)
} else {
    // Return full page
    templates.ExecuteTemplate(w, "base", data)
}
```

### Alternatives Considered

**Templ (Compiled Templates)** - NOT RECOMMENDED for this project
- Provides compile-time type safety and generates Go code from templates
- Adds build complexity and requires learning new DSL
- Better for performance-critical rendering at massive scale
- Not justified for C8S dashboard (performance is adequate with html/template)

**Third-party Engines** (Pongo2, Jet, etc.) - REJECTED
- Unnecessary external dependencies without significant benefits
- Extra learning curve and support burden
- `html/template` is "good enough" for this use case

### Recommendation
**Use html/template with component pattern.** Simplest approach, no dependencies, provides all needed functionality for HTMX rendering.

---

## 3. Real-Time Log Streaming

### Decision
Use **Server-Sent Events (SSE) with HTMX SSE extension** for log streaming.

### Rationale
- Directly streaming logs from running containers requires a persistent connection
- SSE is proven pattern for one-way streaming (logs, notifications, status)
- Go's HTTP streaming with `http.Flusher` provides exactly what needed
- HTMX SSE extension handles client-side complexity (reconnection, event parsing)
- Meets SC-003: <2 seconds latency for live log output

### Technical Implementation
```go
// Server-side (Go)
func StreamLogs(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher := w.(http.Flusher)

    // Subscribe to log channel
    logChan := subscribeToLogs(pipelineRunID)

    for {
        select {
        case <-r.Context().Done():
            return
        case logLine := <-logChan:
            fmt.Fprintf(w, "data: %s\n\n", logLine)
            flusher.Flush()
        }
    }
}
```

```html
<!-- Client-side (HTMX) -->
<div hx-ext="sse" sse-connect="/api/logs/stream?runId=123">
    <div id="log-container" sse-swap="log" hx-swap="beforeend">
        <!-- Log lines appended as they arrive -->
    </div>
</div>
```

### Advanced Patterns
- **Redis pub/sub**: Broadcast logs across multiple API server instances for distributed deployments
- **Backpressure handling**: Buffer logs to prevent slow clients from blocking fast log producers
- **Graceful degradation**: Fall back to polling if SSE connection fails repeatedly

### Alternatives Considered

**WebSocket Streaming** - NOT RECOMMENDED
- Bidirectional protocol when only server→client needed
- Custom message framing adds complexity
- Unnecessary overhead compared to SSE

**HTTP Polling** - FALLBACK ONLY
- Minimum latency equals polling interval (unacceptable for "real-time" logs)
- Creates proportional server load (100 users = 50 req/s minimum)
- Use only as browser compatibility fallback

**WebTransport** - TOO NEW
- Limited browser support as of 2025
- Not appropriate for production requiring broad compatibility

### Recommendation
**SSE is the correct choice for C8S log streaming.** Proven, simple, and efficient for this use case.

---

## 4. Performance Optimization for 100+ Concurrent Users

### Decision
Implement **multi-layer caching** (server-side + HTTP headers + HTMX optimization).

### Rationale
- **HTMX is HTTP-centric**: Works seamlessly with standard HTTP caching (Cache-Control, ETag, If-None-Match)
- **Read-heavy workload**: Pipeline lists and project metadata read frequently but change infrequently → perfect for caching
- **Request optimization essential**: Without debouncing/throttling, HTMX can generate excessive requests that overload backend at scale
- **Multiple layers required**: Single-layer caching (DB-only or HTTP-only) insufficient to meet SC-005 (100+ concurrent users)

### Technical Implementation

**Layer 1: Server-Side Caching (Redis or in-memory)**
```go
func (c *CacheLayer) GetPipelineList(projectID string) ([]PipelineRun, error) {
    key := fmt.Sprintf("project:%s:pipelines", projectID)

    // Try cache
    cached, err := c.redis.Get(ctx, key).Result()
    if err == nil {
        var runs []PipelineRun
        json.Unmarshal([]byte(cached), &runs)
        return runs, nil // Cache hit
    }

    // Cache miss - fetch from DB
    runs, err := c.fetchFromDB(projectID)
    if err != nil {
        return nil, err
    }

    // Store in cache (TTL: 60 seconds for pipeline lists)
    data, _ := json.Marshal(runs)
    c.redis.Set(ctx, key, data, 60*time.Second)

    return runs, nil
}
```

**Layer 2: HTTP Caching Headers**
```go
// Aggressive caching for static assets (1 year)
w.Header().Set("Cache-Control", "public, max-age=31536000")

// Moderate caching for dynamic content (1 minute for pipeline lists)
w.Header().Set("Cache-Control", "private, max-age=60")
w.Header().Set("ETag", calculateETag(data))

// Support conditional requests
if r.Header.Get("If-None-Match") == calculateETag(data) {
    w.WriteHeader(http.StatusNotModified) // Return 304, client uses cached version
    return
}
```

**Layer 3: HTMX Request Optimization**

Debouncing search:
```html
<input type="text" name="search"
       hx-get="/api/pipelines/search"
       hx-trigger="keyup changed delay:500ms"
       hx-target="#results">
```

Throttling status updates:
```html
<div hx-get="/api/pipeline/status"
     hx-trigger="every 1s throttle"
     hx-swap="outerHTML">
</div>
```

Lazy loading with Intersection Observer:
```html
<div hx-get="/api/pipeline-details/123"
     hx-trigger="intersect once"
     hx-swap="innerHTML">
    <div class="loading">Loading...</div>
</div>
```

**Layer 4: CDN for Static Assets**
- Serve CSS, JavaScript, images through CDN (CloudFlare, Fastly, etc.)
- Reduces load on origin servers for global teams

### Success Criteria Mapping
- SC-001 (pipeline load <2s): Server-side cache + HTTP caching
- SC-004 (search/filter <1s): Debouncing + database query optimization
- SC-005 (100+ users): All layers combined prevent bottlenecks
- SC-006 (95% page load <3s): HTTP caching + CDN for static assets

### Alternatives Considered

**Client-side SPA with client-side caching** - REJECTED
- Larger JavaScript bundles, more client memory usage
- Doesn't reduce server load (all requests still reach backend)
- Complexity not justified for C8S dashboard

**No caching** - REJECTED
- Impossible to meet SC-005 without caching
- Database becomes bottleneck under concurrent load

**Aggressive client-side caching only** - INSUFFICIENT
- Doesn't reduce server load (every user still hits origin)
- Must implement server-side caching for scalability

### Recommendation
**Implement all four layers.** This is the proven pattern for scalable HTMX applications. Server-side caching is essential; HTTP headers and HTMX optimization provide multiplicative benefits.

---

## 5. Testing Strategy for HTMX Components

### Decision
**Three-layer testing approach**: Unit tests (Go httptest), Integration tests (multi-handler workflows), E2E tests (Playwright).

### Rationale
- **Go httptest for units**: Fast, isolated, no network overhead. HTMX detection is header-based (`HX-Request: true`), trivial to test
- **HTMX responses are HTTP fragments**: Can be tested identically to any HTTP response (just test for HTML fragment, not full page)
- **Playwright for E2E** (not Cypress): Better multi-browser support (Chrome, Firefox, Safari), faster parallel execution, better TypeScript integration
- **Three layers needed**: Unit tests catch bugs early and run fast; integration tests verify handlers work together; E2E tests verify complete user workflows with real browser

### Technical Implementation

**Unit Testing (Go httptest)**
```go
func TestPipelineListHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/pipelines", nil)
    req.Header.Set("HX-Request", "true") // HTMX request

    rr := httptest.NewRecorder()
    PipelineListHandler(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), `<div id="pipeline-list">`)
    assert.NotContains(t, rr.Body.String(), "<html>") // Fragment, not full page
}

func TestPipelineListHandler_FullPage(t *testing.T) {
    req := httptest.NewRequest("GET", "/pipelines", nil)
    // No HX-Request header - standard request

    rr := httptest.NewRecorder()
    PipelineListHandler(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "<!DOCTYPE html>") // Full page
}
```

**Integration Testing (httptest with test server)**
```go
func TestDashboardWorkflow(t *testing.T) {
    // Set up test router
    mux := http.NewServeMux()
    mux.HandleFunc("/pipelines", PipelineListHandler)
    mux.HandleFunc("/pipeline/", PipelineDetailHandler)

    ts := httptest.NewServer(mux)
    defer ts.Close()

    // Test workflow: list → detail
    resp, err := http.Get(ts.URL + "/pipelines")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // Parse response, extract pipeline ID, fetch detail
    body, _ := io.ReadAll(resp.Body)
    pipelineID := extractIDFromHTML(body)

    resp2, err := http.Get(ts.URL + "/pipeline/" + pipelineID)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp2.StatusCode)
}
```

**E2E Testing (Playwright)**
```typescript
import { test, expect } from '@playwright/test';

test('pipeline list updates in real-time via SSE', async ({ page }) => {
  await page.goto('http://localhost:8080/dashboard');

  // Initial state
  await expect(page.locator('#pipeline-list')).toBeVisible();
  const initialCount = await page.locator('.pipeline-row').count();

  // Simulate new pipeline trigger (via API)
  await triggerPipelineViaAPI(123);

  // Wait for SSE update (should appear within 2 seconds)
  await page.waitForTimeout(2000);

  // Verify new pipeline appeared without page refresh
  const updatedCount = await page.locator('.pipeline-row').count();
  expect(updatedCount).toBe(initialCount + 1);
});

test('log streaming displays in real-time', async ({ page }) => {
  await page.goto('http://localhost:8080/pipeline/run-123');

  const logContainer = page.locator('#log-container');
  const initialLogCount = await logContainer.locator('.log-line').count();

  // Wait for SSE logs to stream
  await page.waitForTimeout(3000);

  const updatedLogCount = await logContainer.locator('.log-line').count();
  expect(updatedLogCount).toBeGreaterThan(initialLogCount);

  // Verify logs are in chronological order
  const lastLog = await logContainer.locator('.log-line').last().textContent();
  expect(lastLog).toMatch(/\d{2}:\d{2}:\d{2}/); // Timestamp format
});
```

### Testing Strategy Summary

| Test Type | Tool | Coverage | Speed | Purpose |
|-----------|------|----------|-------|---------|
| Unit | Go httptest | Individual handler logic, template rendering | Fast (ms) | Verify handlers work in isolation |
| Integration | Go httptest + test server | Multi-handler workflows, data flow | Medium (100ms) | Verify handlers work together correctly |
| E2E | Playwright | Real browser, SSE, HTMX events, user workflows | Slow (seconds) | Verify complete user journeys end-to-end |

### Alternatives Considered

**Cypress for E2E** - REJECTED (2025)
- JavaScript-only test language (Playwright supports Go, Python, JS)
- Historically better Chrome support, weaker Safari/WebKit support
- Slower test execution compared to Playwright
- Use if team already has Cypress expertise, otherwise Playwright is superior

**Manual testing only** - REJECTED
- Real-time features (SSE, polling) error-prone to test manually
- No regression protection as code evolves
- Time-consuming and unreliable for complex workflows

**Selenium** - DEPRECATED
- Older generation browser automation
- Slower, less reliable, worse developer experience
- Use only in legacy projects; don't start new projects with Selenium

**Unit testing HTMX in isolation (mock DOM)** - IMPRACTICAL
- HTMX components depend on HTML structure and server responses
- Isolating them defeats the purpose (most bugs are in integration)
- Focus on integration and E2E tests instead

### Supporting Libraries
- **go-htmx**: Go constants and helpers for HTMX headers (`github.com/donseba/go-htmx`)
- **htmx-go**: Alternative Go helpers (`github.com/angelofallars/htmx-go`)
- **sse**: Production SSE library for Go (`github.com/r3labs/sse`)

### Recommendation
**Implement all three layers.** Unit tests catch bugs early (fast feedback), integration tests verify components work together, E2E tests verify complete workflows. This provides comprehensive coverage while maintaining reasonable CI/CD pipeline speed.

---

## 6. HTMX Extensions and Go Libraries

### Recommended Extensions
- **hx-ext="sse"**: Essential for log streaming (core to MVP)
- **hx-ext="preload"**: Preload linked pages on hover (improves perceived performance, Phase 2)
- **hx-ext="loading-states"**: Show loading indicators during requests (improves UX, Phase 2)

### Recommended Go Libraries
- **github.com/donseba/go-htmx**: Provides HTMX header constants and helper functions
- **github.com/r3labs/sse**: Production-ready Server-Sent Events library
- **github.com/go-chi/chi**: Router (already used in C8S API server)

---

## Summary: Technology Decision Matrix

| Concern | Decision | Why | Alternative | Why Rejected |
|---------|----------|-----|-------------|--------------|
| Real-time updates | Server-Sent Events (SSE) | Optimal for unidirectional streaming; simpler than WebSockets | WebSockets | Bidirectional overhead unnecessary |
| Template engine | Go html/template + components | Standard library, XSS protection, no deps | Templ | Adds complexity without proportional benefit |
| Log streaming | SSE with HTMX extension | Low latency (<2s), automatic reconnection | WebSocket; Polling | Complex; inefficient respectively |
| Performance | Multi-layer caching | Meets 100+ users requirement | No caching | Database becomes bottleneck |
| Testing | httptest + Playwright | Fast units; comprehensive E2E | Cypress; Manual | Less capable; unreliable |

---

## Next Steps

1. **Phase 1a**: Create `data-model.md` defining entities (Project, PipelineRun, PipelineStep, Artifact, User)
2. **Phase 1b**: Create `contracts/api-schema.md` defining REST endpoints and response structures
3. **Phase 1c**: Create `quickstart.md` with developer setup guide
4. **Phase 2**: Generate `tasks.md` with implementation tasks ordered by dependencies
5. **Implementation**: Begin with P1 user stories (pipeline list and detail views)

---

**Research completed**: All NEEDS CLARIFICATION items resolved. Foundation established for Phase 1 design artifacts.
