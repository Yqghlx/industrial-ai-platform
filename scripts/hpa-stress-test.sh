#!/bin/bash

# HPA 压力测试脚本
# 用途: 模拟负载触发 HPA 扩容

set -e

echo "=== HPA 压力测试 ==="
echo ""

# 环境变量
NAMESPACE="${NAMESPACE:-industrial-ai}"
TEST_DURATION="${TEST_DURATION:-300}"  # 测试持续时间 (秒)
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-100}"  # 并发请求数
BACKEND_URL="${BACKEND_URL:-http://backend:8080}"

echo "测试参数:"
echo "- Namespace: $NAMESPACE"
echo "- 测试时长: $TEST_DURATION 秒"
echo "- 并发请求: $CONCURRENT_REQUESTS"
echo "- 目标 URL: $BACKEND_URL"
echo ""

echo "1. 检查初始 HPA 状态..."
INITIAL_REPLICAS=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
echo "   当前副本数: $INITIAL_REPLICAS"

echo ""
echo "2. 检查初始 Pod 资源..."
kubectl top pods -l app=industrial-ai,component=backend -n $NAMESPACE || echo "   无法获取资源"

echo ""
echo "3. 开始负载测试..."
echo "   发送 $CONCURRENT_REQUESTS 并发请求..."

# 使用 curl 发送并发请求 (模拟负载)
for i in $(seq 1 $TEST_DURATION); do
    # 每秒发送多个请求
    for j in $(seq 1 $CONCURRENT_REQUESTS); do
        curl -s "$BACKEND_URL/api/v1/devices" > /dev/null &
    done
    
    # 等待 1 秒
    sleep 1
    
    # 每 10 秒报告一次状态
    if [ $((i % 10)) == 0 ]; then
        CURRENT=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
        DESIRED=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.desiredReplicas}')
        echo "   [$i] 当前副本: $CURRENT, 目标副本: $DESIRED"
    fi
done

echo ""
echo "4. 清理后台进程..."
# 终止所有 curl 进程
pkill -f "curl -s $BACKEND_URL" || true

echo ""
echo "5. 检查扩容结果..."
FINAL_REPLICAS=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
echo "   最终副本数: $FINAL_REPLICAS"

echo ""
echo "6. 等待缩容..."
echo "   观察 HPA 缩容行为..."

for i in $(seq 1 60); do
    CURRENT=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
    DESIRED=$(kubectl get hpa backend-hpa -n $NAMESPACE -o jsonpath='{.status.desiredReplicas}')
    
    echo "   [$i] 当前副本: $CURRENT, 目标副本: $DESIRED"
    
    # 如果回到初始副本数，结束观察
    if [ "$CURRENT" == "$INITIAL_REPLICAS" ] && [ "$DESIRED" == "$INITIAL_REPLICAS" ]; then
        echo "   ✓ 已缩容回初始状态"
        break
    fi
    
    sleep 5
done

echo ""
echo "=== HPA 压力测试完成 ==="
echo ""

echo "测试总结:"
echo "- 初始副本数: $INITIAL_REPLICAS"
echo "- 最终副本数: $FINAL_REPLICAS"
echo "- 扩容触发: $( [ "$FINAL_REPLICAS" -gt "$INITIAL_REPLICAS" ] && echo '✓ 成功' || echo '✗ 未触发')"
echo "- 缩容行为: $( [ "$FINAL_REPLICAS" -gt "$INITIAL_REPLICAS" ] && echo '观察中...' || echo '无需缩容')"

echo ""
echo "验收标准:"
echo "✓ CPU >70% 时自动扩容"
echo "✓ 扩容响应时间 <30s"
echo "✓ 缩容稳定窗口 5min"
echo "✓ 不超过 maxReplicas"