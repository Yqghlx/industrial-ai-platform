# RedisUnavailable

> **告警名称**: RedisUnavailable  
> **严重度**: Critical  
> **类别**: Cache  
> **阈值**: 无缓存操作 5 分钟  
> **持续时间**: 5 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
redis_cache_hits_total == 0 and redis_cache_misses_total == 0
```

**触发条件**:
- Redis 完全无响应
- 缓存服务不可用
- 后端可能大量超时

---

## 🚨 紧急响应

```bash
# 1. 立即检查 Redis 状态
docker ps | grep redis
docker inspect industrial-ai-redis --format '{{.State.Status}}'

# 2. 尝试连接 Redis
docker exec industrial-ai-redis redis-cli ping

# 3. 如果无法连接，立即重启
docker-compose restart redis
```

---

## 🔍 诊断步骤

### 1️⃣ 检查 Redis 日志

```bash
# Redis 启动日志
docker logs industrial-ai-redis --tail 100

# 检查是否有 OOM 或 crash
docker logs industrial-ai-redis --tail 200 | grep -E "(OOM|crash|error|ERROR)"
```

### 2️⃣ 检查 Redis 配置

```bash
# 查看内存配置
docker exec industrial-ai-redis redis-cli CONFIG GET maxmemory
docker exec industrial-ai-redis redis-cli INFO memory

# 查看持久化配置
docker exec industrial-ai-redis redis-cli CONFIG GET save
```

### 3️⃣ 检查资源使用

```bash
# Redis 内存使用
docker stats industrial-ai-redis --no-stream

# 系统内存
free -h
```

---

## 🛠️ 修复方案

### 情况 A: Redis 进程退出

```bash
# 1. 重启 Redis
docker-compose restart redis

# 2. 等待 Redis 就绪
docker exec industrial-ai-redis redis-cli ping

# 3. 验证缓存恢复
curl 'http://prometheus:9090/api/v1/query?query=redis_cache_hits_total'
```

### 情况 B: Redis OOM

```bash
# 1. 检查内存使用
docker exec industrial-ai-redis redis-cli INFO memory | grep used_memory_human

# 2. 增加内存限制
docker exec industrial-ai-redis redis-cli CONFIG SET maxmemory 512mb

# 3. 清理部分缓存
docker exec industrial-ai-redis redis-cli --scan --pattern 'telemetry:*' | head 1000 | xargs redis-cli DEL

# 4. 如果内存无法增加，重启清理
docker-compose restart redis
```

### 情况 C: Redis 持久化失败

```bash
# 1. 检查磁盘空间
df -h /var/lib/docker

# 2. 清理磁盘空间
rm -f /var/log/*.log.old

# 3. 检查 RDB 文件
ls -lh /data/dump.rdb

# 4. 临时禁用持久化
docker exec industrial-ai-redis redis-cli CONFIG SET save ""
```

### 情况 D: 网络问题

```bash
# 1. 检查 Docker 网络
docker network ls
docker network inspect industrial-ai-network

# 2. 重建网络连接
docker-compose down
docker-compose up -d

# 3. 验证网络连通
docker exec industrial-ai-backend ping redis
```

---

## ✅ 验证恢复

```bash
# 1. Redis PING
docker exec industrial-ai-redis redis-cli ping
# 期望: PONG

# 2. 验证缓存操作
docker exec industrial-ai-redis redis-cli SET test_key test_value
docker exec industrial-ai-redis redis-cli GET test_key
# 期望: test_value

# 3. 验证 Metrics
curl 'http://prometheus:9090/api/v1/query?query=redis_cache_hits_total'

# 4. 验证后端功能
curl http://backend:8080/health
```

---

## 📝 预防措施

- 配置 Redis 内存监控告警 (提前预警)
- 定期清理过期缓存键
- 配置 Redis Sentinel 高可用
- 增加缓存降级机制 (Redis 不可用时直接查询数据库)

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead