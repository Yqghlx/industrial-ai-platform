package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Blacklist constants
const (
	BlacklistPrefix      = "jwt_blacklist:" // Token 黑名单 Redis key 前缀
	UserRevocationPrefix = "user_revoke:"   // 用户撤销记录前缀
)

// TokenBlacklistInterface Token 黑名单接口
type TokenBlacklistInterface interface {
	Add(ctx context.Context, tokenID string, duration time.Duration) error
	Exists(ctx context.Context, tokenID string) bool
	AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error
	GetUserRevocation(ctx context.Context, userID int) (time.Time, error)
	Stop()
}

// UserTokenStoreInterface 用户 Token 版本存储接口
type UserTokenStoreInterface interface {
	GetTokenVersion(ctx context.Context, userID int) (int, error)
	UpdateTokenVersion(ctx context.Context, userID int) error
}

// RedisTokenBlacklist Redis 实现的 Token 黑名单
type RedisTokenBlacklist struct {
	client *redis.Client
}

func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	return b.client.Set(ctx, BlacklistPrefix+tokenID, "1", duration).Err()
}

func (b *RedisTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	val, err := b.client.Exists(ctx, BlacklistPrefix+tokenID).Result()
	return err == nil && val > 0
}

func (b *RedisTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	key := fmt.Sprintf("%s%d", UserRevocationPrefix, userID)
	return b.client.Set(ctx, key, revokedAt.Format(time.RFC3339), duration).Err()
}

func (b *RedisTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	key := fmt.Sprintf("%s%d", UserRevocationPrefix, userID)
	val, err := b.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (b *RedisTokenBlacklist) Stop() {}

// revocationEntry 用户撤销记录
type revocationEntry struct {
	revokedAt time.Time
	expiresAt time.Time
}

// MemoryTokenBlacklist 内存实现的 Token 黑名单
// SEC-MEDIUM-02: 添加最大条目限制防止内存耗尽
type MemoryTokenBlacklist struct {
	entries         map[string]time.Time
	userRevocations map[int]revocationEntry
	mu              sync.RWMutex
	shutdown        chan struct{}
	maxEntries      int // 最大条目数限制，防止DoS攻击
}

const DefaultMaxEntries = 10000 // 默认最大条目数

func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	bl := &MemoryTokenBlacklist{
		entries:         make(map[string]time.Time),
		userRevocations: make(map[int]revocationEntry),
		shutdown:        make(chan struct{}),
		maxEntries:      DefaultMaxEntries,
	}
	go bl.cleanupExpiredEntries()
	return bl
}

// NewMemoryTokenBlacklistWithLimit 创建带自定义大小限制的黑名单
func NewMemoryTokenBlacklistWithLimit(maxEntries int) *MemoryTokenBlacklist {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxEntries
	}
	bl := &MemoryTokenBlacklist{
		entries:         make(map[string]time.Time),
		userRevocations: make(map[int]revocationEntry),
		shutdown:        make(chan struct{}),
		maxEntries:      maxEntries,
	}
	go bl.cleanupExpiredEntries()
	return bl
}

func (b *MemoryTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// SEC-MEDIUM-02: 检查大小限制，超过时淘汰条目
	// 优化：O(1) 随机淘汰替代 O(n) 全量扫描，避免写锁长时间阻塞
	if len(b.entries) >= b.maxEntries {
		// 快速尝试：随机取一个条目检查是否过期
		for k, expiry := range b.entries {
			if time.Now().After(expiry) {
				delete(b.entries, k)
				goto added
			}
			break // 只检查第一个，O(1)
		}
		// 没有过期条目：随机淘汰一个（map 迭代顺序随机）
		for k := range b.entries {
			delete(b.entries, k)
			break
		}
	}

added:
	b.entries[BlacklistPrefix+tokenID] = time.Now().Add(duration)
	return nil
}

func (b *MemoryTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	expiry, exists := b.entries[BlacklistPrefix+tokenID]
	if !exists {
		return false
	}
	return !time.Now().After(expiry)
}

func (b *MemoryTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.userRevocations[userID] = revocationEntry{
		revokedAt: revokedAt,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

func (b *MemoryTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	entry, exists := b.userRevocations[userID]
	if !exists || time.Now().After(entry.expiresAt) {
		return time.Time{}, nil
	}
	return entry.revokedAt, nil
}

func (b *MemoryTokenBlacklist) Stop() {
	close(b.shutdown)
}

func (b *MemoryTokenBlacklist) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries)
}

func (b *MemoryTokenBlacklist) cleanupExpiredEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			now := time.Now()
			for tokenID, expiry := range b.entries {
				if now.After(expiry) {
					delete(b.entries, tokenID)
				}
			}
			for userID, entry := range b.userRevocations {
				if now.After(entry.expiresAt) {
					delete(b.userRevocations, userID)
				}
			}
			b.mu.Unlock()
		case <-b.shutdown:
			return
		}
	}
}

// HybridTokenBlacklist 混合实现的 Token 黑名单
type HybridTokenBlacklist struct {
	redisBlacklist  *RedisTokenBlacklist
	memoryBlacklist *MemoryTokenBlacklist
	useRedis        bool // protected by mu
	checkInterval   time.Duration
	shutdown        chan struct{}
	mu              sync.RWMutex // protects useRedis field
}

func NewHybridTokenBlacklist(redisClient *redis.Client) *HybridTokenBlacklist {
	hb := &HybridTokenBlacklist{
		memoryBlacklist: NewMemoryTokenBlacklist(),
		checkInterval:   30 * time.Second,
		shutdown:        make(chan struct{}),
	}
	if redisClient != nil {
		hb.redisBlacklist = NewRedisTokenBlacklist(redisClient)
		hb.useRedis = true
		go hb.checkRedisHealth(redisClient)
	}
	return hb
}

func (b *HybridTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	b.memoryBlacklist.Add(ctx, tokenID, duration)
	b.mu.RLock()
	useRedis := b.useRedis
	b.mu.RUnlock()
	if useRedis && b.redisBlacklist != nil {
		err := b.redisBlacklist.Add(ctx, tokenID, duration)
		if err != nil {
			logger.L().Warn("Redis blacklist write failed", zap.Error(err))
			b.mu.Lock()
			b.useRedis = false
			b.mu.Unlock()
		}
	}
	return nil
}

func (b *HybridTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	b.mu.RLock()
	useRedis := b.useRedis
	b.mu.RUnlock()
	if useRedis && b.redisBlacklist != nil && b.redisBlacklist.Exists(ctx, tokenID) {
		return true
	}
	return b.memoryBlacklist.Exists(ctx, tokenID)
}

func (b *HybridTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	b.memoryBlacklist.AddUserRevocation(ctx, userID, revokedAt, duration)
	b.mu.RLock()
	useRedis := b.useRedis
	b.mu.RUnlock()
	if useRedis && b.redisBlacklist != nil {
		b.redisBlacklist.AddUserRevocation(ctx, userID, revokedAt, duration)
	}
	return nil
}

func (b *HybridTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	b.mu.RLock()
	useRedis := b.useRedis
	b.mu.RUnlock()
	if useRedis && b.redisBlacklist != nil {
		revokedAt, err := b.redisBlacklist.GetUserRevocation(ctx, userID)
		if err == nil && !revokedAt.IsZero() {
			return revokedAt, nil
		}
	}
	return b.memoryBlacklist.GetUserRevocation(ctx, userID)
}

func (b *HybridTokenBlacklist) Stop() {
	close(b.shutdown)
	if b.memoryBlacklist != nil {
		b.memoryBlacklist.Stop()
	}
}

func (b *HybridTokenBlacklist) IsUsingRedis() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.useRedis
}

func (b *HybridTokenBlacklist) checkRedisHealth(client *redis.Client) {
	ticker := time.NewTicker(b.checkInterval)
	defer ticker.Stop()
	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			err := client.Ping(ctx).Err()
			b.mu.RLock()
			currentUseRedis := b.useRedis
			b.mu.RUnlock()
			if err != nil && currentUseRedis {
				logger.L().Warn("Redis unavailable", zap.Error(err))
				b.mu.Lock()
				b.useRedis = false
				b.mu.Unlock()
			} else if err == nil && !currentUseRedis {
				logger.L().Info("Redis recovered")
				b.mu.Lock()
				b.useRedis = true
				b.mu.Unlock()
			}
		case <-b.shutdown:
			return
		}
	}
}
