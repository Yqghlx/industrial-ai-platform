package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentService_Query_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)
	ctx := context.Background()

	query := model.AgentQuery{
		Query:     "分析CNC-001设备状态",
		SessionID: "test-session-123",
	}

	// Expect task log creation
	mock.ExpectQuery("INSERT INTO agent_task_logs").
		WithArgs("test-session-123", "分析CNC-001设备状态", sqlmock.AnyArg(), "设备专家", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Execute query
	response, err := agentService.Query(ctx, query)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-session-123", response.SessionID)
	assert.NotEmpty(t, response.Response)
	assert.NotEmpty(t, response.Agent)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentService_Query_GeneratesSessionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)
	ctx := context.Background()

	query := model.AgentQuery{
		Query: "给我维护建议",
	}

	// Expect task log creation with generated session ID
	mock.ExpectQuery("INSERT INTO agent_task_logs").
		WithArgs(sqlmock.AnyArg(), "给我维护建议", sqlmock.AnyArg(), "维护专家", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Execute query
	response, err := agentService.Query(ctx, query)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.SessionID)
	assert.Contains(t, response.SessionID, "session_")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentService_DetermineAgent_DeviceExpert(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test queries that should route to 设备专家
	queries := []string{
		"分析设备状态",
		"device temperature",
		"温度异常",
		"振动幅度",
		"vibration",
	}

	for _, q := range queries {
		agent := agentService.determineAgent(q)
		assert.Equal(t, "设备专家", agent)
	}
}

func TestAgentService_DetermineAgent_MaintenanceExpert(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test queries that should route to 维护专家
	queries := []string{
		"维护建议",
		"maintenance schedule",
		"工单管理",
		"repair needed",
	}

	for _, q := range queries {
		agent := agentService.determineAgent(q)
		assert.Equal(t, "维护专家", agent)
	}
}

func TestAgentService_DetermineAgent_PredictExpert(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test queries that should route to 预测专家
	queries := []string{
		"预测故障",
		"predict failure",
		"故障风险",
		"fault analysis",
	}

	for _, q := range queries {
		agent := agentService.determineAgent(q)
		assert.Equal(t, "预测专家", agent)
	}
}

func TestAgentService_DetermineAgent_OptimizeExpert(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test queries that should route to 优化专家
	queries := []string{
		"优化生产",
		"optimize efficiency",
		"效率提升",
		"efficiency improvement",
	}

	for _, q := range queries {
		agent := agentService.determineAgent(q)
		assert.Equal(t, "优化专家", agent)
	}
}

func TestAgentService_DetermineAgent_DefaultAgent(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test queries that should route to 通用智能体
	queries := []string{
		"你好",
		"hello",
		"普通问题",
		"general question",
	}

	for _, q := range queries {
		agent := agentService.determineAgent(q)
		assert.Equal(t, "通用智能体", agent)
	}
}

func TestAgentService_MockResponse(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test mock response for each agent type
	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体"}

	for _, agent := range agents {
		response := agentService.mockResponse("test query", agent)
		assert.NotEmpty(t, response)
		assert.Contains(t, response, "**") // Mock responses contain formatted text
	}
}

func TestAgentService_GetTaskLogs_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)
	ctx := context.Background()

	// Expect query for listing task logs
	rows := sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"})
	now := time.Now()
	rows.AddRow(1, "session-1", "query-1", "response-1", "设备专家", now)
	rows.AddRow(2, "session-2", "query-2", "response-2", "维护专家", now)

	mock.ExpectQuery("SELECT id, session_id, query, response, agent, executed_at FROM agent_task_logs").
		WillReturnRows(rows)

	// Execute GetTaskLogs
	logs, err := agentService.GetTaskLogs(ctx, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, logs, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentService_GetTaskLogs_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)
	ctx := context.Background()

	// Expect query with default limit (50)
	rows := sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"})
	mock.ExpectQuery("SELECT.*FROM agent_task_logs.*LIMIT").
		WillReturnRows(rows)

	// Execute GetTaskLogs with limit 0 (should use default)
	logs, err := agentService.GetTaskLogs(ctx, 0)

	// Assertions
	assert.NoError(t, err)
	// Empty result returns empty slice, not nil
	assert.NotNil(t, logs)
	assert.Len(t, logs, 0) // Empty array

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentService_AnalyzeQuery_ExtractDeviceID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test device ID extraction
	analysis := agentService.AnalyzeQuery("分析CNC-001设备状态")

	assert.NotNil(t, analysis)
	assert.Equal(t, "CNC-001", analysis["possible_device_id"])
}

func TestAgentService_AnalyzeQuery_ExtractIntent(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test intent extraction
	testCases := map[string]string{
		"分析设备":     "analyze",
		"预测故障":     "predict",
		"维护设备":     "maintain",
		"优化效率":     "optimize",
		"普通问题":     "query",
		"analyze":  "analyze",
		"predict":  "predict",
		"maintain": "maintain",
		"optimize": "optimize",
	}

	for query, expectedIntent := range testCases {
		analysis := agentService.AnalyzeQuery(query)
		assert.Equal(t, expectedIntent, analysis["intent"])
	}
}

func TestAgentService_BuildSystemPrompt(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(database.NewDBWrapper(db))
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(db))
	telemetryRepo := repository.NewTelemetryRepository(database.NewDBWrapper(db))

	agentService := NewAgentService(taskLogRepo, deviceRepo, telemetryRepo, nil)

	// Test system prompt generation for each agent
	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体"}

	for _, agent := range agents {
		prompt := agentService.buildSystemPrompt(agent)
		assert.NotEmpty(t, prompt)
		assert.Contains(t, prompt, "工业AI平台")
		assert.Contains(t, prompt, agent)
	}
}

// TestAgentService_BuildSystemPrompt_UnknownAgent moved to agent_service_callllm_test.go

func TestGenerateSessionID(t *testing.T) {
	sessionID := generateSessionID()

	assert.NotEmpty(t, sessionID)
	assert.Contains(t, sessionID, "session_")
	// "session_" (8 chars) + 32 hex chars (16 bytes) = 40 chars
	assert.Len(t, sessionID, 40)
}
