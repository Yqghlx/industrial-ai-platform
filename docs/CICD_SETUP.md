# Industrial AI Platform - CI/CD 配置指南

> 方案：GitHub Actions + Gitee Mirror

---

## 📋 架构说明

```
Gitee (主仓库) ──自动同步──> GitHub (镜像仓库)
                                │
                                ▼
                        GitHub Actions CI/CD
                                │
                    ┌───────────┴───────────┐
                    │                       │
              测试 & 构建            Docker 镜像推送
                    │                       │
                    ▼                       ▼
              Coverage 报告           GHCR (镜像仓库)
```

---

## 🚀 快速配置步骤

### Step 1: 创建 GitHub 镜像仓库

1. 登录 GitHub，创建新仓库：
   - 名称：`industrial-ai-platform`（建议同名）
   - 类型：Private（推荐）或 Public
   - 不要初始化 README/gitignore（避免冲突）

2. 获取仓库地址：
   ```
   https://github.com/YOUR_USERNAME/industrial-ai-platform.git
   ```

### Step 2: 推送代码到 GitHub

```bash
# 在 Gitee 仓库目录执行
cd ~/Projects/industrial-ai-platform

# 添加 GitHub 远程仓库
git remote add github https://github.com/YOUR_USERNAME/industrial-ai-platform.git

# 首次推送（强制推送所有分支）
git push github main --force

# 后续推送
git push github main
```

### Step 3: 配置 GitHub Secrets

进入 GitHub 仓库 → Settings → Secrets and variables → Actions

**必需 Secrets：**

| Secret | 说明 | 获取方式 |
|--------|------|----------|
| `GITHUB_TOKEN` | GitHub Token | 自动提供（无需配置） |
| `CODECOV_TOKEN` | Coverage 上传 Token | https://codecov.io 注册获取 |

**可选 Secrets：**

| Secret | 说明 | 用途 |
|--------|------|------|
| `DOCKER_USERNAME` | Docker Hub 用户名 | 推送到 Docker Hub |
| `DOCKER_PASSWORD` | Docker Hub 密码 | 推送到 Docker Hub |

### Step 4: 启用 GitHub Actions

1. 进入 GitHub 仓库 → Actions
2. 如果看到提示，点击 "I understand my workflows, go ahead and enable them"
3. workflows 自动识别并运行

---

## 🔄 同步机制

### 自动同步（推荐）

GitHub Actions 每 6 小时自动检查 Gitee 更新：

```yaml
schedule:
  - cron: '0 */6 * * *'  # 每6小时同步一次
```

### 手动同步

**方式 A：GitHub Actions 手动触发**
- 进入 Actions → Gitee Sync → Run workflow

**方式 B：本地推送**
```bash
# 从 Gitee 推送最新代码
git push github main
```

**方式 C：Gitee Webhook（高级）**
- Gitee 仓库 → 管理 → Webhooks
- 添加 webhook 触发 GitHub Actions
- 需要配置 GitHub Personal Access Token

---

## 📊 CI/CD 流程详解

### 1. Backend Test Job

```yaml
backend-test:
  - Setup Go 1.22
  - go mod download
  - go test -v -short -race -coverprofile=coverage.out
  - Upload to Codecov
  - Coverage threshold check (70%)
```

**触发条件：** push/PR 到 main 分支

### 2. Frontend Test Job

```yaml
frontend-test:
  - Setup Node.js 20
  - npm ci
  - npm run lint
  - tsc --noEmit (类型检查)
  - npm run build
  - E2E tests (可选)
  - Upload dist artifacts
```

### 3. Docker Build Job

```yaml
docker-build:
  - Setup Docker Buildx
  - Login to ghcr.io
  - Build backend image
  - Build frontend image
  - Push to GitHub Container Registry
```

**触发条件：** push 到 main（测试通过后）

**镜像地址：**
- Backend: `ghcr.io/YOUR_USERNAME/industrial-ai-platform/backend:sha-xxx`
- Frontend: `ghcr.io/YOUR_USERNAME/industrial-ai-platform/frontend:sha-xxx`

### 4. Security Scan Job

```yaml
security-scan:
  - Trivy vulnerability scanner
  - Gosec (Go security checker)
```

---

## 🐳 使用构建镜像

### 拉取镜像部署

```bash
# 拉取最新镜像
docker pull ghcr.io/YOUR_USERNAME/industrial-ai-platform/backend:main
docker pull ghcr.io/YOUR_USERNAME/industrial-ai-platform/frontend:main

# 使用镜像运行
docker compose -f docker-compose.ghcr.yml up -d
```

### docker-compose.ghcr.yml

```yaml
services:
  frontend:
    image: ghcr.io/YOUR_USERNAME/industrial-ai-platform/frontend:main
    ports: ["3000:80"]
    
  backend:
    image: ghcr.io/YOUR_USERNAME/industrial-ai-platform/backend:main
    ports: ["8080:8080"]
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
```

---

## 📈 Coverage 报告

### 查看报告

1. 访问 https://codecov.io
2. 添加 GitHub 仓库
3. 查看 Coverage 趋势图

### CI 中的 Coverage 检查

```bash
# 后端 Coverage 70% 门槛
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
if [ "${COVERAGE}" -lt 70 ]; then
  echo "Warning: Coverage below threshold"
fi
```

---

## 🔧 自定义配置

### 修改触发条件

```yaml
on:
  push:
    branches: [main, develop]  # 添加 develop 分支
    paths:
      - 'backend/**'            # 仅后端变更触发
      - 'frontend/**'          # 或仅前端变更
```

### 添加部署 Job

```yaml
deploy:
  needs: [docker-build]
  runs-on: ubuntu-latest
  steps:
    - name: Deploy to server
      uses: appleboy/ssh-action@master
      with:
        host: ${{ secrets.SERVER_HOST }}
        username: ${{ secrets.SERVER_USER }}
        key: ${{ secrets.SERVER_KEY }}
        script: |
          cd /app
          docker compose pull
          docker compose up -d
```

---

## ⚠️ 注意事项

### 1. 同步冲突处理

如果 Gitee 和 GitHub 都有提交，可能产生冲突：

```bash
# 强制同步 Gitee 版本
git fetch gitee
git reset --hard gitee/main
git push github main --force
```

### 2. Secrets 安全

- 不要在 workflow 中硬编码密钥
- 使用 `secrets.` 引用
- GITHUB_TOKEN 自动提供，不要手动设置

### 3. 镜像权限

GHCR 镜像默认私有：
- 公开仓库 → 镜像公开
- 私有仓库 → 需要登录才能拉取

```bash
# 登录 GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### 4. 资源限制

GitHub Actions 免费额度：
- Public 仓库：无限
- Private 仓库：2000 分钟/月

---

## 📋 检查清单

部署前确认：

- [ ] GitHub 仓库已创建
- [ ] GitHub Actions 已启用
- [ ] CODECOV_TOKEN 已配置（可选）
- [ ] 首次同步已完成
- [ ] CI 运行成功（绿色 ✓）
- [ ] Coverage 报告可见
- [ ] Docker 镜像推送成功

---

## 🔗 相关链接

- GitHub Actions 文档：https://docs.github.com/en/actions
- Codecov：https://codecov.io
- GHCR：https://ghcr.io
- Trivy：https://github.com/aquasecurity/trivy