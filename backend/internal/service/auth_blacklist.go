package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
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
type MemoryTokenBlacklist struct {
	entries         map[string]time.Time
	userRevocations map[int]revocationEntry
	mu              sync.RWMutex
	shutdown        chan struct{}
}

func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	bl := &MemoryTokenBlacklist{
		entries:         make(map[string]time.Time),
		userRevocations: make(map[int]revocationEntry),
		shutdown:        make(chan struct{}),
	}
	go bl.cleanupExpiredEntries()
	return bl
}

func (b *MemoryTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
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
	useRedis        bool          // protected by mu
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
			fmt.Printf("Warning: Redis blacklist write failed: %v\n", err)
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
				fmt.Printf("Warning: Redis unavailable: %v\n", err)
				b.mu.Lock()
				b.useRedis = false
				b.mu.Unlock()
			} else if err == nil && !currentUseRedis {
				fmt.Printf("Info: Redis recovered\n")
				b.mu.Lock()
				b.useRedis = true
				b.mu.Unlock()
			}
		case <-b.shutdown:
			return
		}
	}
}