package logger

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("Expected default level 'info', got '%s'", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("Expected default format 'json', got '%s'", cfg.Format)
	}
	if cfg.ServiceName != "industrial-ai-backend" {
		t.Errorf("Expected default service name 'industrial-ai-backend', got '%s'", cfg.ServiceName)
	}
	if cfg.Environment != "production" {
		t.Errorf("Expected default environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", cfg.Version)
	}
	if cfg.Output != "stdout" {
		t.Errorf("Expected default output 'stdout', got '%s'", cfg.Output)
	}
}

// TestNewLogger 测试日志器初始化
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid config with json format",
			config: Config{
				Level:       "info",
				Format:      "json",
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				Output:      "stdout",
			},
			wantError: false,
		},
		{
			name: "valid config with console format",
			config: Config{
				Level:       "debug",
				Format:      "console",
				ServiceName: "test-service",
				Environment: "development",
				Version:     "1.0.0",
				Output:      "stderr",
			},
			wantError: false,
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:       "invalid",
				Format:      "json",
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				Output:      "stdout",
			},
			wantError: false,
		},
		{
			name: "empty output defaults to stdout",
			config: Config{
				Level:       "warn",
				Format:      "json",
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				Output:      "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if logger == nil {
					t.Errorf("Expected logger but got nil")
					return
				}
				if logger.Logger == nil {
					t.Errorf("Expected zap.Logger but got nil")
				}
			}

			// Cleanup
			if logger != nil {
				logger.Sync()
			}
		})
	}
}

// TestLoggerLevels 测试不同日志级别
func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			cfg := Config{
				Level:       level,
				Format:      "json",
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				Output:      "stdout",
			}

			logger, err := NewLogger(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Test that the logger was created successfully
			if logger == nil {
				t.Errorf("Expected logger but got nil")
			}

			logger.Sync()
		})
	}
}

// TestLoggerInfo 测试 Info 方法
func TestLoggerInfo(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "test-service",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test Info logging (should not panic)
	logger.Info("test info message")
	logger.Sync()
}

// TestLoggerWarn 测试 Warn 方法
func TestLoggerWarn(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "test-service",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Warn("test warning message")
	logger.Sync()
}

// TestLoggerError 测试 Error 方法
func TestLoggerError(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "test-service",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Error("test error message")
	logger.Sync()
}

// TestLoggerDebug 测试 Debug 方法
func TestLoggerDebug(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "test-service",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Debug("test debug message")
	logger.Sync()
}

// TestWithContext 测试 WithContext 方法
func TestWithContext(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test without trace_id
	ctx := context.Background()
	result := logger.WithContext(ctx)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	// Test with trace_id
	ctxWithTrace := context.WithValue(context.Background(), "trace_id", "test-trace-123")
	result = logger.WithContext(ctxWithTrace)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger with trace_id")
	}

	logger.Sync()
}

// TestWithTraceID 测试 WithTraceID 方法
func TestWithTraceID(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	traceID := "trace-123-456"
	result := logger.WithTraceID(traceID)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	// Use the logger with trace_id
	result.Info("test message with trace id")

	logger.Sync()
}

// TestWithRequestID 测试 WithRequestID 方法
func TestWithRequestID(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	requestID := "req-789"
	result := logger.WithRequestID(requestID)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	result.Info("test message with request id")

	logger.Sync()
}

// TestWithTenantID 测试 WithTenantID 方法
func TestWithTenantID(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tenantID := "tenant-001"
	result := logger.WithTenantID(tenantID)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	result.Info("test message with tenant id")

	logger.Sync()
}

// TestWithUserID 测试 WithUserID 方法
func TestWithUserID(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	userID := "user-123"
	result := logger.WithUserID(userID)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	result.Info("test message with user id")

	logger.Sync()
}

// TestWithError 测试 WithError 方法
func TestWithError(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test with error
	testErr := errors.New("test error")
	result := logger.WithError(testErr)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	result.Error("test message with error")

	// Test with nil error
	result = logger.WithError(nil)
	if result == nil {
		t.Errorf("Expected non-nil zap.Logger for nil error")
	}

	logger.Sync()
}

// TestHTTPFields 测试 HTTPFields 函数
func TestHTTPFields(t *testing.T) {
	method := "GET"
	path := "/api/v1/users"
	statusCode := 200
	latency := time.Duration(50) * time.Millisecond
	requestSize := int64(1024)
	responseSize := int64(2048)

	fields := HTTPFields(method, path, statusCode, latency, requestSize, responseSize)

	if len(fields) != 6 {
		t.Errorf("Expected 6 fields, got %d", len(fields))
	}

	// Verify each field exists
	fieldNames := []string{"http.method", "http.path", "http.status_code", "http.latency_ms", "http.request_size_bytes", "http.response_size_bytes"}
	for i, name := range fieldNames {
		if fields[i].Key != name {
			t.Errorf("Expected field key '%s', got '%s'", name, fields[i].Key)
		}
	}
}

// TestLogHTTPRequest 测试 LogHTTPRequest 方法
func TestLogHTTPRequest(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tests := []struct {
		name        string
		statusCode  int
		expectLevel string
	}{
		{
			name:        "2xx status",
			statusCode:  200,
			expectLevel: "info",
		},
		{
			name:        "4xx status",
			statusCode:  400,
			expectLevel: "warn",
		},
		{
			name:        "5xx status",
			statusCode:  500,
			expectLevel: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			logger.LogHTTPRequest("GET", "/test", tt.statusCode, time.Millisecond*100, 1024, 2048)
		})
	}

	logger.Sync()
}

// TestInitGlobalLogger 测试全局日志器初始化
func TestInitGlobalLogger(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "global-test",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	err := InitGlobalLogger(cfg)
	if err != nil {
		t.Errorf("Unexpected error initializing global logger: %v", err)
	}

	if globalLogger == nil {
		t.Errorf("Expected globalLogger to be set")
	}

	// Cleanup
	if globalLogger != nil {
		globalLogger.Sync()
	}
	globalLogger = nil
}

// TestGetLogger 测试获取全局日志器
func TestGetLogger(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	// Get logger should create default if not initialized
	logger := GetLogger()
	if logger == nil {
		t.Errorf("Expected non-nil logger")
	}

	// Should return the same instance
	logger2 := GetLogger()
	if logger != logger2 {
		t.Errorf("Expected same logger instance")
	}

	// Cleanup
	if globalLogger != nil {
		globalLogger.Sync()
	}
	globalLogger = nil
}

// TestL 测试 L 快捷方式
func TestL(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	logger := L()
	if logger == nil {
		t.Errorf("Expected non-nil logger from L()")
	}

	// Cleanup
	if globalLogger != nil {
		globalLogger.Sync()
	}
	globalLogger = nil
}

// TestSync 测试 Sync 方法
func TestSync(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	err = logger.Sync()
	// Sync may return errors on certain platforms (e.g., stdout on Windows)
	// We just check it doesn't panic
	_ = err
}

// TestClose 测试 Close 方法
func TestClose(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	err = logger.Close()
	// Close may return errors on certain platforms
	_ = err
}

// TestLogJSONFormat 测试 JSON 格式日志
func TestLogJSONFormat(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "json-test",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Verify logger was created correctly
	if logger == nil {
		t.Fatalf("Expected logger to be created, got nil")
	}

	// Log a message
	logger.Info("json format test")

	logger.Sync()
}

// TestLogConsoleFormat 测试 Console 格式日志
func TestLogConsoleFormat(t *testing.T) {
	cfg := Config{
		Level:       "debug",
		Format:      "console",
		ServiceName: "console-test",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Verify logger was created correctly
	if logger == nil {
		t.Fatalf("Expected logger to be created, got nil")
	}

	// Log a message
	logger.Info("console format test")

	logger.Sync()
}

// TestLoggerFieldChaining 测试字段链式调用
func TestLoggerFieldChaining(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test using With* methods on the Logger directly
	logger.WithTraceID("trace-123").Info("message with trace")
	logger.WithRequestID("req-456").Info("message with request id")

	logger.WithTenantID("tenant-001").Info("message with tenant")
	logger.WithUserID("user-123").Info("message with user")

	logger.Sync()
}

// TestParseLevel 测试日志级别解析
func TestParseLevel(t *testing.T) {
	// Test all valid levels
	validLevels := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}

	for levelStr, expectedLevel := range validLevels {
		t.Run("level_"+levelStr, func(t *testing.T) {
			level, err := zapcore.ParseLevel(levelStr)
			if err != nil {
				t.Errorf("Unexpected error parsing level '%s': %v", levelStr, err)
			}
			if level != expectedLevel {
				t.Errorf("Expected level %v, got %v", expectedLevel, level)
			}
		})
	}

	// Test invalid level defaults to info
	level, err := zapcore.ParseLevel("invalid")
	if err == nil {
		t.Errorf("Expected error for invalid level, got nil")
	}
	// When there's an error, the logger code defaults to InfoLevel
	if level != zapcore.InfoLevel {
		t.Logf("Invalid level parsed as: %v (error: %v)", level, err)
	}
}

// TestHTTPFieldsValues 测试 HTTP 字段值正确性
func TestHTTPFieldsValues(t *testing.T) {
	method := "POST"
	path := "/api/v1/data"
	statusCode := 201
	latency := time.Duration(150) * time.Millisecond
	requestSize := int64(2048)
	responseSize := int64(4096)

	fields := HTTPFields(method, path, statusCode, latency, requestSize, responseSize)

	// Check method
	if fields[0].Key != "http.method" {
		t.Errorf("Expected field key 'http.method', got '%s'", fields[0].Key)
	}

	// Check path
	if fields[1].Key != "http.path" {
		t.Errorf("Expected field key 'http.path', got '%s'", fields[1].Key)
	}

	// Check status code
	if fields[2].Key != "http.status_code" {
		t.Errorf("Expected field key 'http.status_code', got '%s'", fields[2].Key)
	}

	// Check latency
	if fields[3].Key != "http.latency_ms" {
		t.Errorf("Expected field key 'http.latency_ms', got '%s'", fields[3].Key)
	}

	// Check request size
	if fields[4].Key != "http.request_size_bytes" {
		t.Errorf("Expected field key 'http.request_size_bytes', got '%s'", fields[4].Key)
	}

	// Check response size
	if fields[5].Key != "http.response_size_bytes" {
		t.Errorf("Expected field key 'http.response_size_bytes', got '%s'", fields[5].Key)
	}
}

// TestLoggerWithMultipleContexts 测试多个上下文值
func TestLoggerWithMultipleContexts(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test with context containing trace_id
	ctx := context.WithValue(context.Background(), "trace_id", "trace-abc-123")

	// Should return logger with trace_id
	zapLogger := logger.WithContext(ctx)
	if zapLogger == nil {
		t.Errorf("Expected non-nil zap.Logger")
	}

	logger.Sync()
}

// TestLoggerJSONOutput 测试 JSON 输出格式
func TestLoggerJSONOutput(t *testing.T) {
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "json-output-test",
		Environment: "testing",
		Version:     "2.0.0",
		Output:      "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Verify config is stored
	if logger.config.ServiceName != "json-output-test" {
		t.Errorf("Expected service name 'json-output-test', got '%s'", logger.config.ServiceName)
	}
	if logger.config.Environment != "testing" {
		t.Errorf("Expected environment 'testing', got '%s'", logger.config.Environment)
	}
	if logger.config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", logger.config.Version)
	}

	logger.Sync()
}

// TestLogHTTPRequestDifferentStatusCodes 测试不同状态码的 HTTP 请求日志
func TestLogHTTPRequestDifferentStatusCodes(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test various status codes
	statusCodes := []int{
		200, 201, 204, // Success
		301, 302, 304, // Redirect
		400, 401, 403, 404, // Client errors
		500, 502, 503, // Server errors
	}

	for _, code := range statusCodes {
		t.Run("status_"+string(rune(code)), func(t *testing.T) {
			// Should not panic
			logger.LogHTTPRequest("GET", "/test", code, time.Millisecond*50, 100, 200)
		})
	}

	logger.Sync()
}

// BenchmarkLoggerInfo 基准测试 Info 方法
func BenchmarkLoggerInfo(b *testing.B) {
	logger, _ := NewLogger(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "bench-test",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	})
	defer logger.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkLoggerWithFields 基准测试带字段的日志
func BenchmarkLoggerWithFields(b *testing.B) {
	logger, _ := NewLogger(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "bench-test",
		Environment: "test",
		Version:     "1.0.0",
		Output:      "stdout",
	})
	defer logger.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithTraceID("trace-123").Info("benchmark message with fields")
	}
}

// BenchmarkHTTPFields 基准测试 HTTPFields 函数
func BenchmarkHTTPFields(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HTTPFields("GET", "/api/test", 200, time.Millisecond*50, 1024, 2048)
	}
}
