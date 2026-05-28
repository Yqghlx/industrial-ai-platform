package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BaseRepository Tests - 纯函数测试

func TestValidateTableName_ValidTables(t *testing.T) {
	// Test all allowed table names
	allowedTables := []string{
		"users",
		"devices",
		"telemetry_data",
		"alerts",
		"alert_rules",
		"work_orders",
		"notifications",
		"blackbox_records",
		"tenants",
		"roles",
		"permissions",
		"user_roles",
		"role_permissions",
		"agent_task_logs",
		"dashboards",
		"widgets",
		"token_blacklist",
		"user_token_versions",
	}

	for _, table := range allowedTables {
		err := ValidateTableName(table)
		assert.NoError(t, err, "Table %s should be valid", table)
	}
}

func TestValidateTableName_Empty(t *testing.T) {
	err := ValidateTableName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name cannot be empty")
}

func TestValidateTableName_NotAllowed(t *testing.T) {
	invalidTables := []string{
		"invalid_table",
		"malicious_table",
		"sql_injection",
		"another_invalid",
	}

	for _, table := range invalidTables {
		err := ValidateTableName(table)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{"Page1_Size10", 1, 10, 0},
		{"Page2_Size10", 2, 10, 10},
		{"Page3_Size20", 3, 20, 40},
		{"Page0_Size10", 0, 10, 0},
		{"Page-1_Size10", -1, 10, 0},
		{"Page1_Size0", 1, 0, 0},
		{"Page1_Size-1", 1, -1, 0},
		{"Page5_Size50", 5, 50, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := CalculateOffset(tt.page, tt.pageSize)
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedPage   int
		expectedSize   int
	}{
		{"Valid_1_10", 1, 10, 1, 10},
		{"Valid_2_20", 2, 20, 2, 20},
		{"InvalidPage_0_10", 0, 10, 1, 10},
		{"InvalidPage_-1_10", -1, 10, 1, 10},
		{"InvalidSize_1_0", 1, 0, 1, 50},
		{"InvalidSize_1_-1", 1, -1, 1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, pageSize := NormalizePagination(tt.page, tt.pageSize)
			assert.Equal(t, tt.expectedPage, page)
			assert.Equal(t, tt.expectedSize, pageSize)
		})
	}
}

func TestNormalizeLimit(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected int
	}{
		{"Valid_10", 10, 10},
		{"Valid_100", 100, 100},
		{"Invalid_0", 0, 50},
		{"Invalid_-1", -1, 50},
		{"TooLarge_10000", 10000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := NormalizeLimit(tt.limit)
			assert.Equal(t, tt.expected, limit)
		})
	}
}

// BaseRepository Tests - 数据库操作测试

func TestNewBaseRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewBaseRepository(dbWrapper, "users")
	assert.NotNil(t, repo)
}

func TestBaseRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewBaseRepository(dbWrapper, "users")
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBaseRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewBaseRepository(dbWrapper, "users")
	ctx := context.Background()

	// Exists = true
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(ctx, "id", 1)
	require.NoError(t, err)
	assert.True(t, exists)

	// Exists = false
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = repo.Exists(ctx, "id", 999)
	require.NoError(t, err)
	assert.False(t, exists)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBaseRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewBaseRepository(dbWrapper, "users")
	ctx := context.Background()

	// Valid delete
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, "id", 1)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBaseRepository_HandleNoRows(t *testing.T) {
	err := sql.ErrNoRows
	result := HandleNoRows(err, "user", "123")
	require.Error(t, result)
	assert.Contains(t, result.Error(), "user not found")

	err = errors.New("other error")
	result = HandleNoRows(err, "user", "123")
	require.Error(t, result)
	assert.Contains(t, result.Error(), "other error")
}