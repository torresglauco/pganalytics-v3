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

  console.log('✅ Login successful');

  // Navigate to /users
  await page.goto('http://localhost:3000/users');
  await page.waitForTimeout(2000);

  const url = page.url();
  const title = await page.locator('h1, h2').first().textContent();
  const hasTable = await page.locator('table, [role="grid"]').count();
  const hasForm = await page.locator('form').count();
  const bodyText = await page.textContent('body');
  
  console.log('');
  console.log('=== /users Page ===');
  console.log('✅ URL:', url);
  console.log('✅ Page title:', title);
  console.log('✅ Tables/grids:', hasTable);
  console.log('✅ Forms:', hasForm);
  console.log('✅ Has content:', bodyText && bodyText.length > 50);

  await browser.close();
})();
