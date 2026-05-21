# LowCacheHitRate

> **告警名称**: LowCacheHitRate  
> **严重度**: Warning  
> **类别**: Cache  
> **阈值**: 缓存命中率 < 70%  
> **持续时间**: 15 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
redis_cache_hits_total / (redis_cache_hits_total + redis_cache_misses_total) < 0.7
```

**触发条件**:
- 缓存命中率低于 70%
- 大量请求直接打到数据库

---

## 🔍 诊断步骤

### 1️⃣ 检查缓存命中率趋势

```bash
# 历史命中率
curl 'http://prometheus:9090/api/v1/query_range?query=redis_cache_hits_total/(redis_cache_hits_total+redis_cache_misses_total)*100&start=now-1h&end=now&step=60'

# Grafana 查看 Cache Hit Rate 面板
```

### 2️⃣ 分析缓存 Miss 原因

```bash
# 查看缓存键分布
docker exec industrial-ai-redis redis-cli --scan --pattern '*' | sort | uniq -c | sort -nr | head 20

# 查看缓存键 TTL 分布
docker exec industrial-ai-redis redis-cli --scan --pattern '*' | while read key; do docker exec industrial-ai-redis redis-cli TTL "$key"; done | sort | uniq -c
```

### 3️⃣ 检查缓存配置

```bash
# 缓存内存配置
docker exec industrial-ai-redis redis-cli INFO memory | grep -E "(used_memory|maxmemory)"

# 缓存策略
docker exec industrial-ai-redis redis-cli CONFIG GET maxmemory-policy
```

---

## 🛠️ 修复方案

### 情况 A: 缓存内存不足导致 eviction

```bash
# 1. 检查是否有键被 eviction
docker exec industrial-ai-redis redis-cli INFO stats | grep evicted_keys

# 2. 增加缓存内存
docker exec industrial-ai-redis redis-cli CONFIG SET maxmemory 512mb

# 3. 调整 eviction 策略
docker exec industrial-ai-redis redis-cli CONFIG SET maxmemory-policy allkeys-lru

# 4. 验证内存配置
docker exec industrial-ai-redis redis-cli INFO memory
```

### 情况 B: 缓存 TTL 过短

```bash
# 1. 检查热点数据的 TTL
docker exec industrial-ai-redis redis-cli TTL "device:online:*"

# 2. 调整 TTL
# 在配置中增加热点数据的缓存时间

# 3. 手动延长 TTL
docker exec industrial-ai-redis redis-cli --scan --pattern 'device:*' | while read key; do docker exec industrial-ai-redis redis-cli EXPIRE "$key" 3600; done
```

### 情况 C: 缓存键设计问题

```bash
# 1. 分析缓存键命名
docker exec industrial-ai-redis redis-cli --scan --pattern '*' | head 50

# 2. 检查是否有无效键
# 例如: 缓存了不存在的数据

# 3. 清理无效键
docker exec industrial-ai-redis redis-cli --scan --pattern 'invalid:*' | xargs docker exec industrial-ai-redis redis-cli DEL

# 4. 通知开发团队优化缓存策略
```

### 情况 D: 新数据未预热

```bash
# 1. 检查最近新增的数据
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM devices WHERE created_at > now() - interval '1 day';
"

# 2. 手动预热新数据缓存
# 编写脚本批量缓存新设备数据

# 3. 配置数据预热机制
# 在数据创建时自动缓存
```

---

## ✅ 验证恢复

```bash
# 1. 验证命中率回升
curl 'http://prometheus:9090/api/v1/query?query=redis_cache_hits_total/(redis_cache_hits_total+redis_cache_misses_total)*100'
# 期望: > 70%

# 2. 验证缓存内存
docker exec industrial-ai-redis redis-cli INFO memory | grep used_memory_human

# 3. 验证数据库查询减少
curl 'http://prometheus:9090/api/v1/query?query=rate(db_query_duration_seconds_count[5m])'
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead