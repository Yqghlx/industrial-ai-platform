# WebSocketConnectionsDropped

> **告警名称**: WebSocketConnectionsDropped  
> **严重度**: Warning  
> **类别**: WebSocket  
> **阈值**: 连接掉落率 > 10 连接/秒  
> **持续时间**: 5 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
rate(websocket_connections_total[5m]) - rate(websocket_connections_active[5m]) > 10
```

**触发条件**:
- WebSocket 连接掉落速度超过 10 连接/秒
- 可能是网络不稳定或客户端异常

---

## 🔍 诊断步骤

### 1️⃣ 检查连接状态

```bash
# 当前活动连接数
curl 'http://prometheus:9090/api/v1/query?query=websocket_connections_active'

# 连接掉落趋势
curl 'http://prometheus:9090/api/v1/query?query=rate(websocket_connections_total[5m])-rate(websocket_connections_active[5m])'
```

### 2️⃣ 检查 WebSocket 日志

```bash
# 查看连接断开原因
docker logs industrial-ai-backend --tail 500 | grep -E "(websocket|disconnect|close)" | tail 50

# 查看客户端断开码
docker logs industrial-ai-backend --tail 500 | grep -E "close.*code" | tail 20
```

### 3️⃣ 检查网络状态

```bash
# 检查负载均衡
curl -I http://load-balancer/ws

# 检查后端 WebSocket 端点
wscat -c ws://backend:8080/ws

# 检查网络延迟
ping client-network-gateway
```

### 4️⃣ 检查客户端行为

```bash
# 分析客户端 IP 分布
docker logs industrial-ai-backend --tail 200 | grep "websocket connect" | awk '{print $IP}' | sort | uniq -c | sort -nr | head 20

# 检查是否有重连风暴
docker logs industrial-ai-backend --since 10m | grep "connect" | wc -l
```

---

## 🛠️ 修复方案

### 情况 A: 网络不稳定

```bash
# 1. 检查网络设备日志
ssh network-gateway && tail /var/log/network.log

# 2. 临时调整 WebSocket 心跳间隔
# 在配置中增加心跳频率，帮助客户端保持连接

# 3. 配置客户端重连策略
# 建议客户端使用指数退避重连，避免重连风暴
```

### 情况 B: 客户端异常

```bash
# 1. 查看异常客户端 IP
docker logs industrial-ai-backend --tail 500 | grep "disconnect" | awk '{print $IP}' | sort | uniq -c | sort -nr

# 2. 临时限制异常客户端
# 在防火墙中添加临时规则
iptables -A INPUT -s <abnormal-ip> -p tcp --dport 8080 -j DROP

# 3. 联系客户端团队排查
```

### 情况 C: 后端负载过高

```bash
# 1. 检查后端资源
docker stats industrial-ai-backend --no-stream

# 2. 如果 CPU/内存过高
kubectl scale deployment/backend --replicas=5 -n industrial-ai

# 3. 调整 WebSocket 连接限制
# 修改配置中的最大连接数参数
```

### 情况 D: 代理/负载均衡问题

```bash
# 1. 检查 Nginx/负载均衡 WebSocket 配置
# 确保 proxy_read_timeout 和 proxy_send_timeout 合理

# 2. 检查 WebSocket 超时配置
grep -E "proxy_read_timeout|proxy_send_timeout" /etc/nginx/nginx.conf

# 3. 调整超时参数
# 增加 WebSocket 超时时间到 3600s
```

---

## ✅ 验证恢复

```bash
# 1. 验证连接掉落率下降
curl 'http://prometheus:9090/api/v1/query?query=rate(websocket_connections_total[5m])-rate(websocket_connections_active[5m])'

# 2. 验证活动连接稳定
curl 'http://prometheus:9090/api/v1/query?query=websocket_connections_active'

# 3. 手动 WebSocket 连接测试
wscat -c ws://backend:8080/ws -x 300
# 保持连接 5 分钟不断开
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead