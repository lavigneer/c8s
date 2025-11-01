import { test as base, Page } from '@playwright/test';
import { BasePage } from '../pages/base.page';

/**
 * Page Object fixture provider
 * Provides pre-instantiated page objects for use in tests
 */

type PageObjectsFixture = {
  basePage: BasePage;
};

export const test = base.extend<PageObjectsFixture>({
  basePage: async ({ page }, use) => {
    // Create base page object
    const basePage = new BasePage(page);
    await use(basePage);
  },
});

/**
 * Create a page object instance
 * Used in test fixtures to provide page objects
 */
export function createBasePage(page: Page): BasePage {
  return new BasePage(page);
}

export { expect } from '@playwright/test';
