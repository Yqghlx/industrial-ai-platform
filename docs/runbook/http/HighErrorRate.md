# HighErrorRate

> **告警名称**: HighErrorRate  
> **严重度**: Critical  
> **类别**: HTTP  
> **阈值**: HTTP 5xx 错误率 > 5%  
> **持续时间**: 5 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) /
sum(rate(http_requests_total[5m])) > 0.05
```

**触发条件**:
- 过去 5 分钟内，HTTP 5xx 错误率超过 5%
- 可能原因: 服务异常、数据库连接失败、外部 API 调用失败

---

## 🔍 诊断步骤

### 1️⃣ 确认错误类型分布

```bash
# 查看各状态码分布
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total[5m]))by(status)'

# Grafana 查看 HTTP Status Codes 面板
# 关注 500, 502, 503, 504 的具体比例
```

### 2️⃣ 检查后端服务日志

```bash
# Docker
docker logs industrial-ai-backend --tail 500 | grep -E "(error|Error|ERROR|panic|fatal)"

# Kubernetes
kubectl logs -l app=backend -n industrial-ai --tail 500 | grep -i error

# 查看最近的 panic 或 fatal
docker logs industrial-ai-backend --since 10m | grep -E "(panic|fatal|PANIC|FATAL)"
```

### 3️⃣ 检查依赖服务状态

```bash
# 检查 PostgreSQL
docker exec industrial-ai-postgres pg_isready
kubectl exec -it postgres-pod -- pg_isready

# 检查 Redis
docker exec industrial-ai-redis redis-cli ping
kubectl exec -it redis-pod -- redis-cli ping

# 检查数据库连接
docker exec -it industrial-ai-postgres psql -U industrial_user -c "SELECT 1;"
```

### 4️⃣ 检查资源使用

```bash
# CPU/内存
docker stats industrial-ai-backend --no-stream
kubectl top pods -n industrial-ai

# 检查是否 OOM
docker inspect industrial-ai-backend | grep -i oom
kubectl describe pod <pod-name> -n industrial-ai | grep -i oom
```

### 5️⃣ 查看最近变更

```bash
# 检查最近的部署
kubectl rollout history deployment/backend -n industrial-ai

# 检查最近的配置变更
git log --since="1 day ago" --oneline
```

---

## 🛠️ 修复方案

### 情况 A: 服务 Panic/Crash

```bash
# 1. 重启服务
docker-compose restart backend
kubectl rollout restart deployment/backend -n industrial-ai

# 2. 检查重启后状态
kubectl rollout status deployment/backend -n industrial-ai

# 3. 验证错误率下降
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total{status=~"5.."}[5m]))'
```

### 情况 B: 数据库连接失败

```bash
# 1. 检查 PostgreSQL 状态
docker logs industrial-ai-postgres --tail 100

# 2. 检查连接池配置
# 查看 DATABASE_URL 和连接池参数

# 3. 如果数据库重启
docker-compose restart postgres
kubectl rollout restart deployment/postgres -n industrial-ai

# 4. 等待数据库恢复
docker exec industrial-ai-postgres pg_isready

# 5. 重启后端以重建连接池
docker-compose restart backend
```

### 情况 C: Redis 不可用

```bash
# 1. 检查 Redis 状态
docker logs industrial-ai-redis --tail 100

# 2. 检查内存是否溢出
docker exec industrial-ai-redis redis-cli INFO memory

# 3. 重启 Redis
docker-compose restart redis

# 4. 清理缓存 (如果必要)
docker exec industrial-ai-redis redis-cli FLUSHALL
```

### 情况 D: 外部 API 调用失败

```bash
# 1. 检查 AI API 配置
echo $GLM_API_URL
echo $GLM_API_KEY

# 2. 测试 API 连通性
curl -I https://open.bigmodel.cn/api/paas/v4

# 3. 检查 API 返回的错误信息
docker logs industrial-ai-backend --tail 200 | grep -i "api error"

# 4. 如果是 API 服务问题
# - 配置备用 API endpoint
# - 或临时降级 AI 功能
```

### 情况 E: 请求量激增导致过载

```bash
# 1. 检查请求量趋势
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total[5m]))'

# 2. 增加后端实例 (如果使用 Kubernetes)
kubectl scale deployment/backend --replicas=5 -n industrial-ai

# 3. 检查自动扩缩容
kubectl get hpa -n industrial-ai
```

---

## ✅ 验证恢复

```bash
# 1. 等告警自动解除 (等待 5-10 分钟)

# 2. 手动验证错误率
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total{status=~"5.."}[5m]))/sum(rate(http_requests_total[5m]))'

# 3. 验证健康检查
curl http://backend:8080/health

# 4. 验证日志无新错误
docker logs industrial-ai-backend --tail 100 | grep -i error
```

---

## 📝 事后复盘

### 必须记录
- 错误发生时间
- 错误类型分布
- 根本原因
- 修复步骤
- 用户影响评估

### 可能改进
- 如果是新代码 bug → 通知开发团队修复
- 如果是资源不足 → 配置自动扩缩容
- 如果是外部依赖 → 添加降级机制

---

## 🔗 相关 Runbook

- [HighResponseTime](../http/HighResponseTime.md) - 响应时间过长
- [DatabaseConnectionPoolExhausted](../database/DatabaseConnectionPoolExhausted.md) - 数据库连接耗尽
- [ServiceDown](../system/ServiceDown.md) - 服务完全宕机

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead