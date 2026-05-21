# Performance Benchmark Guide

## Overview

This benchmark suite tests the Industrial AI Agent Platform's performance under various load conditions using k6.

## Test Scenarios

| Test | Description | Target |
|------|-------------|--------|
| `api_load_test.js` | REST API load test | 100 concurrent users |
| `ws_stress_test.js` | WebSocket stress test | 200 device connections |
| `ai_throughput_test.js` | AI Agent throughput | 50 queries/sec |

## Prerequisites

### Install k6

```bash
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A36426D57D78446A6EBA16D3D
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows (using Chocolatey)
choco install k6
```

### Verify Installation

```bash
k6 version
```

## Running Benchmarks

### Quick Start

```bash
# Ensure backend is running
cd infra && docker-compose up -d && cd ..

# Run all benchmarks
./benchmarks/run_benchmarks.sh
```

### Individual Tests

```bash
# API load test
k6 run benchmarks/k6/api_load_test.js

# WebSocket stress test
k6 run benchmarks/k6/ws_stress_test.js

# AI throughput test
k6 run benchmarks/k6/ai_throughput_test.js

# Custom base URL
k6 run --env BASE_URL=http://your-server:8080 benchmarks/k6/api_load_test.js
```

### Output Options

```bash
# JSON output
k6 run --out json=results.json benchmarks/k6/api_load_test.js

# Summary export
k6 run --summary-export=summary.json benchmarks/k6/api_load_test.js

# InfluxDB output (for Grafana)
k6 run --out influxdb=http://localhost:8086/k6 benchmarks/k6/api_load_test.js
```

## Performance Thresholds

### Target Metrics

| Endpoint | P95 Latency | Error Rate |
|----------|-------------|------------|
| Health Check | < 50ms | < 0.1% |
| Device List | < 200ms | < 1% |
| Telemetry Submit | < 100ms | < 1% |
| AI Query | < 30s | < 5% |
| ROI Stats (cached) | < 100ms | < 1% |

### Expected Throughput

- **Devices API**: 1000 req/sec (read), 500 req/sec (write)
- **Telemetry**: 2000 data points/sec
- **WebSocket**: 500 concurrent connections
- **AI Agent**: 20 queries/sec (limited by LLM API)

## Interpreting Results

### Key Metrics

- **http_req_duration**: Request latency (min, avg, med, p90, p95, max)
- **http_reqs**: Request count and rate
- **http_req_failed**: Failed request rate
- **iterations**: Test iterations completed
- **vus**: Virtual users active

### Common Issues

| Issue | Possible Cause | Solution |
|-------|---------------|----------|
| High latency (>1s) | DB query slow | Add indexes, optimize queries |
| High error rate | Rate limiting | Adjust rate limits, add caching |
| Connection refused | Server overloaded | Scale horizontally |
| Timeout errors | Network issues | Check infrastructure |

## Continuous Benchmarking

### CI Integration

Add to GitHub Actions:

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmark

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday 2am
  workflow_dispatch:

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup k6
        run: |
          curl -L https://github.com/grafana/k6/releases/download/v0.45.0/k6-v0.45.0-linux-amd64.tar.gz | tar xz
          sudo mv k6-v0.45.0-linux-amd64/k6 /usr/local/bin/
      
      - name: Start backend
        run: |
          cd infra && docker-compose up -d
          sleep 30
      
      - name: Run benchmarks
        run: ./benchmarks/run_benchmarks.sh
      
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmarks/results/
```

## Grafana Dashboard

For real-time monitoring, configure k6 with InfluxDB + Grafana:

1. Start InfluxDB and Grafana:
   ```bash
   docker-compose -f infra/docker-compose.monitoring.yml up -d
   ```

2. Run benchmark with InfluxDB output:
   ```bash
   k6 run --out influxdb=http://localhost:8086/k6 benchmarks/k6/api_load_test.js
   ```

3. Import k6 dashboard in Grafana (ID: 2587)

---

**Benchmark Suite Version**: 1.0.0  
**Last Updated**: 2026-05-13