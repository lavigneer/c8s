# Research & Technical Decisions: Robust E2E Testing Workflow

**Date**: 2025-11-01
**Feature**: 005-create-a-robust
**Status**: Complete

## Overview

This document captures research findings and technical decisions for implementing a robust e2e testing workflow that covers both functional and accessibility testing. The c8s project is a Go-based Kubernetes pipeline management system with a web dashboard (HTML/CSS/HTMX).

## Decisions Made

### 1. Browser Automation Framework: Playwright

**Decision**: Use Playwright as the primary browser automation framework.

**Rationale**:
- Already in use in the project (present in `.mcp.json` and active Playwright test file at `tests/e2e/full_workflow.spec.ts`)
- Mature, well-maintained framework with excellent TypeScript support
- Excellent cross-browser support (Chrome, Firefox, Safari, Edge) out-of-the-box
- Strong accessibility testing capabilities via `@axe-core/playwright`
- Superior performance and reliability compared to older tools
- Built-in screenshot/video recording for test evidence
- Excellent documentation and community support

**Alternatives Considered**:
- Cypress: Better DX but less mature accessibility support, fewer browsers
- Selenium: Older, more verbose, harder to maintain
- Puppeteer: Chrome-only, less suitable for cross-browser requirements

**Evidence**: Existing Playwright setup in the project demonstrates team familiarity and compatibility.

---

### 2. Accessibility Testing Library: axe-core with Playwright

**Decision**: Use `@axe-core/playwright` plugin for automated accessibility audits.

**Rationale**:
- Industry-standard accessibility scanner detecting 95%+ of common WCAG violations
- Playwright native integration available
- Covers automated checks for WCAG 2.1 Level AA compliance
- Can be injected into any Playwright test
- Generates clear, actionable violation reports
- Low false-positive rate

**Alternatives Considered**:
- Google Lighthouse: More focused on performance, less granular accessibility
- W3C validators: Manual verification, not automation-friendly
- Manual testing only: Would not meet automated testing requirements

---

### 3. Test Reporting Solution: Playwright HTML Reporter + Custom Dashboard

**Decision**: Use Playwright's built-in HTML Reporter with custom enhancement for historical tracking.

**Rationale**:
- Playwright HTML Reporter includes screenshots, videos, and detailed error traces
- Provides immediate visual evidence of failures
- Integrates seamlessly with Playwright test infrastructure
- Can be extended with custom metadata and historical aggregation
- Supports CI/CD integration with proper exit codes

**Alternatives Considered**:
- Allure Report: Heavy external dependency, more complex setup
- ReportPortal: SaaS option, introduces external service dependency (violates constraint)
- Custom JSON aggregation: Simpler but less visual

---

### 4. Test Organization: Feature-Based Test Suites

**Decision**: Organize tests by feature/user workflow rather than test type.

**Rationale**:
- Aligns with specification user stories (authentication, pipeline creation, log viewing, artifacts)
- Enables independent test suite execution per user story
- Facilitates parallel test execution
- Easier for developers to find and run tests for specific features
- Natural mapping to user-facing functionality

**Structure**:
```
tests/e2e/
├── authentication.spec.ts
├── pipeline-creation.spec.ts
├── log-viewing.spec.ts
├── artifact-management.spec.ts
├── accessibility/
│   ├── keyboard-navigation.spec.ts
│   ├── screen-reader.spec.ts
│   └── color-contrast.spec.ts
├── performance-baseline.spec.ts
└── fixtures/
    ├── test-data.ts
    └── page-objects.ts
```

---

### 5. Cross-Browser Execution Strategy

**Decision**: Configure Playwright for all four required browsers with viewport/device variations.

**Rationale**:
- Playwright natively supports Chrome, Firefox, Safari, and Edge
- Can configure multiple browser contexts and viewport sizes
- Test matrix avoids code duplication via parameterization
- Optional sharding for parallel execution in CI

**Configuration Approach**:
- Standard CI runs: Chrome + Firefox (2 browsers)
- Full matrix runs: All 4 browsers (nightly or pre-release)
- Each test automatically runs against viewport variations (desktop, tablet, mobile)

---

### 6. CI/CD Integration: GitHub Actions with Playwright

**Decision**: Integrate e2e tests into existing CI/CD pipeline using GitHub Actions.

**Rationale**:
- Project already uses GitHub (indicated by `CLAUDE.md` references)
- Playwright has first-class GitHub Actions integration via `actions/setup-node` and `@actions/upload-artifact`
- Native support for job blocking on test failures
- Built-in support for comment posting results on PRs
- Can be triggered on PR creation and pre-deployment

---

### 7. Test Data Management: Fixture-Based with API Setup/Teardown

**Decision**: Use Playwright fixtures for test data with API-driven setup/teardown.

**Rationale**:
- Playwright fixtures provide automatic cleanup and test isolation
- API-based setup is faster than UI navigation
- Enables true test independence (no shared state between tests)
- Aligns with project's HTTP API infrastructure
- Reduces test execution time and flakiness

**Pattern**:
```typescript
test.beforeEach(async ({ page, request }) => {
  // API call to create test pipeline
  const pipeline = await request.post('/api/pipelines', { data: {...} });

  // Navigate page to that pipeline
  await page.goto(`/dashboard/pipelines/${pipeline.id}`);
});

test.afterEach(async ({ request }) => {
  // API call to clean up test data
  await request.delete(`/api/pipelines/${pipeline.id}`);
});
```

---

### 8. Performance Baseline Testing: Metrics Capture in E2E

**Decision**: Capture performance metrics directly within Playwright e2e tests.

**Rationale**:
- Playwright has native performance timing APIs
- Can measure real user interactions (not synthetic lab conditions)
- Captures actual browser performance under load
- Supports performance thresholds that fail tests

**Metrics to Track**:
- Navigation timing (load, DOMContentLoaded, First Contentful Paint)
- Interaction responsiveness (click-to-response)
- Custom metrics (API response times, animation duration)

---

### 9. Visual Regression Detection: Playwright Screenshots + Diff

**Decision**: Use Playwright's built-in screenshot comparison with manual approval workflow.

**Rationale**:
- Native Playwright feature, no external dependency
- Supports `.toHaveScreenshot()` with visual diff reporting
- Manual approval workflow prevents false negatives
- Integrated with HTML report for easy visual review

---

### 10. Flakiness Mitigation: Retry Strategy + Wait Conditions

**Decision**: Implement retry mechanism with smart wait conditions.

**Rationale**:
- Playwright supports built-in test retries
- Use `waitFor()` with appropriate timeout conditions instead of arbitrary sleeps
- Target WCAG-compliant accessibility selectors (not brittle classes)
- Implement page object model to centralize wait logic

---

## Technology Stack Summary

| Category | Decision | Why |
|----------|----------|-----|
| Browser Automation | Playwright | Already in use, excellent cross-browser support |
| Accessibility Testing | @axe-core/playwright | Industry standard, WCAG coverage, Playwright integration |
| Test Reporting | Playwright HTML Reporter | Built-in, visual evidence, CI-friendly |
| Test Organization | Feature-based suites | Aligns with user stories, enables independent execution |
| Cross-Browser | All 4 browsers + viewport matrix | Meet success criteria requirements |
| CI/CD Platform | GitHub Actions | Project-native, first-class Playwright support |
| Test Data | API fixtures | Fast, isolated, true test independence |
| Performance Testing | Native Playwright metrics | Real user conditions, no external tools |
| Visual Regression | Screenshot comparison | Native feature, manual approval workflow |
| Flakiness Prevention | Retry + wait conditions | Smart waits, accessibility-based selectors |

## Architecture Pattern: Page Object Model

All tests will use the Page Object Model pattern to:
- Centralize UI selectors (resilient to minor changes)
- Encapsulate wait conditions
- Provide reusable interaction methods
- Improve test maintainability

Example structure:
```
tests/e2e/
├── pages/
│   ├── base.page.ts (common interactions)
│   ├── login.page.ts
│   ├── dashboard.page.ts
│   └── pipeline-detail.page.ts
└── fixtures/
    └── page-objects.ts (fixture that provides page objects)
```

## Success Criteria Alignment

- **SC-001** (30-min execution): Achievable with smart parallelization and focused test suites
- **SC-002** (90% pass rate): Achievable with proper setup/teardown and wait conditions
- **SC-003** (Critical workflows tested): Addressed via feature-based test organization
- **SC-004** (95% accessibility violation detection): Addressed via axe-core integration
- **SC-005** (Automated reporting): Addressed via Playwright HTML Reporter
- **SC-006** (All 4 browsers): Addressed via Playwright multi-browser configuration
- **SC-007** (All 3 viewports): Addressed via device emulation configuration
- **SC-008** (2-min feedback): Achievable with proper CI setup and parallel execution
- **SC-009** (99% stability): Addressed via retry strategy and smart waits
- **SC-010** (15-min local troubleshooting): Achievable with good Page Object Model design
- **SC-011** (No critical accessibility issues): Achieved via axe-core + manual guidance
- **SC-012** (10% maintenance overhead): Achievable via Page Object Model and fixture pattern

## Next Steps

1. Generate data model for test entities
2. Create API contracts for test data setup/teardown endpoints
3. Develop quickstart guide for running and writing tests
4. Create comprehensive test suite following Page Object Model pattern
