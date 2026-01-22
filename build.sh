#!/bin/bash

set -e

echo "Building ConnectRPC Catalog..."

# Step 0: Generate protobuf code
echo ""
echo "Step 0: Generating protobuf code..."
make gen

# Step 1: Build the UI
echo ""
echo "Step 1: Building UI..."
cd ui

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing UI dependencies..."
    npm install
fi

# Build the UI
echo "Building UI assets..."
npm run build

cd ..

# Verify UI dist directory exists
if [ ! -d "ui/dist" ]; then
    echo "Error: ui/dist directory not found after build"
    exit 1
fi

# Copy UI assets to Go embed location
echo "Copying UI assets to embed location..."
rm -rf cmd/connectrpc-catalog/dist
cp -r ui/dist cmd/connectrpc-catalog/dist

# Step 2: Build the Go binary
echo ""
echo "Step 2: Building Go binary..."

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build with embedded UI
go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o bin/connectrpc-catalog \
    ./cmd/connectrpc-catalog

echo ""
echo "Build complete! Binary created at: bin/connectrpc-catalog"
echo "Run with: ./bin/connectrpc-catalog"
