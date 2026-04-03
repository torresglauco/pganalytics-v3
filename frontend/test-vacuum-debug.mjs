import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const allErrors = [];

  // Capture all errors
  page.on('console', msg => {
    if (msg.type() === 'error' || msg.text().includes('Error')) {
      allErrors.push({ type: 'console', text: msg.text().substring(0, 200) });
    }
  });

  page.on('response', response => {
    if (!response.ok()) {
      allErrors.push({ type: 'response', status: response.status(), url: response.url() });
    }
  });

  try {
    // Login
    await page.goto('http://localhost:3000/login');
    await page.locator('input').first().fill('admin');
    await page.locator('input').nth(1).fill('admin');
    await page.locator('button').first().click();
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

    console.log('✅ Logged in\n');

    // Get token from storage
    const token = await page.evaluate(() => localStorage.getItem('auth_token'));
    console.log('Token present:', !!token ? 'Yes' : 'No');

    // Navigate to Vacuum Advisor
    console.log('\nNavigating to Vacuum Advisor...');
    allErrors.length = 0; // Reset errors
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(2000);

    console.log('\n📋 API Response Errors:');
    const apiErrors = allErrors.filter(e => e.type === 'response' && e.url.includes('/api'));
    if (apiErrors.length === 0) {
      console.log('  None');
    } else {
      apiErrors.forEach(err => {
        console.log(`  Status ${err.status}: ${err.url.substring(30)}`);
      });
    }

    console.log('\n📋 JavaScript Errors:');
    const jsErrors = allErrors.filter(e => e.type === 'console');
    if (jsErrors.length === 0) {
      console.log('  None');
    } else {
      jsErrors.forEach(err => console.log(`  ${err.text}`));
    }

    // Check for the actual rendered structure
    console.log('\n📄 HTML Structure check:');
    const html = await page.content();
    console.log('  Contains <main>:', html.includes('<main'));
    console.log('  Contains MainLayout div:', html.includes('flex h-screen'));
    console.log('  Contains Sidebar comp:', html.includes('sidebar') || html.includes('nav'));
    console.log('  Contains VACUUM Advisor title:', html.includes('VACUUM Advisor'));

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
