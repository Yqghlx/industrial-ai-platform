package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// AdminHandlerNew Tests
// ============================================

func TestNewAdminHandlerNew(t *testing.T) {
	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockAuthSvc, handler.authSvc)
}

func TestAdminHandlerNew_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.GET("/users", handler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(0), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(50), response["page_size"])
}

func TestAdminHandlerNew_CreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.POST("/users", handler.CreateUser)

	body := map[string]string{
		"username": "newuser",
		"password": "password123",
		"email":    "new@example.com",
		"role":     "admin",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "User created (placeholder)", response["message"])
}

func TestAdminHandlerNew_CreateUser_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.POST("/users", handler.CreateUser)

	body := map[string]string{
		"username": "newuser",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerNew_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.DELETE("/users/:id", handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "User deleted (placeholder)", response["message"])
	assert.Equal(t, "123", response["id"])
}

func TestAdminHandlerNew_GetSystemStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.GET("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "running", response["uptime"])
}

func TestAdminHandlerNew_CreateUser_WithOptionalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc)

	router.POST("/users", handler.CreateUser)

	// Test without email and role (optional fields)
	body := map[string]string{
		"username": "minimaluser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
