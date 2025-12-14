# c8s Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-12

## Active Technologies
- (001-build-a-continuous)
- Go 1.25.0 (002-i-want-to)
- N/A (cluster state managed by chosen local K8s distribution) (002-i-want-to)
- Go 1.24.0 (backend API server), HTML5/CSS3/JavaScript (frontend with HTMX) (004-create-a-front)
- Uses existing C8S infrastructure (Kubernetes, S3-compatible object storage for logs/artifacts) (004-create-a-front)
- Go 1.25.0 (CLI tool), existing C8S backend/frontend (no changes required) + kubectl (Kubernetes client), existing C8S components (API server, controller, frontend) (008-create-a-simple)
- Kubernetes persistent volumes, S3-compatible object storage (already integrated in C8S) (008-create-a-simple)
- Helm 3.x (chart templating), Kubernetes 1.24+, YAML manifests + Helm 3.x CLI, kubectl CLI, existing C8S components (API server, controller, webhook, frontend) (008-create-a-simple)
- Kubernetes persistent volumes, S3-compatible object storage (MinIO or AWS S3) (008-create-a-simple)

## Project Structure
```
cmd/          # Main applications (controller, api-server, webhook, CLI)
pkg/          # Shared libraries and APIs
tests/        # Unit, integration, and E2E tests (Playwright)
chart/c8s/    # Helm chart (values.yaml and values-dev.yaml actively used)
deploy/       # Legacy manifests (use Helm + Tilt for deployment)
config/       # CRD manifests and RBAC
```

**Development Workflow**:
- **Primary**: Tilt (`tilt up`) for local Kubernetes development with live reload
- **Deployment**: Helm chart (chart/c8s/) with values-dev.yaml for local development
- **Testing**: Playwright E2E tests (shell-based *.sh tests are deprecated)

## E2E Testing Framework (005-create-a-robust)

### Quick Start
```bash
npm install                           # Install dependencies
npx playwright install                # Install browser binaries
npm run test:e2e                      # Run all e2e tests
npm run test:e2e:ui                   # Interactive test UI
npm run test:e2e:debug                # Run with Playwright debugger
npm run test:e2e:report               # View HTML test report
```

### Test Structure
```
tests/e2e/
├── specs/                            # Test suites by feature
│   ├── authentication.spec.ts        # Login, session, logout (8 tests)
│   ├── pipeline-creation.spec.ts     # Pipeline CRUD (8 tests)
│   ├── log-viewing.spec.ts          # Log streaming & filtering (6 tests)
│   ├── artifact-management.spec.ts   # Artifact operations (7 tests)
│   ├── cross-browser.spec.ts        # Multi-browser compatibility (14 tests)
│   ├── responsive.spec.ts           # Responsive design (11 tests)
│   ├── performance.spec.ts          # Performance baselines (11 tests)
│   └── accessibility/
│       ├── keyboard-navigation.spec.ts    # WCAG keyboard tests (8 tests)
│       ├── screen-reader.spec.ts         # WCAG screen reader (10 tests)
│       ├── color-contrast.spec.ts        # WCAG contrast (10 tests)
│       └── focus-management.spec.ts      # WCAG focus (10 tests)
├── pages/                            # Page Object Models
│   ├── base.page.ts                 # Base class with accessibility helpers
│   ├── login.page.ts                # Login page interactions
│   ├── dashboard.page.ts            # Dashboard navigation
│   ├── pipeline-detail.page.ts      # Pipeline management
│   ├── log-viewer.page.ts           # Log viewing
│   └── artifact-manager.page.ts     # Artifact operations
└── fixtures/                         # Test utilities
    ├── test-data.ts                 # API request helpers
    ├── auth.ts                      # Authentication setup
    ├── page-objects.ts              # Page object fixtures
    ├── reporting.ts                 # Metrics & reporting
    ├── metrics.ts                   # Performance metrics
    └── constants.ts                 # Test configuration

```

### Test Coverage
- **Total Test Cases**: 120+ automated tests
- **Functional**: 29 tests (authentication, pipelines, logs, artifacts)
- **Accessibility**: 38 tests (keyboard, screen reader, contrast, focus)
- **Cross-Browser**: 14 tests (Chrome, Firefox, Safari, Edge)
- **Responsive**: 11 tests (desktop, tablet, mobile)
- **Performance**: 11 tests (load time, memory, network)
- **Browsers**: 4 major browsers (Chromium, Firefox, WebKit, MSEdge)
- **Viewports**: 3 sizes (desktop 1920x1080, tablet 1024x1366, mobile 390x844)

### Key Features
- ✅ TDD approach - tests written first
- ✅ Page Object Model for maintainability
- ✅ WCAG 2.1 Level AA accessibility testing
- ✅ axe-core integration for automated audits
- ✅ Cross-browser and responsive design testing
- ✅ Performance metrics capture
- ✅ Automatic reporting and metrics aggregation
- ✅ GitHub Actions CI/CD integration
- ✅ Test isolation via API-based test data
- ✅ Graceful error handling

### CI/CD Integration
- Automatically runs on PR creation and push to main
- Matrix strategy tests across 2 browsers × 3 viewports
- Artifacts: HTML reports, videos on failure, metrics JSON
- PR commenting with test summary
- Automatic deployment gate on failures

### Commands

## HTMX Development Philosophy

**Core Principle**: Use HTMX functionality over custom JavaScript logic wherever possible.

### HTMX Best Practices
1. **Prefer HTMX Attributes Over JavaScript**
   - Use `hx-post`, `hx-get`, etc. instead of fetch/XMLHttpRequest
   - Use `hx-validate="true"` for form validation
   - Use `hx-on` for event handlers instead of `.addEventListener()`
   - Use response-targets extension for response code routing

2. **Form Handling with HTMX**
   - ✅ `hx-post="/endpoint"` for form submission
   - ✅ `hx-validate="true"` for HTML5 validation
   - ✅ `htmx:validation:failed` event for custom validation UI
   - ✅ `hx-target-400/422/5*` for error routing
   - ❌ Don't use form submit listeners
   - ❌ Don't use preventDefault() with forms

3. **HTMX Extensions to Leverage**
   - response-targets - Route responses by HTTP status code
   - loading-states - Manage loading/disabled states automatically
   - (others as needed for features)

4. **Test Independence**
   - E2E tests should NOT depend on HTMX internals
   - Don't wait for HTMX events or check window.htmx
   - Test user-visible behavior, not implementation details
   - Tests should work the same if HTMX is removed

5. **JavaScript Usage**
   - Use JavaScript for UI state and accessibility
   - Use JavaScript for event listeners on HTMX events (validation, etc.)
   - Keep custom logic minimal - let HTMX handle the heavy lifting
   - Always prefer HTMX attributes before writing JavaScript

### Example Pattern: Form Validation
```html
<!-- HTML: Let HTMX handle submission -->
<form hx-post="/login" hx-validate="true">
  <input type="text" name="username" required>
  <input type="password" name="password" required>
</form>

<!-- JavaScript: Handle UI updates only -->
document.addEventListener('htmx:validation:failed', (evt) => {
  const field = evt.detail.elt;
  field.setAttribute('aria-invalid', 'true');
  // Show error, set focus, etc.
});
```

## Code Style
Follow standard conventions

## Git Workflow
- **Commit after each significant feature** or bug fix
- **Commit format:** `[Txxx] Feature description` (e.g., `[T084] Implement Dashboard Pipeline Runs`)
- **Always include footer:**
  ```
  🤖 Generated with Claude Code

  Co-Authored-By: Claude <noreply@anthropic.com>
  ```
- **Use git status and git diff --cached --stat before committing**
- Commit logical units of work, not partial implementations

## Development Checklist
Before starting work on a feature:
1. Check current `git status`
2. Create a todo list for the task
3. Work on the feature in small, logical steps
4. Test changes thoroughly
5. Commit with meaningful message when feature is complete

After finishing work:
1. Verify `git log` shows commits for completed work
2. Review changes with `git diff HEAD~N` (where N is number of commits)

## Recent Changes
- 008-create-a-simple: Added Helm 3.x (chart templating), Kubernetes 1.24+, YAML manifests + Helm 3.x CLI, kubectl CLI, existing C8S components (API server, controller, webhook, frontend)
- 008-create-a-simple: Added Go 1.25.0 (CLI tool), existing C8S backend/frontend (no changes required) + kubectl (Kubernetes client), existing C8S components (API server, controller, frontend)
- 005-create-a-robust: **COMPLETED** - Comprehensive E2E testing framework (all 7 phases)
  - Phase 1: Test infrastructure setup (Playwright, axe-core, GitHub Actions)
  - Phase 2: Foundational Page Objects and test fixtures
  - Phase 3: Functional E2E tests (29 test cases)
  - Phase 4: Accessibility E2E tests (38 test cases)
  - Phase 5: Test reporting and performance metrics
  - Phase 6: Cross-browser and responsive design tests (25 test cases)
  - Phase 7: CI/CD integration with deployment gates
  - Pipeline history and status visualization (US1)
  - Real-time log streaming with SSE (US2)
  - Advanced filtering with URL state (US3)
  - Project and webhook management UI (US4)
  - Artifact viewing and download (US5)
  - Keyboard shortcuts (FR-013)
  - Authentication and authorization (FR-010)
  - Mobile-responsive design (FR-012)

## Dashboard Implementation Details
**Status**: COMPLETE, TESTED, AND ENHANCED (Phase 2)

### Phase 1: Initial Implementation (Complete)
- All 5 user stories fully implemented
- 13 feature requirements fulfilled
- 10 commits totaling 600+ lines of code
- Real-time log streaming via SSE
- Advanced filtering with URL persistence
- Keyboard shortcut support
- Artifact management with preview capability
- Responsive design for mobile devices

### Phase 2: Quality-of-Life Enhancements (Complete)
- 3 additional commits with reliability improvements
- Advanced error handling with automatic retry (T091)
  * Retry on 408, 429, and 5xx errors
  * 1-second retry delay
  * Automatic branch fetching from Kubernetes
  * Filter loading feedback with spinner
- Live updates with visual feedback (T092)
  * SSE updates preserve filter state
  * Update notification toast
  * Loading skeleton states for better UX
  * Smooth 200ms transitions
- Dashboard quick stats panel (T093)
  * Dynamic metrics: Total, Success Rate, Failed, Running
  * Color-coded cards
  * No additional API overhead
  * Responsive grid layout

### Key Files Modified
- `cmd/api-server/handlers/` - 7 new/updated handler files
- `cmd/api-server/templates/` - 10+ template files
- `cmd/api-server/static/js/` - Keyboard shortcuts implementation
- `pkg/dashboard/` - DTOs and logging infrastructure
- `DASHBOARD_IMPLEMENTATION.md` - Complete reference documentation

### Testing Checklist
✅ Authentication flow (login/logout)
✅ Pipeline listing and filtering
✅ Log streaming and viewer
✅ Project creation and management
✅ Artifact download and preview
✅ Keyboard shortcuts functionality
✅ URL state persistence
✅ Responsive design
✅ Error handling and automatic retry
✅ Loading skeleton states
✅ Live SSE updates
✅ Filter loading feedback
✅ Dashboard quick stats
✅ Branch dynamic fetching

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
