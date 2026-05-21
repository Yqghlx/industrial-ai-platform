# Industrial AI Platform - 生产检查清单

本文档提供部署到生产环境前的完整检查清单，确保系统安全、稳定、可靠。

## 目录

- [安全检查项](#安全检查项)
- [性能检查项](#性能检查项)
- [可用性检查项](#可用性检查项)
- [监控检查项](#监控检查项)
- [运维检查项](#运维检查项)

---

## 安全检查项

### SEC-001: 认证与授权 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| JWT_SECRET 配置 | ☐ 必须检查 | 生产环境必须设置强密钥 | `kubectl get secret -n industrial-ai` |
| JWT 密钥强度 | ☐ 必须检查 | 密钥长度 ≥ 32 字符 | `openssl rand -base64 32` |
| JWT 过期时间 | ☐ 必须检查 | 合理的过期时间（建议 24h） | 检查配置 |
| 管理员密码强度 | ☐ 必须检查 | 管理员密码 ≥ 12 字符 | 检查 Secret |
| RBAC 配置 | ☐ 必须检查 | Kubernetes RBAC 正确配置 | `kubectl get role,rolebinding -n industrial-ai` |
| ServiceAccount 使用 | ☐ 必须检查 | 使用专用 ServiceAccount | `kubectl get sa -n industrial-ai` |

### SEC-002: 数据安全 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| 数据库 SSL | ☐ 必须检查 | 连接使用 SSL/TLS | 检查 DATABASE_URL 包含 `sslmode=require` |
| 数据库密码强度 | ☐ 必须检查 | PostgreSQL 密码足够强 | 密码 ≥ 16 字符 |
| Redis 密码 | ☐ 必须检查 | Redis 使用密码认证 | 检查 redis-password |
| Secret 加密 | ☐ 必须检查 | Kubernetes Secret 正确配置 | `kubectl get secrets -n industrial-ai` |
| 密钥轮换机制 | ☐ 必须检查 | 启用密钥自动轮换 | `kubectl get cronjob secrets-rotation -n industrial-ai` |
| 备份加密 | ☐ 必须检查 | 数据库备份加密存储 | 检查备份配置 |

### SEC-003: 网络安全 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| TLS/HTTPS | ☐ 必须检查 | Ingress 使用 TLS 证书 | `kubectl get ingress -n industrial-ai` |
| TLS 证书有效期 | ☐ 必须检查 | 证书有效期 ≥ 30 天 | `openssl s_client -connect host:443` |
| CORS 配置 | ☐ 必须检查 | 明确指定 CORS 源，禁止 `*` | 检查 CORS_ORIGINS |
| Network Policy | ☐ 建议检查 | 配置 NetworkPolicy 限制流量 | `kubectl get networkpolicy -n industrial-ai` |
| WAF 配置 | ☐ 必须检查 | 启用 WAF 防护 | 检查 WAF_ENABLED=true |

### SEC-004: 容器安全 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| 非 root 用户运行 | ☐ 必须检查 | 容器以非 root 用户运行 | 检查 Dockerfile USER appuser |
| 只读文件系统 | ☐ 建议检查 | 尽可能使用只读文件系统 | 检查 Pod spec |
| 安全上下文 | ☐ 必须检查 | 配置 securityContext | `kubectl describe pod -n industrial-ai` |
| 镜像签名验证 | ☐ 建议检查 | 验证镜像签名 | 配置镜像策略 |
| 漏洞扫描 | ☐ 必须检查 | 定期扫描镜像漏洞 | `trivy image <image>` |

### SEC-005: 审计与日志 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| 审计日志启用 | ☐ 必须检查 | 启用操作审计日志 | 检查审计配置 |
| 日志聚合 | ☐ 必须检查 | 日志集中收集（Loki） | `kubectl get pods -n logging` |
| 敏感信息过滤 | ☐ 必须检查 | 日志不包含敏感信息 | 检查日志输出 |
| 访问日志记录 | ☐ 必须检查 | 记录所有 API 访问 | 检查中间件配置 |

---

## 性能检查项

### PERF-001: 资源配置 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| CPU Request | ☐ 必须检查 | 设置合理 CPU Request | 100m-500m | `kubectl describe pod -n industrial-ai` |
| CPU Limit | ☐ 必须检查 | 设置合理 CPU Limit | 500m-1000m | `kubectl describe pod -n industrial-ai` |
| Memory Request | ☐ 必须检查 | 设置合理 Memory Request | 128Mi-256Mi | `kubectl describe pod -n industrial-ai` |
| Memory Limit | ☐ 必须检查 | 设置合理 Memory Limit | 256Mi-512Mi | `kubectl describe pod -n industrial-ai` |
| Ephemeral Storage | ☐ 建议检查 | 设置临时存储限制 | 1Gi | `kubectl describe pod -n industrial-ai` |

### PERF-002: 连接池配置 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| DB 连接池大小 | ☐ 必须检查 | 合理的数据库连接池 | 25-50 | 检查 DB_MAX_OPEN_CONNS |
| DB 空闲连接数 | ☐ 必须检查 | 合理的空闲连接数 | 10-25 | 检查 DB_MAX_IDLE_CONNS |
| Redis 连接池 | ☐ 必须检查 | 合理的 Redis 连接池 | 50-100 | 检查 REDIS_POOL_SIZE |
| HTTP 客户端连接池 | ☐ 必须检查 | LLM HTTP 连接池 | 100 | 检查 LLM_MAX_IDLE_CONNS |

### PERF-003: 缓存配置 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| Redis 缓存启用 | ☐ 必须检查 | 启用 Redis 缓存 | true | 检查 CACHE_ENABLED |
| 缓存前缀设置 | ☐ 必须检查 | 设置缓存键前缀 | `iai:` | 检查 CACHE_PREFIX |
| WebSocket 压缩 | ☐ 建议检查 | 启用 WebSocket 压缩 | true | 检查 WS_COMPRESSION_ENABLED |
| 缓存预热策略 | ☐ 建议检查 | 配置缓存预热 | - | 检查缓存预热配置 |

### PERF-004: 限流与负载控制 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| 限流启用 | ☐ 必须检查 | 启用请求限流 | true | 检查 RATE_LIMIT_ENABLED |
| 请求速率限制 | ☐ 必须检查 | 合理的速率限制 | 100/s | 检查 RATE_LIMIT_REQUESTS_PER_SECOND |
| 突发流量控制 | ☐ 必须检查 | 合理的突发限制 | 200 | 检查 RATE_LIMIT_BURST |
| 连接数限制 | ☐ 建议检查 | WebSocket 连接限制 | 200 | 检查配置 |

---

## 可用性检查项

### HA-001: 高可用配置 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| 多副本部署 | ☐ 必须检查 | Deployment ≥ 2 副本 | 2+ | `kubectl get deployment -n industrial-ai` |
| Pod 反亲和性 | ☐ 必须检查 | Pod 分布在不同节点 | 配置 | `kubectl describe deployment -n industrial-ai` |
| 多节点集群 | ☐ 必须检查 | Kubernetes 多节点集群 | 3+ | `kubectl get nodes` |
| 多可用区分布 | ☐ 建议检查 | 跨可用区部署 | 2+ zones | `kubectl get nodes --show-labels` |
| 数据库高可用 | ☐ 建议检查 | PostgreSQL 主备配置 | Patroni | `kubectl get pods -n postgresql` |

### HA-002: 自动扩缩容 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| HPA 配置 | ☐ 必须检查 | 配置 HorizontalPodAutoscaler | min=2, max=10 | `kubectl get hpa -n industrial-ai` |
| CPU 扩容阈值 | ☐ 必须检查 | 合理的 CPU 扩容阈值 | 70% | `kubectl describe hpa -n industrial-ai` |
| 内存扩容阈值 | ☐ 必须检查 | 合理的内存扩容阈值 | 80% | `kubectl describe hpa -n industrial-ai` |
| 扩容稳定窗口 | ☐ 必须检查 | 合理的扩容稳定时间 | 60s | `kubectl describe hpa -n industrial-ai` |
| 缩容稳定窗口 | ☐ 必须检查 | 合理的缩容稳定时间 | 300s | `kubectl describe hpa -n industrial-ai` |

### HA-003: 健康检查 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| Liveness Probe | ☐ 必须检查 | 配置存活探针 | /health/live | `kubectl describe pod -n industrial-ai` |
| Readiness Probe | ☐ 必须检查 | 配置就绪探针 | /health/ready | `kubectl describe pod -n industrial-ai` |
| Startup Probe | ☐ 建议检查 | 配置启动探针 | /health/startup | `kubectl describe pod -n industrial-ai` |
| 探针间隔合理 | ☐ 必须检查 | 合理的探针检查间隔 | 5-10s | `kubectl describe pod -n industrial-ai` |
| 探针超时合理 | ☐ 必须检查 | 合理的探针超时时间 | 3-5s | `kubectl describe pod -n industrial-ai` |

### HA-004: 灾备与恢复 ✅

| 检查项 | 状态 | 描述 | 推荐值 | 验证命令 |
|--------|------|------|--------|----------|
| 数据库备份 | ☐ 必须检查 | 配置数据库定期备份 | 每日 | `kubectl get cronjob -n industrial-ai` |
| 备份存储位置 | ☐ 必须检查 | 备份存储在安全位置 | 外部存储 | 检查备份配置 |
| 备份恢复测试 | ☐ 必须检查 | 定期测试恢复流程 | 每月 | 执行恢复演练 |
| 跨区域备份 | ☐ 建议检查 | 跨区域备份存储 | - | 检查备份配置 |
| PVC 备份 | ☐ 建议检查 | 持久卷数据备份 | - | `velero backup get` |

---

## 监控检查项

### MON-001: 指标收集 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| Prometheus 部署 | ☐ 必须检查 | Prometheus 正常运行 | `kubectl get pods -n monitoring` |
| 指标暴露 | ☐ 必须检查 | 应用暴露 /metrics 端点 | `curl localhost:8080/metrics` |
| Pod annotations | ☐ 必须检查 | Pod 配置 Prometheus annotations | `kubectl describe pod -n industrial-ai` |
| Service Monitor | ☐ 建议检查 | 配置 ServiceMonitor | `kubectl get servicemonitor -n monitoring` |
| 指标完整性 | ☐ 必须检查 | 关键指标正确收集 | 检查 Prometheus UI |

### MON-002: 告警配置 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| Alertmanager 部署 | ☐ 必须检查 | Alertmanager 正常运行 | `kubectl get pods -n monitoring` |
| 基础告警规则 | ☐ 必须检查 | 配置基础告警规则 | `kubectl get prometheusrule -n monitoring` |
| 服务宕机告警 | ☐ 必须检查 | 配置服务宕机告警 | ServiceDown alert |
| 错误率告警 | ☐ 必须检查 | 配置 HTTP 错误率告警 | HighErrorRate alert |
| 响应时间告警 | ☐ 必须检查 | 配置响应时间告警 | HighResponseTime alert |
| 资源告警 | ☐ 必须检查 | 配置资源告警 | CPU/Memory alerts |
| 告警路由配置 | ☐ 必须检查 | 正确的告警路由 | `kubectl get secret alertmanager-config` |

### MON-003: 可视化 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| Grafana 部署 | ☐ 必须检查 | Grafana 正常运行 | `kubectl get pods -n monitoring` |
| 数据源配置 | ☐ 必须检查 | Prometheus 数据源配置 | 检查 Grafana UI |
| 应用 Dashboard | ☐ 必须检查 | 应用监控 Dashboard | 检查 Grafana Dashboard |
| 系统 Dashboard | ☐ 必须检查 | Kubernetes 系统 Dashboard | 检查 Grafana Dashboard |
| Dashboard 变量 | ☐ 建议检查 | Dashboard 使用变量 | 检查 Dashboard 配置 |

### MON-004: 日志收集 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| Loki 部署 | ☐ 建议检查 | Loki 日志收集运行 | `kubectl get pods -n logging` |
| Promtail 配置 | ☐ 建议检查 | Promtail 正常收集日志 | `kubectl get pods -n logging` |
| 日志查询正常 | ☐ 建议检查 | Grafana 可查询日志 | 检查 Grafana Explore |
| 日志保留策略 | ☐ 建议检查 | 合理的日志保留时间 | 检查 Loki 配置 |

### MON-005: 分布式追踪 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| Tempo 部署 | ☐ 建议检查 | Tempo 追踪服务运行 | `kubectl get pods -n tracing` |
| OTel Collector | ☐ 建议检查 | OpenTelemetry Collector 配置 | `kubectl get pods -n tracing` |
| 追踪数据收集 | ☐ 建议检查 | 追踪数据正常收集 | 检查 Grafana Explore |
| 追踪关联日志 | ☐ 建议检查 | 追踪与日志关联 | 检查 Trace ID 关联 |

---

## 运维检查项

### OPS-001: 部署流程 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| GitOps 配置 | ☐ 建议检查 | ArgoCD GitOps 配置 | `kubectl get application -n argocd` |
| 镜像版本管理 | ☐ 必须检查 | 使用明确版本标签 | 避免 `latest` 标签 |
| 部署回滚能力 | ☐ 必须检查 | 可快速回滚部署 | `kubectl rollout undo deployment/backend` |
| 部署变更审批 | ☐ 建议检查 | 生产变更审批流程 | 检查流程 |
| CI/CD 正常 | ☐ 必须检查 | CI/CD 流程正常 | 检查 GitHub Actions |

### OPS-002: 配置管理 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| ConfigMap 管理 | ☐ 必须检查 | ConfigMap 正确配置 | `kubectl get configmap -n industrial-ai` |
| Secret 管理 | ☐ 必须检查 | Secret 正确管理 | `kubectl get secrets -n industrial-ai` |
| 配置版本控制 | ☐ 必须检查 | 配置文件版本控制 | Git 历史检查 |
| 配置变更审计 | ☐ 建议检查 | 配置变更审计记录 | 检查审计日志 |

### OPS-003: 文档与知识库 ✅

| 检查项 | 状态 | 描述 | 验证命令 |
|--------|------|------|----------|
| 部署文档完整 | ☐ 必须检查 | 部署文档完整准确 | 检查 docs/DEPLOYMENT.md |
| Runbook 编写 | ☐ 必须检查 | 运维 Runbook 编写 | 检查 docs/runbook/ |
| API 文档完整 | ☐ 必须检查 | API 文档完整 | 检查 API 文档 |
| 架构文档更新 | ☐ 建议检查 | 架构文档及时更新 | 检查架构文档 |
| 告警处理手册 | ☐ 必须检查 | 告警处理流程文档 | 检查 runbook |

---

## 检查清单使用说明

### 1. 检查频率

- **生产上线前**: 完整检查所有项目
- **定期检查**: 每月检查安全相关项目
- **变更后**: 检查受影响的项目

### 2. 状态标记

- ✅ 通过 - 检查项满足要求
- ☐ 待检查 - 未完成检查
- ⚠️ 部分通过 - 存在部分问题
- ❌ 未通过 - 需要整改

### 3. 优先级说明

- **必须检查**: 生产环境必须满足，否则可能影响安全或可用性
- **建议检查**: 建议满足，可提升系统稳定性或运维效率

### 4. 自动化检查脚本

```bash
#!/bin/bash
# 生产检查脚本示例

echo "=== Industrial AI Platform Production Checklist ==="

# SEC-001: JWT Secret 检查
echo "检查 JWT_SECRET..."
kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath='{.data.jwt-secret}' | base64 -d | wc -c

# HA-001: 副本数检查
echo "检查 Deployment 副本数..."
kubectl get deployment backend -n industrial-ai -o jsonpath='{.spec.replicas}'

# HA-003: 健康检查配置
echo "检查健康探针..."
kubectl describe deployment backend -n industrial-ai | grep -E "LivenessProbe|ReadinessProbe"

# MON-001: Prometheus 检查
echo "检查 Prometheus..."
kubectl get pods -n monitoring -l app=prometheus

echo "=== 检查完成 ==="
```

---

## 相关文档

- [部署文档](./DEPLOYMENT.md)
- [运维 Runbook](./runbook/)
- [高可用配置](./high-availability/)
- [安全文档](./security/)