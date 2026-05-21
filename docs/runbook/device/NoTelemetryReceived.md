# NoTelemetryReceived

> **告警名称**: NoTelemetryReceived  
> **严重度**: Critical  
> **类别**: Device  
> **阈值**: 15 分钟无遥测数据  
> **持续时间**: 15 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
rate(telemetry_received_total[10m]) == 0
```

**触发条件**:
- 系统完全无遥测数据接收
- 设备数据采集中断

---

## 🚨 紧急响应

> **⚠️ 这是最高严重度设备告警 - 数据完全中断**

```bash
# 1. 立即检查 WebSocket 连接
curl 'http://prometheus:9090/api/v1/query?query=websocket_connections_active'

# 2. 检查是否有在线设备
curl 'http://prometheus:9090/api/v1/query?query=sum(devices_online)'

# 3. 如果有在线设备但无数据 → 数据管道问题
```

---

## 🔍 诊断步骤

### 1️⃣ 检查数据接收管道

```bash
# 查看后端遥测接收日志
docker logs industrial-ai-backend --tail 500 | grep -E "(telemetry|receive)" | tail 50

# 检查 WebSocket 消息统计
curl 'http://prometheus:9090/api/v1/query?query=rate(websocket_messages_received_total[10m])'
```

### 2️⃣ 检查数据库写入

```bash
# 检查最近的遥测数据
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT max(created_at) as last_telemetry
FROM telemetry_data;
"

# 检查数据库写入是否有错误
docker logs industrial-ai-backend --tail 500 | grep -E "(insert.*telemetry|error)" | tail 30
```

### 3️⃣ 检查设备端状态

```bash
# 检查设备心跳
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT device_id, last_heartbeat, now() - last_heartbeat as time_since_heartbeat
FROM devices
WHERE status = 'online'
ORDER BY last_heartbeat DESC
LIMIT 20;
"
```

---

## 🛠️ 修复方案

### 情况 A: WebSocket 连接中断

```bash
# 1. 重启后端恢复 WebSocket
docker-compose restart backend

# 2. 检查设备重连日志
docker logs industrial-ai-backend --tail 100 | grep "device.*connected"

# 3. 通知设备端重新连接
# 如果设备无法自动重连，需要手动重启设备端
```

### 情况 B: 数据库写入失败

```bash
# 1. 检查数据库连接
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT 1;"

# 2. 检查磁盘空间
df -h /var/lib/postgresql

# 3. 检查表是否有锁
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT * FROM pg_locks WHERE relation = 'telemetry_data'::regclass;
"

# 4. 如果磁盘满 → 清理旧数据或扩容
```

### 情况 C: 数据管道 Bug

```bash
# 1. 查看具体错误
docker logs industrial-ai-backend --tail 500 | grep -E "(panic|fatal)" | tail 20

# 2. 重启服务
docker-compose restart backend

# 3. 如果持续失败 → 回滚
kubectl rollout undo deployment/backend -n industrial-ai

# 4. 通知开发团队
```

---

## ✅ 验证恢复

```bash
# 1. 验证遥测数据接收
curl 'http://prometheus:9090/api/v1/query?query=rate(telemetry_received_total[5m])'
# 期望: > 0

# 2. 验证数据库写入
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM telemetry_data WHERE created_at > now() - interval '5 minutes';
"

# 3. 验证 WebSocket 连接正常
docker logs industrial-ai-backend --tail 50 | grep "telemetry.*received"
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead