# Session Management

Session-based state management for the ConnectRPC Catalog server.

## Overview

The session package provides per-session state isolation for the catalog server. Each client can maintain their own isolated registry and invoker state by using session IDs.

## Key Features

- **Automatic Session Creation**: Sessions are automatically created when no session ID is provided
- **State Isolation**: Each session has its own Registry and Invoker instances
- **Automatic Cleanup**: Expired sessions are automatically cleaned up based on TTL
- **Concurrent Safe**: All operations are protected by read-write locks

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ X-Session-ID: abc123
       ▼
┌─────────────────────────┐
│   CatalogServer         │
│  ┌──────────────────┐   │
│  │ Session Manager  │   │
│  └────────┬─────────┘   │
│           │             │
│  ┌────────▼─────────┐   │
│  │  Session State   │   │
│  │  ┌────────────┐  │   │
│  │  │ Registry   │  │   │
│  │  ├────────────┤  │   │
│  │  │ Invoker    │  │   │
│  │  └────────────┘  │   │
│  └──────────────────┘   │
└─────────────────────────┘
```

## Usage

### Server-Side

The CatalogServer automatically manages sessions. No manual session handling is required.

```go
import "github.com/opentdf/connectrpc-catalog/internal/server"

// Create server with session management
srv := server.New()
defer srv.Close()

// Server automatically handles session creation and routing
```

### Client-Side

Clients should preserve and send the `X-Session-ID` header in subsequent requests.

```go
import (
	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	"github.com/opentdf/connectrpc-catalog/gen/catalog/v1/catalogv1connect"
)

// Create client
client := catalogv1connect.NewCatalogServiceClient(
	http.DefaultClient,
	"http://localhost:8080",
)

// First request creates a new session
req1 := connect.NewRequest(&catalogv1.LoadProtosRequest{
	Source: &catalogv1.LoadProtosRequest_ProtoPath{
		ProtoPath: "/path/to/protos",
	},
})

resp1, err := client.LoadProtos(ctx, req1)
if err != nil {
	// handle error
}

// Extract session ID from response
sessionID := resp1.Header().Get("X-Session-ID")

// Use session ID in subsequent requests
req2 := connect.NewRequest(&catalogv1.ListServicesRequest{})
req2.Header().Set("X-Session-ID", sessionID)

resp2, err := client.ListServices(ctx, req2)
// Now resp2 will contain services from the same session
```

## Configuration

### Session TTL

Default TTL is 1 hour. Sessions are automatically cleaned up after being inactive for this duration.

```go
import (
	"time"
	"github.com/opentdf/connectrpc-catalog/internal/session"
)

// Create manager with custom TTL
manager := session.NewManager(30 * time.Minute)
defer manager.Close()
```

### Cleanup Interval

Sessions are checked for expiration every 5 minutes by default.

```go
// Defined in session/session.go
const CleanupInterval = 5 * time.Minute
```

## Session Lifecycle

1. **Creation**: Session is created on first request without session ID
2. **Usage**: Client sends session ID in `X-Session-ID` header
3. **Update**: Each request updates the session's `LastUsed` timestamp
4. **Expiration**: Sessions expire after TTL period of inactivity
5. **Cleanup**: Background goroutine removes expired sessions

## Session State

Each session maintains:

```go
type State struct {
	Registry  *registry.Registry  // Descriptor registry
	Invoker   *invoker.Invoker    // gRPC connection pool
	CreatedAt time.Time           // Session creation time
	LastUsed  time.Time           // Last activity timestamp
}
```

## Statistics

Get session statistics:

```go
stats := manager.GetStats()
fmt.Printf("Active Sessions: %d\n", stats.ActiveSessions)
fmt.Printf("Oldest Session: %v\n", stats.OldestSession)
fmt.Printf("Newest Session: %v\n", stats.NewestSession)
```

## Thread Safety

All session operations are thread-safe:
- `GetOrCreate`: Safe for concurrent calls
- `Get`: Safe for concurrent reads
- `Delete`: Safe for concurrent deletion
- `cleanup`: Automatically handles concurrent access

## Best Practices

1. **Preserve Session IDs**: Always send the session ID from the first response
2. **Handle Missing Sessions**: If session expires, server creates a new one automatically
3. **Session Isolation**: Don't share session IDs across different logical clients
4. **Resource Management**: Sessions automatically clean up when expired

## Error Handling

Session creation failures return errors through the standard Connect error system:

```go
state, sessionID, err := manager.GetOrCreate(requestSessionID)
if err != nil {
	return connect.NewError(connect.CodeInternal, err)
}
```

## Performance

- Session lookup: O(1) hash map access
- Session creation: Minimal overhead (ID generation + map insert)
- Cleanup: O(n) but runs infrequently (every 5 minutes)
- Concurrent access: Read-write locks for optimal performance

## Testing

See `session_test.go` for comprehensive unit tests and `session_integration_test.go` for integration tests.

```bash
# Run session tests
go test ./internal/session/...

# Run integration tests
go test ./internal/server/... -run TestSession
```
