# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests/05-user-management.spec.ts >> User Management >> should handle user deletion with confirmation
- Location: e2e/tests/05-user-management.spec.ts:148:3

# Error details

```
Test timeout of 30000ms exceeded while running "beforeEach" hook.
```

```
Error: locator.fill: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('input[name="email"], input[placeholder*="Email"], input[type="email"]').first()

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - img "pgAnalytics" [ref=e6]
      - heading "pgAnalytics" [level=1] [ref=e7]
      - paragraph [ref=e8]: PostgreSQL Observability Platform
    - generic [ref=e9]:
      - generic [ref=e10]:
        - heading "Monitor in Real-Time" [level=3] [ref=e11]
        - paragraph [ref=e12]: Track PostgreSQL logs, metrics, and alerts in a unified dashboard
      - generic [ref=e13]:
        - heading "Deep Analysis" [level=3] [ref=e14]
        - paragraph [ref=e15]: Find slow queries, errors, and performance issues instantly
      - generic [ref=e16]:
        - heading "Proactive Alerting" [level=3] [ref=e17]
        - paragraph [ref=e18]: Get notified before problems impact your users
  - generic [ref=e20]:
    - heading "Sign In" [level=2] [ref=e21]
    - paragraph [ref=e22]: Welcome back to pgAnalytics
    - generic [ref=e23]:
      - generic [ref=e24]:
        - generic [ref=e25]: Username*
        - textbox "admin" [ref=e27]
      - generic [ref=e28]:
        - generic [ref=e29]: Password*
        - textbox "••••••••" [ref=e31]
      - generic [ref=e32]:
        - checkbox "Remember me" [ref=e33]
        - generic [ref=e34]: Remember me
      - button "Sign In" [ref=e35] [cursor=pointer]
    - generic [ref=e36]:
      - generic [ref=e41]: Or continue with
      - button "SSO Login" [ref=e42] [cursor=pointer]
    - paragraph [ref=e43]:
      - text: Don't have an account?
      - link "Sign up" [ref=e44] [cursor=pointer]:
        - /url: /signup
```

# Test source

```ts
  1  | import { Page, expect } from '@playwright/test';
  2  | 
  3  | export class LoginPage {
  4  |   readonly page: Page;
  5  | 
  6  |   // Locators
  7  |   private readonly emailInput = 'input[name="email"], input[placeholder*="Email"], input[type="email"]';
  8  |   private readonly passwordInput = 'input[name="password"], input[placeholder*="Password"], input[type="password"]';
  9  |   private readonly loginButton = 'button:has-text("Sign In"), button:has-text("Login"), button:has-text("Submit")';
  10 |   private readonly logoutButton = '[data-testid="user-menu"], button:has-text("Logout")';
  11 |   private readonly errorMessage = '[data-testid="error"], .error, .alert-danger';
  12 |   private readonly userMenu = '[data-testid="user-menu"], [data-testid="user-dropdown"]';
  13 | 
  14 |   constructor(page: Page) {
  15 |     this.page = page;
  16 |   }
  17 | 
  18 |   async goto() {
  19 |     await this.page.goto('/login');
  20 |     await this.page.waitForLoadState('networkidle');
  21 |   }
  22 | 
  23 |   async login(email: string, password: string) {
  24 |     // Fill email
  25 |     const emailField = this.page.locator(this.emailInput).first();
> 26 |     await emailField.fill(email);
     |                      ^ Error: locator.fill: Test timeout of 30000ms exceeded.
  27 | 
  28 |     // Fill password
  29 |     const passwordField = this.page.locator(this.passwordInput).first();
  30 |     await passwordField.fill(password);
  31 | 
  32 |     // Click login button
  33 |     await this.page.locator(this.loginButton).first().click();
  34 | 
  35 |     // Wait for navigation or success
  36 |     try {
  37 |       await this.page.waitForURL('/dashboard', { timeout: 10000 });
  38 |     } catch {
  39 |       // Try waiting for success element instead
  40 |       await this.page.locator('[data-testid="dashboard"]').first().waitFor({ timeout: 5000 });
  41 |     }
  42 |   }
  43 | 
  44 |   async logout() {
  45 |     // Click user menu if exists
  46 |     const userMenuButton = this.page.locator(this.userMenu).first();
  47 |     if (await userMenuButton.isVisible({ timeout: 2000 }).catch(() => false)) {
  48 |       await userMenuButton.click();
  49 |     }
  50 | 
  51 |     // Click logout
  52 |     const logoutBtn = this.page.locator(this.logoutButton).filter({ hasText: 'Logout' });
  53 |     await logoutBtn.click();
  54 | 
  55 |     // Wait for redirect to login
  56 |     await this.page.waitForURL('/login', { timeout: 5000 });
  57 |   }
  58 | 
  59 |   async expectLoggedIn() {
  60 |     try {
  61 |       await this.page.waitForURL('/dashboard', { timeout: 5000 });
  62 |     } catch {
  63 |       // Alternative: check for dashboard element
  64 |       await expect(this.page.locator('[data-testid="dashboard"]').first()).toBeVisible({
  65 |         timeout: 5000,
  66 |       });
  67 |     }
  68 |   }
  69 | 
  70 |   async expectLoggedOut() {
  71 |     try {
  72 |       await this.page.waitForURL('/login', { timeout: 5000 });
  73 |     } catch {
  74 |       // Alternative: check for login form
  75 |       await expect(this.page.locator(this.loginButton).first()).toBeVisible({
  76 |         timeout: 5000,
  77 |       });
  78 |     }
  79 |   }
  80 | 
  81 |   async expectErrorMessage() {
  82 |     const error = this.page.locator(this.errorMessage).first();
  83 |     await expect(error).toBeVisible({ timeout: 5000 });
  84 |     return error.textContent();
  85 |   }
  86 | 
  87 |   async fillEmail(email: string) {
  88 |     await this.page.locator(this.emailInput).first().fill(email);
  89 |   }
  90 | 
  91 |   async fillPassword(password: string) {
  92 |     await this.page.locator(this.passwordInput).first().fill(password);
  93 |   }
  94 | 
  95 |   async clickLogin() {
  96 |     await this.page.locator(this.loginButton).first().click();
  97 |   }
  98 | }
  99 | 
```