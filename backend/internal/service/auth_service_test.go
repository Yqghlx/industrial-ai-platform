package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain initializes JWT for all tests
func TestMain(m *testing.M) {
	// Initialize JWT with a test secret before running any tests
	testSecret := "test-jwt-secret-for-unit-tests-min-32-chars"
	if err := InitJWT(testSecret); err != nil {
		panic("Failed to initialize JWT for tests: " + err.Error())
	}
	os.Exit(m.Run())
}

// Standard columns for user queries (with token_version and tenant_id)
// nolint:unused
const userQueryColumns = "id, username, password_hash, email, role, COALESCE(token_version, 0), COALESCE(tenant_id, ''), created_at, updated_at"
const userQueryPattern = "SELECT.*FROM users WHERE"

func TestAuthService_Login_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Mock password hash for "password123"
	hashedPassword, err := HashPassword("password123")
	require.NoError(t, err)

	// Expect query for GetByUsername - includes token_version and tenant_id columns
	mock.ExpectQuery(userQueryPattern).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "testuser", hashedPassword, "test@example.com", "user", 0, "tenant-001", time.Now(), time.Now()))

	// Execute login
	user, token, err := authService.Login(ctx, "testuser", "password123")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "user", user.Role)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_InvalidUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Expect query for GetByUsername returning no rows
	mock.ExpectQuery(userQueryPattern).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	// Execute login
	user, token, err := authService.Login(ctx, "nonexistent", "password123")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication failed")
	assert.Nil(t, user)
	assert.Empty(t, token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Hash for correct password
	hashedPassword, err := HashPassword("correctpassword")
	require.NoError(t, err)

	// Expect query for GetByUsername - includes token_version and tenant_id columns
	mock.ExpectQuery(userQueryPattern).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "testuser", hashedPassword, "test@example.com", "user", 0, "tenant-001", time.Now(), time.Now()))

	// Execute login with wrong password
	user, token, err := authService.Login(ctx, "testuser", "wrongpassword")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication failed")
	assert.Nil(t, user)
	assert.Empty(t, token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Register_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	req := &model.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "new@example.com",
		Role:     "user",
	}

	// Expect GetByUsername to return no rows (user doesn't exist)
	mock.ExpectQuery(userQueryPattern).
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	// Expect GetByEmail to return no rows (email doesn't exist)
	mock.ExpectQuery("SELECT.*FROM users WHERE email").
		WithArgs("new@example.com").
		WillReturnError(sql.ErrNoRows)

	// Expect Create to succeed
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("newuser", sqlmock.AnyArg(), "new@example.com", "user", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Execute register
	user, token, err := authService.Register(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "user", user.Role)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	req := &model.RegisterRequest{
		Username: "existinguser",
		Password: "password123",
		Email:    "existing@example.com",
		Role:     "user",
	}

	// Expect GetByUsername to return existing user
	hashedPassword, _ := HashPassword("password123")
	mock.ExpectQuery(userQueryPattern).
		WithArgs("existinguser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "existinguser", hashedPassword, "existing@example.com", "user", 0, "tenant-001", time.Now(), time.Now()))

	// Execute register
	user, token, err := authService.Register(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Username already exists")
	assert.Nil(t, user)
	assert.Empty(t, token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	req := &model.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "existing@example.com",
		Role:     "user",
	}

	// Expect GetByUsername to return no rows
	mock.ExpectQuery(userQueryPattern).
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	// Expect GetByEmail to return existing user
	hashedPassword, _ := HashPassword("password123")
	mock.ExpectQuery("SELECT.*FROM users WHERE email").
		WithArgs("existing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "anotheruser", hashedPassword, "existing@example.com", "user", 0, "tenant-001", time.Now(), time.Now()))

	// Execute register
	user, token, err := authService.Register(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Email already exists")
	assert.Nil(t, user)
	assert.Empty(t, token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	hashedPassword, _ := HashPassword("password123")

	// Expect query for GetByID
	mock.ExpectQuery("SELECT.*FROM users WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "email", "role", "token_version", "tenant_id", "created_at", "updated_at"}).
			AddRow(1, "testuser", hashedPassword, "test@example.com", "user", 0, "tenant-001", time.Now(), time.Now()))

	// Execute GetUserByID
	user, err := authService.GetUserByID(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "testuser", user.Username)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)
	ctx := context.Background()

	// Expect query for GetByID returning no rows
	mock.ExpectQuery("SELECT.*FROM users WHERE id").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Execute GetUserByID
	user, err := authService.GetUserByID(ctx, 999)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper function tests
func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestVerifyPassword_Success(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	result := VerifyPassword(password, hash)
	assert.True(t, result)
}

func TestVerifyPassword_Failure(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	result := VerifyPassword("wrongpassword", hash)
	assert.False(t, result)
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1, "testuser", "admin")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken_Success(t *testing.T) {
	// Generate token
	token, err := GenerateToken(1, "testuser", "admin")
	require.NoError(t, err)

	// Parse token
	claims, err := ParseToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

func TestParseToken_InvalidToken(t *testing.T) {
	claims, err := ParseToken("invalid-token-string")

	assert.Error(t, err)
	assert.Nil(t, claims)
}
