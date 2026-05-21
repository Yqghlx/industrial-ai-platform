package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements CacheService using Redis
type RedisCache struct {
	client    *redis.Client
	config    *Config
	stats     Stats
	available bool
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg *Config) (*RedisCache, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Parse Redis URL
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	rc := &RedisCache{
		client:    client,
		config:    cfg,
		available: true,
		stats: Stats{
			BackendType: "redis",
			Available:   true,
		},
	}

	log.Printf("[Cache] Redis cache connected to %s", opts.Addr)
	return rc, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.available {
		return nil, ErrDisabled
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fullKey := c.prefixKey(key)

	result, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.stats.Misses++
			return nil, ErrNotFound
		}
		c.stats.Errors++
		c.stats.LastError = err.Error()
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	c.stats.Hits++
	return result, nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.available {
		return ErrDisabled
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	fullKey := c.prefixKey(key)

	err := c.client.Set(ctx, fullKey, value, ttl).Err()
	if err != nil {
		c.stats.Errors++
		c.stats.LastError = err.Error()
		return fmt.Errorf("redis set error: %w", err)
	}

	c.stats.Sets++
	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.available {
		return ErrDisabled
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	fullKey := c.prefixKey(key)

	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		c.stats.Errors++
		c.stats.LastError = err.Error()
		return fmt.Errorf("redis del error: %w", err)
	}

	c.stats.Deletes++
	return nil
}

// DeleteByPattern removes keys matching a pattern
func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if !c.available {
		return ErrDisabled
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	fullPattern := c.prefixKey(pattern)

	// Use SCAN to find matching keys (safer than KEYS for production)
	var cursor uint64
	var deletedCount int64

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			c.stats.Errors++
			c.stats.LastError = err.Error()
			return fmt.Errorf("redis scan error: %w", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				c.stats.Errors++
				c.stats.LastError = err.Error()
				return fmt.Errorf("redis del error: %w", err)
			}
			deletedCount += int64(len(keys))
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	c.stats.Deletes += deletedCount
	return nil
}

// Exists checks if a key exists
func (c *RedisCache) Exists(ctx context.Context, key string) bool {
	if !c.available {
		return false
	}

	fullKey := c.prefixKey(key)

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		c.stats.Errors++
		c.stats.LastError = err.Error()
		return false
	}

	return count > 0
}

// GetTTL returns remaining TTL
func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if !c.available {
		return 0, ErrDisabled
	}

	fullKey := c.prefixKey(key)

	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		c.stats.Errors++
		c.stats.LastError = err.Error()
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}

	if ttl < 0 {
		return 0, ErrNotFound
	}

	return ttl, nil
}

// IsAvailable returns whether Redis is available
func (c *RedisCache) IsAvailable() bool {
	if !c.available {
		return false
	}

	// Periodic health check
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		c.available = false
		c.stats.Available = false
		c.stats.LastError = err.Error()
		return false
	}

	return true
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats() Stats {
	stats := c.stats
	stats.Available = c.available

	// Get additional Redis stats
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if c.available {
		// Get key count using DBSIZE
		if keyCount, err := c.client.DBSize(ctx).Result(); err == nil {
			stats.KeysStored = keyCount
		}
	}

	return stats
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	c.available = false
	return c.client.Close()
}

// prefixKey adds the configured prefix to a key
func (c *RedisCache) prefixKey(key string) string {
	if c.config.Prefix != "" && !strings.HasPrefix(key, c.config.Prefix) {
		return c.config.Prefix + key
	}
	return key
}

// Reconnect attempts to reconnect to Redis
func (c *RedisCache) Reconnect(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		c.available = false
		c.stats.Available = false
		c.stats.LastError = err.Error()
		return err
	}

	c.available = true
	c.stats.Available = true
	return nil
}
