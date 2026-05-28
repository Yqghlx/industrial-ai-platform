package service

import (
	"context"
	"testing"
	"time"

	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCacheService for testing AgentOptimizer - implements cache.CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) bool {
	args := m.Called(ctx, key)
	return args.Bool(0)
}

func (m *MockCacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheService) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCacheService) GetStats() cache.Stats {
	args := m.Called()
	return args.Get(0).(cache.Stats)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================
// NewAgentOptimizer Tests
// ============================================

func TestNewAgentOptimizer_WithCache(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("IsAvailable").Return(true)

	optimizer := NewAgentOptimizer(mockCache, 10)

	assert.NotNil(t, optimizer)
	assert.NotNil(t, optimizer.cache)
	assert.NotNil(t, optimizer.semaphore)
	assert.Equal(t, int64(10), optimizer.maxConcurrent)
	assert.Equal(t, 30*time.Minute, optimizer.cacheTTL)
}

func TestNewAgentOptimizer_DefaultMaxConcurrent(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("IsAvailable").Return(true)

	optimizer := NewAgentOptimizer(mockCache, 0) // Zero value should use default

	assert.NotNil(t, optimizer)
	assert.Equal(t, int64(10), optimizer.maxConcurrent) // Default: 10 concurrent LLM calls
}

func TestNewAgentOptimizer_NegativeMaxConcurrent(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("IsAvailable").Return(true)

	optimizer := NewAgentOptimizer(mockCache, -5) // Negative value should use default

	assert.NotNil(t, optimizer)
	assert.Equal(t, int64(10), optimizer.maxConcurrent) // Default: 10 concurrent LLM calls
}

func TestNewAgentOptimizer_NilCache(t *testing.T) {
	optimizer := NewAgentOptimizer(nil, 10)

	assert.NotNil(t, optimizer)
	assert.Nil(t, optimizer.cache)
	assert.NotNil(t, optimizer.semaphore)
	assert.Equal(t, int64(10), optimizer.maxConcurrent)
}

// ============================================
// GetCachedAnswer Tests
// ============================================

func TestGetCachedAnswer_Success(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"
	cachedAnswer := "设备运行正常"

	// Mock cache Get to return cached answer
	mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return([]byte(cachedAnswer), nil)

	answer, found := optimizer.GetCachedAnswer(ctx, query)

	assert.True(t, found)
	assert.Equal(t, cachedAnswer, answer)

	mockCache.AssertExpectations(t)
}

func TestGetCachedAnswer_CacheMiss(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"

	// Mock cache Get to return error (cache miss)
	mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, assert.AnError)

	answer, found := optimizer.GetCachedAnswer(ctx, query)

	assert.False(t, found)
	assert.Empty(t, answer)

	mockCache.AssertExpectations(t)
}

func TestGetCachedAnswer_NilCache(t *testing.T) {
	optimizer := NewAgentOptimizer(nil, 10)

	ctx := context.Background()
	query := "分析设备状态"

	answer, found := optimizer.GetCachedAnswer(ctx, query)

	assert.False(t, found)
	assert.Empty(t, answer)
}

func TestGetCachedAnswer_NilData(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"

	// Mock cache Get to return nil data (should be treated as miss)
	mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, nil)

	answer, found := optimizer.GetCachedAnswer(ctx, query)

	assert.False(t, found)
	assert.Empty(t, answer)

	mockCache.AssertExpectations(t)
}

// ============================================
// CacheAnswer Tests
// ============================================

func TestCacheAnswer_Success(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"
	answer := "设备运行正常"

	// Mock cache Set to succeed
	mockCache.On("Set", ctx, mock.AnythingOfType("string"), []byte(answer), 30*time.Minute).Return(nil)

	optimizer.CacheAnswer(ctx, query, answer)

	mockCache.AssertExpectations(t)
}

func TestCacheAnswer_NilCache(t *testing.T) {
	optimizer := NewAgentOptimizer(nil, 10)

	ctx := context.Background()
	query := "分析设备状态"
	answer := "设备运行正常"

	// Should not panic with nil cache
	optimizer.CacheAnswer(ctx, query, answer)
}

func TestCacheAnswer_EmptyAnswer(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"

	// Should not call cache.Set with empty answer
	optimizer.CacheAnswer(ctx, query, "")

	// No mock expectations - cache should not be called
	mockCache.AssertNotCalled(t, "Set")
}

func TestCacheAnswer_CacheError(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()
	query := "分析设备状态"
	answer := "设备运行正常"

	// Mock cache Set to fail (should not panic, just log warning)
	mockCache.On("Set", ctx, mock.AnythingOfType("string"), []byte(answer), 30*time.Minute).Return(assert.AnError)

	optimizer.CacheAnswer(ctx, query, answer)

	mockCache.AssertExpectations(t)
}

// ============================================
// AcquireSlot Tests
// ============================================

func TestAcquireSlot_Success(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()

	// Should acquire slot without blocking (queue has capacity)
	err := optimizer.AcquireSlot(ctx)

	assert.NoError(t, err)

	// Release after test
	optimizer.ReleaseSlot()
}

func TestAcquireSlot_WithTimeout(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 1) // Only 1 slot

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Acquire first slot
	err := optimizer.AcquireSlot(ctx)
	require.NoError(t, err)

	// Try to acquire another slot - should timeout since only 1 slot available
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	err2 := optimizer.AcquireSlot(ctx2)
	assert.Error(t, err2) // Should fail due to timeout

	// Release the slot
	optimizer.ReleaseSlot()
}

func TestAcquireSlot_CancelledContext(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 1) // Only 1 slot

	// Acquire first slot
	ctx := context.Background()
	err := optimizer.AcquireSlot(ctx)
	require.NoError(t, err)

	// Create cancelled context
	ctx2, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err2 := optimizer.AcquireSlot(ctx2)
	assert.Error(t, err2) // Should fail due to cancelled context

	// Release the slot
	optimizer.ReleaseSlot()
}

// ============================================
// ReleaseSlot Tests
// ============================================

func TestReleaseSlot_Success(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()

	// Acquire slot
	err := optimizer.AcquireSlot(ctx)
	require.NoError(t, err)

	// Release slot - should not panic
	optimizer.ReleaseSlot()

	// Should be able to acquire again
	err2 := optimizer.AcquireSlot(ctx)
	assert.NoError(t, err2)

	// Release again
	optimizer.ReleaseSlot()
}

func TestReleaseSlot_MultipleTimes(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()

	// Acquire slot
	err := optimizer.AcquireSlot(ctx)
	require.NoError(t, err)

	// Release multiple times - semaphore allows this (though not recommended)
	optimizer.ReleaseSlot()
	optimizer.ReleaseSlot()

	// Should still work (semaphore releases more than acquired)
	err2 := optimizer.AcquireSlot(ctx)
	assert.NoError(t, err2)

	// Clean up
	optimizer.ReleaseSlot()
}

// ============================================
// QueueStats Tests (bonus coverage)
// ============================================

func TestQueueStats_EmptyQueue(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	stats := optimizer.QueueStats()

	assert.Equal(t, int64(10), stats["max_concurrent"])
	// current_queue can be 0 or slightly more due to TryAcquire behavior
	assert.GreaterOrEqual(t, stats["current_queue"], int64(0))
	assert.LessOrEqual(t, stats["current_queue"], int64(1))
}

func TestQueueStats_ActiveQueue(t *testing.T) {
	mockCache := new(MockCacheService)
	optimizer := NewAgentOptimizer(mockCache, 10)

	ctx := context.Background()

	// Acquire a slot
	err := optimizer.AcquireSlot(ctx)
	require.NoError(t, err)

	stats := optimizer.QueueStats()

	assert.Equal(t, int64(10), stats["max_concurrent"])
	assert.GreaterOrEqual(t, stats["current_queue"], int64(1)) // At least 1 in queue

	// Release slot
	optimizer.ReleaseSlot()
}