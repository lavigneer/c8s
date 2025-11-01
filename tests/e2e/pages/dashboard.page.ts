import { Page, expect } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Dashboard Page Object
 * Main application dashboard with navigation and filtering
 */
export class DashboardPage extends BasePage {
  // Selectors
  private heading = 'h1:has-text("Pipeline History")';
  private pipelineList = '#pipeline-rows';
  private filterInput = 'input[placeholder*="Search"], input[placeholder*="Filter"]';
  private createButton = 'button:has-text("Create"), button:has-text("New")';
  private navProjects = 'a:has-text("Projects")';
  private navSettings = 'a:has-text("Settings")';

  async goto() {
    await super.goto('/dashboard');
    await this.page.waitForLoadState('networkidle');
    await expect(this.page.locator(this.heading)).toBeVisible();
  }

  async verifyHeading() {
    await expect(this.page.locator(this.heading)).toBeVisible();
  }

  async searchPipeline(query: string) {
    await this.fill(this.filterInput, query);
    await this.page.keyboard.press('Enter');
    await this.page.waitForLoadState('networkidle');
  }

  async clickCreateButton() {
    await this.click(this.createButton);
  }

  async navigateToProjects() {
    await this.click(this.navProjects);
    await this.page.waitForURL(/projects/);
  }

  async navigateToSettings() {
    await this.click(this.navSettings);
    await this.page.waitForURL(/settings/);
  }

  async getPipelineCount(): Promise<number> {
    const rows = this.page.locator(this.pipelineList);
    return await rows.count();
  }

  async isPipelineVisible(pipelineName: string): Promise<boolean> {
    return await this.page.locator(`text="${pipelineName}"`).isVisible();
  }
}
