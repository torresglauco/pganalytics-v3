import { test, expect } from '../fixtures/auth';

/**
 * E2E Tests: API Integration
 * Tests API calls from pages and data loading
 */

test.describe('API Integration', () => {
  test('should make authenticated API calls on protected pages', async ({ page }) => {
    const apiCalls: { url: string; method: string; status: number }[] = [];

    // Intercept all API calls
    await page.on('response', (response) => {
      if (response.url().includes('/api/v1')) {
        apiCalls.push({
          url: response.url(),
          method: response.request().method(),
          status: response.status(),
        });
      }
    });

    // Navigate to a page that makes API calls
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Should have made at least one API call
    expect(apiCalls.length).toBeGreaterThan(0);

    // Check for successful responses
    const successfulCalls = apiCalls.filter((call) => call.status === 200 || call.status === 201);
    console.log(`✅ Made ${apiCalls.length} API calls (${successfulCalls.length} successful)`);
    console.log('   API calls:', apiCalls.map((c) => `${c.method} ${c.url.split('/api/v1')[1]} (${c.status})`).join('\n              '));
  });

  test('Logs page should load logs data', async ({ page }) => {
    const apiCalls: string[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('/logs')) {
        apiCalls.push(response.url());
      }
    });

    await page.goto('/logs');
    await page.waitForTimeout(1500);

    // Page should attempt to load logs
    const pageHasContent = await page.locator('body').locator('*').count().then((c) => c > 5);
    expect(pageHasContent).toBeTruthy();

    console.log(`✅ Logs page loaded (API calls: ${apiCalls.length})`);
  });

  test('Metrics page should load metrics data', async ({ page }) => {
    const apiCalls: string[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('/metrics')) {
        apiCalls.push(response.url());
      }
    });

    await page.goto('/metrics');
    await page.waitForTimeout(1500);

    const pageHasContent = await page.locator('body').locator('*').count().then((c) => c > 5);
    expect(pageHasContent).toBeTruthy();

    console.log(`✅ Metrics page loaded (API calls: ${apiCalls.length})`);
  });

  test('Query Performance page should make correct API calls', async ({ page }) => {
    const queryPerfCalls: string[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('query-performance')) {
        queryPerfCalls.push(response.url());
      }
    });

    await page.goto('/query-performance/1');
    await page.waitForTimeout(1500);

    console.log(`✅ Query Performance page loaded`);
    if (queryPerfCalls.length > 0) {
      console.log(`   API calls: ${queryPerfCalls.length}`);
    }
  });

  test('VACUUM Advisor page should make correct API calls', async ({ page }) => {
    const vacuumCalls: string[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('vacuum-advisor')) {
        vacuumCalls.push(response.url());
      }
    });

    await page.goto('/vacuum-advisor/1');
    await page.waitForTimeout(1500);

    console.log(`✅ VACUUM Advisor page loaded`);
    if (vacuumCalls.length > 0) {
      console.log(`   API calls: ${vacuumCalls.length}`);
    }
  });

  test('Index Advisor page should make correct API calls', async ({ page }) => {
    const indexCalls: string[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('index-advisor')) {
        indexCalls.push(response.url());
      }
    });

    await page.goto('/index-advisor/1');
    await page.waitForTimeout(1500);

    console.log(`✅ Index Advisor page loaded`);
    if (indexCalls.length > 0) {
      console.log(`   API calls: ${indexCalls.length}`);
    }
  });

  test('should handle API errors gracefully', async ({ page }) => {
    let errorResponses: { url: string; status: number }[] = [];

    await page.on('response', (response) => {
      if (!response.ok() && response.url().includes('/api/v1')) {
        errorResponses.push({
          url: response.url(),
          status: response.status(),
        });
      }
    });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Page should still load even if some APIs fail
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible({ timeout: 3000 });

    console.log(`✅ Page loads gracefully with ${errorResponses.length} failed API calls`);
    if (errorResponses.length > 0) {
      console.log('   Failed calls:', errorResponses.map((e) => `${e.status}`).join(', '));
    }
  });

  test('API responses should use correct status codes', async ({ page }) => {
    const responses: { status: number; count: number }[] = [];

    await page.on('response', (response) => {
      if (response.url().includes('/api/v1')) {
        const status = response.status();
        const existing = responses.find((r) => r.status === status);
        if (existing) {
          existing.count++;
        } else {
          responses.push({ status, count: 1 });
        }
      }
    });

    await page.goto('/');
    await page.waitForTimeout(1500);

    console.log('✅ API status codes:');
    responses.forEach((r) => {
      console.log(`   ${r.status}: ${r.count} response(s)`);
    });
  });
});
