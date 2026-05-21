# 安全审计日志指南

> **Industrial AI Platform 安全审计日志最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 安全审计概述

Phase 4 P2 安全审计目标：

| 指标 | 当前状态 | 目标 |
|------|---------|------|
| **操作审计** | 无 | 全量审计 |
| **访问日志** | 基础日志 | 结构化审计 |
| **合规报告** | 手动 | 自动生成 |
| **审计查询** | grep | API 查询 |

---

## 🔄 安全审计架构

### 审计日志流程

```
┌─────────────────────────────────────────┐
│  用户操作                                │
│  - 登录/注销                             │
│  - 数据访问/修改                         │
│  - 配置变更                              │
│  - 权限操作                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  审计日志中间件                          │
│  - 拦截操作                              │
│  - 提取审计信息                          │
│  - 生成审计日志                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  审计日志存储                            │
│  - PostgreSQL audit_logs 表             │
│  - Loki 日志系统                         │
│  - 保留 1 年                             │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  审计日志分析                            │
│  - 实时查询 API                          │
│  - 合规报告生成                          │
│  - 异常行为检测                          │
│  - Grafana 可视化                        │
└─────────────────────────────────────────┘
```

---

## 📝 审计日志格式

### JSON 审计日志结构

```json
{
  "audit_id": "audit-20260513-001",
  "timestamp": "2026-05-13T20:30:45.123Z",
  "event_type": "data_access",
  "event_category": "device",
  "severity": "info",
  "user_id": "user-001",
  "tenant_id": "tenant-001",
  "session_id": "session-001",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0",
  "resource_type": "device",
  "resource_id": "device-001",
  "action": "read",
  "operation": "GET /api/v1/devices/device-001",
  "request_id": "req-001",
  "trace_id": "trace-001",
  "before_state": null,
  "after_state": null,
  "changes": null,
  "result": "success",
  "error_message": null,
  "duration_ms": 45.6,
  "metadata": {
    "device_count": 10,
    "page": 1
  }
}
```

### 审计日志字段规范

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `audit_id` | string | ✅ | 审计日志唯一 ID |
| `timestamp` | string | ✅ | ISO8601 时间戳 |
| `event_type` | string | ✅ | 事件类型 |
| `event_category` | string | ✅ | 事件分类 |
| `severity` | string | ✅ | 严重程度 (info/warn/critical) |
| `user_id` | string | ✅ | 操作用户 ID |
| `tenant_id` | string | ✅ | 租户 ID |
| `ip_address` | string | ✅ | 客户端 IP |
| `resource_type` | string | ❌ | 资源类型 |
| `resource_id` | string | ❌ | 资源 ID |
| `action` | string | ✅ | 操作类型 (read/write/delete) |
| `operation` | string | ✅ | 操作详情 |
| `result` | string | ✅ | 操作结果 (success/failure) |
| `changes` | object | ❌ | 变更内容 |
| `before_state` | object | ❌ | 操作前状态 |
| `after_state` | object | ❌ | 操作后状态 |

---

## 🔧 审计事件类型

### 事件类型分类

| 类别 | 事件类型 | 说明 |
|------|---------|------|
| **认证** | `auth.login` | 用户登录 |
| | `auth.logout` | 用户注销 |
| | `auth.failed` | 认证失败 |
| | `auth.token_refresh` | Token 刷新 |
| **授权** | `authz.grant` | 权限授予 |
| | `authz.revoke` | 权限撤销 |
| | `authz.check` | 权限检查 |
| **数据访问** | `data.read` | 数据读取 |
| | `data.write` | 数据写入 |
| | `data.delete` | 数据删除 |
| | `data.export` | 数据导出 |
| **配置变更** | `config.create` | 配置创建 |
| | `config.update` | 配置更新 |
| | `config.delete` | 配置删除 |
| **系统操作** | `system.start` | 系统启动 |
| | `system.stop` | 系统停止 |
| | `system.restart` | 系统重启 |
| **安全事件** | `security.alert` | 安全告警 |
| | `security.violation` | 安全违规 |
| | `security.blocked` | 操作阻止 |

---

## 📊 PostgreSQL 审计表设计

### audit_logs 表

```sql
CREATE TABLE audit_logs (
    audit_id VARCHAR(64) PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    event_category VARCHAR(32) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'info',
    user_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64),
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(256),
    resource_type VARCHAR(64),
    resource_id VARCHAR(64),
    action VARCHAR(16) NOT NULL,
    operation VARCHAR(256) NOT NULL,
    request_id VARCHAR(64),
    trace_id VARCHAR(64),
    before_state JSONB,
    after_state JSONB,
    changes JSONB,
    result VARCHAR(16) NOT NULL,
    error_message TEXT,
    duration_ms FLOAT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_resource_type ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_result ON audit_logs(result);
CREATE INDEX idx_audit_ip_address ON audit_logs(ip_address);

-- 分区 (按时间)
CREATE TABLE audit_logs_2026_q1 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');
CREATE TABLE audit_logs_2026_q2 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-07-01');
CREATE TABLE audit_logs_2026_q3 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-07-01') TO ('2026-10-01');
CREATE TABLE audit_logs_2026_q4 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-10-01') TO ('2027-01-01');
```

---

## 🔧 Go 审计日志实现

### 审计日志服务

```go
package audit

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/google/uuid"
)

// AuditLog 审计日志结构
type AuditLog struct {
    AuditID      string                 `json:"audit_id"`
    Timestamp    time.Time              `json:"timestamp"`
    EventType    string                 `json:"event_type"`
    EventCategory string                `json:"event_category"`
    Severity     string                 `json:"severity"`
    UserID       string                 `json:"user_id"`
    TenantID     string                 `json:"tenant_id"`
    SessionID    string                 `json:"session_id"`
    IPAddress    string                 `json:"ip_address"`
    UserAgent    string                 `json:"user_agent"`
    ResourceType string                 `json:"resource_type"`
    ResourceID   string                 `json:"resource_id"`
    Action       string                 `json:"action"`
    Operation    string                 `json:"operation"`
    RequestID    string                 `json:"request_id"`
    TraceID      string                 `json:"trace_id"`
    BeforeState  map[string]interface{} `json:"before_state"`
    AfterState   map[string]interface{} `json:"after_state"`
    Changes      map[string]interface{} `json:"changes"`
    Result       string                 `json:"result"`
    ErrorMessage string                 `json:"error_message"`
    DurationMs   float64                `json:"duration_ms"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// AuditService 审计日志服务
type AuditService struct {
    repo AuditRepository
}

// Log 记录审计日志
func (s *AuditService) Log(ctx context.Context, log *AuditLog) error {
    log.AuditID = "audit-" + uuid.New().String()
    log.Timestamp = time.Now()
    
    return s.repo.Create(ctx, log)
}
```

---

## 📈 合规报告生成

### 审计报告类型

| 报告类型 | 频率 | 内容 |
|---------|------|------|
| **日报** | 每日 | 登录统计、异常操作、安全事件 |
| **周报** | 每周 | 用户活跃度、数据访问统计、风险分析 |
| **月报** | 每月 | 合规检查、权限审计、趋势分析 |
| **年报** | 每年 | 全面审计、合规总结、改进建议 |

---

## ✅ 安全审计验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **审计覆盖** | 全量审计 | 检查审计日志 |
| **审计存储** | 1 年保留 | 检查数据库 |
| **审计查询** | API 查询 | 检查 API |
| **合规报告** | 自动生成 | 检查报告 |
| **异常检测** | 实时告警 | 检查告警 |

---

**最后更新**: 2026-05-13  
**审核人**: Security Team