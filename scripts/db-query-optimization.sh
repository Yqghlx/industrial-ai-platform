#!/bin/bash

# PostgreSQL 查询优化脚本
# 用途: 执行常用查询优化操作

set -e

echo "=== PostgreSQL 查询优化 ==="
echo ""

# 环境变量
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-industrial_ai}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

echo "1. 运行性能索引迁移..."
if [ -f "backend/internal/database/migrations/000004_add_performance_indexes.up.sql" ]; then
    PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
        -f backend/internal/database/migrations/000004_add_performance_indexes.up.sql
    echo "   ✓ 性能索引已创建"
else
    echo "   ⚠️ 索引迁移文件不存在"
fi

echo ""
echo "2. 分析表 (更新统计信息)..."
# ANALYZE 更新表统计信息，帮助查询优化器
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- 分析所有表
ANALYZE VERBOSE users;
ANALYZE VERBOSE devices;
ANALYZE VERBOSE device_telemetry;
ANALYZE VERBOSE alerts;
ANALYZE VERBOSE alert_rules;
ANALYZE VERBOSE work_orders;
ANALYZE VERBOSE roles;
ANALYZE VERBOSE user_roles;
ANALYZE VERBOSE permissions;
ANALYZE VERBOSE role_permissions;
EOF
echo "   ✓ 表统计信息已更新"

echo ""
echo "3. 清理旧遥测数据 (可选)..."
echo "   检查遥测数据数量..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT count(*) as total, 
       count(*) FILTER (WHERE timestamp < now() - interval '30 days') as old_data
FROM device_telemetry;
"

read -p "清理 30 天前的旧数据？(yes/no): " cleanup
if [ "$cleanup" == "yes" ]; then
    PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- 分批删除旧数据 (避免锁表)
DELETE FROM device_telemetry 
WHERE timestamp < now() - interval '30 days'
LIMIT 10000;

-- 继续删除直到完成
DO \$\$ BEGIN
    WHILE (SELECT count(*) FROM device_telemetry WHERE timestamp < now() - interval '30 days') > 0 DO
        DELETE FROM device_telemetry 
        WHERE timestamp < now() - interval '30 days'
        LIMIT 10000;
        COMMIT;
    END LOOP;
END \$\$;
EOF
    echo "   ✓ 旧数据已清理"
fi

echo ""
echo "4. 重建高变更表索引..."
read -p "重建 device_telemetry 索引？(yes/no): " reindex
if [ "$reindex" == "yes" ]; then
    PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- REINDEX CONCURRENTLY (PostgreSQL 12+)
REINDEX TABLE CONCURRENTLY device_telemetry;
EOF
    echo "   ✓ 索引已重建"
fi

echo ""
echo "5. 检查表膨胀..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    schemaname,
    relname as table_name,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    round(100.0 * n_dead_tup / nullif(n_live_tup + n_dead_tup, 0), 2) as dead_ratio,
    last_autovacuum,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY dead_ratio DESC;
EOF

echo ""
echo "6. 手动 VACUUM (如果需要)..."
read -p "执行 VACUUM ANALYZE？(yes/no): " vacuum
if [ "$vacuum" == "yes" ]; then
    PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
VACUUM ANALYZE VERBOSE device_telemetry;
VACUUM ANALYZE VERBOSE alerts;
EOF
    echo "   ✓ VACUUM 完成"
fi

echo ""
echo "=== 查询优化完成 ==="
echo ""
echo "后续建议:"
echo "- 监控慢查询日志"
echo "- 定期执行 ANALYZE"
echo "- 配置 autovacuum 参数"
echo "- 定期清理旧数据"