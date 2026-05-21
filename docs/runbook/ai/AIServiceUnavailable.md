# AIServiceUnavailable

> **告警名称**: AIServiceUnavailable  
> **严重度**: Critical  
> **类别**: AI  
> **阈值**: 无 AI 查询但有请求 15 分钟  
> **持续时间**: 15 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
rate(ai_queries_total[10m]) == 0 and sum(http_requests_total{path=~"/api/v1/agent.*"}) > 0
```

**触发条件**:
- 有 AI 相关请求但无 AI 查询成功
- AI Agent 服务不可用

---

## 🚨 紧急响应

```bash
# 1. 立即检查 AI API 配置
echo $GLM_API_URL
echo $GLM_API_KEY

# 2. 测试 AI API 连通性
curl -I https://open.bigmodel.cn/api/paas/v4

# 3. 检查后端 AI 服务日志
docker logs industrial-ai-backend --tail 200 | grep -E "(ai|agent|glm)" | tail 50
```

---

## 🔍 诊断步骤

### 1️⃣ 检查 AI API 状态

```bash
# 测试 GLM API
curl -X POST https://open.bigmodel.cn/api/paas/v4/chat/completions \
  -H "Authorization: Bearer $GLM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"glm-4","prompt":"test"}'

# 检查 API 返回状态码
curl -I https://open.bigmodel.cn/api/paas/v4
```

### 2️⃣ 检查 API Key 状态

```bash
# 检查 API Key 是否过期
# 登录 GLM 控制台检查 API Key 状态

# 检查配额是否耗尽
# 查看 GLM 控制台 → API Usage
```

### 3️⃣ 检查后端 AI 配置

```bash
# 查看 AI 配置
docker exec industrial-ai-backend env | grep -E "(GLM|AI|API)"

# 检查 AI 服务日志
docker logs industrial-ai-backend --tail 500 | grep -E "(api.*error|timeout|quota)" | tail 30
```

---

## 🛠️ 修复方案

### 情况 A: API Key 过期或配额耗尽

```bash
# 1. 登录 GLM 控制台
https://open.bigmodel.cn

# 2. 检查 API Key 状态和配额

# 3. 如果配额耗尽 → 充值或申请更多配额

# 4. 如果 Key 过期 → 生成新 Key

# 5. 更新配置
kubectl create secret generic glm-api-secret \
  --from-literal=GLM_API_KEY=<new-key> \
  -n industrial-ai \
  --dry-run=client -o yaml | kubectl apply -f -

# 6. 重启后端加载新配置
kubectl rollout restart deployment/backend -n industrial-ai
```

### 情况 B: API 服务不可用

```bash
# 1. 检查 GLM 服务状态
# 关注 GLM 官方公告

# 2. 配置备用 API endpoint
# 如果 GLM 不可用，切换到备用 AI 服务

# 3. 临时禁用 AI 功能 (降级)
# 在配置中设置 ai.enabled = false

# 4. 通知用户 AI 功能临时不可用
```

### 情况 C: 网络问题导致无法连接

```bash
# 1. 检查出口网络
curl -I https://open.bigmodel.cn

# 2. 检查防火墙规则
iptables -L -n | grep 443

# 3. 检查 DNS 解析
dig open.bigmodel.cn

# 4. 配置代理 (如果必要)
export HTTPS_PROXY=http://proxy.example.com:8080
```

### 情况 D: 后端 AI 服务 Bug

```bash
# 1. 查看具体错误
docker logs industrial-ai-backend --tail 500 | grep -E "(panic|fatal)" | tail 20

# 2. 重启后端
docker-compose restart backend

# 3. 如果持续失败 → 回滚到稳定版本
kubectl rollout undo deployment/backend -n industrial-ai

# 4. 通知开发团队排查
```

---

## ✅ 验证恢复

```bash
# 1. 测试 AI API
curl -X POST https://open.bigmodel.cn/api/paas/v4/chat/completions \
  -H "Authorization: Bearer $GLM_API_KEY" \
  -d '{"model":"glm-4","prompt":"test"}' | jq .

# 2. 验证 AI 查询 Metrics
curl 'http://prometheus:9090/api/v1/query?query=rate(ai_queries_total[5m])'

# 3. 手动测试 AI 功能
curl http://backend:8080/api/v1/agent/chat \
  -H "Authorization: Bearer <user-token>" \
  -d '{"prompt":"测试消息"}'

# 4. 验证无新错误
docker logs industrial-ai-backend --tail 50 | grep -i "ai.*error"
```

---

**最后更新**: 2026-05-13  
**审核人**: AI Team Lead