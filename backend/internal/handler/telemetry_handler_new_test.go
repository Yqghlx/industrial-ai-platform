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
)

// ============================================
// TelemetryHandlerNew Tests
// ============================================

func TestNewTelemetryHandlerNew(t *testing.T) {
	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockTelemetrySvc, handler.telemetrySvc)
	assert.Equal(t, mockAgentSvc, handler.agentSvc)
}

func TestTelemetryHandlerNew_GetLatestTelemetry_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	telemetry := []model.TelemetryData{
		{DeviceID: "device-1", Timestamp: time.Now(), Temperature: 25.5},
		{DeviceID: "device-2", Timestamp: time.Now(), Pressure: 100.0},
	}

	mockTelemetrySvc.On("GetLatest", mock.Anything).Return(telemetry, nil)

	router.GET("/telemetry/latest", handler.GetLatestTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/latest", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetLatestTelemetry_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	mockTelemetrySvc.On("GetLatest", mock.Anything).Return(nil, assert.AnError)

	router.GET("/telemetry/latest", handler.GetLatestTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/latest", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetDeviceTelemetry_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	telemetry := []model.TelemetryData{
		{DeviceID: "device-1", Timestamp: time.Now(), Temperature: 25.5},
	}

	mockTelemetrySvc.On("GetLatestByDevice", mock.Anything, "device-1", 100).Return(telemetry, nil)

	router.GET("/telemetry/device/:id", handler.GetDeviceTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/device/device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "device-1", response["device_id"])
	data := response["data"].([]interface{})
	assert.Len(t, data, 1)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetDeviceTelemetry_WithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	telemetry := []model.TelemetryData{
		{DeviceID: "device-1", Timestamp: time.Now(), Temperature: 25.5},
	}

	mockTelemetrySvc.On("GetLatestByDevice", mock.Anything, "device-1", 50).Return(telemetry, nil)

	router.GET("/telemetry/device/:id", handler.GetDeviceTelemetry)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/device/device-1?limit=50", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetSystemStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	status := &model.SystemStatus{
		Database:    "connected",
		DBLatency:   5,
		Uptime:      "2 hours",
		Version:     "1.0",
		Timestamp:   time.Now(),
		DeviceCount: 100,
		UserCount:   10,
	}

	mockTelemetrySvc.On("GetSystemStatus", mock.Anything).Return(status, nil)

	router.GET("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetAIStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	taskLogs := []model.AgentTaskLog{
		{ID: 1, Query: "temperature analysis", Response: "completed", Agent: "ai-agent", ExecutedAt: time.Now()},
	}

	mockAgentSvc.On("GetTaskLogs", mock.Anything, 50).Return(taskLogs, nil)

	router.GET("/ai/status", handler.GetAIStatus)

	req := httptest.NewRequest(http.MethodGet, "/ai/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "active", response["status"])
	tasks := response["recent_tasks"].([]interface{})
	assert.Len(t, tasks, 1)

	mockAgentSvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetAIStatus_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	mockAgentSvc.On("GetTaskLogs", mock.Anything, 50).Return(nil, assert.AnError)

	router.GET("/ai/status", handler.GetAIStatus)

	req := httptest.NewRequest(http.MethodGet, "/ai/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAgentSvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_IngestTelemetry_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	router.POST("/telemetry/ingest", handler.IngestTelemetry)

	body := map[string]interface{}{
		"device_id":   "device-1",
		"temperature": 25.5,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/telemetry/ingest", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestTelemetryHandlerNew_AgentQuery_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	response := &model.AgentResponse{
		SessionID: "session-1",
		Response:  "Analysis complete",
		Agent:     "ai-agent",
		Timestamp: time.Now(),
	}

	mockAgentSvc.On("Query", mock.Anything, mock.AnythingOfType("model.AgentQuery")).Return(response, nil)

	router.POST("/ai/query", handler.AgentQuery)

	body := map[string]string{
		"query":     "What is the temperature trend?",
		"device_id": "device-1",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/ai/query", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	agentResp := resp["response"].(map[string]interface{})
	assert.Equal(t, "Analysis complete", agentResp["response"])

	mockAgentSvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_AgentQuery_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	// Empty query still binds successfully - this tests service error handling
	mockAgentSvc.On("Query", mock.Anything, mock.AnythingOfType("model.AgentQuery")).Return(nil, assert.AnError)

	router.POST("/ai/query", handler.AgentQuery)

	req := httptest.NewRequest(http.MethodPost, "/ai/query", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAgentSvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_GetSystemStatus_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	mockTelemetrySvc.On("GetSystemStatus", mock.Anything).Return(nil, assert.AnError)

	router.GET("/telemetry/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockTelemetrySvc.AssertExpectations(t)
}

func TestTelemetryHandlerNew_IngestTelemetry_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockTelemetrySvc := new(MockTelemetryService)
	mockAgentSvc := new(MockAgentService)

	handler := NewTelemetryHandlerNew(mockTelemetrySvc, mockAgentSvc)

	router.POST("/telemetry/ingest", handler.IngestTelemetry)

	req := httptest.NewRequest(http.MethodPost, "/telemetry/ingest", bytes.NewBuffer([]byte("invalid json body")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
