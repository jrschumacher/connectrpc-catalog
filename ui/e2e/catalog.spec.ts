import { test, expect } from '@playwright/test';

/**
 * ConnectRPC Catalog UI E2E Tests
 *
 * Prerequisites:
 * - Backend server must be running at localhost:8080
 * - Test automatically loads proto files before tests run
 *
 * Run with: npm run test:e2e
 */

test.describe('ConnectRPC Catalog UI', () => {
  // Load test protos before running tests
  test.beforeAll(async ({ request }) => {
    // Load the catalog's own proto files as test data
    // This gives us at least one service (CatalogService) to test with
    const response = await request.post('http://localhost:8080/catalog.v1.CatalogService/LoadProtos', {
      headers: {
        'Content-Type': 'application/json',
      },
      data: {
        proto_path: './proto',
      },
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.success).toBeTruthy();
    console.log(`Loaded ${data.serviceCount} services from ${data.fileCount} files`);
  });
  test('should load the application and display header', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Verify the page title and header are visible
    await expect(page.locator('h1')).toContainText('ConnectRPC Catalog');
    await expect(page.locator('text=Browse and test gRPC services')).toBeVisible();
  });

  test('should display services in the sidebar', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load (spinner should disappear)
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Verify sidebar is visible
    const sidebar = page.locator('aside');
    await expect(sidebar).toBeVisible();

    // Verify search input exists
    await expect(sidebar.locator('input[placeholder="Search services..."]')).toBeVisible();
  });

  test('should expand a service and display methods', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Find and click the first service to expand it
    const firstService = page.locator('aside button').first();
    await expect(firstService).toBeVisible();

    const serviceText = await firstService.textContent();
    console.log('First service:', serviceText);

    await firstService.click();

    // Verify at least one method button appears
    // Methods are nested within the service after expansion
    const methodButtons = page.locator('aside button').filter({ hasNotText: serviceText || '' });
    await expect(methodButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should select a method and display method details', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Expand the first service
    const firstService = page.locator('aside button').first();
    await firstService.click();

    // Click the first method
    await page.waitForTimeout(500); // Wait for expansion animation
    const firstMethod = page.locator('aside button').nth(1); // Second button should be the first method
    await firstMethod.click();

    // Verify main content area shows method details
    const mainContent = page.locator('main');

    // Should display request editor card
    await expect(mainContent.locator('text=Request').first()).toBeVisible({ timeout: 5000 });

    // Should display "Send Request" button
    await expect(mainContent.locator('button:has-text("Send Request")')).toBeVisible();
  });

  test('should allow JSON mode toggle in request editor', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Expand and select a method
    await page.locator('aside button').first().click();
    await page.waitForTimeout(500);
    await page.locator('aside button').nth(1).click();

    // Wait for method details to load
    await expect(page.locator('text=Request').first()).toBeVisible({ timeout: 5000 });

    // Find and click the JSON toggle button
    const jsonToggle = page.locator('button:has-text("JSON")');
    if (await jsonToggle.isVisible()) {
      await jsonToggle.click();

      // Verify JSON textarea appears
      await expect(page.locator('textarea#json-input')).toBeVisible();

      // Toggle back to form mode
      await page.locator('button:has-text("Form")').click();

      // Verify textarea is no longer visible
      await expect(page.locator('textarea#json-input')).not.toBeVisible();
    }
  });

  test('should display "No method selected" when no method is chosen', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Verify empty state message is displayed
    await expect(page.locator('text=No method selected')).toBeVisible();
    await expect(page.locator('text=Select a service and method from the sidebar to get started')).toBeVisible();
  });

  test('should allow searching for services', async ({ page }) => {
    // Navigate to the catalog UI
    await page.goto('/');

    // Wait for services to load
    await expect(page.locator('text=Loading services...')).not.toBeVisible({ timeout: 10000 });

    // Get the first service name
    const firstService = page.locator('aside button').first();
    const serviceName = await firstService.textContent();

    if (serviceName) {
      // Extract just the service name (before the method count)
      const searchTerm = serviceName.trim().split(/\s/)[0];

      // Type in the search box
      const searchInput = page.locator('input[placeholder="Search services..."]');
      await searchInput.fill(searchTerm);

      // Verify the service is still visible
      await expect(firstService).toBeVisible();

      // Type something that won't match
      await searchInput.fill('nonexistentservice123456');

      // Verify "No services found" message appears
      await expect(page.locator('text=No services found')).toBeVisible({ timeout: 2000 });
    }
  });
});
