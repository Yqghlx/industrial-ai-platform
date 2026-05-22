package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Redis Connection Error Tests
// ============================================================

func TestRedisConnectionErrors(t *testing.T) {
	t.Run("connection refused error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:         "localhost:9999",
			DialTimeout:  1 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		})
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Test that connection fails
		err := client.Ping(ctx).Err()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection")
	})

	t.Run("GetHealthCheck connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		health, err := GetHealthCheck(client)
		require.Error(t, err)
		require.NotNil(t, health)
		assert.False(t, health.Connected)
		assert.Contains(t, err.Error(), "ping failed")
	})

	t.Run("IsRedisHealthy connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		healthy, warnings := IsRedisHealthy(client)
		assert.False(t, healthy)
		assert.NotEmpty(t, warnings)
		assert.Contains(t, warnings[0], "connection failed")
	})

	t.Run("GetSlowLog connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slowlog")
	})

	t.Run("GetMemoryInfo connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetMemoryInfo(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "memory")
	})

	t.Run("GetCacheStats connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetCacheStats(client)
		require.Error(t, err)
		// GetCacheStats calls GetHealthCheck which returns a ping or info error
		assert.Contains(t, err.Error(), "ping")
	})

	t.Run("CacheWarmup connection error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		keys := map[string]string{"key": "value"}
		err := CacheWarmup(client, keys, 5*time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "warmup failed")
	})
}

// ============================================================
// Set/Get Edge Case Tests
// ============================================================

func TestRedisSetGetEdgeCases(t *testing.T) {
	t.Run("set and get empty value", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		err := client.Set(ctx, "empty_key", "", 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "empty_key").Result()
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("set and get large value", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		largeValue := strings.Repeat("x", 10000)
		err := client.Set(ctx, "large_key", largeValue, 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "large_key").Result()
		require.NoError(t, err)
		assert.Equal(t, largeValue, val)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		_, err := client.Get(ctx, "nonexistent_key").Result()
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("set and get unicode value", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		unicodeValue := "中文测试 日本語 테스트"
		err := client.Set(ctx, "unicode_key", unicodeValue, 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "unicode_key").Result()
		require.NoError(t, err)
		assert.Equal(t, unicodeValue, val)
	})

	t.Run("set with special characters in key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		specialKey := "key:with:colon:and:spaces 123"
		err := client.Set(ctx, specialKey, "value", 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, specialKey).Result()
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		client.Set(ctx, "key", "value1", 0)
		client.Set(ctx, "key", "value2", 0)

		val, err := client.Get(ctx, "key").Result()
		require.NoError(t, err)
		assert.Equal(t, "value2", val)
	})

	t.Run("set multiple keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		for i := 0; i < 100; i++ {
			key := "key:" + string(rune('0'+i%10))
			value := "value:" + string(rune('0'+i%10))
			err := client.Set(ctx, key, value, 0).Err()
			require.NoError(t, err)
		}

		count, err := client.DBSize(ctx).Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(10))
	})

	t.Run("set binary value", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		binaryValue := "\x00\x01\x02\x03\xff"
		err := client.Set(ctx, "binary_key", binaryValue, 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "binary_key").Result()
		require.NoError(t, err)
		assert.Equal(t, binaryValue, val)
	})
}

// ============================================================
// TTL Tests
// ============================================================

func TestRedisTTL(t *testing.T) {
	t.Run("set with TTL and verify expiration", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		err := client.Set(ctx, "ttl_key", "ttl_value", 5*time.Second).Err()
		require.NoError(t, err)

		// Verify key exists immediately
		val, err := client.Get(ctx, "ttl_key").Result()
		require.NoError(t, err)
		assert.Equal(t, "ttl_value", val)

		// Verify TTL is set
		ttl, err := client.TTL(ctx, "ttl_key").Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, 5*time.Second)

		// Fast forward past expiration
		mr.FastForward(6 * time.Second)

		// Verify key has expired
		_, err = client.Get(ctx, "ttl_key").Result()
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("set without TTL (persistent)", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		err := client.Set(ctx, "persistent_key", "persistent_value", 0).Err()
		require.NoError(t, err)

		// Verify TTL is -1 (no expiration)
		ttl, err := client.TTL(ctx, "persistent_key").Result()
		require.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)
	})

	t.Run("set with very short TTL", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		err := client.Set(ctx, "short_ttl_key", "value", 100*time.Millisecond).Err()
		require.NoError(t, err)

		// Verify key exists immediately
		val, err := client.Get(ctx, "short_ttl_key").Result()
		require.NoError(t, err)
		assert.Equal(t, "value", val)

		// Fast forward past expiration
		mr.FastForward(200 * time.Millisecond)

		// Verify key has expired
		_, err = client.Get(ctx, "short_ttl_key").Result()
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("set with long TTL", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		longTTL := 24 * time.Hour
		err := client.Set(ctx, "long_ttl_key", "value", longTTL).Err()
		require.NoError(t, err)

		// Verify TTL is set correctly
		ttl, err := client.TTL(ctx, "long_ttl_key").Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, ttl, 23*time.Hour)
		assert.LessOrEqual(t, ttl, 24*time.Hour)
	})

	t.Run("update TTL on existing key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		client.Set(ctx, "update_ttl_key", "value", 10*time.Second)

		// Update TTL
		err := client.Expire(ctx, "update_ttl_key", 1*time.Minute).Err()
		require.NoError(t, err)

		// Verify TTL is updated
		ttl, err := client.TTL(ctx, "update_ttl_key").Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, ttl, 50*time.Second)
		assert.LessOrEqual(t, ttl, 1*time.Minute)
	})

	t.Run("CacheWarmup with TTL", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		keys := map[string]string{
			"warmup_key1": "value1",
			"warmup_key2": "value2",
			"warmup_key3": "value3",
		}

		ttl := 30 * time.Second
		err := CacheWarmup(client, keys, ttl)
		require.NoError(t, err)

		// Verify all keys exist with TTL
		for key := range keys {
			val, err := client.Get(ctx, key).Result()
			require.NoError(t, err)
			assert.NotEmpty(t, val)

			keyTTL, err := client.TTL(ctx, key).Result()
			require.NoError(t, err)
			assert.GreaterOrEqual(t, keyTTL, time.Duration(0))
			assert.LessOrEqual(t, keyTTL, ttl)
		}

		// Fast forward past TTL
		mr.FastForward(35 * time.Second)

		// Verify all keys have expired
		for key := range keys {
			_, err := client.Get(ctx, key).Result()
			assert.Equal(t, redis.Nil, err)
		}
	})
}

// ============================================================
// Mock Redis Tests (miniredis behavior)
// ============================================================

func TestMiniredisMockBehavior(t *testing.T) {
	t.Run("miniredis supports basic commands", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()

		// SET/GET
		err := client.Set(ctx, "key1", "value1", 0).Err()
		require.NoError(t, err)
		val, err := client.Get(ctx, "key1").Result()
		require.NoError(t, err)
		assert.Equal(t, "value1", val)

		// DEL
		err = client.Del(ctx, "key1").Err()
		require.NoError(t, err)
		_, err = client.Get(ctx, "key1").Result()
		assert.Equal(t, redis.Nil, err)

		// EXISTS
		client.Set(ctx, "exists_key", "value", 0)
		exists, err := client.Exists(ctx, "exists_key").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		exists, err = client.Exists(ctx, "nonexistent").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)

		// INCR
		client.Set(ctx, "counter", "0", 0)
		incr, err := client.Incr(ctx, "counter").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), incr)

		// DBSIZE
		count, err := client.DBSize(ctx).Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("miniredis does not support INFO with sections", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()

		// miniredis INFO command has limitations
		_, err := client.Info(ctx, "server", "memory", "stats").Result()
		// Expect error because miniredis doesn't fully support INFO sections
		assert.Error(t, err)
	})

	t.Run("miniredis does not support SLOWLOG", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// miniredis doesn't support SLOWLOG command
		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slowlog")
	})

	t.Run("miniredis pipeline works", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		pipe := client.Pipeline()
		pipe.Set(ctx, "pipe_key1", "value1", 0)
		pipe.Set(ctx, "pipe_key2", "value2", 0)
		pipe.Set(ctx, "pipe_key3", "value3", 0)

		_, err := pipe.Exec(ctx)
		require.NoError(t, err)

		// Verify all keys were set
		for i := 1; i <= 3; i++ {
			key := "pipe_key" + string(rune('0'+i))
			val, err := client.Get(ctx, key).Result()
			require.NoError(t, err)
			assert.Equal(t, "value"+string(rune('0'+i)), val)
		}
	})

	t.Run("miniredis supports TTL via FastForward", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		client.Set(ctx, "ttl_test", "value", 10*time.Second)

		// Key exists initially
		val, err := client.Get(ctx, "ttl_test").Result()
		require.NoError(t, err)
		assert.Equal(t, "value", val)

		// Fast forward time
		mr.FastForward(15 * time.Second)

		// Key should be expired
		_, err = client.Get(ctx, "ttl_test").Result()
		assert.Equal(t, redis.Nil, err)
	})
}

// ============================================================
// Additional Coverage Tests for Full Function Coverage
// ============================================================

func TestGetHealthCheckFullCoverage(t *testing.T) {
	t.Run("ping failure returns error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		health, err := GetHealthCheck(client)
		require.Error(t, err)
		require.NotNil(t, health)
		assert.False(t, health.Connected)
		assert.Contains(t, err.Error(), "ping failed")
	})

	t.Run("info failure after successful ping", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// miniredis doesn't support INFO with multiple sections
		// so this tests the info failure path
		_, err := GetHealthCheck(client)
		// Expect error because miniredis INFO is limited
		assert.Error(t, err)
	})
}

func TestIsRedisHealthyFullCoverage(t *testing.T) {
	t.Run("connection failure warning", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		healthy, warnings := IsRedisHealthy(client)
		assert.False(t, healthy)
		assert.NotEmpty(t, warnings)
		// Check warning contains "connection failed"
		found := false
		for _, w := range warnings {
			if strings.Contains(w, "connection failed") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestGetSlowLogFullCoverage(t *testing.T) {
	t.Run("slowlog get error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slowlog")
	})

	t.Run("slowlog with miniredis error", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
	})
}

func TestGetMemoryInfoFullCoverage(t *testing.T) {
	t.Run("memory info error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetMemoryInfo(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "memory")
	})
}

func TestGetCacheStatsFullCoverage(t *testing.T) {
	t.Run("cache stats error path", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			DialTimeout: 1 * time.Second,
		})
		defer client.Close()

		_, err := GetCacheStats(client)
		require.Error(t, err)
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkCacheWarmup(b *testing.B) {
	mr, err := miniredis.Run()
	require.NoError(b, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	keys := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CacheWarmup(client, keys, 5*time.Minute)
	}
}

func BenchmarkNewRedisClient(b *testing.B) {
	cfg := &RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolTimeout:  2 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := NewRedisClient(cfg)
		client.Close()
	}
}