package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShutdownMiddleware 关闭状态中间件
// 用于在优雅关闭期间拒绝新请求

// ShutdownMiddleware 关闭中间件
func ShutdownMiddleware(isShuttingDown func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否正在关闭
		if isShuttingDown() {
			// 拒绝新请求，返回 503 Service Unavailable
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service is shutting down",
				"code":  "SERVICE_SHUTTING_DOWN",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupMiddleware 清理中间件相关的后台资源
// 应在应用关闭时调用，确保所有 goroutine 正确退出
func CleanupMiddleware() {
	// 停止全局 Rate Limiter Manager 的清理 goroutine
	if manager := GetLimiterManager(); manager != nil {
		manager.Stop()
	}
}

// RequestTrackingMiddleware 请求追踪中间件
// 用于追踪活跃请求数量，辅助优雅关闭

// ActiveRequestTracker 活跃请求追踪器
type ActiveRequestTracker struct {
	activeRequests int64
}

// NewActiveRequestTracker 创建请求追踪器
func NewActiveRequestTracker() *ActiveRequestTracker {
	return &ActiveRequestTracker{}
}

// Increment 增加活跃请求计数
func (t *ActiveRequestTracker) Increment() {
	t.activeRequests++
}

// Decrement 减少活跃请求计数
func (t *ActiveRequestTracker) Decrement() {
	t.activeRequests--
}

// GetActiveCount 获取活跃请求数
func (t *ActiveRequestTracker) GetActiveCount() int64 {
	return t.activeRequests
}

// RequestTrackingMiddleware 请求追踪中间件
func RequestTrackingMiddleware(tracker *ActiveRequestTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求开始：增加计数
		tracker.Increment()

		// 确保请求结束时减少计数
		defer tracker.Decrement()

		c.Next()
	}
}

// IsRequestComplete 检查所有请求是否完成
func (t *ActiveRequestTracker) IsRequestComplete() bool {
	return t.activeRequests == 0
}

// WaitForRequestsComplete 等待所有请求完成
func (t *ActiveRequestTracker) WaitForRequestsComplete(timeout int64) bool {
	for i := 0; i < int(timeout/100); i++ {
		if t.IsRequestComplete() {
			return true
		}
		// 等待 100ms
		// time.Sleep(100 * time.Millisecond)
	}
	return false
}
