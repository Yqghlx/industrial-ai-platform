# 数据库安全配置指南

> **Industrial AI Platform PostgreSQL 数据库安全最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 数据库安全增强概述

Phase 4 数据库安全加固包含以下改进：

| 功能 | 描述 | 安全收益 |
|------|------|---------|
| **SSL/TLS 连接** | 数据库连接加密 | 防止数据传输泄露 |
| **参数化查询** | 使用 $1, $2 占位符 | 防止 SQL 注入 |
| **连接池限制** | 最大连接数限制 | 防止连接耗尽攻击 |
| **用户权限隔离** | 应用专用数据库用户 | 最小权限原则 |
| **审计日志** | 操作审计记录 | 追踪异常操作 |
| **定期备份** | 自动数据备份 | 数据恢复保障 |

---

## 🔒 SQL 注入防护

### ✅ 当前代码安全性审计

**Industrial AI Platform 代码库已使用参数化查询，防止 SQL 注入：**

```go
// ✅ 正确做法 - 参数化查询
query := `
    SELECT id, username, role, tenant_id
    FROM users
    WHERE username = $1 AND tenant_id = $2
`
err := r.db.QueryRow(query, username, tenantID).Scan(...)

// ❌ 错误做法 - 字符串拼接 (本代码库未使用)
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
```

### 安全查询检查清单

| 检查项 | 当前状态 | 说明 |
|--------|---------|------|
| **参数化查询** | ✅ 已实现 | 所有查询使用 $1, $2 占位符 |
| **字符串拼接** | ✅ 未使用 | 不使用 fmt.Sprintf 构造 SQL |
| **用户输入验证** | ✅ 已实现 | Gin binding 验证输入 |
| **动态 SQL** | ✅ 避免 | 无动态构造 SQL 语句 |

---

## 🔐 PostgreSQL SSL/TLS 连接

### 生产环境 SSL 配置

**1️⃣ PostgreSQL 服务器 SSL 配置**

```bash
# 在 PostgreSQL 容器中生成 SSL 证书
docker exec -it industrial-ai-postgres bash

# 生成服务器证书
openssl req -new -newkey rsa:4096 -nodes \
    -keyout /var/lib/postgresql/data/server.key \
    -out /var/lib/postgresql/data/server.crt \
    -days 365 \
    -subj "/C=CN/ST=Shanghai/L=Shanghai/O=IndustrialAI/CN=postgres"

# 设置权限
chmod 600 /var/lib/postgresql/data/server.key
chmod 644 /var/lib/postgresql/data/server.crt
chown postgres:postgres /var/lib/postgresql/data/server.key
chown postgres:postgres /var/lib/postgresql/data/server.crt
```

**2️⃣ PostgreSQL 配置启用 SSL**

```bash
# postgresql.conf 配置
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
ssl_ca_file = ''  # 如果使用 CA 验证
ssl_min_protocol_version = 'TLSv1.2'
ssl_max_protocol_version = 'TLSv1.3'
```

**3️⃣ 应用端 SSL 连接**

```yaml
# docker-compose.yml (生产环境)
environment:
  - DATABASE_URL=postgres://industrial_user:password@postgres:5432/industrial_ai?sslmode=require
  # 或更严格验证:
  - DATABASE_URL=postgres://industrial_user:password@postgres:5432/industrial_ai?sslmode=verify-full&sslcert=/app/certs/client.crt&sslkey=/app/certs/client.key
```

### SSL 模式说明

| sslmode | 描述 | 安全级别 |
|---------|------|---------|
| `disable` | 不使用 SSL | ❌ 不安全 |
| `allow` | 尝试 SSL，失败则回退 | ⚠️ 低 |
| `prefer` | 优先 SSL | ⚠️ 低 |
| `require` | 强制 SSL | ✅ 中 |
| `verify-ca` | 强制 SSL + CA 验证 | ✅ 高 |
| `verify-full` | 强制 SSL + CA + 主机名验证 | ✅ 最高 |

---

## 📊 连接池安全配置

### 推荐配置参数

```yaml
# 连接池配置
database:
  max_open_conns: 50      # 最大打开连接数
  max_idle_conns: 10      # 最大空闲连接数
  conn_max_lifetime: 1h   # 连接最大生命周期
  conn_max_idle_time: 10m # 空闲连接超时
```

### Go 代码配置

```go
// backend/pkg/database/config.go
func ConfigureConnectionPool(db *sql.DB, maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) {
    db.SetMaxOpenConns(maxOpen)        // 防止连接耗尽
    db.SetMaxIdleConns(maxIdle)        // 保持一定空闲连接
    db.SetConnMaxLifetime(maxLifetime) // 防止长时间连接
    db.SetConnMaxIdleTime(maxIdleTime) // 清理空闲连接
}
```

---

## 👤 数据库用户权限隔离

### 最小权限原则

**创建应用专用用户（非超级用户）：**

```sql
-- 创建应用专用用户
CREATE USER industrial_app WITH PASSWORD 'secure-random-password';

-- 仅授予必要权限
GRANT CONNECT ON DATABASE industrial_ai TO industrial_app;
GRANT USAGE ON SCHEMA public TO industrial_app;

-- 表权限 (仅 SELECT, INSERT, UPDATE)
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO industrial_app;

-- 禁止删除和创建表
-- REVOKE DELETE, CREATE, DROP ON ALL TABLES IN SCHEMA public FROM industrial_app;

-- 序列权限 (用于自增 ID)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO industrial_app;
```

### 权限矩阵

| 操作 | postgres (管理员) | industrial_app (应用) |
|------|------------------|----------------------|
| CONNECT | ✅ | ✅ |
| SELECT | ✅ | ✅ |
| INSERT | ✅ | ✅ |
| UPDATE | ✅ | ✅ |
| DELETE | ✅ | ❌ (通过应用逻辑) |
| CREATE | ✅ | ❌ |
| DROP | ✅ | ❌ |
| ALTER | ✅ | ❌ |
| TRUNCATE | ✅ | ❌ |

---

## 📝 数据库审计日志

### PostgreSQL 审计配置

```bash
# postgresql.conf
log_statement = 'ddl'           # 记录 DDL 操作 (CREATE/DROP/ALTER)
log_connections = on            # 记录连接
log_disconnections = on         # 记录断开连接
log_line_prefix = '%t [%p] [%u@%d] '  # 时间/进程/用户/数据库

# 使用 pgaudit 扩展 (更详细的审计)
shared_preload_libraries = 'pgaudit'
pgaudit.log = 'write, ddl'      # 记录写入和 DDL
pgaudit.log_catalog = on        # 记录系统表操作
```

### 审计日志查询

```sql
-- 查看审计日志
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- 查看最近连接
SELECT * FROM pg_stat_activity ORDER BY backend_start DESC LIMIT 10;

-- 查看用户操作统计
SELECT usename, application_name, count(*) 
FROM pg_stat_activity 
GROUP BY usename, application_name;
```

---

## 💾 数据备份策略

### 自动备份配置

```yaml
# 备份计划
backups:
  full_backup:
    schedule: "0 2 * * *"     # 每天凌晨 2 点
    retention: 30             # 保留 30 天
    storage: "/backup/postgres"
  
  incremental_backup:
    schedule: "0 */6 * * *"   # 每 6 小时
    retention: 7              # 保留 7 天
```

### 备份脚本

```bash
#!/bin/bash
# scripts/db-backup.sh

BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="industrial_ai"

# 全量备份
pg_dump -h postgres -U industrial_backup $DB_NAME | gzip > $BACKUP_DIR/full_$DATE.sql.gz

# 保留 30 天
find $BACKUP_DIR -name "full_*.sql.gz" -mtime +30 -delete

echo "Backup completed: full_$DATE.sql.gz"
```

---

## 🛡️ 安全检查清单

| 检查项 | 要求 | 状态 |
|--------|------|------|
| **SSL 连接** | sslmode=require 或更高 | ⏳ 待配置 |
| **参数化查询** | 所有查询使用占位符 | ✅ 已实现 |
| **用户隔离** | 应用专用数据库用户 | ⏳ 待配置 |
| **连接池限制** | max_open_conns ≤ 100 | ⏳ 待配置 |
| **审计日志** | 记录关键操作 | ⏳ 待配置 |
| **定期备份** | 每日备份 + 保留策略 | ⏳ 待配置 |
| **密码强度** | ≥ 16 字符随机 | ⏳ 待配置 |
| **连接超时** | conn_max_lifetime ≤ 1h | ⏳ 待配置 |

---

## 🔧 Docker Compose 生产配置

```yaml
# docker-compose.prod.yml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres          # 管理员
      - POSTGRES_PASSWORD=${POSTGRES_ADMIN_PASSWORD}  # 强密码
      - POSTGRES_DB=industrial_ai
      - POSTGRES_INITDB_ARGS="--data-checksums"  # 数据校验
    command:
      - "postgres"
      - "-c"
      - "ssl=on"
      - "-c"
      - "ssl_cert_file=/var/lib/postgresql/data/server.crt"
      - "-c"
      - "ssl_key_file=/var/lib/postgresql/data/server.key"
      - "-c"
      - "log_statement=ddl"
      - "-c"
      - "log_connections=on"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./ssl:/var/lib/postgresql/ssl:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d industrial_ai"]
      interval: 30s
      timeout: 10s
      retries: 3

  backend:
    environment:
      - DATABASE_URL=postgres://industrial_app:${APP_DB_PASSWORD}@postgres:5432/industrial_ai?sslmode=require
      - DB_MAX_OPEN_CONNS=50
      - DB_MAX_IDLE_CONNS=10
      - DB_CONN_MAX_LIFETIME=1h
```

---

## ✅ 验证步骤

### 1️⃣ 验证 SSL 连接

```bash
# 检查 SSL 状态
docker exec -it industrial-ai-postgres psql -U postgres -c "SHOW ssl;"

# 测试 SSL 连接
docker exec -it industrial-ai-backend sh -c "psql '$DATABASE_URL' -c 'SELECT 1;'"

# 查看连接 SSL 状态
docker exec -it industrial-ai-postgres psql -U postgres -c "
SELECT pid, usename, ssl, ssl_version, ssl_cipher 
FROM pg_stat_ssl 
JOIN pg_stat_activity USING(pid) 
WHERE ssl = true;
"
```

### 2️⃣ 验证 SQL 注入防护

```bash
# 测试 SQL 注入攻击 (应失败)
curl -X POST http://backend:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin\' OR \'1\'=\'1", "password": "test"}'

# 期望: 401 Unauthorized (攻击被阻止)
```

### 3️⃣ 验证权限隔离

```bash
# 以应用用户连接，尝试创建表 (应失败)
docker exec -it industrial-ai-postgres psql -U industrial_app -d industrial_ai -c "CREATE TABLE test_hack(id int);"

# 期望: ERROR: permission denied
```

---

**最后更新**: 2026-05-13  
**审核人**: DBA + 安全团队