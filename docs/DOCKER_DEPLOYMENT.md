# Industrial AI Platform - Docker 部署指南

> 本指南涵盖完整的 Docker 容器化部署流程，适用于生产环境。

---

## 📋 目录

1. [前置条件](#前置条件)
2. [快速启动](#快速启动)
3. [生产部署](#生产部署)
4. [服务架构](#服务架构)
5. [配置说明](#配置说明)
6. [运维命令](#运维命令)
7. [监控与日志](#监控与日志)
8. [故障排查](#故障排查)

---

## 前置条件

### 必需软件

| 软件 | 版本 | 说明 |
|------|------|------|
| Docker | ≥ 24.0 | 容器引擎 |
| Docker Compose | ≥ 2.20 | 多容器编排 |
| Git | ≥ 2.40 | 代码拉取 |

### 硬件要求

| 配置 | 开发环境 | 生产环境 |
|------|----------|----------|
| CPU | 2核 | 4核+ |
| 内存 | 4GB | 8GB+ |
| 存储 | 20GB | 50GB+ |

### 端口规划

| 服务 | 端口 | 说明 |
|------|------|------|
| Frontend | 3000 | Nginx + React |
| Backend | 8080 | Go API |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |

---

## 快速启动

### 1. 克隆项目

```bash
git clone https://gitee.com/yangqinggang-it/industrial-ai-platform.git
cd industrial-ai-platform
```

### 2. 配置环境变量

```bash
# 复制生产配置模板
cp .env.production .env

# 编辑配置（必须修改以下字段）
vim .env
```

**必须修改的配置项：**
- `POSTGRES_PASSWORD` - 数据库密码
- `JWT_SECRET` - JWT 密钥（使用 `openssl rand -base64 32` 生成）
- `LLM_API_KEY` - AI 服务 API Key

### 3. 启动服务

```bash
# 构建并启动所有服务
docker compose up -d --build

# 查看服务状态
docker compose ps
```

### 4. 验证部署

```bash
# 检查后端健康状态
curl http://localhost:8080/health

# 检查前端服务
curl http://localhost:3000/health

# 访问前端界面
open http://localhost:3000
```

**默认登录账号：**
- 用户名：`admin`
- 密码：`Admin@123456`（首次登录后请修改）

---

## 生产部署

### 1. 安全配置

```bash
# 生成安全的 JWT 密钥
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET=$JWT_SECRET" >> .env

# 生成数据库密码
POSTGRES_PASSWORD=$(openssl rand -base64 24)
echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD" >> .env
```

### 2. 域名与 HTTPS 配置

**使用 Nginx 反向代理 + Let's Encrypt：**

```nginx
# /etc/nginx/conf.d/industrial-ai.conf
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
    }

    location /ws {
        proxy_pass http://localhost:8080/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
    }
}
```

### 3. 资源限制配置

在 `docker-compose.yml` 中添加资源限制：

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

---

## 服务架构

```
┌─────────────────────────────────────────────────────────┐
│                    User / Browser                        │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Frontend (Nginx:3000)                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ React SPA   │  │ Static Files│  │ API Proxy   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Backend (Go:8080)                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ REST API    │  │ WebSocket   │  │ AI Agent    │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
          │                         │
          ▼                         ▼
┌───────────────────┐     ┌───────────────────┐
│ PostgreSQL (5432) │     │ Redis (6379)      │
│ - 设备数据        │     │ - 缓存            │
│ - 遥测历史        │     │ - 会话            │
│ - 用户认证        │     │ - 实时推送        │
└───────────────────┘     └───────────────────┘
```

---

## 配置说明

### 后端环境变量

| 变量 | 说明 | 默认值 | 必需 |
|------|------|--------|------|
| `DATABASE_URL` | PostgreSQL 连接串 | - | ✅ |
| `REDIS_URL` | Redis 连接串 | - | ✅ |
| `JWT_SECRET` | JWT 签名密钥 | - | ✅ |
| `JWT_EXPIRY` | Token 有效期 | 24h | |
| `ADMIN_PASSWORD` | 管理员初始密码 | Admin@123456 | ✅ |
| `LLM_API_KEY` | AI 服务 API Key | - | ✅ |
| `LLM_BASE_URL` | AI 服务地址 | DashScope | |
| `LLM_MODEL` | 模型名称 | glm-5 | |
| `CORS_ORIGINS` | CORS 白名单 | localhost | ✅ |
| `LOG_LEVEL` | 日志级别 | info | |
| `LOG_FORMAT` | 日志格式 | json | |

### 前端构建参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `VITE_API_URL` | API 服务地址 | http://localhost:8080 |

---

## 运维命令

### 服务管理

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启单个服务
docker compose restart backend

# 查看服务日志
docker compose logs -f backend

# 进入容器调试
docker compose exec backend sh
```

### 数据备份

```bash
# PostgreSQL 备份
docker compose exec postgres pg_dump -U postgres industrial_ai > backup.sql

# PostgreSQL 恢复
docker compose exec -T postgres psql -U postgres industrial_ai < backup.sql

# Redis 备份
docker compose exec redis redis-cli BGSAVE
docker cp industrial-ai-redis:/data/dump.rdb ./redis_backup.rdb
```

### 数据库迁移

```bash
# 运行 migrations（首次启动自动执行）
docker compose exec backend ./migrations/run.sh

# 手动迁移
docker compose exec postgres psql -U postgres -d industrial_ai -f /docker-entrypoint-initdb.d/001_init.sql
```

### 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重建并重启
docker compose up -d --build

# 仅更新后端
docker compose up -d --build backend
```

---

## 监控与日志

### 健康检查

```bash
# 后端健康状态
curl http://localhost:8080/health

# PostgreSQL 连接
docker compose exec postgres pg_isready

# Redis 连接
docker compose exec redis redis-cli ping
```

### 日志查看

```bash
# 查看所有日志
docker compose logs -f

# 按服务过滤
docker compose logs -f backend --tail=100

# 按时间过滤
docker compose logs backend --since="2024-01-01T00:00:00"
```

### 性能监控

```bash
# 容器资源使用
docker stats

# 后端性能指标
curl http://localhost:8080/api/v1/system/metrics
```

---

## 故障排查

### 常见问题

#### 1. 后端启动失败

**症状：** `docker compose ps` 显示 backend 状态为 unhealthy

**排查：**
```bash
# 查看日志
docker compose logs backend

# 检查数据库连接
docker compose exec postgres pg_isready -U postgres

# 检查 Redis 连接
docker compose exec redis redis-cli ping
```

**常见原因：**
- PostgreSQL 未就绪 → 增加 `start_period`
- Redis 连接失败 → 检查 `REDIS_URL`
- JWT_SECRET 未设置 → 检查 `.env`

#### 2. 前端无法访问 API

**症状：** 前端页面空白或 API 请求失败

**排查：**
```bash
# 检查 CORS 配置
grep CORS_ORIGINS .env

# 检查 Nginx 代理
docker compose exec frontend cat /etc/nginx/conf.d/default.conf

# 测试 API 直连
curl http://localhost:8080/health
```

#### 3. WebSocket 连接失败

**症状：** 实时数据不更新

**排查：**
```bash
# 检查 WebSocket 端点
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" \
  http://localhost:8080/ws
```

#### 4. AI Agent 无响应

**症状：** AI 对话返回错误

**排查：**
```bash
# 检查 LLM API Key
grep LLM_API_KEY .env

# 测试 API 直连
curl -X POST ${LLM_BASE_URL}/chat/completions \
  -H "Authorization: Bearer ${LLM_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"model":"glm-5","messages":[{"role":"user","content":"test"}]}'
```

### 清理与重置

```bash
# 停止并删除所有容器
docker compose down

# 删除数据卷（清空数据库）
docker compose down -v

# 重新部署
docker compose up -d --build
```

---

## 附录：快速命令参考

```bash
# ========== 服务管理 ==========
docker compose up -d                # 启动
docker compose down                 # 停止
docker compose restart              # 重启
docker compose ps                   # 状态
docker compose logs -f              # 日志

# ========== 更新部署 ==========
git pull && docker compose up -d --build

# ========== 数据备份 ==========
docker compose exec postgres pg_dump -U postgres industrial_ai > backup.sql

# ========== 健康检查 ==========
curl http://localhost:8080/health
curl http://localhost:3000/health
docker compose exec redis redis-cli ping
docker compose exec postgres pg_isready
```

---

**部署完成后，建议：**
1. 修改管理员密码
2. 配置 HTTPS（生产必需）
3. 设置数据库定时备份
4. 启用监控（Prometheus + Grafana）