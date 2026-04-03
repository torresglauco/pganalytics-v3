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

    // Test: Vacuum Advisor - sidebar should appear
    console.log('🔍 Test: Vacuum Advisor - sidebar and layout');
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(2000);

    const sidebarVisible = await page.locator('[data-testid="sidebar"], nav, aside').isVisible({ timeout: 1000 }).catch(() => false);
    const mainLayout = await page.locator('[data-testid="main-layout"], main, [class*="main"]').isVisible({ timeout: 1000 }).catch(() => false);
    const pageTitle = await page.locator('h1:has-text("VACUUM Advisor")').isVisible({ timeout: 1000 }).catch(() => false);

    console.log('   Sidebar visible:', sidebarVisible);
    console.log('   Main layout visible:', mainLayout);
    console.log('   Page title visible:', pageTitle);

    if (sidebarVisible && pageTitle) {
      console.log('\n✅ SUCCESS: Sidebar now appears on Vacuum Advisor page!');
    } else {
      console.log('\n❌ ISSUE: Sidebar or page title not visible');
    }

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
