package service

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// TestHealthService_CheckHealth tests the health check functionality
func TestHealthService_CheckHealth(t *testing.T) {
	// Create a mock database connection (using nil for testing without actual DB)
	// In production, you would use a test database
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Initialize health service
	healthSvc := InitHealthService(db, "1.0.0-test")

	// Perform health check
	ctx := context.Background()
	result := healthSvc.CheckHealth(ctx)

	// Verify response structure
	if result.Version != "1.0.0-test" {
		t.Errorf("Expected version 1.0.0-test, got %s", result.Version)
	}

	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if result.Uptime == "" {
		t.Error("Expected uptime to be set")
	}

	// Check that components are present
	if result.Components.Database.Status == "" {
		t.Error("Expected database status to be set")
	}

	if result.Components.Memory.Status == "" {
		t.Error("Expected memory status to be set")
	}

	if result.Components.Disk.Status == "" {
		t.Error("Expected disk status to be set")
	}

	// Check that LLM API status is present
	if result.Components.LLMAPI.Status == "" {
		t.Error("Expected LLM API status to be set")
	}

	// Verify that overall status is one of the expected values
	validStatuses := map[string]bool{
		"healthy":   true,
		"unhealthy": true,
		"degraded":  true,
	}

	if !validStatuses[result.Status] {
		t.Errorf("Invalid status: %s", result.Status)
	}

	t.Logf("Health check result: Status=%s, Version=%s, Uptime=%s", result.Status, result.Version, result.Uptime)
	t.Logf("Database: %s - %s (latency: %dms)", result.Components.Database.Status, result.Components.Database.Message, result.Components.Database.LatencyMs)
	t.Logf("LLM API: %s - %s", result.Components.LLMAPI.Status, result.Components.LLMAPI.Message)
	t.Logf("Memory: %s - %s", result.Components.Memory.Status, result.Components.Memory.Message)
	t.Logf("Disk: %s - %s", result.Components.Disk.Status, result.Components.Disk.Message)
}

// TestFormatUptime tests the uptime formatting function
func TestFormatUptime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{time.Second * 30, "0h0m30s"},
		{time.Minute * 5, "0h5m0s"},
		{time.Hour * 2, "2h0m0s"},
		{time.Hour*24 + time.Hour*2 + time.Minute*30, "1d2h30m"},
		{time.Hour * 50, "2d2h0m"},
	}

	for _, test := range tests {
		result := formatUptime(test.duration)
		if result != test.expected {
			t.Errorf("For duration %v, expected %s, got %s", test.duration, test.expected, result)
		}
	}
}

// TestHealthStatusTransitions tests that health status transitions correctly
func TestHealthStatusTransitions(t *testing.T) {
	// This test would require a real database connection to test fully
	// For now, we'll just verify the logic works with nil values

	healthSvc := &HealthService{
		db:           nil,
		startTime:    time.Now(),
		version:      "test",
		llmAPIKey:    "",
		llmBaseURL:   "",
		llmModel:     "",
		checkTimeout: 5 * time.Second,
	}

	// Test memory check (should always work)
	memStatus := healthSvc.checkMemory()
	if memStatus.Status == "" {
		t.Error("Memory check should return a status")
	}
	t.Logf("Memory status: %s - %s", memStatus.Status, memStatus.Message)

	// Test disk check (should work)
	diskStatus := healthSvc.checkDisk()
	if diskStatus.Status == "" {
		t.Error("Disk check should return a status")
	}
	t.Logf("Disk status: %s - %s", diskStatus.Status, diskStatus.Message)

	// Test LLM API check (should return unavailable when no API key)
	llmStatus := healthSvc.checkLLMAPI(context.Background())
	if llmStatus.Status != "unavailable" {
		t.Errorf("Expected LLM status to be 'unavailable' when no API key, got %s", llmStatus.Status)
	}
	t.Logf("LLM status: %s - %s", llmStatus.Status, llmStatus.Message)
}
