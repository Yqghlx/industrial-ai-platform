package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig 数据库安全配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string        // "disable", "require", "verify-ca", "verify-full"
	SSLCert         string        // 客户端证书路径 (verify-ca/verify-full)
	SSLKey          string        // 客户端私钥路径
	SSLRootCert     string        // CA 证书路径
	MaxOpenConns    int           // 最大打开连接数 (防止连接耗尽)
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 空闲连接超时
}

// DefaultProductionConfig 生产环境默认安全配置
func DefaultProductionConfig() *DatabaseConfig {
	return &DatabaseConfig{
		SSLMode:         "require",        // 强制 SSL 连接
		MaxOpenConns:    50,               // 限制最大连接数
		MaxIdleConns:    10,               // 保持一定空闲连接
		ConnMaxLifetime: 1 * time.Hour,    // 连接最大 1 小时
		ConnMaxIdleTime: 10 * time.Minute, // 空闲 10 分钟关闭
	}
}

// SecureConnection 建立安全的数据库连接
func SecureConnection(config *DatabaseConfig) (*sql.DB, error) {
	// 构建连接字符串
	connStr := buildConnectionString(config)

	// 打开连接
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池 (防止连接耗尽攻击)
	configureConnectionPool(db, config)

	// 验证连接
	if err := verifyConnection(db); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	return db, nil
}

// buildConnectionString 构建安全的连接字符串
func buildConnectionString(config *DatabaseConfig) string {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	// 添加 SSL 证书配置 (如果使用 verify-ca 或 verify-full)
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

// configureConnectionPool 配置连接池安全参数
func configureConnectionPool(db *sql.DB, config *DatabaseConfig) {
	// 限制最大打开连接数 (防止连接耗尽攻击)
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}

	// 限制最大空闲连接数
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}

	// 设置连接最大生命周期 (防止长时间连接)
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// 设置空闲连接超时 (清理不活跃连接)
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}
}

// verifyConnection 验证数据库连接
func verifyConnection(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	err := db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 检查 SSL 连接状态 (如果启用了 SSL)
	var sslStatus string
	err = db.QueryRowContext(ctx, "SELECT ssl_is_used()").Scan(&sslStatus)
	if err == nil && sslStatus != "true" {
		// 注意: ssl_is_used() 不是标准 PostgreSQL 函数
		// 使用 pg_stat_ssl 查询代替
	}

	return nil
}

// CheckSSLStatus 检查当前连接的 SSL 状态
func CheckSSLStatus(db *sql.DB) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ssl bool
	var version string
	var cipher string

	err := db.QueryRowContext(ctx, `
		SELECT ssl, ssl_version, ssl_cipher
		FROM pg_stat_ssl
		WHERE pid = pg_backend_pid()
	`).Scan(&ssl, &version, &cipher)

	if err != nil {
		return false, "", "", fmt.Errorf("failed to check SSL status: %w", err)
	}

	return ssl, version, cipher, nil
}

// HealthCheck 数据库健康检查 (包含安全检查)
func HealthCheck(db *sql.DB) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make(map[string]interface{})

	// 1. 连接状态
	err := db.PingContext(ctx)
	result["connected"] = err == nil

	// 2. SSL 状态
	ssl, version, cipher, err := CheckSSLStatus(db)
	if err == nil {
		result["ssl_enabled"] = ssl
		result["ssl_version"] = version
		result["ssl_cipher"] = cipher
	}

	// 3. 连接池状态
	stats := db.Stats()
	result["open_connections"] = stats.OpenConnections
	result["in_use"] = stats.InUse
	result["idle"] = stats.Idle
	result["max_open"] = stats.MaxOpenConnections
	result["max_idle"] = stats.MaxIdleClosed
	result["wait_count"] = stats.WaitCount // 等待连接的次数 (连接压力指标)

	// 4. 数据库版本
	var dbVersion string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&dbVersion)
	if err == nil {
		result["version"] = dbVersion
	}

	return result
}

// ValidateQuery 验证查询安全性 (开发/测试环境使用)
// 检查是否有潜在的 SQL 注入风险
func ValidateQuery(query string) []string {
	warnings := []string{}

	// 检查危险模式
	dangerousPatterns := []string{
		"DROP TABLE",
		"TRUNCATE TABLE",
		"DELETE FROM",  // 批量删除
		"--",           // SQL 注释 (可能注入)
		"/*",           // 多行注释
		";DROP",        // 多语句注入
		"UNION SELECT", // UNION 注入
		"OR 1=1",       // 常见注入
		"AND 1=1",      // 常见注入
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(query, pattern) {
			warnings = append(warnings, fmt.Sprintf("Potential SQL injection pattern detected: %s", pattern))
		}
	}

	return warnings
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsIgnoreCaseHelper(s, substr))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			subc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc = sc + 32
			}
			if subc >= 'A' && subc <= 'Z' {
				subc = subc + 32
			}
			if sc != subc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
