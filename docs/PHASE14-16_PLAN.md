# 📋 Phase 14-16: 后续工作执行计划

> **计划日期**: 2026-05-14  
> **目标**: 处理所有后续工作项 (P0+P1+P2)  
> **总工作项**: 13项  
> **预估工时**: 39-55小时

---

## 📊 执行概览

| Phase | 优先级 | 工作项数 | 预估工时 | 执行周期 |
|-------|--------|----------|----------|----------|
| **Phase 14** | P0 必须做 | 4项 | 8-13h | 立即执行 |
| **Phase 15** | P1 建议做 | 4项 | 21-28h | Phase 14后 |
| **Phase 16** | P2 可选做 | 5项 | 10-14h | Phase 15后 |

---

## 🔴 Phase 14: P0 必须做 (4项)

### P0-001: 修复失败测试

**当前失败测试**:

| 模块 | 测试 | 问题类型 |
|------|------|----------|
| `internal/config` | TestConfig_GetCORSOrigins | CORS配置逻辑 |
| `internal/config` | TestConfig_GetWarnings | 配置警告逻辑 |
| `internal/handler` | TestListRules_ServiceError | Handler错误处理 |
| `internal/repository` | TestDeviceRepository_SQLQueryPatterns | SQL模式匹配 |
| `internal/repository` | TestRBACRepo_CreateRole | RBAC仓库创建 |
| `internal/repository` | TestRBACRepo_GetRolePermissions | 权限查询 |
| `pkg/auth` | TestAuthBypass_TokenValidation | JWT认证多测试失败 |

**修复策略**:
1. 读取每个失败测试文件
2. 分析测试期望 vs 实际行为
3. 修复代码逻辑或更新测试预期
4. 验证所有测试通过

---

### P0-002: 生产配置验证脚本

**创建验证脚本**: `scripts/verify-production-config.sh`

检查项:
- JWT_SECRET 配置存在且长度>=32
- DATABASE_URL 包含 sslmode=require
- CORS_ORIGINS 不包含 `*`
- 所有K8s Secret已配置真实值
- TLS证书有效期检查
- 容器非root用户运行

---

### P0-003: 密钥配置指南

**创建文档**: `docs/SECRET_CONFIGURATION.md`

内容:
- 每个Secret的作用说明
- 生成安全密钥的命令
- 替换占位符的步骤
- 配置验证方法

**更新**: `infra/k8s/secrets.yaml`
- 添加注释说明
- 保持占位符(生产时替换)

---

### P0-004: TODO修复

**文件**: `backend/internal/ws/broadcaster.go:179`

当前代码:
```go
return "2026-05-14T00:00:00Z" // TODO: use time.Now().Format(time.RFC3339)
```

修复为:
```go
return time.Now().Format(time.RFC3339)
```

---

## 🟠 Phase 15: P1 建议做 (4项)

### P1-001: 测试覆盖率提升

**目标模块**:

| 模块 | 当前 | 目标 | 新增测试 |
|------|------|------|----------|
| `pkg/cache` | 30.9% | 80% | redis_test.go |
| `internal/config` | 53.3% | 80% | 修复后扩展 |
| `internal/middleware` | 11.8% | 80% | 扩展测试 |
| `internal/model` | 0% | 50% | 模型验证测试 |
| `internal/database` | 0% | 70% | 连接池测试 |

**创建测试文件**:
- `pkg/cache/redis_test.go` (Redis缓存测试)
- `internal/middleware/auth_test.go` (认证中间件测试)
- `internal/middleware/ratelimit_test.go` (扩展)
- `internal/model/device_test.go` (设备模型测试)
- `internal/model/user_test.go` (用户模型测试)
- `internal/database/connection_test.go` (连接池测试)

---

### P1-002: 前端测试扩展

**新增测试文件**:

| 文件 | 测试内容 |
|------|----------|
| `FleetDashboard.test.tsx` | 设备仪表盘渲染、数据刷新 |
| `RuleManager.test.tsx` | 规则CRUD、启用禁用 |
| `NotificationCenter.test.tsx` | 通知列表、标记已读 |
| `WorkOrderBoard.test.tsx` | 工单创建、状态更新 |
| `AuthContext.test.tsx` | 认证状态、登录登出 |
| `hooks/useWebSocket.test.ts` | WebSocket连接、消息处理 |

---

### P1-003: API文档生成

**生成 OpenAPI/Swagger 文档**:

创建文件:
- `docs/API_OPENAPI.yaml` - OpenAPI 3.0规范
- `docs/API_ERROR_CODES.md` - 错误码完整列表
- `docs/API_EXAMPLES.md` - API请求示例
- `docs/WEBSOCKET_PROTOCOL.md` - WebSocket消息格式

---

### P1-004: E2E测试实现

**创建 E2E 测试**: `tests/e2e/`

测试场景:
- `login_flow.spec.ts` - 用户登录流程
- `device_flow.spec.ts` - 设备管理流程
- `alert_flow.spec.ts` - 告警流程
- `workorder_flow.spec.ts` - 工单流程

---

## 🟡 Phase 16: P2 可选做 (5项)

### P2-001: 性能基准测试

**创建基准测试**: `benchmarks/`

| 文件 | 内容 |
|------|------|
| `api_bench_test.go` | API响应时间基准 |
| `websocket_bench_test.go` | WebSocket连接性能 |
| `db_bench_test.go` | 数据库查询性能 |
| `cache_bench_test.go` | 缓存效率测试 |

---

### P2-002: 监控告警调优

**优化配置**: `infra/k8s/monitoring/`

| 配置 | 优化内容 |
|------|----------|
| `prometheus-values.yaml` | 采集频率、保留策略 |
| `alertmanager-config.yaml` | 告警阈值、通知渠道 |
| `grafana-dashboards/` | 业务定制仪表盘 |

---

### P2-003: 国际化完善

**完善 i18n**:

| 文件 | 内容 |
|------|------|
| `locales/zh.json` | 中文翻译补充 |
| `locales/en.json` | 英文翻译补充 |
| `errorHelper.ts` | API错误消息国际化 |
| `dateFormatter.ts` | 日期格式化工具 |

---

### P2-004: 安全扫描验证

**执行扫描验证**:

| 扫描 | 命令 |
|------|------|
| CodeQL | GitHub Actions运行结果 |
| Trivy | `trivy image <image>` |
| govulncheck | `govulncheck ./...` |
| npm audit | `npm audit` |

**创建文档**: `docs/SECURITY_SCAN_REPORT.md`

---

### P2-005: 功能扩展建议

**创建扩展规划**: `docs/FUNCTION_EXTENSION_PLAN.md`

内容:
- 批量操作API设计
- 异步任务系统架构
- GraphQL支持方案
- 移动端适配建议

---

## 🚀 执行策略

### 并行执行模式

```
Loop 1 (Phase 14): P0必须做
├── SubAgent 1: 修复config测试
├── SubAgent 2: 修复handler/repository测试
├── SubAgent 3: 修复auth测试 + TODO修复
└── MainAgent: 生产配置验证脚本 + 密钥配置文档
→ 验证 → 提交

Loop 2 (Phase 15-001): 测试覆盖率
├── SubAgent 1: pkg/cache redis测试
├── SubAgent 2: middleware测试扩展
├── SubAgent 3: model/database测试
→ 验证 → 提交

Loop 3 (Phase 15-002): 前端测试
├── SubAgent 1: FleetDashboard + RuleManager测试
├── SubAgent 2: NotificationCenter + WorkOrderBoard测试
├── SubAgent 3: AuthContext + WebSocket测试
→ 验证 → 提交

Loop 4 (Phase 15-003,004): API文档 + E2E
├── SubAgent 1: OpenAPI文档生成
├── SubAgent 2: 错误码 + WebSocket协议文档
├── SubAgent 3: E2E测试实现
→ 验证 → 提交

Loop 5 (Phase 16): P2可选做
├── SubAgent 1: 性能基准测试
├── SubAgent 2: 监控调优 + 国际化
├── SubAgent 3: 安全扫描 + 功能扩展文档
→ 验证 → 提交
```

---

## ✅ 验证标准

### Phase 14完成标准
```
✅ go test ./... 全部通过
✅ 生产配置验证脚本可执行
✅ 密钥配置文档完整
✅ TODO项已修复
```

### Phase 15完成标准
```
✅ 测试覆盖率 >= 60%
✅ 前端测试文件 >= 18个
✅ OpenAPI文档完整
✅ E2E测试可运行
```

### Phase 16完成标准
```
✅ 性能基准数据收集完成
✅ 监控配置优化
✅ 国际化完整
✅ 安全扫描报告生成
```

---

## 📈 预期成果

| 指标 | 当前 | Phase 14后 | Phase 15后 | Phase 16后 |
|------|------|------------|------------|------------|
| 测试通过率 | 失败多个 | ✅ 100% | ✅ 100% | ✅ 100% |
| 测试覆盖率 | 47.3% | 47.3% | ✅ **60%+** | ✅ 65%+ |
| 前端测试数 | 13 | 13 | ✅ **18+** | ✅ 20+ |
| 文档完整度 | 33个 | ✅ 35个 | ✅ 40个 | ✅ 45个 |
| 性能基准 | 无 | 无 | 无 | ✅ 有 |
| 安全扫描 | 配置 | 配置 | 配置 | ✅ 验证 |

---

## 📝 开始执行

**总工作项**: 13项
**预估总工时**: 39-55小时
**预计执行时间**: ~60分钟 (并行执行)

准备开始自动执行...