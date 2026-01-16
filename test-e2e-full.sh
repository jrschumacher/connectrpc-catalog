#!/usr/bin/env bash

# Full-stack E2E test automation script
# Starts backend, waits for health, runs Playwright tests, captures exit code, stops backend

set -e

# Configuration
BACKEND_HOST="localhost"
BACKEND_PORT="8080"
BACKEND_URL="http://${BACKEND_HOST}:${BACKEND_PORT}"
HEALTH_CHECK_URL="${BACKEND_URL}/catalog.v1.CatalogService/ListServices"
ELIZA_PORT="50051"
ELIZA_HEALTH_URL="http://localhost:${ELIZA_PORT}/health"
MAX_WAIT=30
WAIT_INTERVAL=1
BACKEND_PID=""
ELIZA_PID=""
TEST_EXIT_CODE=0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function - ensures all servers are stopped
cleanup() {
    if [ -n "$ELIZA_PID" ]; then
        log_info "Stopping Eliza test server (PID: $ELIZA_PID)..."
        kill "$ELIZA_PID" 2>/dev/null || true
        wait "$ELIZA_PID" 2>/dev/null || true
        log_success "Eliza test server stopped"
    fi

    if [ -n "$BACKEND_PID" ]; then
        log_info "Stopping backend server (PID: $BACKEND_PID)..."
        kill "$BACKEND_PID" 2>/dev/null || true
        wait "$BACKEND_PID" 2>/dev/null || true
        log_success "Backend server stopped"
    fi
}

# Register cleanup on script exit
trap cleanup EXIT INT TERM

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if Go binary exists or can be built
    if [ ! -f "./bin/connectrpc-catalog" ]; then
        log_warning "Backend binary not found. Building..."
        if ! make build; then
            log_error "Failed to build backend binary"
            exit 1
        fi
        log_success "Backend binary built"
    fi

    # Check if Playwright is installed
    if ! (cd ui && npx playwright --version >/dev/null 2>&1); then
        log_warning "Playwright not found. Installing dependencies..."
        cd ui
        npm install
        npx playwright install --with-deps chromium
        cd ..
        log_success "Playwright installed"
    fi

    log_success "Prerequisites check complete"
}

# Start backend server
start_backend() {
    log_info "Starting backend server on ${BACKEND_URL}..."

    # Start backend in background, redirecting output to file
    ./bin/connectrpc-catalog \
        --host="$BACKEND_HOST" \
        --port="$BACKEND_PORT" \
        > backend-test.log 2>&1 &

    BACKEND_PID=$!

    log_info "Backend server started with PID: $BACKEND_PID"
}

# Start Eliza test server
start_eliza() {
    log_info "Starting Eliza test server on port ${ELIZA_PORT}..."

    # Build and start Eliza test server
    go run ./cmd/test-eliza --port="$ELIZA_PORT" \
        > eliza-test.log 2>&1 &

    ELIZA_PID=$!

    log_info "Eliza test server started with PID: $ELIZA_PID"
}

# Wait for Eliza to be healthy
wait_for_eliza() {
    log_info "Waiting for Eliza test server to be healthy (max ${MAX_WAIT}s)..."

    local elapsed=0
    while [ $elapsed -lt $MAX_WAIT ]; do
        # Try to connect to the health check endpoint
        if curl -s -f "$ELIZA_HEALTH_URL" >/dev/null 2>&1; then
            log_success "Eliza test server is healthy after ${elapsed}s"
            return 0
        fi

        # Check if process is still running
        if ! kill -0 "$ELIZA_PID" 2>/dev/null; then
            log_error "Eliza test server process died during startup"
            log_error "Eliza logs:"
            cat eliza-test.log
            return 1
        fi

        sleep $WAIT_INTERVAL
        elapsed=$((elapsed + WAIT_INTERVAL))
    done

    log_error "Eliza test server did not become healthy within ${MAX_WAIT}s"
    log_error "Eliza logs:"
    cat eliza-test.log
    return 1
}

# Wait for backend to be healthy
wait_for_backend() {
    log_info "Waiting for backend to be healthy (max ${MAX_WAIT}s)..."

    local elapsed=0
    while [ $elapsed -lt $MAX_WAIT ]; do
        # Try to connect to the health check endpoint
        if curl -s -f -X POST \
            -H "Content-Type: application/json" \
            -d '{}' \
            "$HEALTH_CHECK_URL" >/dev/null 2>&1; then
            log_success "Backend is healthy after ${elapsed}s"
            return 0
        fi

        # Check if backend process is still running
        if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
            log_error "Backend process died during startup"
            log_error "Backend logs:"
            cat backend-test.log
            return 1
        fi

        sleep $WAIT_INTERVAL
        elapsed=$((elapsed + WAIT_INTERVAL))
    done

    log_error "Backend did not become healthy within ${MAX_WAIT}s"
    log_error "Backend logs:"
    cat backend-test.log
    return 1
}

# Run Playwright E2E tests
run_tests() {
    log_info "Running Playwright full-stack E2E tests..."

    cd ui

    # Run full-stack tests using the fullstack config
    if npx playwright test --config=playwright.fullstack.config.ts; then
        TEST_EXIT_CODE=0
        log_success "All full-stack E2E tests passed"
    else
        TEST_EXIT_CODE=$?
        log_error "Full-stack E2E tests failed with exit code: $TEST_EXIT_CODE"
    fi

    cd ..

    return $TEST_EXIT_CODE
}

# Main execution flow
main() {
    log_info "==== Full-Stack E2E Test Automation ===="
    log_info ""

    # Step 1: Check prerequisites
    check_prerequisites

    # Step 2: Start Eliza test server (target gRPC service)
    start_eliza

    # Step 3: Wait for Eliza to be healthy
    if ! wait_for_eliza; then
        log_error "Eliza test server health check failed"
        exit 1
    fi

    # Step 4: Start backend server
    start_backend

    # Step 5: Wait for backend to be healthy
    if ! wait_for_backend; then
        log_error "Backend health check failed"
        exit 1
    fi

    # Step 6: Run E2E tests
    run_tests

    # Step 5: Display test results
    log_info ""
    log_info "==== Test Results ===="
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        log_success "✓ All E2E tests passed successfully"
    else
        log_error "✗ E2E tests failed"
    fi

    # Step 6: Cleanup (handled by trap)
    log_info ""
    log_info "==== Cleanup ===="

    # Exit with test exit code
    exit $TEST_EXIT_CODE
}

# Run main function
main
