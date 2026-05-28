package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// FIX-018: 敏感信息关键词列表
var sensitivePatterns = []string{
	"password",
	"passwd",
	"secret",
	"key",
	"token",
	"credential",
	"api_key",
	"apikey",
	"auth",
}

// sensitiveRegex 匹配敏感信息的正则表达式
// FIX-018: 编译一次，提高性能
var sensitiveRegex *regexp.Regexp

func init() {
	// 构建正则表达式: (password|secret|...)=([^\s&]+)
	pattern := "(?i)(" + strings.Join(sensitivePatterns, "|") + `)\s*[=:]\s*[^\s&]+`
	sensitiveRegex = regexp.MustCompile(pattern)
}

// redactSensitiveInfo 过滤日志中的敏感信息
// FIX-018: 将敏感信息替换为 [REDACTED]
func redactSensitiveInfo(msg string) string {
	return sensitiveRegex.ReplaceAllStringFunc(msg, func(match string) string {
		// 保留键名，只替换值
		parts := regexp.MustCompile(`(?i)(` + strings.Join(sensitivePatterns, "|") + `)\s*[=:]\s*`).FindStringSubmatch(match)
		if len(parts) > 1 {
			return parts[0] + "[REDACTED]"
		}
		return "[REDACTED]"
	})
}

// safeLog 安全日志输出，过滤敏感信息
// FIX-018: 包装标准日志函数
func safeLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Print(redactSensitiveInfo(msg))
}

// safeLogf 安全格式化日志输出，过滤敏感信息
// FIX-018: 包装 log.Printf 函数
func safeLogf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("%s", redactSensitiveInfo(msg))
}

// ConnectionConfig 数据库连接配置
type ConnectionConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
	// 连接池配置
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConnectionConfig 默认连接配置
// FIX-012: 从环境变量读取连接池参数
// SEC-HIGH-01: 默认使用 SSL 连接，而不是禁用 SSL
// 在生产环境中，应该使用 verify-full 以验证服务器证书
func DefaultConnectionConfig() *ConnectionConfig {
	// SEC-HIGH-01: 从环境变量读取 SSL 模式，默认使用 require
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require" // 安全默认值：要求 SSL 连接
	}
	return &ConnectionConfig{
		SSLMode:         sslMode,
		MaxOpenConns:    parseEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    parseEnvInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: time.Duration(parseEnvInt("DB_CONN_MAX_LIFETIME", 1800)) * time.Second,
		ConnMaxIdleTime: time.Duration(parseEnvInt("DB_CONN_MAX_IDLE_TIME", 300)) * time.Second,
	}
}

// ProductionConnectionConfig 生产环境连接配置
// FIX-012: 从环境变量读取连接池参数
func ProductionConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		SSLMode:         "require",
		MaxOpenConns:    parseEnvInt("DB_MAX_OPEN_CONNS", 50),
		MaxIdleConns:    parseEnvInt("DB_MAX_IDLE_CONNS", 15),
		ConnMaxLifetime: time.Duration(parseEnvInt("DB_CONN_MAX_LIFETIME", 3600)) * time.Second,
		ConnMaxIdleTime: time.Duration(parseEnvInt("DB_CONN_MAX_IDLE_TIME", 600)) * time.Second,
	}
}

// parseEnvInt 从环境变量解析整数，使用默认值
// FIX-012: 新增辅助函数
func parseEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return i
		}
	}
	return defaultValue
}

// Connect 建立数据库连接并配置连接池
func Connect(config *ConnectionConfig) (*sql.DB, error) {
	connStr := buildConnString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 配置连接池
	ConfigurePool(db, config)

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// FIX-018: 使用安全日志函数，防止敏感信息泄露
	safeLogf("Database connection established: %s:%d/%s", config.Host, config.Port, config.Database)
	return db, nil
}

// ConfigurePool 配置连接池参数
// 推荐配置：
// - MaxOpenConns: 限制最大连接数，防止连接耗尽
// - MaxIdleConns: 保持适量空闲连接，减少连接创建开销
// - ConnMaxLifetime: 连接最大生命周期，防止长时间连接问题
// - ConnMaxIdleTime: 空闲连接超时，自动清理不活跃连接
func ConfigurePool(db *sql.DB, config *ConnectionConfig) {
	// SetMaxOpenConns 设置数据库的最大打开连接数
	// 防止连接数过多导致数据库压力过大或连接耗尽攻击
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
		safeLogf("Connection pool: MaxOpenConns=%d", config.MaxOpenConns)
	}

	// SetMaxIdleConns 设置数据库的最大空闲连接数
	// 保持一定数量的空闲连接可以提高性能，避免频繁创建和销毁连接
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
		safeLogf("Connection pool: MaxIdleConns=%d", config.MaxIdleConns)
	}

	// SetConnMaxLifetime 设置连接的最大可复用时间
	// 超过此时间的连接会在下次使用时关闭，防止长时间连接导致的问题
	// 建议生产环境设置为 30 分钟到 1 小时
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
		safeLogf("Connection pool: ConnMaxLifetime=%v", config.ConnMaxLifetime)
	}

	// SetConnMaxIdleTime 设置空闲连接的最大存活时间
	// 超过此时间的空闲连接会被关闭，有助于清理不活跃的连接
	// 建议设置为 ConnMaxLifetime 的 1/6 到 1/3
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
		safeLogf("Connection pool: ConnMaxIdleTime=%v", config.ConnMaxIdleTime)
	}
}

// buildConnString 构建数据库连接字符串
func buildConnString(config *ConnectionConfig) string {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	if config.SSLCert != "" {
		connStr += fmt.Sprintf(" sslcert=%s", config.SSLCert)
	}
	if config.SSLKey != "" {
		connStr += fmt.Sprintf(" sslkey=%s", config.SSLKey)
	}
	if config.SSLRootCert != "" {
		connStr += fmt.Sprintf(" sslrootcert=%s", config.SSLRootCert)
	}

	return connStr
}

// Close 关闭数据库连接
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// IsConnected 检查数据库连接是否正常
func IsConnected(db *sql.DB) bool {
	if db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return db.PingContext(ctx) == nil
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats(db *sql.DB) map[string]interface{} {
	if db == nil {
		return nil
	}
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
	}
}

// CheckHealth 数据库健康检查（简单版本，只返回错误）
func CheckHealth(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查连接
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查连接池状态
	stats := db.Stats()
	usagePercent := float64(0)
	if stats.MaxOpenConnections > 0 {
		usagePercent = float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	}

	if usagePercent > 90 {
		safeLogf("Warning: Connection pool usage is high: %.1f%%", usagePercent)
	}

	return nil
}
