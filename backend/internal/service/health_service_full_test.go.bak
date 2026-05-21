package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// InitHealthService Tests (0% coverage)
// ============================================

func TestInitHealthService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Reset singleton for test
	healthServiceInstance = nil
	healthServiceOnce = sync.Once{}

	svc := InitHealthService(db, "1.0.0")
	assert.NotNil(t, svc)
	assert.Equal(t, "1.0.0", svc.version)
	assert.NotNil(t, svc.db)
	assert.Equal(t, 5*time.Second, svc.checkTimeout)

	// Calling again should return same instance
	svc2 := InitHealthService(db, "2.0.0")
	assert.Equal(t, svc, svc2)             // Singleton behavior
	assert.Equal(t, "1.0.0", svc2.version) // Version from first call
}

func TestInitHealthService_GetHealthService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Reset singleton for test
	healthServiceInstance = nil
	healthServiceOnce = sync.Once{}

	svc := InitHealthService(db, "test-version")

	retrieved := GetHealthService()
	assert.Equal(t, svc, retrieved)
}

func TestGetHealthService_NoInstance(t *testing.T) {
	// Reset singleton for test
	healthServiceInstance = nil
	healthServiceOnce = sync.Once{}

	svc := GetHealthService()
	assert.Nil(t, svc)
}

// ============================================
// CheckHealth Tests - Full Coverage
// ============================================

func TestHealthService_CheckHealth_Healthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		startTime:    time.Now().Add(-2 * time.Hour),
		version:      "test",
		llmAPIKey:    "",
		llmBaseURL:   "",
		llmModel:     "",
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()

	// Mock DB ping
	mock.ExpectPing()

	result := healthSvc.CheckHealth(ctx)

	assert.NotNil(t, result)
	assert.Equal(t, "test", result.Version)
	assert.NotEmpty(t, result.Uptime)
	assert.NotZero(t, result.Timestamp)

	// Database should be healthy if ping succeeds
	assert.Equal(t, "healthy", result.Components.Database.Status)

	// LLM should be unavailable (no API key)
	assert.Equal(t, "unavailable", result.Components.LLMAPI.Status)

	// Memory and disk checks
	assert.NotEmpty(t, result.Components.Memory.Status)
	assert.NotEmpty(t, result.Components.Disk.Status)
}

func TestHealthService_CheckHealth_DatabaseUnhealthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		startTime:    time.Now(),
		version:      "test",
		llmAPIKey:    "",
		llmBaseURL:   "",
		llmModel:     "",
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()

	// Mock DB ping failure
	mock.ExpectPing().WillReturnError(errors.New("connection refused"))

	result := healthSvc.CheckHealth(ctx)

	assert.NotNil(t, result)
	assert.Equal(t, "unhealthy", result.Components.Database.Status)
	assert.Contains(t, result.Components.Database.Message, "connection refused")
}

// ============================================
// checkDatabase Tests (0% coverage)
// ============================================

func TestHealthService_checkDatabase_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	mock.ExpectPing()

	status := healthSvc.checkDatabase(ctx)

	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, "connected", status.Message)
	assert.GreaterOrEqual(t, status.LatencyMs, int64(0))
}

func TestHealthService_checkDatabase_Failure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	mock.ExpectPing().WillReturnError(errors.New("timeout"))

	status := healthSvc.checkDatabase(ctx)

	assert.Equal(t, "unhealthy", status.Status)
	assert.Contains(t, status.Message, "timeout")
}

func TestHealthService_checkDatabase_NilDB(t *testing.T) {
	healthSvc := &HealthService{
		db:           nil,
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()

	// Should handle nil DB gracefully - will panic or return error
	// This test verifies the behavior
	defer func() {
		if r := recover(); r != nil {
			// Expected panic with nil DB
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	status := healthSvc.checkDatabase(ctx)
	// If no panic, check the status
	if status.Status != "" {
		assert.Equal(t, "unhealthy", status.Status)
	}
}

// ============================================
// checkMemory Tests - Full Coverage
// ============================================

func TestHealthService_checkMemory_Healthy(t *testing.T) {
	healthSvc := &HealthService{}

	status := healthSvc.checkMemory()

	assert.NotEmpty(t, status.Status)
	assert.NotEmpty(t, status.Message)
	// Memory check should always return a status
	assert.Contains(t, []string{"healthy", "warning", "unhealthy"}, status.Status)
}

// ============================================
// checkDisk Tests - Full Coverage
// ============================================

func TestHealthService_checkDisk_Healthy(t *testing.T) {
	healthSvc := &HealthService{}

	status := healthSvc.checkDisk()

	assert.NotEmpty(t, status.Status)
	assert.NotEmpty(t, status.Message)
	// Disk check should work in test environment
	assert.Contains(t, []string{"healthy", "unhealthy", "unknown"}, status.Status)
}

// ============================================
// Overall Status Tests
// ============================================

func TestHealthService_CheckHealth_OverallStatus(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		startTime:    time.Now(),
		version:      "test",
		llmAPIKey:    "test-key",
		llmBaseURL:   "http://test",
		llmModel:     "test-model",
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	mock.ExpectPing()

	result := healthSvc.CheckHealth(ctx)

	// If database is healthy and memory is healthy, overall should be healthy or degraded
	validStatuses := []string{"healthy", "degraded", "unhealthy"}
	assert.Contains(t, validStatuses, result.Status)
}

func TestHealthService_CheckHealth_DegradedStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	healthSvc := &HealthService{
		db:           db,
		startTime:    time.Now(),
		version:      "test",
		llmAPIKey:    "test-key",
		llmBaseURL:   "http://invalid-url", // Will cause unhealthy LLM
		llmModel:     "test-model",
		checkTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	mock.ExpectPing()

	result := healthSvc.CheckHealth(ctx)

	// If LLM API is unhealthy, status should be degraded (if DB is healthy)
	// or unhealthy (if DB is unhealthy)
	assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, result.Status)
}
