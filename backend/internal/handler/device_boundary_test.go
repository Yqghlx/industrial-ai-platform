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
// Device Handler Boundary Tests (补充)
// ============================================

func TestDeviceHandlerNew_CreateDevice_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.POST("/devices", handler.CreateDevice)

	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandlerNew_UpdateDevice_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.PUT("/devices/:id", handler.UpdateDevice)

	req := httptest.NewRequest(http.MethodPut, "/devices/device-1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandlerNew_CreateRule_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.POST("/rules", handler.CreateRule)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandlerNew_CreateUser_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	router.POST("/users", handler.CreateUser)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandlerNew_GetDeviceGraph_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("GetGraph", mock.Anything).Return(nil, assert.AnError)

	router.GET("/devices/graph", handler.GetDeviceGraph)

	req := httptest.NewRequest(http.MethodGet, "/devices/graph", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_ListRules_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockAlertSvc.On("GetRules", mock.Anything).Return(nil, assert.AnError)

	router.GET("/rules", handler.ListRules)

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetLatestTelemetry_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, func(msg model.WSMessage) {})

	mockTelemetrySvc.On("GetLatest", mock.Anything).Return(nil, assert.AnError)

	router.GET("/telemetry/latest", handler.GetLatestTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/latest", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockTelemetrySvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDeviceTelemetry_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, func(msg model.WSMessage) {})

	mockTelemetrySvc.On("GetLatestByDevice", mock.Anything, "device-1", mock.Anything).Return(nil, assert.AnError)

	router.GET("/devices/:id/telemetry", handler.GetDeviceTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/devices/device-1/telemetry", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockTelemetrySvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_DeleteDevice_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("Delete", mock.Anything, "device-1").Return(assert.AnError)

	router.DELETE("/devices/:id", handler.DeleteDevice)

	req := httptest.NewRequest(http.MethodDelete, "/devices/device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDevice_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("GetByID", mock.Anything, "device-1").Return(nil, assert.AnError)

	router.GET("/devices/:id", handler.GetDevice)

	req := httptest.NewRequest(http.MethodGet, "/devices/device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_CreateDevice_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.Device")).Return(assert.AnError)

	router.POST("/devices", handler.CreateDevice)

	body := map[string]string{"id": "new-device", "name": "Test", "type": "sensor"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_UpdateDevice_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("Update", mock.Anything, mock.AnythingOfType("*model.Device")).Return(assert.AnError)

	router.PUT("/devices/:id", handler.UpdateDevice)

	body := map[string]string{"name": "Updated", "type": "sensor"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/devices/device-1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_CreateRule_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockAlertSvc.On("CreateRule", mock.Anything, mock.AnythingOfType("*model.AlertRule")).Return(assert.AnError)

	router.POST("/rules", handler.CreateRule)

	body := map[string]interface{}{"name": "Rule", "metric": "temp", "operator": ">", "threshold": 80.0}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_DeleteRule_ServiceError_New(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockAlertSvc.On("DeleteRule", mock.Anything, 1).Return(assert.AnError)

	router.DELETE("/rules/:id", handler.DeleteRule)

	req := httptest.NewRequest(http.MethodDelete, "/rules/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_ListDevices_ServiceError_New(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(service.MockDeviceService)
	mockAlertSvc := new(service.MockAlertService)
	mockAuthSvc := new(service.MockAuthService)

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, nil, func(msg model.WSMessage) {})

	mockDeviceSvc.On("List", mock.Anything, 1, 20).Return(nil, 0, assert.AnError)

	router.GET("/devices", handler.ListDevices)

	req := httptest.NewRequest(http.MethodGet, "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}
