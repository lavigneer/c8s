import { test as base, Page } from '@playwright/test';

/**
 * Performance Metrics Fixtures
 * Captures performance data during test execution
 */

type MetricsFixture = {
  metrics: PerformanceData;
  captureNavigationMetrics: () => Promise<PerformanceData>;
  captureInteractionMetrics: (label: string) => Promise<number>;
  captureWebVitals: () => Promise<WebVitals>;
};

export interface PerformanceData {
  pageLoadTime: number;
  domContentLoaded: number;
  firstPaint: number;
  firstContentfulPaint: number;
  largestContentfulPaint: number;
  timeToInteractive: number;
}

export interface WebVitals {
  lcp: number; // Largest Contentful Paint
  fid: number; // First Input Delay
  cls: number; // Cumulative Layout Shift
  fcp: number; // First Contentful Paint
  ttfb: number; // Time to First Byte
}

export interface InteractionMetric {
  label: string;
  duration: number;
  timestamp: string;
}

export const test = base.extend<MetricsFixture>({
  metrics: async ({ page }, use) => {
    const metrics: PerformanceData = {
      pageLoadTime: 0,
      domContentLoaded: 0,
      firstPaint: 0,
      firstContentfulPaint: 0,
      largestContentfulPaint: 0,
      timeToInteractive: 0,
    };

    await use(metrics);
  },

  captureNavigationMetrics: async ({ page }, use) => {
    const captureNavigationMetrics = async (): Promise<PerformanceData> => {
      return await page.evaluate(() => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        const paint = performance.getEntriesByType('paint');
        const lcp = performance.getEntriesByType('largest-contentful-paint').pop();
        const tti = performance.getEntriesByName('tti-polyfill', 'measure').pop();

        return {
          pageLoadTime: navigation?.loadEventEnd - navigation?.loadEventStart || 0,
          domContentLoaded:
            navigation?.domContentLoadedEventEnd - navigation?.domContentLoadedEventStart || 0,
          firstPaint:
            paint.find((p) => p.name === 'first-paint')?.startTime ||
            0,
          firstContentfulPaint:
            paint.find((p) => p.name === 'first-contentful-paint')?.startTime ||
            0,
          largestContentfulPaint: lcp?.startTime || 0,
          timeToInteractive: tti?.duration || 0,
        };
      });
    };

    await use(captureNavigationMetrics);
  },

  captureInteractionMetrics: async ({ page }, use) => {
    const captureInteractionMetrics = async (label: string): Promise<number> => {
      const startTime = Date.now();

      // Perform interaction and wait for response
      await page.waitForLoadState('networkidle');

      const duration = Date.now() - startTime;

      return duration;
    };

    await use(captureInteractionMetrics);
  },

  captureWebVitals: async ({ page }, use) => {
    const captureWebVitals = async (): Promise<WebVitals> => {
      return await page.evaluate(() => {
        // LCP - Largest Contentful Paint
        const lcp = (performance.getEntriesByType('largest-contentful-paint').pop() as any)
          ?.startTime || 0;

        // FID - First Input Delay (deprecated, using INP instead for newer browsers)
        const fid = (performance.getEntriesByType('first-input')[0] as any)?.processingDuration || 0;

        // CLS - Cumulative Layout Shift
        let cls = 0;
        try {
          cls = (performance as any).LayoutShift
            ? Array.from(performance.getEntriesByType('layout-shift')).reduce((sum: number, entry: any) => {
                if (!entry.hadRecentInput) {
                  sum += entry.value;
                }
                return sum;
              }, 0)
            : 0;
        } catch {
          // CLS not available
        }

        // FCP - First Contentful Paint
        const fcp = (performance.getEntriesByType('paint').find((p) => p.name === 'first-contentful-paint') as any)
          ?.startTime || 0;

        // TTFB - Time to First Byte
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        const ttfb = navigation?.responseStart - navigation?.fetchStart || 0;

        return {
          lcp,
          fid,
          cls,
          fcp,
          ttfb,
        };
      });
    };

    await use(captureWebVitals);
  },
});

/**
 * Utility to check if metrics meet thresholds
 */
export function meetsPerformanceThresholds(
  metrics: PerformanceData,
  thresholds = {
    pageLoadTime: 3000,
    domContentLoaded: 2000,
    firstContentfulPaint: 1500,
    largestContentfulPaint: 2500,
  }
): boolean {
  return (
    metrics.pageLoadTime <= thresholds.pageLoadTime &&
    metrics.domContentLoaded <= thresholds.domContentLoaded &&
    metrics.firstContentfulPaint <= thresholds.firstContentfulPaint &&
    metrics.largestContentfulPaint <= thresholds.largestContentfulPaint
  );
}

/**
 * Utility to check Core Web Vitals
 */
export function passesWebVitals(vitals: WebVitals): boolean {
  return (
    vitals.lcp <= 2500 && // LCP should be <= 2.5s
    vitals.fid <= 100 && // FID should be <= 100ms
    vitals.cls <= 0.1 && // CLS should be <= 0.1
    vitals.fcp <= 1800 && // FCP should be <= 1.8s
    vitals.ttfb <= 600 // TTFB should be <= 600ms
  );
}

export { expect } from '@playwright/test';
