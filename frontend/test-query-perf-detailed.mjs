import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const consoleLogs = [];
  page.on('console', msg => {
    if (msg.type() === 'error' || msg.type() === 'log') {
      consoleLogs.push({ type: msg.type(), text: msg.text().substring(0, 150) });
    }
  });

  // Login
  await page.goto('http://localhost:3000/login');
  await page.locator('input').first().fill('admin');
  await page.locator('input').nth(1).fill('admin');
  await page.locator('button').first().click();
  await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

  // Navigate to Query Performance
  await page.goto('http://localhost:3000/query-performance/1');
  await page.waitForTimeout(3000);

  const pageText = await page.textContent('body');
  
  console.log('Page contains:');
  if (pageText.includes('Error:')) {
    const match = pageText.match(/Error: [^\\n]+/);
    console.log('  Error message:', match ? match[0] : 'Unknown');
  }
  if (pageText.includes('No query data')) {
    console.log('  Empty state: Yes');
  }
  if (pageText.includes('Query Performance')) {
    console.log('  Title: Yes');
  }

  console.log('\\nConsole logs:');
  consoleLogs.slice(0, 5).forEach(log => {
    console.log(`  ${log.type}: ${log.text}`);
  });

  await browser.close();
})();
