package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RuleRepository Tests

func TestRuleRepository_NewRuleRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestRuleRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	rule := &model.AlertRule{
		Name:        "High Temperature Alert",
		DeviceType:  "CNC",
		Metric:      "temperature",
		Operator:    ">",
		Threshold:   80.0,
		Severity:    "high",
		Actions:     `[{"type":"email","target":"admin@example.com"}]`,
		Enabled:     true,
		CooldownSec: 300,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectQuery(`INSERT INTO alert_rules`).
		WithArgs(
			rule.Name, rule.DeviceType, rule.Metric, rule.Operator, rule.Threshold,
			rule.Severity, sqlmock.AnyArg(), rule.Enabled, rule.CooldownSec,
			rule.CreatedAt, rule.UpdatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, rule)
	assert.NoError(t, err)
	assert.Equal(t, 1, rule.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	rule := &model.AlertRule{
		Name:       "Test Rule",
		DeviceType: "CNC",
		Metric:     "temperature",
		Operator:   ">",
		Threshold:  80.0,
		Severity:   "high",
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO alert_rules`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	actionsJSON := `[{"type":"email","target":"admin@example.com"}]`

	rows := sqlmock.NewRows([]string{
		"id", "name", "device_type", "metric", "operator", "threshold",
		"severity", "actions", "enabled", "cooldown_sec", "created_at", "updated_at",
	}).AddRow(1, "High Temperature", "CNC", "temperature", ">", 80.0, "high", actionsJSON, true, 300, now, now)

	mock.ExpectQuery(`SELECT .* FROM alert_rules WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	rule, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, 1, rule.ID)
	assert.Equal(t, "High Temperature", rule.Name)
	assert.Equal(t, "CNC", rule.DeviceType)
	assert.True(t, rule.Enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM alert_rules WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	rule, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, rule)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	actionsJSON := `[{"type":"email"}]`

	rows := sqlmock.NewRows([]string{
		"id", "name", "device_type", "metric", "operator", "threshold",
		"severity", "actions", "enabled", "cooldown_sec", "created_at", "updated_at",
	}).
		AddRow(1, "Rule 1", "CNC", "temperature", ">", 80.0, "high", actionsJSON, true, 300, now, now).
		AddRow(2, "Rule 2", "INJ", "pressure", "<", 50.0, "medium", actionsJSON, false, 600, now, now)

	mock.ExpectQuery(`SELECT .* FROM alert_rules ORDER BY created_at DESC`).
		WillReturnRows(rows)

	rules, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "Rule 1", rules[0].Name)
	assert.Equal(t, "Rule 2", rules[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_List_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "name", "device_type", "metric", "operator", "threshold",
		"severity", "actions", "enabled", "cooldown_sec", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT .* FROM alert_rules ORDER BY created_at DESC`).
		WillReturnRows(rows)

	rules, err := repo.List(ctx)
	assert.NoError(t, err)
	// Empty result returns nil slice, not empty slice
	assert.Nil(t, rules)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM alert_rules ORDER BY created_at DESC`).
		WillReturnError(errors.New("database error"))

	rules, err := repo.List(ctx)
	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_ListEnabled_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	actionsJSON := `[{"type":"email"}]`

	rows := sqlmock.NewRows([]string{
		"id", "name", "device_type", "metric", "operator", "threshold",
		"severity", "actions", "enabled", "cooldown_sec", "created_at", "updated_at",
	}).
		AddRow(1, "Enabled Rule", "CNC", "temperature", ">", 80.0, "high", actionsJSON, true, 300, now, now)

	mock.ExpectQuery(`SELECT .* FROM alert_rules WHERE enabled = true ORDER BY created_at DESC`).
		WillReturnRows(rows)

	rules, err := repo.ListEnabled(ctx)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "Enabled Rule", rules[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_ListEnabled_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM alert_rules WHERE enabled = true`).
		WillReturnError(errors.New("database error"))

	rules, err := repo.ListEnabled(ctx)
	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	rule := &model.AlertRule{
		ID:          1,
		Name:        "Updated Rule",
		DeviceType:  "CNC",
		Metric:      "temperature",
		Operator:    ">=",
		Threshold:   85.0,
		Severity:    "critical",
		Actions:     `[{"type":"sms","target":"+1234567890"}]`,
		Enabled:     true,
		CooldownSec: 600,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec(`UPDATE alert_rules SET`).
		WithArgs(
			rule.Name, rule.DeviceType, rule.Metric, rule.Operator, rule.Threshold,
			rule.Severity, sqlmock.AnyArg(), rule.Enabled, rule.CooldownSec,
			sqlmock.AnyArg(), rule.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(ctx, rule)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	rule := &model.AlertRule{
		ID:        999,
		Name:      "Updated Rule",
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec(`UPDATE alert_rules SET`).
		WillReturnError(errors.New("database error"))

	err = repo.Update(ctx, rule)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM alert_rules WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM alert_rules WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(ctx, 999)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_ToggleEnabled_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`UPDATE alert_rules SET enabled = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(false, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.ToggleEnabled(ctx, 1, false)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_ToggleEnabled_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRuleRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`UPDATE alert_rules SET enabled =`).
		WillReturnError(errors.New("database error"))

	err = repo.ToggleEnabled(ctx, 999, true)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// AlertRepository Tests

func TestAlertRepository_NewAlertRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestAlertRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	alert := &model.Alert{
		RuleID:      1,
		DeviceID:    "CNC-001",
		Message:     "Temperature exceeded threshold",
		Severity:    "high",
		Status:      "active",
		TriggeredAt: now,
	}

	mock.ExpectQuery(`INSERT INTO alerts`).
		WithArgs(
			alert.RuleID, alert.DeviceID, alert.Message, alert.Severity,
			alert.Status, alert.TriggeredAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, alert)
	assert.NoError(t, err)
	assert.Equal(t, 1, alert.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	alert := &model.Alert{
		RuleID:      1,
		DeviceID:    "CNC-001",
		Message:     "Test Alert",
		Severity:    "high",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO alerts`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, alert)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at",
	}).
		AddRow(1, 1, "CNC-001", "Alert 1", "high", "active", now, nil).
		AddRow(2, 2, "INJ-001", "Alert 2", "medium", "resolved", now, now)

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE 1=1 ORDER BY triggered_at DESC`).
		WillReturnRows(rows)

	alerts, total, err := repo.List(ctx, "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, alerts, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_List_WithStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	// Count query with status filter
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE 1=1 AND status = \$1`).
		WithArgs("active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query with status filter
	rows := sqlmock.NewRows([]string{
		"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at",
	}).AddRow(1, 1, "CNC-001", "Active Alert", "high", "active", now, nil)

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE 1=1 AND status = \$1 ORDER BY triggered_at DESC`).
		WillReturnRows(rows)

	alerts, total, err := repo.List(ctx, "active", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, alerts, 1)
	assert.Equal(t, "active", alerts[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts`).
		WillReturnError(errors.New("database error"))

	alerts, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, alerts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM alerts`).
		WillReturnError(errors.New("database error"))

	alerts, total, err := repo.List(ctx, "", 1, 10)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, alerts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_CountActive_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.CountActive(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_CountActive_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status = 'active'`).
		WillReturnError(errors.New("database error"))

	count, err := repo.CountActive(ctx)
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_Resolve_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`UPDATE alerts SET status = 'resolved', resolved_at = \$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Resolve(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_Resolve_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`UPDATE alerts SET status = 'resolved'`).
		WillReturnError(errors.New("database error"))

	err = repo.Resolve(ctx, 999)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetRecentByDevice_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at",
	}).AddRow(1, 1, "CNC-001", "Recent Alert", "high", "active", now, nil)

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE device_id = \$1 AND rule_id = \$2 AND triggered_at >`).
		WithArgs("CNC-001", 1, 300).
		WillReturnRows(rows)

	alert, err := repo.GetRecentByDevice(ctx, "CNC-001", 1, 300)
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, "CNC-001", alert.DeviceID)
	assert.Equal(t, 1, alert.RuleID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetRecentByDevice_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE device_id = \$1 AND rule_id = \$2`).
		WithArgs("CNC-001", 999, 300).
		WillReturnError(sql.ErrNoRows)

	alert, err := repo.GetRecentByDevice(ctx, "CNC-001", 999, 300)
	assert.NoError(t, err)
	assert.Nil(t, alert)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetRecentByDevice_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE device_id =`).
		WillReturnError(errors.New("database error"))

	alert, err := repo.GetRecentByDevice(ctx, "CNC-001", 1, 300)
	assert.Error(t, err)
	assert.Nil(t, alert)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetRecentByDevice_WithResolvedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAlertRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	now := time.Now()
	resolvedAt := now.Add(1 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at",
	}).AddRow(1, 1, "CNC-001", "Resolved Alert", "medium", "resolved", now, resolvedAt)

	mock.ExpectQuery(`SELECT .* FROM alerts WHERE device_id = \$1 AND rule_id = \$2`).
		WithArgs("CNC-001", 1, 300).
		WillReturnRows(rows)

	alert, err := repo.GetRecentByDevice(ctx, "CNC-001", 1, 300)
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.NotNil(t, alert.ResolvedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}
