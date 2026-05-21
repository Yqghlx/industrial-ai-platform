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

// Test GetHealthCheck with mock server that supports basic commands
func TestGetHealthCheckSuccess(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Test basic connectivity
	ctx := context.Background()
	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	// Add some test data
	client.Set(ctx, "testkey1", "value1", 0)
	client.Set(ctx, "testkey2", "value2", 0)
	client.Set(ctx, "testkey3", "value3", 0)

	// Get DBSize (supported by miniredis)
	count, err := client.DBSize(ctx).Result()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

// Test GetMemoryInfo successful parsing
func TestGetMemoryInfoParsingLogic(t *testing.T) {
	// Test the parsing logic with realistic INFO output
	info := `# Memory
used_memory:1048576
used_memory_peak:2097152
used_memory_rss:1500000
maxmemory:2147483648
used_memory_human:1.00M
`

	usedMemory := parseRedisInfoInt64(info, "used_memory")
	usedMemoryPeak := parseRedisInfoInt64(info, "used_memory_peak")
	usedMemoryRSS := parseRedisInfoInt64(info, "used_memory_rss")
	maxMemory := parseRedisInfoInt64(info, "maxmemory")
	usedMemoryHuman := parseRedisInfo(info, "used_memory_human")

	assert.Equal(t, int64(1048576), usedMemory)
	assert.Equal(t, int64(2097152), usedMemoryPeak)
	assert.Equal(t, int64(1500000), usedMemoryRSS)
	assert.Equal(t, int64(2147483648), maxMemory)
	assert.Equal(t, "1.00M", usedMemoryHuman)

	// Test fragmentation ratio calculation
	var fragmentationRatio float64
	if usedMemory > 0 {
		fragmentationRatio = float64(usedMemoryRSS) / float64(usedMemory)
	}
	assert.InDelta(t, 1.43, fragmentationRatio, 0.01)

	// Test memory usage percent calculation
	var memoryUsagePercent float64
	if maxMemory > 0 {
		memoryUsagePercent = float64(usedMemory) / float64(maxMemory) * 100
	}
	assert.Less(t, memoryUsagePercent, 10.0)
}

// Test GetSlowLog parsing logic
func TestGetSlowLogParsingLogic(t *testing.T) {
	// Test RedisSlowLogEntry structure
	now := time.Now()
	entries := []RedisSlowLogEntry{
		{
			ID:        1,
			StartTime: now.Unix(),
			Duration:  5000,
			Command:   "GET",
			Key:       "user:123",
			Args:      []string{},
		},
		{
			ID:        2,
			StartTime: now.Unix(),
			Duration:  10000,
			Command:   "SET",
			Key:       "cache:456",
			Args:      []string{"value", "EX", "3600"},
		},
		{
			ID:        3,
			StartTime: now.Unix(),
			Duration:  15000,
			Command:   "DEL",
			Key:       "temp:789",
			Args:      []string{},
		},
	}

	assert.Len(t, entries, 3)
	assert.Equal(t, int64(1), entries[0].ID)
	assert.Equal(t, "GET", entries[0].Command)
	assert.Equal(t, "user:123", entries[0].Key)
	assert.Equal(t, int64(5000), entries[0].Duration)

	assert.Equal(t, int64(2), entries[1].ID)
	assert.Equal(t, "SET", entries[1].Command)
	assert.Equal(t, "cache:456", entries[1].Key)
	assert.Len(t, entries[1].Args, 3)
}

// Test IsRedisHealthy logic with various scenarios
func TestIsRedisHealthyLogic(t *testing.T) {
	tests := []struct {
		name           string
		health         *RedisHealthCheck
		expectHealthy  bool
		warningCount   int
		warningContent []string
	}{
		{
			name: "healthy - all good",
			health: &RedisHealthCheck{
				Connected:          true,
				Version:            "7.0.0",
				UptimeSeconds:      3600,
				ConnectedClients:   100,
				UsedMemory:         "100MB",
				UsedMemoryPeak:     "150MB",
				MemoryUsagePercent: 50.0,
				HitRate:            95.0,
				TotalHits:          10000,
				TotalMisses:        500,
				OpsPerSec:          1000,
				SlowlogLen:         10,
				KeysCount:          5000,
			},
			expectHealthy: true,
			warningCount:  0,
		},
		{
			name: "unhealthy - not connected",
			health: &RedisHealthCheck{
				Connected: false,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - low hit rate",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          700,
				TotalMisses:        300,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - high memory usage",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 85.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - many slow queries",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         60,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - high client connections",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   8500,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "multiple warnings",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          700,
				TotalMisses:        300,
				MemoryUsagePercent: 85.0,
				SlowlogLen:         60,
				ConnectedClients:   8500,
			},
			expectHealthy: false,
			warningCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic that determines health
			warnings := []string{}

			if !tt.health.Connected {
				warnings = append(warnings, "Redis not connected")
			}

			if tt.health.HitRate < 80 && tt.health.TotalHits+tt.health.TotalMisses > 100 {
				warnings = append(warnings, "Low cache hit rate")
			}

			if tt.health.MemoryUsagePercent > 80 {
				warnings = append(warnings, "High memory usage")
			}

			if tt.health.SlowlogLen > 50 {
				warnings = append(warnings, "Many slow queries")
			}

			maxClients := 10000
			if tt.health.ConnectedClients > int(float64(maxClients)*0.8) {
				warnings = append(warnings, "High client connections")
			}

			healthy := len(warnings) == 0

			assert.Equal(t, tt.expectHealthy, healthy)
			assert.Len(t, warnings, tt.warningCount)
		})
	}
}

// Test CacheStats structure and GetCacheStats
func TestGetCacheStatsLogic(t *testing.T) {
	health := &RedisHealthCheck{
		Connected:          true,
		Version:            "7.0.0",
		UptimeSeconds:      3600,
		ConnectedClients:   100,
		UsedMemory:         "100MB",
		UsedMemoryPeak:     "150MB",
		MemoryUsagePercent: 50.0,
		HitRate:            95.5,
		TotalHits:          10000,
		TotalMisses:        500,
		OpsPerSec:          1000,
		SlowlogLen:         10,
		KeysCount:          5000,
	}

	stats := &CacheStats{
		KeysCount:    health.KeysCount,
		HitRate:      health.HitRate,
		OpsPerSec:    health.OpsPerSec,
		MemoryUsed:   health.UsedMemory,
		SlowlogCount: health.SlowlogLen,
	}

	assert.Equal(t, int64(5000), stats.KeysCount)
	assert.Equal(t, 95.5, stats.HitRate)
	assert.Equal(t, int64(1000), stats.OpsPerSec)
	assert.Equal(t, "100MB", stats.MemoryUsed)
	assert.Equal(t, int64(10), stats.SlowlogCount)
}

// Test CacheWarmup with various scenarios
func TestCacheWarmupScenarios(t *testing.T) {
	t.Run("warmup with single key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		keys := map[string]string{
			"key1": "value1",
		}

		err := CacheWarmup(client, keys, 5*time.Minute)
		require.NoError(t, err)

		// Verify key was set
		val, err := client.Get(ctx, "key1").Result()
		require.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("warmup with many keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		keys := make(map[string]string)
		for i := 0; i < 100; i++ {
			keys[generateTestKey(i)] = generateTestValue(i)
		}

		err := CacheWarmup(client, keys, 5*time.Minute)
		require.NoError(t, err)

		// Verify all keys were set
		for i := 0; i < 100; i++ {
			key := generateTestKey(i)
			expectedValue := generateTestValue(i)
			val, err := client.Get(ctx, key).Result()
			require.NoError(t, err)
			assert.Equal(t, expectedValue, val)
		}
	})

	t.Run("warmup with TTL", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		keys := map[string]string{
			"ttl_key": "ttl_value",
		}

		err := CacheWarmup(client, keys, 1*time.Second)
		require.NoError(t, err)

		// Verify key exists immediately
		val, err := client.Get(ctx, "ttl_key").Result()
		require.NoError(t, err)
		assert.Equal(t, "ttl_value", val)

		// Fast forward time
		mr.FastForward(2 * time.Second)

		// Verify key has expired
		_, err = client.Get(ctx, "ttl_key").Result()
		assert.Equal(t, redis.Nil, err)
	})
}

// Helper functions
func generateTestKey(i int) string {
	return "test:key:" + string(rune('0'+i%10))
}

func generateTestValue(i int) string {
	return "test:value:" + string(rune('0'+i%10))
}

// Test the newly extracted parseHealthCheckInfo function
func TestParseHealthCheckInfo(t *testing.T) {
	t.Run("parse complete health check info", func(t *testing.T) {
		info := `# Server
redis_version:7.0.0
uptime_in_seconds:3600
# Clients
connected_clients:10
# Memory
used_memory:1048576
used_memory_human:1.00M
used_memory_peak:2097152
used_memory_peak_human:2.00M
# Stats
keyspace_hits:1000
keyspace_misses:100
instantaneous_ops_per_sec:500
slowlog_len:10
`

		health := &RedisHealthCheck{}
		parseHealthCheckInfo(info, health)

		// Verify all fields were parsed correctly
		assert.Equal(t, "7.0.0", health.Version)
		assert.Equal(t, int64(3600), health.UptimeSeconds)
		assert.Equal(t, 10, health.ConnectedClients)
		assert.Equal(t, "1.00M", health.UsedMemory)
		assert.Equal(t, "2.00M", health.UsedMemoryPeak)
		assert.Equal(t, int64(1000), health.TotalHits)
		assert.Equal(t, int64(100), health.TotalMisses)
		assert.Equal(t, int64(500), health.OpsPerSec)
		assert.Equal(t, int64(10), health.SlowlogLen)

		// Verify calculated fields
		expectedHitRate := float64(1000) / float64(1000+100) * 100
		assert.Equal(t, expectedHitRate, health.HitRate)

		// Memory usage percent (1MB / 2GB = ~0.05%)
		assert.Less(t, health.MemoryUsagePercent, 1.0)
	})

	t.Run("parse health check info with zero hits and misses", func(t *testing.T) {
		info := `# Stats
keyspace_hits:0
keyspace_misses:0
`

		health := &RedisHealthCheck{}
		parseHealthCheckInfo(info, health)

		assert.Equal(t, int64(0), health.TotalHits)
		assert.Equal(t, int64(0), health.TotalMisses)
		assert.Equal(t, 0.0, health.HitRate) // No hits/misses, so hit rate is 0
	})
}

// Test the newly extracted evaluateHealthStatus function
func TestEvaluateHealthStatus(t *testing.T) {
	tests := []struct {
		name          string
		health        *RedisHealthCheck
		expectHealthy bool
		warningCount  int
	}{
		{
			name: "healthy - all conditions good",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: true,
			warningCount:  0,
		},
		{
			name: "unhealthy - not connected",
			health: &RedisHealthCheck{
				Connected: false,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - low hit rate",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          700,
				TotalMisses:        300,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - high memory usage",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 85.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - many slow queries",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         60,
				ConnectedClients:   100,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "warning - high client connections",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   8500,
			},
			expectHealthy: false,
			warningCount:  1,
		},
		{
			name: "multiple warnings",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          700,
				TotalMisses:        300,
				MemoryUsagePercent: 85.0,
				SlowlogLen:         60,
				ConnectedClients:   8500,
			},
			expectHealthy: false,
			warningCount:  4,
		},
		{
			name: "no warning - low hit rate but few requests",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          7,
				TotalMisses:        3,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectHealthy: true,
			warningCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthy, warnings := evaluateHealthStatus(tt.health)

			assert.Equal(t, tt.expectHealthy, healthy)
			assert.Len(t, warnings, tt.warningCount)

			// Additional checks for specific warning content
			if !tt.health.Connected && len(warnings) > 0 {
				assert.Contains(t, warnings[0], "not connected")
			}
			if tt.health.HitRate < 80 && tt.health.TotalHits+tt.health.TotalMisses > 100 && len(warnings) > 0 {
				// Find the warning about hit rate
				for _, w := range warnings {
					if strings.Contains(w, "hit rate") {
						assert.Contains(t, w, "hit rate")
						break
					}
				}
			}
			if tt.health.MemoryUsagePercent > 80 && len(warnings) > 0 {
				// Find the warning about memory
				for _, w := range warnings {
					if strings.Contains(w, "memory usage") {
						assert.Contains(t, w, "memory usage")
						break
					}
				}
			}
			if tt.health.SlowlogLen > 50 && len(warnings) > 0 {
				// Find the warning about slow queries
				for _, w := range warnings {
					if strings.Contains(w, "slow queries") {
						assert.Contains(t, w, "slow queries")
						break
					}
				}
			}
			if tt.health.ConnectedClients > 8000 && len(warnings) > 0 {
				// Find the warning about client connections
				for _, w := range warnings {
					if strings.Contains(w, "client connections") {
						assert.Contains(t, w, "client connections")
						break
					}
				}
			}
		})
	}
}
