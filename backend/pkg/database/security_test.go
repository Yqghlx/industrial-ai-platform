package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProductionConfig(t *testing.T) {
	config := DefaultProductionConfig()

	require.NotNil(t, config)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 1*time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, config.ConnMaxIdleTime)
}

func TestDatabaseConfigStruct(t *testing.T) {
	config := &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "testuser",
		Password:        "testpass",
		Database:        "testdb",
		SSLMode:         "verify-full",
		SSLCert:         "/path/to/cert.pem",
		SSLKey:          "/path/to/key.pem",
		SSLRootCert:     "/path/to/ca.pem",
		MaxOpenConns:    100,
		MaxIdleConns:    25,
		ConnMaxLifetime: 2 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, "testpass", config.Password)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "verify-full", config.SSLMode)
	assert.Equal(t, "/path/to/cert.pem", config.SSLCert)
	assert.Equal(t, "/path/to/key.pem", config.SSLKey)
	assert.Equal(t, "/path/to/ca.pem", config.SSLRootCert)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 25, config.MaxIdleConns)
	assert.Equal(t, 2*time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, config.ConnMaxIdleTime)
}

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *DatabaseConfig
		expected string
	}{
		{
			name: "basic connection string",
			config: &DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
		},
		{
			name: "connection string with require SSL",
			config: &DatabaseConfig{
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
			config: &DatabaseConfig{
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
			name: "connection string with only SSL cert",
			config: &DatabaseConfig{
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
		{
			name: "connection string with only SSL key",
			config: &DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "db",
				SSLMode:  "require",
				SSLKey:   "/key.pem",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=db sslmode=require sslkey=/key.pem",
		},
		{
			name: "connection string with only root cert",
			config: &DatabaseConfig{
				Host:        "localhost",
				Port:        5432,
				User:        "user",
				Password:    "pass",
				Database:    "db",
				SSLMode:     "require",
				SSLRootCert: "/ca.pem",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=db sslmode=require sslrootcert=/ca.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		expectWarnings  bool
		warningPatterns []string
	}{
		{
			name:           "safe query",
			query:          "SELECT id, name FROM users WHERE id = $1",
			expectWarnings: false,
		},
		{
			name:           "safe INSERT query",
			query:          "INSERT INTO users (name, email) VALUES ($1, $2)",
			expectWarnings: false,
		},
		{
			name:           "safe UPDATE query",
			query:          "UPDATE users SET name = $1 WHERE id = $2",
			expectWarnings: false,
		},
		{
			name:            "DROP TABLE injection",
			query:           "SELECT * FROM users; DROP TABLE users;",
			expectWarnings:  true,
			warningPatterns: []string{"DROP TABLE"},
		},
		{
			name:            "UNION SELECT injection",
			query:           "SELECT * FROM users WHERE id = 1 UNION SELECT * FROM passwords",
			expectWarnings:  true,
			warningPatterns: []string{"UNION SELECT"},
		},
		{
			name:            "OR 1=1 injection",
			query:           "SELECT * FROM users WHERE id = 1 OR 1=1",
			expectWarnings:  true,
			warningPatterns: []string{"OR 1=1"},
		},
		{
			name:            "AND 1=1 injection",
			query:           "SELECT * FROM users WHERE id = 1 AND 1=1",
			expectWarnings:  true,
			warningPatterns: []string{"AND 1=1"},
		},
		{
			name:            "SQL comment injection",
			query:           "SELECT * FROM users WHERE id = 1 -- AND deleted = false",
			expectWarnings:  true,
			warningPatterns: []string{"--"},
		},
		{
			name:            "multiline comment injection",
			query:           "SELECT * FROM users WHERE id = 1 /* comment */",
			expectWarnings:  true,
			warningPatterns: []string{"/*"},
		},
		{
			name:            "TRUNCATE TABLE injection",
			query:           "TRUNCATE TABLE users",
			expectWarnings:  true,
			warningPatterns: []string{"TRUNCATE TABLE"},
		},
		{
			name:            "DELETE FROM injection",
			query:           "DELETE FROM users WHERE 1=1",
			expectWarnings:  true,
			warningPatterns: []string{"DELETE FROM"},
		},
		{
			name:           "case insensitive detection - drop table",
			query:          "SELECT * FROM users; drop table users;",
			expectWarnings: true,
		},
		{
			name:           "case insensitive detection - union select",
			query:          "SELECT * FROM users union select * FROM passwords",
			expectWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateQuery(tt.query)

			if tt.expectWarnings {
				assert.NotEmpty(t, warnings, "Expected warnings for query: %s", tt.query)
			} else {
				assert.Empty(t, warnings, "Expected no warnings for safe query: %s", tt.query)
			}

			// Verify specific warning patterns are detected
			for _, pattern := range tt.warningPatterns {
				found := false
				for _, warning := range warnings {
					if containsIgnoreCase(warning, pattern) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected warning containing pattern '%s'", pattern)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		expect bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "Hello", true},
		{"Hello World", "hello", true},
		{"Hello World", "lo wo", true},
		{"Hello World", "xyz", false},
		{"", "", true},
		{"a", "", true},
		{"", "a", false},
		{"ABC", "abc", true},
		{"abc", "ABC", true},
		{"AbCdEf", "bCd", true},
		{"AbCdEf", "BcD", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestContainsIgnoreCaseHelper(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		expect bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"ABCDEF", "bcd", true},
		{"ABCDEF", "BCD", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := containsIgnoreCaseHelper(tt.s, tt.substr)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestCheckSSLStatus(t *testing.T) {
	t.Run("returns SSL status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"ssl", "ssl_version", "ssl_cipher"}).
			AddRow(true, "TLSv1.3", "TLS_AES_256_GCM_SHA384")

		mock.ExpectQuery("SELECT ssl, ssl_version, ssl_cipher FROM pg_stat_ssl").
			WillReturnRows(rows)

		ssl, version, cipher, err := CheckSSLStatus(db)

		require.NoError(t, err)
		assert.True(t, ssl)
		assert.Equal(t, "TLSv1.3", version)
		assert.Equal(t, "TLS_AES_256_GCM_SHA384", cipher)
	})

	t.Run("returns false when SSL disabled", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"ssl", "ssl_version", "ssl_cipher"}).
			AddRow(false, "", "")

		mock.ExpectQuery("SELECT ssl, ssl_version, ssl_cipher FROM pg_stat_ssl").
			WillReturnRows(rows)

		ssl, version, cipher, err := CheckSSLStatus(db)

		require.NoError(t, err)
		assert.False(t, ssl)
		assert.Empty(t, version)
		assert.Empty(t, cipher)
	})

	t.Run("handles query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT ssl, ssl_version, ssl_cipher FROM pg_stat_ssl").
			WillReturnError(sql.ErrConnDone)

		ssl, version, cipher, err := CheckSSLStatus(db)

		require.Error(t, err)
		assert.False(t, ssl)
		assert.Empty(t, version)
		assert.Empty(t, cipher)
		assert.Contains(t, err.Error(), "failed to check SSL status")
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("returns health check results", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		sslRows := sqlmock.NewRows([]string{"ssl", "ssl_version", "ssl_cipher"}).
			AddRow(true, "TLSv1.3", "TLS_AES_256_GCM_SHA384")
		mock.ExpectQuery("SELECT ssl, ssl_version, ssl_cipher FROM pg_stat_ssl").
			WillReturnRows(sslRows)

		versionRows := sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 15.2")
		mock.ExpectQuery("SELECT version\\(\\)").
			WillReturnRows(versionRows)

		result := HealthCheck(db)

		assert.True(t, result["connected"].(bool))
		assert.True(t, result["ssl_enabled"].(bool))
		assert.Equal(t, "TLSv1.3", result["ssl_version"])
		assert.Equal(t, "TLS_AES_256_GCM_SHA384", result["ssl_cipher"])
		assert.Equal(t, "PostgreSQL 15.2", result["version"])
	})

	t.Run("handles ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		result := HealthCheck(db)

		assert.False(t, result["connected"].(bool))
	})

	t.Run("handles SSL query failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		mock.ExpectQuery("SELECT ssl, ssl_version, ssl_cipher FROM pg_stat_ssl").
			WillReturnError(sql.ErrConnDone)

		versionRows := sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 15.2")
		mock.ExpectQuery("SELECT version\\(\\)").
			WillReturnRows(versionRows)

		result := HealthCheck(db)

		assert.True(t, result["connected"].(bool))
		assert.Nil(t, result["ssl_enabled"])
	})
}

func TestSecureConnectionConfigPool(t *testing.T) {
	t.Run("configureConnectionPool applies all settings", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := &DatabaseConfig{
			MaxOpenConns:    100,
			MaxIdleConns:    25,
			ConnMaxLifetime: 2 * time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		}

		configureConnectionPool(db, config)

		stats := db.Stats()
		assert.Equal(t, 100, stats.MaxOpenConnections)
	})

	t.Run("configureConnectionPool skips zero values", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Set initial values
		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(10)

		config := &DatabaseConfig{
			MaxOpenConns:    0,
			MaxIdleConns:    0,
			ConnMaxLifetime: 0,
			ConnMaxIdleTime: 0,
		}

		configureConnectionPool(db, config)

		stats := db.Stats()
		assert.Equal(t, 50, stats.MaxOpenConnections)
	})
}

func TestVerifyConnection(t *testing.T) {
	t.Run("returns nil for successful ping", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		err = verifyConnection(db)
		assert.NoError(t, err)
	})

	t.Run("returns error for ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		err = verifyConnection(db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}
