#!/bin/bash

# 熔断器状态检查脚本
# 用途: 检查所有熔断器状态并生成报告

set -e

echo "=== 熔断器状态检查 ==="
echo ""

# 环境变量
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo "1. 检查熔断器状态..."
CB_STATUS=$(curl -s "$BACKEND_URL/circuit-breaker/status")
echo "   状态:"
echo "$CB_STATUS" | jq '.' 2>/dev/null || echo "$CB_STATUS"

echo ""
echo "2. 检查各熔断器详情..."

# 检查 GLM API 熔断器
GLM_STATE=$(echo "$CB_STATUS" | jq '.circuit_breakers.glm_api.state' 2>/dev/null || echo "unknown")
GLM_FAILURES=$(echo "$CB_STATUS" | jq '.circuit_breakers.glm_api.failure_count' 2>/dev/null || echo "0")
GLM_REQUESTS=$(echo "$CB_STATUS" | jq '.circuit_breakers.glm_api.request_count' 2>/dev/null || echo "0")

echo "   GLM API:"
echo "   - 状态: $GLM_STATE"
echo "   - 失败: $GLM_FAILURES"
echo "   - 请求: $GLM_REQUESTS"

# 检查数据库熔断器
DB_STATE=$(echo "$CB_STATUS" | jq '.circuit_breakers.database.state' 2>/dev/null || echo "unknown")
DB_FAILURES=$(echo "$CB_STATUS" | jq '.circuit_breakers.database.failure_count' 2>/dev/null || echo "0")
DB_REQUESTS=$(echo "$CB_STATUS" | jq '.circuit_breakers.database.request_count' 2>/dev/null || echo "0")

echo "   Database:"
echo "   - 状态: $DB_STATE"
echo "   - 失败: $DB_FAILURES"
echo "   - 请求: $DB_REQUESTS"

# 检查 Redis 熔断器
REDIS_STATE=$(echo "$CB_STATUS" | jq '.circuit_breakers.redis.state' 2>/dev/null || echo "unknown")
REDIS_FAILURES=$(echo "$CB_STATUS" | jq '.circuit_breakers.redis.failure_count' 2>/dev/null || echo "0")
REDIS_REQUESTS=$(echo "$CB_STATUS" | jq '.circuit_breakers.redis.request_count' 2>/dev/null || echo "0")

echo "   Redis:"
echo "   - 状态: $REDIS_STATE"
echo "   - 失败: $REDIS_FAILURES"
echo "   - 请求: $REDIS_REQUESTS"

echo ""
echo "3. 检查降级响应..."

# 检查降级请求比例
DEGRADED_COUNT=$(curl -s "$BACKEND_URL/metrics" | grep "http_requests_total{status=\"503\"}" | tail -1 | cut -d' ' -f2 || echo "0")
TOTAL_COUNT=$(curl -s "$BACKEND_URL/metrics" | grep "http_requests_total" | tail -1 | cut -d' ' -f2 || echo "0")

if [ "$TOTAL_COUNT" != "0" ] && [ -n "$TOTAL_COUNT" ]; then
    DEGRADED_RATE=$(awk "BEGIN {printf \"%.2f\", ($DEGRADED_COUNT / $TOTAL_COUNT) * 100}")
    echo "   降级请求比例: ${DEGRADED_RATE}%"
else
    echo "   无法计算降级比例"
fi

echo ""
echo "4. 检查服务依赖状态..."

# 检查 GLM API 连通性
GLM_API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://api.glm.ai/v1/ping" 2>/dev/null || echo "unavailable")
echo "   GLM API 连通性: HTTP $GLM_API_STATUS"

# 检查数据库连通性
DB_PING=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-postgres}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-industrial_ai}" -c "SELECT 1" 2>/dev/null || echo "unavailable")
if [ "$DB_PING" == "1" ]; then
    echo "   Database 连通性: ✓ OK"
else
    echo "   Database 连通性: ✗ Failed"
fi

# 检查 Redis 连通性
REDIS_PING=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT ping 2>/dev/null || echo "unavailable")
if [ "$REDIS_PING" == "PONG" ]; then
    echo "   Redis 连通性: ✓ OK"
else
    echo "   Redis 连通性: ✗ Failed"
fi

echo ""
echo "=== 状态报告 ==="
echo ""

# 生成健康报告
HEALTHY=true

if [ "$GLM_STATE" == "\"open\"" ]; then
    echo "[WARNING] GLM API 熔断器打开"
    HEALTHY=false
fi

if [ "$DB_STATE" == "\"open\"" ]; then
    echo "[CRITICAL] Database 熔断器打开"
    HEALTHY=false
fi

if [ "$REDIS_STATE" == "\"open\"" ]; then
    echo "[CRITICAL] Redis 熔断器打开"
    HEALTHY=false
fi

if $HEALTHY; then
    echo "[OK] 所有熔断器正常"
else
    echo ""
    echo "建议操作:"
    echo "- 检查失败服务状态"
    echo "- 使用 /circuit-breaker/:name/close 手动恢复"
    echo "- 查看详细日志定位问题"
fi

echo ""
echo "熔断器操作命令:"
echo "- 查看状态: curl $BACKEND_URL/circuit-breaker/status"
echo "- 强制打开: curl -X POST $BACKEND_URL/circuit-breaker/<name>/open"
echo "- 强制关闭: curl -X POST $BACKEND_URL/circuit-breaker/<name>/close"