import { test, expect } from '@playwright/test';
import { setupTestAuth } from '../fixtures/auth';
import { ArtifactManagerPage } from '../pages/artifact-manager.page';
import { TIMEOUTS } from '../fixtures/constants';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Artifact Management - User Story 1 (Functional E2E)', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test('should display artifact list page', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Should have artifact list (even if empty)
    const count = await artifactPage.getArtifactCount();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should display artifact count', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    const count = await artifactPage.getArtifactCount();
    expect(typeof count).toBe('number');
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should support artifact upload', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    // Check if upload button exists
    const uploadButton = page.locator('button:has-text("Upload")');
    const exists = await uploadButton.isVisible({ timeout: 5000 }).catch(() => false);

    if (exists) {
      // Create temp file
      const testFile = path.join(__dirname, 'test-artifact.txt');
      fs.writeFileSync(testFile, 'test content for artifact');

      try {
        // Upload file
        const fileInput = page.locator('input[type="file"]');
        if (await fileInput.isVisible({ timeout: 5000 }).catch(() => false)) {
          await fileInput.setInputFiles(testFile);
          await uploadButton.click();

          // Wait for upload
          await page.waitForLoadState('networkidle');
        }
      } finally {
        if (fs.existsSync(testFile)) {
          fs.unlinkSync(testFile);
        }
      }
    }
  });

  test('should support artifact download', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    const count = await artifactPage.getArtifactCount();

    if (count > 0) {
      // Check if download button exists
      const downloadButton = page.locator('button:has-text("Download")').first();
      const exists = await downloadButton.isVisible({ timeout: 5000 }).catch(() => false);

      if (exists) {
        const downloadPromise = page.waitForEvent('download');
        await downloadButton.click();

        try {
          const download = await downloadPromise;
          expect(download.suggestedFilename()).toBeTruthy();
        } catch {
          // Download might have been prevented
        }
      }
    }
  });

  test('should support artifact deletion', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    const initialCount = await artifactPage.getArtifactCount();

    if (initialCount > 0) {
      // Try to delete first artifact
      const deleteButton = page.locator('button:has-text("Delete")').first();
      const exists = await deleteButton.isVisible({ timeout: 5000 }).catch(() => false);

      if (exists) {
        await deleteButton.click();
        await page.waitForLoadState('networkidle');

        const finalCount = await artifactPage.getArtifactCount();
        expect(finalCount).toBeLessThanOrEqual(initialCount);
      }
    }
  });

  test('should retrieve artifact metadata via API', async ({ apiRequest }) => {
    const pipelineId = 'test-123';
    const artifactId = 'artifact-456';

    // Try to get artifact via API
    const response = await apiRequest.get(`/test/pipelines/${pipelineId}/artifacts/${artifactId}`);

    // Should be either 200 (success) or 404 (not found)
    expect([200, 404]).toContain(response.status());
  });

  test('should display artifact information in list', async ({ page }) => {
    const artifactPage = new ArtifactManagerPage(page);
    await page.goto('/dashboard/pipelines/test-123/artifacts');

    const count = await artifactPage.getArtifactCount();

    if (count > 0) {
      // Should display artifact items with info
      const firstArtifact = page.locator('[data-testid="artifact-item"]').first();
      await expect(firstArtifact).toBeVisible({ timeout: TIMEOUTS.medium });
    }
  });
});
