package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/pkg/database"
)

// ============================================
// WithTx Tests - All Repositories
// ============================================

func TestDeviceRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewDeviceRepository(dbWrapper)

	// Begin transaction
	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	// WithTx should return a new repository with the transaction
	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	// The txRepo should use the transaction
	assert.NotNil(t, txRepo.(*DeviceRepository).db)

	// Cleanup
	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewAlertRepository(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTelemetryRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewTelemetryRepository(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewUserRepository(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRBACRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewRBACRepository(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewRuleRepository(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}
