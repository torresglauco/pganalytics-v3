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

    // ✅ UPDATED: Wait for either URL change OR dashboard element
    // (no silent failures - test will fail if neither happens)
    await this.page.waitForFunction(
      () => {
        const isDashboardUrl = window.location.pathname.includes('/dashboard');
        const isDashboardElement = document.querySelector('[data-testid="dashboard"]') !== null;
        return isDashboardUrl || isDashboardElement;
      },
      { timeout: 10000 }
    );
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
    // ✅ UPDATED: Wait for either URL or element (no silent failures)
    await this.page.waitForFunction(
      () => {
        const isDashboardUrl = window.location.pathname.includes('/dashboard');
        const isDashboardElement = document.querySelector('[data-testid="dashboard"]') !== null;
        return isDashboardUrl || isDashboardElement;
      },
      { timeout: 5000 }
    );
  }

  async expectLoggedOut() {
    // ✅ UPDATED: Wait for either URL or element (no silent failures)
    await this.page.waitForFunction(
      () => {
        const isLoginUrl = window.location.pathname.includes('/login');
        const isLoginForm = document.querySelector('button:has-text("Sign In")') !== null ||
                           document.querySelector('button:has-text("Login")') !== null;
        return isLoginUrl || isLoginForm;
      },
      { timeout: 5000 }
    );
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
