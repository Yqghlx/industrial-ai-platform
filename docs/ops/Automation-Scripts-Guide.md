# 自动化运维脚本指南

> **Industrial AI Platform 自动化运维最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 自动化运维概述

Phase 4 P1 运维自动化脚本目标：

| 功能 | 当前状态 | 目标 |
|------|---------|------|
| **部署** | 手动部署 | 一键部署 |
| **备份** | 无备份 | 自动备份 |
| **恢复** | 无恢复 | 一键恢复 |
| **巡检** | 手动检查 | 自动巡检 |

---

## 🔄 自动化运维流程

### 部署流程

```
┌─────────────────────────────────────────┐
│  1. 拉取最新代码                          │
│  git pull origin main                    │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. 构建应用                              │
│  go build / npm build                    │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 构建 Docker 镜像                      │
│  docker build -t app:version .           │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 推送镜像到仓库                        │
│  docker push registry/app:version        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 部署应用                              │
│  kubectl apply / docker-compose up       │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  6. 验证部署                              │
│  health check / smoke test               │
└─────────────────────────────────────────┘
```

### 备份流程

```
┌─────────────────────────────────────────┐
│  1. 检查备份配置                          │
│  读取备份策略                            │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. 执行数据库备份                        │
│  pg_dump / mysqldump                     │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 执行文件备份                          │
│  rsync / tar                             │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 压缩备份文件                          │
│  gzip / tar.gz                           │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 上传到远程存储                        │
│  S3 / OSS / NFS                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  6. 清理旧备份                            │
│  delete backups > retention              │
└─────────────────────────────────────────┘
```

---

## 🔧 部署脚本

### 部署脚本结构

```bash
#!/bin/bash

# 部署脚本
# 用途: 自动化部署 Industrial AI Platform

# 1. 环境检查
check_environment() {
    echo "检查环境..."
    # 检查必要工具
    # 检查配置文件
    # 检查权限
}

# 2. 构建应用
build_application() {
    echo "构建应用..."
    # 编译 Go 后端
    # 构建 React 前端
}

# 3. 构建 Docker 镜像
build_docker_images() {
    echo "构建 Docker 镜像..."
    # 构建后端镜像
    # 构建前端镜像
    # 构建其他服务镜像
}

# 4. 推送镜像
push_docker_images() {
    echo "推送 Docker 镜像..."
    # 推送到 Docker Registry
}

# 5. 部署应用
deploy_application() {
    echo "部署应用..."
    # 使用 Kubernetes 或 Docker Compose
}

# 6. 验证部署
verify_deployment() {
    echo "验证部署..."
    # 健康检查
    # Smoke Test
}

# 主函数
main() {
    check_environment
    build_application
    build_docker_images
    push_docker_images
    deploy_application
    verify_deployment
}

main
```

---

## 📦 备份脚本

### 备份脚本结构

```bash
#!/bin/bash

# 备份脚本
# 用途: 自动备份 Industrial AI Platform 数据

# 备份配置
BACKUP_DIR="/backup/industrial-ai"
RETENTION_DAYS=30
DATABASE_HOST="postgres-primary"
DATABASE_NAME="industrial_ai"

# 1. 创建备份目录
create_backup_directory() {
    mkdir -p $BACKUP_DIR/$(date +%Y-%m-%d)
}

# 2. 备份 PostgreSQL
backup_postgresql() {
    pg_dump -h $DATABASE_HOST -U postgres -d $DATABASE_NAME \
        > $BACKUP_DIR/$(date +%Y-%m-%d)/database.sql
}

# 3. 备份 Redis
backup_redis() {
    redis-cli BGSAVE
    cp /var/lib/redis/dump.rdb $BACKUP_DIR/$(date +%Y-%m-%d)/redis.rdb
}

# 4. 压缩备份
compress_backup() {
    tar -czf $BACKUP_DIR/$(date +%Y-%m-%d).tar.gz \
        $BACKUP_DIR/$(date +%Y-%m-%d)
}

# 5. 上传到远程存储
upload_backup() {
    aws s3 cp $BACKUP_DIR/$(date +%Y-%m-%d).tar.gz \
        s3://backup-bucket/industrial-ai/
}

# 6. 清理旧备份
cleanup_old_backups() {
    find $BACKUP_DIR -mtime +$RETENTION_DAYS -delete
}

# 主函数
main() {
    create_backup_directory
    backup_postgresql
    backup_redis
    compress_backup
    upload_backup
    cleanup_old_backups
}

main
```

---

## 🔄 恢复脚本

### 恢复脚本结构

```bash
#!/bin/bash

# 恢复脚本
# 用途: 从备份恢复 Industrial AI Platform 数据

# 恢复配置
BACKUP_FILE="/backup/industrial-ai/2026-05-13.tar.gz"
DATABASE_HOST="postgres-primary"
DATABASE_NAME="industrial_ai"

# 1. 解压备份
extract_backup() {
    tar -xzf $BACKUP_FILE -C /tmp/restore
}

# 2. 停止服务
stop_services() {
    docker-compose stop backend
    docker-compose stop redis
}

# 3. 恢复 PostgreSQL
restore_postgresql() {
    psql -h $DATABASE_HOST -U postgres -d $DATABASE_NAME \
        < /tmp/restore/database.sql
}

# 4. 恢复 Redis
restore_redis() {
    cp /tmp/restore/redis.rdb /var/lib/redis/dump.rdb
}

# 5. 启动服务
start_services() {
    docker-compose start redis
    docker-compose start backend
}

# 6. 验证恢复
verify_restore() {
    # 检查数据完整性
    # 检查服务状态
}

# 主函数
main() {
    extract_backup
    stop_services
    restore_postgresql
    restore_redis
    start_services
    verify_restore
}

main
```

---

## 🔍 系统巡检脚本

### 巡检脚本结构

```bash
#!/bin/bash

# 系统巡检脚本
# 用途: 全面检查系统健康状态

# 1. 检查系统资源
check_system_resources() {
    echo "=== 系统资源检查 ==="
    # CPU 使用率
    # 内存使用率
    # 磁盘使用率
    # 网络状态
}

# 2. 检查服务状态
check_services() {
    echo "=== 服务状态检查 ==="
    # Docker 服务
    # Kubernetes Pods
    # 应用健康
}

# 3. 检查数据库
check_database() {
    echo "=== 数据库检查 ==="
    # PostgreSQL 连接
    # Redis 连接
    # 数据库性能
}

# 4. 检查日志
check_logs() {
    echo "=== 日志检查 ==="
    # 错误日志
    # 警告日志
    # 关键日志
}

# 5. 检查安全
check_security() {
    echo "=== 安全检查 ==="
    # 防火墙状态
    # 证书有效性
    # 权限检查
}

# 6. 生成报告
generate_report() {
    echo "=== 生成巡检报告 ==="
    # 合并所有检查结果
    # 生成 HTML/PDF 报告
}

# 主函数
main() {
    check_system_resources
    check_services
    check_database
    check_logs
    check_security
    generate_report
}

main
```

---

## 📊 Cron 定时任务

### 定时备份配置

```bash
# 每日凌晨 2 点执行备份
0 2 * * * /scripts/backup.sh >> /logs/backup.log 2>&1

# 每周日凌晨 3 点执行系统巡检
0 3 * * 0 /scripts/inspection.sh >> /logs/inspection.log 2>&1
```

---

## ✅ 自动化运维验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **部署脚本** | 一键部署 | 执行测试 |
| **备份脚本** | 自动备份 | 检查备份文件 |
| **恢复脚本** | 一键恢复 | 恢复测试 |
| **巡检脚本** | 自动巡检 | 检查报告 |
| **定时任务** | 正常运行 | Cron 日志 |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team