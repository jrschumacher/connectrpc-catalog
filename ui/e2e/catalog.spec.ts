import { test, expect } from '@playwright/test';

/**
 * ConnectRPC Inspect UI E2E Tests
 *
 * Prerequisites:
 * - Backend server must be running at localhost:8080 with protos loaded
 *
 * Run with: npm run test:e2e
 */

test.describe('ConnectRPC Inspect UI', () => {
  test('should load the application and display header', async ({ page }) => {
    await page.goto('/');

    // Verify the page title
    await expect(page).toHaveTitle('ConnectRPC Inspect');

    // Verify header is visible with new branding
    await expect(page.locator('header')).toBeVisible();
    await expect(page.locator('header').locator('text=Inspect')).toBeVisible();
  });

  test('should display services in the sidebar', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load
    await page.waitForSelector('aside', { timeout: 10000 });

    // Verify sidebar with search is visible
    const sidebar = page.locator('aside').first();
    await expect(sidebar).toBeVisible();

    // Verify search input exists with new placeholder
    await expect(sidebar.locator('input[placeholder="Search..."]')).toBeVisible();
  });

  test('should show service browser with method counts', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load
    const sidebar = page.locator('aside').first();
    await expect(sidebar).toBeVisible();

    // Should show service and method counts in footer
    await expect(sidebar.locator('text=/\\d+ service/')).toBeVisible({ timeout: 10000 });
    await expect(sidebar.locator('text=/\\d+ methods/')).toBeVisible();
  });

  test('should auto-expand services and show methods', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load and auto-expand
    const sidebar = page.locator('aside').first();

    // Services should auto-expand - look for method buttons in the border-l container
    // Wait for the expanded methods container to appear
    await expect(sidebar.locator('div.border-l')).toBeVisible({ timeout: 10000 });

    // Verify method buttons are visible inside the expanded section
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should select a method and display method details', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load and auto-expand
    const sidebar = page.locator('aside').first();

    // Wait for method buttons to be visible (services auto-expand)
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });

    // Click the first method
    await methodButtons.first().click();

    // Verify center column shows method details
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();

    // Should display method name header (h2)
    await expect(mainContent.locator('h2')).toBeVisible({ timeout: 5000 });

    // Should display HTTP POST badge
    await expect(mainContent.locator('text=POST')).toBeVisible();
  });

  test('should show request editor with Form/JSON toggle', async ({ page }) => {
    await page.goto('/');

    // Wait for method buttons and click one
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    // Wait for schema to load (indicated by Request header appearing in the right panel)
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Should have Form/JSON toggle buttons
    const formButton = rightPanel.locator('button:has-text("Form")');
    const jsonButton = rightPanel.locator('button:has-text("JSON")');

    await expect(formButton).toBeVisible({ timeout: 5000 });
    await expect(jsonButton).toBeVisible();

    // Click JSON mode
    await jsonButton.click();

    // Should show a textarea for JSON input
    await expect(rightPanel.locator('textarea')).toBeVisible();

    // Click Form mode
    await formButton.click();

    // Textarea should not be visible in form mode (form fields instead)
    await expect(rightPanel.locator('textarea')).not.toBeVisible();
  });

  test('should show Send Request button', async ({ page }) => {
    await page.goto('/');

    // Wait for method buttons and click one
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    // Wait for schema to load (indicated by Request header appearing)
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Verify Send Request button exists
    await expect(page.locator('button:has-text("Send Request")')).toBeVisible({ timeout: 5000 });
  });

  test('should display empty state when no method selected', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load
    await page.waitForSelector('aside', { timeout: 10000 });

    // Right panel should show empty state with "Select a method" message
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('text=/Select a method/i')).toBeVisible({ timeout: 5000 });
  });

  test('should allow searching for services', async ({ page }) => {
    await page.goto('/');

    const sidebar = page.locator('aside').first();

    // Wait for services to load
    await expect(sidebar.locator('div.border-l')).toBeVisible({ timeout: 10000 });

    // Type in the search box
    const searchInput = sidebar.locator('input[placeholder="Search..."]');
    await searchInput.fill('nonexistentservice123456');

    // Should show "No matches found" message
    await expect(sidebar.locator('text=No matches found')).toBeVisible({ timeout: 2000 });
  });

  test('should show three-column layout', async ({ page }) => {
    await page.goto('/');

    // Verify three-column structure
    const asides = page.locator('aside');
    const main = page.locator('main');

    // Left sidebar (services)
    await expect(asides.first()).toBeVisible();

    // Center (method details)
    await expect(main).toBeVisible();

    // Right sidebar (request/response)
    await expect(asides.last()).toBeVisible();
  });

  test('should take screenshot of UI with method selected', async ({ page }) => {
    await page.goto('/');

    // Wait for method buttons to be visible
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });

    // Click the first method
    await methodButtons.first().click();

    // Wait for schema to load
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Take screenshot
    await page.screenshot({
      path: 'e2e/screenshots/connectrpc-inspect.png',
      fullPage: true,
    });
  });
});
