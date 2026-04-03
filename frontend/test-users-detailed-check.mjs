import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const consoleLogs = [];
  page.on('console', msg => {
    if (msg.type() === 'error' || msg.type() === 'warning') {
      consoleLogs.push({ type: msg.type(), text: msg.text() });
    }
  });

  // Login
  await page.goto('http://localhost:3000/login');
  await page.locator('input').first().fill('admin');
  await page.locator('input').nth(1).fill('admin');
  await page.locator('button').first().click();
  await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

  console.log('✅ Logged in');

  // Navigate to /users
  await page.goto('http://localhost:3000/users');
  await page.waitForTimeout(3000);

  const url = page.url();
  const title = await page.locator('h1, h2').first().textContent();
  const hasTable = await page.locator('table, [role="grid"]').count();
  const hasButtons = await page.locator('button').count();
  const bodyText = await page.textContent('body');
  
  console.log('');
  console.log('=== Detailed /users Check ===');
  console.log('URL:', url);
  console.log('Title:', title);
  console.log('Tables:', hasTable);
  console.log('Buttons:', hasButtons);
  console.log('Body text length:', bodyText ? bodyText.length : 0);
  
  console.log('');
  console.log('=== Console Errors ===');
  if (consoleLogs.length === 0) {
    console.log('✅ No errors');
  } else {
    consoleLogs.forEach(log => console.log(`${log.type}: ${log.text.substring(0, 200)}`));
  }

  // Try to interact with page
  console.log('');
  console.log('=== User Rows ===');
  const rows = await page.locator('tr, [data-testid="user-item"]');
  const rowCount = await rows.count();
  console.log('User table rows:', rowCount);

  if (rowCount > 0) {
    for (let i = 0; i < Math.min(3, rowCount); i++) {
      const text = await rows.nth(i).textContent();
      console.log(`Row ${i}:`, text.substring(0, 100));
    }
  }

  await browser.close();
})();
