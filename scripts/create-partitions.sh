#!/bin/bash

# 自动创建遥测数据分区脚本
# 用途: 每月自动创建下个月的分区表

set -e

echo "=== Industrial AI Platform 自动创建分区 ==="
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
PRECREATE_MONTHS="${PRECREATE_MONTHS:-3}"  # 预创建未来 3 个月的分区

echo "配置参数:"
echo "- 数据库主机: $DB_HOST"
echo "- 数据库名称: $DB_NAME"
echo "- 分区表: $PARTITION_TABLE"
echo "- 预创建月数: $PRECREATE_MONTHS"
echo ""

# ============================================
# 计算分区时间范围
# ============================================

echo "1. 计算分区时间范围..."

current_year=$(date +%Y)
current_month=$(date +%m | sed 's/^0//')

echo "   当前时间: $current_year 年 $current_month 月"

# ============================================
# 创建分区
# ============================================

echo ""
echo "2. 创建分区..."

for i in $(seq 1 $PRECREATE_MONTHS); do
    # 计算未来的月份
    future_month=$((current_month + i))
    future_year=$current_year
    
    # 处理跨年
    if [ $future_month -gt 12 ]; then
        future_year=$((current_year + 1))
        future_month=$((future_month - 12))
    fi
    
    # 格式化月份 (两位数)
    future_month_padded=$(printf "%02d" $future_month)
    
    # 分区名称
    partition_name="${PARTITION_TABLE}_${future_year}_${future_month_padded}"
    
    # 开始日期
    start_date="${future_year}-${future_month_padded}-01 00:00:00"
    
    # 计算结束日期 (下月第一天)
    if [ $future_month -eq 12 ]; then
        end_year=$((future_year + 1))
        end_month="01"
    else
        end_year=$future_year
        end_month=$(printf "%02d" $((future_month + 1)))
    fi
    
    end_date="${end_year}-${end_month}-01 00:00:00"
    
    echo "   创建分区: $partition_name ($start_date 到 $end_date)"
    
    # 执行创建分区 SQL
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE TABLE IF NOT EXISTS $partition_name 
        PARTITION OF $PARTITION_TABLE 
        FOR VALUES FROM ('$start_date') TO ('$end_date');
    "
    
    if [ $? -eq 0 ]; then
        echo "   ✓ 分区 $partition_name 创建成功"
    else
        echo "   ✗ 分区 $partition_name 创建失败"
    fi
done

# ============================================
# 验证分区
# ============================================

echo ""
echo "3. 验证分区..."

# 查询所有分区
partitions=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT tablename 
    FROM pg_tables 
    WHERE tablename LIKE '${PARTITION_TABLE}_%' 
    ORDER BY tablename;
")

echo "   现有分区列表:"
for partition in $partitions; do
    if [ -n "$partition" ]; then
        # 获取分区大小
        size=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
            SELECT pg_size_pretty(pg_total_relation_size('$partition'));
        ")
        echo "   - $partition ($size)"
    fi
done

# ============================================
# 创建分区索引
# ============================================

echo ""
echo "4. 创建分区索引..."

# 为新分区创建索引
for i in $(seq 1 $PRECREATE_MONTHS); do
    future_month=$((current_month + i))
    future_year=$current_year
    
    if [ $future_month -gt 12 ]; then
        future_year=$((current_year + 1))
        future_month=$((future_month - 12))
    fi
    
    future_month_padded=$(printf "%02d" $future_month)
    partition_name="${PARTITION_TABLE}_${future_year}_${future_month_padded}"
    
    # 创建时间索引
    echo "   创建索引: idx_${partition_name}_time"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE INDEX IF NOT EXISTS idx_${partition_name}_time 
        ON $partition_name(timestamp DESC);
    "
    
    # 创建设备时间复合索引
    echo "   创建索引: idx_${partition_name}_device_time"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE INDEX IF NOT EXISTS idx_${partition_name}_device_time 
        ON $partition_name(device_id, timestamp DESC);
    "
done

# ============================================
# 创建总结
# ============================================

echo ""
echo "=== 分区创建完成 ==="
echo ""

echo "创建总结:"
echo "- 预创建分区数: $PRECREATE_MONTHS"
echo "- 当前年份: $current_year"
echo "- 当前月份: $current_month"

echo ""
echo "分区管理:"
echo "- 查看分区: SELECT * FROM telemetry_partition_stats;"
echo "- 查看索引: SELECT indexname FROM pg_indexes WHERE tablename LIKE 'telemetry_data_%';"
echo "- 查看大小: SELECT pg_size_pretty(pg_total_relation_size('telemetry_data'));"

echo ""
echo "定时任务配置:"
echo "- 每月 1 日凌晨执行: 0 0 1 * * /scripts/create-partitions.sh"
echo "- 或使用 pg_cron: SELECT cron.schedule('create_partitions', '0 0 1 * *', 'SELECT create_telemetry_partition');"

echo ""
echo "✅ Industrial AI Platform 自动创建分区完成！"