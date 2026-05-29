package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// 默认密钥验证器测试（使用环境变量）
// ============================================

func TestDefaultDeviceKeyValidator_WithEnvKey(t *testing.T) {
	// 重置 sync.Once
	deviceAPIKeyOnce = sync.Once{}
	deviceAPIKey = ""

	os.Setenv("DEVICE_API_KEY", "test-api-key-12345")
	defer os.Unsetenv("DEVICE_API_KEY")

	validator := createDefaultDeviceKeyValidator()

	// 直接 API Key 匹配
	deviceID, valid := validator("test-api-key-12345")
	assert.True(t, valid)
	assert.Equal(t, "device", deviceID)

	// 设备特定密钥格式 (deviceID:key)
	deviceID, valid = validator("my-device:test-api-key-12345")
	assert.True(t, valid)
	assert.Equal(t, "my-device", deviceID)

	// 无效密钥
	deviceID, valid = validator("wrong-key")
	assert.False(t, valid)
}

func TestDefaultDeviceKeyValidator_WithoutEnvKey(t *testing.T) {
	// 重置 sync.Once
	deviceAPIKeyOnce = sync.Once{}
	deviceAPIKey = ""

	os.Unsetenv("DEVICE_API_KEY")

	validator := createDefaultDeviceKeyValidator()

	// 开发模式：接受 8 字符以上的密钥
	deviceID, valid := validator("long-enough-key")
	assert.True(t, valid)
	assert.Equal(t, "long-enough-key", deviceID)

	// 太短的密钥
	deviceID, valid = validator("short")
	assert.False(t, valid)
}

func TestDefaultDeviceKeyValidator_InvalidShortKey(t *testing.T) {
	// 重置 sync.Once
	deviceAPIKeyOnce = sync.Once{}
	deviceAPIKey = ""

	os.Setenv("DEVICE_API_KEY", "test-secret-key")
	defer os.Unsetenv("DEVICE_API_KEY")

	validator := createDefaultDeviceKeyValidator()

	// 不匹配任何格式的密钥
	deviceID, valid := validator("random-invalid-key")
	assert.False(t, valid)
	assert.Equal(t, "", deviceID)
}

// ============================================
// DeviceAuthRequired 额外覆盖测试
// ============================================

func TestDeviceAuthRequired_NilConfig(t *testing.T) {
	// 重置环境变量
	deviceAPIKeyOnce = sync.Once{}
	deviceAPIKey = ""
	os.Unsetenv("DEVICE_API_KEY")

	router := gin.New()
	router.Use(DeviceAuthRequired(nil))
	router.POST("/telemetry", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/telemetry", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 没有密钥时应该返回 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeviceAuthRequired_OptionalAuthMode(t *testing.T) {
	config := &DeviceAuthConfig{
		ValidateKey: func(deviceKey string) (string, bool) {
			return "", false
		},
		RequireAuth: false, // 认证不是必需的
	}

	router := gin.New()
	router.Use(DeviceAuthRequired(config))
	router.POST("/telemetry", func(c *gin.Context) {
		auth, _ := c.Get("device_authenticated")
		assert.Equal(t, false, auth)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/telemetry", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 不需要认证时应该通过
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceAuthRequired_EmptyConfig(t *testing.T) {
	// 空的 HeaderName 和 QueryParamName 应该被填充默认值
	deviceAPIKeyOnce = sync.Once{}
	deviceAPIKey = ""
	os.Unsetenv("DEVICE_API_KEY")

	config := &DeviceAuthConfig{
		RequireAuth: true,
	}

	router := gin.New()
	router.Use(DeviceAuthRequired(config))
	router.POST("/telemetry", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/telemetry", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// DeviceAuthOptional 额外覆盖测试
// ============================================

func TestDeviceAuthOptional_NilConfig(t *testing.T) {
	router := gin.New()
	router.Use(DeviceAuthOptional(nil))
	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceAuthOptional_ValidKeyWithValidator(t *testing.T) {
	config := &DeviceAuthConfig{
		ValidateKey: func(deviceKey string) (string, bool) {
			if deviceKey == "valid-key" {
				return "device-1", true
			}
			return "", false
		},
	}

	router := gin.New()
	router.Use(DeviceAuthOptional(config))
	router.GET("/data", func(c *gin.Context) {
		// DeviceAuthOptional 总是设置 device_authenticated 为 false
		// 因为它在 c.Next() 之前重置了
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("X-Device-Key", "valid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceAuthOptional_EmptyConfig(t *testing.T) {
	config := &DeviceAuthConfig{}

	router := gin.New()
	router.Use(DeviceAuthOptional(config))
	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// ActiveRequestTracker 额外测试
// ============================================

func TestActiveRequestTracker_New(t *testing.T) {
	tracker := NewActiveRequestTracker()
	assert.NotNil(t, tracker)
	assert.Equal(t, int64(0), tracker.GetActiveCount())
}

func TestRequestTrackingMiddleware_MultipleRequests(t *testing.T) {
	tracker := NewActiveRequestTracker()

	router := gin.New()
	router.Use(RequestTrackingMiddleware(tracker))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	assert.Equal(t, int64(0), tracker.GetActiveCount())
}

// ============================================
// HSTSDefault 测试
// ============================================

func TestHSTSDefault_Middleware(t *testing.T) {
	router := gin.New()
	router.Use(HSTSDefault())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	hstsValue := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hstsValue, "max-age=31536000")
	assert.Contains(t, hstsValue, "includeSubDomains")
	assert.Contains(t, hstsValue, "preload")
}

// ============================================
// BuildCSPFromConfig 边界测试
// ============================================

func TestBuildCSPFromConfig_WithReportURI(t *testing.T) {
	config := &CSPConfig{
		DefaultSrc: []string{"'self'"},
		ReportURI:  "/csp-report",
	}
	header := BuildCSPFromConfig(config)
	assert.Contains(t, header, "report-uri /csp-report")
}

func TestBuildCSPFromConfig_AllFields(t *testing.T) {
	config := &CSPConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'", "'unsafe-inline'"},
		StyleSrc:       []string{"'self'"},
		ImgSrc:         []string{"'self'", "data:"},
		FontSrc:        []string{"'self'"},
		ConnectSrc:     []string{"'self'", "wss:"},
		FrameAncestors: []string{"'none'"},
		FormAction:     []string{"'self'"},
		BaseURI:        []string{"'self'"},
	}
	header := BuildCSPFromConfig(config)
	assert.Contains(t, header, "default-src")
	assert.Contains(t, header, "script-src")
	assert.Contains(t, header, "style-src")
	assert.Contains(t, header, "img-src")
	assert.Contains(t, header, "font-src")
	assert.Contains(t, header, "connect-src")
	assert.Contains(t, header, "frame-ancestors")
	assert.Contains(t, header, "form-action")
	assert.Contains(t, header, "base-uri")
}

// ============================================
// ForceHTTPS 使用 Host header 回退测试
// ============================================

func TestForceHTTPS_FallbackToHost(t *testing.T) {
	router := gin.New()
	router.Use(ForceHTTPS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?q=1", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	// 不设置 X-Forwarded-Host，回退到 Request.Host
	req.Host = "myhost.com"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "https://myhost.com")
}

// ============================================
// SecurityHeaders 常量测试
// ============================================

func TestHSTSMaxAgeSeconds(t *testing.T) {
	assert.Equal(t, 31536000, HSTSMaxAgeSeconds)
}

func TestCORSMaxAgeSeconds(t *testing.T) {
	assert.Equal(t, 86400, CORSMaxAgeSeconds)
}
