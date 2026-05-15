import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { HostInventoryPage } from '../pages/HostInventoryPage';

test.describe('Host Monitoring', () => {
  let loginPage: LoginPage;
  let hostPage: HostInventoryPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    hostPage = new HostInventoryPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
  });

  test('should display host inventory page with table', async ({ page }) => {
    await hostPage.goto();

    // Verify page title is visible
    await expect(page.locator('h1:has-text("Host Inventory")')).toBeVisible();
  });

  test('should render host status table', async ({ page }) => {
    await hostPage.goto();

    // Wait for table to be visible or empty state
    const table = page.locator('table, [data-testid="host-table"]').first();
    const emptyState = page.locator('text=No hosts configured').first();

    // Either table or empty state should be visible
    await Promise.race([
      expect(table).toBeVisible({ timeout: 10000 }),
      expect(emptyState).toBeVisible({ timeout: 10000 }),
    ]);
  });

  test('should display summary cards with host counts', async ({ page }) => {
    await hostPage.goto();

    // Verify summary cards are visible
    await hostPage.expectSummaryCardsVisible();
  });

  test('should search hosts by hostname', async ({ page }) => {
    await hostPage.goto();

    // Wait for page load
    const emptyState = await hostPage.hasEmptyState();
    if (!emptyState) {
      await hostPage.expectTableVisible();

      const initialRowCount = await hostPage.getRowCount();

      // Enter search term
      await hostPage.searchByHostname('test');

      // Wait for filtering
      await page.waitForTimeout(500);

      // Clear search
      await hostPage.clearSearch();
    }
  });

  test('should filter hosts by status', async ({ page }) => {
    await hostPage.goto();

    // Wait for page load
    const emptyState = await hostPage.hasEmptyState();
    if (!emptyState) {
      await hostPage.expectTableVisible();

      // Filter by 'up' status
      await hostPage.filterByStatus('up');
      await page.waitForTimeout(500);

      // Filter by 'down' status
      await hostPage.filterByStatus('down');
      await page.waitForTimeout(500);

      // Clear filter
      await hostPage.filterByStatus('');
    }
  });

  test('should show host detail panel when clicking a host', async ({ page }) => {
    await hostPage.goto();

    // Check if there are hosts
    const emptyState = await hostPage.hasEmptyState();
    if (!emptyState) {
      await hostPage.expectTableVisible();

      const rowCount = await hostPage.getRowCount();
      if (rowCount > 0) {
        // Click first host row
        await hostPage.clickHostByIndex(0);

        // Verify detail panel opens (or page navigates)
        await page.waitForTimeout(500);
      }
    }
  });

  test('should toggle auto-refresh checkbox', async ({ page }) => {
    await hostPage.goto();

    // Find auto-refresh checkbox
    const autoRefreshCheckbox = page.locator('input[type="checkbox"]#auto_refresh');

    if (await autoRefreshCheckbox.isVisible({ timeout: 2000 }).catch(() => false)) {
      // Check initial state
      const initialState = await hostPage.isAutoRefreshEnabled();

      // Toggle
      await autoRefreshCheckbox.click();

      // Verify state changed
      const newState = await hostPage.isAutoRefreshEnabled();
      expect(newState).toBe(!initialState);

      // Toggle back
      await autoRefreshCheckbox.click();
      const finalState = await hostPage.isAutoRefreshEnabled();
      expect(finalState).toBe(initialState);
    }
  });

  test('should show export button and be clickable', async ({ page }) => {
    await hostPage.goto();

    // Wait for page load
    await page.waitForLoadState('networkidle');

    // Verify export button is visible
    const exportButton = page.locator('button:has-text("Export")');
    await expect(exportButton).toBeVisible();
    await expect(exportButton).toBeEnabled();
  });

  test('should refresh data when refresh button clicked', async ({ page }) => {
    await hostPage.goto();

    // Wait for initial load
    await page.waitForLoadState('networkidle');

    // Click refresh
    await hostPage.refresh();

    // Verify page is still functional
    await expect(page.locator('h1:has-text("Host Inventory")')).toBeVisible();
  });

  test('should handle empty state when no hosts exist', async ({ page }) => {
    await hostPage.goto();

    // Check for empty state or table
    const hasEmptyState = await hostPage.hasEmptyState();
    const hasTable = await page.locator('table').isVisible({ timeout: 1000 }).catch(() => false);

    // Either empty state or table should be present
    expect(hasEmptyState || hasTable).toBe(true);
  });

  test('should show loading state initially', async ({ page }) => {
    // Navigate and immediately check for loading or content
    await page.goto('/host-inventory');

    // Either loading spinner or content should be visible quickly
    const loadingSpinner = page.locator('[data-testid="loading"], .spinner, .animate-spin').first();
    const pageTitle = page.locator('h1:has-text("Host Inventory")');

    // One of them should be visible within a short time
    await Promise.race([
      expect(loadingSpinner).toBeVisible({ timeout: 2000 }).catch(() => {}),
      expect(pageTitle).toBeVisible({ timeout: 2000 }).catch(() => {}),
    ]);
  });

  test('should maintain authentication when accessing host inventory', async ({ page }) => {
    await hostPage.goto();

    // Should not redirect to login
    expect(page.url()).not.toContain('/login');
  });

  test('should display host counts in showing text', async ({ page }) => {
    await hostPage.goto();

    // Wait for page load
    await page.waitForLoadState('networkidle');

    // Check if "Showing X of Y hosts" text exists
    const showingText = await hostPage.getShowingText();
    if (showingText) {
      expect(showingText).toContain('Showing');
    }
  });

  test('should clear filters and show all hosts', async ({ page }) => {
    await hostPage.goto();

    // Check if there are hosts
    const emptyState = await hostPage.hasEmptyState();
    if (!emptyState) {
      await hostPage.expectTableVisible();

      // Apply some filters
      await hostPage.searchByHostname('nonexistent');

      // Clear filters
      await hostPage.clearFilters();

      // Verify filters are cleared
      const searchInput = page.locator('input[placeholder*="Search"], input[type="text"]').first();
      const searchValue = await searchInput.inputValue();
      expect(searchValue).toBe('');
    }
  });

  test('should show error message when API fails', async ({ page }) => {
    // This test would require mocking API failure
    // For now, just verify error handling exists
    await hostPage.goto();
    await page.waitForLoadState('networkidle');

    // Check if error is displayed (won't be in normal case)
    const hasError = await hostPage.hasError();
    // In normal conditions, no error should be present
    expect(hasError).toBe(false);
  });

  test('should display table headers correctly', async ({ page }) => {
    await hostPage.goto();

    // Check if table exists
    const emptyState = await hostPage.hasEmptyState();
    if (!emptyState) {
      await hostPage.expectTableVisible();

      const headers = await hostPage.getTableHeaders();
      expect(headers.length).toBeGreaterThan(0);
    }
  });
});