#!/bin/bash

# Industrial AI Platform 数据库恢复脚本
# 用途: 从备份恢复 PostgreSQL 和 Redis 数据

set -e

echo "=== Industrial AI Platform 数据库恢复 ==="
echo ""

# ============================================
# 配置参数
# ============================================

# 备份文件参数
BACKUP_FILE="${1:-}"
if [ -z "$BACKUP_FILE" ]; then
    echo "用法: ./restore.sh <backup_file.tar.gz>"
    echo ""
    echo "可用的备份文件:"
    ls -lh /backup/industrial-ai/*.tar.gz | tail -10
    exit 1
fi

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    echo "✗ 备份文件不存在: $BACKUP_FILE"
    exit 1
fi

# 解压临时目录
RESTORE_DIR="/tmp/restore-$(date +%Y%m%d-%H%M%S)"

# 数据库配置
DB_HOST="${DB_HOST:-postgres-primary}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-industrial_ai}"

# Redis 配置
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo "恢复参数:"
echo "- 备份文件: $BACKUP_FILE"
echo "- 恢复目录: $RESTORE_DIR"
echo "- 数据库主机: $DB_HOST"
echo "- 数据库名称: $DB_NAME"
echo ""

# ============================================
# 1. 解压备份文件
# ============================================

echo "1. 解压备份文件..."

mkdir -p $RESTORE_DIR
tar -xzf $BACKUP_FILE -C $RESTORE_DIR --strip-components=1

if [ $? -eq 0 ]; then
    echo "   ✓ 备份文件已解压: $RESTORE_DIR"
else
    echo "   ✗ 备份文件解压失败"
    exit 1
fi

# 列出备份内容
echo ""
echo "   备份内容:"
ls -lh $RESTORE_DIR/

# ============================================
# 2. 检查备份清单
# ============================================

echo ""
echo "2. 检查备份清单..."

MANIFEST_FILE=$(find $RESTORE_DIR -name "manifest*.json" | head -1)

if [ -f "$MANIFEST_FILE" ]; then
    echo "   备份清单内容:"
    cat $MANIFEST_FILE | jq '.' || cat $MANIFEST_FILE
else
    echo "   ⚠️ 未找到备份清单，继续恢复..."
fi

# ============================================
# 3. 停止应用服务
# ============================================

echo ""
echo "3. 停止应用服务..."

# 使用 Docker Compose 停止服务
if [ -f "docker-compose.yml" ]; then
    docker-compose stop backend || true
    echo "   ✓ 后端服务已停止"
fi

# 使用 Kubernetes 停止服务
if kubectl get deployment backend -n industrial-ai &> /dev/null; then
    kubectl scale deployment backend --replicas=0 -n industrial-ai
    echo "   ✓ Kubernetes 后端已缩容"
fi

# ============================================
# 4. 恢复 PostgreSQL
# ============================================

echo ""
echo "4. 恢复 PostgreSQL 数据库..."

# 找到数据库备份文件
DB_BACKUP_FILE=$(find $RESTORE_DIR -name "database*.sql.gz" | head -1)

if [ -f "$DB_BACKUP_FILE" ]; then
    echo "   解压数据库备份..."
    gunzip -c $DB_BACKUP_FILE > $RESTORE_DIR/database.sql
    
    # 检查数据库连接
    if PGPASSWORD=$DB_PASSWORD pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
        echo "   ✓ 数据库连接正常"
    else
        echo "   ✗ 数据库连接失败"
        exit 1
    fi
    
    # 恢复数据库
    echo "   恢复数据库数据..."
    PGPASSWORD=$DB_PASSWORD psql \
        -h $DB_HOST \
        -p $DB_PORT \
        -U $DB_USER \
        -d $DB_NAME \
        -f $RESTORE_DIR/database.sql \
        --quiet
    
    if [ $? -eq 0 ]; then
        echo "   ✓ PostgreSQL 数据恢复完成"
    else
        echo "   ✗ PostgreSQL 数据恢复失败"
        exit 1
    fi
else
    echo "   ⚠️ 未找到数据库备份文件，跳过恢复"
fi

# ============================================
# 5. 恢复 Redis
# ============================================

echo ""
echo "5. 恢复 Redis 数据..."

# 找到 Redis 备份文件
REDIS_BACKUP_FILE=$(find $RESTORE_DIR -name "redis*.rdb.gz" | head -1)

if [ -f "$REDIS_BACKUP_FILE" ]; then
    echo "   解压 Redis 备份..."
    gunzip -c $REDIS_BACKUP_FILE > $RESTORE_DIR/redis.rdb
    
    # 停止 Redis 服务
    echo "   停止 Redis 服务..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SHUTDOWN NOSAVE || true
    sleep 2
    
    # 复制 RDB 文件到 Redis 数据目录
    echo "   复制 RDB 文件..."
    if [ -d "/var/lib/redis" ]; then
        cp $RESTORE_DIR/redis.rdb /var/lib/redis/dump.rdb
        echo "   ✓ Redis RDB 文件已复制"
    else
        echo "   ⚠️ Redis 数据目录不存在"
    fi
    
    # 启动 Redis 服务
    echo "   启动 Redis 服务..."
    # 这里假设 Redis 会自动启动，或者需要手动启动
    
    # 等待 Redis 启动
    sleep 5
    
    # 检查 Redis 状态
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping | grep -q "PONG"; then
        echo "   ✓ Redis 服务已启动"
    else
        echo "   ⚠️ Redis 服务启动失败"
    fi
else
    echo "   ⚠️ 未找到 Redis 备份文件，跳过恢复"
fi

# ============================================
# 6. 恢复配置文件 (可选)
# ============================================

echo ""
echo "6. 恢复配置文件..."

CONFIG_BACKUP_FILE=$(find $RESTORE_DIR -name "config*.tar.gz" | head -1)

if [ -f "$CONFIG_BACKUP_FILE" ]; then
    echo "   配置文件备份存在，是否恢复？(y/n)"
    read -t 10 -n 1 RESTORE_CONFIG || RESTORE_CONFIG="n"
    
    if [ "$RESTORE_CONFIG" == "y" ]; then
        tar -xzf $CONFIG_BACKUP_FILE
        echo "   ✓ 配置文件已恢复"
    else
        echo "   ⚠️ 配置文件恢复已跳过"
    fi
else
    echo "   ⚠️ 未找到配置文件备份，跳过恢复"
fi

# ============================================
# 7. 启动应用服务
# ============================================

echo ""
echo "7. 启动应用服务..."

# 使用 Docker Compose 启动服务
if [ -f "docker-compose.yml" ]; then
    docker-compose start backend || docker-compose up -d backend
    echo "   ✓ Docker Compose 后端服务已启动"
fi

# 使用 Kubernetes 启动服务
if kubectl get deployment backend -n industrial-ai &> /dev/null; then
    kubectl scale deployment backend --replicas=2 -n industrial-ai
    echo "   ✓ Kubernetes 后端已扩容"
    
    # 等待 Pods 就绪
    kubectl rollout status deployment/backend -n industrial-ai --timeout=300s
fi

# ============================================
# 8. 验证恢复结果
# ============================================

echo ""
echo "8. 验证恢复结果..."

# 验证数据库
echo "   验证 PostgreSQL..."
DB_TABLE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")
echo "   数据库表数量: $DB_TABLE_COUNT"

# 验证 Redis
echo "   验证 Redis..."
REDIS_KEY_COUNT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT DBSIZE | awk '{print $2}')
echo "   Redis Key 数量: $REDIS_KEY_COUNT"

# 验证应用健康
echo "   验证应用健康..."
sleep 10

HEALTH_STATUS=$(curl -s http://localhost:8080/health/ready || echo "failed")
if echo "$HEALTH_STATUS" | grep -q "ready"; then
    echo "   ✓ 应用健康检查通过"
else
    echo "   ⚠️ 应用健康检查未通过，请手动验证"
fi

# ============================================
# 9. 清理临时文件
# ============================================

echo ""
echo "9. 清理临时文件..."

rm -rf $RESTORE_DIR

echo "   ✓ 临时文件已清理"

# ============================================
# 恢复总结
# ============================================

echo ""
echo "=== 恢复完成 ==="
echo ""

echo "恢复总结:"
echo "- 备份文件: $BACKUP_FILE"
echo "- 恢复时间: $(date)"
echo "- 数据库表数量: $DB_TABLE_COUNT"
echo "- Redis Key 数量: $REDIS_KEY_COUNT"

echo ""
echo "后续操作:"
echo "- 检查数据完整性: 手动验证关键数据"
echo "- 检查服务状态: docker-compose ps 或 kubectl get pods"
echo "- 检查应用日志: docker-compose logs backend 或 kubectl logs deployment/backend"

echo ""
echo "✅ Industrial AI Platform 恢复成功！"