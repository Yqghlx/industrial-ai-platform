package middleware

import (
	"bytes"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ANSI color constants for testing
const (
	Green   = "\x1b[32m"
	Magenta = "\x1b[35m"
	Yellow  = "\x1b[33m"
	Red     = "\x1b[31m"
)

// ============================================
// Logger middleware coverage tests
// ============================================

func TestLogger_WithQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?foo=bar&baz=123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLogger_WithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "req-12345")
		c.Next()
	})
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLogger_WithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.Error(gin.Error{Err: assert.AnError, Type: gin.ErrorTypePrivate})
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestLogger_StatusCodes(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedColor string
	}{
		{"status 200 OK", 200, Green},
		{"status 201 Created", 201, Green},
		{"status 301 Redirect", 301, Magenta},
		{"status 302 Found", 302, Magenta},
		{"status 400 Bad Request", 400, Yellow},
		{"status 404 Not Found", 404, Yellow},
		{"status 500 Internal Error", 500, Red},
		{"status 503 Service Unavailable", 503, Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(Logger())
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

func TestLogger_HTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(Logger())
			router.Handle(method, "/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"method": method})
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		})
	}
}

func TestLogger_PostRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"created": true})
	})

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
}

func TestLogger_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// Recovery middleware coverage tests
// ============================================

func TestRecovery_PanicWithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("intentional panic for testing")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestRecovery_PanicWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic(assert.AnError)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRecovery_PostRequestPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.POST("/test", func(c *gin.Context) {
		panic("post panic")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRecovery_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("panic with query")
	})

	req := httptest.NewRequest("GET", "/test?param=value", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// ============================================
// Combined middleware tests
// ============================================

func TestLoggerAndRecovery_Chain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("combined panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestLoggerAndRecovery_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// Edge case tests
// ============================================

func TestLogger_EmptyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"root": true})
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLogger_LongPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/api/v1/users/123/items/456/details", func(c *gin.Context) {
		c.JSON(200, gin.H{"path": "long"})
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123/items/456/details", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestLogger_SpecialCharacters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?name=测试&emoji=🎉", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Capture output test
func TestLogger_OutputCapture(t *testing.T) {
	// Skip: Logger now uses structured logging (zap), not stdout
	t.Skip("Logger uses structured logging, stdout capture no longer applicable")

	gin.SetMode(gin.TestMode)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-req-id")
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?param=value", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	assert.Contains(t, output, "200")
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "/test")
}

func TestRecovery_OutputCapture(t *testing.T) {
	// Skip: Recovery now uses structured logging (zap), not stdout
	t.Skip("Recovery uses structured logging, stdout capture no longer applicable")

	gin.SetMode(gin.TestMode)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("test panic output")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	assert.Contains(t, output, "PANIC RECOVERED")
	assert.Contains(t, output, "test panic output")
}

// Multiple errors test
func TestLogger_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.Error(gin.Error{Err: assert.AnError, Type: gin.ErrorTypePrivate})
		c.Error(gin.Error{Err: assert.AnError, Type: gin.ErrorTypePublic})
		c.JSON(500, gin.H{"error": "multiple errors"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// Status 3xx redirect test
func TestLogger_RedirectStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/old", func(c *gin.Context) {
		c.Redirect(301, "/new")
	})
	router.GET("/new", func(c *gin.Context) {
		c.JSON(200, gin.H{"location": "new"})
	})

	req := httptest.NewRequest("GET", "/old", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 301 redirect
	assert.Equal(t, 301, w.Code)
}
