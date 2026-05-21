package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/pkg/circuitbreaker"
)

// CircuitBreakerMiddleware 熔断中间件
func CircuitBreakerMiddleware(cb *circuitbreaker.CircuitBreaker, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查熔断器状态
		if cb.GetState() == circuitbreaker.StateOpen {
			// 熔断状态，返回降级响应
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":      "degraded",
				"message":     serviceName + " service temporarily unavailable",
				"fallback":    gin.H{"available": false},
				"retry_after": int(time.Until(cb.GetStats().LastStateChange.Add(30 * time.Second)).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CircuitBreakerHandler 熔断器处理器 (用于外部服务调用)
type CircuitBreakerHandler struct {
	manager *circuitbreaker.CircuitBreakerManager
}

// NewCircuitBreakerHandler 创建熔断器处理器
func NewCircuitBreakerHandler(manager *circuitbreaker.CircuitBreakerManager) *CircuitBreakerHandler {
	return &CircuitBreakerHandler{
		manager: manager,
	}
}

// GetCircuitBreakerStatus 获取熔断器状态
func (h *CircuitBreakerHandler) GetCircuitBreakerStatus(c *gin.Context) {
	stats := h.manager.GetAllStats()

	c.JSON(http.StatusOK, gin.H{
		"circuit_breakers": stats,
		"timestamp":        time.Now().Format(time.RFC3339),
	})
}

// ForceOpenCircuitBreaker 强制打开熔断器
func (h *CircuitBreakerHandler) ForceOpenCircuitBreaker(c *gin.Context) {
	name := c.Param("name")

	cb := h.manager.Get(name)
	if cb == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "circuit breaker not found",
		})
		return
	}

	cb.ForceOpen()

	c.JSON(http.StatusOK, gin.H{
		"message":   "circuit breaker " + name + " forced open",
		"state":     cb.GetState().String(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ForceCloseCircuitBreaker 强制关闭熔断器
func (h *CircuitBreakerHandler) ForceCloseCircuitBreaker(c *gin.Context) {
	name := c.Param("name")

	cb := h.manager.Get(name)
	if cb == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "circuit breaker not found",
		})
		return
	}

	cb.ForceClose()

	c.JSON(http.StatusOK, gin.H{
		"message":   "circuit breaker " + name + " forced close",
		"state":     cb.GetState().String(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// === 降级响应辅助 ===

// DegradedResponse 降级响应
type DegradedResponse struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Fallback   map[string]interface{} `json:"fallback"`
	RetryAfter int                    `json:"retry_after"`
}

// NewDegradedResponse 创建降级响应
func NewDegradedResponse(serviceName string, fallbackData map[string]interface{}) *DegradedResponse {
	return &DegradedResponse{
		Status:     "degraded",
		Message:    serviceName + " service degraded, using fallback data",
		Fallback:   fallbackData,
		RetryAfter: 30,
	}
}

// WriteDegradedResponse 写入降级响应
func WriteDegradedResponse(c *gin.Context, serviceName string, fallbackData map[string]interface{}) {
	resp := NewDegradedResponse(serviceName, fallbackData)
	c.JSON(http.StatusServiceUnavailable, resp)
}

// WriteDegradedResponseJSON 写入降级 JSON 响应
func WriteDegradedResponseJSON(c *gin.Context, serviceName string, data interface{}) {
	resp := gin.H{
		"status":      "degraded",
		"message":     serviceName + " service degraded, using cached data",
		"fallback":    data,
		"source":      "cache",
		"timestamp":   time.Now().Format(time.RFC3339),
		"retry_after": 30,
	}
	c.JSON(http.StatusServiceUnavailable, resp)
}

// === 服务降级策略 ===

// DegradationStrategy 降级策略
type DegradationStrategy struct {
	Name         string
	Priority     int // 0-4, 0 最高
	FallbackFunc func() interface{}
}

// DegradationManager 降级管理器
type DegradationManager struct {
	strategies map[string]*DegradationStrategy
}

// NewDegradationManager 创建降级管理器
func NewDegradationManager() *DegradationManager {
	return &DegradationManager{
		strategies: make(map[string]*DegradationStrategy),
	}
}

// RegisterStrategy 注册降级策略
func (m *DegradationManager) RegisterStrategy(strategy *DegradationStrategy) {
	m.strategies[strategy.Name] = strategy
}

// ExecuteFallback 执行降级回调
func (m *DegradationManager) ExecuteFallback(serviceName string) interface{} {
	strategy := m.strategies[serviceName]
	if strategy == nil {
		return gin.H{"error": "no fallback strategy"}
	}
	return strategy.FallbackFunc()
}

// === 降级响应示例 ===

// AIResponseFallback AI 响应降级
func AIResponseFallback() interface{} {
	return gin.H{
		"response": "AI service is temporarily unavailable. Please try again later.",
		"source":   "fallback",
	}
}

// DeviceListFallback 设备列表降级
func DeviceListFallback() interface{} {
	return gin.H{
		"devices": []interface{}{},
		"message": "Device list unavailable, please use cached data",
		"source":  "fallback",
	}
}

// TelemetryFallback 遥测数据降级
func TelemetryFallback() interface{} {
	return gin.H{
		"telemetry": []interface{}{},
		"message":   "Telemetry data unavailable",
		"source":    "fallback",
	}
}

// RegisterDefaultStrategies 注册默认降级策略
func RegisterDefaultStrategies(m *DegradationManager) {
	m.RegisterStrategy(&DegradationStrategy{
		Name:         "glm_api",
		Priority:     2,
		FallbackFunc: AIResponseFallback,
	})

	m.RegisterStrategy(&DegradationStrategy{
		Name:         "device_list",
		Priority:     1,
		FallbackFunc: DeviceListFallback,
	})

	m.RegisterStrategy(&DegradationStrategy{
		Name:         "telemetry",
		Priority:     2,
		FallbackFunc: TelemetryFallback,
	})
}

// === 熔断器路由注册 ===

// RegisterCircuitBreakerRoutes 注册熔断器路由
func (h *CircuitBreakerHandler) RegisterCircuitBreakerRoutes(r *gin.Engine) {
	cb := r.Group("/circuit-breaker")
	{
		cb.GET("/status", h.GetCircuitBreakerStatus)
		cb.POST("/:name/open", h.ForceOpenCircuitBreaker)
		cb.POST("/:name/close", h.ForceCloseCircuitBreaker)
	}
}

// marshalFallbackData 序列化降级数据
// marshalFallbackData is kept for future use
// nolint:unused
func marshalFallbackData(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
