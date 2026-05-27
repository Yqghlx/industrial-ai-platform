package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// AgentService Tests (using repository mocks)
// ============================================

func TestNewAgentServiceWithClient_Coverage(t *testing.T) {
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}

	// Use existing MockHTTPClient from http_client_test.go
	mockHTTPClient := &MockHTTPClient{}

	config := &AgentServiceConfig{
		HTTPTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		LLMAPIKey:           "test-key",
		LLMBaseURL:          "https://test-api.example.com",
		LLMModel:            "test-model",
	}

	t.Run("WithCustomClient", func(t *testing.T) {
		service := NewAgentServiceWithClient(
			mockTaskLogRepo,
			mockDeviceRepo,
			mockTelemetryRepo,
			config,
			mockHTTPClient,
		)

		assert.NotNil(t, service)
		assert.Equal(t, "test-key", service.apiKey)
		assert.Equal(t, "https://test-api.example.com", service.baseURL)
		assert.Equal(t, "test-model", service.model)
		assert.Equal(t, mockHTTPClient, service.httpClient)
	})

	t.Run("WithNilConfig", func(t *testing.T) {
		service := NewAgentServiceWithClient(
			mockTaskLogRepo,
			mockDeviceRepo,
			mockTelemetryRepo,
			nil,
			mockHTTPClient,
		)

		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.Equal(t, 30*time.Second, service.config.HTTPTimeout)
		assert.Equal(t, 100, service.config.MaxIdleConns)
	})

	t.Run("WithNilClient", func(t *testing.T) {
		service := NewAgentServiceWithClient(
			mockTaskLogRepo,
			mockDeviceRepo,
			mockTelemetryRepo,
			config,
			nil,
		)

		assert.NotNil(t, service)
		assert.NotNil(t, service.httpClient)
	})
}

func TestLoadAgentServiceConfigFromEnv_Coverage(t *testing.T) {
	// Test default behavior (no env vars set)
	config := LoadAgentServiceConfigFromEnv()

	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.HTTPTimeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 10, config.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4", config.LLMBaseURL)
	assert.Equal(t, "glm-4-flash", config.LLMModel)
}

func TestDefaultAgentServiceConfig_Coverage(t *testing.T) {
	config := DefaultAgentServiceConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.HTTPTimeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 10, config.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4", config.LLMBaseURL)
	assert.Equal(t, "glm-4-flash", config.LLMModel)
	assert.Empty(t, config.HTTPProxy)
}

func TestNewAgentServiceWithConfig_Coverage(t *testing.T) {
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}

	config := &AgentServiceConfig{
		HTTPTimeout:         45 * time.Second,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     60 * time.Second,
		LLMAPIKey:           "custom-key",
		LLMBaseURL:          "https://custom-llm.example.com",
		LLMModel:            "custom-model",
		HTTPProxy:           "http://proxy.example.com:8080",
	}

	t.Run("WithCustomConfig", func(t *testing.T) {
		service := NewAgentServiceWithConfig(
			mockTaskLogRepo,
			mockDeviceRepo,
			mockTelemetryRepo,
			config,
		)

		assert.NotNil(t, service)
		assert.Equal(t, "custom-key", service.apiKey)
		assert.Equal(t, "https://custom-llm.example.com", service.baseURL)
		assert.Equal(t, "custom-model", service.model)
		assert.NotNil(t, service.httpClient)
	})

	t.Run("WithNilConfig", func(t *testing.T) {
		service := NewAgentServiceWithConfig(
			mockTaskLogRepo,
			mockDeviceRepo,
			mockTelemetryRepo,
			nil,
		)

		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.Equal(t, 30*time.Second, service.config.HTTPTimeout)
	})
}

func TestNewAgentService_Coverage(t *testing.T) {
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}

	service := NewAgentService(
		mockTaskLogRepo,
		mockDeviceRepo,
		mockTelemetryRepo,
		nil, // OPT-002: No cache for test
	)

	assert.NotNil(t, service)
	assert.NotNil(t, service.httpClient)
	assert.NotNil(t, service.config)
}

// MockHTTPClientWithMock for advanced mock scenarios using testify/mock
type MockHTTPClientWithMock struct {
	mock.Mock
}

func (m *MockHTTPClientWithMock) Do(req interface{}) (interface{}, interface{}) {
	args := m.Called(req)
	return args.Get(0), args.Get(1)
}
