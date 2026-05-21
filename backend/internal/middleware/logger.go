package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Color codes for terminal output
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
)

// Logger middleware with colored output
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

		// Choose color based on status code
		statusColor := Green
		methodColor := Cyan
		pathColor := Gray

		switch {
		case status >= 500:
			statusColor = Red
		case status >= 400:
			statusColor = Yellow
		case status >= 300:
			statusColor = Magenta
		}

		// Format output
		timestamp := time.Now().Format("2006/01/02 - 15:04:05")
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Log the request
		fmt.Printf("%s |%s %3d %s| %13v | %15s |%s %-7s %s %s%s%s\n",
			timestamp,
			statusColor, status, Reset,
			latency,
			clientIP,
			methodColor, method, Reset,
			pathColor, path, Reset,
		)

		// Log query params if present
		if query != "" {
			fmt.Printf("         ├─ Query: %s%s%s\n", Cyan, query, Reset)
		}

		// Log errors if present
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				fmt.Printf("         ├─ Error: %s%s%s\n", Red, err.Error(), Reset)
			}
		}

		// Log request ID for tracing
		if requestID != nil {
			fmt.Printf("         └─ RequestID: %s%s%s\n", Gray, requestID.(string), Reset)
		}
	}
}

// Recovery middleware with colored error output
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with color
				fmt.Printf("\n%s========== PANIC RECOVERED ==========%s\n", Red, Reset)
				fmt.Printf("%sTime:%s     %s\n", Yellow, Reset, time.Now().Format("2006/01/02 - 15:04:05"))
				fmt.Printf("%sPath:%s     %s %s\n", Yellow, Reset, c.Request.Method, c.Request.URL.Path)
				fmt.Printf("%sClient:%s   %s\n", Yellow, Reset, c.ClientIP())
				fmt.Printf("%sError:%s    %v\n", Red, Reset, err)
				fmt.Printf("%sStack:%s\n", Yellow, Reset)
				// You might want to print stack trace here
				fmt.Printf("%s=====================================%s\n\n", Red, Reset)

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
