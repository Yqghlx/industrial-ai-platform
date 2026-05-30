package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
// BE-P2-08: 添加 Context 生命周期管理
type MemoryCache struct {
	data        map[string]*memoryItem
	mu          sync.RWMutex
	ttl         time.Duration
	stats       Stats
	cleanupOnce sync.Once
	stopCleanup chan struct{}

	// BE-P2-08: Context 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
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

	// BE-P2-08: 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	c := &MemoryCache{
		data:        make(map[string]*memoryItem),
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		stats: Stats{
			Available:   true,
			BackendType: "memory",
		},
	}

	// 启动后台清理goroutine
	c.startCleanupRoutine()

	return c
}

// startCleanupRoutine 启动后台清理goroutine
// BE-P2-08: 使用 Context 控制生命周期
func (c *MemoryCache) startCleanupRoutine() {
	c.cleanupOnce.Do(func() {
		// 清理间隔默认为5分钟，或使用TTL的一半（取较小值）
		cleanupInterval := 5 * time.Minute
		if c.ttl > 0 && c.ttl/2 < cleanupInterval {
			cleanupInterval = c.ttl / 2
		}
		if cleanupInterval < time.Minute {
			cleanupInterval = time.Minute
		}

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()

			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					c.Cleanup()
				case <-c.stopCleanup:
					return
				case <-c.ctx.Done():
					// BE-P2-08: Context 取消时也优雅退出
					return
				}
			}
		}()
	})
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
// O(1) 淘汰：超过上限时随机淘汰一个条目，避免写锁长时间阻塞
const cacheMaxEntries = 100000

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}

	// O(1) 淘汰：优先找过期条目，否则随机淘汰（map 迭代顺序随机）
	if len(c.data) >= cacheMaxEntries {
		for k, item := range c.data {
			if time.Now().After(item.expiredAt) {
				delete(c.data, k)
				goto added
			}
			break
		}
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}

added:
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

// Close 关闭缓存并停止清理goroutine
// BE-P2-08: 添加 Context 取消确保优雅退出
func (c *MemoryCache) Close() error {
	// 先取消 context
	if c.cancel != nil {
		c.cancel()
	}
	select {
	case <-c.stopCleanup:
		// 已经关闭
	default:
		close(c.stopCleanup)
	}
	c.wg.Wait()
	return nil
}

// Shutdown 优雅关闭（带超时）
// BE-P2-08: 新增带超时的关闭方法
func (c *MemoryCache) Shutdown(ctx context.Context) error {
	// 取消 context
	if c.cancel != nil {
		c.cancel()
	}
	select {
	case <-c.stopCleanup:
		// 已经关闭
	default:
		close(c.stopCleanup)
	}

	// 等待 goroutine 退出或超时
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
