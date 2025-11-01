# Data Model: E2E Testing Framework Entities

**Date**: 2025-11-01
**Feature**: 005-create-a-robust
**Technology**: Playwright + TypeScript

## Overview

This document defines the key entities and data structures for the e2e testing framework. These are conceptual models representing test infrastructure components, not database schemas.

## Core Entities

### 1. TestSuite

**Purpose**: A collection of related e2e tests organized by feature or user workflow.

**Fields**:
- `id`: string - Unique identifier (e.g., "authentication", "pipeline-creation")
- `name`: string - Human-readable name (e.g., "Authentication Test Suite")
- `description`: string - Purpose and scope of the suite
- `tests`: TestCase[] - Array of test cases in this suite
- `setup`: () => Promise<void> - Shared setup function
- `teardown`: () => Promise<void> - Shared teardown function
- `tags`: string[] - Categorization tags (e.g., ["functional", "core", "daily"])
- `timeout`: number - Default timeout in milliseconds
- `retries`: number - Default retry count for flaky tests

**Relationships**:
- Contains many TestCase
- Generates many TestReport

**Validation Rules**:
- `id` must be unique within the test framework
- `name` must not be empty
- `timeout` must be >= 5000ms (5 seconds)
- `retries` must be 0-3 (prevent infinite retry loops)

---

### 2. TestCase

**Purpose**: An individual automated test that verifies a specific user scenario or functional behavior.

**Fields**:
- `id`: string - Unique identifier within suite (e.g., "should_display_login_form")
- `title`: string - Human-readable test name
- `description`: string - What scenario this test covers
- `category`: string - "functional" | "accessibility" | "performance"
- `priority`: "P1" | "P2" | "P3" - Importance level
- `timeout`: number - Test execution timeout in milliseconds
- `retries`: number - Number of retry attempts on failure
- `tags`: string[] - Labels (e.g., ["smoke", "critical", "mobile"])
- `browserScopes`: BrowserScope[] - Which browsers/viewports this test runs on
- `dependencies`: string[] - IDs of tests that must pass first
- `steps`: TestStep[] - Ordered sequence of test actions

**Relationships**:
- Belongs to one TestSuite
- Generates one TestExecution per run
- Can depend on other TestCase instances

**Validation Rules**:
- `id` must be unique within parent TestSuite
- `title` must be descriptive and start with action verb
- `timeout` must be 5000-60000ms (5-60 seconds)
- `browserScopes` must include at least one browser
- No circular dependencies allowed

---

### 3. TestStep

**Purpose**: An individual action or assertion within a test case.

**Fields**:
- `id`: string - Unique identifier within test (e.g., "click_login_button")
- `action`: "navigate" | "click" | "fill" | "expect" | "wait" | "screenshot" | "audit"
- `selector`: string - CSS selector or accessibility identifier
- `value`: string - Data to fill (for fill action) or expected value
- `description`: string - Human-readable description
- `waitCondition`: WaitCondition - Optional condition before proceeding
- `screenshot`: boolean - Whether to capture screenshot on failure
- `timeout`: number - Step-specific timeout override

**Relationships**:
- Part of ordered sequence in TestCase

**Validation Rules**:
- Action must be valid Playwright action
- `selector` must not be empty for interactions
- `timeout` if specified must be 1000-30000ms
- Order matters (steps execute sequentially)

---

### 4. BrowserScope

**Purpose**: Specification of browser, version, and viewport for test execution.

**Fields**:
- `browser`: "chromium" | "firefox" | "webkit" | "msedge"
- `viewport`: Viewport - Screen dimensions and device config
- `device`: string - Named device preset (e.g., "Desktop Chrome", "iPad Pro")

**Nested: Viewport**:
- `width`: number - Pixels (e.g., 1920)
- `height`: number - Pixels (e.g., 1080)
- `deviceScaleFactor`: number - For retina/high-DPI (default 1)

**Examples**:
```
Desktop Chrome 1920x1080
Tablet iPad Pro 1024x1366
Mobile iPhone 12 390x844
```

**Validation Rules**:
- `width` and `height` must be positive integers
- `deviceScaleFactor` must be >= 1

---

### 5. TestExecution

**Purpose**: Record of a single test run with all results and metadata.

**Fields**:
- `id`: string - Unique identifier (UUID or timestamp-based)
- `testId`: string - ID of TestCase that was executed
- `suiteId`: string - ID of TestSuite
- `status`: "passed" | "failed" | "skipped" | "timeout"
- `startTime`: ISO8601 - When test started
- `endTime`: ISO8601 - When test completed
- `duration`: number - Execution time in milliseconds
- `browserScope`: BrowserScope - Which browser/device was used
- `error`: TestError - Failure details if failed
- `evidence`: TestEvidence - Screenshots, videos, logs
- `metrics`: PerformanceMetrics - Performance data
- `retryAttempt`: number - Which retry attempt this was (0 = first try)
- `environment`: string - Test environment identifier

**Relationships**:
- Generated from one TestCase
- Contributes to one TestReport
- May reference TestError

**Validation Rules**:
- `endTime` must be >= `startTime`
- `duration` = `endTime` - `startTime`
- `status` reflects test outcome

---

### 6. TestError

**Purpose**: Detailed error information for failed tests.

**Fields**:
- `type`: string - Error classification (e.g., "timeout", "assertion", "element_not_found")
- `message`: string - Error message from Playwright or test
- `stack`: string - Stack trace
- `step`: string - Which step failed
- `selector`: string - CSS selector if element-related
- `expectedValue`: any - What was expected
- `actualValue`: any - What was actually observed
- `screenshot`: string - Path to failure screenshot
- `html`: string - Relevant HTML snippet at failure time

**Relationships**:
- Referenced by TestExecution

---

### 7. TestEvidence

**Purpose**: Multimedia evidence of test execution for debugging.

**Fields**:
- `screenshots`: Screenshot[] - Captured at key points
- `video`: string - Path to full test video recording
- `logs`: string[] - Browser console logs
- `networkTraffic`: NetworkRequest[] - API calls made during test
- `accessibility`: AccessibilityAudit - Axe-core audit results

**Nested: Screenshot**:
- `id`: string - Unique identifier
- `path`: string - File path to screenshot image
- `timestamp`: number - When captured
- `step`: string - Which test step this was from
- `description`: string - Why this screenshot was taken

---

### 8. TestReport

**Purpose**: Summary document capturing test execution results for a test run.

**Fields**:
- `id`: string - Unique identifier
- `suiteId`: string - Which TestSuite was run
- `startTime`: ISO8601 - When test run started
- `endTime`: ISO8601 - When test run completed
- `duration`: number - Total execution time
- `summary`: TestSummary - Aggregated results
- `executions`: TestExecution[] - Individual test results
- `environment`: TestEnvironment - Where tests ran
- `browsers`: BrowserScope[] - Which browser combinations tested
- `generatedAt`: ISO8601 - When report was generated

**Nested: TestSummary**:
- `total`: number - Total tests executed
- `passed`: number - Tests that passed
- `failed`: number - Tests that failed
- `skipped`: number - Tests that were skipped
- `timeout`: number - Tests that timed out
- `passRate`: number - Percentage 0-100
- `averageDuration`: number - Average test time in ms

**Validation Rules**:
- `endTime` >= `startTime`
- `passed` + `failed` + `skipped` + `timeout` = `total`
- `passRate` = (passed / total) * 100

---

### 9. PerformanceMetrics

**Purpose**: Performance measurements from test execution.

**Fields**:
- `navigationTiming`: NavigationTiming - Page load performance
- `firstContentfulPaint`: number - FCP in milliseconds
- `largestContentfulPaint`: number - LCP in milliseconds
- `cumulativeLayoutShift`: number - CLS score (0-1)
- `interactionLatency`: number - Time to respond to user action (ms)
- `apiResponseTimes`: ApiMetric[] - Per-endpoint response times
- `customMetrics`: Record<string, number> - App-specific measurements

**Validation Rules**:
- All timing metrics must be >= 0
- CLS must be between 0 and 1
- Custom metrics keys must be alphanumeric

---

### 10. AccessibilityAudit

**Purpose**: Results from automated accessibility scanning using axe-core.

**Fields**:
- `timestamp`: ISO8601 - When audit was run
- `violations`: AccessibilityViolation[]
- `passes`: AccessibilityPass[]
- `inapplicable`: AccessibilityRule[] - Rules that didn't apply
- `incomplete`: AccessibilityRule[] - Rules that need manual review
- `violationCount`: number - Total critical/serious violations
- `passCount`: number - Total passing checks
- `impact`: "critical" | "serious" | "moderate" | "minor"

**Nested: AccessibilityViolation**:
- `id`: string - Violation rule ID (e.g., "color-contrast")
- `impact`: "critical" | "serious" | "moderate" | "minor"
- `description`: string - What WCAG criterion is violated
- `help`: string - How to fix it
- `nodes`: AccessibilityNode[] - DOM elements affected
- `wcagLevel`: "A" | "AA" | "AAA" - WCAG compliance level

**Validation Rules**:
- `violationCount` must equal length of `violations`
- `passCount` must equal length of `passes`

---

### 11. TestEnvironment

**Purpose**: Configuration of where tests are executed.

**Fields**:
- `name`: string - Environment identifier (e.g., "staging", "production-like")
- `baseUrl`: string - Application URL
- `apiBaseUrl`: string - API endpoint URL
- `authToken`: string - Test authentication token (should be injected at runtime)
- `region`: string - Geographic region if applicable
- `timezone`: string - Test execution timezone
- `headless`: boolean - Whether browsers run headless
- `slowMo`: number - Slow down actions by N ms (for debugging)
- `trace`: "on" | "off" | "on-first-retry" - Playwright trace recording
- `video`: "on" | "off" | "on-failure" - Video recording mode

**Validation Rules**:
- `baseUrl` must be valid HTTP(S) URL
- `slowMo` must be >= 0
- Cannot be both headless and have video enabled (conflicting)

---

## Relationships Diagram

```
TestSuite
  ├── contains many → TestCase
  │                    ├── organized as → TestStep[]
  │                    └── generates → TestExecution
  │
  └── generates → TestReport
      ├── contains → TestExecution[]
      │   ├── may include → TestError
      │   └── includes → TestEvidence
      │       ├── screenshots → Screenshot[]
      │       ├── video → string
      │       └── accessibility → AccessibilityAudit
      │
      └── summarized by → TestSummary
```

---

## Validation Rules Summary

| Entity | Key Validation |
|--------|----------------|
| TestSuite | unique ID, non-empty name, timeout >= 5s, retries 0-3 |
| TestCase | unique ID within suite, descriptive title, proper timeout range, no circular dependencies |
| TestStep | valid action type, proper timeout if specified |
| BrowserScope | positive width/height, deviceScaleFactor >= 1 |
| TestExecution | endTime >= startTime, status reflects outcome |
| TestError | valid error type, meaningful message |
| TestReport | times valid, summary counts add up |
| PerformanceMetrics | all timings >= 0, CLS 0-1 range |
| AccessibilityAudit | violation/pass counts accurate |
| TestEnvironment | valid URLs, valid configuration combinations |

---

## State Transitions

### TestExecution Lifecycle

```
Created (pending)
    ↓
Running
    ├→ Passed
    ├→ Failed
    ├→ Timeout
    └→ Skipped
```

### TestCase Status

A test case is considered:
- **Ready**: Has defined steps, proper selectors, valid configuration
- **Flaky**: Passes < 99% of executions (should be flagged for investigation)
- **Deprecated**: Targets removed UI elements (should be updated or deleted)

---

## Notes

- All timestamps use ISO8601 format for consistency
- All IDs should be deterministic or UUID-based for reproducibility
- Performance metrics should align with Web Vitals standards
- Accessibility audit results follow axe-core format specification
