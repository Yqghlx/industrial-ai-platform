// Package benchmarks provides performance benchmark tests for the Industrial AI Platform
// P2-001: Cache Efficiency Benchmarks
// These benchmarks test cache performance using standalone implementations
package benchmarks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// ==========================================
// Memory Cache Implementation (Standalone)
// ==========================================

// MemoryCache implements a simple in-memory cache for benchmarks
type MemoryCache struct {
	data  map[string]*memoryItem
	mu    sync.RWMutex
	ttl   time.Duration
	stats CacheStats
}

type memoryItem struct {
	value     []byte
	expiredAt time.Time
}

type CacheStats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	KeysStored  int64
	Errors      int64
	Available   bool
	BackendType string
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		data: make(map[string]*memoryItem),
		ttl:  ttl,
		stats: CacheStats{
			Available:   true,
			BackendType: "memory",
		},
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, fmt.Errorf("not found")
	}
	
	if time.Now().After(item.expiredAt) {
		c.stats.Misses++
		return nil, fmt.Errorf("not found")
	}
	
	c.stats.Hits++
	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if ttl == 0 {
		ttl = c.ttl
	}
	
	c.data[key] = &memoryItem{
		value:     value,
		expiredAt: time.Now().Add(ttl),
	}
	c.stats.Sets++
	c.stats.KeysStored = int64(len(c.data))
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.data, key)
	c.stats.Deletes++
	c.stats.KeysStored = int64(len(c.data))
	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.data[key]
	if !exists {
		return false
	}
	
	return time.Now().Before(item.expiredAt)
}

func (c *MemoryCache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	count := 0
	for key, item := range c.data {
		if now.After(item.expiredAt) {
			delete(c.data, key)
			count++
		}
	}
	c.stats.KeysStored = int64(len(c.data))
	return count
}

func (c *MemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.stats.KeysStored = int64(len(c.data))
	return c.stats
}

func (c *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for key := range c.data {
		if strings.Contains(pattern, "*") {
			prefix := strings.Replace(pattern, "*", "", -1)
			if strings.HasPrefix(key, prefix) {
				delete(c.data, key)
				c.stats.Deletes++
			}
		} else if key == pattern {
			delete(c.data, key)
			c.stats.Deletes++
		}
	}
	c.stats.KeysStored = int64(len(c.data))
	return nil
}

func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// ==========================================
// Redis-like Mock Implementation
// ==========================================

// RedisLikeCache uses miniredis for Redis-like benchmarks
type RedisLikeCache struct {
	client interface{}
	addr   string
}

// BenchmarkCacheMemoryGet benchmarks memory cache Get operation
func BenchmarkCacheMemoryGet(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		memoryCache.Set(ctx, key, value, 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		_, _ = memoryCache.Get(ctx, key)
	}
}

// BenchmarkCacheMemorySet benchmarks memory cache Set operation
func BenchmarkCacheMemorySet(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		memoryCache.Set(ctx, key, value, 5*time.Minute)
	}
}

// BenchmarkCacheMemoryDelete benchmarks memory cache Delete operation
func BenchmarkCacheMemoryDelete(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Pre-populate cache
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%10000)
		memoryCache.Delete(ctx, key)
	}
}

// BenchmarkCacheMemoryExists benchmarks memory cache Exists operation
func BenchmarkCacheMemoryExists(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		memoryCache.Exists(ctx, key)
	}
}

// BenchmarkCacheMemoryConcurrent benchmarks concurrent memory cache operations
func BenchmarkCacheMemoryConcurrent(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%100)
			if i%3 == 0 {
				memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
			} else if i%3 == 1 {
				memoryCache.Get(ctx, key)
			} else {
				memoryCache.Delete(ctx, key)
			}
			i++
		}
	})
}

// BenchmarkCacheMemoryCleanup benchmarks memory cache cleanup
func BenchmarkCacheMemoryCleanup(b *testing.B) {
	memoryCache := NewMemoryCache(1 * time.Second)
	ctx := context.Background()
	
	// Pre-populate cache with expired items
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 1*time.Nanosecond) // Immediately expired
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memoryCache.Cleanup()
	}
}

// BenchmarkCacheMemoryLargeValue benchmarks large value caching
func BenchmarkCacheMemoryLargeValue(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	// Create large value (100KB)
	largeValue := make([]byte, 100*1024)
	for i := 0; i < len(largeValue); i++ {
		largeValue[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_large_%d", i%10)
		memoryCache.Set(ctx, key, largeValue, 5*time.Minute)
		memoryCache.Get(ctx, key)
	}
}

// BenchmarkCacheRedisGet benchmarks Redis-like cache Get operation
func BenchmarkCacheRedisGet(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Pre-populate Redis
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		mr.Set(key, fmt.Sprintf("bench_value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		_ = mr.Get(key)
	}
}

// BenchmarkCacheRedisSet benchmarks Redis-like cache Set operation
func BenchmarkCacheRedisSet(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		mr.Set(key, fmt.Sprintf("bench_value_%d", i))
	}
}

// BenchmarkCacheRedisDelete benchmarks Redis-like cache Delete operation
func BenchmarkCacheRedisDelete(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Pre-populate Redis
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		mr.Set(key, "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%10000)
		mr.Del(key)
	}
}

// BenchmarkCacheRedisTTL benchmarks Redis TTL operations
func BenchmarkCacheRedisTTL(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Pre-populate Redis with TTL
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_ttl_%d", i)
		mr.Set(key, "value")
		mr.SetTTL(key, 60*time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_ttl_%d", i%1000)
		_ = mr.TTL(key)
	}
}

// BenchmarkCacheRedisExists benchmarks Redis Exists operation
func BenchmarkCacheRedisExists(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Pre-populate Redis
	for i := 0; i < 1000; i++ {
		mr.Set(fmt.Sprintf("key_%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%1000)
		_ = mr.Exists(key)
	}
}

// BenchmarkCacheRedisConcurrent benchmarks concurrent Redis operations
func BenchmarkCacheRedisConcurrent(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Pre-populate
	for i := 0; i < 100; i++ {
		mr.Set(fmt.Sprintf("key_%d", i), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%100)
			if i%3 == 0 {
				mr.Set(key, fmt.Sprintf("value_%d", i))
			} else if i%3 == 1 {
				_ = mr.Get(key)
			} else {
				mr.Del(key)
			}
			i++
		}
	})
}

// BenchmarkCacheRedisLargeValue benchmarks large value caching in Redis
func BenchmarkCacheRedisLargeValue(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Create large value (100KB)
	largeValue := make([]byte, 100*1024)
	for i := 0; i < len(largeValue); i++ {
		largeValue[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_large_%d", i%10)
		mr.Set(key, string(largeValue))
		_ = mr.Get(key)
	}
}

// BenchmarkCacheRedisIncrement benchmarks Redis increment operations
func BenchmarkCacheRedisIncrement(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("counter", "0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mr.Incr("counter")
	}
}

// BenchmarkCacheRedisDecrement benchmarks Redis decrement operations
func BenchmarkCacheRedisDecrement(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("counter", "10000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mr.Decr("counter")
	}
}

// BenchmarkCacheRedisHash benchmarks Redis hash operations
func BenchmarkCacheRedisHash(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("hash_%d", i%10)
		field := fmt.Sprintf("field_%d", i%100)
		value := fmt.Sprintf("value_%d", i)
		mr.HSet(key, field, value)
		_ = mr.HGet(key, field)
	}
}

// BenchmarkCacheRedisList benchmarks Redis list operations
func BenchmarkCacheRedisList(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mr.Push("list", fmt.Sprintf("item_%d", i))
	}
}

// BenchmarkCacheHitRate benchmarks cache hit rate scenarios
func BenchmarkCacheHitRate(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	// Populate 80% of keys (simulate 80% hit rate)
	for i := 0; i < 800; i++ {
		key := fmt.Sprintf("key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000) // 80% of keys exist
			memoryCache.Get(ctx, key)
			i++
		}
	})
}

// BenchmarkCacheMissRate benchmarks cache miss rate scenarios
func BenchmarkCacheMissRate(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	// Populate only 20% of keys (simulate 80% miss rate)
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000) // Only 20% exist
			memoryCache.Get(ctx, key)
			i++
		}
	})
}

// BenchmarkCacheSerialization benchmarks cache value serialization
func BenchmarkCacheSerialization(b *testing.B) {
	data := map[string]interface{}{
		"device_id":   "device-001",
		"temperature": 75.5,
		"vibration":   2.3,
		"pressure":    120.0,
		"status":      "running",
		"timestamp":   time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate serialization
		result := fmt.Sprintf("%s:%v,%s:%v,%s:%v,%s:%v,%s:%v",
			"device_id", data["device_id"],
			"temperature", data["temperature"],
			"vibration", data["vibration"],
			"pressure", data["pressure"],
			"status", data["status"],
		)
		_ = []byte(result)
	}
}

// BenchmarkCacheKeyGeneration benchmarks cache key generation
func BenchmarkCacheKeyGeneration(b *testing.B) {
	prefix := "device_telemetry"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%s:%s:%s", prefix, "device-001", "latest")
		_ = key
	}
}

// BenchmarkCacheMultiGet benchmarks multi-get operations (simulated)
func BenchmarkCacheMultiGet(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		keys[i] = fmt.Sprintf("multi_key_%d", i)
		memoryCache.Set(ctx, keys[i], []byte(fmt.Sprintf("value_%d", i)), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			memoryCache.Get(ctx, key)
		}
	}
}

// BenchmarkCacheMultiSet benchmarks multi-set operations (simulated)
func BenchmarkCacheMultiSet(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			key := fmt.Sprintf("multi_key_%d", j)
			value := []byte(fmt.Sprintf("value_%d_%d", i, j))
			memoryCache.Set(ctx, key, value, 5*time.Minute)
		}
	}
}

// BenchmarkCachePrefixOperations benchmarks prefix-based operations
func BenchmarkCachePrefixOperations(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	prefix := "device:"
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("%s%d:telemetry", prefix, i)
		memoryCache.Set(ctx, key, []byte("data"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memoryCache.DeleteByPattern(ctx, "device:*:telemetry")
	}
}

// BenchmarkCacheExpire benchmarks TTL expiration handling
func BenchmarkCacheExpire(b *testing.B) {
	memoryCache := NewMemoryCache(1 * time.Millisecond)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("expire_key_%d", i)
		memoryCache.Set(ctx, key, []byte("value"), 1*time.Millisecond)
		time.Sleep(2 * time.Millisecond) // Wait for expiration
		memoryCache.Get(ctx, key) // Should miss
	}
}

// BenchmarkCacheAvailabilityCheck benchmarks availability checks
func BenchmarkCacheAvailabilityCheck(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = memoryCache.GetStats().Available
	}
}

// BenchmarkCacheMutex benchmarks mutex operations in cache
func BenchmarkCacheMutex(b *testing.B) {
	var mu sync.RWMutex
	data := make(map[string][]byte)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate read operation
			mu.RLock()
			_ = data["key"]
			mu.RUnlock()
			
			// Simulate write operation
			mu.Lock()
			data["key"] = []byte("value")
			mu.Unlock()
		}
	})
}

// BenchmarkCacheStatsCollection benchmarks stats collection
func BenchmarkCacheStatsCollection(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Populate cache
	for i := 0; i < 1000; i++ {
		memoryCache.Set(ctx, fmt.Sprintf("key_%d", i), []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = memoryCache.GetStats()
	}
}

// BenchmarkCacheSizeCheck benchmarks size checking
func BenchmarkCacheSizeCheck(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	for i := 0; i < 10000; i++ {
		memoryCache.Set(ctx, fmt.Sprintf("key_%d", i), []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = memoryCache.Size()
	}
}

// BenchmarkCachePatternMatch benchmarks pattern matching for deletion
func BenchmarkCachePatternMatch(b *testing.B) {
	pattern := "device:*"
	keys := []string{"device:1", "device:2", "cache:1", "user:1"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			if strings.HasPrefix(key, "device:") {
				_ = key // Match
			}
		}
	}
}

// BenchmarkCacheValueComparison benchmarks value comparison operations
func BenchmarkCacheValueComparison(b *testing.B) {
	value1 := []byte("test_value_12345")
	value2 := []byte("test_value_12345")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Compare values
		equal := string(value1) == string(value2)
		_ = equal
	}
}

// BenchmarkCacheContextTimeout benchmarks context timeout handling
func BenchmarkCacheContextTimeout(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		memoryCache.Set(ctx, "test_key", []byte("test_value"), 5*time.Minute)
		cancel()
	}
}

// BenchmarkCacheHighLoad benchmarks cache under high load
func BenchmarkCacheHighLoad(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Pre-populate
	for i := 0; i < 10000; i++ {
		memoryCache.Set(ctx, fmt.Sprintf("key_%d", i), []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%10000)
			memoryCache.Get(ctx, key)
			i++
		}
	})
}

// BenchmarkCacheWriteHeavy benchmarks write-heavy workload
func BenchmarkCacheWriteHeavy(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("write_key_%d", time.Now().UnixNano())
			memoryCache.Set(ctx, key, []byte("value"), 5*time.Minute)
		}
	})
}

// BenchmarkCacheReadHeavy benchmarks read-heavy workload
func BenchmarkCacheReadHeavy(b *testing.B) {
	memoryCache := NewMemoryCache(5 * time.Minute)
	ctx := context.Background()
	
	// Pre-populate
	for i := 0; i < 1000; i++ {
		memoryCache.Set(ctx, fmt.Sprintf("key_%d", i), []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			memoryCache.Get(ctx, key)
			i++
		}
	})
}