# 工业AI平台修复计划

## Phase 1: P0/CRITICAL 紧急修复（9项）

| ID | 文件位置 | 问题类型 | 修复方案 | 预估时间 |
|-----|---------|---------|---------|---------|
| P0-01 | `pkg/redis/performance.go:49` | 硬编码Redis地址 `localhost:6379` | 使用环境变量 `REDIS_URL` | 0.5h |
| P0-02 | `pkg/validation/uuid.go:128,134,140,146` | 每次调用重新编译正则表达式 | 移到包级别预编译 | 0.5h |
| P0-03 | `internal/service/alert_service.go:599,609,612` | 忽略json.Marshal错误 | 正确处理错误返回值 | 0.5h |
| P0-04 | `pkg/audit/examples.go:30` | 生产代码中使用panic | 移除panic，使用错误处理 | 0.5h |
| P0-05 | `pkg/logger/logger.go:208` | 初始化忽略错误 | 检查错误并处理 | 0.5h |
| P0-06 | `internal/repository/base_repo.go` | 动态SQL拼接风险 | 确保白名单完整覆盖 | 1h |
| P0-07 | `frontend/src/lib/performance.tsx:171` | window.addEventListener未清理 | 添加removeEventListener | 0.5h |
| SEC-CRITICAL-01 | `.secrets.tmp`文件 | 明文密钥/密码泄露 | 立即删除并轮换密钥 | 1h |
| SEC-CRITICAL-02 | `internal/service/agent_service.go:289` | 敏感临时文件写入 | 使用安全文件写入权限0600 | 0.5h |

**Phase 1 总预估时间**: 5h

---

## Phase 2: P1/HIGH 重要修复（21项）

### 后端P1级（9项）

| ID | 文件位置 | 问题类型 | 修复方案 |
|-----|---------|---------|---------|
| P1-01 | `internal/service/telemetry_service.go:76` | 忽略UpdateStatus错误 | 添加错误处理，记录日志 |
| P1-02 | `internal/service/telemetry_service.go:388` | 忽略Count错误 | 处理错误或返回默认值 |
| P1-03 | `internal/service/agent_service.go:289` | 忽略Create错误 | 添加错误处理 |
| P1-04 | `pkg/audit/repository.go:262` | 忽略RowsAffected错误 | 检查删除结果 |
| P1-05 | `pkg/redis/performance_test.go` | 测试中硬编码Redis地址 | 使用mock或测试配置 |
| P1-06 | `internal/handler/health_handler_new.go:26,37` | TODO未实现 | 实现真实状态查询 |
| P1-07 | `internal/service/alert_service.go:589` | 忽略json.Unmarshal错误 | 添加错误检查和fallback |
| P1-08 | `internal/middleware/auth.go:285` | 类型断言无检查 | 使用安全的类型断言模式 |
| P1-09 | `internal/service/factory.go:48` | TODO未实现 | 完善Service初始化逻辑 |

### 前端P1级（8项）

| ID | 文件位置 | 问题类型 | 修复方案 |
|-----|---------|---------|---------|
| P1-01~05 | 多个组件 | eslint-disable绕过依赖检查 | 移除eslint-disable，正确处理依赖 |
| P1-06~08 | 多个组件 | 类型断言问题 | 使用类型守卫 |

### 安全HIGH级（4项）

| ID | 文件位置 | 问题类型 | 修复方案 |
|-----|---------|---------|---------|
| SEC-HIGH-01 | 数据库连接 | SSL禁用(`sslmode=disable`) | 使用`sslmode=require` |
| SEC-HIGH-02 | 遥测公端点 | 无认证机制 | 添加API Key认证 |
| SEC-HIGH-03 | 管理员接口 | 占位实现 | 实现完整认证逻辑 |
| SEC-HIGH-04 | CORS配置 | 生产环境通配符 | 限制允许的源 |

**Phase 2 总预估时间**: 10h

---

## Phase 3: P2/MEDIUM 优化修复（17项）

- 后端P2级（14项）：硬编码URL/端口、魔法数字、Goroutine泄漏风险、context.Background()滥用
- 前端P2级（3项）：缺少React.memo优化、硬编码文本

**Phase 3 总预估时间**: 15h

---

## 执行策略

### Batch Execution Pattern（推荐）

使用 `delegate_task` 启动 2-3 个并行子代理，每个子代理处理一组相关修复：

**Loop 1**: 后端P0级修复（3个子代理）
- 子代理1: P0-01, P0-02 (Redis硬编码 + 正则预编译)
- 子代理2: P0-03, P0-04, P0-05 (错误处理)
- 子代理3: P0-06 + SEC-CRITICAL-02 (SQL安全 + 文件权限)

**Loop 2**: 前端P0级修复 + 安全紧急修复
- 子代理1: P0-07 (前端事件监听器)
- 子代理2: SEC-CRITICAL-01 (删除密钥文件)

**Loop 3**: 后端P1级修复（批量错误处理）

---

## 验证步骤

每完成一个Loop后：
1. 运行 `go build ./...` 验证编译
2. 运行 `go test ./...` 验证测试
3. 提交修复：`git commit -m "fix(phase1): 修复X项P0问题"`
4. 更修复计划状态

---

## 重要约束

**只操作工业AI项目，不要修改Hermes Agent源码！**
- 目标项目：`/Users/yqgmac/yqg/project/industrial-ai-platform`
- 禁止修改：Hermes Agent (`~/.hermes/hermes-agent`) 和 OpenClaw 源码