#!/bin/bash
#
# 跨数据中心数据同步脚本
# 用于北京(主) - 上海(灾备) 数据中心的 PostgreSQL 和 Redis 数据同步
#
# 用法:
#   ./cross-dc-sync.sh [command] [options]
#
# Commands:
#   sync-pg          同步 PostgreSQL 数据
#   sync-redis       同步 Redis 数据
#   sync-all         同步所有数据
#   verify-data      验证数据一致性
#   status           查看同步状态
#   setup-repl       初始化复制配置
#

set -e

# =============================================================================
# 配置部分
# =============================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_section() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# 数据中心配置
DC_BJ_NAME="北京主数据中心"
DC_SH_NAME="上海灾备数据中心"

# PostgreSQL 配置
PG_BJ_HOST="${PG_BJ_HOST:-10.1.1.10}"
PG_BJ_PORT="${PG_BJ_PORT:-5432}"
PG_BJ_USER="${PG_BJ_USER:-replicator}"
PG_BJ_DB="${PG_BJ_DB:-industrial_ai}"
PG_BJ_VIP="${PG_BJ_VIP:-10.1.1.10}"

PG_SH_HOST="${PG_SH_HOST:-10.2.1.10}"
PG_SH_PORT="${PG_SH_PORT:-5432}"
PG_SH_USER="${PG_SH_USER:-replicator}"
PG_SH_DB="${PG_SH_DB:-industrial_ai}"
PG_SH_VIP="${PG_SH_VIP:-10.2.1.10}"

# Redis 配置
REDIS_BJ_HOST="${REDIS_BJ_HOST:-10.1.1.20}"
REDIS_BJ_PORT="${REDIS_BJ_PORT:-6379}"

REDIS_SH_HOST="${REDIS_SH_HOST:-10.2.1.20}"
REDIS_SH_PORT="${REDIS_SH_PORT:-6379}"

# RedisShake 配置
SHAKE_CONFIG="/etc/redis-shake/shake.toml"
SHAKE_PID_FILE="/var/run/redis-shake.pid"

# 同步锁文件
SYNC_LOCK_FILE="/tmp/cross-dc-sync.lock"

# 告警配置
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"
ALERT_EMAIL="${ALERT_EMAIL:-}"

# =============================================================================
# 工具函数
# =============================================================================

# 发送告警
send_alert() {
    local level=$1
    local message=$2
    
    log_${level} "$message"
    
    # 发送企业微信/钉钉告警
    if [ -n "$ALERT_WEBHOOK" ]; then
        curl -s -X POST "$ALERT_WEBHOOK" \
            -H 'Content-Type: application/json' \
            -d "{\"msgtype\":\"text\",\"text\":{\"content\":\"[$level] $message\"}}" > /dev/null 2>&1 || true
    fi
    
    # 发送邮件告警
    if [ -n "$ALERT_EMAIL" ]; then
        echo "$message" | mail -s "[Cross-DC-Sync] $level Alert" "$ALERT_EMAIL" || true
    fi
}

# 获取同步锁
acquire_lock() {
    if [ -f "$SYNC_LOCK_FILE" ]; then
        local pid=$(cat "$SYNC_LOCK_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_error "同步进程已在运行 (PID: $pid)"
            return 1
        else
            log_warn "发现过期的锁文件，清理中..."
            rm -f "$SYNC_LOCK_FILE"
        fi
    fi
    
    echo $$ > "$SYNC_LOCK_FILE"
    trap "rm -f $SYNC_LOCK_FILE" EXIT
    log_info "获取同步锁成功"
}

# 检查 PostgreSQL 连接
check_pg_connection() {
    local host=$1
    local port=$2
    local user=$3
    local db=$4
    
    log_info "检查 PostgreSQL 连接: $host:$port/$db"
    
    if PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "$user" -d "$db" -c "SELECT 1" > /dev/null 2>&1; then
        log_info "PostgreSQL 连接正常"
        return 0
    else
        log_error "PostgreSQL 连接失败"
        return 1
    fi
}

# 检查 Redis 连接
check_redis_connection() {
    local host=$1
    local port=$2
    
    log_info "检查 Redis 连接: $host:$port"
    
    if redis-cli -h "$host" -p "$port" -a "${REDIS_PASSWORD:-}" --no-auth-warning PING > /dev/null 2>&1; then
        log_info "Redis 连接正常"
        return 0
    else
        log_error "Redis 连接失败"
        return 1
    fi
}

# =============================================================================
# PostgreSQL 同步功能
# =============================================================================

# 查看逻辑复制状态
pg_replication_status() {
    log_section "PostgreSQL 复制状态"
    
    check_pg_connection "$PG_BJ_HOST" "$PG_BJ_PORT" "$PG_BJ_USER" "$PG_BJ_DB" || return 1
    
    log_info "查询主数据中心复制状态..."
    PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_BJ_HOST" -p "$PG_BJ_PORT" -U "$PG_BJ_USER" -d "$PG_BJ_DB" <<EOF
\echo '=== 发布状态 ==='
SELECT * FROM pg_publication;

\echo '=== 复制槽状态 ==='
SELECT slot_name, slot_type, active, restart_lsn 
FROM pg_replication_slots;

\echo '=== 复制连接状态 ==='
SELECT 
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes,
    pg_size_pretty(pg_wal_lsn_diff(sent_lsn, replay_lsn)) AS lag_size
FROM pg_stat_replication;
EOF
    
    log_info "查询灾备数据中心订阅状态..."
    check_pg_connection "$PG_SH_HOST" "$PG_SH_PORT" "$PG_SH_USER" "$PG_SH_DB" || return 1
    
    PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" <<EOF
\echo '=== 订阅状态 ==='
SELECT 
    subname,
    subenabled,
    subpublications,
    pg_wal_lsn_diff(received_lsn, latest_end_lsn) AS lag_bytes
FROM pg_subscription s
LEFT JOIN pg_stat_subscription ss ON s.subname = ss.subname;

\echo '=== 订阅详情 ==='
SELECT * FROM pg_stat_subscription;
EOF
}

# 同步 PostgreSQL 数据到灾备
sync_pg_to_dr() {
    log_section "同步 PostgreSQL 到灾备数据中心"
    
    acquire_lock || return 1
    
    # 检查连接
    check_pg_connection "$PG_BJ_HOST" "$PG_BJ_PORT" "$PG_BJ_USER" "$PG_BJ_DB" || return 1
    check_pg_connection "$PG_SH_HOST" "$PG_SH_PORT" "$PG_SH_USER" "$PG_SH_DB" || return 1
    
    log_info "检查灾备数据中心订阅状态..."
    
    # 检查订阅是否存在
    local sub_exists=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -tAc \
        "SELECT COUNT(*) FROM pg_subscription WHERE subname = 'cross_dc_subscription'")
    
    if [ "$sub_exists" -eq 0 ]; then
        log_warn "订阅不存在，创建新的订阅..."
        create_subscription
    else
        # 检查订阅是否启用
        local sub_enabled=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -tAc \
            "SELECT subenabled FROM pg_subscription WHERE subname = 'cross_dc_subscription'")
        
        if [ "$sub_enabled" = "f" ]; then
            log_info "启用订阅..."
            PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -c \
                "ALTER SUBSCRIPTION cross_dc_subscription ENABLE"
        fi
    fi
    
    # 等待同步完成
    wait_for_pg_sync
    
    log_info "PostgreSQL 同步完成"
}

# 创建订阅
create_subscription() {
    log_info "创建逻辑复制订阅..."
    
    local conn_string="host=$PG_BJ_HOST port=$PG_BJ_PORT user=$PG_BJ_USER password=${PG_PASSWORD:-} dbname=$PG_BJ_DB"
    
    PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" <<EOF
-- 创建订阅
CREATE SUBSCRIPTION cross_dc_subscription
    CONNECTION '$conn_string'
    PUBLICATION cross_dc_publication
    WITH (
        copy_data = true,
        create_slot = true,
        slot_name = 'cross_dc_slot',
        synchronous_commit = off,
        binary = true
    );
    
-- 查看订阅状态
SELECT * FROM pg_subscription WHERE subname = 'cross_dc_subscription';
EOF
    
    log_info "订阅创建完成"
}

# 等待 PostgreSQL 同步完成
wait_for_pg_sync() {
    log_info "等待数据同步完成..."
    
    local max_wait=3600  # 最大等待时间 (秒)
    local wait_interval=5
    local elapsed=0
    
    while [ $elapsed -lt $max_wait ]; do
        # 检查订阅状态
        local lag=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -tAc \
            "SELECT COALESCE(pg_wal_lsn_diff(received_lsn, latest_end_lsn), 0) FROM pg_stat_subscription WHERE subname = 'cross_dc_subscription'")
        
        if [ -z "$lag" ] || [ "$lag" = "0" ]; then
            log_info "数据同步完成"
            return 0
        fi
        
        log_info "同步中... 延迟: $lag bytes"
        sleep $wait_interval
        elapsed=$((elapsed + wait_interval))
    done
    
    log_warn "同步超时，请检查网络和服务器状态"
    return 1
}

# 初始化 PostgreSQL 复制
setup_pg_replication() {
    log_section "初始化 PostgreSQL 跨数据中心复制"
    
    acquire_lock || return 1
    
    # 检查主数据中心发布
    log_info "检查主数据中心发布..."
    check_pg_connection "$PG_BJ_HOST" "$PG_BJ_PORT" "$PG_BJ_USER" "$PG_BJ_DB" || return 1
    
    local pub_exists=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_BJ_HOST" -p "$PG_BJ_PORT" -U "$PG_BJ_USER" -d "$PG_BJ_DB" -tAc \
        "SELECT COUNT(*) FROM pg_publication WHERE pubname = 'cross_dc_publication'")
    
    if [ "$pub_exists" -eq 0 ]; then
        log_info "创建发布..."
        PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_BJ_HOST" -p "$PG_BJ_PORT" -U "$PG_BJ_USER" -d "$PG_BJ_DB" <<EOF
-- 创建发布
CREATE PUBLICATION cross_dc_publication FOR ALL TABLES;

-- 授权
GRANT SELECT ON ALL TABLES IN SCHEMA public TO replicator;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO replicator;

-- 查看发布
SELECT * FROM pg_publication;
EOF
    else
        log_info "发布已存在"
    fi
    
    # 在灾备数据中心创建订阅
    create_subscription
    
    log_info "PostgreSQL 复制初始化完成"
}

# =============================================================================
# Redis 同步功能
# =============================================================================

# 查看 Redis 同步状态
redis_sync_status() {
    log_section "Redis 同步状态"
    
    check_redis_connection "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" || return 1
    check_redis_connection "$REDIS_SH_HOST" "$REDIS_SH_PORT" || return 1
    
    log_info "主数据中心 Redis 状态..."
    redis-cli -h "$REDIS_BJ_HOST" -p "$REDIS_BJ_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning <<EOF
INFO replication
DBSIZE
EOF
    
    log_info "灾备数据中心 Redis 状态..."
    redis-cli -h "$REDIS_SH_HOST" -p "$REDIS_SH_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning <<EOF
INFO replication
DBSIZE
EOF
    
    # 检查 RedisShake 状态
    if [ -f "$SHAKE_PID_FILE" ]; then
        local pid=$(cat "$SHAKE_PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_info "RedisShake 运行中 (PID: $pid)"
        else
            log_warn "RedisShake 已停止"
        fi
    else
        log_warn "RedisShake 未运行"
    fi
}

# 启动 Redis 同步
start_redis_shake() {
    log_section "启动 Redis 同步"
    
    if [ ! -f "$SHAKE_CONFIG" ]; then
        log_error "RedisShake 配置文件不存在: $SHAKE_CONFIG"
        return 1
    fi
    
    if [ -f "$SHAKE_PID_FILE" ]; then
        local pid=$(cat "$SHAKE_PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_warn "RedisShake 已在运行 (PID: $pid)"
            return 0
        fi
    fi
    
    # 启动 RedisShake
    nohup redis-shake "$SHAKE_CONFIG" > /var/log/redis-shake/sync.log 2>&1 &
    local pid=$!
    echo $pid > "$SHAKE_PID_FILE"
    
    sleep 2
    
    if ps -p "$pid" > /dev/null 2>&1; then
        log_info "RedisShake 启动成功 (PID: $pid)"
    else
        log_error "RedisShake 启动失败"
        return 1
    fi
}

# 停止 Redis 同步
stop_redis_shake() {
    log_section "停止 Redis 同步"
    
    if [ ! -f "$SHAKE_PID_FILE" ]; then
        log_warn "RedisShake 未运行"
        return 0
    fi
    
    local pid=$(cat "$SHAKE_PID_FILE")
    
    if ps -p "$pid" > /dev/null 2>&1; then
        kill "$pid"
        sleep 2
        
        if ps -p "$pid" > /dev/null 2>&1; then
            kill -9 "$pid"
        fi
        
        rm -f "$SHAKE_PID_FILE"
        log_info "RedisShake 已停止"
    else
        log_warn "RedisShake 进程不存在"
        rm -f "$SHAKE_PID_FILE"
    fi
}

# Redis 数据校验
verify_redis_data() {
    log_section "Redis 数据校验"
    
    check_redis_connection "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" || return 1
    check_redis_connection "$REDIS_SH_HOST" "$REDIS_SH_PORT" || return 1
    
    # 比较键数量
    local bj_keys=$(redis-cli -h "$REDIS_BJ_HOST" -p "$REDIS_BJ_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning DBSIZE | grep -o '[0-9]*')
    local sh_keys=$(redis-cli -h "$REDIS_SH_HOST" -p "$REDIS_SH_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning DBSIZE | grep -o '[0-9]*')
    
    log_info "主数据中心键数量: $bj_keys"
    log_info "灾备数据中心键数量: $sh_keys"
    
    local diff=$((bj_keys - sh_keys))
    if [ $diff -lt 0 ]; then
        diff=$((-diff))
    fi
    
    if [ $diff -gt 100 ]; then
        send_alert warn "Redis 键数量差异较大: $diff"
    else
        log_info "键数量差异在可接受范围内: $diff"
    fi
    
    # 随机抽样检查
    log_info "进行随机抽样检查..."
    local sample_count=20
    local mismatch=0
    
    for i in $(seq 1 $sample_count); do
        local key=$(redis-cli -h "$REDIS_BJ_HOST" -p "$REDIS_BJ_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning RANDOMKEY 2>/dev/null)
        
        if [ -z "$key" ]; then
            continue
        fi
        
        local bj_val=$(redis-cli -h "$REDIS_BJ_HOST" -p "$REDIS_BJ_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning GET "$key" 2>/dev/null)
        local sh_val=$(redis-cli -h "$REDIS_SH_HOST" -p "$REDIS_SH_PORT" -a "${REDIS_PASSWORD:-}" --no-auth-warning GET "$key" 2>/dev/null)
        
        if [ "$bj_val" != "$sh_val" ]; then
            log_warn "数据不一致: $key"
            mismatch=$((mismatch + 1))
        fi
    done
    
    if [ $mismatch -gt 0 ]; then
        send_alert error "Redis 数据不一致: $mismatch/$sample_count"
        return 1
    else
        log_info "随机抽样检查通过"
    fi
    
    log_info "Redis 数据校验完成"
}

# =============================================================================
# PostgreSQL 数据校验
# =============================================================================

verify_pg_data() {
    log_section "PostgreSQL 数据校验"
    
    check_pg_connection "$PG_BJ_HOST" "$PG_BJ_PORT" "$PG_BJ_USER" "$PG_BJ_DB" || return 1
    check_pg_connection "$PG_SH_HOST" "$PG_SH_PORT" "$PG_SH_USER" "$PG_SH_DB" || return 1
    
    # 表列表
    local tables="users devices device_telemetry alert_rules alerts work_orders notifications"
    
    for table in $tables; do
        log_info "检查表: $table"
        
        # 行数比较
        local bj_count=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_BJ_HOST" -p "$PG_BJ_PORT" -U "$PG_BJ_USER" -d "$PG_BJ_DB" -tAc "SELECT COUNT(*) FROM $table" 2>/dev/null || echo "0")
        local sh_count=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -tAc "SELECT COUNT(*) FROM $table" 2>/dev/null || echo "0")
        
        local diff=$((bj_count - sh_count))
        if [ $diff -lt 0 ]; then
            diff=$((-diff))
        fi
        
        if [ $diff -gt 10 ]; then
            log_warn "表 $table 行数差异: 主库=$bj_count, 灾备=$sh_count, 差异=$diff"
        else
            log_info "表 $table 行数一致: $bj_count"
        fi
    done
    
    # 检查复制延迟
    log_info "检查复制延迟..."
    local lag=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$PG_SH_HOST" -p "$PG_SH_PORT" -U "$PG_SH_USER" -d "$PG_SH_DB" -tAc \
        "SELECT COALESCE(pg_wal_lsn_diff(received_lsn, latest_end_lsn), 0) FROM pg_stat_subscription LIMIT 1")
    
    if [ -n "$lag" ] && [ "$lag" != "0" ]; then
        log_warn "复制延迟: $lag bytes"
    else
        log_info "复制延迟: 0 (同步完成)"
    fi
    
    log_info "PostgreSQL 数据校验完成"
}

# =============================================================================
# 全量同步
# =============================================================================

sync_all() {
    log_section "执行全量数据同步"
    
    # PostgreSQL 同步
    sync_pg_to_dr
    
    # Redis 同步
    start_redis_shake
    
    # 等待同步稳定
    log_info "等待同步稳定..."
    sleep 30
    
    # 数据校验
    verify_data
    
    log_info "全量同步完成"
}

# 查看所有状态
show_status() {
    log_section "跨数据中心同步状态总览"
    
    echo ""
    log_info "=== PostgreSQL 状态 ==="
    pg_replication_status
    
    echo ""
    log_info "=== Redis 状态 ==="
    redis_sync_status
    
    echo ""
    log_info "=== 网络连通性 ==="
    
    # 网络延迟检查
    log_info "检查网络延迟..."
    
    local pg_latency=$(ping -c 1 "$PG_SH_HOST" 2>/dev/null | grep 'time=' | awk -F'time=' '{print $2}' | awk '{print $1}')
    local redis_latency=$(ping -c 1 "$REDIS_SH_HOST" 2>/dev/null | grep 'time=' | awk -F'time=' '{print $2}' | awk '{print $1}')
    
    log_info "PostgreSQL 网络延迟: ${pg_latency:-N/A} ms"
    log_info "Redis 网络延迟: ${redis_latency:-N/A} ms"
}

# 验证所有数据
verify_data() {
    log_section "验证数据一致性"
    
    verify_pg_data
    verify_redis_data
    
    log_info "数据验证完成"
}

# =============================================================================
# 主程序
# =============================================================================

usage() {
    echo "用法: $0 <command> [options]"
    echo ""
    echo "命令:"
    echo "  sync-pg          同步 PostgreSQL 数据到灾备数据中心"
    echo "  sync-redis       同步 Redis 数据到灾备数据中心"
    echo "  sync-all         同步所有数据到灾备数据中心"
    echo "  verify-data      验证数据一致性"
    echo "  status           查看同步状态"
    echo "  setup-repl       初始化复制配置"
    echo "  stop-redis       停止 Redis 同步"
    echo ""
    echo "环境变量:"
    echo "  PG_BJ_HOST       北京 PostgreSQL 主机 (默认: 10.1.1.10)"
    echo "  PG_BJ_PORT       北京 PostgreSQL 端口 (默认: 5432)"
    echo "  PG_SH_HOST       上海 PostgreSQL 主机 (默认: 10.2.1.10)"
    echo "  PG_SH_PORT       上海 PostgreSQL 端口 (默认: 5432)"
    echo "  PG_PASSWORD      PostgreSQL 密码"
    echo "  REDIS_BJ_HOST    北京 Redis 主机 (默认: 10.1.1.20)"
    echo "  REDIS_BJ_PORT    北京 Redis 端口 (默认: 6379)"
    echo "  REDIS_SH_HOST    上海 Redis 主机 (默认: 10.2.1.20)"
    echo "  REDIS_SH_PORT    上海 Redis 端口 (默认: 6379)"
    echo "  REDIS_PASSWORD   Redis 密码"
    echo ""
    echo "示例:"
    echo "  $0 sync-all                    # 同步所有数据"
    echo "  $0 status                      # 查看状态"
    echo "  $0 verify-data                 # 验证数据一致性"
    echo "  PG_PASSWORD=xxx $0 setup-repl  # 初始化复制"
}

case "${1:-}" in
    sync-pg)
        sync_pg_to_dr
        ;;
    sync-redis)
        start_redis_shake
        ;;
    sync-all)
        sync_all
        ;;
    verify-data)
        verify_data
        ;;
    status)
        show_status
        ;;
    setup-repl)
        setup_pg_replication
        ;;
    stop-redis)
        stop_redis_shake
        ;;
    -h|--help|help)
        usage
        ;;
    *)
        usage
        exit 1
        ;;
esac