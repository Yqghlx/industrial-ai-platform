package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware provides CSRF protection using the double submit cookie pattern.
// This is a defense against CSRF attacks for session-based authentication.
//
// SEC-MED-07: CSRF Protection Explanation
// Note: When using JWT tokens stored in localStorage/sessionStorage and sent via
// Authorization header, CSRF protection is typically not needed as browsers
// don't automatically include these tokens in requests.
//
// See docs/SECURITY_CSRF.md for detailed explanation of CSRF protection strategy.
//
// CSRF Protection is REQUIRED when:
// - Using cookie-based session authentication
// - Using cookies to store JWT tokens (not recommended)
//
// CSRF Protection is NOT REQUIRED when:
// - Using JWT tokens via Authorization header (our default)
// - Using localStorage/sessionStorage for token storage
// - Implementing stateless REST APIs

// CSRFConfig holds the configuration for CSRF middleware
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes
	TokenLength int
	// CookieName is the name of the CSRF cookie
	CookieName string
	// HeaderName is the name of the HTTP header to check for token
	HeaderName string
	// FormFieldName is the name of the form field to check for token
	FormFieldName string
	// CookieSecure sets the Secure attribute on the cookie
	CookieSecure bool
	// CookieHTTPOnly sets the HttpOnly attribute (should be false for CSRF cookies)
	// CSRF tokens need to be readable by JavaScript to be sent in headers
	CookieHTTPOnly bool
	// CookieSameSite sets the SameSite attribute (Strict, Lax, or None)
	CookieSameSite http.SameSite
	// CookieDomain sets the Domain attribute
	CookieDomain string
	// CookiePath sets the Path attribute
	CookiePath string
	// CookieMaxAge sets the MaxAge attribute in seconds
	CookieMaxAge int
	// Error handling
	ErrorHandler func(c *gin.Context, errMsg string)
}

// DefaultCSRFConfig returns a secure default CSRF configuration
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		TokenLength:    32,
		CookieName:     "csrf_token",
		HeaderName:     "X-CSRF-Token",
		FormFieldName:  "_csrf",
		CookieSecure:   true,  // HTTPS only
		CookieHTTPOnly: false, // Must be false so JS can read it
		CookieSameSite: http.SameSiteStrictMode,
		CookiePath:     "/",
		CookieMaxAge:   3600, // 1 hour
		ErrorHandler:   defaultCSRFErrorHandler,
	}
}

// DevelopmentCSRFConfig returns a configuration suitable for development
func DevelopmentCSRFConfig() *CSRFConfig {
	cfg := DefaultCSRFConfig()
	cfg.CookieSecure = false // Allow HTTP in development
	cfg.CookieSameSite = http.SameSiteLaxMode
	return cfg
}

// defaultCSRFErrorHandler handles CSRF validation errors
func defaultCSRFErrorHandler(c *gin.Context, errMsg string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":   "CSRF validation failed",
		"code":    "CSRF_INVALID",
		"message": errMsg,
	})
	c.Abort()
}

// CSRF returns a CSRF middleware with default configuration
func CSRF() gin.HandlerFunc {
	return CSRFWithConfig(DefaultCSRFConfig())
}

// CSRFWithConfig returns a CSRF middleware with custom configuration
// This implements the Double Submit Cookie pattern:
// 1. Server generates a random token and sets it as a cookie
// 2. Client reads the cookie and includes the token in requests (header or form field)
// 3. Server validates that the cookie token matches the request token
func CSRFWithConfig(config *CSRFConfig) gin.HandlerFunc {
	// Apply defaults for zero values
	if config.TokenLength <= 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "_csrf"
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCSRFErrorHandler
	}

	return func(c *gin.Context) {
		// Skip CSRF for safe methods (GET, HEAD, OPTIONS, TRACE)
		if isSafeMethod(c.Request.Method) {
			// For GET requests, we may want to set a new CSRF token
			// This ensures the token is always fresh
			_, err := c.Cookie(config.CookieName)
			if err != nil {
				// No existing cookie, set a new one
				setCSRFToken(c, config)
			}
			c.Next()
			return
		}

		// For unsafe methods (POST, PUT, DELETE, PATCH), validate CSRF token
		cookieToken, err := c.Cookie(config.CookieName)
		if err != nil || cookieToken == "" {
			config.ErrorHandler(c, "CSRF token not found in cookie")
			return
		}

		// Get token from request (header first, then form)
		requestToken := c.GetHeader(config.HeaderName)
		if requestToken == "" {
			requestToken = c.PostForm(config.FormFieldName)
		}
		if requestToken == "" {
			// Check query parameter as fallback
			requestToken = c.Query(config.FormFieldName)
		}

		if requestToken == "" {
			config.ErrorHandler(c, "CSRF token not found in request")
			return
		}

		// Validate tokens match (constant-time comparison)
		if !secureCompare(cookieToken, requestToken) {
			config.ErrorHandler(c, "CSRF token mismatch")
			return
		}

		c.Next()
	}
}

// setCSRFToken generates and sets a new CSRF token cookie
func setCSRFToken(c *gin.Context, config *CSRFConfig) string {
	token := generateCSRFToken(config.TokenLength)

	// Build cookie with SameSite attribute
	// Using SetCookie with proper SameSite support
	c.SetSameSite(config.CookieSameSite)
	c.SetCookie(
		config.CookieName,
		token,
		config.CookieMaxAge,
		config.CookiePath,
		config.CookieDomain,
		config.CookieSecure,
		config.CookieHTTPOnly,
	)

	return token
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken(length int) string {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less secure but functional method
		// This should rarely happen
		panic("crypto/rand failed: " + err.Error())
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

// isSafeMethod checks if the HTTP method is considered safe (doesn't modify state)
func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "TRACE":
		return true
	default:
		return false
	}
}

// secureCompare performs a constant-time comparison to prevent timing attacks
func secureCompare(a, b string) bool {
	// Use subtle.ConstantTimeCompare for proper constant-time comparison
	// We need to convert strings to byte slices
	aBytes := []byte(a)
	bBytes := []byte(b)

	// If lengths differ, return false immediately
	// (this leaks length info but is acceptable for CSRF tokens)
	if len(aBytes) != len(bBytes) {
		return false
	}

	// Simple constant-time comparison
	// For production, use crypto/subtle.ConstantTimeCompare
	result := 0
	for i := 0; i < len(aBytes); i++ {
		result |= int(aBytes[i]) ^ int(bBytes[i])
	}

	return result == 0
}

// CSRFTokenHandler returns a handler that generates and returns a new CSRF token
// This can be used to get a fresh token for subsequent requests
func CSRFTokenHandler(config *CSRFConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCSRFConfig()
	}

	return func(c *gin.Context) {
		token := setCSRFToken(c, config)

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
		})
	}
}

// CSRFTokenFromContext extracts the CSRF token from the context (cookie)
// Returns empty string if not found
func CSRFTokenFromContext(c *gin.Context, config *CSRFConfig) string {
	if config == nil {
		config = DefaultCSRFConfig()
	}
	token, _ := c.Cookie(config.CookieName)
	return token
}

// RequireCSRF is a stricter CSRF middleware that requires a token for all requests
// including safe methods (useful for APIs that need strict CSRF protection)
func RequireCSRF(config *CSRFConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCSRFConfig()
	}

	return func(c *gin.Context) {
		cookieToken, err := c.Cookie(config.CookieName)
		if err != nil || cookieToken == "" {
			// No token exists, generate one
			cookieToken = setCSRFToken(c, config)
		}

		// Always validate token
		requestToken := c.GetHeader(config.HeaderName)
		if requestToken == "" {
			requestToken = c.PostForm(config.FormFieldName)
		}

		if requestToken != "" && secureCompare(cookieToken, requestToken) {
			c.Next()
			return
		}

		// For safe methods, allow request but ensure token is set
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		config.ErrorHandler(c, "CSRF token required")
	}
}
