# Industrial AI Platform - 生产部署指南

> **目标**: 将 Industrial AI Platform 部署到生产环境
> **前提条件**: Kubernetes 集群已就绪

---

## 📋 部署前检查清单

### 🔴 必须完成 (P0)

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 测试全部通过 | ✅ | 所有单元测试已通过 |
| CI/CD 正常 | ✅ | Security/Code Quality/Docker Build 全部成功 |
| 密钥配置 | ⏳ | 需要配置生产密钥 |
| 环境变量 | ⏳ | 需要配置生产环境变量 |

---

## 🔐 密钥配置指南

### 1. 生成生产密钥

在终端执行以下命令生成安全密钥：

```bash
# JWT Secret (256位)
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET: $JWT_SECRET"

# Encryption Key (AES-256)
ENCRYPTION_KEY=$(openssl rand -base64 32)
echo "ENCRYPTION_KEY: $ENCRYPTION_KEY"

# Redis Password
REDIS_PASSWORD=$(openssl rand -base64 24)
echo "REDIS_PASSWORD: $REDIS_PASSWORD"

# Admin Password (自定义强密码)
ADMIN_PASSWORD="YourStrongAdminPassword123!"
echo "ADMIN_PASSWORD: $ADMIN_PASSWORD"

# Database Password
DB_PASSWORD=$(openssl rand -base64 24)
echo "DB_PASSWORD: $DB_PASSWORD"
```

### 2. 配置 Kubernetes Secrets

将生成的密钥编码为 base64：

```bash
# 编码密钥
JWT_SECRET_B64=$(echo -n "$JWT_SECRET" | base64)
ENCRYPTION_KEY_B64=$(echo -n "$ENCRYPTION_KEY" | base64)
REDIS_PASSWORD_B64=$(echo -n "$REDIS_PASSWORD" | base64)
DB_PASSWORD_B64=$(echo -n "$DB_PASSWORD" | base64)

# 数据库连接字符串
DATABASE_URL="postgres://postgres:${DB_PASSWORD}@postgres:5432/industrial_ai?sslmode=require"
DATABASE_URL_B64=$(echo -n "$DATABASE_URL" | base64)
```

### 3. 更新 secrets.yaml

编辑 `infra/k8s/secrets.yaml`，替换占位符：

```yaml
data:
  jwt-secret: <填入 JWT_SECRET_B64>
  database-url: <填入 DATABASE_URL_B64>
  redis-password: <填入 REDIS_PASSWORD_B64>
  encryption-key: <填入 ENCRYPTION_KEY_B64>
```

**⚠️ 重要**: 不要将真实密钥提交到 Git！

### 4. 或使用 kubectl 直接创建

```bash
kubectl create secret generic industrial-ai-secrets \
  --namespace=industrial-ai \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=database-url="$DATABASE_URL" \
  --from-literal=redis-password="$REDIS_PASSWORD" \
  --from-literal=encryption-key="$ENCRYPTION_KEY"
```

---

## 🌐 环境变量配置

### 1. 配置 .env.production

编辑 `.env.production`：

```bash
# 复制模板
cp .env.production .env

# 编辑配置
vim .env
```

**必须修改的项：**

| 变量 | 当前值 | 需要改为 |
|------|--------|----------|
| `JWT_SECRET` | `your-super-secret...` | 真实的 JWT 密钥 |
| `POSTGRES_PASSWORD` | `your-secure...` | 真实的数据库密码 |
| `ADMIN_PASSWORD` | `Admin@123456` | 强密码 |
| `LLM_API_KEY` | `your-llm-api-key...` | 你的百炼/GLM API Key |
| `CORS_ORIGINS` | `localhost` | 你的生产域名 |

### 2. 配置 CORS

生产环境的 CORS 必须限制为你的域名：

```bash
# 示例：允许你的生产域名
CORS_ORIGINS=https://your-domain.com,https://api.your-domain.com

# ⚠️ 不要使用 * (全部允许)
CORS_ORIGINS=*  # 危险！禁止使用
```

---

## 🚀 部署步骤

### Step 1: 创建 namespace

```bash
kubectl create namespace industrial-ai
```

### Step 2: 应用 secrets

```bash
# 方式 1: 使用 secrets.yaml (需要先填入密钥)
kubectl apply -f infra/k8s/secrets.yaml

# 方式 2: 使用 kubectl create secret (推荐)
kubectl create secret generic industrial-ai-secrets ...
```

### Step 3: 应用其他配置

```bash
kubectl apply -f infra/k8s/
```

### Step 4: 验证部署

```bash
# 检查 pods
kubectl get pods -n industrial-ai

# 检查 services
kubectl get services -n industrial-ai

# 检查健康状态
kubectl exec -it deployment/backend -n industrial-ai -- curl localhost:8080/health
```

---

## 🔒 TLS/HTTPS 配置

### 方式 1: Let's Encrypt (推荐)

```bash
# 安装 cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

# 创建证书
kubectl apply -f infra/k8s/certificate.yaml
```

### 方式 2: 企业证书

```bash
# 编码证书
TLS_CRT_B64=$(cat your-certificate.crt | base64 -w0)
TLS_KEY_B64=$(cat your-private-key.key | base64 -w0)

# 更新 secrets.yaml
# tls.crt: <填入 TLS_CRT_B64>
# tls.key: <填入 TLS_KEY_B64>
```

---

## 📊 监控配置

### Prometheus + Grafana

```bash
# 部署监控栈
kubectl apply -f infra/k8s/monitoring/

# 访问 Grafana
kubectl port-forward svc/grafana -n industrial-ai 3000:80

# 默认账号
# Username: admin
# Password: <在 secrets 中配置>
```

---

## ✅ 部署后验证

### 1. 健康检查

```bash
# API 健康检查
curl https://your-domain.com/health

# 预期响应
{"status":"healthy","version":"1.0.0"}
```

### 2. 功能测试

```bash
# 登录测试
curl -X POST https://your-domain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<你的密码>"}'

# 预期响应
{"token":"eyJhbG...","user":{"id":1,"username":"admin"}}
```

### 3. 安全检查

| 检查项 | 命令 |
|--------|------|
| TLS 有效 | `curl -v https://your-domain.com` |
| CORS 配置 | `curl -H "Origin: https://evil.com" ...` |
| JWT 有效 | `curl -H "Authorization: Bearer <token>" ...` |

---

## 🔄 更新部署

### 滚动更新

```bash
# 更新镜像
kubectl set image deployment/backend backend=ghcr.io/yqghlx/industrial-ai-platform/backend:main -n industrial-ai

# 监控更新
kubectl rollout status deployment/backend -n industrial-ai
```

### 回滚

```bash
kubectl rollout undo deployment/backend -n industrial-ai
```

---

## 📝 部署记录模板

```markdown
## 部署记录 - YYYY-MM-DD

### 配置信息
- JWT_SECRET: 已配置 ✅
- Database URL: 已配置 ✅
- Redis Password: 已配置 ✅
- CORS Origins: https://your-domain.com ✅
- TLS Certificate: Let's Encrypt ✅

### 验证结果
- 健康检查: ✅
- 登录测试: ✅
- API 测试: ✅

### 部署人员
- 操作人: [姓名]
- 时间: YYYY-MM-DD HH:MM
```

---

## ⚠️ 安全注意事项

1. **不要提交密钥到 Git**
   - secrets.yaml 中的占位符不要替换为真实值
   - 使用 kubectl create secret 或 sealed-secrets

2. **定期轮换密钥**
   - JWT Secret: 每 3 个月
   - Database Password: 每 6 个月
   - Encryption Key: 每 3 个月

3. **备份数据库**
   ```bash
   kubectl exec deployment/postgres -n industrial-ai -- pg_dump industrial_ai > backup.sql
   ```

4. **监控告警**
   - 配置 Prometheus 告警规则
   - 配置 Grafana 通知渠道

---

## 🆘 常见问题

### Q: Pods 无法启动
```bash
# 检查日志
kubectl logs deployment/backend -n industrial-ai

# 常见原因
# 1. Secrets 未配置
# 2. Database 连接失败
# 3. 资源不足
```

### Q: API 返回 500
```bash
# 检查健康状态
kubectl exec deployment/backend -n industrial-ai -- curl localhost:8080/health

# 检查日志
kubectl logs -f deployment/backend -n industrial-ai
```

### Q: TLS 证书无效
```bash
# 检查证书
kubectl get certificate -n industrial-ai

# 更新证书
kubectl delete secret industrial-ai-tls-secret -n industrial-ai
kubectl apply -f infra/k8s/certificate.yaml
```

---

**部署完成后，请填写部署记录并通知相关人员！**