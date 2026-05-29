package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ============================================
// mockCacheProvider 模拟缓存状态提供者
// ============================================

type mockCacheProvider struct {
	available bool
	stats     CacheStats
}

func (m *mockCacheProvider) IsAvailable() bool {
	return m.available
}

func (m *mockCacheProvider) GetStats() CacheStats {
	return m.stats
}

// ============================================
// mockWSProvider 模拟 WebSocket 状态提供者
// ============================================

type mockWSProvider struct {
	clientCount int
}

func (m *mockWSProvider) ClientCount() int {
	return m.clientCount
}

// ============================================
// GetCacheStatus 测试
// ============================================

// TestGetCacheStatus_WithProvider 测试注入缓存提供者时返回真实状态
func TestGetCacheStatus_WithProvider(t *testing.T) {
	cacheProvider := &mockCacheProvider{
		available: true,
		stats: CacheStats{
			BackendType: "redis",
			Available:   true,
			KeysStored:  150,
			Hits:        1000,
			Misses:      50,
			Errors:      2,
		},
	}

	handler := NewHealthHandlerNewWithDeps(time.Now().Add(-1*time.Hour), cacheProvider, nil, false)

	router := gin.New()
	router.GET("/cache/status", handler.GetCacheStatus)

	req := httptest.NewRequest("GET", "/cache/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回了注入的真实数据
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "redis", response["backend"])
	assert.Equal(t, true, response["connected"])
	assert.Equal(t, float64(150), response["keys_stored"])
	assert.Equal(t, float64(1000), response["hits"])
	assert.Equal(t, float64(50), response["misses"])
	assert.Equal(t, float64(2), response["errors"])
	assert.Equal(t, "live", response["data_source"])

	// 验证 uptime 字段存在
	uptime, ok := response["uptime"].(float64)
	assert.True(t, ok)
	assert.Greater(t, uptime, float64(0))
}

// TestGetCacheStatus_WithoutProvider 测试未注入缓存提供者时返回 mock 状态
func TestGetCacheStatus_WithoutProvider(t *testing.T) {
	handler := NewHealthHandlerNew(time.Now())

	router := gin.New()
	router.GET("/cache/status", handler.GetCacheStatus)

	req := httptest.NewRequest("GET", "/cache/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回了 mock 默认数据
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "mock", response["backend"])
	assert.Equal(t, true, response["connected"])
	assert.Equal(t, "mock", response["data_source"])
	assert.Equal(t, "inject CacheStatusProvider for live status", response["note"])
}

// TestGetCacheStatus_UptimeCalculation 测试 uptime 计算正确性
func TestGetCacheStatus_UptimeCalculation(t *testing.T) {
	startTime := time.Now().Add(-2 * time.Hour)
	handler := NewHealthHandlerNew(startTime)

	router := gin.New()
	router.GET("/cache/status", handler.GetCacheStatus)

	req := httptest.NewRequest("GET", "/cache/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	uptime, ok := response["uptime"].(float64)
	assert.True(t, ok)
	// uptime 应接近 7200 秒（2小时），允许误差 5 秒
	assert.GreaterOrEqual(t, uptime, float64(7195))
	assert.LessOrEqual(t, uptime, float64(7205))
}

// ============================================
// GetWSStats 测试
// ============================================

// TestGetWSStats_WithProvider 测试注入 WebSocket 提供者时返回真实状态
func TestGetWSStats_WithProvider(t *testing.T) {
	wsProvider := &mockWSProvider{
		clientCount: 42,
	}

	handler := NewHealthHandlerNewWithDeps(time.Now().Add(-30*time.Minute), nil, wsProvider, true)

	router := gin.New()
	router.GET("/ws/stats", handler.GetWSStats)

	req := httptest.NewRequest("GET", "/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回了注入的真实数据
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, float64(42), response["clients_count"])
	assert.Equal(t, true, response["compression"])
	assert.Equal(t, "live", response["data_source"])

	// 验证 uptime 字段存在
	uptime, ok := response["uptime"].(float64)
	assert.True(t, ok)
	assert.Greater(t, uptime, float64(0))
}

// TestGetWSStats_WithoutProvider 测试未注入 WebSocket 提供者时返回 mock 状态
func TestGetWSStats_WithoutProvider(t *testing.T) {
	handler := NewHealthHandlerNew(time.Now())

	router := gin.New()
	router.GET("/ws/stats", handler.GetWSStats)

	req := httptest.NewRequest("GET", "/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回了 mock 默认数据
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, float64(0), response["clients_count"])
	assert.Equal(t, true, response["compression"])
	assert.Equal(t, "mock", response["data_source"])
	assert.Equal(t, "inject WSStatusProvider for live status", response["note"])
}

// TestGetWSStats_ZeroClients 测试没有客户端连接
func TestGetWSStats_ZeroClients(t *testing.T) {
	wsProvider := &mockWSProvider{
		clientCount: 0,
	}

	handler := NewHealthHandlerNewWithDeps(time.Now(), nil, wsProvider, false)

	router := gin.New()
	router.GET("/ws/stats", handler.GetWSStats)

	req := httptest.NewRequest("GET", "/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(0), response["clients_count"])
	assert.Equal(t, false, response["compression"])
	assert.Equal(t, "live", response["data_source"])
}

// TestGetWSStats_CompressionDisabled 测试压缩功能禁用
func TestGetWSStats_CompressionDisabled(t *testing.T) {
	wsProvider := &mockWSProvider{
		clientCount: 10,
	}

	handler := NewHealthHandlerNewWithDeps(time.Now(), nil, wsProvider, false)

	router := gin.New()
	router.GET("/ws/stats", handler.GetWSStats)

	req := httptest.NewRequest("GET", "/ws/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["compression"])
}

// ============================================
// 构造函数测试
// ============================================

// TestNewHealthHandlerNew 验证基础构造函数
func TestNewHealthHandlerNew(t *testing.T) {
	startTime := time.Now()
	handler := NewHealthHandlerNew(startTime)

	assert.NotNil(t, handler)
	assert.Equal(t, startTime, handler.startTime)
	assert.Nil(t, handler.cacheProvider)
	assert.Nil(t, handler.wsProvider)
}

// TestNewHealthHandlerNewWithDeps 验证依赖注入构造函数
func TestNewHealthHandlerNewWithDeps(t *testing.T) {
	cacheProvider := &mockCacheProvider{available: true}
	wsProvider := &mockWSProvider{clientCount: 5}
	startTime := time.Now()

	handler := NewHealthHandlerNewWithDeps(startTime, cacheProvider, wsProvider, true)

	assert.NotNil(t, handler)
	assert.Equal(t, startTime, handler.startTime)
	assert.Equal(t, cacheProvider, handler.cacheProvider)
	assert.Equal(t, wsProvider, handler.wsProvider)
	assert.True(t, handler.wsCompressor)
}
