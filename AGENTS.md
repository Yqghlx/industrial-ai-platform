# Repository Guidelines

本文档为工业 AI 平台的贡献者指南，帮助你快速了解项目结构、开发规范和提交流程。

## Project Structure & Module Organization

```
industrial-ai-platform/
├── backend/                # Go 后端服务
│   ├── internal/
│   │   ├── handler/        # HTTP 处理器 (API 层)
│   │   ├── service/        # 业务逻辑层
│   │   ├── repository/     # 数据访问层
│   │   ├── model/          # 数据模型
│   │   ├── middleware/     # 中间件 (认证、限流等)
│   │   ├── config/         # 配置管理
│   │   └── security/       # 安全模块
│   ├── pkg/                # 公共工具包
│   ├── docs/               # OpenAPI 文档
│   └── main.go
├── frontend/               # React 前端应用
│   ├── src/
│   │   ├── components/     # React 组件 (懒加载)
│   │   ├── lib/            # API 客户端、工具函数
│   │   ├── hooks/          # 自定义 Hooks
│   │   ├── types/          # TypeScript 类型定义
│   │   ├── i18n/           # 国际化配置
│   │   └── utils/          # 工具函数
│   └── tests/              # 测试文件
├── edge/                   # C# 边缘模拟器
├── scripts/                # 部署和工具脚本
├── docker/                 # Docker 配置
├── kubernetes/             # K8s 部署配置
└── docs/                   # 项目文档
```

**架构分层**: `Handler → Service → Repository → Database`

## Build, Test, and Development Commands

### 后端 (Go)

```bash
cd backend

# 开发
go run main.go                    # 启动开发服务器
make build                        # 构建二进制文件

# 测试
make test                         # 运行所有测试 + 覆盖率
make test-unit                    # 单元测试
make coverage-html                # 生成 HTML 覆盖率报告

# 代码质量
make lint                         # golangci-lint 检查
make fmt                          # 格式化代码
make vulncheck                    # 安全漏洞扫描
```

### 前端 (React + TypeScript)

```bash
cd frontend

# 开发
npm run dev                       # 启动 Vite 开发服务器 (http://localhost:5173)
npm run build                     # 生产构建
npm run preview                   # 预览生产构建

# 测试
npm test                          # Vitest 单元测试
npm run test:coverage             # 测试覆盖率报告
npm run test:e2e                  # Playwright E2E 测试

# 代码质量
npm run lint                      # ESLint 检查
npm run format                    # Prettier 格式化
npm run typecheck                 # TypeScript 类型检查
```

### Docker

```bash
docker-compose up -d              # 启动所有服务
docker-compose logs -f            # 查看日志
docker-compose --profile simulator up -d  # 包含边缘模拟器
```

## Coding Style & Naming Conventions

### Go 后端

- **格式化**: 使用 `gofmt` 和 `goimports`
- **命名**: 驼峰式 (`CreateUser`, `getUserByID`)
- **注释**: 函数必须有文档注释，说明功能、参数和返回值
- **错误处理**: 不忽略错误，使用 `pkg/errors` 包装错误上下文
- **接口**: 在 Service 层定义接口，Repository 层实现

```go
// CreateUser creates a new user with the given parameters.
// Returns the created user and any error encountered.
func (s *userService) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
    // 实现...
}
```

### TypeScript 前端

- **格式化**: Prettier (自动格式化)
- **命名**: 
  - 组件: PascalCase (`LoginPage.tsx`)
  - 变量/函数: camelCase (`getUserInfo`)
  - 常量: UPPER_SNAKE_CASE (`API_BASE_URL`)
- **类型**: 所有变量和函数必须有类型定义
- **组件**: 函数式组件 + Hooks，使用 `React.lazy()` 懒加载
- **国际化**: 所有用户可见文本使用 i18n

```typescript
// 组件命名
export const LoginPage: React.FC = () => { ... }

// 类型定义
interface User {
  id: string;
  username: string;
  role: 'admin' | 'operator' | 'viewer';
}
```

## Testing Guidelines

### 后端测试

- **框架**: Go 原生 `testing` 包 + `testify`
- **覆盖率目标**: 70%+
- **测试文件**: `*_test.go`，与源文件同目录
- **命名规范**: `Test<FunctionName>_<Scenario>`，如 `TestAuthService_Create_Success`

```bash
make test                    # 运行测试
make coverage-html           # 覆盖率报告
```

### 前端测试

- **框架**: Vitest + React Testing Library
- **E2E**: Playwright
- **测试文件**: `*.test.ts` / `*.test.tsx`
- **命名规范**: `describe('模块名', () => { it('should ...', () => {}) })`

```bash
npm test                     # 单元测试
npm run test:e2e             # E2E 测试
```

## Commit & Pull Request Guidelines

### 提交信息规范 (Conventional Commits)

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**类型 (type)**:
| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(auth): add JWT refresh mechanism` |
| `fix` | Bug 修复 | `fix(handler): resolve WebSocket memory leak` |
| `refactor` | 重构 | `refactor: replace map[string]interface{} with concrete types` |
| `test` | 测试 | `test(service): add unit tests for alert evaluation` |
| `docs` | 文档 | `docs(api): update OpenAPI specification` |
| `chore` | 构建/工具 | `chore: update dependencies` |
| `style` | 格式化 | `style: format code with prettier` |

### Pull Request 流程

1. **创建分支**: `feature/<name>` 或 `fix/<issue-number>`
2. **本地验证**:
   ```bash
   # 后端
   cd backend && make test && make lint
   
   # 前端
   cd frontend && npm test && npm run lint && npm run typecheck
   ```
3. **提交 PR**: 标题遵循提交规范，描述改动内容和原因
4. **CI 检查**: 所有测试必须通过，代码规范检查必须通过
5. **Code Review**: 至少一位 reviewer 审核通过

### PR 检查清单

- [ ] 所有测试通过
- [ ] 代码覆盖率不降低
- [ ] 无 ESLint / golangci-lint 错误
- [ ] TypeScript 类型检查通过
- [ ] 必要的文档已更新
- [ ] 无安全漏洞引入

## Security & Configuration

### 环境变量

**必需变量** (启动时验证):
```bash
DATABASE_URL=postgres://user:password@host:5432/db?sslmode=disable
JWT_SECRET=your-secret-key-at-least-32-characters
```

**可选变量**:
```bash
LLM_API_KEY=your-dashscope-api-key    # AI 功能
LLM_BASE_URL=https://coding.dashscope.aliyuncs.com/v1
LLM_MODEL=glm-5
```

### 安全要求

- **禁止硬编码敏感信息**: 密码、API Key 等必须通过环境变量配置
- **JWT 认证**: 所有需要认证的 API 必须验证 Token
- **CORS 白名单**: 只允许配置的域名访问
- **限流保护**: 登录、注册等敏感接口有限流保护

## Architecture Overview

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Edge       │────▶│  Backend    │────▶│  Database   │
│  Simulator  │     │  (Go API)   │     │  (TimescaleDB)│
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                          │ WebSocket + REST API
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

**技术栈**:
- **前端**: React 19 + TypeScript (strict) + Vite + TailwindCSS
- **后端**: Go + Gin + PostgreSQL/TimescaleDB
- **AI**: 阿里云百炼 GLM-5 + 规则引擎回退
- **实时通信**: WebSocket + 心跳 + 数据压缩
- **测试**: Vitest + Playwright (前端) + Go testing (后端)
- **CI/CD**: GitHub Actions

## Related Documentation

- [README.md](README.md) - 项目介绍和快速开始
- [CONTRIBUTING.md](CONTRIBUTING.md) - 详细贡献指南
- [SECURITY.md](SECURITY.md) - 安全策略
- [backend/docs/openapi.yaml](backend/docs/openapi.yaml) - API 文档
