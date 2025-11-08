import { test, expect } from '@playwright/test';
import { DashboardPage } from '../pages/dashboard.page';
import { LoginPage } from '../pages/login.page';
import { PipelineDetailPage } from '../pages/pipeline-detail.page';
import { LogViewerPage } from '../pages/log-viewer.page';
import { ArtifactManagerPage } from '../pages/artifact-manager.page';
import { setupTestAuth } from '../fixtures/auth';
import {
  captureScreenshot,
  captureResponsiveScreenshots,
  waitForUIReady,
  printScreenshotSummary,
} from '../utils/screenshot-utils';

/**
 * Screenshot generation tests for documentation
 *
 * These tests capture screenshots of the C8S dashboard for documentation.
 * They are designed to gracefully handle both populated and empty states.
 *
 * Run with: npm run test:e2e -- --grep @screenshots
 * OR use: npm run screenshots
 *
 * Note: Tests will capture the UI as it is. If you want full screenshots with data,
 * ensure your test database has pipeline data, or modify tests to create test data.
 */

test.describe('@screenshots Documentation Screenshots', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);

    // Setup test authentication
    await setupTestAuth(page);
  });

  // ===== AUTHENTICATION SCREENSHOTS =====

  test('capture login page screenshots @screenshots', async ({ page }) => {
    // Don't use setupTestAuth for login page - we want to test the login form itself
    await page.goto('/login');
    await waitForUIReady(page, ['form']);

    // Desktop/default
    await captureScreenshot(page, 'authentication', 'login-page', {
      outputDir: 'docs/screenshots',
      waitTime: 500,
    });

    // Responsive variants
    await captureResponsiveScreenshots(page, 'authentication', 'login-page', {}, {
      outputDir: 'docs/screenshots',
    });
  });

  // ===== DASHBOARD SCREENSHOTS =====

  test('capture dashboard pipeline list @screenshots', async ({ page }) => {
    // Navigate directly to dashboard with auth already set up
    await page.goto('/dashboard');
    await waitForUIReady(page);

    await captureScreenshot(page, 'dashboard', 'pipeline-list', {
      outputDir: 'docs/screenshots',
      waitTime: 1000,
      fullPage: false,
    });
  });

  test('capture dashboard quick stats @screenshots', async ({ page }) => {
    // Navigate directly to dashboard with auth already set up
    await page.goto('/dashboard');
    await waitForUIReady(page);

    // Try to capture stats panel if it exists, otherwise capture what we can
    const statsPanel = page.locator('[data-testid="quick-stats"]');
    if (await statsPanel.isVisible({ timeout: 2000 }).catch(() => false)) {
      await captureScreenshot(page, 'dashboard', 'quick-stats-panel', {
        outputDir: 'docs/screenshots',
        element: '[data-testid="quick-stats"]',
        waitTime: 500,
      });
    }
  });

  test('capture dashboard filters @screenshots', async ({ page }) => {
    // Navigate directly to dashboard with auth already set up
    await page.goto('/dashboard');
    await waitForUIReady(page);

    // Try to open filter panel if it exists
    const filterButton = page.locator('[data-testid="filter-toggle"]');
    if (await filterButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await filterButton.click();
      await page.waitForTimeout(300);

      await captureScreenshot(page, 'dashboard', 'filter-panel', {
        outputDir: 'docs/screenshots',
        element: '[data-testid="filter-panel"]',
      });
    }
  });

  // ===== RESPONSIVE DESIGN SCREENSHOTS =====

  test('capture responsive dashboard @screenshots', async ({ page }) => {
    // Navigate directly to dashboard with auth already set up
    await page.goto('/dashboard');
    await waitForUIReady(page);

    await captureResponsiveScreenshots(page, 'responsive', 'dashboard', {}, {
      outputDir: 'docs/screenshots',
    });
  });

  test('capture responsive login page @screenshots', async ({ page }) => {
    await page.goto('/login');
    await waitForUIReady(page, ['form']);

    await captureResponsiveScreenshots(page, 'responsive', 'login', {}, {
      outputDir: 'docs/screenshots',
    });
  });

  // ===== SUMMARY =====

  test('generate screenshot summary @screenshots', async () => {
    // This test runs after all screenshots are captured and prints a summary
    printScreenshotSummary('docs/screenshots');
  });
});
