#!/bin/bash

# PostgreSQL 数据库备份脚本
# 用途: 自动化数据库备份 + 保留策略

set -e

echo "=== PostgreSQL 数据库备份 ==="
echo ""

# 环境变量
BACKUP_DIR="${BACKUP_DIR:-/backup/postgres}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-industrial_ai}"
DB_USER="${DB_USER:-industrial_backup}"
DB_PASSWORD="${DB_PASSWORD:-backup-password}"

# 备份文件名
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/full_$DATE.sql.gz"

# 保留策略
FULL_BACKUP_RETENTION_DAYS="${FULL_BACKUP_RETENTION_DAYS:-30}"
INCREMENTAL_RETENTION_DAYS="${INCREMENTAL_RETENTION_DAYS:-7}"

echo "1. 创建备份目录..."
mkdir -p "$BACKUP_DIR"

echo "2. 执行全量备份..."
PGPASSWORD="$DB_PASSWORD" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --format=plain \
    --no-owner \
    --no-privileges \
    --clean \
    --if-exists \
    | gzip > "$BACKUP_FILE"

BACKUP_SIZE=$(stat -f%z "$BACKUP_FILE" 2>/dev/null || stat --printf=%s "$BACKUP_FILE" 2>/dev/null)
echo "   ✓ 备份完成: $BACKUP_FILE"
echo "   文件大小: $(numfmt --to=iec $BACKUP_SIZE 2>/dev/null || echo $BACKUP_SIZE bytes)"

echo ""
echo "3. 清理旧备份..."
# 清理超过保留期限的全量备份
find "$BACKUP_DIR" -name "full_*.sql.gz" -type f -mtime +$FULL_BACKUP_RETENTION_DAYS -delete
echo "   ✓ 清理完成 (保留 $FULL_BACKUP_RETENTION_DAYS 天)"

echo ""
echo "4. 备份校验..."
# 验证备份文件完整性
gunzip -t "$BACKUP_FILE"
echo "   ✓ 备份文件完整"

echo ""
echo "5. 备份列表..."
ls -lh "$BACKUP_DIR"/*.sql.gz 2>/dev/null | tail -5

echo ""
echo "=== 备份信息 ==="
echo "备份文件: $BACKUP_FILE"
echo "备份时间: $DATE"
echo "数据库: $DB_NAME"
echo "保留天数: $FULL_BACKUP_RETENTION_DAYS"
echo ""

# 可选: 上传到远程存储
if [ -n "$REMOTE_BACKUP_URL" ]; then
    echo "6. 上传到远程存储..."
    curl -T "$BACKUP_FILE" "$REMOTE_BACKUP_URL"
    echo "   ✓ 上传完成"
fi

echo "✓ 数据库备份完成！"

# 输出 Prometheus 指标 (可选)
echo ""
echo "# Prometheus Metrics"
echo "db_backup_success 1"
echo "db_backup_size_bytes $BACKUP_SIZE"
echo "db_backup_timestamp $(date +%s)"