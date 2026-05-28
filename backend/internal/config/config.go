package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Required fields
	DatabaseURL string
	JWTSecret   string

	// Optional fields with defaults
	Port string

	// FIX-011: Server Configuration
	RequestTimeout int // HTTP 请求超时时间（秒）

	// LLM Configuration (optional)
	LLMAPIKey  string
	LLMBaseURL string
	LLMModel   string

	// FIX-039: HTTP Client Configuration for LLM
	LLMHTTPTimeout         int // HTTP 请求超时时间（秒）
	LLMMaxIdleConns        int // 最大空闲连接数
	LLMMaxIdleConnsPerHost int // 每个主机最大空闲连接数
	LLMIdleConnTimeout     int // 空闲连接超时时间（秒）

	// Cache Configuration (optional)
	RedisURL     string
	CacheEnabled bool
	CachePrefix  string

	// FIX-042: Redis Connection Pool Configuration
	RedisPoolSize     int // Redis 连接池大小
	RedisMinIdleConns int // Redis 最小空闲连接数
	RedisMaxRetries   int // Redis 最大重试次数
	RedisReadTimeout  int // Redis 读超时（秒）
	RedisWriteTimeout int // Redis 写超时（秒）

	// FIX-012: Database Connection Pool Configuration
	DBMaxOpenConns    int // 数据库最大打开连接数
	DBMaxIdleConns    int // 数据库最大空闲连接数
	DBConnMaxLifetime int // 数据库连接最大生命周期（秒）
	DBConnMaxIdleTime int // 数据库空闲连接最大存活时间（秒）

	// WebSocket Compression Configuration (optional)
	WSCompressionEnabled bool
	WSCompressionLevel   int
	WSCompressionMinSize int // Minimum message size in bytes to trigger compression (default 1KB)

	// FIX-042: Rate Limiting Configuration
	RateLimitEnabled           bool // 是否启用限流
	RateLimitRequestsPerSecond int  // 每秒最大请求数
	RateLimitBurst             int  // 突发流量限制
	RateLimitWindow            int  // 限流窗口（秒）

	// FIX-040: WAF Configuration
	WAFEnabled             bool   // 是否启用 WAF
	WAFBlockedUserAgents   string // 阻止的 User-Agent（逗号分隔）
	WAFBlockEmptyUserAgent bool   // 是否阻止空 User-Agent
	WAFMaxRequestSize      string // 最大请求大小（如 "10MB")
	WAFMaxArgsLength       int    // 最大参数长度

	// Other optional fields
	CORSOrigins   string
	AdminPassword string

	// Environment indicator
	Environment string

	// P2-01: Server host for logging URLs (default: localhost)
	ServerHost string

	// P2-04: Fallback LLM API URL for health checks
	LLMFallbackURL string
}

// ValidationError represents configuration validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("configuration error: %s - %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate checks that all required configuration is present and valid
func (c *Config) Validate() error {
	var errs ValidationErrors

	// Required: DATABASE_URL
	if c.DatabaseURL == "" {
		errs = append(errs, ValidationError{
			Field:   "DATABASE_URL",
			Message: "is required. Please set the DATABASE_URL environment variable with your PostgreSQL connection string.",
		})
	}

	// Required for production: JWT_SECRET
	if c.JWTSecret == "" {
		env := c.Environment
		if env == "" {
			env = os.Getenv("ENV")
		}
		if env == "" {
			env = os.Getenv("GO_ENV")
		}

		// In production, JWT_SECRET is required
		if env == "production" || env == "prod" {
			errs = append(errs, ValidationError{
				Field:   "JWT_SECRET",
				Message: "is required in production environment. Please set the JWT_SECRET environment variable.",
			})
		}
	}

	// Set default port if not specified
	if c.Port == "" {
		c.Port = "8080"
	}

	// Validate port format (basic check)
	if c.Port != "" {
		if len(c.Port) > 5 || len(c.Port) < 2 {
			errs = append(errs, ValidationError{
				Field:   "PORT",
				Message: fmt.Sprintf("invalid port number: %s", c.Port),
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	// Validate CORS configuration (only errors in production)
	if err := c.ValidateCORS(); err != nil {
		return err
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	// Parse cache enabled (default: true)
	cacheEnabled := true
	if v := os.Getenv("CACHE_ENABLED"); v == "false" {
		cacheEnabled = false
	}

	// Parse WebSocket compression enabled (default: true)
	wsCompressionEnabled := true
	if v := os.Getenv("WS_COMPRESSION_ENABLED"); v == "false" {
		wsCompressionEnabled = false
	}

	// Parse WebSocket compression level (default: 6, range: 1-9)
	wsCompressionLevel := 6
	if v := os.Getenv("WS_COMPRESSION_LEVEL"); v != "" {
		if level, err := parseCompressionLevel(v); err == nil {
			wsCompressionLevel = level
		}
	}

	// Parse WebSocket compression minimum size (default: 1KB = 1024 bytes)
	wsCompressionMinSize := 1024
	if v := os.Getenv("WS_COMPRESSION_MIN_SIZE"); v != "" {
		if size, err := parseSize(v); err == nil {
			wsCompressionMinSize = size
		}
	}

	// FIX-039: Parse LLM HTTP configuration
	llmHTTPTimeout := parseEnvInt("LLM_HTTP_TIMEOUT", 30) // 默认30秒
	llmMaxIdleConns := parseEnvInt("LLM_MAX_IDLE_CONNS", 100)
	llmMaxIdleConnsPerHost := parseEnvInt("LLM_MAX_IDLE_CONNS_PER_HOST", 10)
	llmIdleConnTimeout := parseEnvInt("LLM_IDLE_CONN_TIMEOUT", 90)

	// FIX-042: Parse Redis configuration
	redisPoolSize := parseEnvInt("REDIS_POOL_SIZE", 50)
	redisMinIdleConns := parseEnvInt("REDIS_MIN_IDLE_CONNS", 10)
	redisMaxRetries := parseEnvInt("REDIS_MAX_RETRIES", 3)
	redisReadTimeout := parseEnvInt("REDIS_READ_TIMEOUT", 3)
	redisWriteTimeout := parseEnvInt("REDIS_WRITE_TIMEOUT", 3)

	// FIX-042: Parse Rate Limiting configuration
	rateLimitEnabled := true
	if v := os.Getenv("RATE_LIMIT_ENABLED"); v == "false" {
		rateLimitEnabled = false
	}
	rateLimitRequestsPerSecond := parseEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 100)
	rateLimitBurst := parseEnvInt("RATE_LIMIT_BURST", 200)
	rateLimitWindow := parseEnvInt("RATE_LIMIT_WINDOW", 60)

	// FIX-040: Parse WAF configuration
	wafEnabled := true
	if v := os.Getenv("WAF_ENABLED"); v == "false" {
		wafEnabled = false
	}
	wafBlockedUserAgents := os.Getenv("WAF_BLOCKED_USER_AGENTS")
	wafBlockEmptyUserAgent := true
	if v := os.Getenv("WAF_BLOCK_EMPTY_USER_AGENT"); v == "false" {
		wafBlockEmptyUserAgent = false
	}
	wafMaxRequestSize := getEnv("WAF_MAX_REQUEST_SIZE", "10MB")
	wafMaxArgsLength := parseEnvInt("WAF_MAX_ARGS_LENGTH", 1000)

	// FIX-011: Parse Server configuration
	requestTimeout := parseEnvInt("REQUEST_TIMEOUT", 15) // 默认15秒

	// FIX-012: Parse Database connection pool configuration
	dbMaxOpenConns := parseEnvInt("DB_MAX_OPEN_CONNS", 25)         // 默认25个
	dbMaxIdleConns := parseEnvInt("DB_MAX_IDLE_CONNS", 10)         // 默认10个
	dbConnMaxLifetime := parseEnvInt("DB_CONN_MAX_LIFETIME", 1800) // 默认30分钟（秒）
	dbConnMaxIdleTime := parseEnvInt("DB_CONN_MAX_IDLE_TIME", 300) // 默认5分钟（秒）

	// P2-01: Server host for logging URLs
	serverHost := getEnv("SERVER_HOST", "localhost")

	// P2-04: Fallback LLM API URL for health checks
	llmFallbackURL := getEnv("LLM_FALLBACK_URL", "https://open.bigmodel.cn/api/paas/v4")

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
		// FIX-011
		RequestTimeout: requestTimeout,
		LLMAPIKey:      os.Getenv("LLM_API_KEY"),
		LLMBaseURL:     os.Getenv("LLM_BASE_URL"),
		LLMModel:       os.Getenv("LLM_MODEL"),
		// FIX-039
		LLMHTTPTimeout:         llmHTTPTimeout,
		LLMMaxIdleConns:        llmMaxIdleConns,
		LLMMaxIdleConnsPerHost: llmMaxIdleConnsPerHost,
		LLMIdleConnTimeout:     llmIdleConnTimeout,
		RedisURL:               os.Getenv("REDIS_URL"),
		CacheEnabled:           cacheEnabled,
		CachePrefix:            os.Getenv("CACHE_PREFIX"),
		// FIX-042
		RedisPoolSize:     redisPoolSize,
		RedisMinIdleConns: redisMinIdleConns,
		RedisMaxRetries:   redisMaxRetries,
		RedisReadTimeout:  redisReadTimeout,
		RedisWriteTimeout: redisWriteTimeout,
		// FIX-012
		DBMaxOpenConns:       dbMaxOpenConns,
		DBMaxIdleConns:       dbMaxIdleConns,
		DBConnMaxLifetime:    dbConnMaxLifetime,
		DBConnMaxIdleTime:    dbConnMaxIdleTime,
		WSCompressionEnabled: wsCompressionEnabled,
		WSCompressionLevel:   wsCompressionLevel,
		WSCompressionMinSize: wsCompressionMinSize,
		// FIX-042
		RateLimitEnabled:           rateLimitEnabled,
		RateLimitRequestsPerSecond: rateLimitRequestsPerSecond,
		RateLimitBurst:             rateLimitBurst,
		RateLimitWindow:            rateLimitWindow,
		// FIX-040
		WAFEnabled:             wafEnabled,
		WAFBlockedUserAgents:   wafBlockedUserAgents,
		WAFBlockEmptyUserAgent: wafBlockEmptyUserAgent,
		WAFMaxRequestSize:      wafMaxRequestSize,
		WAFMaxArgsLength:       wafMaxArgsLength,
		CORSOrigins:            os.Getenv("CORS_ORIGINS"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		Environment:            os.Getenv("ENV"),
		// P2-01 & P2-04
		ServerHost:     serverHost,
		LLMFallbackURL: llmFallbackURL,
	}
}

// parseEnvInt 解析环境变量为整数，使用默认值
// FIX-042: 新增辅助函数
func parseEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return i
		}
	}
	return defaultValue
}

// LoadAndValidate loads configuration from environment and validates it
func LoadAndValidate() (*Config, error) {
	cfg := LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// MustLoad loads configuration and panics if validation fails
// Use this in main when you want the application to fail fast on config errors
func MustLoad() *Config {
	cfg, err := LoadAndValidate()
	if err != nil {
		// Print detailed error message
		fmt.Fprintf(os.Stderr, "\nConfiguration Error:\n")
		fmt.Fprintf(os.Stderr, "==================\n")

		if errs, ok := err.(ValidationErrors); ok {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Field, e.Message)
			}
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
		}

		fmt.Fprintf(os.Stderr, "\nPlease check your environment variables and try again.\n")
		fmt.Fprintf(os.Stderr, "Refer to .env.example for required configuration.\n\n")
		os.Exit(1)
	}
	return cfg
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	env := strings.ToLower(c.Environment)
	return env == "production" || env == "prod"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	env := strings.ToLower(c.Environment)
	return env == "development" || env == "dev" || env == ""
}

// GetCORSOrigins returns CORS origins as a slice
// FIX-024: 安全CORS默认值 - 默认返回空数组而非["*"]
func (c *Config) GetCORSOrigins() []string {
	if c.CORSOrigins == "" {
		// Return empty slice instead of wildcard for security
		// In production, explicit CORS_ORIGINS must be set
		return []string{}
	}
	origins := strings.Split(c.CORSOrigins, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	return origins
}

// ValidateCORS validates CORS configuration
// In production environment:
// 1. CORS_ORIGINS must be explicitly set (cannot be empty)
// 2. Wildcard '*' is not allowed
func (c *Config) ValidateCORS() error {
	if !c.IsProduction() {
		return nil
	}

	// FIX: In production, CORS_ORIGINS must be explicitly configured
	if c.CORSOrigins == "" {
		return ValidationError{
			Field:   "CORS_ORIGINS",
			Message: "must be explicitly set in production environment. Please specify allowed origins (e.g., 'https://example.com,https://app.example.com').",
		}
	}

	origins := c.GetCORSOrigins()

	// FIX: Ensure origins list is not empty after parsing
	if len(origins) == 0 {
		return ValidationError{
			Field:   "CORS_ORIGINS",
			Message: "must contain at least one origin in production environment.",
		}
	}

	for _, origin := range origins {
		if origin == "*" {
			return ValidationError{
				Field:   "CORS_ORIGINS",
				Message: "wildcard '*' is not allowed in production environment. Please specify explicit allowed origins (e.g., 'https://example.com,https://app.example.com').",
			}
		}
		// FIX: Validate origin format (must be a valid URL)
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return ValidationError{
				Field:   "CORS_ORIGINS",
				Message: fmt.Sprintf("invalid origin '%s': must start with http:// or https://", origin),
			}
		}
	}
	return nil
}

// HasCORSWildcard returns true if CORS is configured with wildcard
func (c *Config) HasCORSWildcard() bool {
	origins := c.GetCORSOrigins()
	for _, origin := range origins {
		if origin == "*" {
			return true
		}
	}
	return false
}

// GetWarnings returns non-fatal warnings about configuration
func (c *Config) GetWarnings() []string {
	var warnings []string

	if c.JWTSecret == "" && !c.IsProduction() {
		warnings = append(warnings, "JWT_SECRET is not set. JWT authentication will not work. Please set JWT_SECRET environment variable with at least 32 characters.")
	}

	if c.AdminPassword == "" {
		warnings = append(warnings, "ADMIN_PASSWORD is not set. A random password will be generated on first startup.")
	}

	if c.LLMAPIKey == "" {
		warnings = append(warnings, "LLM_API_KEY is not set. AI agent features may not work properly.")
	}

	// CORS wildcard warning - security concern
	if c.HasCORSWildcard() {
		if c.IsProduction() {
			warnings = append(warnings, "SECURITY WARNING: CORS_ORIGINS is set to '*'. This allows ANY origin to access your API. This configuration is BLOCKED in production - please set explicit allowed origins.")
		} else {
			warnings = append(warnings, "WARNING: CORS_ORIGINS is set to '*'. This allows ANY origin to access your API. Consider restricting allowed origins for better security (e.g., CORS_ORIGINS=http://localhost:3000,https://app.example.com).")
		}
	}

	return warnings
}

// PrintWarnings prints configuration warnings to stderr
func (c *Config) PrintWarnings() {
	warnings := c.GetWarnings()
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\nConfiguration Warnings:\n")
		fmt.Fprintf(os.Stderr, "=======================\n")
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", w)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// Helper to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseCompressionLevel parses WebSocket compression level from string
func parseCompressionLevel(s string) (int, error) {
	level := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		level = level*10 + int(c-'0')
	}
	if level < 1 || level > 9 {
		return 6, fmt.Errorf("invalid compression level: %d (must be 1-9)", level)
	}
	return level, nil
}

// parseSize parses a size string like "1KB", "2MB", or just "1024"
func parseSize(s string) (int, error) {
	s = strings.TrimSpace(s)

	// Try to parse as plain number
	if !strings.HasSuffix(s, "KB") && !strings.HasSuffix(s, "MB") && !strings.HasSuffix(s, "GB") {
		size := 0
		for _, c := range s {
			if c < '0' || c > '9' {
				break
			}
			size = size*10 + int(c-'0')
		}
		return size, nil
	}

	// Parse with suffix
	multiplier := 1
	if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}

	size := 0
	for _, c := range strings.TrimSpace(s) {
		if c < '0' || c > '9' {
			break
		}
		size = size*10 + int(c-'0')
	}

	return size * multiplier, nil
}

// New creates a new Config with defaults and validates it
func New() (*Config, error) {
	return LoadAndValidate()
}
