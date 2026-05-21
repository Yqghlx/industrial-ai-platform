package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/repository"
)

func newTestUserService(t *testing.T) (*UserService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)
	return svc, mock
}

func TestNewUserService(t *testing.T) {
	svc, _ := newTestUserService(t)
	assert.NotNil(t, svc)
}

func TestUserService_Authenticate_Success(t *testing.T) {
	svc, mock := newTestUserService(t)
	defer mock.ExpectationsWereMet()

	hashedPassword, err := HashPassword("password123")
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", hashedPassword, "test@example.com", "admin", 1, "tenant-1", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM users WHERE username = .*`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := svc.Authenticate("testuser", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
}

func TestUserService_Authenticate_UserNotFound(t *testing.T) {
	svc, mock := newTestUserService(t)

	mock.ExpectQuery(`SELECT .* FROM users WHERE username = .*`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := svc.Authenticate("nonexistent", "password")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_Authenticate_InvalidPassword(t *testing.T) {
	svc, mock := newTestUserService(t)

	hashedPassword, err := HashPassword("correctpassword")
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", hashedPassword, "test@example.com", "admin", 1, "", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM users WHERE username = .*`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := svc.Authenticate("testuser", "wrongpassword")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "Authentication failed")
}

func TestUserService_GetByID_Success(t *testing.T) {
	svc, mock := newTestUserService(t)

	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", "hash", "test@example.com", "admin", 1, "tenant-1", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM users WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	user, err := svc.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mock := newTestUserService(t)

	mock.ExpectQuery(`SELECT .* FROM users WHERE id = .*`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	user, err := svc.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_GetTokenVersion(t *testing.T) {
	svc, mock := newTestUserService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT token_version FROM users WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"token_version"}).AddRow(3))

	version, err := svc.GetTokenVersion(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, version)
}

func TestUserService_GetTokenVersion_Error(t *testing.T) {
	svc, mock := newTestUserService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT token_version FROM users WHERE id = .*`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	version, err := svc.GetTokenVersion(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, 0, version)
}

func TestUserService_UpdatePassword(t *testing.T) {
	svc, mock := newTestUserService(t)

	mock.ExpectExec(`UPDATE users SET password_hash = .*`).
		WithArgs("newhash", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdatePassword(1, "newhash")
	assert.NoError(t, err)
}

func TestUserService_UpdatePassword_Error(t *testing.T) {
	svc, mock := newTestUserService(t)

	mock.ExpectExec(`UPDATE users SET password_hash = .*`).
		WillReturnError(errors.New("db error"))

	err := svc.UpdatePassword(999, "newhash")
	assert.Error(t, err)
}

func TestUserService_UpdateTokenVersion(t *testing.T) {
	svc, mock := newTestUserService(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET token_version = token_version \+ 1, updated_at = .* WHERE id = .*`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateTokenVersion(ctx, 1)
	assert.NoError(t, err)
}

func TestUserService_UpdateTokenVersion_Error(t *testing.T) {
	svc, mock := newTestUserService(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET token_version = token_version \+ 1`).
		WithArgs(999).
		WillReturnError(errors.New("db error"))

	err := svc.UpdateTokenVersion(ctx, 999)
	assert.Error(t, err)
}

// Mock tests removed - use testify.Mock pattern instead
// See existing tests for proper mock usage
