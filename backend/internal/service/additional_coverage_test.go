package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// Additional tests to push coverage to 80%+

func TestAlertService_GetAlertByID_Coverage(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}

	service := &AlertService{
		alertRepo: mockAlertRepo,
	}

	ctx := context.Background()

	t.Run("AlertFound", func(t *testing.T) {
		alert := &model.Alert{ID: 1, DeviceID: "device-1", Message: "Alert 1"}
		mockAlertRepo.On("GetByID", ctx, 1).Return(alert, nil).Once()

		result, err := service.GetAlertByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
	})

	t.Run("AlertNotFound", func(t *testing.T) {
		mockAlertRepo.On("GetByID", ctx, 999).Return(nil, fmt.Errorf("sql: no rows in result set")).Once()

		result, err := service.GetAlertByID(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestHybridTokenBlacklist_Add_Coverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewHybridTokenBlacklist(client)
	defer blacklist.Stop()

	ctx := context.Background()
	tokenID := "test-token-coverage"

	err = blacklist.Add(ctx, tokenID, time.Hour)
	assert.NoError(t, err)

	// Verify exists in both Redis and memory
	exists := blacklist.Exists(ctx, tokenID)
	assert.True(t, exists)
}

func TestHybridTokenBlacklist_AddUserRevocation_Coverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewHybridTokenBlacklist(client)
	defer blacklist.Stop()

	ctx := context.Background()
	userID := 123
	revokedAt := time.Now()

	err = blacklist.AddUserRevocation(ctx, userID, revokedAt, time.Hour)
	assert.NoError(t, err)

	// Verify revocation exists
	result, err := blacklist.GetUserRevocation(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, result.IsZero())
}

func TestHybridTokenBlacklist_GetUserRevocation_Coverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewHybridTokenBlacklist(client)
	defer blacklist.Stop()

	ctx := context.Background()
	userID := 456

	t.Run("NoRevocation", func(t *testing.T) {
		result, err := blacklist.GetUserRevocation(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, time.Time{}, result)
	})

	t.Run("WithRevocation", func(t *testing.T) {
		now := time.Now()
		blacklist.AddUserRevocation(ctx, userID, now, time.Hour)

		result, err := blacklist.GetUserRevocation(ctx, userID)
		assert.NoError(t, err)
		assert.False(t, result.IsZero())
	})
}

func TestHybridTokenBlacklist_New_Coverage(t *testing.T) {
	t.Run("WithRedisClient", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		blacklist := NewHybridTokenBlacklist(client)
		assert.NotNil(t, blacklist)
		assert.NotNil(t, blacklist.memoryBlacklist)
		assert.NotNil(t, blacklist.redisBlacklist)
		assert.True(t, blacklist.useRedis)

		blacklist.Stop()
	})

	t.Run("WithoutRedisClient", func(t *testing.T) {
		blacklist := NewHybridTokenBlacklist(nil)
		assert.NotNil(t, blacklist)
		assert.NotNil(t, blacklist.memoryBlacklist)
		assert.Nil(t, blacklist.redisBlacklist)
		assert.False(t, blacklist.useRedis)

		blacklist.Stop()
	})
}

func TestMemoryTokenBlacklist_Exists_Coverage(t *testing.T) {
	blacklist := NewMemoryTokenBlacklist()
	defer blacklist.Stop()
	ctx := context.Background()

	t.Run("TokenExists", func(t *testing.T) {
		blacklist.Add(ctx, "existing-token", time.Hour)
		exists := blacklist.Exists(ctx, "existing-token")
		assert.True(t, exists)
	})

	t.Run("TokenNotExists", func(t *testing.T) {
		exists := blacklist.Exists(ctx, "non-existing-token")
		assert.False(t, exists)
	})

	t.Run("TokenExpired", func(t *testing.T) {
		blacklist.Add(ctx, "expired-token", 100*time.Millisecond)
		time.Sleep(150 * time.Millisecond)
		exists := blacklist.Exists(ctx, "expired-token")
		assert.False(t, exists)
	})
}

func TestMemoryTokenBlacklist_Add_Coverage(t *testing.T) {
	blacklist := NewMemoryTokenBlacklist()
	defer blacklist.Stop()

	ctx := context.Background()
	err := blacklist.Add(ctx, "test-token", time.Hour)
	assert.NoError(t, err)

	// Verify size increased
	size := blacklist.Size()
	assert.Equal(t, 1, size)
}

func TestRedisTokenBlacklist_Exists_Coverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	ctx := context.Background()

	t.Run("TokenExists", func(t *testing.T) {
		blacklist.Add(ctx, "existing-token", time.Hour)
		exists := blacklist.Exists(ctx, "existing-token")
		assert.True(t, exists)
	})

	t.Run("TokenNotExists", func(t *testing.T) {
		exists := blacklist.Exists(ctx, "non-existent-token")
		assert.False(t, exists)
	})
}

func TestLoadAgentServiceConfigFromEnv_Detailed(t *testing.T) {
	config := LoadAgentServiceConfigFromEnv()

	// Verify all fields are properly set
	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.HTTPTimeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 10, config.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4", config.LLMBaseURL)
	assert.Equal(t, "glm-4-flash", config.LLMModel)
}
