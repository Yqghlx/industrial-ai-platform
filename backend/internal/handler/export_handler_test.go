package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

)

// ============================================
// ExportHandler Tests
// ============================================

func TestNewExportHandler(t *testing.T) {
	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockExportSvc, handler.exportSvc)
}

func TestExportHandler_ExportDevices_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test pdf content"),
		Filename: "devices_20260520.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "attachment; filename=\"devices_20260520.pdf\"", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportDevices_XLSX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test xlsx content"),
		Filename: "devices_20260520.xlsx",
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Size:     2048,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=xlsx", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportDevices_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportDevices_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportDevices_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("large file"),
		Filename: "devices_20260520.pdf",
		MimeType: "application/pdf",
		Size:     15 * 1024 * 1024, // 15MB > 10MB limit
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportAlerts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test pdf content"),
		Filename: "alerts_20260520.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts?format=pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportAlerts_WithDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test pdf content"),
		Filename: "alerts_20260520.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts?start_date=2026-05-01&end_date=2026-05-20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportROI_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test pdf content"),
		Filename: "roi_20260520.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi?format=pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportROI_XLSX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test xlsx content"),
		Filename: "roi_20260520.xlsx",
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Size:     2048,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi?format=xlsx", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	mockExportSvc.AssertExpectations(t)
}

func TestParseDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		start, end := parseDateRange(c)
		c.JSON(http.StatusOK, gin.H{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		})
	})

	// Test with provided dates
	req := httptest.NewRequest(http.MethodGet, "/test?start_date=2026-05-01&end_date=2026-05-20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Test with default dates (no query params)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	// Default should be last 24 hours
	assert.NotEmpty(t, response["start"])
	assert.NotEmpty(t, response["end"])
}

func TestExportHandler_ExportDevices_InvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	// 无效格式
	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=doc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportDevices_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte{},
		Filename: "empty.pdf",
		MimeType: "application/pdf",
		Size:     0,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockExportSvc.AssertExpectations(t)
}
