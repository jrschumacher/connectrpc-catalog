# Implementation Plan - ConnectRPC Catalog MVP

## Tech Stack

- **Backend**: Go + ConnectRPC (dogfooding)
- **Frontend**: React + Vite + Tailwind + shadcn/ui
- **Proto Resolution**: Buf for dependency management
- **Module**: `github.com/opentdf/connectrpc-catalog`

## Architecture

```
github.com/opentdf/connectrpc-catalog/
├── cmd/connectrpc-catalog/     # CLI entry point
├── proto/catalog/v1/           # Our own API definition
├── gen/catalog/v1/             # Generated Go code
├── internal/
│   ├── loader/                 # Proto loading (buf integration)
│   ├── registry/               # In-memory descriptor registry
│   ├── invoker/                # gRPC dynamic invocation
│   └── server/                 # Connect handlers
└── ui/                         # TypeScript frontend
    ├── src/
    │   ├── components/
    │   │   ├── ServiceBrowser.tsx
    │   │   ├── MethodDetails.tsx
    │   │   ├── RequestEditor.tsx
    │   │   └── ResponseViewer.tsx
    │   ├── lib/
    │   │   ├── client.ts
    │   │   └── schema.ts
    │   └── App.tsx
    └── dist/                   # Embedded build output
```

## ConnectRPC Service Definition

`proto/catalog/v1/catalog.proto`:
```protobuf
service CatalogService {
  // Load protos from various sources
  rpc LoadProtos(LoadProtosRequest) returns (LoadProtosResponse);

  // List all discovered services
  rpc ListServices(ListServicesRequest) returns (ListServicesResponse);

  // Get schema for a specific service
  rpc GetServiceSchema(GetServiceSchemaRequest) returns (GetServiceSchemaResponse);

  // Invoke a gRPC method (proxy through backend)
  rpc InvokeGRPC(InvokeGRPCRequest) returns (InvokeGRPCResponse);
}
```

## CLI Usage Examples

```bash
# Local directory
connectrpc-catalog serve --proto-path ./protos

# GitHub repo (auto-detects buf.yaml)
connectrpc-catalog serve --proto-repo github.com/connectrpc/eliza

# Buf registry
connectrpc-catalog serve --buf-module buf.build/connectrpc/eliza

# Custom port
connectrpc-catalog serve --proto-path ./protos --port 8080
```

## Task Breakdown

### WAVE 1: Foundation (Sequential)
**Agent**: golang-pro

- [x] **Task 1.1**: Project Structure + Proto Definitions
  - Initialize Go module `github.com/opentdf/connectrpc-catalog`
  - Create `proto/catalog/v1/catalog.proto` with service definition
  - Configure `buf.yaml` and `buf.gen.yaml`
  - Create directory structure
  - **Blocks**: All other tasks

### WAVE 2: Code Generation (Sequential)
**Agent**: golang-pro

- [ ] **Task 2.1**: Generate Go ConnectRPC Code
  - Run `buf generate`
  - Output to `gen/catalog/v1/`
  - **Blocks**: Backend implementation, TypeScript client

### WAVE 3: Parallel Implementation

#### Track A: Go Backend (golang-pro)

- [ ] **Task 3.1**: Proto Loader (buf integration) ⚡ PARALLEL
  - Package: `internal/loader`
  - Functions:
    - `LoadFromPath(path string) (*descriptorpb.FileDescriptorSet, error)`
    - `LoadFromGitHub(repo string) (*descriptorpb.FileDescriptorSet, error)`
    - `LoadFromBufModule(module string) (*descriptorpb.FileDescriptorSet, error)`
  - Uses `buf export` subprocess

- [ ] **Task 3.2**: Descriptor Registry ⚡ PARALLEL
  - Package: `internal/registry`
  - Functions:
    - `Register(fds *descriptorpb.FileDescriptorSet) error`
    - `ListServices() []ServiceInfo`
    - `GetMethodDescriptor(service, method string) (*desc.MethodDescriptor, error)`
    - `GetMessageSchema(msgName string) (*desc.MessageDescriptor, error)`

- [ ] **Task 3.3**: gRPC Dynamic Invoker ⚡ PARALLEL
  - Package: `internal/invoker`
  - Functions:
    - `InvokeUnary(endpoint, service, method string, req json.RawMessage) (json.RawMessage, error)`
  - Uses `grpcurl` patterns with `grpc.NewClient` + dynamic stubs
  - Unary RPCs only for MVP

- [ ] **Task 3.4**: Connect Service Handlers (Sequential after 3.1-3.3)
  - Package: `internal/server`
  - Implements all RPC handlers from `catalog.proto`
  - Integrates loader, registry, invoker

- [ ] **Task 3.5**: Embed UI + HTTP Server (Sequential after 3.4 + Wave 4)
  - Package: `cmd/connectrpc-catalog`
  - `embed.FS` for UI assets
  - Connect handler registration
  - Static file serving with SPA fallback

#### Track B: TypeScript Frontend (typescript-pro)

- [ ] **Task 3.6**: UI Scaffold ⚡ PARALLEL
  - Vite + React + TypeScript setup
  - Tailwind CSS + shadcn/ui initialization
  - Connect client configuration
  - Basic routing structure

- [ ] **Task 3.7**: ServiceBrowser Component ⚡ PARALLEL
  - Sidebar with service/method tree navigation
  - Search and filter functionality
  - Active selection state management

- [ ] **Task 3.8**: Request/Response Components ⚡ PARALLEL
  - `RequestEditor.tsx` - JSON form builder from Protobuf schema
  - `ResponseViewer.tsx` - Formatted response display with syntax highlighting
  - `MethodDetails.tsx` - Display method info and schema
  - Transport selector (Connect/gRPC toggle)

- [ ] **Task 3.9**: Connect Client Integration (Sequential after 3.6-3.8)
  - Call `LoadProtos`, `ListServices`, `GetServiceSchema`
  - Invoke Connect directly from browser (Connect transport)
  - Proxy gRPC through backend (`InvokeGRPC`)
  - Error handling and loading states

### WAVE 4: Integration (Sequential)

- [ ] **Task 4.1**: Build Pipeline
  - Agent: golang-pro
  - Steps:
    1. `cd ui && npm run build`
    2. Go build with embedded assets
    3. Single binary output

- [ ] **Task 4.2**: CLI Wrapper
  - Agent: golang-pro
  - Implement CLI flags:
    - `--proto-path` for local directories
    - `--proto-repo` for GitHub repositories
    - `--buf-module` for Buf registry modules
    - `--port` for HTTP server port (default: 8080)
    - `--endpoint` for default target service endpoint

- [ ] **Task 4.3**: Minimal Tests
  - golang-pro: Integration test for each Connect handler
  - typescript-pro: One E2E Playwright test (load service → invoke method)

## Execution Flow

```
WAVE 1: Foundation (golang-pro)
  └─ Project + Proto (sequential)
        ↓
WAVE 2: Codegen (golang-pro)
  └─ buf generate (sequential)
        ↓
WAVE 3: Implementation (PARALLEL)
  ┌──────────────────┐    ┌──────────────────┐
  │  golang-pro      │    │  typescript-pro  │
  ├──────────────────┤    ├──────────────────┤
  │ • Loader         │    │ • UI Scaffold    │
  │ • Registry       │    │ • ServiceBrowser │
  │ • Invoker        │    │ • Request/Resp   │
  │ • Connect Svc    │    │ • Client         │
  └──────────────────┘    └──────────────────┘
        ↓
WAVE 4: Integration (golang-pro + typescript-pro)
  └─ Build pipeline + CLI + Tests
```

## Key Go Dependencies

- `google.golang.org/protobuf/reflect/protodesc`
- `google.golang.org/grpc`
- `github.com/jhump/protoreflect`
- `connectrpc.com/connect`
- `github.com/bufbuild/buf/private/pkg/...` (for buf integration)

## Key TypeScript Dependencies

- `@connectrpc/connect`
- `@connectrpc/connect-web`
- `react`
- `vite`
- `tailwindcss`
- `@radix-ui/react-*` (via shadcn/ui)

## MVP Constraints

- Unary RPCs only (no streaming)
- No authentication/authorization
- In-memory descriptor storage (no persistence)
- Single concurrent user (no session management)
- UI-based endpoint configuration (no config file)

## Success Criteria

- Load protos from local directory
- Display services and methods in UI
- Invoke unary RPC via Connect transport (from browser)
- Invoke unary RPC via gRPC transport (proxied through Go backend)
- Display formatted request/response JSON
- Single binary deployment
