package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/industrial-ai/platform/pkg/logger"
)

// ============================================
// 结构化日志中间件
// ============================================

// LoggingMiddleware 结构化日志中间件
// 用途: 记录所有 HTTP 请求的结构化日志
func LoggingMiddleware() gin.HandlerFunc {
	log := logger.L()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		requestSize := c.Request.ContentLength
		responseSize := c.Writer.Size()

		// 构建基础日志字段
		fields := []zap.Field{
			zap.String("http.method", method),
			zap.String("http.path", path),
			zap.Int("http.status_code", statusCode),
			zap.Float64("http.latency_ms", float64(latency.Milliseconds())),
			zap.Int64("http.request_size_bytes", requestSize),
			zap.Int("response_size_bytes", responseSize),
			zap.String("http.client_ip", c.ClientIP()),
			zap.String("http.user_agent", c.Request.UserAgent()),
		}

		// 添加追踪 ID (如果有)
		if traceID := c.GetString("trace_id"); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		// 添加请求 ID (如果有)
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// 添加租户 ID (如果有)
		if tenantID := c.GetString("tenant_id"); tenantID != "" {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}

		// 添加用户 ID (如果有)
		if userID := c.GetString("user_id"); userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// 添加错误信息 (如果有)
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("http.errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			log.Error("HTTP request error", fields...)
		} else if statusCode >= 400 {
			log.Warn("HTTP request warning", fields...)
		} else {
			log.Info("HTTP request processed", fields...)
		}
	}
}

// ============================================
// 请求 ID 中间件
// ============================================

// RequestIDMiddleware 请求 ID 中间件
// 用途: 为每个请求生成唯一的请求 ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求 ID
			requestID = generateRequestID()
		}

		// 设置到上下文
		c.Set("request_id", requestID)

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ============================================
// 追踪 ID 中间件
// ============================================

// TraceIDMiddleware 追踪 ID 中间件
// 用途: 从请求头提取或生成分布式追踪 ID
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取追踪 ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			// 尝试从 X-B3-Traceid 获取 (Zipkin 格式)
			traceID = c.GetHeader("X-B3-Traceid")
		}
		if traceID == "" {
			// 尝试从 traceparent 获取 (W3C 格式)
			traceparent := c.GetHeader("traceparent")
			if traceparent != "" {
				// 解析 W3C traceparent 格式: version-traceid-spanid-flags
				parts := splitString(traceparent, "-")
				if len(parts) >= 2 {
					traceID = parts[1]
				}
			}
		}

		// 如果没有追踪 ID，生成新的
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 设置到上下文
		c.Set("trace_id", traceID)

		// 设置响应头
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// generateTraceID 生成追踪 ID
func generateTraceID() string {
	// 生成 16 字节的追踪 ID
	return randomString(16)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[randomInt(len(letters))]
	}
	return string(b)
}

// randomInt 生成随机整数
func randomInt(max int) int {
	// 使用简单的伪随机实现
	// 注意：对于生产环境，应使用 crypto/rand
	seed := time.Now().UnixNano()
	return int(seed % int64(max))
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// ============================================
// 慢请求日志中间件
// ============================================

// SlowRequestLogMiddleware 慢请求日志中间件
// 用途: 记录超过阈值的慢请求
func SlowRequestLogMiddleware(thresholdMs int) gin.HandlerFunc {
	log := logger.L()

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		latencyMs := int(latency.Milliseconds())

		// 检查是否超过阈值
		if latencyMs > thresholdMs {
			fields := []zap.Field{
				zap.String("http.method", c.Request.Method),
				zap.String("http.path", c.Request.URL.Path),
				zap.Int("http.status_code", c.Writer.Status()),
				zap.Int("http.latency_ms", latencyMs),
				zap.Int("threshold_ms", thresholdMs),
				zap.String("http.client_ip", c.ClientIP()),
			}

			// 添加追踪信息
			if traceID := c.GetString("trace_id"); traceID != "" {
				fields = append(fields, zap.String("trace_id", traceID))
			}

			log.Warn("Slow request detected", fields...)
		}
	}
}

// ============================================
// 错误日志中间件
// ============================================

// ErrorLogMiddleware 错误日志中间件
// 用途: 记录所有错误响应的详细日志
func ErrorLogMiddleware() gin.HandlerFunc {
	log := logger.L()

	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				fields := []zap.Field{
					zap.String("http.method", c.Request.Method),
					zap.String("http.path", c.Request.URL.Path),
					zap.Int("http.status_code", c.Writer.Status()),
					zap.Uint64("error_type", uint64(err.Type)),
					zap.String("error.message", err.Error()),
					zap.String("http.client_ip", c.ClientIP()),
				}

				// 添加追踪信息
				if traceID := c.GetString("trace_id"); traceID != "" {
					fields = append(fields, zap.String("trace_id", traceID))
				}

				log.Error("HTTP request error", fields...)
			}
		}
	}
}
