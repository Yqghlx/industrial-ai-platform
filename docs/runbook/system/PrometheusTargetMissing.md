# PrometheusTargetMissing

> **告警名称**: PrometheusTargetMissing  
> **严重度**: Warning  
> **类别**: System  
> **阈值**: Prometheus UP = 0  
> **持续时间**: 10 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
up == 0
```

**触发条件**:
- Prometheus 无法抓取某个服务的 metrics
- 监控数据不完整

---

## 🔍 诊断步骤

### 1️⃣ 检查 Prometheus Targets

```bash
# 查看所有 targets 状态
curl 'http://prometheus:9090/api/v1/targets' | jq '.data.activeTargets[] | {job: .labels.job, health: .health, lastError: .lastError}'

# 或访问 Prometheus UI
https://prometheus:9090/targets
```

### 2️⃣ 检查受影响服务

```bash
# 找出 UP = 0 的服务
curl 'http://prometheus:9090/api/v1/query?query=up==0' | jq .

# 检查服务日志
docker logs <service-name> --tail 100
```

### 3️⃣ 检查网络连通性

```bash
# 从 Prometheus 测试连通性
docker exec industrial-ai-prometheus curl -I http://backend:8080/metrics

# 检查 Prometheus 配置
cat infra/prometheus.yml
```

---

## 🛠️ 修复方案

### 情况 A: 服务 /metrics 端点不可用

```bash
# 1. 检查服务是否运行
docker ps | grep <service-name>

# 2. 如果服务停止 → 启动服务
docker-compose up -d <service-name>

# 3. 检查 metrics 端点
curl http://<service>:8080/metrics
```

### 情况 B: Prometheus 配置错误

```bash
# 1. 检查 scrape_configs
cat infra/prometheus.yml | grep -A 10 "scrape_configs"

# 2. 修复配置
# 确保 targets 地址正确

# 3. 重载 Prometheus 配置
curl -X POST http://prometheus:9090/-/reload
# 或重启
docker-compose restart prometheus
```

### 情况 C: 网络问题

```bash
# 1. 检查 Docker 网络
docker network inspect industrial-ai-network

# 2. 检查服务是否在正确网络
docker inspect <service-name> | grep -A 10 Networks

# 3. 修复网络连接
docker-compose down
docker-compose up -d
```

---

## ✅ 验证恢复

```bash
# 1. 验证 UP 状态
curl 'http://prometheus:9090/api/v1/query?query=up' | jq '.data.result[] | select(.value[1]=="0")'
# 期望: 空 (无 UP=0 的服务)

# 2. 验证 Targets 页面
https://prometheus:9090/targets
# 期望: 所有 targets 状态为 UP

# 3. 验证 Metrics 抓取
curl 'http://prometheus:9090/api/v1/query?query=http_requests_total'
# 期望: 有数据
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead