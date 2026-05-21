#!/bin/bash

# 启动恢复检查脚本
# 用途: 检查应用启动恢复状态

set -e

echo "=== 启动恢复检查 ==="
echo ""

# 环境变量
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo "1. 检查上次关闭状态..."
SHUTDOWN_STATE=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT GET "shutdown_state" 2>/dev/null || echo "")

if [ -n "$SHUTDOWN_STATE" ]; then
    echo "   上次关闭状态: $SHUTDOWN_STATE"
    echo "   解析关闭状态:"
    echo "$SHUTDOWN_STATE" | tr '|' '\n'
else
    echo "   无上次关闭状态 (正常启动)"
fi

echo ""
echo "2. 检查未完成任务..."
UNFINISHED_TASKS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT KEYS "task:*:interrupted" 2>/dev/null || echo "")

if [ -n "$UNFINISHED_TASKS" ]; then
    echo "   未完成任务:"
    echo "$UNFINISHED_TASKS"
else
    echo "   无未完成任务"
fi

echo ""
echo "3. 检查应用启动状态..."
STARTUP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/startup")
if [ "$STARTUP_STATUS" == "200" ]; then
    echo "   ✓ 应用启动完成 (HTTP $STARTUP_STATUS)"
else
    echo "   ⚠️ 应用启动中 (HTTP $STARTUP_STATUS)"
fi

echo ""
echo "4. 检查应用就绪状态..."
READY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/ready")
if [ "$READY_STATUS" == "200" ]; then
    echo "   ✓ 应用就绪 (HTTP $READY_STATUS)"
else
    echo "   ✗ 应用未就绪 (HTTP $READY_STATUS)"
fi

echo ""
echo "5. 检查恢复任务..."
if [ -n "$UNFINISHED_TASKS" ]; then
    echo "   检查任务恢复状态:"
    for TASK_KEY in $UNFINISHED_TASKS; do
        TASK_STATUS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT GET "$TASK_KEY")
        echo "   $TASK_KEY: $TASK_STATUS"
    done
fi

echo ""
echo "6. 检查缓存预热..."
CACHE_KEYS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT KEYS "cache:*" 2>/dev/null | wc -l)
echo "   缓存键数: $CACHE_KEYS"

if [ $CACHE_KEYS -gt 0 ]; then
    echo "   ✓ 缓存已预热"
else
    echo "   ⚠️ 缓存未预热"
fi

echo ""
echo "7. 清理关闭状态..."
if [ -n "$SHUTDOWN_STATE" ]; then
    redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL "shutdown_state"
    echo "   ✓ 关闭状态已清理"
fi

if [ -n "$UNFINISHED_TASKS" ]; then
    for TASK_KEY in $UNFINISHED_TASKS; do
        redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL "$TASK_KEY"
    done
    echo "   ✓ 未完成任务状态已清理"
fi

echo ""
echo "=== 启动恢复检查完成 ==="
echo ""

echo "恢复流程:"
echo "1. 检查上次关闭状态"
echo "2. 恢复未完成任务"
echo "3. 预热缓存数据"
echo "4. 健康检查就绪"
echo "5. 开始接收流量"