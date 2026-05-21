package benchmarks

import (
	"sync"
	"testing"
	"time"
)

// 内存缓存基准测试

type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]string
	ttl  map[string]time.Time
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]string),
		ttl:  make(map[string]time.Time),
	}
}

func (c *MemoryCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		return v, true
	}
	return "", false
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// BenchmarkCacheSet 缓存Set基准测试
func BenchmarkCacheSet(b *testing.B) {
	cache := NewMemoryCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}
}

// BenchmarkCacheGet 缓存Get基准测试
func BenchmarkCacheGet(b *testing.B) {
	cache := NewMemoryCache()
	for i := 0; i < 100; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key0")
	}
}

// BenchmarkCacheGetMiss 缓存Get miss基准测试
func BenchmarkCacheGetMiss(b *testing.B) {
	cache := NewMemoryCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("nonexistent")
	}
}

// BenchmarkCacheDelete 缓存Delete基准测试
func BenchmarkCacheDelete(b *testing.B) {
	cache := NewMemoryCache()
	for i := 0; i < b.N; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete("key" + string(rune(i)))
	}
}

// BenchmarkCacheConcurrentRead 并发读取基准测试
func BenchmarkCacheConcurrentRead(b *testing.B) {
	cache := NewMemoryCache()
	for i := 0; i < 100; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("key0")
		}
	})
}

// BenchmarkCacheConcurrentWrite 并发写入基准测试
func BenchmarkCacheConcurrentWrite(b *testing.B) {
	cache := NewMemoryCache()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set("key"+string(rune(i)), "value")
			i++
		}
	})
}

// BenchmarkCacheMixed 混合读写基准测试
func BenchmarkCacheMixed(b *testing.B) {
	cache := NewMemoryCache()
	for i := 0; i < 100; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Get("key0")
			} else {
				cache.Set("key"+string(rune(i)), "value")
			}
			i++
		}
	})
}
