#!/bin/bash
set -e

echo "🧪 ConnectRPC Catalog - Integration Tests"
echo "==========================================="
echo

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running integration tests...${NC}"
echo

# Run integration tests with verbose output
go test -v ./internal/server/... -run TestIntegration

echo
echo -e "${GREEN}✅ All integration tests passed!${NC}"
echo
echo "Test coverage:"
go test -cover ./internal/server/... -run TestIntegration
