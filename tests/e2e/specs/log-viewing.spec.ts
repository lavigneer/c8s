import { test, expect } from '@playwright/test';
import { setupTestAuth } from '../fixtures/auth';
import { LogViewerPage } from '../pages/log-viewer.page';

test.describe('Log Viewing - User Story 1', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test('should display log viewer for running pipeline', async ({ page }) => {
    const logPage = new LogViewerPage(page);
    await page.goto('/dashboard/pipelines/123/logs');

    await logPage.waitForLogs();
    expect(await page.locator('[data-testid="logs"]').isVisible()).toBeTruthy();
  });

  test('should stream logs in real-time', async ({ page }) => {
    const logPage = new LogViewerPage(page);
    await page.goto('/dashboard/pipelines/123/logs');

    await logPage.waitForLogs();
    const initialCount = await logPage.getLogCount();

    // Wait for more logs
    await page.waitForTimeout(2000);
    const laterCount = await logPage.getLogCount();

    // Count should be same or more (streaming)
    expect(laterCount).toBeGreaterThanOrEqual(initialCount);
  });

  test('should filter logs', async ({ page }) => {
    const logPage = new LogViewerPage(page);
    await page.goto('/dashboard/pipelines/123/logs');

    await logPage.waitForLogs();
    const initialCount = await logPage.getLogCount();

    // Filter logs
    await logPage.filterLogs('ERROR');
    const filteredCount = await logPage.getLogCount();

    // Filtered should have fewer or equal results
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
  });

  test('should search logs', async ({ page }) => {
    const logPage = new LogViewerPage(page);
    await page.goto('/dashboard/pipelines/123/logs');

    await logPage.waitForLogs();

    // Search for specific text
    await logPage.filterLogs('deployment');
    const result = await logPage.getLogText();

    expect(result).toContain('deployment');
  });

  test('should download logs', async ({ page }) => {
    const logPage = new LogViewerPage(page);
    await page.goto('/dashboard/pipelines/123/logs');

    await logPage.waitForLogs();

    // Start download listener
    const downloadPromise = page.waitForEvent('download');
    await logPage.downloadLogs();
    const download = await downloadPromise;

    expect(download.suggestedFilename()).toMatch(/\.log|\.txt/);
  });
});
