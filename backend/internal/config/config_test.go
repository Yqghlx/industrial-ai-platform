package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errFields []string // expected error fields
	}{
		{
			name: "valid config with all required fields",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				JWTSecret:   "secret",
				Port:        "8080",
			},
			wantErr: false,
		},
		{
			name: "missing DATABASE_URL",
			config: &Config{
				JWTSecret: "secret",
				Port:      "8080",
			},
			wantErr:   true,
			errFields: []string{"DATABASE_URL"},
		},
		{
			name: "missing JWT_SECRET in production",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				Environment: "production",
			},
			wantErr:   true,
			errFields: []string{"JWT_SECRET"},
		},
		{
			name: "missing JWT_SECRET not an error in development",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				Environment: "development",
			},
			wantErr: false,
		},
		{
			name: "empty config defaults port",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
			},
			wantErr: false,
		},
		{
			name: "invalid port number",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				Port:        "x",
			},
			wantErr:   true,
			errFields: []string{"PORT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error, got nil")
					return
				}

				// Check that expected fields are in error
				errs, ok := err.(ValidationErrors)
				if !ok {
					t.Errorf("Expected ValidationErrors, got %T", err)
					return
				}

				for _, field := range tt.errFields {
					found := false
					for _, e := range errs {
						if e.Field == field {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field %s, not found", field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfig_DefaultPort(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		Port:        "",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	_ = os.Setenv("JWT_SECRET", "test-secret")
	_ = os.Setenv("PORT", "9090")
	_ = os.Setenv("LLM_API_KEY", "test-key")
	_ = os.Setenv("LLM_BASE_URL", "https://api.example.com")
	_ = os.Setenv("LLM_MODEL", "gpt-4")
	_ = os.Setenv("CORS_ORIGINS", "http://localhost:3000")
	_ = os.Setenv("ADMIN_PASSWORD", "admin123")
	_ = os.Setenv("ENV", "development")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
		os.Unsetenv("LLM_MODEL")
		os.Unsetenv("CORS_ORIGINS")
		os.Unsetenv("ADMIN_PASSWORD")
		os.Unsetenv("ENV")
	}()

	cfg := LoadFromEnv()

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/testdb" {
		t.Errorf("Expected DATABASE_URL, got %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("Expected JWT_SECRET, got %s", cfg.JWTSecret)
	}
	if cfg.Port != "9090" {
		t.Errorf("Expected PORT 9090, got %s", cfg.Port)
	}
	if cfg.LLMAPIKey != "test-key" {
		t.Errorf("Expected LLM_API_KEY, got %s", cfg.LLMAPIKey)
	}
	if cfg.LLMBaseURL != "https://api.example.com" {
		t.Errorf("Expected LLM_BASE_URL, got %s", cfg.LLMBaseURL)
	}
	if cfg.LLMModel != "gpt-4" {
		t.Errorf("Expected LLM_MODEL, got %s", cfg.LLMModel)
	}
	if cfg.CORSOrigins != "http://localhost:3000" {
		t.Errorf("Expected CORS_ORIGINS, got %s", cfg.CORSOrigins)
	}
	if cfg.AdminPassword != "admin123" {
		t.Errorf("Expected ADMIN_PASSWORD, got %s", cfg.AdminPassword)
	}
	if cfg.Environment != "development" {
		t.Errorf("Expected ENV development, got %s", cfg.Environment)
	}
}

func TestConfig_GetCORSOrigins(t *testing.T) {
	tests := []struct {
		name     string
		origins  string
		expected []string
	}{
		{
			name:     "empty returns empty slice (security)",
			origins:  "",
			expected: []string{},
		},
		{
			name:     "single origin",
			origins:  "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "multiple origins",
			origins:  "http://localhost:3000,http://localhost:8080",
			expected: []string{"http://localhost:3000", "http://localhost:8080"},
		},
		{
			name:     "origins with spaces",
			origins:  "http://localhost:3000 , http://localhost:8080",
			expected: []string{"http://localhost:3000", "http://localhost:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{CORSOrigins: tt.origins}
			result := cfg.GetCORSOrigins()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d origins, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected origin %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestConfig_GetWarnings(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedCount int
	}{
		{
			name: "no warnings when all set",
			config: &Config{
				JWTSecret:     "secret",
				AdminPassword: "admin123",
				LLMAPIKey:     "key",
				CORSOrigins:   "http://localhost:3000",
				Environment:   "development",
			},
			expectedCount: 0,
		},
		{
			name: "warnings for missing JWT secret in dev",
			config: &Config{
				JWTSecret:     "",
				AdminPassword: "admin123",
				LLMAPIKey:     "key",
				CORSOrigins:   "http://localhost:3000",
				Environment:   "development",
			},
			expectedCount: 1,
		},
		{
			name: "multiple warnings",
			config: &Config{
				JWTSecret:     "",
				AdminPassword: "",
				LLMAPIKey:     "",
				CORSOrigins:   "",
				Environment:   "development",
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := tt.config.GetWarnings()
			if len(warnings) != tt.expectedCount {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectedCount, len(warnings), warnings)
			}
		})
	}
}

func TestConfig_EnvironmentHelpers(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		isProd      bool
		isDev       bool
	}{
		{
			name:        "production environment",
			environment: "production",
			isProd:      true,
			isDev:       false,
		},
		{
			name:        "prod shorthand",
			environment: "prod",
			isProd:      true,
			isDev:       false,
		},
		{
			name:        "development environment",
			environment: "development",
			isProd:      false,
			isDev:       true,
		},
		{
			name:        "dev shorthand",
			environment: "dev",
			isProd:      false,
			isDev:       true,
		},
		{
			name:        "empty defaults to development",
			environment: "",
			isProd:      false,
			isDev:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}

			if cfg.IsProduction() != tt.isProd {
				t.Errorf("IsProduction() = %v, expected %v", cfg.IsProduction(), tt.isProd)
			}

			if cfg.IsDevelopment() != tt.isDev {
				t.Errorf("IsDevelopment() = %v, expected %v", cfg.IsDevelopment(), tt.isDev)
			}
		})
	}
}

func TestConfig_ValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "DATABASE_URL",
		Message: "is required",
	}

	expected := "configuration error: DATABASE_URL - is required"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestConfig_ValidationErrors(t *testing.T) {
	errs := ValidationErrors{
		{Field: "DATABASE_URL", Message: "is required"},
		{Field: "JWT_SECRET", Message: "is required in production"},
	}

	expected := "configuration error: DATABASE_URL - is required; configuration error: JWT_SECRET - is required in production"
	if errs.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, errs.Error())
	}
}

// --- Tests for LoadAndValidate ---

func TestLoadAndValidate_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := LoadAndValidate()
	if err != nil {
		t.Errorf("LoadAndValidate() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Error("LoadAndValidate() returned nil config")
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("Expected DatabaseURL, got %s", cfg.DatabaseURL)
	}
}

func TestLoadAndValidate_MissingDatabaseURL(t *testing.T) {
	// Clear DATABASE_URL
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")

	cfg, err := LoadAndValidate()
	if err == nil {
		t.Error("LoadAndValidate() expected error for missing DATABASE_URL")
	}
	if cfg != nil {
		t.Error("LoadAndValidate() should return nil config on error")
	}

	errs, ok := err.(ValidationErrors)
	if !ok {
		t.Errorf("Expected ValidationErrors, got %T", err)
		return
	}

	found := false
	for _, e := range errs {
		if e.Field == "DATABASE_URL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for DATABASE_URL field")
	}
}

func TestLoadAndValidate_MissingJWTSecretInProduction(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("ENV", "production")
	os.Unsetenv("JWT_SECRET")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("ENV")
	}()

	cfg, err := LoadAndValidate()
	if err == nil {
		t.Error("LoadAndValidate() expected error for missing JWT_SECRET in production")
	}
	if cfg != nil {
		t.Error("LoadAndValidate() should return nil config on error")
	}
}

// --- Tests for MustLoad ---

func TestMustLoad_Success(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "test-secret")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := MustLoad()
	if cfg == nil {
		t.Error("MustLoad() returned nil config")
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("Expected DatabaseURL, got %s", cfg.DatabaseURL)
	}
}

func TestMustLoad_MissingDatabaseURL_Exits(t *testing.T) {
	// Clear required env vars
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")

	// MustLoad calls os.Exit(1), so we can't easily test it directly
	// Instead, we'll test that LoadAndValidate returns an error
	_, err := LoadAndValidate()
	if err == nil {
		t.Error("LoadAndValidate() expected error for missing DATABASE_URL")
	}
}

// --- Tests for parseCompressionLevel ---

func TestParseCompressionLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		wantError bool
	}{
		{"level 1", "1", 1, false},
		{"level 5", "5", 5, false},
		{"level 9", "9", 9, false},
		{"level 6", "6", 6, false},
		{"level 0 - invalid", "0", 6, true}, // returns default 6 with error
		{"level 10 - invalid", "10", 6, true},
		{"negative - invalid", "-1", 6, true},
		{"empty string", "", 6, true},
		{"non-numeric", "abc", 6, true},
		{"mixed", "5abc", 5, false}, // parses until non-digit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCompressionLevel(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parseCompressionLevel(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseCompressionLevel(%q) unexpected error: %v", tt.input, err)
				}
			}
			if result != tt.expected {
				t.Errorf("parseCompressionLevel(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// --- Tests for parseSize ---

func TestParseSize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		wantError bool
	}{
		{"plain number", "1024", 1024, false},
		{"plain number zero", "0", 0, false},
		{"1KB", "1KB", 1024, false},
		{"2KB", "2KB", 2048, false},
		{"1MB", "1MB", 1024 * 1024, false},
		{"2MB", "2MB", 2 * 1024 * 1024, false},
		{"1GB", "1GB", 1024 * 1024 * 1024, false},
		{"with spaces", "  1024  ", 1024, false},
		{"KB with space", "1 KB", 1024, false},
		{"MB with space", "2 MB", 2 * 1024 * 1024, false},
		{"empty string", "", 0, false},
		{"mixed numeric", "1024abc", 1024, false}, // parses digits only
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parseSize(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseSize(%q) unexpected error: %v", tt.input, err)
				}
			}
			if result != tt.expected {
				t.Errorf("parseSize(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// --- Tests for PrintWarnings ---

func TestPrintWarnings_NoWarnings(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &Config{
		JWTSecret:     "secret",
		AdminPassword: "admin123",
		LLMAPIKey:     "key",
		CORSOrigins:   "http://localhost:3000",
		Environment:   "development",
	}
	cfg.PrintWarnings()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		t.Errorf("PrintWarnings() expected no output, got: %s", output)
	}
}

func TestPrintWarnings_WithWarnings(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &Config{
		JWTSecret:     "",
		AdminPassword: "",
		LLMAPIKey:     "",
		CORSOrigins:   "",
		Environment:   "development",
	}
	cfg.PrintWarnings()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "JWT_SECRET") {
		t.Error("PrintWarnings() expected JWT_SECRET warning")
	}
	if !strings.Contains(output, "ADMIN_PASSWORD") {
		t.Error("PrintWarnings() expected ADMIN_PASSWORD warning")
	}
	if !strings.Contains(output, "LLM_API_KEY") {
		t.Error("PrintWarnings() expected LLM_API_KEY warning")
	}
}

func TestPrintWarnings_CORSWildcardWarning(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &Config{
		JWTSecret:     "secret",
		AdminPassword: "admin123",
		LLMAPIKey:     "key",
		CORSOrigins:   "*",
		Environment:   "development",
	}
	cfg.PrintWarnings()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "CORS") {
		t.Error("PrintWarnings() expected CORS wildcard warning")
	}
}

// --- Tests for New ---

func TestNew_Success(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "test-secret")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := New()
	if err != nil {
		t.Errorf("New() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Error("New() returned nil config")
	}
}

func TestNew_Error(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")

	cfg, err := New()
	if err == nil {
		t.Error("New() expected error for missing DATABASE_URL")
	}
	if cfg != nil {
		t.Error("New() should return nil config on error")
	}
}

// --- Tests for LoadFromEnv with WebSocket compression settings ---

func TestLoadFromEnv_WSCompressionSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("WS_COMPRESSION_ENABLED", "false")
	os.Setenv("WS_COMPRESSION_LEVEL", "5")
	os.Setenv("WS_COMPRESSION_MIN_SIZE", "2KB")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("WS_COMPRESSION_ENABLED")
		os.Unsetenv("WS_COMPRESSION_LEVEL")
		os.Unsetenv("WS_COMPRESSION_MIN_SIZE")
	}()

	cfg := LoadFromEnv()

	if cfg.WSCompressionEnabled != false {
		t.Error("Expected WSCompressionEnabled to be false")
	}
	if cfg.WSCompressionLevel != 5 {
		t.Errorf("Expected WSCompressionLevel 5, got %d", cfg.WSCompressionLevel)
	}
	if cfg.WSCompressionMinSize != 2048 {
		t.Errorf("Expected WSCompressionMinSize 2048, got %d", cfg.WSCompressionMinSize)
	}
}

func TestLoadFromEnv_WSACompressionDefaults(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	defer func() {
		os.Unsetenv("DATABASE_URL")
	}()

	cfg := LoadFromEnv()

	// Check defaults
	if cfg.WSCompressionEnabled != true {
		t.Error("Expected default WSCompressionEnabled to be true")
	}
	if cfg.WSCompressionLevel != 6 {
		t.Errorf("Expected default WSCompressionLevel 6, got %d", cfg.WSCompressionLevel)
	}
	if cfg.WSCompressionMinSize != 1024 {
		t.Errorf("Expected default WSCompressionMinSize 1024, got %d", cfg.WSCompressionMinSize)
	}
}

func TestLoadFromEnv_InvalidWSCompressionLevel(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("WS_COMPRESSION_LEVEL", "15") // Invalid - out of range

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("WS_COMPRESSION_LEVEL")
	}()

	cfg := LoadFromEnv()

	// Should fall back to default
	if cfg.WSCompressionLevel != 6 {
		t.Errorf("Expected default WSCompressionLevel 6, got %d", cfg.WSCompressionLevel)
	}
}

// --- Tests for LoadFromEnv with other settings ---

func TestLoadFromEnv_RedisPoolSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("REDIS_POOL_SIZE", "100")
	os.Setenv("REDIS_MIN_IDLE_CONNS", "20")
	os.Setenv("REDIS_MAX_RETRIES", "5")
	os.Setenv("REDIS_READ_TIMEOUT", "10")
	os.Setenv("REDIS_WRITE_TIMEOUT", "10")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_POOL_SIZE")
		os.Unsetenv("REDIS_MIN_IDLE_CONNS")
		os.Unsetenv("REDIS_MAX_RETRIES")
		os.Unsetenv("REDIS_READ_TIMEOUT")
		os.Unsetenv("REDIS_WRITE_TIMEOUT")
	}()

	cfg := LoadFromEnv()

	if cfg.RedisPoolSize != 100 {
		t.Errorf("Expected RedisPoolSize 100, got %d", cfg.RedisPoolSize)
	}
	if cfg.RedisMinIdleConns != 20 {
		t.Errorf("Expected RedisMinIdleConns 20, got %d", cfg.RedisMinIdleConns)
	}
	if cfg.RedisMaxRetries != 5 {
		t.Errorf("Expected RedisMaxRetries 5, got %d", cfg.RedisMaxRetries)
	}
	if cfg.RedisReadTimeout != 10 {
		t.Errorf("Expected RedisReadTimeout 10, got %d", cfg.RedisReadTimeout)
	}
	if cfg.RedisWriteTimeout != 10 {
		t.Errorf("Expected RedisWriteTimeout 10, got %d", cfg.RedisWriteTimeout)
	}
}

func TestLoadFromEnv_RateLimitSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("RATE_LIMIT_ENABLED", "false")
	os.Setenv("RATE_LIMIT_REQUESTS_PER_SECOND", "200")
	os.Setenv("RATE_LIMIT_BURST", "400")
	os.Setenv("RATE_LIMIT_WINDOW", "120")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("RATE_LIMIT_ENABLED")
		os.Unsetenv("RATE_LIMIT_REQUESTS_PER_SECOND")
		os.Unsetenv("RATE_LIMIT_BURST")
		os.Unsetenv("RATE_LIMIT_WINDOW")
	}()

	cfg := LoadFromEnv()

	if cfg.RateLimitEnabled != false {
		t.Error("Expected RateLimitEnabled to be false")
	}
	if cfg.RateLimitRequestsPerSecond != 200 {
		t.Errorf("Expected RateLimitRequestsPerSecond 200, got %d", cfg.RateLimitRequestsPerSecond)
	}
	if cfg.RateLimitBurst != 400 {
		t.Errorf("Expected RateLimitBurst 400, got %d", cfg.RateLimitBurst)
	}
	if cfg.RateLimitWindow != 120 {
		t.Errorf("Expected RateLimitWindow 120, got %d", cfg.RateLimitWindow)
	}
}

func TestLoadFromEnv_WAFSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("WAF_ENABLED", "false")
	os.Setenv("WAF_BLOCKED_USER_AGENTS", "badbot,evilbot")
	os.Setenv("WAF_BLOCK_EMPTY_USER_AGENT", "false")
	os.Setenv("WAF_MAX_REQUEST_SIZE", "20MB")
	os.Setenv("WAF_MAX_ARGS_LENGTH", "2000")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("WAF_ENABLED")
		os.Unsetenv("WAF_BLOCKED_USER_AGENTS")
		os.Unsetenv("WAF_BLOCK_EMPTY_USER_AGENT")
		os.Unsetenv("WAF_MAX_REQUEST_SIZE")
		os.Unsetenv("WAF_MAX_ARGS_LENGTH")
	}()

	cfg := LoadFromEnv()

	if cfg.WAFEnabled != false {
		t.Error("Expected WAFEnabled to be false")
	}
	if cfg.WAFBlockedUserAgents != "badbot,evilbot" {
		t.Errorf("Expected WAFBlockedUserAgents 'badbot,evilbot', got %s", cfg.WAFBlockedUserAgents)
	}
	if cfg.WAFBlockEmptyUserAgent != false {
		t.Error("Expected WAFBlockEmptyUserAgent to be false")
	}
	if cfg.WAFMaxRequestSize != "20MB" {
		t.Errorf("Expected WAFMaxRequestSize '20MB', got %s", cfg.WAFMaxRequestSize)
	}
	if cfg.WAFMaxArgsLength != 2000 {
		t.Errorf("Expected WAFMaxArgsLength 2000, got %d", cfg.WAFMaxArgsLength)
	}
}

func TestLoadFromEnv_DatabasePoolSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "25")
	os.Setenv("DB_CONN_MAX_LIFETIME", "3600")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "600")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("DB_CONN_MAX_IDLE_TIME")
	}()

	cfg := LoadFromEnv()

	if cfg.DBMaxOpenConns != 50 {
		t.Errorf("Expected DBMaxOpenConns 50, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 25 {
		t.Errorf("Expected DBMaxIdleConns 25, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 3600 {
		t.Errorf("Expected DBConnMaxLifetime 3600, got %d", cfg.DBConnMaxLifetime)
	}
	if cfg.DBConnMaxIdleTime != 600 {
		t.Errorf("Expected DBConnMaxIdleTime 600, got %d", cfg.DBConnMaxIdleTime)
	}
}

func TestLoadFromEnv_LLMHTTPSettings(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("LLM_HTTP_TIMEOUT", "60")
	os.Setenv("LLM_MAX_IDLE_CONNS", "200")
	os.Setenv("LLM_MAX_IDLE_CONNS_PER_HOST", "20")
	os.Setenv("LLM_IDLE_CONN_TIMEOUT", "120")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("LLM_HTTP_TIMEOUT")
		os.Unsetenv("LLM_MAX_IDLE_CONNS")
		os.Unsetenv("LLM_MAX_IDLE_CONNS_PER_HOST")
		os.Unsetenv("LLM_IDLE_CONN_TIMEOUT")
	}()

	cfg := LoadFromEnv()

	if cfg.LLMHTTPTimeout != 60 {
		t.Errorf("Expected LLMHTTPTimeout 60, got %d", cfg.LLMHTTPTimeout)
	}
	if cfg.LLMMaxIdleConns != 200 {
		t.Errorf("Expected LLMMaxIdleConns 200, got %d", cfg.LLMMaxIdleConns)
	}
	if cfg.LLMMaxIdleConnsPerHost != 20 {
		t.Errorf("Expected LLMMaxIdleConnsPerHost 20, got %d", cfg.LLMMaxIdleConnsPerHost)
	}
	if cfg.LLMIdleConnTimeout != 120 {
		t.Errorf("Expected LLMIdleConnTimeout 120, got %d", cfg.LLMIdleConnTimeout)
	}
}

func TestLoadFromEnv_RequestTimeout(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("REQUEST_TIMEOUT", "30")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REQUEST_TIMEOUT")
	}()

	cfg := LoadFromEnv()

	if cfg.RequestTimeout != 30 {
		t.Errorf("Expected RequestTimeout 30, got %d", cfg.RequestTimeout)
	}
}

// --- Tests for ValidateCORS in production ---

func TestValidateCORS_ProductionEmpty(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		JWTSecret:   "secret",
		Environment: "production",
		CORSOrigins: "",
	}

	err := cfg.ValidateCORS()
	if err == nil {
		t.Error("ValidateCORS() expected error for empty CORS in production")
	}
}

func TestValidateCORS_ProductionWildcard(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		JWTSecret:   "secret",
		Environment: "production",
		CORSOrigins: "*",
	}

	err := cfg.ValidateCORS()
	if err == nil {
		t.Error("ValidateCORS() expected error for wildcard CORS in production")
	}
}

func TestValidateCORS_ProductionInvalidOrigin(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		JWTSecret:   "secret",
		Environment: "production",
		CORSOrigins: "invalid-origin",
	}

	err := cfg.ValidateCORS()
	if err == nil {
		t.Error("ValidateCORS() expected error for invalid origin format in production")
	}
}

func TestValidateCORS_ProductionValid(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		JWTSecret:   "secret",
		Environment: "production",
		CORSOrigins: "https://example.com,https://app.example.com",
	}

	err := cfg.ValidateCORS()
	if err != nil {
		t.Errorf("ValidateCORS() unexpected error: %v", err)
	}
}

func TestValidateCORS_DevelopmentAllowsEmpty(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		JWTSecret:   "secret",
		Environment: "development",
		CORSOrigins: "",
	}

	err := cfg.ValidateCORS()
	if err != nil {
		t.Errorf("ValidateCORS() unexpected error in development: %v", err)
	}
}

// --- Test for HasCORSWildcard ---

func TestHasCORSWildcard(t *testing.T) {
	tests := []struct {
		name     string
		origins  string
		expected bool
	}{
		{"wildcard", "*", true},
		{"multiple with wildcard", "http://localhost:3000,*,https://example.com", true},
		{"no wildcard", "http://localhost:3000,https://example.com", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{CORSOrigins: tt.origins}
			if cfg.HasCORSWildcard() != tt.expected {
				t.Errorf("HasCORSWildcard() = %v, expected %v", cfg.HasCORSWildcard(), tt.expected)
			}
		})
	}
}

// --- Test for GetWarnings with CORS wildcard in production ---

func TestGetWarnings_ProductionCORSWildcard(t *testing.T) {
	cfg := &Config{
		JWTSecret:     "secret",
		AdminPassword: "admin",
		LLMAPIKey:     "key",
		CORSOrigins:   "*",
		Environment:   "production",
	}

	warnings := cfg.GetWarnings()

	// Should have a warning about CORS wildcard
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "SECURITY WARNING") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected SECURITY WARNING for CORS wildcard in production")
	}
}
