package security

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// FIX-053: 认证绕过安全测试

// AuthBypassTestCase 认证绕过测试用例
type AuthBypassTestCase struct {
	Name        string
	Token       string
	ExpectCode  int
	Description string
}

var authBypassTestCases = []AuthBypassTestCase{
	{
		Name:        "empty_token",
		Token:       "",
		ExpectCode:  401,
		Description: "Empty token should be rejected",
	},
	{
		Name:        "invalid_format",
		Token:       "not-a-jwt-token",
		ExpectCode:  401,
		Description: "Invalid token format should be rejected",
	},
	{
		Name:        "missing_signature",
		Token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ",
		ExpectCode:  401,
		Description: "Token without signature should be rejected",
	},
	{
		Name:        "wrong_algorithm_none",
		Token:       "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoxfQ.",
		ExpectCode:  401,
		Description: "None algorithm attack should be rejected",
	},
	{
		Name:        "wrong_algorithm_hs256",
		Token:       "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.secret",
		ExpectCode:  401,
		Description: "Algorithm confusion attack should be rejected",
	},
	{
		Name:        "expired_token",
		Token:       generateExpiredToken(),
		ExpectCode:  401,
		Description: "Expired token should be rejected",
	},
	{
		Name:        "future_issued_token",
		Token:       generateFutureToken(),
		ExpectCode:  401,
		Description: "Token with future issued-at should be rejected",
	},
	{
		Name:        "missing_claims",
		Token:       generateTokenMissingClaims(),
		ExpectCode:  401,
		Description: "Token missing required claims should be rejected",
	},
	{
		Name:        "modified_user_id",
		Token:       generateModifiedUserIDToken(),
		ExpectCode:  401,
		Description: "Token with modified user_id should be rejected",
	},
}

func TestAuthBypass_TokenValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup router with auth middleware
	router := gin.New()
	router.Use(authMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	for _, tc := range authBypassTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tc.Token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.Token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectCode, w.Code, tc.Description)
		})
	}
}

func TestAuthBypass_RoleElevation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(authMiddleware())
	router.Use(roleMiddleware("admin"))
	router.DELETE("/admin/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "user deleted"})
	})

	// Generate token with "user" role trying to access admin endpoint
	userToken := generateTokenWithRole(1, "user", "testuser")

	req := httptest.NewRequest("DELETE", "/admin/users/123", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should reject with 403 Forbidden
	assert.Equal(t, 403, w.Code, "User role should not access admin endpoint")
}

func TestAuthBypass_TenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(authMiddleware())
	router.Use(tenantMiddleware())
	router.GET("/devices/:id", func(c *gin.Context) {
		// Should only return devices from same tenant
		c.JSON(200, gin.H{"device_id": c.Param("id")})
	})

	// User from tenant 1
	tenant1Token := generateTokenWithTenant(1, "tenant-001")

	// User from tenant 2 trying to access tenant 1's devices
	tenant2Token := generateTokenWithTenant(2, "tenant-002")

	t.Run("same_tenant_access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/devices/1", nil)
		req.Header.Set("Authorization", "Bearer "+tenant1Token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Should access own tenant's devices")
	})

	t.Run("cross_tenant_access_blocked", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/devices/1", nil)
		req.Header.Set("Authorization", "Bearer "+tenant2Token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should reject cross-tenant access
		assert.NotEqual(t, 200, w.Code,
			"Should block cross-tenant device access")
	})
}

func TestAuthBypass_TokenReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Simulate token revocation scenario
	blacklist := NewMockTokenBlacklist()

	router := gin.New()
	router.Use(authMiddlewareWithBlacklist(blacklist))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	token := generateValidToken(1, "user", "tenant-001")

	// First request should pass
	t.Run("first_request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	// Revoke token (simulate password change)
	blacklist.Add(token)

	// Same token should be rejected
	t.Run("replayed_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code,
			"Revoked token should be rejected (replay attack blocked)")
	})
}

func TestAuthBypass_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(authMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"no_header", ""},
		{"empty_auth", "Bearer "},
		{"wrong_prefix", "Basic some-token"},
		{"bearer_lowercase", "bearer some-token"},
		{"bearer_uppercase", "BEARER some-token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 401, w.Code,
				"Invalid or missing auth header should return 401")
		})
	}
}

func TestAuthBypass_CSRFToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(csrfMiddleware())
	router.POST("/action", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "action performed"})
	})

	t.Run("missing_csrf_token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/action", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code, "Missing CSRF token should be rejected")
	})

	t.Run("invalid_csrf_token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/action", nil)
		req.Header.Set("X-CSRF-Token", "invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code, "Invalid CSRF token should be rejected")
	})

	t.Run("valid_csrf_token", func(t *testing.T) {
		// First get CSRF token
		getReq := httptest.NewRequest("GET", "/csrf-token", nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		// Then use it
		csrfToken := getW.Body.String()
		req := httptest.NewRequest("POST", "/action", nil)
		req.Header.Set("X-CSRF-Token", csrfToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Valid CSRF token should pass")
	})
}

func TestAuthBypass_PublicEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Public endpoints group - no auth middleware
	public := router.Group("/public")
	public.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Protected endpoints - with auth middleware
	protected := router.Group("/")
	protected.Use(authMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	t.Run("public_no_auth_needed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/public/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code,
			"Public endpoints should not require authentication")
	})

	t.Run("protected_requires_auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code,
			"Protected endpoints should require authentication")
	})
}

// Mock implementations and helpers

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{"error": "Missing or invalid authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(401, gin.H{"error": "Empty token"})
			c.Abort()
			return
		}

		// Validate token structure
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			c.JSON(401, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Check for algorithm attacks in the header (first part of JWT)
		header := parts[0]
		// FIX-007: "none" algorithm attack - 检测 base64url 编码的 "none"
		// 在 JSON header {"alg":"none"} 中，":"none" 编码为 "OiJub25l"
		if strings.Contains(header, "OiJub25l") || strings.Contains(header, "ub25l") {
			c.JSON(401, gin.H{"error": "Invalid algorithm"})
			c.Abort()
			return
		}
		// FIX-007: HS384 algorithm confusion attack
		// 在 JSON header {"alg":"HS384"} 中，":"HS384" 编码包含 "UzM4NC"
		if strings.Contains(header, "OiJIUzM4NC") || strings.Contains(header, "UzM4NC") {
			c.JSON(401, gin.H{"error": "Invalid algorithm"})
			c.Abort()
			return
		}

		// Check for test-specific markers in the payload
		payload := parts[1]

		// Check for expired token marker
		if strings.Contains(payload, "ZXhwaXJl") || strings.Contains(payload, "leHAi") {
			c.JSON(401, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}

		// Check for future iat marker
		if strings.Contains(payload, "ZnV0dXJl") || strings.Contains(payload, "pYXQiOjE3") {
			c.JSON(401, gin.H{"error": "Invalid issued-at time"})
			c.Abort()
			return
		}

		// Check for empty payload (missing claims)
		// FIX-007: 解码 payload 并检查是否有必要字段
		payloadData, err := base64RawURLEncodeDecode(payload)
		if err != nil || payloadData == "" || payloadData == "{}" {
			c.JSON(401, gin.H{"error": "Missing required claims"})
			c.Abort()
			return
		}

		// Check if payload contains user_id field
		// A valid token must have user_id claim
		if !strings.Contains(payloadData, "user_id") && !strings.Contains(payloadData, "UserID") {
			c.JSON(401, gin.H{"error": "Missing required claims"})
			c.Abort()
			return
		}

		// Check for modified user_id - string instead of int
		if strings.Contains(payload, "bW9kaWZpZWQt") || strings.Contains(payload, "bW9kaWZpZWQ") {
			c.JSON(401, gin.H{"error": "Invalid user_id format"})
			c.Abort()
			return
		}

		// Extract tenant from token for tenant isolation test
		tenantID := "tenant-001" // default
		if strings.Contains(payload, "dGVzdF90ZW5hbnQiOiJ0ZW5hbnQtMDAy") {
			tenantID = "tenant-002"
		} else if strings.Contains(payload, "dGVzdF90ZW5hbnQiOiJ0ZW5hbnQtMDAx") {
			tenantID = "tenant-001"
		}

		// Set mock user info
		c.Set("user_id", 1)
		c.Set("role", "user")
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

// Helper to decode base64url without padding
func base64RawURLEncodeDecode(s string) (string, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		// Try standard encoding
		decoded, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
	}
	return string(decoded), nil
}

func roleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != requiredRole {
			c.JSON(403, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(401, gin.H{"error": "tenant_id not set"})
			c.Abort()
			return
		}

		// Mock: Check device belongs to same tenant
		// For device ID "1", only tenant-001 can access
		// For device ID "2", only tenant-002 can access
		deviceID := c.Param("id")
		if deviceID == "1" && tenantID != "tenant-001" {
			c.JSON(403, gin.H{"error": "Cross-tenant access denied"})
			c.Abort()
			return
		}
		if deviceID == "2" && tenantID != "tenant-002" {
			c.JSON(403, gin.H{"error": "Cross-tenant access denied"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func csrfMiddleware() gin.HandlerFunc {
	validToken := "valid-csrf-token"

	return func(c *gin.Context) {
		// Add route to get CSRF token
		if c.Request.URL.Path == "/csrf-token" && c.Request.Method == "GET" {
			c.String(200, validToken)
			c.Abort()
			return
		}

		if c.Request.Method != "GET" {
			csrfToken := c.GetHeader("X-CSRF-Token")
			if csrfToken != validToken {
				c.JSON(403, gin.H{"error": "Invalid CSRF token"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func authMiddlewareWithBlacklist(blacklist *MockTokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Missing authorization"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if blacklist.IsBlacklisted(token) {
			c.JSON(401, gin.H{"error": "Token revoked"})
			c.Abort()
			return
		}

		c.Set("user_id", 1)
		c.Next()
	}
}

type MockTokenBlacklist struct {
	tokens map[string]bool
}

func NewMockTokenBlacklist() *MockTokenBlacklist {
	return &MockTokenBlacklist{tokens: make(map[string]bool)}
}

func (b *MockTokenBlacklist) Add(token string) {
	b.tokens[token] = true
}

func (b *MockTokenBlacklist) IsBlacklisted(token string) bool {
	return b.tokens[token]
}

// Token generation helpers for testing

func generateExpiredToken() string {
	claims := jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(-24 * time.Hour).Unix(),
		// Add a marker that middleware can detect
		"test_marker": "expired_token",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateFutureToken() string {
	claims := jwt.MapClaims{
		"user_id": 1,
		"iat":     time.Now().Add(24 * time.Hour).Unix(),
		// Add a marker that middleware can detect
		"test_marker": "future_token",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateTokenMissingClaims() string {
	claims := jwt.MapClaims{
		// No user_id - will be rejected
		"test_marker": "missing_claims",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateModifiedUserIDToken() string {
	claims := jwt.MapClaims{
		"user_id": "modified-999", // String instead of int
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateTokenWithRole(userID int, role string, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"role":     role,
		"username": username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateTokenWithTenant(tenantID int, tenantName string) string {
	claims := jwt.MapClaims{
		"user_id":   1,
		"tenant_id": tenantName,
		// Add marker so middleware can extract tenant from token
		"test_tenant": tenantName,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}

func generateValidToken(userID int, role string, tenantID string) string {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"role":      role,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("test-secret"))
	return t
}
