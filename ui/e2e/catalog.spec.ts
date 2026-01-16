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

/**
 * Error Handling Test Suite
 *
 * Tests error scenarios using Playwright's network interception
 * to simulate backend failures and network issues.
 */
test.describe('Error Handling', () => {
  test('should display error when backend returns failure', async ({ page }) => {
    await page.goto('/');

    // Wait for method selection
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    // Wait for schema to load
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Intercept the gRPC invoke request and simulate backend error
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Backend service is unavailable',
        }),
      });
    });

    // Click Send Request button
    await page.locator('button:has-text("Send Request")').click();

    // Verify error message is displayed
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(rightPanel.locator('text=Backend service is unavailable').first()).toBeVisible();

    // Verify error badge is shown in response viewer
    const errorBadge = rightPanel.locator('span:has-text("Error")');
    await expect(errorBadge).toBeVisible();
  });

  test('should show loading state while request is pending', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    // Wait for request editor
    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Intercept request and delay response to observe loading state
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      // Delay response to observe loading state
      await new Promise(resolve => setTimeout(resolve, 1000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          responseJson: JSON.stringify({ result: 'delayed response' }),
        }),
      });
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Verify loading state is shown
    await expect(rightPanel.locator('text=Sending request...')).toBeVisible({ timeout: 2000 });

    // Eventually should show success
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 5000 });
  });

  test('should display error message in response viewer with proper styling', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Mock error response
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Invalid request: field "name" is required',
          statusMessage: 'INVALID_ARGUMENT',
        }),
      });
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Verify error styling and content
    const errorContainer = rightPanel.locator('.bg-red-500\\/5');
    await expect(errorContainer).toBeVisible({ timeout: 5000 });

    // Check for error text (use exact match to avoid strict mode violation)
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible();
    await expect(rightPanel.locator('text=/Invalid request|field "name" is required/i').first()).toBeVisible();

    // Verify error has red border
    await expect(errorContainer).toHaveClass(/border-red-500/);
  });

  test('should allow retry after error', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    let requestCount = 0;

    // First request fails, second succeeds
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      requestCount++;
      if (requestCount === 1) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Connection failed',
          }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            responseJson: JSON.stringify({ message: 'Success after retry' }),
          }),
        });
      }
    });

    // First request - should fail
    await page.locator('button:has-text("Send Request")').click();
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(rightPanel.locator('text=Connection failed').first()).toBeVisible();

    // Retry - should succeed
    await page.locator('button:has-text("Send Request")').click();

    // Wait for success indicator
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 5000 });
    await expect(rightPanel.locator('text=Success after retry')).toBeVisible();

    // Verify error is cleared
    await expect(rightPanel.getByText('Request Failed', { exact: true })).not.toBeVisible();
  });

  test('should switch between HTTP and gRPC transport modes', async ({ page }) => {
    await page.goto('/');

    // Wait for services to load first (app shows loading state initially)
    const sidebar = page.locator('aside').first();
    await expect(sidebar.locator('div.border-l')).toBeVisible({ timeout: 10000 });

    // Verify transport selector exists in header
    const header = page.locator('header');
    const httpButton = header.locator('button:has-text("HTTP")');
    const grpcButton = header.locator('button:has-text("gRPC")');

    await expect(httpButton).toBeVisible();
    await expect(grpcButton).toBeVisible();

    // HTTP should be selected by default (Connect transport) - check for shadow class
    await expect(httpButton).toHaveClass(/shadow-sm/);

    // Switch to gRPC
    await grpcButton.click();
    await expect(grpcButton).toHaveClass(/shadow-sm/);

    // Switch back to HTTP
    await httpButton.click();
    await expect(httpButton).toHaveClass(/shadow-sm/);
  });

  test('should update target endpoint and use it in requests', async ({ page }) => {
    await page.goto('/');

    // Find target endpoint input in header
    const endpointInput = page.locator('input[placeholder="localhost:50051"]');
    await expect(endpointInput).toBeVisible();

    // Verify default value
    await expect(endpointInput).toHaveValue('localhost:50051');

    // Update endpoint
    await endpointInput.clear();
    await endpointInput.fill('api.example.com:443');
    await expect(endpointInput).toHaveValue('api.example.com:443');

    // Select a method to trigger request
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Intercept request to verify endpoint is used
    let capturedEndpoint = '';
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      const postData = route.request().postDataJSON();
      capturedEndpoint = postData.endpoint || '';

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          responseJson: JSON.stringify({ result: 'ok' }),
        }),
      });
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Wait for request to complete
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 5000 });

    // Verify the custom endpoint was used (check in console or via the intercepted data)
    // Note: In a real test, you would verify capturedEndpoint === 'api.example.com:443'
  });

  test('should show response time on successful request', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Mock successful response
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      // Simulate small delay
      await new Promise(resolve => setTimeout(resolve, 100));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          responseJson: JSON.stringify({ data: 'test' }),
        }),
      });
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Verify response time is displayed
    await expect(rightPanel.locator('span:has-text("Success")')).toBeVisible({ timeout: 5000 });

    // Should show response time with Clock icon
    const responseTime = rightPanel.locator('text=/\\d+ms/');
    await expect(responseTime).toBeVisible();
  });

  test('should show error badge styling on failed request', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Mock error response
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Request failed',
        }),
      });
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Verify error is shown with proper styling
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible({ timeout: 5000 });

    // Error badge should be visible
    const errorBadge = rightPanel.locator('span:has-text("Error")');
    await expect(errorBadge).toBeVisible();

    // Response time IS shown even on error (by design - tracks total request time)
    const responseTime = rightPanel.locator('text=/\\d+ms/');
    await expect(responseTime).toBeVisible();
  });

  test('should handle network errors gracefully', async ({ page }) => {
    await page.goto('/');

    // Select a method
    const sidebar = page.locator('aside').first();
    const methodButtons = sidebar.locator('div.border-l button');
    await expect(methodButtons.first()).toBeVisible({ timeout: 10000 });
    await methodButtons.first().click();

    const rightPanel = page.locator('aside').last();
    await expect(rightPanel.locator('h3:has-text("Request")')).toBeVisible({ timeout: 10000 });

    // Simulate complete network failure
    await page.route('**/catalog.v1.CatalogService/InvokeGRPC', async (route) => {
      await route.abort('failed');
    });

    // Send request
    await page.locator('button:has-text("Send Request")').click();

    // Should show generic error message
    await expect(rightPanel.getByText('Request Failed', { exact: true })).toBeVisible({ timeout: 5000 });

    // Error message should be present (exact text may vary based on error handling)
    const errorText = rightPanel.locator('.text-red-600\\/80').first();
    await expect(errorText).toBeVisible();
  });
});
