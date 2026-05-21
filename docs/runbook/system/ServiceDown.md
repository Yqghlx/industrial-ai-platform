# ServiceDown

> **告警名称**: ServiceDown  
> **严重度**: Critical  
> **类别**: System  
> **阈值**: 服务 UP 状态 = 0  
> **持续时间**: 5 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
up{job="backend"} == 0
```

**触发条件**:
- Backend 服务持续宕机超过 5 分钟
- Prometheus 无法抓取 metrics
- 服务完全不可用

---

## 🚨 紧急响应流程

> **⚠️ 这是最高严重度告警，必须立即处理！**

### Step 1: 确认服务状态 (立即)

```bash
# Docker
docker ps | grep industrial-ai-backend
docker inspect industrial-ai-backend --format '{{.State.Status}}'

# Kubernetes
kubectl get pods -l app=backend -n industrial-ai
kubectl describe pod <pod-name> -n industrial-ai
```

### Step 2: 查看崩溃原因 (1 分钟内)

```bash
# 查看最后日志
docker logs industrial-ai-backend --tail 100

# 如果容器已退出，查看退出前日志
docker logs industrial-ai-backend --previous --tail 200

# Kubernetes 查看事件
kubectl get events -n industrial-ai --sort-by='.lastTimestamp' | grep backend | tail 20

# 查看退出状态
kubectl describe pod <pod-name> -n industrial-ai | grep -A 5 "Last State"
```

### Step 3: 立即重启服务 (2 分钟内)

```bash
# Docker
docker-compose restart backend
docker-compose logs -f backend --tail 50

# Kubernetes
kubectl rollout restart deployment/backend -n industrial-ai
kubectl rollout status deployment/backend -n industrial-ai --timeout=60s
```

### Step 4: 验证恢复 (5 分钟内)

```bash
# 健康检查
curl -f http://backend:8080/health || echo "Service still down!"

# Prometheus UP 状态
curl 'http://prometheus:9090/api/v1/query?query=up{job="backend"}'

# 查看启动日志
docker logs industrial-ai-backend --tail 100 | grep -i "started|listening|ready"
```

---

## 🔍 深入诊断 (重启后)

### 检查崩溃原因

```bash
# 1. OOM (内存溢出)
kubectl describe pod <pod-name> -n industrial-ai | grep -i "OOMKilled"
docker inspect industrial-ai-backend | grep -i "OOMKilled"

# 2. Panic/Crash
docker logs industrial-ai-backend --previous | grep -E "(panic|fatal|FATAL|PANIC)" | tail 20

# 3. 配置错误
docker logs industrial-ai-backend --previous | grep -E "(config|env|invalid|missing)" | tail 20

# 4. 依赖服务不可用
docker logs industrial-ai-backend --previous | grep -E "(postgres|redis|connection.*refused)" | tail 20
```

### 检查资源限制

```bash
# 内存限制
kubectl describe pod <pod-name> -n industrial-ai | grep -A 5 "Limits"
docker inspect industrial-ai-backend | grep -i "Memory"

# CPU 限制
kubectl top pods -n industrial-ai

# 磁盘空间
df -h /var/lib/docker
kubectl exec <pod-name> -- df -h
```

---

## 🛠️ 根本原因修复

### 情况 A: OOM (内存溢出)

```bash
# 1. 增加内存限制
kubectl patch deployment backend -n industrial-ai -p '{"spec":{"template":{"spec":{"containers":[{"name":"backend","resources":{"limits":{"memory":"1Gi"}}}]}}}}'

# 2. 或在 docker-compose.yml 中调整
# memory: 1G

# 3. 分析内存使用
kubectl exec <pod-name> -- cat /proc/self/status | grep -i "VmRSS"
```

### 情况 B: 代码 Panic

```bash
# 1. 收集 panic 日志
docker logs industrial-ai-backend --previous > /tmp/panic.log

# 2. 分析 panic 堆栈
grep -A 30 "panic:" /tmp/panic.log

# 3. 通知开发团队
# 创建 Jira ticket: "Backend Panic: <panic message>"
# 附上完整 panic 日志

# 4. 临时回滚到稳定版本
kubectl rollout undo deployment/backend -n industrial-ai
```

### 情况 C: 配置错误

```bash
# 1. 检查环境变量
kubectl describe pod <pod-name> -n industrial-ai | grep -A 20 "Environment"

# 2. 检查 ConfigMap/Secret
kubectl get configmap -n industrial-ai
kubectl get secret -n industrial-ai

# 3. 修复缺失配置
kubectl create secret generic industrial-ai-secrets --from-literal=JWT_SECRET=<secret> -n industrial-ai
```

### 情况 D: 依赖服务不可用

```bash
# 1. 检查 PostgreSQL
docker ps | grep postgres
docker logs industrial-ai-postgres --tail 100

# 2. 检查 Redis
docker ps | grep redis
docker logs industrial-ai-redis --tail 100

# 3. 启动依赖服务
docker-compose up -d postgres redis

# 4. 等待依赖就绪
docker exec industrial-ai-postgres pg_isready
docker exec industrial-ai-redis redis-cli ping

# 5. 重启后端
docker-compose restart backend
```

---

## ✅ 完整验证清单

```bash
# 1. 服务 UP 状态
curl 'http://prometheus:9090/api/v1/query?query=up{job="backend"}'
# 期望: value = 1

# 2. 健康检查
curl http://backend:8080/health
# 期望: {"status": "healthy"}

# 3. Metrics 可用
curl http://backend:8080/metrics | head -20
# 期望: Prometheus metrics 输出

# 4. API 功能
curl http://backend:8080/api/v1/devices -H "Authorization: Bearer <token>"
# 期望: 正常响应

# 5. WebSocket 连接
wscat -c ws://backend:8080/ws
# 期望: 连接成功

# 6. 数据库连接
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT count(*) FROM devices;"
# 期望: 正常返回数据

# 7. 日志无错误
docker logs industrial-ai-backend --tail 50 | grep -i error
# 期望: 无新错误
```

---

## 📞 紧急联系人

| 角色 | 负责事项 | 联系方式 |
|------|---------|---------|
| **值班 Ops** | 服务重启 | PagerDuty |
| **后端开发** | Panic 分析 | @dev-team |
| **DBA** | 数据库问题 | @dba-team |
| **管理层** | 持续宕机 > 30 分钟 | @manager |

---

## 📝 事后复盘 (必须)

### Critical 告警必须创建复盘报告:

1. **时间线**: 告警触发 → 确认 → 重启 → 恢复
2. **影响**: 用户影响时长、业务损失评估
3. **根本原因**: 具体原因分析
4. **修复方案**: 采取的措施
5. **预防措施**: 如何避免再次发生

### 报告模板:
```markdown
## ServiceDown 事后复盘

**发生时间**: YYYY-MM-DD HH:MM
**持续时间**: XX 分钟
**影响范围**: 所有用户

**根本原因**: 
- [OOM/Panic/Config/Dependency]

**处理步骤**:
1. 收到告警
2. 确认服务宕机
3. 分析日志发现 XXX
4. 执行 XXX 操作
5. 服务恢复

**预防措施**:
- [具体改进建议]

**负责人**: XXX
```

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead