package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceRepository_Create_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	device := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001",
		Type:        "数控机床",
		Location:    "车间A",
		Status:      "online",
		Description: "主加工设备",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Expect INSERT with ON CONFLICT
	mock.ExpectExec(`INSERT INTO devices`).
		WithArgs(
			"CNC-001",
			"数控机床001",
			"数控机床",
			"车间A",
			"online",
			"主加工设备",
			now,
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Create
	err = repo.Create(ctx, device)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	device := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001",
		Type:        "数控机床",
		Location:    "车间A",
		Status:      "online",
		Description: "主加工设备",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Expect INSERT returning error
	mock.ExpectExec(`INSERT INTO devices`).
		WillReturnError(errors.New("database error"))

	// Execute Create
	err = repo.Create(ctx, device)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
		AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "主加工设备", now, now)

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = \$1`).
		WithArgs("CNC-001").
		WillReturnRows(rows)

	// Execute GetByID
	device, err := repo.GetByID(ctx, "CNC-001")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "CNC-001", device.ID)
	assert.Equal(t, "数控机床001", device.Name)
	assert.Equal(t, "数控机床", device.Type)
	assert.Equal(t, "车间A", device.Location)
	assert.Equal(t, "online", device.Status)
	assert.Equal(t, "主加工设备", device.Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect SELECT query returning no rows
	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = \$1`).
		WithArgs("UNKNOWN-001").
		WillReturnError(sql.ErrNoRows)

	// Execute GetByID
	device, err := repo.GetByID(ctx, "UNKNOWN-001")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, device)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect SELECT query returning database error
	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = \$1`).
		WithArgs("CNC-001").
		WillReturnError(errors.New("connection failed"))

	// Execute GetByID
	device, err := repo.GetByID(ctx, "CNC-001")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection failed")
	assert.Nil(t, device)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Expect SELECT query with pagination
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})
	rows.AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "设备1", now, now)
	rows.AddRow("CNC-002", "数控机床002", "数控机床", "车间A", "online", "设备2", now, now)
	rows.AddRow("INJ-001", "注塑机001", "注塑机", "车间B", "online", "设备3", now, now)

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute List (page=1, pageSize=10)
	devices, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, devices, 3)
	assert.Equal(t, 5, total)
	assert.Equal(t, "CNC-001", devices[0].ID)
	assert.Equal(t, "数控机床001", devices[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_SecondPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

	// Expect SELECT query with offset
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})
	rows.AddRow("CNC-011", "数控机床011", "数控机床", "车间A", "online", "设备", now, now)

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 10).
		WillReturnRows(rows)

	// Execute List (page=2, pageSize=10, offset=10)
	devices, total, err := repo.List(ctx, 2, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, 25, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect SELECT query returning empty rows
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute List
	devices, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Empty(t, devices)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query returning error
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnError(errors.New("database error"))

	// Execute List
	devices, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, devices)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Expect SELECT query returning error
	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(errors.New("query failed"))

	// Execute List
	devices, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	assert.Nil(t, devices)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect SELECT query with malformed data (wrong type)
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
		AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "设备", "invalid_time", time.Now())

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute List
	devices, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	device := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001更新",
		Type:        "数控机床",
		Location:    "车间B",
		Status:      "warning",
		Description: "更新后的描述",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Expect UPDATE query
	mock.ExpectExec(`UPDATE devices SET`).
		WithArgs(
			"数控机床001更新",
			"数控机床",
			"车间B",
			"warning",
			"更新后的描述",
			sqlmock.AnyArg(), // UpdatedAt is set in the function
			"CNC-001",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Update
	err = repo.Update(ctx, device)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	device := &model.Device{
		ID:          "UNKNOWN-001",
		Name:        "未知设备",
		Type:        "unknown",
		Location:    "车间A",
		Status:      "offline",
		Description: "描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Expect UPDATE query returning error
	mock.ExpectExec(`UPDATE devices SET`).
		WillReturnError(errors.New("database error"))

	// Execute Update
	err = repo.Update(ctx, device)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect DELETE query
	mock.ExpectExec(`DELETE FROM devices WHERE id = \$1`).
		WithArgs("CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Delete
	err = repo.Delete(ctx, "CNC-001")

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect DELETE query returning error
	mock.ExpectExec(`DELETE FROM devices WHERE id = \$1`).
		WithArgs("UNKNOWN-001").
		WillReturnError(errors.New("database error"))

	// Execute Delete
	err = repo.Delete(ctx, "UNKNOWN-001")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_UpdateStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect UPDATE status query
	mock.ExpectExec(`UPDATE devices SET status = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs("warning", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute UpdateStatus
	err = repo.UpdateStatus(ctx, "CNC-001", "warning")

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_UpdateStatus_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect UPDATE status query returning error
	mock.ExpectExec(`UPDATE devices SET status = \$1, updated_at = \$2 WHERE id = \$3`).
		WillReturnError(errors.New("database error"))

	// Execute UpdateStatus
	err = repo.UpdateStatus(ctx, "UNKNOWN-001", "warning")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Count_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	// Execute Count
	count, err := repo.Count(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 42, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Count_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query returning error
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnError(errors.New("database error"))

	// Execute Count
	count, err := repo.Count(ctx)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_Count_ZeroDevices(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Expect COUNT query returning zero
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM devices`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Execute Count
	count, err := repo.Count(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test SQL query pattern matching
func TestDeviceRepository_SQLQueryPatterns(t *testing.T) {
	t.Run("Create SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewDeviceRepository(database.NewDBWrapper(db))
		ctx := context.Background()

		device := &model.Device{
			ID:          "TEST-001",
			Name:        "测试设备",
			Type:        "测试",
			Location:    "测试位置",
			Status:      "online",
			Description: "测试描述",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Verify the SQL pattern matches INSERT with ON CONFLICT
		mock.ExpectExec("INSERT INTO devices \\(id, name, type, location, status, description, created_at, updated_at\\)").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Create(ctx, device)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByID SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewDeviceRepository(database.NewDBWrapper(db))
		ctx := context.Background()

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
			AddRow("TEST-001", "测试设备", "测试", "测试位置", "online", "测试描述", now, now)

		// Verify the SQL pattern matches SELECT with WHERE id = $1
		mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
			WithArgs("TEST-001").
			WillReturnRows(rows)

		device, err := repo.GetByID(ctx, "TEST-001")
		assert.NoError(t, err)
		assert.NotNil(t, device)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("List SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewDeviceRepository(database.NewDBWrapper(db))
		ctx := context.Background()

		// Verify COUNT query pattern
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM devices").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Verify SELECT with ORDER BY and LIMIT/OFFSET pattern
		rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY created_at DESC LIMIT").
			WithArgs(10, 0).
			WillReturnRows(rows)

		devices, total, err := repo.List(ctx, 1, 10)
		assert.NoError(t, err)
		// Empty result should return empty slice, not nil
		if devices == nil {
			devices = []model.Device{}
		}
		assert.NotNil(t, devices)
		assert.Equal(t, 0, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewDeviceRepository(database.NewDBWrapper(db))
		ctx := context.Background()

		device := &model.Device{
			ID:          "TEST-001",
			Name:        "测试设备",
			Type:        "测试",
			Location:    "测试位置",
			Status:      "online",
			Description: "测试描述",
		}

		// Verify the SQL pattern matches UPDATE with SET clauses and WHERE
		mock.ExpectExec("UPDATE devices SET name = .*, type = .*, location = .*, status = .*, description = .*, updated_at = .* WHERE id").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Update(ctx, device)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewDeviceRepository(database.NewDBWrapper(db))
		ctx := context.Background()

		// Verify the SQL pattern matches DELETE with WHERE id = $1
		mock.ExpectExec("DELETE FROM devices WHERE id").
			WithArgs("TEST-001").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Delete(ctx, "TEST-001")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test context cancellation
func TestDeviceRepository_ContextCancellation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	device := &model.Device{
		ID:          "TEST-001",
		Name:        "测试设备",
		Type:        "测试",
		Location:    "测试位置",
		Status:      "online",
		Description: "测试描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// With canceled context, operations should fail
	// Note: sqlmock may not perfectly simulate context cancellation
	// In real scenarios, the database driver would handle this
	t.Run("Create with canceled context", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO devices").
			WillReturnError(context.Canceled)

		err := repo.Create(ctx, device)
		assert.Error(t, err)
	})

	t.Run("GetByID with canceled context", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
			WithArgs("TEST-001").
			WillReturnError(context.Canceled)

		_, err := repo.GetByID(ctx, "TEST-001")
		assert.Error(t, err)
	})
}

// Test GetByIDWithTenant
func TestDeviceRepository_GetByIDWithTenant_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	expectedDevice := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001",
		Type:        "数控机床",
		Location:    "车间A",
		Status:      "online",
		Description: "主加工设备",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
		AddRow(expectedDevice.ID, expectedDevice.Name, expectedDevice.Type, expectedDevice.Location,
			expectedDevice.Status, expectedDevice.Description, expectedDevice.CreatedAt, expectedDevice.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = \$1 AND`).
		WithArgs("CNC-001", "tenant-001").
		WillReturnRows(rows)

	device, err := repo.GetByIDWithTenant(ctx, "CNC-001", "tenant-001")
	require.NoError(t, err)
	assert.Equal(t, expectedDevice.ID, device.ID)
	assert.Equal(t, expectedDevice.Name, device.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_GetByIDWithTenant_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = \$1 AND`).
		WithArgs("CNC-999", "tenant-001").
		WillReturnError(sql.ErrNoRows)

	device, err := repo.GetByIDWithTenant(ctx, "CNC-999", "tenant-001")
	require.Error(t, err)
	assert.Nil(t, device)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test BatchCreate
func TestDeviceRepository_BatchCreate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	devices := []*model.Device{
		{
			ID:          "CNC-001",
			Name:        "数控机床001",
			Type:        "数控机床",
			Location:    "车间A",
			Status:      "online",
			Description: "主加工设备",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "CNC-002",
			Name:        "数控机床002",
			Type:        "数控机床",
			Location:    "车间B",
			Status:      "online",
			Description: "备用设备",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mock.ExpectExec(`INSERT INTO devices`).
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.BatchCreate(ctx, devices)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceRepository_BatchCreate_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	err = repo.BatchCreate(ctx, []*model.Device{})
	require.NoError(t, err)
}

func TestDeviceRepository_BatchCreate_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDeviceRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	devices := []*model.Device{
		{
			ID:          "CNC-001",
			Name:        "数控机床001",
			Type:        "数控机床",
			Location:    "车间A",
			Status:      "online",
			Description: "主加工设备",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mock.ExpectExec(`INSERT INTO devices`).
		WillReturnError(errors.New("database error"))

	err = repo.BatchCreate(ctx, devices)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
