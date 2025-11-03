import { Page, expect } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Login Page Object
 * Handles authentication workflows
 */
export class LoginPage extends BasePage {
  // Selectors
  private usernameInput = 'input[name="username"]';
  private passwordInput = 'input[name="password"]';
  private submitButton = 'button[type="submit"]';
  private errorMessage = '[role="alert"]';
  private logoutButton = 'button:has-text("Logout")';

  async goto() {
    await super.goto('/login');
    await expect(this.page).toHaveTitle(/Login|Dashboard|Pipeline/);
  }

  async fillUsername(username: string) {
    await this.fill(this.usernameInput, username);
  }

  async fillEmail(email: string) {
    // For backward compatibility, treat email as username
    await this.fill(this.usernameInput, email);
  }

  async fillPassword(password: string) {
    await this.fill(this.passwordInput, password);
  }

  async clickSubmit() {
    // Click submit - may trigger navigation or validation errors
    const submitPromise = this.click(this.submitButton);

    // Try to wait for navigation, but don't fail if it doesn't happen
    // (e.g., when form validation fails)
    await Promise.race([
      this.page.waitForNavigation({ waitUntil: 'networkidle', timeout: 2000 }).catch(() => {}),
      this.page.waitForTimeout(500), // Give time for validation or navigation to happen
      submitPromise,
    ]);

    // Wait for any pending requests
    await this.page.waitForLoadState('networkidle').catch(() => {});
  }

  async login(email: string, password: string) {
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.clickSubmit();
    await this.page.waitForURL(/dashboard|pipeline/);
  }

  async loginWithInvalidCredentials(email: string, password: string) {
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.clickSubmit();
    await this.waitForSelector(this.errorMessage);
  }

  async getErrorMessage(): Promise<string> {
    return await this.getText(this.errorMessage);
  }

  async logout() {
    await this.click(this.logoutButton);
    await this.page.waitForURL(/login/);
  }

  async isLogoutButtonVisible(): Promise<boolean> {
    return await this.isVisible(this.logoutButton);
  }
}
