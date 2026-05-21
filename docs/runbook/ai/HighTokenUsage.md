# HighTokenUsage

> **告警名称**: HighTokenUsage  
> **严重度**: Warning  
> **类别**: AI  
> **阈值**: Token 用量 > 100K/小时  
> **持续时间**: 1 小时

---

## 📊 告警详情

**PromQL 表达式**:
```promql
sum(rate(ai_tokens_used_total[1h])) > 100000
```

**触发条件**:
- AI Token 消耗超过 100,000/小时
- 可能成本失控

---

## 🔍 诊断步骤

### 1️⃣ 分析 Token 消耗来源

```bash
# 查看各模型的 Token 消耗
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(ai_tokens_used_total[1h]))by(model)'

# 查看输入/输出 Token 分布
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(ai_tokens_used_total[1h]))by(type)'
```

### 2️⃣ 检查高频用户

```bash
# 查看各租户的 Token 消耗
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(ai_tokens_used_total[1h]))by(tenant_id)'

# 查看数据库中的 AI 查询记录
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT tenant_id, count(*) as query_count, sum(tokens_used) as total_tokens
FROM ai_queries
WHERE created_at > now() - interval '1 hour'
GROUP BY tenant_id
ORDER BY total_tokens DESC
LIMIT 10;
"
```

### 3️⃣ 检查是否有滥用

```bash
# 查看是否有异常高频查询
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT user_id, count(*) as query_count
FROM ai_queries
WHERE created_at > now() - interval '1 hour'
GROUP BY user_id
ORDER BY query_count DESC
LIMIT 20;
"

# 检查是否有超长 Prompt
docker logs industrial-ai-backend --tail 200 | grep -E "tokens.*>.*1000" | tail 20
```

---

## 🛠️ 修复方案

### 情况 A: 用户正常高频使用

```bash
# 1. 确认用户是否有合理需求
# 联系用户确认使用场景

# 2. 如果合理 → 增加 Token 配额
# 调整租户 Token 限制

# 3. 配置成本预警阈值
# 在 GLM 控制台设置用量告警

# 4. 优化 Prompt 模板
# 减少不必要的 Token 消耗
```

### 情况 B: 用户滥用

```bash
# 1. 识别滥用用户
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT user_id, tenant_id
FROM ai_queries
WHERE created_at > now() - interval '1 hour'
GROUP BY user_id, tenant_id
HAVING count(*) > 100
ORDER BY count(*) DESC;
"

# 2. 临时限制用户 AI 功能
curl -X PUT http://backend:8080/api/v1/users/<user-id>/limits \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"ai_enabled":false}'

# 3. 联系用户沟通使用规范

# 4. 配置限流策略
# 限制每用户 AI 查询频率
```

### 情况 C: 系统问题导致重复请求

```bash
# 1. 检查是否有重试风暴
docker logs industrial-ai-backend --tail 500 | grep -E "retry.*ai.*query" | tail 50

# 2. 检查是否配置了过于激进的重试策略

# 3. 调整重试配置
# 减少重试次数，增加重试间隔

# 4. 修复 Bug (如果有)
# 通知开发团队排查
```

### 情况 D: 模型配置问题

```bash
# 1. 检查使用的模型
# 确认是否使用了 Token 消耗大的模型

# 2. 切换到更经济的模型
# 默认使用 glm-4-flash (成本更低)

# 3. 限制 max_tokens 输出长度
# 配置最大输出 Token 为 500

# 4. 配置输入长度限制
# 限制用户 Prompt 最大长度
```

---

## ✅ 验证恢复

```bash
# 1. 验证 Token 消耗下降
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(ai_tokens_used_total[1h]))'
# 期望: < 100000

# 2. 验证用户查询频率
docker exec -it industrial-ai-postgres psql -U industrial_user -c "
SELECT count(*) FROM ai_queries WHERE created_at > now() - interval '5 minutes';
"
# 期望: 合理范围内

# 3. 验证 GLM 控制台用量
https://open.bigmodel.cn/usage
```

---

**最后更新**: 2026-05-13  
**审核人**: AI Team Lead