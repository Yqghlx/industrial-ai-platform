# 分布式追踪集成指南

> **Industrial AI Platform 分布式追踪最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 分布式追踪概述

Phase 4 P1 运维自动化追踪目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **追踪覆盖** | 无 | 全链路追踪 |
| **追踪可视化** | 无 | Jaeger/Tempo |
| **追踪关联** | 无 | 日志+指标关联 |
| **性能分析** | 手动 | 追踪数据驱动 |

---

## 🔄 分布式追踪架构

### 追踪数据流

```
┌─────────────────────────────────────────┐
│  应用服务                                │
│  - OpenTelemetry SDK                     │
│  - 生成 Span (追踪单元)                  │
│  - 传播 TraceContext                    │
│  - 导出追踪数据                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  OpenTelemetry Collector                 │
│  - 接收追踪数据                          │
│  - 处理和转换                            │
│  - 导出到后端                            │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Jaeger / Tempo (追踪存储)               │
│  - 存储追踪数据                          │
│  - 提供查询 UI                           │
│  - 追踪分析                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Grafana (可视化)                        │
│  - 追踪面板                              │
│  - 与日志/指标关联                       │
│  - 性能分析                              │
└─────────────────────────────────────────┘
```

---

## 📝 OpenTelemetry 核心概念

### Trace 和 Span

```
Trace (追踪):
┌─────────────────────────────────────────┐
│  HTTP Request: /api/v1/devices           │
│                                          │
│  ├─ Span 1: HTTP Handler (45ms)          │
│  │   ├─ Span 2: DB Query (20ms)          │
│  │   ├─ Span 3: Redis Cache (5ms)        │
│  │   └─ Span 4: Response (15ms)          │
│                                          │
│  TraceID: abc123def456ghi789             │
│  Duration: 45ms                          │
│  Spans: 4                                │
└─────────────────────────────────────────┘
```

### Span 属性

| 属性 | 说明 |
|------|------|
| **TraceID** | 追踪唯一标识 |
| **SpanID** | Span 唯一标识 |
| **ParentSpanID** | 父 Span ID |
| **OperationName** | 操作名称 |
| **StartTime** | 开始时间 |
| **Duration** | 持续时间 |
| **Attributes** | 自定义属性 |
| **Events** | 事件记录 |
| **Status** | Span 状态 |

---

## 🔧 Go OpenTelemetry 实现

### OpenTelemetry 配置

```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Config 追踪配置
type Config struct {
    ServiceName    string // 服务名称
    ServiceVersion string // 服务版本
    Environment    string // 环境
    CollectorURL   string // Collector 地址
    SampleRate     float64 // 采样率
}

// InitTracer 初始化追踪器
func InitTracer(ctx context.Context, cfg Config) (*trace.TracerProvider, error) {
    // 创建 OTLP exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.CollectorURL),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }
    
    // 创建资源
    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(cfg.ServiceName),
        semconv.ServiceVersionKey.String(cfg.ServiceVersion),
        semconv.DeploymentEnvironmentKey.String(cfg.Environment),
    )
    
    // 创建 TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
    )
    
    // 设置全局 TracerProvider
    otel.SetTracerProvider(tp)
    
    return tp, nil
}
```

### Gin 追踪中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"
)

// TracingMiddleware 追踪中间件
func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    propagator := otel.GetTextMapPropagator()
    
    return func(c *gin.Context) {
        // 从请求头提取 TraceContext
        ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
        
        // 创建 Span
        spanName := c.Request.Method + " " + c.Request.URL.Path
        ctx, span := tracer.Start(ctx, spanName,
            trace.WithSpanKind(trace.SpanKindServer),
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.host", c.Request.Host),
                attribute.String("http.scheme", c.Request.URL.Scheme),
                attribute.String("http.user_agent", c.Request.UserAgent()),
                attribute.Int64("http.request_content_length", c.Request.ContentLength),
            ),
        )
        
        // 设置到上下文
        c.Request = c.Request.WithContext(ctx)
        
        // 处理请求
        c.Next()
        
        // 设置响应属性
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
            attribute.Int64("http.response_content_length", c.Writer.Size()),
        )
        
        // 设置 Span 状态
        if c.Writer.Status() >= 400 {
            span.SetStatus(codes.Error, c.Errors.String())
        } else {
            span.SetStatus(codes.Ok, "")
        }
        
        // 结束 Span
        span.End()
    }
}
```

---

## 📊 Jaeger 配置

### Jaeger All-in-One 配置

```yaml
# Jaeger 配置
strategy: all-in-one

collector:
  otlp:
    enabled: true
    http:
      hostPort: 0.0.0.0:4318
    grpc:
      hostPort: 0.0.0.0:4317

query:
  base-path: /
  ui-config: /etc/jaeger/ui-config.json

storage:
  type: elasticsearch
  elasticsearch:
    host: elasticsearch:9200
    index-prefix: jaeger

sampling:
  strategies:
    type: probabilistic
    param: 0.1  # 10% 采样率

reporter:
  type: grpc
  hostPort: jaeger-collector:14250
```

---

## 📊 Tempo 配置

### Tempo 配置

```yaml
# Tempo 配置

server:
  http_listen_port: 3200
  grpc_listen_port: 9095

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318

ingester:
  trace_block_size: 100000
  flush_check_period: 10s
  max_block_duration: 30m
  complete_block_timeout: 1h

compactor:
  ring:
    kvstore:
      store: inmemory
  compaction:
    block_retention: 48h
    compaction_window: 1h

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/traces
    wal:
      path: /tmp/tempo/wal
```

---

## 🔍 追踪查询与分析

### Jaeger UI 查询

```
# 搜索追踪
Service: industrial-ai-backend
Operation: GET /api/v1/devices
Tags: http.status_code=200
Min Duration: 100ms
Max Duration: 500ms

# 查看追踪详情
TraceID: abc123def456ghi789
- HTTP Handler (45ms)
  - DB Query (20ms)
  - Redis Cache (5ms)
  - Response (15ms)
```

### Grafana Tempo 查询

```
# TraceQL 查询
{.service.name = "industrial-ai-backend"}

# 条件查询
{.service.name = "industrial-ai-backend" && .http.status_code = 500}

# 延迟查询
{.service.name = "industrial-ai-backend" && duration > 100ms}
```

---

## 📈 追踪监控指标

### Prometheus 追踪指标

```
# 追踪数量
traces_total{service="industrial-ai-backend"}

# Span 数量
spans_total{service="industrial-ai-backend",operation="HTTP Handler"}

# 追踪延迟
trace_duration_seconds{service="industrial-ai-backend"}

# 错误追踪
trace_errors_total{service="industrial-ai-backend"}
```

---

## 🔔 追踪告警规则

### Alertmanager 规则

```yaml
groups:
  - name: tracing
    rules:
      - alert: HighTraceErrorRate
        expr: 
          rate(trace_errors_total[5m]) 
          / rate(traces_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High trace error rate"
          description: "Error rate is {{ $value | humanizePercentage }}"
      
      - alert: SlowTraces
        expr: 
          histogram_quantile(0.95, 
            sum(rate(trace_duration_seconds_bucket[5m])) by (le)
          ) > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow traces detected"
          description: "P95 trace duration is {{ $value }}s"
```

---

## ✅ 分布式追踪验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **追踪数据生成** | OpenTelemetry SDK | 检查追踪器 |
| **追踪数据导出** | Collector 接收 | Jaeger/Tempo UI |
| **追踪可视化** | UI 显示追踪 | Jaeger/Tempo UI |
| **追踪关联** | 日志/指标关联 | Grafana 查询 |
| **追踪采样** | 配置采样率 | 追踪数量 |
| **追踪告警** | 告警触发 | Alertmanager |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team