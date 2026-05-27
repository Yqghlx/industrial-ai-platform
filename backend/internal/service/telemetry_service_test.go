package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/industrial-ai/platform/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryService_Ingest_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	// For telemetry service, we need alert service too, but we'll skip alert evaluation for simplicity
	// In a real scenario, you'd mock the alert service as well

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil, // Skip alert evaluation for this test
	}
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 75.0,
		Vibration:   1.5,
		Pressure:    100.0,
		Humidity:    50.0,
		Power:       200.0,
	}

	// Expect telemetry insert
	mock.ExpectQuery("INSERT INTO device_telemetry").
		WithArgs("CNC-001", sqlmock.AnyArg(), 75.0, 100.0, 1.5, 50.0, 200.0, "normal", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect device status update
	mock.ExpectExec("UPDATE devices SET status").
		WithArgs("online", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Ingest
	err = telemetryService.Ingest(ctx, data)

	// Assertions
	assert.NoError(t, err)
	assert.NotZero(t, data.Timestamp)
	assert.Equal(t, "normal", data.Status)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_Ingest_WarningStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 85.0, // Between HighTemperatureThreshold(80) and CriticalTemperatureThreshold(100), triggers warning
		Vibration:   1.5,
	}

	// Expect telemetry insert with warning status
	mock.ExpectQuery("INSERT INTO device_telemetry").
		WithArgs("CNC-001", sqlmock.AnyArg(), 85.0, 0.0, 1.5, 0.0, 0.0, "warning", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect device status update to warning
	mock.ExpectExec("UPDATE devices SET status").
		WithArgs("warning", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Ingest
	err = telemetryService.Ingest(ctx, data)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "warning", data.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_Ingest_FaultStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 125.0, // Above 120, triggers fault
		Vibration:   6.0,   // Above 5.0, triggers fault
	}

	// Expect telemetry insert with fault status
	mock.ExpectQuery("INSERT INTO device_telemetry").
		WithArgs("CNC-001", sqlmock.AnyArg(), 125.0, 0.0, 6.0, 0.0, 0.0, "fault", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect device status update to fault
	mock.ExpectExec("UPDATE devices SET status").
		WithArgs("fault", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Ingest
	err = telemetryService.Ingest(ctx, data)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "fault", data.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_Ingest_VibrationWarning(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Temperature: 75.0,
		Vibration:   4.0, // Above 3.0, triggers warning
	}

	// Expect telemetry insert with warning status
	mock.ExpectQuery("INSERT INTO device_telemetry").
		WithArgs("CNC-001", sqlmock.AnyArg(), 75.0, 0.0, 4.0, 0.0, 0.0, "warning", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect device status update to warning
	mock.ExpectExec("UPDATE devices SET status").
		WithArgs("warning", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Ingest
	err = telemetryService.Ingest(ctx, data)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "warning", data.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetByDeviceID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	// Expect query for telemetry data
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	rows.AddRow(1, "CNC-001", now, 75.0, 100.0, 1.5, 50.0, 200.0, "normal", "")
	rows.AddRow(2, "CNC-001", now.Add(-10*time.Minute), 74.0, 99.0, 1.4, 49.0, 195.0, "normal", "")

	mock.ExpectQuery("SELECT id, device_id, time, temperature, pressure, vibration, humidity, power, status, message FROM device_telemetry").
		WillReturnRows(rows)

	// Execute GetByDeviceID
	data, err := telemetryService.GetByDeviceID(ctx, "CNC-001", start, end, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetByDeviceID_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	// Expect query with default limit (1000)
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	mock.ExpectQuery("SELECT id, device_id, time, temperature, pressure, vibration, humidity, power, status, message FROM device_telemetry").
		WillReturnRows(rows)

	// Execute GetByDeviceID with limit 0 (should use default 1000)
	data, err := telemetryService.GetByDeviceID(ctx, "CNC-001", start, end, 0)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetLatest_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	now := time.Now()

	// Expect query for latest telemetry
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	rows.AddRow(1, "CNC-001", now, 75.0, 100.0, 1.5, 50.0, 200.0, "normal", "")
	rows.AddRow(2, "INJ-002", now, 80.0, 110.0, 2.0, 55.0, 250.0, "normal", "")

	mock.ExpectQuery("SELECT DISTINCT ON").
		WillReturnRows(rows)

	// Execute GetLatest
	data, err := telemetryService.GetLatest(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetStats_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	// Expect query for stats
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"avg_temp", "avg_pressure", "avg_vibration", "max_temp", "max_pressure", "max_vibration", "count"}).
			AddRow(75.0, 100.0, 1.5, 80.0, 110.0, 2.0, 100))

	// Execute GetStats
	stats, err := telemetryService.GetStats(ctx, "CNC-001", start, end)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "CNC-001", stats.DeviceID)
	assert.Equal(t, 75.0, stats.AvgTemperature)
	assert.Equal(t, 100.0, stats.AvgPressure)
	assert.Equal(t, 1.5, stats.AvgVibration)
	assert.Equal(t, 80.0, stats.MaxTemperature)
	assert.Equal(t, int64(100), stats.DataPoints)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetROIStats_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	alertRepo := repository.NewAlertRepository(database.NewDBWrapper(db))
	workOrderRepo := repository.NewWorkOrderRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertRepo:     alertRepo,
		workOrderRepo: workOrderRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	// Expect device count query
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	// Expect active alerts count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	// Expect open work orders count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	// Expect resolved alerts count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Execute GetROIStats
	stats, err := telemetryService.GetROIStats(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 10, stats.TotalDevices)
	assert.Equal(t, 2, stats.ActiveAlerts)
	assert.Equal(t, 3, stats.OpenWorkOrders)
	assert.Equal(t, 5, stats.ResolvedIssues)
	// Base savings = 10 * 1000 = 10000
	// Resolved savings = 5 * 500 = 2500
	// Alert cost = 2 * 100 = 200
	// Total = 10000 + 2500 - 200 = 12300
	assert.Equal(t, 12300.0, stats.PredictedSavings)
	// Uptime = 100 - (2/10 * 10) = 98.0
	assert.Equal(t, 98.0, stats.UptimePercentage)
	// Avg response time = 1.5 + (2/(5+1) * 2) = 1.5 + 0.67 = 2.17 (approximately)
	assert.InDelta(t, 2.17, stats.AvgResponseTime, 0.1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetSystemStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	// Expect device count query (twice - for DB ping and for count)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Execute GetSystemStatus
	status, err := telemetryService.GetSystemStatus(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Database)
	assert.Equal(t, "1.0.0", status.Version)
	assert.Equal(t, 5, status.DeviceCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetSystemStatus_UnhealthyDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	// Expect device count query returning error
	mock.ExpectQuery("SELECT COUNT").WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery("SELECT COUNT").WillReturnError(sql.ErrConnDone)

	// Execute GetSystemStatus
	status, err := telemetryService.GetSystemStatus(ctx)

	// Assertions
	assert.NoError(t, err) // Function handles error gracefully
	assert.NotNil(t, status)
	assert.Equal(t, "unhealthy", status.Database)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetHistoricalData_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	now := time.Now()

	// Expect query for historical data
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	rows.AddRow(1, "CNC-001", now, 75.0, 100.0, 1.5, 50.0, 200.0, "normal", "")

	mock.ExpectQuery("SELECT id, device_id, time, temperature, pressure, vibration, humidity, power, status, message FROM device_telemetry").
		WillReturnRows(rows)

	// Execute GetHistoricalData
	data, err := telemetryService.GetHistoricalData(ctx, "CNC-001", "1h", 10)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryService_GetHistoricalData_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))

	telemetryService := &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      nil,
	}
	ctx := context.Background()

	// Expect query with default limit (1000)
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	mock.ExpectQuery("SELECT id, device_id, time, temperature, pressure, vibration, humidity, power, status, message FROM device_telemetry").
		WillReturnRows(rows)

	// Execute GetHistoricalData with limit 0 (should use default)
	data, err := telemetryService.GetHistoricalData(ctx, "CNC-001", "1h", 0)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParseTimeRange(t *testing.T) {
	testCases := []struct {
		rangeStr         string
		expectedDuration time.Duration
	}{
		{"1h", 1 * time.Hour},
		{"6h", 6 * time.Hour},
		{"24h", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"30d", 30 * 24 * time.Hour},
		{"invalid", 1 * time.Hour}, // Default
		{"", 1 * time.Hour},        // Default
	}

	for _, tc := range testCases {
		start, end := ParseTimeRange(tc.rangeStr)
		expectedStart := end.Add(-tc.expectedDuration)

		// Allow for small time differences due to execution time
		assert.WithinDuration(t, expectedStart, start, 1*time.Second)
	}
}

func TestGetTimeRanges(t *testing.T) {
	ranges := GetTimeRanges()

	assert.Len(t, ranges, 5)

	expectedValues := []string{"1h", "6h", "24h", "7d", "30d"}
	for i, expected := range expectedValues {
		assert.Equal(t, expected, ranges[i]["value"])
	}
}

func TestFormatTimestamp(t *testing.T) {
	now := time.Now()
	formatted := FormatTimestamp(now)

	assert.NotEmpty(t, formatted)
	assert.Contains(t, formatted, "-") // Date separator
	assert.Contains(t, formatted, ":") // Time separator
	assert.Contains(t, formatted, " ") // Space between date and time
}

func TestValidateTelemetryData_Success(t *testing.T) {
	now := time.Now()
	data := &model.TelemetryData{
		DeviceID:  "CNC-001",
		Timestamp: now,
	}

	err := ValidateTelemetryData(data)
	assert.NoError(t, err)
}

func TestValidateTelemetryData_MissingDeviceID(t *testing.T) {
	data := &model.TelemetryData{
		Timestamp: time.Now(),
	}

	err := ValidateTelemetryData(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device_id is required")
}

func TestValidateTelemetryData_DefaultTimestamp(t *testing.T) {
	data := &model.TelemetryData{
		DeviceID: "CNC-001",
		// Timestamp not set
	}

	err := ValidateTelemetryData(data)
	assert.NoError(t, err)
	assert.NotZero(t, data.Timestamp)
}

func TestPaginationParams_Defaults(t *testing.T) {
	params := &model.PaginationParams{
		Page:     0,  // Should default to 1
		PageSize: 0,  // Should default to 20
		SortBy:   "", // Should default to "created_at"
		Order:    "", // Should default to "desc"
	}

	params.Defaults()

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PageSize)
	assert.Equal(t, "created_at", params.SortBy)
	assert.Equal(t, "desc", params.Order)
}

func TestPaginationParams_Limits(t *testing.T) {
	// PageSize > 100 should be limited to 20
	params := &model.PaginationParams{
		Page:     1,
		PageSize: 200,
	}

	params.Defaults()

	assert.Equal(t, 20, params.PageSize)
}

func TestPaginationParams_ValidValues(t *testing.T) {
	params := &model.PaginationParams{
		Page:     5,
		PageSize: 50,
		SortBy:   "name",
		Order:    "asc",
	}

	params.Defaults()

	// Valid values should not be changed
	assert.Equal(t, 5, params.Page)
	assert.Equal(t, 50, params.PageSize)
	assert.Equal(t, "name", params.SortBy)
	assert.Equal(t, "asc", params.Order)
}
