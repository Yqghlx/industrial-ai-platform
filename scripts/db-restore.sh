#!/bin/bash

# PostgreSQL 数据库恢复脚本
# 用途: 从备份恢复数据库

set -e

echo "=== PostgreSQL 数据库恢复 ==="
echo ""

# 参数
BACKUP_FILE="${1:-}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-industrial_ai}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

if [ -z "$BACKUP_FILE" ]; then
    echo "错误: 请指定备份文件"
    echo "用法: $0 <backup_file.sql.gz>"
    echo ""
    echo "可用备份:"
    ls -lh /backup/postgres/*.sql.gz 2>/dev/null | tail -10
    exit 1
fi

echo "备份文件: $BACKUP_FILE"
echo "目标数据库: $DB_NAME"
echo ""

echo "⚠️ 警告: 恢复将覆盖现有数据！"
read -p "确认继续？(yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "取消恢复"
    exit 0
fi

echo ""
echo "1. 解压备份文件..."
TEMP_SQL="/tmp/restore_$DATE.sql"
gunzip -c "$BACKUP_FILE" > "$TEMP_SQL"
echo "   ✓ 解压完成"

echo ""
echo "2. 检查数据库连接..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;"
echo "   ✓ 连接正常"

echo ""
echo "3. 执行恢复..."
# 注意: 恢复会先清空数据 (--clean --if-exists)
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -f "$TEMP_SQL"

echo "   ✓ 恢复完成"

echo ""
echo "4. 验证数据..."
# 检查关键表
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
SELECT 'users' as table_name, count(*) as count FROM users
UNION ALL
SELECT 'devices', count(*) FROM devices
UNION ALL
SELECT 'tenants', count(*) FROM tenants;
EOF

echo ""
echo "5. 清理临时文件..."
rm "$TEMP_SQL"

echo ""
echo "✓ 数据库恢复完成！"