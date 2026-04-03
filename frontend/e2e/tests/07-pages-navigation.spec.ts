import { test, expect } from '../fixtures/auth';

/**
 * E2E Tests: Page Navigation and Sidebar Visibility
 * Tests that all pages load correctly and sidebar appears properly
 */

test.describe('Pages Navigation & Sidebar', () => {
  test('should render sidebar on all protected pages', async ({ page }) => {
    const pages = [
      { name: 'Home', path: '/' },
      { name: 'Logs', path: '/logs' },
      { name: 'Metrics', path: '/metrics' },
      { name: 'Alerts', path: '/alerts' },
      { name: 'Channels', path: '/channels' },
      { name: 'Collectors', path: '/collectors' },
      { name: 'Settings', path: '/settings' },
      { name: 'Users', path: '/users' },
    ];

    for (const { name, path } of pages) {
      await page.goto(path);

      // Check sidebar is visible
      const sidebar = page.locator('aside, nav[class*="sidebar"]').first();
      const mainContent = page.locator('main');

      await expect(sidebar).toBeVisible({ timeout: 5000 });
      await expect(mainContent).toBeVisible({ timeout: 5000 });

      console.log(`✅ ${name} (${path}): Sidebar and main content visible`);
    }
  });

  test('should navigate between pages using sidebar', async ({ page }) => {
    // Start at home
    await page.goto('/');

    // Click Logs link in sidebar
    await page.click('a[href="/logs"], button:has-text("Logs")');
    await expect(page).toHaveURL(/\/logs/);

    // Click Metrics link
    await page.click('a[href="/metrics"], button:has-text("Metrics")');
    await expect(page).toHaveURL(/\/metrics/);

    // Click Alerts link
    await page.click('a[href="/alerts"], button:has-text("Alerts")');
    await expect(page).toHaveURL(/\/alerts/);

    console.log('✅ Navigation between pages works correctly');
  });

  test('should open/close user menu in header', async ({ page }) => {
    await page.goto('/');

    // Find user menu button (has image)
    const userMenuBtn = page.locator('button').filter({ has: page.locator('img') }).first();

    // Click to open
    await userMenuBtn.click();

    // Check Settings option appears
    const settingsBtn = page.locator('button:has-text("Settings")');
    await expect(settingsBtn).toBeVisible();

    // Check Logout option appears
    const logoutBtn = page.locator('button:has-text("Logout")');
    await expect(logoutBtn).toBeVisible();

    console.log('✅ User menu opens and displays Settings and Logout options');
  });

  test('should navigate to Settings from header menu', async ({ page }) => {
    await page.goto('/');

    // Open user menu
    const userMenuBtn = page.locator('button').filter({ has: page.locator('img') }).first();
    await userMenuBtn.click();

    // Click Settings
    const settingsBtn = page.locator('button:has-text("Settings")');
    await settingsBtn.click();

    // Verify navigation to settings page
    await expect(page).toHaveURL(/\/settings/);
    await expect(page.locator('h1')).toContainText(/settings|users|admin/i);

    console.log('✅ Settings button navigates to /settings page');
  });

  test('should logout from user menu', async ({ page }) => {
    await page.goto('/');

    // Open user menu
    const userMenuBtn = page.locator('button').filter({ has: page.locator('img') }).first();
    await userMenuBtn.click();

    // Click Logout
    const logoutBtn = page.locator('button:has-text("Logout")');
    await logoutBtn.click();

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/);

    console.log('✅ Logout redirects to login page');
  });
});
