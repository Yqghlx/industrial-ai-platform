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
// BusinessHandlerNew Tests
// ============================================

func TestNewBusinessHandlerNew(t *testing.T) {
	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockWorkOrderSvc, handler.workOrderSvc)
	assert.Equal(t, mockNotificationSvc, handler.notificationSvc)
	assert.Equal(t, mockBlackBoxSvc, handler.blackBoxSvc)
	assert.Equal(t, mockReportSvc, handler.reportSvc)
	assert.Equal(t, mockAlertSvc, handler.alertSvc)
}

func TestBusinessHandlerNew_ListWorkOrders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	orders := []model.WorkOrder{
		{ID: 1, Title: "Order 1", Status: "pending", DeviceID: "device-1", CreatedAt: time.Now()},
		{ID: 2, Title: "Order 2", Status: "completed", DeviceID: "device-2", CreatedAt: time.Now()},
	}

	mockWorkOrderSvc.On("List", mock.Anything, "", "", 1, 20).Return(orders, 2, nil)

	router.GET("/work-orders", handler.ListWorkOrders)

	req := httptest.NewRequest(http.MethodGet, "/work-orders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(2), response["total"])
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockWorkOrderSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_ListWorkOrders_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	orders := []model.WorkOrder{
		{ID: 1, Title: "Order 1", Status: "pending", DeviceID: "device-1", CreatedAt: time.Now()},
	}

	mockWorkOrderSvc.On("List", mock.Anything, "pending", "device-1", 1, 20).Return(orders, 1, nil)

	router.GET("/work-orders", handler.ListWorkOrders)

	req := httptest.NewRequest(http.MethodGet, "/work-orders?status=pending&device_id=device-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockWorkOrderSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_CreateWorkOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	mockWorkOrderSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.WorkOrder")).Return(nil)

	router.POST("/work-orders", handler.CreateWorkOrder)

	body := map[string]string{
		"title":     "New Work Order",
		"device_id": "device-1",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/work-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockWorkOrderSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_UpdateWorkOrderStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	mockWorkOrderSvc.On("UpdateStatus", mock.Anything, 1, "completed").Return(nil)

	router.PUT("/work-orders/:id/status", handler.UpdateWorkOrderStatus)

	body := map[string]string{
		"status": "completed",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/work-orders/1/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockWorkOrderSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_ListNotifications_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	notifications := []model.Notification{
		{ID: 1, Type: "alert", Title: "Notification 1", CreatedAt: time.Now()},
	}

	mockNotificationSvc.On("List", mock.Anything, "", 1, 20).Return(notifications, 1, nil)

	router.GET("/notifications", handler.ListNotifications)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockNotificationSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_MarkNotificationRead_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	mockNotificationSvc.On("MarkRead", mock.Anything, 1).Return(nil)

	router.PUT("/notifications/:id/read", handler.MarkNotificationRead)

	req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockNotificationSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_ListBlackBox_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	records := []model.BlackBoxRecord{
		{ID: 1, DeviceID: "device-1", CreatedAt: time.Now()},
	}

	mockBlackBoxSvc.On("List", mock.Anything, "", 1, 20).Return(records, 1, nil)

	router.GET("/black-box", handler.ListBlackBox)

	req := httptest.NewRequest(http.MethodGet, "/black-box", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockBlackBoxSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_ListReports_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	reports := []model.Report{
		{ID: 1, Type: "weekly"},
	}

	mockReportSvc.On("ListReports", mock.Anything, "", 1, 20).Return(reports, 1, nil)

	router.GET("/reports", handler.ListReports)

	req := httptest.NewRequest(http.MethodGet, "/reports", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockReportSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_GenerateReport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	report := &model.Report{ID: 1, Type: "daily"}

	mockReportSvc.On("GenerateReport", mock.Anything, "daily", "device-1").Return(report, nil)

	router.POST("/reports/generate", handler.GenerateReport)

	body := map[string]string{
		"report_type": "daily",
		"device_id":   "device-1",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/reports/generate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockReportSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_GetROIStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	stats := &model.ROIStats{
		TotalDevices:     100,
		ActiveAlerts:     5,
		OpenWorkOrders:   10,
		ResolvedIssues:   50,
		PredictedSavings: 7750,
		UptimePercentage: 99.5,
		AvgResponseTime:  2.5,
	}

	mockReportSvc.On("GetROIStats", mock.Anything).Return(stats, nil)

	router.GET("/roi-stats", handler.GetROIStats)

	req := httptest.NewRequest(http.MethodGet, "/roi-stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockReportSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_GetAlertStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	alerts := []model.Alert{
		{ID: 1, Severity: "critical", Status: "active", TriggeredAt: time.Now()},
		{ID: 2, Severity: "high", Status: "resolved", TriggeredAt: time.Now()},
		{ID: 3, Severity: "medium", Status: "active", TriggeredAt: time.Now()},
	}

	mockAlertSvc.On("GetAlerts", mock.Anything, "all", 1, 1000).Return(alerts, 3, nil)

	router.GET("/alert-stats", handler.GetAlertStats)

	req := httptest.NewRequest(http.MethodGet, "/alert-stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockAlertSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_GetWorkOrder_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.GET("/work-orders/:id", handler.GetWorkOrder)

	req := httptest.NewRequest(http.MethodGet, "/work-orders/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestBusinessHandlerNew_ExportDevices_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestBusinessHandlerNew_ExportAlerts_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestBusinessHandlerNew_ExportROI_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestBusinessHandlerNew_GetBlackBoxData_Placeholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.GET("/black-box/:id", handler.GetBlackBoxData)

	req := httptest.NewRequest(http.MethodGet, "/black-box/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestBusinessHandlerNew_ListWorkOrders_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	mockWorkOrderSvc.On("List", mock.Anything, "", "", 1, 20).Return(nil, 0, assert.AnError)

	router.GET("/work-orders", handler.ListWorkOrders)

	req := httptest.NewRequest(http.MethodGet, "/work-orders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockWorkOrderSvc.AssertExpectations(t)
}

func TestBusinessHandlerNew_CreateWorkOrder_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.POST("/work-orders", handler.CreateWorkOrder)

	req := httptest.NewRequest(http.MethodPost, "/work-orders", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessHandlerNew_UpdateWorkOrderStatus_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.PUT("/work-orders/:id/status", handler.UpdateWorkOrderStatus)

	req := httptest.NewRequest(http.MethodPut, "/work-orders/1/status", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessHandlerNew_GenerateReport_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	router.POST("/reports/generate", handler.GenerateReport)

	req := httptest.NewRequest(http.MethodPost, "/reports/generate", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessHandlerNew_ListNotifications_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockWorkOrderSvc := new(MockWorkOrderService)
	mockNotificationSvc := new(MockNotificationService)
	mockBlackBoxSvc := new(MockBlackBoxService)
	mockReportSvc := new(MockReportService)
	mockAlertSvc := new(MockAlertService)

	broadcastFunc := func(msg model.WSMessage) {}

	handler := NewBusinessHandlerNew(mockWorkOrderSvc, mockNotificationSvc, mockBlackBoxSvc, mockReportSvc, mockAlertSvc, broadcastFunc)

	mockNotificationSvc.On("List", mock.Anything, "", 1, 20).Return(nil, 0, assert.AnError)

	router.GET("/notifications", handler.ListNotifications)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockNotificationSvc.AssertExpectations(t)
}
