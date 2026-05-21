# HPA 自动扩缩容指南

> **Industrial AI Platform Kubernetes HPA 最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 HPA 概述

Phase 4 P1 运维自动化 HPA 目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **扩缩容响应** | 手动 | 自动 |
| **扩容触发时间** | ~5min | <30s |
| **缩容稳定窗口** | 无 | 5min |
| **自定义指标支持** | 无 | Prometheus 指标 |

---

## 🔄 HPA 工作原理

### HPA 扩缩容流程

```
┌─────────────────────────────────────────┐
│  1. Metrics Server 收集指标              │
│  - CPU 使用率                            │
│  - 内存使用率                            │
│  - 自定义 Prometheus 指标                │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. HPA Controller 计算目标副本数         │
│  - currentMetrics / desiredMetrics       │
│  - currentReplicas * utilization        │
│  - 向上取整                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 检查扩缩容限制                         │
│  - minReplicas <= target <= maxReplicas  │
│  - 缩容稳定窗口检查                       │
│  - 扩容冷却时间检查                       │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 更新 Deployment 副本数                │
│  - 扩容: 创建新 Pod                       │
│  - 缩容: 删除多余 Pod                     │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 新 Pod 就绪检查                       │
│  - Readiness Probe 通过                  │
│  - 加入 Service Endpoints                │
│  - 开始接收流量                          │
└─────────────────────────────────────────┘
```

### HPA 计算公式

```
目标副本数 = ceil(当前副本数 * (当前指标值 / 目标指标值))

示例:
- 当前副本数: 2
- 当前 CPU 使用率: 70%
- 目标 CPU 使用率: 50%
- 目标副本数 = ceil(2 * 70/50) = ceil(2.8) = 3
```

---

## 🔧 HPA 配置参数

### 基础 HPA 配置

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2              # 最小副本数
  maxReplicas: 10             # 最大副本数
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70  # CPU 使用率 70% 触发扩容
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80  # 内存使用率 80% 触发扩容
```

### HPA Behavior 配置

```yaml
behavior:
  # 缩容行为
  scaleDown:
    stabilizationWindowSeconds: 300  # 缩容稳定窗口 5 分钟
    policies:
      - type: Percent
        value: 10                     # 每分钟最多缩容 10%
        periodSeconds: 60
      - type: Pods
        value: 1                      # 每分钟最多缩容 1 个 Pod
        periodSeconds: 60
    selectPolicy: Min                 # 选择最保守的策略
  
  # 扩容行为
  scaleUp:
    stabilizationWindowSeconds: 60   # 扩容稳定窗口 1 分钟
    policies:
      - type: Percent
        value: 100                    # 每分钟最多扩容 100%
        periodSeconds: 15
      - type: Pods
        value: 4                      # 每 15 秒最多扩容 4 个 Pod
        periodSeconds: 15
    selectPolicy: Max                 # 选择最激进扩容策略
```

---

## 📊 HPA 指标类型

### Resource 指标 (资源)

| 指标 | 说明 | 推荐阈值 |
|------|------|---------|
| **cpu** | CPU 使用率 | 70% |
| **memory** | 内存使用率 | 80% |

### Pods 指标 (Pod 级别)

| 指标 | 说明 | 推荐阈值 |
|------|------|---------|
| **requests_per_second** | 每秒请求数 | 1000 |
| **connections_active** | 活跃连接数 | 500 |

### Object 指标 (对象级别)

| 指标 | 说明 | 推荐阈值 |
|------|------|---------|
| **queue_length** | 队列长度 | 100 |
| **http_latency_p95** | HTTP 延迟 P95 | 100ms |

---

## 🎯 自定义指标 HPA

### Prometheus Adapter 配置

```yaml
# prometheus-adapter 配置
rules:
  custom:
    - name: http_requests_per_second
      seriesQuery: http_requests_total{job="backend"}
      resources:
        overrides:
          namespace: {resource: "namespace"}
          pod: {resource: "pod"}
      metricsQuery: sum(rate(http_requests_total{<<.LabelMatchers>>}[1m])) by (<<.GroupBy>>)
    
    - name: http_latency_p95_ms
      seriesQuery: http_request_duration_seconds_bucket{job="backend"}
      resources:
        overrides:
          namespace: {resource: "namespace"}
          pod: {resource: "pod"}
      metricsQuery: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{<<.LabelMatchers>>}[1m])) by (le, <<.GroupBy>>))
```

### 自定义指标 HPA 配置

```yaml
metrics:
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: 1000  # 每秒 1000 请求触发扩容
  
  - type: Pods
    pods:
      metric:
        name: http_latency_p95_ms
      target:
        type: AverageValue
        averageValue: 100  # P95 延迟 100ms 触发扩容
```

---

## 📈 HPA 监控指标

### Prometheus 指标

```
# HPA 当前副本数
kube_hpa_status_current_replicas{hpa="backend-hpa"}

# HPA 目标副本数
kube_hpa_status_desired_replicas{hpa="backend-hpa"}

# HPA 最小副本数
kube_hpa_spec_min_replicas{hpa="backend-hpa"}

# HPA 最大副本数
kube_hpa_spec_max_replicas{hpa="backend-hpa"}

# HPA 扩缩容条件
kube_hpa_status_condition{hpa="backend-hpa",condition="ScalingActive"}

# HPA 当前指标值
kube_hpa_status_current_metrics_average_value{hpa="backend-hpa"}
```

---

## 🔔 HPA 告警规则

### Alertmanager 规则

```yaml
groups:
  - name: hpa
    rules:
      - alert: HPAAtMaxReplicas
        expr: kube_hpa_status_current_replicas >= kube_hpa_spec_max_replicas
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "HPA {{ $labels.hpa }} at max replicas"
          description: "HPA is at max replicas {{ $value }}. Consider increasing maxReplicas."
      
      - alert: HPAUnableToScale
        expr: kube_hpa_status_condition{condition="ScalingLimited",status="true"} == 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "HPA {{ $labels.hpa }} unable to scale"
          description: "HPA scaling is limited. Check resource limits or quota."
      
      - alert: HPAScalingFrequent
        expr: rate(kube_hpa_status_current_replicas[5m]) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "HPA {{ $labels.hpa }} scaling frequently"
          description: "HPA is scaling frequently. May indicate unstable workload."
```

---

## 🛠️ HPA 最佳实践

### 推荐配置策略

| 场景 | minReplicas | maxReplicas | CPU 阈值 |
|------|-------------|-------------|---------|
| **Web API** | 2 | 10 | 70% |
| **Worker** | 1 | 5 | 60% |
| **Batch** | 0 | 20 | 80% |
| **AI Service** | 1 | 3 | 50% |

### HPA 注意事项

1. **资源请求设置正确** - HPA 基于 requests 计算，确保 requests 设置合理
2. **避免抖动** - 设置合理的稳定窗口和缩容策略
3. **监控触发原因** - 多指标触发时，关注主导指标
4. **考虑 Pod 启动时间** - 预留启动时间，避免过早扩容
5. **结合 VPA** - VPA 调整 requests，HPA 调整 replicas

---

## ✅ HPA 验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **扩容触发** | CPU >70% 自动扩容 | 压测触发 |
| **缩容触发** | CPU <70% 自动缩容 | 减负载观察 |
| **稳定窗口** | 缩容等待 5min | 检查 behavior |
| **最大副本限制** | 不超过 maxReplicas | 告警检查 |
| **自定义指标** | Prometheus 指标生效 | kubectl hpa |
| **告警规则** | 触发正常 | Alertmanager |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team