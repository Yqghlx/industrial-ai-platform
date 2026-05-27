package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// AgentOptimizer provides LLM call optimization
// P2-3: Queue mechanism + Answer caching
type AgentOptimizer struct {
	cache      cache.CacheService
	semaphore  *semaphore.Weighted // Max concurrent LLM calls
	cacheTTL   time.Duration       // Cache TTL for answers
	maxConcurrent int64            // Max concurrent LLM calls
}

// NewAgentOptimizer creates a new optimizer
func NewAgentOptimizer(cacheSvc cache.CacheService, maxConcurrent int64) *AgentOptimizer {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default: 10 concurrent LLM calls
	}
	return &AgentOptimizer{
		cache:         cacheSvc,
		semaphore:     semaphore.NewWeighted(maxConcurrent),
		cacheTTL:      30 * time.Minute, // Cache answers for 30 minutes
		maxConcurrent: maxConcurrent,
	}
}

// GetCachedAnswer retrieves cached answer if exists
func (o *AgentOptimizer) GetCachedAnswer(ctx context.Context, query string) (string, bool) {
	if o.cache == nil {
		return "", false
	}

	cacheKey := o.generateCacheKey(query)
	data, err := o.cache.Get(ctx, cacheKey)
	if err != nil || data == nil {
		return "", false
	}

	logger.L().Debug("Using cached answer",
		zap.String("query_hash", cacheKey[:8]),
	)
	return string(data), true
}

// CacheAnswer stores answer in cache
func (o *AgentOptimizer) CacheAnswer(ctx context.Context, query string, answer string) {
	if o.cache == nil || answer == "" {
		return
	}

	cacheKey := o.generateCacheKey(query)
	err := o.cache.Set(ctx, cacheKey, []byte(answer), o.cacheTTL)
	if err != nil {
		logger.L().Warn("Failed to cache answer",
			zap.String("query_hash", cacheKey[:8]),
			zap.Error(err),
		)
	}
}

// AcquireSlot acquires a slot for LLM call (blocking if queue full)
func (o *AgentOptimizer) AcquireSlot(ctx context.Context) error {
	return o.semaphore.Acquire(ctx, 1)
}

// ReleaseSlot releases a slot after LLM call completes
func (o *AgentOptimizer) ReleaseSlot() {
	o.semaphore.Release(1)
}

// generateCacheKey creates a unique cache key for query
// Uses standard cache key naming convention: agent:answer:<query_hash>
// This provides namespace isolation and follows the project's cache key standards
func (o *AgentOptimizer) generateCacheKey(query string) string {
	hash := sha256.Sum256([]byte(query))
	return cache.AgentCachePrefix.Build("answer", hex.EncodeToString(hash[:]))
}

// QueueStats returns current queue statistics
func (o *AgentOptimizer) QueueStats() map[string]int64 {
	available := int64(0)
	if o.semaphore.TryAcquire(1) {
		available = 1
		o.semaphore.Release(1)
	}
	currentQueue := o.maxConcurrent - available
	return map[string]int64{
		"max_concurrent": o.maxConcurrent,
		"current_queue":  currentQueue,
	}
}