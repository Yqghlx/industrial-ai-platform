package handler

import (
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
	"github.com/industrial-ai/platform/internal/mocks"
	"github.com/industrial-ai/platform/pkg/errors"
)

// ============================================
// Phase 1 Test: 新架构AlertHandler测试
// ============================================

func TestNewAlertHandler_ListAlerts_Success(t *testing.T) {
	router := gin.New()
	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)

	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	alerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "high", Status: "active", TriggeredAt: time.Now()},
		{ID: 2, DeviceID: "device-2", Severity: "critical", Status: "active", TriggeredAt: time.Now()},
	}

	// P0-03: Use GetAlertsWithFilter instead of GetAlerts
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "active", "", "", 1, 20).Return(alerts, 2, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts?status=active", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "total")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockAlertSvc.AssertExpectations(t)
}

func TestNewAlertHandler_GetAlert_Success(t *testing.T) {
	router := gin.New()
	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)

	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	alert := &model.Alert{
		ID:          1,
		DeviceID:    "device-001",
		Severity:    "critical",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	mockAlertSvc.On("GetAlertByID", mock.Anything, 1).Return(alert, nil)

	router.GET("/alerts/:id", handler.GetAlert)

	req := httptest.NewRequest(http.MethodGet, "/alerts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response model.Alert
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.ID)
	assert.Equal(t, "device-001", response.DeviceID)

	mockAlertSvc.AssertExpectations(t)
}

func TestNewAlertHandler_ResolveAlert_Success(t *testing.T) {
	router := gin.New()
	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)

	broadcastFunc := func(msg model.WSMessage) {
		select {
		case broadcastChan <- msg:
		default: // 防止阻塞
		}
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	mockAlertSvc.On("ResolveAlert", mock.Anything, 1).Return(nil)

	router.POST("/alerts/:id/resolve", handler.ResolveAlert)

	req := httptest.NewRequest(http.MethodPost, "/alerts/1/resolve", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Alert resolved", response["message"])
	assert.Equal(t, float64(1), response["id"])

	mockAlertSvc.AssertExpectations(t)
}

func TestNewAlertHandler_AcknowledgeAlert_Success(t *testing.T) {
	router := gin.New()
	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)

	broadcastFunc := func(msg model.WSMessage) {
		select {
		case broadcastChan <- msg:
		default: // 防止阻塞
		}
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	mockAlertSvc.On("AcknowledgeAlert", mock.Anything, 1).Return(nil)

	router.POST("/alerts/:id/acknowledge", handler.AcknowledgeAlert)

	req := httptest.NewRequest(http.MethodPost, "/alerts/1/acknowledge", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

// ============================================
// AlertHandler Additional Tests (Trend/Ranking/Efficiency)
// ============================================

func TestAlertHandler_GetTrend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	router.GET("/alerts/trend", handler.GetTrend)

	req := httptest.NewRequest(http.MethodGet, "/alerts/trend?period=30d", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "30d", response["period"])
}

func TestAlertHandler_GetRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	router.GET("/alerts/ranking", handler.GetRanking)

	req := httptest.NewRequest(http.MethodGet, "/alerts/ranking?limit=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(20), response["limit"])
}

func TestAlertHandler_GetEfficiency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	router.GET("/alerts/efficiency", handler.GetEfficiency)

	req := httptest.NewRequest(http.MethodGet, "/alerts/efficiency", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "efficiency")
}

func TestAlertHandler_ListAlerts_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	// P0-03: Use GetAlertsWithFilter instead of GetAlerts
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "all", "", "", 1, 20).Return(nil, 0, assert.AnError)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_GetAlert_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	mockAlertSvc.On("GetAlertByID", mock.Anything, 1).Return(nil, errors.NewNotFoundError("Alert", "1"))

	router.GET("/alerts/:id", handler.GetAlert)

	req := httptest.NewRequest(http.MethodGet, "/alerts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ResolveAlert_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	mockAlertSvc.On("ResolveAlert", mock.Anything, 1).Return(assert.AnError)

	router.POST("/alerts/:id/resolve", handler.ResolveAlert)

	req := httptest.NewRequest(http.MethodPost, "/alerts/1/resolve", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_AcknowledgeAlert_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	mockAlertSvc.On("AcknowledgeAlert", mock.Anything, 1).Return(assert.AnError)

	router.POST("/alerts/:id/acknowledge", handler.AcknowledgeAlert)

	req := httptest.NewRequest(http.MethodPost, "/alerts/1/acknowledge", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_WithSeverityFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	// P0-03: Use GetAlertsWithFilter - severity filter is passed to service
	filteredAlerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "critical", Status: "active", TriggeredAt: time.Now()},
		{ID: 3, DeviceID: "device-3", Severity: "critical", Status: "active", TriggeredAt: time.Now()},
	}
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "all", "critical", "", 1, 20).Return(filteredAlerts, 2, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts?severity=critical", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	// Should return 2 critical alerts (filtered at database level)
	assert.Len(t, data, 2)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_WithDeviceIDFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	// P0-03: Use GetAlertsWithFilter - deviceID filter is passed to service
	filteredAlerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "high", Status: "active", TriggeredAt: time.Now()},
	}
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "all", "", "device-1", 1, 20).Return(filteredAlerts, 1, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts?device_id=device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	// Should return 1 alert (filtered at database level)
	assert.Len(t, data, 1)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_WithBothFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	// P0-03: Use GetAlertsWithFilter - both filters passed to service
	filteredAlerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "critical", Status: "active", TriggeredAt: time.Now()},
	}
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "all", "critical", "device-1", 1, 20).Return(filteredAlerts, 1, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts?severity=critical&device_id=device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	// Should return 1 alert (critical + device-1, filtered at database level)
	assert.Len(t, data, 1)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_DefaultStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	alerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "high", Status: "active", TriggeredAt: time.Now()},
	}

	// P0-03: Default status should be "all", use GetAlertsWithFilter
	mockAlertSvc.On("GetAlertsWithFilter", mock.Anything, "all", "", "", 1, 20).Return(alerts, 1, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_CustomPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAlertSvc := new(mocks.MockAlertService)
	broadcastChan := make(chan model.WSMessage, 100)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	handler := NewAlertHandler(mockAlertSvc, broadcastFunc)

	alerts := []model.Alert{
		{ID: 1, DeviceID: "device-1", Severity: "high", Status: "active", TriggeredAt: time.Now()},
	}

	mockAlertSvc.On("GetAlerts", mock.Anything, "all", 2, 50).Return(alerts, 1, nil)

	router.GET("/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/alerts?page=2&page_size=50", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}
