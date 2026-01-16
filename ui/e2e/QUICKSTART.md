# E2E Testing Quick Start

Quick reference for running Playwright E2E tests.

## Setup (One Time)

```bash
# 1. Install dependencies
cd /Users/jschumacher/Projects/connectrpc-catalog/ui
npm install

# 2. Install Playwright browsers
npx playwright install chromium
```

## Running Tests

### Prerequisites
✅ Backend server must be running at `localhost:8080`

```bash
# Start backend (from project root)
go run cmd/server/main.go
```

### Run Tests

```bash
# From ui directory
cd /Users/jschumacher/Projects/connectrpc-catalog/ui

# Headless (CI mode)
npm run test:e2e

# Interactive UI mode (recommended for development)
npm run test:e2e:ui

# Headed mode (watch browser)
npm run test:e2e:headed

# Debug mode (step through)
npm run test:e2e:debug
```

## Test Scripts

| Command | Description |
|---------|-------------|
| `npm run test:e2e` | Run all tests headless |
| `npm run test:e2e:ui` | Interactive mode with UI |
| `npm run test:e2e:headed` | Show browser window |
| `npm run test:e2e:debug` | Debug with breakpoints |

## Common Issues

### ❌ "net::ERR_CONNECTION_REFUSED"
**Fix**: Start backend server first
```bash
go run cmd/server/main.go
```

### ❌ "No services found"
**Fix**: Ensure catalog service is registered (it should auto-register)

### ❌ Browser not found
**Fix**: Install browsers
```bash
npx playwright install chromium
```

## Test Coverage

- ✅ Application loading & header
- ✅ Service browser & search
- ✅ Method selection
- ✅ Request editor (form & JSON modes)
- ✅ Empty states

## Files Created

```
ui/
├── e2e/
│   ├── catalog.spec.ts       # Test suite
│   ├── README.md             # Full documentation
│   └── QUICKSTART.md         # This file
├── playwright.config.ts      # Playwright config
├── package.json              # Updated with scripts
└── .gitignore                # Updated with test artifacts
```

## Next Steps

1. Start backend: `go run cmd/server/main.go`
2. Run tests: `npm run test:e2e:ui`
3. Watch tests execute in Playwright UI
4. Extend tests in `ui/e2e/catalog.spec.ts`
