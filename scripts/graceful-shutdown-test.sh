#!/bin/bash

# 优雅关闭测试脚本
# 用途: 测试应用优雅关闭行为

set -e

echo "=== 优雅关闭测试 ==="
echo ""

# 环境变量
BACKEND_PID="${BACKEND_PID:-}"
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
SHUTDOWN_TIMEOUT="${SHUTDOWN_TIMEOUT:-30}"

echo "1. 检查应用运行状态..."
if [ -z "$BACKEND_PID" ]; then
    echo "   ⚠️ 请设置 BACKEND_PID 环境变量"
    echo "   示例: export BACKEND_PID=$(pgrep -f 'backend')"
    exit 1
fi

if ! ps -p $BACKEND_PID > /dev/null; then
    echo "   ✗ 进程 $BACKEND_PID 不存在"
    exit 1
fi

echo "   ✓ 应用运行中 (PID: $BACKEND_PID)"

echo ""
echo "2. 检查 Readiness 状态..."
READY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/ready")
if [ "$READY_STATUS" == "200" ]; then
    echo "   ✓ 应用就绪 (HTTP $READY_STATUS)"
else
    echo "   ✗ 应用未就绪 (HTTP $READY_STATUS)"
fi

echo ""
echo "3. 发送活跃请求..."
# 发送长请求 (模拟活跃请求)
curl -s "$BACKEND_URL/api/v1/devices" &
REQUEST_PID=$!
echo "   ✓ 请求已发送 (PID: $REQUEST_PID)"

sleep 1

echo ""
echo "4. 发送关闭信号 (SIGTERM)..."
kill -TERM $BACKEND_PID
echo "   ✓ SIGTERM 已发送"

echo ""
echo "5. 监控关闭过程..."
echo "   等待应用优雅关闭 (最多 $SHUTDOWN_TIMEOUT 秒)..."

# 监控 Readiness 变化
for i in $(seq 1 $SHUTDOWN_TIMEOUT); do
    READY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health/ready")
    
    if [ "$READY_STATUS" != "200" ]; then
        echo "   [$i] Readiness 变为 $READY_STATUS (不再接收新请求)"
    fi
    
    # 检查进程是否存活
    if ! ps -p $BACKEND_PID > /dev/null; then
        echo "   [$i] 进程已退出"
        break
    fi
    
    sleep 1
done

echo ""
echo "6. 检查请求处理..."
if ps -p $REQUEST_PID > /dev/null; then
    echo "   ⚠️ 请求仍在处理 (PID: $REQUEST_PID)"
    kill $REQUEST_PID 2>/dev/null
else
    echo "   ✓ 请求已完成"
fi

echo ""
echo "7. 检查关闭日志..."
# 检查最近的关闭日志 (如果可用)
if [ -f "/var/log/backend.log" ]; then
    echo "   最近日志:"
    tail -20 /var/log/backend.log | grep -i "shutdown" || echo "   无关闭日志"
fi

echo ""
echo "=== 优雅关闭测试完成 ==="
echo ""

echo "验收标准:"
echo "✓ 收到 SIGTERM 信号"
echo "✓ Readiness 变为 not_ready"
echo "✓ 等待现有请求完成"
echo "✓ 进程正常退出 (退出码 0)"
echo "✓ 关闭日志记录完整"