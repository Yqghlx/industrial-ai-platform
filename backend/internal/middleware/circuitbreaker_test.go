package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/pkg/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

// ============================================
// CircuitBreakerMiddleware Tests
// ============================================

func TestCircuitBreakerMiddleware_ClosedState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create circuit breaker in closed state
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "test-service",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	router := gin.New()
	router.Use(CircuitBreakerMiddleware(cb, "test-service"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should pass through when circuit is closed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCircuitBreakerMiddleware_OpenState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create circuit breaker and force it open
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "test-service-open",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})
	cb.ForceOpen()

	router := gin.New()
	router.Use(CircuitBreakerMiddleware(cb, "test-service-open"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 503 when circuit is open
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "degraded")
	assert.Contains(t, w.Body.String(), "unavailable")
}

func TestCircuitBreakerMiddleware_AbortsOnOpenState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "abort-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})
	cb.ForceOpen()

	handlerCalled := false
	router := gin.New()
	router.Use(CircuitBreakerMiddleware(cb, "abort-test"))
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should not be called when circuit is open
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// ============================================
// WriteDegradedResponse Tests
// ============================================

func TestWriteDegradedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	fallbackData := map[string]interface{}{
		"available": false,
		"cached":    true,
	}

	WriteDegradedResponse(c, "test-service", fallbackData)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "degraded", resp["status"])
	assert.Contains(t, resp["message"], "test-service")
	assert.NotNil(t, resp["fallback"])
}

func TestWriteDegradedResponse_EmptyFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	WriteDegradedResponse(c, "empty-service", nil)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "degraded", resp["status"])
}

func TestWriteDegradedResponse_ComplexFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	fallbackData := map[string]interface{}{
		"devices": []map[string]interface{}{
			{"id": "1", "name": "Device 1"},
			{"id": "2", "name": "Device 2"},
		},
		"total":  2,
		"source": "cache",
	}

	WriteDegradedResponse(c, "device-service", fallbackData)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "degraded", resp["status"])
	fallback := resp["fallback"].(map[string]interface{})
	assert.Equal(t, float64(2), fallback["total"])
}

// ============================================
// ExecuteFallback Tests
// ============================================

func TestExecuteFallback(t *testing.T) {
	mgr := NewDegradationManager()
	mgr.RegisterStrategy(&DegradationStrategy{
		Name:     "test-service",
		Priority: 1,
		FallbackFunc: func() interface{} {
			return map[string]interface{}{
				"fallback": "data",
				"source":   "cache",
			}
		},
	})

	result := mgr.ExecuteFallback("test-service")
	assert.NotNil(t, result)
}

func TestExecuteFallback_NoStrategy(t *testing.T) {
	mgr := NewDegradationManager()

	result := mgr.ExecuteFallback("unknown-service")
	assert.NotNil(t, result)
}

func TestExecuteFallback_MultipleStrategies(t *testing.T) {
	mgr := NewDegradationManager()

	mgr.RegisterStrategy(&DegradationStrategy{
		Name:     "service-a",
		Priority: 1,
		FallbackFunc: func() interface{} {
			return map[string]interface{}{"service": "a"}
		},
	})

	mgr.RegisterStrategy(&DegradationStrategy{
		Name:     "service-b",
		Priority: 2,
		FallbackFunc: func() interface{} {
			return map[string]interface{}{"service": "b"}
		},
	})

	resultA := mgr.ExecuteFallback("service-a")
	resultB := mgr.ExecuteFallback("service-b")

	assert.NotNil(t, resultA)
	assert.NotNil(t, resultB)
}

// ============================================
// RegisterCircuitBreakerRoutes Tests
// ============================================

func TestRegisterCircuitBreakerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "test-cb",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)

	router := gin.New()
	handler.RegisterCircuitBreakerRoutes(router)

	// Test GET /circuit-breaker/status
	req := httptest.NewRequest("GET", "/circuit-breaker/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterCircuitBreakerRoutes_ForceOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "force-open-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)

	router := gin.New()
	handler.RegisterCircuitBreakerRoutes(router)

	// Test POST /circuit-breaker/:name/open
	req := httptest.NewRequest("POST", "/circuit-breaker/force-open-test/open", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["message"], "forced open")
	assert.Equal(t, "open", resp["state"])
}

func TestRegisterCircuitBreakerRoutes_ForceClose(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "force-close-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	// Force open first
	cb := manager.Get("force-close-test")
	cb.ForceOpen()

	handler := NewCircuitBreakerHandler(manager)

	router := gin.New()
	handler.RegisterCircuitBreakerRoutes(router)

	// Test POST /circuit-breaker/:name/close
	req := httptest.NewRequest("POST", "/circuit-breaker/force-close-test/close", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["message"], "forced close")
	assert.Equal(t, "closed", resp["state"])
}

func TestRegisterCircuitBreakerRoutes_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	// Don't register any circuit breaker

	handler := NewCircuitBreakerHandler(manager)

	router := gin.New()
	handler.RegisterCircuitBreakerRoutes(router)

	// Test POST /circuit-breaker/:name/open with non-existent CB
	req := httptest.NewRequest("POST", "/circuit-breaker/nonexistent/open", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "not found")
}

// ============================================
// marshalFallbackData Tests
// ============================================

func TestMarshalFallbackData(t *testing.T) {
	data := map[string]interface{}{
		"message": "fallback",
		"count":   42,
	}

	result, err := marshalFallbackData(data)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(result, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "fallback", unmarshaled["message"])
	assert.Equal(t, float64(42), unmarshaled["count"])
}

func TestMarshalFallbackData_NilData(t *testing.T) {
	result, err := marshalFallbackData(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("null"), result)
}

func TestMarshalFallbackData_ComplexData(t *testing.T) {
	data := map[string]interface{}{
		"devices": []map[string]interface{}{
			{"id": "device-1", "status": "online"},
			{"id": "device-2", "status": "offline"},
		},
		"metadata": map[string]interface{}{
			"source":    "cache",
			"timestamp": "2024-01-01T00:00:00Z",
		},
	}

	result, err := marshalFallbackData(data)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(result, &unmarshaled)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshaled["devices"])
	assert.NotNil(t, unmarshaled["metadata"])
}

// ============================================
// NewCircuitBreakerHandler Tests
// ============================================

func TestNewCircuitBreakerHandler(t *testing.T) {
	manager := circuitbreaker.NewCircuitBreakerManager()
	handler := NewCircuitBreakerHandler(manager)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.manager)
}

func TestNewCircuitBreakerHandler_NilManager(t *testing.T) {
	NewCircuitBreakerHandler(nil)
	// handler will have nil manager, operations should handle this gracefully
	// The handler struct exists even with nil manager
}

// ============================================
// GetCircuitBreakerStatus Tests
// ============================================

func TestGetCircuitBreakerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "status-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/circuit-breaker/status", nil)

	handler.GetCircuitBreakerStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["circuit_breakers"])
	assert.NotEmpty(t, resp["timestamp"])
}

func TestGetCircuitBreakerStatus_MultipleBreakers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "breaker-1",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})
	manager.Register(circuitbreaker.Config{
		Name:             "breaker-2",
		FailureThreshold: 30,
		MinRequests:      3,
		OpenTimeout:      60 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/circuit-breaker/status", nil)

	handler.GetCircuitBreakerStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	breakers := resp["circuit_breakers"].(map[string]interface{})
	assert.Len(t, breakers, 2)
}

func TestGetCircuitBreakerStatus_EmptyManager(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/circuit-breaker/status", nil)

	handler.GetCircuitBreakerStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	breakers := resp["circuit_breakers"].(map[string]interface{})
	assert.Empty(t, breakers)
}

// ============================================
// ForceOpenCircuitBreaker Tests
// ============================================

func TestForceOpenCircuitBreaker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "force-open",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/force-open/open", nil)
	c.Params = gin.Params{{Key: "name", Value: "force-open"}}

	handler.ForceOpenCircuitBreaker(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["message"], "forced open")
	assert.Equal(t, "open", resp["state"])
	assert.NotEmpty(t, resp["timestamp"])

	// Verify the circuit breaker is actually open
	cb := manager.Get("force-open")
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())
}

func TestForceOpenCircuitBreaker_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/nonexistent/open", nil)
	c.Params = gin.Params{{Key: "name", Value: "nonexistent"}}

	handler.ForceOpenCircuitBreaker(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "not found")
}

func TestForceOpenCircuitBreaker_AlreadyOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "already-open",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	// Pre-open the circuit breaker
	cb := manager.Get("already-open")
	cb.ForceOpen()

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/already-open/open", nil)
	c.Params = gin.Params{{Key: "name", Value: "already-open"}}

	// Should still succeed (idempotent)
	handler.ForceOpenCircuitBreaker(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())
}

// ============================================
// ForceCloseCircuitBreaker Tests
// ============================================

func TestForceCloseCircuitBreaker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "force-close",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	// Pre-open the circuit breaker
	cb := manager.Get("force-close")
	cb.ForceOpen()
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/force-close/close", nil)
	c.Params = gin.Params{{Key: "name", Value: "force-close"}}

	handler.ForceCloseCircuitBreaker(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["message"], "forced close")
	assert.Equal(t, "closed", resp["state"])
	assert.NotEmpty(t, resp["timestamp"])

	// Verify the circuit breaker is actually closed
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())
}

func TestForceCloseCircuitBreaker_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/nonexistent/close", nil)
	c.Params = gin.Params{{Key: "name", Value: "nonexistent"}}

	handler.ForceCloseCircuitBreaker(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "not found")
}

func TestForceCloseCircuitBreaker_AlreadyClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "already-closed",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	cb := manager.Get("already-closed")
	// By default, circuit breaker is closed
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())

	handler := NewCircuitBreakerHandler(manager)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/circuit-breaker/already-closed/close", nil)
	c.Params = gin.Params{{Key: "name", Value: "already-closed"}}

	// Should still succeed (idempotent)
	handler.ForceCloseCircuitBreaker(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())
}

// ============================================
// Circuit Breaker State Transition Tests
// ============================================

func TestCircuitBreakerStateTransition_OpenToClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "transition-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	_ = NewCircuitBreakerHandler(manager)
	cb := manager.Get("transition-test")

	// Start closed
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())

	// Force open
	cb.ForceOpen()
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())

	// Force close
	cb.ForceClose()
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())
}

func TestCircuitBreakerMiddleware_StateTransitions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "middleware-transition",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	router := gin.New()
	router.Use(CircuitBreakerMiddleware(cb, "middleware-transition"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test with closed state
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Force open the circuit breaker
	cb.ForceOpen()

	// Test with open state
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusServiceUnavailable, w2.Code)

	// Force close the circuit breaker
	cb.ForceClose()

	// Test with closed state again
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// ============================================
// Default Strategies Tests
// ============================================

func TestRegisterDefaultStrategies_WithHandler(t *testing.T) {
	mgr := NewDegradationManager()
	RegisterDefaultStrategies(mgr)

	// Test glm_api fallback
	glmFallback := mgr.ExecuteFallback("glm_api")
	assert.NotNil(t, glmFallback)

	// Test device_list fallback
	deviceFallback := mgr.ExecuteFallback("device_list")
	assert.NotNil(t, deviceFallback)

	// Test telemetry fallback
	telemetryFallback := mgr.ExecuteFallback("telemetry")
	assert.NotNil(t, telemetryFallback)
}

// ============================================
// Default Strategies Tests
// ============================================

func TestDegradedResponse_Struct(t *testing.T) {
	resp := &DegradedResponse{
		Status:     "degraded",
		Message:    "test message",
		Fallback:   map[string]interface{}{"key": "value"},
		RetryAfter: 60,
	}

	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "test message", resp.Message)
	assert.NotNil(t, resp.Fallback)
	assert.Equal(t, 60, resp.RetryAfter)
}

func TestNewDegradedResponse_WithFallback(t *testing.T) {
	fallback := map[string]interface{}{
		"data": []string{"cached1", "cached2"},
	}

	resp := NewDegradedResponse("my-service", fallback)

	assert.Equal(t, "degraded", resp.Status)
	assert.Contains(t, resp.Message, "my-service")
	assert.Equal(t, fallback, resp.Fallback)
	assert.Equal(t, 30, resp.RetryAfter)
}

// ============================================
// Integration Tests
// ============================================

func TestCircuitBreakerIntegration_FullFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	manager := circuitbreaker.NewCircuitBreakerManager()
	manager.Register(circuitbreaker.Config{
		Name:             "integration-test",
		FailureThreshold: 50,
		MinRequests:      5,
		OpenTimeout:      30 * time.Second,
	})

	handler := NewCircuitBreakerHandler(manager)
	cb := manager.Get("integration-test")

	router := gin.New()
	handler.RegisterCircuitBreakerRoutes(router)

	// Use middleware with the same circuit breaker
	api := router.Group("/api")
	api.Use(CircuitBreakerMiddleware(cb, "integration-test"))
	api.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "success"})
	})

	// 1. Initial state should be closed, request passes
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Force open via API
	reqOpen := httptest.NewRequest("POST", "/circuit-breaker/integration-test/open", nil)
	wOpen := httptest.NewRecorder()
	router.ServeHTTP(wOpen, reqOpen)
	assert.Equal(t, http.StatusOK, wOpen.Code)

	// 3. Now requests should be blocked
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusServiceUnavailable, w2.Code)

	// 4. Force close via API
	reqClose := httptest.NewRequest("POST", "/circuit-breaker/integration-test/close", nil)
	wClose := httptest.NewRecorder()
	router.ServeHTTP(wClose, reqClose)
	assert.Equal(t, http.StatusOK, wClose.Code)

	// 5. Requests should pass again
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// 6. Check status
	reqStatus := httptest.NewRequest("GET", "/circuit-breaker/status", nil)
	wStatus := httptest.NewRecorder()
	router.ServeHTTP(wStatus, reqStatus)
	assert.Equal(t, http.StatusOK, wStatus.Code)

	var statusResp map[string]interface{}
	json.Unmarshal(wStatus.Body.Bytes(), &statusResp)
	breakers := statusResp["circuit_breakers"].(map[string]interface{})
	assert.NotNil(t, breakers["integration-test"])
}
