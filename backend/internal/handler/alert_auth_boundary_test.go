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

	"github.com/industrial-ai/platform/internal/mocks"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/errors"
)

// ============================================
// Alert Handler Boundary Tests (补充)
// ============================================

func TestAlertHandlerNew_GetAlertByID_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)

	handler := NewAlertHandler(mockAlertSvc, func(msg model.WSMessage) {})

	mockAlertSvc.On("GetAlertByID", mock.Anything, 1).Return(nil, errors.NewNotFoundError("Alert", "1"))

	router.GET("/alerts/:id", handler.GetAlert)

	req := httptest.NewRequest(http.MethodGet, "/alerts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandlerNew_GetTrend_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockAlertSvc.On("GetTrendReport", mock.Anything, "7d").Return(&model.TrendReport{
		Period: "7d",
		Trend:  []model.TrendEntry{},
	}, nil)

	handler := NewAlertHandler(mockAlertSvc, func(msg model.WSMessage) {})

	router.GET("/alerts/trend", handler.GetTrend)

	req := httptest.NewRequest(http.MethodGet, "/alerts/trend?period=7d", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandlerNew_GetRanking_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockAlertSvc.On("GetDeviceRanking", mock.Anything, 10).Return([]model.DeviceRankingEntry{}, nil)

	handler := NewAlertHandler(mockAlertSvc, func(msg model.WSMessage) {})

	router.GET("/alerts/ranking", handler.GetRanking)

	req := httptest.NewRequest(http.MethodGet, "/alerts/ranking?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandlerNew_GetEfficiency_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockAlertSvc.On("GetEfficiencyReport", mock.Anything).Return(&model.EfficiencyReport{
		AvgResolveTime: 0.0,
		AckRate:        0.0,
		TotalAlerts:    0,
		ResolvedAlerts: 0,
	}, nil)

	handler := NewAlertHandler(mockAlertSvc, func(msg model.WSMessage) {})

	router.GET("/alerts/efficiency", handler.GetEfficiency)

	req := httptest.NewRequest(http.MethodGet, "/alerts/efficiency", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandlerNew_CreateRule_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockDeviceSvc := new(mocks.MockDeviceService)
	mockAuthSvc := new(mocks.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.POST("/rules", handler.CreateRule)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertHandlerNew_UpdateRule_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockDeviceSvc := new(mocks.MockDeviceService)
	mockAuthSvc := new(mocks.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.PUT("/rules/:id", handler.UpdateRule)

	req := httptest.NewRequest(http.MethodPut, "/rules/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertHandlerNew_UpdateRule_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	mockDeviceSvc := new(mocks.MockDeviceService)
	mockAuthSvc := new(mocks.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockAlertSvc.On("UpdateRule", mock.Anything, mock.AnythingOfType("*model.AlertRule")).Return(assert.AnError)

	router.PUT("/rules/:id", handler.UpdateRule)

	body := map[string]interface{}{
		"name":      "Updated",
		"metric":    "temperature",
		"operator":  ">",
		"threshold": 90.0,
		"severity":  "high",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/rules/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandlerNew_AcknowledgeAlert_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)

	handler := NewAlertHandler(mockAlertSvc, func(msg model.WSMessage) {})

	mockAlertSvc.On("AcknowledgeAlert", mock.Anything, 1).Return(assert.AnError)

	router.POST("/alerts/:id/acknowledge", handler.AcknowledgeAlert)

	req := httptest.NewRequest(http.MethodPost, "/alerts/1/acknowledge", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

// ============================================
// Auth Handler Boundary Tests (补充)
// ============================================

func TestAuthHandlerNew_Login_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/auth/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_Register_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/auth/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_ChangePassword_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.POST("/auth/change-password", handler.ChangePassword)

	req := httptest.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_RefreshToken_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/auth/refresh-token", handler.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_ValidateToken_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	router.POST("/auth/validate-token", handler.ValidateToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerNew_Login_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Login", mock.Anything, "testuser", "password").Return(nil, "", assert.AnError)

	router.POST("/auth/login", handler.Login)

	body := map[string]string{"username": "testuser", "password": "password"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandlerNew_Register_ServiceError_New(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(mocks.MockAuthService)
	mockUserSvc := new(mocks.MockUserService)

	handler := NewAuthHandlerNew(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, "", assert.AnError)

	router.POST("/auth/register", handler.Register)

	body := map[string]string{"username": "newuser", "password": "password", "email": "test@test.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthSvc.AssertExpectations(t)
}
