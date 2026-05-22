package middleware

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// TokenBucket Boundary Tests
// ============================================

func TestTokenBucket_ExactCapacity(t *testing.T) {
	tb := NewTokenBucket(3.0, 1.0) // Capacity 3, refill rate 1 token/sec

	// Exactly 3 requests should pass
	assert.True(t, tb.Allow(), "Request 1 should pass")
	assert.True(t, tb.Allow(), "Request 2 should pass")
	assert.True(t, tb.Allow(), "Request 3 should pass")

	// 4th request should be blocked
	assert.False(t, tb.Allow(), "Request 4 should be blocked")
}

func TestTokenBucket_RefillRate(t *testing.T) {
	tb := NewTokenBucket(2.0, 1.0) // Capacity 2, refill 1 token/sec

	// Use all tokens
	assert.True(t, tb.Allow())
	assert.True(t, tb.Allow())
	assert.False(t, tb.Allow())

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should have 1 token refilled
	assert.True(t, tb.Allow(), "After 1 second refill, should allow 1 request")
	assert.False(t, tb.Allow(), "Second request should still be blocked")
}

func TestTokenBucket_BurstAllowance(t *testing.T) {
	tb := NewTokenBucket(10.0, 1.0) // High burst capacity

	// Should allow burst of 10 requests
	for i := 0; i < 10; i++ {
		assert.True(t, tb.Allow(), "Request %d should pass in burst", i+1)
	}

	assert.False(t, tb.Allow(), "Request 11 should be blocked")
}

func TestTokenBucket_FastRefill(t *testing.T) {
	tb := NewTokenBucket(2.0, 5.0) // Fast refill rate

	// Use tokens
	assert.True(t, tb.Allow())
	assert.True(t, tb.Allow())

	// Wait short time for refill (0.5 sec = 2.5 tokens)
	time.Sleep(600 * time.Millisecond)

	// Should have enough tokens again
	assert.True(t, tb.Allow(), "Should have refilled quickly")
}

func TestTokenBucket_ZeroCapacity(t *testing.T) {
	tb := NewTokenBucket(0.0, 1.0)

	// Zero capacity means no requests allowed
	assert.False(t, tb.Allow(), "Zero capacity should block all requests")
}

func TestTokenBucket_SlowRefill(t *testing.T) {
	tb := NewTokenBucket(1.0, 0.1) // 1 token per 10 seconds

	// First request passes
	assert.True(t, tb.Allow())

	// Second request blocked
	assert.False(t, tb.Allow())

	// Wait 5 seconds (should still be blocked)
	time.Sleep(5 * time.Second)
	assert.False(t, tb.Allow(), "Should still be blocked after 5 seconds")

	// Wait another 5 seconds (should have 1 token)
	time.Sleep(6 * time.Second)
	assert.True(t, tb.Allow(), "Should allow after 10+ seconds")
}

func TestTokenBucket_LastUsed(t *testing.T) {
	tb := NewTokenBucket(5.0, 1.0)

	// Get initial last used time
	initialTime := tb.LastUsed()

	// Make a request
	tb.Allow()

	// Last used time should update
	newTime := tb.LastUsed()
	assert.True(t, newTime.After(initialTime) || newTime.Equal(initialTime))
}

// ============================================
// RateLimiter Boundary Tests
// ============================================

func TestRateLimiter_GetBucket_DuplicateKey(t *testing.T) {
	rl := NewRateLimiter()

	bucket1 := rl.GetBucket("test-key", 5.0, 1.0)
	bucket2 := rl.GetBucket("test-key", 5.0, 1.0)

	// Should return same bucket for same key
	assert.Equal(t, bucket1, bucket2)
}

func TestRateLimiter_GetBucket_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter()

	bucket1 := rl.GetBucket("key1", 5.0, 1.0)
	bucket2 := rl.GetBucket("key2", 5.0, 1.0)

	// Should return different buckets for different keys
	assert.NotEqual(t, bucket1, bucket2)
}

func TestRateLimiter_CleanupOldBuckets(t *testing.T) {
	rl := NewRateLimiter()

	// Create bucket
	rl.GetBucket("test-key", 5.0, 1.0)

	// Verify bucket exists
	rl.mu.RLock()
	_, exists := rl.buckets["test-key"]
	rl.mu.RUnlock()
	assert.True(t, exists)

	// Cleanup removes buckets not used recently (> 10 minutes)
	// Since we just created the bucket, it shouldn't be removed
	rl.CleanupOldBuckets()

	rl.mu.RLock()
	_, existsAfter := rl.buckets["test-key"]
	rl.mu.RUnlock()
	assert.True(t, existsAfter) // Should still exist
}

// ============================================
// LimiterManager Tests
// ============================================

func TestLimiterManager_GetLimiterManager_Singleton(t *testing.T) {
	manager1 := GetLimiterManager()
	manager2 := GetLimiterManager()

	assert.Equal(t, manager1, manager2)
}

func TestLimiterManager_GetOrCreateLimiter(t *testing.T) {
	manager := GetLimiterManager()

	limiter1 := manager.GetOrCreateLimiter("test-limiter")
	limiter2 := manager.GetOrCreateLimiter("test-limiter")

	assert.Equal(t, limiter1, limiter2)
}

func TestLimiterManager_GetOrCreateLimiter_DifferentNames(t *testing.T) {
	manager := GetLimiterManager()

	// Create limiters with unique names to avoid interference
	name1 := "limiter-different-test-1"
	name2 := "limiter-different-test-2"

	limiter1 := manager.GetOrCreateLimiter(name1)
	limiter2 := manager.GetOrCreateLimiter(name2)

	// Both should exist in the manager's limiters map
	manager.mu.RLock()
	_, exists1 := manager.limiters[name1]
	_, exists2 := manager.limiters[name2]
	manager.mu.RUnlock()

	assert.True(t, exists1)
	assert.True(t, exists2)
	// Both limiters should be different objects
	assert.False(t, limiter1 == limiter2)
}

func TestLimiterManager_Stop(t *testing.T) {
	// Create a new manager for this test
	manager := &LimiterManager{
		limiters: make(map[string]*RateLimiter),
		stopCh:   make(chan struct{}),
	}
	manager.startCleanup()

	// Stop should close the channel
	manager.Stop()

	// Verify stopCh is closed
	select {
	case <-manager.stopCh:
		// Channel is closed, expected
	default:
		assert.Fail(t, "stopCh should be closed after Stop()")
	}
}

func TestLimiterManager_Stop_Idempotent(t *testing.T) {
	manager := &LimiterManager{
		limiters: make(map[string]*RateLimiter),
		stopCh:   make(chan struct{}),
	}
	manager.startCleanup()

	// Multiple stops should not panic
	manager.Stop()
	manager.Stop()
	manager.Stop()

	// Verify channel is closed
	select {
	case <-manager.stopCh:
		// Expected
	default:
		assert.Fail(t, "stopCh should be closed")
	}
}

// ============================================
// RateLimitWithConfig Boundary Tests
// ============================================

func TestRateLimitWithConfig_CustomKeyFunc(t *testing.T) {
	router := gin.New()

	customKeyFunc := func(c *gin.Context) string {
		return "custom-key-" + c.GetHeader("X-Custom-ID")
	}

	router.Use(RateLimitWithConfig(RateLimitConfig{
		Capacity:   2,
		RefillRate: 1,
		KeyFunc:    customKeyFunc,
		Name:       "custom-key-test",
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Two requests with same custom ID should be limited
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Custom-ID", "user-1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Custom-ID", "user-1")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Custom-ID", "user-1")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code) // Blocked

	// Request with different custom ID should pass
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.Header.Set("X-Custom-ID", "user-2")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, 200, w4.Code)
}

func TestRateLimitWithConfig_DefaultKeyFunc(t *testing.T) {
	router := gin.New()

	router.Use(RateLimitWithConfig(RateLimitConfig{
		Capacity:   1,
		RefillRate: 1,
		KeyFunc:    nil, // Should use default (IP-based)
		Name:       "default-key-test",
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
}

func TestRateLimitWithConfig_DefaultName(t *testing.T) {
	router := gin.New()

	router.Use(RateLimitWithConfig(RateLimitConfig{
		Capacity:   10,
		RefillRate: 5,
		KeyFunc:    DefaultKeyFunc,
		Name:       "", // Should use default name
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// ============================================
// Rate Limit Functions Tests
// ============================================

func TestRegisterRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(RegisterRateLimit())
	router.POST("/register", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "registered"})
	})

	// RegisterRateLimit allows 3 requests burst, 0.5 tokens/sec
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/register", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}

	req := httptest.NewRequest("POST", "/register", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestTelemetryRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(TelemetryRateLimit())
	router.POST("/telemetry", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "telemetry received"})
	})

	// TelemetryRateLimit allows 200 requests burst
	for i := 0; i < 200; i++ {
		req := httptest.NewRequest("POST", "/telemetry", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}

	req := httptest.NewRequest("POST", "/telemetry", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestAgentQueryRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(AgentQueryRateLimit())
	router.POST("/agent/query", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "query processed"})
	})

	// AgentQueryRateLimit allows 20 requests burst
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("POST", "/agent/query", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}

	req := httptest.NewRequest("POST", "/agent/query", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestROIStatsRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(ROIStatsRateLimit())
	router.GET("/roi/stats", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "stats"})
	})

	// ROIStatsRateLimit allows 30 requests burst
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest("GET", "/roi/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}

	req := httptest.NewRequest("GET", "/roi/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestWebSocketRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(WebSocketRateLimit())
	router.GET("/ws", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "websocket connected"})
	})

	// WebSocketRateLimit allows 10 requests burst
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/ws", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestUserBasedRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(UserBasedRateLimit(2, 1.0))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Without user_id in context, should fallback to IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestCombinedRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(CombinedRateLimit(2, 1.0, 1, 1.0)) // IP: 2 burst, User: 1 burst
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// IP limit: 2 requests allowed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// User limit: 1 request (but no user_id, so IP-based)
	// Combined should block after IP limit is exhausted
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

// ============================================
// Concurrent Rate Limiting Tests
// ============================================

func TestRateLimit_ConcurrentRequests(t *testing.T) {
	router := gin.New()
	router.Use(RateLimit(5, 1.0)) // 5 burst capacity
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	var wg sync.WaitGroup
	results := make([]int, 10)
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			mu.Lock()
			results[idx] = w.Code
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Count successes - should have at least 5 due to burst capacity
	// (race conditions might allow slightly more or less)
	successCount := 0
	for _, code := range results {
		if code == 200 {
			successCount++
		}
	}

	// Should have at least 4 successes (burst capacity minus race condition tolerance)
	assert.GreaterOrEqual(t, successCount, 4)
}

func TestRateLimitWithContext_BackwardCompatibility(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitWithContext(context.Background(), 2, 1.0, "compatibility-test"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

// ============================================
// min Helper Function Test
// ============================================

func TestMin(t *testing.T) {
	assert.Equal(t, 1.0, min(1.0, 2.0))
	assert.Equal(t, 2.0, min(3.0, 2.0))
	assert.Equal(t, 0.0, min(0.0, 5.0))
	assert.Equal(t, -1.0, min(-1.0, 1.0))
	assert.Equal(t, 5.0, min(5.0, 5.0))
}