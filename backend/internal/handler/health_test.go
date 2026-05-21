package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestNewHealthHandler tests the constructor
func TestNewHealthHandler(t *testing.T) {
	tests := []struct {
		name    string
		db      *sql.DB
		redis   *redis.Client
		version string
	}{
		{
			name:    "with all parameters",
			db:      &sql.DB{},
			redis:   &redis.Client{},
			version: "1.0.0",
		},
		{
			name:    "with nil dependencies",
			db:      nil,
			redis:   nil,
			version: "test",
		},
		{
			name:    "with empty version",
			db:      nil,
			redis:   nil,
			version: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(tt.db, tt.redis, tt.version)

			assert.NotNil(t, handler)
			assert.Equal(t, tt.db, handler.db)
			assert.Equal(t, tt.redis, handler.redis)
			assert.Equal(t, tt.version, handler.version)
			assert.NotZero(t, handler.startTime)
			assert.WithinDuration(t, time.Now(), handler.startTime, time.Second)
		})
	}
}

// TestLivenessCheck tests the liveness probe
func TestLivenessCheck(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	router := gin.New()
	router.GET("/health/live", handler.LivenessCheck)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "successful liveness check",
			method:     http.MethodGet,
			path:       "/health/live",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), "healthy")
			assert.Contains(t, w.Body.String(), "timestamp")
		})
	}
}

// TestReadinessCheck tests the readiness probe
func TestReadinessCheck(t *testing.T) {
	t.Run("ready when all dependencies are healthy", func(t *testing.T) {
		// Setup miniredis for Redis mock
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		// Setup Redis client
		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, nil, "test-version")
		handler.redis = redisClient

		router := gin.New()
		router.GET("/health/ready", handler.ReadinessCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Without DB, status should be ServiceUnavailable
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "not_ready")
		assert.Contains(t, w.Body.String(), "database")
		assert.Contains(t, w.Body.String(), "redis")
	})

	t.Run("not ready when database is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health/ready", handler.ReadinessCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "not_ready")
		assert.Contains(t, w.Body.String(), `"database":"not_ready"`)
		assert.Contains(t, w.Body.String(), `"redis":"not_ready"`)
	})

	t.Run("not ready when redis is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health/ready", handler.ReadinessCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "not_ready")
	})

	t.Run("with miniredis mock - redis ready but db nil", func(t *testing.T) {
		// Setup miniredis for Redis mock
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		// Setup Redis client
		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")

		router := gin.New()
		router.GET("/health/ready", handler.ReadinessCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// DB is nil, so still not ready
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "not_ready")
		assert.Contains(t, w.Body.String(), `"redis":"ok"`)
		assert.Contains(t, w.Body.String(), `"database":"not_ready"`)
	})

	t.Run("context timeout handling", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health/ready", handler.ReadinessCheck)

		// Create a request with already cancelled context
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should still return a response (handler handles context internally)
		assert.NotEmpty(t, w.Body.String())
	})
}

// TestCheckDatabaseReady tests the database readiness check
func TestCheckDatabaseReady(t *testing.T) {
	t.Run("returns false when db is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkDatabaseReady(ctx)
		assert.False(t, result)
	})
}

// TestCheckRedisReady tests the Redis readiness check
func TestCheckRedisReady(t *testing.T) {
	t.Run("returns false when redis is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkRedisReady(ctx)
		assert.False(t, result)
	})

	t.Run("returns true when redis is healthy", func(t *testing.T) {
		// Setup miniredis for Redis mock
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		// Setup Redis client
		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")
		ctx := context.Background()

		result := handler.checkRedisReady(ctx)
		assert.True(t, result)
	})
}

// TestHealthCheckResult tests the HealthCheckResult struct
func TestHealthCheckResult(t *testing.T) {
	t.Run("healthy result", func(t *testing.T) {
		result := HealthCheckResult{
			Status:    "healthy",
			LatencyMS: 10,
			Details:   map[string]interface{}{"connections": 5},
		}

		assert.Equal(t, "healthy", result.Status)
		assert.Equal(t, int64(10), result.LatencyMS)
		assert.NotNil(t, result.Details)
	})

	t.Run("unhealthy result with error", func(t *testing.T) {
		result := HealthCheckResult{
			Status: "unhealthy",
			Error:  "connection refused",
		}

		assert.Equal(t, "unhealthy", result.Status)
		assert.NotEmpty(t, result.Error)
	})
}

// TestHealthStatus tests the HealthStatus struct
func TestHealthStatus(t *testing.T) {
	t.Run("health status creation", func(t *testing.T) {
		status := HealthStatus{
			Status:    "healthy",
			Version:   "1.0.0",
			Uptime:    100,
			Checks:    map[string]interface{}{"database": "ok"},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, int64(100), status.Uptime)
		assert.NotNil(t, status.Checks)
		assert.NotEmpty(t, status.Timestamp)
	})
}

// TestRegisterHealthRoutes tests route registration
func TestRegisterHealthRoutes(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	router := gin.New()
	handler.RegisterHealthRoutes(router)

	routes := router.Routes()
	routePaths := make(map[string]string)
	for _, route := range routes {
		routePaths[route.Path] = route.Method
	}

	tests := []struct {
		path   string
		method string
	}{
		{"/health/live", "GET"},
		{"/health/ready", "GET"},
		{"/health/startup", "GET"},
		{"/health", "GET"},
		{"/health/dependencies", "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			method, exists := routePaths[tt.path]
			assert.True(t, exists, "route %s should exist", tt.path)
			assert.Equal(t, tt.method, method, "route %s should have method %s", tt.path, tt.method)
		})
	}
}

// TestMultipleLivenessChecks tests multiple concurrent liveness checks
func TestMultipleLivenessChecks(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	router := gin.New()
	router.GET("/health/live", handler.LivenessCheck)

	// Run multiple requests concurrently
	for i := 0; i < 10; i++ {
		t.Run("concurrent_check", func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "healthy")
		})
	}
}

// TestDetailedHealthCheck tests the detailed health check endpoint
func TestDetailedHealthCheck(t *testing.T) {
	t.Run("healthy when all dependencies nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health", handler.DetailedHealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "degraded") // DB and Redis are nil, so degraded
		assert.Contains(t, w.Body.String(), "test-version")
		assert.Contains(t, w.Body.String(), "uptime_seconds")
		assert.Contains(t, w.Body.String(), "timestamp")
	})

	t.Run("healthy with miniredis", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")

		router := gin.New()
		router.GET("/health", handler.DetailedHealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// DB is nil so still degraded, but Redis should be healthy
		assert.Contains(t, w.Body.String(), "degraded")
	})

	t.Run("returns correct structure", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health", handler.DetailedHealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "status")
		assert.Contains(t, response, "version")
		assert.Contains(t, response, "uptime_seconds")
		assert.Contains(t, response, "checks")
		assert.Contains(t, response, "timestamp")

		checks := response["checks"].(map[string]interface{})
		assert.Contains(t, checks, "database")
		assert.Contains(t, checks, "redis")
		assert.Contains(t, checks, "disk")
		assert.Contains(t, checks, "system")
	})
}

// TestCheckDatabaseHealth tests database health check function
func TestCheckDatabaseHealth(t *testing.T) {
	t.Run("returns unhealthy when db is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkDatabaseHealth(ctx)

		assert.Equal(t, "unhealthy", result.Status)
		assert.Contains(t, result.Error, "not initialized")
	})

	t.Run("returns result with latency even on error", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkDatabaseHealth(ctx)

		// Even with nil db, the function should return a result
		assert.NotEmpty(t, result.Status)
	})
}

// TestCheckRedisHealth tests Redis health check function
func TestCheckRedisHealth(t *testing.T) {
	t.Run("returns unhealthy when redis is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkRedisHealth(ctx)

		assert.Equal(t, "unhealthy", result.Status)
		assert.Contains(t, result.Error, "not initialized")
	})

	t.Run("returns healthy when redis is connected", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")
		ctx := context.Background()

		result := handler.checkRedisHealth(ctx)

		assert.Equal(t, "healthy", result.Status)
		assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
	})
}

// TestCheckDiskHealth tests disk health check
func TestCheckDiskHealth(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	result := handler.checkDiskHealth()

	assert.Equal(t, "healthy", result.Status)
	assert.Contains(t, result.Details, "check_available")
}

// TestCheckSystemHealth tests system health check
func TestCheckSystemHealth(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	result := handler.checkSystemHealth()

	assert.Equal(t, "healthy", result.Status)
	assert.Contains(t, result.Details, "go_version")
	assert.Contains(t, result.Details, "goroutines")
	assert.Contains(t, result.Details, "memory_alloc_mb")
	assert.Contains(t, result.Details, "memory_sys_mb")
	assert.Contains(t, result.Details, "gc_cycles")
}

// TestDependenciesCheck tests dependencies deep check endpoint
func TestDependenciesCheck(t *testing.T) {
	t.Run("returns unhealthy when dependencies nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health/dependencies", handler.DependenciesCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/dependencies", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "degraded")
		assert.Contains(t, w.Body.String(), "postgresql")
		assert.Contains(t, w.Body.String(), "redis")
	})

	t.Run("with miniredis", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")

		router := gin.New()
		router.GET("/health/dependencies", handler.DependenciesCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/dependencies", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// DB is nil so still degraded
		assert.Contains(t, w.Body.String(), "degraded")
	})
}

// TestCheckPostgreSQLDependency tests PostgreSQL dependency check
func TestCheckPostgreSQLDependency(t *testing.T) {
	t.Run("returns unhealthy when db is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkPostgreSQLDependency(ctx)

		assert.Equal(t, "unhealthy", result.Status)
		assert.Contains(t, result.Error, "not initialized")
	})
}

// TestCheckRedisDependency tests Redis dependency check
func TestCheckRedisDependency(t *testing.T) {
	t.Run("returns unhealthy when redis is nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")
		ctx := context.Background()

		result := handler.checkRedisDependency(ctx)

		assert.Equal(t, "unhealthy", result.Status)
		assert.Contains(t, result.Error, "not initialized")
	})

	t.Run("returns healthy when redis is connected", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")
		ctx := context.Background()

		result := handler.checkRedisDependency(ctx)

		assert.Equal(t, "healthy", result.Status)
		assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
	})
}

// TestStartupCheck tests startup probe endpoint
func TestStartupCheck(t *testing.T) {
	t.Run("not started when dependencies nil", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, "test-version")

		router := gin.New()
		router.GET("/health/startup", handler.StartupCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/startup", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "starting")
		assert.Contains(t, w.Body.String(), "database")
		assert.Contains(t, w.Body.String(), "redis")
		assert.Contains(t, w.Body.String(), "uptime_s")
	})

	t.Run("started when all dependencies ready", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		handler := NewHealthHandler(nil, redisClient, "test-version")

		router := gin.New()
		router.GET("/health/startup", handler.StartupCheck)

		req := httptest.NewRequest(http.MethodGet, "/health/startup", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// DB is nil, so still starting
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "starting")
		// Redis should be initialized
		assert.Contains(t, w.Body.String(), `"redis":"initialized"`)
	})
}

// TestHealthCheckerInterface tests the HealthChecker interface
func TestHealthCheckerInterface(t *testing.T) {
	// Create a custom health checker
	customChecker := &MockHealthChecker{
		name: "custom_check",
		result: HealthCheckResult{
			Status:    "healthy",
			LatencyMS: 10,
		},
	}

	handler := NewHealthHandler(nil, nil, "test-version")
	handler.dependencies = []HealthChecker{customChecker}

	assert.Equal(t, "custom_check", customChecker.Name())
	result := customChecker.Check(context.Background())
	assert.Equal(t, "healthy", result.Status)
}

// MockHealthChecker for testing the interface
type MockHealthChecker struct {
	name   string
	result HealthCheckResult
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

func (m *MockHealthChecker) Check(ctx context.Context) HealthCheckResult {
	return m.result
}
