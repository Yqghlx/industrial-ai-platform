package security

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// FIX-052: SQL 注入安全测试

// SQLInjectionPayload 常见的 SQL 注入 payload
var sqlInjectionPayloads = []string{
	"' OR '1'='1",
	"' OR '1'='1' --",
	"' OR '1'='1' /*",
	"1 OR 1=1",
	"1; DROP TABLE users--",
	"' UNION SELECT * FROM users--",
	"admin'--",
	"' AND 1=1--",
	"' AND 1=2--",
	"1' ORDER BY 1--",
	"1' ORDER BY 2--",
	"' HAVING 1=1--",
	"' HAVING 1=2--",
	"' GROUP BY username--",
	"1' UNION SELECT username, password FROM users WHERE '1'='1",
	"1' UNION SELECT null, null, null--",
	"'; EXEC xp_cmdshell('dir')--",
	"1' WAITFOR DELAY '0:0:5'--",
	"1' BENCHMARK(10000000, SHA1('test'))--",
	"' AND SLEEP(5)--",
	"1' AND (SELECT * FROM (SELECT(SLEEP(5)))a)--",
	"' IF (1=1) SELECT 'yes' ELSE SELECT 'no'--",
	"'; INSERT INTO users VALUES('hacker', 'password')--",
	"' LOAD_FILE('/etc/passwd')--",
	"1' INTO OUTFILE '/tmp/dump.txt'--",
}

func TestSQLInjection_DeviceEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock router with validation
	router := gin.New()
	router.GET("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")

		// Validate ID is numeric
		if !isNumeric(id) {
			c.JSON(400, gin.H{"error": "Invalid device ID"})
			return
		}

		c.JSON(200, gin.H{"device_id": id})
	})

	// Test each SQL injection payload
	for _, payload := range sqlInjectionPayloads {
		t.Run("payload_"+payload, func(t *testing.T) {
			// FIX-052: URL 编码 payload 以避免 httptest 解析错误
			encodedPayload := url.PathEscape(payload)
			req := httptest.NewRequest("GET", "/devices/"+encodedPayload, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should reject with 400 or 404, not 200 or 500
			// FIX-052: 404 也是安全的拒绝响应（路径包含 / 的 payload）
			assert.True(t, w.Code == 400 || w.Code == 404,
				"SQL injection payload should be rejected: %s (got %d)", payload, w.Code)
			// Response should indicate rejection
			assert.True(t, strings.Contains(w.Body.String(), "Invalid") || strings.Contains(w.Body.String(), "404"),
				"Response should indicate invalid input or not found")
		})
	}
}

func TestSQLInjection_SearchEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/devices/search", func(c *gin.Context) {
		query := c.Query("q")

		// Validate query length and content
		if len(query) > 100 {
			c.JSON(400, gin.H{"error": "Query too long"})
			return
		}

		// Check for suspicious patterns
		if containsSQLInjectionPattern(query) {
			c.JSON(400, gin.H{"error": "Invalid query pattern"})
			return
		}

		c.JSON(200, gin.H{"results": []string{}})
	})

	for _, payload := range sqlInjectionPayloads {
		if len(payload) <= 100 {
			t.Run("search_payload", func(t *testing.T) {
				// FIX-052: URL 编码 payload 以避免 httptest 解析错误
				encodedPayload := url.QueryEscape(payload)
				req := httptest.NewRequest("GET", "/devices/search?q="+encodedPayload, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, 400, w.Code,
					"SQL injection in search should be rejected: %s", payload)
			})
		}
	}
}

func TestSQLInjection_LoginEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/auth/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Validate username
		if len(username) > 50 || containsSQLInjectionPattern(username) {
			c.JSON(400, gin.H{"error": "Invalid username"})
			return
		}

		// Validate password
		if len(password) > 100 {
			c.JSON(400, gin.H{"error": "Invalid password"})
			return
		}

		c.JSON(200, gin.H{"message": "login processed"})
	})

	for _, payload := range sqlInjectionPayloads {
		t.Run("login_payload", func(t *testing.T) {
			// FIX-052: URL 编码 payload 以避免表单解析错误
			encodedPayload := url.QueryEscape(payload)
			req := httptest.NewRequest("POST", "/auth/login",
				strings.NewReader("username="+encodedPayload+"&password=test123"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 400, w.Code,
				"SQL injection in login should be rejected: %s", payload)
		})
	}
}

func TestSQLInjection_TenantEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/tenants", func(c *gin.Context) {
		name := c.PostForm("name")

		if len(name) > 100 || containsSQLInjectionPattern(name) {
			c.JSON(400, gin.H{"error": "Invalid tenant name"})
			return
		}

		c.JSON(200, gin.H{"message": "tenant created"})
	})

	for _, payload := range sqlInjectionPayloads {
		if len(payload) <= 100 {
			t.Run("tenant_payload", func(t *testing.T) {
				// FIX-052: URL 编码 payload 以避免表单解析错误
				encodedPayload := url.QueryEscape(payload)
				req := httptest.NewRequest("POST", "/tenants",
					strings.NewReader("name="+encodedPayload))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, 400, w.Code,
					"SQL injection in tenant creation should be rejected: %s", payload)
			})
		}
	}
}

func TestSQLInjection_WorkOrderEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/workorders", func(c *gin.Context) {
		title := c.PostForm("title")
		description := c.PostForm("description")

		// Validate inputs
		if len(title) > 200 || containsSQLInjectionPattern(title) {
			c.JSON(400, gin.H{"error": "Invalid title"})
			return
		}

		if len(description) > 2000 {
			c.JSON(400, gin.H{"error": "Description too long"})
			return
		}

		c.JSON(200, gin.H{"message": "workorder created"})
	})

	// Test title injection
	for _, payload := range sqlInjectionPayloads {
		if len(payload) <= 200 {
			t.Run("workorder_title", func(t *testing.T) {
				// FIX-052: URL 编码 payload 以避免表单解析错误
				encodedPayload := url.QueryEscape(payload)
				req := httptest.NewRequest("POST", "/workorders",
					strings.NewReader("title="+encodedPayload+"&description=normal"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, 400, w.Code,
					"SQL injection in workorder title should be rejected")
			})
		}
	}
}

func TestSQLInjection_AlertRuleEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/alerts/rules", func(c *gin.Context) {
		name := c.PostForm("name")
		condition := c.PostForm("condition")

		// Validate name
		if len(name) > 100 || containsSQLInjectionPattern(name) {
			c.JSON(400, gin.H{"error": "Invalid rule name"})
			return
		}

		// Condition should be validated more strictly
		if len(condition) > 500 {
			c.JSON(400, gin.H{"error": "Condition too complex"})
			return
		}

		c.JSON(200, gin.H{"message": "rule created"})
	})

	for _, payload := range sqlInjectionPayloads {
		if len(payload) <= 100 {
			t.Run("alert_rule_name", func(t *testing.T) {
				// FIX-052: URL 编码 payload 以避免表单解析错误
				encodedPayload := url.QueryEscape(payload)
				req := httptest.NewRequest("POST", "/alerts/rules",
					strings.NewReader("name="+encodedPayload+"&condition=value>100"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, 400, w.Code)
			})
		}
	}
}

// Helper functions for validation

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func containsSQLInjectionPattern(s string) bool {
	patterns := []string{
		"' OR ",
		"' AND ",
		" OR ",  // FIX-052: 添加不带引号的 OR 检测
		" AND ", // FIX-052: 添加不带引号的 AND 检测
		"UNION",
		"SELECT",
		"DROP",
		"INSERT",
		"UPDATE",
		"DELETE",
		"--",
		";",
		"/*",
		"*/",
		"EXEC",
		"xp_cmdshell",
		"SLEEP",
		"BENCHMARK",
		"WAITFOR",
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToUpper(s), strings.ToUpper(pattern)) {
			return true
		}
	}
	return false
}

func TestContainsSQLInjectionPattern_AllPatterns(t *testing.T) {
	for _, payload := range sqlInjectionPayloads {
		detected := containsSQLInjectionPattern(payload)
		assert.True(t, detected, "Should detect injection pattern: %s", payload)
	}
}

func TestContainsSQLInjectionPattern_SafeInputs(t *testing.T) {
	safeInputs := []string{
		"normal device name",
		"sensor-123",
		"temperature reading",
		"user@example.com",
		"alert for high temp",
		"工单描述",
	}

	for _, input := range safeInputs {
		detected := containsSQLInjectionPattern(input)
		assert.False(t, detected, "Should not flag safe input: %s", input)
	}
}
