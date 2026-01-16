# Wave 1 Complete - Foundation

## Deliverables

### 1. Project Structure
```
github.com/opentdf/connectrpc-catalog/
├── cmd/connectrpc-catalog/      # CLI entry point (empty, for Wave 3-4)
├── proto/catalog/v1/            # Proto definitions
│   └── catalog.proto            # ✅ Complete CatalogService definition
├── internal/
│   ├── loader/                  # Proto loading (empty, for Wave 3)
│   ├── registry/                # Descriptor registry (empty, for Wave 3)
│   ├── invoker/                 # gRPC invocation (empty, for Wave 3)
│   └── server/                  # Connect handlers (empty, for Wave 3)
├── ui/                          # React frontend (empty, for Wave 3-4)
├── buf.yaml                     # ✅ Buf configuration
├── buf.gen.yaml                 # ✅ Code generation config
├── go.mod                       # ✅ Go module with dependencies
├── .gitignore                   # ✅ Git ignore patterns
└── BUILD.md                     # ✅ Build instructions
```

### 2. Proto Definitions (/Users/jschumacher/Projects/connectrpc-catalog/proto/catalog/v1/catalog.proto)

Complete ConnectRPC service with 4 RPCs:

- **LoadProtos**: Load proto definitions from local path, GitHub repo, or Buf module
  - Input: `LoadProtosRequest` with oneof source (proto_path | proto_repo | buf_module)
  - Output: `LoadProtosResponse` with success status, error, service/file counts

- **ListServices**: List all discovered services and methods
  - Input: `ListServicesRequest` (empty)
  - Output: `ListServicesResponse` with array of `ServiceInfo`

- **GetServiceSchema**: Get full message schema for a service as JSON
  - Input: `GetServiceSchemaRequest` with service_name
  - Output: `GetServiceSchemaResponse` with service info and message schemas map

- **InvokeGRPC**: Dynamically invoke a gRPC method (proxy through backend)
  - Input: `InvokeGRPCRequest` with endpoint, service, method, request_json, TLS options, timeout, metadata
  - Output: `InvokeGRPCResponse` with success, response_json, error, metadata, status

### 3. Buf Configuration

**buf.yaml**:
- Version: v2
- Module: `buf.build/opentdf/connectrpc-catalog`
- Dependencies: `buf.build/googleapis/googleapis`
- Lint: STANDARD rules
- Breaking change detection: FILE level

**buf.gen.yaml**:
- Managed mode enabled with `go_package_prefix`
- Plugins:
  - `buf.build/protocolbuffers/go` → `gen/` (protobuf messages)
  - `buf.build/connectrpc/go` → `gen/` (ConnectRPC service)
- Output: `paths=source_relative`

### 4. Go Module (/Users/jschumacher/Projects/connectrpc-catalog/go.mod)

Dependencies:
- `connectrpc.com/connect@v1.17.0` - ConnectRPC protocol
- `google.golang.org/protobuf@v1.34.2` - Protocol buffers runtime
- `google.golang.org/grpc@v1.65.0` - gRPC client/server
- `github.com/jhump/protoreflect@v1.16.0` - Dynamic proto reflection

### 5. Build Instructions (/Users/jschumacher/Projects/connectrpc-catalog/BUILD.md)

Complete guide covering:
- Prerequisites (Go 1.23+, Buf CLI, Node.js 18+)
- Installation steps
- Code generation workflow
- Project structure explanation
- Development workflow
- Troubleshooting

## Validation

```bash
# ✅ Buf lint passes
buf lint

# ✅ Go module is valid
go mod download

# ✅ Directory structure matches spec
ls -R cmd/ proto/ internal/ ui/

# ✅ Proto definitions are complete
cat proto/catalog/v1/catalog.proto
```

## Next Steps - Wave 2

**Task 2.1**: Generate Go ConnectRPC Code
```bash
buf generate
```

This will create:
- `gen/catalog/v1/catalog.pb.go` - Protobuf message types
- `gen/catalog/v1/catalogv1connect/catalog.connect.go` - ConnectRPC service interfaces

Blocks: All Wave 3 implementation tasks

## Success Metrics

- [x] Complete directory structure matching spec
- [x] Valid proto definitions with 4 RPCs
- [x] Proper buf.yaml and buf.gen.yaml configuration
- [x] go.mod with all required dependencies
- [x] buf lint passes without errors
- [x] Build instructions document complete
- [x] .gitignore configured for Go + Node.js

## Time Taken

Wave 1 completed in ~5 minutes (structure + proto + config + docs)

## Files Created

1. `/Users/jschumacher/Projects/connectrpc-catalog/proto/catalog/v1/catalog.proto` - 183 lines
2. `/Users/jschumacher/Projects/connectrpc-catalog/buf.yaml` - 12 lines
3. `/Users/jschumacher/Projects/connectrpc-catalog/buf.gen.yaml` - 13 lines
4. `/Users/jschumacher/Projects/connectrpc-catalog/go.mod` - 19 lines
5. `/Users/jschumacher/Projects/connectrpc-catalog/.gitignore` - 39 lines
6. `/Users/jschumacher/Projects/connectrpc-catalog/BUILD.md` - 159 lines
7. `/Users/jschumacher/Projects/connectrpc-catalog/WAVE1_COMPLETE.md` - This file

## Ready for Wave 2

All blocking dependencies resolved. Wave 2 can proceed with `buf generate`.
