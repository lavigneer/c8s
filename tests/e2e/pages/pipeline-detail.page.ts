import { Page } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Pipeline Detail Page Object
 * Pipeline creation and management
 */
export class PipelineDetailPage extends BasePage {
  private nameInput = 'input[name="name"]';
  private repositoryInput = 'input[name="repository"]';
  private submitButton = 'button[type="submit"]:has-text("Create")';
  private statusBadge = '[data-testid="status"]';
  private runButton = 'button:has-text("Run")';

  async fillPipelineName(name: string) {
    await this.fill(this.nameInput, name);
  }

  async fillRepository(repo: string) {
    await this.fill(this.repositoryInput, repo);
  }

  async submitForm() {
    await this.click(this.submitButton);
  }

  async getStatus(): Promise<string> {
    return await this.getText(this.statusBadge);
  }

  async runPipeline() {
    await this.click(this.runButton);
  }
}
