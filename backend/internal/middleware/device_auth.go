package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SEC-002: Device Authentication Middleware
// Device API Key mechanism for authenticating device telemetry endpoints
// Devices authenticate using X-Device-Key header or Device-Key query parameter

// DeviceAuthConfig holds configuration for device authentication
type DeviceAuthConfig struct {
	// HeaderName is the HTTP header to check for device key
	HeaderName string
	// QueryParamName is the query parameter to check for device key
	QueryParamName string
	// ValidateKey is a function to validate the device key
	// Returns device ID if valid, empty string if invalid
	ValidateKey func(deviceKey string) (deviceID string, valid bool)
}

// DefaultDeviceAuthConfig returns the default device auth configuration
func DefaultDeviceAuthConfig() *DeviceAuthConfig {
	return &DeviceAuthConfig{
		HeaderName:     "X-Device-Key",
		QueryParamName: "device_key",
	}
}

// DeviceAuthRequired creates a middleware that requires device authentication
// This is used for device telemetry endpoints where devices need to authenticate
// but don't have user JWT tokens
func DeviceAuthRequired(config *DeviceAuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultDeviceAuthConfig()
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Device-Key"
	}
	if config.QueryParamName == "" {
		config.QueryParamName = "device_key"
	}

	return func(c *gin.Context) {
		// Try to get device key from header first
		deviceKey := c.GetHeader(config.HeaderName)
		if deviceKey == "" {
			// Try query parameter as fallback
			deviceKey = c.Query(config.QueryParamName)
		}

		if deviceKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Device authentication required",
				"code":    "DEVICE_AUTH_REQUIRED",
				"message": "Provide device key via " + config.HeaderName + " header or " + config.QueryParamName + " parameter",
			})
			c.Abort()
			return
		}

		// Validate the device key
		if config.ValidateKey != nil {
			deviceID, valid := config.ValidateKey(deviceKey)
			if !valid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid device key",
					"code":  "INVALID_DEVICE_KEY",
				})
				c.Abort()
				return
			}
			// Set device info in context
			c.Set("device_id", deviceID)
			c.Set("device_authenticated", true)
		} else {
			// Without a validator, we just check that a key was provided
			// This is useful for development/testing
			c.Set("device_key", deviceKey)
			c.Set("device_authenticated", true)
		}

		c.Next()
	}
}

// DeviceAuthOptional creates a middleware that optionally authenticates devices
// If device key is provided, it's validated; if not, the request proceeds
// This allows for gradual migration to authenticated device endpoints
func DeviceAuthOptional(config *DeviceAuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultDeviceAuthConfig()
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Device-Key"
	}
	if config.QueryParamName == "" {
		config.QueryParamName = "device_key"
	}

	return func(c *gin.Context) {
		// Try to get device key from header first
		deviceKey := c.GetHeader(config.HeaderName)
		if deviceKey == "" {
			// Try query parameter as fallback
			deviceKey = c.Query(config.QueryParamName)
		}

		if deviceKey != "" && config.ValidateKey != nil {
			deviceID, valid := config.ValidateKey(deviceKey)
			if valid {
				c.Set("device_id", deviceID)
				c.Set("device_authenticated", true)
			}
		}

		c.Set("device_authenticated", false)
		c.Next()
	}
}

// GenerateDeviceKey generates a device API key from device ID and secret
// The key format: sha256(deviceID + ":" + secret)
// This provides a deterministic but secure key for each device
func GenerateDeviceKey(deviceID, secret string) string {
	h := sha256.New()
	h.Write([]byte(deviceID + ":" + secret))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateDeviceKey validates a device key against the expected key
// Returns the device ID if valid, empty string if invalid
func ValidateDeviceKey(deviceKey, deviceID, secret string) bool {
	expectedKey := GenerateDeviceKey(deviceID, secret)
	return strings.EqualFold(deviceKey, expectedKey)
}

// GetDeviceID extracts device ID from context (set by DeviceAuth middleware)
func GetDeviceID(c *gin.Context) string {
	if deviceID, exists := c.Get("device_id"); exists {
		if id, ok := deviceID.(string); ok {
			return id
		}
	}
	return ""
}

// IsDeviceAuthenticated checks if the request was authenticated as a device
func IsDeviceAuthenticated(c *gin.Context) bool {
	if auth, exists := c.Get("device_authenticated"); exists {
		if authenticated, ok := auth.(bool); ok {
			return authenticated
		}
	}
	return false
}

// DeviceKeyFromRequest extracts device key from request without validation
// Useful for logging or debugging
func DeviceKeyFromRequest(c *gin.Context, config *DeviceAuthConfig) string {
	if config == nil {
		config = DefaultDeviceAuthConfig()
	}

	deviceKey := c.GetHeader(config.HeaderName)
	if deviceKey == "" {
		deviceKey = c.Query(config.QueryParamName)
	}
	return deviceKey
}
