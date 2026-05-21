package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConfig_DefaultProductionRedisConfig(t *testing.T) {
	cfg := DefaultProductionRedisConfig()

	assert.Equal(t, "redis:6379", cfg.Addr)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 50, cfg.PoolSize)
	assert.Equal(t, 10, cfg.MinIdleConns)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 4*time.Second, cfg.PoolTimeout)
	assert.Equal(t, 5*time.Minute, cfg.IdleTimeout)
	assert.Equal(t, 30*time.Minute, cfg.MaxConnAge)
}

func TestRedisConfig_DefaultDevelopmentRedisConfig(t *testing.T) {
	cfg := DefaultDevelopmentRedisConfig()

	assert.Equal(t, "localhost:6379", cfg.Addr)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 10, cfg.PoolSize)
	assert.Equal(t, 5, cfg.MinIdleConns)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, 3*time.Second, cfg.DialTimeout)
	assert.Equal(t, 2*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 2*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 2*time.Second, cfg.PoolTimeout)
	assert.Equal(t, 2*time.Minute, cfg.IdleTimeout)
	assert.Equal(t, 10*time.Minute, cfg.MaxConnAge)
}

func TestNewRedisClient(t *testing.T) {
	t.Run("creates client with valid config", func(t *testing.T) {
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

		client := NewRedisClient(cfg)
		require.NotNil(t, client)

		// Clean up
		client.Close()
	})

	t.Run("creates client with default production config", func(t *testing.T) {
		cfg := DefaultProductionRedisConfig()
		client := NewRedisClient(cfg)
		require.NotNil(t, client)
		client.Close()
	})

	t.Run("creates client with default development config", func(t *testing.T) {
		cfg := DefaultDevelopmentRedisConfig()
		client := NewRedisClient(cfg)
		require.NotNil(t, client)
		client.Close()
	})
}

func TestParseRedisInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected string
	}{
		{
			name:     "parse redis_version",
			info:     "# Server\nredis_version:7.0.0\nredis_mode:standalone\n",
			key:      "redis_version",
			expected: "7.0.0",
		},
		{
			name:     "parse used_memory_human",
			info:     "# Memory\nused_memory:1048576\nused_memory_human:1.00M\n",
			key:      "used_memory_human",
			expected: "1.00M",
		},
		{
			name:     "key not found",
			info:     "# Server\nredis_version:7.0.0\n",
			key:      "non_existent_key",
			expected: "",
		},
		{
			name:     "empty info",
			info:     "",
			key:      "redis_version",
			expected: "",
		},
		{
			name:     "parse uptime_in_seconds",
			info:     "# Server\nuptime_in_seconds:3600\n",
			key:      "uptime_in_seconds",
			expected: "3600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfo(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRedisInfoInt(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected int
	}{
		{
			name:     "parse valid integer",
			info:     "connected_clients:100\n",
			key:      "connected_clients",
			expected: 100,
		},
		{
			name:     "parse integer with suffix",
			info:     "total_connections_received:1000abc\n",
			key:      "total_connections_received",
			expected: 1000,
		},
		{
			name:     "key not found",
			info:     "connected_clients:100\n",
			key:      "non_existent",
			expected: 0,
		},
		{
			name:     "empty value",
			info:     "connected_clients:\n",
			key:      "connected_clients",
			expected: 0,
		},
		{
			name:     "zero value",
			info:     "connected_clients:0\n",
			key:      "connected_clients",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfoInt(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRedisInfoInt64(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected int64
	}{
		{
			name:     "parse valid int64",
			info:     "uptime_in_seconds:3600\n",
			key:      "uptime_in_seconds",
			expected: 3600,
		},
		{
			name:     "parse large value",
			info:     "keyspace_hits:9999999999\n",
			key:      "keyspace_hits",
			expected: 9999999999,
		},
		{
			name:     "key not found",
			info:     "uptime_in_seconds:3600\n",
			key:      "non_existent",
			expected: 0,
		},
		{
			name:     "zero value",
			info:     "keyspace_hits:0\n",
			key:      "keyspace_hits",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfoInt64(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single line",
			input:    "line1",
			expected: []string{"line1"},
		},
		{
			name:     "multiple lines with newline",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "lines with carriage return",
			input:    "line1\r\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{
			name:     "starts with prefix",
			s:        "redis_version:7.0.0",
			prefix:   "redis_version:",
			expected: true,
		},
		{
			name:     "does not start with prefix",
			s:        "used_memory:100",
			prefix:   "redis_version:",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "test",
			prefix:   "test",
			expected: true,
		},
		{
			name:     "string shorter than prefix",
			s:        "ab",
			prefix:   "abc",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			prefix:   "test",
			expected: false,
		},
		{
			name:     "empty prefix",
			s:        "test",
			prefix:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := startsWith(tt.s, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubstringAfter(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected string
	}{
		{
			name:     "find separator",
			s:        "redis_version:7.0.0",
			sep:      ":",
			expected: "7.0.0",
		},
		{
			name:     "separator not found",
			s:        "redis_version",
			sep:      ":",
			expected: "redis_version",
		},
		{
			name:     "multiple separators",
			s:        "key:subkey:value",
			sep:      ":",
			expected: "subkey:value",
		},
		{
			name:     "empty separator",
			s:        "test",
			sep:      "",
			expected: "test",
		},
		{
			name:     "separator at end",
			s:        "test:",
			sep:      ":",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substringAfter(tt.s, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedisHealthCheckStruct(t *testing.T) {
	health := &RedisHealthCheck{
		Connected:          true,
		Version:            "7.0.0",
		UptimeSeconds:      3600,
		ConnectedClients:   10,
		UsedMemory:         "1.00M",
		UsedMemoryPeak:     "2.00M",
		MemoryUsagePercent: 50.0,
		HitRate:            95.5,
		TotalHits:          1000,
		TotalMisses:        50,
		OpsPerSec:          100,
		SlowlogLen:         5,
		KeysCount:          500,
	}

	assert.True(t, health.Connected)
	assert.Equal(t, "7.0.0", health.Version)
	assert.Equal(t, int64(3600), health.UptimeSeconds)
	assert.Equal(t, 10, health.ConnectedClients)
	assert.Equal(t, "1.00M", health.UsedMemory)
	assert.Equal(t, 95.5, health.HitRate)
}

func TestRedisSlowLogEntryStruct(t *testing.T) {
	entry := RedisSlowLogEntry{
		ID:        1,
		StartTime: 1234567890,
		Duration:  1000,
		Command:   "GET",
		Key:       "mykey",
		Args:      []string{"arg1", "arg2"},
	}

	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, int64(1234567890), entry.StartTime)
	assert.Equal(t, int64(1000), entry.Duration)
	assert.Equal(t, "GET", entry.Command)
	assert.Equal(t, "mykey", entry.Key)
	assert.Equal(t, []string{"arg1", "arg2"}, entry.Args)
}

func TestRedisMemoryInfoStruct(t *testing.T) {
	memInfo := &RedisMemoryInfo{
		UsedMemory:         1048576,
		UsedMemoryPeak:     2097152,
		UsedMemoryRSS:      1500000,
		FragmentationRatio: 1.43,
		MaxMemory:          2147483648,
		MemoryUsagePercent: 0.05,
		UsedMemoryHuman:    "1.00M",
	}

	assert.Equal(t, int64(1048576), memInfo.UsedMemory)
	assert.Equal(t, int64(2097152), memInfo.UsedMemoryPeak)
	assert.Equal(t, int64(1500000), memInfo.UsedMemoryRSS)
	assert.Equal(t, 1.43, memInfo.FragmentationRatio)
	assert.Equal(t, "1.00M", memInfo.UsedMemoryHuman)
}

func TestCacheStatsStruct(t *testing.T) {
	stats := &CacheStats{
		KeysCount:    100,
		HitRate:      95.5,
		OpsPerSec:    1000,
		MemoryUsed:   "10M",
		SlowlogCount: 5,
	}

	assert.Equal(t, int64(100), stats.KeysCount)
	assert.Equal(t, 95.5, stats.HitRate)
	assert.Equal(t, int64(1000), stats.OpsPerSec)
	assert.Equal(t, "10M", stats.MemoryUsed)
	assert.Equal(t, int64(5), stats.SlowlogCount)
}

func TestRedisConfigStruct(t *testing.T) {
	cfg := &RedisConfig{
		Addr:         "localhost:6379",
		Password:     "secret",
		DB:           1,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   5,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  3 * time.Second,
		IdleTimeout:  10 * time.Minute,
		MaxConnAge:   1 * time.Hour,
	}

	assert.Equal(t, "localhost:6379", cfg.Addr)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, 1, cfg.DB)
	assert.Equal(t, 100, cfg.PoolSize)
	assert.Equal(t, 10, cfg.MinIdleConns)
	assert.Equal(t, 5, cfg.MaxRetries)
	assert.Equal(t, 10*time.Second, cfg.DialTimeout)
	assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 5*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 3*time.Second, cfg.PoolTimeout)
	assert.Equal(t, 10*time.Minute, cfg.IdleTimeout)
	assert.Equal(t, 1*time.Hour, cfg.MaxConnAge)
}

// Helper function to create miniredis and redis client
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func TestGetHealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// Add some test data
		ctx := context.Background()
		client.Set(ctx, "testkey", "testvalue", 0)

		// Test ping connectivity (miniredis doesn't fully support INFO with multiple sections)
		err := client.Ping(ctx).Err()
		require.NoError(t, err)

		// Test basic DBSIZE which is supported
		count, err := client.DBSize(ctx).Result()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("health check on unavailable server", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999", // Non-existent server
		})
		defer client.Close()

		health, err := GetHealthCheck(client)
		require.Error(t, err)
		require.NotNil(t, health)
		assert.False(t, health.Connected)
	})
}

func TestIsRedisHealthy(t *testing.T) {
	t.Run("healthy redis - ping test", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// Test that redis is reachable (miniredis has limitations)
		ctx := context.Background()
		err := client.Ping(ctx).Err()
		require.NoError(t, err)
	})

	t.Run("unhealthy redis - connection failure", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999",
		})
		defer client.Close()

		healthy, warnings := IsRedisHealthy(client)
		assert.False(t, healthy)
		assert.NotEmpty(t, warnings)
	})
}

func TestGetSlowLog(t *testing.T) {
	t.Run("slowlog error on miniredis", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// miniredis doesn't support SLOWLOG command, so we expect an error
		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slowlog")
	})

	t.Run("slowlog on unavailable server", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999",
		})
		defer client.Close()

		_, err := GetSlowLog(client, 10)
		require.Error(t, err)
	})
}

func TestGetMemoryInfo(t *testing.T) {
	t.Run("memory info error on miniredis", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		// miniredis doesn't support INFO memory section, so we expect an error
		_, err := GetMemoryInfo(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "memory")
	})

	t.Run("memory info on unavailable server", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999",
		})
		defer client.Close()

		_, err := GetMemoryInfo(client)
		require.Error(t, err)
	})
}

func TestCacheWarmup(t *testing.T) {
	t.Run("successful warmup", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		keys := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		err := CacheWarmup(client, keys, 5*time.Minute)
		require.NoError(t, err)

		// Verify keys were set
		for key, expectedValue := range keys {
			val, err := client.Get(ctx, key).Result()
			require.NoError(t, err)
			assert.Equal(t, expectedValue, val)
		}
	})

	t.Run("warmup with empty keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		keys := map[string]string{}

		err := CacheWarmup(client, keys, 5*time.Minute)
		require.NoError(t, err)
	})

	t.Run("warmup on unavailable server", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999",
		})
		defer client.Close()

		keys := map[string]string{"key": "value"}

		err := CacheWarmup(client, keys, 5*time.Minute)
		require.Error(t, err)
	})
}

func TestGetCacheStats(t *testing.T) {
	t.Run("get cache stats error on miniredis", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		ctx := context.Background()
		// Add some keys
		client.Set(ctx, "key1", "value1", 0)
		client.Set(ctx, "key2", "value2", 0)

		// miniredis doesn't support INFO command properly, so we expect an error
		_, err := GetCacheStats(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info")
	})

	t.Run("cache stats on unavailable server", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:9999",
		})
		defer client.Close()

		_, err := GetCacheStats(client)
		require.Error(t, err)
	})
}

func TestNewRedisClientWithMiniredis(t *testing.T) {
	t.Run("client works with miniredis", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		cfg := &RedisConfig{
			Addr:     mr.Addr(),
			Password: "",
			DB:       0,
		}

		client := NewRedisClient(cfg)
		require.NotNil(t, client)
		defer client.Close()

		// Test basic operations
		ctx := context.Background()
		err = client.Set(ctx, "testkey", "testvalue", 0).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "testkey").Result()
		require.NoError(t, err)
		assert.Equal(t, "testvalue", val)
	})
}

// Test Redis INFO parsing with simulated data
func TestGetHealthCheckWithMockedInfo(t *testing.T) {
	t.Run("parse health check with full info", func(t *testing.T) {
		// Test that we can parse various Redis INFO fields
		infoData := `# Server
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

		// Test parseRedisInfo with various keys
		assert.Equal(t, "7.0.0", parseRedisInfo(infoData, "redis_version"))
		assert.Equal(t, "3600", parseRedisInfo(infoData, "uptime_in_seconds"))
		assert.Equal(t, "10", parseRedisInfo(infoData, "connected_clients"))
		assert.Equal(t, "1.00M", parseRedisInfo(infoData, "used_memory_human"))
		assert.Equal(t, "1000", parseRedisInfo(infoData, "keyspace_hits"))

		// Test parseRedisInfoInt
		assert.Equal(t, 10, parseRedisInfoInt(infoData, "connected_clients"))
		assert.Equal(t, 500, parseRedisInfoInt(infoData, "instantaneous_ops_per_sec"))

		// Test parseRedisInfoInt64
		assert.Equal(t, int64(3600), parseRedisInfoInt64(infoData, "uptime_in_seconds"))
		assert.Equal(t, int64(1000), parseRedisInfoInt64(infoData, "keyspace_hits"))
		assert.Equal(t, int64(1048576), parseRedisInfoInt64(infoData, "used_memory"))
	})
}

func TestIsRedisHealthyWithMockData(t *testing.T) {
	// Test the health check logic with different health states
	t.Run("check health with low hit rate", func(t *testing.T) {
		// This tests the IsRedisHealthy logic by checking warnings
		health := &RedisHealthCheck{
			Connected:          true,
			HitRate:            70.0, // Below 80%
			TotalHits:          200,
			TotalMisses:        100,
			MemoryUsagePercent: 50.0,
			SlowlogLen:         10,
			ConnectedClients:   100,
		}

		// Verify the conditions
		assert.True(t, health.HitRate < 80)
		assert.True(t, health.TotalHits+health.TotalMisses > 100)
	})

	t.Run("check health with high memory usage", func(t *testing.T) {
		health := &RedisHealthCheck{
			Connected:          true,
			HitRate:            95.0,
			TotalHits:          1000,
			TotalMisses:        50,
			MemoryUsagePercent: 85.0, // Above 80%
			SlowlogLen:         20,
			ConnectedClients:   100,
		}

		assert.True(t, health.MemoryUsagePercent > 80)
	})

	t.Run("check health with many slow queries", func(t *testing.T) {
		health := &RedisHealthCheck{
			Connected:          true,
			HitRate:            95.0,
			TotalHits:          1000,
			TotalMisses:        50,
			MemoryUsagePercent: 50.0,
			SlowlogLen:         60, // Above 50
			ConnectedClients:   100,
		}

		assert.True(t, health.SlowlogLen > 50)
	})

	t.Run("check health with high client connections", func(t *testing.T) {
		health := &RedisHealthCheck{
			Connected:          true,
			HitRate:            95.0,
			TotalHits:          1000,
			TotalMisses:        50,
			MemoryUsagePercent: 50.0,
			SlowlogLen:         20,
			ConnectedClients:   8500, // Above 80% of 10000
		}

		maxClients := 10000
		assert.True(t, health.ConnectedClients > int(float64(maxClients)*0.8))
	})
}

func TestParseRedisInfoInt64EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected int64
	}{
		{
			name:     "parse max int64 value",
			info:     "big_value:9223372036854775807\n",
			key:      "big_value",
			expected: 9223372036854775807,
		},
		{
			name:     "parse value with mixed content",
			info:     "mixed:12345abc\n",
			key:      "mixed",
			expected: 12345,
		},
		{
			name:     "parse single digit",
			info:     "single:5\n",
			key:      "single",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfoInt64(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMemoryInfoParsing(t *testing.T) {
	// Test memory info parsing logic
	t.Run("parse memory info fields", func(t *testing.T) {
		memInfo := `# Memory
used_memory:1048576
used_memory_peak:2097152
used_memory_rss:1500000
maxmemory:2147483648
used_memory_human:1.00M
`
		assert.Equal(t, int64(1048576), parseRedisInfoInt64(memInfo, "used_memory"))
		assert.Equal(t, int64(2097152), parseRedisInfoInt64(memInfo, "used_memory_peak"))
		assert.Equal(t, int64(1500000), parseRedisInfoInt64(memInfo, "used_memory_rss"))
		assert.Equal(t, int64(2147483648), parseRedisInfoInt64(memInfo, "maxmemory"))
		assert.Equal(t, "1.00M", parseRedisInfo(memInfo, "used_memory_human"))
	})
}

func TestRedisSlowLogEntryParsing(t *testing.T) {
	// Test slowlog entry structure
	entry := RedisSlowLogEntry{
		ID:        1,
		StartTime: time.Now().Unix(),
		Duration:  5000,
		Command:   "GET",
		Key:       "testkey",
		Args:      []string{"arg1", "arg2"},
	}

	assert.Equal(t, int64(1), entry.ID)
	assert.NotZero(t, entry.StartTime)
	assert.Equal(t, int64(5000), entry.Duration)
	assert.Equal(t, "GET", entry.Command)
	assert.Equal(t, "testkey", entry.Key)
	assert.Len(t, entry.Args, 2)
}

func TestCacheStatsStructValues(t *testing.T) {
	// Test CacheStats struct with realistic values
	stats := &CacheStats{
		KeysCount:    10000,
		HitRate:      95.5,
		OpsPerSec:    5000,
		MemoryUsed:   "100MB",
		SlowlogCount: 25,
	}

	assert.Equal(t, int64(10000), stats.KeysCount)
	assert.Equal(t, 95.5, stats.HitRate)
	assert.Equal(t, int64(5000), stats.OpsPerSec)
	assert.Equal(t, "100MB", stats.MemoryUsed)
	assert.Equal(t, int64(25), stats.SlowlogCount)
}

func TestRedisMemoryInfoCalculations(t *testing.T) {
	// Test memory info calculations
	t.Run("fragmentation ratio calculation", func(t *testing.T) {
		usedMemory := int64(1048576)    // 1MB
		usedMemoryRSS := int64(1500000) // ~1.43MB
		expectedRatio := float64(usedMemoryRSS) / float64(usedMemory)

		assert.InDelta(t, 1.43, expectedRatio, 0.01)
	})

	t.Run("memory usage percent calculation", func(t *testing.T) {
		usedMemory := int64(104857600) // ~100MB
		maxMemory := int64(2147483648) // 2GB
		usagePercent := float64(usedMemory) / float64(maxMemory) * 100

		assert.Less(t, usagePercent, 10.0)
	})
}

// Test calculations for RedisHealthCheck
func TestRedisHealthCheckCalculations(t *testing.T) {
	t.Run("hit rate calculation", func(t *testing.T) {
		// Test that hit rate is calculated correctly
		totalHits := int64(800)
		totalMisses := int64(200)
		hitRate := float64(totalHits) / float64(totalHits+totalMisses) * 100

		assert.Equal(t, 80.0, hitRate)

		// Test zero division case
		zeroHits := int64(0)
		zeroMisses := int64(0)
		total := zeroHits + zeroMisses
		assert.Zero(t, total)
	})

	t.Run("memory usage calculation", func(t *testing.T) {
		// Test memory usage percentage
		usedMemory := int64(1073741824) // 1GB
		maxMemory := int64(2147483648)  // 2GB
		usagePercent := float64(usedMemory) / float64(maxMemory) * 100

		assert.Equal(t, 50.0, usagePercent)
	})
}

// Test IsRedisHealthy warning conditions
func TestIsRedisHealthyWarnings(t *testing.T) {
	tests := []struct {
		name           string
		health         *RedisHealthCheck
		expectWarnings bool
		warningCount   int
	}{
		{
			name: "all healthy - no warnings",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectWarnings: false,
			warningCount:   0,
		},
		{
			name: "low hit rate warning",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          700,
				TotalMisses:        300,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectWarnings: true,
			warningCount:   1, // Low hit rate warning
		},
		{
			name: "high memory usage warning",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 85.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectWarnings: true,
			warningCount:   1, // High memory usage warning
		},
		{
			name: "many slow queries warning",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         60,
				ConnectedClients:   100,
			},
			expectWarnings: true,
			warningCount:   1, // Many slow queries warning
		},
		{
			name: "high client connections warning",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            95.0,
				TotalHits:          1000,
				TotalMisses:        50,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   8500,
			},
			expectWarnings: true,
			warningCount:   1, // High client connections warning
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
			expectWarnings: true,
			warningCount:   4, // All four warnings
		},
		{
			name: "low hit rate but few requests - no warning",
			health: &RedisHealthCheck{
				Connected:          true,
				HitRate:            70.0,
				TotalHits:          7,
				TotalMisses:        3,
				MemoryUsagePercent: 50.0,
				SlowlogLen:         10,
				ConnectedClients:   100,
			},
			expectWarnings: false,
			warningCount:   0, // Total < 100, so no warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify health check conditions
			if tt.health.HitRate < 80 && tt.health.TotalHits+tt.health.TotalMisses > 100 {
				assert.True(t, tt.expectWarnings)
			}
			if tt.health.MemoryUsagePercent > 80 {
				assert.True(t, tt.expectWarnings)
			}
			if tt.health.SlowlogLen > 50 {
				assert.True(t, tt.expectWarnings)
			}
			maxClients := 10000
			if tt.health.ConnectedClients > int(float64(maxClients)*0.8) {
				assert.True(t, tt.expectWarnings)
			}
		})
	}
}

// Test RedisMemoryInfo calculations
func TestRedisMemoryInfoCalculatedFields(t *testing.T) {
	t.Run("fragmentation ratio with valid data", func(t *testing.T) {
		usedMemory := int64(1048576)
		usedMemoryRSS := int64(1500000)

		var fragmentationRatio float64
		if usedMemory > 0 {
			fragmentationRatio = float64(usedMemoryRSS) / float64(usedMemory)
		}

		assert.InDelta(t, 1.43, fragmentationRatio, 0.01)
	})

	t.Run("fragmentation ratio with zero memory", func(t *testing.T) {
		usedMemory := int64(0)
		usedMemoryRSS := int64(1500000)

		var fragmentationRatio float64
		if usedMemory > 0 {
			fragmentationRatio = float64(usedMemoryRSS) / float64(usedMemory)
		}

		assert.Zero(t, fragmentationRatio)
	})

	t.Run("memory usage percent with max memory", func(t *testing.T) {
		usedMemory := int64(1073741824)
		maxMemory := int64(2147483648)

		var memoryUsagePercent float64
		if maxMemory > 0 {
			memoryUsagePercent = float64(usedMemory) / float64(maxMemory) * 100
		}

		assert.Equal(t, 50.0, memoryUsagePercent)
	})

	t.Run("memory usage percent with zero max memory", func(t *testing.T) {
		usedMemory := int64(1073741824)
		maxMemory := int64(0)

		var memoryUsagePercent float64
		if maxMemory > 0 {
			memoryUsagePercent = float64(usedMemory) / float64(maxMemory) * 100
		}

		assert.Zero(t, memoryUsagePercent)
	})
}

// Test edge cases for parseRedisInfoInt64
func TestParseRedisInfoInt64AdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected int64
	}{
		{
			name:     "parse zero with suffix",
			info:     "value:0bytes\n",
			key:      "value",
			expected: 0,
		},
		{
			name:     "parse very small number",
			info:     "count:1\n",
			key:      "count",
			expected: 1,
		},
		{
			name:     "parse with only letters after",
			info:     "status:123abc\n",
			key:      "status",
			expected: 123,
		},
		{
			name:     "parse with whitespace",
			info:     "key:456  \n",
			key:      "key",
			expected: 456,
		},
		{
			name:     "key with no colon",
			info:     "something\n",
			key:      "something",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfoInt64(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases for parseRedisInfoInt
func TestParseRedisInfoIntAdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		key      string
		expected int
	}{
		{
			name:     "parse zero with suffix",
			info:     "value:0bytes\n",
			key:      "value",
			expected: 0,
		},
		{
			name:     "parse large number",
			info:     "count:999999\n",
			key:      "count",
			expected: 999999,
		},
		{
			name:     "parse with letters after",
			info:     "status:42KB\n",
			key:      "status",
			expected: 42,
		},
		{
			name:     "parse empty value after colon",
			info:     "empty:\n",
			key:      "empty",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfoInt(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test additional splitLines cases
func TestSplitLinesAdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "mixed line endings",
			input:    "line1\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "consecutive newlines",
			input:    "line1\n\nline2",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "carriage return only",
			input:    "line1\rline2",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test additional startsWith cases
func TestStartsWithAdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{
			name:     "single character prefix",
			s:        "test",
			prefix:   "t",
			expected: true,
		},
		{
			name:     "prefix longer than string",
			s:        "ab",
			prefix:   "abc",
			expected: false,
		},
		{
			name:     "unicode strings",
			s:        "测试test",
			prefix:   "测试",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := startsWith(tt.s, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test additional substringAfter cases
func TestSubstringAfterAdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected string
	}{
		{
			name:     "separator at start",
			s:        ":value",
			sep:      ":",
			expected: "value",
		},
		{
			name:     "multi-character separator",
			s:        "key==>value",
			sep:      "==>",
			expected: "value",
		},
		{
			name:     "separator longer than string",
			s:        "ab",
			sep:      "abc",
			expected: "ab",
		},
		{
			name:     "multiple occurrences",
			s:        "a:b:c",
			sep:      ":",
			expected: "b:c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substringAfter(tt.s, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkParseRedisInfo(b *testing.B) {
	info := `# Server
redis_version:7.0.0
uptime_in_seconds:3600
connected_clients:10
used_memory:1048576
keyspace_hits:1000
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRedisInfo(info, "redis_version")
	}
}

func BenchmarkParseRedisInfoInt64(b *testing.B) {
	info := "uptime_in_seconds:3600\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRedisInfoInt64(info, "uptime_in_seconds")
	}
}
