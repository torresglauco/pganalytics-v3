# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: 06-permissions-access-control.spec.ts >> Permissions and Access Control >> should prevent access to dashboard without login
- Location: e2e/tests/06-permissions-access-control.spec.ts:22:3

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
  5   | test.describe('Permissions and Access Control', () => {
  6   |   let loginPage: LoginPage;
  7   |   let dashboardPage: DashboardPage;
  8   | 
  9   |   test.beforeEach(async ({ page }) => {
  10  |     loginPage = new LoginPage(page);
  11  |     dashboardPage = new DashboardPage(page);
  12  |   });
  13  | 
  14  |   test('should redirect unauthenticated users to login', async ({ page }) => {
  15  |     // Try to access protected route without login
  16  |     await page.goto('/dashboard', { waitUntil: 'networkidle' });
  17  | 
  18  |     // Should redirect to login
  19  |     expect(page.url()).toContain('/login');
  20  |   });
  21  | 
  22  |   test('should prevent access to dashboard without login', async ({ page }) => {
  23  |     // Try direct access to dashboard
  24  |     await page.goto('/dashboard', { waitUntil: 'networkidle' });
  25  | 
  26  |     // Verify login form is displayed
  27  |     const emailInput = page.locator('input[type="email"], input[name="email"]').first();
  28  |     const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
  29  | 
> 30  |     await expect(emailInput).toBeVisible();
      |                              ^ Error: expect(locator).toBeVisible() failed
  31  |     await expect(passwordInput).toBeVisible();
  32  |   });
  33  | 
  34  |   test('should allow access after successful login', async ({ page }) => {
  35  |     // Navigate to login
  36  |     await loginPage.goto();
  37  | 
  38  |     // Login
  39  |     await loginPage.login('demo@pganalytics.com', 'password123');
  40  | 
  41  |     // Should be redirected to dashboard
  42  |     expect(page.url()).toContain('/dashboard');
  43  | 
  44  |     // Dashboard should be visible
  45  |     await dashboardPage.expectLoaded();
  46  |   });
  47  | 
  48  |   test('should restrict collectors page to authenticated users', async ({ page }) => {
  49  |     // Try to access collectors without login
  50  |     await page.goto('/collectors', { waitUntil: 'networkidle' });
  51  | 
  52  |     // Should redirect to login
  53  |     expect(page.url()).toContain('/login');
  54  |   });
  55  | 
  56  |   test('should restrict alerts page to authenticated users', async ({ page }) => {
  57  |     // Try to access alerts without login
  58  |     await page.goto('/alerts', { waitUntil: 'networkidle' });
  59  | 
  60  |     // Should redirect to login
  61  |     expect(page.url()).toContain('/login');
  62  |   });
  63  | 
  64  |   test('should restrict users page to authenticated users', async ({ page }) => {
  65  |     // Try to access users page without login
  66  |     await page.goto('/users', { waitUntil: 'networkidle' });
  67  | 
  68  |     // Should redirect to login
  69  |     expect(page.url()).toContain('/login');
  70  |   });
  71  | 
  72  |   test('should validate session on API calls', async ({ page }) => {
  73  |     // Login first
  74  |     await loginPage.goto();
  75  |     await loginPage.login('demo@pganalytics.com', 'password123');
  76  | 
  77  |     // Intercept API calls to verify auth headers
  78  |     let authHeaderPresent = false;
  79  |     page.on('request', (request) => {
  80  |       const authHeader = request.headers()['authorization'];
  81  |       if (authHeader && authHeader.startsWith('Bearer ')) {
  82  |         authHeaderPresent = true;
  83  |       }
  84  |     });
  85  | 
  86  |     // Navigate to page that makes API calls
  87  |     await dashboardPage.goto();
  88  | 
  89  |     // Wait for API calls to complete
  90  |     await page.waitForLoadState('networkidle');
  91  | 
  92  |     // Verify auth header was sent
  93  |     expect(authHeaderPresent).toBe(true);
  94  |   });
  95  | 
  96  |   test('should reject requests with invalid token', async ({ page }) => {
  97  |     // Manually set invalid token in localStorage
  98  |     await page.goto('/');
  99  |     await page.evaluate(() => {
  100 |       localStorage.setItem('authToken', 'invalid-token-xyz');
  101 |     });
  102 | 
  103 |     // Try to access protected route
  104 |     await page.goto('/dashboard', { waitUntil: 'networkidle' });
  105 | 
  106 |     // Should either redirect to login or show error
  107 |     const isOnLogin = page.url().includes('/login');
  108 |     const isError = page.locator('.alert-danger, [data-testid="error"]').first();
  109 | 
  110 |     expect(isOnLogin || await isError.isVisible({ timeout: 2000 }).catch(() => false)).toBe(true);
  111 |   });
  112 | 
  113 |   test('should handle expired session', async ({ page }) => {
  114 |     // Login
  115 |     await loginPage.goto();
  116 |     await loginPage.login('demo@pganalytics.com', 'password123');
  117 |     await loginPage.expectLoggedIn();
  118 | 
  119 |     // Simulate token expiration by removing it
  120 |     await page.evaluate(() => {
  121 |       localStorage.removeItem('authToken');
  122 |     });
  123 | 
  124 |     // Try to access protected page
  125 |     await page.goto('/dashboard', { waitUntil: 'networkidle' });
  126 | 
  127 |     // Should redirect to login
  128 |     expect(page.url()).toContain('/login');
  129 |   });
  130 | 
```