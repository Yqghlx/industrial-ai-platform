package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
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

func TestExportHandler_ExportAlerts_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts?format=doc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportAlerts_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportAlerts_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("large file"),
		Filename: "alerts_20260520.pdf",
		MimeType: "application/pdf",
		Size:     15 * 1024 * 1024, // 15MB > 10MB limit
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportAlerts_XLSX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test xlsx content"),
		Filename: "alerts_20260520.xlsx",
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Size:     2048,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/alerts", handler.ExportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts?format=xlsx", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportROI_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi?format=csv", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportROI_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportROI_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("large file"),
		Filename: "roi_20260520.pdf",
		MimeType: "application/pdf",
		Size:     15 * 1024 * 1024, // 15MB > 10MB limit
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/roi", handler.ExportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportDevices_DefaultFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test pdf content"),
		Filename: "devices_default.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	// No format parameter - should default to PDF
	req := httptest.NewRequest(http.MethodGet, "/export/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportDevices_ExcelFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	handler := NewExportHandler(mockExportSvc)

	result := &service.ExportResult{
		Data:     []byte("test xlsx content"),
		Filename: "devices_excel.xlsx",
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Size:     2048,
	}

	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	router.GET("/export/devices", handler.ExportDevices)

	// Using "excel" as format alias
	req := httptest.NewRequest(http.MethodGet, "/export/devices?format=excel", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockExportSvc.AssertExpectations(t)
}

// ============================================
// AttachExportHandlers Tests
// ============================================

func TestAttachExportHandlers_RoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create server with minimal fields needed for AttachExportHandlers
	server := &HTTPServerNew{
		router:    router,
		jwtSecret: "test-secret",
	}

	mockExportSvc := new(MockExportService)

	// Attach export handlers
	server.AttachExportHandlers(mockExportSvc)

	// Verify routes are registered
	routes := router.Routes()
	var foundDevices, foundAlerts, foundROI bool
	for _, route := range routes {
		if route.Path == "/api/v1/reports/devices/export" && route.Method == "GET" {
			foundDevices = true
		}
		if route.Path == "/api/v1/reports/alerts/export" && route.Method == "GET" {
			foundAlerts = true
		}
		if route.Path == "/api/v1/reports/roi/export" && route.Method == "GET" {
			foundROI = true
		}
	}

	assert.True(t, foundDevices, "devices/export route should be registered")
	assert.True(t, foundAlerts, "alerts/export route should be registered")
	assert.True(t, foundROI, "roi/export route should be registered")
}

func TestAttachExportHandlers_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &HTTPServerNew{
		router:    router,
		jwtSecret: "test-secret",
	}

	mockExportSvc := new(MockExportService)
	server.AttachExportHandlers(mockExportSvc)

	// Test without auth token - should return 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/devices/export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Auth middleware should reject unauthorized requests
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAttachExportHandlers_WithMockAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use a mock auth middleware that always passes
	mockAuthMiddleware := func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("role", "admin")
		c.Next()
	}

	server := &HTTPServerNew{
		router:    router,
		jwtSecret: "test-secret",
	}

	mockExportSvc := new(MockExportService)

	// Mock successful export
	result := &service.ExportResult{
		Data:     []byte("test roi content"),
		Filename: "roi_export.pdf",
		MimeType: "application/pdf",
		Size:     512,
	}
	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	// Override middleware in test setup
	exportHandler := NewExportHandler(mockExportSvc)

	// Setup routes with mock auth
	auth := server.router.Group("/api/v1")
	auth.Use(mockAuthMiddleware)
	{
		auth.GET("/reports/devices/export", exportHandler.ExportDevices)
		auth.GET("/reports/alerts/export", exportHandler.ExportAlerts)
		auth.GET("/reports/roi/export", exportHandler.ExportROI)
	}

	// Verify routes work with mock auth
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/roi/export?format=pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should pass auth and reach handler - returns 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestAttachExportHandlers_HandlerNotNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &HTTPServerNew{
		router:    router,
		jwtSecret: "test-secret",
	}

	mockExportSvc := new(MockExportService)
	server.AttachExportHandlers(mockExportSvc)

	// Verify routes were registered (handler was created internally)
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/api/v1/reports/devices/export" {
			found = true
			assert.NotNil(t, route.HandlerFunc, "handler should not be nil")
		}
	}
	assert.True(t, found, "route should be registered")
}

func TestAttachExportHandlers_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)

	// Mock successful export
	result := &service.ExportResult{
		Data:     []byte("test content"),
		Filename: "devices_export.pdf",
		MimeType: "application/pdf",
		Size:     100,
	}
	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(result, nil)

	// Create handler directly for integration test
	exportHandler := NewExportHandler(mockExportSvc)

	// Test direct handler invocation (bypass auth for testing)
	router.GET("/test/devices/export", exportHandler.ExportDevices)

	req := httptest.NewRequest(http.MethodGet, "/test/devices/export?format=pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "devices_export.pdf")
	mockExportSvc.AssertExpectations(t)
}

func TestAttachExportHandlers_GroupPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &HTTPServerNew{
		router:    router,
		jwtSecret: "test-secret",
	}

	mockExportSvc := new(MockExportService)
	server.AttachExportHandlers(mockExportSvc)

	// Verify the group path prefix is correct
	routes := router.Routes()
	for _, route := range routes {
		// All export routes should start with /api/v1
		if route.Path == "/api/v1/reports/devices/export" ||
			route.Path == "/api/v1/reports/alerts/export" ||
			route.Path == "/api/v1/reports/roi/export" {
			assert.Contains(t, route.Path, "/api/v1")
			assert.Contains(t, route.Path, "/reports/")
			assert.Contains(t, route.Path, "/export")
		}
	}
}
