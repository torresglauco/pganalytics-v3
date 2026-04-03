import { chromium } from 'playwright';
import fs from 'fs';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // Login
    await page.goto('http://localhost:3000/login');
    await page.locator('input').first().fill('admin');
    await page.locator('input').nth(1).fill('admin');
    await page.locator('button').first().click();
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });

    console.log('✅ Logged in\n');

    // Check Index Advisor
    console.log('📄 Index Advisor page:');
    await page.goto('http://localhost:3000/index-advisor/1');
    await page.waitForTimeout(1000);
    const indexHtml = await page.content();
    const indexHasLayout = indexHtml.includes('flex h-screen');
    const indexHasMain = indexHtml.includes('<main');
    console.log('  Has MainLayout (flex h-screen):', indexHasLayout);
    console.log('  Has <main>:', indexHasMain);

    // Save snippet
    const indexLayoutSection = indexHtml.substring(
      indexHtml.indexOf('<body'),
      indexHtml.indexOf('<body') + 500
    );
    console.log('  Body start:', indexLayoutSection.substring(0, 200));

    // Check Vacuum Advisor
    console.log('\n📄 Vacuum Advisor page:');
    await page.goto('http://localhost:3000/vacuum-advisor/1');
    await page.waitForTimeout(1000);
    const vacuumHtml = await page.content();
    const vacuumHasLayout = vacuumHtml.includes('flex h-screen');
    const vacuumHasMain = vacuumHtml.includes('<main');
    console.log('  Has MainLayout (flex h-screen):', vacuumHasLayout);
    console.log('  Has <main>:', vacuumHasMain);

    // Check if VACUUM title is in the HTML
    const vacuumHasTitle = vacuumHtml.includes('VACUUM Advisor');
    console.log('  Has VACUUM Advisor title:', vacuumHasTitle);

    // Save snippets for inspection
    fs.writeFileSync('/tmp/index-advisor-snippet.html', indexHtml.substring(0, 2000));
    fs.writeFileSync('/tmp/vacuum-advisor-snippet.html', vacuumHtml.substring(0, 2000));
    console.log('\n📁 HTML snippets saved to /tmp/');

  } catch (error) {
    console.error('Test error:', error.message);
  } finally {
    await browser.close();
  }
})();
