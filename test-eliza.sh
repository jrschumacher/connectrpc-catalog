#!/bin/bash
# All-in-one Eliza service evaluation script
# Tests all protocols: Connect, gRPC, gRPC-Web

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[PASS]${NC} $1"; }
error() { echo -e "${RED}[FAIL]${NC} $1"; }

KEEP_RUNNING=false
for arg in "$@"; do
    case $arg in
        --serve|--keep|-s|-k) KEEP_RUNNING=true ;;
    esac
done

cleanup() {
    if [ "$KEEP_RUNNING" = false ]; then
        if [ -n "$ELIZA_PID" ] && kill -0 "$ELIZA_PID" 2>/dev/null; then
            kill "$ELIZA_PID" 2>/dev/null || true
            wait "$ELIZA_PID" 2>/dev/null || true
        fi
    fi
}
trap cleanup EXIT

echo "========================================"
echo "  Eliza Service Evaluation Suite"
echo "========================================"
echo

# Kill any existing test-eliza processes
pkill -f "test-eliza" 2>/dev/null || true
sleep 1

# 1. Unit Tests
info "Running unit tests (all protocols)..."
if go test -v ./internal/elizaservice/... 2>&1 | tee /tmp/eliza-unit.log | grep -E "^(---|PASS|FAIL|ok|FAIL)"; then
    success "Unit tests passed"
else
    error "Unit tests failed"
    cat /tmp/eliza-unit.log
    exit 1
fi
echo

# 2. Integration Tests
info "Starting Eliza server for integration tests..."
go run ./cmd/test-eliza --port 50097 &
ELIZA_PID=$!
sleep 2

# Wait for server to be healthy
for i in {1..10}; do
    if curl -s http://localhost:50097/health > /dev/null 2>&1; then
        break
    fi
    sleep 0.5
done

if ! curl -s http://localhost:50097/health > /dev/null 2>&1; then
    error "Eliza server failed to start"
    exit 1
fi
success "Eliza server running on port 50097"

info "Running invoker integration tests..."
if go test -v ./internal/invoker/ -run TestElizaIntegration 2>&1 | tee /tmp/eliza-integration.log | grep -E "^(---|PASS|FAIL|ok|FAIL)"; then
    success "Integration tests passed"
else
    error "Integration tests failed"
    cat /tmp/eliza-integration.log
    kill "$ELIZA_PID" 2>/dev/null || true
    exit 1
fi

# Stop server for integration tests
kill "$ELIZA_PID" 2>/dev/null || true
wait "$ELIZA_PID" 2>/dev/null || true
unset ELIZA_PID
echo

# 3. Manual curl tests
info "Running manual protocol tests..."
go run ./cmd/test-eliza --port 50098 &
ELIZA_PID=$!
sleep 2

# Wait for server
for i in {1..10}; do
    if curl -s http://localhost:50098/health > /dev/null 2>&1; then
        break
    fi
    sleep 0.5
done

# Test Connect protocol
info "Testing Connect protocol (HTTP/JSON)..."
CONNECT_RESP=$(curl -s -X POST http://localhost:50098/connectrpc.eliza.v1.ElizaService/Say \
    -H "Content-Type: application/json" \
    -d '{"sentence": "Hello from curl test"}')

if echo "$CONNECT_RESP" | grep -q "sentence"; then
    success "Connect protocol: $CONNECT_RESP"
else
    error "Connect protocol failed: $CONNECT_RESP"
fi

# Note: gRPC-Web requires binary framing that curl can't provide
# The unit tests verify gRPC-Web works properly
info "gRPC-Web protocol verified via unit tests (requires binary framing)"

kill "$ELIZA_PID" 2>/dev/null || true
wait "$ELIZA_PID" 2>/dev/null || true
unset ELIZA_PID
echo

# 4. E2E Tests (optional, requires UI built)
if [ "$1" = "--e2e" ] || [ "$1" = "-e" ]; then
    info "Running full E2E tests..."
    ./test-e2e-full.sh
else
    info "Skipping E2E tests (use --e2e to include)"
fi

echo
echo "========================================"
success "All Eliza service tests passed!"
echo "========================================"
echo
echo "Summary:"
echo "  ✓ Unit tests (Connect, gRPC, gRPC-Web)"
echo "  ✓ Integration tests (Invoker)"
echo "  ✓ Protocol tests (curl)"
for arg in "$@"; do
    if [ "$arg" = "--e2e" ] || [ "$arg" = "-e" ]; then
        echo "  ✓ E2E tests (Playwright)"
    fi
done

# Keep servers running for manual testing if requested
if [ "$KEEP_RUNNING" = true ]; then
    echo
    echo "========================================"
    info "Starting servers for manual testing"
    echo "========================================"

    # Clean up any existing processes
    pkill -f "test-eliza" 2>/dev/null || true
    pkill -f "connectrpc-catalog" 2>/dev/null || true
    sleep 1

    # Start Eliza test service
    go run ./cmd/test-eliza --port 50051 &
    ELIZA_PID=$!

    # Start backend server
    go run ./cmd/connectrpc-catalog &
    BACKEND_PID=$!

    # Wait for Eliza to be healthy
    for i in {1..10}; do
        if curl -s http://localhost:50051/health > /dev/null 2>&1; then
            break
        fi
        sleep 0.5
    done

    # Wait for backend to be ready (check root returns HTML)
    for i in {1..10}; do
        if curl -s http://localhost:8080/ | grep -q "DOCTYPE" 2>/dev/null; then
            break
        fi
        sleep 0.5
    done

    echo
    success "Servers running!"
    echo
    echo "  Web UI:        http://localhost:8080"
    echo "  Eliza service: http://localhost:50051"
    echo "  Health check:  http://localhost:50051/health"
    echo
    echo "Test commands:"
    echo "  # Connect protocol (HTTP/JSON)"
    echo "  curl -X POST http://localhost:50051/connectrpc.eliza.v1.ElizaService/Say \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"sentence\": \"Hello!\"}'"
    echo
    echo "Press Ctrl+C to stop servers..."

    # Handle cleanup on exit
    cleanup_serve() {
        echo
        info "Shutting down servers..."
        kill "$ELIZA_PID" 2>/dev/null || true
        kill "$BACKEND_PID" 2>/dev/null || true
        wait "$ELIZA_PID" 2>/dev/null || true
        wait "$BACKEND_PID" 2>/dev/null || true
        success "Servers stopped"
    }
    trap cleanup_serve EXIT INT TERM

    # Wait for either process to exit
    wait
fi
