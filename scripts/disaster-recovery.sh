#!/bin/bash
#
# 灾备切换脚本
# 用于北京(主) - 上海(灾备) 数据中心的灾备切换操作
#
# 用法:
#   ./disaster-recovery.sh <command> [options]
#
# Commands:
#   switch-to-dr      切换到灾备数据中心
#   switch-back       回切到主数据中心
#   verify            验证数据中心状态
#   health-check      健康检查
#   emergency-switch  紧急切换 (跳过部分验证)
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# 数据中心配置
DC_BJ_NAME="北京主数据中心"
DC_SH_NAME="上海灾备数据中心"

# PostgreSQL 配置
PG_BJ_HOST="${PG_BJ_HOST:-10.1.1.10}"
PG_BJ_PORT="${PG_BJ_PORT:-5432}"
PG_BJ_USER="${PG_BJ_USER:-postgres}"
PG_BJ_DB="${PG_BJ_DB:-industrial_ai}"
PG_BJ_VIP="${PG_BJ_VIP:-10.1.1.10}"

PG_SH_HOST="${PG_SH_HOST:-10.2.1.10}"
PG_SH_PORT="${PG_SH_PORT:-5432}"
PG_SH_USER="${PG_SH_USER:-postgres}"
PG_SH_DB="${PG_SH_DB:-industrial_ai}"
PG_SH_VIP="${PG_SH_VIP:-10.2.1.10}"

# Redis 配置
REDIS_BJ_HOST="${REDIS_BJ_HOST:-10.1.1.20}"
REDIS_BJ_PORT="${REDIS_BJ_PORT:-6379}"
REDIS_BJ_SENTINEL_HOSTS="${REDIS_BJ_SENTINEL_HOSTS:-10.1.1.101:26379,10.1.1.102:26379,10.1.1.103:26379}"

REDIS_SH_HOST="${REDIS_SH_HOST:-10.2.1.20}"
REDIS_SH_PORT="${REDIS_SH_PORT:-6379}"
REDIS_SH_SENTINEL_HOSTS="${REDIS_SH_SENTINEL_HOSTS:-10.2.1.101:26379,10.2.1.102:26379,10.2.1.103:26379}"

# Kubernetes 配置
K8S_BJ_CONTEXT="${K8S_BJ_CONTEXT:-dc-bj}"
K8S_SH_CONTEXT="${K8S_SH_CONTEXT:-dc-sh}"
K8S_NAMESPACE="${K8S_NAMESPACE:-industrial-ai}"

# GSLB 配置
GSLB_PROVIDER="${GSLB_PROVIDER:-dns}"  # dns, f5, cloudflare
GSLB_DOMAIN="${GSLB_DOMAIN:-api.industrial-ai.com}"
GSLB_BJ_RECORD="${GSLB_BJ_RECORD:-bj.api.industrial-ai.com}"
GSLB_SH_RECORD="${GSLB_SH_RECORD:-sh.api.industrial-ai.com}"

# Patroni 配置
PATRONI_BJ_NAMESPACE="${PATRONI_BJ_NAMESPACE:-/service/industrial-ai-dc-bj}"
PATRONI_SH_NAMESPACE="${PATRONI_SH_NAMESPACE:-/service/industrial-ai-dc-sh}"

# 告警配置
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"
ALERT_EMAIL="${ALERT_EMAIL:-}"
ALERT_SMS_URL="${ALERT_SMS_URL:-}"

# 切换锁文件
SWITCH_LOCK_FILE="/tmp/disaster-recovery.lock"

# 切换日志
SWITCH_LOG="/var/log/disaster-recovery.log"

# 备份目录
BACKUP_DIR="/var/lib/disaster-recovery"

# =============================================================================
# 工具函数
# =============================================================================

# 发送告警
send_alert() {
    local level=$1
    local message=$2
    
    echo "[$level] $(date '+%Y-%m-%d %H:%M:%S') $message" >> "$SWITCH_LOG"
    
    # 企业微信/钉钉告警
    if [ -n "$ALERT_WEBHOOK" ]; then
        curl -s -X POST "$ALERT_WEBHOOK" \
            -H 'Content-Type: application/json' \
            -d "{\"msgtype\":\"text\",\"text\":{\"content\":\"[$level] 灾备切换: $message\"}}" > /dev/null 2>&1 || true
    fi
    
    # 邮件告警
    if [ -n "$ALERT_EMAIL" ]; then
        echo "$message" | mail -s "[Disaster Recovery] $level Alert" "$ALERT_EMAIL" || true
    fi
    
    # 短信告警 (Critical级别)
    if [ "$level" = "CRITICAL" ] && [ -n "$ALERT_SMS_URL" ]; then
        curl -s -X POST "$ALERT_SMS_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"message\":\"$message\"}" > /dev/null 2>&1 || true
    fi
}

# 获取切换锁
acquire_switch_lock() {
    if [ -f "$SWITCH_LOCK_FILE" ]; then
        local pid=$(cat "$SWITCH_LOCK_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_error "切换进程已在运行 (PID: $pid)"
            return 1
        else
            log_warn "发现过期的锁文件，清理中..."
            rm -f "$SWITCH_LOCK_FILE"
        fi
    fi
    
    echo $$ > "$SWITCH_LOCK_FILE"
    trap "rm -f $SWITCH_LOCK_FILE" EXIT
    log_info "获取切换锁成功"
}

# 检查命令是否存在
check_command() {
    command -v "$1" > /dev/null 2>&1 || {
        log_error "命令不存在: $1"
        return 1
    }
}

# 确认操作
confirm() {
    local message=$1
    local force=${2:-false}
    
    if [ "$force" = "true" ]; then
        return 0
    fi
    
    echo -e "${YELLOW}警告: $message${NC}"
    echo -n "确认执行? (yes/no): "
    read -r response
    
    if [ "$response" != "yes" ]; then
        log_info "操作已取消"
        return 1
    fi
    
    return 0
}

# 记录切换状态
record_switch_state() {
    local from_dc=$1
    local to_dc=$2
    local status=$3
    
    local record_file="$BACKUP_DIR/switch-history.log"
    mkdir -p "$BACKUP_DIR"
    
    echo "$(date '+%Y-%m-%d %H:%M:%S') | $from_dc -> $to_dc | $status | $$" >> "$record_file"
}

# =============================================================================
# 健康检查函数
# =============================================================================

# 检查 PostgreSQL 状态
check_postgresql() {
    local host=$1
    local port=$2
    local dc=$3
    
    log_info "检查 PostgreSQL [$dc]: $host:$port"
    
    # 检查连接
    if ! PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d postgres -c "SELECT 1" > /dev/null 2>&1; then
        log_error "PostgreSQL [$dc] 连接失败"
        return 1
    fi
    
    # 检查角色
    local role=$(PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d postgres -tAc \
        "SELECT pg_is_in_recovery()")
    
    if [ "$role" = "t" ]; then
        log_info "PostgreSQL [$dc] 角色: Standby"
        echo "standby"
    else
        log_info "PostgreSQL [$dc] 角色: Primary"
        echo "primary"
    fi
    
    return 0
}

# 检查 Redis 状态
check_redis() {
    local host=$1
    local port=$2
    local dc=$3
    
    log_info "检查 Redis [$dc]: $host:$port"
    
    # 检查连接
    if ! redis-cli -h "$host" -p "$port" -a "${REDIS_PASSWORD:-}" --no-auth-warning PING > /dev/null 2>&1; then
        log_error "Redis [$dc] 连接失败"
        return 1
    fi
    
    # 检查角色
    local role=$(redis-cli -h "$host" -p "$port" -a "${REDIS_PASSWORD:-}" --no-auth-warning INFO replication | grep role | cut -d: -f2 | tr -d '\r')
    
    log_info "Redis [$dc] 角色: $role"
    echo "$role"
    
    return 0
}

# 检查 Kubernetes 状态
check_kubernetes() {
    local context=$1
    local dc=$2
    
    log_info "检查 Kubernetes [$dc]: $context"
    
    if ! kubectl --context "$context" get ns "$K8S_NAMESPACE" > /dev/null 2>&1; then
        log_warn "Kubernetes [$dc] namespace 不存在或无法访问"
        return 1
    fi
    
    # 检查 Pod 状态
    local pod_count=$(kubectl --context "$context" -n "$K8S_NAMESPACE" get pods --no-headers 2>/dev/null | wc -l)
    local ready_count=$(kubectl --context "$context" -n "$K8S_NAMESPACE" get pods --no-headers 2>/dev/null | grep "Running" | wc -l)
    
    log_info "Kubernetes [$dc] Pod: $ready_count/$pod_count Running"
    
    if [ "$ready_count" -eq 0 ]; then
        log_warn "Kubernetes [$dc] 没有运行中的 Pod"
        return 1
    fi
    
    return 0
}

# 检查应用健康
check_app_health() {
    local host=$1
    local dc=$2
    
    log_info "检查应用健康 [$dc]: http://$host/health"
    
    local response=$(curl -s -o /dev/null -w "%{http_code}" "http://$host/health" --connect-timeout 5 --max-time 10 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ]; then
        log_info "应用健康检查 [$dc] 通过"
        return 0
    else
        log_error "应用健康检查 [$dc] 失败: HTTP $response"
        return 1
    fi
}

# =============================================================================
# PostgreSQL 切换函数
# =============================================================================

# 提升 PostgreSQL 为 Primary
promote_postgresql() {
    local host=$1
    local port=$2
    local dc=$3
    
    log_step "提升 PostgreSQL [$dc] 为 Primary"
    
    # 检查当前状态
    local role=$(check_postgresql "$host" "$port" "$dc")
    
    if [ "$role" = "primary" ]; then
        log_warn "PostgreSQL [$dc] 已经是 Primary"
        return 0
    fi
    
    # 停止订阅 (如果是灾备数据中心)
    log_info "禁用订阅..."
    PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d "$PG_BJ_DB" -c \
        "ALTER SUBSCRIPTION IF EXISTS cross_dc_subscription DISABLE" 2>/dev/null || true
    
    # 通过 Patroni 提升
    if command -v patronictl > /dev/null 2>&1; then
        log_info "使用 Patroni 提升..."
        patronictl -c /etc/patroni.yml switchover --master "$host" --force || {
            log_warn "Patroni 提升 failed, 尝试手动提升"
            # 手动提升
            PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d postgres -c "SELECT pg_promote()"
        }
    else
        log_info "手动提升 PostgreSQL..."
        PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d postgres -c "SELECT pg_promote()"
    fi
    
    # 等待提升完成
    sleep 5
    
    # 验证
    local new_role=$(check_postgresql "$host" "$port" "$dc")
    if [ "$new_role" = "primary" ]; then
        log_info "PostgreSQL [$dc] 提升成功"
        send_alert INFO "PostgreSQL [$dc] 已提升为 Primary"
        return 0
    else
        log_error "PostgreSQL [$dc] 提升失败"
        return 1
    fi
}

# 将 PostgreSQL 降级为 Standby
demote_postgresql() {
    local host=$1
    local port=$2
    local master_host=$3
    local master_port=$4
    local dc=$5
    
    log_step "将 PostgreSQL [$dc] 降级为 Standby"
    
    # 检查当前状态
    local role=$(check_postgresql "$host" "$port" "$dc")
    
    if [ "$role" = "standby" ]; then
        log_warn "PostgreSQL [$dc] 已经是 Standby"
        return 0
    fi
    
    # 创建订阅
    log_info "创建订阅到主数据中心..."
    local conn_string="host=$master_host port=$master_port user=${PG_BJ_USER} password=${PG_PASSWORD:-} dbname=$PG_BJ_DB"
    
    PGPASSWORD="${PG_PASSWORD:-}" psql -h "$host" -p "$port" -U "${PG_BJ_USER}" -d "$PG_BJ_DB" <<EOF
-- 删除旧订阅 (如果存在)
DROP SUBSCRIPTION IF EXISTS cross_dc_subscription;

-- 创建新订阅
CREATE SUBSCRIPTION cross_dc_subscription
    CONNECTION '$conn_string'
    PUBLICATION cross_dc_publication
    WITH (
        copy_data = false,
        create_slot = true,
        slot_name = 'cross_dc_slot_back',
        synchronous_commit = off
    );

-- 启用订阅
ALTER SUBSCRIPTION cross_dc_subscription ENABLE;
EOF
    
    # 等待同步
    log_info "等待初始同步..."
    sleep 10
    
    # 验证
    local new_role=$(check_postgresql "$host" "$port" "$dc")
    if [ "$new_role" = "standby" ]; then
        log_info "PostgreSQL [$dc] 降级成功"
        send_alert INFO "PostgreSQL [$dc] 已降级为 Standby"
        return 0
    else
        log_error "PostgreSQL [$dc] 降级失败"
        return 1
    fi
}

# =============================================================================
# Redis 切换函数
# =============================================================================

# 提升 Redis 为 Master
promote_redis() {
    local host=$1
    local port=$2
    local dc=$3
    
    log_step "提升 Redis [$dc] 为 Master"
    
    # 检查当前状态
    local role=$(check_redis "$host" "$port" "$dc")
    
    if [ "$role" = "master" ]; then
        log_warn "Redis [$dc] 已经是 Master"
        return 0
    fi
    
    # 断开复制
    redis-cli -h "$host" -p "$port" -a "${REDIS_PASSWORD:-}" --no-auth-warning REPLICAOF NO ONE > /dev/null 2>&1
    
    # 等待
    sleep 2
    
    # 更新 Sentinel 配置
    local sentinel_hosts="${REDIS_SH_SENTINEL_HOSTS}"
    if [ "$dc" = "$DC_BJ_NAME" ]; then
        sentinel_hosts="${REDIS_BJ_SENTINEL_HOSTS}"
    fi
    
    IFS=',' read -ra SENTINELS <<< "$sentinel_hosts"
    for sentinel in "${SENTINELS[@]}"; do
        local s_host=$(echo "$sentinel" | cut -d: -f1)
        local s_port=$(echo "$sentinel" | cut -d: -f2)
        log_info "更新 Sentinel: $s_host:$s_port"
        redis-cli -h "$s_host" -p "$s_port" SENTINEL MONITOR industrial-ai-master "$host" "$port" 2 > /dev/null 2>&1 || true
        redis-cli -h "$s_host" -p "$s_port" SENTINEL SET industrial-ai-master auth-pass "${REDIS_PASSWORD:-}" > /dev/null 2>&1 || true
    done
    
    # 验证
    local new_role=$(check_redis "$host" "$port" "$dc")
    if [ "$new_role" = "master" ]; then
        log_info "Redis [$dc] 提升成功"
        send_alert INFO "Redis [$dc] 已提升为 Master"
        return 0
    else
        log_error "Redis [$dc] 提升失败"
        return 1
    fi
}

# 将 Redis 降级为 Replica
demote_redis() {
    local host=$1
    local port=$2
    local master_host=$3
    local master_port=$4
    local dc=$5
    
    log_step "将 Redis [$dc] 降级为 Replica"
    
    # 检查当前状态
    local role=$(check_redis "$host" "$port" "$dc")
    
    if [ "$role" = "slave" ]; then
        log_warn "Redis [$dc] 已经是 Replica"
        return 0
    fi
    
    # 设置为从库
    redis-cli -h "$host" -p "$port" -a "${REDIS_PASSWORD:-}" --no-auth-warning REPLICAOF "$master_host" "$master_port" > /dev/null 2>&1
    
    # 等待同步
    log_info "等待数据同步..."
    sleep 5
    
    # 验证
    local new_role=$(check_redis "$host" "$port" "$dc")
    if [ "$new_role" = "slave" ]; then
        log_info "Redis [$dc] 降级成功"
        send_alert INFO "Redis [$dc] 已降级为 Replica"
        return 0
    else
        log_error "Redis [$dc] 降级失败"
        return 1
    fi
}

# =============================================================================
# Kubernetes 切换函数
# =============================================================================

# 启动 Kubernetes 应用
start_kubernetes_app() {
    local context=$1
    local dc=$2
    
    log_step "启动 Kubernetes 应用 [$dc]"
    
    # 扩容 Deployment
    kubectl --context "$context" -n "$K8S_NAMESPACE" scale deployment --all --replicas=3 2>/dev/null || {
        log_warn "扩容 Deployment 失败"
    }
    
    # 等待 Pod 就绪
    log_info "等待 Pod 就绪..."
    kubectl --context "$context" -n "$K8S_NAMESPACE" rollout status deployment --timeout=300s 2>/dev/null || {
        log_error "Pod 启动超时"
        return 1
    }
    
    # 验证
    local ready_count=$(kubectl --context "$context" -n "$K8S_NAMESPACE" get pods --no-headers 2>/dev/null | grep "Running" | wc -l)
    if [ "$ready_count" -gt 0 ]; then
        log_info "应用启动成功: $ready_count 个 Pod 运行中"
        return 0
    else
        log_error "应用启动失败"
        return 1
    fi
}

# 停止 Kubernetes 应用
stop_kubernetes_app() {
    local context=$1
    local dc=$2
    
    log_step "停止 Kubernetes 应用 [$dc]"
    
    # 缩容 Deployment
    kubectl --context "$context" -n "$K8S_NAMESPACE" scale deployment --all --replicas=0 2>/dev/null || {
        log_warn "缩容 Deployment 失败"
    }
    
    # 等待 Pod 终止
    log_info "等待 Pod 终止..."
    sleep 10
    
    log_info "应用已停止"
}

# =============================================================================
# DNS/GSLB 切换函数
# =============================================================================

# 更新 GSLB
update_gslb() {
    local primary_dc=$1
    local force=${2:-false}
    
    log_step "更新 GSLB 配置"
    
    case "$GSLB_PROVIDER" in
        dns)
            update_dns_record "$primary_dc"
            ;;
        f5)
            update_f5_gslb "$primary_dc"
            ;;
        cloudflare)
            update_cloudflare_lb "$primary_dc"
            ;;
        *)
            log_error "未知的 GSLB 提供商: $GSLB_PROVIDER"
            return 1
            ;;
    esac
}

# 更新 DNS 记录
update_dns_record() {
    local primary_dc=$1
    
    log_info "更新 DNS 记录..."
    
    # 示例: 使用 nsupdate
    if command -v nsupdate > /dev/null 2>&1; then
        local primary_record
        if [ "$primary_dc" = "beijing" ]; then
            primary_record="$GSLB_BJ_RECORD"
        else
            primary_record="$GSLB_SH_RECORD"
        fi
        
        nsupdate <<EOF
server ${DNS_SERVER:-dns.internal}
zone ${GSLB_DOMAIN}
update delete ${GSLB_DOMAIN}
update add ${GSLB_DOMAIN} 60 CNAME ${primary_record}
send
EOF
        
        log_info "DNS 记录已更新: $GSLB_DOMAIN -> $primary_record"
    else
        log_warn "nsupdate 不可用，请手动更新 DNS 记录"
        log_warn "将 $GSLB_DOMAIN 指向 $primary_dc 数据中心"
    fi
    
    # 等待 DNS 传播
    log_info "等待 DNS 传播 (60秒)..."
    sleep 60
}

# 更新 F5 GSLB
update_f5_gslb() {
    local primary_dc=$1
    
    log_info "更新 F5 GSLB 配置..."
    
    # 示例: 使用 F5 REST API
    # 需要 F5 管理 IP 和认证信息
    
    log_warn "F5 GSLB 更新需要手动配置或集成 F5 REST API"
}

# 更新 Cloudflare Load Balancer
update_cloudflare_lb() {
    local primary_dc=$1
    
    log_info "更新 Cloudflare Load Balancer..."
    
    # 示例: 使用 Cloudflare API
    if [ -n "$CF_API_TOKEN" ]; then
        curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/load_balancers/$CF_LB_ID" \
            -H "Authorization: Bearer $CF_API_TOKEN" \
            -H "Content-Type: application/json" \
            --data "{\"default_pool_ids\":[\"$primary_dc-pool\"]}" > /dev/null 2>&1 || {
            log_warn "Cloudflare LB 更新失败"
        }
        log_info "Cloudflare LB 已更新"
    else
        log_warn "CF_API_TOKEN 未设置，跳过 Cloudflare LB 更新"
    fi
}

# =============================================================================
# 切换流程
# =============================================================================

# 切换到灾备数据中心
switch_to_dr() {
    local force=${1:-false}
    
    log_section "开始切换到灾备数据中心 ($DC_SH_NAME)"
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 获取锁
    acquire_switch_lock || return 1
    
    # 发送告警
    send_alert INFO "开始灾备切换: $DC_BJ_NAME -> $DC_SH_NAME"
    
    # 确认操作
    confirm "即将切换到灾备数据中心 ($DC_SH_NAME)，此操作会影响服务可用性" "$force" || return 1
    
    # Step 1: 验证灾备数据中心状态
    log_step "Step 1: 验证灾备数据中心状态"
    check_postgresql "$PG_SH_HOST" "$PG_SH_PORT" "$DC_SH_NAME" || {
        if [ "$force" != "true" ]; then
            send_alert CRITICAL "灾备数据中心 PostgreSQL 不可用，切换中止"
            return 1
        fi
        log_warn "灾备数据中心 PostgreSQL 不可用，强制切换"
    }
    
    check_redis "$REDIS_SH_HOST" "$REDIS_SH_PORT" "$DC_SH_NAME" || {
        if [ "$force" != "true" ]; then
            send_alert CRITICAL "灾备数据中心 Redis 不可用，切换中止"
            return 1
        fi
        log_warn "灾备数据中心 Redis 不可用，强制切换"
    }
    
    # Step 2: 停止主数据中心应用
    log_step "Step 2: 停止主数据中心应用"
    stop_kubernetes_app "$K8S_BJ_CONTEXT" "$DC_BJ_NAME" || true
    
    # Step 3: 提升 PostgreSQL
    log_step "Step 3: 提升灾备 PostgreSQL 为 Primary"
    promote_postgresql "$PG_SH_HOST" "$PG_SH_PORT" "$DC_SH_NAME" || {
        send_alert CRITICAL "PostgreSQL 提升失败"
        return 1
    }
    
    # Step 4: 提升 Redis
    log_step "Step 4: 提升灾备 Redis 为 Master"
    promote_redis "$REDIS_SH_HOST" "$REDIS_SH_PORT" "$DC_SH_NAME" || {
        send_alert CRITICAL "Redis 提升失败"
        return 1
    }
    
    # Step 5: 启动灾备数据中心应用
    log_step "Step 5: 启动灾备数据中心应用"
    start_kubernetes_app "$K8S_SH_CONTEXT" "$DC_SH_NAME" || {
        send_alert CRITICAL "应用启动失败"
        return 1
    }
    
    # Step 6: 更新 DNS/GSLB
    log_step "Step 6: 更新 DNS/GSLB"
    update_gslb "shanghai" "$force" || {
        log_warn "DNS/GSLB 更新失败，请手动处理"
    }
    
    # Step 7: 验证服务
    log_step "Step 7: 验证服务可用性"
    sleep 30  # 等待服务稳定
    
    check_app_health "$GSLB_DOMAIN" "灾备" || {
        send_alert WARNING "服务健康检查失败，请人工验证"
    }
    
    # 记录
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    record_switch_state "$DC_BJ_NAME" "$DC_SH_NAME" "success"
    
    log_section "切换完成"
    log_info "总耗时: ${duration} 秒"
    log_info "RTO: ${duration} 秒 (目标 < 300秒)"
    
    send_alert INFO "灾备切换完成: $DC_BJ_NAME -> $DC_SH_NAME (耗时: ${duration}秒)"
}

# 回切到主数据中心
switch_back() {
    local force=${1:-false}
    
    log_section "开始回切到主数据中心 ($DC_BJ_NAME)"
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 获取锁
    acquire_switch_lock || return 1
    
    # 发送告警
    send_alert INFO "开始回切: $DC_SH_NAME -> $DC_BJ_NAME"
    
    # 确认操作
    confirm "即将回切到主数据中心 ($DC_BJ_NAME)，此操作会影响服务可用性" "$force" || return 1
    
    # Step 1: 验证主数据中心状态
    log_step "Step 1: 验证主数据中心状态"
    check_postgresql "$PG_BJ_HOST" "$PG_BJ_PORT" "$DC_BJ_NAME" || {
        send_alert CRITICAL "主数据中心 PostgreSQL 不可用，回切中止"
        return 1
    }
    
    check_redis "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" "$DC_BJ_NAME" || {
        send_alert CRITICAL "主数据中心 Redis 不可用，回切中止"
        return 1
    }
    
    # Step 2: 同步数据差异 (灾备 -> 主)
    log_step "Step 2: 同步数据差异"
    if [ -f "/usr/local/bin/cross-dc-sync.sh" ]; then
        /usr/local/bin/cross-dc-sync.sh sync-to-dc beijing || {
            log_warn "数据同步失败，继续回切"
        }
    else
        log_warn "cross-dc-sync.sh 不存在，跳过数据同步"
    fi
    
    # Step 3: 停止灾备数据中心应用
    log_step "Step 3: 停止灾备数据中心应用"
    stop_kubernetes_app "$K8S_SH_CONTEXT" "$DC_SH_NAME" || true
    
    # Step 4: 降级灾备 PostgreSQL
    log_step "Step 4: 降级灾备 PostgreSQL 为 Standby"
    demote_postgresql "$PG_SH_HOST" "$PG_SH_PORT" "$PG_BJ_HOST" "$PG_BJ_PORT" "$DC_SH_NAME" || {
        send_alert CRITICAL "PostgreSQL 降级失败"
        return 1
    }
    
    # Step 5: 提升 PostgreSQL
    log_step "Step 5: 提升主 PostgreSQL 为 Primary"
    promote_postgresql "$PG_BJ_HOST" "$PG_BJ_PORT" "$DC_BJ_NAME" || {
        send_alert CRITICAL "PostgreSQL 提升失败"
        return 1
    }
    
    # Step 6: 降级灾备 Redis
    log_step "Step 6: 降级灾备 Redis 为 Replica"
    demote_redis "$REDIS_SH_HOST" "$REDIS_SH_PORT" "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" "$DC_SH_NAME" || {
        send_alert WARNING "Redis 降级失败"
    }
    
    # Step 7: 提升主 Redis
    log_step "Step 7: 提升主 Redis 为 Master"
    promote_redis "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" "$DC_BJ_NAME" || {
        send_alert CRITICAL "Redis 提升失败"
        return 1
    }
    
    # Step 8: 启动主数据中心应用
    log_step "Step 8: 启动主数据中心应用"
    start_kubernetes_app "$K8S_BJ_CONTEXT" "$DC_BJ_NAME" || {
        send_alert CRITICAL "应用启动失败"
        return 1
    }
    
    # Step 9: 更新 DNS/GSLB
    log_step "Step 9: 更新 DNS/GSLB"
    update_gslb "beijing" "$force" || {
        log_warn "DNS/GSLB 更新失败，请手动处理"
    }
    
    # Step 10: 验证服务
    log_step "Step 10: 验证服务可用性"
    sleep 30
    
    check_app_health "$GSLB_DOMAIN" "主" || {
        send_alert WARNING "服务健康检查失败，请人工验证"
    }
    
    # 记录
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    record_switch_state "$DC_SH_NAME" "$DC_BJ_NAME" "success"
    
    log_section "回切完成"
    log_info "总耗时: ${duration} 秒"
    
    send_alert INFO "回切完成: $DC_SH_NAME -> $DC_BJ_NAME (耗时: ${duration}秒)"
}

# 紧急切换
emergency_switch() {
    log_section "紧急切换到灾备数据中心"
    
    send_alert CRITICAL "触发紧急切换"
    
    # 强制切换，跳过所有确认
    switch_to_dr true
}

# 验证数据中心状态
verify_dc() {
    local dc=${1:-all}
    
    log_section "数据中心状态验证"
    
    if [ "$dc" = "beijing" ] || [ "$dc" = "all" ]; then
        log_info "=== $DC_BJ_NAME ==="
        check_postgresql "$PG_BJ_HOST" "$PG_BJ_PORT" "$DC_BJ_NAME"
        check_redis "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" "$DC_BJ_NAME"
        check_kubernetes "$K8S_BJ_CONTEXT" "$DC_BJ_NAME"
        check_app_health "$GSLB_BJ_RECORD" "$DC_BJ_NAME"
    fi
    
    if [ "$dc" = "shanghai" ] || [ "$dc" = "all" ]; then
        log_info "=== $DC_SH_NAME ==="
        check_postgresql "$PG_SH_HOST" "$PG_SH_PORT" "$DC_SH_NAME"
        check_redis "$REDIS_SH_HOST" "$REDIS_SH_PORT" "$DC_SH_NAME"
        check_kubernetes "$K8S_SH_CONTEXT" "$DC_SH_NAME"
        check_app_health "$GSLB_SH_RECORD" "$DC_SH_NAME"
    fi
}

# 健康检查
health_check() {
    log_section "健康检查"
    
    # 检查所有组件
    local status=0
    
    # PostgreSQL
    if ! check_postgresql "$PG_BJ_HOST" "$PG_BJ_PORT" "$DC_BJ_NAME" > /dev/null 2>&1; then
        log_error "PostgreSQL [$DC_BJ_NAME] 异常"
        status=1
    fi
    
    if ! check_postgresql "$PG_SH_HOST" "$PG_SH_PORT" "$DC_SH_NAME" > /dev/null 2>&1; then
        log_error "PostgreSQL [$DC_SH_NAME] 异常"
        status=1
    fi
    
    # Redis
    if ! check_redis "$REDIS_BJ_HOST" "$REDIS_BJ_PORT" "$DC_BJ_NAME" > /dev/null 2>&1; then
        log_error "Redis [$DC_BJ_NAME] 异常"
        status=1
    fi
    
    if ! check_redis "$REDIS_SH_HOST" "$REDIS_SH_PORT" "$DC_SH_NAME" > /dev/null 2>&1; then
        log_error "Redis [$DC_SH_NAME] 异常"
        status=1
    fi
    
    # 应用
    if ! check_app_health "$GSLB_DOMAIN" "Current" > /dev/null 2>&1; then
        log_error "应用健康检查失败"
        status=1
    fi
    
    if [ $status -eq 0 ]; then
        log_info "所有组件健康"
        echo "HEALTHY"
    else
        log_error "存在不健康组件"
        echo "UNHEALTHY"
    fi
    
    return $status
}

# =============================================================================
# 主程序
# =============================================================================

usage() {
    echo "用法: $0 <command> [options]"
    echo ""
    echo "命令:"
    echo "  switch-to-dr      切换到灾备数据中心 (上海)"
    echo "  switch-back       回切到主数据中心 (北京)"
    echo "  verify [dc]       验证数据中心状态 (beijing/shanghai/all)"
    echo "  health-check      执行健康检查"
    echo "  emergency-switch  紧急切换 (跳过确认和部分验证)"
    echo ""
    echo "选项:"
    echo "  -f, --force       强制执行，跳过确认"
    echo "  -h, --help        显示帮助信息"
    echo ""
    echo "环境变量:"
    echo "  PG_BJ_HOST        北京 PostgreSQL 主机"
    echo "  PG_SH_HOST        上海 PostgreSQL 主机"
    echo "  REDIS_BJ_HOST     北京 Redis 主机"
    echo "  REDIS_SH_HOST     上海 Redis 主机"
    echo "  K8S_BJ_CONTEXT    北京 Kubernetes 上下文"
    echo "  K8S_SH_CONTEXT    上海 Kubernetes 上下文"
    echo "  ALERT_WEBHOOK     告警 Webhook URL"
    echo ""
    echo "示例:"
    echo "  $0 switch-to-dr              # 切换到灾备数据中心"
    echo "  $0 switch-back               # 回切到主数据中心"
    echo "  $0 verify beijing            # 验证北京数据中心"
    echo "  $0 health-check             # 健康检查"
    echo "  $0 emergency-switch          # 紧急切换"
}

# 解析参数
force=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--force)
            force=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

case "${1:-}" in
    switch-to-dr)
        switch_to_dr "$force"
        ;;
    switch-back)
        switch_back "$force"
        ;;
    verify)
        verify_dc "${2:-all}"
        ;;
    health-check)
        health_check
        ;;
    emergency-switch)
        emergency_switch
        ;;
    *)
        usage
        exit 1
        ;;
esac