package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// ============================================
// Integration Test Environment Setup
// ============================================

var (
	testDB       *sql.DB
	testDBURL    string
	testRedisURL string
)

func TestMain(m *testing.M) {
	// Setup test environment
	setupTestEnvironment()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestEnvironment()

	os.Exit(code)
}

func setupTestEnvironment() {
	// Get test database URL from environment or use default
	testDBURL = os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		// Use local PostgreSQL (skip if no database available)
		testDBURL = "postgres://postgres@localhost:5432/test_platform?sslmode=disable"
	}

	// Get Redis URL from environment or use default
	testRedisURL = os.Getenv("TEST_REDIS_URL")
	if testRedisURL == "" {
		testRedisURL = "localhost:6379"
	}

	// Connect to test database
	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		fmt.Printf("⚠️  Skipping integration tests: database not available (%v)\n", err)
		return
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		fmt.Printf("⚠️  Skipping integration tests: database ping failed (%v)\n", err)
		db.Close()
		return
	}

	testDB = db

	// Run migrations for test database
	runTestMigrations()

	fmt.Println("✅ Integration test environment ready")
	fmt.Printf("   Database: %s\n", testDBURL)
	fmt.Printf("   Redis: %s\n", testRedisURL)
}

func cleanupTestEnvironment() {
	if testDB != nil {
		// Clean up test data
		cleanupTestData()
		testDB.Close()
	}
}

func runTestMigrations() {
	// Create test tables if not exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE,
			role VARCHAR(50) DEFAULT 'user',
			tenant_id VARCHAR(255) DEFAULT '1',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			token_version INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS devices (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(100),
			status VARCHAR(50) DEFAULT 'active',
			description TEXT DEFAULT '',
			location VARCHAR(255) DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id SERIAL PRIMARY KEY,
			device_id VARCHAR(255),
			rule_id INTEGER,
			message TEXT,
			severity VARCHAR(50),
			status VARCHAR(50) DEFAULT 'active',
			triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			resolved_at TIMESTAMP,
			acknowledged_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			metric VARCHAR(255),
			operator VARCHAR(10),
			threshold FLOAT,
			severity VARCHAR(50),
			device_type VARCHAR(100),
			actions TEXT,
			enabled BOOLEAN DEFAULT true,
			cooldown_sec INTEGER DEFAULT 300,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS telemetry (
			id SERIAL PRIMARY KEY,
			device_id VARCHAR(255),
			metric VARCHAR(255),
			value FLOAT,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			resource VARCHAR(255),
			action VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id INTEGER REFERENCES roles(id),
			permission_id INTEGER REFERENCES permissions(id),
			PRIMARY KEY (role_id, permission_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id INTEGER REFERENCES users(id),
			role_id INTEGER REFERENCES roles(id),
			PRIMARY KEY (user_id, role_id)
		)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) DEFAULT '1'`,
		`ALTER TABLE devices ALTER COLUMN location SET DEFAULT ''`,
	}

	for _, query := range queries {
		if _, err := testDB.Exec(query); err != nil {
			// Log but don't panic - table might already exist
			fmt.Printf("Migration warning: %v\n", err)
		}
	}
}

func cleanupTestData() {
	// Clean up all test data
	queries := []string{
		"DELETE FROM telemetry",
		"DELETE FROM alerts",
		"DELETE FROM alert_rules",
		"DELETE FROM devices",
		"DELETE FROM user_roles",
		"DELETE FROM role_permissions",
		"DELETE FROM permissions",
		"DELETE FROM roles",
		"DELETE FROM users",
	}

	for _, query := range queries {
		testDB.Exec(query)
	}
}

// ============================================
// Test Helper Functions
// ============================================

func newTestContext(t *testing.T) context.Context {
	return context.Background()
}

// setupGinTestMode is kept for future use
// nolint:unused
func setupGinTestMode() {
	gin.SetMode(gin.TestMode)
}

func truncateAllTables(t *testing.T) {
	queries := []string{
		"TRUNCATE TABLE telemetry CASCADE",
		"TRUNCATE TABLE alerts CASCADE",
		"TRUNCATE TABLE alert_rules CASCADE",
		"TRUNCATE TABLE devices CASCADE",
		"TRUNCATE TABLE user_roles CASCADE",
		"TRUNCATE TABLE role_permissions CASCADE",
		"TRUNCATE TABLE permissions CASCADE",
		"TRUNCATE TABLE roles CASCADE",
		"TRUNCATE TABLE users CASCADE",
	}

	for _, query := range queries {
		_, err := testDB.Exec(query)
		require.NoError(t, err, "Failed to truncate tables")
	}
}

// ============================================
// Integration Test Verification
// ============================================

func TestIntegration_DatabaseConnection(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	require.NotNil(t, testDB, "Test database should be initialized")

	ctx := newTestContext(t)

	// Test basic query
	var result int
	err := testDB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 1, result)

	fmt.Println("✅ Database connection verified")
}

func TestIntegration_TablesExist(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	ctx := newTestContext(t)

	tables := []string{
		"users", "devices", "alerts", "alert_rules",
		"telemetry", "roles", "permissions",
	}

	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = $1
			)
		`
		err := testDB.QueryRowContext(ctx, query, table).Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "Table %s should exist", table)
	}

	fmt.Println("✅ All required tables exist")
}

func TestIntegration_TableOperations(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := newTestContext(t)

	// Test INSERT
	_, err := testDB.ExecContext(ctx,
		"INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)",
		"testuser", "hash123", "test@example.com",
	)
	require.NoError(t, err)

	// Test SELECT
	var username string
	err = testDB.QueryRowContext(ctx,
		"SELECT username FROM users WHERE username = $1",
		"testuser",
	).Scan(&username)
	require.NoError(t, err)
	require.Equal(t, "testuser", username)

	// Test DELETE
	_, err = testDB.ExecContext(ctx, "DELETE FROM users WHERE username = $1", "testuser")
	require.NoError(t, err)

	// Verify deletion
	var count int
	err = testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	fmt.Println("✅ Table operations verified")
}
