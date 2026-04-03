import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // Clear cache
    const context = browser.contexts()[0];
    await context.clearCookies();

    // Login
    console.log('Logging in...');
    await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle' });
    await page.locator('input').first().fill('admin');
    await page.locator('input').nth(1).fill('admin');
    await page.locator('button').first().click();
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

    console.log('✅ Logged in\n');

    // Navigate to Vacuum Advisor with hard refresh
    console.log('Navigating to Vacuum Advisor with hard refresh...');
    await page.goto('http://localhost:3000/vacuum-advisor/1', { waitUntil: 'networkidle' });

    // Hard refresh (Ctrl+Shift+R equivalent)
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Check if sidebar appears
    const sidebarVisible = await page.locator('aside').isVisible({ timeout: 1000 }).catch(() => false);
    const mainVisible = await page.locator('main').isVisible({ timeout: 1000 }).catch(() => false);
    const titleVisible = await page.locator('h1').first().isVisible({ timeout: 1000 }).catch(() => false);

    console.log('After hard refresh:');
    console.log('  Sidebar visible:', sidebarVisible);
    console.log('  Main element visible:', mainVisible);
    console.log('  Title visible:', titleVisible);

    if (sidebarVisible && mainVisible) {
      console.log('\n✅ SUCCESS: Sidebar now appears!');
    } else {
      console.log('\n❌ Still not working');

      // Check HTML
      const html = await page.content();
      const hasLayout = html.includes('flex h-screen');
      console.log('  HTML has MainLayout structure:', hasLayout);
    }

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
