# E2E 测试报告 (2026-05-17)

## 测试环境

| 服务 | 状态 | 版本 |
|------|------|------|
| PostgreSQL | ✅ 运行中 | v16 |
| Redis | ✅ 运行中 | Latest |
| Backend | ✅ 端口 8080 | Go 1.22+ |
| Frontend | ✅ 端口 5173 | React 19 + Vite |

## 测试结果

### 登录流程 (auth/login.spec.ts)

| 测试用例 | 结果 | 时间 |
|----------|------|------|
| 成功登录 - 管理员 | ✅ PASS | 2.7s |
| 成功登录 - 操作员 | ✅ PASS | 3.1s |
| 登录失败 - 错误密码 | ⏭️ SKIP | - |
| 登录失败 - 空用户名 | ✅ PASS | 1.7s |
| 登录失败 - 空密码 | ✅ PASS | 1.8s |
| 登出功能 | ✅ PASS | 2.7s |
| 登录页记住我功能 | ✅ PASS | 1.7s |
| 登录后页面跳转 | ✅ PASS | 2.2s |
| 移动端登录页面 | ✅ PASS | 1.7s |
| 平板端登录页面 | ✅ PASS | 1.7s |

**结果：9 passed, 1 skipped**

### AI Agent (ai/ai-agent.spec.ts)

| 测试用例 | 结果 | 备注 |
|----------|------|------|
| 显示 AI 对话界面 | ✅ PASS | 界面渲染正常 |
| 发送问题并接收回答 | ⏭️ SKIP | 需真实 LLM API |
| 多轮对话 | ⏭️ SKIP | 需真实 LLM API |
| 设备相关查询 | ⏭️ SKIP | 需真实 LLM API |
| 建议问题 | ✅ PASS | UI 功能正常 |
| 空问题验证 | ✅ PASS | 表单验证正常 |
| 长问题处理 | ✅ PASS | UI 处理正常 |

**结果：3 passed, 7 skipped (需真实 LLM API)**

### 设备列表 (devices/device-list.spec.ts)

| 测试用例 | 结果 | 原因 |
|----------|------|------|
| 显示设备列表 | ❌ FAIL | 数据库设备数为 0 |
| 搜索设备 | ❌ TIMEOUT | 无设备数据 |

**问题：数据库设备表为空，需添加测试数据**

### 告警/遥测/报告/国际化

部分测试因设备数据缺失而超时或失败。

## API 验证

| API | 状态 | 响应时间 |
|-----|------|----------|
| /health | ✅ 200 | ~1ms |
| /api/v1/auth/login | ✅ 200 | ~50ms |
| /api/v1/devices | ✅ 200 | ~1ms (空数据) |
| /ws | ✅ 连接成功 | WebSocket 正常 |

## 有效设备类型

创建测试设备需使用以下类型：
- CNC (加工中心)
- InjectionMolder (注塑机)
- AssemblyRobot (装配机器人)
- Conveyor (输送带)
- Sensor/sensor (传感器)
- gauge (仪表)
- PLC (可编程控制器)
- robot (机器人)
- motor (电机)
- pump (泵)
- valve (阀门)
- heater (加热器)
- cooler (冷却器)

## 下一步

1. 添加测试设备数据到数据库
2. 重新运行完整 E2E 测试
3. 验证设备 CRUD 功能
4. 验证告警触发逻辑

## 测试用户

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | Admin@123456 | 管理员 |
| operator | Operator@123 | 操作员 |