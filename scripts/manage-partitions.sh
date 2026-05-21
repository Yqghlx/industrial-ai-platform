#!/bin/bash

# 分区管理脚本
# 用途: 管理遥测数据分区 (归档/删除/统计)

set -e

echo "=== Industrial AI Platform 分区管理 ==="
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

# 分区配置
PARTITION_TABLE="telemetry_data"
ARCHIVE_TABLE="telemetry_archive"

# 保留策略
HOT_RETENTION_MONTHS="${HOT_RETENTION_MONTHS:-3}"   # 热数据保留 3 个月
WARM_RETENTION_MONTHS="${WARM_RETENTION_MONTHS:-6}"  # 温数据保留 6 个月
COLD_RETENTION_MONTHS="${COLD_RETENTION_MONTHS:-12}" # 冷数据保留 12 个月

# 操作类型
OPERATION="${OPERATION:-status}"  # status/archive/delete

echo "配置参数:"
echo "- 数据库主机: $DB_HOST"
echo "- 数据库名称: $DB_NAME"
echo "- 分区表: $PARTITION_TABLE"
echo "- 热数据保留: $HOT_RETENTION_MONTHS 个月"
echo "- 温数据保留: $WARM_RETENTION_MONTHS 个月"
echo "- 冷数据保留: $COLD_RETENTION_MONTHS 个月"
echo "- 操作类型: $OPERATION"
echo ""

# ============================================
# 分区状态检查
# ============================================

check_partition_status() {
    echo "1. 检查分区状态..."

    # 查询所有分区
    partitions=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT 
            tablename AS partition_name,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
            pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
        FROM pg_tables 
        WHERE tablename LIKE '${PARTITION_TABLE}_%'
        ORDER BY tablename;
    ")

    echo "$partitions"

    # 查询总大小
    total_size=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT pg_size_pretty(pg_total_relation_size('$PARTITION_TABLE'));
    ")

    echo ""
    echo "   总大小: $total_size"

    # 查询总行数
    total_rows=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT count(*) FROM $PARTITION_TABLE;
    ")

    echo "   总行数: $total_rows"
}

# ============================================
# 分区归档
# ============================================

archive_old_partitions() {
    echo ""
    echo "2. 归档旧分区..."

    # 计算归档截止日期
    archive_cutoff=$(date -d "$WARM_RETENTION_MONTHS months ago" +%Y-%m-01)
    echo "   归档截止: $archive_cutoff"

    # 查询需要归档的分区
    archive_partitions=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE '${PARTITION_TABLE}_%' 
        AND tablename < '${PARTITION_TABLE}_${archive_cutoff}'
        ORDER BY tablename;
    ")

    for partition in $archive_partitions; do
        if [ -n "$partition" ]; then
            echo "   归档分区: $partition"
            
            # 获取分区行数
            rows=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
                SELECT count(*) FROM $partition;
            ")
            
            echo "   数据行数: $rows"
            
            # 归档数据到归档表
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
                INSERT INTO $ARCHIVE_TABLE
                SELECT *, NOW() AS archived_at FROM $partition;
            "
            
            if [ $? -eq 0 ]; then
                echo "   ✓ 分区 $partition 归档成功"
            else
                echo "   ✗ 分区 $partition 归档失败"
            fi
        fi
    done
}

# ============================================
# 分区删除
# ============================================

delete_old_partitions() {
    echo ""
    echo "3. 删除旧分区..."

    # 计算删除截止日期
    delete_cutoff=$(date -d "$COLD_RETENTION_MONTHS months ago" +%Y-%m-01)
    echo "   删除截止: $delete_cutoff"

    # 查询需要删除的分区
    delete_partitions=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE '${PARTITION_TABLE}_%' 
        AND tablename < '${PARTITION_TABLE}_${delete_cutoff}'
        ORDER BY tablename;
    ")

    for partition in $delete_partitions; do
        if [ -n "$partition" ]; then
            echo "   删除分区: $partition"
            
            # 获取分区大小
            size=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
                SELECT pg_size_pretty(pg_total_relation_size('$partition'));
            ")
            
            echo "   分区大小: $size"
            
            # 删除分区
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
                DROP TABLE IF EXISTS $partition;
            "
            
            if [ $? -eq 0 ]; then
                echo "   ✓ 分区 $partition 删除成功"
            else
                echo "   ✗ 分区 $partition 删除失败"
            fi
        fi
    done
}

# ============================================
# 分区压缩
# ============================================

compress_partition() {
    partition=$1
    
    echo "   压缩分区: $partition"
    
    # 执行 VACUUM FULL 压缩
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        VACUUM FULL ANALYZE $partition;
    "
    
    if [ $? -eq 0 ]; then
        echo "   ✓ 分区 $partition 压缩成功"
    else
        echo "   ✗ 分区 $partition 压缩失败"
    fi
}

# ============================================
# 分区统计
# ============================================

generate_partition_report() {
    echo ""
    echo "4. 生成分区统计报告..."

    report_file="/logs/partition-report-$(date +%Y-%m-%d).md"

    cat > $report_file << EOF
# Industrial AI Platform 分区统计报告

**报告日期**: $(date '+%Y-%m-%d %H:%M:%S')
**分区表**: $PARTITION_TABLE

---

## 分区状态

EOF

    # 查询分区统计
    stats=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT 
            tablename AS partition_name,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
            pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
        FROM pg_tables 
        WHERE tablename LIKE '${PARTITION_TABLE}_%'
        ORDER BY tablename;
    ")

    echo "$stats" >> $report_file

    echo ""
    echo "   报告已生成: $report_file"
}

# ============================================
# 执行操作
# ============================================

case $OPERATION in
    status)
        check_partition_status
        ;;
    archive)
        check_partition_status
        archive_old_partitions
        ;;
    delete)
        check_partition_status
        delete_old_partitions
        ;;
    full)
        check_partition_status
        archive_old_partitions
        delete_old_partitions
        generate_partition_report
        ;;
    *)
        echo "未知操作: $OPERATION"
        echo "支持的操作: status/archive/delete/full"
        exit 1
        ;;
esac

# ============================================
# 完成总结
# ============================================

echo ""
echo "=== 分区管理完成 ==="
echo ""

echo "管理总结:"
echo "- 操作类型: $OPERATION"
echo "- 热数据保留: $HOT_RETENTION_MONTHS 个月"
echo "- 温数据保留: $WARM_RETENTION_MONTHS 个月"
echo "- 冷数据保留: $COLD_RETENTION_MONTHS 个月"

echo ""
echo "定时任务配置:"
echo "- 每月归档: 0 3 1 * * /scripts/manage-partitions.sh archive"
echo "- 每月删除: 0 4 1 * * /scripts/manage-partitions.sh delete"
echo "- 每月报告: 0 5 1 * * /scripts/manage-partitions.sh status"

echo ""
echo "✅ Industrial AI Platform 分区管理完成！"