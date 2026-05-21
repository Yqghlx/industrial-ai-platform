#!/bin/bash

# Run device_repo tests
echo "Running device_repo tests..."
cd /Users/yqgvirtualmacos/Projects/industrial-ai-platform/backend
go test -v ./internal/repository -run TestDeviceRepository

echo ""
echo "Test execution complete!"