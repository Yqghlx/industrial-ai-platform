# TooManyActiveCriticalAlerts

> **告警名称**: TooManyActiveCriticalAlerts  
> **严重度**: Critical  
> **类别**: Alert  
> **阈值**: Critical 告警数 > 5  
> **持续时间**: 5 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
sum(alerts_active{severity="critical"}) > 5
```

**触发条件**:
- 系统中有超过 5 个 Critical 告警同时活动
- 可能是告警风暴，需要紧急处理

---

## 🚨 紧急响应

> **⚠️ 这是业务告警，可能表示多个设备/系统同时出现严重问题**

```bash
# 1. 查看所有 Critical 告警详情
curl http://backend:8080/api/v1/alerts?severity=critical&status=active

# 2. 查看 Grafana 告警面板
https://grafana.industrial-ai.example.com/d/industrial-ai-business

# 3. 确认是否是真实告警还是误报
```

---

## 🔍 诊断步骤

### 1️⃣ 分析告警分布

```bash
# 查看告警类型分布
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT alert_type, severity, count(*) as count
FROM alerts
WHERE status = 'active' AND severity = 'critical'
GROUP BY alert_type, severity
ORDER BY count DESC;
"

# 查看告警设备分布
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT device_id, count(*) as alert_count
FROM alerts
WHERE status = 'active' AND severity = 'critical'
GROUP BY device_id
ORDER BY alert_count DESC
LIMIT 10;
"
```

### 2️⃣ 检查告警触发原因

```bash
# 查看最近告警触发日志
docker logs industrial-ai-backend --tail 200 | grep -E "alert.*triggered.*critical" | tail 30

# 检查是否有告警规则变更
git log --since="1 day ago" --oneline -- "*alert*"
```

### 3️⃣ 确认是否是告警风暴

```bash
# 查看告警触发速率
curl 'http://prometheus:9090/api/v1/query?query=rate(alerts_triggered_total{severity="critical"}[5m])'

# 如果触发速率很高 → 可能是告警风暴
```

---

## 🛠️ 处理方案

### 情况 A: 真实的多个设备告警

```bash
# 1. 查看告警详情
curl http://backend:8080/api/v1/alerts?severity=critical&status=active | jq .

# 2. 优先处理最严重的告警
# 例如: 设备温度过高、设备故障等

# 3. 协调现场团队处理设备问题

# 4. 确认告警后标记为已处理
curl -X PUT http://backend:8080/api/v1/alerts/<alert-id>/resolve \
  -H "Authorization: Bearer <admin-token>"
```

### 情况 B: 告警风暴 (同一问题触发多个告警)

```bash
# 1. 找出根本原因
# 例如: 一个网关故障导致所有连接设备告警

# 2. 处理根本问题
# 修复网关或网络设备

# 3. 批量处理相关告警
curl -X PUT http://backend:8080/api/v1/alerts/bulk-resolve \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"alert_type":"device_offline","device_ids":["gateway-01"]}'
```

### 情景 C: 告警规则配置问题

```bash
# 1. 检查告警规则配置
cat backend/config/alert_rules.yaml

# 2. 查看是否有过于敏感的阈值
# 例如: 温度告警阈值设置过低

# 3. 调整告警阈值
# 修改告警规则配置

# 4. 重启后端加载新配置
docker-compose restart backend
```

### 情况 D: 误报 (传感器故障)

```bash
# 1. 检查设备传感器数据
curl http://backend:8080/api/v1/devices/<device-id>/telemetry | jq .

# 2. 如果传感器数据异常 → 标记为误报
curl -X PUT http://backend:8080/api/v1/alerts/<alert-id> \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"status":"false_positive","note":"传感器故障误报"}'

# 3. 维护或更换传感器
# 通知现场团队检查设备
```

---

## ✅ 验证恢复

```bash
# 1. 验证 Critical 告警数下降
curl 'http://prometheus:9090/api/v1/query?query=sum(alerts_active{severity="critical"})'

# 2. 验证所有告警已处理
curl http://backend:8080/api/v1/alerts?severity=critical&status=active | jq 'length'

# 3. 验证告警触发速率下降
curl 'http://prometheus:9090/api/v1/query?query=rate(alerts_triggered_total[5m])'
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead