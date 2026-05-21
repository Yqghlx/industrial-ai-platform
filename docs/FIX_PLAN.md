# Industrial AI Platform 修复计划

> **创建日期**: 2026-05-14  
> **目标**: 修复代码审查发现的 67 个问题  
> **预计工期**: 4-6 周

---

## 📋 修复阶段概览

| 阶段 | 时间 | 任务数 | 优先级 |
|------|------|--------|--------|
| Phase 1 | Week 1 | 8 | P0 CRITICAL |
| Phase 2 | Week 2-3 | 15 | P0 + P1 HIGH |
| Phase 3 | Week 3-4 | 20 | P1 MEDIUM |
| Phase 4 | Week 5-6 | 24 | P2 LOW |

---

## 🔴 Phase 1: CRITICAL - Week 1 (8任务)

> **目标**: 修复所有安全漏洞和编译错误  
> **预计工时**: 16-20 小时

### FIX-001: 移除 JWT Secret 硬编码

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/internal/service/auth_helpers.go:23` |
| **问题** | 默认密钥硬编码，攻击者可伪造 Token |
| **修复方案** | 1. 移除默认值<br>2. 启动时强制检查 JWT_SECRET 环境变量<br>3. 添加密钥强度验证 (≥32字符) |
| **预估工时** | 2h |
| **验收标准** | ✅ 无硬编码密钥<br>✅ 未设置 JWT_SECRET 时启动失败<br>✅ 密钥长度检查 |

**代码修改**:
```go
// auth_helpers.go
var jwtSecret []byte

func InitJWT(secret string) error {
    if len(secret) < 32 {
        return errors.New("JWT_SECRET must be at least 32 characters")
    }
    jwtSecret = []byte(secret)
    return nil
}

// main.go
if err := auth.InitJWT(appCfg.JWTSecret); err != nil {
    log.Fatalf("JWT initialization failed: %v", err)
}
```

---

### FIX-002: 移除密码明文日志

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/internal/handler/server.go:513` |
| **问题** | 管理员密码直接打印到日志 |
| **修复方案** | 1. 移除密码日志打印<br>2. 通过环境变量 ADMIN_PASSWORD 设置<br>3. 未设置时生成临时密码并只显示一次 |
| **预估工时** | 1h |
| **验收标准** | ✅ 密码不出现在日志<br>✅ 支持环境变量配置<br>✅ 临时密码仅显示在终端 |

---

### FIX-003: 修复 database 包导入缺失

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/internal/handler/server.go:466` |
| **问题** | 使用 `database.NewMigrator` 但未导入包 |
| **修复方案** | 添加导入语句 |
| **预估工时** | 0.5h |
| **验收标准** | ✅ go build 成功 |

---

### FIX-004: 修复 handler 变量名冲突

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/internal/handler/server.go:331` |
| **问题** | `handler.NewAuthHandler` 与包名冲突 |
| **修复方案** | 在同一包内直接调用 `NewAuthHandler` |
| **预估工时** | 0.5h |
| **验收标准** | ✅ 编译成功<br>✅ 无变量名冲突警告 |

---

### FIX-005: 统一 JWT 过期时间

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `auth_helpers.go:16` + `jwt_helpers.go:33` |
| **问题** | AccessToken 15分钟 vs 默认Token 24小时 |
| **修复方案** | 1. 移除旧 GenerateToken API<br>2. 统一使用 AccessToken 15分钟 + RefreshToken 7天<br>3. 添加配置化过期时间 |
| **预估工时** | 2h |
| **验收标准** | ✅ 所有 Token 15分钟有效期<br>✅ RefreshToken 7天<br>✅ 配置可调 |

---

### FIX-006: 使用 crypto/rand 替代 math/rand

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/internal/service/agent_service.go:323,337-342` |
| **问题** | Session ID 使用非安全随机数 |
| **修复方案** | 替换为 crypto/rand |
| **预估工时** | 1h |
| **验收标准** | ✅ 使用 crypto/rand<br>✅ Session ID 不可预测 |

**代码修改**:
```go
func generateSecureID() string {
    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        panic(err) // crypto/rand 失败是严重错误
    }
    return hex.EncodeToString(b)
}
```

---

### FIX-007: 修复前端类型错误

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | 4处前端文件 |
| **问题** | 类型定义不一致 |
| **修复方案** | 1. `DeviceDetail.tsx`: 导入 DeviceStats<br>2. `KnowledgeGraph.tsx`: 使用 DeviceGraph<br>3. `BlackBoxCenter.tsx`: 统一数据类型<br>4. `errorHelper.ts`: 修正 UNAUTHORIZED 值 |
| **预估工时** | 2h |
| **验收标准** | ✅ TypeScript 编译无错误<br>✅ 无 any 类型滥用 |

---

### FIX-008: 修复 context 使用

| 项目 | 详情 |
|------|------|
| **优先级** | P0 CRITICAL |
| **文件** | `backend/main.go:46` |
| **问题** | context.WithCancel 创建后未使用 |
| **修复方案** | 正确实现 graceful shutdown |
| **预估工时** | 1h |
| **验收标准** | ✅ Graceful shutdown 正确工作<br>✅ 所有连接优雅关闭 |

---

## 🟠 Phase 2: HIGH - Week 2-3 (15任务)

> **目标**: 完成核心功能修复和基础测试  
> **预计工时**: 30-40 小时

### FIX-009: 实现 RevokeAllUserTokens

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | `backend/internal/service/auth_helpers.go:238-243` |
| **问题** | 函数只返回 nil，Token 未撤销 |
| **修复方案** | 1. 在 JWT Claims 添加 TokenVersion<br>2. 密码修改时更新版本号<br>3. 验证时检查版本号 |
| **预估工时** | 3h |
| **验收标准** | ✅ 修改密码后旧 Token 失效<br>✅ 测试覆盖 |

---

### FIX-010: 添加 auth_handler 测试

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | 新建 `backend/internal/handler/auth_handler_test.go` |
| **问题** | Handler 测试 0% 覆盖 |
| **修复方案** | 测试覆盖：<br>1. 登录/注册正常流程<br>2. 无效凭证测试<br>3. Token 验证测试<br>4. Refresh Token 测试 |
| **预估工时** | 4h |
| **验收标准** | ✅ 覆盖率 >80%<br>✅ 所有测试通过 |

---

### FIX-011: 添加 device_handler 测试

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | 新建 `backend/internal/handler/device_handler_test.go` |
| **问题** | Handler 测试 0% 覆盖 |
| **修复方案** | 测试覆盖：<br>1. CRUD 操作测试<br>2. 权限验证测试<br>3. 输入验证测试<br>4. 分页测试 |
| **预估工时** | 3h |
| **验收标准** | ✅ 覆盖率 >80% |

---

### FIX-012: 添加 device_repo 测试

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | 新建 `backend/internal/repository/device_repo_test.go` |
| **问题** | Repository 测试 0% 覆盖 |
| **修复方案** | 使用 go-sqlmock 测试：<br>1. CRUD SQL 查询<br>2. 边界条件<br>3. 错误处理 |
| **预估工时** | 3h |
| **验收标准** | ✅ SQL 查询验证<br>✅ 锆误处理覆盖 |

---

### FIX-013: 添加 tenant_repo 测试

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | 新建 `backend/internal/repository/tenant_repo_test.go` |
| **预估工时** | 2h |

---

### FIX-014: 添加 rbac_repo 测试

| 项目 | 详情 |
|------|------|
| **优先级** | P0 HIGH |
| **文件** | 新建 `backend/internal/repository/rbac_repo_test.go` |
| **预估工时** | 3h |

---

### FIX-015: 修复 CORS 默认配置

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `backend/internal/config/config.go:212` |
| **问题** | CORS 默认 `*` 过宽松 |
| **修复方案** | 1. 移除默认 `*`<br>2. 生产环境强制配置 CORS_ORIGINS<br>3. 添加启动警告 |
| **预估工时** | 1h |
| **验收标准** | ✅ 生产环境禁止 `*`<br>✅ 明确配置要求 |

---

### FIX-016: WebSocket CheckOrigin 验证

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `backend/internal/handler/server.go:196-208` |
| **问题** | 无 Origin header 的请求被允许 |
| **修复方案** | 1. 生产环境严格验证 Origin<br>2. Origin 必须在允许列表<br>3. WebSocket 认证 |
| **预估工时** | 2h |
| **验收标准** | ✅ 无合法 Origin 时拒绝连接<br>✅ 认证 Token 验证 |

---

### FIX-017: 添加输入验证

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `device_handler.go`, `tenant_handler.go` 等 |
| **问题** | page/pageSize 参数无边界检查 |
| **修复方案** | 1. 添加 maxPageSize=100<br>2. strconv.Atoi 错误处理<br>3. 字符串长度限制 |
| **预估工时** | 2h |
| **验收标准** | ✅ 所有数值参数验证<br>✅ 错误返回 400 |

---

### FIX-018: 提高密码复杂度

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `backend/internal/model/auth_models.go:6,34` |
| **问题** | 密码验证过于简单 (min=6) |
| **修复方案** | 1. 最小长度 12<br>2. 正则验证：大小写+数字+特殊字符<br>3. 密码强度评分 |
| **预估工时** | 2h |
| **验收标准** | ✅ 密码强度验证<br>✅ 清晰错误提示 |

---

### FIX-019: 修复 HTTP Client 未复用

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `backend/internal/service/agent_service.go:172` |
| **问题** | 每次 LLM 调用创建新 Client |
| **修复方案** | Service 初始化时创建共享 Client |
| **预估工时** | 1h |
| **验收标准** | ✅ 单一 HTTP Client<br>✅ 连接复用 |

---

### FIX-020: 统一 WebSocket 广播器

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `telemetry_service.go` + `handler.go` |
| **问题** | 重复实现 |
| **修复方案** | 统一为单一实现 |
| **预估工时** | 3h |
| **验收标准** | ✅ 单一 WebSocket 管理<br>✅ 无重复代码 |

---

### FIX-021: 添加 CSRF 防护

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | 新建 middleware |
| **问题** | 缺少 CSRF Token |
| **修复方案** | 1. 添加 CSRF 中间件<br>2. POST/PUT/DELETE 验证 Token<br>3. SameSite Cookie |
| **预估工时** | 3h |
| **验收标准** | ✅ CSRF Token 生成和验证<br>✅ 测试覆盖 |

---

### FIX-022: 修复 Rate Limiter goroutine

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | `backend/internal/middleware/ratelimit.go:102-107` |
| **问题** | 每次 RateLimit 启动新 ticker goroutine |
| **修复方案** | 使用单例模式管理清理 goroutine |
| **预估工时** | 2h |
| **验收标准** | ✅ 无 goroutine 泄漏<br>✅ pprof 验证 |

---

### FIX-023: 添加安全审计日志

| 项目 | 详情 |
|------|------|
| **优先级** | P1 HIGH |
| **文件** | 新建 audit middleware |
| **问题** | 缺少敏感操作审计 |
| **修复方案** | 1. 创建 audit_service<br>2. 记录登录/密码修改/用户操作<br>3. 持久化存储 |
| **预估工时** | 4h |
| **验收标准** | ✅ 关键操作审计<br>✅ 审计日志可查询 |

---

## 🟡 Phase 3: MEDIUM - Week 3-4 (20任务)

> **目标**: 优化性能和代码质量  
> **预计工时**: 40-50 小时

### 前端优化 (10任务)

| ID | 任务 | 工时 |
|----|------|------|
| FIX-024 | 添加 API 返回类型定义 | 3h |
| FIX-025 | useCallback 包裹 loadDevices | 1h |
| FIX-026 | useCallback 包裹 loadOrders | 1h |
| FIX-027 | WebSocket 单连接管理 | 2h |
| FIX-028 | AITeamDashboard 消息限制 | 0.5h |
| FIX-029 | wsCompression payload 类型化 | 1h |
| FIX-030 | 添加 Navigator 类型声明 | 1h |
| FIX-031 | 提取颜色映射函数 | 1h |
| FIX-032 | 状态颜色国际化 | 1h |
| FIX-033 | Tailwind 动态类名 safelist | 0.5h |

### 后端优化 (10任务)

| ID | 任务 | 工时 |
|----|------|------|
| FIX-034 | 统一 CORS 中间件 | 1h |
| FIX-035 | 统一 Security Headers | 1h |
| FIX-036 | crypto/rand RequestID | 1h |
| FIX-037 | 移除全局变量 | 3h |
| FIX-038 | 添加 Service 接口定义 | 3h |
| FIX-039 | HTTP Timeout 配置化 | 1h |
| FIX-040 | WAF User-Agent 配置 | 1h |
| FIX-041 | Token 黑名单 fallback | 2h |
| FIX-042 | 配置参数外部化 | 2h |
| FIX-043 | 统一日志库 zap | 2h |

---

## 🟢 Phase 4: LOW - Week 5-6 (24任务)

> **目标**: 代码风格和文档完善  
> **预计工时**: 30-40 小时

### 测试补充 (10任务)

| ID | 任务 | 工时 |
|----|------|------|
| FIX-044 | telemetry_handler 测试 | 2h |
| FIX-045 | alert_handler 测试 | 2h |
| FIX-046 | tenant_handler 测试 | 2h |
| FIX-047 | rbac_handler 测试 | 2h |
| FIX-048 | export_handler 测试 | 1h |
| FIX-049 | middleware/auth 测试 | 2h |
| FIX-050 | middleware/rbac 测试 | 2h |
| FIX-051 | middleware/ratelimit 测试 | 1h |
| FIX-052 | SQL 注入安全测试 | 2h |
| FIX-053 | 认证绕过测试 | 2h |

### 代码风格 (14任务)

| ID | 任务 | 工时 |
|----|------|------|
| FIX-054 | 统一导入路径 | 1h |
| FIX-055 | 移除 itoa 自实现 | 0.5h |
| FIX-056 | TTL 变量常量化 | 1h |
| FIX-057 | 添加业务逻辑注释 | 2h |
| FIX-058 | Server 结构体拆分 | 3h |
| FIX-059 | 移除 init() goroutine | 1h |
| FIX-060 | CSP 严格配置 | 2h |
| FIX-061 | 数据库连接池优化 | 1h |
| FIX-062 | 前端 hooks 命名修复 | 0.5h |
| FIX-063 | useVirtualList useMemo | 1h |
| FIX-064 | 国际化插值支持 | 2h |
| FIX-065 | ErrorBoundary 国际化 | 1h |
| FIX-066 | 代码规范检查 CI | 2h |
| FIX-067 | 更新 README | 1h |

---

## 📊 工时汇总

| 阶段 | 任务数 | 工时 | 完成标准 |
|------|--------|------|----------|
| Phase 1 | 8 | 16-20h | 安全漏洞修复、编译通过 |
| Phase 2 | 15 | 30-40h | 核心测试 >60% 覆盖 |
| Phase 3 | 20 | 40-50h | 性能优化、代码质量提升 |
| Phase 4 | 24 | 30-40h | 测试 >70%、代码规范 |
| **总计** | **67** | **116-150h** | 生产就绪 |

---

## 🎯 里程碑

| 里程碑 | 日期 | 目标 |
|--------|------|------|
| M1 | Week 1 结束 | 无安全漏洞、编译通过 |
| M2 | Week 2 结束 | Handler/Repo 测试 >60% |
| M3 | Week 3 结束 | 前端优化完成 |
| M4 | Week 4 结束 | 性能优化完成 |
| M5 | Week 6 结束 | 测试 >70%、代码质量达标 |

---

## 📝 执行建议

### 每日流程

```
1. 选择当日任务 (2-3个)
2. 创建 feature 分支
3. 编写代码 + 测试
4. 本地验证 (go build, npm run test)
5. 提交 PR
6. Code Review
7. 合并到 main
```

### 验收标准

- ✅ 所有修复有对应测试
- ✅ go build 无错误
- ✅ npm run type-check 无错误
- ✅ 测试覆盖率达标
- ✅ CI Pipeline 通过

---

## 🔧 工具支持

### 验证命令

```bash
# 后端
cd backend
go build ./...
go test ./... -cover
go vet ./...

# 前端
cd frontend
npm run type-check
npm run test
npm run lint

# 安全检查
gosec ./...
npm audit
```

---

**修复计划创建完成！建议按 Phase 顺序执行。**