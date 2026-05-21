# 数据库性能优化指南

> **Industrial AI Platform PostgreSQL 性能优化最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 性能优化概述

Phase 4 P1 性能优化目标：

| 指标 | 当前状态 | 目标值 | 提升幅度 |
|------|---------|--------|---------|
| **查询 P95 响应时间** | ~300ms | <100ms | 3x |
| **数据库连接池利用率** | ~80% | <70% | 更稳定 |
| **慢查询数量** | ~50/天 | <5/天 | 10x |
| **索引命中率** | ~60% | >90% | 30% |
| **缓存命中率** | ~60% | >85% | 25% |

---

## 📊 索引优化

### 现有索引审计

**当前已创建索引：**

| 表 | 索引名 | 字段 | 状态 |
|----|--------|------|------|
| users | idx_users_tenant_id | tenant_id | ✅ |
| devices | idx_devices_tenant_id | tenant_id | ✅ |
| device_telemetry | idx_telemetry_tenant_id | tenant_id | ✅ |
| alert_rules | idx_alert_rules_tenant_id | tenant_id | ✅ |
| alerts | idx_alerts_tenant_id | tenant_id | ✅ |
| work_orders | idx_work_orders_tenant_id | tenant_id | ✅ |
| roles | idx_roles_tenant_id | tenant_id | ✅ |
| roles | idx_roles_name | name | ✅ |
| user_roles | idx_user_roles_user_id | user_id | ✅ |
| user_roles | idx_user_roles_role_id | role_id | ✅ |

### 需新增索引

| 表 | 索引名 | 字段 | 原因 |
|----|--------|------|------|
| devices | idx_devices_tenant_status | tenant_id, status | 高频查询组合 |
| devices | idx_devices_tenant_created | tenant_id, created_at | 列表排序 |
| device_telemetry | idx_telemetry_device_time | device_id, timestamp | 遥测查询 |
| device_telemetry | idx_telemetry_tenant_device_time | tenant_id, device_id, timestamp | 多租户查询 |
| alerts | idx_alerts_status_created | status, created_at | 告警列表 |
| alerts | idx_alerts_tenant_device | tenant_id, device_id | 设备告警查询 |
| users | idx_users_tenant_username | tenant_id, username | 登录查询 |

### 索引创建 SQL

```sql
-- 性能优化索引迁移
-- migration: 000004_add_performance_indexes.up.sql

-- 设备表索引优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_status 
ON devices (tenant_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_created 
ON devices (tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_last_heartbeat 
ON devices (last_heartbeat DESC);

-- 遥测数据索引优化 (高频查询)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_device_time 
ON device_telemetry (device_id, timestamp DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_tenant_device_time 
ON device_telemetry (tenant_id, device_id, timestamp DESC);

-- 告警表索引优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_status_created 
ON alerts (status, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_tenant_device 
ON alerts (tenant_id, device_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_severity_status 
ON alerts (severity, status);

-- 用户表索引优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_username 
ON users (tenant_id, username);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_last_login 
ON users (last_login DESC);

-- 告警规则索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_enabled 
ON alert_rules (enabled);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_tenant_enabled 
ON alert_rules (tenant_id, enabled);

-- 工单索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_status_created 
ON work_orders (status, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_tenant_status 
ON work_orders (tenant_id, status);
```

---

## 🔧 PostgreSQL 配置优化

### 连接池配置

```yaml
# postgresql.conf
max_connections = 200             # 最大连接数
shared_buffers = 256MB            # 共享内存缓冲池
work_mem = 16MB                   # 单个查询工作内存
maintenance_work_mem = 256MB      # 维护操作内存
effective_cache_size = 768MB      # 估计可用缓存
random_page_cost = 1.1            # SSD 低成本
effective_io_concurrency = 200    # SSD 并发 I/O
```

### 查询优化配置

```yaml
# postgresql.conf
# 查询计划优化
jit = off                         # 禁用 JIT (小查询不适合)
track_activities = on             # 活动追踪
track_counts = on                 # 统计追踪
track_functions = all             # 函数追踪
track_io_timing = on              # I/O 时间追踪

# 慢查询日志
log_min_duration_statement = 100  # 100ms 以上记录慢查询
log_statement = 'ddl'             # DDL 操作日志
log_line_prefix = '%t [%p] [%u@%d] ' # 日志前缀
```

---

## 📈 pg_stat_statements 配置

### 启用扩展

```sql
-- 启用 pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 配置 postgresql.conf
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.max = 10000
pg_stat_statements.track = all
```

### 查询分析 SQL

```sql
-- 查看最慢查询 (Top 20)
SELECT 
    query,
    calls,
    total_time / calls as avg_time_ms,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
ORDER BY avg_time_ms DESC
LIMIT 20;

-- 查看最频繁查询
SELECT 
    query,
    calls,
    total_time / 1000 as total_time_s,
    rows
FROM pg_stat_statements
ORDER BY calls DESC
LIMIT 20;

-- 查看缓存命中率低的查询
SELECT 
    query,
    calls,
    shared_blks_read,
    shared_blks_hit,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
WHERE shared_blks_read > 100
ORDER BY hit_percent ASC
LIMIT 20;
```

---

## 🔄 连接池参数 (应用端)

### Go 连接池配置

```go
// backend/pkg/database/config.go
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
    return &ConnectionPoolConfig{
        MaxOpenConns:    50,              // 最大打开连接
        MaxIdleConns:    10,              // 最大空闲连接
        ConnMaxLifetime: 1 * time.Hour,   // 连接最大生命周期
        ConnMaxIdleTime: 10 * time.Minute,// 空闲超时
    }
}
```

### 参数说明

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| **MaxOpenConns** | 50 | 防止连接耗尽，根据并发请求调整 |
| **MaxIdleConns** | 10 | 保持一定空闲连接，减少新建连接开销 |
| **ConnMaxLifetime** | 1h | 防止长时间连接，定期刷新 |
| **ConnMaxIdleTime** | 10m | 清理不活跃连接，释放资源 |

---

## 📊 性能监控查询

### 实时监控 SQL

```sql
-- 当前活跃查询
SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state,
    usename,
    application_name
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
AND state != 'idle'
ORDER BY duration DESC;

-- 连接池状态
SELECT 
    state,
    count(*)
FROM pg_stat_activity
GROUP BY state;

-- 表大小统计
SELECT 
    table_name,
    pg_size_pretty(pg_total_relation_size(table_name::text)) as size,
    pg_stat_get_tuples_returned(table_name::regclass) as tuples_returned
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY pg_total_relation_size(table_name::text) DESC
LIMIT 10;

-- 索引使用统计
SELECT 
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC
LIMIT 20;

-- 未使用索引 (考虑删除)
SELECT 
    indexrelname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) as wasted_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
AND indexrelname NOT LIKE '%_pkey'
ORDER BY pg_relation_size(indexrelid) DESC;
```

---

## ⚡ 查询优化建议

### 常见优化模式

| 模式 | 问题 | 解决方案 |
|------|------|---------|
| **全表扫描** | WHERE 条件无索引 | 添加合适索引 |
| **排序慢** | ORDER BY 无索引 | 添加排序字段索引 |
| **JOIN 慢** | 关联字段无索引 | 添加外键索引 |
| **LIMIT + OFFSET** | 大 OFFSET 性差 | 使用 keyset pagination |
| **SELECT *** | 返回过多列 | 只返回需要的列 |
| **COUNT(*)** | 大表计数慢 | 使用估算或缓存 |

### 分页优化

```sql
-- ❌ 低效分页 (OFFSET)
SELECT * FROM device_telemetry
ORDER BY timestamp DESC
LIMIT 20 OFFSET 10000;

-- ✅ 高效分页 (Keyset)
SELECT * FROM device_telemetry
WHERE timestamp < '2024-01-01 00:00:00'
ORDER BY timestamp DESC
LIMIT 20;
```

---

## ✅ 优化验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **索引覆盖率** | >90% | `pg_stat_user_indexes` |
| **查询 P95** | <100ms | Prometheus histogram |
| **慢查询** | <5/天 | PostgreSQL slow log |
| **缓存命中率** | >85% | pg_stat_statements |
| **连接池健康** | 无耗尽 | Prometheus metrics |

---

## 🔧 优化脚本

```bash
# 运行索引优化迁移
psql -h postgres -U industrial_app -d industrial_ai \
    -f migrations/000004_add_performance_indexes.up.sql

# 启用 pg_stat_statements
psql -h postgres -U postgres -d industrial_ai \
    -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"

# 查看慢查询
psql -h postgres -U industrial_app -d industrial_ai \
    -c "SELECT query, calls, total_time/calls as avg_ms FROM pg_stat_statements ORDER BY avg_ms DESC LIMIT 20;"
```

---

**最后更新**: 2026-05-13  
**审核人**: DBA Team