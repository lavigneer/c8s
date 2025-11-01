import { test as base, APIRequestContext } from '@playwright/test';

/**
 * Test data fixtures for e2e testing
 * Provides API request helper and test data management
 */

type TestDataFixture = {
  apiRequest: APIRequestContext;
  testToken: string;
  baseUrl: string;
  apiUrl: string;
};

export const test = base.extend<TestDataFixture>({
  testToken: async ({}, use) => {
    // Inject test authentication token
    const token = process.env.TEST_AUTH_TOKEN || 'test_token_' + Date.now();
    await use(token);
  },

  baseUrl: async ({}, use) => {
    const url = process.env.BASE_URL || 'http://localhost:8080';
    await use(url);
  },

  apiUrl: async ({}, use) => {
    const url = process.env.API_URL || 'http://localhost:8080/api';
    await use(url);
  },

  apiRequest: async ({ page }, use) => {
    // Create API request context using page's context to share cookies
    // This ensures auth cookies from login are included in API requests
    const request = page.context().request;
    await use(request);
  },
});

/**
 * Helper to create test pipeline via API
 */
export async function createTestPipeline(
  request: APIRequestContext,
  name: string = `test-pipeline-${Date.now()}`
) {
  const response = await request.post('/test/pipelines', {
    data: {
      name,
      repository: 'github.com/test/repo',
      branches: ['main', 'develop'],
      timeout: 3600,
    },
  });

  return await response.json();
}

/**
 * Helper to delete test pipeline via API
 */
export async function deleteTestPipeline(request: APIRequestContext, pipelineId: string) {
  return await request.delete(`/test/pipelines/${pipelineId}`);
}

/**
 * Helper to create test project via API
 */
export async function createTestProject(
  request: APIRequestContext,
  name: string = `test-project-${Date.now()}`
) {
  const response = await request.post('/test/projects', {
    data: {
      name,
      description: 'Test project for e2e testing',
    },
  });

  return await response.json();
}

/**
 * Helper to delete test project via API
 */
export async function deleteTestProject(request: APIRequestContext, projectId: string) {
  return await request.delete(`/test/projects/${projectId}`);
}

/**
 * Helper to cleanup all test data
 */
export async function cleanupTestData(request: APIRequestContext, testRunId: string) {
  return await request.post('/test/cleanup', {
    data: {
      testRunId,
    },
  });
}

export { expect } from '@playwright/test';
