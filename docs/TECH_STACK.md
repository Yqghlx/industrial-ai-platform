# 技术栈文档 (TECH_STACK)

> 文档版本: 1.0.0  
> 更新日期: 2026-05-14  
> 项目: Industrial AI Platform

---

## 1. 后端技术栈详情

### 1.1 核心框架与语言

| 技术 | 版本 | 用途 | 说明 |
|------|------|------|------|
| **Go** | 1.25.0 | 编程语言 | 高性能、并发友好 |
| **Gin** | 1.9.1 | Web框架 | 轻量级HTTP路由框架 |

### 1.2 数据存储

| 技术 | 版本 | 用途 | 配置 |
|------|------|------|------|
| **PostgreSQL** | (lib/pq 1.2.0) | 主数据库 | 连接池: MaxOpen=25, MaxIdle=5, Lifetime=5min |
| **Redis** | 9.7.0 (go-redis) | 缓存层 | 支持内存缓存fallback |
| **sqlx** | 1.3.5 | SQL扩展 | 结构化查询映射 |

### 1.3 认证与安全

| 技术 | 版本 | 用途 |
|------|------|------|
| **jwt/v5** | 5.2.0 | JWT Token生成与验证 |
| **golang.org/x/crypto** | 0.49.0 | 密码哈希(bcrypt) |

### 1.4 实时通信

| 技术 | 版本 | 用途 | 特性 |
|------|------|------|------|
| **gorilla/websocket** | 1.5.1 | WebSocket服务 | 压缩支持、心跳机制 |

### 1.5 监控与追踪

| 技术 | 版本 | 用途 |
|------|------|------|
| **Prometheus client_golang** | 1.23.2 | 指标采集与导出 |
| **OpenTelemetry** | 1.43.0 | 分布式追踪 |
| **Zap** | 1.26.0 | 结构化日志 |

### 1.6 其他依赖

| 技术 | 版本 | 用途 |
|------|------|------|
| **google/uuid** | 1.6.0 | UUID生成 |
| **stretchr/testify** | 1.11.1 | 单元测试框架 |
| **DATA-DOG/go-sqlmock** | 1.5.2 | 数据库Mock测试 |

### 1.7 后端目录结构

```
backend/
├── cmd/
│   └── server/          # 主入口
├── internal/
│   ├── handler/         # API处理器
│   │   ├── server.go    # 服务核心
│   │   ├── routes.go    # 路由定义
│   │   ├── auth_handler.go
│   │   ├── device_handler.go
│   │   ├── telemetry_handler.go
│   │   ├── business_handler.go
│   │   ├── rbac_handler.go
│   │   ├── tenant_handler.go
│   │   ├── export_handler.go
│   │   ├── websocket.go
│   │   └── admin.go
│   ├── middleware/      # 中间件
│   │   ├── auth.go      # JWT认证
│   │   ├── ratelimit.go # 限流
│   │   ├── cors.go      # CORS
│   │   ├── logging.go   # 日志
│   │   ├── tracing.go   # 追踪
│   │   ├── tenant.go    # 租户隔离
│   │   ├── csrf.go      # CSRF防护
│   │   └── waf.go       # WAF防护
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── validator/       # 数据验证
├── pkg/
│   ├── cache/           # 缓存服务
│   ├── cache_service/   # 缓存集成
│   └── wscompression/   # WebSocket压缩
├── docs/                # 静态文档
└── go.mod               # Go模块定义
```

---

## 2. 前端技术栈详情

### 2.1 核心框架

| 技术 | 版本 | 用途 | 说明 |
|------|------|------|------|
| **React** | 19.0.0 | UI框架 | 组件化前端框架 |
| **React DOM** | 19.0.0 | DOM渲染 | React渲染器 |
| **React Router DOM** | 6.20.0 | 路由管理 | SPA路由 |
| **TypeScript** | 5.3.0 | 类型系统 | 强类型JavaScript |

### 2.2 构建工具

| 技术 | 版本 | 用途 | 特性 |
|------|------|------|------|
| **Vite** | 5.0.0 | 构建工具 | 快速冷启动、HMR |
| **@vitejs/plugin-react** | 4.2.0 | React插件 | JSX转换 |
| **esbuild** | (内置) | 压缩 | Go编写的高性能压缩 |
| **rollup-plugin-visualizer** | 5.12.0 | 打包分析 | 可视化依赖图 |

### 2.3 UI与样式

| 技术 | 版本 | 用途 |
|------|------|------|
| **Tailwind CSS** | 3.4.0 | CSS框架 |
| **PostCSS** | 8.4.32 | CSS处理器 |
| **Autoprefixer** | 10.4.16 | CSS兼容 |

### 2.4 图表与可视化

| 技术 | 版本 | 用途 |
|------|------|------|
| **Recharts** | 2.12.0 | 数据图表库 |
| **react-force-graph-2d** | 1.25.0 | 知识图谱可视化 |

### 2.5 图标与工具

| 技术 | 版本 | 用途 |
|------|------|------|
| **lucide-react** | 0.460.0 | 图标库 |
| **pako** | 2.1.0 | 数据压缩解压 |

### 2.6 测试框架

| 技术 | 版本 | 用途 |
|------|------|------|
| **Vitest** | 1.2.0 | 单元测试 |
| **@vitest/coverage-v8** | 1.2.0 | 覆盖率报告 |
| **@testing-library/react** | 14.2.0 | React测试工具 |
| **@testing-library/jest-dom** | 6.4.0 | DOM断言扩展 |
| **@playwright/test** | 1.45.0 | E2E测试 |
| **jsdom** | 24.0.0 | DOM模拟 |

### 2.7 代码质量

| 技术 | 版本 | 用途 |
|------|------|------|
| **ESLint** | 8.55.0 | 代码检查 |
| **@typescript-eslint** | 7.0.0 | TS规则 |
| **eslint-plugin-react-hooks** | 4.6.0 | Hook规则 |
| **Prettier** | 3.2.0 | 代码格式化 |

### 2.8 前端目录结构

```
frontend/
├── src/
│   ├── components/      # React组件
│   │   ├── App.tsx      # 主应用
│   │   ├── LoginPage.tsx
│   │   ├── FleetDashboard.tsx
│   │   ├── DeviceManager.tsx
│   │   ├── DeviceDetail.tsx
│   │   ├── AITeamDashboard.tsx
│   │   ├── WorkOrderBoard.tsx
│   │   ├── RuleManager.tsx
│   │   ├── NotificationCenter.tsx
│   │   ├── ReportCenter.tsx
│   │   ├── ROIDashboard.tsx
│   │   ├── SystemStatus.tsx
│   │   ├── UserManager.tsx
│   │   ├── lazyRoutes.tsx
│   │   └── UI/          # 通用UI组件
│   ├── hooks/           # 自定义Hooks
│   │   └── useWebSocket.ts
│   ├── lib/             # 工具库
│   │   ├── performance.ts
│   │   ├── useSwipe.ts
│   │   └── security.ts
│   ├── i18n/            # 国际化
│   ├── types/           # TypeScript类型
│   │   ├── api.ts
│   │   └── index.ts
│   ├── utils/           # 工具函数
│   ├── test/            # 测试配置
│   ├── main.tsx         # 入口文件
│   └── index.tsx
├── e2e/                 # E2E测试
│   ├── auth/
│   ├── devices/
│   ├── telemetry/
│   ├── alerts/
│   ├── reports/
│   ├── ai/
│   └── i18n/
├── vite.config.ts       # Vite配置
├── vitest.config.ts     # Vitest配置
├── playwright.config.ts # Playwright配置
├── package.json         # 依赖定义
└── tsconfig.json        # TS配置
```

---

## 3. 第三方依赖说明

### 3.1 后端核心依赖详解

#### Gin Web框架
- **用途**: HTTP路由与中间件
- **特性**: 轻量、高性能、中间件链
- **配置**: 默认运行模式可通过环境变量切换

#### PostgreSQL
- **用途**: 主数据存储
- **连接池**: 
  - MaxOpenConns: 25
  - MaxIdleConns: 5
  - ConnMaxLifetime: 5分钟
- **数据表**: devices, users, telemetry_data, rules, alerts, work_orders, notifications, blackbox_events, reports, tenants, roles, permissions

#### Redis缓存
- **用途**: 
  - 会话缓存
  - API响应缓存
  - 限流计数器
  - WebSocket状态
- **Fallback**: 内存缓存（Redis不可用时）
- **缓存前缀**: `iai:`
- **内存限制**: 100MB

#### JWT认证
- **算法**: HS256
- **AccessToken**: 15分钟有效期
- **RefreshToken**: 7天有效期
- **Claims**: user_id, username, role, tenant_id, token_version

### 3.2 前端核心依赖详解

#### Vite构建工具
- **优势**: 
  - 开发服务器: 毫秒级冷启动
  - HMR: 即时热更新
  - 生产构建: Rollup优化
- **配置**: 
  - 代码分割: react-core, charts, icons, graph, vendor
  - CSS分割: 按组件分割
  - 目标: ES2020

#### Tailwind CSS
- **配置**: JIT模式
- **断点**: sm(640px), md(768px), lg(1024px)
- **主题**: Slate色系深色主题

#### Recharts图表
- **使用**: LineChart, BarChart, AreaChart
- **数据格式**: 数组对象 `{name, value}`

---

## 4. 版本兼容性

### 4.1 Go版本兼容

| Go版本 | 兼容状态 | 说明 |
|--------|----------|------|
| 1.25.x | ✅ 完全兼容 | 当前使用版本 |
| 1.24.x | ⚠️ 可能兼容 | 需测试验证 |
| 1.23.x | ⚠️ 可能兼容 | 需测试验证 |
| 1.22.x及以下 | ❌ 不兼容 | 缺少新特性 |

### 4.2 Node.js版本兼容

| Node版本 | 兼容状态 | 说明 |
|----------|----------|------|
| 20.x | ✅ 完全兼容 | 推荐 LTS版本 |
| 18.x | ✅ 完全兼容 | LTS版本 |
| 16.x | ⚠️ 部分兼容 | 可能缺少某些特性 |
| 14.x及以下 | ❌ 不兼容 | 不支持 |

### 4.3 PostgreSQL版本兼容

| PG版本 | 兼容状态 |
|--------|----------|
| 15.x | ✅ 推荐 |
| 14.x | ✅ 兼容 |
| 13.x | ⚠️ 基本兼容 |
| 12.x及以下 | ❌ 不兼容 |

### 4.4 Redis版本兼容

| Redis版本 | 兼容状态 |
|-----------|----------|
| 7.x | ✅ 推荐 |
| 6.x | ✅ 兼容 |
| 5.x | ⚠️ 基本兼容 |

### 4.5 浏览器兼容

| 浏览器 | 最低版本 | 说明 |
|--------|----------|------|
| Chrome | 90+ | 完全支持 |
| Firefox | 90+ | 完全支持 |
| Safari | 14+ | 完全支持 |
| Edge | 90+ | 完全支持 |
| IE | ❌ | 不支持 |

---

## 5. 开发环境配置

### 5.1 后端开发环境

```bash
# 必需环境变量
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
JWT_SECRET=your-secret-key
REDIS_URL=redis://host:6379
ADMIN_PASSWORD=admin-secret

# 可选环境变量
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
CACHE_ENABLED=true
CACHE_PREFIX=iai:
ENVIRONMENT=development
WS_COMPRESSION_ENABLED=true
```

### 5.2 前端开发环境

```bash
# 安装依赖
cd frontend && npm install

# 开发模式
npm run dev          # 启动开发服务器 (端口3000)

# 构建
npm run build        # 生产构建
npm run build:analyze # 构建分析

# 测试
npm run test         # 单元测试
npm run test:coverage # 覆盖率
npm run test:e2e     # E2E测试

# 代码质量
npm run lint         # ESLint检查
npm run format       # Prettier格式化
npm run typecheck    # TypeScript检查
```

### 5.3 代理配置

Vite开发服务器代理配置:

```typescript
proxy: {
  '/api': { target: 'http://localhost:8080' },
  '/ws': { target: 'ws://localhost:8080', ws: true },
  '/health': { target: 'http://localhost:8080' },
  '/docs': { target: 'http://localhost:8080' }
}
```

---

## 6. 技术选型理由

### 6.1 后端选型

| 技术 | 选型理由 |
|------|----------|
| Go | 高并发性能、编译速度快、部署简单 |
| Gin | 轻量级、API友好、中间件丰富 |
| PostgreSQL | 企业级稳定性、JSON支持、丰富扩展 |
| Redis | 高性能缓存、限流计数器、发布订阅 |
| JWT | 无状态认证、跨服务友好 |
| Prometheus | 标准化指标、Grafana集成 |
| OpenTelemetry | 统一追踪标准、云原生友好 |

### 6.2 前端选型

| 技术 | 选型理由 |
|------|----------|
| React | 组件化、生态丰富、团队熟悉 |
| TypeScript | 类型安全、IDE支持、重构友好 |
| Vite | 快速开发体验、现代化构建 |
| Tailwind CSS | 原子化CSS、快速开发、一致性 |
| Recharts | React原生、声明式配置 |
| Vitest | Vite集成、快速测试 |
| Playwright | 跨浏览器、可靠E2E |

---

## 7. 未来技术演进

### 7.1 短期演进 (1-3个月)

- [ ] 后端: 引入GraphQL端点
- [ ] 前端: 引入状态管理库(Zustand)
- [ ] 数据库: 添加读写分离
- [ ] 监控: Grafana仪表板完善

### 7.2 中期演进 (3-6个月)

- [ ] 后端: Kubernetes部署优化
- [ ] 前端: PWA完整支持
- [ ] 数据库: PostgreSQL分库分表
- [ ] 缓存: Redis集群部署

### 7.3 长期演进 (6-12个月)

- [ ] 微服务拆分
- [ ] GraphQL API优先
- [ ] 实时流处理(Kafka)
- [ ] 容器化全面改造

---

## 附录

### A. 依赖版本锁定策略

- 后端: go.mod精确版本锁定
- 前端: package.json使用精确版本(无^/~)

### B. 安全依赖更新

- 定期检查依赖安全漏洞
- 使用 `npm audit` 和 `govulncheck`
- 依赖更新: 每月检查，按需更新

---

*文档维护: 开发团队*  
*最后审核: 2026-05-14*