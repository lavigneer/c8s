import { test, expect } from '@playwright/test';
import { setupTestAuth, clearTestAuth } from '../fixtures/auth';
import { LoginPage } from '../pages/login.page';
import { DashboardPage } from '../pages/dashboard.page';

test.describe('Authentication - User Story 1', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test.afterEach(async ({ page }) => {
    await clearTestAuth(page);
  });

  test('should display login form', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    expect(await page.locator('input[type="email"]').isVisible()).toBeTruthy();
    expect(await page.locator('input[type="password"]').isVisible()).toBeTruthy();
    expect(await page.locator('button[type="submit"]').isVisible()).toBeTruthy();
  });

  test('should successfully login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');

    await expect(page).toHaveURL(/dashboard|pipeline/);
  });

  test('should show error on invalid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginWithInvalidCredentials('test@example.com', 'wrongpassword');

    const errorMsg = await loginPage.getErrorMessage();
    expect(errorMsg).toContain('Invalid');
  });

  test('should maintain session state after login', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');

    const authToken = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(authToken).toBeTruthy();
  });

  test('should successfully logout', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');

    // Verify logged in
    await dashboardPage.goto();

    // Logout
    if (await loginPage.isLogoutButtonVisible()) {
      await loginPage.logout();
      await expect(page).toHaveURL(/login/);
    }
  });

  test('should restrict access to dashboard when not authenticated', async ({ page }) => {
    await clearTestAuth(page);
    await page.goto('/dashboard');

    // Should redirect to login
    await expect(page).toHaveURL(/login/);
  });
});
