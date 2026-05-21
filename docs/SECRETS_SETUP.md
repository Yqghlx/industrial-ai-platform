# 🔐 密钥配置清单

> **生成时间**: 2026-05-19 13:46
> **状态**: ⚠️ 需要手动配置
> **重要**: 不要将此文件提交到 Git！

---

## 📋 GitHub Secrets 配置清单

### ✅ 已自动提供
- `GITHUB_TOKEN` - GitHub Actions 自动提供

### ⚠️ 必须配置 (P0)

| Secret 名称 | 用途 | 状态 |
|-------------|------|------|
| `DEPLOY_TOKEN` | Docker 部署 Token | ⏳ 待配置 |
| `KUBE_CONFIG_STAGING` | Kubernetes staging 配置 | ⏳ 待配置 |
| `KUBE_CONFIG_PRODUCTION` | Kubernetes production 配置 | ⏳ 待配置 |

### 📝 可选配置 (P1)

| Secret 名称 | 用途 | 状态 |
|-------------|------|------|
| `CODECOV_TOKEN` | Codecov 覆盖率报告 | ⏳ 待配置 |
| `SLACK_BOT_TOKEN` | Slack 通知 | ⏳ 待配置 |

---

## 🛠️ 配置步骤

### Step 1: 创建 GitHub Personal Access Token

1. 前往: https://github.com/settings/tokens
2. 点击 "Generate new token (classic)"
3. 设置:
   - Token name: `Industrial AI Deploy Token`
   - Expiration: `No expiration` 或 `1 year`
   - Permissions: `repo`, `write:packages`, `read:packages`
4. 生成并保存 Token

### Step 2: 配置 GitHub Secrets

前往: https://github.com/Yqghlx/industrial-ai-platform/settings/secrets/actions

点击 "New repository secret"，添加:

| Secret 名称 | 值 | 说明 |
|-------------|-----|------|
| `DEPLOY_TOKEN` | <刚才生成的 Token> | GitHub PAT |

### Step 3: 配置 Kubernetes Secrets (可选)

如果有 Kubernetes 集群:

1. 获取 kubeconfig 文件:
   ```bash
   cat ~/.kube/config | base64
   ```

2. 将 base64 编码的内容添加到 GitHub Secrets:
   - `KUBE_CONFIG_STAGING` - staging 环境配置
   - `KUBE_CONFIG_PRODUCTION` - production 环境配置

---

## 🚀 快速生成应用密钥

运行密钥生成脚本:

```bash
cd ~/Projects/industrial-ai-platform
./scripts/generate-secrets.sh
```

脚本会自动:
- 生成所有安全密钥 (JWT, Encryption, Redis, DB)
- 保存到 `.secrets.tmp` (不提交到 Git)
- 提供 GitHub Secrets 配置指引
- 提供 Kubernetes Secrets 配置命令

---

## ⚠️ 安全提醒

1. **不要提交密钥到 Git！**
   - `.secrets.tmp` 文件已在 `.gitignore` 中
   - `secrets.yaml` 不要填入真实密钥后提交

2. **保存密钥到安全地方**
   - 使用密码管理器 (1Password, Bitwarden)
   - 或保存到本地加密文件

3. **定期更换密钥**
   - JWT_SECRET 建议每 90 天更换
   - 数据库密码建议每 180 天更换

---

## 📝 配置完成验证

配置完成后，执行以下命令验证:

```bash
# 检查 GitHub Secrets 是否配置 (需要先 gh auth login)
gh secret list

# 检查 Kubernetes Secrets 是否创建
kubectl get secrets -n industrial-ai

# 检查 .env.production 是否配置
grep -E "JWT_SECRET|DB_PASSWORD" .env.production
```

---

## 🚀 配置完成后

密钥配置完成后，可以:
1. 运行 Build and Deploy workflow
2. 部署到生产环境
3. 测试生产环境连接

---

**配置状态**: ⏳ 等待用户完成
**下一步**: 用户生成密钥后，继续部署流程