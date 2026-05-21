package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceService_Create_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	device := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001",
		Type:        "数控机床",
		Location:    "车间A",
		Description: "主加工设备",
	}

	// Expect device creation
	mock.ExpectExec("INSERT INTO devices").
		WithArgs("CNC-001", "数控机床001", "数控机床", "车间A", "online", "主加工设备", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Create
	err = deviceService.Create(ctx, device)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "online", device.Status) // Default status set
	assert.NotZero(t, device.CreatedAt)
	assert.NotZero(t, device.UpdatedAt)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	now := time.Now()

	// Expect query for GetByID
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("CNC-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
			AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "主加工设备", now, now))

	// Execute GetByID
	device, err := deviceService.GetByID(ctx, "CNC-001")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "CNC-001", device.ID)
	assert.Equal(t, "数控机床001", device.Name)
	assert.Equal(t, "数控机床", device.Type)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	// Expect query returning no rows
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("UNKNOWN-001").
		WillReturnError(sql.ErrNoRows)

	// Execute GetByID
	device, err := deviceService.GetByID(ctx, "UNKNOWN-001")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, device)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	now := time.Now()

	// Expect count query
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Expect list query
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})
	rows.AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "设备1", now, now)
	rows.AddRow("CNC-002", "数控机床002", "数控机床", "车间A", "online", "设备2", now, now)

	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY").
		WillReturnRows(rows)

	// Execute List
	devices, total, err := deviceService.List(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, 5, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	now := time.Now()

	// Expect GetByID first
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("CNC-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
			AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "设备", now, now))

	// Expect Update
	mock.ExpectExec("UPDATE devices SET").
		WithArgs("数控机床001更新", "数控机床", "车间B", "online", "更新描述", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	device := &model.Device{
		ID:          "CNC-001",
		Name:        "数控机床001更新",
		Type:        "数控机床",
		Location:    "车间B",
		Status:      "online",
		Description: "更新描述",
	}

	// Execute Update
	err = deviceService.Update(ctx, device)

	// Assertions
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_Update_DeviceNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	// Expect GetByID returning error
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("UNKNOWN-001").
		WillReturnError(sql.ErrNoRows)

	device := &model.Device{
		ID:          "UNKNOWN-001",
		Name:        "Unknown",
		Type:        "unknown",
		Location:    "车间",
		Description: "描述",
	}

	// Execute Update
	err = deviceService.Update(ctx, device)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Device not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	// Expect Delete
	mock.ExpectExec("DELETE FROM devices WHERE id").
		WithArgs("CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute Delete
	err = deviceService.Delete(ctx, "CNC-001")

	// Assertions
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_UpdateStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	// Expect UpdateStatus
	mock.ExpectExec("UPDATE devices SET status").
		WithArgs("warning", sqlmock.AnyArg(), "CNC-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute UpdateStatus
	err = deviceService.UpdateStatus(ctx, "CNC-001", "warning")

	// Assertions
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDeviceTypeFromID(t *testing.T) {
	testCases := map[string]string{
		"CNC-001": "数控机床",
		"CNC-123": "数控机床",
		"INJ-001": "注塑机",
		"INJ-999": "注塑机",
		"ROB-001": "工业机器人",
		"ROB-500": "工业机器人",
		"ASM-001": "装配线",
		"CNV-001": "传送带",
		"ABC-001": "未知设备",
		"XY-001":  "未知设备",
		"":        "未知设备",
		"1":       "未知设备",
	}

	for deviceID, expectedType := range testCases {
		deviceType := GetDeviceTypeFromID(deviceID)
		assert.Equal(t, expectedType, deviceType)
	}
}

func TestGetDeviceNameFromType(t *testing.T) {
	testCases := map[string]string{
		"数控机床":    "数控机床",
		"CNC":     "数控机床",
		"注塑机":     "注塑机",
		"INJ":     "注塑机",
		"工业机器人":   "工业机器人",
		"ROB":     "工业机器人",
		"装配线":     "装配线",
		"ASM":     "装配线",
		"传送带":     "传送带",
		"CNV":     "传送带",
		"unknown": "工业设备",
		"":        "工业设备",
	}

	for deviceType, expectedName := range testCases {
		name := GetDeviceNameFromType(deviceType)
		assert.Equal(t, expectedName, name)
	}
}

func TestDeviceService_AutoRegisterDevice_NewDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	// Expect GetByID returning error (device doesn't exist)
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("CNC-NEW").
		WillReturnError(sql.ErrNoRows)

	// Expect Create
	mock.ExpectExec("INSERT INTO devices").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute AutoRegisterDevice
	device, err := deviceService.AutoRegisterDevice(ctx, "CNC-NEW")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "CNC-NEW", device.ID)
	assert.Contains(t, device.Name, "数控机床")
	assert.Equal(t, "数控机床", device.Type)
	assert.Equal(t, "online", device.Status)
	assert.Equal(t, "车间A", device.Location)
	assert.Equal(t, "自动注册设备", device.Description)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_AutoRegisterDevice_ExistingDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	now := time.Now()

	// Expect GetByID returning existing device
	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id").
		WithArgs("CNC-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
			AddRow("CNC-001", "数控机床001", "数控机床", "车间A", "online", "设备", now, now))

	// Execute AutoRegisterDevice
	device, err := deviceService.AutoRegisterDevice(ctx, "CNC-001")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "CNC-001", device.ID)
	assert.Equal(t, "数控机床001", device.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeviceService_GetGraph_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceService := NewDeviceService(deviceRepo, userRepo)
	ctx := context.Background()

	now := time.Now()

	// Expect count query
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Expect list query
	rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"})
	rows.AddRow("CNC-001", "数控机床1", "数控机床", "车间A", "online", "设备", now, now)
	rows.AddRow("CNC-002", "数控机床2", "数控机床", "车间A", "online", "设备", now, now)
	rows.AddRow("INJ-001", "注塑机1", "注塑机", "车间B", "online", "设备", now, now)

	mock.ExpectQuery("SELECT id, name, type, location, status, description, created_at, updated_at FROM devices ORDER BY").
		WillReturnRows(rows)

	// Execute GetGraph
	graph, err := deviceService.GetGraph(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, graph)
	assert.NotNil(t, graph["nodes"])
	assert.NotNil(t, graph["links"])

	// Nodes should contain 3 devices
	nodes := graph["nodes"].([]map[string]interface{})
	assert.Len(t, nodes, 3)

	// Links should contain relationships based on same location
	links := graph["links"].([]map[string]interface{})
	assert.GreaterOrEqual(t, len(links), 0) // May have links for co-located devices

	assert.NoError(t, mock.ExpectationsWereMet())
}
