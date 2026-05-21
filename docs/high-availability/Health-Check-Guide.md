# 健康检查增强指南

> **Industrial AI Platform 多层级健康检查最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 健康检查概述

Phase 4 P1 高可用健康检查目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **健康检查覆盖率** | 基础 | 多层级 |
| **故障检测时间** | ~30s | <10s |
| **故障恢复时间** | ~5min | <30s |
| **依赖检查** | 无 | 全面 |
| **K8s 集成** | 基础 | 增强 |

---

## 🏥 多层级健康检查

### 检查层级架构

```
┌─────────────────────────────────────────┐
│  Level 1: 基础存活检查 (Liveness)        │  ← K8s livenessProbe
│  - 应用进程存活                          │
│  - HTTP 服务响应                        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Level 2: 就绪检查 (Readiness)           │  ← K8s readinessProbe
│  - 应用初始化完成                        │
│  - 关键依赖可用                          │
│  - 数据库连接正常                        │
│  - Redis 连接正常                       │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Level 3: 详细健康检查 (Health)          │  ← /health 端点
│  - 数据库连接池状态                      │
│  - Redis 内存/命中率                     │
│  - 磁盘空间检查                          │
│  - 外部服务状态                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Level 4: 依赖深度检查 (Dependency)      │  ← /health/dependencies
│  - 数据库查询响应时间                    │
│  - Redis 操作延迟                        │
│  - GLM API 可用性                        │
│  - 网络连通性                            │
└─────────────────────────────────────────┘
```

---

## 🔍 Kubernetes 探针配置

### Liveness Probe (存活探针)

```yaml
# 存活探针: 检查应用是否存活
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30   # 容器启动后 30 秒开始检查
  periodSeconds: 10         # 每 10 秒检查一次
  timeoutSeconds: 5         # 请求超时 5 秒
  failureThreshold: 3       # 连续 3 次失败重启容器
  successThreshold: 1       # 1 次成功即认为健康
```

### Readiness Probe (就绪探针)

```yaml
# 就绪探针: 检查应用是否就绪接收流量
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10   # 容器启动后 10 秒开始检查
  periodSeconds: 5          # 每 5 秒检查一次
  timeoutSeconds: 3         # 请求超时 3 秒
  failureThreshold: 3       # 连续 3 次失败从 Service 移除
  successThreshold: 1       # 1 次成功加入 Service
```

### Startup Probe (启动探针)

```yaml
# 启动探针: 慢启动应用保护 (防止 livenessProbe 过早杀死)
startupProbe:
  httpGet:
    path: /health/startup
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 30      # 最多等待 300 秒 (30*10)
  successThreshold: 1
```

---

## 📊 健康检查端点

### 端点设计

| 端点 | 用途 | 检查内容 |
|------|------|---------|
| `/health/live` | Liveness | 仅检查进程存活 |
| `/health/ready` | Readiness | 检查关键依赖 |
| `/health` | 详细健康 | 全系统状态 |
| `/health/dependencies` | 依赖详情 | 各依赖深度检查 |

### 响应格式

```json
// /health/live (Liveness)
{
  "status": "healthy",
  "timestamp": "2026-05-13T10:00:00Z"
}

// /health/ready (Readiness)
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "redis": "ok"
  },
  "timestamp": "2026-05-13T10:00:00Z"
}

// /health (详细健康)
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 5,
      "pool_stats": {
        "open_connections": 10,
        "in_use": 2,
        "idle": 8
      }
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 2,
      "memory_used": "512MB",
      "hit_rate": 92.5
    },
    "disk": {
      "status": "healthy",
      "available_gb": 50,
      "used_percent": 40
    }
  },
  "timestamp": "2026-05-13T10:00:00Z"
}

// /health/dependencies (依赖详情)
{
  "status": "healthy",
  "dependencies": {
    "postgresql": {
      "status": "healthy",
      "latency_ms": 5,
      "version": "15.4",
      "ssl": true
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 2,
      "version": "7.0",
      "uptime_seconds": 86400
    },
    "glm_api": {
      "status": "healthy",
      "latency_ms": 150,
      "rate_limit_remaining": 100
    }
  },
  "timestamp": "2026-05-13T10:00:00Z"
}
```

---

## 🔧 Go 健康检查实现

### 健康检查 Handler

```go
// backend/internal/handler/health.go

// HealthStatus 健康状态
type HealthStatus struct {
    Status    string                 `json:"status"`
    Version   string                 `json:"version"`
    Uptime    int64                  `json:"uptime_seconds"`
    Checks    map[string]interface{} `json:"checks"`
    Timestamp string                 `json:"timestamp"`
}

// LivenessCheck 存活检查
func LivenessCheck(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}

// ReadinessCheck 就绪检查
func ReadinessCheck(c *gin.Context) {
    checks := map[string]string{}
    status := "ready"
    
    // 检查数据库
    if !isDatabaseReady() {
        checks["database"] = "not_ready"
        status = "not_ready"
    } else {
        checks["database"] = "ok"
    }
    
    // 检查 Redis
    if !isRedisReady() {
        checks["redis"] = "not_ready"
        status = "not_ready"
    } else {
        checks["redis"] = "ok"
    }
    
    if status == "ready" {
        c.JSON(200, gin.H{"status": status, "checks": checks})
    } else {
        c.JSON(503, gin.H{"status": status, "checks": checks})
    }
}
```

---

## 📈 健康检查指标

### Prometheus 指标

```
# 健康检查状态 (1=健康, 0=不健康)
health_status{component="database"} 1
health_status{component="redis"} 1
health_status{component="glm_api"} 1

# 健康检查延迟
health_check_latency_ms{component="database"} 5
health_check_latency_ms{component="redis"} 2

# 健康检查失败计数
health_check_failures{component="database"} 0
health_check_failures{component="redis"} 3
```

---

## 🔔 健康检查告警

### Alertmanager 规则

```yaml
# 健康检查告警规则
groups:
  - name: health
    rules:
      - alert: ServiceNotReady
        expr: health_status{component="application"} == 0
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "Service not ready"
          description: "Application health check failed for 30 seconds"

      - alert: DatabaseHealthDegraded
        expr: health_status{component="database"} == 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Database health degraded"
          description: "Database health check failed"

      - alert: RedisHealthDegraded
        expr: health_status{component="redis"} == 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Redis health degraded"
          description: "Redis health check failed"

      - alert: HealthCheckLatencyHigh
        expr: health_check_latency_ms > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Health check latency high"
          description: "Health check latency exceeds 100ms"
```

---

## ✅ 健康检查验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **Liveness Probe** | 响应 200 | curl /health/live |
| **Readiness Probe** | 依赖检查 | curl /health/ready |
| **详细健康检查** | 全状态返回 | curl /health |
| **依赖检查** | 延迟测试 | curl /health/dependencies |
| **K8s 集成** | 探针生效 | kubectl describe pod |
| **告警规则** | 触发正常 | Alertmanager UI |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team