package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
)

// FIX-016: AuthConfig for public endpoint whitelist configuration
// AuthConfig holds configuration for authentication middleware
type AuthConfig struct {
	// JWTSecret is the secret key for JWT validation
	JWTSecret string
	// PublicEndpoints is a list of endpoints that don't require authentication
	// These endpoints are intentionally public (e.g., telemetry, health checks)
	PublicEndpoints []string
	// PublicEndpointPolicy documents the reason for public access
	PublicEndpointPolicy string
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		PublicEndpoints: []string{
			"/health", // Health check endpoint - intentionally public
		},
		PublicEndpointPolicy: "Health check endpoint is intentionally public for operational reasons",
	}
}

// isPublicEndpoint checks if the request path matches a public endpoint
// Supports exact match and prefix match (e.g., "/api/v1/telemetry/*")
func isPublicEndpoint(path string, publicEndpoints []string) bool {
	for _, endpoint := range publicEndpoints {
		// Exact match
		if path == endpoint {
			return true
		}
		// Prefix match for wildcard endpoints (e.g., "/api/v1/telemetry")
		if strings.HasSuffix(endpoint, "*") {
			prefix := strings.TrimSuffix(endpoint, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

// AuthRequiredWithConfig validates JWT token with configurable public endpoints
// This is the enhanced version that supports public endpoint whitelist
func AuthRequiredWithConfig(config *AuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAuthConfig()
	}
	if config.JWTSecret == "" {
		// Fallback to global JWT secret
		config.JWTSecret = string(GetJWTSecret())
	}

	return func(c *gin.Context) {
		// FIX-016: Check if this is a public endpoint
		if isPublicEndpoint(c.Request.URL.Path, config.PublicEndpoints) {
			// Mark request as public for logging/auditing
			c.Set("public_endpoint", true)
			c.Set("public_endpoint_reason", config.PublicEndpointPolicy)
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		tokenString := ExtractToken(authHeader)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// 使用增强版 ParseToken (支持黑名单检查)
		claims, err := service.ParseToken(tokenString)
		if err != nil {
			// 区分不同的错误类型
			errorMsg := "Invalid or expired token"
			errorCode := "INVALID_TOKEN"

			if strings.Contains(err.Error(), "revoked") {
				errorMsg = "Token has been revoked"
				errorCode = "TOKEN_REVOKED"
			} else if strings.Contains(err.Error(), "expired") {
				errorMsg = "Token has expired"
				errorCode = "TOKEN_EXPIRED"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errorMsg,
				"code":  errorCode,
			})
			c.Abort()
			return
		}

		// 验证是否是 Access Token (不能使用 Refresh Token 访问 API)
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Refresh token cannot be used for API access",
				"code":  "INVALID_TOKEN_TYPE",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("role", claims.Role) // 兼容旧代码
		c.Set("tenant_id", claims.TenantID)
		c.Set("token_id", claims.TokenID)

		c.Next()
	}
}

// AuthRequired validates JWT token and sets user info in context (增强版)
// 支持 Access Token + Refresh Token 机制
// This is the backward-compatible version without public endpoint whitelist
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	// Use AuthRequiredWithConfig with empty public endpoints for backward compatibility
	config := &AuthConfig{
		JWTSecret:       jwtSecret,
		PublicEndpoints: []string{}, // No public endpoints for backward compatibility
	}
	return AuthRequiredWithConfig(config)
}

// AuthOptional creates a middleware that optionally authenticates requests.
// If a token is provided, it's validated; if not, the request proceeds as public.
// Use this for endpoints that can serve both authenticated and public data.
func AuthOptional(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token - mark as public request
			c.Set("optional_auth", true)
			c.Set("authenticated", false)
			c.Next()
			return
		}

		tokenString := ExtractToken(authHeader)
		if tokenString == "" {
			// Invalid format - still proceed as public
			c.Set("optional_auth", true)
			c.Set("authenticated", false)
			c.Next()
			return
		}

		// Try to validate the token
		claims, err := service.ParseToken(tokenString)
		if err != nil {
			// Invalid token - proceed as public
			c.Set("optional_auth", true)
			c.Set("authenticated", false)
			c.Next()
			return
		}

		// Valid token - set user info
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("role", claims.Role)
		c.Set("tenant_id", claims.TenantID)
		c.Set("token_id", claims.TokenID)
		c.Set("authenticated", true)
		c.Next()
	}
}

// IsPublicEndpointRequest checks if the current request is to a public endpoint
func IsPublicEndpointRequest(c *gin.Context) bool {
	if public, exists := c.Get("public_endpoint"); exists {
		if isPublic, ok := public.(bool); ok {
			return isPublic
		}
	}
	return false
}

// GetPublicEndpointReason returns the reason/policy for public endpoint access
func GetPublicEndpointReason(c *gin.Context) string {
	if reason, exists := c.Get("public_endpoint_reason"); exists {
		if r, ok := reason.(string); ok {
			return r
		}
	}
	return ""
}

// IsAuthenticated checks if the request is authenticated (either via JWT or optional auth)
func IsAuthenticated(c *gin.Context) bool {
	if auth, exists := c.Get("authenticated"); exists {
		if isAuthenticated, ok := auth.(bool); ok {
			return isAuthenticated
		}
	}
	// Check if user_id is set (from AuthRequired)
	if _, exists := c.Get("user_id"); exists {
		return true
	}
	return false
}

// ExtractToken 从 Authorization Header 提取 Token
func ExtractToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// AdminRequired requires user to have admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantRequired requires valid tenant ID in context
func TenantRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists || tenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant ID required",
				"code":  "MISSING_TENANT",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts user ID from context
// P1-08: 安全类型断言 - 避免panic
func GetUserID(c *gin.Context) int {
	if id, exists := c.Get("user_id"); exists {
		if userID, ok := id.(int); ok {
			return userID
		}
	}
	return 0
}

// GetUsername extracts username from context
// P1-08: 安全类型断言 - 避免panic
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if usernameStr, ok := username.(string); ok {
			return usernameStr
		}
	}
	return ""
}

// GetUserRole extracts user role from context
// P1-08: 安全类型断言 - 避免panic
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr
		}
	}
	return ""
}

// GetTenantID extracts tenant ID from context
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		return tenantID.(string)
	}
	return ""
}

// GetTokenID extracts token ID from context
func GetTokenID(c *gin.Context) string {
	if tokenID, exists := c.Get("token_id"); exists {
		return tokenID.(string)
	}
	return ""
}
