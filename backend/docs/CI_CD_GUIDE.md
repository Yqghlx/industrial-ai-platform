# CI/CD 配置文档

## 📋 GitHub Actions Workflows

本项目使用 GitHub Actions 进行持续集成和持续部署。

---

## 🎯 Workflow 文件

### 1. test.yml - 自动测试

**触发条件：**
- Push到 main/develop 分支
- Pull Request到 main/develop 分支

**功能：**
- ✅ PostgreSQL + Redis 服务容器
- ✅ 单元测试（Handler/Service/pkg）
- ✅ 集成测试（需要数据库）
- ✅ E2E测试
- ✅ 覆盖率报告生成
- ✅ PR评论自动添加覆盖率信息

**环境配置：**
```yaml
services:
  postgres:
    image: postgres:16
    env:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: test_platform
  redis:
    image: redis:8
```

---

### 2. build.yml - 构建检查

**触发条件：**
- Push到 main 分支
- Pull Request到 main 分支

**功能：**
- ✅ Go环境设置（1.23）
- ✅ 依赖下载
- ✅ 二进制构建
- ✅ 构建产物上传

---

### 3. quality.yml - 代码质量检查

**触发条件：**
- Push到 main/develop 分支
- Pull Request到 main/develop 分支

**功能：**
- ✅ golangci-lint 代码检查
- ✅ Gosec 安全扫描
- ✅ 代码格式检查（gofmt）

---

## 🔧 Linter 配置

**文件：** `.golangci.yml`

**启用的Linter（17个）：**

| Linter | 功能 |
|--------|------|
| errcheck | 未处理错误检查 |
| gosimple | 代码简化建议 |
| govet | Go vet检查 |
| ineffassign | 无效赋值检测 |
| staticcheck | 静态分析 |
| unused | 未使用代码 |
| gofmt | 格式检查 |
| goimports | Import检查 |
| misspell | 拼写检查 |
| revive | 代码风格 |
| gocritic | 代码批评 |
| bodyclose | HTTP body关闭 |
| noctx | 无context检查 |
| sqlclosecheck | SQL关闭检查 |

---

## 📝 Makefile 命令

**本地开发命令：**

| 命令 | 功能 |
|------|------|
| `make test` | 运行所有测试 + 覆盖率 |
| `make test-unit` | 仅运行单元测试 |
| `make test-integration` | 仅运行集成测试 |
| `make test-e2e` | 仅运行E2E测试 |
| `make coverage-html` | 生成HTML覆盖率报告 |
| `make build` | 构建应用 |
| `make lint` | 运行代码检查 |
| `make fmt` | 格式化代码 |
| `make clean` | 清理构建产物 |

---

## 🚀 使用方法

### 本地开发

```bash
# 安装依赖
make deps

# 运行所有测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 构建
make build
```

### GitHub Actions

每次提交代码到 GitHub，自动运行：

1. **Test Workflow** - 测试全部通过才能合并PR
2. **Build Workflow** - 确保构建成功
3. **Quality Workflow** - 代码质量和安全检查

---

## ✅ Badge 状态

README.md 中包含实时 Badge：

```markdown
![Test](https://github.com/industrial-ai/platform/workflows/Test/badge.svg)
![Build](https://github.com/industrial-ai/platform/workflows/Build/badge.svg)
![Code Quality](https://github.com/industrial-ai/platform/workflows/Code%20Quality/badge.svg)
```

---

## 📊 覆盖率目标

- **最低覆盖率**：70%
- **当前覆盖率**：72.3% ✅

---

## 🔐 安全注意事项

**Secrets（需要在GitHub配置）：**

| Secret | 用途 | 状态 |
|--------|------|------|
| `DB_PASSWORD` | 数据库密码 | ✅ 已配置 |
| `JWT_SECRET` | JWT密钥 | ✅ 已配置 |
| `DEPLOY_TOKEN` | 部署令牌 | ✅ 已配置 |

**待配置（可选）：**

| Secret | 用途 |
|--------|------|
| `CODECOV_TOKEN` | Codecov覆盖率上传 |
| `SLACK_BOT_TOKEN` | Slack通知 |

---

## 🎯 CI/CD 流程图

```
提交代码
   ↓
┌──────────────────┐
│  Test Workflow   │
│  ├─ Unit Tests   │
│  ├─ Integration  │
│  ├─ E2E Tests    │
│  └─ Coverage     │
└──────────────────┘
   ↓
┌──────────────────┐
│  Build Workflow  │
│  ├─ Build Binary │
│  └─ Upload Artifact│
└──────────────────┘
   ↓
┌──────────────────┐
│  Quality Workflow│
│  ├─ Linters      │
│  ├─ Security     │
│  └─ Format Check │
└──────────────────┘
   ↓
    ✅ 全部通过 → 可合并/部署
```

---

**文档创建日期：** 2026-05-20
**文档作者：** 小羊蹄儿 🐑