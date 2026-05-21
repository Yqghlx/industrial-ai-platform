package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TelemetryRepository Tests

func TestTelemetryRepository_Insert_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Timestamp:   now,
		Temperature: 25.5,
		Pressure:    101.3,
		Vibration:   0.02,
		Humidity:    45.0,
		Power:       1500.0,
		Status:      "normal",
		Message:     "正常运行",
	}

	// Expect INSERT with RETURNING id
	mock.ExpectQuery(`INSERT INTO device_telemetry`).
		WithArgs(
			"CNC-001",
			now,
			25.5,
			101.3,
			0.02,
			45.0,
			1500.0,
			"normal",
			"正常运行",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Execute Insert
	err = repo.Insert(ctx, data)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(1), data.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_Insert_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	data := &model.TelemetryData{
		DeviceID:    "CNC-001",
		Timestamp:   time.Now(),
		Temperature: 25.5,
		Pressure:    101.3,
		Vibration:   0.02,
		Humidity:    45.0,
		Power:       1500.0,
		Status:      "normal",
		Message:     "正常运行",
	}

	// Expect INSERT returning error
	mock.ExpectQuery(`INSERT INTO device_telemetry`).
		WillReturnError(errors.New("database error"))

	// Execute Insert
	err = repo.Insert(ctx, data)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetByDeviceID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT query with time range
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	rows.AddRow(1, "CNC-001", now, 25.5, 101.3, 0.02, 45.0, 1500.0, "normal", "正常运行")
	rows.AddRow(2, "CNC-001", now.Add(-1*time.Hour), 26.0, 102.0, 0.03, 46.0, 1550.0, "normal", "")
	rows.AddRow(3, "CNC-001", now.Add(-2*time.Hour), 24.5, 100.5, 0.01, 44.0, 1450.0, "warning", "温度略高")

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id = .* AND time >= .* AND time <= .* ORDER BY time DESC LIMIT .*`).
		WithArgs("CNC-001", start, end, 100).
		WillReturnRows(rows)

	// Execute GetByDeviceID
	data, err := repo.GetByDeviceID(ctx, "CNC-001", start, end, 100)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, data, 3)
	assert.Equal(t, "CNC-001", data[0].DeviceID)
	assert.Equal(t, 25.5, data[0].Temperature)
	assert.Equal(t, "normal", data[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetByDeviceID_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT query returning empty rows
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id = .* AND time >= .* AND time <= .* ORDER BY time DESC LIMIT .*`).
		WithArgs("UNKNOWN-001", start, end, 100).
		WillReturnRows(rows)

	// Execute GetByDeviceID
	data, err := repo.GetByDeviceID(ctx, "UNKNOWN-001", start, end, 100)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0) // Should return empty slice, not nil
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetByDeviceID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT query returning error
	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id`).
		WillReturnError(errors.New("database error"))

	// Execute GetByDeviceID
	data, err := repo.GetByDeviceID(ctx, "CNC-001", start, end, 100)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, data)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetByDeviceID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT query with malformed data
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"}).
		AddRow(1, "CNC-001", "invalid_time", 25.5, 101.3, 0.02, 45.0, 1500.0, "normal", "test")

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id`).
		WillReturnRows(rows)

	// Execute GetByDeviceID
	data, err := repo.GetByDeviceID(ctx, "CNC-001", start, end, 100)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetLatest_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Expect SELECT query for latest telemetry
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	rows.AddRow(1, "CNC-001", now, 25.5, 101.3, 0.02, 45.0, 1500.0, "normal", "")
	rows.AddRow(2, "INJ-001", now, 30.0, 105.0, 0.05, 50.0, 2000.0, "normal", "")
	rows.AddRow(3, "ASM-001", now, 22.0, 100.0, 0.01, 40.0, 1000.0, "warning", "振动异常")

	mock.ExpectQuery(`SELECT DISTINCT ON \(device_id\) .* FROM device_telemetry ORDER BY device_id, time DESC`).
		WillReturnRows(rows)

	// Execute GetLatest
	data, err := repo.GetLatest(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, data, 3)
	assert.Equal(t, "CNC-001", data[0].DeviceID)
	assert.Equal(t, "INJ-001", data[1].DeviceID)
	assert.Equal(t, "ASM-001", data[2].DeviceID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetLatest_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning empty rows
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})

	mock.ExpectQuery(`SELECT DISTINCT ON \(device_id\) .* FROM device_telemetry ORDER BY device_id, time DESC`).
		WillReturnRows(rows)

	// Execute GetLatest
	data, err := repo.GetLatest(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetLatest_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning error
	mock.ExpectQuery(`SELECT DISTINCT ON \(device_id\) .* FROM device_telemetry ORDER BY device_id, time DESC`).
		WillReturnError(errors.New("database error"))

	// Execute GetLatest
	data, err := repo.GetLatest(ctx)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, data)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetStats_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT stats query with aggregation
	rows := sqlmock.NewRows([]string{"avg_temp", "avg_pressure", "avg_vibration", "max_temp", "max_pressure", "max_vibration", "count"}).
		AddRow(25.5, 101.3, 0.02, 30.0, 105.0, 0.05, 100)

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id = .* AND time >= .* AND time <= .*`).
		WithArgs("CNC-001", start, end).
		WillReturnRows(rows)

	// Execute GetStats
	stats, err := repo.GetStats(ctx, "CNC-001", start, end)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "CNC-001", stats.DeviceID)
	assert.Equal(t, 25.5, stats.AvgTemperature)
	assert.Equal(t, 101.3, stats.AvgPressure)
	assert.Equal(t, 0.02, stats.AvgVibration)
	assert.Equal(t, 30.0, stats.MaxTemperature)
	assert.Equal(t, 105.0, stats.MaxPressure)
	assert.Equal(t, 0.05, stats.MaxVibration)
	assert.Equal(t, int64(100), stats.DataPoints)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetStats_NoData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT stats query with no data (COALESCE returns 0)
	rows := sqlmock.NewRows([]string{"avg_temp", "avg_pressure", "avg_vibration", "max_temp", "max_pressure", "max_vibration", "count"}).
		AddRow(0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0)

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id = .* AND time >= .* AND time <= .*`).
		WithArgs("UNKNOWN-001", start, end).
		WillReturnRows(rows)

	// Execute GetStats
	stats, err := repo.GetStats(ctx, "UNKNOWN-001", start, end)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "UNKNOWN-001", stats.DeviceID)
	assert.Equal(t, 0.0, stats.AvgTemperature)
	assert.Equal(t, int64(0), stats.DataPoints)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_GetStats_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT stats query returning error
	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id`).
		WillReturnError(errors.New("database error"))

	// Execute GetStats
	stats, err := repo.GetStats(ctx, "CNC-001", start, end)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test NullFloat64 handling in GetByDeviceID
func TestTelemetryRepository_GetByDeviceID_NullFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now

	// Expect SELECT query with NULL values for optional fields
	rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})
	// Add row with NULL values (represented as nil in sqlmock)
	rows.AddRow(1, "CNC-001", now, nil, nil, nil, nil, nil, "offline", nil)

	mock.ExpectQuery(`SELECT .* FROM device_telemetry WHERE device_id`).
		WithArgs("CNC-001", start, end, 100).
		WillReturnRows(rows)

	// Execute GetByDeviceID
	data, err := repo.GetByDeviceID(ctx, "CNC-001", start, end, 100)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, "CNC-001", data[0].DeviceID)
	assert.Equal(t, "offline", data[0].Status)
	// Null values should be converted to 0.0 or empty string
	assert.Equal(t, 0.0, data[0].Temperature)
	assert.Equal(t, 0.0, data[0].Pressure)
	assert.Equal(t, "", data[0].Message)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// SQL Pattern Tests for TelemetryRepository

func TestTelemetryRepository_SQLQueryPatterns(t *testing.T) {
	t.Run("Insert SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewTelemetryRepository(db)
		ctx := context.Background()

		data := &model.TelemetryData{
			DeviceID:    "TEST-001",
			Timestamp:   time.Now(),
			Temperature: 25.0,
			Pressure:    100.0,
			Vibration:   0.01,
			Humidity:    50.0,
			Power:       1000.0,
			Status:      "normal",
			Message:     "",
		}

		mock.ExpectQuery("INSERT INTO device_telemetry").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err = repo.Insert(ctx, data)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByDeviceID SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewTelemetryRepository(db)
		ctx := context.Background()

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"}).
			AddRow(1, "TEST-001", now, 25.0, 100.0, 0.01, 50.0, 1000.0, "normal", "")

		mock.ExpectQuery("SELECT .* FROM device_telemetry WHERE device_id = .* AND time >= .* AND time <= .* ORDER BY time DESC LIMIT").
			WillReturnRows(rows)

		data, err := repo.GetByDeviceID(ctx, "TEST-001", now.Add(-1*time.Hour), now, 10)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetLatest SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewTelemetryRepository(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"})

		mock.ExpectQuery(`SELECT DISTINCT ON \(device_id\) .* FROM device_telemetry ORDER BY device_id, time DESC`).
			WillReturnRows(rows)

		data, err := repo.GetLatest(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetStats SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewTelemetryRepository(db)
		ctx := context.Background()

		now := time.Now()
		rows := sqlmock.NewRows([]string{"avg_temp", "avg_pressure", "avg_vibration", "max_temp", "max_pressure", "max_vibration", "count"}).
			AddRow(0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0)

		mock.ExpectQuery(`SELECT COALESCE\(AVG\(temperature\), 0\).* FROM device_telemetry WHERE device_id`).
			WillReturnRows(rows)

		stats, err := repo.GetStats(ctx, "TEST-001", now.Add(-1*time.Hour), now)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Context cancellation tests

func TestTelemetryRepository_ContextCancellation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	data := &model.TelemetryData{
		DeviceID:    "TEST-001",
		Timestamp:   time.Now(),
		Temperature: 25.0,
		Status:      "normal",
	}

	t.Run("Insert with canceled context", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO device_telemetry").
			WillReturnError(context.Canceled)

		err := repo.Insert(ctx, data)
		assert.Error(t, err)
	})

	t.Run("GetByDeviceID with canceled context", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM device_telemetry").
			WillReturnError(context.Canceled)

		now := time.Now()
		_, err := repo.GetByDeviceID(ctx, "TEST-001", now.Add(-1*time.Hour), now, 10)
		assert.Error(t, err)
	})

	t.Run("GetLatest with canceled context", func(t *testing.T) {
		mock.ExpectQuery(`SELECT DISTINCT ON \(device_id\) .* FROM device_telemetry ORDER BY device_id, time DESC`).
			WillReturnError(context.Canceled)

		_, err := repo.GetLatest(ctx)
		assert.Error(t, err)
	})

	t.Run("GetStats with canceled context", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM device_telemetry WHERE device_id").
			WillReturnError(context.Canceled)

		now := time.Now()
		_, err := repo.GetStats(ctx, "TEST-001", now.Add(-1*time.Hour), now)
		assert.Error(t, err)
	})
}
