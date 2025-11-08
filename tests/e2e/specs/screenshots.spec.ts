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
 * These tests capture desktop-sized (1920x1080) screenshots of the C8S dashboard
 * for use in user-facing documentation.
 *
 * Run with: npm run test:e2e -- --grep @screenshots
 * OR use: npm run screenshots
 *
 * Note: Tests capture the UI as it is. If you want full screenshots with data,
 * ensure your test database has pipeline data.
 */

test.describe('@screenshots Documentation Screenshots', () => {
  test.beforeEach(async ({ page }) => {
    // Setup test authentication for authenticated pages
    await setupTestAuth(page);
    // Set viewport to standard desktop size
    await page.setViewportSize({ width: 1920, height: 1080 });
  });

  // ===== AUTHENTICATION SCREENSHOTS =====

  test('capture login page @screenshots', async ({ page }) => {
    // Don't use auth for login page - we want to capture the login form itself
    await page.goto('/login');
    await waitForUIReady(page, ['form']);

    await captureScreenshot(page, 'authentication', 'login-page', {
      outputDir: 'docs/screenshots',
      waitTime: 500,
    });
  });

  // ===== DASHBOARD SCREENSHOTS =====

  test('capture pipeline history dashboard @screenshots', async ({ page }) => {
    await page.goto('/dashboard');
    await waitForUIReady(page);

    await captureScreenshot(page, 'dashboard', 'pipeline-history', {
      outputDir: 'docs/screenshots',
      waitTime: 1000,
      fullPage: false,
    });
  });

  // ===== PIPELINE DETAIL SCREENSHOTS =====

  test('capture pipeline detail page @screenshots', async ({ page }) => {
    // Try to navigate to a pipeline detail page
    await page.goto('/dashboard');
    await waitForUIReady(page);

    // Look for a pipeline run to click on
    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      await captureScreenshot(page, 'pipeline', 'pipeline-detail', {
        outputDir: 'docs/screenshots',
        waitTime: 500,
      });
    }
  });

  test('capture pipeline logs view @screenshots', async ({ page }) => {
    await page.goto('/dashboard');
    await waitForUIReady(page);

    // Try to navigate to logs for a pipeline
    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      // Try to find and click logs tab
      const logsTab = page.locator('[data-testid="logs-tab"]');
      if (await logsTab.isVisible({ timeout: 2000 }).catch(() => false)) {
        await logsTab.click();
        await page.waitForTimeout(500);

        await captureScreenshot(page, 'logs', 'log-viewer', {
          outputDir: 'docs/screenshots',
          waitTime: 500,
        });
      }
    }
  });

  test('capture pipeline artifacts view @screenshots', async ({ page }) => {
    await page.goto('/dashboard');
    await waitForUIReady(page);

    // Try to navigate to artifacts for a pipeline
    const pipelineLink = page.locator('[data-testid="pipeline-run-item"]').first();
    if (await pipelineLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await pipelineLink.click();
      await page.waitForTimeout(1000);

      // Try to find and click artifacts tab
      const artifactsTab = page.locator('[data-testid="artifacts-tab"]');
      if (await artifactsTab.isVisible({ timeout: 2000 }).catch(() => false)) {
        await artifactsTab.click();
        await page.waitForTimeout(500);

        await captureScreenshot(page, 'artifacts', 'artifact-list', {
          outputDir: 'docs/screenshots',
          waitTime: 500,
        });
      }
    }
  });

  // ===== PROJECTS SCREENSHOTS =====

  test('capture projects page @screenshots', async ({ page }) => {
    await page.goto('/projects');
    await waitForUIReady(page);

    await captureScreenshot(page, 'projects', 'projects-list', {
      outputDir: 'docs/screenshots',
      waitTime: 500,
    });
  });

  // ===== SUMMARY =====

  test('generate screenshot summary @screenshots', async () => {
    // This test runs after all screenshots are captured and prints a summary
    printScreenshotSummary('docs/screenshots');
  });
});
