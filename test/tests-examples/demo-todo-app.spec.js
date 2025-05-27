// @ts-check
import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  
});


test.describe('Post', () => {
  test('Submit a new post', async ({ page }) => {
	// Arrange 
	await page.goto('http://localhost:54324');
	const menuSubmit = await page.$('a[href="/posts/submit"]')
	menuSubmit?.click();

	// Act
	await page.fill('input[name="title"]', 'a new title');
	await page.fill('input[name="url"]', 'https://sprinteins.com');
	await page.fill('textarea[name="description"]', 'a cool website');

	await page.click('button[type="submit"]');

	// Assert

	// await new Promise((r) => setTimeout(r, 100_000))
  })
})