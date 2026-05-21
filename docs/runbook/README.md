# Industrial AI Platform Runbook

> **运维告警处理标准手册**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13  
> **维护团队**: Industrial AI Ops Team

---

## 📚 目录

| 分类 | Runbook | 严重度 |
|------|---------|--------|
| **HTTP/WebSocket** | [HighErrorRate](./http/HighErrorRate.md) | Critical |
| | [HighResponseTime](./http/HighResponseTime.md) | Warning |
| | [TooManyInFlightRequests](./http/TooManyInFlightRequests.md) | Warning |
| | [WebSocketConnectionsDropped](./websocket/WebSocketConnectionsDropped.md) | Warning |
| | [TooManyWebSocketConnections](./websocket/TooManyWebSocketConnections.md) | Warning |
| **Database** | [HighDatabaseQueryTime](./database/HighDatabaseQueryTime.md) | Warning |
| | [TooManyDatabaseConnections](./database/TooManyDatabaseConnections.md) | Warning |
| | [DatabaseConnectionPoolExhausted](./database/DatabaseConnectionPoolExhausted.md) | Critical |
| **Cache** | [LowCacheHitRate](./cache/LowCacheHitRate.md) | Warning |
| | [RedisUnavailable](./cache/RedisUnavailable.md) | Critical |
| **Device** | [LowDeviceOnlineRate](./device/LowDeviceOnlineRate.md) | Warning |
| | [NoTelemetryReceived](./device/NoTelemetryReceived.md) | Critical |
| **Alert** | [TooManyActiveCriticalAlerts](./alert/TooManyActiveCriticalAlerts.md) | Critical |
| | [HighAlertTriggerRate](./alert/HighAlertTriggerRate.md) | Warning |
| **AI** | [HighAIQueryTime](./ai/HighAIQueryTime.md) | Warning |
| | [AIServiceUnavailable](./ai/AIServiceUnavailable.md) | Critical |
| | [HighTokenUsage](./ai/HighTokenUsage.md) | Warning |
| **System** | [ServiceDown](./system/ServiceDown.md) | Critical |
| | [PrometheusTargetMissing](./system/PrometheusTargetMissing.md) | Warning |

---

## 🚨 严重度分级

| 级别 | 响应时间 | 通知渠道 | 处理时效 |
|------|---------|---------|---------|
| **Critical** | 立即 (5分钟内) | PagerDuty + Slack + Email | 30分钟内解决 |
| **Warning** | 30分钟内 | Email + Slack | 4小时内解决 |
| **Info** | 1小时内 | Email | 24小时内评估 |

---

## 🔄 通用处理流程

### 1️⃣ 告警确认
```bash
# 查看 Grafana Dashboard
https://grafana.industrial-ai.example.com

# 查看 Prometheus Metrics
curl http://backend:8080/metrics

# 查看 Alertmanager
https://alertmanager.industrial-ai.example.com
```

### 2️⃣ 初步诊断
```bash
# 检查服务状态
docker ps | grep industrial-ai
kubectl get pods -n industrial-ai

# 查看日志
docker logs industrial-ai-backend --tail 200
kubectl logs -l app=backend -n industrial-ai --tail 200

# 检查资源
docker stats industrial-ai-backend
kubectl top pods -n industrial-ai
```

### 3️⃣ 问题定位
- 检查相关组件日志
- 查询 Prometheus 对应指标
- 分析时间趋势图
- 检查最近变更记录

### 4️⃣ 执行修复
- 按具体 Runbook 指导操作
- 记录操作步骤
- 更新状态到 Slack

### 5️⃣ 验证恢复
```bash
# 等待告警自动解除 (通常 5-10 分钟)
# 手动验证指标恢复正常
curl http://backend:8080/metrics | grep <metric_name>

# 验证功能正常
curl http://backend:8080/health
```

### 6️⃣ 总结报告
- 在 Jira/工单系统创建记录
- 更新 Runbook (如有新发现)
- 发送事后复盘邮件 (Critical 告警)

---

## 🔧 常用诊断命令

### 服务状态
```bash
# Docker
docker-compose ps
docker-compose logs --tail 100 backend
docker-compose restart backend

# Kubernetes
kubectl get all -n industrial-ai
kubectl describe pod <pod-name> -n industrial-ai
kubectl logs <pod-name> -n industrial-ai --previous
kubectl rollout restart deployment/backend -n industrial-ai
```

### 数据库
```bash
# PostgreSQL 连接
docker exec -it industrial-ai-postgres psql -U industrial_user -d industrial_ai

# 查看活动连接
SELECT count(*) FROM pg_stat_activity;

# 查看慢查询
SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

# 查看锁等待
SELECT * FROM pg_locks WHERE NOT granted;
```

### Redis
```bash
# Redis 连接
docker exec -it industrial-ai-redis redis-cli

# 查看内存使用
INFO memory

# 查看命中率
INFO stats | grep hits

# 清空缓存 (谨慎使用)
FLUSHALL
```

### Prometheus
```bash
# 查询指标
curl 'http://prometheus:9090/api/v1/query?query=http_requests_total'

# 查看告警规则
curl 'http://prometheus:9090/api/v1/rules'

# 查看活动告警
curl 'http://prometheus:9090/api/v1/alerts'
```

---

## 📞 紧急联系人

| 角色 | 姓名 | 电话 | Slack | 邮箱 |
|------|------|------|-------|------|
| **值班 Ops** | Ops Team | +86-xxx-xxxx | @ops-team | ops@industrial-ai.example.com |
| **DBA** | DBA Team | +86-xxx-xxxx | @dba-team | dba@industrial-ai.example.com |
| **后端开发** | Dev Team | +86-xxx-xxxx | @dev-team | dev@industrial-ai.example.com |
| **AI Team** | AI Team | +86-xxx-xxxx | @ai-team | ai@industrial-ai.example.com |
| **管理层** | Manager | +86-xxx-xxxx | @manager | manager@industrial-ai.example.com |

**值班轮换**: 每周轮换，查看值班表 https://wiki.industrial-ai.example.com/ops/rotation

---

## 📋 告警抑制规则

当 `ServiceDown` 告警触发时，以下告警会被自动抑制：
- 所有同服务的 Warning 级别告警
- 同 category 的 Warning 告警

**目的**: 避免告警风暴，聚焦核心问题

---

## 📝 Runbook 更新记录

| 日期 | 更新内容 | 作者 |
|------|---------|------|
| 2026-05-13 | 创建初始版本 | Ops Team |

---

## 🔗 相关链接

- **Grafana**: https://grafana.industrial-ai.example.com
- **Prometheus**: https://prometheus.industrial-ai.example.com
- **Alertmanager**: https://alertmanager.industrial-ai.example.com
- **Wiki**: https://wiki.industrial-ai.example.com/ops
- **值班表**: https://wiki.industrial-ai.example.com/ops/rotation
- **工单系统**: https://jira.industrial-ai.example.com
- **部署文档**: https://docs.industrial-ai.example.com/deployment