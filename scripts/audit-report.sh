#!/bin/bash

# 审计日志合规报告生成脚本
# 用途: 生成安全审计合规报告

set -e

echo "=== Industrial AI Platform 审计日志合规报告 ==="
echo ""

# ============================================
# 配置参数
# ============================================

# 数据库配置
DB_HOST="${DB_HOST:-postgres-primary}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-industrial_ai}"

# 报告配置
REPORT_DIR="${REPORT_DIR:-/logs/audit-reports}"
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_TIME=$(date +%H%M%S)

# 报告类型
REPORT_TYPE="${REPORT_TYPE:-daily}"  # daily/weekly/monthly/yearly

# 创建报告目录
mkdir -p $REPORT_DIR

echo "报告参数:"
echo "- 报告目录: $REPORT_DIR"
echo "- 报告类型: $REPORT_TYPE"
echo "- 报告日期: $REPORT_DATE"
echo ""

# ============================================
# 生成报告
# ============================================

REPORT_FILE="$REPORT_DIR/audit-report-$REPORT_TYPE-$REPORT_DATE-$REPORT_TIME.md"

cat > $REPORT_FILE << EOF
# Industrial AI Platform 安全审计合规报告

**报告类型**: $REPORT_TYPE
**报告日期**: $REPORT_DATE
**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')

---

## 📋 审计概览

EOF

# ============================================
# 1. 审计日志统计
# ============================================

echo "1. 生成审计日志统计..."

# 计算时间范围
case $REPORT_TYPE in
    daily)
        START_TIME=$(date -d "1 day ago" '+%Y-%m-%d %H:%M:%S')
        END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
        ;;
    weekly)
        START_TIME=$(date -d "7 days ago" '+%Y-%m-%d %H:%M:%S')
        END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
        ;;
    monthly)
        START_TIME=$(date -d "30 days ago" '+%Y-%m-%d %H:%M:%S')
        END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
        ;;
    yearly)
        START_TIME=$(date -d "365 days ago" '+%Y-%m-%d %H:%M:%S')
        END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
        ;;
    *)
        START_TIME=$(date -d "1 day ago" '+%Y-%m-%d %H:%M:%S')
        END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
        ;;
esac

echo "   时间范围: $START_TIME 到 $END_TIME"

# 查询总数
TOTAL_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COUNT(*) FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
")

# 查询失败数
FAILURE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COUNT(*) FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND result = 'failure'
")

# 计算失败率
FAILURE_RATE=$(echo "scale=2; $FAILURE_COUNT * 100 / $TOTAL_COUNT" | bc)

cat >> $REPORT_FILE << EOF

**审计总数**: $TOTAL_COUNT
**失败数**: $FAILURE_COUNT
**失败率**: $FAILURE_RATE%
**时间范围**: $START_TIME 到 $END_TIME

EOF

# ============================================
# 2. 事件类型统计
# ============================================

echo ""
echo "2. 生成事件类型统计..."

cat >> $REPORT_FILE << EOF

### 事件类型统计

EOF

EVENT_TYPE_STATS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT event_type, COUNT(*) as count
    FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    GROUP BY event_type
    ORDER BY count DESC
" || echo "N/A")

echo "$EVENT_TYPE_STATS" >> $REPORT_FILE

# ============================================
# 3. 用户活动统计
# ============================================

echo ""
echo "3. 生成用户活动统计..."

cat >> $REPORT_FILE << EOF

### Top 10 用户活动

EOF

TOP_USERS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT user_id, COUNT(*) as operations
    FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    GROUP BY user_id
    ORDER BY operations DESC
    LIMIT 10
" || echo "N/A")

echo "$TOP_USERS" >> $REPORT_FILE

# ============================================
# 4. 安全事件统计
# ============================================

echo ""
echo "4. 生成安全事件统计..."

cat >> $REPORT_FILE << EOF

### 安全事件统计

EOF

SECURITY_EVENTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT event_type, severity, COUNT(*) as count
    FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND event_category = 'security'
    GROUP BY event_type, severity
    ORDER BY count DESC
" || echo "N/A")

echo "$SECURITY_EVENTS" >> $REPORT_FILE

# ============================================
# 5. 异常 IP 地址统计
# ============================================

echo ""
echo "5. 生成异常 IP 地址统计..."

cat >> $REPORT_FILE << EOF

### 异常 IP 地址统计 (失败次数 > 10)

EOF

ABNORMAL_IPS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT ip_address, COUNT(*) as failure_count
    FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND result = 'failure'
    GROUP BY ip_address
    HAVING COUNT(*) > 10
    ORDER BY failure_count DESC
" || echo "N/A")

echo "$ABNORMAL_IPS" >> $REPORT_FILE

# ============================================
# 6. 合规检查
# ============================================

echo ""
echo "6. 生成合规检查..."

cat >> $REPORT_FILE << EOF

---

## 🔒 合规检查

EOF

# 检查认证失败
AUTH_FAILURE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COUNT(*) FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND event_type = 'auth.failed'
")

cat >> $REPORT_FILE << EOF

**认证失败数**: $AUTH_FAILURE_COUNT
EOF

if [ "$AUTH_FAILURE_COUNT" -gt 100 ]; then
    echo "**⚠️ 警告**: 认证失败次数过多，可能存在安全风险" >> $REPORT_FILE
fi

# 检查权限违规
VIOLATION_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COUNT(*) FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND event_type = 'security.violation'
")

cat >> $REPORT_FILE << EOF

**权限违规数**: $VIOLATION_COUNT
EOF

if [ "$VIOLATION_COUNT" -gt 10 ]; then
    echo "**⚠️ 警告**: 权限违规次数过多，需要检查权限配置" >> $REPORT_FILE
fi

# 检查租户隔离违规
TENANT_VIOLATION_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COUNT(*) FROM audit_logs
    WHERE timestamp >= '$START_TIME' AND timestamp <= '$END_TIME'
    AND operation LIKE '%Tenant%'
    AND result = 'failure'
")

cat >> $REPORT_FILE << EOF

**租户隔离违规数**: $TENANT_VIOLATION_COUNT
EOF

if [ "$TENANT_VIOLATION_COUNT" -gt 0 ]; then
    echo "**⚠️ 警告**: 存在租户隔离违规，需要立即处理" >> $REPORT_FILE
fi

# ============================================
# 7. 建议与改进
# ============================================

echo ""
echo "7. 生成建议与改进..."

cat >> $REPORT_FILE << EOF

---

## 💡 建议与改进

EOF

# 生成建议
if [ "$AUTH_FAILURE_COUNT" -gt 100 ]; then
    echo "1. 加强认证安全措施，考虑启用双因素认证" >> $REPORT_FILE
fi

if [ "$VIOLATION_COUNT" -gt 10 ]; then
    echo "2. 检查并优化权限配置，减少权限违规" >> $REPORT_FILE
fi

if [ "$TENANT_VIOLATION_COUNT" -gt 0 ]; then
    echo "3. 加强租户隔离检查，防止跨租户数据访问" >> $REPORT_FILE
fi

if [ "$FAILURE_RATE" -gt 5 ]; then
    echo "4. 分析失败原因，优化系统稳定性" >> $REPORT_FILE
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
echo "=== 报告生成完成 ==="
echo ""

echo "报告总结:"
echo "- 审计总数: $TOTAL_COUNT"
echo "- 失败数: $FAILURE_COUNT"
echo "- 失败率: $FAILURE_RATE%"
echo "- 认证失败: $AUTH_FAILURE_COUNT"
echo "- 权限违规: $VIOLATION_COUNT"
echo "- 租户隔离违规: $TENANT_VIOLATION_COUNT"

echo ""
echo "查看报告:"
echo "- cat $REPORT_FILE"
echo "- 或使用浏览器打开"

echo ""
echo "✅ Industrial AI Platform 审计日志合规报告生成完成！"