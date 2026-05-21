// Package cache provides a unified caching layer supporting Redis and in-memory fallback.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
)

var (
	// ErrNotFound indicates the key was not found in cache
	ErrNotFound = errors.New("cache: key not found")
	// ErrDisabled indicates caching is disabled
	ErrDisabled = errors.New("cache: disabled")
)

// CacheService provides caching operations
type CacheService interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error
	// DeleteByPattern removes all keys matching a pattern
	DeleteByPattern(ctx context.Context, pattern string) error
	// Exists checks if a key exists
	Exists(ctx context.Context, key string) bool
	// GetTTL returns the remaining TTL for a key
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	// IsAvailable returns whether the cache backend is available
	IsAvailable() bool
	// GetStats returns cache statistics
	GetStats() Stats
	// Close closes the cache connection
	Close() error
}

// Stats holds cache statistics
type Stats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	Errors      int64
	KeysStored  int64
	LastError   string
	Available   bool
	BackendType string
}

// CacheKeyBuilder helps build cache keys with prefixes
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a key builder with prefix
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build creates a cache key
func (b *CacheKeyBuilder) Build(parts ...string) string {
	key := b.prefix
	for _, part := range parts {
		if part != "" {
			key += ":" + part
		}
	}
	return key
}

// Common cache key prefixes
var (
	// DeviceCachePrefix for device-related cache
	DeviceCachePrefix = NewCacheKeyBuilder("device")
	// ROICachePrefix for ROI statistics cache
	ROICachePrefix = NewCacheKeyBuilder("roi")
	// AlertCachePrefix for alert statistics cache
	AlertCachePrefix = NewCacheKeyBuilder("alert")
	// TelemetryCachePrefix for telemetry cache
	TelemetryCachePrefix = NewCacheKeyBuilder("telemetry")
)

// Config holds cache configuration
type Config struct {
	// RedisURL is the Redis connection URL (optional)
	RedisURL string
	// Enabled determines if caching is enabled
	Enabled bool
	// DefaultTTL is the default TTL for cached items
	DefaultTTL time.Duration
	// MaxMemorySize is the max size for in-memory cache (in bytes)
	MaxMemorySize int64
	// Prefix is applied to all cache keys
	Prefix string
}

// DefaultConfig returns default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		DefaultTTL:    5 * time.Minute,
		MaxMemorySize: 100 * 1024 * 1024, // 100MB
		Prefix:        "iai:",
	}
}

// New creates a new CacheService based on configuration
func New(cfg *Config) CacheService {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if !cfg.Enabled {
		log.Println("[Cache] Cache is disabled, using noop cache")
		return NewNoopCache()
	}

	// Try Redis first if URL is provided
	if cfg.RedisURL != "" {
		redisCache, err := NewRedisCache(cfg)
		if err != nil {
			log.Printf("[Cache] Redis connection failed: %v, falling back to memory cache", err)
			return NewMemoryCache(cfg)
		}
		log.Println("[Cache] Redis cache initialized successfully")
		return redisCache
	}

	log.Println("[Cache] No Redis URL provided, using memory cache")
	return NewMemoryCache(cfg)
}

// Helper functions for serialization

// MarshalJSON marshals a value to JSON bytes
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals JSON bytes to a value
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GetOrSet retrieves from cache or sets the value from the loader function
func GetOrSet(ctx context.Context, cache CacheService, key string, loader func() ([]byte, error), ttl time.Duration) ([]byte, error) {
	// Try to get from cache
	data, err := cache.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	// If cache is disabled or unavailable, just run loader
	if errors.Is(err, ErrDisabled) || !cache.IsAvailable() {
		return loader()
	}

	// Key not found, load and cache
	data, err = loader()
	if err != nil {
		return nil, err
	}

	// Cache the result (ignore errors, cache failures are non-critical)
	_ = cache.Set(ctx, key, data, ttl)

	return data, nil
}

// GetOrSetJSON retrieves JSON from cache or loads and caches it
func GetOrSetJSON(ctx context.Context, cache CacheService, key string, loader func() (interface{}, error), ttl time.Duration, result interface{}) error {
	data, err := GetOrSet(ctx, cache, key, func() ([]byte, error) {
		v, err := loader()
		if err != nil {
			return nil, err
		}
		return MarshalJSON(v)
	}, ttl)
	if err != nil {
		return err
	}

	return UnmarshalJSON(data, result)
}

// TTL constants for different cache types
var (
	// DeviceListTTL - 5 minutes for device list
	DeviceListTTL = 5 * time.Minute
	// ROIStatsTTL - 10 minutes for ROI statistics
	ROIStatsTTL = 10 * time.Minute
	// AlertStatsTTL - 1 minute for alert statistics
	AlertStatsTTL = 1 * time.Minute
	// TelemetryLatestTTL - 30 seconds for latest telemetry
	TelemetryLatestTTL = 30 * time.Second
)

// SetTTLs allows customizing TTL values
func SetTTLs(deviceList, roiStats, alertStats time.Duration) {
	DeviceListTTL = deviceList
	ROIStatsTTL = roiStats
	AlertStatsTTL = alertStats
}
