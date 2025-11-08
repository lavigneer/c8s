import { Page, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Screenshot utilities for documentation generation
 * Provides helpers to capture and organize screenshots for embedding in docs
 */

export interface ScreenshotOptions {
  /** Directory to save screenshots (relative to project root) */
  outputDir?: string;
  /** Wait time before capturing screenshot (ms) */
  waitTime?: number;
  /** Whether to scroll to element before capturing */
  scrollIntoView?: boolean;
  /** Full page screenshot instead of viewport */
  fullPage?: boolean;
  /** Element selector to screenshot (if null, captures full page) */
  element?: string;
  /** Custom filename (without .png extension) */
  filename?: string;
}

/**
 * Capture a screenshot for documentation
 * Automatically organizes screenshots by feature/page
 */
export async function captureScreenshot(
  page: Page,
  feature: string,
  screenName: string,
  options: ScreenshotOptions = {}
) {
  const {
    outputDir = 'docs/screenshots',
    waitTime = 500,
    scrollIntoView = true,
    fullPage = false,
    element,
    filename,
  } = options;

  // Create feature subdirectory
  const featureDir = path.join(outputDir, feature);
  if (!fs.existsSync(featureDir)) {
    fs.mkdirSync(featureDir, { recursive: true });
  }

  // Wait for dynamic content to load
  await page.waitForTimeout(waitTime);

  // Handle element-specific screenshot
  if (element) {
    const elementHandle = await page.locator(element);
    await expect(elementHandle).toBeVisible();
    if (scrollIntoView) {
      await elementHandle.scrollIntoViewIfNeeded();
      await page.waitForTimeout(200);
    }
  }

  // Generate filename
  const screenshotName = filename || screenName;
  const screenshotPath = path.join(featureDir, `${screenshotName}.png`);

  // Capture screenshot
  if (element) {
    const elementHandle = await page.locator(element).first();
    await elementHandle.screenshot({ path: screenshotPath });
  } else {
    await page.screenshot({
      path: screenshotPath,
      fullPage,
    });
  }

  console.log(`✓ Screenshot saved: ${screenshotPath}`);
  return screenshotPath;
}

/**
 * Capture multiple screenshots at different viewports
 * Useful for responsive design documentation
 */
export async function captureResponsiveScreenshots(
  page: Page,
  feature: string,
  screenName: string,
  viewports: Record<string, { width: number; height: number }> = {},
  options: ScreenshotOptions = {}
) {
  const defaultViewports = {
    desktop: { width: 1920, height: 1080 },
    tablet: { width: 1024, height: 1366 },
    mobile: { width: 390, height: 844 },
    ...viewports,
  };

  const screenshots: Record<string, string> = {};

  for (const [viewportName, dimensions] of Object.entries(defaultViewports)) {
    await page.setViewportSize(dimensions);
    await page.waitForTimeout(300);

    const featureDir = path.join(options.outputDir || 'docs/screenshots', feature);
    if (!fs.existsSync(featureDir)) {
      fs.mkdirSync(featureDir, { recursive: true });
    }

    const filename = `${screenName}-${viewportName}`;
    const screenshotPath = path.join(featureDir, `${filename}.png`);

    if (options.element) {
      const elementHandle = await page.locator(options.element).first();
      await elementHandle.screenshot({ path: screenshotPath });
    } else {
      await page.screenshot({
        path: screenshotPath,
        fullPage: options.fullPage || false,
      });
    }

    screenshots[viewportName] = screenshotPath;
    console.log(`✓ Responsive screenshot (${viewportName}): ${screenshotPath}`);
  }

  return screenshots;
}

/**
 * Generate markdown snippet for embedding screenshot
 * Use this to quickly add screenshots to your documentation
 */
export function generateMarkdownSnippet(
  screenshotPath: string,
  altText: string,
  title?: string
): string {
  // Make path relative to docs folder
  const relativePath = screenshotPath.replace('docs/', '');
  const markdownTitle = title ? `${title}\n\n` : '';

  return `${markdownTitle}![${altText}](./${relativePath})`;
}

/**
 * Generate markdown snippet with responsive screenshots
 */
export function generateResponsiveMarkdownSnippet(
  screenshots: Record<string, string>,
  altText: string,
  title?: string
): string {
  const markdownTitle = title ? `${title}\n\n` : '';

  let markdown = markdownTitle;
  markdown += `**Desktop:**\n`;
  markdown += `![${altText} - Desktop](./${screenshots.desktop.replace('docs/', '')})\n\n`;

  markdown += `**Tablet:**\n`;
  markdown += `![${altText} - Tablet](./${screenshots.tablet.replace('docs/', '')})\n\n`;

  markdown += `**Mobile:**\n`;
  markdown += `![${altText} - Mobile](./${screenshots.mobile.replace('docs/', '')})\n`;

  return markdown;
}

/**
 * Wait for specific UI elements to be ready before screenshot
 * Useful for ensuring dynamic content is loaded
 */
export async function waitForUIReady(
  page: Page,
  readinessSelectors: string[] = []
) {
  // Default selectors that indicate the page is ready
  const defaultSelectors = [
    '[data-testid="page-loaded"]',
    '.dashboard-content',
    'main',
  ];

  const selectorsToWait = readinessSelectors.length > 0 ? readinessSelectors : defaultSelectors;

  for (const selector of selectorsToWait) {
    try {
      await page.locator(selector).first().waitFor({ state: 'visible', timeout: 5000 });
    } catch {
      // Selector may not exist on all pages, continue
      continue;
    }
  }
}

/**
 * Create a screenshot index file documenting all captured screenshots
 */
export function createScreenshotIndex(
  outputDir: string = 'docs/screenshots'
): Record<string, string[]> {
  const index: Record<string, string[]> = {};

  if (!fs.existsSync(outputDir)) {
    return index;
  }

  const features = fs.readdirSync(outputDir);

  for (const feature of features) {
    const featurePath = path.join(outputDir, feature);
    if (fs.statSync(featurePath).isDirectory()) {
      const screenshots = fs.readdirSync(featurePath)
        .filter(file => file.endsWith('.png'))
        .map(file => path.join(feature, file));

      if (screenshots.length > 0) {
        index[feature] = screenshots;
      }
    }
  }

  return index;
}

/**
 * Print a summary of all captured screenshots
 */
export function printScreenshotSummary(
  outputDir: string = 'docs/screenshots'
): void {
  const index = createScreenshotIndex(outputDir);
  const totalScreenshots = Object.values(index).reduce((sum, screens) => sum + screens.length, 0);

  console.log('\n═══════════════════════════════════════════');
  console.log('📸 Screenshot Generation Summary');
  console.log('═══════════════════════════════════════════\n');

  for (const [feature, screenshots] of Object.entries(index)) {
    console.log(`📁 ${feature}/`);
    screenshots.forEach(screenshot => {
      console.log(`   └─ ${path.basename(screenshot)}`);
    });
    console.log();
  }

  console.log(`Total: ${totalScreenshots} screenshots captured`);
  console.log(`Location: ${path.resolve(outputDir)}`);
  console.log('\n═══════════════════════════════════════════\n');
}
