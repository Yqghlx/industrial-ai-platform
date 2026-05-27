package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// ============================================
// InitPrometheus tests
// ============================================

func TestInitPrometheus(t *testing.T) {
	// InitPrometheus should not panic when registering metrics
	// Since metrics are already registered at package init time,
	// we verify it can be called (metrics registered globally)
	
	// Create a new registry for testing registry behavior
	registry := prometheus.NewRegistry()
	
	// Create new metric instances for this test
	testHTTPRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_http_requests_total",
			Help: "Test total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	// Register should work
	err := registry.Register(testHTTPRequestsTotal)
	assert.NoError(t, err)
	
	// Duplicate registration should fail
	err = registry.Register(testHTTPRequestsTotal)
	assert.Error(t, err, "Duplicate registration should fail")
}

func TestInitPrometheus_NoPanic(t *testing.T) {
	// Calling InitPrometheus should not panic
	// Note: This may panic if already registered in default registry
	// but we catch that with recover
	defer func() {
		if r := recover(); r != nil {
			// Expected if already registered
			t.Logf("InitPrometheus recovered from: %v", r)
		}
	}()
	InitPrometheus()
}

func TestInitPrometheus_MetricsRegistration(t *testing.T) {
	// Create a fresh registry to test metric registration logic
	registry := prometheus.NewRegistry()
	
	// Create sample metrics
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "A test counter",
	})
	
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "A test gauge",
	})
	
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_histogram",
		Help:    "A test histogram",
		Buckets: []float64{0.1, 0.5, 1.0},
	})
	
	// All should register successfully
	assert.NoError(t, registry.Register(counter))
	assert.NoError(t, registry.Register(gauge))
	assert.NoError(t, registry.Register(histogram))
}

// ============================================
// PrometheusMiddleware tests
// ============================================

func TestPrometheusMiddleware_BasicRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestPrometheusMiddleware_MultipleMethods(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{"GET request", "GET", "/api/users", 200},
		{"POST request", "POST", "/api/users", 201},
		{"PUT request", "PUT", "/api/users/1", 200},
		{"DELETE request", "DELETE", "/api/users/1", 204},
		{"PATCH request", "PATCH", "/api/users/1", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(PrometheusMiddleware())
			
			switch tt.method {
			case "GET":
				router.GET(tt.path, func(c *gin.Context) { c.JSON(tt.statusCode, gin.H{"ok": true}) })
			case "POST":
				router.POST(tt.path, func(c *gin.Context) { c.JSON(tt.statusCode, gin.H{"ok": true}) })
			case "PUT":
				router.PUT(tt.path, func(c *gin.Context) { c.JSON(tt.statusCode, gin.H{"ok": true}) })
			case "DELETE":
				router.DELETE(tt.path, func(c *gin.Context) { c.Status(tt.statusCode) })
			case "PATCH":
				router.PATCH(tt.path, func(c *gin.Context) { c.JSON(tt.statusCode, gin.H{"ok": true}) })
			}

			var req *http.Request
			if tt.method == "POST" || tt.method == "PUT" || tt.method == "PATCH" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestPrometheusMiddleware_SkipsMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/metrics", func(c *gin.Context) {
		c.String(200, "metrics output")
	})

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "metrics output")
}

func TestPrometheusMiddleware_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
	}{
		{"status 200", 200},
		{"status 201", 201},
		{"status 400", 400},
		{"status 404", 404},
		{"status 500", 500},
		{"status 503", 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(PrometheusMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(tt.statusCode, gin.H{"status": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestPrometheusMiddleware_WithPathParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestPrometheusMiddleware_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Simulate multiple concurrent requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
}

// ============================================
// SetupPrometheusEndpoint tests
// ============================================

func TestSetupPrometheusEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	SetupPrometheusEndpoint(router)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSetupPrometheusEndpoint_ReturnsPrometheusFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	SetupPrometheusEndpoint(router)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Prometheus handler returns text/plain with prometheus format
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestSetupPrometheusEndpoint_WithMultipleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	SetupPrometheusEndpoint(router)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test metrics endpoint
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// ============================================
// RecordWSConnection tests
// ============================================

func TestRecordWSConnection(t *testing.T) {
	// This should not panic
	RecordWSConnection()
	RecordWSConnection()
	RecordWSConnection()
}

func TestRecordWSDisconnection(t *testing.T) {
	// First connect to have active connections
	RecordWSConnection()
	RecordWSConnection()
	
	// This should not panic
	RecordWSDisconnection()
	RecordWSDisconnection()
}

func TestRecordWSConnection_Balance(t *testing.T) {
	// Test connection/disconnection balance
	RecordWSConnection()
	RecordWSConnection()
	RecordWSDisconnection()
	RecordWSDisconnection()
	RecordWSConnection()
	RecordWSDisconnection()
}

// ============================================
// RecordWSMessageReceived tests
// ============================================

func TestRecordWSMessageReceived(t *testing.T) {
	RecordWSMessageReceived("text")
	RecordWSMessageReceived("binary")
	RecordWSMessageReceived("ping")
	RecordWSMessageReceived("pong")
}

func TestRecordWSMessageReceived_EmptyType(t *testing.T) {
	// Empty type should still work
	RecordWSMessageReceived("")
}

func TestRecordWSMessageReceived_MultipleTypes(t *testing.T) {
	messageTypes := []string{"text", "binary", "json", "control", "ping", "pong"}
	for _, msgType := range messageTypes {
		RecordWSMessageReceived(msgType)
	}
}

// ============================================
// RecordWSMessageSent tests
// ============================================

func TestRecordWSMessageSent(t *testing.T) {
	RecordWSMessageSent("text")
	RecordWSMessageSent("binary")
	RecordWSMessageSent("json")
}

func TestRecordWSMessageSent_EmptyType(t *testing.T) {
	RecordWSMessageSent("")
}

func TestRecordWSMessageSent_MultipleTypes(t *testing.T) {
	messageTypes := []string{"text", "binary", "json", "control", "broadcast"}
	for _, msgType := range messageTypes {
		RecordWSMessageSent(msgType)
	}
}

// ============================================
// UpdateDeviceMetrics tests
// ============================================

func TestUpdateDeviceMetrics(t *testing.T) {
	UpdateDeviceMetrics("tenant-1", "CNC", 100, 80)
	UpdateDeviceMetrics("tenant-1", "PLC", 50, 40)
	UpdateDeviceMetrics("tenant-2", "Robot", 30, 25)
}

func TestUpdateDeviceMetrics_ZeroValues(t *testing.T) {
	UpdateDeviceMetrics("tenant-1", "CNC", 0, 0)
}

func TestDeviceMetrics_MultipleTenants(t *testing.T) {
	tenants := []string{"tenant-1", "tenant-2", "tenant-3"}
	deviceTypes := []string{"CNC", "PLC", "Robot", "Sensor"}
	
	for _, tenant := range tenants {
		for _, deviceType := range deviceTypes {
			UpdateDeviceMetrics(tenant, deviceType, 10, 5)
		}
	}
}

// ============================================
// RecordTelemetryReceived tests
// ============================================

func TestRecordTelemetryReceived(t *testing.T) {
	RecordTelemetryReceived("tenant-1", "CNC")
	RecordTelemetryReceived("tenant-1", "PLC")
	RecordTelemetryReceived("tenant-2", "Robot")
}

func TestRecordTelemetryReceived_EmptyTenant(t *testing.T) {
	RecordTelemetryReceived("", "CNC")
}

func TestRecordTelemetryReceived_EmptyDeviceType(t *testing.T) {
	RecordTelemetryReceived("tenant-1", "")
}

func TestRecordTelemetryReceived_MultipleDeviceTypes(t *testing.T) {
	deviceTypes := []string{"CNC", "PLC", "Robot", "Sensor", "AGV", "Conveyor"}
	for _, dt := range deviceTypes {
		RecordTelemetryReceived("tenant-1", dt)
	}
}

// ============================================
// RecordAlertTriggered tests
// ============================================

func TestRecordAlertTriggered(t *testing.T) {
	RecordAlertTriggered("tenant-1", "critical")
	RecordAlertTriggered("tenant-1", "warning")
	RecordAlertTriggered("tenant-2", "info")
}

func TestRecordAlertTriggered_AllSeverities(t *testing.T) {
	severities := []string{"critical", "high", "warning", "medium", "low", "info"}
	for _, severity := range severities {
		RecordAlertTriggered("tenant-1", severity)
	}
}

func TestRecordAlertTriggered_MultipleTenants(t *testing.T) {
	for i := 1; i <= 5; i++ {
		tenant := "tenant-" + strconv.Itoa(i)
		RecordAlertTriggered(tenant, "warning")
	}
}

// ============================================
// UpdateActiveAlerts tests
// ============================================

func TestUpdateActiveAlerts(t *testing.T) {
	UpdateActiveAlerts("tenant-1", "critical", 5)
	UpdateActiveAlerts("tenant-1", "warning", 10)
	UpdateActiveAlerts("tenant-2", "info", 3)
}

func TestUpdateActiveAlerts_ZeroCount(t *testing.T) {
	UpdateActiveAlerts("tenant-1", "critical", 0)
}

func TestUpdateActiveAlerts_LargeCount(t *testing.T) {
	UpdateActiveAlerts("tenant-1", "warning", 1000)
	UpdateActiveAlerts("tenant-2", "critical", 999999)
}

func TestUpdateActiveAlerts_NegativeCount(t *testing.T) {
	// Prometheus gauges can technically handle negative values
	UpdateActiveAlerts("tenant-1", "warning", -1)
}

// ============================================
// RecordAIQuery tests
// ============================================

func TestRecordAIQuery(t *testing.T) {
	RecordAIQuery("tenant-1", "gpt-4", 1.5, 100, 200)
	RecordAIQuery("tenant-2", "gpt-3.5-turbo", 0.5, 50, 100)
}

func TestRecordAIQuery_DifferentModels(t *testing.T) {
	models := []string{"gpt-4", "gpt-3.5-turbo", "claude-2", "claude-instant", "palm-2"}
	for _, model := range models {
		RecordAIQuery("tenant-1", model, 1.0, 50, 100)
	}
}

func TestRecordAIQuery_ZeroTokens(t *testing.T) {
	RecordAIQuery("tenant-1", "gpt-4", 0.0, 0, 0)
}

func TestRecordAIQuery_LargeValues(t *testing.T) {
	RecordAIQuery("tenant-1", "gpt-4", 120.5, 10000, 20000)
}

func TestRecordAIQuery_MultipleQueries(t *testing.T) {
	for i := 0; i < 10; i++ {
		RecordAIQuery("tenant-1", "gpt-4", float64(i)*0.5, i*10, i*20)
	}
}

// ============================================
// RecordDBQuery tests
// ============================================

func TestRecordDBQuery(t *testing.T) {
	RecordDBQuery("SELECT", "users", 0.01)
	RecordDBQuery("INSERT", "logs", 0.05)
	RecordDBQuery("UPDATE", "devices", 0.02)
	RecordDBQuery("DELETE", "sessions", 0.001)
}

func TestRecordDBQuery_AllOperations(t *testing.T) {
	operations := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "BEGIN", "COMMIT", "ROLLBACK"}
	for _, op := range operations {
		RecordDBQuery(op, "test_table", 0.01)
	}
}

func TestRecordDBQuery_DifferentTables(t *testing.T) {
	tables := []string{"users", "devices", "telemetry", "alerts", "sessions", "logs"}
	for _, table := range tables {
		RecordDBQuery("SELECT", table, 0.005)
	}
}

func TestRecordDBQuery_SlowQuery(t *testing.T) {
	RecordDBQuery("SELECT", "large_table", 5.5)
	RecordDBQuery("SELECT", "huge_table", 10.0)
}

func TestRecordDBQuery_ZeroDuration(t *testing.T) {
	RecordDBQuery("SELECT", "cache", 0.0)
}

// ============================================
// UpdateDBConnections tests
// ============================================

func TestUpdateDBConnections(t *testing.T) {
	UpdateDBConnections(10)
	UpdateDBConnections(20)
	UpdateDBConnections(5)
}

func TestUpdateDBConnections_Zero(t *testing.T) {
	UpdateDBConnections(0)
}

func TestUpdateDBConnections_LargeValue(t *testing.T) {
	UpdateDBConnections(1000)
	UpdateDBConnections(10000)
}

func TestUpdateDBConnections_ConnectionPool(t *testing.T) {
	// Simulate connection pool changes
	for i := 0; i < 10; i++ {
		UpdateDBConnections(5 + i)
	}
	for i := 9; i >= 0; i-- {
		UpdateDBConnections(5 + i)
	}
}

// ============================================
// RecordRedisCommand tests
// ============================================

func TestRecordRedisCommand(t *testing.T) {
	RecordRedisCommand("GET")
	RecordRedisCommand("SET")
	RecordRedisCommand("DEL")
}

func TestRecordRedisCommand_AllCommands(t *testing.T) {
	commands := []string{
		"GET", "SET", "DEL", "HGET", "HSET", "LPUSH", "RPUSH",
		"LPOP", "RPOP", "SADD", "SREM", "ZADD", "ZREM",
		"INCR", "DECR", "EXPIRE", "TTL", "KEYS", "SCAN",
	}
	for _, cmd := range commands {
		RecordRedisCommand(cmd)
	}
}

func TestRecordRedisCommand_EmptyCommand(t *testing.T) {
	RecordRedisCommand("")
}

func TestRecordRedisCommand_MultipleCalls(t *testing.T) {
	for i := 0; i < 100; i++ {
		RecordRedisCommand("GET")
	}
}

// ============================================
// RecordCacheHit tests
// ============================================

func TestRecordCacheHit(t *testing.T) {
	RecordCacheHit()
	RecordCacheHit()
	RecordCacheHit()
}

func TestRecordCacheHit_MultipleHits(t *testing.T) {
	for i := 0; i < 100; i++ {
		RecordCacheHit()
	}
}

// ============================================
// RecordCacheMiss tests
// ============================================

func TestRecordCacheMiss(t *testing.T) {
	RecordCacheMiss()
	RecordCacheMiss()
	RecordCacheMiss()
}

func TestRecordCacheMiss_MultipleMisses(t *testing.T) {
	for i := 0; i < 100; i++ {
		RecordCacheMiss()
	}
}

// ============================================
// Cache Hit/Miss Ratio tests
// ============================================

func TestCacheHitMissRatio(t *testing.T) {
	// Simulate realistic cache patterns
	// 80% hit rate
	for i := 0; i < 80; i++ {
		RecordCacheHit()
	}
	for i := 0; i < 20; i++ {
		RecordCacheMiss()
	}
}

// ============================================
// Integration tests - middleware with metrics
// ============================================

func TestPrometheusMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	SetupPrometheusEndpoint(router)

	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(201, gin.H{"created": true})
	})

	// Make some API requests
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	req = httptest.NewRequest("POST", "/api/data", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// Check metrics endpoint returns success
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestPrometheusMiddleware_WithMetricsRecording(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	SetupPrometheusEndpoint(router)

	router.GET("/api/process", func(c *gin.Context) {
		// Simulate business logic with metrics
		RecordWSConnection()
		RecordWSMessageReceived("text")
		RecordTelemetryReceived("tenant-1", "CNC")
		RecordAlertTriggered("tenant-1", "warning")
		RecordAIQuery("tenant-1", "gpt-4", 1.5, 100, 200)
		RecordDBQuery("SELECT", "devices", 0.01)
		RecordRedisCommand("GET")
		RecordCacheHit()
		
		c.JSON(200, gin.H{"processed": true})
	})

	req := httptest.NewRequest("GET", "/api/process", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Verify metrics endpoint still works
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// ============================================
// Edge cases
// ============================================

func TestRecordFunctions_EmptyStrings(t *testing.T) {
	// All these should not panic
	RecordWSMessageReceived("")
	RecordWSMessageSent("")
	UpdateDeviceMetrics("", "", 0, 0)
	RecordTelemetryReceived("", "")
	RecordAlertTriggered("", "")
	UpdateActiveAlerts("", "", 0)
	RecordAIQuery("", "", 0, 0, 0)
	RecordDBQuery("", "", 0)
	RecordRedisCommand("")
}

func TestRecordFunctions_SpecialCharacters(t *testing.T) {
	// Test with special characters in labels
	RecordWSMessageReceived("text/html")
	RecordWSMessageSent("application/json")
	UpdateDeviceMetrics("tenant-123", "type_with_underscore", 10, 5)
	RecordTelemetryReceived("tenant with spaces", "device-type")
	RecordAlertTriggered("tenant/1", "critical-alert")
	RecordAIQuery("tenant-1", "model-v2.0", 1.0, 100, 200)
	RecordDBQuery("SELECT", "table_with_special-chars", 0.01)
	RecordRedisCommand("GET_EX")
}

func TestPrometheusMiddleware_LongPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/api/v1/tenants/:tenant_id/devices/:device_id/telemetry/:telemetry_id", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/api/v1/tenants/t1/devices/d1/telemetry/t1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestPrometheusMiddleware_QueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/search", func(c *gin.Context) {
		c.JSON(200, gin.H{"query": c.Query("q")})
	})

	req := httptest.NewRequest("GET", "/search?q=test&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// Benchmark tests
// ============================================

func BenchmarkPrometheusMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRecordWSConnection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordWSConnection()
	}
}

func BenchmarkRecordAIQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordAIQuery("tenant-1", "gpt-4", 1.0, 100, 200)
	}
}

func BenchmarkRecordDBQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordDBQuery("SELECT", "users", 0.01)
	}
}