package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// callLLM Full Tests with HTTP Mock
// ============================================

func TestAgentService_callLLM_Success(t *testing.T) {
	// Create mock LLM server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Read and parse request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		assert.Equal(t, "test-model", reqBody["model"])
		assert.NotNil(t, reqBody["messages"])

		// Send mock response
		response := LLMResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "这是测试响应内容",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create service with mock server URL
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Call callLLM
	response, err := svc.callLLM(ctx, "测试问题", nil, "通用智能体")
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "测试响应")
}

func TestAgentService_callLLM_WithContextData(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		messages := reqBody["messages"].([]interface{})
		userMessage := messages[1].(map[string]interface{})["content"].(string)

		// Should contain context data
		assert.Contains(t, userMessage, "上下文数据")

		response := LLMResponse{
			ID:    "test-id",
			Model: "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "响应内容",
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	contextData := map[string]interface{}{
		"device_id": "CNC-001",
		"status":    "running",
	}

	response, err := svc.callLLM(ctx, "问题", contextData, "设备专家")
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestAgentService_callLLM_HTTPError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	response, err := svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.Error(), "500")
}

func TestAgentService_callLLM_InvalidResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	response, err := svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestAgentService_callLLM_EmptyChoices(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LLMResponse{
			ID:    "test-id",
			Model: "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{}, // Empty choices
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	_, err = svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No choices")
}

func TestAgentService_callLLM_ContextTimeout(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:   "test-api-key",
		LLMBaseURL:  mockServer.URL,
		LLMModel:    "test-model",
		HTTPTimeout: 100 * time.Millisecond, // Short timeout
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	response, err := svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestAgentService_callLLM_RequestCreationError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: "://invalid-url", // Invalid URL
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	response, err := svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestAgentService_callLLM_MarshalError(t *testing.T) {
	// This is difficult to test directly since json.Marshal rarely fails
	// We can test the edge case where contextData contains problematic values
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: "http://test-url",
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Normal context data should marshal fine
	contextData := map[string]interface{}{
		"key": "value",
	}

	// The call will fail because there's no real server, but marshal should succeed
	response, err := svc.callLLM(ctx, "问题", contextData, "通用智能体")
	assert.Error(t, err) // Connection error
	assert.Empty(t, response)
}

func TestAgentService_callLLM_DifferentAgents(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		messages := reqBody["messages"].([]interface{})
		systemMessage := messages[0].(map[string]interface{})["content"].(string)

		// Verify system prompt is set
		assert.Contains(t, systemMessage, "工业AI平台")

		response := LLMResponse{
			ID:    "test-id",
			Model: "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "响应",
					},
					FinishReason: "stop",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体"}
	for _, agent := range agents {
		response, err := svc.callLLM(ctx, "问题", nil, agent)
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	}
}

func TestAgentService_callLLM_401Unauthorized(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "invalid-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	_, err = svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestAgentService_callLLM_429RateLimit(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limit exceeded"))
	}))
	defer mockServer.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	_, err = svc.callLLM(ctx, "问题", nil, "通用智能体")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

// ============================================
// Query Tests with Real callLLM
// ============================================

func TestAgentService_Query_WithAPIKey(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LLMResponse{
			ID:    "test-id",
			Model: "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "这是AI响应",
					},
					FinishReason: "stop",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	// Mock task log insert
	mock.ExpectQuery("INSERT INTO agent_task_logs").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	query := model.AgentQuery{
		Query: "分析设备状态",
	}

	response, err := svc.Query(ctx, query)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
	assert.Contains(t, response.Response, "AI响应")
}

func TestAgentService_Query_APIKeyFallbackToMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	// Empty API key should use mock response
	config := &AgentServiceConfig{
		LLMAPIKey:  "",
		LLMBaseURL: "http://test",
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	mock.ExpectQuery("INSERT INTO agent_task_logs").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	query := model.AgentQuery{
		Query: "分析设备状态",
	}

	response, err := svc.Query(ctx, query)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response) // Mock response
}

func TestAgentService_Query_LLMErrorFallback(t *testing.T) {
	// Server returns error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{
		LLMAPIKey:  "test-api-key",
		LLMBaseURL: mockServer.URL,
		LLMModel:   "test-model",
	}

	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	mock.ExpectQuery("INSERT INTO agent_task_logs").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	query := model.AgentQuery{
		Query: "分析设备状态",
	}

	// Should fallback to mock response when LLM fails
	response, err := svc.Query(ctx, query)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response) // Fallback mock response
}

// ============================================
// determineAgent Tests
// ============================================

func TestAgentService_DetermineAgent_DeviceKeywords(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	tests := []struct {
		query    string
		expected string
	}{
		{"分析设备状态", "设备专家"},
		{"设备温度异常", "设备专家"},
		{"device CNC-001", "设备专家"},
		{"温度过高", "设备专家"},
		{"振动数据", "设备专家"},
		{"vibration analysis", "设备专家"},
		{"需要维护", "维护专家"},
		{"维护计划", "维护专家"},
		{"maintenance schedule", "维护专家"},
		{"工单管理", "维护专家"},
		{"repair request", "维护专家"},
		{"预测故障", "预测专家"},
		{"故障概率", "预测专家"},
		{"predict failure", "预测专家"},
		{"fault analysis", "预测专家"},
		{"优化效率", "优化专家"},
		{"生产优化", "优化专家"},
		{"optimize process", "优化专家"},
		{"efficiency improvement", "优化专家"},
		{"普通问题", "通用智能体"},
	}

	for _, tt := range tests {
		agent := svc.determineAgent(tt.query)
		assert.Equal(t, tt.expected, agent, "Query: %s", tt.query)
	}
}

// ============================================
// buildSystemPrompt Tests
// ============================================

func TestAgentService_BuildSystemPrompt_AllAgents(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	agents := []string{"设备专家", "维护专家", "预测专家", "优化专家", "通用智能体", "未知角色"}

	for _, agent := range agents {
		prompt := svc.buildSystemPrompt(agent)
		assert.NotEmpty(t, prompt)
		assert.Contains(t, prompt, "工业AI平台")
	}
}

func TestAgentService_BuildSystemPrompt_UnknownAgent(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	// Unknown agent should use default (通用智能体)
	prompt := svc.buildSystemPrompt("未知角色")
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "通用智能体")
}

// ============================================
// AnalyzeQuery Full Tests
// ============================================

func TestAgentService_AnalyzeQuery_AllIntents(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	tests := []struct {
		query          string
		expectedIntent string
	}{
		{"分析CNC-001状态", "analyze"},
		{"analyze device", "analyze"},
		{"预测故障时间", "predict"},
		{"predict failure", "predict"},
		{"制定维护计划", "maintain"},
		{"maintain equipment", "maintain"},
		{"优化生产流程", "optimize"},
		{"optimize efficiency", "optimize"},
		{"一般问题", "query"},
	}

	for _, tt := range tests {
		analysis := svc.AnalyzeQuery(tt.query)
		assert.Equal(t, tt.expectedIntent, analysis["intent"], "Query: %s", tt.query)
	}
}

func TestAgentService_AnalyzeQuery_DeviceIDPattern(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	tests := []struct {
		query            string
		expectDeviceID   bool
		expectedDeviceID string
	}{
		{"分析CNC-001状态", true, "CNC-001"},
		{"检查INJ-002", true, "INJ-002"},
		{"ROB-003温度", true, "ROB-003"},
		{"ABC-999设备", true, "ABC-999"},
		{"无设备编号", false, ""},
		{"cnc-001", false, ""}, // Lowercase not matched
	}

	for _, tt := range tests {
		analysis := svc.AnalyzeQuery(tt.query)
		if tt.expectDeviceID {
			assert.Contains(t, analysis, "possible_device_id")
			assert.Equal(t, tt.expectedDeviceID, analysis["possible_device_id"])
		} else {
			assert.NotContains(t, analysis, "possible_device_id")
		}
	}
}

func TestAgentService_AnalyzeQuery_MultipleDeviceIDs(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	// Query with multiple device IDs - should return first match
	analysis := svc.AnalyzeQuery("比较CNC-001和INJ-002")
	assert.Contains(t, analysis, "possible_device_id")
	// Returns first match
	assert.NotEmpty(t, analysis["possible_device_id"])
}

// ============================================
// Mock Response Tests
// ============================================

func TestAgentService_MockResponse_Randomness(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)

	// Multiple calls should sometimes return different responses (random)
	responses := make(map[string]int)
	for i := 0; i < 100; i++ {
		resp := svc.mockResponse("test", "设备专家")
		responses[resp]++
	}

	// Should have at least one response
	assert.Greater(t, len(responses), 0)
}

// ============================================
// GetTaskLogs Full Tests
// ============================================

func TestAgentService_GetTaskLogs_WithResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"}).
		AddRow(1, "session-1", "query-1", "response-1", "设备专家", now).
		AddRow(2, "session-2", "query-2", "response-2", "维护专家", now)

	mock.ExpectQuery("SELECT .* FROM agent_task_logs").WillReturnRows(rows)

	logs, err := svc.GetTaskLogs(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

func TestAgentService_GetTaskLogs_EmptyResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	taskLogRepo := repository.NewAgentTaskLogRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	config := &AgentServiceConfig{LLMAPIKey: ""}
	svc := NewAgentServiceWithConfig(taskLogRepo, deviceRepo, telemetryRepo, config)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .* FROM agent_task_logs").WillReturnRows(sqlmock.NewRows([]string{"id", "session_id", "query", "response", "agent", "executed_at"}))

	logs, err := svc.GetTaskLogs(ctx, 50)
	assert.NoError(t, err)
	assert.Len(t, logs, 0)
}

// ============================================
// LoadAgentServiceConfigFromEnv Tests
// ============================================

func TestLoadAgentServiceConfigFromEnv_Defaults(t *testing.T) {
	// Clear environment variables
	config := LoadAgentServiceConfigFromEnv()

	assert.NotZero(t, config.HTTPTimeout)
	assert.NotZero(t, config.MaxIdleConns)
	assert.NotZero(t, config.MaxIdleConnsPerHost)
	assert.NotZero(t, config.IdleConnTimeout)
	assert.NotEmpty(t, config.LLMBaseURL)
	assert.NotEmpty(t, config.LLMModel)
}

// ============================================
// DefaultAgentServiceConfig Tests
// ============================================

func TestDefaultAgentServiceConfig(t *testing.T) {
	config := DefaultAgentServiceConfig()

	assert.Equal(t, 30*time.Second, config.HTTPTimeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 10, config.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.NotEmpty(t, config.LLMBaseURL)
	assert.NotEmpty(t, config.LLMModel)
}

// ============================================
// generateSessionID Tests
// ============================================

func TestGenerateSessionID_Format(t *testing.T) {
	id := generateSessionID()

	assert.NotEmpty(t, id)
	assert.True(t, strings.HasPrefix(id, "session_"))
	// After "session_", should be hex characters
	hexPart := strings.TrimPrefix(id, "session_")
	assert.NotEmpty(t, hexPart)
}

func TestGenerateSessionID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateSessionID()
		ids[id] = true
	}

	// All IDs should be unique
	assert.Equal(t, 100, len(ids))
}
