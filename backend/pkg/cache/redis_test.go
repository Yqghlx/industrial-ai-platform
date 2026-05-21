package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup a miniredis server for testing
func setupMiniRedis(t *testing.T) *miniredis.Miniredis {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")
	return mr
}

func TestRedisCache_NewRedisCache(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		mr := setupMiniRedis(t)
		defer mr.Close()

		cfg := &Config{
			RedisURL:   "redis://" + mr.Addr(),
			DefaultTTL: 5 * time.Minute,
			Prefix:     "test:",
		}

		cache, err := NewRedisCache(cfg)
		require.NoError(t, err)
		require.NotNil(t, cache)
		assert.True(t, cache.available)
		assert.Equal(t, "redis", cache.stats.BackendType)

		cache.Close()
	})

	t.Run("with nil config (default)", func(t *testing.T) {
		mr := setupMiniRedis(t)
		defer mr.Close()

		// Create config with miniredis address
		cfg := DefaultConfig()
		cfg.RedisURL = "redis://" + mr.Addr()

		cache, err := NewRedisCache(cfg)
		require.NoError(t, err)
		require.NotNil(t, cache)

		cache.Close()
	})

	t.Run("with invalid Redis URL", func(t *testing.T) {
		cfg := &Config{
			RedisURL: "invalid-url",
		}

		cache, err := NewRedisCache(cfg)
		assert.Error(t, err)
		assert.Nil(t, cache)
		assert.Contains(t, err.Error(), "failed to parse Redis URL")
	})

	t.Run("with unreachable Redis", func(t *testing.T) {
		cfg := &Config{
			RedisURL: "redis://localhost:9999", // Non-existent port
		}

		cache, err := NewRedisCache(cfg)
		assert.Error(t, err)
		assert.Nil(t, cache)
		assert.Contains(t, err.Error(), "failed to connect to Redis")
	})
}

func TestRedisCache_Set_Get(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("set and get value", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		got, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, got)

		// Check stats
		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats.Hits)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, err := cache.Get(ctx, "non-existent-key")
		assert.ErrorIs(t, err, ErrNotFound)

		stats := cache.GetStats()
		assert.GreaterOrEqual(t, stats.Misses, int64(1))
	})

	t.Run("set with custom TTL", func(t *testing.T) {
		key := "custom-ttl-key"
		value := []byte("value")

		err := cache.Set(ctx, key, value, 2*time.Second)
		assert.NoError(t, err)

		// Should exist immediately
		got, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, got)
	})

	t.Run("set with zero TTL uses default", func(t *testing.T) {
		key := "default-ttl-key"
		value := []byte("value")

		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		// Key should exist
		assert.True(t, cache.Exists(ctx, key))
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		key := "overwrite-key"
		value1 := []byte("value1")
		value2 := []byte("value2")

		_ = cache.Set(ctx, key, value1, 0)
		_ = cache.Set(ctx, key, value2, 0)

		got, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value2, got)
	})
}

func TestRedisCache_Delete(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete-key"
		value := []byte("value")

		_ = cache.Set(ctx, key, value, 0)
		assert.True(t, cache.Exists(ctx, key))

		err := cache.Delete(ctx, key)
		assert.NoError(t, err)

		assert.False(t, cache.Exists(ctx, key))
		_, err = cache.Get(ctx, key)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		err := cache.Delete(ctx, "non-existent-key")
		assert.NoError(t, err) // Redis DEL returns success even for non-existent keys
	})
}

func TestRedisCache_DeleteByPattern(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("delete by wildcard pattern", func(t *testing.T) {
		// Set up keys with prefix
		_ = cache.Set(ctx, "device:1", []byte("v1"), 0)
		_ = cache.Set(ctx, "device:2", []byte("v2"), 0)
		_ = cache.Set(ctx, "device:3", []byte("v3"), 0)
		_ = cache.Set(ctx, "other:1", []byte("v4"), 0)

		// Delete by pattern
		err := cache.DeleteByPattern(ctx, "device:*")
		assert.NoError(t, err)

		// Verify device keys are gone
		assert.False(t, cache.Exists(ctx, "device:1"))
		assert.False(t, cache.Exists(ctx, "device:2"))
		assert.False(t, cache.Exists(ctx, "device:3"))

		// Verify other keys still exist
		assert.True(t, cache.Exists(ctx, "other:1"))
	})

	t.Run("delete exact match when no wildcard", func(t *testing.T) {
		_ = cache.Set(ctx, "exact-key", []byte("v1"), 0)
		_ = cache.Set(ctx, "exact-key-other", []byte("v2"), 0)

		err := cache.DeleteByPattern(ctx, "exact-key")
		assert.NoError(t, err)

		assert.False(t, cache.Exists(ctx, "exact-key"))
		assert.True(t, cache.Exists(ctx, "exact-key-other"))
	})
}

func TestRedisCache_Exists(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("exists returns true for valid key", func(t *testing.T) {
		key := "exists-key"
		value := []byte("value")

		_ = cache.Set(ctx, key, value, 0)
		assert.True(t, cache.Exists(ctx, key))
	})

	t.Run("exists returns false for non-existent key", func(t *testing.T) {
		assert.False(t, cache.Exists(ctx, "non-existent"))
	})
}

func TestRedisCache_GetTTL(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("GetTTL returns remaining time", func(t *testing.T) {
		key := "ttl-key"
		value := []byte("value")

		ttl := 5 * time.Second
		_ = cache.Set(ctx, key, value, ttl)

		remaining, err := cache.GetTTL(ctx, key)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, remaining.Milliseconds(), int64(4000))
		assert.LessOrEqual(t, remaining.Milliseconds(), int64(5000))
	})

	t.Run("GetTTL for non-existent key", func(t *testing.T) {
		_, err := cache.GetTTL(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestRedisCache_IsAvailable(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	t.Run("is available when connected", func(t *testing.T) {
		assert.True(t, cache.IsAvailable())
	})

	t.Run("unavailable after close", func(t *testing.T) {
		cache.Close()
		assert.False(t, cache.IsAvailable())
	})
}

func TestRedisCache_GetStats(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("stats are tracked correctly", func(t *testing.T) {
		// Set
		_ = cache.Set(ctx, "key1", []byte("v1"), 0)
		_ = cache.Set(ctx, "key2", []byte("v2"), 0)

		stats := cache.GetStats()
		assert.Equal(t, int64(2), stats.Sets)
		assert.True(t, stats.Available)
		assert.Equal(t, "redis", stats.BackendType)

		// Get (hit)
		_, _ = cache.Get(ctx, "key1")
		stats = cache.GetStats()
		assert.Equal(t, int64(1), stats.Hits)

		// Get (miss)
		_, _ = cache.Get(ctx, "non-existent")
		stats = cache.GetStats()
		assert.GreaterOrEqual(t, stats.Misses, int64(1))

		// Delete
		_ = cache.Delete(ctx, "key1")
		stats = cache.GetStats()
		assert.GreaterOrEqual(t, stats.Deletes, int64(1))
	})
}

func TestRedisCache_Close(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	t.Run("close returns nil", func(t *testing.T) {
		err := cache.Close()
		assert.NoError(t, err)
		assert.False(t, cache.available)
	})
}

func TestRedisCache_PrefixKey(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	t.Run("prefix is added to keys", func(t *testing.T) {
		cfg := &Config{
			RedisURL:   "redis://" + mr.Addr(),
			DefaultTTL: 5 * time.Minute,
			Prefix:     "iai:",
		}

		cache, err := NewRedisCache(cfg)
		require.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()
		_ = cache.Set(ctx, "mykey", []byte("value"), 0)

		// Check that the key in Redis has the prefix
		assert.True(t, mr.Exists("iai:mykey"))
	})

	t.Run("no duplicate prefix if key already has it", func(t *testing.T) {
		cfg := &Config{
			RedisURL:   "redis://" + mr.Addr(),
			DefaultTTL: 5 * time.Minute,
			Prefix:     "iai:",
		}

		cache, err := NewRedisCache(cfg)
		require.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()
		_ = cache.Set(ctx, "iai:mykey", []byte("value"), 0)

		// Check that the key in Redis doesn't have double prefix
		assert.True(t, mr.Exists("iai:mykey"))
		assert.False(t, mr.Exists("iai:iai:mykey"))
	})

	t.Run("empty prefix", func(t *testing.T) {
		cfg := &Config{
			RedisURL:   "redis://" + mr.Addr(),
			DefaultTTL: 5 * time.Minute,
			Prefix:     "",
		}

		cache, err := NewRedisCache(cfg)
		require.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()
		_ = cache.Set(ctx, "mykey", []byte("value"), 0)

		// Key should be stored without prefix
		assert.True(t, mr.Exists("mykey"))
	})
}

func TestRedisCache_Reconnect(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("reconnect succeeds when Redis is available", func(t *testing.T) {
		err := cache.Reconnect(ctx)
		assert.NoError(t, err)
		assert.True(t, cache.available)
	})
}

func TestRedisCache_DisabledCache(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Mark as unavailable
	cache.available = false

	t.Run("get returns ErrDisabled", func(t *testing.T) {
		_, err := cache.Get(ctx, "key")
		assert.ErrorIs(t, err, ErrDisabled)
	})

	t.Run("set returns ErrDisabled", func(t *testing.T) {
		err := cache.Set(ctx, "key", []byte("value"), 0)
		assert.ErrorIs(t, err, ErrDisabled)
	})

	t.Run("delete returns ErrDisabled", func(t *testing.T) {
		err := cache.Delete(ctx, "key")
		assert.ErrorIs(t, err, ErrDisabled)
	})

	t.Run("exists returns false", func(t *testing.T) {
		assert.False(t, cache.Exists(ctx, "key"))
	})
}

func TestRedisCache_ContextCancellation(t *testing.T) {
	mr := setupMiniRedis(t)
	defer mr.Close()

	cfg := &Config{
		RedisURL:   "redis://" + mr.Addr(),
		DefaultTTL: 5 * time.Minute,
		Prefix:     "test:",
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	defer cache.Close()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("get with cancelled context", func(t *testing.T) {
		_, err := cache.Get(ctx, "key")
		assert.Error(t, err) // Context error
	})

	t.Run("set with cancelled context", func(t *testing.T) {
		err := cache.Set(ctx, "key", []byte("value"), 0)
		assert.Error(t, err) // Context error
	})

	t.Run("delete with cancelled context", func(t *testing.T) {
		err := cache.Delete(ctx, "key")
		assert.Error(t, err) // Context error
	})
}
