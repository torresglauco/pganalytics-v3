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

  console.log('✅ Logged in');

  // Navigate to Index Advisor
  await page.goto('http://localhost:3000/index-advisor/1');
  await page.waitForTimeout(3000);

  const url = page.url();
  const title = await page.locator('h1, h2').first().textContent();
  
  // Check for error
  const errorVisible = await page.locator('[data-testid="error"], .alert-danger, p:has-text(/Error|Failed/i)').first().isVisible({ timeout: 1000 }).catch(() => false);
  
  // Check for empty state message
  const emptyMsg = await page.locator('text=/No recommendations|no data|empty/i').isVisible({ timeout: 1000 }).catch(() => false);
  
  const allText = await page.textContent('body');
  
  console.log('');
  console.log('=== Index Advisor Result ===');
  console.log('URL:', url);
  console.log('Title:', title);
  console.log('Has error message:', errorVisible);
  console.log('Has empty state:', emptyMsg);
  console.log('Page loads successfully:', url.includes('index-advisor') && !errorVisible);

  await browser.close();
})();
