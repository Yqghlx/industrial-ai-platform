// Package benchmarks provides performance benchmark tests for the Industrial AI Platform
// P2-001: API Response Time Benchmarks
// These benchmarks test API endpoint performance using httptest
package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Device represents a device entity for benchmark tests
type Device struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TelemetryData represents telemetry data for benchmark tests
type TelemetryData struct {
	DeviceID    string                 `json:"device_id"`
	DeviceType  string                 `json:"device_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Metrics     map[string]interface{} `json:"metrics"`
	Status      string                 `json:"status"`
}

// AlertRule represents an alert rule for benchmark tests
type AlertRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Severity  string `json:"severity"`
	Enabled   bool   `json:"enabled"`
}

// LoginRequest represents login request for benchmark tests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BenchmarkAPIHealthCheck benchmarks the health check endpoint
func BenchmarkAPIHealthCheck(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPILogin benchmarks the authentication login endpoint
func BenchmarkAPILogin(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		// Simulate auth processing
		time.Sleep(1 * time.Millisecond) // Simulate password hash comparison
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"token": "benchmark-test-token",
			},
		})
	})

	loginPayload := LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	body, _ := json.Marshal(loginPayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIDeviceList benchmarks listing devices with pagination
func BenchmarkAPIDeviceList(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/devices", func(c *gin.Context) {
		// Simulate database query with mock data
		devices := make([]Device, 100)
		for i := 0; i < 100; i++ {
			devices[i] = Device{
				ID:        fmt.Sprintf("device-%d", i),
				Name:      fmt.Sprintf("Test Device %d", i),
				Type:      "CNC",
				Location:  "Factory-A",
				Status:    "running",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    devices,
			"total":   100,
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/devices?page=1&limit=100", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIDeviceGet benchmarks retrieving a single device
func BenchmarkAPIDeviceGet(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/devices/:id", func(c *gin.Context) {
		device := Device{
			ID:        c.Param("id"),
			Name:      "Test Device",
			Type:      "CNC",
			Location:  "Factory-A",
			Status:    "running",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    device,
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/device-001", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPITelemetrySubmit benchmarks submitting telemetry data
func BenchmarkAPITelemetrySubmit(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/devices/telemetry", func(c *gin.Context) {
		var telemetry TelemetryData
		if err := c.ShouldBindJSON(&telemetry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		// Simulate data processing
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"device_id": telemetry.DeviceID,
				"received":  true,
			},
		})
	})

	telemetryPayload := TelemetryData{
		DeviceID:    "device-bench-001",
		DeviceType:  "CNC",
		Timestamp:   time.Now(),
		Metrics: map[string]interface{}{
			"temperature":       75.5,
			"vibration":         2.3,
			"pressure":          120.0,
			"power_consumption": 550.0,
		},
		Status: "running",
	}
	body, _ := json.Marshal(telemetryPayload)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/telemetry", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPITelemetryLatest benchmarks getting latest telemetry
func BenchmarkAPITelemetryLatest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/devices/latest", func(c *gin.Context) {
		// Simulate returning latest telemetry
		latestData := make([]map[string]interface{}, 10)
		for i := 0; i < 10; i++ {
			latestData[i] = map[string]interface{}{
				"device_id": fmt.Sprintf("device-%d", i),
				"metrics": map[string]interface{}{
					"temperature": 75.0 + float64(i),
					"vibration":   2.0 + float64(i)*0.1,
				},
				"timestamp": time.Now().Unix(),
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    latestData,
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/latest", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIRulesList benchmarks listing alert rules
func BenchmarkAPIRulesList(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/rules", func(c *gin.Context) {
		rules := make([]AlertRule, 50)
		for i := 0; i < 50; i++ {
			rules[i] = AlertRule{
				ID:        fmt.Sprintf("rule-%d", i),
				Name:      fmt.Sprintf("Temperature Alert %d", i),
				Condition: "temperature > 80",
				Severity:  "high",
				Enabled:   true,
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    rules,
			"total":   50,
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/rules", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIROIStats benchmarks ROI statistics endpoint (cached)
func BenchmarkAPIROIStats(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/roi/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"total_savings":            150000.00,
				"energy_reduction":         25.5,
				"maintenance_optimization": 35.2,
				"uptime_improvement":       15.8,
				"ai_queries_processed":     5000,
				"alerts_prevented":         120,
			},
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/roi/stat", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIJSONSerialization benchmarks JSON serialization performance
func BenchmarkAPIJSONSerialization(b *testing.B) {
	data := gin.H{
		"success": true,
		"data": gin.H{
			"devices": make([]Device, 1000),
			"total":   1000,
		},
	}
	
	for i := 0; i < 1000; i++ {
		data["data"].(gin.H)["devices"].([]Device)[i] = Device{
			ID:        fmt.Sprintf("device-%d", i),
			Name:      fmt.Sprintf("Device %d", i),
			Type:      "CNC",
			Location:  fmt.Sprintf("Location-%d", i%10),
			Status:    "running",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

// BenchmarkAPIJSONDeserialization benchmarks JSON deserialization performance
func BenchmarkAPIJSONDeserialization(b *testing.B) {
	payload := TelemetryData{
		DeviceID:    "device-001",
		DeviceType:  "CNC",
		Timestamp:   time.Now(),
		Metrics: map[string]interface{}{
			"temperature":       75.5,
			"vibration":         2.3,
			"pressure":          120.0,
			"power_consumption": 550.0,
		},
		Status: "running",
	}
	body, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data TelemetryData
		_ = json.Unmarshal(body, &data)
	}
}

// BenchmarkAPIMiddlewareChain benchmarks middleware chain performance
func BenchmarkAPIMiddlewareChain(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	
	// Add multiple middlewares
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "bench-001")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		c.Set("start_time", time.Now())
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		// Simulate auth check
		c.Set("user_id", "user-001")
		c.Next()
	})
	
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"request_id": c.GetString("request_id"),
			"user_id":    c.GetString("user_id"),
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIConcurrentRequests benchmarks concurrent request handling
func BenchmarkAPIConcurrentRequests(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/concurrent", func(c *gin.Context) {
		// Simulate some processing
		time.Sleep(100 * time.Microsecond)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/concurrent", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPILargePayload benchmarks handling large JSON payloads
func BenchmarkAPILargePayload(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/bulk", func(c *gin.Context) {
		var data []TelemetryData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"received": len(data),
		})
	})

	// Create large payload with 500 telemetry entries
	largePayload := make([]TelemetryData, 500)
	for i := 0; i < 500; i++ {
		largePayload[i] = TelemetryData{
			DeviceID:    fmt.Sprintf("device-%d", i),
			DeviceType:  "CNC",
			Timestamp:   time.Now(),
			Metrics: map[string]interface{}{
				"temperature": float64(70 + i%20),
				"vibration":   float64(2 + i%3),
			},
			Status: "running",
		}
	}
	body, _ := json.Marshal(largePayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bulk", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIMetricsCollection benchmarks Prometheus metrics collection
func BenchmarkAPIMetricsCollection(b *testing.B) {
	var mu sync.Mutex
	metrics := make(map[string]int64)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			metrics["requests_total"]++
			metrics["requests_in_progress"]++
			metrics["requests_in_progress"]--
			mu.Unlock()
		}
	})
}

// BenchmarkAPIRequestWithContext benchmarks request with context handling
func BenchmarkAPIRequestWithContext(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/context", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		
		// Simulate work with context
		select {
		case <-time.After(1 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"success": true})
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/context", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIResponseWriting benchmarks response writing performance
func BenchmarkAPIResponseWriting(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	// Large response payload
	largeData := make([]byte, 64*1024) // 64KB
	for i := 0; i < len(largeData); i++ {
		largeData[i] = byte(i % 256)
	}
	
	router := gin.New()
	router.GET("/api/v1/large", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/octet-stream", largeData)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/large", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIStreamResponse benchmarks streaming response performance
func BenchmarkAPIStreamResponse(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/stream", func(c *gin.Context) {
		c.Stream(func(w io.Writer) bool {
			data := []byte(fmt.Sprintf(`{"data":"item-%d"}`, time.Now().UnixNano()%1000))
			w.Write(data)
			return false // Stop streaming
		})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIDeviceCreate benchmarks device creation endpoint
func BenchmarkAPIDeviceCreate(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/devices", func(c *gin.Context) {
		var device Device
		if err := c.ShouldBindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		device.CreatedAt = time.Now()
		device.UpdatedAt = time.Now()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    device,
		})
	})

	devicePayload := Device{
		ID:          "device-new-001",
		Name:        "New Benchmark Device",
		Type:        "CNC",
		Location:    "Factory-B",
		Status:      "active",
		Description: "Benchmark test device",
	}
	body, _ := json.Marshal(devicePayload)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIDeviceUpdate benchmarks device update endpoint
func BenchmarkAPIDeviceUpdate(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.PUT("/api/v1/devices/:id", func(c *gin.Context) {
		var device Device
		if err := c.ShouldBindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		device.ID = c.Param("id")
		device.UpdatedAt = time.Now()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    device,
		})
	})

	devicePayload := Device{
		Name:        "Updated Device",
		Type:        "PLC",
		Location:    "Factory-C",
		Status:      "maintenance",
		Description: "Updated benchmark device",
	}
	body, _ := json.Marshal(devicePayload)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPut, "/api/v1/devices/device-001", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIDeviceDelete benchmarks device deletion endpoint
func BenchmarkAPIDeviceDelete(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.DELETE("/api/v1/devices/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Device deleted successfully",
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/devices/device-001", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAPIMetricsEndpoint benchmarks Prometheus metrics endpoint
func BenchmarkAPIMetricsEndpoint(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/metrics", func(c *gin.Context) {
		// Simulate Prometheus metrics output
		metricsOutput := `# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",endpoint="/api/v1/devices"} 1234
# HELP http_request_duration_seconds HTTP request latency
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 100
http_request_duration_seconds_bucket{le="0.5"} 200
http_request_duration_seconds_bucket{le="1.0"} 300
http_request_duration_seconds_bucket{le="+Inf"} 400
http_request_duration_seconds_sum 150.5
http_request_duration_seconds_count 400
# HELP ws_connections_active Active WebSocket connections
# TYPE ws_connections_active gauge
ws_connections_active 50
`
		c.String(http.StatusOK, metricsOutput)
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}