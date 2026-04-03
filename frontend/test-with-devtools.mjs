import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const logs = [];
  const errors = [];

  page.on('console', msg => {
    logs.push({ type: msg.type(), text: msg.text().substring(0, 200) });
    if (msg.type() === 'error' || msg.type() === 'warn') {
      console.log(`[${msg.type()}] ${msg.text().substring(0, 150)}`);
    }
  });

  page.on('pageerror', err => {
    errors.push(err.message);
    console.log(`[PAGE ERROR] ${err.message}`);
  });

  try {
    // Login
    await page.goto('http://localhost:3000/login');
    await page.locator('input').first().fill('admin');
    await page.locator('input').nth(1).fill('admin');
    await page.locator('button').first().click();
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

    console.log('✅ Logged in\n');

    // Navigate to Vacuum Advisor
    console.log('Navigating to Vacuum Advisor (watch for errors)...\n');
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(3000);

    console.log('\n📋 Summary:');
    console.log('  Errors:', errors.length);
    console.log('  Console messages:', logs.length);

    // Check component structure
    const html = await page.content();
    const startOfRoot = html.indexOf('<div id="root">');
    const nextPart = html.substring(startOfRoot, startOfRoot + 300);
    console.log('\n📄 HTML starts with:', nextPart.substring(0, 200));

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
