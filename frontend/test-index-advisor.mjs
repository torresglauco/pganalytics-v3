import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const consoleLogs = [];
  const networkRequests = [];

  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleLogs.push(msg.text());
    }
  });

  page.on('response', response => {
    if (response.url().includes('/api/') && response.status() >= 400) {
      networkRequests.push({
        url: response.url(),
        status: response.status(),
        method: response.request().method()
      });
    }
  });

  // Login
  await page.goto('http://localhost:3000/login');
  await page.locator('input').first().fill('admin');
  await page.locator('input').nth(1).fill('admin');
  await page.locator('button').first().click();
  await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

  console.log('✅ Logged in');

  // Navigate to Index Advisor
  await page.goto('http://localhost:3000/index-advisor/1');
  await page.waitForTimeout(3000);

  const url = page.url();
  const title = await page.locator('h1, h2').first().textContent();
  const errorMsg = await page.locator('[data-testid="error"], .alert-danger, p').filter({ hasText: /failed|error|unauthorized/i }).first().textContent();
  
  console.log('');
  console.log('=== Index Advisor Check ===');
  console.log('URL:', url);
  console.log('Title:', title);
  console.log('Error message:', errorMsg);
  
  console.log('');
  console.log('=== Network Errors ===');
  networkRequests.forEach(req => {
    console.log(`${req.method} ${req.url} -> ${req.status}`);
  });

  console.log('');
  console.log('=== Console Errors ===');
  consoleLogs.forEach(log => console.log(log.substring(0, 150)));

  await browser.close();
})();
