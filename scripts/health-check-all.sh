#!/bin/bash

# Industrial AI Platform 综合健康检查脚本
# 用途: 检查所有系统组件健康状态

set -e

echo "=== Industrial AI Platform 综合健康检查 ==="
echo ""

# ============================================
# 配置参数
# ============================================

NAMESPACE="${NAMESPACE:-industrial-ai}"
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_TIME=$(date +%H%M%S)

echo "检查参数:"
echo "- Namespace: $NAMESPACE"
echo "- 时间: $REPORT_DATE $REPORT_TIME"
echo ""

# ============================================
# 健康检查函数
# ============================================

check_service() {
    local name=$1
    local url=$2
    local expected=$3
    
    result=$(curl -s $url 2>/dev/null || echo "down")
    
    if echo "$result" | grep -q "$expected"; then
        echo "✓ $name: 正常"
        return 0
    else
        echo "✗ $name: 异常 ($result)"
        return 1
    fi
}

check_port() {
    local name=$1
    local host=$2
    local port=$3
    
    if nc -z $host $port 2>/dev/null; then
        echo "✓ $name ($host:$port): 正常"
        return 0
    else
        echo "✗ $name ($host:$port): 异常"
        return 1
    fi
}

# ============================================
# 1. 基础服务健康检查
# ============================================

echo "1. 基础服务健康检查..."

# 后端应用
check_service "Backend" "http://localhost:8080/health" "healthy"

# PostgreSQL
check_port "PostgreSQL" "postgres-primary" 5432

# Redis
check_port "Redis" "redis" 6379

echo ""

# ============================================
# 2. 监控系统健康检查
# ============================================

echo "2. 监控系统健康检查..."

# Prometheus
check_service "Prometheus" "http://localhost:9090/-/ready" "Ready"

# Grafana
check_service "Grafana" "http://localhost:3000/api/health" "ok"

# Alertmanager
check_service "Alertmanager" "http://localhost:9093/-/healthy" "healthy"

echo ""

# ============================================
# 3. 日志系统健康检查
# ============================================

echo "3. 日志系统健康检查..."

# Loki
check_service "Loki" "http://localhost:3100/ready" "ready"

# Promtail
check_service "Promtail" "http://localhost:9080/ready" "ready"

echo ""

# ============================================
# 4. 追踪系统健康检查
# ============================================

echo "4. 追踪系统健康检查..."

# Jaeger
check_service "Jaeger" "http://localhost:16686/health" "OK"

# Tempo
check_service "Tempo" "http://localhost:3200/ready" "ready"

# OTEL Collector
check_service "OTEL Collector" "http://localhost:8888/health" "OK"

echo ""

# ============================================
# 5. Kubernetes 资源检查
# ============================================

echo "5. Kubernetes 资源检查..."

if command -v kubectl &> /dev/null; then
    # Pods 状态
    PODS_READY=$(kubectl get pods -n $NAMESPACE -o json | jq '.items[] | select(.status.phase=="Running") | .status.conditions[] | select(.type=="Ready") | .status' | grep -c "True" || echo "0")
    PODS_TOTAL=$(kubectl get pods -n $NAMESPACE --no-headers | wc -l || echo "0")
    
    echo "   Pods 就绪: $PODS_READY/$PODS_TOTAL"
    
    # Deployments 状态
    DEPLOY_READY=$(kubectl get deployments -n $NAMESPACE -o json | jq '.items[] | select(.status.readyReplicas==.spec.replicas) | .metadata.name' | wc -l || echo "0")
    DEPLOY_TOTAL=$(kubectl get deployments -n $NAMESPACE --no-headers | wc -l || echo "0")
    
    echo "   Deployments 就绪: $DEPLOY_READY/$DEPLOY_TOTAL"
    
    # Services 状态
    SERVICES_COUNT=$(kubectl get services -n $NAMESPACE --no-headers | wc -l || echo "0")
    echo "   Services 数量: $SERVICES_COUNT"
    
    # HPA 状态
    HPA_COUNT=$(kubectl get hpa -n $NAMESPACE --no-headers 2>/dev/null | wc -l || echo "0")
    echo "   HPA 数量: $HPA_COUNT"
fi

echo ""

# ============================================
# 6. 数据库健康检查
# ============================================

echo "6. 数据库健康检查..."

# PostgreSQL 连接数
PG_CONN=$(PGPASSWORD=postgres psql -h postgres-primary -U postgres -d industrial_ai -t -c "SELECT count(*) FROM pg_stat_activity" 2>/dev/null || echo "0")
echo "   PostgreSQL 连接数: $PG_CONN"

# PostgreSQL 复制状态
PG_REPLICATION=$(PGPASSWORD=postgres psql -h postgres-primary -U postgres -t -c "SELECT count(*) FROM pg_stat_replication" 2>/dev/null || echo "0")
echo "   PostgreSQL 复制连接: $PG_REPLICATION"

# Redis 内存
REDIS_MEM=$(redis-cli -h redis INFO memory 2>/dev/null | grep used_memory_human | cut -d: -f2 | tr -d '\r' || echo "N/A")
echo "   Redis 内存: $REDIS_MEM"

# Redis Key 数量
REDIS_KEYS=$(redis-cli -h redis DBSIZE 2>/dev/null | awk '{print $2}' || echo "0")
echo "   Redis Key 数量: $REDIS_KEYS"

echo ""

# ============================================
# 7. 系统资源健康检查
# ============================================

echo "7. 系统资源健康检查..."

# CPU 使用率
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d% -f1 || echo "0")
echo "   CPU 使用率: $CPU_USAGE%"

# 内存使用率
MEM_USAGE=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}' || echo "0")
echo "   内存使用率: $MEM_USAGE%"

# 磁盘使用率
DISK_USAGE=$(df -h / | awk '{print $5}' | tail -1 | cut -d% -f1 || echo "0")
echo "   磁盘使用率: $DISK_USAGE%"

echo ""

# ============================================
# 8. 生成健康报告
# ============================================

echo "8. 生成健康报告..."

# 计算健康项数
HEALTHY_ITEMS=0
TOTAL_ITEMS=20

# 检查各项
if nc -z postgres-primary 5430 &>/dev/null; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if nc -z redis 6379 &>/dev/null; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if curl -s http://localhost:8080/health | grep -q "healthy"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if curl -s http://localhost:9090/-/ready | grep -q "Ready"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if curl -s http://localhost:3000/api/health | grep -q "ok"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if curl -s http://localhost:3100/ready | grep -q "ready"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if curl -s http://localhost:16686/health | grep -q "OK"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$CPU_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$MEM_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$DISK_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi

HEALTH_SCORE=$((HEALTHY_ITEMS * 100 / TOTAL_ITEMS))

echo ""
echo "=== 健康报告 ==="
echo ""

echo "健康评分: $HEALTH_SCORE%"
echo "正常项数: $HEALTHY_ITEMS/$TOTAL_ITEMS"

# 健康状态判断
if [ $HEALTH_SCORE -ge 90 ]; then
    echo "状态: ✓ 系统健康"
elif [ $HEALTH_SCORE -ge 70 ]; then
    echo "状态: ⚠ 系统部分异常"
else
    echo "状态: ✗ 系统异常"
fi

echo ""
echo "详细检查命令:"
echo "- 后端: curl http://localhost:8080/health"
echo "- Prometheus: curl http://localhost:9090/-/ready"
echo "- Grafana: curl http://localhost:3000/api/health"
echo "- Loki: curl http://localhost:3100/ready"
echo "- Jaeger: curl http://localhost:16686/health"
echo "- Tempo: curl http://localhost:3200/ready"

echo ""
echo "✅ Industrial AI Platform 综合健康检查完成！"