# Cache Strategy Documentation

## Overview

This document describes the caching strategy implemented for the Industrial AI Platform backend. The cache layer provides performance optimization for frequently accessed data while maintaining data consistency.

## Architecture

### Cache Backend

The cache system supports two backends:

1. **Redis** (recommended for production)
   - Distributed caching
   - Persistence support
   - Better scalability
   - Connection URL: `redis://localhost:6379/0`

2. **In-Memory Cache** (fallback)
   - Automatic fallback when Redis is unavailable
   - Suitable for development/testing
   - Limited to single instance

### Cache Layers

```
┌─────────────────────────────────────────────────────┐
│                    Application                       │
├─────────────────────────────────────────────────────┤
│                 CacheService                         │
│  ┌─────────────┐     ┌──────────────┐               │
│  │   Redis     │ ──► │ Memory Cache │ (fallback)    │
│  └─────────────┘     └──────────────┘               │
├─────────────────────────────────────────────────────┤
│                   Database                          │
└─────────────────────────────────────────────────────┘
```

## Cache Strategy

### Cached Data Types

| Data Type       | TTL        | Cache Key Pattern        | Invalidation Trigger       |
|-----------------|------------|--------------------------|----------------------------|
| Device List     | 5 minutes  | `device:list:*`          | Device create/update/delete|
| Device Details  | 5 minutes  | `device:{id}`            | Device update/delete       |
| ROI Statistics  | 10 minutes | `roi:stats`              | Work order change          |
| Alert Statistics| 1 minute   | `alert:stats:*`          | New alert/resolved alert   |
| Latest Telemetry| 30 seconds | `telemetry:latest`       | New telemetry data         |
| Device Telemetry| 1 minute   | `telemetry:{device_id}`  | New telemetry for device   |

### Cache Key Naming Convention

All cache keys follow the pattern: `{prefix}:{category}:{subcategory}:{identifier}`

Examples:
- `iai:device:list:page1_size20` - Device list for page 1, size 20
- `iai:device:CNC-001` - Device with ID CNC-001
- `iai:roi:stats` - ROI statistics
- `iai:alert:stats:daily` - Daily alert statistics
- `iai:telemetry:CNC-001` - Telemetry for device CNC-001

## Cache Invalidation

### Strategies

1. **Time-based (TTL)**
   - Automatic expiration after TTL
   - Guarantees eventual consistency

2. **Write-through**
   - Cache cleared on data modification
   - Immediate consistency for critical data

3. **Pattern-based**
   - Clear all keys matching a pattern
   - Efficient for bulk invalidation

### Invalidation Rules

```go
// Device changes - clear device list and specific device cache
func InvalidateDeviceCache(ctx context.Context, deviceID string) {
    cache.DeleteByPattern(ctx, "device:list*")
    cache.Delete(ctx, "device:" + deviceID)
}

// Work order changes - clear ROI stats
func InvalidateWorkOrderCache(ctx context.Context) {
    cache.Delete(ctx, "roi:stats")
}

// New alert - clear alert stats
func InvalidateAlertCache(ctx context.Context) {
    cache.DeleteByPattern(ctx, "alert:stats*")
}
```

## Configuration

### Environment Variables

```bash
# Redis connection (optional)
REDIS_URL=redis://localhost:6379/0

# Enable/disable caching (default: true)
CACHE_ENABLED=true

# Cache key prefix (default: iai:)
CACHE_PREFIX=iai:
```

### Docker Compose

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru

  backend:
    environment:
      - REDIS_URL=redis://redis:6379/0
      - CACHE_ENABLED=true
```

## Monitoring

### Cache Statistics Endpoint

```bash
curl http://localhost:8080/cache/status
```

Response:
```json
{
  "available": true,
  "backend_type": "redis",
  "hit_rate": 85.5,
  "hits": 1234,
  "misses": 213,
  "keys_stored": 45,
  "sets": 100,
  "deletes": 20,
  "errors": 0,
  "last_error": ""
}
```

### Metrics to Monitor

1. **Hit Rate**
   - Target: > 80%
   - Low hit rate indicates cache inefficiency

2. **Keys Stored**
   - Monitor for memory usage
   - Redis: Check against maxmemory

3. **Errors**
   - Should be 0
   - Non-zero indicates backend issues

## Best Practices

### When to Use Cache

✅ Use cache for:
- Frequently read data
- Expensive computations
- Database query results
- Aggregation/statistics

❌ Don't cache:
- Real-time data (live telemetry)
- User-specific session data
- Sensitive data requiring strict consistency
- Very large objects (> 1MB)

### Cache Size Guidelines

- Individual cache value: < 100KB
- Total cache size: < 256MB (Redis)
- Consider pagination for large datasets

### Error Handling

```go
// Cache errors are non-critical
data, err := cache.Get(ctx, key)
if err == cache.ErrNotFound {
    // Cache miss - load from database
    data = loadFromDatabase()
    // Set cache (ignore errors)
    _ = cache.Set(ctx, key, data, ttl)
}
```

## Performance Impact

### Expected Improvements

| Endpoint              | Before (ms) | After (ms) | Improvement |
|-----------------------|-------------|------------|-------------|
| Device List           | 150         | 5          | 97%         |
| ROI Statistics        | 300         | 10         | 97%         |
| Alert Statistics      | 200         | 5          | 98%         |
| Device Details        | 50          | 2          | 96%         |

### Redis vs Memory Cache

| Metric          | Redis    | Memory   |
|-----------------|----------|----------|
| Hit Rate        | 95%      | 85%      |
| Latency         | 1-2ms    | < 1ms    |
| Scalability     | Multi-instance | Single instance |
| Persistence     | Yes      | No       |
| Memory Limit    | Configurable | 100MB   |

## Troubleshooting

### Common Issues

1. **Low Hit Rate**
   - Check TTL settings
   - Review cache invalidation logic
   - Verify key naming consistency

2. **Redis Connection Errors**
   - Verify Redis URL
   - Check Redis service status
   - System automatically falls back to memory cache

3. **High Memory Usage**
   - Reduce TTL values
   - Clear unnecessary cached data
   - Increase Redis maxmemory

### Debug Commands

```bash
# Redis CLI
redis-cli
> INFO memory
> KEYS iai:*
> TTL iai:device:list

# Check cache status
curl http://localhost:8080/cache/status

# Clear all cache (Redis)
redis-cli FLUSHDB
```

## Cache Warmup

Cache warmup is performed automatically on server startup:

1. Device list (first 100 devices)
2. ROI statistics
3. Active alert summary

Warmup runs asynchronously to not delay server startup.