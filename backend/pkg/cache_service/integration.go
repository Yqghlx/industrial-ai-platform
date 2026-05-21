// Package cache_service provides integration layer for caching in business services
package cache_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// CacheServiceIntegration provides caching helpers for business services
type CacheServiceIntegration struct {
	cache  cache.CacheService
	warmup *cache.WarmupService
}

// NewCacheServiceIntegration creates a new cache integration
func NewCacheServiceIntegration(cfg *cache.Config) *CacheServiceIntegration {
	cacheSvc := cache.New(cfg)
	warmupSvc := cache.NewWarmupService(cacheSvc)

	return &CacheServiceIntegration{
		cache:  cacheSvc,
		warmup: warmupSvc,
	}
}

// GetCache returns the underlying cache service
func (csi *CacheServiceIntegration) GetCache() cache.CacheService {
	return csi.cache
}

// GetWarmup returns the warmup service
func (csi *CacheServiceIntegration) GetWarmup() *cache.WarmupService {
	return csi.warmup
}

// Device caching helpers

// GetDeviceList retrieves device list from cache or database
func (csi *CacheServiceIntegration) GetDeviceList(ctx context.Context, page, pageSize int, loader func() ([]interface{}, int, error)) ([]interface{}, int, error) {
	key := cache.DeviceCachePrefix.Build("list", fmt.Sprintf("page%d_size%d", page, pageSize))

	var result struct {
		Data  []interface{} `json:"data"`
		Total int           `json:"total"`
	}

	err := cache.GetOrSetJSON(ctx, csi.cache, key,
		func() (interface{}, error) {
			data, total, err := loader()
			if err != nil {
				return nil, err
			}
			return struct {
				Data  []interface{} `json:"data"`
				Total int           `json:"total"`
			}{Data: data, Total: total}, nil
		},
		cache.DeviceListTTL,
		&result,
	)

	return result.Data, result.Total, err
}

// InvalidateDeviceCache clears device-related cache
func (csi *CacheServiceIntegration) InvalidateDeviceCache(ctx context.Context, deviceID string) {
	// Clear device list cache
	_ = csi.cache.DeleteByPattern(ctx, "device:list*")

	// Clear specific device cache
	if deviceID != "" {
		_ = csi.cache.Delete(ctx, cache.DeviceCachePrefix.Build(deviceID))
		_ = csi.cache.Delete(ctx, cache.DeviceCachePrefix.Build(deviceID, "stats"))
		_ = csi.cache.Delete(ctx, cache.DeviceCachePrefix.Build(deviceID, "telemetry"))
	}

	logger.L().Debug("Device cache invalidated",
		zap.String("device_id", deviceID),
	)
}

// ROI caching helpers

// GetROIStats retrieves ROI statistics from cache or database
func (csi *CacheServiceIntegration) GetROIStats(ctx context.Context, loader func() (interface{}, error)) (interface{}, error) {
	key := cache.ROICachePrefix.Build("stats")

	var result interface{}
	err := cache.GetOrSetJSON(ctx, csi.cache, key,
		loader,
		cache.ROIStatsTTL,
		&result,
	)

	return result, err
}

// InvalidateROICache clears ROI statistics cache
func (csi *CacheServiceIntegration) InvalidateROICache(ctx context.Context) {
	_ = csi.cache.DeleteByPattern(ctx, "roi:*")
	logger.L().Debug("ROI cache invalidated")
}

// Alert caching helpers

// GetAlertStats retrieves alert statistics from cache or database
func (csi *CacheServiceIntegration) GetAlertStats(ctx context.Context, period string, loader func() (interface{}, error)) (interface{}, error) {
	key := cache.AlertCachePrefix.Build("stats", period)

	var result interface{}
	err := cache.GetOrSetJSON(ctx, csi.cache, key,
		loader,
		cache.AlertStatsTTL,
		&result,
	)

	return result, err
}

// InvalidateAlertCache clears alert statistics cache
func (csi *CacheServiceIntegration) InvalidateAlertCache(ctx context.Context) {
	_ = csi.cache.DeleteByPattern(ctx, "alert:*")
	logger.L().Debug("Alert cache invalidated")
}

// Telemetry caching helpers

// GetLatestTelemetry retrieves latest telemetry from cache or database
func (csi *CacheServiceIntegration) GetLatestTelemetry(ctx context.Context, loader func() ([]interface{}, error)) ([]interface{}, error) {
	key := cache.TelemetryCachePrefix.Build("latest")

	var result []interface{}
	err := cache.GetOrSetJSON(ctx, csi.cache, key,
		func() (interface{}, error) {
			data, err := loader()
			if err != nil {
				return nil, err
			}
			return data, nil
		},
		cache.TelemetryLatestTTL,
		&result,
	)

	return result, err
}

// InvalidateTelemetryCache clears telemetry cache
func (csi *CacheServiceIntegration) InvalidateTelemetryCache(ctx context.Context, deviceID string) {
	if deviceID != "" {
		_ = csi.cache.Delete(ctx, cache.TelemetryCachePrefix.Build(deviceID))
		_ = csi.cache.Delete(ctx, cache.TelemetryCachePrefix.Build("latest"))
	} else {
		_ = csi.cache.DeleteByPattern(ctx, "telemetry:*")
	}
	logger.L().Debug("Telemetry cache invalidated",
		zap.String("device_id", deviceID),
	)
}

// Generic caching helpers

// GetJSON retrieves JSON data from cache or loader
func (csi *CacheServiceIntegration) GetJSON(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration, result interface{}) error {
	return cache.GetOrSetJSON(ctx, csi.cache, key, loader, ttl, result)
}

// GetRaw retrieves raw bytes from cache or loader
func (csi *CacheServiceIntegration) GetRaw(ctx context.Context, key string, loader func() ([]byte, error), ttl time.Duration) ([]byte, error) {
	return cache.GetOrSet(ctx, csi.cache, key, loader, ttl)
}

// Set stores data in cache
func (csi *CacheServiceIntegration) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return csi.cache.Set(ctx, key, data, ttl)
}

// Delete removes a key from cache
func (csi *CacheServiceIntegration) Delete(ctx context.Context, key string) error {
	return csi.cache.Delete(ctx, key)
}

// DeleteByPattern removes keys matching pattern
func (csi *CacheServiceIntegration) DeleteByPattern(ctx context.Context, pattern string) error {
	return csi.cache.DeleteByPattern(ctx, pattern)
}

// Stats returns cache statistics
func (csi *CacheServiceIntegration) Stats() cache.Stats {
	return csi.cache.GetStats()
}

// IsAvailable checks if cache is available
func (csi *CacheServiceIntegration) IsAvailable() bool {
	return csi.cache.IsAvailable()
}

// Warmup executes cache warmup
func (csi *CacheServiceIntegration) Warmup(ctx context.Context) error {
	return csi.warmup.Warmup(ctx)
}

// WarmupAsync executes async cache warmup
func (csi *CacheServiceIntegration) WarmupAsync() {
	csi.warmup.WarmupAsync()
}

// Close closes the cache service
func (csi *CacheServiceIntegration) Close() error {
	return csi.cache.Close()
}

// GetCacheHealth returns cache health status
func (csi *CacheServiceIntegration) GetCacheHealth() map[string]interface{} {
	stats := csi.cache.GetStats()
	hitRate := 0.0
	if stats.Hits+stats.Misses > 0 {
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	}

	return map[string]interface{}{
		"available":    stats.Available,
		"backend_type": stats.BackendType,
		"hit_rate":     hitRate,
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"keys_stored":  stats.KeysStored,
		"sets":         stats.Sets,
		"deletes":      stats.Deletes,
		"errors":       stats.Errors,
		"last_error":   stats.LastError,
	}
}
