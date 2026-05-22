package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// FIX-051: Rate Limiter 测试
// 注意: 这些测试使用实际的 RateLimit API (RateLimit(capacity, refillRate))
// 而不是之前错误的 NewLimiter API

func TestRateLimit_BasicLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use unique limiter name to avoid interference from other tests
	router := gin.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Capacity:   2,
		RefillRate: 100.0, // Very fast refill to ensure we have tokens
		KeyFunc:    DefaultKeyFunc,
		Name:       "basic-limit-test-unique",
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// First request should pass
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should pass
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Third request should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoginRateLimit())
	router.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "login"})
	})

	// Multiple login attempts from same IP
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// LoginRateLimit allows 5 requests burst, 1 token per second
		// First 5 should pass
		assert.Equal(t, 200, w.Code, "Request %d should pass", i+1)
	}
}

func TestAPIRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(APIRateLimit())
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// API rate limit allows 100 requests burst
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
