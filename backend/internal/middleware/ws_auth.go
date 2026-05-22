package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SEC-MED-01: WebSocket Authentication Middleware
// WebSocket endpoints can be protected using JWT token authentication
// via query parameter or header before the WebSocket upgrade.

// WSAuthConfig holds configuration for WebSocket authentication
type WSAuthConfig struct {
	// JWTSecret is the secret key for JWT validation
	JWTSecret string
	// TokenQueryParam is the query parameter name for JWT token
	TokenQueryParam string
	// TokenHeader is the header name for JWT token (before WebSocket upgrade)
	TokenHeader string
	// AllowPublic controls whether public (unauthenticated) connections are allowed
	// Set to true for intentionally public WebSocket endpoints
	AllowPublic bool
	// PublicPolicy documents the policy for public WebSocket access
	PublicPolicy string
}

// DefaultWSAuthConfig returns default WebSocket auth configuration
func DefaultWSAuthConfig() *WSAuthConfig {
	return &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
		AllowPublic:     false, // Default: require authentication
		PublicPolicy:    "",
	}
}

// WSAuthRequired creates a middleware that requires JWT authentication for WebSocket
// connections. The JWT token must be provided via query parameter or Authorization header.
// Note: WebSocket connections cannot use standard Authorization headers after upgrade,
// so the token must be validated BEFORE the WebSocket upgrade occurs.
func WSAuthRequired(config *WSAuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultWSAuthConfig()
	}
	if config.TokenQueryParam == "" {
		config.TokenQueryParam = "token"
	}
	if config.TokenHeader == "" {
		config.TokenHeader = "Authorization"
	}

	return func(c *gin.Context) {
		// Skip auth for non-WebSocket upgrade requests
		if !isWebSocketUpgrade(c) {
			c.Next()
			return
		}

		// Extract JWT token from query parameter or header
		token := extractWSToken(c, config)
		if token == "" {
			// No token provided - check if public access is allowed
			if config.AllowPublic {
				// Public access is documented and intentional
				c.Set("ws_public", true)
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "WebSocket authentication required",
				"code":    "WS_AUTH_REQUIRED",
				"message": "Provide JWT token via 'token' query parameter or Authorization header",
			})
			c.Abort()
			return
		}

		// Validate JWT token
		if config.JWTSecret != "" {
			userID, err := validateJWTToken(token, config.JWTSecret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid WebSocket token",
					"code":    "WS_INVALID_TOKEN",
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			// Set user info in context
			c.Set("user_id", userID)
			c.Set("ws_authenticated", true)
		}

		c.Next()
	}
}

// WSAuthOptional creates a middleware that optionally authenticates WebSocket connections.
// If a token is provided, it's validated; if not, the connection proceeds as public.
// Use this when WebSocket can serve both authenticated and public data.
func WSAuthOptional(config *WSAuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultWSAuthConfig()
	}
	config.AllowPublic = true // Allow public access

	return WSAuthRequired(config)
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade
func isWebSocketUpgrade(c *gin.Context) bool {
	return strings.ToLower(c.GetHeader("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade")
}

// extractWSToken extracts JWT token from query parameter or header
func extractWSToken(c *gin.Context, config *WSAuthConfig) string {
	// First try query parameter (common for WebSocket connections)
	token := c.Query(config.TokenQueryParam)
	if token != "" {
		return token
	}

	// Try Authorization header (Bearer token)
	authHeader := c.GetHeader(config.TokenHeader)
	if authHeader != "" {
		// Extract Bearer token
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	return ""
}

// GetWSUserID extracts the authenticated user ID from WebSocket context
func GetWSUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// IsWSAuthenticated checks if WebSocket connection is authenticated
func IsWSAuthenticated(c *gin.Context) bool {
	if auth, exists := c.Get("ws_authenticated"); exists {
		if authenticated, ok := auth.(bool); ok {
			return authenticated
		}
	}
	return false
}

// IsWSPublic checks if WebSocket connection is public (unauthenticated)
func IsWSPublic(c *gin.Context) bool {
	if public, exists := c.Get("ws_public"); exists {
		if isPublic, ok := public.(bool); ok {
			return isPublic
		}
	}
	return false
}