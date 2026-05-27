package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// Server Helper Functions Tests
// ============================================

func TestGetRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx, cancel := getRequestContext(c)
		defer cancel()

		assert.NotNil(t, ctx)
		_, ok := ctx.Deadline()
		assert.True(t, ok)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetRequestContext_HasTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx, cancel := getRequestContext(c)
		defer cancel()

		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.True(t, deadline.After(time.Now()))

		c.JSON(http.StatusOK, gin.H{"deadline": deadline.Format(time.RFC3339)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGenerateRandomPassword(t *testing.T) {
	password1 := generateRandomPassword(16)
	password2 := generateRandomPassword(16)

	assert.Len(t, password1, 16)
	assert.Len(t, password2, 16)
	assert.NotEqual(t, password1, password2) // Should be random
}

func TestGenerateRandomPassword_DifferentLengths(t *testing.T) {
	password8 := generateRandomPassword(8)
	password32 := generateRandomPassword(32)

	assert.Len(t, password8, 8)
	assert.Len(t, password32, 32)
}

func TestGenerateFallbackPassword(t *testing.T) {
	password := generateFallbackPassword(16)

	assert.Len(t, password, 16)
	// Should be numeric
	for _, c := range password {
		assert.True(t, c >= '0' && c <= '9')
	}
}

func TestGenerateFallbackPassword_DifferentLengths(t *testing.T) {
	password8 := generateFallbackPassword(8)
	password16 := generateFallbackPassword(16)

	assert.Len(t, password8, 8)
	assert.Len(t, password16, 16)
}

// ============================================
// Backward Compatibility Tests
// ============================================

func TestCompatAuthSvc_Login(t *testing.T) {
	mockUserSvc := new(MockUserService)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	expectedUser := &model.User{ID: 1, Username: "testuser"}
	mockUserSvc.On("Authenticate", "testuser", "password").Return(expectedUser, nil)

	user, token, err := compatSvc.Login(context.Background(), "testuser", "password")

	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	assert.Equal(t, "token", token)

	mockUserSvc.AssertExpectations(t)
}

func TestCompatAuthSvc_Login_Error(t *testing.T) {
	mockUserSvc := new(MockUserService)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	mockUserSvc.On("Authenticate", "testuser", "wrongpassword").Return(nil, assert.AnError)

	user, token, err := compatSvc.Login(context.Background(), "testuser", "wrongpassword")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)

	mockUserSvc.AssertExpectations(t)
}

func TestCompatAuthSvc_GetUserByID(t *testing.T) {
	mockUserSvc := new(MockUserService)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	expectedUser := &model.User{ID: 1, Username: "testuser"}
	mockUserSvc.On("GetByID", 1).Return(expectedUser, nil)

	user, err := compatSvc.GetUserByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockUserSvc.AssertExpectations(t)
}

func TestCompatAuthSvc_Register(t *testing.T) {
	mockUserSvc := new(MockUserService)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	user, token, err := compatSvc.Register(context.Background(), &model.RegisterRequest{
		Username: "newuser",
		Password: "password",
	})

	require.NoError(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
}

func TestNewAuthHandler_BackwardCompat(t *testing.T) {
	mockUserSvc := new(MockUserService)
	jwtSecret := "test-secret"

	handler := NewAuthHandler(mockUserSvc, jwtSecret)

	assert.NotNil(t, handler)
}

// ============================================
// Pagination Tests (from validation.go)
// ============================================

func TestGetPagination_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		pagination := GetPagination(c)
		c.JSON(http.StatusOK, gin.H{
			"page":      pagination.Page,
			"page_size": pagination.PageSize,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(20), response["page_size"])
}

func TestGetPagination_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		pagination := GetPagination(c)
		c.JSON(http.StatusOK, gin.H{
			"page":      pagination.Page,
			"page_size": pagination.PageSize,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?page=5&page_size=100", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(5), response["page"])
	assert.Equal(t, float64(100), response["page_size"])
}

func TestGetPagination_InvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		pagination := GetPagination(c)
		c.JSON(http.StatusOK, gin.H{
			"page":      pagination.Page,
			"page_size": pagination.PageSize,
		})
	})

	// Negative values should be corrected
	req := httptest.NewRequest(http.MethodGet, "/test?page=-1&page_size=-10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should use defaults for invalid values
	assert.GreaterOrEqual(t, response["page"], float64(1))
	assert.GreaterOrEqual(t, response["page_size"], float64(1))
}

func TestValidatePagination(t *testing.T) {
	pagination := ValidatePagination(0, 0)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 20, pagination.PageSize)

	pagination = ValidatePagination(5, 100)
	assert.Equal(t, 5, pagination.Page)
	assert.Equal(t, 100, pagination.PageSize)

	pagination = ValidatePagination(-1, -10)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 20, pagination.PageSize)
}

func TestValidatePagination_MaxPageSize(t *testing.T) {
	// Test max page size limit
	pagination := ValidatePagination(1, 10000)
	assert.Equal(t, 100, pagination.PageSize) // Should be capped at MaxPageSize (100)
}

// ============================================
// HTTPServerNew Simple Tests
// ============================================

func TestHTTPServerNew_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a minimal server for testing
	startTime := time.Now()
	server := &HTTPServerNew{
		router:    router,
		startTime: startTime,
	}

	router.GET("/health", server.healthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "uptime")
}

func TestHTTPServerNew_GetRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &HTTPServerNew{
		router: router,
	}

	result := server.GetRouter()

	assert.NotNil(t, result)
	assert.Equal(t, router, result)
}

func TestHTTPServerNew_getCacheStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockHealthHandler := NewHealthHandlerNew(time.Now())

	server := &HTTPServerNew{
		router:        router,
		healthHandler: mockHealthHandler,
	}

	router.GET("/cache-status", server.getCacheStatus)

	req := httptest.NewRequest(http.MethodGet, "/cache-status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// FIX-015: WebSocket Origin Validation Tests
// ============================================

func TestIsLocalhostOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		// HTTP localhost
		{"http localhost", "http://localhost", true},
		{"http localhost with port 3000", "http://localhost:3000", true},
		{"http localhost with port 8080", "http://localhost:8080", true},
		{"http localhost with port 5173", "http://localhost:5173", true},
		{"http localhost with path", "http://localhost/app", true},

		// HTTPS localhost
		{"https localhost", "https://localhost", true},
		{"https localhost with port", "https://localhost:443", true},

		// HTTP 127.0.0.1
		{"http 127.0.0.1", "http://127.0.0.1", true},
		{"http 127.0.0.1 with port", "http://127.0.0.1:3000", true},
		{"http 127.0.0.1 with port 8080", "http://127.0.0.1:8080", true},

		// HTTPS 127.0.0.1
		{"https 127.0.0.1", "https://127.0.0.1", true},
		{"https 127.0.0.1 with port", "https://127.0.0.1:443", true},

		// IPv6 localhost
		{"http [::1]", "http://[::1]", true},
		{"http [::1] with port", "http://[::1]:3000", true},
		{"https [::1]", "https://[::1]", true},

		// Non-localhost origins (should be false)
		{"external domain", "https://example.com", false},
		{"external domain with port", "https://example.com:443", false},
		{"http external", "http://external.host", false},
		{"empty origin", "", false},
		{"malformed origin", "not-a-url", false},
		{"partial localhost match", "http://localhost.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalhostOrigin(tt.origin)
			assert.Equal(t, tt.expected, result, "isLocalhostOrigin(%q) = %v, want %v", tt.origin, result, tt.expected)
		})
	}
}

func TestWebSocketOriginCheck_Development(t *testing.T) {
	// Test that development environment allows localhost origins
	gin.SetMode(gin.TestMode)

	// Simulate development environment CheckOrigin function
	isProduction := false
	wsAllowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:8080": true,
	}

	// Add default localhost origins for dev
	wsAllowedOrigins["http://localhost"] = true
	wsAllowedOrigins["http://127.0.0.1"] = true

	checkOrigin := func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			if !isProduction {
				return true
			}
			return false
		}
		if wsAllowedOrigins[origin] {
			return true
		}
		if wsAllowedOrigins["*"] {
			return true
		}
		if !isProduction {
			return isLocalhostOrigin(origin)
		}
		return false
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"no origin - dev allows", "", true},
		{"localhost in allowed", "http://localhost:3000", true},
		{"127.0.0.1 in allowed", "http://127.0.0.1:8080", true},
		{"localhost not in allowed but isLocalhost", "http://localhost:5173", true},
		{"127.0.0.1 not in allowed but isLocalhost", "http://127.0.0.1:3001", true},
		{"external domain should fail", "https://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			result := checkOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebSocketOriginCheck_Production(t *testing.T) {
	// Test that production environment strictly checks origins
	gin.SetMode(gin.TestMode)

	// Simulate production environment CheckOrigin function
	isProduction := true
	wsAllowedOrigins := map[string]bool{
		"https://app.example.com": true,
		"https://admin.example.com": true,
	}

	checkOrigin := func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			if !isProduction {
				return true
			}
			return false
		}
		if wsAllowedOrigins[origin] {
			return true
		}
		if wsAllowedOrigins["*"] {
			return true
		}
		if !isProduction {
			return isLocalhostOrigin(origin)
		}
		return false
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"no origin - prod denies", "", false},
		{"allowed origin 1", "https://app.example.com", true},
		{"allowed origin 2", "https://admin.example.com", true},
		{"localhost should fail in prod", "http://localhost:3000", false},
		{"external domain not allowed", "https://evil.com", false},
		{"http variant not allowed", "http://app.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			result := checkOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebSocketOriginCheck_Wildcard(t *testing.T) {
	// Test wildcard origin handling (not recommended for production)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		isProduction      bool
		wsAllowedOrigins  map[string]bool
		origin            string
		expected          bool
	}{
		{
			name:             "wildcard allows any origin",
			isProduction:     false,
			wsAllowedOrigins: map[string]bool{"*": true},
			origin:           "https://any-domain.com",
			expected:         true,
		},
		{
			name:             "wildcard in production allows external",
			isProduction:     true,
			wsAllowedOrigins: map[string]bool{"*": true},
			origin:           "https://external.com",
			expected:         true,
		},
		{
			name:             "no wildcard and not in allowed - dev localhost",
			isProduction:     false,
			wsAllowedOrigins: map[string]bool{},
			origin:           "http://localhost:9999",
			expected:         true, // localhost allowed in dev
		},
		{
			name:             "no wildcard and not in allowed - dev external",
			isProduction:     false,
			wsAllowedOrigins: map[string]bool{},
			origin:           "https://external.com",
			expected:         false, // external not allowed even in dev
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkOrigin := func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					if !tt.isProduction {
						return true
					}
					return false
				}
				if tt.wsAllowedOrigins[origin] {
					return true
				}
				if tt.wsAllowedOrigins["*"] {
					return true
				}
				if !tt.isProduction {
					return isLocalhostOrigin(origin)
				}
				return false
			}

			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			result := checkOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}
