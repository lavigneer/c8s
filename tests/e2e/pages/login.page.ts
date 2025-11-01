import { Page, expect } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Login Page Object
 * Handles authentication workflows
 */
export class LoginPage extends BasePage {
  // Selectors
  private emailInput = 'input[type="email"]';
  private passwordInput = 'input[type="password"]';
  private submitButton = 'button[type="submit"]';
  private errorMessage = '[role="alert"]';
  private logoutButton = 'button:has-text("Logout")';

  async goto() {
    await super.goto('/login');
    await expect(this.page).toHaveTitle(/Login|Authentication/);
  }

  async fillEmail(email: string) {
    await this.fill(this.emailInput, email);
  }

  async fillPassword(password: string) {
    await this.fill(this.passwordInput, password);
  }

  async clickSubmit() {
    await this.click(this.submitButton);
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
