import { Page, expect } from '@playwright/test';

export class LoginPage {
  readonly page: Page;

  // Locators
  private readonly emailInput = 'input[name="email"], input[placeholder*="Email"], input[type="email"]';
  private readonly passwordInput = 'input[name="password"], input[placeholder*="Password"], input[type="password"]';
  private readonly loginButton = 'button:has-text("Sign In"), button:has-text("Login"), button:has-text("Submit")';
  private readonly logoutButton = '[data-testid="user-menu"], button:has-text("Logout")';
  private readonly errorMessage = '[data-testid="error"], .error, .alert-danger';
  private readonly userMenu = '[data-testid="user-menu"], [data-testid="user-dropdown"]';

  constructor(page: Page) {
    this.page = page;
  }

  async goto() {
    await this.page.goto('/login');
    await this.page.waitForLoadState('networkidle');
  }

  async login(email: string, password: string) {
    // Fill email
    const emailField = this.page.locator(this.emailInput).first();
    await emailField.fill(email);

    // Fill password
    const passwordField = this.page.locator(this.passwordInput).first();
    await passwordField.fill(password);

    // Click login button
    await this.page.locator(this.loginButton).first().click();

    // Wait for navigation or success
    try {
      await this.page.waitForURL('/dashboard', { timeout: 10000 });
    } catch {
      // Try waiting for success element instead
      await this.page.locator('[data-testid="dashboard"]').first().waitFor({ timeout: 5000 });
    }
  }

  async logout() {
    // Click user menu if exists
    const userMenuButton = this.page.locator(this.userMenu).first();
    if (await userMenuButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await userMenuButton.click();
    }

    // Click logout
    const logoutBtn = this.page.locator(this.logoutButton).filter({ hasText: 'Logout' });
    await logoutBtn.click();

    // Wait for redirect to login
    await this.page.waitForURL('/login', { timeout: 5000 });
  }

  async expectLoggedIn() {
    try {
      await this.page.waitForURL('/dashboard', { timeout: 5000 });
    } catch {
      // Alternative: check for dashboard element
      await expect(this.page.locator('[data-testid="dashboard"]').first()).toBeVisible({
        timeout: 5000,
      });
    }
  }

  async expectLoggedOut() {
    try {
      await this.page.waitForURL('/login', { timeout: 5000 });
    } catch {
      // Alternative: check for login form
      await expect(this.page.locator(this.loginButton).first()).toBeVisible({
        timeout: 5000,
      });
    }
  }

  async expectErrorMessage() {
    const error = this.page.locator(this.errorMessage).first();
    await expect(error).toBeVisible({ timeout: 5000 });
    return error.textContent();
  }

  async fillEmail(email: string) {
    await this.page.locator(this.emailInput).first().fill(email);
  }

  async fillPassword(password: string) {
    await this.page.locator(this.passwordInput).first().fill(password);
  }

  async clickLogin() {
    await this.page.locator(this.loginButton).first().click();
  }
}
