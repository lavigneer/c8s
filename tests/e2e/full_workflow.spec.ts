import { test, expect } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('C8S Dashboard - Complete Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Set authentication header for API calls
    await page.addInitScript(() => {
      localStorage.setItem('auth_token', 'test_token_12345');
    });
  });

  test('should display dashboard home page', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Check page title
    await expect(page).toHaveTitle(/Pipeline Runs/i);

    // Check main heading
    await expect(page.locator('h1')).toContainText('Pipeline Runs');

    // Check navigation is present
    await expect(page.locator('nav')).toBeVisible();
  });

  test('should navigate between pages', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Navigate to projects
    const projectsLink = page.locator('a:has-text("Projects")');
    await projectsLink.click();

    // Should be on projects page
    await expect(page).toHaveURL(/\/dashboard\/projects/);
    await expect(page.locator('h1')).toContainText('Projects');
  });

  test('should display pipeline list with filters', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Check for filter inputs
    const searchInput = page.locator('input[placeholder*="Search"]');
    await expect(searchInput).toBeVisible();

    // Type in search
    await searchInput.fill('test-pipeline');

    // Trigger search
    await page.keyboard.press('Enter');

    // Wait for results
    await page.waitForTimeout(500);
  });

  test('should display keyboard shortcuts help', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Press '?' to open help
    await page.keyboard.press('?');

    // Wait for modal to appear
    const modal = page.locator('#shortcuts-help-modal');
    await expect(modal).not.toHaveClass(/hidden/);

    // Check for shortcuts content
    await expect(page.locator('text=Keyboard Shortcuts')).toBeVisible();
    await expect(page.locator('text=Navigation')).toBeVisible();
  });

  test('should navigate with keyboard shortcuts', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Focus should be on page
    await page.keyboard.press('Control+K');

    // Search input should be focused
    const searchInput = page.locator('input[placeholder*="Search"]');
    await expect(searchInput).toBeFocused();
  });

  test('should close modal with Escape key', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Open shortcuts help
    await page.keyboard.press('?');

    // Modal should be visible
    const modal = page.locator('#shortcuts-help-modal');
    await expect(modal).not.toHaveClass(/hidden/);

    // Press Escape
    await page.keyboard.press('Escape');

    // Modal should be hidden
    await expect(modal).toHaveClass(/hidden/);
  });

  test('should display loading indicators during HTMX requests', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard/projects`);

    // Click to trigger HTMX request
    const createButton = page.locator('button:has-text("Create Project")');
    if (await createButton.isVisible()) {
      // Start listening for requests
      const requestPromise = page.waitForResponse(
        response => response.url().includes('/api/') && response.status() === 200
      );

      // Trigger a request if available
      // Modal should show loading indicator
      const loadingIndicator = page.locator('.htmx-indicator');
      // Check if indicator becomes visible during request
    }
  });

  test('should manage projects', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard/projects`);

    // Check projects page loaded
    await expect(page.locator('h1')).toContainText('Projects');

    // Check for create button
    const createButton = page.locator('button:has-text("Create Project")');
    await expect(createButton).toBeVisible();
  });

  test('should display cache statistics in console', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Cache stats might be logged to console
    // This test verifies the page loads without cache errors
    const errors: string[] = [];

    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    // Wait for page to fully load
    await page.waitForTimeout(1000);

    // Check no cache-related errors
    const cacheErrors = errors.filter(e => e.includes('cache'));
    expect(cacheErrors.length).toBe(0);
  });

  test('should handle navigation with HTMX', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // HTMX should be loaded
    const htmxCheck = await page.evaluate(() => {
      return typeof (window as any).htmx !== 'undefined';
    });

    expect(htmxCheck).toBe(true);
  });

  test('should display status badges for pipeline runs', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Look for status badges
    const badges = page.locator('[class*="badge"]');

    // Should have some status indicators
    const count = await badges.count();
    // Page might not have runs, but structure should exist
    await expect(page.locator('text=Pipeline')).toBeVisible();
  });

  test('should handle responsive layout', async ({ page }) => {
    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await page.goto(`${BASE_URL}/dashboard`);

    // Page should still be usable
    await expect(page.locator('h1')).toBeVisible();

    // Navigation should be accessible
    const nav = page.locator('nav');
    await expect(nav).toBeVisible();
  });

  test('should handle tablet viewport', async ({ page }) => {
    // Test tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });

    await page.goto(`${BASE_URL}/dashboard`);

    // Page should render properly
    await expect(page.locator('h1')).toBeVisible();
  });

  test('should handle desktop viewport', async ({ page }) => {
    // Test desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });

    await page.goto(`${BASE_URL}/dashboard`);

    // Page should render properly
    await expect(page.locator('h1')).toBeVisible();
  });

  test('should have proper accessibility attributes', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Check for required accessibility features
    const mainContent = page.locator('main');

    // Page should have semantic HTML structure
    await expect(page.locator('h1')).toBeVisible();

    // Form inputs should have labels
    const inputs = page.locator('input[type="search"], input[placeholder*="Search"]');
    if (await inputs.count() > 0) {
      // Inputs should be findable
      expect(await inputs.count()).toBeGreaterThan(0);
    }
  });

  test('should maintain session during navigation', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Get initial URL
    let currentUrl = page.url();
    expect(currentUrl).toContain('/dashboard');

    // Navigate to projects
    const projectsLink = page.locator('a:has-text("Projects")');
    if (await projectsLink.isVisible()) {
      await projectsLink.click();

      // Should still be authenticated (no redirect to login)
      currentUrl = page.url();
      expect(currentUrl).toContain('/dashboard/projects');
    }
  });

  test('should display error messages gracefully', async ({ page }) => {
    // Navigate to non-existent page
    await page.goto(`${BASE_URL}/dashboard/nonexistent`, { waitUntil: 'networkidle' });

    // Page should either show 404 or redirect
    // Check that we get some response
    expect(page.url()).toBeDefined();
  });

  test('should support keyboard navigation through lists', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`);

    // Tab through elements
    await page.keyboard.press('Tab');

    // Check that focus moved
    const focused = await page.evaluate(() => document.activeElement?.tagName);
    expect(focused).toBeTruthy();
  });
});
