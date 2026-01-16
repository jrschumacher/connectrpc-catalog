Below is a **concise but real spec** you could hand to an engineer (or use as a README / design doc) that clearly states intent, scope, and non‑goals—*without accidentally reinventing Swagger or REST*.

I’ll write this as if the working name is **`connectrpc-inspect`**, but that’s easily swappable.

---

# connectrpc-inspect — Specification (Draft)

## 1. Purpose

`connectrpc-inspect` is an **interactive, proto‑native API exploration tool** for ConnectRPC services.

It provides a **Swagger‑like look and feel** for *humans*, while preserving the **schema‑first, transport‑agnostic philosophy** of Protobuf and ConnectRPC.

The tool is **not an API contract** and **not a replacement for SDKs**.  
Its role is discovery, experimentation, and debugging.

---

## 2. Core Principles

1. **Protobuf is the source of truth**
   - `.proto` files define the API
   - No OpenAPI or REST schema is required

2. **Transport is a choice, not an identity**
   - Every RPC may be invoked via:
     - Connect (HTTP/JSON)
     - gRPC (binary)
   - The UI makes this explicit and selectable

3. **SDKs remain the primary interface**
   - This tool exists for humans, not production integrations

4. **Swagger‑like UX, not Swagger semantics**
   - Familiar layout and flow
   - No REST bias (paths, verbs, status codes)

---

## 3. Non‑Goals

- ❌ Replacing generated SDKs
- ❌ Acting as a public API portal
- ❌ Defining compatibility or versioning rules
- ❌ Generating OpenAPI as a primary output
- ❌ Supporting arbitrary REST endpoints

---

## 4. Target Users

- Backend engineers
- Platform teams
- API consumers onboarding to a service
- SREs debugging live services

---

## 5. High‑Level Feature Set

### 5.1 Service Discovery

The tool MUST support discovering services via:

1. **Protobuf sources**
   - Local `.proto` files
   - Buf modules
2. **Runtime reflection (optional)**
   - gRPC reflection
   - Connect metadata endpoints (if available)

---

### 5.2 Schema Browsing (Swagger‑Like UI)

The UI SHOULD resemble Swagger UI in layout:

- Left sidebar:
  - Package
  - Service
  - RPC method
- Main panel:
  - Method description
  - Request message schema
  - Response message schema
  - Comments from `.proto` files

But MUST differ conceptually:

| Swagger | connectrpc-inspect |
|------|-------------------|
| HTTP paths | RPC methods |
| HTTP verbs | Method types (Unary / Streaming) |
| JSON schema | Protobuf messages |

---

### 5.3 Method Invocation (“Try It Out”)

Each RPC method MAY be invoked interactively.

#### Invocation Capabilities
- Unary RPCs (required)
- Server streaming (phase 2)
- Client / bidi streaming (phase 3)

#### Request Editing
- Structured form generated from Protobuf
- Raw JSON editor (advanced mode)
- Field defaults derived from schema

---

### 5.4 Transport Selection

Each invocation MUST allow selecting transport:

- ✅ **Connect (HTTP/JSON)**
- ✅ **gRPC (binary)**

The UI SHOULD:
- Explain differences briefly
- Show headers / metadata per transport
- Make transport choice explicit

---

### 5.5 Response Inspection

Responses SHOULD display:

- Structured decoded response
- Raw response (JSON or binary metadata)
- gRPC status / Connect error
- Trailers and metadata (advanced view)

---

## 6. Architecture Overview

### 6.1 Inputs

- `.proto` files
- Buf modules
- Optional reflection endpoints

### 6.2 Internal Model

- Protobuf descriptors (`FileDescriptorSet`)
- Service graph
- Message schemas

### 6.3 Execution Layer

| Transport | Implementation |
|--------|----------------|
| Connect | HTTP client using Connect protocol |
| gRPC | gRPC client with reflection support |

---

## 7. Security Model

- Intended for **trusted environments**
- No authentication baked in
- Pluggable auth:
  - Headers
  - mTLS
  - OAuth tokens (manual input)

---
## 8. Extensibility

### 8.1 Plugin Hooks (Future)

- Auth providers
- Custom field renderers
- Organization‑specific annotations

---

## 9. CLI + UI Split (Recommended)

### CLI (`connectrpc inspect`)
- Loads schemas
- Starts local UI
- Configures endpoints
- Handles auth

### UI
- Pure frontend
- Talks to CLI backend
- No direct service access

---

## 10. Comparison Matrix

| Tool | Proto‑Native | Interactive | Multi‑Transport |
|----|----|----|----|
| Swagger UI | ❌ | ✅ | ❌ |
| grpcui | ✅ | ✅ | ❌ |
| Buf Docs | ✅ | ❌ | ✅ |
| connectrpc‑inspect | ✅ | ✅ | ✅ |

---

## 11. Why This Is Straightforward to Build

This tool is primarily **composition**, not invention:

- Protobuf descriptors → already solved
- Schema rendering → grpcui, Buf Docs
- Connect invocation → Connect SDKs
- gRPC invocation → grpcurl / grpcui patterns
- UI patterns → Swagger UI mental model

The innovation is **unifying these into a transport‑agnostic, proto‑first explorer**.

---

## 12. Success Criteria

The tool succeeds if:

- Engineers stop asking “where is Swagger?”
- API consumers can experiment without writing code
- No one mistakes it for the actual API contract
- It reinforces (not undermines) SDK‑first usage

---

## One‑Sentence Description (For the Repo)

> `connectrpc-inspect` is a proto‑native, transport‑agnostic API explorer for ConnectRPC services, offering a familiar Swagger‑like experience without compromising schema‑first design.

---

The focus will be on UX. We want this to look and feel really good like modern software.

---

## Getting Started (MVP Implementation)

### Prerequisites

- Go 1.24+
- Node.js 18+
- npm 9+

### Quick Build

```bash
# Build everything (UI + Go binary)
make build

# Or use the build script directly
./build.sh
```

This produces a single binary at `bin/connectrpc-catalog` with the UI embedded.

### Run

```bash
# Run the built binary
./bin/connectrpc-catalog

# Or build and run in one step
make run

# Custom port and host
./bin/connectrpc-catalog -port 3000 -host 0.0.0.0
```

The server will start on http://localhost:8080 by default.

### Development Mode

For faster iteration during development, run the UI and backend separately:

```bash
# Terminal 1: Run backend server
make backend

# Terminal 2: Run UI dev server with hot reload
make ui
```

The UI dev server (http://localhost:5173) will proxy API requests to the backend (http://localhost:8080).

## Current MVP Features

✅ Load protobuf definitions from:
- Local file paths
- GitHub repositories
- Buf Schema Registry

✅ Browse loaded services with full schema details

✅ Test gRPC endpoints with dynamic invocation (unary methods)

✅ Modern React UI with embedded backend (single binary)

✅ ConnectRPC API for all operations

## Project Structure

```
connectrpc-catalog/
├── cmd/
│   └── connectrpc-catalog/      # Main application entry point
│       └── main.go              # HTTP server with embedded UI
├── internal/
│   ├── loader/                  # Proto loading logic
│   ├── registry/                # Descriptor registry
│   ├── invoker/                 # Dynamic gRPC invocation
│   └── server/                  # ConnectRPC handlers
├── proto/
│   └── catalog/v1/              # Service API definitions
├── gen/                         # Generated protobuf code
├── ui/                          # React frontend
│   ├── src/
│   ├── dist/                    # Built UI assets (embedded)
│   └── package.json
├── bin/                         # Built binaries
├── build.sh                     # Build script
└── Makefile                     # Build automation
```

## Development Commands

```bash
make build       # Build UI and Go binary (default)
make clean       # Clean build artifacts
make dev         # Show development mode instructions
make ui          # Run UI development server
make backend     # Run backend server
make test        # Run tests
make run         # Build and run the binary
make install-ui  # Install UI dependencies
make gen         # Generate protobuf code
make fmt         # Format code
make help        # Show help message
```

## MVP Limitations

- Only unary RPC methods supported (no streaming)
- In-memory registry only (no persistence)
- Basic error handling
- No authentication/authorization
- No request history
- Connect transport only (gRPC binary transport coming in future phases)
