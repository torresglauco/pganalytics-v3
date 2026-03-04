# E2E Testing Implementation Plan - Playwright
**Date**: March 4, 2026
**Framework**: Playwright
**Coverage Target**: 40% → 80%+
**Timeline**: Mar 11-22 (2 weeks)
**Effort**: 16-20 hours

---

## 📋 Overview

This document outlines the plan to implement E2E (end-to-end) tests using Playwright, covering 7 critical user workflows in pgAnalytics v3.2.0.

### Current State
- E2E Coverage: ~40%
- Missing Scenarios: 7 critical flows
- Test Framework: Need Playwright setup

### Target State
- E2E Coverage: >80%
- All 7 scenarios covered
- Automated in CI/CD pipeline
- Runs on every PR and release

---

## 🎯 Critical Scenarios to Test

### 1. Login/Logout Flow
**User Story**: User logs in to pgAnalytics, accesses dashboard, then logs out

**Steps**:
```
1. Navigate to login page (http://localhost:3000/login)
2. Enter credentials (demo@pganalytics.com / password123)
3. Click "Sign In" button
4. Verify redirect to dashboard
5. Verify user name displayed in header
6. Click logout button
7. Verify redirect to login page
8. Verify can't access dashboard without token
```

**Expected Results**:
- Login succeeds within 2 seconds
- Dashboard loads with collector data
- Logout clears session and redirects
- Unauthorized access returns 401

---

### 2. Collector Registration
**User Story**: User registers a new PostgreSQL database to be monitored

**Steps**:
```
1. Navigate to /collectors/register
2. Enter database details:
   - Hostname: db-prod-01
   - Port: 5432
   - User: monitoring
   - Password: secure_password
3. Click "Test Connection"
4. Verify "Connection successful" message
5. Click "Register Collector"
6. Verify success message with collector ID
7. Verify collector appears in collector list
```

**Expected Results**:
- Connection test succeeds for valid credentials
- Collector registration creates collector in database
- Collector appears in UI immediately after registration
- TLS certificate auto-generated

---

### 3. Collector Management
**User Story**: User edits, pauses, and deletes collectors

**Scenarios**:
a) **Edit Collector**
```
1. Navigate to /collectors
2. Find collector "db-prod-01"
3. Click "Edit" button
4. Change interval from 60 to 30 seconds
5. Click "Save"
6. Verify interval updated
```

b) **Pause Collector**
```
1. Click "Pause" button on collector
2. Verify status changes to "paused"
3. Verify metrics stop being collected
4. Verify dashboard shows "paused" state
```

c) **Delete Collector**
```
1. Click "Delete" button
2. Confirm deletion dialog
3. Verify collector removed from list
4. Verify no error in console
```

---

### 4. Dashboard Visualization
**User Story**: User views dashboard with metrics and charts

**Steps**:
```
1. Login successfully
2. Navigate to dashboard
3. Verify dashboard loads within 3 seconds
4. Verify all 9 pre-built dashboards are available:
   - Query Performance
   - Replication Health
   - System Metrics
   - Infrastructure
   - Query Stats
   - Advanced Features
   - System Metrics Breakdown
   - Replication Collector
   - Multi-Collector Monitor
5. Click each dashboard
6. Verify charts load without errors
7. Verify no console errors (404s, etc)
```

**Expected Results**:
- Dashboard loads fast (<3s)
- Charts display data
- No JavaScript errors
- Responsive on mobile (optional)

---

### 5. Alert Creation & Management
**User Story**: User creates, edits, and deletes alert rules

**Steps**:
```
a) Create Alert:
1. Navigate to /alerts
2. Click "Create Alert"
3. Fill form:
   - Name: "High CPU Alert"
   - Metric: "cpu_usage"
   - Condition: > 80%
   - Action: Send email
4. Click "Create"
5. Verify alert appears in list

b) Edit Alert:
1. Click "Edit" on alert
2. Change threshold to 85%
3. Click "Save"
4. Verify change persisted

c) Delete Alert:
1. Click "Delete"
2. Confirm
3. Verify alert removed
```

---

### 6. User Management
**User Story**: Admin creates, edits, and deletes users

**Steps**:
```
a) Create User:
1. Navigate to /admin/users (admin only)
2. Click "Add User"
3. Fill form:
   - Email: newuser@company.com
   - Name: John Doe
   - Role: Analyst
4. Click "Create"
5. Verify user appears in list

b) Edit User:
1. Click "Edit" on user
2. Change role to Admin
3. Click "Save"

c) Delete User:
1. Click "Delete"
2. Confirm
3. Verify removed from list
```

---

### 7. Permission Testing
**User Story**: User without permissions can't access restricted areas

**Steps**:
```
a) Anonymous User (no token):
1. Navigate to /dashboard (requires auth)
2. Verify redirects to /login
3. Cannot access /api/v1/collectors

b) Limited User (analyst role):
1. Login as analyst
2. Can view dashboards
3. Cannot access /admin/users
4. Cannot delete collectors
5. Cannot edit system config

c) Admin User:
1. Login as admin
2. Can access all pages
3. Can modify all resources
```

---

## 🛠️ Implementation Steps

### Step 1: Setup Playwright (2-3 hours)

#### 1.1 Install Playwright
```bash
cd frontend

npm install -D @playwright/test

# Or using yarn
yarn add -D @playwright/test
```

#### 1.2 Create Configuration
**File**: `frontend/playwright.config.ts`

```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
```

#### 1.3 Create Test Directory Structure
```bash
mkdir -p frontend/e2e
mkdir -p frontend/e2e/fixtures
mkdir -p frontend/e2e/pages
mkdir -p frontend/e2e/tests

# Page Object Model helpers
touch frontend/e2e/pages/LoginPage.ts
touch frontend/e2e/pages/DashboardPage.ts
touch frontend/e2e/pages/CollectorPage.ts
```

---

### Step 2: Create Page Object Models (3-4 hours)

Page objects help organize tests and make them maintainable.

**File**: `frontend/e2e/pages/LoginPage.ts`

```typescript
import { Page, expect } from '@playwright/test';

export class LoginPage {
  constructor(readonly page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.page.fill('input[name="email"]', email);
    await this.page.fill('input[name="password"]', password);
    await this.page.click('button:has-text("Sign In")');

    // Wait for navigation
    await this.page.waitForURL('/dashboard');
  }

  async logout() {
    await this.page.click('[data-testid="user-menu"]');
    await this.page.click('button:has-text("Logout")');
    await this.page.waitForURL('/login');
  }

  async expectLoggedIn() {
    await expect(this.page).toHaveURL('/dashboard');
  }

  async expectLoggedOut() {
    await expect(this.page).toHaveURL('/login');
  }
}
```

**File**: `frontend/e2e/pages/DashboardPage.ts`

```typescript
import { Page, expect } from '@playwright/test';

export class DashboardPage {
  constructor(readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard');
  }

  async expectLoaded() {
    // Wait for main dashboard element
    await expect(this.page.locator('[data-testid="dashboard"]')).toBeVisible();
  }

  async selectDashboard(name: string) {
    await this.page.click(`[data-testid="dashboard-${name}"]`);
  }

  async expectDashboardLoaded(name: string) {
    await expect(
      this.page.locator(`text=${name}`)
    ).toBeVisible({ timeout: 3000 });
  }

  async getChartCount() {
    return await this.page.locator('[data-testid="chart"]').count();
  }
}
```

**File**: `frontend/e2e/pages/CollectorPage.ts`

```typescript
import { Page, expect } from '@playwright/test';

export class CollectorPage {
  constructor(readonly page: Page) {}

  async goto() {
    await this.page.goto('/collectors');
  }

  async clickRegister() {
    await this.page.click('button:has-text("Register Collector")');
  }

  async fillRegistrationForm(data: {
    hostname: string;
    port: number;
    database: string;
  }) {
    await this.page.fill('input[name="hostname"]', data.hostname);
    await this.page.fill('input[name="port"]', data.port.toString());
    await this.page.fill('input[name="database"]', data.database);
  }

  async testConnection() {
    await this.page.click('button:has-text("Test Connection")');
    await expect(
      this.page.locator('text=Connection successful')
    ).toBeVisible({ timeout: 5000 });
  }

  async registerCollector() {
    await this.page.click('button:has-text("Register")');
    await expect(
      this.page.locator('text=Collector registered')
    ).toBeVisible({ timeout: 5000 });
  }

  async expectCollectorInList(hostname: string) {
    await expect(
      this.page.locator(`text=${hostname}`)
    ).toBeVisible();
  }
}
```

---

### Step 3: Write Test Files (8-10 hours)

**File**: `frontend/e2e/tests/01-login-logout.spec.ts`

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Login/Logout Flow', () => {
  test('should login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);

    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Verify user name in header
    await expect(page.locator('[data-testid="user-name"]')).toContainText('Demo User');
  });

  test('should logout and return to login page', async ({ page }) => {
    const loginPage = new LoginPage(page);

    // First login
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
    await loginPage.expectLoggedIn();

    // Then logout
    await loginPage.logout();
    await loginPage.expectLoggedOut();
  });

  test('should prevent access without authentication', async ({ page }) => {
    await page.goto('/dashboard');

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should handle invalid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);

    await loginPage.goto();
    await loginPage.login('wrong@example.com', 'wrongpassword');

    // Should show error
    await expect(page.locator('text=Invalid credentials')).toBeVisible();

    // Should stay on login page
    await expect(page).toHaveURL('/login');
  });
});
```

**File**: `frontend/e2e/tests/02-collector-registration.spec.ts`

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { CollectorPage } from '../pages/CollectorPage';

test.describe('Collector Registration', () => {
  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@pganalytics.com', 'password123');
  });

  test('should register a new collector', async ({ page }) => {
    const collectorPage = new CollectorPage(page);

    await collectorPage.goto();
    await collectorPage.clickRegister();

    await collectorPage.fillRegistrationForm({
      hostname: 'test-db-01.example.com',
      port: 5432,
      database: 'postgres',
    });

    await collectorPage.testConnection();
    await collectorPage.registerCollector();

    await collectorPage.expectCollectorInList('test-db-01.example.com');
  });
});
```

---

### Step 4: Setup CI/CD (2-3 hours)

**File**: `.github/workflows/e2e-tests.yml`

```yaml
name: E2E Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    timeout-minutes: 60
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: pganalytics
          POSTGRES_DB: pganalytics
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      api:
        image: ghcr.io/torresglauco/pganalytics-api:latest
        ports:
          - 8080:8080
        env:
          DATABASE_URL: postgres://postgres:pganalytics@postgres:5432/pganalytics
          JWT_SECRET: test-secret

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: 'frontend/package-lock.json'

      - name: Install dependencies
        run: cd frontend && npm install

      - name: Install Playwright browsers
        run: cd frontend && npx playwright install --with-deps

      - name: Run E2E tests
        run: cd frontend && npm run test:e2e
        env:
          BASE_URL: http://localhost:3000

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: frontend/playwright-report/
          retention-days: 30
```

---

## 📊 Success Criteria

### Coverage
- [ ] 7 critical scenarios covered
- [ ] >80% E2E coverage achieved
- [ ] All tests passing on CI/CD

### Quality
- [ ] Tests run in <5 minutes
- [ ] Zero flakiness
- [ ] Clear error messages on failures
- [ ] Screenshots on failures

### Maintainability
- [ ] Page object model used
- [ ] No hard-coded waits (use proper waits)
- [ ] DRY code (no duplication)
- [ ] Well-documented

---

## 🚀 Running Tests

### Local Development
```bash
cd frontend

# Run all tests
npm run test:e2e

# Run specific test
npm run test:e2e -- --grep "Login"

# Run with UI
npm run test:e2e -- --ui

# Run headless
npm run test:e2e -- --headed=false
```

### CI/CD
```bash
# Tests run automatically on:
# - Push to main/develop
# - Pull requests

# View results:
# GitHub Actions → E2E Tests → Artifacts → playwright-report
```

---

## 📝 Package.json Updates

Add to `frontend/package.json`:

```json
{
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:debug": "playwright test --debug",
    "test:e2e:headed": "playwright test --headed"
  },
  "devDependencies": {
    "@playwright/test": "^1.40.0"
  }
}
```

---

## ⏱️ Timeline

| Week | Task | Hours | Status |
|------|------|-------|--------|
| W2 (Mar 11-15) | Setup Playwright + Page Objects | 5 | Not Started |
| W2 (Mar 11-15) | Tests 1-2 (Login + Collectors) | 6 | Not Started |
| W2 (Mar 11-15) | Tests 3-4 (Management + Dashboard) | 4 | Not Started |
| W3 (Mar 18-22) | Tests 5-7 (Alerts + Users + Perms) | 3 | Not Started |
| W3 (Mar 18-22) | CI/CD Integration | 2 | Not Started |
| **Total** | | **20** | |

---

## 🎯 Next Steps

1. ✅ Approve this plan
2. ⏳ Schedule work for Week 2-3
3. ⏳ Assign frontend developer
4. ⏳ Setup testing infrastructure
5. ⏳ Begin implementation

---

**Created**: March 4, 2026
**Status**: Ready for Implementation
**Assigned to**: Frontend Lead
**Start Date**: March 11, 2026 (Week 2)
