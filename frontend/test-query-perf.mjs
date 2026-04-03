import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const networkErrors = [];
  page.on('response', response => {
    if (response.url().includes('/api/') && response.status() >= 400) {
      networkErrors.push({
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

  // Navigate to Query Performance
  await page.goto('http://localhost:3000/query-performance/1');
  await page.waitForTimeout(3000);

  const url = page.url();
  const title = await page.locator('h1, h2').first().textContent();
  const errorMsg = await page.locator('[data-testid="error"], .alert-danger, p').filter({ hasText: /failed|error/i }).first().textContent();
  
  console.log('');
  console.log('=== Query Performance Check ===');
  console.log('URL:', url);
  console.log('Title:', title);
  console.log('Error message:', errorMsg);
  
  console.log('');
  console.log('=== Network Errors ===');
  networkErrors.forEach(req => {
    console.log(`${req.method} ${req.url.split('?')[0]} -> ${req.status}`);
  });

  await browser.close();
})();
