#!/bin/bash

# Debug LazyArchon with delve remote debugging
# Usage: ./debug.sh [port]

PORT=${1:-2345}

echo "Starting LazyArchon with delve remote debugging on port $PORT"
echo "Connect VS Code debugger to localhost:$PORT"
echo ""
echo "To connect:"
echo "1. Open VS Code"
echo "2. Go to Run and Debug (Ctrl+Shift+D)"
echo "3. Select 'Attach to running LazyArchon'"
echo "4. Press F5"
echo ""
echo "Press Ctrl+C to stop debugging"
echo ""

# Build and start with delve
go build -gcflags="all=-N -l" -o lazyarchon-debug cmd/lazyarchon/main.go

# Start delve in headless mode
dlv --listen=:$PORT --headless=true --api-version=2 --accept-multiclient exec ./lazyarchon-debug
