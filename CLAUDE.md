# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

工业 AI 智能体平台（Industrial AI Agent Platform）— 全栈工业设备监控与告警系统，包含 Go 后端、React 前端、C# 边缘模拟器和 SDK。

## 常用命令

### 后端（Go，在 `backend/` 目录下）

```bash
make build            # 编译到 bin/industrial-ai-platform
make test             # 运行全部测试（含覆盖率）
make test-unit        # 仅运行单元测试 (handler/service/pkg 层)
make test-integration # 集成测试（需要本地数据库）
make lint             # golangci-lint（5分钟超时）
make fmt              # gofmt 格式化
make security-scan    # govulncheck + golangci-lint(gosec)
make coverage-html    # 生成 HTML 覆盖率报告
```

运行单个测试：`go test ./internal/service/... -run TestFunctionName -v`

### 前端（React/TypeScript，在 `frontend/` 目录下）

```bash
npm run dev           # Vite 开发服务器 (端口 3000，代理 /api→8080, /ws→ws://8080)
npm run build         # TypeScript 编译 + 生产构建
npm test              # Vitest watch 模式
npm run test:run      # Vitest 单次运行
npm run test:coverage # 含覆盖率报告（阈值 50%）
npm run test:e2e      # Playwright E2E 测试
npm run lint          # ESLint
npm run typecheck     # tsc --noEmit 类型检查
npm run format        # Prettier 格式化
```

运行单个测试：`npx vitest run src/components/App.test.tsx`

### Docker

```bash
docker compose up -d          # 启动全部服务（frontend:3000, backend:8080, postgres, redis）
```

另有 `docker-compose.dev.yml`（开发）、`docker-compose.prod-ssl.yml`（生产 SSL）、`docker-compose.ghcr.yml`（GHCR 镜像）。

## 架构

### 后端分层架构（Go + Gin）

入口：`backend/main.go` → `handler.NewHTTPServerNew()`（`backend/internal/handler/server_new.go`）

```
Handler (HTTP 层) → Service (业务逻辑) → Repository (数据访问) → PostgreSQL/TimescaleDB
```

- **工厂模式**：`HandlerFactory`（handler/factory.go）、`ServiceFactory`（service/factory.go）和 `RepositoryFactory`（repository/factory.go）三层工厂负责依赖注入，所有组件通过工厂创建，测试时可通过替换工厂实现 mock
- **数据库**：PostgreSQL + TimescaleDB 时序扩展。`device_telemetry` 表按 `time` 分区为 hypertable，7 天前数据自动压缩（节省 ~90% 空间），90 天前数据自动删除。连接池 MaxOpen=100/MaxIdle=25
- **缓存**：Redis 优先，自动降级到内存缓存（`pkg/cache/`），支持 GetOrSet 模式和启动预热
- **WebSocket**：`internal/ws/` 处理实时推送，支持 pako 压缩，心跳保活，最大 1000 连接
- **认证**：JWT（24h 有效期）+ Token 版本控制（密码修改后旧 Token 自动失效）+ RBAC，密钥最少 32 字符
- **中间件栈**：`internal/middleware/` 包含认证、限流（登录 1/s burst 5、注册 1/s burst 3、API 100/s burst 200）、CORS、CSRF、安全头、审计日志、熔断器、Prometheus 指标等
- **AI 集成**：阿里云百炼 GLM-5 大模型（`LLM_API_KEY` 配置），支持多智能体（设备专家、维护专家、预测专家、优化专家）。未配置 API Key 时自动回退到规则引擎 Mock
- **可观测性**：OpenTelemetry tracing + Prometheus metrics
- **API 文档**：Swagger UI 在 `/docs/`，OpenAPI 3.0 规范在 `backend/docs/openapi.yaml`

路由结构（在 `server_new.go` 的 `setupRoutes()` 中注册）：
- `/health` — 健康检查
- `/docs/` — Swagger UI
- `/api/v1/auth/*` — 公开认证端点
- `/api/v1/devices/telemetry` — 遥测数据上报（设备端）
- `/ws` — WebSocket（含限流）
- `/api/v1/*`（需认证）— 设备、告警、导出、工单、AI 查询、通知、报告、ROI 等业务 API
- `/api/v1/admin/*`（需管理员）— 用户管理、系统配置

### 前端架构（React 19 + Vite + TypeScript）

- **路由**：React Router v6，所有页面组件通过 `lazyRoutes.tsx` 懒加载（`React.lazy()` + `Suspense`）
- **代码分割**：Vite 手动分块策略（react-core, react-router, charts, icons, graph, vendor），首屏 ~150KB
- **状态管理**：React Context（AuthContext、I18nProvider、MobileProvider、ToastProvider）+ 自定义 Hooks（`src/lib/hooks.ts`）
- **国际化**：`src/i18n/` + `src/locales/`（zh.ts / en.ts），支持中英文，基于 React Context 实现
- **API 客户端**：`src/lib/api.ts`，JWT 存储在 sessionStorage，30s 默认超时（AI 查询 60s），支持 AbortController 请求取消
- **移动端适配**：`src/lib/mobileOptimizations.ts`，底部导航栏响应式布局，触摸优化
- **样式**：TailwindCSS + 自定义主题

### 边缘层

- `edge/` — C# .NET 8 设备模拟器，支持多设备类型（CNC、注塑机、机器人等）和故障注入（5% 随机故障率）。环境变量：`DEVICE_COUNT`、`REPORT_INTERVAL_MS`、`API_BASE_URL`
- `sdk/` — C# Edge SDK，提供 HTTP 和 WebSocket 客户端供边缘设备集成

### CI/CD（GitHub Actions）

主要工作流：
- `backend.yml` — lint、单元测试、构建、安全扫描（govulncheck + gosec + trivy）、集成测试（PostgreSQL + Redis 服务）
- `frontend.yml` — ESLint + 类型检查、Vitest 覆盖率、生产构建、bundle 大小检查（>5MB 告警）
- `lint.yml` — 全量代码规范检查（golangci-lint + ESLint + Prettier + TypeScript）
- `security.yml` — 安全扫描
- `docker.yml` — Docker 镜像构建
- `e2e.yml` — Playwright E2E 测试（多浏览器 + 移动端）
- `benchmark.yml` — 性能基准测试

## 环境变量

后端必需：`DATABASE_URL`、`JWT_SECRET`（≥32字符）、`ADMIN_PASSWORD`

后端可选：`REDIS_URL`、`LLM_API_KEY`、`LLM_BASE_URL`（默认 `https://dashscope.aliyuncs.com/compatible-mode/v1`）、`LLM_MODEL`（默认 `qwen-plus`）、`CORS_ORIGINS`、`PORT`（默认 8080）、`WS_MAX_CONNECTIONS`、`WS_HEARTBEAT_INTERVAL`

前端：`VITE_API_URL`（默认通过 Vite 代理）

本地开发默认管理员：`admin` / `admin123`

## 代码风格

- 后端：gofmt + golangci-lint，驼峰命名，函数需有注释
- 前端：ESLint + Prettier，TypeScript strict 模式，函数式组件 + Hooks
- 提交规范：Conventional Commits（`feat(scope): 描述`）
- 所有注释和日志使用中文

## 测试策略

- 后端：单元测试分布在各层（handler/service/repository/pkg），覆盖率目标 70%+，集成测试需本地数据库。Mock 通过替换 HandlerFactory/ServiceFactory 注入
- 前端：Vitest 单元测试（jsdom, 50% 阈值）+ Playwright E2E（Chrome/Firefox/Safari + Pixel 5/iPhone 12），测试文件与组件同目录。Mock 使用 `vi.mock()`
- 测试配置：Vitest 用 `vitest.config.ts`（v8 覆盖率提供者），Playwright 用 `playwright.config.ts`
