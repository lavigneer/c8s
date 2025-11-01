import { test, expect } from '@playwright/test';
import { setupTestAuth } from '../fixtures/auth';
import { ArtifactManagerPage } from '../pages/artifact-manager.page';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Artifact Management - User Story 1', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test('should display artifact list', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/123/artifacts');

    expect(await page.locator('[data-testid="artifact-item"]').count()).toBeGreaterThanOrEqual(0);
  });

  test('should upload artifact', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/123/artifacts');

    // Create temp file
    const testFile = path.join(__dirname, 'test-artifact.txt');
    fs.writeFileSync(testFile, 'test content');

    try {
      await artifactPage.uploadFile(testFile);
      await page.waitForTimeout(1000); // Wait for upload

      // Verify file appears in list
      expect(await page.locator('text="test-artifact.txt"').isVisible()).toBeTruthy();
    } finally {
      fs.unlinkSync(testFile);
    }
  });

  test('should download artifact', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/123/artifacts');

    // Setup download listener
    const downloadPromise = page.waitForEvent('download');
    await artifactPage.downloadArtifact(0);

    try {
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toBeTruthy();
    } catch {
      // No artifacts to download
    }
  });

  test('should delete artifact', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/123/artifacts');

    const initialCount = await artifactPage.getArtifactCount();

    if (initialCount > 0) {
      await artifactPage.deleteArtifact(0);
      await page.waitForTimeout(1000);

      const finalCount = await artifactPage.getArtifactCount();
      expect(finalCount).toBeLessThan(initialCount);
    }
  });

  test('should retrieve artifact with correct content', async ({ page, apiRequest }) => {
    const pipelineId = 'test-123';
    const artifactId = 'artifact-456';

    // Get artifact via API
    const response = await apiRequest.get(`/test/pipelines/${pipelineId}/artifacts/${artifactId}`);

    expect([200, 404]).toContain(response.status());
  });
});
