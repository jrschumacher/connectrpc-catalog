# ConnectRPC Catalog UI - E2E Tests

End-to-end tests for the ConnectRPC Catalog UI using Playwright.

## Prerequisites

Before running E2E tests, ensure:

1. **Backend server is running** at `localhost:8080`
2. **At least one service is registered** in the catalog
3. **Node dependencies are installed**: `npm install`
4. **Playwright browsers are installed**: `npx playwright install`

## Installation

```bash
# From the ui directory
cd /Users/jschumacher/Projects/connectrpc-catalog/ui

# Install dependencies (includes Playwright)
npm install

# Install Playwright browsers
npx playwright install chromium
```

## Running Tests

### Basic Commands

```bash
# Run all tests (headless mode)
npm run test:e2e

# Run tests with Playwright UI (interactive mode)
npm run test:e2e:ui

# Run tests in headed mode (see browser)
npm run test:e2e:headed

# Run tests in debug mode (step through tests)
npm run test:e2e:debug
```

### Advanced Playwright Commands

```bash
# Run specific test file
npx playwright test catalog.spec.ts

# Run specific test by name
npx playwright test -g "should load the application"

# Run tests in specific browser
npx playwright test --project=chromium

# Generate test report
npx playwright show-report

# Run tests with trace
npx playwright test --trace on
```

## Test Coverage

The E2E test suite covers:

### ✅ Application Loading
- Page header and title display
- Initial loading state
- Services list loading

### ✅ Service Browser
- Sidebar visibility
- Search functionality
- Service expansion/collapse
- Method list display

### ✅ Method Selection
- Method selection interaction
- Method details display
- Request editor rendering

### ✅ Request Editor
- Form mode display
- JSON mode toggle
- Input field rendering
- Send button functionality

### ✅ Empty States
- No method selected message
- No services found message

## Test Structure

```
ui/e2e/
├── catalog.spec.ts     # Main E2E test suite
└── README.md           # This file

ui/
├── playwright.config.ts # Playwright configuration
└── package.json         # Updated with test scripts
```

## Configuration

Tests are configured in `playwright.config.ts`:

- **Base URL**: `http://localhost:8080`
- **Browser**: Chromium (Desktop Chrome)
- **Timeout**: 30 seconds per test
- **Screenshots**: On failure only
- **Trace**: On first retry
- **Retries**: 2 on CI, 0 locally

## Troubleshooting

### Tests Fail with "page.goto: net::ERR_CONNECTION_REFUSED"

**Solution**: Ensure the backend server is running at `localhost:8080`

```bash
# Start the backend server (from project root)
go run cmd/server/main.go
```

### Tests Fail with "No services found"

**Solution**: Ensure at least one service is registered in the catalog. The catalog service itself should be available.

### Tests Timeout

**Solution**: Increase timeout in `playwright.config.ts`:

```typescript
use: {
  timeout: 60000, // 60 seconds
}
```

### Browser Installation Issues

**Solution**: Manually install browsers:

```bash
npx playwright install --with-deps chromium
```

## Extending Tests

To add new tests:

1. Open `ui/e2e/catalog.spec.ts`
2. Add new test cases within the `test.describe` block:

```typescript
test('should handle method invocation', async ({ page }) => {
  await page.goto('/');
  // Your test logic
});
```

3. Run tests: `npm run test:e2e`

## CI/CD Integration

For CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Install dependencies
  run: cd ui && npm ci

- name: Install Playwright browsers
  run: cd ui && npx playwright install --with-deps chromium

- name: Start backend server
  run: go run cmd/server/main.go &

- name: Wait for server
  run: sleep 5

- name: Run E2E tests
  run: cd ui && npm run test:e2e
```

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Playwright Test API](https://playwright.dev/docs/api/class-test)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
