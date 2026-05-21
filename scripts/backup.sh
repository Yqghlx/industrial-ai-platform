#!/bin/bash

# Industrial AI Platform 数据库备份脚本
# 用途: 自动备份 PostgreSQL 和 Redis 数据

set -e

echo "=== Industrial AI Platform 数据库备份 ==="
echo ""

# ============================================
# 配置参数
# ============================================

# 备份配置
BACKUP_BASE_DIR="${BACKUP_DIR:-/backup/industrial-ai}"
BACKUP_DATE=$(date +%Y-%m-%d)
BACKUP_TIME=$(date +%H%M%S)
BACKUP_DIR="$BACKUP_BASE_DIR/$BACKUP_DATE"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

# 数据库配置
DB_HOST="${DB_HOST:-postgres-primary}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-industrial_ai}"

# Redis 配置
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# 远程存储配置 (可选)
S3_BUCKET="${S3_BUCKET:-s3://backup-bucket/industrial-ai}"
S3_ENABLED="${S3_ENABLED:-false}"

echo "备份参数:"
echo "- 备份目录: $BACKUP_DIR"
echo "- 数据库主机: $DB_HOST"
echo "- 数据库名称: $DB_NAME"
echo "- Redis 主机: $REDIS_HOST"
echo "- 保留天数: $RETENTION_DAYS"
echo ""

# ============================================
# 1. 创建备份目录
# ============================================

echo "1. 创建备份目录..."

mkdir -p $BACKUP_DIR

echo "   ✓ 备份目录已创建: $BACKUP_DIR"

# ============================================
# 2. 备份 PostgreSQL
# ============================================

echo ""
echo "2. 备份 PostgreSQL 数据库..."

# 执行数据库备份
BACKUP_SQL="$BACKUP_DIR/database-$BACKUP_TIME.sql"

PGPASSWORD=$DB_PASSWORD pg_dump \
    -h $DB_HOST \
    -p $DB_PORT \
    -U $DB_USER \
    -d $DB_NAME \
    --format=plain \
    --verbose \
    > $BACKUP_SQL

if [ $? -eq 0 ]; then
    # 压缩 SQL 文件
    gzip $BACKUP_SQL
    
    BACKUP_SIZE=$(du -h "$BACKUP_SQL.gz" | cut -f1)
    echo "   ✓ PostgreSQL 备份完成: $BACKUP_SQL.gz ($BACKUP_SIZE)"
else
    echo "   ✗ PostgreSQL 备份失败"
    exit 1
fi

# ============================================
# 3. 备份 Redis
# ============================================

echo ""
echo "3. 备份 Redis 数据..."

# 触发 Redis BGSAVE
redis-cli -h $REDIS_HOST -p $REDIS_PORT BGSAVE

# 等待 BGSAVE 完成
sleep 5

# 检查 BGSAVE 状态
BGSAVE_STATUS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT LASTSAVE)
echo "   Redis LASTSAVE: $BGSAVE_STATUS"

# 备份 RDB 文件 (假设 Redis 数据目录在 /var/lib/redis)
BACKUP_RDB="$BACKUP_DIR/redis-$BACKUP_TIME.rdb"

if [ -f "/var/lib/redis/dump.rdb" ]; then
    cp /var/lib/redis/dump.rdb $BACKUP_RDB
    
    # 压缩 RDB 文件
    gzip $BACKUP_RDB
    
    BACKUP_SIZE=$(du -h "$BACKUP_RDB.gz" | cut -f1)
    echo "   ✓ Redis 备份完成: $BACKUP_RDB.gz ($BACKUP_SIZE)"
else
    echo "   ⚠️ Redis RDB 文件不存在，跳过备份"
fi

# ============================================
# 4. 备份配置文件
# ============================================

echo ""
echo "4. 备份配置文件..."

BACKUP_CONFIG="$BACKUP_DIR/config-$BACKUP_TIME.tar.gz"

tar -czf $BACKUP_CONFIG \
    infra/postgresql/*.conf \
    infra/redis/*.conf \
    infra/k8s/*.yaml \
    .env \
    config.yaml 2>/dev/null || true

if [ -f $BACKUP_CONFIG ]; then
    BACKUP_SIZE=$(du -h $BACKUP_CONFIG | cut -f1)
    echo "   ✓ 配置文件备份完成: $BACKUP_CONFIG ($BACKUP_SIZE)"
else
    echo "   ⚠️ 配置文件备份失败，跳过"
fi

# ============================================
# 5. 创建备份清单
# ============================================

echo ""
echo "5. 创建备份清单..."

BACKUP_MANIFEST="$BACKUP_DIR/manifest-$BACKUP_TIME.json"

# 计算总备份大小
TOTAL_SIZE=$(du -sh $BACKUP_DIR | cut -f1)

# 创建 JSON 清单
cat > $BACKUP_MANIFEST << EOF
{
    "backup_date": "$BACKUP_DATE",
    "backup_time": "$BACKUP_TIME",
    "backup_type": "full",
    "database": {
        "host": "$DB_HOST",
        "port": $DB_PORT,
        "name": "$DB_NAME",
        "file": "database-$BACKUP_TIME.sql.gz"
    },
    "redis": {
        "host": "$REDIS_HOST",
        "port": $REDIS_PORT,
        "file": "redis-$BACKUP_TIME.rdb.gz"
    },
    "config_file": "config-$BACKUP_TIME.tar.gz",
    "total_size": "$TOTAL_SIZE",
    "retention_days": $RETENTION_DAYS
}
EOF

echo "   ✓ 备份清单已创建: $BACKUP_MANIFEST"

# ============================================
# 6. 压缩整个备份目录
# ============================================

echo ""
echo "6. 压缩备份目录..."

BACKUP_ARCHIVE="$BACKUP_BASE_DIR/$BACKUP_DATE-$BACKUP_TIME.tar.gz"

tar -czf $BACKUP_ARCHIVE $BACKUP_DIR

if [ $? -eq 0 ]; then
    ARCHIVE_SIZE=$(du -h $BACKUP_ARCHIVE | cut -f1)
    echo "   ✓ 备份压缩完成: $BACKUP_ARCHIVE ($ARCHIVE_SIZE)"
else
    echo "   ✗ 备份压缩失败"
    exit 1
fi

# ============================================
# 7. 上传到远程存储 (可选)
# ============================================

if [ "$S3_ENABLED" == "true" ]; then
    echo ""
    echo "7. 上传到远程存储..."
    
    aws s3 cp $BACKUP_ARCHIVE $S3_BUCKET/$BACKUP_DATE/
    
    if [ $? -eq 0 ]; then
        echo "   ✓ 备份已上传到 S3: $S3_BUCKET/$BACKUP_DATE/"
    else
        echo "   ⚠️ S3 上传失败"
    fi
fi

# ============================================
# 8. 清理旧备份
# ============================================

echo ""
echo "8. 清理旧备份 (保留 $RETENTION_DAYS 天)..."

# 删除超过保留天数的备份
find $BACKUP_BASE_DIR -type f -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete
find $BACKUP_BASE_DIR -type d -empty -delete

CLEANED_COUNT=$(find $BACKUP_BASE_DIR -type f -name "*.tar.gz" -mtime +$RETENTION_DAYS | wc -l)
echo "   ✓ 清理完成，删除了 $CLEANED_COUNT 个旧备份"

# ============================================
# 9. 验证备份完整性
# ============================================

echo ""
echo "9. 验证备份完整性..."

# 检查备份文件是否存在
if [ -f "$BACKUP_ARCHIVE" ]; then
    # 检查文件大小
    ARCHIVE_SIZE_BYTES=$(stat -f%z "$BACKUP_ARCHIVE" || stat -c%s "$BACKUP_ARCHIVE")
    
    if [ $ARCHIVE_SIZE_BYTES -gt 0 ]; then
        echo "   ✓ 备份文件完整性验证通过"
    else
        echo "   ✗ 备份文件大小为 0"
        exit 1
    fi
else
    echo "   ✗ 备份文件不存在"
    exit 1
fi

# ============================================
# 备份总结
# ============================================

echo ""
echo "=== 备份完成 ==="
echo ""

echo "备份总结:"
echo "- 备份时间: $BACKUP_DATE $BACKUP_TIME"
echo "- 备份目录: $BACKUP_DIR"
echo "- 备份文件: $BACKUP_ARCHIVE"
echo "- 总大小: $ARCHIVE_SIZE"
echo "- 保留天数: $RETENTION_DAYS"

echo ""
echo "备份文件列表:"
ls -lh $BACKUP_DIR/

echo ""
echo "恢复命令:"
echo "- 恢复数据库: ./scripts/restore.sh $BACKUP_ARCHIVE"
echo "- 查看备份: tar -tzf $BACKUP_ARCHIVE"

echo ""
echo "✅ Industrial AI Platform 备份成功！"