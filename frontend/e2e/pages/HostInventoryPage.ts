import { Page, expect } from '@playwright/test';

/**
 * Page Object Model for Host Inventory Page
 * Handles navigation and interaction with host inventory monitoring
 */
export class HostInventoryPage {
  readonly page: Page;

  // Locators
  private readonly pageTitle = 'h1:has-text("Host Inventory")';
  private readonly hostTable = 'table, [data-testid="host-table"]';
  private readonly tableRows = 'tbody tr, [data-testid="host-row"]';
  private readonly loadingSpinner = '[data-testid="loading"], .spinner, .animate-spin';
  private readonly errorMessage = '.bg-red-50, [data-testid="error"], .alert-danger';
  private readonly exportButton = 'button:has-text("Export")';
  private readonly refreshButton = 'button:has-text("Refresh")';
  private readonly searchInput = 'input[placeholder*="Search"], input[type="text"]';
  private readonly statusFilter = 'select';
  private readonly autoRefreshCheckbox = 'input[type="checkbox"]#auto_refresh';
  private readonly summaryCards = '[data-testid="summary-card"], .summary-card';
  private readonly detailPanel = '[data-testid="host-detail-panel"], .host-detail';
  private readonly closeDetailButton = 'button:has-text("Close"), button:has-text("Back")';
  private readonly emptyState = 'text=No hosts configured, text=No hosts match';
  private readonly noHostsMessage = 'text=No hosts configured';

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to host inventory page
   */
  async goto() {
    await this.page.goto('/host-inventory');
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
   * Wait for host table to render
   */
  async expectTableVisible() {
    const table = this.page.locator(this.hostTable).first();
    await expect(table).toBeVisible({ timeout: 10000 });
  }

  /**
   * Enter search term to filter hosts by hostname
   */
  async searchByHostname(term: string) {
    const input = this.page.locator(this.searchInput).first();
    await input.fill(term);
    // Wait for filtering to apply
    await this.page.waitForTimeout(500);
  }

  /**
   * Clear search input
   */
  async clearSearch() {
    const input = this.page.locator(this.searchInput).first();
    await input.fill('');
    await this.page.waitForTimeout(500);
  }

  /**
   * Select status filter (up, down, unknown)
   */
  async filterByStatus(status: 'up' | 'down' | 'unknown' | '') {
    const filter = this.page.locator(this.statusFilter).first();
    await filter.selectOption(status);
    await this.page.waitForTimeout(500);
  }

  /**
   * Get the number of visible hosts
   */
  async getRowCount(): Promise<number> {
    await this.expectTableVisible();
    return await this.page.locator(this.tableRows).count();
  }

  /**
   * Click on a host row to view details
   */
  async clickHost(collectorId: string) {
    await this.expectTableVisible();
    const row = this.page.locator(this.tableRows).filter({ hasText: collectorId }).first();
    await row.click();
    // Wait for detail panel or navigation
    await this.page.waitForTimeout(500);
  }

  /**
   * Click on a host row by index
   */
  async clickHostByIndex(index: number) {
    await this.expectTableVisible();
    const row = this.page.locator(this.tableRows).nth(index);
    await row.click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Verify detail panel opens after clicking a host
   */
  async expectDetailPanelOpen() {
    const panel = this.page.locator(this.detailPanel).first();
    await expect(panel).toBeVisible({ timeout: 5000 });
  }

  /**
   * Close the detail panel
   */
  async closeDetailPanel() {
    const button = this.page.locator(this.closeDetailButton).first();
    await button.click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Toggle auto-refresh checkbox
   */
  async toggleAutoRefresh() {
    const checkbox = this.page.locator(this.autoRefreshCheckbox);
    await checkbox.check();

    // Toggle the checkbox
    const isChecked = await checkbox.isChecked();
    if (isChecked) {
      await checkbox.uncheck();
    } else {
      await checkbox.check();
    }
  }

  /**
   * Check if auto-refresh is enabled
   */
  async isAutoRefreshEnabled(): Promise<boolean> {
    const checkbox = this.page.locator(this.autoRefreshCheckbox);
    return await checkbox.isChecked();
  }

  /**
   * Click the export button to download CSV
   */
  async exportCsv() {
    const button = this.page.locator(this.exportButton);
    await expect(button).toBeEnabled({ timeout: 5000 });
    await button.click();
  }

  /**
   * Click the export button and wait for download
   */
  async exportCsvAndWaitForDownload(): Promise<string | null> {
    const [download] = await Promise.all([
      this.page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
      this.exportCsv(),
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
   * Verify summary cards are visible
   */
  async expectSummaryCardsVisible() {
    const cards = this.page.locator(this.summaryCards);
    const count = await cards.count();
    expect(count).toBeGreaterThan(0);
  }

  /**
   * Get summary card values (up, down, unknown counts)
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
   * Check if empty state is displayed (no hosts)
   */
  async hasEmptyState(): Promise<boolean> {
    const emptyState = this.page.locator(this.emptyState);
    return await emptyState.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Check if no hosts message is displayed
   */
  async hasNoHostsMessage(): Promise<boolean> {
    const message = this.page.locator(this.noHostsMessage);
    return await message.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Get all visible host hostnames
   */
  async getVisibleHostnames(): Promise<string[]> {
    await this.expectTableVisible();
    const rows = this.page.locator(this.tableRows);
    const count = await rows.count();
    const hostnames: string[] = [];

    for (let i = 0; i < count; i++) {
      // Assuming hostname is in the first column
      const hostname = await rows.nth(i).locator('td').first().textContent();
      if (hostname) {
        hostnames.push(hostname.trim());
      }
    }

    return hostnames;
  }

  /**
   * Get status indicator for a host
   */
  async getHostStatus(hostname: string): Promise<string | null> {
    const row = this.page.locator(this.tableRows).filter({ hasText: hostname }).first();
    if (await row.isVisible({ timeout: 1000 }).catch(() => false)) {
      // Status might be in a badge or icon
      const statusBadge = row.locator('.badge, [data-testid="status"]');
      if (await statusBadge.isVisible().catch(() => false)) {
        return await statusBadge.textContent();
      }
    }
    return null;
  }

  /**
   * Clear all filters
   */
  async clearFilters() {
    await this.clearSearch();
    await this.filterByStatus('');
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

  /**
   * Get showing X of Y hosts text
   */
  async getShowingText(): Promise<string | null> {
    const text = this.page.locator('text=Showing');
    if (await text.isVisible({ timeout: 1000 }).catch(() => false)) {
      return await text.textContent();
    }
    return null;
  }
}