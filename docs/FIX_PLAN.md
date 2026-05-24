# 工业AI平台修复计划

---

## Phase 1 (Week 1): P0/CRITICAL 修复

### 后端 P0 (5项)

| ID | 文件 | 问题 | 修复方案 | 工时 |
|---|---|---|---|---|
| FIX-001 | repository/factory.go | 4个方法返回nil导致panic | 返回mock实例或error | 2h |
| FIX-002 | handler/alert_handler_new.go | 内存过滤性能问题 | 使用数据库查询+索引 | 3h |
| FIX-003 | constants.go | 硬编码连接字符串 | 从环境变量读取 | 1h |
| FIX-004 | mock_server/main.go | 硬编码JWT secret | 使用config | 1h |
| FIX-005 | handler/server_new.go | CreateRBAC返回nil | 实现完整逻辑 | 2h |

### 前端 P0 (6项)

| ID | 文件 | 问题 | 修复方案 | 工时 |
|---|---|---|---|---|
| FIX-006 | AlertsPage.tsx | localStorage key不一致 | 统一使用token | 1h |
| FIX-007 | AlertsPage.tsx | 硬编码中文 | 添加i18n调用 | 2h |
| FIX-008 | TelemetryPage.tsx | 硬编码中文 | 添加i18n调用 | 1h |
| FIX-009 | App.tsx | 硬编码标题"Industrial AI" | 使用i18n | 0.5h |
| FIX-010 | AlertsPage.tsx | 硬编码告警消息 | 使用i18n模板 | 1h |
| FIX-011 | AITeamDashboard.tsx | 硬编码"AI智能体对话" | 使用i18n | 0.5h |

### 安全 HIGH (2项)

| ID | 文件 | 问题 | 修复方案 | 工时 |
|---|---|---|---|---|
| FIX-012 | docker-compose.yml | 默认密码postgres | 使用.env文件 | 1h |
| FIX-013 | docker-compose.yml | Redis无认证 | 添加requirepass | 1h |

---

## Phase 2 (Week 2): P1/HIGH 修复

### 后端 P1 (9项)

| ID | 文件 | 问题 | 修复方案 | 工时 |
|---|---|---|---|---|
| FIX-014 | 多文件 | context.TODO()滥用 | 传递真实context | 3h |
| FIX-015 | 多文件 | deprecated函数未删除 | 清理死代码 | 2h |
| FIX-016 | auth_handler_new.go | 占位实现未完成 | 完成token刷新 | 2h |
| FIX-017 | auth_handler_new.go | ChangePassword未实现 | 实现密码修改 | 2h |
| FIX-018 | repository/base.go | 字符串拼接SQL | 表名白名单 | 1h |

### 前端 P1 (7项)

| ID | 文件 | 问题 | 修复方案 | 工时 |
|---|---|---|---|---|
| FIX-019 | AlertsPage.tsx | 类型断言 | 使用typeGuards | 1h |
| FIX-020 | TelemetryPage.tsx | 类型断言 | 使用typeGuards | 1h |
| FIX-021 | Sidebar.tsx | 菜单重建 | useMemo包裹 | 0.5h |
| FIX-022 | AITeamDashboard.tsx | React key不稳定 | 使用session_id | 0.5h |
| FIX-023 | AlertsPage.tsx | severityConfig硬编码 | i18n labels | 1h |

---

## Phase 3 (Week 3): P2/MEDIUM 修复

- 前端：重复useEffect、eslint-disable、按钮实现规范
- 安全：WAF强制启用、密码复杂度、HSTS preload

---

## 总工时估算

| Phase | 工时 |
|---|---|
| Phase 1 (P0) | ~16小时 |
| Phase 2 (P1) | ~12小时 |
| Phase 3 (P2) | ~8小时 |

---

## 执行策略

1. **并行修复**：使用3个sub-agent同时处理（后端、前端、安全）
2. **批量提交**：每完成一组相关修复后commit
3. **验证**：每phase完成后运行测试+压测

---

**待用户批准后开始执行**