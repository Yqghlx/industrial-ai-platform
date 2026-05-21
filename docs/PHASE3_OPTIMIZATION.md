# Phase 3 后续优化任务清单

**生成日期**: 2026-05-13
**项目状态**: Phase 3 已100%完成，但存在多个集成和实现缺口

---

## 📊 优化任务总览

| 分类 | 任务数 | 优先级 |
|------|--------|--------|
| 🔴 P0 - 关键修复 | 4 | 立即处理 |
| 🟡 P1 - 功能完善 | 6 | 近期处理 |
| 🟢 P2 - 代码质量 | 5 | 中期处理 |
| 🔵 P3 - 文档更新 | 4 | 可延后 |
| **总计** | **19** | - |

---

## 🔴 P0 - 关键修复 (数据安全/功能缺口)

### P0-1: 多租户数据隔离未实现 ⚠️ **最高优先级**

**问题描述**: 
- 数据模型已添加 `tenant_id` 字段
- 中间件 `TenantIsolation` 已定义但未实际过滤数据
- **所有 Repository 的 SQL 查询都没有 tenant_id 过滤**，存在数据泄露风险

**涉及文件**:
```
backend/internal/repository/device_repo.go      - Create/List/GetByID/Update/Delete 全部缺少 tenant_id
backend/internal/repository/telemetry_repo.go   - Insert/GetByDeviceID/GetLatest/GetStats 缺少 tenant_id
backend/internal/repository/rule_repo.go        - 缺少 tenant_id 过滤
backend/internal/repository/work_order.go       - 缺少 tenant_id 过滤
backend/internal/repository/notification.go     - 缺少 tenant_id 过滤
```

**修复方案**:
```go
// 示例: device_repo.go List 方法修改
func (r *DeviceRepository) List(ctx context.Context, tenantID string, page, pageSize int) ([]model.Device, int, error) {
    whereClause := "WHERE tenant_id = $1"
    args := []interface{}{tenantID}
    
    // Count with tenant filter
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM devices %s", whereClause)
    err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    
    // Query with tenant filter
    query := fmt.Sprintf(`
        SELECT id, name, type, location, status, description, created_at, updated_at, tenant_id
        FROM devices %s ORDER BY created_at DESC LIMIT $2 OFFSET $3
    `, whereClause)
    args = append(args, pageSize, offset)
    ...
}
```

**验收标准**:
- 所有 Repository 方法都接受 tenantID 参数
- SQL 查询包含 `WHERE tenant_id = ?` 过滤
- 单元测试覆盖多租户隔离场景
- 超级管理员 (admin role) 可以绕过租户隔离

---

### P0-2: RBAC 权限中间件未应用到路由

**问题描述**:
- `PermissionRequired` 中间件已定义
- **server.go 中所有业务路由都没有使用权限检查中间件**
- 用户可以访问任何 API，仅需登录即可

**涉及文件**:
```
backend/internal/handler/server.go  - setupRoutes() 方法需要添加权限中间件
```

**修复方案**:
```go
// server.go setupRoutes 修改示例
auth := s.router.Group("/api/v1")
auth.Use(middleware.AuthRequired(s.jwtSecret))
auth.Use(middleware.TenantIsolation())  // 添加租户隔离
{
    // Device endpoints - 需要 devices:read 权限
    devices := auth.Group("/devices")
    devices.Use(middleware.PermissionRequired(s.rbacSvc, "devices", "read"))
    {
        devices.GET("", s.listDevices)
        devices.GET("/:id", s.getDevice)
    }
    
    // Device management - 需要 devices:manage 权限
    devicesManage := auth.Group("/devices")
    devicesManage.Use(middleware.PermissionRequired(s.rbacSvc, "devices", "manage"))
    {
        devicesManage.POST("", s.createDevice)
        devicesManage.PUT("/:id", s.updateDevice)
        devicesManage.DELETE("/:id", s.deleteDevice)
    }
    
    // Rules - 需要 rules:read/manage
    rules := auth.Group("/rules")
    rules.Use(middleware.PermissionRequired(s.rbacSvc, "rules", "read"))
    rules.GET("", s.listRules)
    // ...
}
```

**验收标准**:
- 所有敏感 API 都有权限检查
- 无权限返回 403 Forbidden
- 权限配置表与 API 映射文档化

---

### P0-3: Import 路径不一致导致编译失败

**问题描述**:
- `go.mod` 定义模块为 `github.com/industrial-ai/platform`
- 但 4 个文件使用了错误路径 `github.com/yqghlx/industrial-ai-platform/backend/internal/...`

**涉及文件**:
```
backend/internal/repository/tenant_repo.go    - 错误 import
backend/internal/service/tenant_service.go   - 错误 import
backend/internal/handler/tenant_handler.go   - 错误 import
```

**修复方案**:
```go
// 修改前
import "github.com/yqghlx/industrial-ai-platform/backend/internal/model"

// 修改后
import "github.com/industrial-ai/platform/internal/model"
```

**验收标准**:
- 所有 import 路径一致
- `go build ./...` 成功
- `go test ./...` 成功

---

### P0-4: go.sum 缺失依赖导致测试无法运行

**问题描述**:
```
missing go.sum entry for go.mod file; to add it:
    go mod download github.com/DATA-DOG/go-sqlmock
missing go.sum entry for module providing package github.com/redis/go-redis/v9
```

**修复方案**:
```bash
cd backend
go mod tidy
go mod download github.com/DATA-DOG/go-sqlmock
go mod download github.com/redis/go-redis/v9
```

**验收标准**:
- `go test ./...` 全部通过
- 所有 85 个测试用例执行成功

---

## 🟡 P1 - 功能完善

### P1-1: 前端租户上下文缺失

**问题描述**:
- `AuthContext.tsx` 只有 `user` 和 `isAdmin`
- **缺少 `tenant_id`, `tenant_name`, `permissions` 字段**
- 无法实现租户切换和权限控制 UI

**修复方案**:
```typescript
// AuthContext.tsx 扩展
interface User {
  id: number;
  username: string;
  role: string;
  tenantId?: string;       // 新增
  tenantName?: string;     // 新增
  permissions: string[];   // 新增
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  tenantId: string | null;       // 新增
  permissions: string[];         // 新增
  hasPermission: (resource: string, action: string) => boolean;  // 新增
  switchTenant: (tenantId: string) => Promise<void>;  // 新增 (超级管理员)
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
  isAdmin: boolean;
}
```

**涉及文件**:
```
frontend/src/components/AuthContext.tsx
frontend/src/lib/api.ts           - 登录响应需要返回 tenantId 和 permissions
frontend/src/types/api.ts         - User 类型扩展
```

---

### P1-2: 前端权限控制 UI 组件

**问题描述**:
- Sidebar 只检查 `isAdmin`，未检查细粒度权限
- 组件无法根据权限隐藏/禁用功能

**修复方案**:
```typescript
// 新建 frontend/src/components/PermissionGuard.tsx
interface PermissionGuardProps {
  resource: string;
  action: string;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function PermissionGuard({ resource, action, children, fallback }: PermissionGuardProps) {
  const { hasPermission } = useAuth();
  
  if (!hasPermission(resource, action)) {
    return fallback || null;
  }
  
  return <>{children}</>;
}

// 使用示例
<PermissionGuard resource="devices" action="manage">
  <button onClick={handleDelete}>删除设备</button>
</PermissionGuard>
```

**涉及文件**:
```
frontend/src/components/PermissionGuard.tsx (新建)
frontend/src/components/Sidebar.tsx         - 使用权限控制菜单项
frontend/src/components/UserManager.tsx     - 使用权限控制
```

---

### P1-3: JWT Claims 扩展包含 tenant_id

**问题描述**:
- JWT token 未包含 `tenant_id`
- 登录时后端未返回用户权限列表

**涉及文件**:
```
backend/internal/middleware/jwt_helpers.go
backend/internal/service/auth_service.go
backend/internal/handler/auth_handler.go
```

**修复方案**:
```go
// JWT Claims 扩展
type Claims struct {
    UserID      int      `json:"user_id"`
    Username    string   `json:"username"`
    Role        string   `json:"role"`
    TenantID    string   `json:"tenant_id"`     // 新增
    Permissions []string `json:"permissions"`   // 新增
    jwt.RegisteredClaims
}

// 登录响应扩展
type LoginResponse struct {
    Token       string   `json:"token"`
    User        User     `json:"user"`
    TenantID    string   `json:"tenant_id"`     // 新增
    Permissions []string `json:"permissions"`   // 新增
}
```

---

### P1-4: Edge SDK 与后端集成测试

**问题描述**:
- SDK 功能完整，但未验证与实际后端的通信
- 缺少 SDK 连接、遥测发送、WebSocket 接收的端到端测试

**修复方案**:
```bash
# 新建 sdk/IntegrationTests 目录
# 测试场景:
1. SDK 连接后端健康检查
2. 设备自动注册
3. 遥测数据发送和持久化验证
4. WebSocket 实时消息接收
5. 多租户隔离验证 (tenant_id 传递)
```

**验收标准**:
- SDK 示例程序能连接运行中的后端
- 遥测数据出现在数据库中
- WebSocket 能接收告警推送

---

### P1-5: 数据库迁移脚本统一和验证

**问题描述**:
- 存在两套迁移脚本目录:
  - `backend/migrations/` (包含 000001-000006)
  - `backend/internal/database/migrations/` (包含 000001-000003)
- **目录结构混乱，执行顺序不明确**

**修复方案**:
```bash
# 统一迁移脚本目录
# 决策: 使用 backend/internal/database/migrations 作为唯一目录
# 合并 000006_add_tenants.up.sql 到主目录
# 删除 backend/migrations/ 目录
```

**验收标准**:
- 单一迁移目录
- 迁移脚本可按顺序执行
- `migrator.Up()` 成功创建所有表和索引

---

### P1-6: 租户设备限额验证

**问题描述**:
- Tenant 模型有 `max_devices` 字段
- 创建设备时未验证租户限额

**修复方案**:
```go
// device_service.go 添加限额检查
func (s *DeviceService) Create(ctx context.Context, tenantID string, device *model.Device) error {
    // 获取租户信息
    tenant, err := s.tenantRepo.GetByID(tenantID)
    if err != nil {
        return err
    }
    
    // 检查设备数量
    count, err := s.deviceRepo.CountByTenant(ctx, tenantID)
    if count >= tenant.MaxDevices {
        return ErrTenantDeviceLimitExceeded
    }
    
    // 创建设备...
}
```

---

## 🟢 P2 - 代码质量

### P2-1: Repository 单元测试覆盖多租户

**问题描述**:
- 现有测试未覆盖多租户隔离场景
- 需要添加 tenant_id 过滤的测试

**修复方案**:
```go
// device_repo_test.go 新增测试
func TestDeviceRepository_TenantIsolation(t *testing.T) {
    // 创建两个租户的设备
    repo.Create(ctx, &Device{ID: "d1", TenantID: "tenant-a"})
    repo.Create(ctx, &Device{ID: "d2", TenantID: "tenant-b"})
    
    // 验证租户 A 只能看到自己的设备
    devices, _ := repo.List(ctx, "tenant-a", 1, 10)
    assert.Equal(t, 1, len(devices))
    assert.Equal(t, "d1", devices[0].ID)
    
    // 验证租户 B 只能看到自己的设备
    devices, _ = repo.List(ctx, "tenant-b", 1, 10)
    assert.Equal(t, 1, len(devices))
    assert.Equal(t, "d2", devices[0].ID)
}
```

---

### P2-2: RBAC 权限种子数据验证

**问题描述**:
- `InitializeDefaultRBAC` 创建默认角色和权限
- 缺少验证确保数据正确创建

**修复方案**:
```go
// 添加 RBAC 初始化验证测试
func TestRBACService_DefaultRolesAndPermissions(t *testing.T) {
    svc := NewRBACService(...)
    svc.InitializeDefaultRBAC(ctx)
    
    // 验证角色存在
    roles, _ := svc.ListRoles(ctx, "")
    assert.Contains(t, roles, Role{Name: "admin"})
    assert.Contains(t, roles, Role{Name: "operator"})
    assert.Contains(t, roles, Role{Name: "viewer"})
    
    // 验证权限分配
    adminPerms, _ := svc.GetRolePermissions(ctx, adminRoleID)
    assert.Greater(t, len(adminPerms), 10)
}
```

---

### P2-3: WebSocket 压缩测试未导入包

**问题描述**:
```
pkg/wscompression/compressor_test.go:8:2: "github.com/gorilla/websocket" imported and not used
```

**修复方案**:
移除未使用的 import 或添加使用该包的测试代码。

---

### P2-4: 错误处理标准化

**问题描述**:
- 部分代码返回 `errors.New()` 而非自定义错误类型
- 缺少统一的错误码定义

**修复方案**:
```go
// backend/internal/errors/errors.go (新建)
package errors

type ErrorCode string

const (
    ErrTenantNotFound        ErrorCode = "TENANT_NOT_FOUND"
    ErrTenantDeviceLimit     ErrorCode = "TENANT_DEVICE_LIMIT"
    ErrPermissionDenied      ErrorCode = "PERMISSION_DENIED"
    ErrTenantIsolation       ErrorCode = "TENANT_ISOLATION"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Details interface{}
}

func (e *AppError) Error() string {
    return e.Message
}
```

---

### P2-5: SQL 注入参数化检查

**问题描述**:
- 部分 Repository 使用 `fmt.Sprintf` 构建 SQL
- 需确保所有动态值使用参数化查询

**涉及文件**:
```
backend/internal/repository/telemetry_repo.go - List 方法使用 fmt.Sprintf
backend/internal/repository/work_order.go     - List 方法使用 fmt.Sprintf
```

**修复方案**:
确保所有动态值通过 `$1, $2, $3...` 参数传递，而非字符串拼接。

---

## 🔵 P3 - 文档更新

### P3-1: API 权限矩阵文档

**需要创建**:
```
docs/API_PERMISSIONS.md

内容:
- 所有 API 端点列表
- 所需权限 (resource:action)
- 所需角色
- 租户隔离说明
```

---

### P3-2: 多租户架构文档

**需要创建**:
```
docs/MULTI_TENANT_ARCHITECTURE.md

内容:
- 租户隔离实现原理
- 数据模型 tenant_id 字段说明
- 超级管理员权限说明
- 租户切换流程
- 最佳实践和安全考虑
```

---

### P3-3: RBAC 配置指南

**需要更新**:
```
docs/RBAC_GUIDE.md (已存在则更新，不存在则创建)

内容:
- 默认角色和权限列表
- 如何创建自定义角色
- 如何分配角色给用户
- 权限检查中间件使用说明
```

---

### P3-4: Edge SDK 集成指南补充

**需要补充**:
```
sdk/IndustrialAI.EdgeSDK/README.md

补充内容:
- tenant_id 配置说明 (多租户场景)
- 与后端 API 版本兼容性
- 错误处理最佳实践
- 生产环境部署建议
```

---

## 📅 建议执行顺序

| 周次 | 任务 | 说明 |
|------|------|------|
| Week 1 | P0-3, P0-4 | 先修复编译问题，让测试能运行 |
| Week 1 | P0-1 | 修复数据安全漏洞，最高优先级 |
| Week 2 | P0-2 | 应用 RBAC 中间件到所有路由 |
| Week 2 | P1-3 | 扩展 JWT Claims |
| Week 3 | P1-1, P1-2 | 前端租户和权限 UI |
| Week 3 | P1-5, P1-6 | 迁移脚本统一和限额验证 |
| Week 4 | P1-4 | Edge SDK 集成测试 |
| Week 4 | P2-1, P2-2 | 单元测试补充 |
| Week 5 | P2-3, P2-4, P2-5 | 代码质量改进 |
| Week 5 | P3-1 ~ P3-4 | 文档更新 |

---

## 🎯 验收清单

Phase 3 后续优化完成后，应满足:

- [ ] 所有 Repository SQL 包含 tenant_id 过滤
- [ ] 所有 API 路由有权限检查
- [ ] 前端显示用户租户和权限
- [ ] JWT 包含 tenant_id 和 permissions
- [ ] Edge SDK 能成功连接后端
- [ ] go test ./... 全部通过 (85+ tests)
- [ ] 迁移脚本可顺序执行
- [ ] 文档完整覆盖多租户和 RBAC

---

**生成工具**: Hermes Agent
**文档版本**: 1.0.0