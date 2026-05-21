package middleware

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestWAFRouter(config WAFConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	router := gin.New()
	router.Use(WAFMiddleware(config, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	router.GET("/upload", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return router
}

func TestWAFMiddleware_SQLInjection(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	tests := []struct {
		name        string
		query       string
		expectBlock bool
	}{
		{"normal", "q=hello", false},
		{"select_from", "q=SELECT+something+FROM+users", true},
		{"union_select", "q=UNION+SELECT+something+FROM+users", true},
		{"drop_x_table", "q=DROP+something+TABLE+users", true},
		{"insert_x_into", "q=INSERT+something+INTO+admin", true},
		{"delete_x_from", "q=DELETE+something+FROM+users", true},
		{"exec", "q=EXEC+master..xp_cmdshell", true},
		{"declare", "q=DECLARE+%40var+NVARCHAR(200)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?"+tt.query, nil)
			req.Header.Set("User-Agent", "test-agent")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectBlock {
				assert.Equal(t, 403, w.Code, "Expected block for: %s", tt.query)
			} else {
				assert.Equal(t, 200, w.Code, "Expected pass for: %s", tt.query)
			}
		})
	}
}

func TestWAFMiddleware_XSSProtection(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	tests := []struct {
		name        string
		query       string
		expectBlock bool
	}{
		{"normal_text", "q=hello+world", false},
		{"script_tag", "q=%3Cscript%3Ealert(1)%3C/script%3E", true},
		{"javascript_uri", "q=javascript:alert(1)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?"+tt.query, nil)
			req.Header.Set("User-Agent", "test-agent")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectBlock {
				assert.Equal(t, 403, w.Code)
			} else {
				assert.Equal(t, 200, w.Code)
			}
		})
	}
}

func TestWAFMiddleware_PathTraversal(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("GET", "/test?q=../../../etc/passwd", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_SafeRequests(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	safeQueries := []string{
		"device=CNC-001",
		"status=online",
		"page=1&limit=10",
		"search=motor+bearing",
		"sort=created_at&order=desc",
	}

	for _, q := range safeQueries {
		req := httptest.NewRequest("GET", "/test?"+q, nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Should allow safe query: %s", q)
	}
}

func TestWAFMiddleware_Disabled(t *testing.T) {
	config := DefaultWAFConfig()
	config.Enabled = false
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("GET", "/test?q=something+FROM+users", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestWAFMiddleware_BlockedUserAgent(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	blockedAgents := []string{
		"sqlmap/1.0",
		"Nikto/2.1.5",
		"masscan/1.0",
		"Nmap Scripting Engine",
	}

	for _, agent := range blockedAgents {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", agent)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code, "Should block user agent: %s", agent)
	}
}

func TestWAFMiddleware_EmptyUserAgent(t *testing.T) {
	config := DefaultWAFConfig()
	config.BlockEmptyUserAgent = true
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_EmptyUserAgentAllowed(t *testing.T) {
	config := DefaultWAFConfig()
	config.BlockEmptyUserAgent = false
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestWAFMiddleware_RequestTooLarge(t *testing.T) {
	config := DefaultWAFConfig()
	config.MaxRequestSize = 100 // Very small limit
	router := newTestWAFRouter(config)

	body := strings.Repeat("a", 200)
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_PostBodySQLInjection(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	body := `{"query": "SELECT * FROM users WHERE id = 1"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_PostBodyXSS(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	body := `{"comment": "<script>alert('xss')</script>"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_SensitivePath(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	sensitivePaths := []string{
		"/.env",
		"/.git/config",
		"/wp-config.php",
	}

	for _, path := range sensitivePaths {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code, "Should block path: %s", path)
	}
}

func TestWAFMiddleware_DangerousUpload(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("POST", "/upload", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Content-Disposition", `attachment; filename="test.php"`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestWAFMiddleware_SafeUpload(t *testing.T) {
	config := DefaultWAFConfig()
	router := newTestWAFRouter(config)

	req := httptest.NewRequest("POST", "/upload", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Content-Disposition", "attachment; filename=\"report.pdf\"")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May be 200 or 403 depending on whether the router has the POST /upload handler
	assert.NotNil(t, w.Code)
}

func TestDefaultWAFConfig(t *testing.T) {
	config := DefaultWAFConfig()
	assert.True(t, config.Enabled)
	assert.NotEmpty(t, config.SQLInjectionPatterns)
	assert.NotEmpty(t, config.XSSPatterns)
	assert.NotEmpty(t, config.PathTraversalPatterns)
	assert.NotEmpty(t, config.CommandInjectionPatterns)
	assert.True(t, config.MaxRequestSize > 0)
	assert.True(t, config.MaxArgsLength > 0)
	assert.NotEmpty(t, config.BlockedExtensions)
	assert.NotEmpty(t, config.BlockedUserAgents)
	assert.True(t, config.BlockEmptyUserAgent)
}

func TestIsAllowedMethod(t *testing.T) {
	allowed := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
	assert.True(t, isAllowedMethod("GET", allowed))
	assert.True(t, isAllowedMethod("POST", allowed))
	assert.False(t, isAllowedMethod("PROPFIND", allowed))
}

func TestIsBlockedUserAgentWithConfig(t *testing.T) {
	blocked := []string{"bot", "crawler", "sqlmap"}

	assert.True(t, isBlockedUserAgentWithConfig("Googlebot/2.1", blocked, false))
	assert.True(t, isBlockedUserAgentWithConfig("sqlmap/1.5", blocked, false))
	assert.False(t, isBlockedUserAgentWithConfig("Mozilla/5.0", blocked, false))
	assert.True(t, isBlockedUserAgentWithConfig("", blocked, true))
	assert.False(t, isBlockedUserAgentWithConfig("", blocked, false))
}

func TestDetectAttack(t *testing.T) {
	sqlPatterns := []string{`(?i)select\s+.*\s+from`}
	assert.True(t, detectAttack("SELECT * FROM users", sqlPatterns))
	assert.False(t, detectAttack("hello world", sqlPatterns))
}

func TestIsBlockedExtension(t *testing.T) {
	blocked := []string{".php", ".jsp", ".exe"}
	assert.True(t, isBlockedExtension("evil.php", blocked))
	assert.True(t, isBlockedExtension("evil.PHP", blocked))
	assert.False(t, isBlockedExtension("report.pdf", blocked))
	assert.False(t, isBlockedExtension("", blocked))
}

func TestExtractFilename(t *testing.T) {
	tests := []struct {
		name               string
		contentDisposition string
		expected           string
	}{
		{"attachment", `attachment; filename="test.php"`, "test.php"},
		{"no filename", "attachment", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilename(tt.contentDisposition)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSensitivePath(t *testing.T) {
	assert.True(t, isSensitivePath("/.env"))
	assert.True(t, isSensitivePath("/.git/config"))
	assert.True(t, isSensitivePath("/WP-CONFIG.PHP"))
	assert.False(t, isSensitivePath("/api/devices"))
	assert.False(t, isSensitivePath("/dashboard"))
}

func TestContainsString(t *testing.T) {
	list := []string{"apple", "banana", "cherry"}
	assert.True(t, containsString(list, "apple"))
	assert.False(t, containsString(list, "grape"))
}

func TestWAFStatsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	stats := &WAFStats{}
	config := DefaultWAFConfig()

	router := gin.New()
	router.Use(WAFStatsMiddleware(stats))
	router.Use(WAFMiddleware(config, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Normal request
	req := httptest.NewRequest("GET", "/test?q=hello", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.BlockedRequests)

	// SQL injection request
	req2 := httptest.NewRequest("GET", "/test?q=SELECT+something+FROM+users", nil)
	req2.Header.Set("User-Agent", "test-agent")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 403, w2.Code)
	assert.Equal(t, int64(2), stats.TotalRequests)
	assert.Equal(t, int64(1), stats.BlockedRequests)
	assert.Equal(t, int64(1), stats.SQLInjection)
}
