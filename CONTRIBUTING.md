# Contributing to Industrial AI Platform

感谢您有兴趣为工业AI平台做出贡献！本文档将帮助您了解如何参与项目开发。

## 📋 目录

- [代码贡献流程](#代码贡献流程)
- [开发环境设置](#开发环境设置)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [测试要求](#测试要求)

---

## 代码贡献流程

### 1. Fork 和 Clone

```bash
# Fork 项目到您的 GitHub 账户
# 然后 clone 到本地
git clone https://github.com/YOUR_USERNAME/industrial-ai-platform.git
cd industrial-ai-platform

# 添加上游仓库
git remote add upstream https://github.com/Yqghlx/industrial-ai-platform.git
```

### 2. 创建分支

```bash
# 创建功能分支
git checkout -b feature/your-feature-name

# 或修复分支
git checkout -b fix/issue-number
```

### 3. 开发和测试

```bash
# 后端开发
cd backend
go mod download
go test ./...

# 前端开发
cd frontend
npm install
npm test
npm run typecheck
```

---

## 开发环境设置

### 必需工具

| 工具 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 后端开发 |
| Node.js | 20+ | 前端开发 |
| Docker | 最新 | 容器化部署 |
| PostgreSQL | 15+ | 数据库 |
| Redis | 7+ | 缓存 |

### 环境变量

参考 `.env.example` 文件配置必要的环境变量：

```bash
# 复制示例文件
cp .env.example .env

# 必需变量
DATABASE_URL=postgres://...
JWT_SECRET=your-secret-key-min-32-characters
LLM_API_KEY=your-api-key
```

---

## 代码规范

### Go 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 指南
- 使用 `gofmt` 格式化代码
- 函数和变量命名使用驼峰式
- 添加必要的注释和文档

```go
// Example: 函数注释规范
// CreateUser creates a new user with the given parameters.
// Returns the created user and any error encountered.
func CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
    // ...
}
```

### TypeScript 代码规范

- 使用 ESLint 和 Prettier
- 所有变量和函数必须有类型定义
- React 组件使用函数式组件 + Hooks
- 国际化文本使用 i18n 系统

```typescript
// Example: 类型定义规范
interface User {
  id: string;
  username: string;
  role: 'admin' | 'operator' | 'viewer';
}
```

---

## 提交规范

使用 Conventional Commits 规范：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### 类型 (type)

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug修复 |
| docs | 文档更新 |
| style | 代码格式（不影响逻辑） |
| refactor | 重构 |
| test | 测试相关 |
| chore | 构建/工具相关 |

### 示例

```bash
feat(auth): add JWT token refresh mechanism
fix(handler): resolve memory leak in WebSocket connection
docs(api): update OpenAPI specification
test(service): add unit tests for alert evaluation
```

---

## Pull Request 流程

### 1. 提交前检查

```bash
# 运行所有测试
go test ./...
npm test

# 代码格式检查
gofmt -s -w .
npm run lint

# 类型检查
npm run typecheck
```

### 2. 创建 Pull Request

- PR 标题遵循提交规范
- 描述清楚改动内容和原因
- 关联相关 Issue（如有）
- 等待 CI 通过和代码审查

### 3. PR 要求

- ✅ 所有测试通过
- ✅ 代码覆盖率不降低
- ✅ 无 ESLint/编译错误
- ✅ 有必要的文档更新
- ✅ 无安全漏洞引入

---

## 测试要求

### 后端测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包测试
go test ./internal/service -v

# 测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 前端测试

```bash
# 单元测试
npm test

# E2E测试
npm run test:e2e

# 测试覆盖率
npm run test:coverage
```

### 测试命名规范

```go
// Go 测试
func TestUserService_Create(t *testing.T) {}
func TestUserService_Create_Error(t *testing.T) {}
```

```typescript
// TypeScript 测试
describe('UserService', () => {
  it('should create user successfully', () => {});
  it('should handle error when creating user', () => {});
});
```

---

## 📚 相关文档

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - 系统架构
- [API文档](docs/openapi.yaml) - OpenAPI规范
- [SECURITY.md](SECURITY.md) - 安全策略

---

## 💬 获取帮助

- GitHub Issues: 提交问题或建议
- Pull Requests: 代码贡献
- Discussions: 讨论想法和方案

感谢您的贡献！🎉