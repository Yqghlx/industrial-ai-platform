package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket for rate limiting using token bucket algorithm
type TokenBucket struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimiter manages rate limiters per key
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

// Global limiter manager - singleton pattern to prevent goroutine leaks
var (
	globalLimiterManager *LimiterManager
	limiterOnce          sync.Once
)

// LimiterManager manages all rate limiters and cleanup goroutine
// This is a singleton to ensure only one cleanup goroutine runs
type LimiterManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// GetLimiterManager returns the global singleton limiter manager
func GetLimiterManager() *LimiterManager {
	limiterOnce.Do(func() {
		globalLimiterManager = &LimiterManager{
			limiters: make(map[string]*RateLimiter),
			stopCh:   make(chan struct{}),
		}
		// Start single cleanup goroutine
		globalLimiterManager.startCleanup()
	})
	return globalLimiterManager
}

// startCleanup starts a single goroutine to clean up old buckets
func (lm *LimiterManager) startCleanup() {
	lm.wg.Add(1)
	go func() {
		defer lm.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lm.cleanupAllLimiters()
			case <-lm.stopCh:
				return
			}
		}
	}()
}

// cleanupAllLimiters removes old buckets from all limiters
func (lm *LimiterManager) cleanupAllLimiters() {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	for _, limiter := range lm.limiters {
		limiter.CleanupOldBuckets()
	}
}

// Stop stops the cleanup goroutine (call on shutdown)
// This method is idempotent and safe to call multiple times
func (lm *LimiterManager) Stop() {
	select {
	case <-lm.stopCh:
		// Already stopped
	default:
		close(lm.stopCh)
	}
	lm.wg.Wait()
}

// GetOrCreateLimiter gets or creates a rate limiter for a given name
func (lm *LimiterManager) GetOrCreateLimiter(name string) *RateLimiter {
	lm.mu.RLock()
	limiter, exists := lm.limiters[name]
	lm.mu.RUnlock()

	if exists {
		return limiter
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Double check
	if limiter, exists := lm.limiters[name]; exists {
		return limiter
	}

	limiter = NewRateLimiter()
	lm.limiters[name] = limiter
	return limiter
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// LastUsed returns the last time the bucket was used
func (tb *TokenBucket) LastUsed() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.lastRefill
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

// GetBucket gets or creates a bucket for a key
func (rl *RateLimiter) GetBucket(key string, capacity, refillRate float64) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double check
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}

	bucket = NewTokenBucket(capacity, refillRate)
	rl.buckets[key] = bucket
	return bucket
}

// CleanupOldBuckets removes buckets not used recently
func (rl *RateLimiter) CleanupOldBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		if now.Sub(bucket.LastUsed()) > 10*time.Minute {
			delete(rl.buckets, key)
		}
	}
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Capacity   int                         // Maximum number of tokens (burst capacity)
	RefillRate float64                     // Tokens refilled per second
	KeyFunc    func(c *gin.Context) string // Function to extract rate limit key
	Name       string                      // Name for the limiter (for management)
}

// DefaultKeyFunc returns the client IP as the rate limit key
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimit middleware using IP-based token bucket
// This version uses the global singleton limiter manager to prevent goroutine leaks
func RateLimit(capacity int, refillRate float64) gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   capacity,
		RefillRate: refillRate,
		KeyFunc:    DefaultKeyFunc,
		Name:       "default",
	})
}

// RateLimitWithConfig creates a rate limit middleware with custom configuration
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	// Apply defaults
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	if config.Name == "" {
		config.Name = "default"
	}

	// Get the global limiter manager (singleton)
	manager := GetLimiterManager()

	// Get or create limiter for this configuration
	limiter := manager.GetOrCreateLimiter(config.Name)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)
		bucket := limiter.GetBucket(key, float64(config.Capacity), config.RefillRate)

		if !bucket.Allow() {
			c.JSON(429, gin.H{
				"error":   "Rate limit exceeded",
				"code":    "RATE_LIMIT",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithContext creates a rate limit middleware with context for proper shutdown
// DEPRECATED: Use RateLimitWithConfig with the global LimiterManager instead.
// This function is kept for backward compatibility but may cause goroutine leaks
// if the context is never cancelled. Prefer using RateLimit() or RateLimitWithConfig()
// which uses the singleton manager with proper lifecycle management.
func RateLimitWithContext(_ context.Context, capacity int, refillRate float64, name string) gin.HandlerFunc {
	// Use the global singleton manager instead of creating isolated limiters
	// This prevents goroutine leaks by reusing the single cleanup goroutine
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   capacity,
		RefillRate: refillRate,
		Name:       name,
		KeyFunc:    DefaultKeyFunc,
	})
}

// DefaultRateLimit 全局默认速率限制
// SEC-MEDIUM-04: 添加全局默认限流，防止未配置限流的端点被滥用
func DefaultRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   60, // 每分钟 60 次
		RefillRate: 1,  // 每秒补充 1 个令牌
		Name:       "default_global",
	})
}

// LoginRateLimit specific rate limit for login endpoints
func LoginRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   5,
		RefillRate: 1,
		Name:       "login",
	})
}

// RegisterRateLimit specific rate limit for registration
func RegisterRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   3,
		RefillRate: 0.5,
		Name:       "register",
	})
}

// APIRateLimit general API rate limit
func APIRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   100,
		RefillRate: 10,
		Name:       "api",
	})
}

// TelemetryRateLimit rate limit for telemetry ingestion endpoints
// This is a high-frequency API, allow more requests but with burst control
func TelemetryRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   200,
		RefillRate: 50,
		Name:       "telemetry",
	})
}

// AgentQueryRateLimit rate limit for AI agent query endpoints
// AI queries are resource-intensive, use stricter limits
func AgentQueryRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   20,
		RefillRate: 2,
		Name:       "agent_query",
	})
}

// ROIStatsRateLimit rate limit for ROI statistics endpoints
func ROIStatsRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   30,
		RefillRate: 5,
		Name:       "roi_stats",
	})
}

// WebSocketRateLimit rate limit for WebSocket connection endpoints
// WebSocket connections are expensive but needed for real-time data
// SEC-001: Add rate limiting to WebSocket endpoint for public access
func WebSocketRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   10,  // Allow up to 10 concurrent connection attempts per IP
		RefillRate: 0.5, // Refill 1 token every 2 seconds (slow refill for connections)
		Name:       "websocket",
	})
}

// UserBasedRateLimit creates a rate limit based on user ID instead of IP
// Useful when you want to rate limit authenticated users specifically
func UserBasedRateLimit(capacity int, refillRate float64) gin.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Capacity:   capacity,
		RefillRate: refillRate,
		Name:       "user_based",
		KeyFunc: func(c *gin.Context) string {
			// Try to get user ID from context (set by auth middleware)
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(string); ok && uid != "" {
					return "user:" + uid
				}
			}
			// Fallback to IP-based rate limiting
			return "ip:" + c.ClientIP()
		},
	})
}

// CombinedRateLimit creates a rate limit that combines IP and user rate limiting
// This allows for both IP-level and user-level rate limiting
func CombinedRateLimit(ipCapacity int, ipRefillRate float64, userCapacity int, userRefillRate float64) gin.HandlerFunc {
	ipLimiter := RateLimitWithConfig(RateLimitConfig{
		Capacity:   ipCapacity,
		RefillRate: ipRefillRate,
		Name:       "combined_ip",
	})

	userLimiter := UserBasedRateLimit(userCapacity, userRefillRate)

	return func(c *gin.Context) {
		// Check IP rate limit first
		ipLimiter(c)
		if c.IsAborted() {
			return
		}

		// Then check user rate limit
		userLimiter(c)
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
