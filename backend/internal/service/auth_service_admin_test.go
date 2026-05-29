package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/pkg/database"

	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// EnsureDefaultAdmin 测试
// ============================================

// TestAuthService_EnsureDefaultAdmin_AlreadyExists 测试 admin 已存在时直接返回
func TestAuthService_EnsureDefaultAdmin_AlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Mock: GetByUsername 找到已存在的 admin 用户
	hashedPassword, err := HashPassword("existing_password")
	require.NoError(t, err)

	mock.ExpectQuery(userQueryPattern).
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "admin", hashedPassword, "admin@industrial.ai", "admin", 0, "", time.Now(), time.Now()))

	// 执行
	err = authService.EnsureDefaultAdmin(ctx, "newpassword")

	// 断言: admin 已存在，不应报错，不应创建新用户
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAuthService_EnsureDefaultAdmin_NotExists 测试 admin 不存在时创建
func TestAuthService_EnsureDefaultAdmin_NotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Mock: GetByUsername 返回 admin 不存在
	mock.ExpectQuery(userQueryPattern).
		WithArgs("admin").
		WillReturnError(sql.ErrNoRows)

	// Mock: Create 创建 admin 用户成功
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("admin", sqlmock.AnyArg(), "admin@industrial.ai", "admin", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// 执行
	err = authService.EnsureDefaultAdmin(ctx, "admin123")

	// 断言: 成功创建 admin
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAuthService_EnsureDefaultAdmin_CreateError 测试创建 admin 失败时的错误处理
func TestAuthService_EnsureDefaultAdmin_CreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Mock: GetByUsername 返回 admin 不存在
	mock.ExpectQuery(userQueryPattern).
		WithArgs("admin").
		WillReturnError(sql.ErrNoRows)

	// Mock: Create 创建 admin 失败（数据库错误）
	mock.ExpectQuery("INSERT INTO users").
		WillReturnError(errors.New("database connection lost"))

	// 执行
	err = authService.EnsureDefaultAdmin(ctx, "admin123")

	// 断言: 创建失败应返回错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection lost")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAuthService_EnsureDefaultAdmin_EmptyPassword 测试空密码时自动生成
func TestAuthService_EnsureDefaultAdmin_EmptyPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Mock: GetByUsername 返回 admin 不存在
	mock.ExpectQuery(userQueryPattern).
		WithArgs("admin").
		WillReturnError(sql.ErrNoRows)

	// Mock: Create 创建 admin 用户成功（空密码时自动生成随机密码）
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("admin", sqlmock.AnyArg(), "admin@industrial.ai", "admin", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// 执行: 传入空密码
	err = authService.EnsureDefaultAdmin(ctx, "")

	// 断言: 空密码应自动生成，不应报错
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
