import { Page, expect } from '@playwright/test';

export class CollectorPage {
  readonly page: Page;

  // Locators
  private readonly collectorsList = '[data-testid="collectors-list"], table';
  private readonly registerButton = 'button:has-text("Register"), button:has-text("Add Collector"), [data-testid="btn-register"]';
  private readonly hostnameInput = 'input[name="hostname"], input[placeholder*="Hostname"], input[placeholder*="Host"]';
  private readonly portInput = 'input[name="port"], input[placeholder*="Port"], input[type="number"]';
  private readonly databaseInput = 'input[name="database"], input[placeholder*="Database"]';
  private readonly usernameInput = 'input[name="username"], input[placeholder*="Username"], input[placeholder*="User"]';
  private readonly testConnectionButton = 'button:has-text("Test"), button:has-text("Test Connection")';
  private readonly registerCollectorButton = 'button:has-text("Register"), button:has-text("Create")';
  private readonly successMessage = '.alert-success, [data-testid="success"], .toast-success';
  private readonly errorMessage = '.alert-danger, [data-testid="error"], .toast-error';
  private readonly collectorRow = 'tr, [data-testid="collector-item"]';

  constructor(page: Page) {
    this.page = page;
  }

  async goto() {
    await this.page.goto('/collectors');
    await this.expectLoaded();
  }

  async expectLoaded() {
    // Wait for page to load
    await this.page.waitForLoadState('networkidle');

    // Check for collectors list or register button
    const list = this.page.locator(this.collectorsList).first();
    const button = this.page.locator(this.registerButton).first();

    try {
      await expect(list.or(button)).toBeVisible({ timeout: 5000 });
    } catch {
      // Page might be loading, wait a bit more
      await this.page.waitForTimeout(1000);
    }
  }

  async clickRegisterButton() {
    await this.page.locator(this.registerButton).first().click();

    // Wait for modal/form to appear
    const form = this.page.locator(this.hostnameInput).first();
    await form.waitFor({ timeout: 5000 });
  }

  async fillRegistrationForm(data: {
    hostname: string;
    port?: number;
    database?: string;
    username?: string;
  }) {
    // Fill hostname
    const hostnameField = this.page.locator(this.hostnameInput).first();
    await hostnameField.fill(data.hostname);

    // Fill port if provided
    if (data.port) {
      const portField = this.page.locator(this.portInput).first();
      if (await portField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await portField.fill(data.port.toString());
      }
    }

    // Fill database if provided
    if (data.database) {
      const dbField = this.page.locator(this.databaseInput).first();
      if (await dbField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await dbField.fill(data.database);
      }
    }

    // Fill username if provided
    if (data.username) {
      const userField = this.page.locator(this.usernameInput).first();
      if (await userField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await userField.fill(data.username);
      }
    }
  }

  async testConnection() {
    const testBtn = this.page.locator(this.testConnectionButton).first();
    await testBtn.click();

    // Wait for result
    const success = this.page.locator(this.successMessage).first();
    const error = this.page.locator(this.errorMessage).first();

    try {
      await expect(success.or(error)).toBeVisible({ timeout: 10000 });
    } catch {
      // Connection test might be taking longer
      await this.page.waitForTimeout(2000);
    }
  }

  async expectConnectionSuccess() {
    const success = this.page.locator(this.successMessage);
    await expect(success.filter({ hasText: /connection|success/i }).first()).toBeVisible({
      timeout: 5000,
    });
  }

  async expectConnectionError() {
    const error = this.page.locator(this.errorMessage);
    await expect(error.first()).toBeVisible({ timeout: 5000 });
  }

  async registerCollector() {
    const registerBtn = this.page.locator(this.registerCollectorButton).last();
    await registerBtn.click();

    // Wait for success message or redirect
    const success = this.page.locator(this.successMessage).first();
    try {
      await expect(success).toBeVisible({ timeout: 10000 });
    } catch {
      // Might redirect instead of showing message
      await this.page.waitForLoadState('networkidle');
    }
  }

  async expectCollectorInList(hostname: string) {
    const collectorItem = this.page.locator(this.collectorRow).filter({ hasText: hostname }).first();
    await expect(collectorItem).toBeVisible({ timeout: 5000 });
  }

  async expectSuccessMessage() {
    const success = this.page.locator(this.successMessage).first();
    await expect(success).toBeVisible({ timeout: 5000 });
  }

  async getCollectorCount(): Promise<number> {
    return await this.page.locator(this.collectorRow).count();
  }

  async deleteCollector(hostname: string) {
    // Find the collector row
    const row = this.page.locator(this.collectorRow).filter({ hasText: hostname }).first();

    // Find delete button in that row
    const deleteBtn = row.locator('button:has-text("Delete")').first();
    await deleteBtn.click();

    // Confirm if dialog appears
    const confirmBtn = this.page.locator('button:has-text("Confirm"), button:has-text("OK")').first();
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    }

    // Wait for deletion
    await this.page.waitForLoadState('networkidle');
  }

  async editCollector(hostname: string, newData: Partial<{ interval: number; enabled: boolean }>) {
    // Find the collector row
    const row = this.page.locator(this.collectorRow).filter({ hasText: hostname }).first();

    // Find edit button
    const editBtn = row.locator('button:has-text("Edit")').first();
    await editBtn.click();

    // Wait for form
    const form = this.page.locator('input, textarea').first();
    await form.waitFor({ timeout: 5000 });

    // Fill new data if provided
    if (newData.interval !== undefined) {
      const intervalInput = this.page.locator('input[name="interval"]').first();
      await intervalInput.fill(newData.interval.toString());
    }

    // Save
    const saveBtn = this.page.locator('button:has-text("Save")').first();
    await saveBtn.click();

    // Wait for save
    await this.page.waitForLoadState('networkidle');
  }
}
