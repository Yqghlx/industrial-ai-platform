# 第四轮审计 - 代码清理检查报告

**审计时间**: 2026-05-26
**审计范围**: /Users/yqgmac/yqg/project/industrial-ai-platform/backend

---

## 一、清理建议汇总表

| 问题类型 | 文件位置 | 清理建议 | 优先级 |
|---------|---------|---------|--------|
| 未使用变量 | `cmd/mock_server/main.go:19-20` | 删除 `requestCount` 和 `countMutex` | P3 |
| 未使用函数 | `internal/handler/websocket.go:207` | 删除 `broadcast` 函数或添加调用 | P2 |
| 未使用函数 | `internal/handler/websocket.go:232` | 删除 `getWSCompressionStats` 或添加调用 | P2 |
| 未使用函数 | `internal/middleware/circuitbreaker.go:234` | 删除 `marshalFallbackData` | P3 |
| 未使用函数 | `internal/middleware/waf.go:380` | 删除 `isBlockedUserAgent` | P3 |
| 未使用变量 | `internal/security/sql_injection_enhanced.go:105` | 删除 `suspiciousChars` 变量 | P3 |
| 未使用常量 | `internal/service/auth_service_test.go:30` | 删除 `userQueryColumns` | P3 |
| 未使用函数 | `tests/integration/setup_test.go:210` | 删除 `setupGinTestMode` | P3 |
| append结果未使用 | `benchmarks/api_bench_test.go:87` | 修复 append 返回值使用 | P2 |
| 空指针风险 | `internal/config/config_test.go:373,443` | 添加 nil 检查 | P1 |
| 空指针风险 | `pkg/logger/logger_test.go:108,530,557` | 添加 nil 检查 | P1 |
| nil Context传递 | `internal/handler/alert_handler_test.go` (15处) | 使用 `context.Background()` 或 `context.TODO()` | P2 |
| 废弃API使用 | `pkg/tracing/tracer.go:92` | 将 `WithSpanLimits` 替换为 `WithRawSpanLimits` | P2 |
| 代码简化 | `pkg/audit/repository.go:185-231` | 移除冗余 nil check，直接使用 `len()` | P3 |
| 代码简化 | `internal/service/agent_llm.go:56` | 移除冗余 nil check | P3 |
| 导入格式 | 50+ 个文件 | 运行 `goimports -w .` 统一格式 | P3 |
| TODO注释 | `internal/service/factory.go:48` | 实现或移除 "实现完整的 Service 初始化" | P2 |
| TODO注释 | `internal/handler/admin_handler_new.go:27` | 扩展 AuthServiceInterface 添加 List 方法 | P2 |
| TODO注释 | `internal/handler/health_handler_new.go:26,37` | 实现真实的缓存/WebSocket状态查询 | P2 |
| TODO注释 | `internal/handler/factory_coverage_test.go:136,149` | 实现统一接口 | P3 |
| 过时文件 | `internal/middleware/cors_csrf.go` | 整个文件已废弃，可删除 | P2 |
| fmt.Printf使用 | `internal/middleware/logger.go` (多处) | 改用结构化日志 | P2 |
| fmt.Printf使用 | `internal/service/auth_blacklist.go` (3处) | 改用结构化日志 | P2 |
| fmt.Printf使用 | `internal/service/auth_service.go` (1处) | 改用结构化日志 | P2 |
| fmt.Printf使用 | `internal/service/auth_jwt.go` (3处) | 改用结构化日志 | P2 |
| 大量不可达函数 | `internal/handler/mock_common_test.go` | 评估是否需要这些Mock方法 | P3 |
| 大量不可达函数 | `internal/handler/test_helper_test.go` | 评估是否需要这些测试辅助函数 | P3 |

---

## 二、详细分析

### 2.1 未使用代码（高优先级）

#### staticcheck U1000 检测结果

```
cmd/mock_server/main.go:19:2: var requestCount is unused
cmd/mock_server/main.go:20:2: var countMutex is unused
internal/handler/websocket.go:207:18: func broadcast is unused
internal/handler/websocket.go:232:18: func getWSCompressionStats is unused
internal/middleware/circuitbreaker.go:234:6: func marshalFallbackData is unused
internal/middleware/waf.go:380:6: func isBlockedUserAgent is unused
internal/security/sql_injection_enhanced.go:105:2: var suspiciousChars is unused
internal/service/auth_service_test.go:30:7: const userQueryColumns is unused
tests/integration/setup_test.go:210:6: func setupGinTestMode is unused
```

**建议**: 逐一审查这些未使用的变量/函数，确认是否可以删除或需要添加调用点。

### 2.2 空指针风险（P1）

**文件**: `internal/config/config_test.go`
- 第373行和第443行存在可能的空指针解引用

**文件**: `pkg/logger/logger_test.go`
- 第108、530、557行存在可能的空指针解引用

**建议**: 在解引用前添加 nil 检查。

### 2.3 nil Context 传递（P2）

**文件**: `internal/handler/alert_handler_test.go`
- 共15处传递 nil Context 给函数

**建议**: 将所有 `nil` 替换为 `context.Background()` 或 `context.TODO()`。

### 2.4 废弃API使用（P2）

**文件**: `pkg/tracing/tracer.go:92`
```go
sdktrace.WithSpanLimits // 已废弃
```

**建议**: 替换为 `sdktrace.WithRawSpanLimits`。

### 2.5 TODO 注释（共6处）

| 文件 | 行号 | 内容 | 建议 |
|-----|------|------|------|
| internal/service/factory.go | 48 | 实现完整的 Service 初始化 | P2 - 创建任务跟踪 |
| internal/handler/admin_handler_new.go | 27 | 需要扩展 AuthServiceInterface 添加 List 方法 | P2 - 创建任务跟踪 |
| internal/handler/health_handler_new.go | 26 | 实现真实的缓存状态查询 | P2 - 创建任务跟踪 |
| internal/handler/health_handler_new.go | 37 | 实现真实的WebSocket状态查询 | P2 - 创建任务跟踪 |
| internal/handler/factory_coverage_test.go | 136 | returns nil until unified interface | P3 - 评估 |
| internal/handler/factory_coverage_test.go | 149 | returns nil until unified interface | P3 - 评估 |

### 2.6 过时/废弃文件

**文件**: `internal/middleware/cors_csrf.go`
- 整个文件已标记为 DEPRECATED
- 功能已迁移到 `cors.go`
- 建议删除此文件

### 2.7 代码简化机会（S1009）

**文件**: `pkg/audit/repository.go` (多处)
```go
// 当前代码
if m != nil && len(m) > 0 { ... }

// 可简化为（len() 对 nil map 返回 0）
if len(m) > 0 { ... }
```

**文件**: `internal/service/agent_llm.go:56`
- 同样的冗余 nil check

### 2.8 日志规范化

**需要改用结构化日志的位置**:
- `internal/middleware/logger.go` - fmt.Printf (多处)
- `internal/service/auth_blacklist.go` - fmt.Printf (3处)
- `internal/service/auth_service.go` - fmt.Printf (1处)
- `internal/service/auth_jwt.go` - fmt.Printf (3处)

### 2.9 goimports 格式问题

约50个文件存在导入格式不规范，建议运行：
```bash
goimports -w .
```

---

## 三、deadcode 检测统计

大量不可达函数集中在测试辅助文件：

| 文件 | 不可达函数数量 |
|-----|--------------|
| internal/handler/mock_common_test.go | 36个 |
| internal/handler/test_helper_test.go | 14个 |
| internal/middleware/prometheus.go | 12个 |
| internal/handler/server_new_integration_test.go | 7个 |

**建议**: 评估这些测试辅助代码是否真正需要，或是否应该重构测试以使用它们。

---

## 四、清理优先级排序

### P1 - 立即处理（影响代码正确性）
1. 空指针风险 - config_test.go, logger_test.go

### P2 - 尽快处理（影响代码质量）
1. nil Context 传递 - alert_handler_test.go
2. 废弃API使用 - tracer.go
3. TODO注释跟踪
4. 删除废弃文件 - cors_csrf.go
5. 日志规范化

### P3 - 可延后处理（代码整洁）
1. 删除未使用变量/函数
2. goimports 格式化
3. 代码简化
4. 评估测试辅助代码

---

## 五、快速修复命令

```bash
# 1. 格式化导入
goimports -w .

# 2. 运行 staticcheck
staticcheck ./...

# 3. 查找未使用代码
go run golang.org/x/tools/cmd/deadcode@latest -test ./...
```

---

**审计完成**