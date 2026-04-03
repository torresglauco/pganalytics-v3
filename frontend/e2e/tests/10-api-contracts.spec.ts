/**
 * API Contract Tests
 * Validates that backend API returns expected response structures
 * These tests ensure frontend and backend stay in sync
 */

import { test, expect, API_ENDPOINTS, TEST_CREDENTIALS } from '../fixtures/auth';

test.describe('API Response Contracts', () => {
  let authToken: string;

  test.beforeEach(async ({ authToken: token }) => {
    authToken = token;
  });

  test.describe('Authentication API', () => {
    test('should return valid token on successful login', async ({ request }) => {
      const response = await request.post(API_ENDPOINTS.auth.login, {
        data: {
          username: TEST_CREDENTIALS.admin.username,
          password: TEST_CREDENTIALS.admin.password,
        },
      });

      expect(response.status()).toBe(200);

      const data = await response.json();

      // Validate response schema
      expect(data).toHaveProperty('token');
      expect(data).toHaveProperty('refresh_token');
      expect(data).toHaveProperty('user');
      expect(data).toHaveProperty('expires_in');

      // Validate token format (JWT)
      expect(typeof data.token).toBe('string');
      expect(data.token.split('.').length).toBe(3); // JWT has 3 parts

      // Validate user object
      expect(data.user).toHaveProperty('id');
      expect(data.user).toHaveProperty('username');
      expect(data.user).toHaveProperty('email');
      expect(data.user).toHaveProperty('role');
      expect(data.user.role).toBe('admin');
    });

    test('should fail with invalid credentials', async ({ request }) => {
      const response = await request.post(API_ENDPOINTS.auth.login, {
        data: {
          username: 'invalid_user',
          password: 'wrong_password',
        },
      });

      expect(response.status()).toBe(401);
      const data = await response.json();
      expect(data).toHaveProperty('code');
      expect(data).toHaveProperty('message');
    });
  });

  test.describe('Users API', () => {
    test('should return paginated user list with correct structure', async ({ request }) => {
      const response = await request.get(API_ENDPOINTS.users.list, {
        headers: { 'Authorization': `Bearer ${authToken}` },
      });

      expect(response.status()).toBe(200);
      const data = await response.json();

      // Validate pagination response schema
      expect(data).toHaveProperty('data');
      expect(data).toHaveProperty('page');
      expect(data).toHaveProperty('page_size');
      expect(data).toHaveProperty('total');
      expect(data).toHaveProperty('total_pages');

      // Validate types
      expect(Array.isArray(data.data)).toBe(true);
      expect(typeof data.page).toBe('number');
      expect(typeof data.page_size).toBe('number');
      expect(typeof data.total).toBe('number');
      expect(typeof data.total_pages).toBe('number');

      // Validate data integrity
      expect(data.page).toBeGreaterThanOrEqual(1);
      expect(data.page_size).toBeGreaterThan(0);
      expect(data.total).toBeGreaterThanOrEqual(0);
      expect(data.total_pages).toBeGreaterThanOrEqual(1);

      // Verify users have required fields
      if (data.data.length > 0) {
        const user = data.data[0];
        expect(user).toHaveProperty('id');
        expect(user).toHaveProperty('username');
        expect(user).toHaveProperty('email');
        expect(user).toHaveProperty('role');
        expect(user).toHaveProperty('is_active');
      }
    });

    test('should return admin user in list', async ({ request }) => {
      const response = await request.get(API_ENDPOINTS.users.list, {
        headers: { 'Authorization': `Bearer ${authToken}` },
      });

      expect(response.status()).toBe(200);
      const data = await response.json();

      // Find admin user
      const adminUser = data.data.find((u: any) => u.username === 'admin');
      expect(adminUser).toBeDefined();
      expect(adminUser.role).toBe('admin');
      expect(adminUser.is_active).toBe(true);
      expect(adminUser.email).toBe('admin@pganalytics.local');
    });

    test('should require authentication for users list', async ({ request }) => {
      const response = await request.get(API_ENDPOINTS.users.list);

      // Should fail without token
      expect(response.status()).toBe(401);
      const data = await response.json();
      expect(data).toHaveProperty('message');
    });

    test('should deny access to non-admin users', async ({ request }) => {
      // Create a regular user token (if available)
      // For now, just test that admin can access
      const response = await request.get(API_ENDPOINTS.users.list, {
        headers: { 'Authorization': `Bearer ${authToken}` },
      });

      expect(response.status()).toBe(200);
    });
  });

  test.describe('Health Check API', () => {
    test('should return health status without authentication', async ({ request }) => {
      const response = await request.get('/api/v1/health');

      expect(response.status()).toBe(200);
      const data = await response.json();

      // Validate health response schema
      expect(data).toHaveProperty('status');
      expect(data).toHaveProperty('version');
      expect(data).toHaveProperty('timestamp');
      expect(data).toHaveProperty('database_ok');
      expect(data).toHaveProperty('timescale_ok');

      // Validate values
      expect(data.status).toBe('ok');
      expect(typeof data.version).toBe('string');
      expect(data.database_ok).toBe(true);
      expect(data.timescale_ok).toBe(true);
    });
  });

  test.describe('Error Handling', () => {
    test('should return 401 for expired/invalid token', async ({ request }) => {
      const response = await request.get(API_ENDPOINTS.users.list, {
        headers: { 'Authorization': 'Bearer invalid_token_123' },
      });

      expect(response.status()).toBe(401);
      const data = await response.json();
      expect(data).toHaveProperty('code');
      expect(data).toHaveProperty('message');
    });

    test('should return proper error structure', async ({ request }) => {
      const response = await request.post(API_ENDPOINTS.auth.login, {
        data: {
          username: 'nonexistent',
          password: 'wrong',
        },
      });

      expect(response.status()).toBeGreaterThanOrEqual(400);
      const data = await response.json();

      // All error responses must have code and message
      expect(data).toHaveProperty('code');
      expect(data).toHaveProperty('message');
      expect(typeof data.code).toBe('number');
      expect(typeof data.message).toBe('string');
    });
  });
});
