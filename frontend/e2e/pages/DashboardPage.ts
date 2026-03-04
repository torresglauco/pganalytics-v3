import { Page, expect } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;

  // Locators
  private readonly dashboardContainer = '[data-testid="dashboard"], main, .dashboard';
  private readonly navigationMenu = 'nav, [role="navigation"]';
  private readonly collectorsLink = 'a:has-text("Collectors"), a:has-text("Servers"), [data-testid="nav-collectors"]';
  private readonly alertsLink = 'a:has-text("Alerts"), [data-testid="nav-alerts"]';
  private readonly usersLink = 'a:has-text("Users"), a:has-text("Admin"), [data-testid="nav-users"]';
  private readonly chartsContainer = '[data-testid="chart"], .chart, canvas';
  private readonly loadingSpinner = '[data-testid="loading"], .spinner, .loader';
  private readonly errorAlert = '.alert-danger, [data-testid="error"]';

  constructor(page: Page) {
    this.page = page;
  }

  async goto() {
    await this.page.goto('/dashboard');
    await this.expectLoaded();
  }

  async expectLoaded() {
    const container = this.page.locator(this.dashboardContainer).first();
    await expect(container).toBeVisible({ timeout: 5000 });

    // Wait for loading to complete
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 10000 });
    }
  }

  async expectNoErrors() {
    const errors = this.page.locator(this.errorAlert);
    const count = await errors.count();
    expect(count).toBe(0);
  }

  async navigateToCollectors() {
    const link = this.page.locator(this.collectorsLink).first();
    await link.click();
    await this.page.waitForURL('**/collectors', { timeout: 5000 });
  }

  async navigateToAlerts() {
    const link = this.page.locator(this.alertsLink).first();
    await link.click();
    await this.page.waitForURL('**/alerts', { timeout: 5000 });
  }

  async navigateToUsers() {
    const link = this.page.locator(this.usersLink).first();
    await link.click();
    await this.page.waitForURL('**/users', { timeout: 5000 });
  }

  async getChartCount(): Promise<number> {
    return await this.page.locator(this.chartsContainer).count();
  }

  async expectChartsLoaded(minCharts: number = 1) {
    const chartCount = await this.getChartCount();
    expect(chartCount).toBeGreaterThanOrEqual(minCharts);
  }

  async waitForDataLoad() {
    // Wait for loading to complete
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 10000 });
    }

    // Wait a bit more for data to render
    await this.page.waitForTimeout(1000);
  }

  async getMetricsCount(): Promise<number> {
    // Count metric cards or items on dashboard
    return await this.page.locator('[data-testid="metric"], .metric-card').count();
  }

  async expectMetricsDisplayed(minCount: number = 1) {
    const metricsCount = await this.getMetricsCount();
    expect(metricsCount).toBeGreaterThanOrEqual(minCount);
  }
}
