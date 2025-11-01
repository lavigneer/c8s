/**
 * Test constants and configuration
 */

// Test timeouts
export const TIMEOUTS = {
  short: 5000, // 5 seconds
  medium: 10000, // 10 seconds
  long: 30000, // 30 seconds
  veryLong: 60000, // 60 seconds
};

// Test user credentials (for demo/testing only)
export const TEST_USERS = {
  admin: {
    email: 'admin@test.example.com',
    password: 'admin-password-123',
  },
  developer: {
    email: 'dev@test.example.com',
    password: 'dev-password-456',
  },
};

// Test data defaults
export const TEST_DATA = {
  pipeline: {
    name: `test-pipeline-${Date.now()}`,
    repository: 'github.com/test/repo',
    timeout: 3600,
  },
  project: {
    name: `test-project-${Date.now()}`,
    description: 'Test project for e2e tests',
  },
};

// Viewport sizes
export const VIEWPORTS = {
  desktop: { width: 1920, height: 1080 },
  tablet: { width: 1024, height: 1366 },
  mobile: { width: 390, height: 844 },
};

// Browser list for cross-browser testing
export const BROWSERS = ['chromium', 'firefox', 'webkit', 'msedge'];

// Retry configuration
export const RETRY_CONFIG = {
  maxAttempts: 3,
  delayMs: 1000,
  backoffMultiplier: 2,
};

// Test environment URLs
export const URLS = {
  base: process.env.BASE_URL || 'http://localhost:8080',
  api: process.env.API_URL || 'http://localhost:8080/api',
  login: '/login',
  dashboard: '/dashboard',
  projects: '/dashboard/projects',
  settings: '/dashboard/settings',
};

// Common selectors (fallback patterns)
export const SELECTORS = {
  // Generic
  button: 'button',
  input: 'input',
  link: 'a',
  heading: 'h1, h2, h3',

  // Common patterns
  modal: '[role="dialog"]',
  alert: '[role="alert"]',
  spinner: '[data-testid="spinner"], .spinner, .loading',
  errorMessage: '[role="alert"][aria-live="polite"]',
};

// WCAG compliance levels
export const WCAG_LEVELS = {
  A: 'wcag2a',
  AA: 'wcag2aa',
  AAA: 'wcag2aaa',
};

// Performance thresholds (milliseconds)
export const PERFORMANCE_THRESHOLDS = {
  pageLoad: 3000,
  interaction: 500,
  apiResponse: 2000,
};

// Accessibility checklist
export const ACCESSIBILITY_CHECKS = {
  keyboard: true,
  screenReader: true,
  colorContrast: true,
  focusManagement: true,
  wcagLevel: 'wcag2aa',
};

export default {
  TIMEOUTS,
  TEST_USERS,
  TEST_DATA,
  VIEWPORTS,
  BROWSERS,
  RETRY_CONFIG,
  URLS,
  SELECTORS,
  WCAG_LEVELS,
  PERFORMANCE_THRESHOLDS,
  ACCESSIBILITY_CHECKS,
};
