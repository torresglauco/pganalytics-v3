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
    console.log('📋 SIDEBAR VISIBILITY CHECK:\n');

    // Test each page
    const pages = [
      { name: 'Home', url: '/' },
      { name: 'Dashboard', url: '/' },
      { name: 'Query Performance', url: '/query-performance/1' },
      { name: 'Log Analysis', url: '/log-analysis/1' },
      { name: 'Index Advisor', url: '/index-advisor/1' },
      { name: 'VACUUM Advisor', url: '/vacuum-advisor/1' },
    ];

    for (const test of pages) {
      await page.goto(`http://localhost:3000${test.url}`);
      await page.waitForTimeout(1500);

      const hasSidebar = await page.locator('aside').isVisible({ timeout: 500 }).catch(() => false);
      const hasMain = await page.locator('main').isVisible({ timeout: 500 }).catch(() => false);
      const status = hasSidebar && hasMain ? '✅' : '❌';

      console.log(`${status} ${test.name.padEnd(20)} | Sidebar: ${hasSidebar} | Main: ${hasMain}`);
    }

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
