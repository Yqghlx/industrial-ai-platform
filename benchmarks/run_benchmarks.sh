#!/bin/bash

# Performance Benchmark Runner
# Industrial AI Agent Platform

set -e

echo "=========================================="
echo "  Industrial AI Platform - Benchmark Suite"
echo "=========================================="

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
RESULTS_DIR="benchmarks/results"
DATE=$(date +%Y%m%d_%H%M%S)

# Create results directory
mkdir -p "$RESULTS_DIR/$DATE"

echo ""
echo "Base URL: $BASE_URL"
echo "Results: $RESULTS_DIR/$DATE"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo "ERROR: k6 is not installed"
    echo ""
    echo "Install k6:"
    echo "  macOS:   brew install k6"
    echo "  Linux:   https://k6.io/docs/getting-started/installation/"
    echo ""
    exit 1
fi

echo "k6 version: $(k6 version)"
echo ""

# Function to run a benchmark
run_benchmark() {
    local name=$1
    local script=$2
    
    echo "----------------------------------------"
    echo "Running: $name"
    echo "----------------------------------------"
    
    k6 run \
        --env BASE_URL="$BASE_URL" \
        --out json="$RESULTS_DIR/$DATE/${name}.json" \
        --summary-export="$RESULTS_DIR/$DATE/${name}_summary.json" \
        "$script"
    
    echo ""
    echo "Results saved to: $RESULTS_DIR/$DATE/${name}.json"
    echo ""
}

# Ensure backend is running
echo "Checking backend health..."
if ! curl -s "$BASE_URL/health" | grep -q "healthy"; then
    echo ""
    echo "WARNING: Backend may not be running at $BASE_URL"
    echo "Starting backend with Docker Compose..."
    echo ""
    
    cd infra && docker-compose up -d && cd ..
    
    echo "Waiting for backend to start..."
    sleep 10
    
    if ! curl -s "$BASE_URL/health" | grep -q "healthy"; then
        echo "ERROR: Backend failed to start"
        exit 1
    fi
    
    echo "Backend started successfully"
fi

echo ""
echo "=========================================="
echo "  Starting Benchmark Tests"
echo "=========================================="
echo ""

# Run benchmarks
run_benchmark "api_load" "benchmarks/k6/api_load_test.js"
run_benchmark "ws_stress" "benchmarks/k6/ws_stress_test.js"
run_benchmark "ai_throughput" "benchmarks/k6/ai_throughput_test.js"

# Generate summary report
echo "=========================================="
echo "  Generating Summary Report"
echo "=========================================="
echo ""

python3 benchmarks/generate_report.py "$RESULTS_DIR/$DATE" > "$RESULTS_DIR/$DATE/benchmark_report.md"

echo ""
echo "Benchmark completed!"
echo ""
echo "View results:"
echo "  $RESULTS_DIR/$DATE/benchmark_report.md"
echo ""