package cache

import (
	"context"
	"time"
)

// NoopCache is a cache implementation that does nothing (used when cache is disabled)
type NoopCache struct {
	stats Stats
}

// NewNoopCache creates a new noop cache
func NewNoopCache() *NoopCache {
	return &NoopCache{
		stats: Stats{
			Available:   false,
			BackendType: "noop",
		},
	}
}

// Get always returns ErrNotFound
func (c *NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrNotFound
}

// Set does nothing
func (c *NoopCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// DeleteByPattern does nothing
func (c *NoopCache) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}

// Exists always returns false
func (c *NoopCache) Exists(ctx context.Context, key string) bool {
	return false
}

// GetTTL always returns 0
func (c *NoopCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, ErrNotFound
}

// IsAvailable always returns false
func (c *NoopCache) IsAvailable() bool {
	return false
}

// GetStats returns empty stats
func (c *NoopCache) GetStats() Stats {
	return Stats{
		Available:   false,
		BackendType: "noop",
	}
}

// Close does nothing
func (c *NoopCache) Close() error {
	return nil
}
