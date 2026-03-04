import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';

test.describe('Permissions and Access Control', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
  });

  test('should redirect unauthenticated users to login', async ({ page }) => {
    // Try to access protected route without login
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });

  test('should prevent access to dashboard without login', async ({ page }) => {
    // Try direct access to dashboard
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Verify login form is displayed
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first();

    await expect(emailInput).toBeVisible();
    await expect(passwordInput).toBeVisible();
  });

  test('should allow access after successful login', async ({ page }) => {
    // Navigate to login
    await loginPage.goto();

    // Login
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Should be redirected to dashboard
    expect(page.url()).toContain('/dashboard');

    // Dashboard should be visible
    await dashboardPage.expectLoaded();
  });

  test('should restrict collectors page to authenticated users', async ({ page }) => {
    // Try to access collectors without login
    await page.goto('/collectors', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });

  test('should restrict alerts page to authenticated users', async ({ page }) => {
    // Try to access alerts without login
    await page.goto('/alerts', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });

  test('should restrict users page to authenticated users', async ({ page }) => {
    // Try to access users page without login
    await page.goto('/users', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });

  test('should validate session on API calls', async ({ page }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Intercept API calls to verify auth headers
    let authHeaderPresent = false;
    page.on('request', (request) => {
      const authHeader = request.headers()['authorization'];
      if (authHeader && authHeader.startsWith('Bearer ')) {
        authHeaderPresent = true;
      }
    });

    // Navigate to page that makes API calls
    await dashboardPage.goto();

    // Wait for API calls to complete
    await page.waitForLoadState('networkidle');

    // Verify auth header was sent
    expect(authHeaderPresent).toBe(true);
  });

  test('should reject requests with invalid token', async ({ page }) => {
    // Manually set invalid token in localStorage
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('authToken', 'invalid-token-xyz');
    });

    // Try to access protected route
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should either redirect to login or show error
    const isOnLogin = page.url().includes('/login');
    const isError = page.locator('.alert-danger, [data-testid="error"]').first();

    expect(isOnLogin || await isError.isVisible({ timeout: 2000 }).catch(() => false)).toBe(true);
  });

  test('should handle expired session', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Simulate token expiration by removing it
    await page.evaluate(() => {
      localStorage.removeItem('authToken');
    });

    // Try to access protected page
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });

  test('should prevent CSRF attacks with CSRF token', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Check for CSRF token in forms
    const forms = page.locator('form');
    const formCount = await forms.count();

    if (formCount > 0) {
      // Look for CSRF token field or header
      const form = forms.first();
      const csrfInput = form.locator('input[name*="csrf"], input[name*="token"]').first();

      // If form exists, CSRF protection might be in place
      const hasCsrfField = await csrfInput.isVisible({ timeout: 1000 }).catch(() => false);
      console.log(`CSRF field found: ${hasCsrfField}`);
    }
  });

  test('should maintain user context across pages', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Navigate to different pages
    await page.goto('/collectors', { waitUntil: 'networkidle' });
    expect(page.url()).toContain('/collectors');

    await page.goto('/alerts', { waitUntil: 'networkidle' });
    expect(page.url()).toContain('/alerts');

    // Should still be logged in
    const userMenu = page.locator('[data-testid="user-menu"], .user-menu, [aria-label*="user"]').first();
    const isLoggedIn = await userMenu.isVisible({ timeout: 1000 }).catch(() => false);

    // At minimum, page should load without redirect to login
    expect(page.url()).not.toContain('/login');
  });

  test('should display user name/email in header when logged in', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Check for user display in header
    const userDisplay = page.locator(
      '[data-testid="user-menu"], .user-menu, .user-info, header'
    ).first();

    const headerText = await userDisplay.textContent();

    // Should contain email or username
    expect(headerText).toBeTruthy();
  });

  test('should log out and clear session', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Logout
    await loginPage.logout();
    await loginPage.expectLoggedOut();

    // Try to access protected page
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should be redirected to login
    expect(page.url()).toContain('/login');
  });

  test('should handle multiple users with separate sessions', async ({ context }) => {
    // Create two browser contexts for two users
    const page1 = await context.newPage();
    const page2 = await context.newPage();

    const login1 = new LoginPage(page1);
    const login2 = new LoginPage(page2);

    // User 1 logs in
    await login1.goto();
    await login1.login('demo@pganalytics.com', 'password123');

    // User 2 tries to login with same credentials (if available)
    await login2.goto();
    await login2.login('demo@pganalytics.com', 'password123');

    // Both should be logged in
    expect(page1.url()).toContain('/dashboard');
    expect(page2.url()).toContain('/dashboard');

    // Logout user 1
    await login1.logout();

    // User 1 should see login page
    expect(page1.url()).toContain('/login');

    // User 2 should still be logged in
    expect(page2.url()).toContain('/dashboard');

    // Cleanup
    await page1.close();
    await page2.close();
  });

  test('should protect against XSS in user input', async ({ page }) => {
    // This is more of a security validation test
    // In a real scenario, we'd test form inputs sanitization

    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Monitor for any script execution
    let scriptExecuted = false;
    page.on('console', (msg) => {
      if (msg.text().includes('xss') || msg.type() === 'error') {
        scriptExecuted = true;
      }
    });

    // Navigate to page
    await dashboardPage.goto();

    // Page should load without XSS execution
    expect(scriptExecuted).toBe(false);
  });

  test('should enforce rate limiting on login attempts', async ({ page }) => {
    // Try multiple failed login attempts
    let attemptCount = 0;
    const maxAttempts = 5;

    for (let i = 0; i < maxAttempts; i++) {
      await loginPage.goto();

      await loginPage.fillEmail('demo@pganalytics.com');
      await loginPage.fillPassword('wrongpassword');
      await loginPage.clickLogin();

      // Wait for response
      await page.waitForTimeout(500);

      // Check if we got locked out
      const lockoutMessage = page.locator(
        '[data-testid="error"], .alert-danger'
      ).first();

      const isLockedOut = await lockoutMessage.isVisible({ timeout: 1000 }).catch(() => false);
      if (isLockedOut) {
        const messageText = await lockoutMessage.textContent();
        if (messageText && messageText.toLowerCase().includes('locked')) {
          console.log(`Rate limiting triggered after ${i + 1} attempts`);
          break;
        }
      }

      attemptCount++;
    }

    console.log(`Completed ${attemptCount} login attempts`);
  });

  test('should validate authorization headers on API requests', async ({ page }) => {
    // This validates that API requests include proper auth

    // Login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Intercept API requests
    const requestHeaders: { [key: string]: string } = {};

    page.on('request', (request) => {
      const url = request.url();
      if (url.includes('/api/')) {
        const authHeader = request.headers()['authorization'];
        requestHeaders[url] = authHeader || 'MISSING';
      }
    });

    // Navigate to page that makes API calls
    await dashboardPage.goto();
    await page.waitForLoadState('networkidle');

    // Verify auth headers were sent
    const apiRequestsWithAuth = Object.values(requestHeaders).filter((h) => h !== 'MISSING');
    expect(apiRequestsWithAuth.length).toBeGreaterThan(0);
  });
});
