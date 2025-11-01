import { Page, expect } from '@playwright/test';
import { injectAxe, checkA11y } from 'axe-playwright';

/**
 * Base Page Object class providing common functionality for all page objects
 * Includes navigation, waiting, accessibility testing, and screenshot utilities
 */
export class BasePage {
  constructor(protected page: Page) {}

  /**
   * Navigate to a URL relative to baseURL
   */
  async goto(path: string = '') {
    await this.page.goto(path);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Wait for a specific URL pattern
   */
  async waitForURL(urlPattern: string | RegExp) {
    await this.page.waitForURL(urlPattern);
  }

  /**
   * Wait for element to be visible
   */
  async waitForSelector(selector: string, timeout = 5000) {
    await this.page.waitForSelector(selector, { state: 'visible', timeout });
  }

  /**
   * Click an element with implicit wait
   */
  async click(selector: string) {
    await this.page.click(selector);
  }

  /**
   * Fill a text input
   */
  async fill(selector: string, value: string) {
    await this.page.fill(selector, value);
  }

  /**
   * Get text content of an element
   */
  async getText(selector: string): Promise<string> {
    return await this.page.textContent(selector) || '';
  }

  /**
   * Check if element is visible
   */
  async isVisible(selector: string): Promise<boolean> {
    try {
      return await this.page.isVisible(selector);
    } catch {
      return false;
    }
  }

  /**
   * Take a screenshot for debugging/comparison
   */
  async screenshot(name: string) {
    await this.page.screenshot({ path: `test-results/screenshots/${name}.png` });
  }

  /**
   * Inject axe accessibility audit into page
   */
  async injectAccessibilityAudit() {
    await injectAxe(this.page);
  }

  /**
   * Run accessibility audit on page
   * Checks for WCAG 2.1 Level AA compliance
   */
  async runAccessibilityAudit(options?: any) {
    try {
      await checkA11y(this.page, null, {
        detailedReport: true,
        detailedReportOptions: {
          html: true,
        },
        ...options,
      });
    } catch (error) {
      // Accessibility violations found - will be logged
      throw error;
    }
  }

  /**
   * Verify keyboard navigation is working
   * Presses Tab key multiple times and verifies focus moved
   */
  async verifyKeyboardNavigation(numberOfTabs = 5) {
    const initialFocus = await this.page.evaluate(() => {
      return document.activeElement?.tagName;
    });

    for (let i = 0; i < numberOfTabs; i++) {
      await this.page.keyboard.press('Tab');
    }

    const finalFocus = await this.page.evaluate(() => {
      return document.activeElement?.tagName;
    });

    // Focus should have changed (ideally to different elements)
    return initialFocus !== finalFocus;
  }

  /**
   * Verify focus trap in modal/dialog
   * Presses Tab multiple times and ensures focus stays within container
   */
  async verifyFocusTrap(containerSelector: string) {
    const focusedElements: string[] = [];

    for (let i = 0; i < 10; i++) {
      await this.page.keyboard.press('Tab');
      const focused = await this.page.evaluate((selector) => {
        const container = document.querySelector(selector);
        const activeEl = document.activeElement;
        return container?.contains(activeEl as Node);
      }, containerSelector);

      if (!focused) {
        return false; // Focus escaped container
      }
    }

    return true; // Focus remained trapped
  }

  /**
   * Measure page performance metrics
   */
  async getPerformanceMetrics() {
    return await this.page.evaluate(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const paint = performance.getEntriesByType('paint');
      const lcp = performance.getEntriesByType('largest-contentful-paint').pop();

      return {
        navigationTiming: {
          domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
          loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
        },
        firstPaint: paint.find((p) => p.name === 'first-paint')?.startTime || 0,
        firstContentfulPaint: paint.find((p) => p.name === 'first-contentful-paint')?.startTime || 0,
        largestContentfulPaint: lcp?.startTime || 0,
      };
    });
  }

  /**
   * Get page title
   */
  async getTitle(): Promise<string> {
    return await this.page.title();
  }

  /**
   * Verify page has accessibility tree (basic ARIA check)
   */
  async hasAccessibleContent(): Promise<boolean> {
    const roles = await this.page.evaluate(() => {
      const elements = document.querySelectorAll('[role]');
      return elements.length > 0;
    });

    const labels = await this.page.evaluate(() => {
      const inputs = document.querySelectorAll('input');
      return Array.from(inputs).some((input) => {
        return input.hasAttribute('aria-label') || input.hasAttribute('aria-labelledby');
      });
    });

    return roles || labels;
  }
}
