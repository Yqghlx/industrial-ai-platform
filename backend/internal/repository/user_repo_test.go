package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UserRepository Tests

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()
	user := &model.User{
		Username:  "testuser",
		Password:  "hashed_password_123",
		Email:     "test@example.com",
		Role:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Expect INSERT with RETURNING id
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(
			"testuser",
			"hashed_password_123",
			"test@example.com",
			"admin",
			now,
			now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Execute Create
	err = repo.Create(ctx, user)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:  "testuser",
		Password:  "hashed_password",
		Email:     "test@example.com",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Expect INSERT returning error
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(errors.New("duplicate key value"))

	// Execute Create
	err = repo.Create(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", "hashed_password", "test@example.com", "admin", 1, "tenant-001", now, now)

	mock.ExpectQuery(`SELECT .* FROM users WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	// Execute GetByID
	user, err := repo.GetByID(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "admin", user.Role)
	assert.Equal(t, 1, user.TokenVersion)
	assert.Equal(t, "tenant-001", user.TenantID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning no rows
	mock.ExpectQuery(`SELECT .* FROM users WHERE id = .*`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Execute GetByID
	user, err := repo.GetByID(ctx, 999)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", "hashed_password", "test@example.com", "admin", 0, "", now, now)

	mock.ExpectQuery(`SELECT .* FROM users WHERE username = .*`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Execute GetByUsername
	user, err := repo.GetByUsername(ctx, "testuser")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning no rows
	mock.ExpectQuery(`SELECT .* FROM users WHERE username = .*`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	// Execute GetByUsername
	user, err := repo.GetByUsername(ctx, "nonexistent")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
		AddRow(1, "testuser", "hashed_password", "test@example.com", "viewer", 2, "tenant-002", now, now)

	mock.ExpectQuery(`SELECT .* FROM users WHERE email = .*`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	// Execute GetByEmail
	user, err := repo.GetByEmail(ctx, "test@example.com")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning no rows
	mock.ExpectQuery(`SELECT .* FROM users WHERE email = .*`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	// Execute GetByEmail
	user, err := repo.GetByEmail(ctx, "nonexistent@example.com")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Expect SELECT query with pagination
	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "created_at", "updated_at"})
	rows.AddRow(1, "admin", "hash1", "admin@example.com", "admin", now, now)
	rows.AddRow(2, "operator", "hash2", "operator@example.com", "operator", now, now)
	rows.AddRow(3, "viewer", "hash3", "viewer@example.com", "viewer", now, now)

	mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT .* OFFSET .*`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute List (page=1, pageSize=10)
	users, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, 3, total)
	assert.Equal(t, "admin", users[0].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect SELECT query returning empty rows
	rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT .* OFFSET .*`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute List
	users, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Empty(t, users)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect COUNT query returning error
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(errors.New("database error"))

	// Execute List
	users, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, users)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Expect SELECT query returning error
	mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT .* OFFSET .*`).
		WithArgs(10, 0).
		WillReturnError(errors.New("query failed"))

	// Execute List
	users, total, err := repo.List(ctx, 1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	assert.Nil(t, users)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()
	user := &model.User{
		ID:        1,
		Username:  "updateduser",
		Email:     "updated@example.com",
		Role:      "operator",
		UpdatedAt: now,
	}

	// Expect UPDATE query
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(
			"updateduser",
			"updated@example.com",
			"operator",
			sqlmock.AnyArg(), // UpdatedAt is set in the function
			1,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute Update
	err = repo.Update(ctx, user)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		ID:       999,
		Username: "nonexistent",
		Email:    "nonexistent@example.com",
		Role:     "viewer",
	}

	// Expect UPDATE query returning error
	mock.ExpectExec(`UPDATE users SET`).
		WillReturnError(errors.New("database error"))

	// Execute Update
	err = repo.Update(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect UPDATE password query
	mock.ExpectExec(`UPDATE users SET password_hash = .* updated_at = .* WHERE id = .*`).
		WithArgs("new_hashed_password", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute UpdatePassword
	err = repo.UpdatePassword(ctx, 1, "new_hashed_password")

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect UPDATE password query returning error
	mock.ExpectExec(`UPDATE users SET password_hash`).
		WillReturnError(errors.New("database error"))

	// Execute UpdatePassword
	err = repo.UpdatePassword(ctx, 999, "new_hashed_password")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect DELETE query
	mock.ExpectExec(`DELETE FROM users WHERE id = .*`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute Delete
	err = repo.Delete(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect DELETE query returning error
	mock.ExpectExec(`DELETE FROM users WHERE id = .*`).
		WillReturnError(errors.New("database error"))

	// Execute Delete
	err = repo.Delete(ctx, 999)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Count_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Execute Count
	count, err := repo.Count(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Count_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect COUNT query returning error
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(errors.New("database error"))

	// Execute Count
	count, err := repo.Count(ctx)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetTokenVersion_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect SELECT token_version query
	rows := sqlmock.NewRows([]string{"token_version"}).AddRow(3)

	mock.ExpectQuery(`SELECT token_version FROM users WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	// Execute GetTokenVersion
	version, err := repo.GetTokenVersion(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 3, version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetTokenVersion_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect SELECT query returning no rows
	mock.ExpectQuery(`SELECT token_version FROM users WHERE id = .*`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Execute GetTokenVersion
	version, err := repo.GetTokenVersion(ctx, 999)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, 0, version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateTokenVersion_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect UPDATE token_version query (increment)
	mock.ExpectExec(`UPDATE users SET token_version = token_version \+ 1, updated_at = .* WHERE id = .*`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute UpdateTokenVersion
	err = repo.UpdateTokenVersion(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateTokenVersion_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Expect UPDATE token_version query returning error
	mock.ExpectExec(`UPDATE users SET token_version`).
		WillReturnError(errors.New("database error"))

	// Execute UpdateTokenVersion
	err = repo.UpdateTokenVersion(ctx, 999)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// SQL Pattern Tests for UserRepository

func TestUserRepository_SQLQueryPatterns(t *testing.T) {
	t.Run("Create SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(db)
		ctx := context.Background()

		user := &model.User{
			Username:  "patternuser",
			Password:  "hash",
			Email:     "pattern@example.com",
			Role:      "viewer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery("INSERT INTO users").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err = repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByID SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(db)
		ctx := context.Background()

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "testuser", "hash", "test@example.com", "admin", 0, "", now, now)

		mock.ExpectQuery("SELECT .* FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		user, err := repo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(db)
		ctx := context.Background()

		user := &model.User{
			ID:       1,
			Username: "updated",
			Email:    "updated@example.com",
			Role:     "admin",
		}

		mock.ExpectExec("UPDATE users SET username = .*, email = .*, role = .*, updated_at = .* WHERE id").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Update(ctx, user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete SQL pattern", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(db)
		ctx := context.Background()

		mock.ExpectExec("DELETE FROM users WHERE id").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Delete(ctx, 1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
