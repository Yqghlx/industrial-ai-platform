# 缓存策略优化指南

> **Industrial AI Platform 缓存优化最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 缓存策略概述

Phase 4 P1 缓存策略优化目标：

| 指标 | 当前状态 | 目标值 | 提升幅度 |
|------|---------|--------|---------|
| **缓存命中率** | ~60% | >90% | 30% |
| **热点数据覆盖** | 手动 | 自动识别 | 智能化 |
| **预热效率** | 无 | 启动预热 | 新增 |
| **缓存一致性** | 基础 | 强一致性 | 提升 |

---

## 🔄 缓存模式

### Cache-Aside (旁路缓存)

**最常用模式，适合读多写少场景：**

```go
// 读取流程
func GetDevice(deviceID string) (*Device, error) {
    // 1. 先查缓存
    cached, err := cache.Get(deviceID)
    if err == nil {
        return cached, nil // 缓存命中
    }
    
    // 2. 缓存未命中，查数据库
    device, err := db.GetDevice(deviceID)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    cache.Set(deviceID, device, 5*time.Minute)
    
    return device, nil
}

// 写入流程
func UpdateDevice(deviceID string, device *Device) error {
    // 1. 更新数据库
    err := db.UpdateDevice(deviceID, device)
    if err != nil {
        return err
    }
    
    // 2. 删除缓存 (下次读取时重新加载)
    cache.Delete(deviceID)
    
    return nil
}
```

### Write-Through (穿透写)

**写入时同步更新缓存：**

```go
func UpdateDevice(deviceID string, device *Device) error {
    // 1. 同时写入缓存和数据库
    cache.Set(deviceID, device, 5*time.Minute)
    err := db.UpdateDevice(deviceID, device)
    if err != nil {
        // 写入失败，删除缓存
        cache.Delete(deviceID)
        return err
    }
    return nil
}
```

### Write-Behind (异步写)

**写入缓存，异步写入数据库：**

```go
func UpdateDevice(deviceID string, device *Device) error {
    // 1. 写入缓存
    cache.Set(deviceID, device, 5*time.Minute)
    
    // 2. 异步写入数据库
    queue.Enqueue(&WriteTask{
        Type: "device_update",
        ID: deviceID,
        Data: device,
    })
    
    return nil
}
```

---

## 🔥 热点数据缓存

### 热点数据识别

**自动识别高频访问数据：**

| 数据类型 | 访问频率 | 缓存优先级 |
|----------|---------|-----------|
| **在线设备状态** | 极高 | **P0 预热** |
| **活跃告警列表** | 高 | **P0 预热** |
| **租户配置** | 高 | **P1 预热** |
| **用户 Session** | 中 | **P2 按需** |
| **遥测历史** | 中 | **P2 按需** |
| **AI 响应** | 低 | **P3 按需** |

### 热点数据 TTL

```yaml
# 缓存 TTL 配置
cache_ttl:
  device_status: 5m       # 设备状态 (频繁变化)
  device_list: 30m        # 设备列表 (相对稳定)
  alerts_active: 10m      # 活跃告警 (实时性要求)
  telemetry_recent: 1h    # 近期遥测 (历史数据)
  user_session: 15m       # 用户 Session (安全刷新)
  user_profile: 1h        # 用户信息 (相对稳定)
  tenant_config: 24h      # 租户配置 (极少变化)
  ai_response: 30m        # AI 响应 (可重复使用)
```

---

## 🌡️ 缓存预热策略

### 启动预热

**应用启动时预热热点数据：**

```go
func WarmupOnStartup(cache *CacheService, db *Database) {
    log.Println("Cache warmup on startup...")
    
    // 1. 预热所有租户的在线设备
    tenants := db.GetAllTenantIDs()
    for _, tenantID := range tenants {
        devices := db.GetOnlineDevices(tenantID)
        cache.WarmupDeviceStatus(tenantID, devices)
    }
    
    // 2. 预热活跃告警
    for _, tenantID := range tenants {
        alerts := db.GetActiveAlerts(tenantID)
        cache.WarmupActiveAlerts(tenantID, alerts)
    }
    
    // 3. 预热租户配置
    for _, tenantID := range tenants {
        config := db.GetTenantConfig(tenantID)
        cache.SetTenantConfig(tenantID, config)
    }
    
    log.Println("Cache warmup completed")
}
```

### 定时预热

**定时更新热点数据缓存：**

```go
// 每 5 分钟刷新设备状态缓存
func ScheduleDeviceStatusWarmup(cache *CacheService, db *Database) {
    ticker := time.NewTicker(5 * time.Minute)
    for {
        select {
        case <-ticker.C:
            tenants := db.GetAllTenantIDs()
            for _, tenantID := range tenants {
                devices := db.GetOnlineDevices(tenantID)
                cache.WarmupDeviceStatus(tenantID, devices)
            }
        }
    }
}

// 每 10 分钟刷新告警缓存
func ScheduleAlertsWarmup(cache *CacheService, db *Database) {
    ticker := time.NewTicker(10 * time.Minute)
    for {
        select {
        case <-ticker.C:
            tenants := db.GetAllTenantIDs()
            for _, tenantID := range tenants {
                alerts := db.GetActiveAlerts(tenantID)
                cache.WarmupActiveAlerts(tenantID, alerts)
            }
        }
    }
}
```

---

## 🔑 缓存键命名规范

### 标准格式

```
{namespace}:{tenant_id}:{entity}:{entity_id}:{field}

# 示例
cache:tenant_001:device:dev_123:status
cache:tenant_001:device:dev_123:info
cache:tenant_001:alerts:active
cache:tenant_001:devices:online
cache:tenant_001:user:user_456:profile
cache:tenant_001:tenant:config
```

### 特殊键

```
# JWT 黑名单
jwt_blacklist:token_id_abc123

# 访问计数 (热点识别)
access_count:cache:tenant_001:device:dev_123:status

# 分布式锁
lock:tenant_001:device:dev_123:update
```

---

## 🔄 缓存一致性

### 一致性策略

| 场景 | 策略 | 说明 |
|------|------|------|
| **数据更新** | 先更新 DB，再删除缓存 | 避免脏数据 |
| **数据删除** | 先删除缓存，再删除 DB | 防止读取已删除数据 |
| **批量更新** | 删除相关缓存，下次加载 | 简化逻辑 |
| **租户切换** | 清除租户所有缓存 | 安全隔离 |

### 缓存失效策略

```go
// 数据更新时失效缓存
func InvalidateCache(cache *CacheService, tenantID, entity, entityID string) {
    // 1. 删除实体缓存
    cache.Delete(tenantID, entity, entityID)
    
    // 2. 删除关联缓存 (如列表)
    cache.Delete(tenantID, entity, "list")
    cache.Delete(tenantID, entity, "all")
}

// 租户数据变更时清除所有缓存
func InvalidateTenantCache(cache *CacheService, tenantID string) {
    cache.DeleteByPattern(fmt.Sprintf("cache:%s:*", tenantID))
}
```

---

## 🛡️ 缓存穿透/击穿/雪崩

### 缓存穿透

**查询不存在数据，缓存和 DB 都无数据：**

```go
// 解决方案: 缓存空值
func GetDevice(deviceID string) (*Device, error) {
    cached, err := cache.Get(deviceID)
    if err == nil {
        if cached == "NULL" {
            return nil, errors.New("device not found")
        }
        return cached, nil
    }
    
    device, err := db.GetDevice(deviceID)
    if err != nil {
        // 缓存空值，防止穿透
        cache.Set(deviceID, "NULL", 5*time.Minute)
        return nil, err
    }
    
    cache.Set(deviceID, device, 5*time.Minute)
    return device, nil
}
```

### 缓存击穿

**热点数据过期，大量请求同时访问 DB：**

```go
// 解决方案: 分布式锁
func GetDeviceWithLock(deviceID string) (*Device, error) {
    cached, err := cache.Get(deviceID)
    if err == nil {
        return cached, nil
    }
    
    // 获取分布式锁
    lock := cache.AcquireLock(deviceID, 10*time.Second)
    if lock.Acquired() {
        // 查询 DB
        device, err := db.GetDevice(deviceID)
        if err != nil {
            cache.Set(deviceID, "NULL", 5*time.Minute)
        } else {
            cache.Set(deviceID, device, 5*time.Minute)
        }
        lock.Release()
        return device, err
    }
    
    // 等待其他请求完成
    time.Sleep(100 * time.Millisecond)
    return cache.Get(deviceID)
}
```

### 缓存雪崩

**大量缓存同时过期：**

```go
// 解决方案: TTL 随机化
func SetCache(key, value string, baseTTL time.Duration) {
    // 添加随机偏移 (0-20%)
    randomOffset := time.Duration(rand.Intn(int(baseTTL) / 5))
    actualTTL := baseTTL + randomOffset
    cache.Set(key, value, actualTTL)
}

// 永不过期策略 (配合主动更新)
func SetCacheWithUpdate(key, value string) {
    cache.Set(key, value, 0) // TTL=0 表示永不过期
    // 后台定时更新
}
```

---

## ✅ 缓存优化验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **缓存命中率** | >90% | Redis INFO stats |
| **热点数据覆盖** | P0 数据预热 | 启动日志 |
| **预热时间** | <30s | 启动耗时 |
| **缓存一致性** | 更新后失效 | 功能测试 |
| **穿透防护** | 空值缓存 | 异常测试 |
| **击穿防护** | 分布式锁 | 并发测试 |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team