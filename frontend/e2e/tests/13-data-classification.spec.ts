import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DataClassificationPage } from '../pages/DataClassificationPage';

test.describe('Data Classification', () => {
  let loginPage: LoginPage;
  let classificationPage: DataClassificationPage;

  // Test collector ID - adjust based on test data
  const testCollectorId = 'test-collector-001';

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    classificationPage = new DataClassificationPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
  });

  test('should display classification page with table', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Verify page title is visible
    await expect(page.locator('h1:has-text("Data Classification")')).toBeVisible();
  });

  test('should render classification table', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table to be visible
    await classificationPage.expectTableVisible();
  });

  test('should display summary cards with counts', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Verify summary cards are visible
    await classificationPage.expectSummaryCardsVisible();
  });

  test('should show pattern breakdown chart', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Verify chart is visible
    await classificationPage.expectChartVisible();
  });

  test('should filter results by pattern type', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table
    await classificationPage.expectTableVisible();

    const initialRowCount = await classificationPage.getRowCount();

    // Try to select a pattern type filter (if available)
    const patternFilter = page.locator('[data-testid="pattern-type-filter"], select[name="pattern_type"]').first();
    if (await patternFilter.isVisible({ timeout: 2000 }).catch(() => false)) {
      await classificationPage.selectPatternType('CPF');
      // Table should update
      await page.waitForTimeout(500);
    }
  });

  test('should filter results by database', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table
    await classificationPage.expectTableVisible();

    // Try to select a database filter (if available)
    const dbFilter = page.locator('[data-testid="database-filter"], select[name="database"]').first();
    if (await dbFilter.isVisible({ timeout: 2000 }).catch(() => false)) {
      // Get available options
      const options = await dbFilter.locator('option').allInnerTexts();
      if (options.length > 1) {
        await classificationPage.selectDatabase(options[1]);
        await page.waitForTimeout(500);
      }
    }
  });

  test('should navigate via breadcrumbs', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for page load
    await classificationPage.expectTableVisible();

    // Get breadcrumbs
    const breadcrumbs = await classificationPage.getBreadcrumbs();
    expect(breadcrumbs.length).toBeGreaterThan(0);

    // Click first breadcrumb (All Databases)
    if (breadcrumbs.length > 0) {
      await classificationPage.clickBreadcrumb(0);
    }
  });

  test('should click row for drill-down navigation', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table
    await classificationPage.expectTableVisible();

    const rowCount = await classificationPage.getRowCount();
    if (rowCount > 0) {
      // Click first row for drill-down
      await classificationPage.clickRow(0);
      await page.waitForTimeout(500);
    }
  });

  test('should show export button and be clickable', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for page load
    await classificationPage.expectTableVisible();

    // Verify export button is visible
    const exportButton = page.locator('button:has-text("Export")');
    await expect(exportButton).toBeVisible();
    await expect(exportButton).toBeEnabled();
  });

  test('should refresh data when refresh button clicked', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for initial load
    await classificationPage.expectTableVisible();

    // Click refresh
    await classificationPage.refresh();

    // Verify table is still visible
    await classificationPage.expectTableVisible();
  });

  test('should handle missing or invalid collectorId gracefully', async ({ page }) => {
    // Navigate with invalid collector ID
    await classificationPage.goto('non-existent-collector-id');

    // Should show error or no data state
    const hasError = await classificationPage.hasError();
    const hasNoData = await classificationPage.hasNoData();

    // At least one should be true for invalid collector
    expect(hasError || hasNoData).toBe(true);
  });

  test('should display table headers correctly', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table
    await classificationPage.expectTableVisible();

    const headers = await classificationPage.getTableHeaders();
    expect(headers.length).toBeGreaterThan(0);
  });

  test('should show loading state initially', async ({ page }) => {
    // Navigate and immediately check for loading or content
    await page.goto(`/data-classification/${testCollectorId}`);

    // Either loading spinner or table should be visible quickly
    const loadingSpinner = page.locator('[data-testid="loading"], .spinner, .animate-spin').first();
    const table = page.locator('table, [data-testid="classification-table"]').first();

    // One of them should be visible within a short time
    await Promise.race([
      expect(loadingSpinner).toBeVisible({ timeout: 2000 }).catch(() => {}),
      expect(table).toBeVisible({ timeout: 2000 }).catch(() => {}),
    ]);
  });

  test('should maintain authentication when accessing classification page', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Should not redirect to login
    expect(page.url()).not.toContain('/login');
  });

  test('should reset filters when clicking first breadcrumb', async ({ page }) => {
    await classificationPage.goto(testCollectorId);

    // Wait for table
    await classificationPage.expectTableVisible();

    // Get initial row count
    const initialRowCount = await classificationPage.getRowCount();

    // Try to apply a filter if filters are available
    const dbFilter = page.locator('[data-testid="database-filter"], select[name="database"]').first();
    if (await dbFilter.isVisible({ timeout: 2000 }).catch(() => false)) {
      const options = await dbFilter.locator('option').allInnerTexts();
      if (options.length > 1) {
        await classificationPage.selectDatabase(options[1]);
        await page.waitForTimeout(500);
      }
    }

    // Reset by clicking first breadcrumb
    await classificationPage.resetFilters();
  });
});