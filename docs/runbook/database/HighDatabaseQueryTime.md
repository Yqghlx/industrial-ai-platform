# HighDatabaseQueryTime

> **告警名称**: HighDatabaseQueryTime  
> **严重度**: Warning  
> **类别**: Database  
> **阈值**: P95 查询时间 > 500ms  
> **持续时间**: 10 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le)) > 0.5
```

**触发条件**:
- 95% 的数据库查询超过 500ms
- 查询性能严重下降

---

## 🔍 诊断步骤

### 1️⃣ 分析慢查询

```bash
# PostgreSQL 慢查询统计
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT query, calls, total_time/calls as avg_time_ms, rows
FROM pg_stat_statements
ORDER BY avg_time_ms DESC
LIMIT 20;
"

# 启用 pg_stat_statements (如果未启用)
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
"
```

### 2️⃣ 检查查询类型分布

```bash
# 查看各操作的查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le,operation))'
```

### 3️⃣ 检查索引状态

```bash
# 查看缺失索引的表
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT schemaname, relname, attname, n_distinct, correlation
FROM pg_stats
WHERE n_distinct > 100 AND correlation < 0.1
ORDER BY n_distinct DESC;
"

# 查看索引使用率
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT indexrelname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC
LIMIT 20;
"
```

### 4️⃣ 检查表大小

```bash
# 查看表大小
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT table_name, 
       pg_size_pretty(pg_total_relation_size(table_name::text)) as size,
       (SELECT count(*) FROM information_schema.columns WHERE table_name = t.table_name) as columns
FROM information_schema.tables t
WHERE table_schema = 'public'
ORDER BY pg_total_relation_size(table_name::text) DESC
LIMIT 10;
"
```

---

## 🛠️ 修复方案

### 情况 A: 缺失索引

```bash
# 1. 分析慢查询需要什么索引
# 例如: tenant_id + status 组合查询慢

# 2. 创建索引
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
CREATE INDEX CONCURRENTLY idx_devices_tenant_status_created 
ON devices (tenant_id, status, created_at);
"

# 3. 验证索引生效
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
EXPLAIN ANALYZE SELECT * FROM devices WHERE tenant_id = 'xxx' AND status = 'online';
"
```

### 情况 B: 查询返回大量数据

```bash
# 1. 检查是否有全表扫描
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT query, rows, calls
FROM pg_stat_statements
WHERE rows > 1000
ORDER BY rows DESC LIMIT 10;
"

# 2. 添加查询限制
# 修改应用代码，添加 LIMIT 和分页

# 3. 配置查询超时
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
ALTER SYSTEM SET statement_timeout = '30s';
"
```

### 情况 C: 表数据量过大

```bash
# 1. 检查遥测数据量
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM telemetry_data WHERE created_at < now() - interval '30 days';
"

# 2. 清理旧数据
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
DELETE FROM telemetry_data WHERE created_at < now() - interval '90 days';
"

# 3. 配置分区表 (长期)
# 按时间分区遥测数据表
```

### 情况 D: 数据库资源不足

```bash
# 1. 检查数据库资源
docker stats industrial-ai-postgres --no-stream

# 2. 增加数据库资源
# 调整 shared_buffers 和 work_mem
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET work_mem = '64MB';
"

# 3. 重启数据库生效
docker-compose restart postgres
```

---

## ✅ 验证恢复

```bash
# 1. 验证查询时间下降
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le))'

# 2. 验证索引使用增加
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT indexrelname, idx_scan FROM pg_stat_user_indexes WHERE indexrelname LIKE 'idx_devices%';
"

# 3. 手动查询测试
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
EXPLAIN ANALYZE SELECT * FROM devices WHERE tenant_id = 'test-tenant' LIMIT 10;
"
```

---

**最后更新**: 2026-05-13  
**审核人**: DBA Team Lead