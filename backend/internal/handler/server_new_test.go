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
