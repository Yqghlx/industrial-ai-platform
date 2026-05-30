package service

import (
	"context"
	"testing"
	"time"
)

// TestMemoryTokenBlacklist_Add_Eviction 测试达到上限时淘汰条目
func TestMemoryTokenBlacklist_Add_Eviction(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(3)
	ctx := context.Background()

	// 添加 3 个 token（填满）
	for i := 1; i <= 3; i++ {
		tokenID := string(rune('A' + i - 1))
		err := blacklist.Add(ctx, tokenID, time.Hour)
		if err != nil {
			t.Errorf("Add token %d failed: %v", i, err)
		}
	}

	// 添加第 4 个 - 应触发淘汰
	err := blacklist.Add(ctx, "D", time.Hour)
	if err != nil {
		t.Errorf("Add token 4 failed: %v", err)
	}

	// 验证新 token 存在
	exists := blacklist.Exists(ctx, "D")
	if !exists {
		t.Error("New token 'D' should exist")
	}

	// 验证条目数不超过 maxEntries
	blacklist.mu.RLock()
	entryCount := len(blacklist.entries)
	blacklist.mu.RUnlock()
	if entryCount > 3 {
		t.Errorf("Expected at most 3 entries, got %d", entryCount)
	}
}

// TestMemoryTokenBlacklist_Add_NoEviction 测试正常添加不触发淘汰
func TestMemoryTokenBlacklist_Add_NoEviction(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(10)
	ctx := context.Background()

	tokenID := "test_token_123"
	err := blacklist.Add(ctx, tokenID, time.Hour)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	exists := blacklist.Exists(ctx, tokenID)
	if !exists {
		t.Error("Token should exist")
	}
}

// TestMemoryTokenBlacklist_Add_MultipleEvictions 测试多次淘汰后条目数不超限
func TestMemoryTokenBlacklist_Add_MultipleEvictions(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(2)
	ctx := context.Background()

	// 添加 5 个 token - 应触发 3 次淘汰
	for i := 1; i <= 5; i++ {
		tokenID := string(rune('A' + i - 1))
		err := blacklist.Add(ctx, tokenID, time.Hour)
		if err != nil {
			t.Errorf("Add token %d failed: %v", i, err)
		}
	}

	// 验证条目数不超过 maxEntries
	blacklist.mu.RLock()
	entryCount := len(blacklist.entries)
	blacklist.mu.RUnlock()
	if entryCount > 2 {
		t.Errorf("Expected at most 2 entries, got %d", entryCount)
	}

	// 验证最新 token 存在
	exists := blacklist.Exists(ctx, "E")
	if !exists {
		t.Error("Latest token 'E' should exist")
	}
}

// TestMemoryTokenBlacklist_Add_ExpiredEntryEviction 测试优先淘汰过期条目
func TestMemoryTokenBlacklist_Add_ExpiredEntryEviction(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(2)
	ctx := context.Background()

	// 添加一个即将过期的 token
	blacklist.Add(ctx, "expiring", 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond) // 等待过期

	// 添加一个正常 token
	blacklist.Add(ctx, "normal", time.Hour)

	// 添加第三个 - 应淘汰过期的 "expiring"
	blacklist.Add(ctx, "new", time.Hour)

	// "normal" 和 "new" 应该存在
	if !blacklist.Exists(ctx, "normal") {
		t.Error("Token 'normal' should exist")
	}
	if !blacklist.Exists(ctx, "new") {
		t.Error("Token 'new' should exist")
	}
}
