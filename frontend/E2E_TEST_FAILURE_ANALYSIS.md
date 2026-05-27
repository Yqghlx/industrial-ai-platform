# E2E 测试失败分析报告

## 问题诊断

### 测试状态
- **通过**: 32 个测试
- **失败**: 298 个测试
- **失败类型**: 登录相关测试（成功登录、登出、跳转）

### 根本原因

| 问题 | 详情 |
|------|------|
| **密码不匹配** | 测试 fixtures 使用 `Admin@123456`，但实际 admin 密码是 `Admin@TPby8q1dmPk`（来自环境变量 ADMIN_PASSWORD） |
| **operator 用户不存在** | 数据库只有 `admin` 和 `testuser`，没有 `operator` 用户 |
| **viewer 用户不存在** | 同上 |

### 验证证据

1. **后端日志显示 401 Unauthorized**:
```
"status":401,"path":"/api/v1/auth/login"
```

2. **页面快照显示错误信息**:
```yaml
- generic [ref=e34]: 用户名或密码错误
```

3. **数据库用户查询**:
```sql
SELECT username FROM users;
-- 结果: admin, testuser (没有 operator)
```

4. **实际密码验证**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"username":"admin","password":"Admin@TPby8q1dmPk"}'
# 成功返回 token

curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"username":"admin","password":"Admin@123456"}'
# 返回 {"error":"Invalid credentials"}
```

## 修复方案

### 方案 A: 创建测试用户（推荐）

在数据库中创建 operator 和 viewer 测试用户：

```bash
# 1. 设置环境变量
export ADMIN_PASSWORD="Admin@TPby8q1dmPk"  # 从 docker exec industrial-ai-backend printenv ADMIN_PASSWORD 获取

# 2. 运行初始化脚本
./scripts/e2e-test-users.sh
```

### 方案 B: 使用环境变量运行测试

```bash
# 设置正确的密码
export E2E_ADMIN_PASSWORD="Admin@TPby8q1dmPk"

# 运行测试（只使用 admin 用户）
npx playwright test e2e/auth/login.spec.ts --grep "成功登录 - 管理员"
```

### 方案 C: 修改 playwright.config.ts

添加测试前的用户初始化：

```typescript
webServer: {
  command: 'npm run dev && ../scripts/e2e-test-users.sh',
  url: 'http://localhost:3000',
  reuseExistingServer: true,
  timeout: 60000,
},
```

## 执行步骤

1. **立即修复** - 创建测试用户:
```bash
cd /Users/yqgmac/yqg/project/industrial-ai-platform

# 获取 admin 密码
ADMIN_PASSWORD=$(docker exec industrial-ai-backend printenv ADMIN_PASSWORD)

# 获取 admin token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"${ADMIN_PASSWORD}\"}" | jq -r '.token')

# 创建 operator 用户
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"username":"operator","password":"Operator@123","email":"operator@industrial.ai","role":"operator"}'

# 创建 viewer 用户
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"username":"viewer","password":"Viewer@123","email":"viewer@industrial.ai","role":"viewer"}'
```

2. **运行测试验证**:
```bash
export E2E_ADMIN_PASSWORD="${ADMIN_PASSWORD}"
cd frontend
npx playwright test e2e/auth/login.spec.ts
```

## 长期建议

1. **统一密码管理**: 将测试密码写入 `.env.e2e` 文件，与 docker-compose 共享
2. **文档更新**: 更新 `frontend/e2e/README.md` 中的测试用户说明
3. **CI/CD 集成**: 在测试前自动创建测试用户

## 相关文件

- `/Users/yqgmac/yqg/project/industrial-ai-platform/frontend/e2e/fixtures/test-fixtures.ts` - 测试密码配置
- `/Users/yqgmac/yqg/project/industrial-ai-platform/frontend/playwright.config.ts` - E2E 测试配置
- `/Users/yqgmac/yqg/project/industrial-ai-platform/.env` - 环境变量（包含 ADMIN_PASSWORD）
- `/Users/yqgmac/yqg/project/industrial-ai-platform/scripts/e2e-test-users.sh` - 用户初始化脚本（新创建）

---
**分析日期**: 2026-05-26
**分析工具**: Hermes Agent