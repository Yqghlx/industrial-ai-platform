# API 错误码完整列表

本文档详细说明工业AI代理平台API返回的所有错误码、HTTP状态码、错误消息及解决方案。

## 错误响应格式

所有API错误响应遵循统一的JSON格式：

```json
{
  "error": "错误消息描述",
  "code": "ERROR_CODE",
  "detail": "详细错误信息（可选）"
}
```

## 按模块分类的错误码

### 1. 认证模块 (AUTH)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `AUTH_FAILED` | 401 | 用户名或密码错误 | 登录凭据无效 | 检查用户名和密码是否正确 |
| `INVALID_AUTH_FORMAT` | 401 | 无效的认证格式 | Authorization头部格式错误 | 使用 `Bearer <token>` 格式 |
| `TOKEN_EXPIRED` | 401 | Token已过期 | JWT Token已过期 | 使用Refresh Token获取新的Access Token |
| `TOKEN_INVALID` | 401 | 无效的Token | JWT Token无效或被篡改 | 重新登录获取新Token |
| `SESSION_EXPIRED` | 401 | 会话已过期 | 用户会话已过期 | 重新登录 |
| `ACCOUNT_LOCKED` | 403 | 账户已被锁定 | 连续登录失败次数过多 | 等待15分钟后重试或联系管理员 |
| `PASSWORD_WEAK` | 400 | 密码强度不足 | 密码不符合安全要求 | 使用至少12位，包含大小写字母、数字和特殊字符的密码 |
| `PASSWORD_MISMATCH` | 400 | 旧密码不正确 | 修改密码时旧密码验证失败 | 确认旧密码正确 |
| `UNAUTHORIZED` | 401 | 未授权访问 | 缺少认证信息 | 在请求头中添加有效的Bearer Token |

### 2. 设备模块 (DEVICE)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `DEVICE_NOT_FOUND` | 404 | 设备不存在 | 指定的设备ID不存在 | 检查设备ID是否正确 |
| `DUPLICATE_DEVICE_ID` | 409 | 设备ID已存在 | 创建设备时ID重复 | 使用唯一的设备ID |
| `INVALID_DEVICE_ID` | 400 | 无效的设备ID格式 | 设备ID格式不符合要求 | 使用有效的UUID或指定格式的ID |
| `INVALID_DEVICE_TYPE` | 400 | 无效的设备类型 | 设备类型不在允许列表中 | 使用有效的设备类型：CNC, InjectionMolder, AssemblyRobot, Conveyor, Sensor, Other |
| `DEVICE_AUTH_REQUIRED` | 401 | 设备认证失败 | 设备遥测数据上报缺少认证 | 添加有效的设备认证密钥 |
| `INVALID_DEVICE_KEY` | 401 | 无效的设备密钥 | 设备认证密钥无效 | 检查设备密钥是否正确 |
| `DEVICE_OFFLINE` | 503 | 设备离线 | 设备当前处于离线状态 | 等待设备上线后重试 |

### 3. 权限控制模块 (RBAC)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `FORBIDDEN` | 403 | 访问被拒绝 | 用户没有执行该操作的权限 | 联系管理员分配相应权限 |
| `INSUFFICIENT_PERMISSIONS` | 403 | 权限不足 | 用户角色权限不够 | 需要更高级别的角色权限 |
| `ROLE_NOT_FOUND` | 404 | 角色不存在 | 指定的角色ID不存在 | 检查角色ID是否正确 |
| `PERMISSION_DENIED` | 403 | 权限拒绝 | 用户没有特定资源的访问权限 | 申请相应的资源访问权限 |
| `ADMIN_REQUIRED` | 403 | 需要管理员权限 | 操作需要管理员角色 | 使用管理员账户执行操作 |

### 4. 租户模块 (TENANT)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `TENANT_NOT_FOUND` | 404 | 租户不存在 | 指定的租户ID不存在 | 检查租户ID是否正确 |
| `INVALID_TENANT_ID` | 400 | 无效的租户ID格式 | 租户ID格式不符合要求 | 使用有效的UUID格式 |
| `TENANT_SUSPENDED` | 403 | 租户已被暂停 | 租户账户已被暂停 | 联系平台管理员恢复租户 |
| `TENANT_DELETED` | 410 | 租户已删除 | 租户已被删除 | 联系平台管理员 |
| `QUOTA_EXCEEDED` | 403 | 配额已超限 | 租户资源配额已用完 | 升级租户套餐或清理资源 |
| `PLAN_LIMIT_REACHED` | 403 | 套餐限制 | 当前套餐不支持此操作 | 升级租户套餐 |

### 5. 告警模块 (ALERT)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `RULE_NOT_FOUND` | 404 | 告警规则不存在 | 指定的告警规则不存在 | 检查规则ID是否正确 |
| `INVALID_RULE_CONFIG` | 400 | 无效的规则配置 | 告警规则配置参数无效 | 检查阈值和操作符配置 |
| `ALERT_NOT_FOUND` | 404 | 告警不存在 | 指定的告警记录不存在 | 检查告警ID是否正确 |
| `DUPLICATE_RULE_NAME` | 409 | 规则名称已存在 | 同一租户下规则名称重复 | 使用唯一的规则名称 |

### 6. 遥测数据模块 (TELEMETRY)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `INVALID_TIMESTAMP` | 400 | 无效的时间戳 | 时间戳格式错误或超出范围 | 使用ISO 8601格式的时间戳 |
| `MISSING_FIELD` | 400 | 缺少必填字段 | 请求缺少必填字段 | 检查请求体是否完整 |
| `TELEMETRY_TOO_LARGE` | 413 | 遥测数据过大 | 单次上传数据量超过限制 | 分批上传数据 |
| `INVALID_METRIC_VALUE` | 400 | 无效的指标值 | 指标值超出合理范围 | 检查传感器数据是否正常 |

### 7. AI智能体模块 (AGENT)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `EMPTY_QUERY` | 400 | 查询不能为空 | AI查询内容为空 | 提供有效的查询内容 |
| `AI_SERVICE_ERROR` | 503 | AI服务暂时不可用 | 后端AI服务异常 | 稍后重试或联系技术支持 |
| `AI_TIMEOUT` | 504 | AI查询超时 | AI处理请求超时 | 简化查询或稍后重试 |
| `SESSION_NOT_FOUND` | 404 | 会话不存在 | AI会话已过期或不存在 | 创建新的会话 |

### 8. 工单模块 (WORK_ORDER)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `WORK_ORDER_NOT_FOUND` | 404 | 工单不存在 | 指定的工单不存在 | 检查工单ID是否正确 |
| `INVALID_STATUS_TRANSITION` | 400 | 无效的状态转换 | 工单状态转换不合法 | 按照工作流程正确转换状态 |
| `ASSIGNEE_NOT_FOUND` | 404 | 指派人不存在 | 指定的指派人不存在 | 检查指派人ID是否正确 |

### 9. 报告模块 (REPORT)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `REPORT_NOT_FOUND` | 404 | 报告不存在 | 指定的报告不存在 | 检查报告ID是否正确 |
| `REPORT_GENERATION_FAILED` | 500 | 报告生成失败 | AI报告生成过程中出错 | 稍后重试或联系技术支持 |

### 10. 导出模块 (EXPORT)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `EXPORT_FORMAT_NOT_SUPPORTED` | 400 | 不支持的导出格式 | 指定的导出格式不被支持 | 使用CSV或JSON格式 |
| `EXPORT_TOO_LARGE` | 413 | 导出数据量过大 | 导出数据超过限制 | 缩小时间范围或分批导出 |
| `EXPORT_FAILED` | 500 | 导出失败 | 导出过程中发生错误 | 稍后重试 |

### 11. 存储模块 (STORAGE)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `STORAGE_ERROR` | 500 | 存储服务错误 | 数据库或存储服务异常 | 稍后重试或联系技术支持 |
| `DATABASE_ERROR` | 500 | 数据库操作失败 | 数据库操作异常 | 稍后重试或联系技术支持 |
| `CACHE_ERROR` | 500 | 缓存服务错误 | Redis缓存服务异常 | 稍后重试 |

### 12. 通用错误 (GENERAL)

| 错误码 | HTTP状态 | 错误消息 | 描述 | 解决方案 |
|--------|----------|----------|------|----------|
| `INVALID_INPUT` | 400 | 输入参数无效 | 请求参数验证失败 | 检查请求参数格式和类型 |
| `VALIDATION_ERROR` | 400 | 验证失败 | 请求数据验证失败 | 查看detail字段了解具体错误 |
| `NOT_FOUND` | 404 | 资源不存在 | 请求的资源不存在 | 检查请求路径和资源ID |
| `CONFLICT` | 409 | 资源冲突 | 资源状态冲突（如并发修改） | 刷新资源状态后重试 |
| `RATE_LIMITED` | 429 | 请求过于频繁 | 触发API限流 | 降低请求频率，稍后重试 |
| `INTERNAL_ERROR` | 500 | 内部服务器错误 | 服务器内部错误 | 稍后重试或联系技术支持 |
| `SERVICE_UNAVAILABLE` | 503 | 服务暂时不可用 | 服务正在维护或过载 | 稍后重试 |

## HTTP状态码对照表

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 200 | OK | 请求成功 |
| 201 | Created | 资源创建成功 |
| 204 | No Content | 成功但无返回内容 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未认证 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 413 | Payload Too Large | 请求体过大 |
| 429 | Too Many Requests | 请求限流 |
| 500 | Internal Server Error | 服务器内部错误 |
| 503 | Service Unavailable | 服务不可用 |
| 504 | Gateway Timeout | 网关超时 |

## 限流配置

各API端点的限流策略：

| 端点类型 | 限流策略 | 窗口期 |
|----------|----------|--------|
| 登录 (`/auth/login`) | 5次/IP | 15分钟 |
| 注册 (`/auth/register`) | 3次/IP | 1小时 |
| 遥测上报 (`/devices/telemetry`) | 100次/设备 | 1分钟 |
| AI查询 (`/agent/query`) | 20次/用户 | 1分钟 |
| ROI统计 (`/roi/stats`) | 10次/用户 | 1分钟 |
| WebSocket连接 | 10次/IP | 1分钟 |
| 默认API | 100次/用户 | 1分钟 |

## 错误处理最佳实践

### 客户端建议

1. **检查HTTP状态码**: 先根据状态码判断请求是否成功
2. **解析错误码**: 使用`code`字段进行程序化错误处理
3. **展示友好消息**: 使用`error`字段向用户展示友好的错误消息
4. **实现重试逻辑**: 对于`RATE_LIMITED`和`SERVICE_UNAVAILABLE`错误，实现指数退避重试
5. **记录详细信息**: 开发环境下记录`detail`字段用于调试

### 错误处理示例

```javascript
async function handleApiResponse(response) {
  if (!response.ok) {
    const error = await response.json();
    
    switch (error.code) {
      case 'TOKEN_EXPIRED':
      case 'SESSION_EXPIRED':
        // 尝试刷新Token
        await refreshToken();
        return retryOriginalRequest();
        
      case 'RATE_LIMITED':
        // 等待后重试
        await sleep(getRetryAfter(response));
        return retryOriginalRequest();
        
      case 'AUTH_FAILED':
      case 'ACCOUNT_LOCKED':
        // 重定向到登录页
        redirectToLogin();
        break;
        
      default:
        // 显示错误消息
        showErrorToast(error.error);
    }
  }
  
  return response.json();
}
```

## 版本历史

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| 1.0.0 | 2024-01-15 | 初始版本 |