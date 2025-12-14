# E2E Testing with Playwright

End-to-end testing suite for C8S dashboard and API using Playwright.

## Overview

This directory contains 120+ comprehensive E2E tests covering:
- **Functional testing** (authentication, pipelines, logs, artifacts)
- **Accessibility testing** (WCAG 2.1 Level AA compliance)
- **Cross-browser testing** (Chromium, Firefox, WebKit, Edge)
- **Responsive design testing** (desktop, tablet, mobile)
- **Performance testing** (load times, memory, network)

## Quick Start

```bash
# Install dependencies
npm install

# Install Playwright browsers (one-time)
npx playwright install

# Run all tests
npm run test:e2e

# Run in UI mode (recommended for development)
npm run test:e2e:ui

# Run with debugger
npm run test:e2e:debug
```

## Test Coverage

| Category | Tests | Description |
|----------|-------|-------------|
| **Functional** | 29 | Core features (auth, pipelines, logs, artifacts) |
| **Accessibility** | 38 | WCAG 2.1 AA compliance (keyboard, screen reader, contrast, focus) |
| **Cross-Browser** | 14 | Chrome, Firefox, Safari, Edge compatibility |
| **Responsive** | 11 | Desktop, tablet, mobile layouts |
| **Performance** | 11 | Page load, memory, network metrics |
| **Total** | **120+** | Comprehensive coverage |

## Directory Structure

```
tests/e2e/
├── README.md                          # This file
├── specs/                             # Test suites
│   ├── authentication.spec.ts         # Login, logout, session (8 tests)
│   ├── pipeline-creation.spec.ts      # Create, edit, delete pipelines (8 tests)
│   ├── log-viewing.spec.ts           # Log streaming, filtering (6 tests)
│   ├── artifact-management.spec.ts    # Upload, download artifacts (7 tests)
│   ├── cross-browser.spec.ts         # Multi-browser tests (14 tests)
│   ├── responsive.spec.ts            # Responsive design (11 tests)
│   ├── performance.spec.ts           # Performance metrics (11 tests)
│   └── accessibility/                 # WCAG compliance tests
│       ├── keyboard-navigation.spec.ts     # Keyboard accessibility (8 tests)
│       ├── screen-reader.spec.ts          # Screen reader support (10 tests)
│       ├── color-contrast.spec.ts         # Color contrast ratios (10 tests)
│       └── focus-management.spec.ts       # Focus indicators (10 tests)
├── pages/                             # Page Object Models
│   ├── base.page.ts                  # Base page class
│   ├── login.page.ts                 # Login page interactions
│   ├── dashboard.page.ts             # Dashboard navigation
│   ├── pipeline-detail.page.ts       # Pipeline management
│   ├── log-viewer.page.ts            # Log viewing
│   └── artifact-manager.page.ts      # Artifact operations
├── fixtures/                          # Test utilities
│   ├── test-data.ts                  # API test data helpers
│   ├── auth.ts                       # Authentication fixtures
│   ├── page-objects.ts               # Page object fixtures
│   ├── reporting.ts                  # Metrics & reporting
│   ├── metrics.ts                    # Performance metrics
│   └── constants.ts                  # Test configuration
└── utils/                            # Helper functions
    └── accessibility.ts              # axe-core helpers
```

## Page Object Model

We use the Page Object Model pattern for maintainability:

**Example: LoginPage**
```typescript
// pages/login.page.ts
export class LoginPage extends BasePage {
  async login(username: string, password: string) {
    await this.page.fill('[name="username"]', username);
    await this.page.fill('[name="password"]', password);
    await this.page.click('button[type="submit"]');
    await this.page.waitForURL('/dashboard');
  }

  async getErrorMessage() {
    return await this.page.locator('.error-message').textContent();
  }
}
```

**Usage in tests**:
```typescript
// specs/authentication.spec.ts
test('should login successfully', async ({ loginPage }) => {
  await loginPage.login('admin', 'password');
  await expect(loginPage.page).toHaveURL(/.*dashboard/);
});
```

## Running Tests

### All Tests

```bash
npm run test:e2e
```

### Specific Test File

```bash
npx playwright test tests/e2e/specs/authentication.spec.ts
```

### Specific Test

```bash
npx playwright test tests/e2e/specs/authentication.spec.ts -g "should login successfully"
```

### By Browser

```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
```

### By Tag

```bash
npx playwright test --grep @smoke     # Smoke tests only
npx playwright test --grep @critical  # Critical path tests
```

### Interactive Mode

```bash
npm run test:e2e:ui
```

Opens Playwright UI where you can:
- Run tests interactively
- See live browser preview
- Inspect DOM and network
- Time-travel through test steps

### Debug Mode

```bash
npm run test:e2e:debug
```

Runs with Playwright Inspector:
- Step through test execution
- Pause and resume
- Inspect page state
- Generate test code

## Test Fixtures

Fixtures provide reusable setup:

### Authentication Fixture

```typescript
// fixtures/auth.ts
export const authFixture = test.extend({
  authenticatedPage: async ({ page }, use) => {
    await page.goto('/login');
    await page.fill('[name="username"]', 'testuser');
    await page.fill('[name="password"]', 'testpass');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard');
    await use(page);
  },
});
```

### Page Object Fixture

```typescript
// fixtures/page-objects.ts
export const pageObjectFixture = test.extend({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  dashboardPage: async ({ page }, use) => {
    await use(new DashboardPage(page));
  },
});
```

## Accessibility Testing

We use axe-core for automated accessibility testing:

```typescript
import { injectAxe, checkA11y } from 'axe-playwright';

test('dashboard should be accessible', async ({ page }) => {
  await page.goto('/dashboard');
  await injectAxe(page);
  await checkA11y(page, null, {
    detailedReport: true,
    detailedReportOptions: { html: true },
  });
});
```

**WCAG 2.1 Level AA Compliance Checks**:
- Keyboard navigation
- Screen reader support
- Color contrast ratios (4.5:1 for normal text)
- Focus indicators
- Semantic HTML
- ARIA labels

## Performance Testing

Performance metrics are captured for key pages:

```typescript
test('dashboard should load quickly', async ({ page }) => {
  const startTime = Date.now();
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');
  const loadTime = Date.now() - startTime;

  expect(loadTime).toBeLessThan(3000); // 3 second threshold
});
```

**Metrics Tracked**:
- Page load time
- Time to Interactive (TTI)
- First Contentful Paint (FCP)
- Largest Contentful Paint (LCP)
- Memory usage
- Network request count

## CI/CD Integration

Tests run automatically in GitHub Actions:

**.github/workflows/e2e-tests.yml**:
- Triggers on frontend code changes
- Runs across 2 browsers × 3 viewports = 6 configurations
- Uploads test reports and videos on failure
- Comments on PR with test results

## Test Reports

After running tests:

```bash
# View HTML report
npm run test:e2e:report

# Report location
# playwright-report/index.html
```

Report includes:
- Test results summary
- Failed test details
- Screenshots on failure
- Test execution timeline
- Browser console logs

## Writing New Tests

### 1. Choose Test Type

- **Functional**: User workflows and features
- **Accessibility**: WCAG compliance
- **Performance**: Load times and metrics
- **Responsive**: Layout on different devices

### 2. Use Page Objects

Don't query DOM directly in tests. Use page objects:

```typescript
// ❌ Bad: Direct DOM queries
await page.click('.submit-button');

// ✅ Good: Page object
await loginPage.clickSubmit();
```

### 3. Wait for Elements

Playwright auto-waits, but be explicit when needed:

```typescript
// Wait for navigation
await page.waitForURL('/dashboard');

// Wait for element
await page.waitForSelector('.pipeline-list');

// Wait for API response
await page.waitForResponse(response =>
  response.url().includes('/api/v1/pipelines') && response.status() === 200
);
```

### 4. Use Assertions

```typescript
// Visibility
await expect(page.locator('.success-message')).toBeVisible();

// Text content
await expect(page.locator('.title')).toHaveText('Dashboard');

// URL
await expect(page).toHaveURL(/.*dashboard/);

// Count
await expect(page.locator('.pipeline-item')).toHaveCount(5);
```

### 5. Test Independence

Each test should be independent:

```typescript
test.beforeEach(async ({ page }) => {
  // Reset state before each test
  await page.goto('/');
  // Create fresh test data via API
});

test.afterEach(async ({ page }) => {
  // Clean up test data
});
```

## Best Practices

### ✅ DO

- Use Page Object Model
- Test user workflows, not implementation
- Use semantic selectors (role, label, text)
- Wait for elements properly
- Make tests independent
- Use fixtures for common setup
- Add accessibility checks
- Capture performance metrics

### ❌ DON'T

- Query DOM directly in tests
- Use CSS selectors for dynamic elements
- Hardcode wait times (use waitFor*)
- Make tests depend on each other
- Test implementation details
- Skip accessibility tests
- Ignore performance thresholds

## Debugging Tests

### View Test in Browser

```bash
npx playwright test --headed
```

### Slow Down Execution

```bash
npx playwright test --slow-mo=1000  # 1 second delay
```

### Pause Test

```typescript
await page.pause();  // Opens Playwright Inspector
```

### Screenshots on Failure

Automatically captured in `test-results/`

### Trace Viewer

```bash
npx playwright show-trace trace.zip
```

Shows:
- Test steps
- Screenshots
- Network activity
- Console logs
- Source code

## Troubleshooting

### "Playwright browsers not installed"

```bash
npx playwright install
```

### "Tests timing out"

Increase timeout in playwright.config.ts:
```typescript
timeout: 60000, // 60 seconds
```

### "Flaky tests"

1. Add proper waits (`waitForLoadState`, `waitForSelector`)
2. Increase retry count in config
3. Check for race conditions
4. Use `page.waitForResponse()` for API calls

### "Can't find element"

1. Check selector is correct
2. Wait for element to appear
3. Check element isn't in shadow DOM
4. Use Playwright Inspector to debug

## Resources

- [Playwright Documentation](https://playwright.dev/)
- [Page Object Model](https://playwright.dev/docs/pom)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Accessibility Testing](https://playwright.dev/docs/accessibility-testing)
- [axe-core Rules](https://github.com/dequelabs/axe-core/blob/develop/doc/rule-descriptions.md)

## Contributing

When adding new E2E tests:

1. Follow Page Object Model pattern
2. Add to appropriate spec file or create new one
3. Update this README if adding new patterns
4. Ensure tests pass locally before PR
5. Add accessibility checks for new pages
6. Document any new fixtures or utilities

---

**Questions?** See main [tests/README.md](../README.md) or [CLAUDE.md](../../CLAUDE.md#e2e-testing-framework)
