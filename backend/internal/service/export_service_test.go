package service

import (
	"context"
	"testing"
	"time"

	"github.com/industrial-ai/platform/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

func newTestExportService(t *testing.T) (*ExportService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))
	reportRepo := repository.NewReportRepository(database.NewDBWrapper(db))
	notificationRepo := repository.NewNotificationRepository(database.NewDBWrapper(db))

	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, notificationRepo)

	svc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	return svc, mock
}

func TestNewExportService(t *testing.T) {
	svc, _ := newTestExportService(t)
	assert.NotNil(t, svc)
}

func TestExportService_Export_UnsupportedReportType(t *testing.T) {
	svc, _ := newTestExportService(t)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "unsupported",
		Format:     FormatPDF,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	result, err := svc.Export(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unsupported report type")
}

func TestExportService_Export_UnsupportedFormat(t *testing.T) {
	svc, _ := newTestExportService(t)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "devices",
		Format:     "xml",
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	result, err := svc.Export(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unsupported export format")
}

func TestExportService_Export_DefaultDates(t *testing.T) {
	svc, _ := newTestExportService(t)
	ctx := context.Background()

	// Empty dates should be auto-filled
	req := &ExportRequest{
		ReportType: "devices",
		Format:     FormatPDF,
		StartDate:  time.Time{}, // zero
		EndDate:    time.Time{}, // zero
	}

	// This will fail at DB level but it tests the default date logic path
	_, _ = svc.Export(ctx, req)
	// Just verifying it doesn't panic
}

func TestGetExportFilename(t *testing.T) {
	tests := []struct {
		name       string
		reportType string
		format     ExportFormat
		expected   string
	}{
		{
			name:       "devices PDF",
			reportType: "devices",
			format:     FormatPDF,
			expected:   ".pdf",
		},
		{
			name:       "alerts XLSX",
			reportType: "alerts",
			format:     FormatXLSX,
			expected:   ".xlsx",
		},
		{
			name:       "roi PDF",
			reportType: "roi",
			format:     FormatPDF,
			expected:   ".pdf",
		},
		{
			name:       "unknown type",
			reportType: "unknown",
			format:     FormatPDF,
			expected:   ".pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := GetExportFilename(tt.reportType, tt.format)
			assert.Contains(t, filename, tt.expected)
			assert.Contains(t, filename, time.Now().Format("20060102"))
		})
	}
}

func TestGetExportFilename_Prefixes(t *testing.T) {
	tests := []struct {
		reportType     string
		expectedPrefix string
	}{
		{"devices", "设备状态报告"},
		{"alerts", "告警统计报告"},
		{"roi", "ROI分析报告"},
		{"unknown", "报告"},
	}

	for _, tt := range tests {
		t.Run(tt.reportType, func(t *testing.T) {
			filename := GetExportFilename(tt.reportType, FormatPDF)
			assert.Contains(t, filename, tt.expectedPrefix)
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		name     string
		format   ExportFormat
		expected string
	}{
		{"PDF", FormatPDF, "application/pdf"},
		{"XLSX", FormatXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"Unknown", "unknown", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime := GetMimeType(tt.format)
			assert.Equal(t, tt.expected, mime)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExportResult_Fields(t *testing.T) {
	result := &ExportResult{
		Data:     []byte("test data"),
		Filename: "test.pdf",
		MimeType: "application/pdf",
		Size:     9,
	}
	assert.Equal(t, []byte("test data"), result.Data)
	assert.Equal(t, "test.pdf", result.Filename)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.Equal(t, int64(9), result.Size)
}

func TestExportFormat_Constants(t *testing.T) {
	assert.Equal(t, ExportFormat("pdf"), FormatPDF)
	assert.Equal(t, ExportFormat("xlsx"), FormatXLSX)
}

func TestDeviceReportData(t *testing.T) {
	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Devices: []model.Device{
			{ID: "d1", Name: "Device 1", Type: "sensor", Location: "Room A", Status: "online"},
		},
		Summary: DeviceSummary{
			TotalDevices:   1,
			OnlineDevices:  1,
			OfflineDevices: 0,
		},
	}
	assert.Equal(t, 1, data.Summary.TotalDevices)
}

func TestAlertReportData(t *testing.T) {
	data := &AlertReportData{
		GeneratedAt: time.Now(),
		Alerts:      []model.Alert{},
		AlertStats: AlertStats{
			TotalAlerts: 5,
		},
	}
	assert.Equal(t, 5, data.AlertStats.TotalAlerts)
}

func TestROIReportData(t *testing.T) {
	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     10,
			PredictedSavings: 50000.0,
			UptimePercentage: 99.5,
		},
	}
	assert.Equal(t, 10, data.ROIStats.TotalDevices)
}


// ============================================
// PDF/XLSX 文件名和 MIME 类型验证
// ============================================

func TestExportPDF_FilenameAndMimeType(t *testing.T) {
	svc := &ExportService{}
	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary:     DeviceSummary{TotalDevices: 1, OnlineDevices: 1},
		Devices:     []model.Device{{ID: "d1", Name: "Device1", Type: "pump", Location: "A", Status: "online"}},
	}

	result, err := svc.exportPDF(data, "devices", "report")
	require.NoError(t, err)
	assert.Equal(t, "report.pdf", result.Filename)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.True(t, len(result.Data) > 0)
}

func TestExportXLSX_FilenameAndMimeType(t *testing.T) {
	svc := &ExportService{}
	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary:     DeviceSummary{TotalDevices: 1, OnlineDevices: 1},
		Devices:     []model.Device{{ID: "d1", Name: "Device1", Type: "pump", Location: "A", Status: "online"}},
	}

	result, err := svc.exportXLSX(data, "devices", "report")
	require.NoError(t, err)
	assert.Equal(t, "report.xlsx", result.Filename)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.MimeType)
	assert.True(t, len(result.Data) > 0)
}

func TestExportPDF_AlertReport(t *testing.T) {
	svc := &ExportService{}
	data := &AlertReportData{
		GeneratedAt: time.Now(),
		AlertStats:  AlertStats{TotalAlerts: 10, ActiveAlerts: 3, ResolvedAlerts: 7},
	}

	result, err := svc.exportPDF(data, "alerts", "alert_report")
	require.NoError(t, err)
	assert.Equal(t, "alert_report.pdf", result.Filename)
	assert.Contains(t, string(result.Data), "告警统计报告")
}

func TestExportPDF_ROIReport(t *testing.T) {
	svc := &ExportService{}
	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats:    model.ROIStats{TotalDevices: 5, PredictedSavings: 10000, UptimePercentage: 99.5},
	}

	result, err := svc.exportPDF(data, "roi", "roi_report")
	require.NoError(t, err)
	assert.Equal(t, "roi_report.pdf", result.Filename)
	assert.Contains(t, string(result.Data), "ROI分析报告")
}
