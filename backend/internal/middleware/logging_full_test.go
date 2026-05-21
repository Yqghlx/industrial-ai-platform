package middleware

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// ErrorLogMiddleware Full Tests
// ============================================

func TestErrorLogMiddleware_WithErrors(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("test error 1"))
		c.Error(errors.New("test error 2"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_WithErrorType(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("binding error")).SetType(gin.ErrorTypeBind)
		c.JSON(400, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestErrorLogMiddleware_WithPrivateError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("private error")).SetType(gin.ErrorTypePrivate)
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_WithPublicError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("public error")).SetType(gin.ErrorTypePublic)
		c.JSON(400, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestErrorLogMiddleware_WithContextValues(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "trace-123")
		c.Set("request_id", "req-456")
		c.Set("tenant_id", "tenant-789")
		c.Set("user_id", "user-001")
		c.Next()
	})
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("error with context"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_MultipleErrors(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Add multiple errors of different types
		c.Error(errors.New("error 1")).SetType(gin.ErrorTypeBind)
		c.Error(errors.New("error 2")).SetType(gin.ErrorTypePrivate)
		c.Error(errors.New("error 3")).SetType(gin.ErrorTypePublic)
		c.JSON(500, gin.H{"errors": "multiple"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_WithClientIP(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("error from client"))
		c.JSON(400, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestErrorLogMiddleware_WithUserAgent(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("error with user agent"))
		c.JSON(400, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestClient/1.0")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestErrorLogMiddleware_PostRequest(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.Error(errors.New("post error"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_PutRequest(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.PUT("/test", func(c *gin.Context) {
		c.Error(errors.New("put error"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_DeleteRequest(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.DELETE("/test", func(c *gin.Context) {
		c.Error(errors.New("delete error"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_WithTraceID(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, exists := c.Get("trace_id")
		assert.True(t, exists)
		assert.NotEmpty(t, traceID)
		c.Error(errors.New("error with trace"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_WithProvidedTraceID(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Set("trace_id", "provided-trace-123")
		c.Error(errors.New("error"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", "header-trace-456")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestErrorLogMiddleware_AbortedRequest(t *testing.T) {
	router := gin.New()
	router.Use(ErrorLogMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("abort error"))
		c.AbortWithStatus(403)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

// ============================================
// SlowRequestLogMiddleware Full Tests
// ============================================

func TestSlowRequestLogMiddleware_SlowRequest(t *testing.T) {
	router := gin.New()
	router.Use(SlowRequestLogMiddleware(1)) // 1ms threshold - everything is "slow"
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware_FastRequest(t *testing.T) {
	router := gin.New()
	router.Use(SlowRequestLogMiddleware(10000)) // 10s threshold - nothing is "slow"
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware_WithTraceID(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.Use(SlowRequestLogMiddleware(1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware_WithContextTraceID(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "slow-trace-123")
		c.Next()
	})
	router.Use(SlowRequestLogMiddleware(1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware_PostRequest(t *testing.T) {
	router := gin.New()
	router.Use(SlowRequestLogMiddleware(1))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSlowRequestLogMiddleware_ErrorStatus(t *testing.T) {
	router := gin.New()
	router.Use(SlowRequestLogMiddleware(1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// ============================================
// LoggingMiddleware Full Tests
// ============================================

func TestLoggingMiddleware_SuccessStatus(t *testing.T) {
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLoggingMiddleware_ClientErrorStatus(t *testing.T) {
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(400, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestLoggingMiddleware_ServerErrorStatus(t *testing.T) {
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestLoggingMiddleware_WithAllContextValues(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "trace-123")
		c.Set("request_id", "req-456")
		c.Set("tenant_id", "tenant-789")
		c.Set("user_id", "user-001")
		c.Next()
	})
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLoggingMiddleware_WithGinErrors(t *testing.T) {
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("handler error"))
		c.JSON(500, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestLoggingMiddleware_PostWithContentLength(t *testing.T) {
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Length", "100")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// RequestIDMiddleware Full Tests
// ============================================

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestIDMiddleware_UsesProvidedID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.Equal(t, "provided-id-123", requestID)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "provided-id-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "provided-id-123", w.Header().Get("X-Request-ID"))
}

// ============================================
// TraceIDMiddleware Full Tests
// ============================================

func TestTraceIDMiddleware_GeneratesID(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, exists := c.Get("trace_id")
		assert.True(t, exists)
		assert.NotEmpty(t, traceID)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
}

func TestTraceIDMiddleware_UsesProvidedID(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, exists := c.Get("trace_id")
		assert.True(t, exists)
		assert.Equal(t, "provided-trace-456", traceID)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", "provided-trace-456")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "provided-trace-456", w.Header().Get("X-Trace-ID"))
}

func TestTraceIDMiddleware_UsesB3TraceID(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		assert.Equal(t, "b3-trace-id-789", traceID)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-B3-Traceid", "b3-trace-id-789")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestTraceIDMiddleware_UsesTraceparent(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		// traceparent format: version-traceid-spanid-flags
		// 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
		assert.NotEmpty(t, traceID)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestTraceIDMiddleware_InvalidTraceparent(t *testing.T) {
	router := gin.New()
	router.Use(TraceIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		// Should generate new trace ID since traceparent is invalid
		assert.NotEmpty(t, traceID)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("traceparent", "invalid-format")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// splitString Tests
// ============================================

func TestSplitString_MultipleParts(t *testing.T) {
	parts := splitString("a-b-c-d", "-")
	assert.Equal(t, []string{"a", "b", "c", "d"}, parts)
}

func TestSplitString_SinglePart(t *testing.T) {
	parts := splitString("single", "-")
	assert.Equal(t, []string{"single"}, parts)
}

func TestSplitString_EmptyString(t *testing.T) {
	parts := splitString("", "-")
	assert.Equal(t, []string{""}, parts)
}

func TestSplitString_NoSeparator(t *testing.T) {
	parts := splitString("abc", "")
	assert.NotNil(t, parts)
}

func TestSplitString_TrailingSeparator(t *testing.T) {
	parts := splitString("a-b-", "-")
	assert.Equal(t, []string{"a", "b", ""}, parts)
}

func TestSplitString_MultipleCharacterSeparator(t *testing.T) {
	parts := splitString("a==b==c", "==")
	// Note: splitString splits by single character
	assert.NotNil(t, parts)
}

// ============================================
// randomString Tests
// ============================================

func TestRandomString_DifferentLengths(t *testing.T) {
	s1 := randomString(5)
	s2 := randomString(10)
	s3 := randomString(20)

	assert.Len(t, s1, 5)
	assert.Len(t, s2, 10)
	assert.Len(t, s3, 20)
}

func TestRandomString_ZeroLength(t *testing.T) {
	s := randomString(0)
	assert.Len(t, s, 0)
}

// ============================================
// randomInt Tests
// ============================================

func TestRandomInt_MultipleCalls(t *testing.T) {
	// Multiple calls should produce values in range
	for i := 0; i < 100; i++ {
		n := randomInt(100)
		assert.True(t, n >= 0 && n < 100)
	}
}

func TestRandomInt_OneMax(t *testing.T) {
	// Max of 1 should always return 0
	n := randomInt(1)
	assert.Equal(t, 0, n)
}
