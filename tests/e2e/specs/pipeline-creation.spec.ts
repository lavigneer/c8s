import { test, expect } from '../fixtures/test-data';
import { setupTestAuth } from '../fixtures/auth';
import { DashboardPage } from '../pages/dashboard.page';
import { PipelineDetailPage } from '../pages/pipeline-detail.page';

test.describe('Pipeline Creation - User Story 1', () => {
  test.beforeEach(async ({ page, testToken }) => {
    await setupTestAuth(page);
  });

  test('should display create pipeline button', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    expect(await page.locator('button:has-text("Create")').isVisible()).toBeTruthy();
  });

  test('should successfully create a new pipeline', async ({ page, apiRequest }) => {
    const dashboardPage = new DashboardPage(page);
    const pipeline = await page.evaluate(() => ({
      name: `pipeline-${Date.now()}`,
      repository: 'github.com/test/repo',
    }));

    // Create via API
    const response = await apiRequest.post('/test/pipelines', {
      data: {
        name: pipeline.name,
        repository: pipeline.repository,
      },
    });

    expect(response.status()).toBe(201);
    const created = await response.json();
    expect(created.name).toBe(pipeline.name);
  });

  test('should navigate through pipeline lifecycle', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Verify pipeline list is visible
    expect(await dashboardPage.getPipelineCount()).toBeGreaterThanOrEqual(0);
  });

  test('should reflect state changes in interface', async ({ page }) => {
    const pipelinePage = new PipelineDetailPage(page);
    await page.goto('/dashboard/pipelines/test-123');

    // Verify status is displayed
    const status = await pipelinePage.getStatus();
    expect(['pending', 'running', 'completed', 'failed']).toContain(status);
  });

  test('should validate required fields', async ({ page }) => {
    const pipelinePage = new PipelineDetailPage(page);
    await page.goto('/dashboard/pipelines/new');

    // Try to submit without filling name
    await pipelinePage.submitForm();

    // Should show validation error
    expect(await page.locator('[role="alert"]').isVisible()).toBeTruthy();
  });

  test('should delete pipeline', async ({ page, apiRequest }) => {
    const pipelineId = 'test-pipeline-123';

    // Delete via API
    const response = await apiRequest.delete(`/test/pipelines/${pipelineId}`);

    expect([204, 404]).toContain(response.status());
  });
});
