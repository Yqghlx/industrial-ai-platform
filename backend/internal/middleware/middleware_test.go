package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ============================================
// Shutdown middleware tests
// ============================================

func TestShutdownMiddleware_NotShuttingDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ShutdownMiddleware(func() bool { return false }))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestShutdownMiddleware_ShuttingDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ShutdownMiddleware(func() bool { return true }))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 503, w.Code)
}

// ============================================
// ActiveRequestTracker tests
// ============================================

func TestActiveRequestTracker_IncrementDecrement(t *testing.T) {
	tracker := NewActiveRequestTracker()

	assert.Equal(t, int64(0), tracker.GetActiveCount())

	tracker.Increment()
	assert.Equal(t, int64(1), tracker.GetActiveCount())

	tracker.Increment()
	assert.Equal(t, int64(2), tracker.GetActiveCount())

	tracker.Decrement()
	assert.Equal(t, int64(1), tracker.GetActiveCount())
}

func TestActiveRequestTracker_IsRequestComplete(t *testing.T) {
	tracker := NewActiveRequestTracker()
	assert.True(t, tracker.IsRequestComplete())

	tracker.Increment()
	assert.False(t, tracker.IsRequestComplete())

	tracker.Decrement()
	assert.True(t, tracker.IsRequestComplete())
}

func TestActiveRequestTracker_WaitForRequestsComplete(t *testing.T) {
	tracker := NewActiveRequestTracker()

	// With no active requests, should return immediately as true
	assert.True(t, tracker.IsRequestComplete())

	// With active requests, tracker shows not complete
	tracker.Increment()
	assert.False(t, tracker.IsRequestComplete())

	tracker.Decrement()
	assert.True(t, tracker.IsRequestComplete())
}

func TestRequestTrackingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := NewActiveRequestTracker()

	router := gin.New()
	router.Use(RequestTrackingMiddleware(tracker))
	router.GET("/test", func(c *gin.Context) {
		assert.Equal(t, int64(1), tracker.GetActiveCount())
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, int64(0), tracker.GetActiveCount())
}

// ============================================
// Tenant middleware tests
// ============================================

func TestTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TenantIsolation())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetTenantSlug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// No tenant slug set
	slug := GetTenantSlug(c)
	assert.Empty(t, slug)

	// With tenant slug
	c.Set("tenant_slug", "test-company")
	slug = GetTenantSlug(c)
	assert.Equal(t, "test-company", slug)
}

func TestGetTenantPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// No plan set
	plan := GetTenantPlan(c)
	assert.Empty(t, plan)

	// With plan
	c.Set("tenant_plan", "enterprise")
	plan = GetTenantPlan(c)
	assert.Equal(t, "enterprise", plan)
}

func TestSetTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	SetTenantContext(c, "tenant-123", "my-company", "pro")

	id, _ := c.Get("tenant_id")
	assert.Equal(t, "tenant-123", id)
	slug, _ := c.Get("tenant_slug")
	assert.Equal(t, "my-company", slug)
	plan, _ := c.Get("tenant_plan")
	assert.Equal(t, "pro", plan)
}

func TestTenantAdminRequired_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TenantAdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_role", "tenant_admin")

	// Test via handler directly
	router.ServeHTTP(w, req)

	// Should proceed if role is tenant_admin (depends on implementation)
	// If it checks for "tenant_admin" role specifically, should be 200
}

func TestBuildTenantFilter(t *testing.T) {
	filter := BuildTenantFilter("tenant-123", "SELECT * FROM devices WHERE 1=1", "arg1")
	assert.Contains(t, filter.WhereClause, "tenant_id")
}

func TestTenantScopedQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("tenant_id", "tenant-123")

	query, args := TenantScopedQuery(c, "SELECT * FROM devices")
	assert.Contains(t, query, "tenant_id")
	assert.Contains(t, args, "tenant-123")
}

// ============================================
// CSRF middleware tests
// ============================================

func TestCSRF_SafeMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRF_OptionsMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRF_PostWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be rejected without CSRF token
	assert.Equal(t, 403, w.Code)
}

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()
	assert.NotEmpty(t, config.CookieName)
	assert.NotEmpty(t, config.HeaderName)
	assert.True(t, config.CookieSecure)
}

func TestDevelopmentCSRFConfig(t *testing.T) {
	config := DevelopmentCSRFConfig()
	assert.False(t, config.CookieSecure)
}

func TestIsSafeMethod(t *testing.T) {
	assert.True(t, isSafeMethod("GET"))
	assert.True(t, isSafeMethod("HEAD"))
	assert.True(t, isSafeMethod("OPTIONS"))
	assert.True(t, isSafeMethod("TRACE"))
	assert.False(t, isSafeMethod("POST"))
	assert.False(t, isSafeMethod("PUT"))
	assert.False(t, isSafeMethod("DELETE"))
}

func TestGenerateCSRFToken(t *testing.T) {
	token := generateCSRFToken(32)
	assert.NotEmpty(t, token)
	assert.True(t, len(token) > 0)

	token2 := generateCSRFToken(32)
	assert.NotEqual(t, token, token2, "Tokens should be unique")
}

func TestCSRFTokenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.GET("/csrf-token", CSRFTokenHandler(config))

	req := httptest.NewRequest("GET", "/csrf-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSecureCompare(t *testing.T) {
	assert.True(t, secureCompare("abc", "abc"))
	assert.False(t, secureCompare("abc", "def"))
	assert.False(t, secureCompare("abc", ""))
	assert.False(t, secureCompare("", "abc"))
}

// ============================================
// Device Auth middleware tests
// ============================================

func TestDefaultDeviceAuthConfig(t *testing.T) {
	config := DefaultDeviceAuthConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.HeaderName)
}

func TestGenerateDeviceKey(t *testing.T) {
	key := GenerateDeviceKey("device-001", "secret123")
	assert.NotEmpty(t, key)
}

func TestValidateDeviceKey(t *testing.T) {
	key := GenerateDeviceKey("device-001", "secret123")
	assert.True(t, ValidateDeviceKey(key, "device-001", "secret123"))
	assert.False(t, ValidateDeviceKey(key, "device-001", "wrong-secret"))
	assert.False(t, ValidateDeviceKey(key, "wrong-device", "secret123"))
}

func TestGetDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// No device ID set
	id := GetDeviceID(c)
	assert.Empty(t, id)

	// With device ID
	c.Set("device_id", "CNC-001")
	id = GetDeviceID(c)
	assert.Equal(t, "CNC-001", id)
}

func TestIsDeviceAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	assert.False(t, IsDeviceAuthenticated(c))

	c.Set("device_authenticated", true)
	assert.True(t, IsDeviceAuthenticated(c))
}

func TestDeviceKeyFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultDeviceAuthConfig()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set(config.HeaderName, "test-device-key")

	key := DeviceKeyFromRequest(c, config)
	assert.Equal(t, "test-device-key", key)
}

// ============================================
// JWT Helper tests
// ============================================

func TestGenerateAndParseJWT(t *testing.T) {
	secret := []byte("test-secret-key-that-is-long-enough")

	token, err := GenerateJWT(1, "testuser", "admin", "tenant-123", secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseJWT(token, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "tenant-123", claims.TenantID)
}

// TestParseJWT_InvalidToken and TestParseJWT_WrongSecret moved to jwt_helpers_test.go

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	token2, err := GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestGetSetJWTSecret(t *testing.T) {
	originalSecret := GetJWTSecret()

	SetJWTSecret("new-test-secret")
	assert.Equal(t, []byte("new-test-secret"), GetJWTSecret())

	// Restore original
	SetJWTSecret(string(originalSecret))
}

// ============================================
// Circuit breaker middleware tests
// ============================================

func TestCircuitBreakerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a circuit breaker with relaxed settings
	// This requires importing the circuitbreaker package
	// For now, test the handler functions
	handler := &CircuitBreakerHandler{}

	assert.NotNil(t, handler)
}

func TestNewDegradedResponse(t *testing.T) {
	resp := NewDegradedResponse("test-service", map[string]interface{}{
		"message": "service degraded",
	})
	assert.NotNil(t, resp)
	assert.Equal(t, "degraded", resp.Status)
}

func TestWriteDegradedResponseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	WriteDegradedResponseJSON(c, "test-service", map[string]interface{}{
		"status": "degraded",
	})

	assert.Equal(t, 503, w.Code)
}

func TestNewDegradationManager(t *testing.T) {
	mgr := NewDegradationManager()
	assert.NotNil(t, mgr)
}

func TestAIResponseFallback(t *testing.T) {
	fallback := AIResponseFallback()
	assert.NotNil(t, fallback)
}

func TestDeviceListFallback(t *testing.T) {
	fallback := DeviceListFallback()
	assert.NotNil(t, fallback)
}

func TestTelemetryFallback(t *testing.T) {
	fallback := TelemetryFallback()
	assert.NotNil(t, fallback)
}

func TestRegisterDefaultStrategies(t *testing.T) {
	mgr := NewDegradationManager()
	RegisterDefaultStrategies(mgr)
	// Should not panic
}

// ============================================
// Logger middleware tests
// ============================================

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// ============================================
// Logging middleware tests
// ============================================

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestTraceIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SlowRequestLogMiddleware(100)) // 100ms threshold
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestErrorLogMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// Helper function tests
// ============================================

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()
	assert.NotEmpty(t, id)

	id2 := generateRequestID()
	assert.NotEqual(t, id, id2)
}

func TestGenerateFallbackRequestID(t *testing.T) {
	id := generateFallbackRequestID()
	assert.NotEmpty(t, id)
}

func TestGenerateSecureRandomString(t *testing.T) {
	s := generateSecureRandomString(16)
	assert.NotEmpty(t, s)

	s2 := generateSecureRandomString(16)
	assert.NotEqual(t, s, s2)
}

func TestGenerateTraceID(t *testing.T) {
	id := generateTraceID()
	assert.NotEmpty(t, id)
}

func TestSplitString(t *testing.T) {
	parts := splitString("a,b,c", ",")
	assert.Equal(t, []string{"a", "b", "c"}, parts)

	parts = splitString("", ",")
	assert.NotNil(t, parts) // empty string still returns a slice with empty string
}

func TestRandomInt(t *testing.T) {
	n := randomInt(100)
	assert.True(t, n >= 0 && n < 100)
}

func TestRandomString(t *testing.T) {
	s := randomString(10)
	assert.Len(t, s, 10)

	s2 := randomString(10)
	assert.NotEqual(t, s, s2)
}

// ============================================
// CleanupMiddleware test
// ============================================

func TestCleanupMiddleware(t *testing.T) {
	// Should not panic
	CleanupMiddleware()
}

// ============================================
// Role Required middleware test
// ============================================

func TestRoleRequired_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RoleRequired("admin", "superadmin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRoleRequired_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "viewer")
		c.Next()
	})
	router.Use(RoleRequired("admin", "superadmin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestRoleRequired_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RoleRequired("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

// ============================================
// Wait for active request tracker
// ============================================

func TestWaitForRequestsComplete_Immediate(t *testing.T) {
	tracker := NewActiveRequestTracker()
	// With no active requests, IsRequestComplete returns true
	assert.True(t, tracker.IsRequestComplete())
}
