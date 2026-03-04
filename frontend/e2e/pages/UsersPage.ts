import { Page, expect } from '@playwright/test';

export class UsersPage {
  readonly page: Page;

  // Locators
  private readonly usersList = '[data-testid="users-list"], table';
  private readonly createUserButton = 'button:has-text("Create User"), button:has-text("Add User"), [data-testid="btn-create-user"]';
  private readonly emailInput = 'input[name="email"], input[type="email"], input[placeholder*="Email"]';
  private readonly passwordInput = 'input[name="password"], input[type="password"], input[placeholder*="Password"]';
  private readonly nameInput = 'input[name="name"], input[name="fullName"], input[placeholder*="Name"]';
  private readonly roleSelect = 'select[name="role"], [data-testid="role-select"]';
  private readonly saveButton = 'button:has-text("Save"), button:has-text("Create")';
  private readonly deleteButton = 'button:has-text("Delete")';
  private readonly editButton = 'button:has-text("Edit")';
  private readonly confirmButton = 'button:has-text("Confirm"), button:has-text("OK")';
  private readonly successMessage = '.alert-success, [data-testid="success"], .toast-success';
  private readonly errorMessage = '.alert-danger, [data-testid="error"], .toast-error';
  private readonly userRow = 'tr, [data-testid="user-item"]';
  private readonly changePasswordButton = 'button:has-text("Change Password")';

  constructor(page: Page) {
    this.page = page;
  }

  async goto() {
    await this.page.goto('/users');
    await this.expectLoaded();
  }

  async expectLoaded() {
    await this.page.waitForLoadState('networkidle');

    const list = this.page.locator(this.usersList).first();
    const button = this.page.locator(this.createUserButton).first();

    try {
      await expect(list.or(button)).toBeVisible({ timeout: 5000 });
    } catch {
      await this.page.waitForTimeout(1000);
    }
  }

  async clickCreateUser() {
    await this.page.locator(this.createUserButton).first().click();
    const form = this.page.locator(this.emailInput).first();
    await form.waitFor({ timeout: 5000 });
  }

  async fillUserForm(data: {
    email: string;
    password?: string;
    name?: string;
    role?: string;
  }) {
    const emailField = this.page.locator(this.emailInput).first();
    await emailField.fill(data.email);

    if (data.password) {
      const passwordField = this.page.locator(this.passwordInput).first();
      if (await passwordField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await passwordField.fill(data.password);
      }
    }

    if (data.name) {
      const nameField = this.page.locator(this.nameInput).first();
      if (await nameField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await nameField.fill(data.name);
      }
    }

    if (data.role) {
      const roleField = this.page.locator(this.roleSelect).first();
      if (await roleField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await roleField.selectOption(data.role);
      }
    }
  }

  async saveUser() {
    const saveBtn = this.page.locator(this.saveButton).first();
    await saveBtn.click();

    const success = this.page.locator(this.successMessage).first();
    try {
      await expect(success).toBeVisible({ timeout: 10000 });
    } catch {
      await this.page.waitForLoadState('networkidle');
    }
  }

  async deleteUser(email: string) {
    const row = this.page.locator(this.userRow).filter({ hasText: email }).first();
    const deleteBtn = row.locator(this.deleteButton).first();
    await deleteBtn.click();

    const confirmBtn = this.page.locator(this.confirmButton).first();
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    }

    await this.page.waitForLoadState('networkidle');
  }

  async editUser(email: string, newData: Partial<{
    name: string;
    role: string;
  }>) {
    const row = this.page.locator(this.userRow).filter({ hasText: email }).first();
    const editBtn = row.locator(this.editButton).first();
    await editBtn.click();

    const form = this.page.locator(this.emailInput).first();
    await form.waitFor({ timeout: 5000 });

    if (newData.name) {
      const nameField = this.page.locator(this.nameInput).first();
      if (await nameField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await nameField.fill(newData.name);
      }
    }

    if (newData.role) {
      const roleField = this.page.locator(this.roleSelect).first();
      if (await roleField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await roleField.selectOption(newData.role);
      }
    }

    const saveBtn = this.page.locator(this.saveButton).first();
    await saveBtn.click();

    await this.page.waitForLoadState('networkidle');
  }

  async expectUserInList(email: string) {
    const userItem = this.page.locator(this.userRow).filter({ hasText: email }).first();
    await expect(userItem).toBeVisible({ timeout: 5000 });
  }

  async getUserCount(): Promise<number> {
    return await this.page.locator(this.userRow).count();
  }

  async expectSuccessMessage() {
    const success = this.page.locator(this.successMessage).first();
    await expect(success).toBeVisible({ timeout: 5000 });
  }

  async expectErrorMessage() {
    const error = this.page.locator(this.errorMessage).first();
    await expect(error).toBeVisible({ timeout: 5000 });
  }

  async changePassword(email: string, newPassword: string) {
    const row = this.page.locator(this.userRow).filter({ hasText: email }).first();
    const changeBtn = row.locator(this.changePasswordButton).first();

    if (await changeBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await changeBtn.click();

      const passwordField = this.page.locator(this.passwordInput).first();
      await passwordField.waitFor({ timeout: 3000 });
      await passwordField.fill(newPassword);

      const saveBtn = this.page.locator(this.saveButton).first();
      await saveBtn.click();

      await this.page.waitForLoadState('networkidle');
    }
  }
}
