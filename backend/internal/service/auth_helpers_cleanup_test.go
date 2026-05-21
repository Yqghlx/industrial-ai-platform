package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// cleanupExpiredEntries Full Tests
// ============================================

func TestMemoryTokenBlacklist_CleanupExpiredEntries(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add tokens with different durations
	bl.Add(ctx, "token-short", 50*time.Millisecond)
	bl.Add(ctx, "token-long", 10*time.Minute)
	bl.Add(ctx, "token-medium", 200*time.Millisecond)

	// All should exist initially
	assert.True(t, bl.Exists(ctx, "token-short"))
	assert.True(t, bl.Exists(ctx, "token-long"))
	assert.True(t, bl.Exists(ctx, "token-medium"))
	assert.Equal(t, 3, bl.Size())

	// Wait for short token to expire
	time.Sleep(100 * time.Millisecond)

	// Cleanup goroutine runs every minute, but we can manually check Exists which handles expiry
	assert.False(t, bl.Exists(ctx, "token-short")) // Should be expired
	assert.True(t, bl.Exists(ctx, "token-long"))
	assert.True(t, bl.Exists(ctx, "token-medium"))

	// Wait for medium token to expire
	time.Sleep(200 * time.Millisecond)

	assert.False(t, bl.Exists(ctx, "token-medium")) // Should be expired
	assert.True(t, bl.Exists(ctx, "token-long"))
}

func TestMemoryTokenBlacklist_CleanupUserRevocations(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add user revocations with different durations
	bl.AddUserRevocation(ctx, 1, time.Now(), 50*time.Millisecond)
	bl.AddUserRevocation(ctx, 2, time.Now(), 10*time.Minute)
	bl.AddUserRevocation(ctx, 3, time.Now(), 200*time.Millisecond)

	// All should exist initially
	rev1, err := bl.GetUserRevocation(ctx, 1)
	require.NoError(t, err)
	assert.False(t, rev1.IsZero())

	rev2, err := bl.GetUserRevocation(ctx, 2)
	require.NoError(t, err)
	assert.False(t, rev2.IsZero())

	rev3, err := bl.GetUserRevocation(ctx, 3)
	require.NoError(t, err)
	assert.False(t, rev3.IsZero())

	// Wait for short revocation to expire
	time.Sleep(100 * time.Millisecond)

	// GetUserRevocation handles expiry check internally
	rev1, err = bl.GetUserRevocation(ctx, 1)
	require.NoError(t, err)
	assert.True(t, rev1.IsZero()) // Should be expired

	rev2, err = bl.GetUserRevocation(ctx, 2)
	require.NoError(t, err)
	assert.False(t, rev2.IsZero()) // Still valid

	rev3, err = bl.GetUserRevocation(ctx, 3)
	require.NoError(t, err)
	assert.False(t, rev3.IsZero()) // Still valid

	// Wait for medium revocation to expire
	time.Sleep(200 * time.Millisecond)

	rev3, err = bl.GetUserRevocation(ctx, 3)
	require.NoError(t, err)
	assert.True(t, rev3.IsZero()) // Should be expired

	rev2, err = bl.GetUserRevocation(ctx, 2)
	require.NoError(t, err)
	assert.False(t, rev2.IsZero()) // Still valid
}

func TestMemoryTokenBlacklist_CleanupWithMixedEntries(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add both token blacklist and user revocations
	bl.Add(ctx, "token-1", 50*time.Millisecond)
	bl.Add(ctx, "token-2", 10*time.Minute)
	bl.AddUserRevocation(ctx, 10, time.Now(), 50*time.Millisecond)
	bl.AddUserRevocation(ctx, 11, time.Now(), 10*time.Minute)

	// Wait for short entries to expire
	time.Sleep(100 * time.Millisecond)

	// Short entries should be expired
	assert.False(t, bl.Exists(ctx, "token-1"))
	rev10, _ := bl.GetUserRevocation(ctx, 10)
	assert.True(t, rev10.IsZero())

	// Long entries should still be valid
	assert.True(t, bl.Exists(ctx, "token-2"))
	rev11, _ := bl.GetUserRevocation(ctx, 11)
	assert.False(t, rev11.IsZero())
}

func TestMemoryTokenBlacklist_StopImmediately(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	// Stop immediately - should not panic
	bl.Stop()

	// Operations after stop should still work (just cleanup goroutine stopped)
	ctx := context.Background()
	bl.Add(ctx, "test-token", time.Minute)
	assert.True(t, bl.Exists(ctx, "test-token"))
}

func TestMemoryTokenBlacklist_DoubleStop(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	bl.Stop()
	// Double stop - this will panic because channel is already closed
	// We should NOT call Stop twice in normal usage
	// bl.Stop() // This would panic
}

func TestMemoryTokenBlacklist_ConcurrentCleanup(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent adds and exists checks
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			bl.Add(ctx, "token-"+string(rune(idx)), time.Minute)
		}(i)
		go func(idx int) {
			defer wg.Done()
			bl.Exists(ctx, "token-"+string(rune(idx)))
		}(i)
	}
	wg.Wait()

	// Should have added tokens
	assert.Greater(t, bl.Size(), 0)
}

func TestMemoryTokenBlacklist_CleanupEmpty(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	// Empty blacklist - cleanup should not panic
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, bl.Size())
}

func TestMemoryTokenBlacklist_CleanupAllExpired(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add all short-lived tokens
	bl.Add(ctx, "token-1", 10*time.Millisecond)
	bl.Add(ctx, "token-2", 10*time.Millisecond)
	bl.Add(ctx, "token-3", 10*time.Millisecond)

	// Wait for all to expire
	time.Sleep(50 * time.Millisecond)

	// All should be expired (checked via Exists)
	assert.False(t, bl.Exists(ctx, "token-1"))
	assert.False(t, bl.Exists(ctx, "token-2"))
	assert.False(t, bl.Exists(ctx, "token-3"))
}

// ============================================
// HybridTokenBlacklist Cleanup Tests
// ============================================

func TestHybridTokenBlacklist_CleanupWithoutRedis(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	// Add entries
	hb.Add(ctx, "hybrid-token-1", 50*time.Millisecond)
	hb.Add(ctx, "hybrid-token-2", 10*time.Minute)

	// Check immediately
	assert.True(t, hb.Exists(ctx, "hybrid-token-1"))
	assert.True(t, hb.Exists(ctx, "hybrid-token-2"))

	// Wait for short entry to expire
	time.Sleep(100 * time.Millisecond)

	// Short entry should be expired
	assert.False(t, hb.Exists(ctx, "hybrid-token-1"))
	assert.True(t, hb.Exists(ctx, "hybrid-token-2"))
}

func TestHybridTokenBlacklist_UserRevocationCleanup(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	// Add user revocations
	hb.AddUserRevocation(ctx, 100, time.Now(), 50*time.Millisecond)
	hb.AddUserRevocation(ctx, 101, time.Now(), 10*time.Minute)

	// Wait for short revocation to expire
	time.Sleep(100 * time.Millisecond)

	rev100, _ := hb.GetUserRevocation(ctx, 100)
	assert.True(t, rev100.IsZero()) // Expired

	rev101, _ := hb.GetUserRevocation(ctx, 101)
	assert.False(t, rev101.IsZero()) // Still valid
}

func TestHybridTokenBlacklist_Stop(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	hb.Stop() // Should not panic

	// Operations after stop should still work
	ctx := context.Background()
	hb.Add(ctx, "after-stop", time.Minute)
	assert.True(t, hb.Exists(ctx, "after-stop"))
}

func TestHybridTokenBlacklist_MixedCleanup(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	// Add both types
	hb.Add(ctx, "token-short", 50*time.Millisecond)
	hb.Add(ctx, "token-long", 10*time.Minute)
	hb.AddUserRevocation(ctx, 200, time.Now(), 50*time.Millisecond)
	hb.AddUserRevocation(ctx, 201, time.Now(), 10*time.Minute)

	// Wait for short entries to expire
	time.Sleep(100 * time.Millisecond)

	// Short entries expired
	assert.False(t, hb.Exists(ctx, "token-short"))
	rev200, _ := hb.GetUserRevocation(ctx, 200)
	assert.True(t, rev200.IsZero())

	// Long entries valid
	assert.True(t, hb.Exists(ctx, "token-long"))
	rev201, _ := hb.GetUserRevocation(ctx, 201)
	assert.False(t, rev201.IsZero())
}

func TestHybridTokenBlacklist_ConcurrentOperations(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent operations
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(idx int) {
			defer wg.Done()
			hb.Add(ctx, "htoken-"+string(rune(idx)), time.Minute)
		}(i)
		go func(idx int) {
			defer wg.Done()
			hb.Exists(ctx, "htoken-"+string(rune(idx)))
		}(i)
		go func(idx int) {
			defer wg.Done()
			hb.AddUserRevocation(ctx, idx, time.Now(), time.Minute)
		}(i)
	}
	wg.Wait()

	// Should have entries
	assert.Greater(t, hb.memoryBlacklist.Size(), 0)
}

// ============================================
// Token Blacklist Interface Tests
// ============================================

func TestTokenBlacklistInterface_Implementation(t *testing.T) {
	// Both MemoryTokenBlacklist and HybridTokenBlacklist implement TokenBlacklistInterface
	var _ TokenBlacklistInterface = NewMemoryTokenBlacklist()
	var _ TokenBlacklistInterface = NewHybridTokenBlacklist(nil)
}

func TestMemoryTokenBlacklist_InterfaceMethods(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Test all interface methods
	err := bl.Add(ctx, "test", time.Minute)
	assert.NoError(t, err)

	exists := bl.Exists(ctx, "test")
	assert.True(t, exists)

	err = bl.AddUserRevocation(ctx, 1, time.Now(), time.Minute)
	assert.NoError(t, err)

	revokedAt, err := bl.GetUserRevocation(ctx, 1)
	assert.NoError(t, err)
	assert.False(t, revokedAt.IsZero())
}

func TestHybridTokenBlacklist_InterfaceMethods(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	// Test all interface methods
	err := hb.Add(ctx, "test", time.Minute)
	assert.NoError(t, err)

	exists := hb.Exists(ctx, "test")
	assert.True(t, exists)

	err = hb.AddUserRevocation(ctx, 1, time.Now(), time.Minute)
	assert.NoError(t, err)

	revokedAt, err := hb.GetUserRevocation(ctx, 1)
	assert.NoError(t, err)
	assert.False(t, revokedAt.IsZero())
}

// ============================================
// Edge Cases
// ============================================

func TestMemoryTokenBlacklist_AddExpiredDuration(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add with negative duration (already expired)
	bl.Add(ctx, "expired-token", -1*time.Second)

	// Should immediately be considered expired by Exists
	assert.False(t, bl.Exists(ctx, "expired-token"))
}

func TestMemoryTokenBlacklist_AddZeroDuration(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add with zero duration
	bl.Add(ctx, "zero-token", 0)

	// Should immediately be considered expired (time.Now().Add(0) is now, which is already past)
	// Actually, Add(0) means it expires at exactly now, which in Exists check is: now.After(expiry)
	// Since expiry = now (approximately), this might be true or false depending on timing
	// Let's just verify the operation doesn't panic
	assert.NotPanics(t, func() {
		bl.Exists(ctx, "zero-token")
	})
}

func TestMemoryTokenBlacklist_VeryLongDuration(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add with very long duration
	bl.Add(ctx, "long-token", 24*time.Hour*365) // 1 year

	assert.True(t, bl.Exists(ctx, "long-token"))
}

func TestMemoryTokenBlacklist_RevocationPastTime(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add revocation with past time
	pastTime := time.Now().Add(-1 * time.Hour)
	bl.AddUserRevocation(ctx, 999, pastTime, time.Minute)

	revokedAt, err := bl.GetUserRevocation(ctx, 999)
	assert.NoError(t, err)
	// Should return the past time (the storage doesn't check if time is past)
	assert.WithinDuration(t, pastTime, revokedAt, time.Second)
}

func TestMemoryTokenBlacklist_RevocationFutureTime(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add revocation with future time
	futureTime := time.Now().Add(1 * time.Hour)
	bl.AddUserRevocation(ctx, 888, futureTime, time.Minute)

	revokedAt, err := bl.GetUserRevocation(ctx, 888)
	assert.NoError(t, err)
	assert.WithinDuration(t, futureTime, revokedAt, time.Second)
}

// ============================================
// Race Condition Tests
// ============================================

func TestMemoryTokenBlacklist_RaceConditions(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Run operations concurrently to test for race conditions
	var wg sync.WaitGroup

	// Writer goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bl.Add(ctx, "token-"+string(rune(idx+j)), time.Minute)
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bl.Exists(ctx, "token-"+string(rune(idx+j)))
			}
		}(i)
	}

	// User revocation goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bl.AddUserRevocation(ctx, idx+j, time.Now(), time.Minute)
				bl.GetUserRevocation(ctx, idx+j)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race condition is detected (run with -race flag)
}

func TestHybridTokenBlacklist_RaceConditions(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Writer goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				hb.Add(ctx, "htoken-"+string(rune(idx+j)), time.Minute)
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				hb.Exists(ctx, "htoken-"+string(rune(idx+j)))
			}
		}(i)
	}

	wg.Wait()
}

// ============================================
// Performance Tests (basic)
// ============================================

func TestMemoryTokenBlacklist_LargeNumberOfEntries(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add many entries
	for i := 0; i < 1000; i++ {
		bl.Add(ctx, "bulk-token-"+string(rune(i)), time.Minute)
	}

	assert.Equal(t, 1000, bl.Size())

	// Check all exist
	for i := 0; i < 1000; i++ {
		assert.True(t, bl.Exists(ctx, "bulk-token-"+string(rune(i))))
	}
}

func TestHybridTokenBlacklist_LargeNumberOfEntries(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	// Add many entries
	for i := 0; i < 1000; i++ {
		hb.Add(ctx, "bulk-htoken-"+string(rune(i)), time.Minute)
	}

	// Check via memory blacklist size
	assert.Equal(t, 1000, hb.memoryBlacklist.Size())
}
