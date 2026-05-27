package service

import (
	"context"
	"testing"
	"time"
)

// TestMemoryTokenBlacklist_Add_Eviction tests LRU eviction when maxEntries reached
func TestMemoryTokenBlacklist_Add_Eviction(t *testing.T) {
	// Create blacklist with maxEntries=3
	blacklist := NewMemoryTokenBlacklistWithLimit(3)
	ctx := context.Background()
	
	// Add 3 tokens (fill to max)
	for i := 1; i <= 3; i++ {
		tokenID := string(rune('A' + i - 1))
		err := blacklist.Add(ctx, tokenID, time.Hour)
		if err != nil {
			t.Errorf("Add token %d failed: %v", i, err)
		}
	}
	
	// Add 4th token - should trigger eviction
	token4 := "D"
	err := blacklist.Add(ctx, token4, time.Hour)
	if err != nil {
		t.Errorf("Add token 4 failed: %v", err)
	}
	
	// Verify oldest token was evicted
	// Token "A" should be evicted (oldest)
	exists := blacklist.Exists(ctx, "A")
	if exists {
		t.Error("Oldest token 'A' should have been evicted")
	}
	
	// Verify new token exists
	exists = blacklist.Exists(ctx, "D")
	if !exists {
		t.Error("New token 'D' should exist")
	}
	
	// Verify total entries still = maxEntries
	blacklist.mu.RLock()
	entryCount := len(blacklist.entries)
	blacklist.mu.RUnlock()
	if entryCount != 3 {
		t.Errorf("Expected 3 entries, got %d", entryCount)
	}
}

// TestMemoryTokenBlacklist_Add_NoEviction tests normal add without eviction
func TestMemoryTokenBlacklist_Add_NoEviction(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(10)
	ctx := context.Background()
	
	// Add token below maxEntries
	tokenID := "test_token_123"
	err := blacklist.Add(ctx, tokenID, time.Hour)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}
	
	// Verify token exists
	exists := blacklist.Exists(ctx, tokenID)
	if !exists {
		t.Error("Token should exist")
	}
}

// TestMemoryTokenBlacklist_Add_MultipleEvictions tests multiple eviction cycles
func TestMemoryTokenBlacklist_Add_MultipleEvictions(t *testing.T) {
	blacklist := NewMemoryTokenBlacklistWithLimit(2)
	ctx := context.Background()
	
	// Add 5 tokens - should trigger 3 evictions
	for i := 1; i <= 5; i++ {
		tokenID := string(rune('A' + i - 1))
		err := blacklist.Add(ctx, tokenID, time.Hour)
		if err != nil {
			t.Errorf("Add token %d failed: %v", i, err)
		}
		
		// Verify eviction happened for i > 2
		if i > 2 {
			// Oldest token should be evicted
			evictedToken := string(rune('A' + i - 3))
			exists := blacklist.Exists(ctx, evictedToken)
			if exists {
				t.Errorf("Token '%s' should have been evicted", evictedToken)
			}
		}
	}
	
	// Verify only 2 entries remain (maxEntries)
	blacklist.mu.RLock()
	entryCount := len(blacklist.entries)
	blacklist.mu.RUnlock()
	if entryCount != 2 {
		t.Errorf("Expected 2 entries, got %d", entryCount)
	}
	
	// Verify latest 2 tokens exist
	exists := blacklist.Exists(ctx, "D")
	if !exists {
		t.Error("Token 'D' should exist")
	}
	exists = blacklist.Exists(ctx, "E")
	if !exists {
		t.Error("Token 'E' should exist")
	}
}
