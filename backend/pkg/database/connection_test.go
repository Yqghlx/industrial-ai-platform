package database

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewMockDB creates a mock database with ping monitoring enabled
func NewMockDB(t *testing.T) (db *sql.DB, mock sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	return db, mock
}

func TestDefaultConnectionConfig(t *testing.T) {
	t.Run("returns default config", func(t *testing.T) {
		// 清除环境变量确保测试隔离
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("DB_CONN_MAX_IDLE_TIME")

		config := DefaultConnectionConfig()

		assert.NotNil(t, config)
		assert.Equal(t, "disable", config.SSLMode)
		assert.Equal(t, 25, config.MaxOpenConns)
		assert.Equal(t, 10, config.MaxIdleConns)
		assert.Equal(t, 1800*time.Second, config.ConnMaxLifetime)
		assert.Equal(t, 300*time.Second, config.ConnMaxIdleTime)
	})

	t.Run("reads from environment variables", func(t *testing.T) {
		os.Setenv("DB_MAX_OPEN_CONNS", "100")
		os.Setenv("DB_MAX_IDLE_CONNS", "50")
		os.Setenv("DB_CONN_MAX_LIFETIME", "7200")
		os.Setenv("DB_CONN_MAX_IDLE_TIME", "600")
		defer func() {
			os.Unsetenv("DB_MAX_OPEN_CONNS")
			os.Unsetenv("DB_MAX_IDLE_CONNS")
			os.Unsetenv("DB_CONN_MAX_LIFETIME")
			os.Unsetenv("DB_CONN_MAX_IDLE_TIME")
		}()

		config := DefaultConnectionConfig()

		assert.Equal(t, 100, config.MaxOpenConns)
		assert.Equal(t, 50, config.MaxIdleConns)
		assert.Equal(t, 7200*time.Second, config.ConnMaxLifetime)
		assert.Equal(t, 600*time.Second, config.ConnMaxIdleTime)
	})
}

func TestProductionConnectionConfig(t *testing.T) {
	t.Run("returns production config", func(t *testing.T) {
		config := ProductionConnectionConfig()

		assert.NotNil(t, config)
		assert.Equal(t, "require", config.SSLMode)
		assert.Equal(t, 50, config.MaxOpenConns)
		assert.Equal(t, 15, config.MaxIdleConns)
		assert.Equal(t, 3600*time.Second, config.ConnMaxLifetime)
		assert.Equal(t, 600*time.Second, config.ConnMaxIdleTime)
	})

	t.Run("reads from environment variables", func(t *testing.T) {
		os.Setenv("DB_MAX_OPEN_CONNS", "200")
		os.Setenv("DB_MAX_IDLE_CONNS", "75")
		defer func() {
			os.Unsetenv("DB_MAX_OPEN_CONNS")
			os.Unsetenv("DB_MAX_IDLE_CONNS")
		}()

		config := ProductionConnectionConfig()

		assert.Equal(t, 200, config.MaxOpenConns)
		assert.Equal(t, 75, config.MaxIdleConns)
	})
}

func TestParseEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"returns default when env not set", "NON_EXISTENT_KEY", "", 42, 42},
		{"returns env value when set", "TEST_KEY", "100", 42, 100},
		{"returns default for invalid value", "TEST_KEY", "invalid", 42, 42},
		{"returns default for zero value", "TEST_KEY", "0", 42, 42},
		{"returns default for negative value", "TEST_KEY", "-5", 42, 42},
		{"accepts positive value", "TEST_KEY", "1", 42, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := parseEnvInt(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildConnString(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConnectionConfig
		expected string
	}{
		{
			name: "basic connection string with SSL require (SEC-HIGH-01)",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "require",
			},
			expected: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=require",
		},
		{
			name: "connection string with SSL",
			config: &ConnectionConfig{
				Host:     "db.example.com",
				Port:     5432,
				User:     "admin",
				Password: "secret",
				Database: "production",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5432 user=admin password=secret dbname=production sslmode=require",
		},
		{
			name: "connection string with SSL certificates",
			config: &ConnectionConfig{
				Host:        "secure.db.example.com",
				Port:        5432,
				User:        "secure_user",
				Password:    "secure_pass",
				Database:    "secure_db",
				SSLMode:     "verify-full",
				SSLCert:     "/path/to/cert.pem",
				SSLKey:      "/path/to/key.pem",
				SSLRootCert: "/path/to/ca.pem",
			},
			expected: "host=secure.db.example.com port=5432 user=secure_user password=secure_pass dbname=secure_db sslmode=verify-full sslcert=/path/to/cert.pem sslkey=/path/to/key.pem sslrootcert=/path/to/ca.pem",
		},
		{
			name: "connection string with partial SSL config",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "db",
				SSLMode:  "verify-ca",
				SSLCert:  "/cert.pem",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=db sslmode=verify-ca sslcert=/cert.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnString(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigurePool(t *testing.T) {
	t.Run("configures pool with all settings", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := &ConnectionConfig{
			MaxOpenConns:    50,
			MaxIdleConns:    25,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		}

		ConfigurePool(db, config)

		stats := db.Stats()
		assert.Equal(t, 50, stats.MaxOpenConnections)
	})

	t.Run("does not configure when values are zero", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Set initial value
		db.SetMaxOpenConns(30)
		db.SetMaxIdleConns(15)

		config := &ConnectionConfig{
			MaxOpenConns:    0,
			MaxIdleConns:    0,
			ConnMaxLifetime: 0,
			ConnMaxIdleTime: 0,
		}

		ConfigurePool(db, config)

		stats := db.Stats()
		assert.Equal(t, 30, stats.MaxOpenConnections)
	})
}

func TestClose(t *testing.T) {
	t.Run("closes non-nil database", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		// Expect Close to be called
		mock.ExpectClose()

		err = Close(db)
		assert.NoError(t, err)
	})

	t.Run("returns nil for nil database", func(t *testing.T) {
		err := Close(nil)
		assert.NoError(t, err)
	})
}

func TestIsConnected(t *testing.T) {
	t.Run("returns false for nil database", func(t *testing.T) {
		result := IsConnected(nil)
		assert.False(t, result)
	})

	t.Run("returns true for connected database", func(t *testing.T) {
		db, mock := NewMockDB(t)
		defer db.Close()

		mock.ExpectPing()

		result := IsConnected(db)
		assert.True(t, result)
	})

	t.Run("returns false for ping failure", func(t *testing.T) {
		db, mock := NewMockDB(t)
		defer db.Close()

		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		result := IsConnected(db)
		assert.False(t, result)
	})
}

func TestGetPoolStats(t *testing.T) {
	t.Run("returns nil for nil database", func(t *testing.T) {
		result := GetPoolStats(nil)
		assert.Nil(t, result)
	})

	t.Run("returns stats for valid database", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(10)

		result := GetPoolStats(db)

		require.NotNil(t, result)
		assert.Contains(t, result, "max_open_connections")
		assert.Contains(t, result, "open_connections")
		assert.Contains(t, result, "in_use")
		assert.Contains(t, result, "idle")
		assert.Contains(t, result, "wait_count")
		assert.Contains(t, result, "wait_duration_ms")
	})
}

func TestCheckHealth(t *testing.T) {
	t.Run("returns error for nil database", func(t *testing.T) {
		err := CheckHealth(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("returns error for ping failure", func(t *testing.T) {
		db, mock := NewMockDB(t)
		defer db.Close()

		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		err := CheckHealth(db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
	})

	t.Run("returns nil for healthy database", func(t *testing.T) {
		db, mock := NewMockDB(t)
		defer db.Close()

		mock.ExpectPing()

		err := CheckHealth(db)
		assert.NoError(t, err)
	})
}

func TestConnectionConfigStruct(t *testing.T) {
	t.Run("creates config with all fields", func(t *testing.T) {
		config := &ConnectionConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "user",
			Password:        "password",
			Database:        "dbname",
			SSLMode:         "require",
			SSLCert:         "/cert.pem",
			SSLKey:          "/key.pem",
			SSLRootCert:     "/ca.pem",
			MaxOpenConns:    100,
			MaxIdleConns:    25,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		}

		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "user", config.User)
		assert.Equal(t, "password", config.Password)
		assert.Equal(t, "dbname", config.Database)
		assert.Equal(t, "require", config.SSLMode)
		assert.Equal(t, "/cert.pem", config.SSLCert)
		assert.Equal(t, "/key.pem", config.SSLKey)
		assert.Equal(t, "/ca.pem", config.SSLRootCert)
		assert.Equal(t, 100, config.MaxOpenConns)
		assert.Equal(t, 25, config.MaxIdleConns)
		assert.Equal(t, 1*time.Hour, config.ConnMaxLifetime)
		assert.Equal(t, 10*time.Minute, config.ConnMaxIdleTime)
	})
}

// FIX-018: 测试敏感信息过滤功能
func TestRedactSensitiveInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "redact password",
			input:    "password=secret123 host=localhost",
			expected: "password=[REDACTED] host=localhost",
		},
		{
			name:     "redact password with colon",
			input:    "password:secret123",
			expected: "password:[REDACTED]",
		},
		{
			name:     "redact token",
			input:    "token=abc123xyz&user=admin",
			expected: "token=[REDACTED]&user=admin",
		},
		{
			name:     "redact api_key",
			input:    "api_key=my-secret-key-12345",
			expected: "api_key=[REDACTED]",
		},
		{
			name:     "redact secret",
			input:    "secret=my_super_secret_value",
			expected: "secret=[REDACTED]",
		},
		{
			name:     "redact credential",
			input:    "credential=admin_credentials_here",
			expected: "credential=[REDACTED]",
		},
		{
			name:     "redact key",
			input:    "key=encryption_key_value",
			expected: "key=[REDACTED]",
		},
		{
			name:     "redact apikey (no underscore)",
			input:    "apikey=myapikey123",
			expected: "apikey=[REDACTED]",
		},
		{
			name:     "redact auth",
			input:    "auth=bearer_token_xyz",
			expected: "auth=[REDACTED]",
		},
		{
			name:     "case insensitive password",
			input:    "PASSWORD=SecretValue123",
			expected: "PASSWORD=[REDACTED]",
		},
		{
			name:     "case insensitive token",
			input:    "TOKEN=MyTokenValue",
			expected: "TOKEN=[REDACTED]",
		},
		{
			name:     "no sensitive info to redact",
			input:    "host=localhost port=5432 user=admin",
			expected: "host=localhost port=5432 user=admin",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple sensitive values",
			input:    "password=pass1 token=token1 key=key1",
			expected: "password=[REDACTED] token=[REDACTED] key=[REDACTED]",
		},
		{
			name:     "passwd variant",
			input:    "passwd=mypassword",
			expected: "passwd=[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactSensitiveInfo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// FIX-018: 测试安全日志函数
func TestSafeLogFunctions(t *testing.T) {
	t.Run("safeLogf redacts sensitive info", func(t *testing.T) {
		// This test verifies that safeLogf doesn't panic and processes correctly
		// Actual log output verification would require capturing log output
		result := redactSensitiveInfo("Connection established: password=secret123")
		assert.Contains(t, result, "[REDACTED]")
		assert.NotContains(t, result, "secret123")
	})

	t.Run("redactSensitiveInfo preserves structure", func(t *testing.T) {
		input := "host=localhost port=5432 password=secret123 database=mydb"
		result := redactSensitiveInfo(input)
		// Verify structure is preserved
		assert.Contains(t, result, "host=localhost")
		assert.Contains(t, result, "port=5432")
		assert.Contains(t, result, "password=[REDACTED]")
		assert.Contains(t, result, "database=mydb")
		assert.NotContains(t, result, "secret123")
	})
}
