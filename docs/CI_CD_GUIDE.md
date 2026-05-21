# GitHub Actions CI/CD 配置指南

## 概述

本项目使用 GitHub Actions 实现完整的 CI/CD 自动化流水线。

## 工作流列表

| 工作流 | 触发条件 | 功能 |
|--------|---------|------|
| `backend.yml` | backend 目录变更 | Go 后端测试 + 构建 + 安全扫描 |
| `frontend.yml` | frontend 目录变更 | 前端构建 + Vitest 测试 + Bundle 检查 |
| `e2e.yml` | main 分支/定时/手动 | Playwright E2E 测试 |
| `deploy.yml` | main 分支/tag | Docker 构建 + Kubernetes 部署 |
| `release.yml` | v* tags | 发布 Release + 构建各平台二进制 |
| `security.yml` | push/PR/定时 | CodeQL + Trivy + Secret 扫描 |
| `benchmark.yml` | main 分支/定时 | k6 性能基准测试 |

---

## 🔧 工作流详解

### 1. Backend CI (`backend.yml`)

```yaml
触发条件:
  - push: main, develop (backend/**)
  - PR: main (backend/**)

执行步骤:
  1. golangci-lint 代码检查
  2. Go 单元测试 + 覆盖率
  3. 构建 Linux 二进制
  4. Gosec 安全扫描
  5. Trivy 漏洞扫描
  6. 集成测试 (PostgreSQL + Redis)
```

**覆盖率上传**: Codecov 自动收集

---

### 2. Frontend CI (`frontend.yml`)

```yaml
触发条件:
  - push: main, develop (frontend/**)
  - PR: main (frontend/**)

执行步骤:
  1. ESLint + TypeScript 检查
  2. Vitest 单元测试 + 覆盖率
  3. Vite 生产构建
  4. Bundle 大小检查 (< 5MB)
```

---

### 3. E2E Tests (`e2e.yml`)

```yaml
触发条件:
  - push: main (全项目)
  - PR: main
  - 定时: 每日 02:00 UTC
  - 手动触发 (可选浏览器)

执行步骤:
  1. 启动 PostgreSQL + Redis 服务
  2. 构建后端 + 启动服务
  3. 安装 Playwright 浏览器
  4. 启动前端开发服务器
  5. 运行 E2E 测试
  6. 上传测试报告 + 失败截图

浏览器策略:
  - PR: 仅 Chromium (快速反馈)
  - main/定时: 全浏览器 (Chromium/Firefox/Safari + Mobile)
```

---

### 4. Build & Deploy (`deploy.yml`)

```yaml
触发条件:
  - push: main 分支
  - tag: v* 版本
  - 手动触发 (staging/production)

执行步骤:
  1. 构建 Backend Docker 镜像 (多平台: amd64/arm64)
  2. 构建 Frontend Docker 阆像
  3. 推送到 GitHub Container Registry (ghcr.io)
  4. 部署到 Staging 环境
  5. 部署到 Production 环境 (仅 tag)
  6. 验证部署健康状态
  7. 失败自动回滚

部署策略:
  - Staging: 每次 main 分支推送
  - Production: 仅版本 tag + 手动确认
  - Blue-Green Deployment
```

---

### 5. Release (`release.yml`)

```yaml
触发条件:
  - tag: v*.*.* 版本号

执行步骤:
  1. 创建 GitHub Release
  2. 构建各平台二进制:
     - linux-amd64
     - linux-arm64
     - windows-amd64
     - darwin-amd64 (Mac Intel)
     - darwin-arm64 (Mac M1/M2)
  3. 构建前端产物打包
  4. 上传 Release Assets
  5. 更新 CHANGELOG.md
  6. 发送 Slack 通知
```

---

### 6. Security Scan (`security.yml`)

```yaml
触发条件:
  - push: main
  - PR: main
  - 定时: 每周一 03:00 UTC

执行步骤:
  1. CodeQL 分析 (Go + TypeScript)
  2. 依赖审查 (PR)
  3. Trivy 文件系统扫描
  4. Gosec Go 安全扫描
  5. TruffleHog + Gitleaks Secret 扫描
  6. NPM Audit
```

---

### 7. Performance Benchmark (`benchmark.yml`)

```yaml
触发条件:
  - push: main (backend/**, benchmarks/**)
  - 定时: 每周日 04:00 UTC

执行步骤:
  1. 启动测试环境
  2. 运行 k6 API 负载测试
  3. 运行 WebSocket 压力测试
  4. 生成基准报告
  5. 运行 Go 内置 benchmark
  6. 比较历史数据 + 告警
```

---

## 📊 测试矩阵

| 测试类型 | 工具 | 频率 | 覆盖率目标 |
|---------|------|------|-----------|
| 单元测试 | Go test + Vitest | 每次提交 | > 80% |
| E2E 测试 | Playwright | main + 每日 | 核心场景 |
| 安全扫描 | CodeQL + Trivy | 每次提交 + 每周 | N/A |
| 性能测试 | k6 | 每周 | N/A |

---

## 🔐 Secrets 配置

需要在 GitHub Settings > Secrets 中配置：

| Secret | 用途 |
|--------|------|
| `GITHUB_TOKEN` | 自动提供，用于 Release/Registry |
| `KUBE_CONFIG_STAGING` | Staging K8s 部署凭证 (base64) |
| `KUBE_CONFIG_PRODUCTION` | Production K8s 部署凭证 (base64) |
| `DEPLOY_TOKEN` | 部署配置仓库访问 |
| `SLACK_BOT_TOKEN` | Slack 通知 |
| `CODECOV_TOKEN` | 覆盖率上传 (可选) |
| `GITLEAKS_LICENSE` | Gitleaks 组织版 (可选) |

---

## 🚀 使用指南

### 本地验证

```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm run test:run

# E2E 测试
cd frontend
npm run test:e2e

# 安全扫描
trivy fs .
gosec ./backend/...
```

### 触发特定工作流

```bash
# 手动触发 E2E 测试
gh workflow run e2e.yml -f browser=chromium

# 手动部署到 staging
gh workflow run deploy.yml -f environment=staging

# 手动触发安全扫描
gh workflow run security.yml
```

### 查看运行状态

```bash
# 列出最近运行
gh run list

# 查看特定运行详情
gh run view <run-id>

# 下载 artifacts
gh run download <run-id>
```

---

## 📋 Branch Protection Rules

建议在 GitHub Settings > Branches 中配置：

| 规则 | 设置 |
|------|------|
| Require PR reviews | 至少 1 人审核 |
| Require status checks | backend-ci, frontend-ci |
| Require E2E tests | 仅 main 分支 |
| Require signed commits | 可选 |
| Include administrators | 建议启用 |

---

## 🔄 CI/CD 流程图

```
提交代码
   │
   ├── backend/** → Backend CI
   │     ├── lint → test → build → security
   │
   ├── frontend/** → Frontend CI  
   │     ├── lint → test → build
   │
   └── main 分支 → E2E Tests
         └── setup → e2e → report
   
合并到 main
   │
   ├── Build & Deploy → Staging
   │     ├── Docker Build → Push → Deploy
   │
   └── 每日定时 → Security + Benchmark
   
发布版本 (v* tag)
   │
   ├── Release → 构建 Assets
   ├── Deploy → Production
   └── Notify → Slack
```

---

## ⚡ 性能优化

| 优化项 | 实现 |
|--------|------|
| 依赖缓存 | npm ci + go mod download |
| Docker 缓存 | GitHub Actions cache |
| 并行测试 | matrix strategy |
| E2E 快速反馈 | PR 仅测 Chromium |
| 资源限制 | timeout-minutes 配置 |

---

**CI/CD 框架版本**: 1.0.0  
**更新日期**: 2026-05-13  
**维护者**: Industrial AI Team