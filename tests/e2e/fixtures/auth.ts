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

  // Set localStorage before page navigation using addInitScript
  // This ensures localStorage is available before any page loads
  await page.addInitScript(({ token }) => {
    try {
      localStorage.setItem('auth_token', token);
      localStorage.setItem('auth_user_id', 'test-user-123');
      localStorage.setItem('auth_user_name', 'Test User');
    } catch (e) {
      // localStorage might not be available in some contexts (e.g., error pages)
      // This is fine - the backend auth cookie is what matters
    }
  }, { token: testToken });
}

/**
 * Clear authentication from page
 */
export async function clearTestAuth(page: Page) {
  try {
    await page.evaluate(() => {
      try {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user_id');
        localStorage.removeItem('auth_user_name');
      } catch (e) {
        // localStorage might not be available
      }
    });
  } catch (e) {
    // Page context might not be available
  }
}

/**
 * Get current test auth token
 */
export async function getTestAuthToken(page: Page): Promise<string> {
  try {
    return (
      (await page.evaluate(() => {
        try {
          return localStorage.getItem('auth_token');
        } catch {
          return null;
        }
      })) || 'test_token_default'
    );
  } catch (e) {
    return 'test_token_default';
  }
}

/**
 * Verify auth token is present
 */
export async function verifyAuthTokenPresent(page: Page): Promise<boolean> {
  try {
    return await page.evaluate(() => {
      try {
        const token = localStorage.getItem('auth_token');
        return !!token && token.length > 0;
      } catch {
        return false;
      }
    });
  } catch (e) {
    return false;
  }
}

export { expect };
