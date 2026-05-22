package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

// RuleRepository handles alert rule data access
type RuleRepository struct {
	db database.QueryExecutor
}

// NewRuleRepository creates a new rule repository
func NewRuleRepository(db database.QueryExecutor) *RuleRepository {
	return &RuleRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *RuleRepository) WithTx(tx database.TransactionInterface) *RuleRepository {
	return &RuleRepository{db: tx}
}

// Create creates a new alert rule
func (r *RuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	query := `
		INSERT INTO alert_rules (name, device_type, metric, operator, threshold, severity, actions, enabled, cooldown_sec, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	actionsJSON, _ := json.Marshal(rule.Actions)
	return r.db.QueryRow(ctx, query,
		rule.Name, rule.DeviceType, rule.Metric, rule.Operator, rule.Threshold,
		rule.Severity, string(actionsJSON), rule.Enabled, rule.CooldownSec,
		rule.CreatedAt, rule.UpdatedAt,
	).Scan(&rule.ID)
}

// GetByID retrieves an alert rule by ID
func (r *RuleRepository) GetByID(ctx context.Context, id int) (*model.AlertRule, error) {
	query := `
		SELECT id, name, device_type, metric, operator, threshold, severity, actions, enabled, cooldown_sec, created_at, updated_at
		FROM alert_rules WHERE id = $1
	`
	rule := &model.AlertRule{}
	var actionsJSON string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.DeviceType, &rule.Metric, &rule.Operator,
		&rule.Threshold, &rule.Severity, &actionsJSON, &rule.Enabled, &rule.CooldownSec,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(actionsJSON), &rule.Actions)
	return rule, nil
}

// List retrieves all alert rules
func (r *RuleRepository) List(ctx context.Context) ([]model.AlertRule, error) {
	query := `
		SELECT id, name, device_type, metric, operator, threshold, severity, actions, enabled, cooldown_sec, created_at, updated_at
		FROM alert_rules ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.AlertRule
	for rows.Next() {
		var rule model.AlertRule
		var actionsJSON string
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.DeviceType, &rule.Metric, &rule.Operator,
			&rule.Threshold, &rule.Severity, &actionsJSON, &rule.Enabled, &rule.CooldownSec,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(actionsJSON), &rule.Actions)
		rules = append(rules, rule)
	}

	return rules, nil
}

// ListEnabled retrieves all enabled alert rules
func (r *RuleRepository) ListEnabled(ctx context.Context) ([]model.AlertRule, error) {
	query := `
		SELECT id, name, device_type, metric, operator, threshold, severity, actions, enabled, cooldown_sec, created_at, updated_at
		FROM alert_rules WHERE enabled = true ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.AlertRule
	for rows.Next() {
		var rule model.AlertRule
		var actionsJSON string
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.DeviceType, &rule.Metric, &rule.Operator,
			&rule.Threshold, &rule.Severity, &actionsJSON, &rule.Enabled, &rule.CooldownSec,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(actionsJSON), &rule.Actions)
		rules = append(rules, rule)
	}

	return rules, nil
}

// Update updates an alert rule
func (r *RuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	query := `
		UPDATE alert_rules SET
			name = $1, device_type = $2, metric = $3, operator = $4, threshold = $5,
			severity = $6, actions = $7, enabled = $8, cooldown_sec = $9, updated_at = $10
		WHERE id = $11
	`
	actionsJSON, _ := json.Marshal(rule.Actions)
	rule.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		rule.Name, rule.DeviceType, rule.Metric, rule.Operator, rule.Threshold,
		rule.Severity, string(actionsJSON), rule.Enabled, rule.CooldownSec,
		rule.UpdatedAt, rule.ID,
	)
	return err
}

// Delete removes an alert rule
func (r *RuleRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM alert_rules WHERE id = $1", id)
	return err
}

// ToggleEnabled enables or disables a rule
func (r *RuleRepository) ToggleEnabled(ctx context.Context, id int, enabled bool) error {
	_, err := r.db.Exec(ctx,
		"UPDATE alert_rules SET enabled = $1, updated_at = $2 WHERE id = $3",
		enabled, time.Now(), id,
	)
	return err
}

// AlertRepository handles alert data access
type AlertRepository struct {
	db database.QueryExecutor
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db database.QueryExecutor) *AlertRepository {
	return &AlertRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *AlertRepository) WithTx(tx database.TransactionInterface) *AlertRepository {
	return &AlertRepository{db: tx}
}

// Create creates a new alert
func (r *AlertRepository) Create(ctx context.Context, alert *model.Alert) error {
	query := `
		INSERT INTO alerts (rule_id, device_id, message, severity, status, triggered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		alert.RuleID, alert.DeviceID, alert.Message, alert.Severity,
		alert.Status, alert.TriggeredAt,
	).Scan(&alert.ID)
}

// List retrieves alerts with filters
func (r *AlertRepository) List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, rule_id, device_id, message, severity, status, triggered_at, resolved_at
		FROM alerts %s ORDER BY triggered_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		var resolvedAt sql.NullTime
		if err := rows.Scan(
			&a.ID, &a.RuleID, &a.DeviceID, &a.Message, &a.Severity,
			&a.Status, &a.TriggeredAt, &resolvedAt,
		); err != nil {
			return nil, 0, err
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}

	return alerts, total, nil
}

// CountActive counts active alerts
func (r *AlertRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM alerts WHERE status = 'active'",
	).Scan(&count)
	return count, err
}

// Resolve resolves an alert
func (r *AlertRepository) Resolve(ctx context.Context, id int) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		"UPDATE alerts SET status = 'resolved', resolved_at = $1 WHERE id = $2",
		now, id,
	)
	return err
}

// UpdateStatus updates alert status (for acknowledge, etc.)
func (r *AlertRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE alerts SET status = $1 WHERE id = $2",
		status, id,
	)
	return err
}

// GetRecentByDevice checks if a device has recent alerts (for cooldown)
func (r *AlertRepository) GetRecentByDevice(ctx context.Context, deviceID string, ruleID int, cooldownSec int) (*model.Alert, error) {
	query := `
		SELECT id, rule_id, device_id, message, severity, status, triggered_at, resolved_at
		FROM alerts
		WHERE device_id = $1 AND rule_id = $2 AND triggered_at > NOW() - INTERVAL '1 second' * $3
		ORDER BY triggered_at DESC LIMIT 1
	`
	var a model.Alert
	var resolvedAt sql.NullTime
	err := r.db.QueryRow(ctx, query, deviceID, ruleID, cooldownSec).Scan(
		&a.ID, &a.RuleID, &a.DeviceID, &a.Message, &a.Severity,
		&a.Status, &a.TriggeredAt, &resolvedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return &a, nil
}
// GetRecentAlertsByDeviceBatch 批量查询设备最近的告警（用于 cooldown 检查，避免 N+1 查询）
// FIX-P1-01: N+1 查询优化
func (r *AlertRepository) GetRecentAlertsByDeviceBatch(ctx context.Context, deviceID string, ruleIDs []int, cooldownSec int) (map[int]*model.Alert, error) {
	if len(ruleIDs) == 0 {
		return nil, nil
	}

	// 使用单次查询获取所有规则的最近告警
	query := `
		SELECT DISTINCT ON (rule_id) id, rule_id, device_id, message, severity, status, triggered_at, resolved_at
		FROM alerts
		WHERE device_id = $1 AND rule_id = ANY($2) AND triggered_at > NOW() - INTERVAL '1 second' * $3
		ORDER BY rule_id, triggered_at DESC
	`

	rows, err := r.db.Query(ctx, query, deviceID, ruleIDs, cooldownSec)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]*model.Alert)
	for rows.Next() {
		var a model.Alert
		var resolvedAt sql.NullTime
		if err := rows.Scan(
			&a.ID, &a.RuleID, &a.DeviceID, &a.Message, &a.Severity,
			&a.Status, &a.TriggeredAt, &resolvedAt,
		); err != nil {
			return nil, err
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		result[a.RuleID] = &a
	}

	return result, nil
}
