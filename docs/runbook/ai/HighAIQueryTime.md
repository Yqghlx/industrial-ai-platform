# HighAIQueryTime

> **告警名称**: HighAIQueryTime  
> **严重度**: Warning  
> **类别**: AI  
> **阈值**: P95 AI 查询时间 > 60 秒  
> **持续时间**: 10 分钟

---

## 📊 告警详情

**PromQL 表达式**:
```promql
histogram_quantile(0.95, sum(rate(ai_query_duration_seconds_bucket[5m])) by (le)) > 60
```

**触发条件**:
- 95% 的 AI 查询超过 60 秒
- AI 响应非常慢

---

## 🔍 诊断步骤

### 1️⃣ 检查各模型响应时间

```bash
# 查看各模型的查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(ai_query_duration_seconds_bucket[5m]))by(le,model))'
```

### 2️⃣ 检查 GLM API 响应

```bash
# 测试 API 响应时间
time curl -X POST https://open.bigmodel.cn/api/paas/v4/chat/completions \
  -H "Authorization: Bearer $GLM_API_KEY" \
  -d '{"model":"glm-4","prompt":"测试"}'
```

### 3️⃣ 检查请求复杂度

```bash
# 查看查询 Token 数量分布
docker logs industrial-ai-backend --tail 200 | grep -E "tokens.*input|prompt.*length" | tail 20
```

---

## 🛠️ 修复方案

### 情况 A: GLM API 响应慢

```bash
# 1. 检查 GLM 服务状态
# 关注 GLM 官方状态页面

# 2. 切换到更快的模型
# 配置默认使用 glm-4-flash (更快)

# 3. 配置请求超时
# 设置 AI 查询超时为 30 秒，避免长时间等待

# 4. 如果持续慢 → 考虑备用 AI 服务
```

### 情况 B: 查询过于复杂

```bash
# 1. 分析慢查询的 Prompt
docker logs industrial-ai-backend --tail 500 | grep -E "duration.*>.*30" | tail 20

# 2. 优化 Prompt 长度
# 限制用户输入最大长度

# 3. 配置模型参数
# 减少 max_tokens，降低输出长度
```

### 情况 C: 网络延迟

```bash
# 1. 检查网络延迟
ping open.bigmodel.cn

# 2. 配置代理加速
# 如果网络延迟高，配置更快的代理

# 3. 检查 SSL 握手时间
openssl s_client -connect open.bigmodel.cn:443 -servername open.bigmodel.cn
```

---

## ✅ 验证恢复

```bash
# 1. 验证 AI 查询时间
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(ai_query_duration_seconds_bucket[5m]))by(le))'

# 2. 验证 API 响应
time curl -X POST https://open.bigmodel.cn/api/paas/v4/chat/completions \
  -H "Authorization: Bearer $GLM_API_KEY" \
  -d '{"model":"glm-4-flash","prompt":"test"}'
# 期望: < 10 秒
```

---

**最后更新**: 2026-05-13  
**审核人**: AI Team Lead