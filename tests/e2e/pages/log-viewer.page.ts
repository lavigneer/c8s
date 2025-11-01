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
    const logs = this.page.locator(this.logLine);
    return await logs.count();
  }

  async filterLogs(query: string) {
    await this.fill(this.filterInput, query);
    await this.page.waitForLoadState('networkidle');
  }

  async downloadLogs() {
    await this.click(this.downloadButton);
  }

  async getLogText(): Promise<string> {
    return await this.getText(this.logContainer);
  }
}
