#!/bin/bash

# Redis 性能分析脚本
# 用途: 分析 Redis 性能瓶颈

set -e

echo "=== Redis 性能分析 ==="
echo ""

# 环境变量
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

echo "1. 连接 Redis..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" PING
echo "   ✓ Redis 连接正常"

echo ""
echo "2. 服务器信息..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO server | grep -E "redis_version|uptime_in_seconds|tcp_port"

echo ""
echo "3. 内存使用统计..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO memory | grep -E "used_memory_human|used_memory_peak_human|used_memory_rss_human|maxmemory_human|mem_fragmentation_ratio"

echo ""
echo "4. 缓存命中率..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep -E "keyspace_hits|keyspace_misses|instantaneous_ops_per_sec"

echo ""
echo "5. 计算缓存命中率..."
HITS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep keyspace_hits | cut -d: -f2 | tr -d '\r')
MISSES=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep keyspace_misses | cut -d: -f2 | tr -d '\r')

if [ -n "$HITS" ] && [ -n "$MISSES" ]; then
    TOTAL=$((HITS + MISSES))
    if [ $TOTAL -gt 0 ]; then
        HIT_RATE=$(echo "scale=2; $HITS * 100 / $TOTAL" | bc)
        echo "   缓存命中率: ${HIT_RATE}%"
        echo "   总命中: $HITS"
        echo "   总未命中: $MISSES"
    fi
fi

echo ""
echo "6. 慢查询日志..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" SLOWLOG GET 10

echo ""
echo "7. 客户端连接..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO clients | grep -E "connected_clients|blocked_clients"

echo ""
echo "8. 持久化状态..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO persistence | grep -E "rdb_last_save_time|rdb_changes_since_last_save|aof_enabled|aof_rewrite_in_progress|aof_last_rewrite_time_sec"

echo ""
echo "9. 键空间统计..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO keyspace

echo ""
echo "10. 大键分析..."
echo "   扫描大键 (可能耗时较长)..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" --bigkeys

echo ""
echo "11. 延迟测试..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" --latency-history

echo ""
echo "=== Redis 性能分析完成 ==="
echo ""
echo "建议操作:"
echo "1. 如果缓存命中率 <80%，优化缓存策略"
echo "2. 如果内存碎片率 >1.5，执行 MEMORY PURGE"
echo "3. 如果有慢查询，优化相关操作"
echo "4. 如果有大键，拆分或压缩数据"