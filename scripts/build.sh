#!/bin/bash
set -e

# Simple LazyArchon Build Script

BUILD_DIR=${BUILD_DIR:-"bin"}

echo "Building LazyArchon..."

# Create build directory
mkdir -p "$BUILD_DIR"

# Build for current platform
go build -o "$BUILD_DIR/lazyarchon" ./cmd/lazyarchon

echo "✓ Built: $BUILD_DIR/lazyarchon"