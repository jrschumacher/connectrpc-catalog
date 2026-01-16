# Testing Strategy

This document describes the testing approach for the ConnectRPC Catalog project.

## Test Categories

### 1. Integration Tests

**Location**: `internal/server/integration_test.go`

Integration tests verify end-to-end API functionality using real HTTP servers and Connect clients.

**Coverage**:
- ✅ LoadProtos API (local path, GitHub, Buf modules)
- ✅ ListServices API
- ✅ GetServiceSchema API
- ✅ InvokeGRPC API validation
- ✅ Error handling and edge cases

### 2. Unit Tests

**Location**: Various `*_test.go` files throughout the codebase

- `internal/loader/loader_test.go` - Proto loading logic
- `internal/registry/registry_test.go` - Service registry
- `internal/invoker/invoker_test.go` - gRPC invocation
- `internal/server/server_test.go` - Server logic

### 3. End-to-End Tests (Planned)

**Location**: `ui/src/__tests__/` (Playwright)

- UI component tests
- Full workflow tests (load protos → browse services → invoke methods)
- Cross-browser compatibility

## Running Tests

### Quick Commands

```bash
# Run all tests
go test ./...

# Run integration tests only
go test -v ./internal/server/... -run TestIntegration

# Run unit tests only
go test ./... -short

# Run with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Using the Test Script

```bash
# Run integration tests with formatted output
./test-integration.sh
```

## Integration Test Details

### Test Suite Overview

| Test | Purpose | Validates |
|------|---------|-----------|
| `TestIntegrationLoadProtos_LocalPath` | Load protos from filesystem | ✅ Success response, service/file counts |
| `TestIntegrationLoadProtos_InvalidPath` | Error handling for invalid paths | ✅ Error messages, graceful failure |
| `TestIntegrationListServices` | List loaded services | ✅ Service metadata, method info |
| `TestIntegrationGetServiceSchema` | Retrieve service schemas | ✅ Message schemas, field definitions |
| `TestIntegrationGetServiceSchema_InvalidService` | Error handling for missing service | ✅ Error messages |
| `TestIntegrationListServices_EmptyRegistry` | Empty registry behavior | ✅ Empty list handling |
| `TestIntegrationInvokeGRPC_MissingFields` | Request validation | ✅ Required field validation |
| `TestIntegrationMultipleLoadProtos` | Multiple proto loads | ✅ Registry accumulation |

### Test Execution Flow

```
┌─────────────────────────────────────────────────────┐
│  1. Create CatalogServer instance                   │
│  2. Set up HTTP test server with Connect handlers   │
│  3. Create Connect client pointing to test server   │
│  4. Execute API calls (LoadProtos, ListServices)    │
│  5. Verify responses and error handling             │
│  6. Cleanup resources                               │
└─────────────────────────────────────────────────────┘
```

## Test Data

### Proto Files

Integration tests use the project's own proto files located at:
- `proto/catalog/v1/catalog.proto`

The test helper `getTestProtoPath()` automatically locates proto files relative to the test location.

### Expected Test Outputs

When running integration tests against the catalog proto:
- **Services**: 1 (CatalogService)
- **Methods**: 4 (LoadProtos, ListServices, GetServiceSchema, InvokeGRPC)
- **Message Types**: 13 (Request/Response messages)

## Writing New Tests

### Integration Test Template

```go
func TestIntegration<Feature>(t *testing.T) {
    // Setup: Create test server
    catalogServer := server.New()
    defer catalogServer.Close()

    mux := http.NewServeMux()
    path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
    mux.Handle(path, handler)

    testServer := httptest.NewServer(mux)
    defer testServer.Close()

    client := catalogv1connect.NewCatalogServiceClient(
        http.DefaultClient,
        testServer.URL,
    )

    ctx := context.Background()

    // Test: Your test logic here
    req := connect.NewRequest(&catalogv1.YourRequest{})
    resp, err := client.YourMethod(ctx, req)

    // Verify: Assertions
    if err != nil {
        t.Fatalf("YourMethod failed: %v", err)
    }

    // Add specific assertions
}
```

### Best Practices

1. **Use t.Helper()** in helper functions
2. **Clean up resources** with `defer`
3. **Test both success and error paths**
4. **Use descriptive test names** (`TestIntegration<Feature>_<Scenario>`)
5. **Log useful information** with `t.Logf()`
6. **Verify all critical fields** in responses

## Continuous Integration

### GitHub Actions (Planned)

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Install buf
        run: |
          curl -sSL https://github.com/bufbuild/buf/releases/download/latest/buf-Linux-x86_64 \
            -o /usr/local/bin/buf
          chmod +x /usr/local/bin/buf
      - name: Run tests
        run: go test -v -race -cover ./...
```

## Test Coverage Goals

- **Integration Tests**: 100% of API endpoints
- **Unit Tests**: >80% code coverage
- **E2E Tests**: Critical user workflows

## Debugging Failed Tests

### Common Issues

**Proto files not found**:
```bash
# Verify proto directory exists
ls -la proto/catalog/v1/

# Run tests from project root
cd /Users/jschumacher/Projects/connectrpc-catalog
go test ./internal/server/...
```

**Buf not installed**:
```bash
# Check buf installation
buf --version

# Install buf if missing
brew install bufbuild/buf/buf
```

**Port already in use**:
```bash
# Integration tests use httptest, which allocates random ports
# This should not happen, but if it does, check for leaked processes
lsof -ti:8080 | xargs kill -9
```

## Performance Benchmarks (Planned)

```bash
# Run benchmarks
go test -bench=. ./...

# Benchmark with memory profiling
go test -bench=. -benchmem ./...
```

## Test Maintenance

### When to Update Tests

- **Adding new API endpoints**: Add corresponding integration tests
- **Changing API contracts**: Update request/response validations
- **Bug fixes**: Add regression tests
- **Performance improvements**: Add benchmark comparisons

### Test Review Checklist

- [ ] Tests cover success and error paths
- [ ] Resources are properly cleaned up
- [ ] Test names clearly describe what's being tested
- [ ] Assertions verify all critical fields
- [ ] Helper functions use `t.Helper()`
- [ ] Tests are independent and can run in any order

## Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Connect-Go Testing](https://connectrpc.com/docs/go/testing)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
