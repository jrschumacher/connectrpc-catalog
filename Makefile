.PHONY: all build clean dev ui backend test test-e2e run help

# Default target
all: build

# Build both UI and Go binary
build: gen
	@echo "Building ConnectRPC Catalog..."
	@./build.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ui/dist
	@rm -rf bin
	@rm -rf ui/node_modules
	@echo "Clean complete"

# Development mode - run UI and backend separately
dev:
	@echo "Starting development servers..."
	@echo "Backend will run on http://localhost:8080"
	@echo "UI will run on http://localhost:5173"
	@echo ""
	@echo "Run these in separate terminals:"
	@echo "  Terminal 1: make backend"
	@echo "  Terminal 2: make ui"

# Run UI development server
ui:
	@echo "Starting UI development server..."
	@cd ui && npm run dev

# Run backend server (without embedded UI)
backend:
	@echo "Starting backend server..."
	@go run ./cmd/connectrpc-catalog

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run full-stack E2E tests (starts backend, runs Playwright tests, stops backend)
test-e2e: build
	@echo "Running full-stack E2E tests..."
	@./test-e2e-full.sh

# Run the built binary
run: build
	@echo "Starting ConnectRPC Catalog..."
	@./bin/connectrpc-catalog

# Install UI dependencies
install-ui:
	@echo "Installing UI dependencies..."
	@cd ui && npm install

# Generate protobuf code
gen:
	@echo "Generating protobuf code..."
	@buf generate
	@echo "Generating eliza service protos (test dependency)..."
	@buf generate buf.build/connectrpc/eliza

# Format code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Formatting UI code..."
	@cd ui && npm run lint

# Help
help:
	@echo "ConnectRPC Catalog - Available targets:"
	@echo ""
	@echo "  make build       - Build UI and Go binary (default)"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make dev         - Show development mode instructions"
	@echo "  make ui          - Run UI development server"
	@echo "  make backend     - Run backend server"
	@echo "  make test        - Run tests"
	@echo "  make test-e2e    - Run full-stack E2E tests (backend + Playwright)"
	@echo "  make run         - Build and run the binary"
	@echo "  make install-ui  - Install UI dependencies"
	@echo "  make gen         - Generate protobuf code"
	@echo "  make fmt         - Format code"
	@echo "  make help        - Show this help message"
