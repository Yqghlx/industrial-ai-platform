#!/bin/bash

# 缓存预热脚本
# 用途: 启动时预热常用数据到缓存

set -e

echo "=== Industrial AI Platform 缓存预热 ==="
echo ""

# ============================================
# 配置参数
# ============================================

# Redis 配置
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"

# API 配置
API_URL="${API_URL:-http://localhost:8080}"

# 预热配置
PREHEAT_DEVICES="${PREHEAT_DEVICES:-true}"
PREHEAT_CONFIG="${PREHEAT_CONFIG:-true}"
PREHEAT_STATISTICS="${PREHEAT_STATISTICS:-true}"

echo "配置参数:"
echo "- Redis 主机: $REDIS_HOST"
echo "- Redis 端口: $REDIS_PORT"
echo "- API URL: $API_URL"
echo "- 预热设备: $PREHEAT_DEVICES"
echo "- 预热配置: $PREHEAT_CONFIG"
echo "- 预热统计: $PREHEAT_STATISTICS"
echo ""

# ============================================
# 检查 Redis 连接
# ============================================

echo "1. 检查 Redis 连接..."

redis_status=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT ping || echo "down")
if [ "$redis_status" == "PONG" ]; then
    echo "   ✓ Redis 连接正常"
else
    echo "   ✗ Redis 连接失败"
    exit 1
fi

# ============================================
# 预热设备列表
# ============================================

if [ "$PREHEAT_DEVICES" == "true" ]; then
    echo ""
    echo "2. 预热设备列表..."

    # 获取设备列表
    devices=$(curl -s $API_URL/api/v1/devices | jq '.data' || echo "[]")

    # 缓存设备列表
    cache_key="cache:default:devices:list"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "$cache_key" "$devices" EX 300

    echo "   ✓ 设备列表已缓存"
fi

# ============================================
# 预热配置数据
# ============================================

if [ "$PREHEAT_CONFIG" == "true" ]; then
    echo ""
    echo "3. 预热配置数据..."

    # 获取配置数据
    config=$(curl -s $API_URL/api/v1/config | jq '.' || echo "{}")

    # 缓存配置数据
    cache_key="cache:default:config:main"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "$cache_key" "$config" EX 3600

    echo "   ✓ 配置数据已缓存"
fi

# ============================================
# 预热统计数据
# ============================================

if [ "$PREHEAT_STATISTICS" == "true" ]; then
    echo ""
    echo "4. 预热统计数据..."

    # 获取统计数据
    stats=$(curl -s $API_URL/api/v1/statistics/devices | jq '.' || echo "{}")

    # 缓存统计数据
    cache_key="cache:default:statistics:devices"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "$cache_key" "$stats" EX 3600

    echo "   ✓ 统计数据已缓存"
fi

# ============================================
# 验证缓存
# ============================================

echo ""
echo "5. 验证缓存..."

# 查询缓存 Key 数量
key_count=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT DBSIZE | awk '{print $2}')
echo "   缓存 Key 数量: $key_count"

# 查询缓存命中率
hit_rate=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT INFO stats | grep keyspace_hits || echo "0")
echo "   缓存命中率: $hit_rate"

# ============================================
# 预热完成
# ============================================

echo ""
echo "=== 缓存预热完成 ==="
echo ""

echo "预热总结:"
echo "- Redis 连接: ✓"
echo "- 设备列表: ✓"
echo "- 配置数据: ✓"
echo "- 统计数据: ✓"
echo "- 缓存 Key 数量: $key_count"

echo ""
echo "查看缓存:"
echo "- redis-cli -h $REDIS_HOST -p $REDIS_PORT KEYS 'cache:*'"
echo "- redis-cli -h $REDIS_HOST -p $REDIS_PORT INFO stats"

echo ""
echo "✅ Industrial AI Platform 缓存预热完成！"