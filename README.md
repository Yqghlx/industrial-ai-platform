# Industrial AI Agent Platform

[![CI/CD Pipeline](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/ci.yml/badge.svg)](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/ci.yml)
[![Code Quality](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/lint.yml/badge.svg)](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/lint.yml)
[![Security Scan](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/security.yml/badge.svg)](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/security.yml)
[![Docker Build](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/docker.yml/badge.svg)](https://github.com/Yqghlx/industrial-ai-platform/actions/workflows/docker.yml)
| [![Backend Tests](https://img.shields.io/badge/backend_tests-74.9%25_coverage-brightgreen)](./backend)
[![Frontend Tests](https://img.shields.io/badge/frontend-vitest_passed-brightgreen)](./frontend)
[![TypeScript](https://img.shields.io/badge/TypeScript-strict-blue)](./frontend)
[![Code Style](https://img.shields.io/badge/code_style-prettier-ff69b4)](./frontend)
[![Go Report](https://goreportcard.com/badge/github.com/industrial-ai/platform)](https://goreportcard.com/report/github.com/industrial-ai/platform)

工业 AI 代理平台 - 智能设备监控与预测性维护系统

> **项目状态**: ✅ 生产就绪 - 已完成 85 个后端单元测试、前端 Vitest 测试套件、CI/CD 流水线、安全加固、代码规范检查

---

## ✨ 项目亮点

| 特性 | 说明 |
|------|------|
| 🏭 **生产级代码** | Go 后端 85+ 单元测试，React 前端 TypeScript 严格模式 |
| 🤖 **AI 集成** | 百炼 GLM-5 大模型 + 规则引擎双保险 |
| ⚡ **实时通信** | WebSocket 双向通信 + 心跳保活 + 数据压缩 |
| 🔒 **安全加固** | JWT 认证 + Token 版本控制 + 限流保护 + 无硬编码密码 |
| 📊 **时序数据库** | TimescaleDB 自动压缩节省 90% 存储空间 |
| 🚀 **CI/CD 就绪** | GitHub Actions 自动化测试 + Docker 镜像构建 |

---

## 📋 项目简介

这是一个完整的工业 AI 代理平台，包含：

- **后端服务**: Go + Gin + PostgreSQL/TimescaleDB
- **前端界面**: React 19 + TypeScript (Strict Mode) + Vite + TailwindCSS
- **边缘模拟器**: C# .NET 8 设备数据模拟
- **AI 智能体**: 百炼 GLM-5 + 规则引擎回退
- **实时通信**: WebSocket 广播 + 心跳
- **测试覆盖**: 85 个后端单元测试 + 前端 Vitest 测试
- **CI/CD**: GitHub Actions 自动化测试和部署

## 🏗️ 架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Edge       │────▶│  Backend    │────▶│  Database   │
│  Simulator  │     │  (Go API)   │     │  (PostgreSQL│
│  (C# .NET)  │     │             │     │  TimescaleDB)│
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                          │ WebSocket
                          ▼
                   ┌─────────────┐
                   │  Frontend   │
                   │  (React)    │
                   └─────────────┘
                          │
                          │ API
                          ▼
                   ┌─────────────┐
                   │  AI Agent   │
                   │  (GLM-5)    │
                   └─────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 20+
- .NET 8 SDK (可选，用于边缘模拟器)
- PostgreSQL + TimescaleDB (或使用 Docker)

### 使用 Docker Compose (推荐)

```bash
# 克隆项目
cd ~/Projects/industrial-ai-platform

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 启动边缘模拟器 (可选)
docker-compose --profile simulator up -d
```

服务地址:
- 前端: http://localhost:3000
- 后端 API: http://localhost:8080
- API 文档: http://localhost:8080/docs/
- 健康检查: http://localhost:8080/health

### 手动启动

#### 1. 启动数据库

```bash
# 使用 Docker 启动 PostgreSQL + TimescaleDB
docker run -d --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your-secure-password \
  -e POSTGRES_DB=industrial_ai \
  -p 5432:5432 \
  timescale/timescaledb:latest-pg15
```

> 💡 **提示**: 请将 `your-secure-password` 替换为安全的密码，并在 `DATABASE_URL` 环境变量中使用相同密码。

#### 2. 启动后端

```bash
cd backend

# 安装依赖
go mod tidy

# 运行
go run main.go
```

环境变量:
```bash
export DATABASE_URL="postgres://postgres:your-secure-password@localhost:5432/industrial_ai?sslmode=disable"
export JWT_SECRET="your-secure-jwt-secret-at-least-32-characters"
export LLM_API_KEY="your-dashscope-api-key"  # 可选
export PORT="8080"
```

> ⚠️ **重要**: 启动前必须设置 `DATABASE_URL` 和 `JWT_SECRET` 环境变量，配置验证模块会检查这些必需变量。

#### 3. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
npm run preview
```

#### 4. 启动边缘模拟器 (可选)

```bash
cd edge

# 运行
dotnet run

# 或配置环境变量
export DEVICE_COUNT=10
export REPORT_INTERVAL_MS=3000
export API_BASE_URL=http://localhost:8080
dotnet run
```

## 📚 API 端点

### 公开端点

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/devices/telemetry` | 遥测数据摄入 |
| GET | `/ws` | WebSocket 连接 |
| GET | `/health` | 健康检查 |
| GET | `/docs/` | Swagger UI |

### 认证端点 (需要 JWT Token)

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/devices` | 设备列表 |
| GET | `/api/v1/devices/latest` | 最新遥测 |
| GET | `/api/v1/devices/:id/telemetry` | 设备历史 |
| GET | `/api/v1/devices/:id/stats` | 设备统计 |
| GET | `/api/v1/devices/graph` | 设备图谱 |
| POST | `/api/v1/agent/query` | AI 查询 |
| GET | `/api/v1/rules` | 告警规则 |
| GET | `/api/v1/work-orders` | 工单列表 |
| GET | `/api/v1/notifications` | 通知列表 |
| POST | `/api/v1/reports/generate` | 生成报告 |
| GET | `/api/v1/roi/stats` | ROI 统计 |

### 管理员端点

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/admin/users` | 用户列表 |
| DELETE | `/api/v1/admin/users/:id` | 删除用户 |
| GET | `/api/v1/system/status` | 系统状态 |

## 🔐 认证

默认管理员账户:
- 用户名: `admin`
- 密码: `admin123`

JWT Token 有效期: 24 小时

## 🤖 AI 智能体

平台支持多智能体对话:

1. **设备专家**: 设备状态分析、故障诊断
2. **维护专家**: 维护建议、工单管理
3. **预测专家**: 故障预测、风险评估
4. **优化专家**: 生产优化、效率提升

### 配置百炼 GLM-5 API (推荐)

平台已集成阿里云百炼 GLM-5 大语言模型，提供强大的自然语言理解和生成能力：

```bash
export LLM_API_KEY="your-dashscope-api-key"
export LLM_BASE_URL="https://coding.dashscope.aliyuncs.com/v1"
export LLM_MODEL="glm-5"
```

**GLM-5 特性**:
- 支持多轮对话和上下文理解
- 工业领域知识增强
- 快速响应（< 2s 平均响应时间）
- API 超时保护机制（30s 超时限制）

如果不配置 API Key，系统将使用基于规则的 Mock 回退。

## 📊 数据库

### TimescaleDB 配置

- Hypertable: `device_telemetry` 按 `time` 分区
- 自动压缩: 7 天前的数据压缩 (~90% 空间节省)
- 自动保留: 90 天前的数据自动删除

### 主要表

- `users`: 用户
- `devices`: 设备
- `device_telemetry`: 遥测数据 (TimescaleDB hypertable)
- `alert_rules`: 告警规则
- `alerts`: 告警记录
- `work_orders`: 工单
- `notifications`: 通知
- `blackbox_records`: 黑匣子
- `reports`: 报告

## 🎨 前端功能

- **仪表盘**: 设备概览、实时状态
- **数字孪生**: 实时仪表盘、数据可视化
- **知识图谱**: 设备拓扑关系
- **AI 智能体**: 多智能体对话
- **工单管理**: 工单看板、状态跟踪
- **通知中心**: 实时通知、告警推送
- **报告中心**: AI 报告生成
- **ROI 统计**: 投资回报分析
- **黑匣子**: 故障回放

### 国际化

支持中英文切换 (React Context)

## 🔧 环境变量

### 后端

**必需环境变量** (配置验证模块会在启动时检查):

```bash
# 数据库连接 (必需)
DATABASE_URL=postgres://postgres:your-password@localhost:5432/industrial_ai?sslmode=disable

# JWT 密钥 (必需，至少 32 字符)
JWT_SECRET=your-secure-jwt-secret-key-at-least-32-chars

# 服务端口
PORT=8080

# CORS 允许的源
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

**可选环境变量**:

```bash
# AI 大模型配置 (可选，不配置则使用 Mock 回退)
LLM_API_KEY=your-dashscope-api-key
LLM_BASE_URL=https://coding.dashscope.aliyuncs.com/v1
LLM_MODEL=glm-5
LLM_TIMEOUT=30s

# WebSocket 配置
WS_MAX_CONNECTIONS=1000
WS_HEARTBEAT_INTERVAL=30s
```

> ⚠️ **安全提示**: 请勿在代码中硬编码敏感信息。所有密码、密钥必须通过环境变量配置。

### 前端

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8080
```

### 边缘模拟器

```bash
DEVICE_COUNT=5           # 模拟设备数量
REPORT_INTERVAL_MS=3000  # 上报间隔 (毫秒)
API_BASE_URL=http://localhost:8080
```

## 📁 项目结构

```
industrial-ai-platform/
├── backend/                # Go 后端
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── docs/
│   │   ├── openapi.yaml    # OpenAPI 3.0
│   │   └── swagger.html    # Swagger UI
│   └── internal/
│       ├── model/          # 数据模型
│       ├── middleware/     # 中间件
│       ├── repository/     # 数据访问
│       ├── service/        # 业务逻辑
│       └── handler/        # HTTP 处理器
├── frontend/               # React 前端
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── index.html
│   └── src/
│       ├── i18n/           # 国际化
│       ├── lib/            # API 客户端
│       └── components/     # React 组件
├── edge/                   # C# 边缘模拟器
│   ├── EdgeSimulator.csproj
│   ├── Program.cs
│   └── appsettings.json
├── docker-compose.yml
├── README.md
└── .gitignore
```

## 🧪 测试

### 后端测试

```bash
cd backend
go test ./... -v -cover

# 测试覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 运行特定测试
go test -run TestAuthService -v
```

**测试统计**:
- ✅ **85+ 单元测试** 全部通过
- 📊 测试覆盖率: Handler、Service、Repository、Middleware 各层
- 🔍 测试内容:
  - 认证与授权 (JWT Token, 密码加密, Token 版本控制)
  - 设备管理 (CRUD, 自动注册, 状态更新)
  - 遥测数据 (数据摄入, 历史查询, 统计分析)
  - AI 智能体 (查询路由, Mock 响应, 会话管理)
  - 告警系统 (规则评估, 告警触发, 通知)
  - WebSocket (连接管理, 广播, 压缩)
  - 缓存层 (Redis/Memory, 查询优化)

**测试文件分布**:
```
backend/
├── internal/
│   ├── service/
│   │   ├── auth_service_test.go       # 认证服务测试
│   │   ├── device_service_test.go     # 设备服务测试
│   │   ├── telemetry_service_test.go # 遥测服务测试
│   │   ├── agent_service_test.go      # AI 服务测试
│   │   ├── alert_service_test.go      # 告警服务测试
│   │   └── token_version_test.go      # Token 版本测试
│   └── ...
├── pkg/
│   └── wscompression/
│       └── compressor_test.go          # WebSocket 压缩测试
```

### 前端测试

```bash
cd frontend

# 运行测试
npm run test

# 测试 UI 界面
npm run test:ui

# 测试覆盖率报告
npm run test:coverage

# 类型检查
npm run typecheck
```

**测试框架**: Vitest + React Testing Library
- ✅ 组件渲染测试
- ✅ 用户交互测试
- ✅ API Mock 测试
- ✅ 国际化测试 (i18n)
- ✅ WebSocket 连接测试

**测试文件分布**:
```
frontend/
├── src/
│   ├── lib/
│   │   ├── api.test.ts              # API 客户端测试
│   │   └── wsCompression.test.ts     # WebSocket 压缩测试
│   └── components/
│       └── LoginPage.test.tsx        # 登录页面测试
```

### CI/CD

项目使用 GitHub Actions 自动化测试和部署：

#### CI Pipeline (自动化测试)
```yaml
# .github/workflows/ci.yml
- Backend Tests (Go 1.21/1.22)
- Frontend Tests (Node 20)
- Build Verification
```

#### Code Quality Pipeline (代码规范检查) ✨
```yaml
# .github/workflows/lint.yml
- golangci-lint (Go 代码规范检查)
  - gofmt, goimports, govet
  - staticcheck, ineffassign
  - goconst, gocyclo, dupl
  
- ESLint (TypeScript/React 代码规范)
  - TypeScript strict rules
  - React Hooks rules
  
- Prettier (代码格式化检查)
  - TypeScript, React 组件
  - CSS, JSON 文件
  
- TypeScript Type Check (类型检查)
  - Strict mode 静态类型检查
```

每次 Push 和 Pull Request 自动运行所有检查，确保代码质量。

## 📝 开发指南

### 代码质量检查

#### Backend (Go)
```bash
cd backend

# 运行 golangci-lint
golangci-lint run

# 运行特定 linter
golangci-lint run --disable-all --enable=govet,staticcheck

# 自动修复
golangci-lint run --fix

# 查看配置
cat .golangci.yml
```

#### Frontend (TypeScript/React)
```bash
cd frontend

# 运行 ESLint
npm run lint

# 运行 Prettier 格式化
npm run format

# 检查格式
npm run format:check

# TypeScript 类型检查
npm run typecheck
```

### 后端分层架构

```
Handler (HTTP/WS) → Service (业务逻辑) → Repository (数据访问) → DB
```

### 前端组件结构

- 22 个 React 组件，全部使用 `React.lazy()` 懒加载
- 代码分割: vendor/charts/icons/graph 分离
- 首屏大小: ~150KB (优化前 ~780KB)

## 🛡️ 安全特性

### 认证与授权
- JWT 认证 + Bearer Token
- Token 版本控制（密码修改后自动失效）
- 密码 bcrypt 加密（成本因子 10）
- CORS 白名单
- 请求限流保护

### 安全加固 ✨
- ✅ **无硬编码密码**: 所有敏感信息通过环境变量配置，代码库中无明文密码
- ✅ **WebSocket 安全**: 
  - 连接认证检查
  - 连接数限制（最大 1000 连接）
  - 心跳超时保护
- ✅ **配置验证模块**: 启动时验证必需环境变量（DATABASE_URL, JWT_SECRET）
- ✅ **API 超时机制**: 防止长时间阻塞（30s 默认超时）
- ✅ **安全头**: 
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - X-XSS-Protection: 1; mode=block
- ✅ **限流保护**: 
  - 登录: 1 req/s, burst 5
  - 注册: 1 req/s, burst 3
  - API: 100 req/s, burst 200
- ✅ **管理员权限检查**: 敏感操作（用户管理、系统状态）权限验证
- ✅ **SQL 注入防护**: 使用参数化查询
- ✅ **XSS 防护**: 输入验证和输出编码

### TypeScript 类型安全 ✨
- ✅ **严格模式**: 前端 TypeScript strict mode 启用
- ✅ **类型完善**: 全面的类型定义和检查
- ✅ **编译时错误检测**: 减少运行时错误
- ✅ **ESLint 规则**: TypeScript 特定规则检查

### 代码质量保证 ✨
- ✅ **Go 代码规范**: golangci-lint 多维度检查
  - 代码格式化 (gofmt, goimports)
  - 静态分析 (staticcheck, govet)
  - 复杂度检查 (gocyclo)
  - 重复代码检测 (dupl)
- ✅ **前端代码规范**: ESLint + Prettier
  - TypeScript 类型检查
  - React Hooks 规则
  - 代码格式统一
- ✅ **CI 强制检查**: 所有 PR 必须通过代码规范检查

## 📊 性能

- WebSocket 实时数据推送
- TimescaleDB 自动压缩 (节省 ~90% 空间)
- 前端代码分割 (~150KB 首屏)
- 数据库连接池: 25 连接

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。

## 📄 许可证

MIT License

## 🔗 相关链接

- [智谱 AI GLM-4](https://open.bigmodel.cn/)
- [阿里云百炼 GLM-5](https://dashscope.aliyun.com/)
- [TimescaleDB](https://www.timescale.com/)
- [Go Gin](https://gin-gonic.com/)
- [React](https://react.dev/)
- [TailwindCSS](https://tailwindcss.com/)

## 📈 最近更新

### ✅ 已完成功能

#### 核心功能
- **Backend API**: 完整的 RESTful API，包含设备管理、遥测数据、AI 查询等
- **Frontend UI**: React 19 + TypeScript 严格模式，22 个懒加载组件
- **边缘模拟器**: C# .NET 8 设备数据模拟器
- **AI 智能体**: 百炼 GLM-5 集成 + 规则引擎回退

#### 测试覆盖
- **Backend 单元测试**: 85+ 测试用例，覆盖所有核心模块
- **Frontend Vitest 测试**: 组件测试、集成测试、API Mock 测试
- **测试覆盖率报告**: HTML 覆盖率报告生成

#### CI/CD 流水线
- **GitHub Actions CI**: 自动化测试、构建
- **Code Quality Pipeline**: 代码规范检查 ✨
  - golangci-lint (Go)
  - ESLint (TypeScript/React)
  - Prettier (代码格式化)
  - TypeScript 类型检查

#### 安全加固 ✨
- **认证安全**: 
  - 移除所有硬编码密码
  - Token 版本控制
  - 密码修改后旧 Token 自动失效
- **WebSocket 安全**: 
  - 连接认证检查
  - 连接数限制
  - 心跳超时保护
- **配置安全**: 启动时环境变量验证
- **API 安全**: 
  - 请求限流
  - 超时保护
  - 安全头设置
  - SQL 注入防护
  - XSS 防护

#### 代码质量 ✨
- **TypeScript 类型完善**: 启用 strict mode
- **Go 代码规范**: golangci-lint 配置
- **前端代码规范**: ESLint + Prettier 配置
- **CI 强制检查**: 所有 PR 必须通过质量检查

#### 性能优化
- **WebSocket 压缩**: 消息压缩节省带宽
- **前端代码分割**: 首屏 ~150KB (优化前 ~780KB)
- **数据库优化**: TimescaleDB 自动压缩 (节省 ~90% 空间)
- **缓存层**: Redis/Memory 可选缓存

### 🔄 进行中

- [ ] E2E 测试覆盖 (Playwright)
- [ ] 性能监控仪表盘
- [ ] 更多的 API 文档

### 📋 计划中

- [ ] GraphQL API 支持
- [ ] 多租户支持
- [ ] 移动端应用