# Industrial AI Platform - 部署文档

本文档描述了 Industrial AI Platform 的部署配置和流程。

## 目录

- [环境变量清单](#环境变量清单)
- [必需配置项](#必需配置项)
- [可选配置项](#可选配置项)
- [配置验证方法](#配置验证方法)
- [常见问题排查](#常见问题排查)
- [Kubernetes 部署](#kubernetes-部署)
- [Docker 部署](#docker-部署)

---

## 环境变量清单

### 数据库配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `DATABASE_URL` | 必需 | - | PostgreSQL 连接字符串。格式: `postgres://user:password@host:5432/dbname?sslmode=require` |
| `DB_MAX_OPEN_CONNS` | 可选 | 25 | 数据库最大打开连接数 |
| `DB_MAX_IDLE_CONNS` | 可选 | 10 | 数据库最大空闲连接数 |
| `DB_CONN_MAX_LIFETIME` | 可选 | 1800 | 数据库连接最大生命周期（秒） |
| `DB_CONN_MAX_IDLE_TIME` | 可选 | 300 | 数据库空闲连接最大存活时间（秒） |

### Redis 缓存配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `REDIS_URL` | 必需 | - | Redis 连接字符串。格式: `redis://host:6379/0`。**Phase 1 P0-01修复：从硬编码改为环境变量** |
| `REDIS_PASSWORD` | 可选 | - | Redis 密码（K8s 环境通过 Secret 提供） |
| `REDIS_POOL_SIZE` | 可选 | 50 | Redis 连接池大小 |
| `REDIS_MIN_IDLE_CONNS` | 可选 | 10 | Redis 最小空闲连接数 |
| `REDIS_MAX_RETRIES` | 可选 | 3 | Redis 最大重试次数 |
| `REDIS_READ_TIMEOUT` | 可选 | 3 | Redis 读超时（秒） |
| `REDIS_WRITE_TIMEOUT` | 可选 | 3 | Redis 写超时（秒） |
| `CACHE_ENABLED` | 可选 | true | 是否启用缓存 |
| `CACHE_PREFIX` | 可选 | `iai:` | 缓存键前缀 |

### 认证配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `JWT_SECRET` | 必需(生产) | - | JWT 签名密钥。**生产环境必须设置**，建议使用 256 位随机字符串 |
| `ADMIN_PASSWORD` | 可选 | 随机生成 | 管理员密码。如未设置，首次启动时自动生成并输出到日志 |

### LLM/AI 配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `LLM_API_KEY` | 可选 | - | LLM API 密钥（百炼 GLM-5） |
| `LLM_BASE_URL` | 可选 | `https://coding.dashscope.aliyuncs.com/v1` | LLM API 地址 |
| `LLM_MODEL` | 可选 | `glm-5` | LLM 模型名称 |
| `LLM_HTTP_TIMEOUT` | 可选 | 30 | LLM HTTP 请求超时（秒） |
| `LLM_MAX_IDLE_CONNS` | 可选 | 100 | LLM HTTP 客户端最大空闲连接数 |
| `LLM_MAX_IDLE_CONNS_PER_HOST` | 可选 | 10 | LLM HTTP 客户端每个主机最大空闲连接数 |
| `LLM_IDLE_CONN_TIMEOUT` | 可选 | 90 | LLM HTTP 客户端空闲连接超时（秒） |

### 服务器配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `PORT` | 可选 | 8080 | 服务监听端口 |
| `REQUEST_TIMEOUT` | 可选 | 15 | HTTP 请求超时（秒） |
| `CORS_ORIGINS` | 必需(生产) | - | CORS 允许的源（逗号分隔）。**生产环境不允许 `*`**（SEC-HIGH-04修复） |

### WebSocket 配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `WS_COMPRESSION_ENABLED` | 可选 | true | 是否启用 WebSocket 压缩 |
| `WS_COMPRESSION_LEVEL` | 可选 | 6 | 压缩级别 (1-9)，6 为平衡速度和压缩率 |
| `WS_COMPRESSION_MIN_SIZE` | 可选 | 1024 | 触发压缩的最小消息大小（字节），支持 KB/MB 格式 |

### 安全配置 (WAF)

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `WAF_ENABLED` | 可选 | true | 是否启用 WAF |
| `WAF_BLOCKED_USER_AGENTS` | 可选 | - | 阻止的 User-Agent（逗号分隔） |
| `WAF_BLOCK_EMPTY_USER_AGENT` | 可选 | true | 是否阻止空 User-Agent |
| `WAF_MAX_REQUEST_SIZE` | 可选 | 10MB | 最大请求大小 |
| `WAF_MAX_ARGS_LENGTH` | 可选 | 1000 | 最大参数长度 |

### 限流配置

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `RATE_LIMIT_ENABLED` | 可选 | true | 是否启用限流 |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | 可选 | 100 | 每秒最大请求数 |
| `RATE_LIMIT_BURST` | 可选 | 200 | 突发流量限制 |
| `RATE_LIMIT_WINDOW` | 可选 | 60 | 限流窗口（秒） |

### 环境标识

| 变量名 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| `ENV` | 可选 | development | 环境标识：`development`、`production` |
| `GIN_MODE` | 可选 | release | Gin 模式：`debug`、`release` |

---

## 必需配置项

生产环境部署必须配置以下环境变量：

### 1. DATABASE_URL (必需)

```bash
# PostgreSQL 连接字符串
DATABASE_URL=postgres://industrial_user:your_password@postgres-host:5432/industrial_ai?sslmode=require
```

**注意事项：**
- 生产环境必须使用 SSL (`sslmode=require` 或 `sslmode=verify-full`)
- 密码不要包含特殊字符 `#` 或 `?`，可能影响 URL 解析
- 连接字符串应存储在 Kubernetes Secret 中

### 2. JWT_SECRET (生产必需)

```bash
# JWT 签名密钥（256位随机字符串）
JWT_SECRET=$(openssl rand -base64 32)
```

**生成方式：**
```bash
# 方法1: openssl
openssl rand -base64 32

# 方法2: Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

### 3. CORS_ORIGINS (生产必需)

```bash
# 生产环境必须显式指定允许的源
CORS_ORIGINS=https://your-domain.com,https://app.your-domain.com
```

**注意：** 生产环境禁止使用 `*` 通配符，这是安全风险。

---

## 可选配置项

### 缓存配置

如果 Redis 可用，启用缓存可显著提升性能：

```bash
REDIS_URL=redis://redis-host:6379/0
CACHE_ENABLED=true
CACHE_PREFIX=iai:
```

### AI 功能配置

启用 AI Agent 功能需要配置 LLM：

```bash
LLM_API_KEY=your_api_key
LLM_BASE_URL=https://coding.dashscope.aliyuncs.com/v1
LLM_MODEL=glm-5
```

### 性能调优配置

高负载场景建议调优连接池：

```bash
# 数据库连接池
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=1800

# Redis 连接池
REDIS_POOL_SIZE=100
REDIS_MIN_IDLE_CONNS=20
```

---

## 配置验证方法

### 1. 启动时验证

应用启动时会自动验证必需配置：

```bash
# 启动时检查配置
./server

# 如果缺少必需配置，会输出错误信息：
# Configuration Error:
# ===================
#   - DATABASE_URL: is required. Please set...
#   - JWT_SECRET: is required in production...
```

### 2. 配置检查命令

```bash
# 检查环境变量是否设置
env | grep -E "DATABASE_URL|JWT_SECRET|REDIS_URL|LLM_API_KEY"

# 验证数据库连接
psql "$DATABASE_URL" -c "SELECT 1"

# 验证 Redis 连接
redis-cli -u "$REDIS_URL" ping
```

### 3. 健康检查 API

```bash
# 基础健康检查
curl http://localhost:8080/health

# 存活检查
curl http://localhost:8080/health/live

# 就绪检查
curl http://localhost:8080/health/ready
```

### 4. Kubernetes 配置验证

```bash
# 检查 Secret 是否正确配置
kubectl get secrets -n industrial-ai

# 检查 Secret 内容（base64 解码）
kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath='{.data.jwt-secret}' | base64 -d

# 检查 ConfigMap
kubectl get configmaps -n industrial-ai

# 验证 Pod 状态
kubectl get pods -n industrial-ai
kubectl describe pod backend-xxx -n industrial-ai
```

---

## 常见问题排查

### 1. 数据库连接失败

**症状：** `connection refused` 或 `no such host`

**排查步骤：**

```bash
# 检查数据库是否运行
docker ps | grep postgres

# 检查网络连通性
ping postgres-host
telnet postgres-host 5432

# 检查 DATABASE_URL 格式
echo $DATABASE_URL

# 检查 SSL 配置（如果使用 sslmode=require）
openssl s_client -connect postgres-host:5432
```

**解决方案：**
- 确保 PostgreSQL 服务运行
- 检查防火墙规则
- 验证连接字符串格式
- 生产环境使用 SSL

### 2. Redis 连接超时

**症状：** `dial tcp: connection refused`

**排查步骤：**

```bash
# 检查 Redis 是否运行
redis-cli ping

# 检查连接
redis-cli -u "$REDIS_URL" ping
```

**解决方案：**
- 如果 Redis 不可用，应用会自动使用内存缓存作为 fallback
- 检查 Redis 密码配置

### 3. JWT 认证失败

**症状：** `invalid token` 或 `token expired`

**排查步骤：**

```bash
# 检查 JWT_SECRET 是否一致
kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath='{.data.jwt-secret}' | base64 -d

# 如果使用多副本，确保所有 Pod 使用相同密钥
kubectl exec -it backend-xxx -n industrial-ai -- env | grep JWT_SECRET
```

**解决方案：**
- 确保所有 Pod 使用相同的 JWT_SECRET
- 检查密钥轮换后是否重启了 Pod

### 4. CORS 错误

**症状：** 前端请求被浏览器拦截

**排查步骤：**

```bash
# 检查 CORS 配置
curl -H "Origin: https://your-domain.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/api/v1/endpoint
```

**解决方案：**
- 确保 `CORS_ORIGINS` 包含前端域名
- 生产环境不要使用 `*`
- 检查端口是否匹配（80, 443, 3000 等）

### 5. AI 功能不工作

**症状：** Agent 响应超时或失败

**排查步骤：**

```bash
# 检查 API Key
curl -H "Authorization: Bearer $LLM_API_KEY" \
     $LLM_BASE_URL/models

# 检查超时配置
echo $LLM_HTTP_TIMEOUT
```

**解决方案：**
- 验证 LLM API Key 有效
- 检查网络可达性
- 调整超时时间（AI 响应可能较慢）

### 6. 健康检查失败

**症状：** Pod 被重启

**排查步骤：**

```bash
# 查看 Pod 事件
kubectl describe pod backend-xxx -n industrial-ai

# 查看 Pod 日志
kubectl logs backend-xxx -n industrial-ai --previous

# 手动测试健康检查
kubectl exec -it backend-xxx -n industrial-ai -- curl http://localhost:8080/health/live
```

**解决方案：**
- 检查启动时间是否超过 `initialDelaySeconds`
- 检查应用是否有依赖启动问题（数据库、Redis）
- 调整健康检查参数

### 7. 内存不足 / OOM Killed

**症状：** Pod 被 OOM Killed

**排查步骤：**

```bash
# 查看 Pod 内存使用
kubectl top pods -n industrial-ai

# 查看 Pod 事件
kubectl describe pod backend-xxx -n industrial-ai | grep -i oom
```

**解决方案：**
- 增加内存限制
- 检查内存泄漏
- 优化数据库连接池大小

---

## Kubernetes 部署

### 部署前准备

1. **创建 Namespace：**

```bash
kubectl apply -f infra/k8s/deployment.yaml
```

2. **创建 Secrets：**

```bash
# 使用 kubectl 创建 Secret（推荐）
kubectl create secret generic industrial-ai-secrets \
    --namespace=industrial-ai \
    --from-literal=jwt-secret=$(openssl rand -base64 32) \
    --from-literal=database-url='postgres://user:password@host:5432/db?sslmode=require' \
    --from-literal=redis-password='your-redis-password'

# 或使用 YAML 文件
kubectl apply -f infra/k8s/secrets.yaml
```

3. **部署应用：**

```bash
# 部署 Deployment 和 Service
kubectl apply -f infra/k8s/deployment.yaml
kubectl apply -f infra/k8s/deployment-health.yaml

# 配置 HPA
kubectl apply -f infra/k8s/hpa.yaml

# 配置 Ingress
kubectl apply -f infra/k8s/ingress-tls.yaml
```

### 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n industrial-ai

# 检查服务状态
kubectl get svc -n industrial-ai

# 检查 Ingress
kubectl get ingress -n industrial-ai

# 测试服务
kubectl port-forward svc/backend 8080:8080 -n industrial-ai
curl http://localhost:8080/health
```

---

## Docker 部署

### 使用 docker-compose

```bash
# 开发环境
docker-compose -f docker-compose.dev.yml up -d

# 生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod-ssl.yml up -d

# 带 SSL 的开发环境
docker-compose -f docker-compose.yml -f docker-compose.dev-ssl.yml up -d
```

### 环境配置文件

复制 `.env.example` 到 `.env` 并填入配置：

```bash
cp .env.example .env
# 编辑 .env 文件
vim .env
```

### 构建自定义镜像

```bash
# 使用优化的 Dockerfile
docker build -f docker/Dockerfile.backend \
    --build-arg VERSION=1.0.0 \
    --build-arg COMMIT_SHA=$(git rev-parse HEAD) \
    -t industrial-ai-backend:1.0.0 .
```

---

## 监控配置

### Prometheus 部署

```bash
# 部署 Prometheus
kubectl apply -f infra/prometheus.yml

# 部署告警规则
kubectl apply -f infra/prometheus/alert_rules.yml
```

### Grafana 部署

```bash
# 部署 Grafana
kubectl apply -f infra/grafana/

# 配置数据源
kubectl apply -f infra/grafana/datasources/datasources.yaml
```

---

## 相关文档

- [生产检查清单](./PRODUCTION_CHECKLIST.md)
- [安全文档](./security/)
- [运维手册](./runbook/)
- [高可用配置](./high-availability/)