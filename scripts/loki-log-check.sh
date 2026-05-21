#!/bin/bash

# Loki 日志检查脚本
# 用途: 检查 Loki 日志系统状态并生成报告

set -e

echo "=== Loki 日志系统状态检查 ==="
echo ""

# 环境变量
LOKI_URL="${LOKI_URL:-http://localhost:3100}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3001}"
SERVICE_NAME="${SERVICE_NAME:-industrial-ai-backend}"

echo "1. 检查 Loki 服务状态..."
LOKI_READY=$(curl -s "$LOKI_URL/ready" || echo "down")
if [ "$LOKI_READY" == "ready" ]; then
    echo "   ✓ Loki 服务正常"
else
    echo "   ✗ Loki 服务异常: $LOKI_READY"
fi

echo ""
echo "2. 检查 Loki 健康状态..."
LOKI_HEALTH=$(curl -s "$LOKI_URL/health" || echo "unhealthy")
echo "   Loki 健康状态: $LOKI_HEALTH"

echo ""
echo "3. 检查 Loki 指标..."
LOKI_METRICS=$(curl -s "$LOKI_URL/metrics" | grep -E "loki_ingester_chunks_total|loki_query_requests_total" | head -5 || echo "no metrics")
echo "$LOKI_METRICS"

echo ""
echo "4. 检查 Promtail 状态..."
PROMTAIL_STATUS=$(docker ps --filter "name=promtail" --format "{{.Status}}" || echo "not running")
echo "   Promtail 状态: $PROMTAIL_STATUS"

echo ""
echo "5. 检查日志接收状态..."
# 查询最近 5 分钟的日志数量
LOG_COUNT=$(curl -s -G "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode "query=count_over_time({service=\"$SERVICE_NAME\"}[5m])" \
    --data-urlencode "start=$(date -d '5 minutes ago' +%s)000000000" \
    --data-urlencode "end=$(date +%s)000000000" \
    --data-urlencode "step=60" | jq '.data.result[0].values[-1][1]' || echo "0")

echo "   最近 5 分钟日志数: $LOG_COUNT"

echo ""
echo "6. 检查错误日志..."
ERROR_COUNT=$(curl -s -G "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode "query=count_over_time({service=\"$SERVICE_NAME\",level=\"error\"}[5m])" \
    --data-urlencode "start=$(date -d '5 minutes ago' +%s)000000000" \
    --data-urlencode "end=$(date +%s)000000000" \
    --data-urlencode "step=60" | jq '.data.result[0].values[-1][1]' || echo "0")

echo "   最近 5 分钟错误日志数: $ERROR_COUNT"

if [ "$ERROR_COUNT" -gt 0 ]; then
    echo "   ⚠️ 存在错误日志，请检查"
fi

echo ""
echo "7. 检查 HTTP 请求日志..."
HTTP_REQUEST_COUNT=$(curl -s -G "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode "query=count_over_time({service=\"$SERVICE_NAME\"} | json | http_method [5m])" \
    --data-urlencode "start=$(date -d '5 minutes ago' +%s)000000000" \
    --data-urlencode "end=$(date +%s)000000000" \
    --data-urlencode "step=60" | jq '.data.result | length' || echo "0")

echo "   HTTP 请求日志类型数: $HTTP_REQUEST_COUNT"

echo ""
echo "8. 检查慢请求日志..."
SLOW_REQUEST_COUNT=$(curl -s -G "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode "query=count_over_time({service=\"$SERVICE_NAME\"} | json | http_latency > 1000 [5m])" \
    --data-urlencode "start=$(date -d '5 minutes ago' +%s)000000000" \
    --data-urlencode "end=$(date +%s)000000000" \
    --data-urlencode "step=60" | jq '.data.result[0].values[-1][1]' || echo "0")

echo "   慢请求 (>1s) 数: $SLOW_REQUEST_COUNT"

if [ "$SLOW_REQUEST_COUNT" -gt 10 ]; then
    echo "   ⚠️ 慢请求较多，性能可能有问题"
fi

echo ""
echo "9. 查看最近日志样本..."
RECENT_LOGS=$(curl -s -G "$LOKI_URL/loki/api/v1/query" \
    --data-urlencode "query={service=\"$SERVICE_NAME\"} | json | line_format \"{{.level}} {{.message}}\"" \
    --data-urlencode "limit=5" | jq '.data.result[0].values[][1]' || echo "no logs")

echo "   最近日志样本:"
echo "$RECENT_LOGS" | head -5

echo ""
echo "10. 检查 Grafana 连接..."
# 使用环境变量进行认证（避免硬编码）
GRAFANA_CREDS="${GRAFANA_ADMIN_USER:-admin}:${GRAFANA_ADMIN_PASSWORD:-admin}"
GRAFANA_DATASOURCE=$(curl -s "$GRAFANA_URL/api/datasources" -u "$GRAFANA_CREDS" | jq '.[] | select(.type=="loki") | .name' || echo "not configured")
echo "   Grafana Loki 数据源: $GRAFANA_DATASOURCE"

echo ""
echo "=== Loki 日志系统健康报告 ==="
echo ""

# 生成健康报告
HEALTHY=true

# 检查 Loki 服务
if [ "$LOKI_READY" != "ready" ]; then
    echo "[CRITICAL] Loki 服务异常"
    HEALTHY=false
fi

# 检查日志接收
if [ "$LOG_COUNT" == "0" ]; then
    echo "[WARNING] 没有收到日志"
    HEALTHY=false
fi

# 检查错误日志
if [ "$ERROR_COUNT" -gt 100 ]; then
    echo "[WARNING] 错误日志过多: $ERROR_COUNT"
    HEALTHY=false
fi

# 检查慢请求
if [ "$SLOW_REQUEST_COUNT" -gt 10 ]; then
    echo "[WARNING] 慢请求过多: $SLOW_REQUEST_COUNT"
    HEALTHY=false
fi

if $HEALTHY; then
    echo "[OK] Loki 日志系统状态正常"
else
    echo "[ACTION] 需要采取行动"
fi

echo ""
echo "操作命令:"
echo "- 查看 Grafana: http://$GRAFANA_URL"
echo "- Loki API: $LOKI_URL"
echo "- 查询日志: curl -G $LOKI_URL/loki/api/v1/query --data-urlencode 'query={service=\"$SERVICE_NAME\"}'"
echo "- 查看错误: curl -G $LOKI_URL/loki/api/v1/query --data-urlencode 'query={service=\"$SERVICE_NAME\",level=\"error\"}'"