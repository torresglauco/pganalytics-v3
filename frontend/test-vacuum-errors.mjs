import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const errors = [];
  const logs = [];

  page.on('console', msg => {
    const text = msg.text();
    if (msg.type() === 'error') {
      errors.push(text);
    } else if (msg.type() === 'log' || msg.type() === 'warn') {
      logs.push({ type: msg.type(), text });
    }
  });

  page.on('pageerror', err => {
    errors.push('Page error: ' + err.message);
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
    console.log('Navigating to Vacuum Advisor...');
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(2000);

    console.log('\n📋 Errors found:');
    if (errors.length === 0) {
      console.log('  None');
    } else {
      errors.forEach(err => console.log('  ❌', err.substring(0, 120)));
    }

    console.log('\n📋 Console messages:');
    logs.slice(0, 10).forEach(log => {
      console.log(`  [${log.type}] ${log.text.substring(0, 100)}`);
    });

    // Check page structure
    const mainContent = await page.locator('main').isVisible({ timeout: 500 }).catch(() => false);
    const header = await page.locator('header, [class*="header"]').first().isVisible({ timeout: 500 }).catch(() => false);
    const sidebar = await page.locator('aside, [class*="sidebar"]').first().isVisible({ timeout: 500 }).catch(() => false);

    console.log('\n📄 Page structure:');
    console.log('  Main content:', mainContent);
    console.log('  Header:', header);
    console.log('  Sidebar:', sidebar);

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
