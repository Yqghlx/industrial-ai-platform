# 结构化日志与 Loki 集成指南

> **Industrial AI Platform 结构化日志与 Loki 最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 结构化日志概述

Phase 4 P1 运维自动化日志目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **日志格式** | 纯文本 | JSON 结构化 |
| **日志收集** | 本地文件 | Loki 集中存储 |
| **日志查询** | grep 搜索 | Grafana LogQL |
| **日志分析** | 手动分析 | 自动仪表盘 |

---

## 🔄 日志架构

### 日志收集架构

```
┌─────────────────────────────────────────┐
│  应用服务                                │
│  - 生成 JSON 结构化日志                  │
│  - 输出到 stdout/stderr                  │
│  - 包含 TraceID/请求ID                   │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Promtail (日志采集器)                   │
│  - 读取容器日志                          │
│  - 解析 JSON 日志                        │
│  - 提取标签和字段                        │
│  - 推送到 Loki                           │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Loki (日志存储)                         │
│  - 存储压缩日志                          │
│  - 多租户支持                            │
│  - 标签索引                              │
│  - LogQL 查询引擎                        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Grafana (可视化)                        │
│  - LogQL 查询界面                        │
│  - 日志仪表盘                            │
│  - 日志与指标关联                        │
│  - Alert 基于日志                        │
└─────────────────────────────────────────┘
```

---

## 📝 结构化日志格式

### JSON 日志结构

```json
{
  "timestamp": "2026-05-13T20:30:45.123Z",
  "level": "info",
  "message": "HTTP request processed",
  "logger": "http.handler",
  "service": "industrial-ai-backend",
  "version": "1.0.0",
  "environment": "production",
  "trace_id": "abc123def456",
  "span_id": "span789",
  "request_id": "req-001",
  "tenant_id": "tenant-001",
  "user_id": "user-001",
  "http": {
    "method": "GET",
    "path": "/api/v1/devices",
    "status_code": 200,
    "latency_ms": 45.6,
    "request_size_bytes": 256,
    "response_size_bytes": 1024
  },
  "source": {
    "file": "handler/device.go",
    "line": 123,
    "function": "GetDevices"
  },
  "extra": {
    "device_count": 10,
    "page": 1
  }
}
```

### 日志字段规范

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `timestamp` | string | ✅ | ISO8601 时间戳 |
| `level` | string | ✅ | 日志级别 (debug/info/warn/error) |
| `message` | string | ✅ | 日志消息 |
| `logger` | string | ✅ | 日志记录器名称 |
| `service` | string | ✅ | 服务名称 |
| `trace_id` | string | ❌ | 分布式追踪 ID |
| `request_id` | string | ❌ | 请求 ID |
| `tenant_id` | string | ❌ | 租户 ID |
| `error` | object | ❌ | 错误详情 |
| `http` | object | ❌ | HTTP 请求详情 |
| `source` | object | ❌ | 源代码位置 |

---

## 🔧 Go 结构化日志实现

### Zap 日志配置

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Config 日志配置
type Config struct {
    Level       string // 日志级别
    Format      string // 日志格式 (json/console)
    ServiceName string // 服务名称
    Environment string // 环境
    Version     string // 版本
}

// NewLogger 创建结构化日志器
func NewLogger(cfg Config) (*zap.Logger, error) {
    // 解析日志级别
    level, err := zapcore.ParseLevel(cfg.Level)
    if err != nil {
        level = zapcore.InfoLevel
    }
    
    // 配置编码器
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "timestamp",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "source",
        MessageKey:     "message",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.StringDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    
    // 创建编码器
    var encoder zapcore.Encoder
    if cfg.Format == "json" {
        encoder = zapcore.NewJSONEncoder(encoderConfig)
    } else {
        encoder = zapcore.NewConsoleEncoder(encoderConfig)
    }
    
    // 创建 Core
    core := zapcore.NewCore(
        encoder,
        zapcore.AddSync(os.Stdout),
        level,
    )
    
    // 创建 Logger
    logger := zap.New(core,
        zap.AddCaller(),
        zap.Fields(
            zap.String("service", cfg.ServiceName),
            zap.String("environment", cfg.Environment),
            zap.String("version", cfg.Version),
        ),
    )
    
    return logger, nil
}
```

### Gin 日志中间件

```go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// LoggingMiddleware 结构化日志中间件
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 记录日志
        latency := time.Since(start)
        status := c.Writer.Status()
        
        fields := []zap.Field{
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status_code", status),
            zap.Float64("latency_ms", float64(latency.Milliseconds())),
            zap.Int64("request_size_bytes", c.Request.ContentLength),
            zap.Int("response_size_bytes", c.Writer.Size()),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        }
        
        // 添加追踪 ID
        if traceID := c.GetString("trace_id"); traceID != "" {
            fields = append(fields, zap.String("trace_id", traceID))
        }
        
        // 添加请求 ID
        if requestID := c.GetString("request_id"); requestID != "" {
            fields = append(fields, zap.String("request_id", requestID))
        }
        
        // 添加租户 ID
        if tenantID := c.GetString("tenant_id"); tenantID != "" {
            fields = append(fields, zap.String("tenant_id", tenantID))
        }
        
        // 根据状态码选择日志级别
        if status >= 500 {
            logger.Error("HTTP request error", fields...)
        } else if status >= 400 {
            logger.Warn("HTTP request warning", fields...)
        } else {
            logger.Info("HTTP request processed", fields...)
        }
    }
}
```

---

## 📊 Loki 配置

### Loki 基础配置

```yaml
# loki-config.yaml

auth_enabled: false  # 单租户模式

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

ingester:
  lifecycler:
    address: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
  chunk_idle_period: 5m       # 空闲 chunk 等待时间
  chunk_block_size: 262144    # chunk 块大小
  chunk_target_size: 1572864  # 目标 chunk 大小
  chunk_retain_period: 30s    # chunk 保留时间
  max_chunk_age: 1h           # 最大 chunk 年龄

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  max_entries_per_query: 10000
  max_streams_per_user: 10000
  max_global_streams_per_user: 50000

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/boltdb-shipper-active
    cache_location: /loki/boltdb-shipper-cache
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks

compactor:
  working_directory: /loki/compactor
  shared_store: filesystem
  compaction_interval: 10m
  retention_enabled: true
  retention_delete_reference_count: 0
  retention_delete_delay: 2h

ruler:
  storage:
    type: local
    local:
      directory: /loki/rules
  rule_path: /loki/rules-temp
  alertmanager_url: http://alertmanager:9093
  ring:
    kvstore:
      store: inmemory
  enable_api: true
```

### Promtail 配置

```yaml
# promtail-config.yaml

server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push
    tenant_id: industrial-ai

scrape_configs:
  # Docker 容器日志
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["com.docker.compose.project=industrial-ai"]
    relabel_configs:
      - source_labels: [__meta_docker_container_name]
        regex: /(.*)
        target_label: container
      - source_labels: [__meta_docker_container_log_stream]
        target_label: stream
      - source_labels: [__meta_docker_container_label_com_docker_compose_service]
        target_label: service
    pipeline_config:
      stages:
        # 解析 JSON 日志
        - json:
            expressions:
              timestamp: timestamp
              level: level
              message: message
              logger: logger
              service: service
              trace_id: trace_id
              request_id: request_id
              tenant_id: tenant_id
              http_method: http.method
              http_path: http.path
              http_status: http.status_code
              http_latency: http.latency_ms
        
        # 设置时间戳
        - timestamp:
            source: timestamp
            format: RFC3339Nano
        
        # 设置标签
        - labels:
            level:
            service:
            logger:
        
        # 设置日志级别颜色
        - match:
            selector: '{level="error"}'
            stages:
              - labels:
                  level:
        
        # 提取指标
        - metrics:
            http_requests_total:
              type: Counter
              description: "Total HTTP requests"
              source: http_method
              config:
                action: inc
            http_latency_seconds:
              type: Histogram
              description: "HTTP request latency"
              source: http_latency
              config:
                buckets: [10, 50, 100, 200, 500, 1000]

  # Kubernetes Pod 日志
  - job_name: kubernetes
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - industrial-ai
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_container_name]
        target_label: container
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
      - source_labels: [__meta_kubernetes_pod_label_app]
        target_label: app
    pipeline_config:
      stages:
        - json:
            expressions:
              timestamp: timestamp
              level: level
              message: message
        - timestamp:
            source: timestamp
            format: RFC3339Nano
        - labels:
            level:
```

---

## 🔍 LogQL 查询语法

### 基础查询

```logql
# 查询所有日志
{service="industrial-ai-backend"}

# 查询错误日志
{service="industrial-ai-backend", level="error"}

# 查询特定租户日志
{service="industrial-ai-backend", tenant_id="tenant-001"}

# 查询特定容器日志
{container="backend"}
```

### 过滤查询

```logql
# 包含特定消息
{service="industrial-ai-backend"} |= "device"

# 不包含特定消息
{service="industrial-ai-backend"} != "debug"

# 正则匹配
{service="industrial-ai-backend"} |~ "error.*timeout"

# 正则不匹配
{service="industrial-ai-backend"} !~ "debug.*info"
```

### JSON 解析查询

```logql
# 解析 JSON 日志
{service="industrial-ai-backend"} | json | level="error"

# 提取字段
{service="industrial-ai-backend"} | json | http_status >= 400

# 计算 HTTP 延迟
{service="industrial-ai-backend"} | json | http_latency > 100
```

### 聚合查询

```logql
# 日志计数
count_over_time({service="industrial-ai-backend"} [1h])

# 错误日志计数
count_over_time({service="industrial-ai-backend", level="error"} [1h])

# HTTP 请求计数
count_over_time({service="industrial-ai-backend"} | json | http_method="GET" [5m])

# HTTP 延迟统计
rate({service="industrial-ai-backend"} | json | unwrap http_latency [5m])

# HTTP 状态码分布
count_over_time({service="industrial-ai-backend"} | json | http_status [1h]) by (http_status)
```

---

## 📈 Grafana 日志仪表盘

### Loki 数据源配置

```yaml
# Grafana Loki 数据源
apiVersion: 1
datasources:
  - name: Loki
    type: loki
    url: http://loki:3100
    access: proxy
    isDefault: false
    jsonData:
      maxLines: 1000
      derivedFields:
        - name: TraceID
          matcherRegex: '"trace_id":"(\w+)"'
          url: '$${__value.raw}'
          datasourceUid: tempo
```

### 日志仪表盘面板

**错误日志统计面板**
```
Query: count_over_time({service="industrial-ai-backend", level="error"} [5m])
Visualization: Stat
Title: 错误日志计数 (5分钟)
```

**HTTP 请求延迟分布面板**
```
Query: quantile_over_time(0.95, {service="industrial-ai-backend"} | json | unwrap http_latency [5m])
Visualization: Graph
Title: HTTP P95 延迟
```

**日志流面板**
```
Query: {service="industrial-ai-backend"} | json
Visualization: Logs
Title: 应用日志流
```

---

## ✅ 结构化日志验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **日志格式** | JSON 结构化 | 查看日志输出 |
| **日志级别** | 正确分类 | Loki 查询 |
| **Loki 存储** | 正常写入 | Grafana 查询 |
| **Promtail 采集** | 正常采集 | Promtail 指标 |
| **Grafana 面板** | 正常显示 | Grafana UI |
| **LogQL 查询** | 正常响应 | Grafana Explore |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team