package logger

import (
	"testing"
	"go.uber.org/zap/zapcore"
)

// TestLogger_SetLevel tests Logger.SetLevel method
func TestLogger_SetLevel(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	// Test setting different levels
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		err := logger.SetLevel(level)
		if err != nil {
			t.Errorf("SetLevel(%s) failed: %v", level, err)
		}
		current := logger.GetLevel()
		if current != level {
			t.Errorf("SetLevel(%s) failed, got %s", level, current)
		}
	}
}

// TestLogger_GetLevel tests Logger.GetLevel method
func TestLogger_GetLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "info"
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	level := logger.GetLevel()
	if level != "info" {
		t.Errorf("GetLevel failed, expected 'info', got '%s'", level)
	}
}

// TestLogger_GetLevelZap tests Logger.GetLevelZap method
func TestLogger_GetLevelZap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "warn"
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	zapLevel := logger.GetLevelZap()
	if zapLevel != zapcore.WarnLevel {
		t.Errorf("GetLevelZap failed, expected zapcore.WarnLevel, got %v", zapLevel)
	}
}

// TestLogger_SetFormat tests Logger.SetFormat method
func TestLogger_SetFormat(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	// Test setting different formats
	formats := []string{"json", "console"}
	for _, format := range formats {
		logger.SetFormat(format)
		current := logger.GetFormat()
		if current != format {
			t.Errorf("SetFormat(%s) failed, got %s", format, current)
		}
	}
}

// TestLogger_GetFormat tests Logger.GetFormat method
func TestLogger_GetFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Format = "json"
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	format := logger.GetFormat()
	if format != "json" {
		t.Errorf("GetFormat failed, expected 'json', got '%s'", format)
	}
}

// TestLogger_GetConfig tests Logger.GetConfig method
func TestLogger_GetConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "info"
	cfg.Format = "json"
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	
	config := logger.GetConfig()
	if config.Level != "info" {
		t.Errorf("GetConfig level failed, expected 'info', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("GetConfig format failed, expected 'json', got '%s'", config.Format)
	}
}

// TestSetGlobalLevel tests SetGlobalLevel function
func TestSetGlobalLevel(t *testing.T) {
	// Initialize global logger
	cfg := DefaultConfig()
	InitGlobalLogger(cfg)
	
	// Test setting global level
	err := SetGlobalLevel("error")
	if err != nil {
		t.Errorf("SetGlobalLevel failed: %v", err)
	}
	
	level := GetGlobalLevel()
	if level != "error" {
		t.Errorf("SetGlobalLevel failed, expected 'error', got '%s'", level)
	}
}

// TestGetGlobalLevel tests GetGlobalLevel function
func TestGetGlobalLevel(t *testing.T) {
	// Initialize global logger
	cfg := DefaultConfig()
	InitGlobalLogger(cfg)
	
	SetGlobalLevel("warn")
	level := GetGlobalLevel()
	if level != "warn" {
		t.Errorf("GetGlobalLevel failed, expected 'warn', got '%s'", level)
	}
}

// TestReloadGlobalLogger tests ReloadGlobalLogger function
func TestReloadGlobalLogger(t *testing.T) {
	// Initialize global logger
	cfg1 := DefaultConfig()
	cfg1.Level = "info"
	InitGlobalLogger(cfg1)
	
	// Reload with new config
	cfg2 := DefaultConfig()
	cfg2.Level = "debug"
	cfg2.Format = "console"
	
	err := ReloadGlobalLogger(cfg2)
	if err != nil {
		t.Errorf("ReloadGlobalLogger failed: %v", err)
	}
	
	// Verify new config
	level := GetGlobalLevel()
	if level != "debug" {
		t.Errorf("ReloadGlobalLogger failed, level should be 'debug', got '%s'", level)
	}
}
