#!/bin/bash

# PostgreSQL 性能分析脚本
# 用途: 分析数据库性能瓶颈

set -e

echo "=== PostgreSQL 性能分析 ==="
echo ""

# 环境变量
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-industrial_ai}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

echo "1. 检查 pg_stat_statements 扩展..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT extname, extversion FROM pg_extension WHERE extname = 'pg_stat_statements';
" || echo "   ⚠️ pg_stat_statements 未启用"

echo ""
echo "2. 启用 pg_stat_statements 扩展..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
" && echo "   ✓ pg_stat_statements 已启用"

echo ""
echo "3. 最慢查询分析 (Top 20)..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    substring(query, 1, 100) as query_preview,
    calls,
    round(total_time / calls::numeric, 2) as avg_time_ms,
    round(total_time / 1000::numeric, 2) as total_time_s,
    rows,
    round(100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0), 2) AS hit_percent
FROM pg_stat_statements
ORDER BY total_time / calls DESC
LIMIT 20;
EOF

echo ""
echo "4. 最频繁查询分析..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    substring(query, 1, 100) as query_preview,
    calls,
    round(total_time / 1000::numeric, 2) as total_time_s,
    rows
FROM pg_stat_statements
ORDER BY calls DESC
LIMIT 10;
EOF

echo ""
echo "5. 缓存命中率低的查询..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    substring(query, 1, 100) as query_preview,
    calls,
    shared_blks_read,
    shared_blks_hit,
    round(100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0), 2) AS hit_percent
FROM pg_stat_statements
WHERE shared_blks_read > 100
ORDER BY hit_percent ASC
LIMIT 10;
EOF

echo ""
echo "6. 表大小统计..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    table_name,
    pg_size_pretty(pg_total_relation_size(table_name::regclass)) as total_size,
    pg_size_pretty(pg_relation_size(table_name::regclass)) as table_size,
    pg_size_pretty(pg_indexes_size(table_name::regclass)) as index_size
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_type = 'BASE TABLE'
ORDER BY pg_total_relation_size(table_name::regclass) DESC
LIMIT 10;
EOF

echo ""
echo "7. 索引使用统计..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    indexrelname as index_name,
    relname as table_name,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC
LIMIT 20;
EOF

echo ""
echo "8. 未使用索引 (建议删除)..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    indexrelname as index_name,
    relname as table_name,
    idx_scan as scans,
    pg_size_pretty(pg_relation_size(indexrelid)) as wasted_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
AND indexrelname NOT LIKE '%_pkey'
ORDER BY pg_relation_size(indexrelid) DESC
LIMIT 10;
EOF

echo ""
echo "9. 连接池状态..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    state,
    count(*) as count,
    max(now() - query_start) as max_duration
FROM pg_stat_activity
GROUP BY state
ORDER BY count DESC;
EOF

echo ""
echo "10. 长时间运行查询..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
SELECT 
    pid,
    now() - query_start AS duration,
    substring(query, 1, 100) as query_preview,
    state,
    usename
FROM pg_stat_activity
WHERE (now() - query_start) > interval '1 minute'
AND state != 'idle'
ORDER BY duration DESC
LIMIT 10;
EOF

echo ""
echo "=== 性能分析完成 ==="
echo ""
echo "建议操作:"
echo "1. 针对慢查询添加索引"
echo "2. 删除未使用的索引"
echo "3. 优化缓存命中率低的查询"
echo "4. 监控长时间运行的查询"