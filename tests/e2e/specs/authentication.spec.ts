import { test, expect } from '@playwright/test';
import { setupTestAuth, clearTestAuth, getTestAuthToken } from '../fixtures/auth';
import { LoginPage } from '../pages/login.page';
import { DashboardPage } from '../pages/dashboard.page';
import { TIMEOUTS } from '../fixtures/constants';

test.describe('Authentication - User Story 1 (Functional E2E)', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test.afterEach(async ({ page }) => {
    await clearTestAuth(page);
  });

  test('should display login form with all required fields', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    // Verify email input
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: TIMEOUTS.medium });

    // Verify password input
    await expect(page.locator('input[type="password"]')).toBeVisible();

    // Verify submit button
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should successfully login with valid credentials and store session', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    // Fill and submit
    await loginPage.fillEmail('test@example.com');
    await loginPage.fillPassword('password123');
    await loginPage.clickSubmit();

    // Should navigate away from login
    await page.waitForURL(/dashboard|pipeline|home/, { timeout: TIMEOUTS.medium });
    await expect(page).not.toHaveURL(/login/);

    // Verify auth token is stored
    const token = await getTestAuthToken(page);
    expect(token).toBeTruthy();
    expect(token.length).toBeGreaterThan(0);
  });

  test('should maintain session state after successful login', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');

    // Get token after login
    const authToken = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(authToken).toBeTruthy();

    // Navigate away and back - token should still exist
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    const tokenAfterNav = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(tokenAfterNav).toBe(authToken);
  });

  test('should show error message on invalid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    // Try login with wrong password
    await loginPage.fillEmail('test@example.com');
    await loginPage.fillPassword('wrongpassword');
    await loginPage.clickSubmit();

    // Wait for error message
    const errorLocator = page.locator('[role="alert"]');
    await expect(errorLocator).toBeVisible({ timeout: TIMEOUTS.medium });

    const errorText = await errorLocator.textContent();
    expect(errorText).toBeTruthy();
  });

  test('should keep user on login page on failed login attempt', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    // Try login with invalid credentials
    await loginPage.fillEmail('nonexistent@example.com');
    await loginPage.fillPassword('wrongpassword');
    await loginPage.clickSubmit();

    // Should still be on login page
    await expect(page).toHaveURL(/login/, { timeout: TIMEOUTS.medium });
  });

  test('should successfully logout and clear session', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);

    // First login
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');

    // Navigate to dashboard
    await page.goto('/dashboard');

    // Verify logged in
    const tokenBeforeLogout = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(tokenBeforeLogout).toBeTruthy();

    // Logout
    if (await loginPage.isLogoutButtonVisible()) {
      await loginPage.logout();

      // Should redirect to login
      await expect(page).toHaveURL(/login/, { timeout: TIMEOUTS.medium });

      // Token should be cleared
      await page.waitForLoadState('networkidle');
      const tokenAfterLogout = await page.evaluate(() => localStorage.getItem('auth_token'));
      // After logout, token might be cleared or reset
      expect(tokenAfterLogout === null || tokenAfterLogout === undefined).toBeTruthy();
    }
  });

  test('should restrict access to protected dashboard route when not authenticated', async ({ page }) => {
    // Clear auth first
    await clearTestAuth(page);

    // Try to access protected route
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should redirect to login
    await expect(page).toHaveURL(/login/, { timeout: TIMEOUTS.medium });
  });

  test('should prevent navigation to protected pages without valid token', async ({ page }) => {
    const token = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(token).toBeTruthy(); // Setup should have token

    // Remove token
    await page.evaluate(() => localStorage.removeItem('auth_token'));

    // Try to access protected page
    await page.goto('/dashboard/pipelines', { waitUntil: 'networkidle' });

    // Should redirect to login
    await expect(page).toHaveURL(/login/, { timeout: TIMEOUTS.medium });
  });
});
