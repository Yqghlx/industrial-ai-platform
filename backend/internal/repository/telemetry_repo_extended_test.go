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

// WorkOrderRepository Tests

func TestWorkOrderRepository_NewWorkOrderRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestWorkOrderRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	now := time.Now()
	assignedTo := 123
	wo := &model.WorkOrder{
		Title:       "Fix CNC Machine",
		Description: "Temperature sensor malfunction",
		DeviceID:    "CNC-001",
		Priority:    "high",
		Status:      "pending",
		AssignedTo:  &assignedTo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectQuery(`INSERT INTO work_orders`).
		WithArgs(
			wo.Title, wo.Description, wo.DeviceID, wo.Priority, wo.Status,
			wo.AssignedTo, wo.CreatedAt, wo.UpdatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, wo)
	assert.NoError(t, err)
	assert.Equal(t, 1, wo.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	wo := &model.WorkOrder{
		Title:     "Test Work Order",
		DeviceID:  "CNC-001",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO work_orders`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, wo)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	now := time.Now()
	assignedTo := int64(123)

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "device_id", "priority", "status", "assigned_to", "created_at", "updated_at",
	}).AddRow(1, "Fix Machine", "Description", "CNC-001", "high", "pending", assignedTo, now, now)

	mock.ExpectQuery(`SELECT .* FROM work_orders WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	wo, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, wo)
	assert.Equal(t, 1, wo.ID)
	assert.Equal(t, "Fix Machine", wo.Title)
	assert.NotNil(t, wo.AssignedTo)
	assert.Equal(t, 123, *wo.AssignedTo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM work_orders WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	wo, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, wo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "device_id", "priority", "status", "assigned_to", "created_at", "updated_at",
	}).
		AddRow(1, "Work Order 1", "Description 1", "CNC-001", "high", "pending", nil, now, now).
		AddRow(2, "Work Order 2", "Description 2", "INJ-001", "medium", "in_progress", nil, now, now)

	mock.ExpectQuery(`SELECT .* FROM work_orders WHERE 1=1 ORDER BY created_at DESC`).
		WillReturnRows(rows)

	orders, total, err := repo.List(ctx, "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, orders, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query with filters
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE 1=1 AND status = \$1 AND device_id = \$2`).
		WithArgs("pending", "CNC-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query with filters
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "device_id", "priority", "status", "assigned_to", "created_at", "updated_at",
	}).AddRow(1, "Pending Order", "Description", "CNC-001", "high", "pending", nil, now, now)

	mock.ExpectQuery(`SELECT .* FROM work_orders WHERE 1=1 AND status = \$1 AND device_id = \$2 ORDER BY created_at DESC`).
		WillReturnRows(rows)

	orders, total, err := repo.List(ctx, "pending", "CNC-001", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, orders, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders`).
		WillReturnError(errors.New("database error"))

	orders, total, err := repo.List(ctx, "", "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM work_orders`).
		WillReturnError(errors.New("database error"))

	orders, total, err := repo.List(ctx, "", "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_UpdateStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE work_orders SET status = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs("completed", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateStatus(ctx, 1, "completed")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkOrderRepository_UpdateStatus_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE work_orders SET status =`).
		WillReturnError(errors.New("database error"))

	err = repo.UpdateStatus(ctx, 999, "completed")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// NotificationRepository Tests

func TestNotificationRepository_NewNotificationRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestNotificationRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	now := time.Now()
	deviceID := "CNC-001"
	n := &model.Notification{
		Type:      "alert",
		Title:     "High Temperature Alert",
		Message:   "Device temperature exceeded threshold",
		DeviceID:  &deviceID,
		Read:      false,
		CreatedAt: now,
	}

	mock.ExpectQuery(`INSERT INTO notifications`).
		WithArgs(
			n.Type, n.Title, n.Message, n.DeviceID, n.Read, n.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, n)
	assert.NoError(t, err)
	assert.Equal(t, 1, n.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	n := &model.Notification{
		Type:      "alert",
		Title:     "Test Notification",
		Message:   "Test Message",
		Read:      false,
		CreatedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO notifications`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, n)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "type", "title", "message", "device_id", "read", "created_at",
	}).
		AddRow(1, "alert", "Alert 1", "Message 1", "CNC-001", false, now).
		AddRow(2, "info", "Info 1", "Message 2", nil, true, now)

	mock.ExpectQuery(`SELECT .* FROM notifications WHERE 1=1 ORDER BY created_at DESC`).
		WillReturnRows(rows)

	notifications, total, err := repo.List(ctx, "", false, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, notifications, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query with filters
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications WHERE 1=1 AND type = \$1 AND read = false`).
		WithArgs("alert").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query with filters
	rows := sqlmock.NewRows([]string{
		"id", "type", "title", "message", "device_id", "read", "created_at",
	}).AddRow(1, "alert", "Alert", "Message", "CNC-001", false, now)

	mock.ExpectQuery(`SELECT .* FROM notifications WHERE 1=1 AND type = \$1 AND read = false ORDER BY created_at DESC`).
		WillReturnRows(rows)

	notifications, total, err := repo.List(ctx, "alert", true, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, notifications, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications`).
		WillReturnError(errors.New("database error"))

	notifications, total, err := repo.List(ctx, "", false, 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, notifications)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM notifications`).
		WillReturnError(errors.New("database error"))

	notifications, total, err := repo.List(ctx, "", false, 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, notifications)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_MarkRead_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE notifications SET read = true WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.MarkRead(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_MarkRead_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE notifications SET read = true WHERE id = \$1`).
		WillReturnError(errors.New("database error"))

	err = repo.MarkRead(ctx, 999)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// BlackBoxRepository Tests

func TestBlackBoxRepository_NewBlackBoxRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestBlackBoxRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	now := time.Now()
	record := &model.BlackBoxRecord{
		DeviceID:    "CNC-001",
		TriggerType: "alert",
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now,
		Summary:     "Temperature spike detected",
		CreatedAt:   now,
	}

	mock.ExpectQuery(`INSERT INTO blackbox_records`).
		WithArgs(
			record.DeviceID, record.TriggerType, record.StartTime, record.EndTime,
			record.Summary, record.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, record)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), record.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlackBoxRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	record := &model.BlackBoxRecord{
		DeviceID:    "CNC-001",
		TriggerType: "alert",
		StartTime:   time.Now(),
		CreatedAt:   time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO blackbox_records`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, record)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlackBoxRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM blackbox_records`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "device_id", "trigger_type", "start_time", "end_time", "summary", "created_at",
	}).
		AddRow(1, "CNC-001", "alert", now.Add(-2*time.Hour), now.Add(-1*time.Hour), "Alert triggered", now).
		AddRow(2, "INJ-001", "manual", now.Add(-1*time.Hour), now, "Manual snapshot", now)

	mock.ExpectQuery(`SELECT .* FROM blackbox_records ORDER BY created_at DESC`).
		WillReturnRows(rows)

	records, total, err := repo.List(ctx, "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, records, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlackBoxRepository_List_WithDeviceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query with device_id filter
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM blackbox_records WHERE device_id = \$1`).
		WithArgs("CNC-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query with device_id filter
	rows := sqlmock.NewRows([]string{
		"id", "device_id", "trigger_type", "start_time", "end_time", "summary", "created_at",
	}).AddRow(1, "CNC-001", "alert", now.Add(-1*time.Hour), now, "Alert triggered", now)

	mock.ExpectQuery(`SELECT .* FROM blackbox_records WHERE device_id = \$1 ORDER BY created_at DESC`).
		WillReturnRows(rows)

	records, total, err := repo.List(ctx, "CNC-001", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, records, 1)
	assert.Equal(t, "CNC-001", records[0].DeviceID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlackBoxRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM blackbox_records`).
		WillReturnError(errors.New("database error"))

	records, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, records)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlackBoxRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBlackBoxRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM blackbox_records`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM blackbox_records`).
		WillReturnError(errors.New("database error"))

	records, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, records)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ReportRepository Tests

func TestReportRepository_NewReportRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestReportRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	now := time.Now()
	deviceID := "CNC-001"
	report := &model.Report{
		Title:       "Daily Performance Report",
		Type:        "performance",
		DeviceID:    &deviceID,
		Content:     "Report content here",
		GeneratedAt: now,
	}

	mock.ExpectQuery(`INSERT INTO reports`).
		WithArgs(
			report.Title, report.Type, report.DeviceID, report.Content, report.GeneratedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, report)
	assert.NoError(t, err)
	assert.Equal(t, 1, report.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	report := &model.Report{
		Title:       "Test Report",
		Type:        "test",
		Content:     "Test content",
		GeneratedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO reports`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, report)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reports`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "title", "type", "device_id", "content", "generated_at",
	}).
		AddRow(1, "Report 1", "performance", "CNC-001", "Content 1", now).
		AddRow(2, "Report 2", "maintenance", nil, "Content 2", now)

	mock.ExpectQuery(`SELECT .* FROM reports ORDER BY generated_at DESC`).
		WillReturnRows(rows)

	reports, total, err := repo.List(ctx, "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, reports, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_List_WithType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Count query with type filter
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reports WHERE type = \$1`).
		WithArgs("performance").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query with type filter
	rows := sqlmock.NewRows([]string{
		"id", "title", "type", "device_id", "content", "generated_at",
	}).AddRow(1, "Performance Report", "performance", "CNC-001", "Content", now)

	mock.ExpectQuery(`SELECT .* FROM reports WHERE type = \$1 ORDER BY generated_at DESC`).
		WillReturnRows(rows)

	reports, total, err := repo.List(ctx, "performance", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, reports, 1)
	assert.Equal(t, "performance", reports[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reports`).
		WillReturnError(errors.New("database error"))

	reports, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, reports)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReportRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reports`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM reports`).
		WillReturnError(errors.New("database error"))

	reports, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, reports)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// AgentTaskLogRepository Tests

func TestAgentTaskLogRepository_NewAgentTaskLogRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestAgentTaskLogRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	ctx := context.Background()

	now := time.Now()
	log := &model.AgentTaskLog{
		SessionID:  "session-123",
		Query:      "Analyze temperature data",
		Response:   "Temperature is within normal range",
		Agent:      "data-analyst",
		ExecutedAt: now,
	}

	mock.ExpectQuery(`INSERT INTO agent_task_logs`).
		WithArgs(
			log.SessionID, log.Query, log.Response, log.Agent, log.ExecutedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), log.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentTaskLogRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	ctx := context.Background()

	log := &model.AgentTaskLog{
		SessionID:  "session-123",
		Query:      "Test query",
		Response:   "Test response",
		Agent:      "test-agent",
		ExecutedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO agent_task_logs`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, log)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentTaskLogRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	ctx := context.Background()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "session_id", "query", "response", "agent", "executed_at",
	}).
		AddRow(1, "session-1", "Query 1", "Response 1", "agent-1", now).
		AddRow(2, "session-2", "Query 2", "Response 2", "agent-2", now)

	mock.ExpectQuery(`SELECT .* FROM agent_task_logs ORDER BY executed_at DESC LIMIT \$1`).
		WithArgs(10).
		WillReturnRows(rows)

	logs, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, "session-1", logs[0].SessionID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentTaskLogRepository_List_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "session_id", "query", "response", "agent", "executed_at",
	})

	mock.ExpectQuery(`SELECT .* FROM agent_task_logs ORDER BY executed_at DESC LIMIT \$1`).
		WithArgs(10).
		WillReturnRows(rows)

	logs, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.Len(t, logs, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentTaskLogRepository_List_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAgentTaskLogRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM agent_task_logs ORDER BY executed_at DESC LIMIT \$1`).
		WillReturnError(errors.New("database error"))

	logs, err := repo.List(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test unmarshalActions helper
func TestUnmarshalActions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectLen  int
		expectZero bool
	}{
		{
			name:      "valid JSON array",
			input:     `[{"type":"email","target":"admin@example.com"}]`,
			expectLen: 1,
		},
		{
			name:       "empty string",
			input:      "",
			expectZero: true,
		},
		{
			name:       "invalid JSON",
			input:      `invalid json`,
			expectZero: true,
		},
		{
			name:      "multiple actions",
			input:     `[{"type":"email"},{"type":"sms"}]`,
			expectLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalActions(tt.input)
			if tt.expectZero {
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, tt.expectLen)
			}
		})
	}
}
