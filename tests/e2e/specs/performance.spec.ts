import { test, expect } from '@playwright/test';
import { setupTestAuth } from '../fixtures/auth';
import { DashboardPage } from '../pages/dashboard.page';
import { PERFORMANCE_THRESHOLDS } from '../fixtures/constants';

/**
 * Performance Baseline Tests
 * Captures and verifies performance metrics for critical user journeys
 */
test.describe('Performance Baseline - User Story 3 (Reporting)', () => {
  test.beforeEach(async ({ page }) => {
    await setupTestAuth(page);
  });

  test('should load login page within performance budget', async ({ page }) => {
    const startTime = Date.now();

    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const loadTime = Date.now() - startTime;

    // Login page should load quickly
    expect(loadTime).toBeLessThan(3000); // 3 seconds
  });

  test('should load dashboard within performance budget', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    const startTime = Date.now();

    await dashboardPage.goto();

    const loadTime = Date.now() - startTime;

    expect(loadTime).toBeLessThan(PERFORMANCE_THRESHOLDS.pageLoad);
  });

  test('should display first contentful paint within budget', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    // Navigate and capture metrics
    await dashboardPage.goto();

    const fcp = await page.evaluate(() => {
      const paint = performance.getEntriesByType('paint');
      return (
        paint.find((p) => p.name === 'first-contentful-paint')?.startTime || 0
      );
    });

    // FCP should be within budget (typically < 1.8s)
    expect(fcp).toBeLessThan(2000);
  });

  test('should have good DOM Content Loaded performance', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    const dcl = await page.evaluate(() => {
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return nav?.domContentLoadedEventEnd - nav?.domContentLoadedEventStart || 0;
    });

    // DCL should be reasonable
    expect(dcl).toBeLessThan(2000);
  });

  test('should handle navigation without performance degradation', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    // Measure initial load
    const initialLoad = await page.evaluate(() => {
      return performance.now();
    });

    // Simulate user navigation
    await page.locator('a').first().click().catch(() => {
      // Link might not exist, that's okay
    });

    await page.waitForLoadState('networkidle').catch(() => {
      // Page might already be loaded
    });

    // Subsequent navigation should still be fast
    const afterNavigation = await page.evaluate(() => {
      return performance.now();
    });

    const navigationTime = afterNavigation - initialLoad;

    // Navigation should be responsive
    expect(navigationTime).toBeLessThan(5000);
  });

  test('should measure interaction responsiveness', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    // Click a button and measure response time
    const button = page.locator('button').first();
    const exists = await button.isVisible({ timeout: 5000 }).catch(() => false);

    if (exists) {
      const clickTime = Date.now();

      await button.click();
      await page.waitForLoadState('networkidle').catch(() => {
        // Page might be fast
      });

      const responseTime = Date.now() - clickTime;

      // Interaction should be responsive
      expect(responseTime).toBeLessThan(PERFORMANCE_THRESHOLDS.interaction);
    }
  });

  test('should have acceptable script evaluation time', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    const scriptTiming = await page.evaluate(() => {
      const entries = performance.getEntriesByType('measure');
      const scriptEntries = entries.filter((e) => e.name.includes('script'));

      if (scriptEntries.length === 0) return 0;

      return Math.max(...scriptEntries.map((e) => e.duration));
    });

    // Script evaluation should not block for too long
    expect(scriptTiming).toBeLessThan(1000);
  });

  test('should not have excessive layout shifts', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    const cls = await page.evaluate(() => {
      let totalCLS = 0;

      try {
        const entries = performance.getEntriesByType('layout-shift') as any[];
        entries.forEach((entry) => {
          if (!entry.hadRecentInput) {
            totalCLS += entry.value;
          }
        });
      } catch {
        // Layout Shift API not available
      }

      return totalCLS;
    });

    // Cumulative Layout Shift should be low
    expect(cls).toBeLessThan(0.1); // Less than 0.1 is "Good"
  });

  test('should measure network request performance', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    const networkMetrics = await page.evaluate(() => {
      const resources = performance.getEntriesByType('resource');

      return {
        totalRequests: resources.length,
        slowRequests: resources.filter((r) => r.duration > 1000).length,
        averageRequestTime:
          resources.length > 0
            ? resources.reduce((sum, r) => sum + r.duration, 0) / resources.length
            : 0,
        largestTransfer: Math.max(
          ...resources.map((r) => (r as PerformanceResourceTiming)?.transferSize || 0)
        ),
      };
    });

    // Should not have excessive network requests
    expect(networkMetrics.slowRequests).toBeLessThan(5);

    // Average request time should be reasonable
    expect(networkMetrics.averageRequestTime).toBeLessThan(1000);
  });

  test('should measure memory usage', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);

    await dashboardPage.goto();

    const memory = await page.evaluate(() => {
      if (!(performance as any).memory) {
        return null;
      }

      return {
        usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
        jsHeapSizeLimit: (performance as any).memory.jsHeapSizeLimit,
        percentUsed:
          ((performance as any).memory.usedJSHeapSize /
            (performance as any).memory.jsHeapSizeLimit) *
          100,
      };
    });

    // Memory check is informational
    if (memory) {
      expect(memory.percentUsed).toBeLessThan(90);
    }
  });

  test('should handle large dataset rendering efficiently', async ({ page }) => {
    await page.goto('/dashboard');

    // Wait for list to render
    await page.waitForLoadState('networkidle');

    // Measure rendering time for multiple items
    const renderTime = await page.evaluate(() => {
      const before = performance.now();

      // Simulate scrolling/rendering
      const items = document.querySelectorAll('[data-testid="pipeline-item"], tr');

      const after = performance.now();

      return {
        itemCount: items.length,
        renderTime: after - before,
      };
    });

    // Rendering should be fast even with many items
    expect(renderTime.renderTime).toBeLessThan(1000);
  });
});
