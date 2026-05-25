# 🚀 Industrial AI Platform - 生产部署配置指南

> **生成日期**: 2026-05-25
> **用途**: 生产环境部署前的安全配置

---

## 🔐 必须配置的密钥/凭证

### 1. 本地开发环境 (.env)

| 配置项 | 当前值 | 需要配置 |
|--------|--------|---------|
| REDIS_PASSWORD | 空 | 配置强密码（≥16字符） |
| LLM_API_KEY | 空 | 配置百炼 API Key |
| ADMIN_PASSWORD | admin123 | 强密码（≥12字符，含特殊字符） |

### 2. Kubernetes Secrets (infra/k8s/secrets.yaml)

所有 `<base64-encoded-xxx>` 占位符需要替换：

```bash
# 生成 JWT Secret
openssl rand -base64 32 | base64

# 生成 Encryption Key
openssl rand -base64 32 | base64

# 生成 Redis Password
openssl rand -base64 24 | base64

# 编码 Database URL (启用 SSL)
echo -n "postgres://user:password@host:5432/db?sslmode=require" | base64

# 编码 API Key
echo -n "your-api-key" | base64
```

---

## 📋 生产部署检查清单

### ✅ 安全配置

- [ ] JWT_SECRET ≥ 32字符
- [ ] ADMIN_PASSWORD ≥ 12字符，含大小写+数字+特殊字符
- [ ] REDIS_PASSWORD 已配置
- [ ] DATABASE_SSL 已启用
- [ ] TLS 证书已配置
- [ ] CORS 配置禁止 `*`

### ✅ 网络安全

- [ ] HTTPS 已启用
- [ ] 证书有效期 ≥ 30天
- [ ] WAF 防护已启用（可选）

### ✅ 数据安全

- [ ] 数据库 SSL 连接
- [ ] 数据库密码强度验证
- [ ] Redis 密码认证
- [ ] Kubernetes Secret 配置

### ✅ 容器安全

- [ ] 非 root 用户运行
- [ ] 安全上下文配置
- [ ] 镜像漏洞扫描

---

## 🔧 快速修复步骤

### 步骤 1: 配置 .env 密钥

```bash
cd /path/to/industrial-ai-platform

# 编辑 .env 文件
vim .env

# 修改以下项：
REDIS_PASSWORD=your-redis-password-here
LLM_API_KEY=your-bailian-api-key
ADMIN_PASSWORD=StrongAdmin@2026!
```

### 步骤 2: 配置 Kubernetes Secrets

```bash
cd infra/k8s

# 生成密钥脚本
./scripts/generate-secrets.sh

# 或手动更新 secrets.yaml
vim secrets.yaml
```

### 步骤 3: 启用数据库 SSL

修改 DATABASE_URL：
```
postgres://user:password@host:5432/db?sslmode=require
```

---

## ⚠️ 安全警告

**以下配置必须修改，否则存在安全风险**：

| 配置项 | 风险等级 | 说明 |
|--------|---------|------|
| REDIS_PASSWORD=空 | 🔴 高危 | Redis 无密码暴露 |
| LLM_API_KEY=空 | 🔴 高危 | AI 功能无法使用 |
| ADMIN_PASSWORD=admin123 | 🟠 中危 | 弱密码易被破解 |
| DATABASE_SSL=disable | 🟠 中危 | 数据传输未加密 |

---

## 📝 备注

1. 本指南仅用于生产环境部署
2. 本地开发环境可保持当前配置
3. 所有密钥应妥善保管，避免泄露
4. 建议使用密钥管理服务（如 Vault）

---

**帅老大，请根据此指南配置生产环境密钥！**