import { test, expect } from '@playwright/test';
import { DashboardPage } from '../pages/dashboard.page';
import { LoginPage } from '../pages/login.page';
import { PipelineDetailPage } from '../pages/pipeline-detail.page';
import { LogViewerPage } from '../pages/log-viewer.page';
import { ArtifactManagerPage } from '../pages/artifact-manager.page';
import {
  captureScreenshot,
  captureResponsiveScreenshots,
  waitForUIReady,
  printScreenshotSummary,
} from '../utils/screenshot-utils';

/**
 * Screenshot generation tests for documentation
 * Run with: npm run test:e2e -- --grep @screenshots
 * OR use: npm run screenshots (when script is added to package.json)
 */

test.describe('@screenshots Documentation Screenshots', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
  });

  // ===== AUTHENTICATION SCREENSHOTS =====

  test('capture login page screenshots @screenshots', async ({ page }) => {
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
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    await captureScreenshot(page, 'dashboard', 'pipeline-list', {
      outputDir: 'docs/screenshots',
      waitTime: 1000,
      fullPage: false,
    });
  });

  test('capture dashboard quick stats @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    // Capture just the stats panel
    await captureScreenshot(page, 'dashboard', 'quick-stats-panel', {
      outputDir: 'docs/screenshots',
      element: '[data-testid="quick-stats"]',
      waitTime: 500,
    });
  });

  test('capture dashboard filters @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    // Open filter panel if it exists
    const filterButton = page.locator('[data-testid="filter-toggle"]');
    if (await filterButton.isVisible()) {
      await filterButton.click();
      await page.waitForTimeout(300);
    }

    await captureScreenshot(page, 'dashboard', 'filter-panel', {
      outputDir: 'docs/screenshots',
      element: '[data-testid="filter-panel"]',
    });
  });

  // ===== PIPELINE DETAIL SCREENSHOTS =====

  test('capture pipeline detail page @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    // Navigate to first pipeline
    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible()) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      const pipelineDetailPage = new PipelineDetailPage(page);
      await pipelineDetailPage.waitForPageLoad();

      await captureScreenshot(page, 'pipeline', 'pipeline-detail', {
        outputDir: 'docs/screenshots',
        waitTime: 500,
      });
    }
  });

  // ===== LOG VIEWER SCREENSHOTS =====

  test('capture log viewer @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    // Navigate to a pipeline with logs
    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible()) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      const logViewerPage = new LogViewerPage(page);
      await logViewerPage.waitForLogsLoaded();

      await captureScreenshot(page, 'logs', 'log-viewer', {
        outputDir: 'docs/screenshots',
        element: '[data-testid="log-viewer"]',
        waitTime: 500,
      });
    }
  });

  test('capture log search functionality @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible()) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      const logViewerPage = new LogViewerPage(page);
      await logViewerPage.waitForLogsLoaded();

      // Perform a search
      const searchInput = page.locator('[data-testid="log-search"]');
      if (await searchInput.isVisible()) {
        await searchInput.fill('error');
        await page.waitForTimeout(300);

        await captureScreenshot(page, 'logs', 'log-search', {
          outputDir: 'docs/screenshots',
          element: '[data-testid="log-viewer"]',
        });
      }
    }
  });

  // ===== ARTIFACT SCREENSHOTS =====

  test('capture artifact manager @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible()) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      // Navigate to artifacts tab
      const artifactsTab = page.locator('[data-testid="artifacts-tab"]');
      if (await artifactsTab.isVisible()) {
        await artifactsTab.click();
        await page.waitForTimeout(500);

        const artifactManagerPage = new ArtifactManagerPage(page);
        await artifactManagerPage.waitForArtifactsLoaded();

        await captureScreenshot(page, 'artifacts', 'artifact-list', {
          outputDir: 'docs/screenshots',
          element: '[data-testid="artifact-list"]',
        });
      }
    }
  });

  // ===== RESPONSIVE DESIGN SCREENSHOTS =====

  test('capture responsive dashboard @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    await captureResponsiveScreenshots(page, 'responsive', 'dashboard', {}, {
      outputDir: 'docs/screenshots',
    });
  });

  test('capture responsive pipeline detail @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible()) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      const pipelineDetailPage = new PipelineDetailPage(page);
      await pipelineDetailPage.waitForPageLoad();

      await captureResponsiveScreenshots(page, 'responsive', 'pipeline-detail', {}, {
        outputDir: 'docs/screenshots',
      });
    }
  });

  // ===== KEYBOARD SHORTCUTS SCREENSHOTS =====

  test('capture keyboard shortcuts help @screenshots', async ({ page }) => {
    await loginPage.navigateToLogin();
    await loginPage.login('test@example.com', 'password');
    await dashboardPage.waitForDashboard();

    // Trigger help modal with '?'
    await page.keyboard.press('Shift+Slash'); // '?' key
    await page.waitForTimeout(300);

    const helpModal = page.locator('[data-testid="help-modal"]');
    if (await helpModal.isVisible()) {
      await captureScreenshot(page, 'keyboard-shortcuts', 'help-modal', {
        outputDir: 'docs/screenshots',
        element: '[data-testid="help-modal"]',
      });
    }
  });

  // ===== SUMMARY =====

  test('generate screenshot summary @screenshots', async () => {
    // This test runs after all screenshots are captured and prints a summary
    printScreenshotSummary('docs/screenshots');
  });
});
