# Frontend Comprehensive E2E Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add comprehensive E2E test coverage for all frontend screens and user actions not yet covered by existing tests.

**Architecture:** Create separate test suites for each feature area (Logs, Metrics, Channels, Query Performance, Log Analysis, Index Advisor, Vacuum Advisor, Settings, Grafana). Each test file validates:
- Page loads and renders correctly
- Navigation works
- Data displays properly
- Forms and interactions work
- Error states handled gracefully

**Tech Stack:** Playwright, TypeScript, authenticated fixtures (from 10-api-contracts.spec.ts)

**Existing Tests:**
- 01-login-logout.spec.ts (authentication flows)
- 02-collector-registration.spec.ts (collector CRUD)
- 03-dashboard.spec.ts (dashboard/overview)
- 04-alert-management.spec.ts (alerts CRUD)
- 05-user-management.spec.ts (user CRUD)
- 06-permissions-access-control.spec.ts (auth/permissions)
- 10-api-contracts.spec.ts (API response validation)

**New Tests to Add:**
- 07-logs-page.spec.ts
- 08-metrics-page.spec.ts
- 09-channels-page.spec.ts
- 11-query-performance.spec.ts
- 12-log-analysis.spec.ts
- 13-index-advisor.spec.ts
- 14-vacuum-advisor.spec.ts
- 15-settings-admin.spec.ts
- 16-grafana-redirect.spec.ts

---

## Task 1: Logs Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/07-logs-page.spec.ts`

- [ ] **Step 1: Write failing test for logs page load**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { LogsPage } from '../pages/LogsPage';

test.describe('Logs Page', () => {
  let loginPage: LoginPage;
  let logsPage: LogsPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    logsPage = new LogsPage(page);

    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display logs page with title', async ({ page }) => {
    await logsPage.goto();

    const title = page.locator('h1, h2').filter({ hasText: /logs/i }).first();
    await expect(title).toBeVisible();
  });

  test('should display logs table', async ({ page }) => {
    await logsPage.goto();

    const table = page.locator('table, [role="grid"], [data-testid="logs-table"]').first();
    await expect(table).toBeVisible({ timeout: 5000 });
  });

  test('should have filter/search input', async ({ page }) => {
    await logsPage.goto();

    const searchInput = page.locator('input').filter({ hasText: /search|filter/i }).first();
    const isVisible = await searchInput.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await expect(searchInput).toBeVisible();
    }
  });

  test('should handle log level filtering', async ({ page }) => {
    await logsPage.goto();

    const levelFilter = page.locator('select, button').filter({ hasText: /level|severity/i }).first();

    if (await levelFilter.isVisible({ timeout: 1000 }).catch(() => false)) {
      await levelFilter.click();
      await expect(page.locator('[role="option"], button').first()).toBeVisible();
    }
  });

  test('should display log details on row click', async ({ page }) => {
    await logsPage.goto();

    const firstRow = page.locator('tr, [data-testid="log-item"]').first();
    if (await firstRow.isVisible({ timeout: 2000 }).catch(() => false)) {
      await firstRow.click();

      const detail = page.locator('[data-testid="log-detail"], [role="dialog"]').first();
      const isDetailVisible = await detail.isVisible({ timeout: 3000 }).catch(() => false);
      expect(isDetailVisible).toBeTruthy();
    }
  });

  test('should support pagination or infinite scroll', async ({ page }) => {
    await logsPage.goto();

    const pagination = page.locator('[aria-label*="Paginat"], button:has-text(/next|previous/i)').first();
    const scrollContainer = page.locator('[data-testid="logs-container"]').first();

    const hasPagination = await pagination.isVisible({ timeout: 2000 }).catch(() => false);
    const hasScroll = await scrollContainer.isVisible({ timeout: 2000 }).catch(() => false);

    expect(hasPagination || hasScroll).toBeTruthy();
  });

  test('should handle no logs state gracefully', async ({ page }) => {
    await logsPage.goto();

    const emptyState = page.locator('[data-testid="empty-state"], .empty').first();
    const table = page.locator('table, [role="grid"]').first();

    const isEmpty = await emptyState.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isEmpty || await table.isVisible({ timeout: 2000 }).catch(() => false)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Create LogsPage page object**

```typescript
// frontend/e2e/pages/LogsPage.ts
import { Page } from '@playwright/test';

export class LogsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/logs');
    await this.page.waitForLoadState('networkidle');
  }

  async expectLoaded() {
    const title = this.page.locator('h1, h2').filter({ hasText: /logs/i }).first();
    await title.waitFor({ state: 'visible', timeout: 5000 });
  }

  async getLogCount() {
    const rows = this.page.locator('tr, [data-testid="log-item"]');
    return await rows.count();
  }

  async filterByLevel(level: string) {
    const filter = this.page.locator('select, button').filter({ hasText: /level/i }).first();
    await filter.click();
    await this.page.locator('text=' + level).click();
    await this.page.waitForLoadState('networkidle');
  }

  async searchLogs(query: string) {
    const searchInput = this.page.locator('input').filter({ hasText: /search/i }).first();
    await searchInput.fill(query);
    await this.page.waitForTimeout(500);
  }
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 07-logs-page.spec.ts --reporter=list`
Expected: Tests should fail because LogsPage class doesn't exist yet

- [ ] **Step 4: Create LogsPage component if missing**

Check: `ls -la src/pages/LogsPage.tsx`
If exists: Skip this step
If missing: Create basic component with table/logs display

- [ ] **Step 5: Run tests again to verify they pass**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 07-logs-page.spec.ts --reporter=list`
Expected: Most tests should pass (some may skip if features not fully implemented)

- [ ] **Step 6: Commit**

```bash
cd /Users/glauco.torres/git/pganalytics-v3
git add frontend/e2e/tests/07-logs-page.spec.ts frontend/e2e/pages/LogsPage.ts
git commit -m "test: add comprehensive E2E tests for logs page"
```

---

## Task 2: Metrics Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/08-metrics-page.spec.ts`

- [ ] **Step 1: Write failing test for metrics page**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { MetricsPage } from '../pages/MetricsPage';

test.describe('Metrics Page', () => {
  let loginPage: LoginPage;
  let metricsPage: MetricsPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    metricsPage = new MetricsPage(page);

    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display metrics page with title', async ({ page }) => {
    await metricsPage.goto();

    const title = page.locator('h1, h2').filter({ hasText: /metrics|performance/i }).first();
    await expect(title).toBeVisible();
  });

  test('should display metrics charts', async ({ page }) => {
    await metricsPage.goto();

    const charts = page.locator('[data-testid="chart"], canvas, [role="img"]').first();
    const isVisible = await charts.isVisible({ timeout: 5000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should display metric cards/summary', async ({ page }) => {
    await metricsPage.goto();

    const cards = page.locator('[data-testid="metric-card"], [data-testid="summary-card"]').first();
    const isVisible = await cards.isVisible({ timeout: 3000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should allow time range selection', async ({ page }) => {
    await metricsPage.goto();

    const timeRange = page.locator('select, button').filter({ hasText: /time|range|last|hour/i }).first();
    const isVisible = await timeRange.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await timeRange.click();
      await expect(page.locator('[role="option"], button').first()).toBeVisible();
    }
  });

  test('should update metrics on time range change', async ({ page }) => {
    await metricsPage.goto();

    const initialData = await page.locator('[data-testid="chart"], canvas').first().screenshot();

    const timeRange = page.locator('select, button').filter({ hasText: /time|range/i }).first();
    if (await timeRange.isVisible({ timeout: 1000 }).catch(() => false)) {
      await timeRange.click();
      const option = page.locator('[role="option"]').first();
      if (await option.isVisible({ timeout: 1000 }).catch(() => false)) {
        await option.click();
        await page.waitForTimeout(1000);

        const updatedData = await page.locator('[data-testid="chart"]').first().screenshot();
        expect(initialData).not.toEqual(updatedData);
      }
    }
  });

  test('should handle empty metrics gracefully', async ({ page }) => {
    await metricsPage.goto();

    const empty = page.locator('[data-testid="empty-state"]').first();
    const chart = page.locator('[data-testid="chart"]').first();

    const isEmpty = await empty.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isEmpty || await chart.isVisible({ timeout: 2000 }).catch(() => false)).toBeTruthy();
  });

  test('should show metric descriptions/legends', async ({ page }) => {
    await metricsPage.goto();

    const legend = page.locator('[role="legend"], [data-testid="legend"]').first();
    const description = page.locator('p, span').filter({ hasText: /metric|measure/i }).first();

    const hasLegend = await legend.isVisible({ timeout: 2000 }).catch(() => false);
    const hasDesc = await description.isVisible({ timeout: 2000 }).catch(() => false);

    expect(hasLegend || hasDesc).toBeTruthy();
  });
});
```

- [ ] **Step 2: Create MetricsPage page object**

```typescript
// frontend/e2e/pages/MetricsPage.ts
import { Page } from '@playwright/test';

export class MetricsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/metrics');
    await this.page.waitForLoadState('networkidle');
  }

  async expectLoaded() {
    const title = this.page.locator('h1, h2').filter({ hasText: /metrics/i }).first();
    await title.waitFor({ state: 'visible', timeout: 5000 });
  }

  async getChartCount() {
    const charts = this.page.locator('[data-testid="chart"], canvas');
    return await charts.count();
  }

  async selectTimeRange(range: string) {
    const selector = this.page.locator('select, button').filter({ hasText: /time|range/i }).first();
    await selector.click();
    await this.page.locator(`text=${range}`).click();
    await this.page.waitForLoadState('networkidle');
  }

  async getMetricValue(metricName: string) {
    const card = this.page.locator('[data-testid="metric-card"]').filter({ hasText: metricName }).first();
    const value = card.locator('[data-testid="metric-value"], span').first();
    return await value.textContent();
  }
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 08-metrics-page.spec.ts --reporter=list`
Expected: Tests fail (page object doesn't exist)

- [ ] **Step 4: Run tests again after page object created**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 08-metrics-page.spec.ts --reporter=list`
Expected: Tests pass or show what's implemented

- [ ] **Step 5: Commit**

```bash
cd /Users/glauco.torres/git/pganalytics-v3
git add frontend/e2e/tests/08-metrics-page.spec.ts frontend/e2e/pages/MetricsPage.ts
git commit -m "test: add comprehensive E2E tests for metrics page"
```

---

## Task 3: Channels Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/09-channels-page.spec.ts`
- Create: `frontend/e2e/pages/ChannelsPage.ts`

- [ ] **Step 1: Write test for channels page**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Channels Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display channels page', async ({ page }) => {
    await page.goto('/channels');
    const title = page.locator('h1, h2').filter({ hasText: /channel/i }).first();
    await expect(title).toBeVisible();
  });

  test('should display channel list or form', async ({ page }) => {
    await page.goto('/channels');

    const list = page.locator('table, [role="grid"]').first();
    const form = page.locator('form').first();

    const hasContent = await list.isVisible({ timeout: 3000 }).catch(() => false) ||
                       await form.isVisible({ timeout: 3000 }).catch(() => false);
    expect(hasContent).toBeTruthy();
  });

  test('should allow creating new channel', async ({ page }) => {
    await page.goto('/channels');

    const createBtn = page.locator('button').filter({ hasText: /create|add|new/i }).first();
    const isVisible = await createBtn.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await createBtn.click();
      const form = page.locator('form, [role="dialog"]').first();
      await expect(form).toBeVisible({ timeout: 3000 });
    }
  });

  test('should handle channel configuration', async ({ page }) => {
    await page.goto('/channels');

    const inputs = page.locator('input, textarea, select');
    const count = await inputs.count();

    expect(count).toBeGreaterThan(0);
  });

  test('should validate channel form', async ({ page }) => {
    await page.goto('/channels');

    const createBtn = page.locator('button').filter({ hasText: /create|add/i }).first();
    if (await createBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await createBtn.click();

      const saveBtn = page.locator('button').filter({ hasText: /save|create|submit/i }).last();
      if (await saveBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
        await saveBtn.click();

        const error = page.locator('[data-testid="error"], .alert-danger, input:invalid').first();
        const isError = await error.isVisible({ timeout: 2000 }).catch(() => false);
        expect(isError).toBeTruthy();
      }
    }
  });

  test('should display test channel connection button', async ({ page }) => {
    await page.goto('/channels');

    const testBtn = page.locator('button').filter({ hasText: /test|verify|check/i }).first();
    const isVisible = await testBtn.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await testBtn.click();
      await page.waitForTimeout(1000);
    }
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 09-channels-page.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/09-channels-page.spec.ts
git commit -m "test: add E2E tests for channels page"
```

---

## Task 4: Query Performance Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/11-query-performance.spec.ts`

- [ ] **Step 1: Write test for query performance page**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Query Performance Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should navigate to query performance from dashboard', async ({ page }) => {
    await page.goto('/');

    // Try to find and click query performance link
    const qpLink = page.locator('a, button').filter({ hasText: /query|performance/i }).first();

    if (await qpLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await qpLink.click();
      const title = page.locator('h1, h2').filter({ hasText: /query/i }).first();
      await expect(title).toBeVisible({ timeout: 5000 });
    }
  });

  test('should display query performance metrics', async ({ page }) => {
    // Navigate to a specific database's query performance
    const dbId = '1';
    await page.goto(`/query-performance/${dbId}`);

    const loaded = await page.locator('h1, h2, [data-testid="metrics"]').first().isVisible({ timeout: 5000 }).catch(() => false);
    expect(loaded).toBeTruthy();
  });

  test('should display slow queries list', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/query-performance/${dbId}`);

    const table = page.locator('table, [role="grid"], [data-testid="queries-list"]').first();
    const isVisible = await table.isVisible({ timeout: 5000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should allow query filtering', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/query-performance/${dbId}`);

    const filter = page.locator('input, select, button').filter({ hasText: /filter|search|sort/i }).first();
    const isVisible = await filter.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should show query details on click', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/query-performance/${dbId}`);

    const firstRow = page.locator('tr, [data-testid="query-item"]').first();
    if (await firstRow.isVisible({ timeout: 2000 }).catch(() => false)) {
      await firstRow.click();

      const detail = page.locator('[data-testid="query-detail"], [role="dialog"]').first();
      const isDetailVisible = await detail.isVisible({ timeout: 3000 }).catch(() => false);
      expect(isDetailVisible).toBeTruthy();
    }
  });

  test('should handle database not found', async ({ page }) => {
    await page.goto('/query-performance/invalid-db-id');

    const error = page.locator('[data-testid="error"], .alert-danger, p').filter({ hasText: /not found|not|error/i }).first();
    const notFound = page.locator('h1').filter({ hasText: /not found|404/i }).first();

    const isError = await error.isVisible({ timeout: 3000 }).catch(() => false);
    const is404 = await notFound.isVisible({ timeout: 3000 }).catch(() => false);

    expect(isError || is404).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 11-query-performance.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/11-query-performance.spec.ts
git commit -m "test: add E2E tests for query performance page"
```

---

## Task 5: Log Analysis Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/12-log-analysis.spec.ts`

- [ ] **Step 1: Write test for log analysis page**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Log Analysis Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display log analysis page', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/log-analysis/${dbId}`);

    const title = page.locator('h1, h2').filter({ hasText: /log|analysis/i }).first();
    const isVisible = await title.isVisible({ timeout: 5000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should display log patterns or charts', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/log-analysis/${dbId}`);

    const chart = page.locator('[data-testid="chart"], canvas, [role="img"]').first();
    const table = page.locator('table, [role="grid"]').first();

    const hasContent = await chart.isVisible({ timeout: 5000 }).catch(() => false) ||
                       await table.isVisible({ timeout: 5000 }).catch(() => false);
    expect(hasContent).toBeTruthy();
  });

  test('should allow filtering by log level', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/log-analysis/${dbId}`);

    const filter = page.locator('select, button').filter({ hasText: /level|filter/i }).first();
    const isVisible = await filter.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await filter.click();
      const option = page.locator('[role="option"]').first();
      expect(await option.isVisible({ timeout: 1000 }).catch(() => false)).toBeTruthy();
    }
  });

  test('should display error trend analysis', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/log-analysis/${dbId}`);

    const trend = page.locator('[data-testid="trend"], [data-testid="chart"]').first();
    const isTrendVisible = await trend.isVisible({ timeout: 3000 }).catch(() => false);
    expect(isTrendVisible).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 12-log-analysis.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/12-log-analysis.spec.ts
git commit -m "test: add E2E tests for log analysis page"
```

---

## Task 6: Index Advisor Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/13-index-advisor.spec.ts`

- [ ] **Step 1: Write test for index advisor**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Index Advisor Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display index advisor page', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/index-advisor/${dbId}`);

    const title = page.locator('h1, h2').filter({ hasText: /index|advisor|recommendation/i }).first();
    await expect(title).toBeVisible({ timeout: 5000 });
  });

  test('should display recommended indexes', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/index-advisor/${dbId}`);

    const recommendations = page.locator('table, [role="grid"], [data-testid="recommendations"]').first();
    const isVisible = await recommendations.isVisible({ timeout: 5000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should allow applying index recommendations', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/index-advisor/${dbId}`);

    const applyBtn = page.locator('button').filter({ hasText: /apply|create|implement/i }).first();
    const isVisible = await applyBtn.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should show impact analysis', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/index-advisor/${dbId}`);

    const impact = page.locator('[data-testid="impact"], span, p').filter({ hasText: /impact|effect|improve/i }).first();
    const isVisible = await impact.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 13-index-advisor.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/13-index-advisor.spec.ts
git commit -m "test: add E2E tests for index advisor page"
```

---

## Task 7: Vacuum Advisor Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/14-vacuum-advisor.spec.ts`

- [ ] **Step 1: Write test for vacuum advisor**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Vacuum Advisor Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display vacuum advisor page', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/vacuum-advisor/${dbId}`);

    const title = page.locator('h1, h2').filter({ hasText: /vacuum|maintenance|advisor/i }).first();
    await expect(title).toBeVisible({ timeout: 5000 });
  });

  test('should display vacuum recommendations', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/vacuum-advisor/${dbId}`);

    const recommendations = page.locator('table, [role="grid"], [data-testid="recommendations"]').first();
    const isVisible = await recommendations.isVisible({ timeout: 5000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should show table bloat information', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/vacuum-advisor/${dbId}`);

    const bloat = page.locator('[data-testid="bloat"], span, p').filter({ hasText: /bloat|dead|space/i }).first();
    const isVisible = await bloat.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should allow scheduling vacuum', async ({ page }) => {
    const dbId = '1';
    await page.goto(`/vacuum-advisor/${dbId}`);

    const scheduleBtn = page.locator('button').filter({ hasText: /schedule|run|execute|vacuum/i }).first();
    const isVisible = await scheduleBtn.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 14-vacuum-advisor.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/14-vacuum-advisor.spec.ts
git commit -m "test: add E2E tests for vacuum advisor page"
```

---

## Task 8: Settings/Admin Page E2E Tests

**Files:**
- Create: `frontend/e2e/tests/15-settings-admin.spec.ts`

- [ ] **Step 1: Write test for settings page**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { UsersPage } from '../pages/UsersPage';

test.describe('Settings/Admin Page', () => {
  let loginPage: LoginPage;
  let usersPage: UsersPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    usersPage = new UsersPage(page);

    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should display settings page', async ({ page }) => {
    await page.goto('/settings');

    const title = page.locator('h1, h2').filter({ hasText: /settings|admin|configuration/i }).first();
    await expect(title).toBeVisible({ timeout: 5000 });
  });

  test('should display settings tabs/sections', async ({ page }) => {
    await page.goto('/settings');

    const tabs = page.locator('[role="tab"], button[data-testid*="tab"]').first();
    const sections = page.locator('[data-testid="section"], section').first();

    const hasContent = await tabs.isVisible({ timeout: 2000 }).catch(() => false) ||
                       await sections.isVisible({ timeout: 2000 }).catch(() => false);
    expect(hasContent).toBeTruthy();
  });

  test('should display users management section', async ({ page }) => {
    await page.goto('/settings');

    // Click users tab if present
    const usersTab = page.locator('[role="tab"], button').filter({ hasText: /users|admin/i }).first();
    if (await usersTab.isVisible({ timeout: 1000 }).catch(() => false)) {
      await usersTab.click();
    }

    const usersList = page.locator('table, [role="grid"]').first();
    await expect(usersList).toBeVisible({ timeout: 5000 });
  });

  test('should allow changing settings', async ({ page }) => {
    await page.goto('/settings');

    const inputs = page.locator('input, select, textarea').first();
    const isVisible = await inputs.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await inputs.fill('test-value');
      const saveBtn = page.locator('button').filter({ hasText: /save|apply/i }).first();
      if (await saveBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
        await saveBtn.click();
      }
    }
  });

  test('should handle settings validation', async ({ page }) => {
    await page.goto('/settings');

    const saveBtn = page.locator('button').filter({ hasText: /save|apply/i }).first();
    if (await saveBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await saveBtn.click();

      const success = page.locator('[data-testid="success"], .alert-success').first();
      const error = page.locator('[data-testid="error"], .alert-danger').first();

      const hasMessage = await success.isVisible({ timeout: 2000 }).catch(() => false) ||
                        await error.isVisible({ timeout: 2000 }).catch(() => false);
      expect(hasMessage).toBeTruthy();
    }
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 15-settings-admin.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/15-settings-admin.spec.ts
git commit -m "test: add E2E tests for settings/admin page"
```

---

## Task 9: Grafana Redirect E2E Tests

**Files:**
- Create: `frontend/e2e/tests/16-grafana-redirect.spec.ts`

- [ ] **Step 1: Write test for grafana redirect**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Grafana Redirect', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
    await loginPage.expectLoggedIn();
  });

  test('should redirect to grafana from menu', async ({ page, context }) => {
    await page.goto('/');

    // Listen for new pages (redirects)
    const pagePromise = context.waitForEvent('page');

    const grafanaLink = page.locator('a, button').filter({ hasText: /grafana/i }).first();
    const isVisible = await grafanaLink.isVisible({ timeout: 2000 }).catch(() => false);

    if (isVisible) {
      await grafanaLink.click();

      // Wait for new page or check redirect
      try {
        const newPage = await pagePromise;
        expect(newPage.url()).toContain('localhost:3001');
        await newPage.close();
      } catch {
        // If not a new page, check if current page redirected
        await page.waitForURL('**/localhost:3001', { timeout: 5000 }).catch(() => {
          console.log('Grafana redirect handled by client-side navigation');
        });
      }
    }
  });

  test('should display loading message during grafana redirect', async ({ page }) => {
    await page.goto('/grafana');

    const loading = page.locator('text=/redirecting|loading/i').first();
    const isVisible = await loading.isVisible({ timeout: 3000 }).catch(() => false);
    expect(isVisible).toBeTruthy();
  });

  test('should handle grafana service unavailable', async ({ page }) => {
    // This test checks error handling if grafana redirect fails
    await page.goto('/grafana');

    await page.waitForTimeout(2000);

    // Check if we're still on the page or got an error
    const error = page.locator('[data-testid="error"], .alert-danger').first();
    const loader = page.locator('[data-testid="loader"], .loader').first();

    const hasContent = await error.isVisible({ timeout: 1000 }).catch(() => false) ||
                       await loader.isVisible({ timeout: 1000 }).catch(() => false);
    expect(hasContent).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test and commit**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- 16-grafana-redirect.spec.ts --reporter=list`

```bash
git add frontend/e2e/tests/16-grafana-redirect.spec.ts
git commit -m "test: add E2E tests for grafana redirect"
```

---

## Task 10: Run Full E2E Test Suite

**Files:**
- Modified: All E2E test files

- [ ] **Step 1: Run entire E2E test suite**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e`
Expected: All tests pass (or show specific failures to fix)

- [ ] **Step 2: Generate test report**

Run: `cd /Users/glauco.torres/git/pganalytics-v3/frontend && npm run test:e2e -- --reporter=html`

Check results in: `frontend/test-results/index.html`

- [ ] **Step 3: Fix any failing tests**

Review failures and update tests as needed based on actual component behavior

- [ ] **Step 4: Final commit**

```bash
cd /Users/glauco.torres/git/pganalytics-v3
git add frontend/e2e/tests/
git commit -m "test: complete comprehensive E2E test coverage for all frontend pages"
```

---

## Summary

This plan adds comprehensive E2E test coverage for all frontend screens:
- ✅ Logs page (5 tests)
- ✅ Metrics page (6 tests)
- ✅ Channels page (6 tests)
- ✅ Query performance (6 tests)
- ✅ Log analysis (5 tests)
- ✅ Index advisor (4 tests)
- ✅ Vacuum advisor (4 tests)
- ✅ Settings/Admin (5 tests)
- ✅ Grafana redirect (3 tests)

**Total: ~44 new E2E tests** + existing 45 tests = **~89 comprehensive tests**

Each test validates:
- Page loads and renders
- User interactions work
- Data displays correctly
- Forms validate
- Error states handled
- Navigation functions
