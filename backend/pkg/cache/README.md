# Cache Package

统一缓存层，支持 Redis 和内存缓存，并提供自动降级功能。

## 功能特性

- **双缓存后端**: 支持 Redis 和内存缓存
- **自动降级**: Redis 不可用时自动切换到内存缓存
- **可选缓存**: 可完全禁用缓存功能
- **缓存预热**: 支持启动时预加载热点数据
- **监控统计**: 提供缓存命中率等统计信息
- **模式删除**: 支持按模式批量删除缓存键

## 配置

### 环境变量

```bash
# Redis URL (可选)
REDIS_URL=redis://localhost:6379/0

# 是否启用缓存 (默认: true)
CACHE_ENABLED=true

# 缓存键前缀 (默认: iai:)
CACHE_PREFIX=iai:
```

### 配置对象

```go
import "github.com/industrial-ai/platform/pkg/cache"

cfg := &cache.Config{
    RedisURL:      "redis://localhost:6379/0",
    Enabled:       true,
    DefaultTTL:    5 * time.Minute,
    MaxMemorySize: 100 * 1024 * 1024, // 100MB
    Prefix:        "iai:",
}

cacheService := cache.New(cfg)
```

## 使用方法

### 基本操作

```go
// Get
data, err := cacheService.Get(ctx, "device:list")
if err == cache.ErrNotFound {
    // 缓存未命中，从数据库加载
}

// Set
cacheService.Set(ctx, "device:list", jsonData, 5*time.Minute)

// Delete
cacheService.Delete(ctx, "device:list")

// DeleteByPattern
cacheService.DeleteByPattern(ctx, "device:*")
```

### GetOrSet 模式

```go
// 自动处理缓存未命中情况
data, err := cache.GetOrSet(ctx, cacheService, "device:list",
    func() ([]byte, error) {
        // 从数据库加载
        devices, err := deviceRepo.List(ctx)
        if err != nil {
            return nil, err
        }
        return json.Marshal(devices)
    },
    5*time.Minute,
)
```

### GetOrSetJSON 模式

```go
// JSON 序列化/反序列化自动处理
var devices []model.Device
err := cache.GetOrSetJSON(ctx, cacheService, "device:list",
    func() (interface{}, error) {
        return deviceRepo.List(ctx)
    },
    5*time.Minute,
    &devices,
)
```

### 缓存键构建

```go
// 使用预定义的键构建器
key := cache.DeviceCachePrefix.Build("list", "page1")
// 结果: "device:list:page1"

key := cache.ROICachePrefix.Build("stats")
// 结果: "roi:stats"

key := cache.AlertCachePrefix.Build("stats", "daily")
// 结果: "alert:stats:daily"

// 自定义键构建器
builder := cache.NewCacheKeyBuilder("custom")
key := builder.Build("part1", "part2")
// 结果: "custom:part1:part2"
```

## 预定义 TTL

```go
// 设备列表缓存 - 5分钟
cache.DeviceListTTL

// ROI 统计缓存 - 10分钟
cache.ROIStatsTTL

// 告警统计缓存 - 1分钟
cache.AlertStatsTTL

// 最新遥测数据缓存 - 30秒
cache.TelemetryLatestTTL
```

## 缓存预热

```go
warmup := cache.NewWarmupService(cacheService)

// 注册预热加载器
warmup.RegisterLoader(func(ctx context.Context, cache cache.CacheService) error {
    devices, err := deviceRepo.List(ctx, 1, 100)
    if err != nil {
        return err
    }
    data, _ := json.Marshal(devices)
    return cache.Set(ctx, "device:list", data, cache.DeviceListTTL)
})

// 执行预热
warmup.Warmup(ctx)

// 异步预热
warmup.WarmupAsync()

// 定期预热
stopChan := warmup.ScheduleWarmup(30 * time.Minute)
close(stopChan) // 停止定期预热
```

## 监控统计

```go
stats := cacheService.GetStats()
fmt.Printf("缓存命中率: %.2f%%\n", 
    float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100)
fmt.Printf("缓存键数量: %d\n", stats.KeysStored)
fmt.Printf("后端类型: %s\n", stats.BackendType)
fmt.Printf("是否可用: %v\n", stats.Available)
```

## 缓存失效策略

### 数据更新时清除缓存

```go
// 更新设备后清除设备列表缓存
func (s *DeviceService) Update(ctx context.Context, device *model.Device) error {
    // 1. 更新数据库
    err := s.deviceRepo.Update(ctx, device)
    if err != nil {
        return err
    }
    
    // 2. 清除相关缓存
    s.cache.Delete(ctx, "device:list")
    s.cache.Delete(ctx, "device:"+device.ID)
    
    return nil
}
```

### 模式删除

```go
// 清除所有设备相关缓存
s.cache.DeleteByPattern(ctx, "device:*")

// 清除所有 ROI 统计缓存
s.cache.DeleteByPattern(ctx, "roi:*")
```

## 降级处理

缓存系统自动处理以下降级场景：

1. **Redis 不可用**: 自动切换到内存缓存
2. **禁用缓存**: 使用 noop 缓存，所有操作返回 ErrNotFound
3. **缓存操作失败**: GetOrSet 模式自动跳过缓存直接加载

## 线程安全

所有缓存实现都是线程安全的，可以在多线程环境下安全使用。

## 性能建议

1. **合理设置 TTL**: 根据数据更新频率设置适当的 TTL
2. **避免缓存大对象**: 大对象会占用大量内存，考虑分片缓存
3. **使用模式删除**: 数据更新时清除相关所有缓存，避免遗漏
4. **监控命中率**: 低命中率可能意味着 TTL 设置不当或缓存键策略问题