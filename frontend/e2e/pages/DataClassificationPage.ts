import { Page, expect } from '@playwright/test';

/**
 * Page Object Model for Data Classification Page
 * Handles navigation and interaction with data classification results
 */
export class DataClassificationPage {
  readonly page: Page;

  // Locators
  private readonly pageTitle = 'h1:has-text("Data Classification")';
  private readonly classificationTable = 'table, [data-testid="classification-table"]';
  private readonly tableRows = 'tbody tr, [data-testid="classification-row"]';
  private readonly loadingSpinner = '[data-testid="loading"], .spinner, .animate-spin';
  private readonly errorMessage = '.bg-red-50, [data-testid="error"], .alert-danger';
  private readonly exportButton = 'button:has-text("Export")';
  private readonly refreshButton = 'button:has-text("Refresh")';
  private readonly databaseFilter = 'select[name="database"], [data-testid="database-filter"]';
  private readonly schemaFilter = 'select[name="schema"], [data-testid="schema-filter"]';
  private readonly tableFilter = 'select[name="table"], [data-testid="table-filter"]';
  private readonly patternTypeFilter = 'select[name="pattern_type"], [data-testid="pattern-type-filter"]';
  private readonly categoryFilter = 'select[name="category"], [data-testid="category-filter"]';
  private readonly summaryCards = '[data-testid="summary-card"], .summary-card';
  private readonly patternBreakdownChart = '[data-testid="pattern-breakdown-chart"], canvas';
  private readonly breadcrumbs = 'nav button';
  private readonly noCollectorMessage = 'text=No collector selected';
  private readonly noDataMessage = 'text=No classification data';

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to data classification page for a specific collector
   */
  async goto(collectorId: string) {
    await this.page.goto(`/data-classification/${collectorId}`);
    await this.expectPageLoaded();
  }

  /**
   * Wait for page to be fully loaded
   */
  async expectPageLoaded() {
    // Wait for page title
    await expect(this.page.locator(this.pageTitle)).toBeVisible({ timeout: 10000 });

    // Wait for loading to complete
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 15000 });
    }
  }

  /**
   * Wait for classification table to render
   */
  async expectTableVisible() {
    const table = this.page.locator(this.classificationTable).first();
    await expect(table).toBeVisible({ timeout: 10000 });
  }

  /**
   * Select a database from the dropdown filter
   */
  async selectDatabase(name: string) {
    const filter = this.page.locator(this.databaseFilter).first();
    await filter.selectOption({ label: name });
    // Wait for table to update
    await this.page.waitForTimeout(500);
  }

  /**
   * Select a schema from the dropdown filter
   */
  async selectSchema(name: string) {
    const filter = this.page.locator(this.schemaFilter).first();
    await filter.selectOption({ label: name });
    await this.page.waitForTimeout(500);
  }

  /**
   * Select a table from the dropdown filter
   */
  async selectTable(name: string) {
    const filter = this.page.locator(this.tableFilter).first();
    await filter.selectOption({ label: name });
    await this.page.waitForTimeout(500);
  }

  /**
   * Select a pattern type from the dropdown filter
   */
  async selectPatternType(type: string) {
    const filter = this.page.locator(this.patternTypeFilter).first();
    await filter.selectOption({ label: type });
    await this.page.waitForTimeout(500);
  }

  /**
   * Select a category from the dropdown filter
   */
  async selectCategory(category: string) {
    const filter = this.page.locator(this.categoryFilter).first();
    await filter.selectOption({ label: category });
    await this.page.waitForTimeout(500);
  }

  /**
   * Get the number of visible table rows
   */
  async getRowCount(): Promise<number> {
    await this.expectTableVisible();
    return await this.page.locator(this.tableRows).count();
  }

  /**
   * Click on a table row by index for drill-down navigation
   */
  async clickRow(index: number) {
    await this.expectTableVisible();
    const row = this.page.locator(this.tableRows).nth(index);
    await row.click();
    // Wait for navigation/filters to update
    await this.page.waitForTimeout(500);
  }

  /**
   * Click on the first row that matches a pattern type
   */
  async clickRowByPattern(patternType: string) {
    await this.expectTableVisible();
    const row = this.page.locator(this.tableRows).filter({ hasText: patternType }).first();
    await row.click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Verify summary cards are visible
   */
  async expectSummaryCardsVisible() {
    const cards = this.page.locator(this.summaryCards);
    const count = await cards.count();
    expect(count).toBeGreaterThan(0);
  }

  /**
   * Get summary card values (total classifications, by category, etc.)
   */
  async getSummaryValues(): Promise<Record<string, string>> {
    const cards = this.page.locator(this.summaryCards);
    const values: Record<string, string> = {};
    const count = await cards.count();

    for (let i = 0; i < count; i++) {
      const card = cards.nth(i);
      const label = await card.locator('h3, .label').first().textContent();
      const value = await card.locator('.value, p').first().textContent();
      if (label && value) {
        values[label.trim()] = value.trim();
      }
    }

    return values;
  }

  /**
   * Verify pattern breakdown chart is visible
   */
  async expectChartVisible() {
    const chart = this.page.locator(this.patternBreakdownChart).first();
    await expect(chart).toBeVisible({ timeout: 5000 });
  }

  /**
   * Click the export button
   */
  async exportReport() {
    const button = this.page.locator(this.exportButton);
    await expect(button).toBeEnabled({ timeout: 5000 });
    await button.click();
  }

  /**
   * Click the export button and wait for download
   */
  async exportReportAndWaitForDownload(): Promise<string | null> {
    const [download] = await Promise.all([
      this.page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
      this.exportReport(),
    ]);

    return download?.suggestedFilename() || null;
  }

  /**
   * Click the refresh button
   */
  async refresh() {
    const button = this.page.locator(this.refreshButton);
    await expect(button).toBeEnabled({ timeout: 5000 });
    await button.click();

    // Wait for loading to complete
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 15000 });
    }
  }

  /**
   * Check if error message is displayed
   */
  async hasError(): Promise<boolean> {
    const error = this.page.locator(this.errorMessage);
    return await error.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Get error message text if displayed
   */
  async getErrorMessage(): Promise<string | null> {
    const error = this.page.locator(this.errorMessage);
    if (await error.isVisible({ timeout: 1000 }).catch(() => false)) {
      return await error.textContent();
    }
    return null;
  }

  /**
   * Check if no collector message is displayed
   */
  async hasNoCollectorMessage(): Promise<boolean> {
    const message = this.page.locator(this.noCollectorMessage);
    return await message.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Check if no data message is displayed
   */
  async hasNoData(): Promise<boolean> {
    const message = this.page.locator(this.noDataMessage);
    return await message.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Get breadcrumb items
   */
  async getBreadcrumbs(): Promise<string[]> {
    const crumbs = this.page.locator(this.breadcrumbs);
    const count = await crumbs.count();
    const items: string[] = [];

    for (let i = 0; i < count; i++) {
      const text = await crumbs.nth(i).textContent();
      if (text) {
        items.push(text.trim());
      }
    }

    return items;
  }

  /**
   * Click on a breadcrumb item by index
   */
  async clickBreadcrumb(index: number) {
    const crumbs = this.page.locator(this.breadcrumbs);
    await crumbs.nth(index).click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Get all visible pattern types in the table
   */
  async getVisiblePatternTypes(): Promise<string[]> {
    await this.expectTableVisible();
    // Assuming pattern type is in a specific column
    const patternCells = this.page.locator('tbody td:nth-child(4)'); // Adjust column index as needed
    const count = await patternCells.count();
    const types: string[] = [];

    for (let i = 0; i < count; i++) {
      const text = await patternCells.nth(i).textContent();
      if (text && !types.includes(text.trim())) {
        types.push(text.trim());
      }
    }

    return types;
  }

  /**
   * Reset all filters
   */
  async resetFilters() {
    // Click on the first breadcrumb to reset to all databases
    await this.clickBreadcrumb(0);
  }

  /**
   * Get table headers
   */
  async getTableHeaders(): Promise<string[]> {
    await this.expectTableVisible();
    const headers = this.page.locator('thead th');
    const count = await headers.count();
    const headerTexts: string[] = [];

    for (let i = 0; i < count; i++) {
      const text = await headers.nth(i).textContent();
      if (text) {
        headerTexts.push(text.trim());
      }
    }

    return headerTexts;
  }
}