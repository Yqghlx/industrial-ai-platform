#!/bin/bash

# PostgreSQL 复制状态检查脚本
# 用途: 检查 PostgreSQL 主从复制状态

set -e

echo "=== PostgreSQL 复制状态检查 ==="
echo ""

# 环境变量
PG_PRIMARY="${PG_PRIMARY:-postgres-primary}"
PG_REPLICA="${PG_REPLICA:-postgres-replica}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_DB="${PG_DB:-industrial_ai}"
PATRONI_PORT="${PATRONI_PORT:-8008}"

echo "1. 检查主节点状态..."
PRIMARY_PATRONI=$(curl -s "http://$PG_PRIMARY:$PATRONI_PORT/patroni" 2>/dev/null || echo "unavailable")

if [ "$PRIMARY_PATRONI" != "unavailable" ]; then
    PRIMARY_ROLE=$(echo "$PRIMARY_PATRONI" | jq '.role' || echo "unknown")
    PRIMARY_STATE=$(echo "$PRIMARY_PATRONI" | jq '.state' || echo "unknown")
    PRIMARY_LSN=$(echo "$PRIMARY_PATRONI" | jq '.xlog.position' || echo "unknown")
    
    echo "   角色: $PRIMARY_ROLE"
    echo "   状态: $PRIMARY_STATE"
    echo "   LSN: $PRIMARY_LSN"
else
    echo "   ⚠️ Patroni API 不可用"
    
    # 直接连接数据库检查
    PRIMARY_RECOVERY=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_PRIMARY -U $PG_USER -d $PG_DB -t -c "SELECT pg_is_in_recovery()")
    if [ "$PRIMARY_RECOVERY" == "f" ]; then
        echo "   数据库角色: 主节点 (primary)"
    else
        echo "   数据库角色: 副本节点 (replica)"
    fi
fi

echo ""
echo "2. 检查副本节点状态..."
REPLICA_PATRONI=$(curl -s "http://$PG_REPLICA:$PATRONI_PORT/patroni" 2>/dev/null || echo "unavailable")

if [ "$REPLICA_PATRONI" != "unavailable" ]; then
    REPLICA_ROLE=$(echo "$REPLICA_PATRONI" | jq '.role' || echo "unknown")
    REPLICA_STATE=$(echo "$REPLICA_PATRONI" | jq '.state' || echo "unknown")
    REPLICA_LSN=$(echo "$REPLICA_PATRONI" | jq '.xlog.position' || echo "unknown")
    REPLICA_LAG=$(echo "$REPLICA_PATRONI" | jq '.replication.lag' || echo "0")
    
    echo "   角色: $REPLICA_ROLE"
    echo "   状态: $REPLICA_STATE"
    echo "   LSN: $REPLICA_LSN"
    echo "   复制延迟: $REPLICA_LAG bytes"
else
    echo "   ⚠️ Patroni API 不可用"
    
    # 直接连接数据库检查
    REPLICA_RECOVERY=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_REPLICA -U $PG_USER -d $PG_DB -t -c "SELECT pg_is_in_recovery()")
    if [ "$REPLICA_RECOVERY" == "t" ]; then
        echo "   数据库角色: 副本节点 (replica)"
    else
        echo "   ⚠️ 数据库角色: 主节点 (异常)"
    fi
fi

echo ""
echo "3. 检查复制连接..."
REPLICATION_STATUS=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_PRIMARY -U $PG_USER -d $PG_DB -c "
SELECT 
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes
FROM pg_stat_replication;
" || echo "failed")

echo "$REPLICATION_STATUS"

echo ""
echo "4. 检查复制延迟详情..."
if [ "$REPLICA_LAG" != "0" ] && [ -n "$REPLICA_LAG" ]; then
    LAG_MB=$((REPLICA_LAG / 1024 / 1024))
    echo "   复制延迟: ${LAG_MB} MB"
    
    if [ $LAG_MB -gt 10 ]; then
        echo "   ⚠️ 复制延迟较大 (>10MB)"
    elif [ $LAG_MB -gt 1 ]; then
        echo "   ⚠️ 复制延迟中等 (>1MB)"
    else
        echo "   ✓ 复制延迟正常 (<1MB)"
    fi
fi

echo ""
echo "5. 检查 WAL 状态..."
WAL_STATUS=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_PRIMARY -U $PG_USER -d $PG_DB -c "
SELECT 
    pg_current_wal_lsn() AS current_lsn,
    pg_walfile_name(pg_current_wal_lsn()) AS wal_file,
    pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0')) AS wal_size;
" || echo "failed")

echo "$WAL_STATUS"

echo ""
echo "6. 检查复制槽..."
REPLICATION_SLOTS=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_PRIMARY -U $PG_USER -d $PG_DB -c "
SELECT 
    slot_name,
    slot_type,
    active,
    restart_lsn,
    pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS retained_wal
FROM pg_replication_slots;
" || echo "failed")

echo "$REPLICATION_SLOTS"

echo ""
echo "7. 检查数据库连接..."
CONN_COUNT=$(PGPASSWORD="$PG_PASSWORD" psql -h $PG_PRIMARY -U $PG_USER -d $PG_DB -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'industrial_ai'")
echo "   当前连接数: $CONN_COUNT"

echo ""
echo "=== PostgreSQL 复制状态报告 ==="
echo ""

# 生成健康报告
HEALTHY=true

# 检查主节点
if [ "$PRIMARY_ROLE" != "\"master\"" ] && [ "$PRIMARY_ROLE" != "\"primary\"" ] && [ "$PRIMARY_ROLE" != "\"leader\"" ]; then
    echo "[WARNING] 主节点角色异常: $PRIMARY_ROLE"
    HEALTHY=false
fi

# 检查副本节点
if [ "$REPLICA_ROLE" != "\"replica\"" ] && [ "$REPLICA_ROLE" != "\"standby_leader\"" ]; then
    echo "[WARNING] 副本节点角色异常: $REPLICA_ROLE"
    HEALTHY=false
fi

# 检查复制延迟
if [ -n "$REPLICA_LAG" ] && [ "$REPLICA_LAG" -gt 1048576 ]; then
    echo "[WARNING] 复制延迟较大: $REPLICA_LAG bytes"
    HEALTHY=false
fi

if $HEALTHY; then
    echo "[OK] PostgreSQL 复制状态正常"
else
    echo "[ACTION] 需要采取行动"
fi

echo ""
echo "操作命令:"
echo "- 故障转移: ./scripts/postgresql-failover.sh"
echo "- 手动切换: curl -X POST http://$PG_REPLICA:$PATRONI_PORT/switchover"
echo "- 查看状态: curl http://$PG_PRIMARY:$PATRONI_PORT/patroni"