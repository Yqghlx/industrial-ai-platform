# 紧急预案 (Emergency Procedures)

> **Industrial AI Platform 紧急响应手册**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 🚨 紧急级别定义

| 级别 | 定义 | 响应时间 | 负责人 |
|------|------|---------|--------|
| **P0 - 灾难级** | 全服务宕机、数据丢失风险 | 立即 (5分钟内) | 全团队 + 管理层 |
| **P1 - 严重级** | 核心功能不可用 | 15分钟内 |值班 Ops + 开发 |
| **P2 - 重要级** | 性能严重下降、部分功能异常 | 30分钟内 | 值班 Ops |
| **P3 - 一般级** | 单点告警、非核心问题 | 4小时内 | 值班 Ops |

---

## 📋 P0 - 全服务宕机预案

### 触发条件
- Backend + PostgreSQL + Redis 全部宕机
- 数据中心故障
- 网络完全中断

### 紧急响应流程

#### 1️⃣ 立即通知 (0-5分钟)
```bash
# 紧急电话通知
值班 Ops: +86-xxx-xxxx
后端开发 Lead: +86-xxx-xxxx
管理层: +86-xxx-xxxx

# Slack 紧急频道
#alerts-critical: "🚨 P0: 全服务宕机，立即响应！"
```

#### 2️⃣ 快速诊断 (5-10分钟)
```bash
# 检查基础设施
docker ps -a
kubectl get nodes
kubectl get pods -A

# 检查网络
ping internal-network
curl http://load-balancer/health

# 检查存储
df -h
kubectl get pvc -A
```

#### 3️⃣ 恢复服务 (10-30分钟)
```bash
# 场景 A: Docker 全重启
docker-compose down
docker-compose up -d

# 场景 B: Kubernetes 全重启
kubectl rollout restart deployment -n industrial-ai

# 场景 C: 数据中心故障 → 启用备用环境
# 切换到备用数据中心或云环境
```

#### 4️⃣ 数据验证 (30-60分钟)
```bash
# 检查数据库完整性
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM devices;
SELECT count(*) FROM users;
SELECT count(*) FROM alerts;
"

# 检查是否有数据丢失
# 对比最近备份和数据现状
```

---

## 📋 P1 - 核心功能不可用预案

### 触发条件
- Backend 服务宕机
- 数据库连接池耗尽
- Redis 宕机导致无缓存

### 紧急响应流程

#### 1️⃣ 确认影响范围 (0-5分钟)
```bash
# 哪些服务受影响
curl http://backend:8080/health
curl http://frontend:80/health

# 哪些功能不可用
curl http://backend:8080/api/v1/devices -H "Authorization: Bearer xxx"
```

#### 2️⃣ 重启受影响服务 (5-15分钟)
```bash
# 重启 Backend
docker-compose restart backend

# 重启 PostgreSQL (如果必要)
docker-compose restart postgres

# 重启 Redis (如果必要)
docker-compose restart redis
```

#### 3️⃣ 验证恢复 (15-30分钟)
```bash
# 功能测试
curl http://backend:8080/api/v1/devices
curl http://backend:8080/api/v1/alerts
```

---

## 📋 P2 - 性能严重下降预案

### 触发条件
- HTTP P95 > 5s
- 数据库查询 P95 > 2s
- 缓存命中率 < 30%

### 紧急响应流程

#### 1️⃣ 确认瓶颈 (0-10分钟)
```bash
# 查看资源使用
docker stats --no-stream
kubectl top pods -n industrial-ai

# 查看请求量
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total[5m]))'
```

#### 2️⃣ 快速缓解 (10-30分钟)
```bash
# 扩容
kubectl scale deployment/backend --replicas=10 -n industrial-ai

# 清理长时间事务
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle in transaction' AND now() - query_start > interval '60 seconds';
"

# 重启 Redis 清理缓存
docker-compose restart redis
```

---

## 📋 数据恢复预案

### 数据库恢复

```bash
# 1. 检查备份文件
ls -lh /backup/postgres/

# 2. 恢复最近备份
docker exec -i industrial-ai-postgres psql -U industrial_user -d industrial_ai < /backup/postgres/latest.sql

# 3. 验证数据完整性
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT table_name, 
       pg_size_pretty(pg_total_relation_size(table_name::text)) as size
FROM information_schema.tables 
WHERE table_schema = 'public';
"
```

### Redis 数据恢复

```bash
# Redis 持久化配置
docker exec industrial-ai-redis redis-cli CONFIG GET save

# 从 RDB 文件恢复
# Redis 自动加载 dump.rdb
docker-compose restart redis
```

---

## 📋 网络故障预案

### 内网故障

```bash
# 1. 检查 DNS
nslookup internal-service

# 2. 检查防火墙
iptables -L -n

# 3. 检查负载均衡
curl http://load-balancer/status

# 4. 临时绕过负载均衡直连服务
curl http://backend-pod-ip:8080/health
```

### 外网故障 (用户无法访问)

```bash
# 1. 检查入口
curl -I https://industrial-ai.example.com

# 2. 检查 CDN/DNS
dig industrial-ai.example.com

# 3. 临时切换到备用域名
# 更新 DNS 或通知用户备用入口
```

---

## 📋 安全事件预案

### 数据泄露风险

```bash
# 1. 立即关闭外部访问
kubectl scale deployment/frontend --replicas=0 -n industrial-ai

# 2. 审查访问日志
docker logs industrial-ai-backend --since 24h | grep -i "unauthorized|leak"

# 3. 检查敏感数据访问
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT * FROM audit_log 
WHERE action LIKE '%sensitive%' 
ORDER BY timestamp DESC LIMIT 100;
"

# 4. 通知安全团队和管理层
```

### DDoS 攻击

```bash
# 1. 启用限流
# 调整 middleware 限流参数

# 2. 配置 IP 黑名单
iptables -A INPUT -s <attacker-ip> -j DROP

# 3. 启用 WAF 规则
# 如果有 WAF，启用更严格的规则

# 4. 联系 CDN/云服务商协助
```

---

## 📞 紧急联系人矩阵

### 值班轮换

| 周次 | 主值班 | 备值班 | 时间 |
|------|--------|--------|------|
| Week 1 | Ops A | Ops B | 周一 09:00 - 周一 09:00 |
| Week 2 | Ops C | Ops D | 周一 09:00 - 周一 09:00 |
| ... | ... | ... | ... |

**值班表链接**: https://wiki.industrial-ai.example.com/ops/rotation

### 紧急联系顺序

```
P0 → 值班 Ops → 开发 Lead → 管理层 → 全团队
P1 → 值班 Ops → 相关开发
P2 → 值班 Ops
P3 → 值班 Ops (工作时间)
```

### 外部支持

| 服务商 | 用途 | 联系方式 |
|--------|------|---------|
| **云服务商** | 基础设施故障 | 客服热线 |
| **数据库厂商** | PostgreSQL 专家支持 | 技术支持热线 |
| **安全厂商** | 安全事件响应 | 安全热线 |

---

## 📝 紧急事件记录模板

```markdown
## 紧急事件报告

**事件级别**: P0/P1/P2/P3
**发生时间**: YYYY-MM-DD HH:MM:SS
**发现方式**: 告警/用户反馈/巡检
**持续时间**: XX 分钟/小时

**影响范围**:
- 用户数: XX
- 功能: XXX
- 业务损失: XX

**响应人员**:
- 主值班: XXX
- 协助: XXX, XXX

**处理时间线**:
- HH:MM 收到告警
- HH:MM 确认问题
- HH:MM 执行 XXX
- HH:MM 服务恢复

**根本原因**:
- [具体原因分析]

**预防措施**:
- [改进建议]

**附件**:
- 日志文件: /logs/emergency-YYYYMMDD.log
- 监控截图: [Grafana 截图链接]
```

---

## 🔗 相关资源

- **Runbook 主页**: https://docs.industrial-ai.example.com/runbook
- **值班表**: https://wiki.industrial-ai.example.com/ops/rotation
- **备份位置**: /backup/industrial-ai/
- **监控仪表盘**: https://grafana.industrial-ai.example.com
- **工单系统**: https://jira.industrial-ai.example.com

---

**最后更新**: 2026-05-13  
**审核人**: Ops Team Lead + 管理层