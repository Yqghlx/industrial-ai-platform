package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that require a real Redis server
// These tests will be skipped if Redis is not available

func getIntegrationRedisClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skip("Redis not available for integration tests")
	}

	return client
}

func TestIntegrationGetHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	// Add some test data
	ctx := context.Background()
	client.Set(ctx, "test_health_key1", "value1", 0)
	client.Set(ctx, "test_health_key2", "value2", 0)
	client.Set(ctx, "test_health_key3", "value3", 0)

	health, err := GetHealthCheck(client)
	require.NoError(t, err)
	require.NotNil(t, health)

	// Verify health check results
	assert.True(t, health.Connected)
	assert.NotEmpty(t, health.Version)
	assert.GreaterOrEqual(t, health.UptimeSeconds, int64(0))
	assert.GreaterOrEqual(t, health.ConnectedClients, 1)
	assert.NotEmpty(t, health.UsedMemory)
	assert.NotEmpty(t, health.UsedMemoryPeak)
	assert.GreaterOrEqual(t, health.KeysCount, int64(3))

	// Clean up
	client.Del(ctx, "test_health_key1", "test_health_key2", "test_health_key3")
}

func TestIntegrationIsRedisHealthy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	healthy, warnings := IsRedisHealthy(client)
	assert.True(t, healthy)
	assert.Empty(t, warnings)
}

func TestIntegrationGetSlowLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	// Get slowlog (may be empty if no slow queries)
	entries, err := GetSlowLog(client, 10)
	require.NoError(t, err)
	// Slowlog may be empty, so just verify no error
	assert.NotNil(t, entries)
}

func TestIntegrationGetMemoryInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	memInfo, err := GetMemoryInfo(client)
	require.NoError(t, err)
	require.NotNil(t, memInfo)

	// Verify memory info results
	assert.GreaterOrEqual(t, memInfo.UsedMemory, int64(0))
	assert.GreaterOrEqual(t, memInfo.UsedMemoryPeak, int64(0))
	assert.GreaterOrEqual(t, memInfo.UsedMemoryRSS, int64(0))
	assert.NotEmpty(t, memInfo.UsedMemoryHuman)
	assert.GreaterOrEqual(t, memInfo.FragmentationRatio, 0.0)
}

func TestIntegrationGetCacheStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	// Add some test data
	ctx := context.Background()
	client.Set(ctx, "test_stats_key1", "value1", 0)
	client.Set(ctx, "test_stats_key2", "value2", 0)
	client.Set(ctx, "test_stats_key3", "value3", 0)

	stats, err := GetCacheStats(client)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify cache stats
	assert.GreaterOrEqual(t, stats.KeysCount, int64(3))
	assert.GreaterOrEqual(t, stats.OpsPerSec, int64(0))
	assert.NotEmpty(t, stats.MemoryUsed)

	// Clean up
	client.Del(ctx, "test_stats_key1", "test_stats_key2", "test_stats_key3")
}

func TestIntegrationCacheWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	keys := map[string]string{
		"test_warmup_key1": "value1",
		"test_warmup_key2": "value2",
		"test_warmup_key3": "value3",
	}

	err := CacheWarmup(client, keys, 5*time.Minute)
	require.NoError(t, err)

	// Verify keys were set
	for key, expectedValue := range keys {
		val, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, expectedValue, val)
	}

	// Clean up
	for key := range keys {
		client.Del(ctx, key)
	}
}

func TestIntegrationSetGetWithTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	// Set with TTL
	err := client.Set(ctx, "test_ttl_key", "test_ttl_value", 10*time.Second).Err()
	require.NoError(t, err)

	// Verify key exists
	val, err := client.Get(ctx, "test_ttl_key").Result()
	require.NoError(t, err)
	assert.Equal(t, "test_ttl_value", val)

	// Verify TTL is set
	ttl, err := client.TTL(ctx, "test_ttl_key").Result()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, ttl, 5*time.Second)
	assert.LessOrEqual(t, ttl, 10*time.Second)

	// Clean up
	client.Del(ctx, "test_ttl_key")
}

func TestIntegrationPipelineOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getIntegrationRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	// Use pipeline for multiple operations
	pipe := client.Pipeline()
	pipe.Set(ctx, "test_pipe_key1", "value1", 0)
	pipe.Set(ctx, "test_pipe_key2", "value2", 0)
	pipe.Set(ctx, "test_pipe_key3", "value3", 0)
	pipe.Get(ctx, "test_pipe_key1")

	_, err := pipe.Exec(ctx)
	require.NoError(t, err)

	// Verify keys were set
	for i := 1; i <= 3; i++ {
		key := "test_pipe_key" + string(rune('0'+i))
		val, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, "value"+string(rune('0'+i)), val)
	}

	// Clean up
	client.Del(ctx, "test_pipe_key1", "test_pipe_key2", "test_pipe_key3")
}
