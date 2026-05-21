# 后端代码质量审查报告

**审查日期**: 2025-05-14
**审查范围**: backend/ 目录所有 Go 代码
**审查重点**: 代码结构、错误处理、死代码、硬编码、性能、并发安全、代码重复

---

## P0 级问题 (严重) - 需立即修复

### 1. WebSocket Broadcaster Goroutine 泄漏风险
**文件**: `internal/service/telemetry_service.go` (第 111-177 行)
**问题描述**: `MemoryTokenBlacklist.cleanupExpiredEntries()` 的 goroutine 永久运行，没有停止机制。`for range ticker.C` 会一直阻塞，无法被外部停止。
```go
func (b *MemoryTokenBlacklist) cleanupExpiredEntries() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {  // 永久运行，无法停止！
        b.mu.Lock()
        // ...
        b.mu.Unlock()
    }
}
```
**建议修复方案**: 
1. 添加 `stopCh` channel 和 context 参数
2. 在 select 中监听 stop channel
3. 提供明确的 `Stop()` 方法
**预估工时**: 2h

---

### 2. HybridTokenBlacklist Redis健康检查Goroutine无法停止
**文件**: `internal/service/auth_helpers.go` (第 192-212 行)
**问题描述**: `checkRedisHealth` goroutine 使用 `for range ticker.C` 无法被停止，存在泄漏风险。
```go
func (b *HybridTokenBlacklist) checkRedisHealth(client *redis.Client) {
    ticker := time.NewTicker(b.checkInterval)
    defer ticker.Stop()
    
    for range ticker.C {  // 无法停止
        // ...
    }
}
```
**建议修复方案**: 添加 stopCh channel，使用 select 监听关闭信号。
**预估工时**: 1.5h

---

### 3. Context 缺失超时控制
**文件**: `internal/handler/device_handler.go` (第 16-175 行)
**问题描述**: 所有 handler 使用 `context.Background()` 没有超时设置，可能导致请求长时间阻塞。
```go
func (s *Server) listDevices(c *gin.Context) {
    ctx := context.Background()  // 无超时！
    devices, total, err := s.deviceRepo.List(ctx, pagination.Page, pagination.PageSize)
}
```
**建议修复方案**: 
1. 使用 `context.WithTimeout(context.Background(), 30*time.Second)`
2. 或从 gin.Context 提取 request context: `c.Request.Context()`
**预估工时**: 3h (需修改所有 handler)

---

## P1 级问题 (高) - 本周内修复

### 4. N+1 查询风险 - AlertService.EvaluateRules
**文件**: `internal/service/alert_service.go` (第 50-100 行)
**问题描述**: 每次遥测数据入库都会查询所有规则并逐个检查 cooldown，大量数据时性能问题严重。
```go
func (s *AlertService) EvaluateRules(ctx context.Context, data *model.TelemetryData) error {
    rules, err := s.ruleRepo.ListEnabled(ctx)  // 查询所有规则
    for _, rule := range rules {
        // 每个规则都查数据库检查 cooldown
        recentAlert, err := s.alertRepo.GetRecentByDevice(ctx, data.DeviceID, rule.ID, rule.CooldownSec)
    }
}
```
**建议修复方案**: 
1. 批量查询 cooldown 状态
2. 使用缓存暂存规则和 cooldown 状态
3. 考虑规则分组评估
**预估工时**: 4h

---

### 5. 全局变量滥用导致并发安全风险
**文件**: `internal/service/auth_helpers.go` (第 482 行)
**问题描述**: `globalJWTService` 作为全局变量，初始化后无法安全更新，多实例场景下可能存在问题。
```go
var globalJWTService *JWTService  // 全局变量，难以管理
```
**建议修复方案**: 
1. 使用依赖注入替代全局变量
2. 在 Server 结构体中管理 JWTService 实例
3. 移除 deprecated 函数的使用
**预估工时**: 2h

---

### 6. 统计数据并发安全问题
**文件**: `pkg/wscompression/compressor.go` (第 138-143 行)
**问题描述**: `ShouldCompress` 中跳过消息统计使用了锁，但 `Compress` 中的统计在锁外更新 `TotalMessages`。
```go
func (c *Compressor) ShouldCompress(data []byte) bool {
    if len(data) < c.config.MinSize {
        c.mu.Lock()
        c.stats.SkippedMessages++  // 这里加锁
        c.mu.Unlock()
        return false
    }
    return true
}

func (c *Compressor) Compress(data []byte) ([]byte, error) {
    // ...
    c.mu.Lock()
    c.stats.CompressedMessages++
    c.stats.TotalMessages++  // 这里有锁
    c.mu.Unlock()
    // ...
}
```
以及 `WriteCompressed` (第 230-232 行):
```go
c.mu.Lock()
c.stats.TotalMessages++  // 重复更新
c.mu.Unlock()
```
**建议修复方案**: 统一统计更新逻辑，避免重复计数和锁竞争。
**预估工时**: 1h

---

### 7. 数据库连接池参数硬编码
**文件**: `internal/handler/server.go` (第 127-130 行)
**问题描述**: 数据库连接池参数硬编码，无法根据负载动态调整。
```go
db.SetMaxOpenConns(25)   // 硬编码
db.SetMaxIdleConns(5)    // 硬编码
db.SetConnMaxLifetime(5 * time.Minute)  // 硬编码
```
**建议修复方案**: 
1. 从配置读取连接池参数
2. 添加到 Config 结构体
3. 支持环境变量覆盖
**预估工时**: 1h

---

### 8. Repository 错误处理不完整
**文件**: `internal/repository/rule_repo.go` (第 54-55, 81-82 行)
**问题描述**: `json.Unmarshal` 错误被忽略，可能导致数据损坏。
```go
json.Unmarshal([]byte(actionsJSON), &rule.Actions)  // 忽略错误！
```
**建议修复方案**: 
1. 检查并处理 Unmarshal 错误
2. 使用 json.Decoder 更严格验证
3. 添加日志记录解析失败
**预估工时**: 1h

---

## P2 级问题 (中) - 两周内修复

### 9. 代码重复 - Pagination 处理
**文件**: `internal/handler/device_handler.go`, `internal/handler/auth_handler.go` 等
**问题描述**: 每个 handler 重复调用 `GetPagination(c)`，PaginationParams 结构重复定义。
**建议修复方案**: 
1. 提取 Pagination 处理到 middleware
2. 使用 gin binding 自动绑定
3. 创建统一的 Pagination 响应结构
**预估工时**: 2h

---

### 10. 魔法数字散布代码中
**文件**: 多处
**问题描述**: 代码中大量硬编码数字：
- `internal/service/telemetry_service.go`: 温度阈值 100, 120; 振动阈值 3.0, 5.0
- `pkg/cache/redis.go`: 扫描批次 100
- `internal/service/alert_service.go`: cooldown 300, 180, 600
- `internal/middleware/ratelimit.go`: 限流参数 5, 1, 3, 0.5, 100, 10 等

```go
if data.Temperature > 100 || data.Vibration > 3.0 {  // 魔法数字
    data.Status = "warning"
}
```
**建议修复方案**: 
1. 定义常量或从配置读取
2. 创建 ThresholdConfig 结构体
3. 支持动态调整阈值
**预估工时**: 2h

---

### 11. Repository 代码结构重复
**文件**: `internal/repository/telemetry_repo.go`, `internal/repository/rule_repo.go` 等
**问题描述**: List 函数结构高度重复：分页、过滤、排序逻辑几乎相同。
```go
// WorkOrderRepository.List
whereClause := "WHERE 1=1"
args := []interface{}{}
argIdx := 1
// ... 类似代码在多个 repo 中重复
```
**建议修复方案**: 
1. 创建通用的 QueryBuilder
2. 使用泛型或接口抽象通用列表查询
3. 提取公共的 count/offset 逻辑
**预估工时**: 3h

---

### 12. Cache 配置默认值重复
**文件**: `pkg/cache/cache.go`, `pkg/cache/redis.go`, `pkg/cache/memory.go`
**问题描述**: 默认配置值在多处重复定义（TTL: 5min, Prefix: "iai:" 等）。
**建议修复方案**: 统一到 Config 的 DefaultConfig() 函数。
**预估工时**: 0.5h

---

### 13. 缺少 Request Validation
**文件**: `internal/handler/device_handler.go` (第 51-68 行)
**问题描述**: createDevice/updateDevice 没有验证输入字段，可能注入恶意数据。
```go
var device model.Device
if err := c.ShouldBindJSON(&device); err != nil {
    // 只检查 JSON 格式，不验证字段值
}
```
**建议修复方案**: 
1. 添加 validator middleware
2. 使用 go-playground/validator 结构体验证
3. 添加字段长度、格式限制
**预估工时**: 2h

---

## P3 级问题 (低) - 下月修复

### 14. 测试代码错误 - 类型不匹配
**文件**: `internal/handler/tenant_handler_test.go` (第 97 行)
**问题描述**: `go vet` 报告类型错误：`strPtr("Updated Tenant")` 应为 `string` 而非 `*string`。
```go
vet: internal/handler/tenant_handler_test.go:97:17: cannot use strPtr("Updated Tenant") (value of type *string) as string value in struct literal
```
**建议修复方案**: 修复测试代码中的类型使用。
**预估工时**: 0.5h

---

### 15. 未使用导入
**文件**: `internal/repository/tenant_repo_test.go` (第 4 行)
**问题描述**: `context` 包导入但未使用。
```go
vet: internal/repository/tenant_repo_test.go:4:2: "context" imported and not used
```
**建议修复方案**: 移除未使用导入或添加测试代码使用。
**预估工时**: 0.25h

---

### 16. 注释风格不一致
**文件**: 多处
**问题描述**: 部分使用 FIX-xxx 注释标记，部分使用普通注释，部分无注释。
**建议修复方案**: 
1. 统一注释规范
2. 添加 godoc 标准注释
3. 移除或更新过时的 FIX-xxx 注释
**预估工时**: 1h

---

### 17. 日志输出不完整
**文件**: `internal/service/agent_service.go`, `pkg/audit/service.go`
**问题描述**: 关键操作缺少日志记录，难以排查问题。
**建议修复方案**: 
1. 使用统一的 logger 包
2. 关键操作添加结构化日志
3. 添加 traceID 支持
**预估工时**: 1h

---

### 18. 死代码风险 - 未使用的 WebSocketManager
**文件**: `internal/handler/websocket.go` (第 23-39 行)
**问题描述**: `WebSocketManager` 结构体定义但从未被使用，Server 仍使用内置的 wsClients 管理。
```go
type WebSocketManager struct {
    clients    map[*websocket.Conn]bool
    // ... 定义但未使用
}
```
**建议修复方案**: 
1. 统一使用 WebSocketManager
2. 或移除未使用的结构体
**预估工时**: 1h

---

## 代码架构改进建议

### 1. 依赖注入改进
当前代码使用大量全局变量和直接初始化，建议：
- 使用 wire 或 dig 进行依赖注入
- 创建 App 结构体管理所有服务实例
- 移除 SetXXXSecret 类全局设置函数

### 2. 错误处理标准化
建议创建统一的错误类型：
```go
type AppError struct {
    Code    string
    Message string
    Details map[string]interface{}
}
```

### 3. Context 传递规范化
所有数据库操作应使用带超时的 context：
- Handler: 从 request context 创建带超时的子 context
- Service: 接收并传递 context
- Repository: 使用 context 做超时控制

---

## 审查总结

| 优先级 | 问题数量 | 预估总工时 |
|--------|----------|-----------|
| P0     | 3        | 6.5h      |
| P1     | 5        | 8h        |
| P2     | 5        | 7.5h      |
| P3     | 5        | 2.75h     |
| **总计** | **18** | **24.75h** |

---

## 建议修复顺序

1. **第一周**: P0 级问题 - Goroutine 泄漏和 Context 超时
2. **第二周**: P1 级问题 - N+1 查询、并发安全、配置外部化
3. **第三周**: P2 级问题 - 代码重构、验证改进
4. **第四周**: P3 级问题 - 测试修复、文档完善

---

*审查完成。建议按优先级逐步修复，同时进行架构层面的持续改进。*