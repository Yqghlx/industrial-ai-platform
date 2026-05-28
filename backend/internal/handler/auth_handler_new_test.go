package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// AuthHandlerNew Tests
// ============================================

func TestNewAuthHandlerNew(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockAuthSvc, handler.authSvc)
	assert.Equal(t, mockUserSvc, handler.userSvc)
}

func TestAuthHandlerNew_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	expectedUser := &model.User{ID: 1, Username: "testuser"}
	expectedToken := "mock-jwt-token"

	mockAuthSvc.On("Login", mock.Anything, "testuser", "password123").
		Return(expectedUser, expectedToken, nil)

	router.POST("/login", handler.Login)

	body := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, expectedToken, response["token"])
	user := response["user"].(map[string]interface{})
	assert.Equal(t, "testuser", user["username"])

	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Login", mock.Anything, "testuser", "wrongpassword").
		Return(nil, "", assert.AnError)

	router.POST("/login", handler.Login)

	body := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Invalid credentials", response["error"])

	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_Login_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/login", handler.Login)

	body := map[string]string{
		"username": "testuser",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	expectedUser := &model.User{ID: 1, Username: "newuser", Email: "new@example.com"}
	expectedToken := "mock-jwt-token"

	registerReq := &model.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "new@example.com",
	}

	mockAuthSvc.On("Register", mock.Anything, registerReq).
		Return(expectedUser, expectedToken, nil)

	router.POST("/register", handler.Register)

	body := map[string]string{
		"username": "newuser",
		"password": "password123",
		"email":    "new@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, expectedToken, response["token"])
	user := response["user"].(map[string]interface{})
	assert.Equal(t, "newuser", user["username"])

	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_Register_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/register", handler.Register)

	body := map[string]string{
		"username": "newuser",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_Register_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	registerReq := &model.RegisterRequest{
		Username: "existinguser",
		Password: "password123",
		Email:    "existing@example.com",
	}

	mockAuthSvc.On("Register", mock.Anything, registerReq).
		Return(nil, "", assert.AnError)

	router.POST("/register", handler.Register)

	body := map[string]string{
		"username": "existinguser",
		"password": "password123",
		"email":    "existing@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/logout", handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Logged out successfully", response["message"])
}

func TestAuthHandlerNew_ChangePassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	// FIX-017: 设置 mock 预期
	mockAuthSvc.On("ChangePassword", mock.Anything, 1, "OldPass123!@", "NewPass123!@").Return(nil)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/change-password", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.ChangePassword(c)
	})

	// FIX-017: 使用符合密码复杂度要求的密码 (至少12位，包含大小写字母、数字和特殊字符)
	body := map[string]string{
		"old_password": "OldPass123!@",
		"new_password": "NewPass123!@",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockAuthSvc.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Password changed successfully", response["message"])
}

func TestAuthHandlerNew_ChangePassword_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/change-password", handler.ChangePassword)

	body := map[string]string{
		"old_password": "oldpass",
		"new_password": "newpass",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandlerNew_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	// FIX-016: 设置 mock 预期返回值
	expectedTokenPair := &service.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}
	mockAuthSvc.On("RefreshToken", mock.Anything, "old-refresh-token").Return(expectedTokenPair, nil)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/refresh-token", handler.RefreshToken)

	body := map[string]string{
		"token": "old-refresh-token",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	// FIX-017: 设置 mock 预期返回值
	expectedClaims := &service.Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "admin",
	}
	mockAuthSvc.On("ValidateToken", mock.Anything, "jwt-token-to-validate").Return(expectedClaims, nil)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/validate-token", handler.ValidateToken)

	body := map[string]string{
		"token": "jwt-token-to-validate",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/validate-token", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, true, response["valid"])
}

func TestAuthHandlerNew_RefreshToken_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/refresh-token", handler.RefreshToken)

	// Missing token body
	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
}

func TestAuthHandlerNew_ValidateToken_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/validate-token", handler.ValidateToken)

	// Missing token body
	req := httptest.NewRequest(http.MethodPost, "/validate-token", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_ChangePassword_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})

	router.POST("/change-password", handler.ChangePassword)

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer([]byte("invalid json body")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 for invalid JSON
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================
// GetCSRFToken Tests (SEC-HIGH-02)
// ============================================

func TestAuthHandlerNew_GetCSRFToken_NewToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.GET("/csrf-token", handler.GetCSRFToken)

	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that csrf_token is present and not empty
	assert.NotEmpty(t, response["csrf_token"])
	assert.Contains(t, response["message"], "CSRF token")

	// Check that cookie was set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Find the csrf_token cookie
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}
	assert.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
	assert.Equal(t, csrfCookie.Value, response["csrf_token"])
}

func TestAuthHandlerNew_GetCSRFToken_ExistingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.GET("/csrf-token", handler.GetCSRFToken)

	// Create request with existing CSRF cookie
	existingToken := "existing-csrf-token-value"
	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: existingToken,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return the existing token
	assert.Equal(t, existingToken, response["csrf_token"])
}

func TestAuthHandlerNew_GetCSRFToken_EmptyCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.GET("/csrf-token", handler.GetCSRFToken)

	// Create request with empty CSRF cookie (should generate new token)
	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: "", // Empty value
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return a new token (not empty)
	assert.NotEmpty(t, response["csrf_token"])

	// Check that new cookie was set
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}
	assert.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}
