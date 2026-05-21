package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data  map[string]*memoryItem
	mu    sync.RWMutex
	ttl   time.Duration
	stats Stats
}

type memoryItem struct {
	value     []byte
	expiredAt time.Time
}

// NewMemoryCache 创建内存缓存 (支持 Config 或 time.Duration)
func NewMemoryCache(cfgOrTtl interface{}) *MemoryCache {
	var ttl time.Duration
	if cfg, ok := cfgOrTtl.(*Config); ok {
		ttl = cfg.DefaultTTL
	} else if d, ok := cfgOrTtl.(time.Duration); ok {
		ttl = d
	} else {
		ttl = 5 * time.Minute
	}

	return &MemoryCache{
		data: make(map[string]*memoryItem),
		ttl:  ttl,
		stats: Stats{
			Available:   true,
			BackendType: "memory",
		},
	}
}

// Get 获取缓存值
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, ErrNotFound
	}

	if time.Now().After(item.expiredAt) {
		c.stats.Misses++
		return nil, ErrNotFound
	}

	c.stats.Hits++
	return item.value, nil
}

// Set 设置缓存值
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

// Delete 删除缓存
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	c.stats.Deletes++
	c.stats.KeysStored = int64(len(c.data))
	return nil
}

// Exists 检查键是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false
	}

	return time.Now().Before(item.expiredAt)
}

// GetTTL 获取键的剩余 TTL
func (c *MemoryCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return 0, ErrNotFound
	}

	remaining := time.Until(item.expiredAt)
	if remaining < 0 {
		return 0, ErrNotFound
	}
	return remaining, nil
}

// IsAvailable 返回缓存是否可用
func (c *MemoryCache) IsAvailable() bool {
	return true
}

// GetStats 返回缓存统计
func (c *MemoryCache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.stats.KeysStored = int64(len(c.data))
	return c.stats
}

// Close 关闭缓存 (内存缓存无需操作)
func (c *MemoryCache) Close() error {
	return nil
}

// DeleteByPattern 模式删除
func (c *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if strings.Contains(pattern, "*") {
			prefix := strings.ReplaceAll(pattern, "*", "")
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

// Cleanup 清理过期条目
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

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
