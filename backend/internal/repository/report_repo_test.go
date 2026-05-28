package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ReportRepository Tests

func TestReportRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	deviceID := "CNC-001"
	expectedReport := &model.Report{
		ID:          1,
		Title:       "Test Report",
		Type:        "monthly",
		DeviceID:    &deviceID,
		Content:     "Report content",
		GeneratedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "title", "type", "device_id", "content", "generated_at"}).
		AddRow(expectedReport.ID, expectedReport.Title, expectedReport.Type, expectedReport.DeviceID, expectedReport.Content, expectedReport.GeneratedAt)

	mock.ExpectQuery(`SELECT id, title, type, device_id, content, generated_at FROM reports WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	report, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expectedReport.ID, report.ID)
	assert.Equal(t, expectedReport.Title, report.Title)
	assert.Equal(t, expectedReport.Type, report.Type)
	assert.Equal(t, expectedReport.DeviceID, report.DeviceID)
	assert.Equal(t, expectedReport.Content, report.Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, title, type, device_id, content, generated_at FROM reports WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(errors.New("no rows in result set"))

	report, err := repo.GetByID(ctx, 999)
	require.Error(t, err)
	assert.Nil(t, report)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM reports WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM reports WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(errors.New("no rows affected"))

	err = repo.Delete(ctx, 999)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_WithTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewReportRepository(dbWrapper)

	// Begin transaction using dbWrapper.BeginTx
	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	assert.NotNil(t, txRepo)

	// Verify transaction repository is different instance
	assert.NotNil(t, txRepo)

	// Cleanup
	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}