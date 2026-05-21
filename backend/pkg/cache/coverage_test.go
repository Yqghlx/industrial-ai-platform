package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// NoopCache Tests
// ============================================

func TestNoopCache_Get(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	data, err := cache.Get(ctx, "test-key")
	assert.Nil(t, data)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestNoopCache_Set(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	err := cache.Set(ctx, "test-key", []byte("value"), time.Minute)
	assert.NoError(t, err)
}

func TestNoopCache_Delete(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	err := cache.Delete(ctx, "test-key")
	assert.NoError(t, err)
}

func TestNoopCache_DeleteByPattern(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	err := cache.DeleteByPattern(ctx, "test:*")
	assert.NoError(t, err)
}

func TestNoopCache_Exists(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	assert.False(t, cache.Exists(ctx, "test-key"))
}

func TestNoopCache_GetTTL(t *testing.T) {
	cache := NewNoopCache()
	ctx := context.Background()

	ttl, err := cache.GetTTL(ctx, "test-key")
	assert.Error(t, err)
	assert.Equal(t, time.Duration(0), ttl)
}

func TestNoopCache_IsAvailable(t *testing.T) {
	cache := NewNoopCache()
	assert.False(t, cache.IsAvailable())
}

func TestNoopCache_GetStats(t *testing.T) {
	cache := NewNoopCache()
	stats := cache.GetStats()
	assert.NotNil(t, stats)
}

func TestNoopCache_Close(t *testing.T) {
	cache := NewNoopCache()
	assert.NoError(t, cache.Close())
}

// ============================================
// WarmupService Tests
// ============================================

func TestNewWarmupService(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)
	assert.NotNil(t, ws)
}

func TestWarmupService_RegisterLoader(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)
	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return cache.Set(ctx, "warmup:1", []byte("value1"), time.Minute)
	})
	assert.Len(t, ws.loaders, 1)
}

func TestWarmupService_RegisterConfigLoader(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)
	ws.RegisterConfigLoader(WarmupConfig{
		Key: "warmup:config",
		Load: func(ctx context.Context) ([]byte, error) {
			return []byte("config-data"), nil
		},
		TTL: time.Minute,
	})
	assert.Len(t, ws.loaders, 1)
}

func TestWarmupService_Warmup_Success(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return cache.Set(ctx, "warmup:test", []byte("data"), time.Minute)
	})

	ctx := context.Background()
	err := ws.Warmup(ctx)
	assert.NoError(t, err)

	data, err := cache.Get(ctx, "warmup:test")
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	assert.False(t, ws.GetWarmupTime().IsZero())
	assert.True(t, ws.IsWarmupComplete())
}

func TestWarmupService_Warmup_CacheNotAvailable(t *testing.T) {
	cache := NewNoopCache()
	ws := NewWarmupService(cache)

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return cache.Set(ctx, "warmup:test", []byte("data"), time.Minute)
	})

	ctx := context.Background()
	err := ws.Warmup(ctx)
	assert.NoError(t, err)
	assert.True(t, ws.warmupAt.IsZero())
}

func TestWarmupService_Warmup_LoaderError(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return assert.AnError
	})

	ctx := context.Background()
	err := ws.Warmup(ctx)
	assert.Error(t, err)
}

func TestWarmupService_Warmup_ContextCancelled(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ws.Warmup(ctx)
	_ = err
}

func TestWarmupService_WarmupAsync(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return cache.Set(ctx, "warmup:async", []byte("async-data"), time.Minute)
	})

	ws.WarmupAsync()
	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	data, err := cache.Get(ctx, "warmup:async")
	require.NoError(t, err)
	assert.Equal(t, []byte("async-data"), data)
}

func TestWarmupService_GetWarmupTime(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)
	assert.True(t, ws.GetWarmupTime().IsZero())

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error { return nil })
	ws.Warmup(context.Background())
	assert.False(t, ws.GetWarmupTime().IsZero())
}

func TestWarmupService_IsWarmupComplete(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)
	assert.False(t, ws.IsWarmupComplete())

	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error { return nil })
	ws.Warmup(context.Background())
	assert.True(t, ws.IsWarmupComplete())
}

func TestWarmupService_ScheduleWarmup(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	// Register loader that sets a cache key to verify warmup ran
	ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		return cache.Set(ctx, "warmup:scheduled", []byte("done"), time.Minute)
	})

	stopChan := ws.ScheduleWarmup(30 * time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	close(stopChan)

	// Verify warmup ran by checking cache
	ctx := context.Background()
	data, err := cache.Get(ctx, "warmup:scheduled")
	require.NoError(t, err)
	assert.Equal(t, []byte("done"), data)
}

func TestWarmupService_ConcurrentLoaders(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ws := NewWarmupService(cache)

	for i := 0; i < 10; i++ {
		idx := i
		ws.RegisterLoader(func(ctx context.Context, cache CacheService) error {
			return cache.Set(ctx, "warmup:concurrent"+string(rune(idx)), []byte("data"), time.Minute)
		})
	}

	ctx := context.Background()
	err := ws.Warmup(ctx)
	assert.NoError(t, err)
	assert.True(t, ws.IsWarmupComplete())
}

// ============================================
// Cache New Factory Tests
// ============================================

func TestNew_Disabled(t *testing.T) {
	cfg := &Config{Enabled: false}
	cache := New(cfg)
	assert.NotNil(t, cache)
	assert.False(t, cache.IsAvailable())
}

func TestNew_NilConfig(t *testing.T) {
	cache := New(nil)
	assert.NotNil(t, cache)
	assert.True(t, cache.IsAvailable())
}

// ============================================
// GetOrSetJSON Tests
// ============================================

type TestCacheStruct struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func TestGetOrSetJSON_CacheMiss(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ctx := context.Background()

	var result TestCacheStruct
	err := GetOrSetJSON(ctx, cache, "json-key", func() (interface{}, error) {
		return TestCacheStruct{Name: "Test", ID: 42}, nil
	}, time.Minute, &result)

	require.NoError(t, err)
	assert.Equal(t, "Test", result.Name)
	assert.Equal(t, 42, result.ID)

	// Should now be cached
	var result2 TestCacheStruct
	err = GetOrSetJSON(ctx, cache, "json-key", func() (interface{}, error) {
		return TestCacheStruct{Name: "Wrong", ID: 0}, nil
	}, time.Minute, &result2)

	require.NoError(t, err)
	assert.Equal(t, "Test", result2.Name)
}

func TestGetOrSetJSON_LoaderError(t *testing.T) {
	cache := NewMemoryCache(DefaultConfig())
	ctx := context.Background()

	var result TestCacheStruct
	err := GetOrSetJSON(ctx, cache, "error-json", func() (interface{}, error) {
		return nil, assert.AnError
	}, time.Minute, &result)

	assert.Error(t, err)
}

// ============================================
// SetTTLs Tests
// ============================================

func TestSetTTLs(t *testing.T) {
	origDevice := DeviceListTTL
	origROI := ROIStatsTTL
	origAlert := AlertStatsTTL
	defer func() {
		DeviceListTTL = origDevice
		ROIStatsTTL = origROI
		AlertStatsTTL = origAlert
	}()

	SetTTLs(1*time.Minute, 2*time.Minute, 30*time.Second)
	assert.Equal(t, 1*time.Minute, DeviceListTTL)
	assert.Equal(t, 2*time.Minute, ROIStatsTTL)
	assert.Equal(t, 30*time.Second, AlertStatsTTL)
}

// ============================================
// MarshalJSON/UnmarshalJSON Tests
// ============================================

func TestMarshalJSON_Helpers(t *testing.T) {
	data := map[string]string{"key": "value"}
	bytes, err := MarshalJSON(data)
	require.NoError(t, err)
	assert.NotEmpty(t, bytes)

	var result map[string]string
	err = UnmarshalJSON([]byte(`{"key":"value"}`), &result)
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}
