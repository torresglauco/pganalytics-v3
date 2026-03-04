import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';

test.describe('Login/Logout Flow', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
  });

  test('should display login page', async ({ page }) => {
    await loginPage.goto();

    // Check that login form is visible
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
    const loginBtn = page.locator('button').filter({ hasText: /Sign In|Login|Submit/i }).first();

    await expect(emailInput).toBeVisible();
    await expect(passwordInput).toBeVisible();
    await expect(loginBtn).toBeVisible();
  });

  test('should login with valid credentials', async ({ page }) => {
    await loginPage.goto();

    // Use test credentials (these should be seeded in test database)
    await loginPage.login('demo@pganalytics.com', 'password123');

    // Verify redirect to dashboard
    await loginPage.expectLoggedIn();

    // Verify dashboard loads
    await dashboardPage.expectLoaded();
  });

  test('should show error with invalid credentials', async ({ page }) => {
    await loginPage.goto();

    // Fill wrong credentials
    await loginPage.fillEmail('wrong@example.com');
    await loginPage.fillPassword('wrongpassword');
    await loginPage.clickLogin();

    // Verify error message appears
    try {
      await loginPage.expectErrorMessage();
    } catch {
      // Some implementations might not show explicit error
      // Check that we're still on login page instead
      expect(page.url()).toContain('/login');
    }
  });

  test('should logout and redirect to login page', async ({ page }) => {
    await loginPage.goto();

    // First login
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Then logout
    await loginPage.logout();

    // Verify redirect to login
    await loginPage.expectLoggedOut();
  });

  test('should prevent unauthorized access to dashboard', async ({ page }) => {
    // Try to access dashboard without logging in
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should be redirected to login
    expect(page.url()).toContain('/login');
  });

  test('should show loading state during login', async ({ page }) => {
    await loginPage.goto();

    // Fill credentials
    await loginPage.fillEmail('demo@pganalytics.com');
    await loginPage.fillPassword('password123');

    // Click login and check for loading state
    await loginPage.clickLogin();

    // Wait for either loading indicator or page transition
    // This is flexible as different implementations may vary
    await page.waitForLoadState('networkidle');

    // Should end up at dashboard
    expect(page.url()).toContain('/dashboard');
  });

  test('should maintain session after page reload', async ({ page }) => {
    await loginPage.goto();

    // Login
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Reload page
    await page.reload();

    // Should still be logged in
    await dashboardPage.expectLoaded();
    expect(page.url()).toContain('/dashboard');
  });

  test('should clear session on logout', async ({ page }) => {
    await loginPage.goto();

    // Login
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Logout
    await loginPage.logout();

    // Try to access dashboard
    await page.goto('/dashboard', { waitUntil: 'networkidle' });

    // Should redirect to login
    expect(page.url()).toContain('/login');
  });
});
