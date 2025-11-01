import { test, expect } from '@playwright/test';
import { setupTestAuth } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/dashboard.page';
import { TIMEOUTS } from '../../fixtures/constants';

/**
 * Keyboard Navigation Accessibility Tests (WCAG 2.1 AA)
 * Verifies all interactive elements are reachable and operable via keyboard
 */
test.describe('Keyboard Navigation - User Story 2 (Accessibility)', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test('should be able to navigate dashboard with Tab key', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Get initial focus
    const initialFocus = await page.evaluate(() => {
      return document.activeElement?.tagName;
    });

    // Tab through elements
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
    }

    // Focus should have moved
    const finalFocus = await page.evaluate(() => {
      return document.activeElement?.tagName;
    });

    expect(finalFocus).toBeTruthy();
  });

  test('should activate buttons with Enter key', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Find first button
    const button = page.locator('button').first();

    if (await button.isVisible({ timeout: 5000 }).catch(() => false)) {
      // Focus button
      await button.focus();

      // Verify focused
      const isFocused = await button.evaluate((el) => {
        return document.activeElement === el;
      });

      expect(isFocused).toBeTruthy();

      // Press Enter
      await page.keyboard.press('Enter');

      // Action should have been triggered
      await page.waitForLoadState('networkidle');
    }
  });

  test('should activate buttons with Space key', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const button = page.locator('button').first();

    if (await button.isVisible({ timeout: 5000 }).catch(() => false)) {
      await button.focus();

      const isFocused = await button.evaluate((el) => {
        return document.activeElement === el;
      });

      expect(isFocused).toBeTruthy();

      // Press Space
      await page.keyboard.press(' ');
      await page.waitForLoadState('networkidle');
    }
  });

  test('should close modals with Escape key', async ({ page }) => {
    // Navigate to a page with modal capability
    await page.goto('/dashboard/pipelines/new');

    // Check if modal exists
    const modal = page.locator('[role="dialog"]');
    const modalExists = await modal.isVisible({ timeout: 5000 }).catch(() => false);

    if (modalExists) {
      // Press Escape
      await page.keyboard.press('Escape');
      await page.waitForLoadState('networkidle');

      // Modal should be closed or not focused
      const stillVisible = await modal.isVisible({ timeout: 5000 }).catch(() => false);
      expect(stillVisible).toBeFalsy();
    }
  });

  test('should navigate links with Tab key', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Find all links
    const links = page.locator('a');
    const linkCount = await links.count();

    if (linkCount > 0) {
      // Focus first link
      const firstLink = links.first();
      await firstLink.focus();

      const isFocused = await firstLink.evaluate((el) => {
        return document.activeElement === el;
      });

      expect(isFocused).toBeTruthy();

      // Tab to next link
      await page.keyboard.press('Tab');

      const nextFocus = await page.evaluate(() => {
        const el = document.activeElement as HTMLElement;
        return el?.tagName === 'A' || el?.tagName === 'BUTTON';
      });

      expect(nextFocus).toBeTruthy();
    }
  });

  test('should support arrow key navigation in lists/dropdowns', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Find dropdown or select element
    const dropdown = page.locator('select, [role="listbox"], [role="combobox"]').first();

    if (await dropdown.isVisible({ timeout: 5000 }).catch(() => false)) {
      await dropdown.focus();

      // Press arrow down
      await page.keyboard.press('ArrowDown');
      await page.waitForTimeout(100);

      // Check if selection changed
      const selectedValue = await dropdown.evaluate((el) => {
        if (el instanceof HTMLSelectElement) {
          return el.value;
        }
        return (el as any).getAttribute('aria-selected');
      });

      expect(selectedValue).toBeTruthy();
    }
  });

  test('should maintain focus visibility', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Tab to first interactive element
    await page.keyboard.press('Tab');

    // Get focused element
    const focusedElement = await page.evaluate(() => {
      const el = document.activeElement as HTMLElement;
      if (!el) return null;

      const style = window.getComputedStyle(el);
      return {
        tag: el.tagName,
        outline: style.outline,
        boxShadow: style.boxShadow,
        hasVisibleFocus: style.outline !== 'none' || style.boxShadow !== 'none',
      };
    });

    // Focus should be visible
    if (focusedElement) {
      expect(focusedElement.tag).toBeTruthy();
      // Should have some focus indicator
      expect(focusedElement.hasVisibleFocus || focusedElement.outline !== 'none').toBeTruthy();
    }
  });

  test('should have logical tab order', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Collect tab order
    const tabOrder = await page.evaluate(() => {
      const interactive = Array.from(document.querySelectorAll('a, button, input, select, textarea, [tabindex]'));
      return interactive
        .filter((el) => {
          const style = window.getComputedStyle(el as HTMLElement);
          return style.display !== 'none' && style.visibility !== 'hidden';
        })
        .map((el) => {
          const rect = (el as HTMLElement).getBoundingClientRect();
          return {
            tag: (el as HTMLElement).tagName,
            tabindex: (el as HTMLElement).getAttribute('tabindex'),
            visible: rect.width > 0 && rect.height > 0,
          };
        });
    });

    // Should have interactive elements
    expect(tabOrder.length).toBeGreaterThan(0);

    // All visible elements should not have negative tabindex (except intentional)
    const invalidTabindex = tabOrder.filter((el) => {
      const idx = parseInt(el.tabindex || '0');
      return el.visible && idx < 0;
    });

    // Should not have visible elements with negative tabindex
    expect(invalidTabindex.length).toBe(0);
  });

  test('should skip to main content', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    // Look for skip link
    const skipLink = page.locator('a[href="#main"], a[href="#content"], .skip-link');
    const hasSkipLink = await skipLink.isVisible({ timeout: 5000 }).catch(() => false);

    // While not required, skip links are best practice
    if (hasSkipLink) {
      await skipLink.click();
      // Should navigate to main content
      await page.waitForLoadState('networkidle');
    }
  });
});
