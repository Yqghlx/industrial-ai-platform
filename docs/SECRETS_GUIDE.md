# 密钥配置指南 (Secrets Configuration Guide)

本文档详细说明 Industrial AI Platform 生产环境所需的所有密钥配置，包括生成方法、配置步骤和安全最佳实践。

---

## 目录

1. [密钥清单](#密钥清单)
2. [密钥生成命令](#密钥生成命令)
3. [配置步骤](#配置步骤)
4. [配置验证](#配置验证)
5. [密钥轮换](#密钥轮换)
6. [安全最佳实践](#安全最佳实践)
7. [故障排除](#故障排除)

---

## 密钥清单

### 必需密钥

| 密钥名称 | 环境变量 | 用途 | 最小长度 | 轮换周期 |
|---------|---------|------|---------|---------|
| JWT Secret | `JWT_SECRET` | JWT 令牌签名密钥 | 32 字符 | 90 天 |
| Database URL | `DATABASE_URL` | 数据库连接字符串 | - | 按需 |
| Database Password | `POSTGRES_PASSWORD` / `DB_PASSWORD` | 数据库用户密码 | 16 字符 | 180 天 |
| Redis Password | `REDIS_PASSWORD` | Redis 认证密码 | 16 字符 | 90 天 |
| Admin Password | `ADMIN_PASSWORD` | 系统管理员密码 | 16 字符 | 90 天 |

### 可选密钥

| 密钥名称 | 环境变量 | 用途 | 最小长度 |
|---------|---------|------|---------|
| Encryption Key | `ENCRYPTION_KEY` | 数据加密密钥 | 32 字符 |
| Grafana Password | `GRAFANA_PASSWORD` | Grafana 管理员密码 | 12 字符 |
| GLM API Key | `GLM_API_KEY` | 智谱 AI API 密钥 | - |

---

## 密钥生成命令

### 1. JWT Secret (JWT 签名密钥)

```bash
# 生成 256 位 (32 字节) Base64 编码密钥
openssl rand -base64 32

# 或使用更长的密钥 (推荐)
openssl rand -base64 48

# 示例输出:
# xK9vM2nR5tW8yB1cF4gH7jK0lN3qS6uV9zA2bD5eG8hJ1kM4nO7pQ0rS3tV6wY9z
```

### 2. Database Password (数据库密码)

```bash
# 生成 24 字符密码 (移除可能引起问题的特殊字符)
openssl rand -base64 24 | tr -d '/+=' | head -c 24

# 或使用 openssl rand -hex
openssl rand -hex 16

# 示例输出:
# a1b2c3d4e5f6g7h8i9j0k1l2
```

### 3. Database URL (数据库连接字符串)

```bash
# 构建 PostgreSQL 连接字符串
# 格式: postgres://用户名:密码@主机:端口/数据库名?sslmode=require

# 手动构建
DATABASE_URL="postgres://industrial_user:YOUR_PASSWORD@postgres-host:5432/industrial_ai?sslmode=require"

# 使用脚本生成
./scripts/generate-secrets.sh
```

### 4. Redis Password (Redis 密码)

```bash
# 生成 24 字符密码
openssl rand -base64 24 | tr -d '/+=' | head -c 24

# 或
openssl rand -hex 16
```

### 5. Admin Password (管理员密码)

```bash
# 生成强密码
openssl rand -base64 18 | tr -d '/+=' | head -c 18

# 或使用 memorable 密码生成器
# brew install pwgen  # macOS
pwgen -s 18 1
```

### 6. Encryption Key (加密密钥)

```bash
# 生成 AES-256 加密密钥
openssl rand -base64 32
```

### 一键生成所有密钥

```bash
# 使用项目提供的脚本生成所有密钥
./scripts/generate-secrets.sh

# 脚本将:
# 1. 生成所有必需密钥
# 2. 保存到 ./secrets/ 目录
# 3. 生成 Kubernetes Secrets YAML
# 4. 设置正确的文件权限 (400)
```

---

## 配置步骤

### 步骤 1: 准备环境变量文件

```bash
# 复制示例配置
cp .env.example .env

# 编辑配置文件
vim .env  # 或使用您喜欢的编辑器
```

### 步骤 2: 替换占位符

打开 `.env` 文件，替换以下占位符：

```bash
# ❌ 替换前 (示例)
JWT_SECRET=REPLACE_WITH_YOUR_SECRET
DATABASE_URL=postgres://industrial_user:REPLACE_PASSWORD@localhost:5432/industrial_ai?sslmode=disable
ADMIN_PASSWORD=REPLACE_WITH_YOUR_PASSWORD

# ✅ 替换后 (示例)
JWT_SECRET=YOUR_GENERATED_SECRET_HERE
DATABASE_URL=postgres://industrial_user:YOUR_PASSWORD_HERE@postgres:5432/industrial_ai?sslmode=require
ADMIN_PASSWORD=YOUR_ADMIN_PASSWORD_HERE
```

### 步骤 3: Docker Compose 配置

对于 Docker Compose 部署，可以：

**方式 A: 使用环境变量文件**

```yaml
# docker-compose.yml
services:
  backend:
    env_file:
      - .env
```

**方式 B: 直接配置环境变量**

```yaml
# docker-compose.prod.yml
services:
  backend:
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
```

启动时传入变量：

```bash
# 从 .env 文件加载
export $(cat .env | xargs) && docker-compose up -d

# 或直接传入
JWT_SECRET="your-secret" DATABASE_URL="your-url" docker-compose up -d
```

### 步骤 4: Kubernetes Secrets 配置

#### 方式 A: 使用 kubectl 创建

```bash
# 创建命名空间
kubectl create namespace industrial-ai

# 创建 Secrets
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=database-url="postgres://industrial_user:password@postgres:5432/industrial_ai?sslmode=require" \
  --from-literal=redis-password=$(openssl rand -base64 24) \
  --from-literal=admin-password=$(openssl rand -base64 18) \
  -n industrial-ai

# 验证创建成功
kubectl get secrets -n industrial-ai
kubectl describe secret industrial-ai-secrets -n industrial-ai
```

#### 方式 B: 使用 YAML 文件

```yaml
# infra/k8s/secrets.yaml (Base64 编码)
apiVersion: v1
kind: Secret
metadata:
  name: industrial-ai-secrets
  namespace: industrial-ai
type: Opaque
data:
  # 使用 base64 编码的值
  # echo -n "REPLACE_WITH_SECRET" | base64
  jwt-secret: BASE64_ENCODED_PLACEHOLDER
  database-url: BASE64_ENCODED_PLACEHOLDER
  redis-password: BASE64_ENCODED_PLACEHOLDER
  admin-password: BASE64_ENCODED_PLACEHOLDER
```

应用配置：

```bash
# 编码密钥
echo -n "your-actual-secret-value" | base64

# 应用 YAML
kubectl apply -f infra/k8s/secrets.yaml

# 验证
kubectl get secret industrial-ai-secrets -n industrial-ai -o yaml
```

#### 方式 C: 使用生成脚本

```bash
# 运行密钥生成脚本
./scripts/generate-secrets.sh

# 脚本会生成:
# - ./secrets/jwt_secret.txt
# - ./secrets/database_url.txt
# - ./secrets/redis_password.txt
# - ./secrets/admin_password.txt
# - ./secrets/k8s-secrets.yaml (可选)

# 应用生成的 K8s Secrets
kubectl apply -f ./secrets/k8s-secrets.yaml
```

### 步骤 5: 在 Deployment 中引用 Secrets

```yaml
# infra/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: industrial-ai-backend
  namespace: industrial-ai
spec:
  template:
    spec:
      containers:
        - name: backend
          image: industrial-ai-backend:latest
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
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: redis-password
            - name: ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: admin-password
```

---

## 配置验证

### 使用验证脚本

```bash
# 运行生产配置验证脚本
./scripts/verify-production-config.sh

# 指定环境变量文件
./scripts/verify-production-config.sh --env-file .env.production

# 启用 Kubernetes 检查
./scripts/verify-production-config.sh --k8s
```

### 手动验证

#### 1. 验证 JWT_SECRET

```bash
# 检查 JWT_SECRET 是否设置
echo $JWT_SECRET

# 检查长度
echo -n "$JWT_SECRET" | wc -c
# 输出应 >= 32

# 检查是否为弱密钥
echo "$JWT_SECRET" | grep -iE "your|change|secret|password|test|default|example"
# 应无输出
```

#### 2. 验证 DATABASE_URL

```bash
# 检查 SSL 配置
echo $DATABASE_URL | grep "sslmode=require"
# 应有输出

# 禁止的配置
echo $DATABASE_URL | grep "sslmode=disable"
# 应无输出
```

#### 3. 验证 CORS_ORIGINS

```bash
# 检查是否使用通配符
echo $CORS_ORIGINS | grep "\*"
# 应无输出 (生产环境禁止)
```

#### 4. 验证 Kubernetes Secrets

```bash
# 列出所有 Secrets
kubectl get secrets -n industrial-ai

# 查看 Secret 详情
kubectl describe secret industrial-ai-secrets -n industrial-ai

# 获取并解码 Secret 值
kubectl get secret industrial-ai-secrets -n industrial-ai \
  -o jsonpath='{.data.jwt-secret}' | base64 -d

# 检查 JWT Secret 长度
kubectl get secret industrial-ai-secrets -n industrial-ai \
  -o jsonpath='{.data.jwt-secret}' | base64 -d | wc -c
# 应 >= 32
```

#### 5. 验证应用启动

```bash
# Docker Compose
docker-compose logs backend | grep -i "jwt\|database\|error"

# Kubernetes
kubectl logs -n industrial-ai deployment/industrial-ai-backend | grep -i "jwt\|database\|error"
```

---

## 密钥轮换

### JWT Secret 轮换

```bash
# 使用轮换脚本
./scripts/rotate-jwt-secret.sh

# 或手动轮换
# 1. 生成新密钥
NEW_JWT_SECRET=$(openssl rand -base64 32)

# 2. 更新 K8s Secret
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret="$NEW_JWT_SECRET" \
  -n industrial-ai --dry-run=client -o yaml | kubectl apply -f -

# 3. 重启应用
kubectl rollout restart deployment/industrial-ai-backend -n industrial-ai

# 4. 验证
kubectl rollout status deployment/industrial-ai-backend -n industrial-ai
```

### Database Password 轮换

```bash
# 1. 生成新密码
NEW_DB_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)

# 2. 更新数据库密码
kubectl exec -it -n industrial-ai deployment/postgres -- \
  psql -U postgres -c "ALTER USER industrial_user PASSWORD '$NEW_DB_PASSWORD';"

# 3. 更新 Secret
NEW_DB_URL="postgres://industrial_user:$NEW_DB_PASSWORD@postgres:5432/industrial_ai?sslmode=require"
kubectl create secret generic industrial-ai-secrets \
  --from-literal=database-url="$NEW_DB_URL" \
  -n industrial-ai --dry-run=client -o yaml | kubectl apply -f -

# 4. 重启应用
kubectl rollout restart deployment/industrial-ai-backend -n industrial-ai
```

### Redis Password 轮换

```bash
# 1. 生成新密码
NEW_REDIS_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)

# 2. 更新 Redis 配置
kubectl exec -it -n industrial-ai deployment/redis -- \
  redis-cli CONFIG SET requirepass "$NEW_REDIS_PASSWORD"

# 3. 更新 Secret
kubectl create secret generic industrial-ai-secrets \
  --from-literal=redis-password="$NEW_REDIS_PASSWORD" \
  -n industrial-ai --dry-run=client -o yaml | kubectl apply -f -

# 4. 重启应用
kubectl rollout restart deployment/industrial-ai-backend -n industrial-ai
```

---

## 安全最佳实践

### 1. 密钥存储

- ✅ **推荐**: 使用 Kubernetes Secrets、HashiCorp Vault 或云服务密钥管理
- ✅ **开发环境**: 使用 `.env` 文件，并确保在 `.gitignore` 中
- ❌ **禁止**: 将密钥提交到 Git 仓库
- ❌ **禁止**: 在日志或错误消息中输出密钥

### 2. 密钥强度

| 密钥类型 | 最小长度 | 推荐长度 | 格式 |
|---------|---------|---------|------|
| JWT Secret | 32 字符 | 48+ 字符 | Base64 编码随机字节 |
| Database Password | 16 字符 | 24+ 字符 | 混合字符 |
| Redis Password | 16 字符 | 24+ 字符 | 混合字符 |
| Admin Password | 16 字符 | 20+ 字符 | 混合大小写+数字+符号 |

### 3. 密钥轮换

- **JWT Secret**: 每 90 天轮换一次
- **Database Password**: 每 180 天轮换一次
- **Redis Password**: 每 90 天轮换一次
- **Admin Password**: 每 90 天轮换一次
- **密钥泄露后立即轮换**

### 4. 访问控制

```yaml
# Kubernetes RBAC 示例
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: industrial-ai
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["industrial-ai-secrets"]
    verbs: ["get"]
```

### 5. 审计日志

```bash
# 启用 Kubernetes 审计日志
# 在 kube-apiserver 配置中添加:
--audit-log-path=/var/log/kubernetes/audit.log
--audit-log-maxage=30
--audit-policy-file=/etc/kubernetes/audit-policy.yaml

# 查看 Secret 访问日志
kubectl logs -n kube-system -l component=kube-apiserver | grep "secrets"
```

### 6. 加密存储

```yaml
# Kubernetes Secrets 加密配置
# encryption-config.yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: <base64-encoded-32-byte-key>
      - identity: {}
```

### 7. 网络安全

```yaml
# 限制数据库访问
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: database-policy
  namespace: industrial-ai
spec:
  podSelector:
    matchLabels:
      app: postgres
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: backend
      ports:
        - protocol: TCP
          port: 5432
```

### 8. 环境隔离

```bash
# 为不同环境使用不同密钥
# 开发环境
kubectl create secret generic industrial-ai-secrets \
  -n industrial-ai-dev \
  --from-literal=jwt-secret=$(openssl rand -base64 32)

# 生产环境
kubectl create secret generic industrial-ai-secrets \
  -n industrial-ai-prod \
  --from-literal=jwt-secret=$(openssl rand -base64 48)  # 更强的密钥
```

---

## 故障排除

### 常见问题

#### 1. JWT_SECRET 未配置或长度不足

**症状**: 应用启动失败，日志显示 `JWT_SECRET is required` 或 `JWT_SECRET must be at least 32 characters`

**解决方案**:
```bash
# 检查环境变量
echo $JWT_SECRET
echo -n "$JWT_SECRET" | wc -c

# 设置正确的 JWT_SECRET
export JWT_SECRET=$(openssl rand -base64 32)

# 或更新 .env 文件
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
```

#### 2. 数据库连接失败

**症状**: `connection refused` 或 `SSL required`

**解决方案**:
```bash
# 确保 DATABASE_URL 包含 sslmode=require
echo $DATABASE_URL | grep "sslmode=require"

# 如果缺少，添加 SSL 参数
export DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require"
```

#### 3. CORS 错误

**症状**: 前端请求被阻止，控制台显示 CORS 错误

**解决方案**:
```bash
# 确保 CORS_ORIGINS 不使用通配符
echo $CORS_ORIGINS
# 应为: https://domain1.com,https://domain2.com

# 设置正确的 CORS 配置
export CORS_ORIGINS="https://your-domain.com,https://app.your-domain.com"
```

#### 4. Kubernetes Secret 不存在

**症状**: Pod 启动失败，显示 `secret "industrial-ai-secrets" not found`

**解决方案**:
```bash
# 检查 Secret 是否存在
kubectl get secrets -n industrial-ai

# 创建 Secret
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=database-url="postgres://..." \
  -n industrial-ai
```

#### 5. Secret 值为占位符

**症状**: 应用使用默认值或启动失败

**解决方案**:
```bash
# 检查 Secret 值
kubectl get secret industrial-ai-secrets -n industrial-ai \
  -o jsonpath='{.data.jwt-secret}' | base64 -d

# 如果值为占位符，更新为真实值
kubectl create secret generic industrial-ai-secrets \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  -n industrial-ai --dry-run=client -o yaml | kubectl apply -f -
```

### 调试命令

```bash
# 查看所有环境变量
kubectl exec -it deployment/backend -n industrial-ai -- env | grep -E "JWT|DATABASE|REDIS"

# 查看 Secret 内容 (Base64 解码)
kubectl get secret industrial-ai-secrets -n industrial-ai -o yaml

# 查看应用日志
kubectl logs -f deployment/backend -n industrial-ai

# 进入容器调试
kubectl exec -it deployment/backend -n industrial-ai -- sh
```

---

## 检查清单

部署前请确认以下项目：

- [ ] JWT_SECRET 已配置且长度 >= 32 字符
- [ ] DATABASE_URL 包含 `sslmode=require`
- [ ] CORS_ORIGINS 不包含通配符 `*`
- [ ] 所有密码不包含弱密钥模式
- [ ] .env 文件已添加到 .gitignore
- [ ] Kubernetes Secrets 已创建且使用真实值
- [ ] TLS 证书已配置且未过期
- [ ] 容器以非 root 用户运行
- [ ] 管理员密码已修改且强度足够
- [ ] 运行验证脚本 `./scripts/verify-production-config.sh` 通过

---

## 参考链接

- [Kubernetes Secrets 文档](https://kubernetes.io/docs/concepts/configuration/secret/)
- [HashiCorp Vault](https://www.vaultproject.io/)
- [OWASP 密码存储备忘录](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)

---

*最后更新: 2024*
*维护者: Industrial AI Platform Team*