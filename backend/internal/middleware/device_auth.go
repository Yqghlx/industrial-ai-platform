package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// SEC-002: Device Authentication Middleware
// SEC-HIGH-02: Device API Key mechanism for authenticating device telemetry endpoints
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
	// RequireAuth determines if authentication is required (default: true for SEC-HIGH-02)
	RequireAuth bool
}

// DefaultDeviceAuthConfig returns the default device auth configuration
func DefaultDeviceAuthConfig() *DeviceAuthConfig {
	return &DeviceAuthConfig{
		HeaderName:     "X-Device-Key",
		QueryParamName: "device_key",
		RequireAuth:    true,
	}
}

// SEC-HIGH-02: Global device API key validator
// Uses environment variable DEVICE_API_KEY for validation
var (
	deviceAPIKey     string
	deviceAPIKeyOnce sync.Once
)

// getDeviceAPIKey returns the device API key from environment
func getDeviceAPIKey() string {
	deviceAPIKeyOnce.Do(func() {
		deviceAPIKey = os.Getenv("DEVICE_API_KEY")
		if deviceAPIKey == "" {
			logger.L().Warn("DEVICE_API_KEY not set, device authentication will be less secure")
		}
	})
	return deviceAPIKey
}

// DeviceAuthRequired creates a middleware that requires device authentication
// SEC-HIGH-02: Used for device telemetry endpoints where devices need to authenticate
// This is the primary authentication mechanism for the telemetry endpoint
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

		// SEC-HIGH-02: 如果没有配置验证器，使用环境变量验证
		if config.ValidateKey == nil {
			config.ValidateKey = createDefaultDeviceKeyValidator()
		}

		// SEC-HIGH-02: 如果认证是必需的，但没有提供密钥
		if config.RequireAuth && deviceKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Device authentication required",
				"code":    "DEVICE_AUTH_REQUIRED",
				"message": "Provide device key via " + config.HeaderName + " header or " + config.QueryParamName + " parameter",
			})
			c.Abort()
			return
		}

		// 如果提供了密钥，验证它
		if deviceKey != "" {
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
			logger.L().Info("Device authenticated",
				zap.String("device_id", deviceID),
				zap.String("path", c.Request.URL.Path))
		} else {
			// 如果认证不是必需的，标记为未认证
			c.Set("device_authenticated", false)
		}

		c.Next()
	}
}

// createDefaultDeviceKeyValidator creates a default validator using environment variable
// SEC-HIGH-02: 使用环境变量 DEVICE_API_KEY 进行验证
func createDefaultDeviceKeyValidator() func(deviceKey string) (deviceID string, valid bool) {
	apiKey := getDeviceAPIKey()

	return func(deviceKey string) (string, bool) {
		// 如果没有设置 DEVICE_API_KEY，使用宽松验证（仅检查格式）
		if apiKey == "" {
			// 开发环境：接受任何非空密钥
			if deviceKey != "" && len(deviceKey) >= 8 {
				logger.L().Warn("Device authenticated without DEVICE_API_KEY validation (development mode)")
				return deviceKey, true
			}
			return "", false
		}

		// 生产环境：严格验证
		// 密钥可以是直接的 API Key 或设备特定密钥 (格式: deviceID:key)
		if deviceKey == apiKey {
			return "device", true
		}

		// 验证设备特定密钥格式
		parts := strings.SplitN(deviceKey, ":", 2)
		if len(parts) == 2 {
			deviceID := parts[0]
			key := parts[1]
			// 验证密钥是否匹配 API Key
			if key == apiKey {
				return deviceID, true
			}
			// 或验证 SHA256 哈希格式
			expectedHash := GenerateDeviceKey(deviceID, apiKey)
			if strings.EqualFold(deviceKey, expectedHash) {
				return deviceID, true
			}
		}

		return "", false
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
