# ConnectRPC Catalog - Build Instructions

## Prerequisites

- **Go**: 1.23 or later
- **Buf CLI**: Install from https://buf.build/docs/installation
- **Node.js**: 18+ (for UI in later waves)

## Quick Start

### 1. Install Buf CLI

```bash
# macOS/Linux
brew install bufbuild/buf/buf

# Or download directly
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

### 2. Generate Go Code from Protos

```bash
# Generate Go + ConnectRPC code
buf generate

# This creates:
# - gen/catalog/v1/catalog.pb.go (protobuf messages)
# - gen/catalog/v1/catalogv1connect/catalog.connect.go (ConnectRPC service)
```

### 3. Install Go Dependencies

```bash
go mod download
go mod tidy
```

### 4. Verify Installation

```bash
# Check generated files exist
ls -la gen/catalog/v1/

# Expected output:
# - catalog.pb.go
# - catalogv1connect/catalog.connect.go
```

## Project Structure

```
github.com/opentdf/connectrpc-catalog/
├── cmd/connectrpc-catalog/      # CLI entry point (Wave 3-4)
├── proto/catalog/v1/            # Proto definitions
│   └── catalog.proto            # CatalogService definition
├── gen/catalog/v1/              # Generated Go code (created by buf)
│   ├── catalog.pb.go
│   └── catalogv1connect/
│       └── catalog.connect.go
├── internal/
│   ├── loader/                  # Proto loading (Wave 3)
│   ├── registry/                # Descriptor registry (Wave 3)
│   ├── invoker/                 # gRPC invocation (Wave 3)
│   └── server/                  # Connect handlers (Wave 3)
├── ui/                          # React frontend (Wave 3-4)
├── buf.yaml                     # Buf configuration
├── buf.gen.yaml                 # Code generation config
└── go.mod                       # Go module dependencies
```

## Key Dependencies

### Go Packages

- **connectrpc.com/connect**: ConnectRPC protocol implementation
- **google.golang.org/protobuf**: Protocol buffers runtime
- **google.golang.org/grpc**: gRPC client/server
- **github.com/jhump/protoreflect**: Dynamic proto reflection

### Buf Configuration

- **buf.yaml**: Module definition and dependencies
- **buf.gen.yaml**: Code generation pipeline

## Development Workflow

### Modify Protos

```bash
# 1. Edit proto/catalog/v1/catalog.proto
# 2. Regenerate code
buf generate

# 3. Update Go code to match new definitions
```

### Add Proto Dependencies

```bash
# Add to buf.yaml deps section
deps:
  - buf.build/googleapis/googleapis
  - buf.build/your/module
```

### Linting and Breaking Change Detection

```bash
# Lint proto files
buf lint

# Check for breaking changes
buf breaking --against '.git#branch=main'
```

## Next Steps (Wave 2+)

- **Wave 2**: Run `buf generate` and verify generated code
- **Wave 3**: Implement Go backend packages (loader, registry, invoker, server)
- **Wave 3**: Build React UI with ConnectRPC client
- **Wave 4**: Integration, build pipeline, CLI wrapper

## Troubleshooting

### Buf Not Found

```bash
# Install buf CLI
brew install bufbuild/buf/buf
# or follow https://buf.build/docs/installation
```

### Go Version Mismatch

```bash
# Check Go version
go version

# Update if needed (must be 1.23+)
```

### Generation Errors

```bash
# Clean and regenerate
rm -rf gen/
buf generate
```

## References

- **ConnectRPC**: https://connectrpc.com/docs/go/getting-started
- **Buf**: https://buf.build/docs
- **protoreflect**: https://github.com/jhump/protoreflect
