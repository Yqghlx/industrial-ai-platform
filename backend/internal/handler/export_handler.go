package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/middleware"
	"github.com/industrial-ai/platform/internal/service"
)

// ExportHandler handles export requests
type ExportHandler struct {
	exportSvc service.ExportServiceInterface
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportSvc service.ExportServiceInterface) *ExportHandler {
	return &ExportHandler{
		exportSvc: exportSvc,
	}
}

// ExportDevices exports device report
// GET /api/v1/reports/devices/export?format=pdf|xlsx
func (h *ExportHandler) ExportDevices(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "pdf"
	}

	// Parse format
	exportFormat := service.FormatPDF
	if format == "xlsx" || format == "excel" {
		exportFormat = service.FormatXLSX
	} else if format != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use 'pdf' or 'xlsx'"})
		return
	}

	// Parse date range
	startDate, endDate := parseDateRange(c)

	// Create export request
	req := &service.ExportRequest{
		ReportType: "devices",
		Format:     exportFormat,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Export
	result, err := h.exportSvc.Export(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to export devices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	// Check file size constraint (<10MB)
	if result.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导出文件过大，请缩小时间范围"})
		return
	}

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	c.Header("Content-Type", result.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", result.Size))

	// Send file content
	c.Data(http.StatusOK, result.MimeType, result.Data)
}

// ExportAlerts exports alert report
// GET /api/v1/reports/alerts/export?format=pdf|xlsx
func (h *ExportHandler) ExportAlerts(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "pdf"
	}

	// Parse format
	exportFormat := service.FormatPDF
	if format == "xlsx" || format == "excel" {
		exportFormat = service.FormatXLSX
	} else if format != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use 'pdf' or 'xlsx'"})
		return
	}

	// Parse date range
	startDate, endDate := parseDateRange(c)

	// Create export request
	req := &service.ExportRequest{
		ReportType: "alerts",
		Format:     exportFormat,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Export
	result, err := h.exportSvc.Export(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to export alerts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	// Check file size constraint (<10MB)
	if result.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导出文件过大，请缩小时间范围"})
		return
	}

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	c.Header("Content-Type", result.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", result.Size))

	// Send file content
	c.Data(http.StatusOK, result.MimeType, result.Data)
}

// ExportROI exports ROI report
// GET /api/v1/reports/roi/export?format=pdf|xlsx
func (h *ExportHandler) ExportROI(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "pdf"
	}

	// Parse format
	exportFormat := service.FormatPDF
	if format == "xlsx" || format == "excel" {
		exportFormat = service.FormatXLSX
	} else if format != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use 'pdf' or 'xlsx'"})
		return
	}

	// Create export request
	req := &service.ExportRequest{
		ReportType: "roi",
		Format:     exportFormat,
	}

	// Export
	result, err := h.exportSvc.Export(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to export ROI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	// Check file size constraint (<10MB)
	if result.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导出文件过大"})
		return
	}

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	c.Header("Content-Type", result.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", result.Size))

	// Send file content
	c.Data(http.StatusOK, result.MimeType, result.Data)
}

// parseDateRange parses start and end date from query parameters
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	// Default to last 24 hours if not specified
	if startDate.IsZero() {
		startDate = time.Now().Add(-24 * time.Hour)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	return startDate, endDate
}

// AttachExportHandlers attaches export handlers to the server
func (s *Server) AttachExportHandlers(exportSvc service.ExportServiceInterface) {
	exportHandler := NewExportHandler(exportSvc)

	// Export routes
	auth := s.router.Group("/api/v1")
	auth.Use(middleware.AuthRequired(string(s.jwtSecret)))
	{
		auth.GET("/reports/devices/export", exportHandler.ExportDevices)
		auth.GET("/reports/alerts/export", exportHandler.ExportAlerts)
		auth.GET("/reports/roi/export", exportHandler.ExportROI)
	}
}
