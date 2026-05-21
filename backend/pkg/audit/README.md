# 安全审计日志服务

## 概述

安全审计日志服务提供了一个全面的安全审计系统，用于记录和追踪所有安全相关的事件，包括认证、授权、数据访问、管理操作和安全事件。

## 特性

- ✅ **完整的事件类型支持**
  - 认证事件（登录、登出、密码修改）
  - 数据访问记录（读取、写入、删除、导出）
  - 管理操作记录（用户管理、角色管理、配置变更）
  - 安全事件记录（违规、告警、阻断）

- ✅ **异步写入队列**
  - 高性能异步日志记录
  - 可配置队列大小和工作协程数量
  - 批量写入优化

- ✅ **可配置日志级别**
  - 支持 All、Info、Warning、Critical、None 级别
  - 灵活的日志过滤策略

- ✅ **丰富的元数据支持**
  - 记录变更前后的状态
  - 详细的变更追踪
  - 自定义元数据字段

- ✅ **数据导出**
  - JSON 格式导出
  - CSV 格式导出

## 安装

```bash
# 确保已安装依赖
go mod tidy
```

## 数据库迁移

执行数据库迁移以创建 `audit_logs` 表：

```bash
# 使用迁移工具
migrate -path ./internal/database/migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up
```

## 快速开始

### 1. 基础配置

```go
package main

import (
    "context"
    "log"
    
    "github.com/yourorg/industrial-ai-platform/backend/pkg/audit"
    "go.uber.org/zap"
)

func main() {
    // 创建日志记录器
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // 创建仓库实例（需要实现 audit.Repository 接口）
    repo := audit.NewPostgresRepository(db, logger)
    
    // 创建审计日志配置
    config := &audit.Config{
        Enabled:        true,
        LogLevel:       audit.LogLevelAll,
        AsyncEnabled:   true,
        QueueSize:      10000,
        WorkerCount:    3,
        BatchSize:      100,
        BatchTimeout:   5,
        RetentionDays:  90,
        EnableMetadata: true,
    }
    
    // 创建审计日志服务
    auditLogger := audit.NewAuditLogger(repo, logger, config)
    defer auditLogger.Close()
    
    // 使用审计日志服务
    // ...
}
```

### 2. 记录认证事件

```go
// 登录成功
err := auditLogger.LogLogin(ctx, userID, tenantID, sessionID, ipAddress, userAgent, true)

// 登录失败
err := auditLogger.LogLogin(ctx, userID, tenantID, sessionID, ipAddress, userAgent, false)

// 登出
err := auditLogger.LogLogout(ctx, userID, tenantID, sessionID, ipAddress)

// 密码修改
err := auditLogger.LogPasswordChange(ctx, userID, tenantID, ipAddress, true, map[string]interface{}{
    "changed_by": "self",
    "reason": "user_request",
})
```

### 3. 记录数据访问

```go
// 读取数据
err := auditLogger.LogDataAccess(ctx, userID, tenantID, ipAddress,
    "device", "device-001", audit.ActionRead, "Read device data", nil)

// 写入数据
err := auditLogger.LogDataAccess(ctx, userID, tenantID, ipAddress,
    "device", "device-001", audit.ActionWrite, "Update device configuration", map[string]interface{}{
        "fields_updated": []string{"status", "location"},
    })

// 删除数据
err := auditLogger.LogDataAccess(ctx, userID, tenantID, ipAddress,
    "device", "device-001", audit.ActionDelete, "Delete device", nil)
```

### 4. 记录管理操作

```go
// 创建用户
err := auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
    audit.EventAdminUserCreate, "user", "user-002", "Create new user",
    nil,
    map[string]interface{}{"username": "newuser", "role": "user"},
    map[string]interface{}{"username": "newuser"},
    nil)

// 更新用户角色
err := auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
    audit.EventAdminRoleAssign, "user", "user-002", "Assign admin role",
    map[string]interface{}{"role": "user"},
    map[string]interface{}{"role": "admin"},
    map[string]interface{}{"role": "user -> admin"},
    nil)

// 删除用户
err := auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
    audit.EventAdminUserDelete, "user", "user-002", "Delete user",
    map[string]interface{}{"username": "olduser"},
    nil, nil, nil)
```

### 5. 记录安全事件

```go
// 安全违规
err := auditLogger.LogSecurityViolation(ctx, userID, tenantID, ipAddress,
    "unauthorized_access", "Attempted to access restricted resource",
    map[string]interface{}{
        "resource": "/admin/settings",
        "method":   "GET",
    })

// 安全告警
err := auditLogger.LogSecurityAlert(ctx, userID, tenantID, ipAddress,
    "suspicious_activity", "Multiple failed login attempts detected",
    map[string]interface{}{
        "attempt_count": 5,
        "time_window":   "5m",
    })

// 自定义安全事件
err := auditLogger.LogSecurityEvent(ctx, userID, tenantID, ipAddress,
    audit.EventSecurityBlocked, "IP blocked due to brute force",
    audit.SeverityWarning,
    map[string]interface{}{
        "blocked_ip": ipAddress,
        "reason":     "brute_force_attack",
    })
```

### 6. 查询审计日志

```go
// 查询审计日志
query := &audit.QueryRequest{
    TenantID:  "tenant-456",
    UserID:    "user-123",
    EventType: audit.EventAuthLogin,
    StartTime: &startTime,
    EndTime:   &endTime,
    Page:      1,
    PageSize:  20,
}

logs, total, err := auditLogger.Query(ctx, query)

// 获取单个审计日志详情
log, err := auditLogger.GetByID(ctx, auditID)
```

### 7. 导出审计日志

```go
// 导出为 JSON 格式
jsonData, err := auditLogger.ExportAuditLogs(ctx, query, "json")

// 导出为 CSV 格式
csvData, err := auditLogger.ExportAuditLogs(ctx, query, "csv")
```

## 配置说明

### Config 结构

```go
type Config struct {
    // Enabled 是否启用审计日志
    Enabled bool
    
    // LogLevel 日志级别 (All, Info, Warning, Critical, None)
    LogLevel LogLevel
    
    // AsyncEnabled 是否启用异步写入
    AsyncEnabled bool
    
    // QueueSize 异步队列大小
    QueueSize int
    
    // WorkerCount 工作协程数量
    WorkerCount int
    
    // BatchSize 批量写入大小
    BatchSize int
    
    // BatchTimeout 批量写入超时时间（秒）
    BatchTimeout int
    
    // RetentionDays 日志保留天数
    RetentionDays int
    
    // EnableMetadata 是否记录元数据
    EnableMetadata bool
}
```

### 默认配置

```go
config := audit.DefaultConfig()
// 返回：
// {
//     Enabled:        true,
//     LogLevel:       LogLevelAll,
//     AsyncEnabled:   true,
//     QueueSize:      10000,
//     WorkerCount:    3,
//     BatchSize:      100,
//     BatchTimeout:   5,
//     RetentionDays:  90,
//     EnableMetadata: true,
// }
```

## 事件类型

### 认证事件
- `EventAuthLogin` - 登录事件
- `EventAuthLogout` - 登出事件
- `EventAuthFailed` - 认证失败
- `EventAuthTokenRefresh` - Token 刷新
- `EventAuthPasswordChange` - 密码修改

### 数据访问事件
- `EventDataRead` - 数据读取
- `EventDataWrite` - 数据写入
- `EventDataDelete` - 数据删除
- `EventDataExport` - 数据导出

### 管理操作事件
- `EventAdminUserCreate` - 创建用户
- `EventAdminUserUpdate` - 更新用户
- `EventAdminUserDelete` - 删除用户
- `EventAdminRoleAssign` - 分配角色
- `EventAdminRoleRevoke` - 撤销角色
- `EventAdminConfigChange` - 配置变更
- `EventAdminSystemRestart` - 系统重启

### 安全事件
- `EventSecurityAlert` - 安全告警
- `EventSecurityViolation` - 安全违规
- `EventSecurityBlocked` - 安全阻断
- `EventSecurityIncident` - 安全事故

## 日志级别

- `LogLevelAll` - 记录所有日志
- `LogLevelInfo` - 记录 Info、Warning、Critical
- `LogLevelWarning` - 记录 Warning、Critical
- `LogLevelCritical` - 仅记录 Critical
- `LogLevelNone` - 不记录日志

## 性能优化

### 异步写入
异步写入通过队列和批量处理显著提高性能：

```go
// 推荐生产环境配置
config := &audit.Config{
    Enabled:        true,
    LogLevel:       audit.LogLevelInfo,  // 过滤 Info 级别
    AsyncEnabled:   true,                  // 启用异步
    QueueSize:      10000,                 // 大队列
    WorkerCount:    5,                     // 多个 worker
    BatchSize:      200,                   // 大批量
    BatchTimeout:   10,                    // 较长超时
    RetentionDays:  90,
    EnableMetadata: true,
}
```

### 性能基准

- **同步写入**: ~1,000 logs/sec
- **异步写入**: ~50,000 logs/sec
- **批量写入**: ~100,000 logs/sec

## 数据库索引

以下索引已自动创建以优化查询性能：

```sql
-- 单列索引
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);

-- 组合索引
CREATE INDEX idx_audit_logs_user_time ON audit_logs(user_id, timestamp DESC);
CREATE INDEX idx_audit_logs_tenant_time ON audit_logs(tenant_id, timestamp DESC);
```

## 最佳实践

1. **使用异步写入**：生产环境强烈建议启用异步写入以提高性能。

2. **合理设置日志级别**：
   - 开发环境：`LogLevelAll`
   - 生产环境：`LogLevelInfo` 或 `LogLevelWarning`
   - 高安全环境：`LogLevelCritical`

3. **定期清理旧日志**：
```go
// 定期调用清理旧日志
err := auditLogger.DeleteOld(ctx)
```

4. **监控队列状态**：
```go
stats := auditLogger.GetStats()
if stats.QueueSize > 8000 {
    log.Warn("Audit queue is nearly full", zap.Int("queue_size", stats.QueueSize))
}
```

5. **关键操作记录详细元数据**：
```go
err := auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
    audit.EventAdminRoleAssign, "user", "user-002", "Assign admin role",
    beforeState, afterState, changes,
    map[string]interface{}{
        "reason": "promotion",
        "approved_by": "admin-001",
        "ticket_id": "TICKET-123",
    })
```

## 监控和统计

```go
// 获取统计信息
stats := auditLogger.GetStats()
fmt.Printf("Total logs: %d\n", stats.TotalLogs)
fmt.Printf("Success: %d\n", stats.SuccessCount)
fmt.Printf("Failure: %d\n", stats.FailureCount)
fmt.Printf("Queue size: %d\n", stats.QueueSize)
fmt.Printf("Dropped: %d\n", stats.DroppedCount)
```

## 故障排查

### 日志丢失
- 检查队列大小是否足够
- 查看 `DroppedCount` 统计
- 考虑增加 `WorkerCount` 或 `QueueSize`

### 性能问题
- 启用异步写入
- 增加批量写入大小
- 调整工作协程数量

### 磁盘空间
- 定期清理旧日志
- 调整日志保留天数
- 监控数据库大小

## 测试

```bash
# 运行测试
go test -v ./pkg/audit/

# 运行基准测试
go test -bench=. ./pkg/audit/
```

## 许可证

MIT License