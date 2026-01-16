# Server Handler Tests

This directory contains integration tests for the ConnectRPC handlers in `server.go`.

## Test Coverage

### Core Handler Tests

1. **TestLoadProtos_WithTestData** - Tests proto loading using programmatically created test data
2. **TestListServices** - Tests listing services after loading protos
3. **TestGetServiceSchema** - Tests retrieving service schema for a known service
4. **TestInvokeGRPC** - Tests InvokeGRPC validation (connection validation only)

### Edge Case Tests

- **TestLoadProtos_InvalidPath** - Error handling for invalid paths
- **TestListServices_Empty** - Listing services when registry is empty
- **TestGetServiceSchema_NotFound** - Error handling for unknown service
- **TestGetServiceSchema_EmptyName** - Validation for empty service name
- **TestInvokeGRPC_MissingEndpoint** - Validation for missing endpoint
- **TestInvokeGRPC_MissingService** - Validation for missing service
- **TestInvokeGRPC_MissingMethod** - Validation for missing method

### Utility Tests

- **TestServerValidation** - Tests ValidateSetup method
- **TestServerStats** - Tests GetStats method
- **TestClearRegistry** - Tests clearing the registry

## Test Approach

### Test Data Generation

The tests use programmatically created test data via `createTestFileDescriptorSet()` in `testdata_helper.go`. This approach:

- Avoids dependency on buf CLI for basic tests
- Provides a minimal but complete test service (test.v1.TestService)
- Enables fast, reliable unit tests without external dependencies

### Skipped Tests

- **TestLoadProtos** - Requires buf CLI and actual proto files. Skipped by default.
  - For full integration testing with real proto files, run with buf CLI installed
  - Use environment-specific integration test suites for end-to-end validation

## Running Tests

```bash
# Run all tests
go test ./internal/server/...

# Run with verbose output
go test ./internal/server/... -v

# Run a specific test
go test ./internal/server/... -run TestListServices
```

## Test Structure

Each test follows this pattern:
1. Create new server instance
2. Setup test data (register descriptors or prepare request)
3. Execute handler method
4. Validate response and behavior
5. Cleanup (via defer)

## Future Improvements

- Add integration tests with real buf CLI and proto files
- Add tests for streaming methods (when MVP support is added)
- Add performance/benchmark tests
- Add tests for concurrent operations
- Add tests for error recovery scenarios
