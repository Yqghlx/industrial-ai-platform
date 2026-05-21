# 缓存策略实施计划

> **任务**: Task #6 - 缓存策略扩展
> **目标**: API响应时间 -80%（缓存命中时）
> **预估时间**: 30分钟

---

## 📊 当前状态

**✅ 已有基础设施**:
- 缓存服务（pkg/cache）
- CacheKeyBuilder
- Cache Prefix（DeviceCachePrefix, AlertCachePrefix, etc）
- handler 层已集成 cacheSvc

**⏳ 待实施**:
- handler 层使用缓存策略
- 缓存失效机制

---

## 🎯 需要添加缓存的 API

| API | 缓存 Key | TTL | 优先级 |
|-----|----------|-----|--------|
| `/devices` | `devices:list:{tenant}:{page}` | 5min | P1 |
| `/devices/{id}` | `device:{id}` | 5min | P1 |
| `/alerts/stats` | `alerts:stats:{tenant}` | 1min | P1 |
| `/reports/roi` | `roi:stats:{tenant}` | 10min | P2 |

---

## 📝 实施策略

### Phase 1: 设备列表缓存（10分钟）

**修改**: `internal/handler/device_handler.go`

```go
// 缓存 Key 格式
key := cache.DeviceCachePrefix.Build("list", tenantID, strconv.Itoa(page))

// 使用 GetOrSetJSON
err := cache.GetOrSetJSON(ctx, h.cache, key, func() (interface{}, error) {
    return h.deviceRepo.List(ctx, page, pageSize)
}, 5*time.Minute, &result)
```

---

### Phase 2: 告警统计缓存（10分钟）

**修改**: `internal/handler/business_handler.go`

```go
// 缓存 Key 格式
key := cache.AlertCachePrefix.Build("stats", tenantID)

// 使用 GetOrSetJSON
err := cache.GetOrSetJSON(ctx, h.cache, key, func() (interface{}, error) {
    return s.calculateAlertStats(ctx, tenantID)
}, 1*time.Minute, &stats)
```

---

### Phase 3: 缓存失效机制（10分钟）

**需要在以下事件时失效缓存**:
- 设备创建/更新/删除 → 失效 `devices:list:*`
- 告警创建/状态变更 → 失效 `alerts:stats:*`

---

## ⚠️ 实施注意

**由于 handler 覆盖率较低（56.1%），直接修改可能影响测试**

**建议方案**:
1. 先完善 handler Mock（架构重构）
2. 再添加缓存策略
3. 或者直接添加缓存策略，后续补充测试

---

## 🎯 当前建议

**Option A**: 等待 handler Mock 重构后实施（推荐）
**Option B**: 直接实施缓存策略（风险：影响测试覆盖率）
**Option C**: 创建缓存策略文档，标记待实施

---

## ✅ 缓存基础设施已完成

**已有功能**:
- Memory Cache（内存缓存）
- Redis Cache（Redis 缓存）
- 自动降级（Redis → Memory）
- Cache Key Builder
- GetOrSetJSON helper
- TTL 管理

**仅需**: handler 层集成使用

---

**状态**: ⏳ 等待决策
**下一步**: 选择实施方案