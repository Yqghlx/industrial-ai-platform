# Redis 性能优化指南

> **Industrial AI Platform Redis 缓存优化最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 Redis 优化概述

Phase 4 P1 Redis 性能优化目标：

| 指标 | 当前状态 | 目标值 | 提升幅度 |
|------|---------|--------|---------|
| **缓存命中率** | ~60% | >90% | 30% |
| **Redis 响应 P95** | ~50ms | <10ms | 5x |
| **内存利用率** | ~40% | <70% | 稳定 |
| **持久化延迟** | ~100ms | <50ms | 2x |
| **连接池健康** | 基础 | 优化 | 更稳定 |

---

## 🔧 Redis 配置优化

### 生产环境配置

```yaml
# redis.conf (生产环境)
# 网络配置
bind 0.0.0.0
port 6379
protected-mode yes
tcp-backlog 511
tcp-keepalive 300

# 内存配置
maxmemory 2gb                 # 最大内存限制
maxmemory-policy allkeys-lru  # 内存淘汰策略 (LRU)
maxmemory-samples 5           # LRU 样本数

# 持久化配置 (AOF + RDB 混合)
appendonly yes                # 启用 AOF
appendfsync everysec          # 每秒同步
appendfilename "appendonly.aof"
aof-use-rdb-preamble yes      # RDB 前言 (混合持久化)

# RDB 快照
save 900 1                    # 900 秒内 1 次修改
save 300 10                   # 300 秒内 10 次修改
save 60 10000                 # 60 秒内 10000 次修改
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes

# 性能优化
slowlog-max-len 128           # 慢日志最大长度
slowlog-log-slower-than 10000 # 慢日志阈值 (微秒)
latency-monitor-threshold 100 # 延迟监控阈值

# 客户端连接
maxclients 10000              # 最大客户端连接
timeout 0                     # 连接超时 (不超时)
```

### 开发环境配置

```yaml
# redis.conf (开发环境)
maxmemory 512mb
maxmemory-policy allkeys-lru

# 仅 AOF 持久化 (简化)
appendonly yes
appendfsync everysec

# 禁用 RDB 快照 (开发环境)
save ""
```

---

## 📊 内存淘汰策略

### 策略对比

| 策略 | 描述 | 适用场景 |
|------|------|---------|
| **noeviction** | 不淘汰，内存满报错 | 缓存必须完整 |
| **allkeys-lru** | 所有键 LRU 淘汰 | **推荐缓存场景** |
| **volatile-lru** | 仅带 TTL 键 LRU 淘汰 | 有 TTL 的缓存 |
| **allkeys-lfu** | 所有键 LFU 淘汰 | 热点数据识别 |
| **volatile-lfu** | 仅带 TTL 键 LFU 淘汰 | 有 TTL 热点 |
| **allkeys-random** | 随机淘汰 | 无热点场景 |
| **volatile-random** | 随机淘汰 TTL 键 | 有 TTL 场景 |
| **volatile-ttl** | 淘汰 TTL 最短键 | TTL 优先 |

### 推荐配置

```yaml
# 缓存场景 (推荐)
maxmemory-policy allkeys-lru

# 热点数据场景 (可选)
maxmemory-policy allkeys-lfu
maxmemory-samples 10
```

---

## 💾 持久化策略

### AOF + RDB 混合持久化 (推荐)

**优势：**
- AOF 数据完整性高 (最多丢失 1 秒)
- RDB 前言加速重启恢复
- 混合模式兼顾性能和安全

**配置：**

```yaml
# redis.conf
appendonly yes
appendfsync everysec          # 性能 + 安全平衡
aof-use-rdb-preamble yes      # 混合持久化

# RDB 快照 (备份)
save 900 1
save 300 10
save 60 10000
```

### 持久化性能对比

| 策略 | 数据安全 | 性能影响 | 恢复速度 |
|------|---------|---------|---------|
| **仅 RDB** | 分钟级丢失 | 低 | 快 |
| **仅 AOF (always)** | 秒级丢失 | 高 | 中 |
| **仅 AOF (everysec)** | 最多 1 秒丢失 | 中 | 中 |
| **AOF + RDB 混合** | 最多 1 秒丢失 | 中 | 快 |

---

## 🔗 Redis 集群准备

### 主从复制配置

```yaml
# 主节点
replica-serve-stale-data yes
replica-read-only yes

# 从节点配置
replicaof <master-ip> <master-port>
replica-serve-stale-data yes
replica-read-only yes
```

### 哨兵模式配置

```yaml
# sentinel.conf
sentinel monitor industrial-ai-redis <master-ip> <master-port> 2
sentinel down-after-milliseconds industrial-ai-redis 30000
sentinel parallel-syncs industrial-ai-redis 1
sentinel failover-timeout industrial-ai-redis 180000
```

### 集群模式 (Redis Cluster)

```yaml
# redis.conf (集群节点)
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
cluster-replica-validity-factor 10
cluster-migration-barrier 1
cluster-require-full-coverage yes
```

---

## 📈 Redis 性能监控

### 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|---------|
| **hit_rate** | 缓存命中率 | <80% 告警 |
| **memory_usage** | 内存使用率 | >80% 告警 |
| **connected_clients** | 连接数 | >maxclients*0.8 告警 |
| **blocked_clients** | 阻塞客户端 | >10 告警 |
| **slowlog_len** | 慢查询数 | >50 告警 |
| **latency** | 响应延迟 | >50ms 告警 |
| **instantaneous_ops_per_sec** | QPS | 监控 |

### Redis INFO 命令

```bash
# 内存信息
redis-cli INFO memory

# 统计信息
redis-cli INFO stats

# 持久化信息
redis-cli INFO persistence

# 客户端信息
redis-cli INFO clients

# 复制信息
redis-cli INFO replication

# CPU 信息
redis-cli INFO cpu
```

### Prometheus 集成

```yaml
# redis_exporter 配置
redis_exporter:
  redis.addr: "redis://redis:6379"
  redis.password: "${REDIS_PASSWORD}"
  
# Grafana Dashboard
# 导入 Redis Dashboard: https://grafana.com/grafana/dashboards/763
```

---

## 🔄 缓存策略优化

### 缓存键命名规范

```
# 推荐格式
{namespace}:{tenant_id}:{entity}:{id}:{field}

# 示例
device:tenant_001:device:dev_123:status
device:tenant_001:device:dev_123:telemetry
user:tenant_001:user:user_456:profile
cache:tenant_001:alerts:active
cache:tenant_001:devices:online

# 黑名单键 (JWT)
jwt_blacklist:token_id_abc123
```

### 缓存过期策略

| 数据类型 | TTL 建议 | 原因 |
|----------|---------|------|
| **设备状态** | 5 分钟 | 状态变化频繁 |
| **遥测数据** | 1 小时 | 近期数据热 |
| **用户 Session** | 15 分钟 | 安全刷新 |
| **告警列表** | 10 分钟 | 实时性要求 |
| **设备列表** | 30 分钟 | 列表稳定 |
| **用户信息** | 1 小时 | 信息稳定 |
| **JWT 黑名单** | 7 天 | Token 有效期 |

### 缓存预热策略

```go
// 应用启动时预热热点数据
func CacheWarmup(redis *RedisClient) {
    // 1. 预热在线设备列表
    devices := GetOnlineDevices()
    for _, device := range devices {
        redis.Set("device:"+device.ID+":status", device.Status, 5*time.Minute)
    }

    // 2. 预热活跃告警
    alerts := GetActiveAlerts()
    redis.Set("cache:active_alerts", alerts, 10*time.Minute)

    // 3. 预热用户 Session (如有)
    // ...
}
```

---

## 🛠️ Go Redis 客户端优化

### 连接池配置

```go
// backend/pkg/redis/config.go
func DefaultRedisConfig() *RedisConfig {
    return &RedisConfig{
        Addr:         "redis:6379",
        Password:     "",
        DB:           0,
        PoolSize:     50,           // 连接池大小
        MinIdleConns: 10,           // 最小空闲连接
        MaxRetries:   3,            // 最大重试次数
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
        IdleTimeout:  5 * time.Minute,
        MaxConnAge:   30 * time.Minute,
    }
}
```

### Pipeline 批量操作

```go
// 批量写入 (减少网络往返)
func BatchSet(redis *RedisClient, keys map[string]string) error {
    pipe := redis.Pipeline()
    for key, value := range keys {
        pipe.Set(key, value, 5*time.Minute)
    }
    _, err := pipe.Exec()
    return err
}
```

---

## ✅ Redis 优化验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **缓存命中率** | >90% | redis-cli INFO stats |
| **内存使用** | <70% | redis-cli INFO memory |
| **慢查询** | <10/分钟 | redis-cli SLOWLOG |
| **延迟** | P95 <10ms | redis-cli --latency |
| **连接池** | 无耗尽 | Prometheus metrics |

---

## 🔧 Redis 优化脚本

```bash
# 检查缓存命中率
redis-cli INFO stats | grep hits

# 查看慢查询日志
redis-cli SLOWLOG GET 10

# 测试延迟
redis-cli --latency-history

# 查看内存使用
redis-cli INFO memory | grep used_memory_human

# 查看持久化状态
redis-cli INFO persistence

# 触发 RDB 备份
redis-cli BGSAVE

# 触发 AOF 重写
redis-cli BGREWRITEAOF
```

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team