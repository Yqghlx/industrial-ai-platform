package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// ============================================
// Redis Token Blacklist Tests
// ============================================

func TestRedisTokenBlacklist_New(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)
	assert.NotNil(t, blacklist)
}

func TestRedisTokenBlacklist_Add(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	ctx := context.Background()
	tokenID := "test-token-123"
	expiry := time.Hour

	err = blacklist.Add(ctx, tokenID, expiry)
	assert.NoError(t, err)
}

func TestRedisTokenBlacklist_Exists(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	ctx := context.Background()
	tokenID := "test-token-123"

	// 未添加前不存在
	exists := blacklist.Exists(ctx, tokenID)
	assert.False(t, exists)

	// 添加后存在
	blacklist.Add(ctx, tokenID, time.Hour)
	exists = blacklist.Exists(ctx, tokenID)
	assert.True(t, exists)
}

func TestRedisTokenBlacklist_AddUserRevocation(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	ctx := context.Background()
	userID := 1
	revokedAt := time.Now()
	duration := time.Hour

	err = blacklist.AddUserRevocation(ctx, userID, revokedAt, duration)
	assert.NoError(t, err)
}

func TestRedisTokenBlacklist_GetUserRevocation(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	ctx := context.Background()
	userID := 1

	// 未添加前返回 0
	version, err := blacklist.GetUserRevocation(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}, version)

	// 添加后返回时间
	blacklist.AddUserRevocation(ctx, userID, time.Now(), time.Hour)
	version, err = blacklist.GetUserRevocation(ctx, userID)
	assert.NoError(t, err)
	assert.NotEqual(t, time.Time{}, version)
}

func TestRedisTokenBlacklist_Stop(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewRedisTokenBlacklist(client)

	// Stop 方法（Redis 无需清理）
	blacklist.Stop()
}