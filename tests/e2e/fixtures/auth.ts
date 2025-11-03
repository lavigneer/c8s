import { Page, expect } from '@playwright/test';

/**
 * Authentication fixture setup
 * Injects test authentication token into localStorage
 */

export async function setupTestAuth(page: Page) {
  /**
   * Set test authentication cookie
   * The backend expects an 'auth_token' HTTP cookie for authentication
   */
  const testToken = 'test_token_' + Date.now();

  // Set the auth_token cookie that the backend expects
  await page.context().addCookies([
    {
      name: 'auth_token',
      value: testToken,
      url: 'http://localhost:8080'
    }
  ]);

  // Also set localStorage for any client-side usage
  await page.evaluate(({ token }) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('auth_user_id', 'test-user-123');
    localStorage.setItem('auth_user_name', 'Test User');
  }, { token: testToken });
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
