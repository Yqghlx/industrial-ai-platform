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
	"github.com/industrial-ai/platform/pkg/database"
)

// ============================================
// NewAgentServiceWithConfig Tests (0% coverage)
// ============================================

func TestNewAgentServiceWithConfig_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{
		HTTPTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		LLMAPIKey:           "test-key",
		LLMBaseURL:          "http://test-url",
		LLMModel:            "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.httpClient)
	assert.Equal(t, config.HTTPTimeout, svc.httpClient.Timeout)
	assert.Equal(t, "test-key", svc.apiKey)
	assert.Equal(t, "http://test-url", svc.baseURL)
	assert.Equal(t, "test-model", svc.model)
}

func TestNewAgentServiceWithConfig_NilConfig(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	// Nil config should use defaults
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, nil)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.config)
	assert.NotZero(t, svc.config.HTTPTimeout)
}

// ============================================
// GetDeviceContext Tests (0% coverage)
// ============================================

func TestAgentService_GetDeviceContext_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Mock device lookup
	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC Machine", "cnc", "Line 1", "online", "tenant1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices WHERE id").
		WithArgs("CNC-001").
		WillReturnRows(deviceRows)

	// Note: GetDeviceContext always returns nil error, errors are handled internally
	contextData, err := svc.GetDeviceContext(ctx, "CNC-001")
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	// Device should be present
	assert.Contains(t, contextData, "device")
}

func TestAgentService_GetDeviceContext_DeviceNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .* FROM devices WHERE id").
		WithArgs("nonexistent").
		WillReturnError(errors.New("not found"))

	// GetDeviceContext returns empty map when device not found, not an error
	contextData, err := svc.GetDeviceContext(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.NotNil(t, contextData)
	// Should be empty map
	assert.Empty(t, contextData)
}

func TestAgentService_GetDeviceContext_TelemetryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	deviceRows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "tenant_id", "created_at", "updated_at"}).
		AddRow("CNC-001", "CNC Machine", "cnc", "Line 1", "online", "tenant1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .* FROM devices WHERE id").WillReturnRows(deviceRows)

	// Telemetry error - GetDeviceContext handles this internally and continues
	// Note: telemetry query has time range parameters, may not match exactly
	mock.ExpectQuery("SELECT .* FROM telemetry").
		WillReturnError(errors.New("no telemetry"))

	contextData, err := svc.GetDeviceContext(ctx, "CNC-001")
	assert.NoError(t, err) // Should handle telemetry error gracefully
	assert.NotNil(t, contextData)
	assert.Contains(t, contextData, "device")
	// Telemetry may not be present due to error
}

// ============================================
// callLLM Tests (0% coverage)
// ============================================

func TestAgentService_callLLM_NoAPIKey(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{
		LLMAPIKey:  "", // No API key
		LLMBaseURL: "http://test",
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Without API key, callLLM should fail or return error
	_, err = svc.callLLM(ctx, "test query", nil, "通用智能体")
	assert.Error(t, err)
}

// ============================================
// Query Tests - Additional Coverage
// ============================================

func TestAgentService_Query_WithSessionID2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{
		LLMAPIKey: "", // No API key, use mock response
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	query := model.AgentQuery{
		Query:     "分析设备状态",
		SessionID: "existing-session-123",
	}

	mock.ExpectQuery("INSERT INTO agent_task_logs").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	response, err := svc.Query(ctx, query)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "existing-session-123", response.SessionID)
}

// ============================================
// generateSessionID Tests - Additional Coverage
// ============================================

func TestGenerateSessionID_Unique2(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should generate unique IDs
}

// ============================================
// mockResponse Tests - Additional Coverage
// ============================================

func TestAgentService_MockResponse_AllAgents2(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体", "unknown_agent"}

	for _, agent := range agents {
		response := svc.mockResponse("test query", agent)
		assert.NotEmpty(t, response)
	}
}

// ============================================
// GetTaskLogs Tests - Additional Coverage
// ============================================

func TestAgentService_GetTaskLogs_NegativeLimit2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Negative limit should default to 50
	mock.ExpectQuery("SELECT .* FROM agent_task_logs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"}))

	logs, err := svc.GetTaskLogs(ctx, -10)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
}

func TestAgentService_GetTaskLogs_ZeroLimit2(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .* FROM agent_task_logs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"}))

	logs, err := svc.GetTaskLogs(ctx, 0)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
}

// ============================================
// AnalyzeQuery Tests
// ============================================

func TestAgentService_AnalyzeQuery2(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	// AnalyzeQuery takes a string and returns map[string]interface{}
	analysis := svc.AnalyzeQuery("设备温度异常")

	assert.NotNil(t, analysis)
}

// ============================================
// buildSystemPrompt Tests - Additional Coverage
// ============================================

func TestAgentService_BuildSystemPrompt2(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体", "unknown"}

	for _, agent := range agents {
		prompt := svc.buildSystemPrompt(agent)
		assert.NotEmpty(t, prompt)
	}
}
