# 🔧 Industrial AI Platform - 全面修复计划

> **计划日期**: 2026-05-14  
> **总发现数**: 62项  
> **预估工时**: 40+小时  
> **执行周期**: 6周 (Phase 5-8)

---

## 📋 修复计划总览

| Phase | 优先级 | 发现数 | 预估工时 | 执行周期 |
|-------|--------|--------|----------|----------|
| **Phase 5** | P0 Critical | 7项 | 10h | Week 1 |
| **Phase 6** | P1 High | 18项 | 20h | Week 2-3 |
| **Phase 7** | P2 Medium | 23项 | 12h | Week 4 |
| **Phase 8** | P3 Low + Tests | 14项 + 测试 | 15h+ | Week 5-6 |

---

## 🔴 Phase 5: P0 Critical (Week 1)

### 目标
修复可能导致生产事故的严重问题：内存泄漏、请求阻塞、类型安全

### BE-P0: 后端 (3项)

#### FIX-001: Goroutine泄漏修复 - Token黑名单
```
文件: backend/pkg/auth/token_blacklist.go
问题: cleanupExpiredEntries() goroutine永久运行无法停止
修复: 
  1. 添加 shutdown channel
  2. 在 select 中监听 context.Done()
  3. 实现 Stop() 方法
预估: 2小时
验收: goroutine 在服务关闭时正确退出
```

#### FIX-002: Goroutine泄漏修复 - Hybrid黑名单
```
文件: backend/pkg/auth/hybrid_blacklist.go
问题: checkRedisHealth() 同样的goroutine泄漏
修复:
  1. 统一实现 shutdown 机制
  2. 添加 Stop() 方法
预估: 1小时
验收: Stop() 调用后所有goroutine退出
```

#### FIX-003: Context超时控制
```
文件: backend/internal/handler/*.go (所有handler)
问题: 所有handler使用context.Background()无超时控制
修复:
  1. 改用请求传入的context
  2. 设置默认30秒超时
  3. 数据库操作使用context.WithTimeout
预估: 3小时
验收: 请求超时后正确返回504，不阻塞
```

### FE-P0: 前端 (4项)

#### FIX-004: 错误枚举值修复
```
文件: frontend/src/lib/errorHelper.ts:13
问题: UNAUTHORIZED='***' 枚举值错误
修复: 改为 UNAUTHORIZED='UNAUTHORIZED'
预估: 0.5小时
验收: 认证错误类型判断正常工作
```

#### FIX-005: 类型安全修复 - BlackBoxCenter
```
文件: frontend/src/components/BlackBoxCenter.tsx:33
问题: setSelectedRecord(res as unknown as BlackBoxRecord) 双重断言
修复:
  1. 定义API响应类型 BlackBoxRecordResponse
  2. 使用类型守卫 validateBlackBoxRecord()
预估: 1小时
验收: 类型检查通过，无运行时错误
```

#### FIX-006: useEffect依赖修复
```
文件: frontend/src/components/FleetDashboard.tsx:37-41
问题: loadData在空依赖数组[]内，导航时不刷新
修复: 添加navigation state依赖或使用事件监听
预估: 0.5小时
验收: 页面导航后数据自动刷新
```

#### FIX-007: AuthContext类型修复
```
文件: frontend/src/context/AuthContext.tsx:44
问题: setUser(response.user as User) 无验证
修复:
  1. 添加 validateUser() 函数
  2. 验证必填字段存在
预估: 1小时
验收: isAdmin计算正确，异常用户数据不进入state
```

---

## 🟠 Phase 6: P1 High (Week 2-3)

### BE-P1: 后端性能/并发 (5项)

#### FIX-008: N+1查询优化
```
文件: backend/internal/repository/*.go
问题: GetList后循环查详情，N+1性能问题
修复:
  1. tenant_repo.go: 使用JOIN获取租户详情
  2. device_repo.go: 批量查询设备状态
  3. rule_repo.go: 预加载关联规则
预估: 4小时
验收: 查询次数从O(N)降为O(1)
```

#### FIX-009: 全局变量重构
```
文件: backend/pkg/auth/*.go
问题: jwtSecret, tokenBlacklist全局变量滥用
修复:
  1. 封装为 AuthService 结构体
  2. 通过依赖注入传递
  3. 移除全局变量
预估: 3小时
验收: 所有认证功能通过DI使用，无全局变量
```

#### FIX-010: 并发安全修复
```
文件: backend/pkg/cache/memory.go
问题: map无锁读写，并发不安全
修复:
  1. 使用 sync.RWMutex 保护map
  2. Get用RLock，Set用Lock
预估: 2小时
验收: 并发测试通过，无data race
```

#### FIX-011: 配置硬编码修复
```
文件: backend/internal/config/config.go
问题: 端口8080、超时15s等硬编码
修复:
  1. 添加环境变量支持 SERVER_PORT, REQUEST_TIMEOUT
  2. 配置文件读取
预估: 2小时
验收: 可通过环境变量修改所有关键配置
```

#### FIX-012: 连接池配置化
```
文件: backend/pkg/database/connection.go
问题: MaxOpenConns=25, MaxIdleConns=5 硬编码
修复: 添加DB_MAX_OPEN, DB_MAX_IDLE环境变量
预估: 1小时
验收: 连接池参数可通过配置调整
```

### FE-P1: 前端类型/国际化 (11项)

#### FIX-013: 类型断言消除 (批量)
```
文件: DeviceManager.tsx, RuleManager.tsx, WorkOrderBoard.tsx等
问题: 大量 as Type 类型断言
修复:
  1. 定义API响应泛型类型 ApiResponse<T>
  2. 创建类型守卫函数
  3. 替换所有断言为守卫调用
预估: 4小时
验收: tsc无隐式any警告
```

#### FIX-014: useMemo优化 - IntersectionObserver
```
文件: frontend/src/lib/hooks.ts:232-243
问题: options参数每次新对象导致Observer重建
修复: 使用useMemo包裹options或useRef存储
预估: 1小时
验收: Observer创建次数大幅减少
```

#### FIX-015-021: 国际化修复 (7处)
```
文件: Sidebar.tsx, ReportCenter.tsx, SystemStatus.tsx, ROIDashboard.tsx, 
      BlackBoxCenter.tsx, Toast.tsx, ExportButton.tsx
问题: 硬编码中文文本
修复: 添加t()函数调用，补充i18n翻译键
预估: 3小时
验收: 所有UI文本通过i18n获取
```

#### FIX-022: 可访问性增强
```
文件: 所有组件
问题: 仅5处aria-label，模态框缺少role="dialog"
修复:
  1. 添加 aria-label 到所有按钮
  2. 模态框添加 role="dialog" + aria-modal
  3. 输入框添加 aria-describedby
预估: 2小时
验收: axe-core无a11y警告
```

#### FIX-023: API错误处理统一
```
文件: KnowledgeGraph.tsx, DigitalTwinPanel.tsx, DeviceDetail.tsx
问题: catch仅console.error，无用户提示
修复: 使用统一的showError toast服务
预估: 2小时
验收: 所有API错误显示用户友好提示
```

### SEC-P1: 安全修复 (1项)

#### FIX-024: CORS默认值修复
```
文件: backend/internal/config/config.go
问题: ParseOrigins返回["*"]作为默认值
修复:
  1. 生产环境强制指定origins
  2. 默认返回空数组触发错误
预估: 1小时
验收: 生产环境不允CORS *
```

---

## 🟡 Phase 7: P2 Medium (Week 4)

### BE-P2: 后端代码质量 (5项)

#### FIX-025: CRUD逻辑提取
```
文件: backend/internal/repository/*.go
问题: 相似CRUD代码重复
修复:
  1. 创建 BaseRepository[T] 泛型
  2. 提供通用CRUD方法
预估: 3小时
验收: 代码重复减少50%+
```

#### FIX-026: 魔法数字常量化
```
文件: 多处
问题: 100, 200, 15等魔法数字
修复:
  1. 定义 DefaultPageSize=20, MaxPageSize=100
  2. DefaultTimeout=15s
预估: 2小时
验收: 无硬编码数字
```

#### FIX-027: 输入验证增强
```
文件: backend/internal/handler/*.go
问题: device_id等参数无格式验证
修复:
  1. 使用Gin binding验证标签
  2. 自定义验证器UUID格式
预估: 2小时
验收: 无效输入返回400而非500
```

#### FIX-028: 日志级别配置
```
文件: backend/pkg/logger/logger.go
问题: 日志级别固定INFO
修复: 添加LOG_LEVEL环境变量
预估: 1小时
验收: 可动态调整日志级别
```

#### FIX-029: 错误处理统一
```
文件: backend/internal/service/*.go
问题: 错误处理不一致，有的返回string有的返回error
修复:
  1. 定义业务错误类型 BusinessError
  2. 统一返回error接口
预估: 2小时
验收: 所有service返回统一error类型
```

### FE-P2: 前端性能优化 (11项)

#### FIX-030-035: useMemo批量添加 (6处)
```
文件: FleetDashboard.tsx, DeviceManager.tsx, hooks.ts等
问题: 计算在渲染时执行，无memo优化
修复: 添加useMemo包裹计算逻辑
预估: 2小时
验收: 渲染次数减少
```

#### FIX-036: 批量API调用
```
文件: frontend/src/components/NotificationCenter.tsx
问题: handleMarkAllRead循环串行调用API
修复: 使用Promise.all或批量API端点
预估: 1小时
验收: 全部标记完成时间<1s
```

#### FIX-037: Tailwind动态类名
```
文件: frontend/src/components/ReportCenter.tsx:151
问题: bg-${color}-500可能被purge删除
修复: 使用完整类名映射对象
预估: 0.5小时
验收: 所有颜色正确显示
```

#### FIX-038: CRUD Hook提取
```
文件: DeviceManager.tsx, RuleManager.tsx, WorkOrderBoard.tsx
问题: 相似CRUD组件逻辑重复
修复: 创建useCRUD<T>通用Hook
预估: 4小时
验收: CRUD组件代码减少40%+
```

#### FIX-039: 类型定义统一
```
文件: frontend/src/components/FleetDashboard.tsx
问题: 本地Device/Telemetry与api.ts冲突
修复: 移除本地定义，使用types/api.ts
预估: 1小时
验收: 无类型重复定义
```

#### FIX-040: 自定义确认框
```
文件: DeviceManager.tsx, RuleManager.tsx, UserManager.tsx
问题: 使用原生confirm()
修复: 创建ConfirmDialog组件
预估: 2小时
验收: 确认框样式统一
```

### SEC-P2: 安全增强 (7项)

#### FIX-041: WebSocket认证
```
文件: backend/internal/handler/routes.go
问题: /ws端点无认证
修复:
  1. 添加JWT认证中间件
  2. 或明确为公开端点并添加限流
预估: 2小时
验收: 未认证用户无法连接WebSocket
```

#### FIX-042: 遥测端点安全
```
文件: backend/internal/handler/routes.go
问题: /devices/telemetry公开
修复: 添加设备API Key认证
预估: 2小时
验收: 无认证设备无法提交遥测
```

#### FIX-043: 密钥轮换自动化
```
文件: k8s/secrets-rotation-reminder.yaml
问题: 仅提醒不执行轮换
修复: 实现自动轮换Job
预估: 3小时
验收: 密钥每90天自动轮换
```

#### FIX-044: 输入格式验证
```
文件: backend/internal/handler/device_handler.go
问题: device_id无UUID验证
修复: 添加正则验证 ^[a-f0-9-]{36}$
预估: 1小时
验收: 无效ID返回400
```

#### FIX-045: SQL注入检测升级
```
文件: backend/pkg/database/security.go
问题: containsSQLInjectionPattern简单字符串匹配
修复: 升级为正则+编码检测
预估: 2小时
验收: 绕过测试用例全部拦截
```

#### FIX-046: 密钥输出移除
```
文件: scripts/generate-secrets.sh
问题: 打印密钥到终端
修复: 仅写入文件，移除echo
预估: 0.5小时
验收: 终端不显示密钥
```

#### FIX-047: CSRF文档补充
```
文件: docs/API_SECURITY.md (新建)
问题: JWT无CSRF保护缺少文档说明
修复: 编写安全配置指南
预估: 1小时
验收: 文档明确cookie vs header使用场景
```

---

## 🟢 Phase 8: P3 Low + 测试覆盖 (Week 5-6)

### BE-P3: 后端低优先级 (5项)

#### FIX-048-052: 代码清理
```
问题:
  - 测试代码mock错误
  - 未使用导入
  - 死代码
  - 注释缺失
  - 日志格式不一致
修复: golangci-lint自动清理 + 手动补充注释
预估: 2小时
```

### FE-P3: 前端低优先级 (12项)

#### FIX-053-064: 前端清理
```
问题:
  - 5处硬编码文本国际化
  - key使用i而非session_id
  - ×符号语义化
  - console.log移除
  - console.error统一
  - types/api.ts截断
  - ComponentType<any>泛型化
  - useMemo依赖优化
修复: 批量处理 + 代码审查
预估: 3小时
```

### SEC-P3: 安全低优先级 (4项)

#### FIX-065-068: 安全增强
```
问题:
  - 开发环境CSP宽松
  - 密钥长度日志
  - 漏洞扫描自动化
  - 环境变量存储文档
修复: 添加CI扫描 + 文档补充
预估: 1小时
```

### TEST: 测试覆盖率提升

#### TEST-001: pkg/cache测试
```
目标: backend/pkg/cache/* 从0%到80%
内容:
  - memory_test.go: CRUD、过期、并发
  - redis_test.go: 连接、重连、分布式锁
预估: 4小时
```

#### TEST-002: pkg/circuitbreaker测试
```
目标: backend/pkg/circuitbreaker/* 从0%到80%
内容: breaker_test.go: 状态转换、熔断恢复
预估: 2小时
```

#### TEST-003: pkg/auth测试
```
目标: backend/pkg/auth/* 到90%
内容:
  - jwt_helpers_test.go扩展
  - token_blacklist_test.go
预估: 3小时
```

#### TEST-004: service层测试
```
目标: backend/internal/service/* 到80%
内容:
  - tenant_service_test.go扩展
  - rbac_service_test.go扩展
预估: 4小时
```

#### TEST-005: 前端Hook测试
```
目标: frontend/src/hooks/* 到80%
内容:
  - useWebSocket.test.ts
  - useVirtualList.test.ts扩展
预估: 3小时
```

#### TEST-006: 前端核心组件测试
```
目标: frontend/src/components/* 到60%
内容:
  - AuthContext.test.tsx
  - DeviceManager.test.tsx
  - FleetDashboard.test.tsx
预估: 4小时
```

---

## 📊 执行策略

### 批量执行模式 (推荐)

使用 `delegate_task` 3个并行子代理，每个处理一组相关修复：

```
Loop 1 (Phase 5):
  - SubAgent 1: FIX-001,002,003 (后端P0)
  - SubAgent 2: FIX-004,005,006,007 (前端P0)
  - SubAgent 3: FIX-024 (安全P1)
→ 验证 → 提交 → 更新进度

Loop 2 (Phase 6 BE):
  - SubAgent 1: FIX-008,009,010 (性能/并发)
  - SubAgent 2: FIX-011,012 (配置)
  - SubAgent 3: FIX-013,014 (前端类型)
→ 验证 → 提交 → 更新进度

... 依此类推
```

### 验证标准

每个修复完成后执行：

**后端**:
```bash
go build ./...       # 编译通过
go test ./... -race  # 无data race
```

**前端**:
```bash
npm run typecheck    # 0 errors
npm run lint         # 0 warnings
```

**安全**:
```bash
# 检查配置文件正确性
# 运行安全测试套件
```

---

## 📈 预期成果

| 指标 | 当前 | Phase 5后 | Phase 8后 |
|------|------|-----------|-----------|
| Goroutine泄漏 | 存在 | ✅ 0 | ✅ 0 |
| Context超时 | ❌ 无 | ✅ 有 | ✅ 有 |
| 类型断言 | 20+处 | 15处 | 0处 |
| 国际化缺失 | 15+处 | 10处 | 0处 |
| 安全漏洞 | 12项 | 6项 | 0项 |
| 测试覆盖率 | 13.5% | 15% | 50%+ |
| 代码质量评分 | B | B+ | A |

---

## 📝 执行确认

**是否开始执行此计划？**

选项：
- **A**: 立即开始 Phase 5 (P0 Critical)
- **B**: 从 Phase 5 开始，完整执行所有Phase
- **C**: 仅执行特定Phase (请指定)
- **D**: 修改计划后再执行