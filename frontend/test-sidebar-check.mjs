import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // Login
    await page.goto('http://localhost:3000/login');
    await page.locator('input').first().fill('admin');
    await page.locator('input').nth(1).fill('admin');
    await page.locator('button').first().click();
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

    console.log('✅ Logged in\n');

    // Test 1: Dashboard (home) - should have sidebar
    console.log('Test 1: Dashboard');
    await page.goto('http://localhost:3000/');
    await page.waitForTimeout(1000);
    const dashboardSidebar = await page.locator('aside, nav, [class*="sidebar"]').first().isVisible({ timeout: 1000 }).catch(() => false);
    console.log('  Dashboard sidebar visible:', dashboardSidebar);

    // Test 2: Index Advisor
    console.log('\nTest 2: Index Advisor');
    await page.goto('http://localhost:3000/index-advisor/1');
    await page.waitForTimeout(1000);
    const indexSidebar = await page.locator('aside, nav, [class*="sidebar"]').first().isVisible({ timeout: 1000 }).catch(() => false);
    const indexTitle = await page.locator('h1').first().isVisible({ timeout: 1000 }).catch(() => false);
    console.log('  Index Advisor sidebar visible:', indexSidebar);
    console.log('  Index Advisor title visible:', indexTitle);

    // Test 3: Vacuum Advisor
    console.log('\nTest 3: Vacuum Advisor');
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(1000);
    const vacuumSidebar = await page.locator('aside, nav, [class*="sidebar"]').first().isVisible({ timeout: 1000 }).catch(() => false);
    const vacuumTitle = await page.locator('h1').first().isVisible({ timeout: 1000 }).catch(() => false);
    console.log('  Vacuum Advisor sidebar visible:', vacuumSidebar);
    console.log('  Vacuum Advisor title visible:', vacuumTitle);

    // Debug: Check what's in the page
    console.log('\n📋 Debug info:');
    const pageContent = await page.content();
    const hasSidebar = pageContent.includes('sidebar') || pageContent.includes('Sidebar') || pageContent.includes('nav');
    console.log('  Page contains sidebar elements:', hasSidebar);

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
