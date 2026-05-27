package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================
// validateJWTToken Tests
// ============================================

func TestValidateJWTToken_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	// Create a valid JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.NoError(t, err)
	assert.Equal(t, "12345", userID)
}

func TestValidateJWTToken_ValidToken_FloatUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	// Create token with float64 user_id (JSON number)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(12345), // JSON decodes numbers as float64
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.NoError(t, err)
	assert.Equal(t, "12345", userID)
}

func TestValidateJWTToken_EmptySecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
	})
	tokenString, err := token.SignedString([]byte("some-secret"))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, "")
	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.Contains(t, err.Error(), "JWT secret not configured")
}

func TestValidateJWTToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	userID, err := validateJWTToken("invalid-token-string", secret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
}

func TestValidateJWTToken_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "correct-secret"
	wrongSecret := "wrong-secret"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, wrongSecret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
}

func TestValidateJWTToken_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	// Create an expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
}

func TestValidateJWTToken_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.Contains(t, err.Error(), "invalid token claims")
}

func TestValidateJWTToken_InvalidUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": []string{"invalid"}, // Invalid type
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.Contains(t, err.Error(), "invalid user_id type")
}

func TestValidateJWTToken_InvalidSigningMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	// Create token with none signing method (will be rejected)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": "12345",
	})
	// Need to use unsigned string for none method
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	userID, err := validateJWTToken(tokenString, secret)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
}

// ============================================
// DefaultWSAuthConfig Tests
// ============================================

func TestDefaultWSAuthConfig(t *testing.T) {
	config := DefaultWSAuthConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "token", config.TokenQueryParam)
	assert.Equal(t, "Authorization", config.TokenHeader)
	assert.False(t, config.AllowPublic)
	assert.Empty(t, config.PublicPolicy)
}

func TestDefaultWSAuthConfig_Immutable(t *testing.T) {
	config1 := DefaultWSAuthConfig()
	config2 := DefaultWSAuthConfig()

	// Modify config1
	config1.TokenQueryParam = "modified"
	config1.AllowPublic = true

	// config2 should not be affected
	assert.Equal(t, "token", config2.TokenQueryParam)
	assert.False(t, config2.AllowPublic)
}

// ============================================
// isWebSocketUpgrade Tests
// ============================================

func TestIsWebSocketUpgrade_ValidUpgrade(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Upgrade", "websocket")
	c.Request.Header.Set("Connection", "Upgrade")

	assert.True(t, isWebSocketUpgrade(c))
}

func TestIsWebSocketUpgrade_ValidUpgrade_Lowercase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Upgrade", "WebSocket") // Mixed case
	c.Request.Header.Set("Connection", "upgrade")

	assert.True(t, isWebSocketUpgrade(c))
}

func TestIsWebSocketUpgrade_NoUpgradeHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Connection", "Upgrade")

	assert.False(t, isWebSocketUpgrade(c))
}

func TestIsWebSocketUpgrade_NoConnectionHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Upgrade", "websocket")

	assert.False(t, isWebSocketUpgrade(c))
}

func TestIsWebSocketUpgrade_WrongUpgradeValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Upgrade", "h2c")
	c.Request.Header.Set("Connection", "Upgrade")

	assert.False(t, isWebSocketUpgrade(c))
}

func TestIsWebSocketUpgrade_RegularHTTPRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/test", nil)

	assert.False(t, isWebSocketUpgrade(c))
}

// ============================================
// extractWSToken Tests
// ============================================

func TestExtractWSToken_FromQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws?token=my-jwt-token", nil)

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractWSToken_FromQueryParam_CustomParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws?access_token=my-jwt-token", nil)

	config := &WSAuthConfig{
		TokenQueryParam: "access_token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractWSToken_FromAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Authorization", "Bearer my-jwt-token")

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractWSToken_FromAuthorizationHeader_NoBearerPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Authorization", "my-jwt-token")

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractWSToken_FromCustomHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("X-Auth-Token", "my-jwt-token")

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "X-Auth-Token",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractWSToken_QueryPriorityOverHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws?token=query-token", nil)
	c.Request.Header.Set("Authorization", "Bearer header-token")

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Equal(t, "query-token", token)
}

func TestExtractWSToken_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ws", nil)

	config := &WSAuthConfig{
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	token := extractWSToken(c, config)
	assert.Empty(t, token)
}

// ============================================
// WSAuthRequired Middleware Tests
// ============================================

func TestWSAuthRequired_SkipsNonWebSocketRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/api/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthRequired_WebSocketNoToken_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
		AllowPublic:     false,
	}

	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "WS_AUTH_REQUIRED")
}

func TestWSAuthRequired_WebSocketWithValidToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	// Create valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	config := &WSAuthConfig{
		JWTSecret:       secret,
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		assert.Equal(t, "12345", GetWSUserID(c))
		assert.True(t, IsWSAuthenticated(c))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws?token="+tokenString, nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthRequired_WebSocketWithInvalidToken_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws?token=invalid-token", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "WS_INVALID_TOKEN")
}

func TestWSAuthRequired_WebSocketWithBearerToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-789",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	config := &WSAuthConfig{
		JWTSecret:       secret,
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		assert.Equal(t, "user-789", GetWSUserID(c))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthRequired_PublicAllowed_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
		AllowPublic:     true,
		PublicPolicy:    "Public dashboard WebSocket",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		assert.True(t, IsWSPublic(c))
		assert.False(t, IsWSAuthenticated(c))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthRequired_NilConfig_UsesDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that nil config doesn't panic
	router := gin.New()
	router.Use(WSAuthRequired(nil))
	router.GET("/ws", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return unauthorized because no token and no JWT secret
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWSAuthRequired_EmptyJWTSecret_SkipsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "", // Empty secret
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws?token=some-token", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler is called because JWT secret is empty (no validation)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthRequired_ConfigDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that empty TokenQueryParam and TokenHeader use defaults
	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "",
		TokenHeader:     "",
	}

	secret := "test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "12345",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Use default query param "token"
	req := httptest.NewRequest("GET", "/ws?token="+tokenString, nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// WSAuthOptional Middleware Tests
// ============================================

func TestWSAuthOptional_NoToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &WSAuthConfig{
		JWTSecret:       "test-secret",
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthOptional(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		assert.True(t, IsWSPublic(c))
		assert.False(t, IsWSAuthenticated(c))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthOptional_WithValidToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-key"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "99999",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	config := &WSAuthConfig{
		JWTSecret:       secret,
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthOptional(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		assert.Equal(t, "99999", GetWSUserID(c))
		assert.True(t, IsWSAuthenticated(c))
		assert.False(t, IsWSPublic(c))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws?token="+tokenString, nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWSAuthOptional_NilConfig_UsesDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthOptional(nil))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// GetWSUserID Tests
// ============================================

func TestGetWSUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "12345")

	userID := GetWSUserID(c)
	assert.Equal(t, "12345", userID)
}

func TestGetWSUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := GetWSUserID(c)
	assert.Empty(t, userID)
}

func TestGetWSUserID_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 12345) // Integer instead of string

	userID := GetWSUserID(c)
	assert.Empty(t, userID)
}

// ============================================
// IsWSAuthenticated Tests
// ============================================

func TestIsWSAuthenticated_True(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_authenticated", true)

	assert.True(t, IsWSAuthenticated(c))
}

func TestIsWSAuthenticated_False(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_authenticated", false)

	assert.False(t, IsWSAuthenticated(c))
}

func TestIsWSAuthenticated_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	assert.False(t, IsWSAuthenticated(c))
}

func TestIsWSAuthenticated_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_authenticated", "yes") // String instead of bool

	assert.False(t, IsWSAuthenticated(c))
}

// ============================================
// IsWSPublic Tests
// ============================================

func TestIsWSPublic_True(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_public", true)

	assert.True(t, IsWSPublic(c))
}

func TestIsWSPublic_False(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_public", false)

	assert.False(t, IsWSPublic(c))
}

func TestIsWSPublic_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	assert.False(t, IsWSPublic(c))
}

func TestIsWSPublic_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("ws_public", "yes") // String instead of bool

	assert.False(t, IsWSPublic(c))
}

// ============================================
// Integration Tests
// ============================================

func TestWSAuthIntegration_FullFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "integration-test-secret"

	config := &WSAuthConfig{
		JWTSecret:       secret,
		TokenQueryParam: "auth_token",
		TokenHeader:     "X-WS-Auth",
		AllowPublic:     false,
	}

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "integration-user",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	handlerCalled := false
	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		handlerCalled = true
		userID := GetWSUserID(c)
		assert.Equal(t, "integration-user", userID)
		assert.True(t, IsWSAuthenticated(c))
		assert.False(t, IsWSPublic(c))
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Test with token in custom header
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("X-WS-Auth", tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "integration-user")
}

func TestWSAuthIntegration_ExpiredTokenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "expiry-test-secret"

	config := &WSAuthConfig{
		JWTSecret:       secret,
		TokenQueryParam: "token",
		TokenHeader:     "Authorization",
	}

	// Create an expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "expired-user",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired
	})
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	router := gin.New()
	router.Use(WSAuthRequired(config))
	router.GET("/ws", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach"})
	})

	req := httptest.NewRequest("GET", "/ws?token="+tokenString, nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "WS_INVALID_TOKEN")
}
