#!/bin/bash

# 健康检查脚本
# 用途: 执行多层级健康检查并生成报告

set -e

echo "=== 健康检查 ==="
echo ""

# 环境变量
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo "1. Level 1: 存活检查 (Liveness)..."
LIVE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/live")
if [ "$LIVE_STATUS" == "200" ]; then
    echo "   ✓ 应用存活 (HTTP $LIVE_STATUS)"
else
    echo "   ✗ 应用不存活 (HTTP $LIVE_STATUS)"
fi

echo ""
echo "2. Level 2: 就绪检查 (Readiness)..."
READY_RESPONSE=$(curl -s "$BACKEND_URL/health/ready")
READY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/ready")
echo "   HTTP Status: $READY_STATUS"
echo "   Response: $READY_RESPONSE"

if [ "$READY_STATUS" == "200" ]; then
    echo "   ✓ 应用就绪"
else
    echo "   ✗ 应用未就绪"
fi

echo ""
echo "3. Level 3: 详细健康检查..."
HEALTH_RESPONSE=$(curl -s "$BACKEND_URL/health")
echo "   Response:"
echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"

echo ""
echo "4. Level 4: 依赖深度检查..."
DEPS_RESPONSE=$(curl -s "$BACKEND_URL/health/dependencies")
echo "   Response:"
echo "$DEPS_RESPONSE" | jq '.' 2>/dev/null || echo "$DEPS_RESPONSE"

echo ""
echo "5. 数据库直接检查..."
PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h $DB_HOST -p $DB_PORT -U "${DB_USER:-postgres}" -d "${DB_NAME:-industrial_ai}" -c "
SELECT 
    'database' as component,
    'healthy' as status,
    pg_database_size(current_database()) as size_bytes,
    count(*) as table_count
FROM information_schema.tables 
WHERE table_schema = 'public';
" || echo "   ⚠️ 数据库连接失败"

echo ""
echo "6. Redis 直接检查..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping || echo "   ⚠️ Redis 连接失败"

echo ""
echo "7. 连接池状态..."
POOL_RESPONSE=$(curl -s "$BACKEND_URL/health" | jq '.checks.database.details' 2>/dev/null)
if [ -n "$POOL_RESPONSE" ]; then
    echo "   连接池状态:"
    echo "$POOL_RESPONSE"
else
    echo "   ⚠️ 无法获取连接池状态"
fi

echo ""
echo "8. 系统资源..."
SYS_RESPONSE=$(curl -s "$BACKEND_URL/health" | jq '.checks.system.details' 2>/dev/null)
if [ -n "$SYS_RESPONSE" ]; then
    echo "   系统状态:"
    echo "$SYS_RESPONSE"
else
    echo "   ⚠️ 无法获取系统状态"
fi

echo ""
echo "=== 健康检查完成 ==="
echo ""

# 生成健康报告
echo "健康报告摘要:"
echo "==============="
echo ""

# 计算健康状态
HEALTHY=true

if [ "$LIVE_STATUS" != "200" ]; then
    echo "[CRITICAL] 应用存活检查失败"
    HEALTHY=false
fi

if [ "$READY_STATUS" != "200" ]; then
    echo "[WARNING] 应用就绪检查失败"
    HEALTHY=false
fi

if $HEALTHY; then
    echo "[OK] 所有健康检查通过"
else
    echo "[ACTION] 需要采取行动"
fi

echo ""
echo "建议操作:"
echo "- 如果存活检查失败: 检查应用进程和日志"
echo "- 如果就绪检查失败: 检查数据库/Redis 连接"
echo "- 如果依赖检查失败: 检查具体依赖状态"