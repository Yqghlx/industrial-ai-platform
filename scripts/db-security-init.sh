#!/bin/bash

# PostgreSQL 数据库安全初始化脚本
# 用途: 创建应用专用用户 + 配置最小权限

set -e

echo "=== PostgreSQL 数据库安全初始化 ==="
echo ""

# 环境变量
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-industrial_ai}"
ADMIN_USER="${ADMIN_USER:-postgres}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-postgres}"

# 应用用户配置
APP_USER="${APP_USER:-industrial_app}"
APP_PASSWORD="${APP_PASSWORD:-$(openssl rand -base64 24)}"

echo "1. 连接 PostgreSQL..."
PGPASSWORD="$ADMIN_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -d $DB_NAME -c "SELECT version();"

echo ""
echo "2. 创建应用专用用户..."
PGPASSWORD="$ADMIN_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -d $DB_NAME << EOF
-- 创建应用用户 (非超级用户)
DO \$\$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$APP_USER') THEN
        CREATE USER $APP_USER WITH PASSWORD '$APP_PASSWORD';
    END IF;
END \$\$;

-- 授予连接权限
GRANT CONNECT ON DATABASE $DB_NAME TO $APP_USER;

-- 授予 schema 使用权限
GRANT USAGE ON SCHEMA public TO $APP_USER;

-- 授予表权限 (仅 SELECT, INSERT, UPDATE)
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO $APP_USER;

-- 授予序列权限 (用于自增 ID)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $APP_USER;

-- 设置默认权限 (新建表自动授权)
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE ON TABLES TO $APP_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO $APP_USER;
EOF

echo "   ✓ 应用用户创建完成: $APP_USER"
echo ""

echo "3. 创建备份用户..."
BACKUP_USER="${BACKUP_USER:-industrial_backup}"
BACKUP_PASSWORD="${BACKUP_PASSWORD:-$(openssl rand -base64 24)}"

PGPASSWORD="$ADMIN_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -d $DB_NAME << EOF
-- 创建备份用户
DO \$\$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$BACKUP_USER') THEN
        CREATE USER $BACKUP_USER WITH PASSWORD '$BACKUP_PASSWORD';
    END IF;
END \$\$;

-- 授予备份权限
GRANT CONNECT ON DATABASE $DB_NAME TO $BACKUP_USER;
GRANT USAGE ON SCHEMA public TO $BACKUP_USER;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO $BACKUP_USER;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $BACKUP_USER;
EOF

echo "   ✓ 备份用户创建完成: $BACKUP_USER"
echo ""

echo "4. 验证权限..."
PGPASSWORD="$APP_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $APP_USER -d $DB_NAME -c "SELECT current_user, session_user;"

echo ""
echo "=== 用户信息 ==="
echo "应用用户: $APP_USER"
echo "应用密码: $APP_PASSWORD"
echo "备份用户: $BACKUP_USER"
echo "备份密码: $BACKUP_PASSWORD"
echo ""
echo "⚠️ 请妥善保存密码！建议存储到 Secrets 管理系统"
echo ""

echo "5. 配置 SSL 连接 (如果需要)..."
echo "   请参考 docs/security/Database-Security-Guide.md 配置 SSL"
echo ""

echo "✓ 数据库安全初始化完成！"