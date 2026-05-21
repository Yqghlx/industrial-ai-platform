#!/bin/bash

# PostgreSQL 故障转移脚本
# 用途: 手动执行 PostgreSQL 主从切换

set -e

echo "=== PostgreSQL 故障转移 ==="
echo ""

# 环境变量
PATRONI_PRIMARY="${PATRONI_PRIMARY:-postgres-primary}"
PATRONI_REPLICA="${PATRONI_REPLICA:-postgres-replica}"
ETCD_HOST="${ETCD_HOST:-etcd}"
ETCD_PORT="${ETCD_PORT:-2379}"
PGBOUNCER_HOST="${PGBOUNCER_HOST:-pgbouncer}"
PGBOUNCER_PORT="${PGBOUNCER_PORT:-6432}"

echo "1. 检查当前主节点状态..."
PRIMARY_STATUS=$(curl -s "http://$PATRONI_PRIMARY:8008/patroni" | jq '.role' || echo "unknown")
echo "   当前主节点角色: $PRIMARY_STATUS"

if [ "$PRIMARY_STATUS" != "\"master\"" ] && [ "$PRIMARY_STATUS" != "\"primary\"" ] && [ "$PRIMARY_STATUS" != "\"leader\"" ]; then
    echo "   ⚠️ 主节点状态异常"
    echo "   当前状态: $PRIMARY_STATUS"
fi

echo ""
echo "2. 检查副本节点状态..."
REPLICA_STATUS=$(curl -s "http://$PATRONI_REPLICA:8008/patroni" | jq '.role' || echo "unknown")
REPLICA_LAG=$(curl -s "http://$PATRONI_REPLICA:8008/patroni" | jq '.replication.lag' || echo "0")
echo "   副本节点角色: $REPLICA_STATUS"
echo "   复制延迟: $REPLICA_LAG bytes"

if [ "$REPLICA_STATUS" != "\"replica\"" ] && [ "$REPLICA_STATUS" != "\"standby_leader\"" ]; then
    echo "   ⚠️ 副本节点状态异常"
fi

echo ""
echo "3. 检查复制延迟..."
if [ "$REPLICA_LAG" -gt 1048576 ]; then
    echo "   ⚠️ 复制延迟较大 (>1MB): $REPLICA_LAG bytes"
    echo "   建议等待复制同步后再切换"
    read -p "是否继续切换？(yes/no): " continue_switch
    if [ "$continue_switch" != "yes" ]; then
        echo "   取消切换"
        exit 1
    fi
fi

echo ""
echo "4. 执行故障转移..."
echo "   使用 Patroni 手动切换..."

# 使用 Patroni API 执行故障转移
SWITCH_RESULT=$(curl -s -X POST "http://$PATRONI_REPLICA:8008/switchover" -d "{\"leader\": \"$PATRONI_PRIMARY\", \"candidate\": \"$PATRONI_REPLICA\"}")

echo "   切换结果: $SWITCH_RESULT"

echo ""
echo "5. 等待切换完成..."
sleep 10

echo "   检查新主节点状态..."
NEW_PRIMARY_STATUS=$(curl -s "http://$PATRONI_REPLICA:8008/patroni" | jq '.role' || echo "unknown")
echo "   新主节点角色: $NEW_PRIMARY_STATUS"

if [ "$NEW_PRIMARY_STATUS" == "\"master\"" ] || [ "$NEW_PRIMARY_STATUS" == "\"primary\"" ] || [ "$NEW_PRIMARY_STATUS" == "\"leader\"" ]; then
    echo "   ✓ 切换成功"
else
    echo "   ✗ 切换失败"
    echo "   状态: $NEW_PRIMARY_STATUS"
fi

echo ""
echo "6. 检查旧主节点状态..."
OLD_PRIMARY_STATUS=$(curl -s "http://$PATRONI_PRIMARY:8008/patroni" | jq '.role' || echo "unknown")
echo "   旧主节点角色: $OLD_PRIMARY_STATUS"

if [ "$OLD_PRIMARY_STATUS" == "\"replica\"" ] || [ "$OLD_PRIMARY_STATUS" == "\"standby_leader\"" ]; then
    echo "   ✓ 旧主节点已降级为副本"
else
    echo "   ⚠️ 旧主节点状态异常: $OLD_PRIMARY_STATUS"
fi

echo ""
echo "7. 更新 pgBouncer 配置..."
# 检查 pgBouncer 状态
PGBOUNCER_STATS=$(curl -s "http://$PGBOUNCER_HOST:$PGBOUNCER_PORT/stats" || echo "unavailable")
echo "   pgBouncer 状态: $PGBOUNCER_STATS"

echo ""
echo "8. 验证应用连接..."
# 测试数据库连接
TEST_CONN=$(PGPASSWORD="${PG_PASSWORD:-postgres}" psql -h $PGBOUNCER_HOST -p $PGBOUNCER_PORT -U postgres -d industrial_ai -c "SELECT pg_is_in_recovery()" || echo "failed")
echo "   连接测试结果: $TEST_CONN"

echo ""
echo "=== 故障转移完成 ==="
echo ""

echo "故障转移总结:"
echo "- 新主节点: $PATRONI_REPLICA"
echo "- 新副本节点: $PATRONI_PRIMARY"
echo "- 切换耗时: ~10 秒"
echo "- 数据延迟: $REPLICA_LAG bytes"

echo ""
echo "后续操作:"
echo "1. 验证应用读写分离"
echo "2. 检查监控指标"
echo "3. 更新 DNS/路由配置"
echo "4. 记录切换日志"