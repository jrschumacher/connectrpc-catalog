#!/bin/bash
set -e

echo "🔍 Wave 1 Validation Script"
echo "============================"
echo

# Check directory structure
echo "✓ Checking directory structure..."
for dir in cmd/connectrpc-catalog proto/catalog/v1 internal/loader internal/registry internal/invoker internal/server ui; do
  if [ -d "$dir" ]; then
    echo "  ✅ $dir exists"
  else
    echo "  ❌ $dir missing"
    exit 1
  fi
done
echo

# Check required files
echo "✓ Checking required files..."
for file in proto/catalog/v1/catalog.proto buf.yaml buf.gen.yaml go.mod .gitignore BUILD.md; do
  if [ -f "$file" ]; then
    echo "  ✅ $file exists"
  else
    echo "  ❌ $file missing"
    exit 1
  fi
done
echo

# Validate proto syntax
echo "✓ Validating proto definitions..."
if buf lint; then
  echo "  ✅ Proto definitions are valid"
else
  echo "  ❌ Proto validation failed"
  exit 1
fi
echo

# Check Go module
echo "✓ Checking Go module..."
if go mod download; then
  echo "  ✅ Go dependencies are valid"
else
  echo "  ❌ Go module error"
  exit 1
fi
echo

# Check buf CLI
echo "✓ Checking buf CLI..."
if command -v buf &> /dev/null; then
  BUF_VERSION=$(buf --version)
  echo "  ✅ buf CLI installed ($BUF_VERSION)"
else
  echo "  ❌ buf CLI not found"
  exit 1
fi
echo

# Check Go version
echo "✓ Checking Go version..."
GO_VERSION=$(go version)
echo "  ✅ $GO_VERSION"
echo

# Count proto RPCs
echo "✓ Checking proto completeness..."
RPC_COUNT=$(grep -c "rpc " proto/catalog/v1/catalog.proto || true)
if [ "$RPC_COUNT" -eq 4 ]; then
  echo "  ✅ All 4 RPCs defined"
else
  echo "  ❌ Expected 4 RPCs, found $RPC_COUNT"
  exit 1
fi
echo

echo "================================"
echo "✅ Wave 1 validation successful!"
echo "================================"
echo
echo "Next step: Run 'buf generate' to proceed to Wave 2"
