# API功能清单 (API_FEATURES)

> 文档版本: 1.1.0  
> 更新日期: 2026-05-28  
> 项目: Industrial AI Platform

---

## 1. 已实现API列表

### 1.1 公共路由 (Public Routes - 无需认证)

| 端点 | 方法 | 功能描述 | 限流策略 |
|------|------|----------|----------|
| `/health` | GET | 峻康检查，返回服务状态、运行时间 | 无 |
| `/cache/status` | GET | 缓存状态监控 | 无 |
| `/ws/stats` | GET | WebSocket压缩统计 | 无 |
| `/docs/*` | GET | 静态文档服务 | 无 |
| `/ws` | GET | WebSocket实时数据连接 | WebSocketRateLimit |
| `/api/v1/auth/login` | POST | 用户登录，返回Access+Refresh Token | LoginRateLimit |
| `/api/v1/auth/register` | POST | 用户注册 | RegisterRateLimit |
| `/api/v1/auth/refresh` | POST | Token刷新 | 无 |
| `/api/v1/devices/telemetry` | POST | 遥测数据上报（设备端） | TelemetryRateLimit |

### 1.2 认证路由 (Authenticated Routes - 需JWT认证)

#### 认证管理 (Auth)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/auth/logout` | POST | 用户登出，撤销Token |
| `/api/v1/auth/change-password` | POST | 修改密码 |
| `/api/v1/auth/validate` | GET | 验证Token有效性 |

#### 设备管理 (Devices)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/devices` | GET | 设备列表（分页） |
| `/api/v1/devices` | POST | 创建设备 |
| `/api/v1/devices/latest` | GET | 获取最新遥测数据 |
| `/api/v1/devices/graph` | GET | 设备知识图谱 |
| `/api/v1/devices/:id` | GET | 获取单个设备详情 |
| `/api/v1/devices/:id` | PUT | 更新设备信息 |
| `/api/v1/devices/:id` | DELETE | 删除设备 |
| `/api/v1/devices/:id/telemetry` | GET | 设备遥测历史数据 |
| `/api/v1/devices/:id/stats` | GET | 设备统计数据 |

#### 规则管理 (Rules)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/rules` | GET | 规则列表 |
| `/api/v1/rules` | POST | 创建告警规则 |
| `/api/v1/rules/:id` | GET | 获取规则详情 |
| `/api/v1/rules/:id` | PUT | 更新规则 |
| `/api/v1/rules/:id/toggle` | PUT | 启用/禁用规则 |

#### AI智能体 (AI Agent)
| 端点 | 方法 | 功能描述 | 限流策略 |
|------|------|----------|----------|
| `/api/v1/agent/query` | POST | AI查询请求 | AgentQueryRateLimit |
| `/api/v1/ai/status` | GET | AI服务状态 |

#### 工单管理 (Work Orders)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/work-orders` | GET | 工单列表（支持状态/设备筛选） |
| `/api/v1/work-orders` | POST | 创建工单 |
| `/api/v1/work-orders/:id` | GET | 工单详情 |
| `/api/v1/work-orders/:id/status` | PUT | 更新工单状态 |

#### 通知管理 (Notifications)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/notifications` | GET | 通知列表 |
| `/api/v1/notifications/:id/read` | PUT | 标记通知已读 |

#### 黑匣子记录 (Black Box)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/blackbox` | GET | 黑匣子事件列表 |
| `/api/v1/blackbox/:id/data` | GET | 黑匣子详细数据 |

#### 报表中心 (Reports)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/reports` | GET | 报表列表 |
| `/api/v1/reports/generate` | POST | 生成报表 |

#### ROI分析 (ROI Stats)
| 端点 | 方法 | 功能描述 | 限流策略 |
|------|------|----------|----------|
| `/api/v1/roi/stats` | GET | ROI统计数据 | ROIStatsRateLimit |

#### 数据导出 (Export)
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/reports/devices/export` | GET | 设备数据导出 (CSV/JSON) |
| `/api/v1/reports/alerts/export` | GET | 告警数据导出 |
| `/api/v1/reports/roi/export` | GET | ROI数据导出 |

### 1.3 管理员路由 (Admin Routes - 需Admin权限)

| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/admin/users` | GET | 用户列表 |
| `/api/v1/admin/users` | POST | 创建用户 |
| `/api/v1/admin/users/:id` | DELETE | 删除用户 |
| `/api/v1/system/status` | GET | 系统状态 |
| `/api/v1/rules/:id` | DELETE | 删除规则 |
| `/api/v1/tenants` | POST | 创建租户 |
| `/api/v1/tenants` | GET | 租户列表 |
| `/api/v1/tenants/:id` | DELETE | 删除租户 |

### 1.4 租户路由 (Tenant Routes)

| 端点 | 方法 | 功能描述 | 权限 |
|------|------|----------|------|
| `/api/v1/tenants/:id` | GET | 获取租户详情 | 认证用户 |
| `/api/v1/tenants/:id` | PUT | 更新租户信息 | 认证用户 |

### 1.5 RBAC路由 (角色权限管理)

#### 认证用户路由
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/roles` | GET | 角色列表 |
| `/api/v1/roles/:id` | GET | 角色详情（含权限） |
| `/api/v1/permissions` | GET | 权限列表 |
| `/api/v1/users/:id/roles` | GET | 用户角色列表 |

#### 管理员路由
| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/api/v1/roles` | POST | 创建角色 |
| `/api/v1/roles/:id` | PUT | 更新角色 |
| `/api/v1/roles/:id` | DELETE | 删除角色 |
| `/api/v1/users/:id/roles` | POST | 分配角色 |
| `/api/v1/users/:id/roles/:role_id` | DELETE | 移除角色 |
| `/api/v1/roles/:id/permissions` | POST | 分配权限 |
| `/api/v1/roles/:id/permissions/:perm_id` | DELETE | 移除权限 |

### 1.6 Prometheus监控

| 端点 | 方法 | 功能描述 |
|------|------|----------|
| `/metrics` | GET | Prometheus指标导出 |

---

## 2. API版本管理

### 2.1 版本策略

当前API版本: **v1**

- **URL路径版本**: `/api/v1/...`
- **向后兼容**: 新版本发布时，旧版本API保持至少6个月的支持周期
- **版本弃用通知**: 通过响应头 `X-API-Version-Deprecated` 提示

### 2.2 版本规划

| 版本 | 状态 | 说明 |
|------|------|------|
| v1 | 当前稳定版本 | 所有核心功能已实现 |
| v2 (规划) | 待开发 | GraphQL支持、批量操作优化 |

### 2.3 API响应格式

所有API响应遵循统一格式:

```json
{
  "success": true,
  "data": {...},
  "message": "操作成功",
  "error": null
}
```

错误响应:
```json
{
  "success": false,
  "error": "错误描述",
  "code": "ERROR_CODE"
}
```

---

## 3. 认证机制说明

### 3.1 JWT Token认证

**认证流程**:
1. 用户登录 → 获取 `AccessToken` + `RefreshToken`
2. 使用 `AccessToken` 访问API (Header: `Authorization: Bearer <token>`)
3. Token过期 → 使用 `RefreshToken` 刷新获取新Token Pair

**Token配置**:
- AccessToken有效期: 15分钟
- RefreshToken有效期: 7天
- 签名算法: HS256
- Token版本验证: 支持Token撤销机制

### 3.2 Token Claims结构

```json
{
  "user_id": 1,
  "username": "admin",
  "role": "admin",
  "tenant_id": "default",
  "token_version": 1,
  "exp": 1716000000,
  "type": "access" // 或 "refresh"
}
```

### 3.3 认证中间件

| 中间件 | 功能 |
|--------|------|
| `AuthRequired` | JWT验证，提取用户信息 |
| `AdminRequired` | 管理员权限检查 |
| `TenantRequired` | 租户隔离验证 |

### 3.4 限流策略

| 限流器 | 触发端点 | 限制 |
|--------|----------|------|
| LoginRateLimit | `/auth/login` | 5次/分钟 |
| RegisterRateLimit | `/auth/register` | 3次/小时 |
| TelemetryRateLimit | `/devices/telemetry` | 100次/分钟 |
| AgentQueryRateLimit | `/agent/query` | 10次/分钟 |
| ROIStatsRateLimit | `/roi/stats` | 20次/分钟 |
| WebSocketRateLimit | `/ws` | 10连接/分钟 |

---

## 4. 扩展建议

### 4.1 功能扩展

#### 高优先级
- [ ] **批量操作API**: POST `/api/v1/devices/batch` 支持批量创建/更新
- [ ] **异步任务API**: 长时间操作返回任务ID，通过 `/tasks/:id` 查询进度
- [ ] **GraphQL端点**: `/graphql` 提供灵活查询能力

#### 中优先级
- [ ] **Webhook回调**: POST `/api/v1/webhooks` 配置事件回调
- [ ] **API密钥管理**: 支持服务端API密钥认证（替代JWT）
- [ ] **审计日志查询**: GET `/api/v1/audit-logs` 查询操作记录

#### 低优先级
- [ ] **API文档自动生成**: Swagger/OpenAPI集成
- [ ] **版本自动发现**: GET `/api/version` 返回支持的版本列表
- [ ] **请求签名验证**: HMAC签名防篡改

### 4.2 性能优化建议

1. **响应缓存**: 对GET请求添加ETag/Cache-Control支持
2. **分页优化**: 大数据集使用游标分页替代偏移分页
3. **连接池调优**: 根据负载调整数据库连接池参数
4. **批量查询**: 减少N+1查询问题

### 4.3 安全增强建议

1. **API密钥轮换**: 定期轮换JWT签名密钥
2. **请求签名**: 关键操作添加请求签名验证
3. **IP白名单**: 管理API添加IP访问控制
4. **审计增强**: 记录所有API调用详情

---

## 附录

### A. HTTP状态码规范

| 状态码 | 使用场景 |
|--------|----------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证/Token无效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 429 | 请求频率超限 |
| 500 | 服务器内部错误 |

### B. 错误代码列表

| 错误代码 | 描述 |
|----------|------|
| INVALID_REQUEST | 请求格式错误 |
| VALIDATION_ERROR | 数据验证失败 |
| NOT_FOUND | 资源不存在 |
| DATABASE_ERROR | 数据库操作失败 |
| CREATE_FAILED | 创建操作失败 |
| UPDATE_FAILED | 更新操作失败 |
| AUTH_ERROR | 认证失败 |
| RATE_LIMIT_EXCEEDED | 超过限流阈值 |

---

*文档维护: 开发团队*  
*最后审核: 2026-05-28*