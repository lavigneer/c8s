import { Page, expect } from '@playwright/test';

/**
 * Authentication fixture setup
 * Injects test authentication token into localStorage
 */

export async function setupTestAuth(page: Page) {
  /**
   * Inject test authentication token into page before navigation
   * This allows tests to access protected resources
   */
  await page.addInitScript(() => {
    const testToken = localStorage.getItem('auth_token') || 'test_token_' + Date.now();
    localStorage.setItem('auth_token', testToken);
    localStorage.setItem('auth_user_id', 'test-user-123');
    localStorage.setItem('auth_user_name', 'Test User');
  });
}

/**
 * Clear authentication from page
 */
export async function clearTestAuth(page: Page) {
  await page.evaluate(() => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user_id');
    localStorage.removeItem('auth_user_name');
  });
}

/**
 * Get current test auth token
 */
export async function getTestAuthToken(page: Page): Promise<string> {
  return (
    (await page.evaluate(() => localStorage.getItem('auth_token'))) || 'test_token_default'
  );
}

/**
 * Verify auth token is present
 */
export async function verifyAuthTokenPresent(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    const token = localStorage.getItem('auth_token');
    return !!token && token.length > 0;
  });
}

export { expect };
