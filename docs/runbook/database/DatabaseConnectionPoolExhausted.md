# DatabaseConnectionPoolExhausted

> **告警名称**: DatabaseConnectionPoolExhausted  
> **严重度**: Critical  
> **类别**: Database  
> **阈值**: 活动连接数 > 100  
> **持续时间**: 2 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
db_connections_active > 100
```

**触发条件**:
- PostgreSQL 活动连接数超过 100
- 连接池已耗尽，新请求无法获取连接
- 服务可能完全不可用

---

## 🔍 诊断步骤

### 1️⃣ 检查当前连接数

```bash
# PostgreSQL 活动连接
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) as total, 
       count(*) FILTER (WHERE state = 'active') as active,
       count(*) FILTER (WHERE state = 'idle') as idle,
       count(*) FILTER (WHERE state = 'idle in transaction') as idle_in_transaction
FROM pg_stat_activity;
"

# 查看连接来源
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT client_addr, usename, application_name, state, count(*)
FROM pg_stat_activity
GROUP BY client_addr, usename, application_name, state
ORDER BY count(*) DESC;
"
```

### 2️⃣ 查看长时间事务

```bash
# 查看超过 30 秒的事务
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT pid, usename, state, query_start, now() - query_start as duration, query
FROM pg_stat_activity
WHERE state IN ('active', 'idle in transaction')
AND now() - query_start > interval '30 seconds'
ORDER BY duration DESC;
"
```

### 3️⃣ 检查是否有锁等待

```bash
# 查看锁等待
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT blocked_locks.pid AS blocked_pid,
       blocked_locks.query AS blocked_query,
       blocking_locks.pid AS blocking_pid,
       blocking_locks.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_locks blocking_locks
  ON blocking_locks.locktype = blocked_locks.locktype
  AND blocking_locks.relation = blocked_locks.relation
  AND blocking_locks.granted
WHERE NOT blocked_locks.granted;
"
```

### 4️⃣ 检查后端配置

```bash
# 查看后端最大连接配置
docker logs industrial-ai-backend --tail 500 | grep -i "max.*connection"

# 检查 PostgreSQL 最大连接
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SHOW max_connections;"
```

---

## 🛠️ 修复方案

### 情况 A: 长时间事务阻塞

```bash
# 1. 找出阻塞的 PID
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT pid, usename, query_start, now() - query_start as duration, query
FROM pg_stat_activity
WHERE state = 'idle in transaction' AND now() - query_start > interval '60 seconds'
ORDER BY duration DESC;
"

# 2. 终止长时间事务
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle in transaction' 
AND now() - query_start > interval '60 seconds';
"

# 3. 验证连接数下降
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT count(*) FROM pg_stat_activity;"
```

### 情况 B: 锁等待导致连接堆积

```bash
# 1. 找出阻塞源头
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT blocking_pid, blocking_query
FROM pg_locks blocked
JOIN pg_locks blocking ON blocking.locktype = blocked.locktype
WHERE NOT blocked.granted AND blocking.granted;
"

# 2. 终止阻塞事务
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT pg_terminate_backend(<blocking_pid>);"

# 3. 验证锁解除
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT count(*) FROM pg_locks WHERE NOT granted;"
```

### 情况 C: 后端连接池配置错误

```bash
# 1. 重启后端释放连接
docker-compose restart backend
kubectl rollout restart deployment/backend -n industrial-ai

# 2. 修改连接池配置
# 在 ServerConfig 中调整:
# - DatabaseMaxOpenConns: 50 (默认)
# - DatabaseMaxIdleConns: 10
# - DatabaseConnMaxLifetime: 30m

# 3. 增加 PostgreSQL 最大连接 (如果必要)
docker exec -it industrial-ai-postgres psql -U industrial_user -c "ALTER SYSTEM SET max_connections = 200;"
docker-compose restart postgres
```

### 情况 D: 流量激增

```bash
# 1. 扩容后端实例
kubectl scale deployment/backend --replicas=5 -n industrial-ai

# 2. 配置连接池分片
# 每个实例的连接池大小需要小于 PostgreSQL max_connections / replicas

# 3. 临时增加 PostgreSQL 连接限制
docker exec -it industrial-ai-postgres psql -U industrial_user -c "ALTER SYSTEM SET max_connections = 300;"
docker-compose restart postgres
```

---

## ✅ 验证恢复

```bash
# 1. 验证连接数正常 (< 50)
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT count(*) FROM pg_stat_activity;"

# 2. 验证无长时间事务
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM pg_stat_activity 
WHERE state = 'idle in transaction' AND now() - query_start > interval '60 seconds';
"

# 3. 验证无锁等待
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT count(*) FROM pg_locks WHERE NOT granted;"

# 4. 验证服务正常
curl http://backend:8080/health
```

---

## 📝 事后复盘

### 必须检查
- 是否有慢查询导致事务长时间运行
- 是否有代码 bug 导致连接未释放
- 连接池配置是否合理

### 长期优化
- 添加事务超时机制
- 配置连接池健康检查
- 定期清理 idle in transaction 连接
- 添加连接数告警阈值 (Warning: 50)

---

## 🔗 相关 Runbook

- [HighDatabaseQueryTime](../database/HighDatabaseQueryTime.md) - 查询慢导致连接堆积
- [HighErrorRate](../http/HighErrorRate.md) - 连接池耗尽导致 500 错误
- [ServiceDown](../system/ServiceDown.md) - 数据库完全不可用

---

**最后更新**: 2026-05-13  
**审核人**: DBA Team Lead