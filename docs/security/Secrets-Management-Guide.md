# 密钥管理安全指南

> **Industrial AI Platform Secrets 管理最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 密钥管理概述

| 功能 | 描述 | 安全收益 |
|------|------|---------|
| **环境变量加密** | Secrets 不明文存储 | 防止泄露 |
| **Kubernetes Secrets** | 集中 Secrets 管理 | 安全存储 |
| **密钥轮换** | 定期更换密钥 | 减少风险窗口 |
| **访问控制** | RBAC 限制密钥访问 | 最小权限 |
| **审计日志** | 密钥操作记录 | 追踪异常 |
| **备份恢复** | 密钥备份策略 | 灾难恢复 |

---

## 🔐 需管理的密钥清单

### 应用密钥

| 密钥名称 | 用途 | 安全级别 | 轮换周期 |
|----------|------|---------|---------|
| `JWT_SECRET` | JWT Token 签名 | **Critical** | 90 天 |
| `DATABASE_URL` | 数据库连接 | **Critical** | 180 天 |
| `REDIS_PASSWORD` | Redis 认证 | **High** | 90 天 |
| `GLM_API_KEY` | GLM AI API | **High** | 手动 |
| `ENCRYPTION_KEY` | 数据加密 | **Critical** | 365 天 |
| `SMTP_PASSWORD` | 邮件发送 | **Medium** | 180 天 |

### 基础设施密钥

| 密钥名称 | 用途 | 安全级别 |
|----------|------|---------|
| `POSTGRES_PASSWORD` | PostgreSQL 管理员密码 | **Critical** |
| `GRAFANA_ADMIN_PASSWORD` | Grafana 管理员密码 | **High** |
| `PROMETHEUS_BASIC_AUTH` | Prometheus 认证 | **Medium** |
| `SSL_CERTIFICATE_KEY` | SSL 证书私钥 | **Critical** |

---

## 🛡️ 密钥存储方案

### 方案 A: Kubernetes Secrets (推荐)

```yaml
# infra/k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: industrial-ai-secrets
  namespace: industrial-ai
type: Opaque
data:
  # 所有值必须 base64 编码
  jwt-secret: <base64-encoded-value>
  database-url: <base64-encoded-value>
  redis-password: <base64-encoded-value>
  glm-api-key: <base64-encoded-value>
---
apiVersion: v1
kind: Secret
metadata:
  name: industrial-ai-db-credentials
  namespace: industrial-ai
type: Opaque
data:
  postgres-password: <base64-encoded-value>
  postgres-user: aW5kdXN0cmlhbF9hcHA=  # industrial_app
---
apiVersion: v1
kind: Secret
metadata:
  name: industrial-ai-tls-secret
  namespace: industrial-ai
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
```

### 创建 Secrets

```bash
# 从文件创建 Secret
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=database-url="postgres://industrial_app:password@postgres:5432/industrial_ai?sslmode=require" \
  --namespace=industrial-ai

# 从文件创建 TLS Secret
kubectl create secret tls industrial-ai-tls-secret \
  --cert=fullchain.pem \
  --key=privkey.pem \
  --namespace=industrial-ai

# 验证 Secret 创建
kubectl get secrets -n industrial-ai
kubectl describe secret industrial-ai-secrets -n industrial-ai
```

---

### 方案 B: HashiCorp Vault (企业级)

```bash
# Vault 配置示例
vault kv put secret/industrial-ai/jwt-secret value=$(openssl rand -base64 32)
vault kv put secret/industrial-ai/database-url value="postgres://..."
vault kv put secret/industrial-ai/redis-password value=$(openssl rand -base64 24)

# 读取密钥
vault kv get -field=value secret/industrial-ai/jwt-secret
```

**Vault 集成配置：**

```yaml
# docker-compose.yml (Vault 集成)
services:
  vault:
    image: hashicorp/vault:latest
    environment:
      - VAULT_DEV_ROOT_TOKEN_ID=root
    ports:
      - "8200:8200"
    
  backend:
    environment:
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
```

---

### 方案 C: Docker Secrets (Docker Swarm)

```bash
# 创建 Docker Secrets
docker secret create jwt_secret ./jwt_secret.txt
docker secret create database_url ./database_url.txt

# 服务使用 Secrets
docker service create \
  --name industrial-ai-backend \
  --secret jwt_secret \
  --secret database_url \
  industrial-ai/backend:latest
```

---

## 🔧 环境变量安全配置

### .env 文件安全规则

```bash
# ❌ 错误做法 - 硬编码密钥
JWT_SECRET=hardcoded-secret-key

# ✅ 正确做法 - 使用强随机密钥
JWT_SECRET=$(openssl rand -base64 32)

# ✅ 生产环境 - 不使用 .env 文件
# 使用 Secrets 管理系统 (K8s Secrets / Vault)
```

### .env.example 模板

```bash
# .env.example (不含真实密钥，仅示例)
JWT_SECRET=your-jwt-secret-min-32-characters
DATABASE_URL=postgres://user:password@host:5432/db?sslmode=require
REDIS_PASSWORD=your-redis-password
GLM_API_KEY=your-glm-api-key
ENCRYPTION_KEY=your-encryption-key-min-32-characters

# 提示: 生产环境请使用 Kubernetes Secrets 或 Vault
# 所有密钥应至少 32 字符随机生成
```

### 环境变量注入安全

```yaml
# Kubernetes Deployment (使用 Secrets)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      containers:
        - name: backend
          env:
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: jwt-secret
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: database-url
```

---

## 🔄 密钥轮换策略

### JWT Secret 轮换

```bash
# 生成新 JWT Secret
NEW_JWT_SECRET=$(openssl rand -base64 32)

# 更新 Kubernetes Secret
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret=$NEW_JWT_SECRET \
  --namespace=industrial-ai \
  --dry-run=client -o yaml | kubectl apply -f -

# 重启后端服务加载新密钥
kubectl rollout restart deployment/backend -n industrial-ai

# 注意: 所有现有 JWT Token 将失效，用户需重新登录
```

### 密钥轮换时间表

| 密钥类型 | 轮换周期 | 影响范围 | 轮换步骤 |
|----------|---------|---------|---------|
| **JWT Secret** | 90 天 | 所有用户 Token 失效 | 1. 生成新密钥 2. 更新 Secret 3. 重启服务 4. 通知用户 |
| **Database Password** | 180 天 | 数据库连接中断 | 1. 创建新用户 2. 更新连接字符串 3. 删除旧用户 |
| **Redis Password** | 90 天 | 缓存连接中断 | 1. 生成新密码 2. 更新配置 3. 重启 Redis |
| **API Key** | 手动 | API 调用失败 | 1. 申请新 Key 2. 更新配置 3. 测试验证 |

---

## 📝 密钥生成最佳实践

### 强密钥生成命令

```bash
# 32 字节 (256 位) 随机密钥
openssl rand -base64 32

# 64 字节 (512 位) 超强密钥
openssl rand -base64 64

# 十六进制格式
openssl rand -hex 32

# UUID 格式
uuidgen | tr -d '-'
```

### 密钥强度要求

| 密钥类型 | 最小长度 | 格式 |
|----------|---------|------|
| **JWT Secret** | 32 字符 | Base64 随机 |
| **Database Password** | 24 字符 | 随机字母数字 + 特殊字符 |
| **Encryption Key** | 32 字节 | AES-256 |
| **API Key** | 32 字符 | 唯一标识 |

---

## 🛡️ 密钥访问控制

### Kubernetes RBAC 配置

```yaml
# 仅允许特定 ServiceAccount 读取 Secrets
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secrets-reader
  namespace: industrial-ai
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["industrial-ai-secrets"]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: backend-secrets-reader
  namespace: industrial-ai
subjects:
  - kind: ServiceAccount
    name: backend-sa
roleRef:
  kind: Role
  name: secrets-reader
  apiGroup: rbac.authorization.k8s.io
```

---

## ✅ 安全检查清单

| 检查项 | 要求 | 状态 |
|--------|------|------|
| **密钥不硬编码** | 不在代码中硬编码 | ✅ 已实现 |
| **.env 不提交** | .env 文件不提交 Git | ⏳ 待配置 |
| **K8s Secrets 使用** | 生产环境使用 Secrets | ⏳ 待配置 |
| **密钥长度** | ≥ 32 字符随机 | ⏳ 待配置 |
| **密钥轮换** | 定期轮换机制 | ⏳ 待配置 |
| **访问控制** | RBAC 限制密钥访问 | ⏳ 待配置 |
| **审计日志** | 密钥操作记录 | ⏳ 待配置 |
| **备份恢复** | 密钥备份策略 | ⏳ 待配置 |

---

## 🔧 配置示例

### 生产环境部署

```yaml
# infra/k8s/deployment-with-secrets.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: industrial-ai
spec:
  replicas: 2
  template:
    spec:
      serviceAccountName: backend-sa
      containers:
        - name: backend
          image: industrial-ai/backend:v1.0.0
          env:
            # 从 Secrets 加载密钥
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: jwt-secret
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: database-url
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: redis-password
            - name: GLM_API_KEY
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: glm-api-key
          volumeMounts:
            - name: tls-certs
              mountPath: /etc/ssl/industrial-ai
              readOnly: true
      volumes:
        - name: tls-certs
          secret:
            secretName: industrial-ai-tls-secret
```

---

**最后更新**: 2026-05-13  
**审核人**: 安全团队