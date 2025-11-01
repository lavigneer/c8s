# E2E Testing Framework - Quick Start Guide

**Date**: 2025-11-01
**Framework**: Playwright + TypeScript
**Status**: Implementation Guide

## Overview

This guide helps you quickly get started with the C8S e2e testing framework for testing both functionality and accessibility.

## Prerequisites

- Node.js 16+ (check with `node --version`)
- npm or yarn package manager
- C8S application running locally or in staging environment
- Playwright browsers installed

## Installation

### 1. Install Playwright

```bash
npm install -D @playwright/test @axe-core/playwright
```

### 2. Install Accessibility Testing Tools

```bash
npm install -D axe-playwright @axe-core/playwright
```

### 3. Install Playwright Browsers

```bash
npx playwright install
```

## Project Structure

```
tests/e2e/
├── pages/                          # Page Object Models
│   ├── base.page.ts               # Shared page methods
│   ├── login.page.ts
│   ├── dashboard.page.ts
│   ├── pipeline-detail.page.ts
│   ├── log-viewer.page.ts
│   └── artifact-manager.page.ts
│
├── fixtures/                        # Test Fixtures
│   ├── test-data.ts               # Test data generators
│   ├── page-objects.ts            # Fixture providers
│   └── auth.ts                    # Authentication setup
│
├── specs/                          # Test Suites
│   ├── authentication.spec.ts
│   ├── pipeline-creation.spec.ts
│   ├── log-viewing.spec.ts
│   ├── artifact-management.spec.ts
│   └── accessibility/
│       ├── keyboard-navigation.spec.ts
│       ├── screen-reader.spec.ts
│       └── color-contrast.spec.ts
│
└── playwright.config.ts            # Playwright configuration
```

## Running Tests

### Run All Tests

```bash
npx playwright test
```

### Run Single Test Suite

```bash
npx playwright test tests/e2e/specs/authentication.spec.ts
```

### Run Tests with Specific Browser

```bash
npx playwright test --project=chromium
```

### Run Tests in Headed Mode (See Browser)

```bash
npx playwright test --headed
```

### Run Tests with Debug Mode (Interactive Debugger)

```bash
npx playwright test --debug
```

### Run Tests in UI Mode (Test Dashboard)

```bash
npx playwright test --ui
```

## Writing a New Test

### Basic Structure

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/login.page';
import { DashboardPage } from '../pages/dashboard.page';

test.describe('Pipeline Creation', () => {
  test('should create a new pipeline', async ({ page, context }) => {
    // Arrange: Setup test data
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);

    // Act: Execute user actions
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.clickCreatePipeline();

    // Assert: Verify outcomes
    await expect(page).toHaveURL(/\/pipelines\/\d+/);
  });
});
```

### Page Object Model Example

```typescript
// pages/login.page.ts
import { Page, expect } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
    await expect(this.page).toHaveTitle(/Login/);
  }

  async login(email: string, password: string) {
    await this.page.fill('input[type="email"]', email);
    await this.page.fill('input[type="password"]', password);
    await this.page.click('button[type="submit"]');
    await this.page.waitForURL('/dashboard');
  }

  async fillEmail(email: string) {
    await this.page.fill('input[type="email"]', email);
  }

  async fillPassword(password: string) {
    await this.page.fill('input[type="password"]', password);
  }

  async clickSubmit() {
    await this.page.click('button[type="submit"]');
  }
}
```

## Testing Accessibility

### Basic Accessibility Audit

```typescript
import { injectAxe, checkA11y } from 'axe-playwright';

test('should have no accessibility violations', async ({ page }) => {
  await page.goto('/dashboard');

  // Inject axe script
  await injectAxe(page);

  // Run accessibility check
  await checkA11y(page);
});
```

### Keyboard Navigation Test

```typescript
test('should be keyboard navigable', async ({ page }) => {
  await page.goto('/dashboard');

  // Tab through elements
  for (let i = 0; i < 5; i++) {
    await page.keyboard.press('Tab');
  }

  // Focus should have moved through interactive elements
  const focused = await page.evaluate(() => document.activeElement?.tagName);
  expect(['A', 'BUTTON', 'INPUT']).toContain(focused);
});
```

### Screen Reader Test

```typescript
test('form labels should be announced', async ({ page }) => {
  await page.goto('/dashboard/create-pipeline');

  // Get accessibility tree
  const accessibility = await page.evaluate(() => {
    return (document.querySelector('label')?.textContent);
  });

  expect(accessibility).toContain('Pipeline Name');
});
```

### Color Contrast Test

```typescript
test('should have sufficient color contrast', async ({ page }) => {
  await page.goto('/dashboard');
  await injectAxe(page);

  // Check for color contrast violations
  const results = await checkA11y(page, undefined, {
    rules: {
      'color-contrast': { enabled: true }
    }
  });

  expect(results).toEqual([]);
});
```

## Test Data Setup/Teardown

### Using Fixtures for Setup

```typescript
// fixtures/test-data.ts
import { test as base, APIRequestContext } from '@playwright/test';

type TestDataFixture = {
  api: APIRequestContext;
  testPipeline: any;
  testProject: any;
};

export const test = base.extend<TestDataFixture>({
  testPipeline: async ({ api }, use) => {
    // Create pipeline
    const response = await api.post('/api/test/pipelines', {
      data: {
        name: `test-pipeline-${Date.now()}`,
        repository: 'github.com/test/repo'
      }
    });
    const pipeline = await response.json();

    // Use in test
    await use(pipeline);

    // Cleanup after test
    await api.delete(`/api/test/pipelines/${pipeline.id}`);
  }
});
```

### Using the Fixture

```typescript
import { test } from '../fixtures/test-data';

test('should list pipelines', async ({ page, testPipeline }) => {
  // testPipeline is automatically created and will be deleted after test
  await page.goto('/dashboard');
  await expect(page.locator(`text=${testPipeline.name}`)).toBeVisible();
});
```

## Performance Testing

### Measure Page Load

```typescript
test('should load dashboard in under 2 seconds', async ({ page }) => {
  const startTime = Date.now();

  await page.goto('/dashboard');

  const loadTime = Date.now() - startTime;
  expect(loadTime).toBeLessThan(2000);
});
```

### Measure Web Vitals

```typescript
test('should have good web vitals', async ({ page }) => {
  await page.goto('/dashboard');

  const vitals = await page.evaluate(() => ({
    FCP: performance.getEntriesByName('first-contentful-paint')[0]?.startTime,
    LCP: performance.getEntriesByType('largest-contentful-paint').pop()?.startTime,
  }));

  expect(vitals.FCP).toBeLessThan(1500);
  expect(vitals.LCP).toBeLessThan(2500);
});
```

## Visual Regression Testing

### Capture Baseline

```typescript
test('should match dashboard snapshot', async ({ page }) => {
  await page.goto('/dashboard');

  // On first run, generates baseline screenshot
  await expect(page).toHaveScreenshot('dashboard.png');
});
```

### Compare Against Baseline

```bash
npx playwright test --update-snapshots  # Generate baselines
npx playwright test                      # Compare against baselines
```

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on:
  pull_request:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '18'

      - run: npm install
      - run: npx playwright install --with-deps

      - run: npm run test:e2e

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: playwright-report/
```

## Troubleshooting

### Test is Flaky (Intermittently Fails)

1. **Add explicit waits**: Use `waitForURL()`, `waitForSelector()` instead of sleeps
2. **Use accessibility selectors**: Target by role, label instead of brittle CSS
3. **Enable retries**: Add `test.only()` or configure retries in `playwright.config.ts`

```typescript
// Use waitFor instead of sleep
await page.waitForURL('/dashboard');  // Good
await page.waitForTimeout(2000);      // Bad
```

### Element Not Found

1. Check if element is visible: `await expect(element).toBeVisible()`
2. Use Playwright Inspector: `npx playwright test --debug`
3. Take screenshot at failure: `await page.screenshot({ path: 'debug.png' })`

### Timeout Errors

1. Increase timeout in `playwright.config.ts`:
   ```typescript
   timeout: 60 * 1000  // 60 seconds
   ```
2. Or per test:
   ```typescript
   test('slow test', async ({ page }) => {
     // test code
   }, { timeout: 120_000 });
   ```

### Accessibility Audit Failing

1. Run specific rule: `await checkA11y(page, undefined, { runOnly: { type: 'rule', values: ['color-contrast'] } })`
2. View detailed violations: Enable verbose logging
3. Manually inspect in axe DevTools browser extension

## Best Practices

1. **One assertion per test step**: Makes failures easier to diagnose
2. **Use Page Object Model**: Encapsulates UI selectors, enables reuse
3. **Test user workflows**: Not implementation details
4. **Keep tests independent**: No reliance on test execution order
5. **Use fixtures**: For setup/teardown and test isolation
6. **Name tests clearly**: Describe the user scenario, not the implementation
7. **Capture evidence**: Screenshots/videos for debugging
8. **Tag tests**: For selective execution (`@smoke`, `@critical`, `@accessibility`)

## Common Commands Cheat Sheet

| Command | Purpose |
|---------|---------|
| `npm run test:e2e` | Run all e2e tests |
| `npm run test:e2e -- --headed` | Run tests with visible browser |
| `npm run test:e2e -- --debug` | Run with Playwright debugger |
| `npm run test:e2e -- --ui` | Run with interactive dashboard |
| `npm run test:e2e -- --project=chromium` | Run on specific browser |
| `npx playwright show-report` | View test report in HTML |
| `npm run test:e2e -- --update-snapshots` | Update visual regression baselines |

## Next Steps

1. Review existing test examples in `tests/e2e/specs/`
2. Create Page Objects for your features
3. Write your first test using the patterns above
4. Run tests locally: `npm run test:e2e`
5. Check the HTML report: `npx playwright show-report`

## Additional Resources

- [Playwright Documentation](https://playwright.dev)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [axe DevTools Browser Extension](https://www.deque.com/axe/devtools/)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
