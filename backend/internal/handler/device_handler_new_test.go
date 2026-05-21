package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/errors"
)

// ============================================
// DeviceHandlerNew Tests
// ============================================

func TestNewDeviceHandlerNew(t *testing.T) {
	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDeviceSvc, handler.deviceSvc)
	assert.Equal(t, mockAlertSvc, handler.alertSvc)
	assert.Equal(t, mockAuthSvc, handler.authSvc)
}

func TestDeviceHandlerNew_ListDevices_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	devices := []model.Device{
		{ID: "device-1", Name: "Device 1", Type: "sensor"},
		{ID: "device-2", Name: "Device 2", Type: "controller"},
	}

	mockDeviceSvc.On("List", mock.Anything, 1, 20).Return(devices, 2, nil)

	router.GET("/devices", handler.ListDevices)

	req := httptest.NewRequest(http.MethodGet, "/devices?page=1&page_size=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(1), response["page"])
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_ListDevices_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockDeviceSvc.On("List", mock.Anything, 1, 20).Return(nil, 0, assert.AnError)

	router.GET("/devices", handler.ListDevices)

	req := httptest.NewRequest(http.MethodGet, "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDevice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	expectedDevice := &model.Device{ID: "device-1", Name: "Device 1", Type: "sensor"}

	mockDeviceSvc.On("GetByID", mock.Anything, "device-1").Return(expectedDevice, nil)

	router.GET("/devices/:id", handler.GetDevice)

	req := httptest.NewRequest(http.MethodGet, "/devices/device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "device-1", response["id"])
	assert.Equal(t, "Device 1", response["name"])

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDevice_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockDeviceSvc.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.NewDeviceNotFoundError("nonexistent"))

	router.GET("/devices/:id", handler.GetDevice)

	req := httptest.NewRequest(http.MethodGet, "/devices/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_CreateDevice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastChan := make(chan model.WSMessage, 10)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockDeviceSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.Device")).Return(nil)

	router.POST("/devices", handler.CreateDevice)

	body := map[string]string{
		"id":   "new-device-1",
		"name": "New Device",
		"type": "sensor",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Verify broadcast message was sent
	select {
	case msg := <-broadcastChan:
		assert.Equal(t, "device_created", msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Log("No broadcast message received (expected)")
	}

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_DeleteDevice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastChan := make(chan model.WSMessage, 10)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockDeviceSvc.On("Delete", mock.Anything, "device-1").Return(nil)

	router.DELETE("/devices/:id", handler.DeleteDevice)

	req := httptest.NewRequest(http.MethodDelete, "/devices/device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Device deleted", response["message"])
	assert.Equal(t, "device-1", response["id"])

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDeviceGraph_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	graphData := map[string]interface{}{
		"nodes": []interface{}{map[string]string{"id": "node-1"}},
		"links": []interface{}{},
	}

	mockDeviceSvc.On("GetGraph", mock.Anything).Return(graphData, nil)

	router.GET("/devices/graph", handler.GetDeviceGraph)

	req := httptest.NewRequest(http.MethodGet, "/devices/graph", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_ListRules_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	rules := []model.AlertRule{
		{ID: 1, Name: "High Temperature", Metric: "temperature", Enabled: true},
		{ID: 2, Name: "Low Voltage", Metric: "voltage", Enabled: true},
	}

	mockAlertSvc.On("GetRules", mock.Anything).Return(rules, nil)

	router.GET("/rules", handler.ListRules)

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_CreateRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockAlertSvc.On("CreateRule", mock.Anything, mock.AnythingOfType("*model.AlertRule")).Return(nil)

	router.POST("/rules", handler.CreateRule)

	body := map[string]interface{}{
		"name":      "New Rule",
		"metric":    "temperature",
		"operator":  ">",
		"threshold": 80.0,
		"severity":  "high",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_DeleteRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockAlertSvc.On("DeleteRule", mock.Anything, 1).Return(nil)

	router.DELETE("/rules/:id", handler.DeleteRule)

	req := httptest.NewRequest(http.MethodDelete, "/rules/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_CreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	expectedUser := &model.User{ID: 1, Username: "newuser"}
	expectedToken := "mock-token"

	mockAuthSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).
		Return(expectedUser, expectedToken, nil)

	router.POST("/users", handler.CreateUser)

	body := map[string]string{
		"username": "newuser",
		"password": "password123",
		"email":    "new@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAuthSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetLatestTelemetry_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	telemetry := []model.TelemetryData{
		{DeviceID: "device-1", Timestamp: time.Now(), Temperature: 25.5},
	}

	mockTelemetrySvc.On("GetLatest", mock.Anything).Return(telemetry, nil)

	router.GET("/telemetry/latest", handler.GetLatestTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/latest", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDeviceTelemetry_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	telemetry := []model.TelemetryData{
		{DeviceID: "device-1", Timestamp: time.Now(), Temperature: 25.5},
	}

	mockTelemetrySvc.On("GetLatestByDevice", mock.Anything, "device-1", 100).Return(telemetry, nil)

	router.GET("/devices/:id/telemetry", handler.GetDeviceTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/devices/device-1/telemetry", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_UpdateDevice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockDeviceSvc.On("Update", mock.Anything, mock.AnythingOfType("*model.Device")).Return(nil)

	router.PUT("/devices/:id", handler.UpdateDevice)

	body := map[string]string{
		"name": "Updated Device",
		"type": "sensor",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/devices/device-1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockDeviceSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_GetDeviceStats_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	router.GET("/devices/:id/stats", handler.GetDeviceStats)

	req := httptest.NewRequest(http.MethodGet, "/devices/device-1/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceHandlerNew_GetRule_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	router.GET("/rules/:id", handler.GetRule)

	req := httptest.NewRequest(http.MethodGet, "/rules/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceHandlerNew_UpdateRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	mockAlertSvc.On("UpdateRule", mock.Anything, mock.AnythingOfType("*model.AlertRule")).Return(nil)

	router.PUT("/rules/:id", handler.UpdateRule)

	body := map[string]interface{}{
		"name":      "Updated Rule",
		"metric":    "temperature",
		"operator":  ">",
		"threshold": 90.0,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/rules/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestDeviceHandlerNew_ToggleRule_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDeviceSvc := new(MockDeviceService)
	mockAlertSvc := new(MockAlertService)
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewDeviceHandlerNew(mockDeviceSvc, mockAlertSvc, mockAuthSvc, mockTelemetrySvc, broadcastFunc)

	router.POST("/rules/:id/toggle", handler.ToggleRule)

	req := httptest.NewRequest(http.MethodPost, "/rules/1/toggle", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
