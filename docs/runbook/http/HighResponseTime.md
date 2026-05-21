# HighResponseTime

> **告警名称**: HighResponseTime  
> **严重度**: Warning  
> **类别**: HTTP  
> **阈值**: P95 响应时间 > 1 秒  
> **持续时间**: 10 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 1
```

**触发条件**:
- 过去 10 分钟内，95% 的请求响应时间超过 1 秒
- 用户可能感到明显的响应延迟

---

## 🔍 诊断步骤

### 1️⃣ 确认响应时间分布

```bash
# 查看各百分位响应时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.50,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))'
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))'
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.99,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))'

# Grafana 查看 Response Time 面板
```

### 2️⃣ 按路径分析慢请求

```bash
# 查看各路径的响应时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,path))'

# 找出最慢的路径
curl 'http://prometheus:9090/api/v1/query?query=topk(5,histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,path)))'
```

### 3️⃣ 检查数据库查询时间

```bash
# 查看数据库查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le))'

# 查看 PostgreSQL 慢查询
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT query, calls, total_time/calls as avg_time 
FROM pg_stat_statements 
ORDER BY avg_time DESC 
LIMIT 10;
"
```

### 4️⃣ 检查缓存命中率

```bash
# 缓存命中率
curl 'http://prometheus:9090/api/v1/query?query=redis_cache_hits_total/(redis_cache_hits_total+redis_cache_misses_total)'

# Redis INFO
docker exec industrial-ai-redis redis-cli INFO stats | grep hits
```

### 5️⃣ 检查并发请求量

```bash
# 并发请求数
curl 'http://prometheus:9090/api/v1/query?query=http_requests_in_flight'

# 资源使用
docker stats industrial-ai-backend --no-stream
kubectl top pods -n industrial-ai
```

---

## 🛠️ 修复方案

### 情况 A: 数据库查询慢

```bash
# 1. 查看慢查询详情
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT query, calls, rows, total_time, mean_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC
LIMIT 20;
"

# 2. 检查是否有缺失索引
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT schemaname, relname, attname, n_distinct, correlation
FROM pg_stats
WHERE n_distinct > 100 AND correlation < 0.1;
"

# 3. 为高频查询添加索引 (根据具体查询)
# 示例:
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
CREATE INDEX CONCURRENTLY idx_devices_tenant_status 
ON devices (tenant_id, status);
"

# 4. 如果查询无法优化 → 增加缓存
# 调整缓存 TTL 或缓存策略
```

### 情况 B: 缓存命中率低

```bash
# 1. 检查缓存配置
docker exec industrial-ai-redis redis-cli CONFIG GET maxmemory
docker exec industrial-ai-redis redis-cli CONFIG GET maxmemory-policy

# 2. 增加缓存内存
docker exec industrial-ai-redis redis-cli CONFIG SET maxmemory 512mb

# 3. 检查缓存键分布
docker exec industrial-ai-redis redis-cli --scan --pattern '*' | head -100

# 4. 手动预热热点数据 (如果必要)
# 根据业务逻辑预热常用设备/告警数据
```

### 情况 C: 并发请求过多

```bash
# 1. 检查请求来源
docker logs industrial-ai-backend --tail 200 | grep "User-Agent" | sort | uniq -c

# 2. 如果是正常流量增长 → 扩容
kubectl scale deployment/backend --replicas=5 -n industrial-ai

# 3. 如果是异常流量 → 配置限流
# 修改 middleware 限流参数
```

### 情况 D: AI 查询慢

```bash
# 1. 检查 AI 查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(ai_query_duration_seconds_bucket[5m]))by(le))'

# 2. 检查是否是特定模型慢
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(ai_query_duration_seconds_bucket[5m]))by(le,model))'

# 3. 切换到更快的模型 (临时)
# 或配置请求超时

# 4. 检查 AI API 响应时间
curl -w "Time: %{time_total}s\n" -X POST https://open.bigmodel.cn/api/paas/v4/chat/completions \
  -H "Authorization: Bearer $GLM_API_KEY" \
  -d '{"model":"glm-4","prompt":"test"}'
```

---

## ✅ 验证恢复

```bash
# 1. 验证响应时间下降
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))'

# 2. 验证数据库查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le))'

# 3. 验证缓存命中率
curl 'http://prometheus:9090/api/v1/query?query=redis_cache_hits_total/(redis_cache_hits_total+redis_cache_misses_total)*100'
```

---

## 📝 事后复盘

### 可能原因
- 数据库索引缺失
- 查询返回大量数据
- 缓存未命中
- 外部 API 响应慢
- 资源不足

### 长期优化建议
- 定期分析慢查询日志
- 配置自动索引推荐
- 增加热点数据缓存
- 配置 HPA 自动扩缩容
- 优化 AI 查询超时和重试

---

## 🔗 相关 Runbook

- [HighDatabaseQueryTime](../database/HighDatabaseQueryTime.md) - 数据库查询慢
- [LowCacheHitRate](../cache/LowCacheHitRate.md) - 缓存命中率低
- [HighAIQueryTime](../ai/HighAIQueryTime.md) - AI 查询慢

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead