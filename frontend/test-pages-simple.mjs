import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  // Login
  await page.goto('http://localhost:3000/login');
  await page.locator('input').first().fill('admin');
  await page.locator('input').nth(1).fill('admin');
  await page.locator('button').first().click();
  await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

  console.log('✅ Logged in\n');

  const testPages = [
    { path: '/query-performance/1', name: 'Query Performance' },
    { path: '/log-analysis/1', name: 'Log Analysis' },
    { path: '/vacuum-advisor/1', name: 'Vacuum Advisor' },
  ];

  for (const testPage of testPages) {
    await page.goto(`http://localhost:3000${testPage.path}`);
    await page.waitForTimeout(2000);

    const title = await page.locator('h1, h2').first().textContent().catch(() => null);
    const bodyText = await page.textContent('body').catch(() => '');
    const hasError = bodyText.includes('Error') || bodyText.includes('error') || bodyText.includes('failed');
    const hasContent = bodyText.length > 100;
    
    console.log(`📄 ${testPage.name}:`);
    console.log(`   URL: ${page.url()}`);
    console.log(`   Title: ${title}`);
    console.log(`   Has content: ${hasContent}`);
    console.log(`   Has error: ${hasError}\n`);
  }

  await browser.close();
})();
