#!/bin/bash

# 缓存分析脚本
# 用途: 分析缓存使用情况和热点数据

set -e

echo "=== 缓存分析 ==="
echo ""

# 环境变量
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
CACHE_PREFIX="${CACHE_PREFIX:-cache}"

echo "1. 缓存键统计..."
TOTAL_KEYS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" DBSIZE | tr -d '\r')
echo "   总键数: $TOTAL_KEYS"

CACHE_KEYS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "$CACHE_PREFIX:*" | wc -l)
echo "   缓存键数: $CACHE_KEYS"

echo ""
echo "2. 缓存键类型分布..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "$CACHE_PREFIX:*" | \
    sed 's/.*://' | cut -d':' -f1 | sort | uniq -c | sort -rn

echo ""
echo "3. 缓存命中率..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep -E "keyspace_hits|keyspace_misses"

HITS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep keyspace_hits | cut -d: -f2 | tr -d '\r')
MISSES=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO stats | grep keyspace_misses | cut -d: -f2 | tr -d '\r')

if [ -n "$HITS" ] && [ -n "$MISSES" ]; then
    TOTAL=$((HITS + MISSES))
    if [ $TOTAL -gt 0 ]; then
        HIT_RATE=$(awk "BEGIN {printf \"%.2f\", ($HITS / $TOTAL) * 100}")
        echo "   缓存命中率: ${HIT_RATE}%"
    fi
fi

echo ""
echo "4. 内存使用..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" INFO memory | grep -E "used_memory_human|used_memory_peak_human|maxmemory_human"

echo ""
echo "5. 大键分析..."
echo "   扫描大键 (可能耗时)..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" --bigkeys | tail -10

echo ""
echo "6. 缓存键 TTL 分析..."
echo "   查看键 TTL 分布..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "$CACHE_PREFIX:*" | \
    while read key; do
        TTL=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" TTL "$key")
        echo "$key: TTL=$TTL seconds"
    done | head -20

echo ""
echo "7. 访问最频繁键 (热点数据)..."
echo "   统计 access_count:* 键..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "access_count:$CACHE_PREFIX:*" | \
    while read key; do
        COUNT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" GET "$key")
        echo "$key: count=$COUNT"
    done | sort -t= -k2 -rn | head -10

echo ""
echo "8. 缓存键过期时间检查..."
EXPIRED_KEYS=0
NEVER_EXPIRE=0

redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "$CACHE_PREFIX:*" | \
    while read key; do
        TTL=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" TTL "$key")
        if [ "$TTL" = "-1" ]; then
            NEVER_EXPIRE=$((NEVER_EXPIRE + 1))
        elif [ "$TTL" = "-2" ]; then
            EXPIRED_KEYS=$((EXPIRED_KEYS + 1))
        fi
    done

echo "   永不过期键: $NEVER_EXPIRE"
echo "   已过期键: $EXPIRED_KEYS"

echo ""
echo "9. 清理访问计数 (可选)..."
read -p "清理 access_count 键？(yes/no): " cleanup
if [ "$cleanup" == "yes" ]; then
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" KEYS "access_count:*" | \
        xargs redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" DEL
    echo "   ✓ 访问计数已清理"
fi

echo ""
echo "=== 缓存分析完成 ==="
echo ""
echo "建议:"
echo "1. 如果命中率 <80%，检查缓存策略"
echo "2. 如果有永不过期键，设置合理的 TTL"
echo "3. 如果有大键，考虑拆分或压缩"
echo "4. 根据热点数据优化缓存预热策略"