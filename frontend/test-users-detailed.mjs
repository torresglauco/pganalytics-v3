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

  console.log('✅ Logged in');

  // Navigate to /users
  await page.goto('http://localhost:3000/users');
  await page.waitForTimeout(2000);

  const url = page.url();
  const pageContent = await page.content();
  const hasSettingsAdmin = pageContent.includes('SettingsAdmin') || pageContent.includes('settings');
  const bodyText = await page.textContent('body');
  
  console.log('');
  console.log('=== /users Page Analysis ===');
  console.log('Current URL:', url);
  console.log('Page has SettingsAdmin component:', hasSettingsAdmin);
  console.log('Page has visible text:', bodyText && bodyText.length > 20);
  console.log('Body text length:', bodyText ? bodyText.length : 0);
  console.log('First 200 chars of body:', bodyText ? bodyText.substring(0, 200) : 'NO TEXT');
  
  // Check for specific elements
  const hasTable = await page.locator('table').count();
  const hasForm = await page.locator('form').count();
  const hasButton = await page.locator('button').count();
  
  console.log('');
  console.log('Elements found:');
  console.log('- Tables:', hasTable);
  console.log('- Forms:', hasForm);
  console.log('- Buttons:', hasButton);
  
  // Check if page is blank/loading
  const allElements = await page.locator('*').count();
  console.log('- Total elements on page:', allElements);

  await browser.close();
})();
