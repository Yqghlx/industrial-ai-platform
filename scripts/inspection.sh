#!/bin/bash

# Industrial AI Platform 系统巡检脚本
# 用途: 全面检查系统健康状态并生成报告

set -e

echo "=== Industrial AI Platform 系统巡检 ==="
echo ""

# ============================================
# 配置参数
# ============================================

REPORT_DIR="${REPORT_DIR:-/logs/inspection}"
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_TIME=$(date +%H%M%S)
REPORT_FILE="$REPORT_DIR/inspection-$REPORT_DATE-$REPORT_TIME.md"

# 服务配置
NAMESPACE="${NAMESPACE:-industrial-ai}"

# 创建报告目录
mkdir -p $REPORT_DIR

echo "巡检参数:"
echo "- 报告目录: $REPORT_DIR"
echo "- 报告文件: $REPORT_FILE"
echo "- Namespace: $NAMESPACE"
echo ""

# ============================================
# 开始生成报告
# ============================================

cat > $REPORT_FILE << EOF
# Industrial AI Platform 系统巡检报告

**巡检时间**: $(date '+%Y-%m-%d %H:%M:%S')
**巡检人员**: Automated Script
**Namespace**: $NAMESPACE

---

## 📋 巡检概览

EOF

# ============================================
# 1. 系统资源检查
# ============================================

echo "1. 检查系统资源..."

cat >> $REPORT_FILE << EOF

### 1. 系统资源检查

EOF

# CPU 使用率
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d% -f1 || echo "N/A")
echo "   CPU 使用率: $CPU_USAGE%"
cat >> $REPORT_FILE << EOF

**CPU 使用率**: $CPU_USAGE%

EOF

# 内存使用率
MEM_USAGE=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
MEM_TOTAL=$(free -h | grep Mem | awk '{print $2}')
MEM_USED=$(free -h | grep Mem | awk '{print $3}')
echo "   内存使用率: $MEM_USAGE% ($MEM_USED / $MEM_TOTAL)"
cat >> $REPORT_FILE << EOF

**内存使用率**: $MEM_USAGE% ($MEM_USED / $MEM_TOTAL)

EOF

# 磁盘使用率
DISK_USAGE=$(df -h / | awk '{print $5}' | tail -1 | cut -d% -f1)
DISK_TOTAL=$(df -h / | awk '{print $2}' | tail -1)
DISK_USED=$(df -h / | awk '{print $3}' | tail -1)
echo "   磁盘使用率: $DISK_USAGE% ($DISK_USED / $DISK_TOTAL)"
cat >> $REPORT_FILE << EOF

**磁盘使用率**: $DISK_USAGE% ($DISK_USED / $DISK_TOTAL)

EOF

# 网络连接数
NET_CONNECTIONS=$(ss -s | awk '/estab/ {print $4}' || echo "N/A")
echo "   网络连接数: $NET_CONNECTIONS"
cat >> $REPORT_FILE << EOF

**网络连接数**: $NET_CONNECTIONS

EOF

# ============================================
# 2. 服务状态检查
# ============================================

echo ""
echo "2. 检查服务状态..."

cat >> $REPORT_FILE << EOF

### 2. 服务状态检查

EOF

# Docker 服务状态
if command -v docker &> /dev/null; then
    DOCKER_STATUS=$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "N/A")
    echo "$DOCKER_STATUS" >> $REPORT_FILE
fi

# Kubernetes Pods 状态
if command -v kubectl &> /dev/null; then
    PODS_STATUS=$(kubectl get pods -n $NAMESPACE -o wide 2>/dev/null || echo "N/A")
    echo "" >> $REPORT_FILE
    echo "**Kubernetes Pods**:" >> $REPORT_FILE
    echo "$PODS_STATUS" >> $REPORT_FILE
fi

# ============================================
# 3. 数据库检查
# ============================================

echo ""
echo "3. 检查数据库状态..."

cat >> $REPORT_FILE << EOF

### 3. 数据库检查

EOF

# PostgreSQL 连接检查
PG_READY=$(pg_isready -h postgres-primary -p 5432 || echo "down")
if echo "$PG_READY" | grep -q "accepting"; then
    echo "   PostgreSQL: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**PostgreSQL**: ✓ 正常

EOF
    
    # 数据库大小
    DB_SIZE=$(PGPASSWORD=postgres psql -h postgres-primary -U postgres -d industrial_ai -t -c "SELECT pg_size_pretty(pg_database_size('industrial_ai'))" || echo "N/A")
    echo "   数据库大小: $DB_SIZE"
    echo "**数据库大小**: $DB_SIZE" >> $REPORT_FILE
    
    # 数据库连接数
    DB_CONN=$(PGPASSWORD=postgres psql -h postgres-primary -U postgres -d industrial_ai -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'industrial_ai'" || echo "0")
    echo "   数据库连接数: $DB_CONN"
    echo "**数据库连接数**: $DB_CONN" >> $REPORT_FILE
else
    echo "   PostgreSQL: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**PostgreSQL**: ✗ 异常

EOF
fi

# Redis 连接检查
REDIS_PING=$(redis-cli -h redis ping || echo "down")
if [ "$REDIS_PING" == "PONG" ]; then
    echo "   Redis: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**Redis**: ✓ 正常

EOF
    
    # Redis 内存使用
    REDIS_MEM=$(redis-cli -h redis INFO memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
    echo "   Redis 内存: $REDIS_MEM"
    echo "**Redis 内存**: $REDIS_MEM" >> $REPORT_FILE
    
    # Redis Key 数量
    REDIS_KEYS=$(redis-cli -h redis DBSIZE | awk '{print $2}')
    echo "   Redis Key 数量: $REDIS_KEYS"
    echo "**Redis Key 数量**: $REDIS_KEYS" >> $REPORT_FILE
else
    echo "   Redis: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**Redis**: ✗ 异常

EOF
fi

# ============================================
# 4. 应用健康检查
# ============================================

echo ""
echo "4. 检查应用健康..."

cat >> $REPORT_FILE << EOF

### 4. 应用健康检查

EOF

# 健康检查
HEALTH_STATUS=$(curl -s http://localhost:8080/health || echo "down")
if echo "$HEALTH_STATUS" | grep -q "healthy"; then
    echo "   应用健康: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**应用健康**: ✓ 正常

EOF
else
    echo "   应用健康: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**应用健康**: ✗ 异常

EOF
fi

# HTTP 服务检查
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/devices || echo "000")
echo "   HTTP 状态码: $HTTP_STATUS"
cat >> $REPORT_FILE << EOF

**HTTP 状态码**: $HTTP_STATUS

EOF

# ============================================
# 5. 监控系统检查
# ============================================

echo ""
echo "5. 检查监控系统..."

cat >> $REPORT_FILE << EOF

### 5. 监控系统检查

EOF

# Prometheus 检查
PROM_READY=$(curl -s http://localhost:9090/-/ready || echo "down")
if [ "$PROM_READY" == "Prometheus is Ready." ]; then
    echo "   Prometheus: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**Prometheus**: ✓ 正常

EOF
else
    echo "   Prometheus: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**Prometheus**: ✗ 异常

EOF
fi

# Grafana 检查
GRAFANA_READY=$(curl -s http://localhost:3000/api/health || echo "down")
if echo "$GRAFANA_READY" | grep -q "ok"; then
    echo "   Grafana: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**Grafana**: ✓ 正常

EOF
else
    echo "   Grafana: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**Grafana**: ✗ 异常

EOF
fi

# Loki 检查
LOKI_READY=$(curl -s http://localhost:3100/ready || echo "down")
if [ "$LOKI_READY" == "ready" ]; then
    echo "   Loki: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**Loki**: ✓ 正常

EOF
else
    echo "   Loki: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**Loki**: ✗ 异常

EOF
fi

# Jaeger 检查
JAEGER_HEALTH=$(curl -s http://localhost:16686/health || echo "down")
if [ "$JAEGER_HEALTH" == "OK" ]; then
    echo "   Jaeger: ✓ 正常"
    cat >> $REPORT_FILE << EOF

**Jaeger**: ✓ 正常

EOF
else
    echo "   Jaeger: ✗ 异常"
    cat >> $REPORT_FILE << EOF

**Jaeger**: ✗ 异常

EOF
fi

# ============================================
# 6. 安全检查
# ============================================

echo ""
echo "6. 检查安全状态..."

cat >> $REPORT_FILE << EOF

### 6. 安全检查

EOF

# 防火墙状态
FIREWALL_STATUS=$(systemctl is-active firewalld || ufw status | head -1 || echo "N/A")
echo "   防火墙状态: $FIREWALL_STATUS"
cat >> $REPORT_FILE << EOF

**防火墙状态**: $FIREWALL_STATUS

EOF

# SSL 证书检查 (假设证书在 /etc/ssl)
if [ -f "/etc/ssl/certs/server.crt" ]; then
    CERT_EXPIRY=$(openssl x509 -enddate -noout -in /etc/ssl/certs/server.crt | cut -d= -f2)
    echo "   SSL 证书有效期: $CERT_EXPIRY"
    cat >> $REPORT_FILE << EOF

**SSL 证书有效期**: $CERT_EXPIRY

EOF
fi

# ============================================
# 7. 日志检查
# ============================================

echo ""
echo "7. 检查日志状态..."

cat >> $REPORT_FILE << EOF

### 7. 日志检查

EOF

# 检查错误日志数量 (最近 1 小时)
ERROR_COUNT=$(grep -c "ERROR" /logs/backend.log 2>/dev/null || echo "0")
echo "   错误日志数 (最近): $ERROR_COUNT"
cat >> $REPORT_FILE << EOF

**错误日志数**: $ERROR_COUNT

EOF

# ============================================
# 8. 生成健康评分
# ============================================

echo ""
echo "8. 生成健康评分..."

cat >> $REPORT_FILE << EOF

---

## 📊 健康评分

EOF

# 计算健康评分
HEALTHY_ITEMS=0
TOTAL_ITEMS=10

if [ "$CPU_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$MEM_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$DISK_USAGE" -lt 80 ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if echo "$PG_READY" | grep -q "accepting"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$REDIS_PING" == "PONG" ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if echo "$HEALTH_STATUS" | grep -q "healthy"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$PROM_READY" == "Prometheus is Ready." ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if echo "$GRAFANA_READY" | grep -q "ok"; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$LOKI_READY" == "ready" ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi
if [ "$JAEGER_HEALTH" == "OK" ]; then HEALTHY_ITEMS=$((HEALTHY_ITEMS+1)); fi

HEALTH_SCORE=$((HEALTHY_ITEMS * 100 / TOTAL_ITEMS))

echo "   健康评分: $HEALTH_SCORE% ($HEALTHY_ITEMS/$TOTAL_ITEMS)"
cat >> $REPORT_FILE << EOF

**健康评分**: $HEALTH_SCORE% ($HEALTHY_ITEMS/$TOTAL_ITEMS 项正常)

EOF

# ============================================
# 9. 生成建议
# ============================================

echo ""
echo "9. 生成建议..."

cat >> $REPORT_FILE << EOF

---

## 💡 建议与行动项

EOF

# 根据检查结果生成建议
if [ "$CPU_USAGE" -gt 80 ]; then
    echo "   [WARNING] CPU 使用率过高"
    echo "- **CPU**: 使用率过高 ($CPU_USAGE%)，建议优化应用或扩容" >> $REPORT_FILE
fi

if [ "$MEM_USAGE" -gt 80 ]; then
    echo "   [WARNING] 内存使用率过高"
    echo "- **内存**: 使用率过高 ($MEM_USAGE%)，建议清理内存或扩容" >> $REPORT_FILE
fi

if [ "$DISK_USAGE" -gt 80 ]; then
    echo "   [WARNING] 磁盘使用率过高"
    echo "- **磁盘**: 使用率过高 ($DISK_USAGE%)，建议清理磁盘或扩容" >> $REPORT_FILE
fi

if [ "$ERROR_COUNT" -gt 100 ]; then
    echo "   [WARNING] 错误日志过多"
    echo "- **日志**: 错误日志过多 ($ERROR_COUNT)，建议检查应用问题" >> $REPORT_FILE
fi

# ============================================
# 报告完成
# ============================================

cat >> $REPORT_FILE << EOF

---

**报告生成时间**: $(date '+%Y-%m-%d %H:%M:%S')
**报告路径**: $REPORT_FILE

EOF

echo ""
echo "=== 系统巡检完成 ==="
echo ""

echo "巡检总结:"
echo "- 健康评分: $HEALTH_SCORE%"
echo "- 报告文件: $REPORT_FILE"
echo "- 正常项数: $HEALTHY_ITEMS/$TOTAL_ITEMS"

echo ""
echo "查看报告:"
echo "- cat $REPORT_FILE"
echo "- 或使用浏览器打开"

echo ""
echo "✅ Industrial AI Platform 系统巡检完成！"