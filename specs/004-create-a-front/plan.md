# Implementation Plan: Web Dashboard for C8S CI Workflows

**Branch**: `004-create-a-front` | **Date**: 2025-10-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-create-a-front/spec.md`

## Summary

Implement a real-time web dashboard for C8S that enables developers to monitor pipeline execution, view logs, manage projects, and access artifacts. The dashboard will be built using HTMX with server-side rendering for dynamic updates without requiring a heavy JavaScript framework. The API server will provide RESTful endpoints for pipeline data, logs, and configuration. HTMX handles real-time updates through WebSocket/polling, while the server generates HTML templates with dynamic content.

**Key approach**: Server-driven architecture where the API server (existing Go service) renders HTML templates that HTMX gradually enhances with real-time capabilities (live logs, status updates, search/filtering).

## Technical Context

**Language/Version**: Go 1.24.0 (backend API server), HTML5/CSS3/JavaScript (frontend with HTMX)
**Primary Dependencies**:
- Backend: chi (routing), already used in C8S API server
- Frontend: HTMX (dynamic updates), Tauri/htmx-extensions (optional WebSocket support)
- Template engine: Go's html/template (renders pages server-side)

**Storage**: Uses existing C8S infrastructure (Kubernetes, S3-compatible object storage for logs/artifacts)
**Testing**:
- Backend: Go testing (unit), integration tests for API endpoints
- Frontend: Browser automation tests (Playwright/Selenium) for HTMX interactions
**Target Platform**: Web browser (Chrome, Firefox, Safari, Edge)
**Project Type**: Web application with server-side rendering
**Performance Goals**:
- Dashboard loads in <3 seconds
- Real-time log updates: <2 seconds latency
- Search/filter on 1000 runs: <1 second response
- Support 100+ concurrent users
**Constraints**:
- Must work without JavaScript (graceful degradation)
- Real-time features optional but preferred
- Session security via existing auth mechanism
**Scale/Scope**:
- ~50-100 screens/components
- 5 main user stories (P1: 2, P2: 2, P3: 1)
- Estimated Phase 1 (MVP): P1 + P2 stories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Principle I - Specification-First Development**: ✅ PASS
- Feature has complete specification with user scenarios, requirements, success criteria

**Principle II - User Story-Driven Architecture**: ✅ PASS
- 5 prioritized user stories (P1, P2, P3) each independently testable and deployable

**Principle III - Constitution Gates**: ⏳ PENDING RE-CHECK
- Architecture complexity justified: HTMX + Go templates is simpler than building a full SPA
- No architectural violations anticipated

**Principle IV - Test Independence**: ✅ PASS
- Each user story can be tested independently (pipeline list, detail view, search, configuration, artifacts)

**Principle V - Documentation as Artifact**: ✅ PASS
- Plan.md, research.md, data-model.md, contracts/ all generated through this workflow
- Version-controlled alongside code

## Project Structure

### Documentation (this feature)

```
specs/004-create-a-front/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - HTMX/Go architecture research
├── data-model.md        # Phase 1 output - entities and data structures
├── quickstart.md        # Phase 1 output - developer quickstart guide
├── contracts/           # Phase 1 output - API schemas
│   ├── api-schema.md    # Dashboard API endpoints
│   └── websocket-spec.md # Real-time log streaming spec
├── checklists/requirements.md  # Specification quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code Structure

**Selected: Integrated with existing C8S API server**

The dashboard will extend the existing C8S API server (cmd/api-server) rather than being a separate project:

```
cmd/api-server/
├── main.go                      # Existing API server
├── handlers/
│   ├── dashboard.go             # NEW: Dashboard page handlers
│   ├── pipeline_runs.go         # NEW: Pipeline run API endpoints
│   ├── logs.go                  # NEW: Log streaming handlers
│   ├── artifacts.go             # NEW: Artifact listing/download
│   └── projects.go              # NEW: Project configuration handlers
├── templates/                   # NEW: HTML templates for dashboard
│   ├── layout.html              # Base layout
│   ├── pipeline_list.html       # Pipeline runs list
│   ├── pipeline_detail.html     # Pipeline run detail view
│   ├── logs.html                # Log viewer component
│   ├── artifacts.html           # Artifacts list
│   ├── projects.html            # Projects management
│   └── components/              # Reusable HTMX components
│       ├── step_status.html
│       ├── log_viewer.html
│       └── filter_panel.html
└── static/                      # NEW: Static assets
    ├── css/
    │   └── dashboard.css        # Dashboard styling
    ├── js/
    │   └── htmx.min.js          # HTMX library
    └── img/                     # Icons, logos

pkg/api/
├── dashboard.go                 # NEW: Dashboard service logic
└── [...existing handlers...]

tests/
├── integration/
│   ├── dashboard_test.go        # NEW: Dashboard endpoint tests
│   ├── pipeline_runs_test.go    # NEW: Pipeline API tests
│   └── [...existing...]
└── unit/
    └── [...existing...]
```

**Structure Decision**: Integrated approach extends the existing API server with dashboard-specific handlers and templates. This avoids duplication of auth, database access, and core logic. HTMX components are rendered server-side, reducing JavaScript complexity while maintaining dynamic interactivity.

## Complexity Tracking

*No constitution violations - architecture decisions fully justified:*

| Decision | Justification | Alternatives |
|----------|---------------|--------------|
| Integrated dashboard (not separate frontend service) | Reuses auth, Kubernetes access, and database connections from existing API server | Separate frontend would duplicate auth logic and require new deployment target |
| HTMX + Go templates (not SPA) | Simpler architecture with less JavaScript, faster initial page load, works without JS | React/Vue SPAs require more build tooling and client-side state management |
| Server-side rendering (not client-side) | Leverages Go's strengths; reduces frontend complexity; enables progressive enhancement | Client-side rendering requires API-first architecture and more complex state management |
