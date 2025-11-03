import { Page } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Log Viewer Page Object
 * Real-time log viewing and streaming
 */
export class LogViewerPage extends BasePage {
  private logContainer = '[data-testid="logs"]';
  private logLine = '[data-testid="log-line"]';
  private filterInput = 'input[placeholder*="Filter"]';
  private downloadButton = 'button:has-text("Download")';

  async waitForLogs() {
    await this.page.waitForSelector(this.logContainer);
  }

  async getLogCount(): Promise<number> {
    try {
      const logs = this.page.locator(this.logLine);
      const count = await logs.count();
      return count;
    } catch {
      // Log elements might not exist yet
      return 0;
    }
  }

  async filterLogs(query: string) {
    try {
      const filterInput = this.page.locator(this.filterInput);
      const exists = await filterInput.isVisible({ timeout: 1000 }).catch(() => false);
      if (exists) {
        await this.fill(this.filterInput, query);
        await this.page.waitForLoadState('networkidle');
      } else {
        // Filter input not found, skip
        throw new Error('Filter input not available');
      }
    } catch (e) {
      // Filter might not be available
      throw e;
    }
  }

  async downloadLogs() {
    await this.click(this.downloadButton);
  }

  async getLogText(): Promise<string> {
    return await this.getText(this.logContainer);
  }
}
