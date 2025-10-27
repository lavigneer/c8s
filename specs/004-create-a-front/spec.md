# Feature Specification: Web Dashboard for C8S CI Workflows

**Feature Branch**: `004-create-a-front`
**Created**: 2025-10-26
**Status**: Draft
**Input**: User description: "Create a front-end for the c8s CI workflows, logics, and projects"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Pipeline History and Current Status (Priority: P1)

A developer needs to understand the status of their project's pipelines - which ones are currently running, which recently completed successfully, and which failed. They access a dashboard that shows a list of all pipeline runs for their project, with each run displaying its status (running/passed/failed), when it started, how long it took, and which commit triggered it. The developer can click into any run for more details.

**Why this priority**: Pipeline visibility is essential for developers to track their work and identify failures immediately. Without this, developers must rely on notifications or polling the CLI, reducing visibility and responsiveness.

**Independent Test**: Can be fully tested by creating several pipeline runs in the system, accessing the dashboard, and verifying that all runs appear with correct status, timestamps, and commit information. Delivers immediate value by centralizing pipeline visibility in one place.

**Acceptance Scenarios**:

1. **Given** multiple pipeline runs exist for a project, **When** a developer accesses the dashboard, **Then** they see a list of recent runs sorted by start time (newest first)
2. **Given** a pipeline is currently executing, **When** viewing the dashboard, **Then** the active run displays a running status with elapsed time updating in real-time
3. **Given** a pipeline has completed, **When** viewing the dashboard, **Then** the run shows final status (passed/failed), total duration, and timestamp
4. **Given** a developer views a pipeline run list, **When** a new run is triggered, **Then** it appears at the top of the list without requiring a page refresh
5. **Given** multiple projects exist, **When** a developer navigates to a project's dashboard, **Then** they see only pipeline runs for that specific project

---

### User Story 2 - Monitor Step-by-Step Execution and Logs (Priority: P1)

A developer has triggered a pipeline and wants to see which steps are currently executing, how long each step is taking, and what the output is. They click on a running pipeline to see a detailed view showing all pipeline steps, their individual status (pending/running/succeeded/failed), execution time, and a live log viewer showing output from the currently active step. If a step fails, the developer immediately sees the error output.

**Why this priority**: Real-time step visibility and logs are critical for debugging failures and understanding execution flow. This is often the most time-consuming part of pipeline debugging, so providing comprehensive visibility directly in the dashboard saves significant developer time.

**Independent Test**: Can be tested by running a multi-step pipeline and accessing the detailed pipeline view, verifying that step status updates in real-time, logs stream live, and step timings are accurate. Delivers value by enabling faster debugging of pipeline failures.

**Acceptance Scenarios**:

1. **Given** a pipeline run is executing, **When** viewing the pipeline detail page, **Then** all pipeline steps are displayed with their current status
2. **Given** a step is currently executing, **When** viewing the log section, **Then** live logs from that step stream continuously without requiring manual refresh
3. **Given** a step fails, **When** viewing the pipeline detail page, **Then** the failed step is clearly highlighted and its error output is visible
4. **Given** a step has completed, **When** scrolling through logs, **Then** the developer can see all output from that step including stdout and stderr
5. **Given** a pipeline contains dependent steps, **When** viewing the detail page, **Then** the dependency relationships are visible (e.g., which steps block which other steps)

---

### User Story 3 - Search and Filter Pipelines (Priority: P2)

A developer needs to find a specific pipeline run from several weeks ago or filter to see only failed runs for a particular branch. The dashboard provides search and filter capabilities allowing the developer to find pipelines by commit hash, branch name, status (passed/failed), or date range. Results update dynamically as filters are applied.

**Why this priority**: Search and filtering become essential as teams run hundreds of pipelines weekly. Developers frequently need to correlate a specific commit with its pipeline results or identify patterns in failures for a particular branch.

**Independent Test**: Can be tested by creating multiple pipeline runs with different statuses and branches, applying various filters and search criteria, and verifying that results are accurate and update in real-time. Delivers value by reducing time spent manually browsing pipeline history.

**Acceptance Scenarios**:

1. **Given** the pipeline list is displayed, **When** a developer enters a commit SHA in the search field, **Then** the list filters to show only runs for that commit
2. **Given** a developer wants to view only failures, **When** they select "Status: Failed" filter, **Then** only failed pipeline runs are displayed
3. **Given** a developer wants to review a specific branch, **When** they select a branch from the branch filter, **Then** only runs for that branch appear
4. **Given** multiple filters are applied, **When** filters are updated, **Then** results update dynamically without requiring a page reload
5. **Given** a developer needs historical data, **When** they select a date range, **Then** only runs within that date range are displayed

---

### User Story 4 - Configure Projects and Webhooks (Priority: P2)

An organization administrator needs to set up a project in C8S and connect it to their Git repository. They access a project configuration page where they can create a new project, specify the Git repository URL, and view the webhook URL that needs to be registered with their Git platform. The administrator registers the webhook with GitHub/GitLab/Bitbucket, and subsequent pushes automatically trigger pipelines.

**Why this priority**: Without the ability to configure projects through the UI, administrators must use the CLI or direct API calls. A configuration interface makes C8S more accessible and reduces setup friction for new teams. However, the system can function initially with CLI-based setup.

**Independent Test**: Can be tested by creating a new project through the dashboard, obtaining the webhook URL, and verifying that new commits trigger pipeline runs. Delivers value by reducing setup friction and centralizing configuration management.

**Acceptance Scenarios**:

1. **Given** an administrator accesses the projects section, **When** they click "Create Project", **Then** a form appears with fields for project name and repository URL
2. **Given** a project is created, **When** viewing the project settings, **Then** the webhook URL is displayed and can be copied to clipboard
3. **Given** a webhook is registered with a Git platform, **When** a developer pushes code, **Then** a pipeline run is automatically created without manual intervention
4. **Given** an administrator has multiple projects, **When** viewing the projects list, **Then** they can see all projects with their status and last run information
5. **Given** a project needs to be deleted, **When** an administrator requests deletion, **Then** all associated data is cleaned up and the project no longer accepts webhooks

---

### User Story 5 - View and Manage Artifacts and Outputs (Priority: P3)

A developer has run a pipeline that produces build artifacts (compiled binaries, Docker images, test reports) and wants to download or inspect these artifacts. They access the pipeline run detail page, see a list of artifacts generated by the pipeline, and can download them or view their metadata (size, type, creation time). For certain artifact types (like test reports), the dashboard renders them directly in the browser.

**Why this priority**: Artifact management is valuable for sharing build outputs and test reports, but the system can initially function without this feature. Pipelines can still generate artifacts; developers just need alternative methods to access them. This feature improves user experience but isn't blocking for core functionality.

**Independent Test**: Can be tested by running a pipeline configured to produce artifacts, accessing the artifacts list in the dashboard, downloading an artifact, and verifying its contents. Delivers value by centralizing artifact access in the dashboard and reducing reliance on external storage browsing tools.

**Acceptance Scenarios**:

1. **Given** a pipeline step produces artifacts, **When** viewing the pipeline run detail page, **Then** all generated artifacts are listed with their names, sizes, and types
2. **Given** an artifact is listed, **When** a developer clicks the artifact, **Then** they can download it to their local machine
3. **Given** a pipeline produces a test report artifact, **When** viewing the pipeline detail page, **Then** the test report is rendered directly in the dashboard showing test results
4. **Given** multiple artifacts exist, **When** a developer wants to batch download them, **Then** they can select multiple artifacts and download as a compressed archive
5. **Given** artifacts are retained in storage, **When** accessing a run from several months ago, **Then** all artifacts are still available for download

---

### Edge Cases

- **Long-running pipelines**: Dashboard displays elapsed time and allows users to cancel stuck pipelines per spec.md US2 edge case. System shows last-updated timestamp to distinguish active vs. stalled executions.
- **Network interruption**: Dashboard gracefully handles disconnections. SSE reconnects automatically within 10 seconds (SC-012); page doesn't break if logs fail to load temporarily.
- **Concurrent updates**: Multiple team members viewing the same pipeline run see synchronized updates through SSE in real-time.
- **Very large log outputs**: Dashboard streams logs and implements pagination/virtual scrolling to avoid rendering performance degradation.
- **Deleted projects or pipeline metadata**: Dashboard handles gracefully by showing archived/read-only view with message explaining that data is no longer maintained.
- **Permission mismatches**: Users without access to a project cannot view its runs or data; appropriate error messages guide them to request access.

### Filter Persistence and Sharing (Per A12 - Search & Filter)

**Filter State Management**:
1. **URL-based filter state** (RECOMMENDED for MVP):
   - Filters stored as query parameters: `/dashboard?status=failed&branch=main&search=abc123`
   - Users can bookmark filtered views
   - Shared links preserve filters
   - Back/forward browser buttons work with filters
   - Example: `/dashboard?status=running,failed&branch=*&created_after=2025-10-01&created_before=2025-10-31`

2. **Session-based filter persistence** (OPTIONAL enhancement):
   - If URL-based filters are cleared, check browser localStorage
   - Store last-used filter combination in localStorage
   - Restore on next dashboard visit
   - Max 10 saved filter presets per user (stored server-side)

**Filter Parameters Supported**:
- `status` - Comma-separated: `passed,failed,running,cancelled`
- `branch` - Git branch name (supports wildcards: `feature/*`)
- `search` - Commit SHA or author name (substring match)
- `created_after` - ISO date: `2025-10-01`
- `created_before` - ISO date: `2025-10-31`
- `project_id` - Filter to specific project

**Acceptance Criteria for FR-006 (Search & Filter)**:
1. Filters update URL query params immediately
2. Page reload preserves active filters
3. Shared links include filter state
4. Invalid filter values ignored (no error page)
5. Filters reset button clears all params
6. "Save filter preset" feature (optional for P2)

**Example Filter Scenarios**:
- Failed runs in production: `/dashboard?status=failed&branch=main`
- Recent commits by alice: `/dashboard?search=alice&created_after=2025-10-25`
- All running tests: `/dashboard?status=running&branch=*-test`

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a paginated list of pipeline runs for a selected project, sorted by creation time (newest first)
- **FR-002**: System MUST display real-time status updates for running pipelines without requiring page refresh
- **FR-003**: System MUST show detailed pipeline run information including: status, commit SHA, branch, author, start time, duration, and trigger source
- **FR-004**: System MUST provide a detailed view of individual pipeline runs showing all steps, their status, duration, and dependencies
- **FR-005**: System MUST stream real-time logs from executing pipeline steps with automatic scrolling to latest output
- **FR-006**: System MUST allow users to search pipeline runs by commit SHA and filter by branch, status, and date range
- **FR-007**: System MUST display project configuration interface for creating projects and obtaining webhook URLs
- **FR-008**: System MUST show artifacts generated by pipeline runs with download capability
- **FR-009**: System MUST render HTML and markdown artifact types (test reports, coverage reports, documentation) directly in the dashboard
- **FR-010**: System MUST authenticate users and enforce access control based on project membership
- **FR-011**: System SHOULD support dark mode as a post-launch enhancement; initial release focuses on light mode
- **FR-012**: System MUST be responsive and function on mobile devices (tablets, phones) for basic viewing scenarios
- **FR-013**: System MUST provide keyboard shortcuts for common actions (e.g., refresh, cancel pipeline, search)
- **FR-014**: System MUST maintain session security through appropriate token handling and HTTPS enforcement
- **FR-015**: System MUST cache frequently accessed data (pipeline lists, project metadata) to reduce API load and improve perceived performance
- **FR-016**: System MUST handle WebSocket or polling disconnections gracefully, with automatic reconnection and user notification

### Keyboard Shortcuts (Per FR-013)

Dashboard users can access the following keyboard shortcuts from any page:

| Shortcut | Action | Context | Notes |
|----------|--------|---------|-------|
| `?` | Show keyboard shortcuts help modal | Any page | Display modal with shortcut reference |
| `Ctrl/Cmd + K` | Focus search box | Pipeline list, Project list | Allows quick search without mouse |
| `Ctrl/Cmd + R` | Refresh current page | Any page | Force refresh of pipeline/project data |
| `Ctrl/Cmd + L` | Jump to latest log line | Pipeline detail (logs visible) | Auto-scroll to bottom of log viewer |
| `Esc` | Close modal/dropdown | Modal or filter menu open | Dismiss currently open UI element |
| `Ctrl/Cmd + Enter` | Submit form | Project creation or search filters | Submit form without clicking button |
| `J` | Jump to next run | Pipeline list (focus in table) | Navigate down one row in pipeline list |
| `K` | Jump to previous run | Pipeline list (focus in table) | Navigate up one row in pipeline list |
| `X` | Cancel selected pipeline | Pipeline detail (running pipeline) | Terminate currently executing pipeline |
| `D` | Download artifact | Artifact list (focus on artifact) | Download selected artifact |
| `V` | View artifact | Artifact list (focus on artifact) | Open/preview selected artifact |
| `Ctrl/Cmd + /` | Toggle search filter panel | Pipeline list | Show/hide advanced filter options |
| `Ctrl/Cmd + S` | Save filter preset | Pipeline list (filter active) | Save current filter combination as reusable preset |

**Acceptance Criteria for FR-013**:
1. All keyboard shortcuts defined above work on supported pages
2. Help modal (triggered by `?`) displays all available shortcuts for current page
3. Shortcuts respect modifier key conventions (Cmd on Mac, Ctrl on Windows/Linux)
4. Keyboard navigation doesn't interfere with text input fields (search, form fields)
5. Accessibility: All shortcut actions have equivalent mouse/UI alternatives (no functionality keyboard-only)

### Key Entities

- **Project**: Represents a connected Git repository with CI/CD configuration. Attributes: name, repository URL, webhook URL, creation date, last activity date, member list, permissions.
- **PipelineRun**: Represents a single execution of a pipeline triggered by a commit or manual trigger. Attributes: ID, project reference, status, commit SHA, branch, author, start time, end time, trigger source, overall duration.
- **PipelineStep**: Represents a single step within a pipeline run. Attributes: name, status, start time, end time, duration, image, commands, exit code, resources used.
- **Artifact**: Represents a file or output generated by a pipeline step. Attributes: name, type, size, MIME type, step reference, URL, created time.
- **User**: Represents a dashboard user. Attributes: ID, username, email, projects (array of project access), roles (admin, viewer, editor).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can view their most recent 10 pipeline runs within 2 seconds of accessing the dashboard
- **SC-002**: New pipeline run appears in the dashboard within 5 seconds of being triggered
- **SC-003**: Real-time log streaming shows output with less than 2 seconds of latency behind actual execution
- **SC-004**: Users can search/filter a list of 1000 pipeline runs and see results in under 1 second
- **SC-005**: Dashboard successfully handles 100+ concurrent users without degradation in response time
- **SC-006**: 95% of page loads complete within 3 seconds on standard internet connections
- **SC-007**: Dashboard is usable on mobile devices with screen sizes down to 375px wide (iPhone SE)
- **SC-008**: Users can access dashboard features with keyboard navigation (no mouse required)
- **SC-009**: Dashboard achieves 80+ Lighthouse accessibility score
- **SC-010**: Project setup (create project, obtain webhook, register webhook) can be completed in under 5 minutes
- **SC-011**: Artifact downloads complete in under 30 seconds for artifacts up to 500MB
- **SC-012**: System recovers from network interruptions and reconnects within 10 seconds

### Performance Test Environment Specification (Per A3)

All performance success criteria (SC-001, SC-002, SC-003, SC-004, SC-005, SC-006, SC-011) MUST be validated in a defined test environment to ensure reproducibility and consistency.

**Test Environment Configuration**:

| Component | Specification | Notes |
|-----------|---------------|-------|
| **Server Hardware** | 4 vCPU, 8GB RAM, SSD storage | Represents mid-tier production node |
| **Kubernetes Cluster** | Local k3d cluster with 1 control plane + 2 worker nodes | Simulates multi-node environment |
| **Database** | PostgreSQL 14+ (if used for metadata) | Native performance, no managed service overhead |
| **Object Storage** | Local MinIO or S3 compatible | Represents artifact storage layer |
| **Network Latency** | 10-50ms RTT (simulating WAN) | Can use `tc` (traffic control) to simulate latency |
| **Bandwidth** | 100 Mbps (simulating standard broadband) | Can use `tc` to throttle bandwidth |
| **Browser/Client** | Chrome/Firefox latest stable | Desktop viewport (1920x1080) |
| **Load Generator** | Apache JMeter or Locust for concurrent user tests | For SC-005 (100+ concurrent) |
| **Monitoring** | Prometheus + Grafana or built-in Go profiling | Track request latency, throughput, error rates |

**Performance Testing Procedure**:
1. Warm up database/cache with baseline data (e.g., 1000 pipeline runs for SC-004)
2. Clear caches before each test run
3. Measure 3 test runs, report average + p99 latency
4. Validate against success criteria
5. Document any deviations or environmental factors

**Client Network Conditions** (for SC-006):
- **"Standard internet connection"** defined as:
  - Downlink: 10-25 Mbps (typical cable/fiber)
  - Uplink: 5-10 Mbps
  - Latency: 20-40ms
  - Packet loss: < 0.1%
  - Can be simulated using Chrome DevTools throttling or network simulator

**Success Criteria Validation**:
- **SC-001** (2sec load): Measure time from initial dashboard page request to render complete with data
- **SC-002** (5sec new run): Create pipeline run via API, measure time to appearance in dashboard list (via SSE)
- **SC-003** (2sec log latency): Compare log timestamp with SSE event timestamp (measure backend queue delay)
- **SC-004** (1sec search): Insert 1000 test pipeline runs, measure search filter response time
- **SC-005** (100+ concurrent): Use load generator, maintain target response time under concurrent load
- **SC-006** (3sec page load): Simulate "standard internet" network conditions, measure full page load time
- **SC-011** (30sec download): Download 500MB artifact file, measure download completion time

---

## Assumptions

- **Authentication**: Assumes OAuth2 or session-based authentication is available through the existing C8S API server
- **Backend APIs**: Assumes REST API endpoints exist for fetching pipeline runs, steps, logs, and artifacts (or these will be implemented concurrently)
- **Real-time requirements**: Real-time updates will be achieved through either WebSocket connections or polling at reasonable intervals (5-10 second intervals acceptable for non-critical updates)
- **Artifact storage**: Artifacts are already stored and accessible through the S3-compatible storage system mentioned in the README
- **Browser compatibility**: Dashboard supports modern browsers (Chrome, Firefox, Safari, Edge) released in the last 2 years
- **Deployment**: Dashboard will be deployed alongside existing C8S components (controller, API server, webhook service)
- **Data retention**: Assumes the system has defined policies for log and artifact retention (data older than X days is archived/deleted)
