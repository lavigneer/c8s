import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for E2E testing
 * Includes multi-browser support, viewport variations, and accessibility testing
 */
const config = defineConfig({
  testDir: './tests/e2e/specs',
  testMatch: '**/*.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['list']
  ],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: process.env.ALL_BROWSERS === 'true' ? [
    // Desktop browsers
    {
      name: 'chromium-desktop',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1920, height: 1080 },
      },
    },
    {
      name: 'firefox-desktop',
      use: {
        ...devices['Desktop Firefox'],
        viewport: { width: 1920, height: 1080 },
      },
    },
    {
      name: 'webkit-desktop',
      use: {
        ...devices['Desktop Safari'],
        viewport: { width: 1920, height: 1080 },
      },
    },
    {
      name: 'msedge-desktop',
      use: {
        ...devices['Desktop Edge'],
        viewport: { width: 1920, height: 1080 },
      },
    },

    // Tablet viewports
    {
      name: 'chromium-tablet',
      use: {
        ...devices['iPad Pro'],
        viewport: { width: 1024, height: 1366 },
      },
    },
    {
      name: 'firefox-tablet',
      use: {
        ...devices['Firefox Tablet'],
        viewport: { width: 1024, height: 1366 },
      },
    },

    // Mobile viewports
    {
      name: 'chromium-mobile',
      use: {
        ...devices['iPhone 12'],
        viewport: { width: 390, height: 844 },
      },
    },
    {
      name: 'firefox-mobile',
      use: {
        ...devices['Firefox Mobile'],
        viewport: { width: 390, height: 844 },
      },
    },
  ] : [
    // Development: Run only chromium-desktop by default
    {
      name: 'chromium-desktop',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1920, height: 1080 },
      },
    },
  ],

  webServer: {
    command: 'echo "Ensure Tilt is running: tilt up"',
    port: 8080,
    reuseExistingServer: true,
    timeout: 5000,
  },
});

export default config;
