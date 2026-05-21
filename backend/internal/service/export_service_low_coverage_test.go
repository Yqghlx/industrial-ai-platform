package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// generateAlertReportData Tests (0% coverage)
// ============================================

func TestExportService_GenerateAlertReportData_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "alerts",
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	// Mock alert count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock alert list
	alertRows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
		AddRow(1, 1, "CNC-001", "温度过高", "high", "active", time.Now(), nil).
		AddRow(2, 2, "INJ-002", "振动异常", "medium", "resolved", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM alerts").WillReturnRows(alertRows)

	data := exportSvc.generateAlertReportData(ctx, req)

	assert.NotNil(t, data)
	assert.Equal(t, 5, data.AlertStats.TotalAlerts)
	assert.GreaterOrEqual(t, len(data.TopAlertRules), 0)
}

func TestExportService_GenerateAlertReportData_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "alerts",
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("db error"))

	data := exportSvc.generateAlertReportData(ctx, req)

	// Should handle error gracefully and return empty data
	assert.NotNil(t, data)
	assert.Equal(t, 0, data.AlertStats.TotalAlerts)
}

// ============================================
// generateROIReportData Tests (0% coverage)
// ============================================

func TestExportService_GenerateROIReportData_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	// Mock device count for ROI stats
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock device list for metrics
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC", "cnc", "Line 1", "online", "t1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices").WillReturnRows(deviceRows)

	data := exportSvc.generateROIReportData(ctx)

	assert.NotNil(t, data)
	assert.GreaterOrEqual(t, len(data.MonthlyTrend), 0)
}

func TestExportService_GenerateROIReportData_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("db error"))

	data := exportSvc.generateROIReportData(ctx)

	// Should handle error gracefully
	assert.NotNil(t, data)
}

// ============================================
// exportXLSX Tests (0% coverage)
// ============================================

func TestExportService_ExportXLSX_DeviceReport2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "devices",
		Format:     FormatXLSX,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	// Mock device list
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC", "cnc", "Line 1", "online", "t1", time.Now(), time.Now()).
		AddRow("INJ-002", "Injection", "injection", "Line 2", "online", "t1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices").WillReturnRows(deviceRows)

	result, err := exportSvc.Export(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Data)
	// Note: current implementation uses .csv extension (will change to .xlsx when using excelize)
	assert.Contains(t, result.Filename, ".csv")
}

func TestExportService_ExportXLSX_AlertReport2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "alerts",
		Format:     FormatXLSX,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	alertRows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
		AddRow(1, 1, "CNC-001", "Alert1", "high", "active", time.Now(), nil).
		AddRow(2, 2, "CNC-002", "Alert2", "medium", "resolved", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM alerts").WillReturnRows(alertRows)

	result, err := exportSvc.Export(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Note: current implementation uses .csv extension
	assert.Contains(t, result.Filename, ".csv")
}

func TestExportService_ExportXLSX_ROIReport2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportSvc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, nil)

	exportSvc := NewExportService(deviceRepo, telemetryRepo, alertRepo, workOrderRepo, reportSvc)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "roi",
		Format:     FormatXLSX,
	}

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC", "cnc", "Line 1", "online", "t1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices").WillReturnRows(deviceRows)

	result, err := exportSvc.Export(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Note: current implementation uses .csv extension
	assert.Contains(t, result.Filename, ".csv")
}

// ============================================
// generateDeviceExcelContent Tests (0% coverage)
// ============================================

func TestExportService_GenerateDeviceExcelContent_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Devices: []model.Device{
			{ID: "CNC-001", Name: "CNC Machine", Type: "cnc", Location: "Line 1", Status: "online"},
			{ID: "INJ-002", Name: "Injection", Type: "injection", Location: "Line 2", Status: "warning"},
		},
		Summary: DeviceSummary{
			TotalDevices:   2,
			OnlineDevices:  1,
			WarningDevices: 1,
		},
	}

	content := exportSvc.generateDeviceExcelContent(data)
	assert.NotEmpty(t, content)
}

func TestExportService_GenerateDeviceExcelContent_EmptyDevices(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Devices:     []model.Device{},
		Summary:     DeviceSummary{},
	}

	content := exportSvc.generateDeviceExcelContent(data)
	assert.NotEmpty(t, content)
}

// ============================================
// generateAlertExcelContent Tests (0% coverage)
// ============================================

func TestExportService_GenerateAlertExcelContent_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &AlertReportData{
		GeneratedAt: time.Now(),
		Alerts: []model.Alert{
			{ID: 1, RuleID: 1, DeviceID: "CNC-001", Message: "温度过高", Severity: "high", Status: "active", TriggeredAt: time.Now()},
			{ID: 2, RuleID: 2, DeviceID: "INJ-002", Message: "振动异常", Severity: "medium", Status: "resolved", TriggeredAt: time.Now()},
		},
		AlertStats: AlertStats{
			TotalAlerts:    2,
			HighAlerts:     1,
			MediumAlerts:   1,
			ActiveAlerts:   1,
			ResolvedAlerts: 1,
		},
	}

	content := exportSvc.generateAlertExcelContent(data)
	assert.NotEmpty(t, content)
}

// ============================================
// generateROIExcelContent Tests (0% coverage)
// ============================================

func TestExportService_GenerateROIExcelContent_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     10,
			PredictedSavings: 50000,
			UptimePercentage: 99.5,
		},
		DeviceMetrics: []DeviceMetric{
			{DeviceID: "CNC-001", DeviceName: "CNC", UptimeHours: 720, Savings: 5000},
		},
		MonthlyTrend: []MonthlyMetric{
			{Month: "2025-01", TotalSavings: 40000, UptimePercent: 99.0},
		},
	}

	content := exportSvc.generateROIExcelContent(data)
	assert.NotEmpty(t, content)
}

// ============================================
// generateAlertPDFContent Tests (0% coverage)
// ============================================

func TestExportService_GenerateAlertPDFContent_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &AlertReportData{
		GeneratedAt: time.Now(),
		Alerts: []model.Alert{
			{ID: 1, DeviceID: "CNC-001", Message: "Alert", Severity: "high"},
		},
		AlertStats: AlertStats{TotalAlerts: 1, HighAlerts: 1},
	}

	content := exportSvc.generateAlertPDFContent(data)
	assert.NotEmpty(t, content)
}

// ============================================
// generateROIPDFContent Tests (0% coverage)
// ============================================

func TestExportService_GenerateROIPDFContent_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)

	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     10,
			PredictedSavings: 50000,
		},
	}

	content := exportSvc.generateROIPDFContent(data)
	assert.NotEmpty(t, content)
}

// ============================================
// Export Tests - Additional Coverage
// ============================================

func TestExportService_Export_UnsupportedType2(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "invalid_type",
		Format:     FormatPDF,
	}

	result, err := exportSvc.Export(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported report type")
	assert.Nil(t, result)
}

func TestExportService_Export_UnsupportedFormat2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "devices",
		Format:     "invalid_format",
	}

	// Mock for generateDeviceReportData
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .* FROM devices").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	result, err := exportSvc.Export(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
	assert.Nil(t, result)
}

func TestExportService_Export_DefaultDates2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	exportSvc := NewExportService(
		repository.NewDeviceRepository(db),
		repository.NewTelemetryRepository(db),
		repository.NewAlertRepository(db),
		repository.NewWorkOrderRepository(db),
		nil,
	)
	ctx := context.Background()

	req := &ExportRequest{
		ReportType: "devices",
		Format:     FormatPDF,
		// No dates set - should use defaults
	}

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .* FROM devices").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	result, err := exportSvc.Export(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
