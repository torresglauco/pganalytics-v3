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

  console.log('✅ Logged in\n');

  // Test 1: Vacuum Advisor - sidebar issue
  console.log('🔍 Test 1: Vacuum Advisor - check sidebar');
  await page.goto('http://localhost:3000/vacuum-advisor/1');
  await page.waitForTimeout(2000);
  
  const sidebarVisible = await page.locator('[data-testid="sidebar"], nav, aside').isVisible({ timeout: 1000 }).catch(() => false);
  const mainLayout = await page.locator('[data-testid="main-layout"], main').isVisible({ timeout: 1000 }).catch(() => false);
  
  console.log('   Sidebar visible:', sidebarVisible);
  console.log('   Main layout visible:', mainLayout);

  // Test 2: Settings button - click settings
  console.log('\n🔍 Test 2: Settings button - top right corner');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(1000);
  
  const settingsBtn = await page.locator('button, a').filter({ hasText: /settings|gear|⚙️/i }).first();
  const settingsVisible = await settingsBtn.isVisible({ timeout: 1000 }).catch(() => false);
  
  console.log('   Settings button visible:', settingsVisible);
  
  if (settingsVisible) {
    const settingsBtnText = await settingsBtn.textContent();
    console.log('   Settings button text:', settingsBtnText);
    
    // Try clicking it
    await settingsBtn.click();
    await page.waitForTimeout(1500);
    
    const currentUrl = page.url();
    console.log('   URL after click:', currentUrl);
    console.log('   Navigated to settings:', currentUrl.includes('/settings'));
  }

  await browser.close();
})();
