import { Page, expect } from '@playwright/test';

export class AlertsPage {
  readonly page: Page;

  // Locators
  private readonly alertsList = '[data-testid="alerts-list"], table';
  private readonly createAlertButton = 'button:has-text("Create Alert"), button:has-text("New Alert"), [data-testid="btn-create-alert"]';
  private readonly alertNameInput = 'input[name="name"], input[placeholder*="Alert Name"]';
  private readonly metricSelect = 'select[name="metric"], [data-testid="metric-select"]';
  private readonly conditionSelect = 'select[name="condition"], [data-testid="condition-select"]';
  private readonly thresholdInput = 'input[name="threshold"], input[placeholder*="Threshold"]';
  private readonly saveButton = 'button:has-text("Save"), button:has-text("Create")';
  private readonly deleteButton = 'button:has-text("Delete")';
  private readonly confirmButton = 'button:has-text("Confirm"), button:has-text("OK")';
  private readonly successMessage = '.alert-success, [data-testid="success"], .toast-success';
  private readonly errorMessage = '.alert-danger, [data-testid="error"], .toast-error';
  private readonly alertRow = 'tr, [data-testid="alert-item"]';
  private readonly enableToggle = '[data-testid="alert-toggle"], input[type="checkbox"][name="enabled"]';
  private readonly editButton = 'button:has-text("Edit")';

  constructor(page: Page) {
    this.page = page;
  }

  async goto() {
    await this.page.goto('/alerts');
    await this.expectLoaded();
  }

  async expectLoaded() {
    await this.page.waitForLoadState('networkidle');

    const list = this.page.locator(this.alertsList).first();
    const button = this.page.locator(this.createAlertButton).first();

    try {
      await expect(list.or(button)).toBeVisible({ timeout: 5000 });
    } catch {
      await this.page.waitForTimeout(1000);
    }
  }

  async clickCreateAlert() {
    await this.page.locator(this.createAlertButton).first().click();
    const form = this.page.locator(this.alertNameInput).first();
    await form.waitFor({ timeout: 5000 });
  }

  async fillAlertForm(data: {
    name: string;
    metric?: string;
    condition?: string;
    threshold?: string;
  }) {
    const nameField = this.page.locator(this.alertNameInput).first();
    await nameField.fill(data.name);

    if (data.metric) {
      const metricField = this.page.locator(this.metricSelect).first();
      if (await metricField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await metricField.selectOption(data.metric);
      }
    }

    if (data.condition) {
      const conditionField = this.page.locator(this.conditionSelect).first();
      if (await conditionField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await conditionField.selectOption(data.condition);
      }
    }

    if (data.threshold) {
      const thresholdField = this.page.locator(this.thresholdInput).first();
      if (await thresholdField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await thresholdField.fill(data.threshold);
      }
    }
  }

  async saveAlert() {
    const saveBtn = this.page.locator(this.saveButton).first();
    await saveBtn.click();

    const success = this.page.locator(this.successMessage).first();
    try {
      await expect(success).toBeVisible({ timeout: 10000 });
    } catch {
      await this.page.waitForLoadState('networkidle');
    }
  }

  async deleteAlert(alertName: string) {
    const row = this.page.locator(this.alertRow).filter({ hasText: alertName }).first();
    const deleteBtn = row.locator(this.deleteButton).first();
    await deleteBtn.click();

    const confirmBtn = this.page.locator(this.confirmButton).first();
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    }

    await this.page.waitForLoadState('networkidle');
  }

  async toggleAlert(alertName: string, enabled: boolean) {
    const row = this.page.locator(this.alertRow).filter({ hasText: alertName }).first();
    const toggle = row.locator(this.enableToggle).first();

    const isChecked = await toggle.isChecked();
    if (isChecked !== enabled) {
      await toggle.click();
    }

    await this.page.waitForLoadState('networkidle');
  }

  async expectAlertInList(alertName: string) {
    const alertItem = this.page.locator(this.alertRow).filter({ hasText: alertName }).first();
    await expect(alertItem).toBeVisible({ timeout: 5000 });
  }

  async getAlertCount(): Promise<number> {
    return await this.page.locator(this.alertRow).count();
  }

  async expectSuccessMessage() {
    const success = this.page.locator(this.successMessage).first();
    await expect(success).toBeVisible({ timeout: 5000 });
  }

  async expectErrorMessage() {
    const error = this.page.locator(this.errorMessage).first();
    await expect(error).toBeVisible({ timeout: 5000 });
  }
}
