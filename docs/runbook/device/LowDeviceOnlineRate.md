# LowDeviceOnlineRate

> **告警名称**: LowDeviceOnlineRate  
> **严重度**: Warning  
> **类别**: Device  
> **阈值**: 设备在线率 < 50%  
> **持续时间**: 30 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
sum(devices_online) / sum(devices_total) < 0.5
```

**触发条件**:
- 超过一半的设备离线
- 可能是网络问题或设备异常

---

## 🔍 诊断步骤

### 1️⃣ 确认设备在线率

```bash
# 设备总数和在线数
curl 'http://prometheus:9090/api/v1/query?query=sum(devices_total)'
curl 'http://prometheus:9090/api/v1/query?query=sum(devices_online)'

# 查看各租户设备状态
curl 'http://prometheus:9090/api/v1/query?query=sum(devices_online)by(tenant_id)'
```

### 2️⃣ 查看离线设备分布

```bash
# 查看最近心跳时间
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT tenant_id, device_type, count(*) as offline_count
FROM devices
WHERE last_heartbeat < now() - interval '30 minutes'
GROUP BY tenant_id, device_type
ORDER BY offline_count DESC;
"
```

### 3️⃣ 检查 WebSocket 连接

```bash
# WebSocket 连接数
curl 'http://prometheus:9090/api/v1/query?query=websocket_connections_active'

# 查看是否有连接掉落
docker logs industrial-ai-backend --tail 200 | grep -E "(device.*disconnect|heartbeat.*timeout)"
```

### 4️⃣ 检查网络连通性

```bash
# 测试设备网关连通性
ping device-gateway.example.com

# 检查 MQTT Broker 状态 (如果使用 MQTT)
docker ps | grep mqtt
docker logs mqtt-broker --tail 100
```

---

## 🛠️ 修复方案

### 情况 A: WebSocket 连接问题

```bash
# 1. 检查 WebSocket 配置
grep -E "websocket.*timeout|heartbeat" backend/config.yaml

# 2. 调整心跳超时
# 增加设备心跳超时时间到 5 分钟

# 3. 重启后端恢复连接
docker-compose restart backend
```

### 情况 B: 网络问题

```bash
# 1. 检查网络设备状态
ssh network-admin && show device-gateway status

# 2. 检查防火墙规则
iptables -L -n | grep 8080

# 3. 检查 DNS 解析
dig device-gateway.example.com

# 4. 通知网络团队
```

### 情况 C: 设端固件问题

```bash
# 1. 检查设备日志 (如果有)
# 查看设备端是否有错误日志

# 2. 检查设备固件版本
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT firmware_version, count(*) 
FROM devices 
WHERE last_heartbeat < now() - interval '30 minutes'
GROUP BY firmware_version;
"

# 3. 如果是特定固件版本问题 → 通知设备团队推送更新
```

### 情况 D: 租户配置问题

```bash
# 1. 检查是否有租户配置变更
docker logs industrial-ai-backend --tail 500 | grep -E "tenant.*config.*change"

# 2. 检查租户设备配额
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT tenant_id, device_quota, count(*) as current_devices
FROM devices d
JOIN tenants t ON d.tenant_id = t.id
GROUP BY tenant_id, device_quota;
"

# 3. 检查是否有租户批量操作
# 查看最近是否有批量设备删除/禁用操作
```

---

## ✅ 验证恢复

```bash
# 1. 验证设备在线率回升
curl 'http://prometheus:9090/api/v1/query?query=sum(devices_online)/sum(devices_total)'

# 2. 验证新设备连接
docker logs industrial-ai-backend --tail 100 | grep "device.*connected" | wc -l

# 3. 验证遥测数据接收
curl 'http://prometheus:9090/api/v1/query?query=rate(telemetry_received_total[5m])'
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead