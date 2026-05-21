# PostgreSQL 高可用指南

> **Industrial AI Platform PostgreSQL 高可用最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 PostgreSQL 高可用概述

Phase 4 P1 高可用数据库目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **数据库可用性** | 单节点 | 主从复制 |
| **故障恢复时间** | 手动恢复 | <30s 自动切换 |
| **数据一致性** | 无保障 | WAL 流复制 |
| **故障检测** | 无 | Patroni 监控 |

---

## 🔄 PostgreSQL 高可用架构

### 主从复制架构

```
┌─────────────────────────────────────────┐
│  Patroni (HA 管理器)                     │
│  - 故障检测                              │
│  - 自动故障转移                          │
│  - 健康检查                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  PostgreSQL Primary (主节点)             │  ← 写入节点
│  - 接收所有写请求                        │
│  - WAL 流复制到副本                      │
│  - 自动故障转移候选                      │
└─────────────────────────────────────────┘
          ↓ (WAL Stream)
┌─────────────────────────────────────────┐
│  PostgreSQL Replica (副本节点)           │  ← 读节点
│  - 流复制同步主节点数据                  │
│  - 接收读请求                            │
│  - 故障转移后可升级为主节点              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  pgBouncer (连接池)                      │  ← 连接池管理
│  - 连接池管理                            │
│  - 读/写分离                             │
│  - 主节点故障时自动切换                  │
└─────────────────────────────────────────┘
```

---

## 🔧 PostgreSQL 主从复制配置

### 主节点配置 (postgresql.conf)

```conf
# postgresql.conf (主节点)

# === 基础配置 ===
listen_addresses = '*'
port = 5432
max_connections = 200

# === WAL 配置 ===
wal_level = replica             # 启用 WAL 复制
max_wal_senders = 10            # 最大 WAL 发送进程
wal_keep_size = 1GB             # 保留 WAL 大小
max_replication_slots = 10      # 最大复制槽

# === 复制配置 ===
synchronous_commit = on         # 同步提交
wal_sender_delay = 200ms        # WAL 发送延迟

# === 性能优化 ===
shared_buffers = 1GB
work_mem = 64MB
maintenance_work_mem = 256MB
checkpoint_completion_target = 0.9
wal_buffers = 64MB

# === 监控配置 ===
track_activities = on
track_counts = on
track_io_timing = on
track_functions = all
```

### 副本节点配置 (postgresql.conf)

```conf
# postgresql.conf (副本节点)

# === 基础配置 ===
listen_addresses = '*'
port = 5432
max_connections = 200

# === 复制配置 ===
hot_standby = on                # 启用热备份
hot_standby_feedback = on       # 反馈机制

# === 连接配置 ===
primary_conninfo = 'host=postgres-primary port=5432 user=replicator password=<password>'

# === 复制槽 ===
# primary_slot_name = 'replica_1'
```

### 复制用户配置

```sql
-- 创建复制用户
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD '<password>';

-- 授权复制用户
GRANT pg_read_all_data TO replicator;
```

---

## 🛡️ Patroni 高可用管理

### Patroni 配置

```yaml
# patroni.yml (主节点)

scope: industrial-ai-cluster
name: postgres-primary

restapi:
  listen: 0.0.0.0:8008
  connect: postgres-primary:8008

postgresql:
  listen: 0.0.0.0:5432
  connect: postgres-primary:5432
  data_dir: /var/lib/postgresql/data
  
  authentication:
    superuser:
      username: postgres
      password: <password>
    replication:
      username: replicator
      password: <password>
  
  parameters:
    wal_level: replica
    max_wal_senders: 10
    wal_keep_size: 1GB
    max_replication_slots: 10
    hot_standby: on
    synchronous_commit: on
  
  # 复制配置
  create_replica_methods:
    - pg_basebackup
  
  pg_basebackup:
    command: pg_basebackup -h postgres-primary -U replicator -D /var/lib/postgresql/data -P -R

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
    
    postgresql:
      use_pg_rewind: true
      
      parameters:
        wal_level: replica
        max_wal_senders: 10
        wal_keep_size: 1GB
        max_replication_slots: 10
        hot_standby: on
        synchronous_commit: on

# DCS 配置 (使用 etcd 或 Consul)
etcd:
  host: etcd:2379
```

---

## 🔄 故障转移机制

### Patroni 故障转移流程

```
┌─────────────────────────────────────────┐
│  1. 主节点故障检测                        │
│  - Patroni 健康检查失败                   │
│  - 连续 N 次失败后确认故障                │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. 选举新主节点                          │
│  - 副本节点竞选                          │
│  - 最小延迟副本优先                      │
│  - DCS 锁定选举结果                      │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 升级新主节点                          │
│  - 副本节点升级为主节点                  │
│  - 关闭读模式                            │
│  - 开始接收写请求                        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 更新路由                              │
│  - pgBouncer 更新主节点地址              │
│  - 应用重新连接                          │
│  - 开始处理请求                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 修复故障节点                          │
│  - 重启故障节点                          │
│  - 同步 WAL 数据                         │
│  - 作为副本重新加入集群                  │
└─────────────────────────────────────────┘
```

---

## 📊 数据库监控

### 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|---------|
| **replication_lag** | 复制延迟 | >10MB 告警 |
| **primary_status** | 主节点状态 | down 告警 |
| **replica_count** | 副本节点数 | <1 告警 |
| **connection_pool_usage** | 连接池使用率 | >80% 告警 |
| **wal_files_count** | WAL 文件数 | >100 告警 |

### PostgreSQL 监控视图

```sql
-- 复制状态监控视图
SELECT * FROM pg_stat_replication;

-- 复制延迟监控
SELECT 
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replication_lag_bytes
FROM pg_stat_replication;

-- 主节点状态检查
SELECT pg_is_in_recovery();
```

---

## 🔄 pgBouncer 连接池

### pgBouncer 配置

```ini
# pgbouncer.ini

[databases]
industrial_ai = host=postgres-primary port=5432 dbname=industrial_ai
industrial_ai_read = host=postgres-replica port=5432 dbname=industrial_ai

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 50
min_pool_size = 10
reserve_pool_size = 10
reserve_pool_timeout = 5

admin_users = postgres
stats_users = postgres

logfile = /var/log/pgbouncer/pgbouncer.log
pidfile = /var/run/pgbouncer/pgbouncer.pid
```

---

## ✅ PostgreSQL 高可用验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **主从复制** | 数据同步 | pg_stat_replication |
| **故障转移** | 自动切换 | Patroni 状态 |
| **连接池** | 正常工作 | pgBouncer 状态 |
| **读/写分离** | 正常路由 | 应用测试 |
| **监控指标** | 正常采集 | Prometheus |
| **告警规则** | 正常触发 | Alertmanager |

---

**最后更新**: 2026-05-13  
**审核人**: DBA Team