package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
)

// AuthRequired validates JWT token and sets user info in context (增强版)
// 支持 Access Token + Refresh Token 机制
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
func GetUserID(c *gin.Context) int {
	if id, exists := c.Get("user_id"); exists {
		return id.(int)
	}
	return 0
}

// GetUsername extracts username from context
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("user_role"); exists {
		return role.(string)
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
