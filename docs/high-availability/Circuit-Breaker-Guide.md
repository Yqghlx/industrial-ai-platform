# 服务降级与熔断指南

> **Industrial AI Platform 服务降级与熔断最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 熔断与降级概述

Phase 4 P1 高可用熔断与降级目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **故障传播** | 无防护 | 熔断隔离 |
| **降级响应** | 无策略 | 有序降级 |
| **故障恢复** | 手动 | 自动恢复 |
| **用户体验** | 完全失败 | 部分可用 |

---

## 🔥 熔断器模式

### 熔断器状态机

```
┌─────────────────────────────────────────┐
│  CLOSED (关闭状态)                        │  ← 正常状态
│  - 所有请求正常通过                       │
│  - 记录失败次数                           │
│  - 失败率 < 阈值时保持关闭                 │
└─────────────────────────────────────────┘
          ↓ (失败率 > 阈值)
┌─────────────────────────────────────────┐
│  OPEN (打开状态)                          │  ← 熔断状态
│  - 拒绝所有请求                           │
│  - 返回降级响应                           │
│  - 等待超时后进入半开状态                  │
└─────────────────────────────────────────┘
          ↓ (超时后)
┌─────────────────────────────────────────┐
│  HALF-OPEN (半开状态)                     │  ← 试探状态
│  - 允许少量请求通过                       │
│  - 成功 → 进入关闭状态                    │
│  - 失败 → 进入打开状态                    │
└─────────────────────────────────────────┘
```

### 熔断器配置参数

| 参数 | 说明 | 推荐值 |
|------|------|--------|
| **failureThreshold** | 失败率阈值 | 50% |
| **minRequests** | 最小请求数 | 10 |
| **openTimeout** | 打开状态超时 | 30s |
| **halfOpenRequests** | 半开状态请求数 | 3 |
| **successThreshold** | 成功阈值 | 5 |

---

## 🔧 熔断器配置

### 关键服务熔断配置

```yaml
# 熔断器配置
circuit_breakers:
  # GLM API 熔断器
  glm_api:
    failure_threshold: 50      # 失败率 >50% 熔断
    min_requests: 10           # 最少 10 次请求后判断
    open_timeout: 30s          # 熔断后 30s 后试探
    half_open_requests: 3      # 半开状态允许 3 次请求
    success_threshold: 5       # 连续 5 次成功后恢复
  
  # 数据库熔断器
  database:
    failure_threshold: 30      # 失败率 >30% 熔断
    min_requests: 5            # 最少 5 次请求后判断
    open_timeout: 60s          # 熔断后 60s 后试探
    half_open_requests: 2      # 半开状态允许 2 次请求
    success_threshold: 3       # 连续 3 次成功后恢复
  
  # Redis 熔断器
  redis:
    failure_threshold: 40      # 失败率 >40% 熔断
    min_requests: 10           # 最少 10 次请求后判断
    open_timeout: 30s          # 熔断后 30s 后试探
    half_open_requests: 3      # 半开状态允许 3 次请求
    success_threshold: 5       # 连续 5 次成功后恢复
```

---

## 📉 服务降级策略

### 降级优先级

| 优先级 | 服务 | 降级策略 |
|--------|------|---------|
| **P0** | 核心 API | 不降级 |
| **P1** | 设备管理 | 返回缓存数据 |
| **P2** | AI 功能 | 返回默认响应 |
| **P3** | 历史数据 | 返回空数据 |
| **P4** | 统计报表 | 返回延迟加载 |

### 降级响应格式

```json
// AI 功能降级响应
{
  "status": "degraded",
  "message": "AI service temporarily unavailable",
  "fallback": {
    "response": "Default response for degraded mode",
    "source": "cache"
  },
  "retry_after": 30
}

// 数据库降级响应
{
  "status": "degraded",
  "message": "Database service degraded, using cached data",
  "fallback": {
    "data": [...],
    "source": "cache",
    "timestamp": "2026-05-13T10:00:00Z"
  },
  "retry_after": 60
}
```

---

## 🛡️ 熔断器实现

### Go 熔断器库

```go
// backend/pkg/circuitbreaker/breaker.go

type CircuitBreaker struct {
    name              string
    failureThreshold  int
    minRequests       int
    openTimeout       time.Duration
    halfOpenRequests  int
    successThreshold  int
    
    state             State
    failureCount      int
    successCount      int
    requestCount      int
    lastFailureTime   time.Time
    
    mutex             sync.RWMutex
    onStateChange     func(old, new State)
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    switch cb.state {
    case StateOpen:
        return ErrCircuitBreakerOpen
    
    case StateHalfOpen:
        if cb.requestCount >= cb.halfOpenRequests {
            return ErrCircuitBreakerOpen
        }
        cb.requestCount++
        
    case StateClosed:
        cb.requestCount++
    }
    
    // 执行请求
    err := fn()
    
    // 更新状态
    if err != nil {
        cb.recordFailure()
    } else {
        cb.recordSuccess()
    }
    
    return err
}
```

---

## 📊 熔断器监控

### Prometheus 指标

```
# 熔断器状态
circuit_breaker_state{name="glm_api"} 0  # 0=closed, 1=open, 2=half_open

# 熔断器请求计数
circuit_breaker_requests{name="glm_api",state="closed"} 100
circuit_breaker_requests{name="glm_api",state="open"} 50

# 熔断器失败计数
circuit_breaker_failures{name="glm_api"} 20

# 熔断器成功计数
circuit_breaker_successes{name="glm_api"} 80

# 熔断器打开次数
circuit_breaker_opens{name="glm_api"} 2

# 熔断器恢复次数
circuit_breaker_recovers{name="glm_api"} 1
```

---

## 🔔 熔断器告警

### Alertmanager 规则

```yaml
groups:
  - name: circuit_breaker
    rules:
      - alert: CircuitBreakerOpen
        expr: circuit_breaker_state == 1
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Circuit breaker {{ $labels.name }} is open"
          description: "Service {{ $labels.name }} is circuit broken"

      - alert: CircuitBreakerHighFailures
        expr: circuit_breaker_failures / circuit_breaker_requests > 0.3
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High failure rate for {{ $labels.name }}"
          description: "Failure rate is {{ $value }}%"
```

---

## ✅ 熔断器验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **状态转换** | Closed/Open/Half-Open | 状态监控 |
| **失败检测** | 失败率计算正确 | 日志检查 |
| **自动恢复** | Half-Open 成功后恢复 | 时间监控 |
| **降级响应** | 返回降级数据 | 功能测试 |
| **监控指标** | Prometheus 指标正常 | Grafana 检查 |
| **告警触发** | 熔断告警正常 | Alertmanager 检查 |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team