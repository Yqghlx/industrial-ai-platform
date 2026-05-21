# Redis 跨数据中心同步方案

## 1. 方案概述

### 1.1 目标
- 实现北京(主)与上海(灾备)数据中心的 Redis 数据同步
- RPO < 15 分钟
- 自动故障切换能力
- 数据一致性保障

### 1.2 架构拓扑

```
┌─────────────────────────────────────────────────────────────────┐
│                     主数据中心 (北京 DC-BJ)                     │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Redis Sentinel                       │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                 │   │
│  │  │Sentinel1│  │Sentinel2│  │Sentinel3│                 │   │
│  │  └────┬────┘  └────┬────┘  └────┬────┘                 │   │
│  │       └────────────┼────────────┘                      │   │
│  │                    │ 监控 & 故障切换                    │   │
│  └────────────────────┼────────────────────────────────────┘   │
│                       │                                          │
│       ┌───────────────┼───────────────┐                        │
│       │               │               │                        │
│  ┌────▼────┐    ┌────▼────┐    ┌────▼────┐                    │
│  │ Master  │───▶│ Replica1│───▶│ Replica2│                    │
│  │10.1.1.20│    │10.1.1.21│    │10.1.1.22│                    │
│  └────┬────┘    └─────────┘    └─────────┘                    │
│       │                                                        │
│       │ 跨DC同步 (CRDT/RedisShake)                             │
└───────┼────────────────────────────────────────────────────────┘
        │
        │ 专用线路/VPN
        │
┌───────┼────────────────────────────────────────────────────────┐
│       │                                                        │
│  ┌────▼────┐    ┌─────────┐    ┌─────────┐                    │
│  │ Master  │───▶│ Replica1│───▶│ Replica2│                    │
│  │10.2.1.20│    │10.2.1.21│    │10.2.1.22│                    │
│  └─────────┘    └─────────┘    └─────────┘                    │
│       │                                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Redis Sentinel                       │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                 │   │
│  │  │Sentinel1│  │Sentinel2│  │Sentinel3│                 │   │
│  │  └─────────┘  └─────────┘  └─────────┘                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│                     灾备数据中心 (上海 DC-SH)                    │
└─────────────────────────────────────────────────────────────────┘
```

## 2. 同步方案选择

### 2.1 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|-----|------|------|---------|
| **RedisShake** | 成熟稳定、支持全量+增量 | 单向同步、需外部工具 | 主备模式 |
| **Redis-CRDT** | 双向同步、冲突自动解决 | 社区版有限制 | 双活模式 |
| **自定义脚本** | 灵活可控 | 开发维护成本高 | 特殊需求 |
| **Redis Enterprise** | 原生支持、功能强大 | 商业版本、成本高 | 企业级 |

### 2.2 推荐方案

**混合方案: Redis Sentinel (数据中心内) + RedisShake (跨数据中心)**

- 数据中心内: Sentinel 高可用 + 主从复制
- 跨数据中心: RedisShake 数据同步
- 故障切换: Sentinel 自动管理

## 3. 数据中心内配置

### 3.1 主数据中心 (北京) - Sentinel 配置

#### redis.conf (Master)
```conf
# redis-master.conf

# 基础配置
bind 0.0.0.0
port 6379
protected-mode no
daemonize yes
pidfile /var/run/redis/redis-server.pid
logfile /var/log/redis/redis-server.log
dir /var/lib/redis

# 内存配置
maxmemory 16gb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec
appendfilename "appendonly.aof"

# 复制配置
replica-serve-stale-data yes
replica-read-only yes
repl-diskless-sync yes
repl-timeout 60

# 性能优化
tcp-backlog 511
tcp-keepalive 300
timeout 0

# 安全配置
requirepass your-secure-password
masterauth your-secure-password

# 跨DC同步专用键空间通知
notify-keyspace-events Ex
```

#### sentinel.conf
```conf
# sentinel.conf - 主数据中心

port 26379
daemonize yes
pidfile /var/run/redis/redis-sentinel.pid
logfile /var/log/redis/redis-sentinel.log
dir /var/lib/redis

# 监控配置
sentinel monitor industrial-ai-master 10.1.1.20 6379 2
sentinel auth-pass industrial-ai-master your-secure-password

# 故障检测配置
sentinel down-after-milliseconds industrial-ai-master 10000
sentinel parallel-syncs industrial-ai-master 1
sentinel failover-timeout industrial-ai-master 60000

# 通知脚本
sentinel notification-script industrial-ai-master /etc/redis/notify.sh
sentinel client-reconfig-script industrial-ai-master /etc/redis/reconfig.sh

# 禁止自动故障切换 (跨DC时手动控制)
# sentinel deny-scripts-reconfig yes

# 暴露给客户端的 Master 名称
sentinel announce-ip 10.1.1.20
sentinel announce-port 26379
```

### 3.2 灾备数据中心 (上海) - Sentinel 配置

#### redis.conf (初始为 Replica)
```conf
# redis-slave.conf

# 基础配置 (同主数据中心)
bind 0.0.0.0
port 6379
protected-mode no
daemonize yes
pidfile /var/run/redis/redis-server.pid
logfile /var/log/redis/redis-server.log
dir /var/lib/redis

# 内存配置
maxmemory 16gb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec

# 复制配置 - 初始从北京主库同步
replicaof 10.1.1.20 6379
replica-serve-stale-data yes
replica-read-only yes
repl-diskless-sync yes
repl-timeout 60

# 安全配置
requirepass your-secure-password
masterauth your-secure-password
```

#### sentinel.conf
```conf
# sentinel.conf - 灾备数据中心

port 26379
daemonize yes
pidfile /var/run/redis/redis-sentinel.pid
logfile /var/log/redis/redis-sentinel.log
dir /var/lib/redis

# 监控配置
sentinel monitor industrial-ai-master 10.2.1.20 6379 2
sentinel auth-pass industrial-ai-master your-secure-password

# 故障检测配置
sentinel down-after-milliseconds industrial-ai-master 10000
sentinel parallel-syncs industrial-ai-master 1
sentinel failover-timeout industrial-ai-master 60000

# 通知脚本
sentinel notification-script industrial-ai-master /etc/redis/notify.sh
```

## 4. 跨数据中心同步

### 4.1 RedisShake 配置

#### 安装 RedisShake
```bash
# 在两个数据中心都安装
wget https://github.com/alibaba/RedisShake/releases/download/v4.0.0/redisshake-linux-amd64.tar.gz
tar -xzf redisshake-linux-amd64.tar.gz -C /usr/local/bin/
chmod +x /usr/local/bin/redis-shake
```

#### shake.toml (北京 → 上海)
```toml
# shake.toml - 北京主数据中心同步到上海

# 同步模式
type = "sync"

# 源 Redis (北京主库)
[source]
address = "10.1.1.20:6379"
password = "your-secure-password"

# 目标 Redis (上海灾备)
[target]
address = "10.2.1.20:6379"
password = "your-secure-password"

# 过滤规则 (可选)
# 只同步特定前缀的键
#[filter]
#allow_prefixes = ["session:", "cache:", "device:"]

# 高级配置
[advanced]
# 并发数
parallel = 32

# 批量大小
batch_size = 100

# 重试配置
retry_times = 3
retry_interval = 5

# 性能监控
[metrics]
# Prometheus 监控
listen = ":9321"

# 日志
[log]
level = "info"
file = "/var/log/redis-shake/sync.log"
```

#### Systemd 服务
```ini
# /etc/systemd/system/redis-shake.service

[Unit]
Description=RedisShake Sync Service
After=network.target redis.service
Wants=redis.service

[Service]
Type=simple
User=redis
Group=redis
ExecStart=/usr/local/bin/redis-shake /etc/redis-shake/shake.toml
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 4.2 双向同步方案 (可选)

如果需要双向同步，使用 Redis CRDT 模块：

#### 安装 Redis CRDT
```bash
# 两个数据中心都安装
git clone https://github.com/redis/redis-crdt.git
cd redis-crdt
make
cp crdt.so /usr/lib/redis/modules/
```

#### redis.conf 添加模块
```conf
# 加载 CRDT 模块
loadmodule /usr/lib/redis/modules/crdt.so

# CRDT 配置
crdt-gid 1  # 北京 GID=1, 上海 GID=2
crdt-vclock-expire 86400
```

## 5. 数据一致性策略

### 5.1 键分类策略

| 键前缀 | 一致性要求 | 同步策略 | TTL |
|-------|-----------|---------|-----|
| `session:` | 最终一致性 | 异步复制 | 30分钟 |
| `cache:` | 最终一致性 | 异步复制 | 1小时 |
| `device:state:` | 强一致性 | 同步写 | 永久 |
| `lock:` | 强一致性 | Redlock | 30秒 |
| `counter:` | 最终一致性 | 合并 | 永久 |

### 5.2 冲突解决策略

#### Last-Write-Wins (LWW)
```lua
-- 在写入时添加时间戳
local time = redis.call('TIME')
redis.call('HSET', KEYS[1], 'value', ARGV[1], 'timestamp', time[1]..time[2])
```

#### CRDT Counter
```lua
-- 使用 Redis CRDT 的 PN-Counter
redis.call('CRDT.INCR', 'counter:visits', 1)
```

### 5.3 数据校验

```bash
# 定期校验脚本 (每日执行)
# /usr/local/bin/redis-verify.sh

#!/bin/bash

BEIJING_REDIS="10.1.1.20:6379"
SHANGHAI_REDIS="10.2.1.20:6379"
PASSWORD="your-secure-password"

# 比较键数量
BJ_KEYS=$(redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD DBSIZE | grep -o '[0-9]*')
SH_KEYS=$(redis-cli -h 10.2.1.20 -p 6379 -a $PASSWORD DBSIZE | grep -o '[0-9]*')

echo "北京键数: $BJ_KEYS"
echo "上海键数: $SH_KEYS"

# 检查差异
DIFF=$((BJ_KEYS - SH_KEYS))
if [ $DIFF -lt 0 ]; then
    DIFF=$((-DIFF))
fi

if [ $DIFF -gt 100 ]; then
    echo "警告: 键数量差异较大 ($DIFF)"
    # 发送告警
fi

# 随机抽样检查
for i in {1..10}; do
    KEY=$(redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD --no-auth-warning RANDOMKEY)
    BJ_VAL=$(redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD --no-auth-warning GET "$KEY")
    SH_VAL=$(redis-cli -h 10.2.1.20 -p 6379 -a $PASSWORD --no-auth-warning GET "$KEY")
    
    if [ "$BJ_VAL" != "$SH_VAL" ]; then
        echo "差异发现: $KEY"
        echo "北京: $BJ_VAL"
        echo "上海: $SH_VAL"
    fi
done
```

## 6. 监控与告警

### 6.1 关键指标

| 指标 | 说明 | 告警阈值 |
|-----|------|---------|
| `connected_clients` | 连接数 | > 1000 |
| `blocked_clients` | 阻塞客户端 | > 10 |
| `used_memory` | 内存使用 | > 80% |
| `replication_lag` | 复制延迟 | > 1000ms |
| `sync_full` | 全量同步次数 | > 0 |
| `sync_partial_err` | 部分同步错误 | > 0 |

### 6.2 监控脚本

```bash
# /usr/local/bin/redis-monitor.sh

#!/bin/bash

REDIS_HOST="10.1.1.20"
REDIS_PORT="6379"
REDIS_PASS="your-secure-password"

# 获取复制延迟
REPLICATION_LAG=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning INFO replication | grep "slave0:" | grep -o 'lag=[0-9]*' | cut -d= -f2)

# 获取内存使用
MEMORY_USED=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning INFO memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')

# 获取连接数
CLIENTS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning INFO clients | grep connected_clients | cut -d: -f2 | tr -d '\r')

echo "redis_replication_lag_seconds{dc=\"beijing\"} $REPLICATION_LAG"
echo "redis_memory_used_bytes{dc=\"beijing\"} $MEMORY_USED"
echo "redis_connected_clients{dc=\"beijing\"} $CLIENTS"
```

### 6.3 Prometheus 配置

```yaml
# prometheus.yml 添加

scrape_configs:
  - job_name: 'redis-exporter-beijing'
    static_configs:
      - targets: ['10.1.1.20:9121']
    relabel_configs:
      - target_label: dc
        replacement: beijing

  - job_name: 'redis-exporter-shanghai'
    static_configs:
      - targets: ['10.2.1.20:9121']
    relabel_configs:
      - target_label: dc
        replacement: shanghai
```

## 7. 故障切换流程

### 7.1 数据中心内故障切换

Sentinel 自动管理，无需人工干预。

### 7.2 跨数据中心故障切换

#### 提升上海为独立集群

```bash
#!/bin/bash
# promote-shanghai-standalone.sh

# 1. 断开与北京的复制
redis-cli -h 10.2.1.20 -p 6379 -a $PASSWORD REPLICAOF NO ONE

# 2. 更新 Sentinel 配置
redis-cli -h 10.2.1.101 -p 26379 SENTINEL MONITOR industrial-ai-master 10.2.1.20 6379 2
redis-cli -h 10.2.1.102 -p 26379 SENTINEL MONITOR industrial-ai-master 10.2.1.20 6379 2
redis-cli -h 10.2.1.103 -p 26379 SENTINEL MONITOR industrial-ai-master 10.2.1.20 6379 2

# 3. 停止 RedisShake 同步
systemctl stop redis-shake

# 4. 更新应用连接
# 由 GSLB 处理

echo "上海 Redis 已提升为独立 Master"
```

### 7.3 回切流程

```bash
#!/bin/bash
# switch-back-to-beijing.sh

# 1. 确认北京 Redis 已恢复
redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD PING

# 2. 将北京设为上海的从库
redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD REPLICAOF 10.2.1.20 6379

# 3. 等待数据同步完成
# 监控复制偏移量
while true; do
    SYNC=$(redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD INFO replication | grep master_sync_in_progress | cut -d: -f2)
    if [ "$SYNC" = "0" ]; then
        break
    fi
    sleep 1
done

# 4. 提升北京为主库
redis-cli -h 10.1.1.20 -p 6379 -a $PASSWORD REPLICAOF NO ONE

# 5. 将上海设为北京的从库
redis-cli -h 10.2.1.20 -p 6379 -a $PASSWORD REPLICAOF 10.1.1.20 6379

# 6. 重启 RedisShake 同步
systemctl start redis-shake

echo "已回切到北京主数据中心"
```

## 8. 最佳实践

### 8.1 网络优化

```conf
# /etc/sysctl.conf 添加

# TCP 优化
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_keepalive_time = 60
net.ipv4.tcp_keepalive_intvl = 10
net.ipv4.tcp_keepalive_probes = 6
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30

# 应用
sysctl -p
```

### 8.2 连接池配置

```go
// Go 应用连接池配置
&redis.Options{
    Addr:         "10.1.1.20:6379",
    Password:     "your-secure-password",
    DB:           0,
    PoolSize:     100,        // 连接池大小
    MinIdleConns: 20,         // 最小空闲连接
    MaxRetries:   3,          // 最大重试次数
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,
}
```

### 8.3 健康检查

```bash
# 健康检查脚本
#!/bin/bash

REDIS_HOST="${REDIS_HOST:-10.1.1.20}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASS}"

# 检查连接
if ! redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning PING > /dev/null 2>&1; then
    echo "Redis connection failed"
    exit 1
fi

# 检查角色
ROLE=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning INFO replication | grep role | cut -d: -f2 | tr -d '\r')
echo "Redis role: $ROLE"

# 检查复制状态
if [ "$ROLE" = "slave" ]; then
    LINK=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASS --no-auth-warning INFO replication | grep master_link_status | cut -d: -f2 | tr -d '\r')
    if [ "$LINK" != "up" ]; then
        echo "Replication link down"
        exit 1
    fi
fi

echo "Redis health check passed"
exit 0
```

## 9. 运维手册

### 9.1 日常运维

```bash
# 查看复制状态
redis-cli -h <host> -p <port> -a <password> INFO replication

# 查看延迟
redis-cli -h <host> -p <port> -a <password> --latency

# 查看慢日志
redis-cli -h <host> -p <port> -a <password> SLOWLOG GET 10

# 监控实时命令
redis-cli -h <host> -p <port> -a <password> MONITOR
```

### 9.2 故障排查

```bash
# 检查 Sentinel 状态
redis-cli -h <sentinel_host> -p 26379 SENTINEL MASTER industrial-ai-master
redis-cli -h <sentinel_host> -p 26379 SENTINEL REPLICAS industrial-ai-master

# 手动故障切换
redis-cli -h <sentinel_host> -p 26379 SENTINEL FAILOVER industrial-ai-master

# 检查网络延迟
redis-cli -h <remote_host> -p <port> -a <password> --latency-history
```

### 9.3 备份与恢复

```bash
# 备份
redis-cli -h <host> -p <port> -a <password> BGSAVE
# RDB 文件位置: /var/lib/redis/dump.rdb

# 恢复
# 1. 停止 Redis
systemctl stop redis

# 2. 复制 RDB 文件
cp /backup/dump.rdb /var/lib/redis/dump.rdb
chown redis:redis /var/lib/redis/dump.rdb

# 3. 启动 Redis
systemctl start redis
```

## 10. 附录

### 10.1 配置文件清单

- `redis.conf` - Redis 主配置
- `sentinel.conf` - Sentinel 配置
- `shake.toml` - RedisShake 同步配置
- `redis-verify.sh` - 数据校验脚本
- `redis-monitor.sh` - 监控脚本

### 10.2 参考文档

- [Redis Replication](https://redis.io/docs/management/replication/)
- [Redis Sentinel](https://redis.io/docs/management/sentinel/)
- [RedisShake](https://github.com/alibaba/RedisShake)
- [Redis CRDT](https://github.com/redis/redis-crdt)