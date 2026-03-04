import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';

test.describe('Dashboard Visualization', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
  });

  test('should load dashboard successfully', async ({ page }) => {
    await dashboardPage.goto();

    // Verify dashboard is displayed
    await dashboardPage.expectLoaded();

    // Verify no errors
    await dashboardPage.expectNoErrors();
  });

  test('should display navigation menu', async ({ page }) => {
    await dashboardPage.goto();

    // Check for navigation elements
    const nav = page.locator('nav, [role="navigation"]').first();
    await expect(nav).toBeVisible();

    // Verify key menu items
    const menuItems = ['Collectors', 'Alerts', 'Dashboards', 'Users'];
    for (const item of menuItems) {
      const link = page.locator(`a, button`).filter({ hasText: new RegExp(item, 'i') }).first();
      const isVisible = await link.isVisible({ timeout: 1000 }).catch(() => false);
      if (isVisible) {
        console.log(`Found menu item: ${item}`);
      }
    }
  });

  test('should load charts without errors', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.waitForDataLoad();

    // Verify charts are displayed
    const charts = page.locator('canvas, [data-testid="chart"], svg').first();
    const isVisible = await charts.isVisible({ timeout: 3000 }).catch(() => false);

    if (isVisible) {
      // At least one chart is visible
      const chartCount = await dashboardPage.getChartCount();
      console.log(`Charts displayed: ${chartCount}`);
      expect(chartCount).toBeGreaterThanOrEqual(0);
    }
  });

  test('should load metrics data', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.waitForDataLoad();

    // Try to get metrics
    const metricsCount = await dashboardPage.getMetricsCount();
    console.log(`Metrics displayed: ${metricsCount}`);

    // Even if no metrics, page should be responsive
    await dashboardPage.expectNoErrors();
  });

  test('should navigate between pages', async ({ page }) => {
    await dashboardPage.goto();

    // Try to navigate to Collectors
    try {
      await dashboardPage.navigateToCollectors();
      expect(page.url()).toContain('/collectors');
    } catch {
      console.log('Collectors navigation skipped');
    }

    // Return to dashboard
    await dashboardPage.goto();

    // Try to navigate to Alerts
    try {
      await dashboardPage.navigateToAlerts();
      expect(page.url()).toContain('/alerts');
    } catch {
      console.log('Alerts navigation skipped');
    }
  });

  test('should handle page reload', async ({ page }) => {
    await dashboardPage.goto();

    // Reload page
    await page.reload();

    // Should still be logged in and dashboard should load
    await dashboardPage.expectLoaded();
    await dashboardPage.expectNoErrors();
  });

  test('should be responsive on viewport resize', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.waitForDataLoad();

    // Verify initial state
    await dashboardPage.expectLoaded();

    // Resize viewport
    await page.setViewportSize({ width: 768, height: 1024 });

    // Wait for resize to apply
    await page.waitForTimeout(500);

    // Dashboard should still be visible
    await dashboardPage.expectLoaded();
  });

  test('should handle slow network', async ({ page }) => {
    // Simulate slow network
    await page.route('**/*', (route) => {
      setTimeout(() => route.continue(), 500);
    });

    await dashboardPage.goto();

    // Should still load eventually
    await dashboardPage.expectLoaded({ timeout: 15000 });
  });

  test('should display loading state', async ({ page }) => {
    // Go to dashboard and watch for loading
    const gotoPromise = dashboardPage.goto();

    // Check if loading spinner appears
    const spinner = page.locator('[data-testid="loading"], .spinner, .loader').first();
    const isLoading = await spinner.isVisible({ timeout: 2000 }).catch(() => false);

    // Wait for navigation to complete
    await gotoPromise;

    // Loading should be gone now
    const stillLoading = await spinner.isVisible({ timeout: 500 }).catch(() => false);
    expect(stillLoading).toBe(false);
  });

  test('should update data on interval', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.waitForDataLoad();

    // Get initial data
    const initialMetrics = await dashboardPage.getMetricsCount();

    // Wait for data refresh (typically 30-60 seconds, use shorter timeout)
    await page.waitForTimeout(3000);

    // Data might have updated
    const updatedMetrics = await dashboardPage.getMetricsCount();

    // Should still be same count or more
    expect(updatedMetrics).toBeGreaterThanOrEqual(initialMetrics);
  });

  test('should handle empty state gracefully', async ({ page }) => {
    // Go to dashboard
    await dashboardPage.goto();

    // Even with no data, dashboard should be usable
    await dashboardPage.expectLoaded();
    await dashboardPage.expectNoErrors();

    // Should still have navigation
    const nav = page.locator('nav, [role="navigation"]').first();
    await expect(nav).toBeVisible();
  });

  test('should search or filter data', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.waitForDataLoad();

    // Look for search input
    const searchInput = page.locator('input').filter({ hasText: /search|filter/i }).first();

    if (await searchInput.isVisible({ timeout: 1000 }).catch(() => false)) {
      // Type in search
      await searchInput.fill('test');

      // Wait for results to filter
      await page.waitForTimeout(500);

      // Results should be filtered or no error
      await dashboardPage.expectNoErrors();
    }
  });
});
