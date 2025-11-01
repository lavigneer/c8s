import { test as base, Page } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Test Reporting Fixtures
 * Provides utilities for capturing test metrics and generating reports
 */

type ReportingFixture = {
  captureMetrics: (name: string) => Promise<void>;
  recordTestResult: (testName: string, passed: boolean, duration: number) => void;
  generateTestReport: () => void;
};

interface TestMetric {
  testName: string;
  timestamp: string;
  duration: number;
  passed: boolean;
  browser: string;
  viewport: string;
}

const metricsFile = path.join(process.cwd(), 'test-results', 'metrics.json');

export const test = base.extend<ReportingFixture>({
  captureMetrics: async ({ page }, use) => {
    const captureMetrics = async (name: string) => {
      try {
        const metrics = await page.evaluate(() => {
          const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
          const paint = performance.getEntriesByType('paint');
          const lcp = performance.getEntriesByType('largest-contentful-paint').pop();

          return {
            domContentLoaded: navigation?.domContentLoadedEventEnd - navigation?.domContentLoadedEventStart,
            pageLoadTime: navigation?.loadEventEnd - navigation?.loadEventStart,
            firstPaint: paint.find((p) => p.name === 'first-paint')?.startTime || 0,
            firstContentfulPaint: paint.find((p) => p.name === 'first-contentful-paint')?.startTime || 0,
            largestContentfulPaint: lcp?.startTime || 0,
          };
        });

        // Save metrics
        const metricsData = {
          testName: name,
          timestamp: new Date().toISOString(),
          metrics,
        };

        console.log(`[Metrics] ${name}:`, metricsData);
      } catch (error) {
        // Metrics capture failed, continue
        console.log(`Could not capture metrics for ${name}`);
      }
    };

    await use(captureMetrics);
  },

  recordTestResult: async ({}, use) => {
    const recordTestResult = (testName: string, passed: boolean, duration: number) => {
      try {
        // Ensure directory exists
        const dir = path.dirname(metricsFile);
        if (!fs.existsSync(dir)) {
          fs.mkdirSync(dir, { recursive: true });
        }

        // Read existing metrics
        let metrics: TestMetric[] = [];
        if (fs.existsSync(metricsFile)) {
          try {
            metrics = JSON.parse(fs.readFileSync(metricsFile, 'utf-8'));
          } catch {
            metrics = [];
          }
        }

        // Add new metric
        const newMetric: TestMetric = {
          testName,
          timestamp: new Date().toISOString(),
          duration,
          passed,
          browser: process.env.BROWSER || 'chromium',
          viewport: process.env.VIEWPORT || 'desktop',
        };

        metrics.push(newMetric);

        // Save updated metrics
        fs.writeFileSync(metricsFile, JSON.stringify(metrics, null, 2));
      } catch (error) {
        console.error('Failed to record test result:', error);
      }
    };

    await use(recordTestResult);
  },

  generateTestReport: async ({}, use) => {
    const generateTestReport = () => {
      try {
        if (!fs.existsSync(metricsFile)) {
          console.log('No metrics file found');
          return;
        }

        const metrics: TestMetric[] = JSON.parse(fs.readFileSync(metricsFile, 'utf-8'));

        // Calculate statistics
        const totalTests = metrics.length;
        const passedTests = metrics.filter((m) => m.passed).length;
        const failedTests = totalTests - passedTests;
        const passRate = ((passedTests / totalTests) * 100).toFixed(1);
        const avgDuration = (metrics.reduce((sum, m) => sum + m.duration, 0) / totalTests).toFixed(0);

        // Group by browser
        const byBrowser = metrics.reduce(
          (acc, m) => {
            if (!acc[m.browser]) {
              acc[m.browser] = [];
            }
            acc[m.browser].push(m);
            return acc;
          },
          {} as Record<string, TestMetric[]>
        );

        // Generate report
        const report = {
          generatedAt: new Date().toISOString(),
          summary: {
            totalTests,
            passedTests,
            failedTests,
            passRate: `${passRate}%`,
            averageDuration: `${avgDuration}ms`,
          },
          byBrowser: Object.entries(byBrowser).map(([browser, tests]) => ({
            browser,
            count: tests.length,
            passed: tests.filter((t) => t.passed).length,
            passRate: ((tests.filter((t) => t.passed).length / tests.length) * 100).toFixed(1),
          })),
          timeline: metrics
            .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
            .slice(-20),
        };

        // Save report
        const reportFile = path.join(process.cwd(), 'test-results', 'test-report.json');
        fs.writeFileSync(reportFile, JSON.stringify(report, null, 2));

        console.log('\n=== Test Report Summary ===');
        console.log(`Total Tests: ${totalTests}`);
        console.log(`Passed: ${passedTests}`);
        console.log(`Failed: ${failedTests}`);
        console.log(`Pass Rate: ${passRate}%`);
        console.log(`Average Duration: ${avgDuration}ms`);
        console.log('\n=== By Browser ===');
        report.byBrowser.forEach((browser) => {
          console.log(`${browser.browser}: ${browser.passed}/${browser.count} (${browser.passRate}%)`);
        });
      } catch (error) {
        console.error('Failed to generate test report:', error);
      }
    };

    await use(generateTestReport);
  },
});

/**
 * Helper to get historical test data
 */
export function getTestMetrics() {
  try {
    if (fs.existsSync(metricsFile)) {
      return JSON.parse(fs.readFileSync(metricsFile, 'utf-8'));
    }
  } catch {
    return [];
  }
  return [];
}

/**
 * Helper to calculate test stability
 */
export function calculateStability(testName: string): number {
  const metrics = getTestMetrics();
  const testResults = metrics.filter((m: TestMetric) => m.testName === testName);

  if (testResults.length === 0) return 0;

  const passCount = testResults.filter((m: TestMetric) => m.passed).length;
  return (passCount / testResults.length) * 100;
}

/**
 * Helper to get trending data
 */
export function getTrendingData(days: number = 7) {
  const metrics = getTestMetrics();
  const cutoffDate = new Date();
  cutoffDate.setDate(cutoffDate.getDate() - days);

  const recentMetrics = metrics.filter((m: TestMetric) => {
    return new Date(m.timestamp) >= cutoffDate;
  });

  const dailyStats = recentMetrics.reduce(
    (acc, m: TestMetric) => {
      const date = new Date(m.timestamp).toISOString().split('T')[0];
      if (!acc[date]) {
        acc[date] = { passed: 0, failed: 0, total: 0 };
      }
      acc[date].total++;
      if (m.passed) {
        acc[date].passed++;
      } else {
        acc[date].failed++;
      }
      return acc;
    },
    {} as Record<string, { passed: number; failed: number; total: number }>
  );

  return Object.entries(dailyStats).map(([date, stats]) => ({
    date,
    passRate: ((stats.passed / stats.total) * 100).toFixed(1),
    ...stats,
  }));
}

export { expect } from '@playwright/test';
