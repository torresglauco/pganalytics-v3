import { test as base, Page } from '@playwright/test';

/**
 * Default test credentials for pgAnalytics
 * These credentials MUST match actual database state
 */
export const TEST_CREDENTIALS = {
  admin: {
    username: 'admin',
    password: 'admin',
    email: 'admin@pganalytics.local',
    role: 'admin' as const,
  },
  user: {
    username: `testuser-${Date.now()}`,
    password: 'TestPassword123!',
    email: `testuser-${Date.now()}@example.com`,
    role: 'user' as const,
  },
  viewer: {
    username: `testviewer-${Date.now()}`,
    password: 'ViewerPass123!',
    email: `testviewer-${Date.now()}@example.com`,
    role: 'viewer' as const,
  },
} as const;

/**
 * API endpoints
 */
export const API_ENDPOINTS = {
  auth: {
    login: '/api/v1/auth/login',
    logout: '/api/v1/auth/logout',
    me: '/api/v1/auth/me',
    refresh: '/api/v1/auth/refresh',
  },
  users: {
    list: '/api/v1/users',
    create: '/api/v1/users',
    get: (id: string | number) => `/api/v1/users/${id}`,
    update: (id: string | number) => `/api/v1/users/${id}`,
    delete: (id: string | number) => `/api/v1/users/${id}`,
    resetPassword: (id: string | number) => `/api/v1/users/${id}/reset-password`,
  },
} as const;

/**
 * Authenticated page fixture
 * Automatically logs in before each test using admin credentials
 */
type AuthenticatedPage = {
  page: Page;
  authToken: string;
};

export const test = base.extend<AuthenticatedPage>({
  page: async ({ page }, use) => {
    // Navigate to app
    await page.goto('/');
    await use(page);
  },

  authToken: async ({ page }, use) => {
    // Login via API to get token
    const response = await page.request.post(API_ENDPOINTS.auth.login, {
      data: {
        username: TEST_CREDENTIALS.admin.username,
        password: TEST_CREDENTIALS.admin.password,
      },
    });

    if (!response.ok()) {
      throw new Error(`Failed to authenticate: ${response.status()}`);
    }

    const data = await response.json();
    const token = data.token;

    if (!token) {
      throw new Error('No auth token received from API');
    }

    await use(token);
  },
});

export { expect } from '@playwright/test';
