package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ============================================
// HealthService Tests - New Architecture
// ============================================

func TestHealthService_InitHealthService(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	config := HealthServiceConfig{
		LLMAPIKey:    "test-key",
		LLMBseURL:    "https://test-url",
		LLMModel:     "test-model",
		CheckTimeout: 5 * time.Second,
	}

	svc := InitHealthService(db, "1.0.0", config)
	assert.NotNil(t, svc)
	assert.Equal(t, "1.0.0", svc.version)
	assert.Equal(t, config.LLMAPIKey, svc.config.LLMAPIKey)
}

func TestHealthService_InitHealthServiceWithClient(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	config := DefaultHealthServiceConfig()
	mockClient := &MockHTTPClient{
		Response: &http.Response{StatusCode: 200},
		Error:    nil,
	}

	svc := InitHealthServiceWithClient(db, "1.0.0", config, mockClient)
	assert.NotNil(t, svc)
	assert.Equal(t, mockClient, svc.httpClient)
}

func TestHealthService_CheckHealth_DatabaseOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock Ping
	mock.ExpectPing()

	config := DefaultHealthServiceConfig()
	mockClient := &MockHTTPClient{
		Response: &http.Response{StatusCode: 200},
	}

	svc := InitHealthServiceWithClient(db, "1.0.0", config, mockClient)

	ctx := context.Background()
	result := svc.CheckHealth(ctx)

	assert.NotNil(t, result)
	assert.Contains(t, []string{"healthy", "degraded"}, result.Status)
}

func TestHealthService_CheckHealth_LLMDisabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	config := HealthServiceConfig{
		LLMAPIKey:    "", // 禁用 LLM
		CheckTimeout: 5 * time.Second,
	}

	svc := InitHealthService(db, "1.0.0", config)

	ctx := context.Background()
	result := svc.CheckHealth(ctx)

	assert.NotNil(t, result)
	assert.Contains(t, []string{"healthy", "degraded"}, result.Status)
	assert.Equal(t, "unavailable", result.Components.LLMAPI.Status)
}

func TestHealthService_CheckHealth_MemoryOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	config := DefaultHealthServiceConfig()
	svc := InitHealthService(db, "1.0.0", config)

	ctx := context.Background()
	result := svc.CheckHealth(ctx)

	assert.NotNil(t, result)
	assert.Contains(t, []string{"healthy", "degraded"}, result.Components.Memory.Status)
}
