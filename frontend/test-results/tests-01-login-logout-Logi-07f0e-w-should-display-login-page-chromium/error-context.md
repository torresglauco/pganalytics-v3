# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests/01-login-logout.spec.ts >> Login/Logout Flow >> should display login page
- Location: e2e/tests/01-login-logout.spec.ts:14:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('input[type="email"], input[name="email"]').first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('input[type="email"], input[name="email"]').first()

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
  1   | import { test, expect } from '@playwright/test';
  2   | import { LoginPage } from '../pages/LoginPage';
  3   | import { DashboardPage } from '../pages/DashboardPage';
  4   | 
  5   | test.describe('Login/Logout Flow', () => {
  6   |   let loginPage: LoginPage;
  7   |   let dashboardPage: DashboardPage;
  8   | 
  9   |   test.beforeEach(async ({ page }) => {
  10  |     loginPage = new LoginPage(page);
  11  |     dashboardPage = new DashboardPage(page);
  12  |   });
  13  | 
  14  |   test('should display login page', async ({ page }) => {
  15  |     await loginPage.goto();
  16  | 
  17  |     // Check that login form is visible
  18  |     const emailInput = page.locator('input[type="email"], input[name="email"]').first();
  19  |     const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
  20  |     const loginBtn = page.locator('button').filter({ hasText: /Sign In|Login|Submit/i }).first();
  21  | 
> 22  |     await expect(emailInput).toBeVisible();
      |                              ^ Error: expect(locator).toBeVisible() failed
  23  |     await expect(passwordInput).toBeVisible();
  24  |     await expect(loginBtn).toBeVisible();
  25  |   });
  26  | 
  27  |   test('should login with valid credentials', async ({ page }) => {
  28  |     await loginPage.goto();
  29  | 
  30  |     // Use test credentials (these should be seeded in test database)
  31  |     await loginPage.login('demo@pganalytics.com', 'password123');
  32  | 
  33  |     // Verify redirect to dashboard
  34  |     await loginPage.expectLoggedIn();
  35  | 
  36  |     // Verify dashboard loads
  37  |     await dashboardPage.expectLoaded();
  38  |   });
  39  | 
  40  |   test('should show error with invalid credentials', async ({ page }) => {
  41  |     await loginPage.goto();
  42  | 
  43  |     // Fill wrong credentials
  44  |     await loginPage.fillEmail('wrong@example.com');
  45  |     await loginPage.fillPassword('wrongpassword');
  46  |     await loginPage.clickLogin();
  47  | 
  48  |     // Verify error message appears
  49  |     try {
  50  |       await loginPage.expectErrorMessage();
  51  |     } catch {
  52  |       // Some implementations might not show explicit error
  53  |       // Check that we're still on login page instead
  54  |       expect(page.url()).toContain('/login');
  55  |     }
  56  |   });
  57  | 
  58  |   test('should logout and redirect to login page', async ({ page }) => {
  59  |     await loginPage.goto();
  60  | 
  61  |     // First login
  62  |     await loginPage.login('demo@pganalytics.com', 'password123');
  63  |     await loginPage.expectLoggedIn();
  64  | 
  65  |     // Then logout
  66  |     await loginPage.logout();
  67  | 
  68  |     // Verify redirect to login
  69  |     await loginPage.expectLoggedOut();
  70  |   });
  71  | 
  72  |   test('should prevent unauthorized access to dashboard', async ({ page }) => {
  73  |     // Try to access dashboard without logging in
  74  |     await page.goto('/dashboard', { waitUntil: 'networkidle' });
  75  | 
  76  |     // Should be redirected to login
  77  |     expect(page.url()).toContain('/login');
  78  |   });
  79  | 
  80  |   test('should show loading state during login', async ({ page }) => {
  81  |     await loginPage.goto();
  82  | 
  83  |     // Fill credentials
  84  |     await loginPage.fillEmail('demo@pganalytics.com');
  85  |     await loginPage.fillPassword('password123');
  86  | 
  87  |     // Click login and check for loading state
  88  |     await loginPage.clickLogin();
  89  | 
  90  |     // Wait for either loading indicator or page transition
  91  |     // This is flexible as different implementations may vary
  92  |     await page.waitForLoadState('networkidle');
  93  | 
  94  |     // Should end up at dashboard
  95  |     expect(page.url()).toContain('/dashboard');
  96  |   });
  97  | 
  98  |   test('should maintain session after page reload', async ({ page }) => {
  99  |     await loginPage.goto();
  100 | 
  101 |     // Login
  102 |     await loginPage.login('demo@pganalytics.com', 'password123');
  103 |     await loginPage.expectLoggedIn();
  104 | 
  105 |     // Reload page
  106 |     await page.reload();
  107 | 
  108 |     // Should still be logged in
  109 |     await dashboardPage.expectLoaded();
  110 |     expect(page.url()).toContain('/dashboard');
  111 |   });
  112 | 
  113 |   test('should clear session on logout', async ({ page }) => {
  114 |     await loginPage.goto();
  115 | 
  116 |     // Login
  117 |     await loginPage.login('demo@pganalytics.com', 'password123');
  118 |     await loginPage.expectLoggedIn();
  119 | 
  120 |     // Logout
  121 |     await loginPage.logout();
  122 | 
```