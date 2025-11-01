import { Page } from '@playwright/test';
import { BasePage } from './base.page';

/**
 * Artifact Manager Page Object
 * Artifact upload, download, and management
 */
export class ArtifactManagerPage extends BasePage {
  private uploadInput = 'input[type="file"]';
  private uploadButton = 'button:has-text("Upload")';
  private artifactList = '[data-testid="artifact-item"]';
  private downloadButton = 'button:has-text("Download")';
  private deleteButton = 'button:has-text("Delete")';

  async uploadFile(filePath: string) {
    await this.page.locator(this.uploadInput).setInputFiles(filePath);
    await this.click(this.uploadButton);
  }

  async getArtifactCount(): Promise<number> {
    const artifacts = this.page.locator(this.artifactList);
    return await artifacts.count();
  }

  async downloadArtifact(index: number = 0) {
    const downloadButtons = this.page.locator(this.downloadButton);
    await downloadButtons.nth(index).click();
  }

  async deleteArtifact(index: number = 0) {
    const deleteButtons = this.page.locator(this.deleteButton);
    await deleteButtons.nth(index).click();
  }

  async isArtifactVisible(name: string): Promise<boolean> {
    return await this.page.locator(`text="${name}"`).isVisible();
  }
}
