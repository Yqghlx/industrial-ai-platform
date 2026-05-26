package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// Logger middleware with structured logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Writer.Status()

		// Get request ID
		requestID, _ := c.Get("request_id")

		// Log the request with structured logging
		logger.L().Info("HTTP Request",
			zap.String("timestamp", time.Now().Format("2006/01/02 - 15:04:05")),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
		)

		// Log query params if present
		if query != "" {
			logger.L().Debug("Query params", zap.String("query", query))
		}

		// Log errors if present
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.L().Error("Request error", zap.String("error", err.Error()))
			}
		}

		// Log request ID for tracing
		if requestID != nil {
			logger.L().Debug("Request ID", zap.String("request_id", requestID.(string)))
		}
	}
}

// Recovery middleware with structured logging
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with structured logging
				logger.L().Error("PANIC RECOVERED",
					zap.String("time", time.Now().Format("2006/01/02 - 15:04:05")),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("error", err),
				)

				c.JSON(500, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
