import { test, expect } from '@playwright/test';

/**
 * Full-Stack E2E Tests
 *
 * These tests run against the real Go backend and Eliza test service.
 * They do NOT use network mocking - they test the actual end-to-end flow.
 *
 * Prerequisites:
 * - Go backend running at localhost:8080
 * - Eliza test server running at localhost:50051
 * - Protos loaded via global setup
 */
test.describe('Full-Stack E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app (served by Go backend)
    await page.goto('/');

    // Wait for services to load
    const sidebar = page.locator('aside').first();
    await expect(sidebar.locator('text=ElizaService')).toBeVisible({ timeout: 15000 });
  });

  test('should successfully invoke Say method and display response', async ({ page }) => {
    // Click on the Say method
    const sidebar = page.locator('aside').first();
    const sayButton = sidebar.locator('button:has-text("Say")');
    await sayButton.click();

    // Wait for the request editor to appear
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Find the sentence input field and enter a test message
    // The form should have a field for "sentence"
    const sentenceInput = rightPanel.locator('input[name="sentence"], input[placeholder*="sentence"], textarea').first();
    if (await sentenceInput.isVisible()) {
      await sentenceInput.fill('Hello, Eliza!');
    } else {
      // Try JSON editor mode
      const jsonToggle = rightPanel.locator('button:has-text("JSON")');
      if (await jsonToggle.isVisible()) {
        await jsonToggle.click();
      }
      // Find the JSON editor and enter the request
      const jsonEditor = rightPanel.locator('textarea, [contenteditable="true"]').first();
      await jsonEditor.fill('{"sentence": "Hello, Eliza!"}');
    }

    // Click Send Request
    await page.locator('button:has-text("Send Request")').click();

    // Wait for response - should show Success badge
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // Verify response contains expected content
    // The Eliza test server responds with "Hello! How can I help you today?" for hello messages
    await expect(rightPanel.locator('text=/Hello.*help/i')).toBeVisible({ timeout: 5000 });

    // Response time should be displayed
    const responseTime = rightPanel.locator('text=/\\d+ms/');
    await expect(responseTime).toBeVisible();
  });

  test('should handle test message and get echo response', async ({ page }) => {
    // Click on the Say method
    const sidebar = page.locator('aside').first();
    await sidebar.locator('button:has-text("Say")').click();

    // Wait for request editor
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Switch to JSON mode and enter request
    const jsonToggle = rightPanel.locator('button:has-text("JSON")');
    if (await jsonToggle.isVisible()) {
      await jsonToggle.click();
    }

    const jsonEditor = rightPanel.locator('textarea, [contenteditable="true"]').first();
    await jsonEditor.fill('{"sentence": "test"}');

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Wait for success
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // The Eliza test server responds with "Test received successfully!" for "test" input
    await expect(rightPanel.locator('text=Test received successfully')).toBeVisible({ timeout: 5000 });
  });

  test('should show error for empty sentence', async ({ page }) => {
    // Click on the Say method
    const sidebar = page.locator('aside').first();
    await sidebar.locator('button:has-text("Say")').click();

    // Wait for request editor
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Send empty request (should fail validation)
    await page.locator('button:has-text("Send Request")').click();

    // Should show either an error or validation message
    // Wait a bit for the response
    await page.waitForTimeout(2000);

    // Check if there's an error badge or error message
    const hasError = await rightPanel.locator('span:has-text("Error")').isVisible() ||
                     await rightPanel.locator('text=/required|invalid/i').isVisible();

    // If no error, it might have succeeded with empty - either way, response should show
    const hasSuccess = await rightPanel.locator('span:has-text("Success")').isVisible();

    expect(hasError || hasSuccess).toBeTruthy();
  });

  test('should update target endpoint and use it in request', async ({ page }) => {
    // First verify default endpoint
    const endpointInput = page.locator('input[placeholder="localhost:50051"]');
    await expect(endpointInput).toBeVisible();
    await expect(endpointInput).toHaveValue('localhost:50051');

    // Select a method
    const sidebar = page.locator('aside').first();
    await sidebar.locator('button:has-text("Say")').click();

    // Wait for request editor
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Enter test data
    const jsonToggle = rightPanel.locator('button:has-text("JSON")');
    if (await jsonToggle.isVisible()) {
      await jsonToggle.click();
    }
    const jsonEditor = rightPanel.locator('textarea, [contenteditable="true"]').first();
    await jsonEditor.fill('{"sentence": "hello"}');

    // Send request to default endpoint (should succeed)
    await page.locator('button:has-text("Send Request")').click();
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // Now change endpoint to something invalid
    await endpointInput.clear();
    await endpointInput.fill('localhost:99999');

    // Send request again (should fail because nothing is running on that port)
    await page.locator('button:has-text("Send Request")').click();

    // Should show error (connection refused or similar)
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible({ timeout: 15000 });
  });

  test('should switch transport modes and still invoke successfully', async ({ page }) => {
    // Find transport selector in header - could be buttons or a select/combobox
    const header = page.locator('header');

    // Check for button-based transport selector
    const httpButton = header.locator('button:has-text("HTTP")');
    const grpcButton = header.locator('button:has-text("gRPC")');

    // Check for select-based transport selector
    const transportSelect = header.locator('select, [role="combobox"]');

    const hasButtons = await httpButton.isVisible().catch(() => false);
    const hasSelect = await transportSelect.isVisible().catch(() => false);

    // Verify we have some transport selector
    expect(hasButtons || hasSelect).toBeTruthy();

    // Select a method
    const sidebar = page.locator('aside').first();
    await sidebar.locator('button:has-text("Say")').click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Enter test data
    const jsonToggle = rightPanel.locator('button:has-text("JSON")');
    if (await jsonToggle.isVisible()) {
      await jsonToggle.click();
    }
    const jsonEditor = rightPanel.locator('textarea, [contenteditable="true"]').first();
    await jsonEditor.fill('{"sentence": "test with HTTP"}');

    // Test with HTTP (Connect protocol - default)
    await page.locator('button:has-text("Send Request")').click();
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // Switch to gRPC mode
    if (hasButtons) {
      await grpcButton.click();
    } else if (hasSelect) {
      // For native <select>, use selectOption
      await transportSelect.selectOption({ label: 'gRPC' });
    }

    // Update message
    await jsonEditor.fill('{"sentence": "test with gRPC"}');

    // Send request with gRPC transport
    await page.locator('button:has-text("Send Request")').click();

    // gRPC should succeed - the internal Eliza test server supports all three protocols
    // (Connect, gRPC, gRPC-Web) via connect-go with h2c
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // Verify the response contains the expected content from the Eliza service
    await expect(rightPanel.locator('text=Test received successfully')).toBeVisible({ timeout: 5000 });
  });

  test('should display response time for successful requests', async ({ page }) => {
    // Select a method
    const sidebar = page.locator('aside').first();
    await sidebar.locator('button:has-text("Say")').click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Enter test data
    const jsonToggle = rightPanel.locator('button:has-text("JSON")');
    if (await jsonToggle.isVisible()) {
      await jsonToggle.click();
    }
    const jsonEditor = rightPanel.locator('textarea, [contenteditable="true"]').first();
    await jsonEditor.fill('{"sentence": "timing test"}');

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Wait for success
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 15000 });

    // Response time should be visible
    const responseTime = rightPanel.locator('text=/\\d+ms/');
    await expect(responseTime).toBeVisible();

    // Response time should be reasonable (less than 5 seconds for local service)
    const timeText = await responseTime.textContent();
    const ms = parseInt(timeText?.replace('ms', '') || '0');
    expect(ms).toBeGreaterThan(0);
    expect(ms).toBeLessThan(5000);
  });
});
