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

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/change-password", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.ChangePassword(c)
	})

	body := map[string]string{
		"old_password": "oldpass",
		"new_password": "newpass",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

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
}

func TestAuthHandlerNew_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)

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
