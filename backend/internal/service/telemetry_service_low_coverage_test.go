package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/industrial-ai/platform/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// NewTelemetryService Tests (0% coverage)
// ============================================

func TestNewTelemetryService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	assert.NotNil(t, svc)
}

// ============================================
// InitTelemetryService Tests (0% coverage)
// ============================================

func TestInitTelemetryService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))

	// Create alert service
	ruleRepo := repository.NewRuleRepository(database.NewDBWrapper(db))
	notificationRepo := repository.NewNotificationRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))
	blackBoxRepo := repository.NewBlackBoxRepository(database.NewDBWrapper(db))
	alertSvc := NewAlertService(ruleRepo, alertRepo, notificationRepo, workOrderRepo, blackBoxRepo, telemetryRepo, deviceRepo, AlertServiceConfig{})

	svc := InitTelemetryService(alertSvc, telemetryRepo, deviceRepo)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.telemetryRepo)
	assert.NotNil(t, svc.deviceRepo)
	assert.NotNil(t, svc.alertSvc)
}

// ============================================
// GetROIStats Tests (Additional Coverage)
// ============================================

func TestTelemetryService_GetROIStats_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("db error"))

	stats, err := svc.GetROIStats(ctx)
	assert.Error(t, err)
	assert.Nil(t, stats)
}

// ============================================
// GetSystemStatus Tests (Additional Coverage)
// ============================================

func TestTelemetryService_GetSystemStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	ctx := context.Background()

	// GetSystemStatus calls deviceRepo.Count twice:
	// 1. First for DB ping
	// 2. Second for device count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	status, err := svc.GetSystemStatus(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Database)
	assert.Equal(t, 10, status.DeviceCount)
	assert.NotZero(t, status.Timestamp)
}

func TestTelemetryService_GetSystemStatus_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("db error"))

	status, err := svc.GetSystemStatus(ctx)
	assert.NoError(t, err) // Should handle error gracefully
	assert.Equal(t, "unhealthy", status.Database)
}

// ============================================
// Ingest Tests - Additional Coverage
// ============================================

func TestTelemetryService_Ingest_WithAlertService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	// Create alert service
	ruleRepo := repository.NewRuleRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	notificationRepo := repository.NewNotificationRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))
	blackBoxRepo := repository.NewBlackBoxRepository(database.NewDBWrapper(db))
	alertSvc := NewAlertService(ruleRepo, alertRepo, notificationRepo, workOrderRepo, blackBoxRepo, telemetryRepo, deviceRepo, AlertServiceConfig{})

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, alertSvc)
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 75.0,
		Vibration:   2.0,
		Timestamp:   time.Now(),
		Status:      "normal",
	}

	// Mock telemetry insert
	mock.ExpectQuery("INSERT INTO device_telemetry").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock device status update
	mock.ExpectExec("UPDATE devices SET status").WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock alert evaluation
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC", "cnc", "Line 1", "online", "t1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices WHERE id").WillReturnRows(deviceRows)
	mock.ExpectQuery("SELECT .* FROM alert_rules WHERE enabled").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	err = svc.Ingest(ctx, data)
	assert.NoError(t, err)
}

func TestTelemetryService_Ingest_StatusWarning(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 85.0, // Between HighTemperatureThreshold(80) and CriticalTemperatureThreshold(100), triggers warning
		Vibration:   2.0,
		Timestamp:   time.Now(),
	}

	mock.ExpectQuery("INSERT INTO device_telemetry").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("UPDATE devices SET status").WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.Ingest(ctx, data)
	assert.NoError(t, err)
	assert.Equal(t, "warning", data.Status)
}

func TestTelemetryService_Ingest_StatusFault(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	svc := NewTelemetryService(telemetryRepo, deviceRepo, alertRepo, workOrderRepo, nil)
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 125.0, // Fault level
		Vibration:   6.0,   // Also fault level
		Timestamp:   time.Now(),
	}

	mock.ExpectQuery("INSERT INTO device_telemetry").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("UPDATE devices SET status").WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.Ingest(ctx, data)
	assert.NoError(t, err)
	assert.Equal(t, "fault", data.Status)
}
