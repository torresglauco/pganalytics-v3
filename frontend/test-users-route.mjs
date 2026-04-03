import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  // First, login via the login page properly
  console.log('1. Navigating to login...');
  await page.goto('http://localhost:3000/login');
  
  // Wait for login form to load
  await page.waitForSelector('input', { timeout: 5000 }).catch(() => console.log('No input found'));
  
  // Try to find and fill username input
  const usernameInput = await page.locator('input').first();
  const isVisible = await usernameInput.isVisible().catch(() => false);
  
  console.log('2. Username input visible:', isVisible);
  
  if (isVisible) {
    console.log('3. Filling username...');
    await usernameInput.fill('admin');
    
    const passwordInput = await page.locator('input').nth(1);
    console.log('4. Filling password...');
    await passwordInput.fill('admin');
    
    const submitBtn = await page.locator('button').first();
    console.log('5. Clicking submit...');
    await submitBtn.click();
    
    console.log('6. Waiting for redirect...');
    await page.waitForURL('http://localhost:3000/', { timeout: 10000 });
    
    console.log('✅ Login successful!');
    
    // Check localStorage
    const token = await page.evaluate(() => localStorage.getItem('auth_token'));
    console.log('7. Token in localStorage:', token ? '✅ YES' : '❌ NO');
    
    // Now navigate to /users
    console.log('8. Navigating to /users...');
    await page.goto('http://localhost:3000/users');
    
    const finalUrl = page.url();
    const title = await page.locator('h1, h2').first().textContent();
    
    console.log('');
    console.log('=== Final Result ===');
    console.log('Current URL:', finalUrl);
    console.log('Expected URL: http://localhost:3000/users');
    console.log('Matches expected:', finalUrl === 'http://localhost:3000/users');
    console.log('Page title:', title);
  }

  await browser.close();
})();
