package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================
// GetDeviceContext Tests
// ============================================

func TestGetDeviceContext_WithDeviceID(t *testing.T) {
	// Setup mock database
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	deviceRepo := repository.NewDeviceRepository(dbWrapper)
	telemetryRepo := repository.NewTelemetryRepository(dbWrapper)
	taskLogRepo := repository.NewAgentTaskLogRepository(dbWrapper)

	// Create AgentService with mocks
	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device query
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
		AddRow(deviceID, "CNC Machine 001", "CNC", "Factory A", "running", "Test device", time.Now(), time.Now())
	mockDB.ExpectQuery("SELECT.*FROM devices WHERE id =").
		WithArgs(deviceID).
		WillReturnRows(deviceRows)

	// Mock telemetry query
	telemetryRows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"}).
		AddRow(1, deviceID, time.Now(), 75.0, 1.0, 1.2, 50.0, 100.0, "normal", "OK")
	mockDB.ExpectQuery("SELECT.*FROM device_telemetry").
		WithArgs(deviceID, sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
		WillReturnRows(telemetryRows)

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Contains(t, contextData, "device")
	assert.Contains(t, contextData, "telemetry")

	device := contextData["device"].(*model.Device)
	assert.Equal(t, deviceID, device.ID)
	assert.Equal(t, "CNC Machine 001", device.Name)

	telemetry := contextData["telemetry"].([]model.TelemetryData)
	assert.Len(t, telemetry, 1)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetDeviceContext_DeviceError(t *testing.T) {
	// Setup mock database
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	deviceRepo := repository.NewDeviceRepository(dbWrapper)
	telemetryRepo := repository.NewTelemetryRepository(dbWrapper)
	taskLogRepo := repository.NewAgentTaskLogRepository(dbWrapper)

	// Create AgentService with mocks
	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device query to return error
	mockDB.ExpectQuery("SELECT.*FROM devices WHERE id =").
		WithArgs(deviceID).
		WillReturnError(errors.New("device not found"))

	// Mock telemetry query (should still try)
	telemetryRows := sqlmock.NewRows([]string{"id", "device_id", "time", "temperature", "pressure", "vibration", "humidity", "power", "status", "message"}).
		AddRow(1, deviceID, time.Now(), 75.0, 1.0, 1.2, 50.0, 100.0, "normal", "OK")
	mockDB.ExpectQuery("SELECT.*FROM device_telemetry").
		WithArgs(deviceID, sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
		WillReturnRows(telemetryRows)

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should not return error, just skip device in context
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.NotContains(t, contextData, "device") // Device should not be in context due to error
	assert.Contains(t, contextData, "telemetry") // Telemetry should still be present

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetDeviceContext_TelemetryError(t *testing.T) {
	// Setup mock database
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	deviceRepo := repository.NewDeviceRepository(dbWrapper)
	telemetryRepo := repository.NewTelemetryRepository(dbWrapper)
	taskLogRepo := repository.NewAgentTaskLogRepository(dbWrapper)

	// Create AgentService with mocks
	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device query to return device
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
		AddRow(deviceID, "CNC Machine 001", "CNC", "Factory A", "running", "Test device", time.Now(), time.Now())
	mockDB.ExpectQuery("SELECT.*FROM devices WHERE id =").
		WithArgs(deviceID).
		WillReturnRows(deviceRows)

	// Mock telemetry query to return error
	mockDB.ExpectQuery("SELECT.*FROM device_telemetry").
		WithArgs(deviceID, sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
		WillReturnError(errors.New("no telemetry data"))

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should not return error, just skip telemetry in context
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Contains(t, contextData, "device")      // Device should be in context
	assert.NotContains(t, contextData, "telemetry") // Telemetry should not be in context

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetDeviceContext_EmptyDeviceID(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	deviceRepo := repository.NewDeviceRepository(dbWrapper)
	telemetryRepo := repository.NewTelemetryRepository(dbWrapper)
	taskLogRepo := repository.NewAgentTaskLogRepository(dbWrapper)

	// Create AgentService with mocks
	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	ctx := context.Background()
	deviceID := "" // Empty device ID

	// Execute GetDeviceContext with empty device ID
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should return empty context without calling repositories
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Empty(t, contextData) // Should be empty since deviceID is empty
}

func TestGetDeviceContext_BothErrors(t *testing.T) {
	// Setup mock database
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	deviceRepo := repository.NewDeviceRepository(dbWrapper)
	telemetryRepo := repository.NewTelemetryRepository(dbWrapper)
	taskLogRepo := repository.NewAgentTaskLogRepository(dbWrapper)

	// Create AgentService with mocks
	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock both queries to return errors
	mockDB.ExpectQuery("SELECT.*FROM devices WHERE id =").
		WithArgs(deviceID).
		WillReturnError(errors.New("device not found"))

	mockDB.ExpectQuery("SELECT.*FROM device_telemetry").
		WithArgs(deviceID, sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
		WillReturnError(errors.New("no telemetry"))

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should not return error, return empty context
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Empty(t, contextData) // Should be empty since both failed

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

// ============================================
// Alternative test using repository mocks
// ============================================

func TestGetDeviceContext_WithRepositoryMocks(t *testing.T) {
	// Use repository mocks directly
	mockDeviceRepo := new(MockDeviceRepoForLLMTest)
	mockTelemetryRepo := new(MockTelemetryRepoForLLMTest)
	mockTaskLogRepo := new(MockTaskLogRepoForLLMTest)

	// Create AgentService with mocks
	agentService := &AgentService{
		taskLogRepo:   mockTaskLogRepo,
		deviceRepo:    mockDeviceRepo,
		telemetryRepo: mockTelemetryRepo,
	}

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device repository to return device
	expectedDevice := &model.Device{
		ID:       deviceID,
		Name:     "CNC Machine 001",
		Type:     "CNC",
		Status:   "running",
		Location: "Factory A",
	}
	mockDeviceRepo.On("GetByID", ctx, deviceID).Return(expectedDevice, nil)

	// Mock telemetry repository to return telemetry data
	expectedTelemetry := []model.TelemetryData{
		{
			ID:          1,
			DeviceID:    deviceID,
			Timestamp:   time.Now(),
			Temperature: 75.0,
			Vibration:   1.2,
		},
	}
	mockTelemetryRepo.On("GetByDeviceID", ctx, deviceID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 10).Return(expectedTelemetry, nil)

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Contains(t, contextData, "device")
	assert.Contains(t, contextData, "telemetry")

	device := contextData["device"].(*model.Device)
	assert.Equal(t, deviceID, device.ID)

	telemetry := contextData["telemetry"].([]model.TelemetryData)
	assert.Len(t, telemetry, 1)

	mockDeviceRepo.AssertExpectations(t)
	mockTelemetryRepo.AssertExpectations(t)
}

func TestGetDeviceContext_WithRepositoryMocks_DeviceError(t *testing.T) {
	// Use repository mocks directly
	mockDeviceRepo := new(MockDeviceRepoForLLMTest)
	mockTelemetryRepo := new(MockTelemetryRepoForLLMTest)
	mockTaskLogRepo := new(MockTaskLogRepoForLLMTest)

	// Create AgentService with mocks
	agentService := &AgentService{
		taskLogRepo:   mockTaskLogRepo,
		deviceRepo:    mockDeviceRepo,
		telemetryRepo: mockTelemetryRepo,
	}

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device repository to return error
	mockDeviceRepo.On("GetByID", ctx, deviceID).Return(nil, errors.New("device not found"))

	// Mock telemetry repository to return telemetry data
	expectedTelemetry := []model.TelemetryData{
		{
			ID:          1,
			DeviceID:    deviceID,
			Timestamp:   time.Now(),
		},
	}
	mockTelemetryRepo.On("GetByDeviceID", ctx, deviceID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 10).Return(expectedTelemetry, nil)

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should not return error, just skip device in context
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.NotContains(t, contextData, "device") // Device should not be in context due to error
	assert.Contains(t, contextData, "telemetry")

	mockDeviceRepo.AssertExpectations(t)
	mockTelemetryRepo.AssertExpectations(t)
}

func TestGetDeviceContext_WithRepositoryMocks_TelemetryError(t *testing.T) {
	// Use repository mocks directly
	mockDeviceRepo := new(MockDeviceRepoForLLMTest)
	mockTelemetryRepo := new(MockTelemetryRepoForLLMTest)
	mockTaskLogRepo := new(MockTaskLogRepoForLLMTest)

	// Create AgentService with mocks
	agentService := &AgentService{
		taskLogRepo:   mockTaskLogRepo,
		deviceRepo:    mockDeviceRepo,
		telemetryRepo: mockTelemetryRepo,
	}

	ctx := context.Background()
	deviceID := "CNC-001"

	// Mock device repository to return device
	expectedDevice := &model.Device{
		ID:   deviceID,
		Name: "CNC Machine 001",
	}
	mockDeviceRepo.On("GetByID", ctx, deviceID).Return(expectedDevice, nil)

	// Mock telemetry repository to return error
	mockTelemetryRepo.On("GetByDeviceID", ctx, deviceID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 10).Return(nil, errors.New("no telemetry"))

	// Execute GetDeviceContext
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should not return error, just skip telemetry in context
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Contains(t, contextData, "device")
	assert.NotContains(t, contextData, "telemetry")

	mockDeviceRepo.AssertExpectations(t)
	mockTelemetryRepo.AssertExpectations(t)
}

func TestGetDeviceContext_WithRepositoryMocks_EmptyDeviceID(t *testing.T) {
	// Use repository mocks directly
	mockDeviceRepo := new(MockDeviceRepoForLLMTest)
	mockTelemetryRepo := new(MockTelemetryRepoForLLMTest)
	mockTaskLogRepo := new(MockTaskLogRepoForLLMTest)

	// Create AgentService with mocks
	agentService := &AgentService{
		taskLogRepo:   mockTaskLogRepo,
		deviceRepo:    mockDeviceRepo,
		telemetryRepo: mockTelemetryRepo,
	}

	ctx := context.Background()
	deviceID := "" // Empty device ID

	// Execute GetDeviceContext with empty device ID
	contextData, err := agentService.GetDeviceContext(ctx, deviceID)

	// Assertions - should return empty context without calling repositories
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	assert.Empty(t, contextData)

	// No repository calls should be made
	mockDeviceRepo.AssertNotCalled(t, "GetByID")
	mockTelemetryRepo.AssertNotCalled(t, "GetByDeviceID")
}

// ============================================
// Mock Implementations for LLM Tests
// ============================================

// MockDeviceRepoForLLMTest for GetDeviceContext tests
type MockDeviceRepoForLLMTest struct {
	mock.Mock
}

func (m *MockDeviceRepoForLLMTest) GetByID(ctx context.Context, id string) (*model.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepoForLLMTest) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) GetByIDWithTenant(ctx context.Context, id string, tenantID string) (*model.Device, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepoForLLMTest) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Device), args.Get(1).(int), args.Error(2)
}

func (m *MockDeviceRepoForLLMTest) Update(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockDeviceRepoForLLMTest) WithTx(tx database.TransactionInterface) repository.DeviceRepositoryInterface {
	return m
}

func (m *MockDeviceRepoForLLMTest) BatchCreate(ctx context.Context, devices []*model.Device) error {
	args := m.Called(ctx, devices)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) BatchUpdate(ctx context.Context, devices []*model.Device) error {
	args := m.Called(ctx, devices)
	return args.Error(0)
}

func (m *MockDeviceRepoForLLMTest) BatchUpdateStatus(ctx context.Context, deviceStatuses map[string]string) error {
	args := m.Called(ctx, deviceStatuses)
	return args.Error(0)
}

// MockTelemetryRepoForLLMTest for GetDeviceContext tests
type MockTelemetryRepoForLLMTest struct {
	mock.Mock
}

func (m *MockTelemetryRepoForLLMTest) Insert(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTelemetryRepoForLLMTest) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, start, end, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryRepoForLLMTest) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryRepoForLLMTest) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	args := m.Called(ctx, deviceID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DeviceStats), args.Error(1)
}

func (m *MockTelemetryRepoForLLMTest) GetStatsBatch(ctx context.Context, deviceIDs []string, start, end time.Time) (map[string]*model.DeviceStats, error) {
	args := m.Called(ctx, deviceIDs, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*model.DeviceStats), args.Error(1)
}

// MockTaskLogRepoForLLMTest for GetDeviceContext tests
type MockTaskLogRepoForLLMTest struct {
	mock.Mock
}

func (m *MockTaskLogRepoForLLMTest) Create(ctx context.Context, log *model.AgentTaskLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockTaskLogRepoForLLMTest) List(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AgentTaskLog), args.Error(1)
}

func (m *MockTaskLogRepoForLLMTest) GetBySessionID(ctx context.Context, sessionID string) (*model.AgentTaskLog, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentTaskLog), args.Error(1)
}

func (m *MockTaskLogRepoForLLMTest) WithTx(tx database.TransactionInterface) repository.AgentTaskLogRepositoryInterface {
	return m
}