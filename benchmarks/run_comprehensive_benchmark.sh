#!/bin/bash
# Comprehensive Performance Benchmark Runner
# Industrial AI Platform - Performance Test Suite

set -e

# Configuration
PROJECT_DIR="/Users/yqgmac/yqg/project/industrial-ai-platform"
REPORT_DIR="${PROJECT_DIR}/docs"
BENCHMARK_DIR="${PROJECT_DIR}/benchmarks"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="${REPORT_DIR}/PERFORMANCE_BENCHMARK.md"
RESULTS_FILE="${BENCHMARK_DIR}/results/benchmark_${TIMESTAMP}.txt"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Industrial AI Platform - Performance Benchmark${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Timestamp: $(date)"
echo "Report Directory: ${REPORT_DIR}"
echo ""

# Create directories
mkdir -p "${REPORT_DIR}"
mkdir -p "${BENCHMARK_DIR}/results"

# Initialize results file
echo "Performance Benchmark Results - ${TIMESTAMP}" > "${RESULTS_FILE}"
echo "===========================================" >> "${RESULTS_FILE}"

# 1. Gateway Health Check
echo -e "\n${GREEN}=== 1. Gateway Health Check ===${NC}"
echo -e "\n=== 1. Gateway Health Check ===" >> "${RESULTS_FILE}"

health_response=$(curl -s http://localhost:80/health 2>/dev/null)
echo "Response: ${health_response}"

# Parse uptime
uptime=$(echo "${health_response}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uptime', 'N/A'))" 2>/dev/null || echo "N/A")
echo "Gateway Uptime: ${uptime} seconds"
echo "Gateway Uptime: ${uptime} seconds" >> "${RESULTS_FILE}"

# 2. Authentication Test
echo -e "\n${GREEN}=== 2. Authentication Test ===${NC}"
echo -e "\n=== 2. Authentication Test ===" >> "${RESULTS_FILE}"

LOGIN_URL="http://localhost:80/api/v1/auth/login"
LOGIN_DATA='{"username":"admin","password":"Admin@123456"}'

# Use python for curl measurement to avoid macOS compatibility issues
login_result=$(python3 << 'PYEOF'
import urllib.request
import json
import time

url = "http://localhost:80/api/v1/auth/login"
data = json.dumps({"username": "admin", "password": "Admin@123456"}).encode('utf-8')

start_time = time.time()
try:
    req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=10) as response:
        body = response.read().decode('utf-8')
        elapsed = time.time() - start_time
        status = response.status
        result = json.loads(body)
        token = result.get('token', result.get('access_token', result.get('data', {}).get('token', '')))
        print(f"SUCCESS|{status}|{elapsed:.3f}|{token}")
except Exception as e:
    elapsed = time.time() - start_time
    print(f"FAILED|0|{elapsed:.3f}|{str(e)}")
PYEOF
)

login_status=$(echo "${login_result}" | cut -d'|' -f1)
login_http_code=$(echo "${login_result}" | cut -d'|' -f2)
login_time=$(echo "${login_result}" | cut -d'|' -f3)
TOKEN=$(echo "${login_result}" | cut -d'|' -f4)

echo "Login Status: ${login_status}"
echo "HTTP Code: ${login_http_code}"
echo "Response Time: ${login_time}s"
echo "Login Response Time: ${login_time}s" >> "${RESULTS_FILE}"

if [ "${login_status}" = "SUCCESS" ] && [ -n "${TOKEN}" ]; then
    echo -e "Token: ${GREEN}Extracted Successfully${NC}"
    echo "Token: Extracted" >> "${RESULTS_FILE}"
else
    echo -e "Token: ${YELLOW}Not available${NC}"
    echo "Token: Not available" >> "${RESULTS_FILE}"
fi

# 3. API Response Time Tests
echo -e "\n${GREEN}=== 3. API Response Time Tests ===${NC}"
echo -e "\n=== 3. API Response Time Tests ===" >> "${RESULTS_FILE}"

# Function to measure API response times using Python
measure_api() {
    local url=$1
    local name=$2
    local method=${3:-GET}
    local auth_token=${4:-}
    local body=${5:-}
    
    python3 << PYEOF
import urllib.request
import json
import time
import ssl

url = "${url}"
method = "${method}"
token = "${auth_token}"
body_data = """${body}"""

headers = {'Content-Type': 'application/json'}
if token:
    headers['Authorization'] = f'Bearer {token}'

times = []
status_code = 0
for i in range(10):
    try:
        data = body_data.encode('utf-8') if body_data else None
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        start = time.time()
        with urllib.request.urlopen(req, timeout=15) as resp:
            resp.read()
            elapsed = time.time() - start
            times.append(elapsed)
            status_code = resp.status
    except Exception as e:
        elapsed = time.time() - start if 'start' in dir() else 0
        times.append(elapsed)
        status_code = 500

if times:
    sorted_times = sorted(times)
    avg = sum(times) / len(times)
    min_t = min(times)
    max_t = max(times)
    p50 = sorted_times[4] if len(sorted_times) >= 5 else sorted_times[-1]
    p95 = sorted_times[8] if len(sorted_times) >= 9 else sorted_times[-1]
    print(f"${name}|{status_code}|{min_t:.3f}|{avg:.3f}|{p50:.3f}|{p95:.3f}|{max_t:.3f}")
else:
    print(f"${name}|0|0.000|0.000|0.000|0.000|0.000")
PYEOF
}

# 3.1 Health Check
echo -e "\n${YELLOW}3.1 Health Check (Public)${NC}"
result=$(measure_api "http://localhost:80/health" "health")
echo "${result}" | awk -F'|' '{printf "  Status: %s\n  Min: %.3fs, Avg: %.3fs, P50: %.3fs, P95: %.3fs, Max: %.3fs\n", $2, $3, $4, $5, $6, $7}'
echo "Health API: ${result}" >> "${RESULTS_FILE}"

# 3.2 Device List
echo -e "\n${YELLOW}3.2 Device List API${NC}"
result=$(measure_api "http://localhost:80/api/v1/devices" "devices" "GET" "${TOKEN}")
echo "${result}" | awk -F'|' '{printf "  Status: %s\n  Min: %.3fs, Avg: %.3fs, P50: %.3fs, P95: %.3fs, Max: %.3fs\n", $2, $3, $4, $5, $6, $7}'
echo "Devices API: ${result}" >> "${RESULTS_FILE}"

# 3.3 Alerts List
echo -e "\n${YELLOW}3.3 Alerts List API${NC}"
result=$(measure_api "http://localhost:80/api/v1/alerts" "alerts" "GET" "${TOKEN}")
echo "${result}" | awk -F'|' '{printf "  Status: %s\n  Min: %.3fs, Avg: %.3fs, P50: %.3fs, P95: %.3fs, Max: %.3fs\n", $2, $3, $4, $5, $6, $7}'
echo "Alerts API: ${result}" >> "${RESULTS_FILE}"

# 3.4 Telemetry Latest
echo -e "\n${YELLOW}3.4 Telemetry Latest API${NC}"
result=$(measure_api "http://localhost:80/api/v1/telemetry/latest" "telemetry" "GET" "${TOKEN}")
echo "${result}" | awk -F'|' '{printf "  Status: %s\n  Min: %.3fs, Avg: %.3fs, P50: %.3fs, P95: %.3fs, Max: %.3fs\n", $2, $3, $4, $5, $6, $7}'
echo "Telemetry API: ${result}" >> "${RESULTS_FILE}"

# 3.5 Telemetry POST
echo -e "\n${YELLOW}3.5 Telemetry Submit API (POST)${NC}"
telemetry_body='{"device_id":"bench-test-001","device_type":"CNC","timestamp":"2026-05-25T13:00:00Z","metrics":{"temperature":75,"vibration":2.5},"status":"normal"}'
result=$(measure_api "http://localhost:80/api/v1/devices/telemetry" "telemetry_post" "POST" "" "${telemetry_body}")
echo "${result}" | awk -F'|' '{printf "  Status: %s\n  Min: %.3fs, Avg: %.3fs, P50: %.3fs, P95: %.3fs, Max: %.3fs\n", $2, $3, $4, $5, $6, $7}'
echo "Telemetry POST: ${result}" >> "${RESULTS_FILE}"

# 4. Database Performance Tests
echo -e "\n${GREEN}=== 4. Database Performance Tests ===${NC}"
echo -e "\n=== 4. Database Performance Tests ===" >> "${RESULTS_FILE}"

# 4.1 PostgreSQL
echo -e "\n${YELLOW}4.1 PostgreSQL Connection & Query Performance${NC}"

for db in iai-auth-db iai-device-db iai-telemetry-db iai-alert-db; do
    echo -n "  ${db}: "
    db_result=$(python3 << PYEOF
import subprocess
import time

db_name = "${db}"
try:
    start = time.time()
    result = subprocess.run(
        ['docker', 'exec', db_name, 'psql', '-U', 'postgres', '-c', 'SELECT 1'],
        capture_output=True, timeout=5
    )
    elapsed = time.time() - start
    if result.returncode == 0:
        print(f"Connected ({elapsed:.3f}s)")
    else:
        print(f"Error: {result.returncode}")
except Exception as e:
    print(f"Failed: {str(e)[:30]}")
PYEOF
)
    echo "${db_result}"
    echo "  ${db}: ${db_result}" >> "${RESULTS_FILE}"
done

# 4.2 Redis
echo -e "\n${YELLOW}4.2 Redis Performance${NC}"
redis_result=$(python3 << 'PYEOF'
import subprocess
import time

try:
    # Test connection
    start = time.time()
    result = subprocess.run(['redis-cli', '-h', 'localhost', '-p', '6379', 'ping'], 
                          capture_output=True, timeout=5)
    ping_time = time.time() - start
    
    if result.stdout.decode().strip() == 'PONG':
        print(f"Redis Status: Connected (ping: {ping_time*1000:.1f}ms)")
        
        # Get stats
        info_result = subprocess.run(['redis-cli', '-h', 'localhost', '-p', '6379', 'info', 'stats'],
                                    capture_output=True, timeout=5)
        stats = info_result.stdout.decode()
        
        for line in stats.split('\n'):
            if 'keyspace_hits' in line or 'keyspace_misses' in line or 'total_commands_processed' in line:
                print(line)
    else:
        print("Redis Status: Not responding")
except Exception as e:
    print(f"Redis Status: Failed - {str(e)[:30]}")
PYEOF
)
echo "${redis_result}"
echo "${redis_result}" >> "${RESULTS_FILE}"

# 5. Concurrent Load Tests
echo -e "\n${GREEN}=== 5. Concurrent Load Tests (k6) ===${NC}"
echo -e "\n=== 5. Concurrent Load Tests (k6) ===" >> "${RESULTS_FILE}"

cd "${BENCHMARK_DIR}/k6"

# Run simplified k6 test for quick benchmark
echo "Running k6 load test (30 seconds)..."

k6_output=$(k6 run --duration 30s --vus 10 \
    --out json=results/quick_benchmark_${TIMESTAMP}.json \
    --env BASE_URL=http://localhost:80 \
    --env TEST_USER=admin \
    --env TEST_PASS='Admin@123456' \
    comprehensive_benchmark.js 2>&1 || true)

echo "${k6_output}" | grep -E "checks|http_req_duration|iterations|vus|data_received|data_sent" || echo "k6 output captured"
echo "k6 test completed" >> "${RESULTS_FILE}"

# Extract key metrics from k6 output
echo "${k6_output}" | tail -50 > "${BENCHMARK_DIR}/results/k6_summary_${TIMESTAMP}.txt"

# 6. WebSocket Test
echo -e "\n${GREEN}=== 6. WebSocket Endpoint Test ===${NC}"
echo -e "\n=== 6. WebSocket Endpoint Test ===" >> "${RESULTS_FILE}"

ws_result=$(python3 << 'PYEOF'
import urllib.request
import time

try:
    start = time.time()
    req = urllib.request.Request("http://localhost:80/ws")
    with urllib.request.urlopen(req, timeout=5) as resp:
        elapsed = time.time() - start
        print(f"WebSocket endpoint HTTP: {resp.status} ({elapsed:.3f}s)")
except urllib.error.HTTPError as e:
    elapsed = time.time() - start
    # 400/426 are expected for WebSocket without upgrade headers
    if e.code in [400, 426]:
        print(f"WebSocket endpoint: OK (HTTP {e.code}, expected for non-WS request, {elapsed:.3f}s)")
    else:
        print(f"WebSocket endpoint: HTTP {e.code}")
except Exception as e:
    print(f"WebSocket test: Failed - {str(e)[:30]}")
PYEOF
)
echo "${ws_result}"
echo "${ws_result}" >> "${RESULTS_FILE}"

# 7. Service Summary
echo -e "\n${GREEN}=== 7. Service Health Summary ===${NC}"
echo -e "\n=== 7. Service Health Summary ===" >> "${RESULTS_FILE}"

# Check services
check_service() {
    local port=$1
    local name=$2
    
    python3 << PYEOF
import urllib.request
import time

try:
    start = time.time()
    req = urllib.request.Request("http://localhost:${port}/health")
    with urllib.request.urlopen(req, timeout=3) as resp:
        elapsed = time.time() - start
        print(f"  ${name}: Healthy ({elapsed*1000:.0f}ms)")
except Exception:
    # Try TCP connection
    import socket
    try:
        start = time.time()
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(2)
        s.connect(('localhost', ${port}))
        elapsed = time.time() - start
        s.close()
        print(f"  ${name}: Port Open ({elapsed*1000:.0f}ms)")
    except Exception:
        print(f"  ${name}: Not Responding")
PYEOF
}

check_service 80 "Gateway"
check_service 8004 "Auth Service"
check_service 8001 "Device Service"
check_service 6379 "Redis"
check_service 5432 "PostgreSQL"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Benchmark completed at $(date)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Results saved to: ${RESULTS_FILE}"