#!/bin/bash

# 分布式追踪检查脚本
# 用途: 检查 OpenTelemetry/Jaeger/Tempo 追踪系统状态

set -e

echo "=== 分布式追踪系统状态检查 ==="
echo ""

# 环境变量
JAEGER_URL="${JAEGER_URL:-http://localhost:16686}"
TEMPO_URL="${TEMPO_URL:-http://localhost:3200}"
OTEL_URL="${OTEL_URL:-http://localhost:8888}"
SERVICE_NAME="${SERVICE_NAME:-industrial-ai-backend}"

echo "1. 检查 OpenTelemetry Collector 状态..."
OTEL_HEALTH=$(curl -s "$OTEL_URL/health" || echo "down")
if [ "$OTEL_HEALTH" == "OK" ]; then
    echo "   ✓ OTEL Collector 正常"
else
    echo "   ✗ OTEL Collector 异常: $OTEL_HEALTH"
fi

echo ""
echo "2. 检查 Jaeger 服务状态..."
JAEGER_HEALTH=$(curl -s "$JAEGER_URL/health" || echo "down")
if [ "$JAEGER_HEALTH" == "OK" ]; then
    echo "   ✓ Jaeger 正常"
else
    echo "   ✗ Jaeger 异常: $JAEGER_HEALTH"
fi

echo ""
echo "3. 检查 Tempo 服务状态..."
TEMPO_READY=$(curl -s "$TEMPO_URL/ready" || echo "down")
if [ "$TEMPO_READY" == "ready" ]; then
    echo "   ✓ Tempo 正常"
else
    echo "   ✗ Tempo 异常: $TEMPO_READY"
fi

echo ""
echo "4. 检查 OTEL Collector 指标..."
OTEL_METRICS=$(curl -s "$OTEL_URL/metrics" | grep -E "otelcol_receiver_accepted_spans|otelcol_exporter_sent_spans" | head -5 || echo "no metrics")
echo "$OTEL_METRICS"

echo ""
echo "5. 检查 Jaeger 服务列表..."
JAEGER_SERVICES=$(curl -s "$JAEGER_URL/api/services" | jq '.data[]' || echo "no services")
echo "$JAEGER_SERVICES" | head -10

echo ""
echo "6. 检查最近追踪数据..."
# 查询最近 5 分钟的追踪
TRACES_COUNT=$(curl -s -G "$JAEGER_URL/api/traces" \
    --data-urlencode "service=$SERVICE_NAME" \
    --data-urlencode "limit=10" \
    | jq '.data | length' || echo "0")

echo "   最近追踪数量: $TRACES_COUNT"

if [ "$TRACES_COUNT" -gt 0 ]; then
    echo "   ✓ 追踪数据正常接收"
else
    echo "   ⚠️ 暂无追踪数据"
fi

echo ""
echo "7. 检查追踪错误..."
ERROR_TRACES=$(curl -s -G "$JAEGER_URL/api/traces" \
    --data-urlencode "service=$SERVICE_NAME" \
    --data-urlencode "lookback=5m" \
    --data-urlencode "tags={\"error\":\"true\"}" \
    | jq '.data | length' || echo "0")

echo "   错误追踪数: $ERROR_TRACES"

if [ "$ERROR_TRACES" -gt 10 ]; then
    echo "   ⚠️ 错误追踪较多，请检查"
fi

echo ""
echo "8. 检查慢追踪..."
SLOW_TRACES=$(curl -s -G "$JAEGER_URL/api/traces" \
    --data-urlencode "service=$SERVICE_NAME" \
    --data-urlencode "lookback=5m" \
    --data-urlencode "minDuration=500ms" \
    | jq '.data | length' || echo "0")

echo "   慢追踪 (>500ms) 数: $SLOW_TRACES"

if [ "$SLOW_TRACES" -gt 5 ]; then
    echo "   ⚠️ 慢追踪较多，性能可能有问题"
fi

echo ""
echo "9. 检查追踪详情样本..."
if [ "$TRACES_COUNT" -gt 0 ]; then
    TRACE_ID=$(curl -s -G "$JAEGER_URL/api/traces" \
        --data-urlencode "service=$SERVICE_NAME" \
        --data-urlencode "limit=1" \
        | jq '.data[0].traceID' || echo "")
    
    if [ "$TRACE_ID" != "" ]; then
        echo "   最新追踪 ID: $TRACE_ID"
        
        # 查询追踪详情
        TRACE_DETAIL=$(curl -s "$JAEGER_URL/api/traces/$TRACE_ID" | jq '.data[0].spans | length' || echo "0")
        echo "   Span 数量: $TRACE_DETAIL"
    fi
fi

echo ""
echo "10. 检查 Grafana Tempo 数据源..."
# 使用环境变量进行认证（避免硬编码）
GRAFANA_CREDS="${GRAFANA_ADMIN_USER:-admin}:${GRAFANA_ADMIN_PASSWORD:-admin}"
GRAFANA_DATASOURCE=$(curl -s "http://localhost:3002/api/datasources" -u "$GRAFANA_CREDS" | jq '.[] | select(.type=="tempo") | .name' || echo "not configured")
echo "   Grafana Tempo 数据源: $GRAFANA_DATASOURCE"

echo ""
echo "=== 分布式追踪系统健康报告 ==="
echo ""

# 生成健康报告
HEALTHY=true

# 检查 OTEL Collector
if [ "$OTEL_HEALTH" != "OK" ]; then
    echo "[CRITICAL] OpenTelemetry Collector 异常"
    HEALTHY=false
fi

# 检查 Jaeger
if [ "$JAEGER_HEALTH" != "OK" ]; then
    echo "[CRITICAL] Jaeger 异常"
    HEALTHY=false
fi

# 检查 Tempo
if [ "$TEMPO_READY" != "ready" ]; then
    echo "[WARNING] Tempo 异常"
    HEALTHY=false
fi

# 检查追踪数据
if [ "$TRACES_COUNT" == "0" ]; then
    echo "[WARNING] 没有追踪数据"
    HEALTHY=false
fi

# 检查错误追踪
if [ "$ERROR_TRACES" -gt 10 ]; then
    echo "[WARNING] 错误追踪较多: $ERROR_TRACES"
    HEALTHY=false
fi

# 检查慢追踪
if [ "$SLOW_TRACES" -gt 5 ]; then
    echo "[WARNING] 慢追踪较多: $SLOW_TRACES"
    HEALTHY=false
fi

if $HEALTHY; then
    echo "[OK] 分布式追踪系统状态正常"
else
    echo "[ACTION] 需要采取行动"
fi

echo ""
echo "操作命令:"
echo "- 查看 Jaeger UI: http://$JAEGER_URL"
echo "- 查看 Tempo API: $TEMPO_URL"
echo "- 查看 OTEL Metrics: $OTEL_URL/metrics"
echo "- 查询追踪: curl -G $JAEGER_URL/api/traces --data-urlencode 'service=$SERVICE_NAME'"
echo "- 查询错误: curl -G $JAEGER_URL/api/traces --data-urlencode 'service=$SERVICE_NAME' --data-urlencode 'tags={\"error\":\"true\"}'"