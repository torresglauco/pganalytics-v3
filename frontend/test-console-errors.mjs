import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  // Capture console messages
  const consoleLogs = [];
  page.on('console', msg => {
    consoleLogs.push({ type: msg.type(), text: msg.text() });
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
  await page.waitForTimeout(2000);

  console.log('');
  console.log('=== Console Messages ===');
  consoleLogs.forEach(log => {
    if (log.type === 'error' || log.type === 'warning') {
      console.log(`${log.type.toUpperCase()}: ${log.text}`);
    }
  });

  if (consoleLogs.filter(l => l.type === 'error').length === 0) {
    console.log('No errors in console');
  }

  // Check MainLayout
  const hasMainLayout = await page.locator('[data-testid="main-layout"], main, .main-layout').count();
  console.log('');
  console.log('=== Layout Elements ===');
  console.log('Main layout elements:', hasMainLayout);

  // Try to find PageWrapper
  const allDivs = await page.locator('div').count();
  console.log('Total divs:', allDivs);

  await browser.close();
})();
