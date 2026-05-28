package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ============================================
// PostgreSQL Audit Log Repository
// ============================================

// PostgresRepository represents PostgreSQL audit log repository
type PostgresRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a PostgreSQL audit log repository
func NewPostgresRepository(db *sqlx.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// ============================================
// Repository Interface Implementation
// ============================================

// Create creates an audit log
func (r *PostgresRepository) Create(ctx context.Context, log *AuditLog) error {
	// Serialize JSON fields
	beforeState, err := json.Marshal(log.BeforeState)
	if err != nil {
		return fmt.Errorf("marshal before state: %w", err)
	}
	afterState, err := json.Marshal(log.AfterState)
	if err != nil {
		return fmt.Errorf("marshal after state: %w", err)
	}
	changes, err := json.Marshal(log.Changes)
	if err != nil {
		return fmt.Errorf("marshal changes: %w", err)
	}
	metadata, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			audit_id, timestamp, event_type, event_category, severity,
			user_id, tenant_id, session_id, ip_address, user_agent,
			resource_type, resource_id, action, operation, request_id, trace_id,
			before_state, after_state, changes, result, error_message,
			duration_ms, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		log.AuditID, log.Timestamp, log.EventType, log.EventCategory, log.Severity,
		log.UserID, log.TenantID, log.SessionID, log.IPAddress, log.UserAgent,
		log.ResourceType, log.ResourceID, log.Action, log.Operation, log.RequestID, log.TraceID,
		beforeState, afterState, changes, log.Result, log.ErrorMessage,
		log.DurationMs, metadata, log.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to insert audit log",
			zap.String("audit_id", log.AuditID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Query queries audit logs
func (r *PostgresRepository) Query(ctx context.Context, query *QueryRequest) ([]*AuditLog, int64, error) {
	// Build query conditions
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if query.TenantID != "" {
		whereClause += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, query.TenantID)
		argIndex++
	}

	if query.UserID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, query.UserID)
		argIndex++
	}

	if query.EventType != "" {
		whereClause += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, query.EventType)
		argIndex++
	}

	if query.Category != "" {
		whereClause += fmt.Sprintf(" AND event_category = $%d", argIndex)
		args = append(args, query.Category)
		argIndex++
	}

	if query.ResourceType != "" {
		whereClause += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, query.ResourceType)
		argIndex++
	}

	if query.ResourceID != "" {
		whereClause += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, query.ResourceID)
		argIndex++
	}

	if query.Result != "" {
		whereClause += fmt.Sprintf(" AND result = $%d", argIndex)
		args = append(args, query.Result)
		argIndex++
	}

	if query.IPAddress != "" {
		whereClause += fmt.Sprintf(" AND ip_address = $%d", argIndex)
		args = append(args, query.IPAddress)
		argIndex++
	}

	if query.StartTime != nil {
		whereClause += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, query.StartTime)
		argIndex++
	}

	if query.EndTime != nil {
		whereClause += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, query.EndTime)
		argIndex++
	}

	// Query total count
	countQuery := "SELECT COUNT(*) FROM audit_logs " + whereClause
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Set default pagination parameters
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// Calculate offset
	offset := (query.Page - 1) * query.PageSize

	// Query data
	dataQuery := `
		SELECT 
			audit_id, timestamp, event_type, event_category, severity,
			user_id, tenant_id, session_id, ip_address, user_agent,
			resource_type, resource_id, action, operation, request_id, trace_id,
			before_state, after_state, changes, result, error_message,
			duration_ms, metadata, created_at
		FROM audit_logs
		` + whereClause + `
		ORDER BY timestamp DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, query.PageSize, offset)

	var logs []*AuditLog
	err = r.db.SelectContext(ctx, &logs, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Parse JSON fields
	for _, log := range logs {
		if log.BeforeState != nil && len(log.BeforeState) > 0 {
			// BeforeState already is map, no conversion needed from DB
		}
		if log.AfterState != nil && len(log.AfterState) > 0 {
			// AfterState already is map, no conversion needed from DB
		}
		if log.Changes != nil && len(log.Changes) > 0 {
			// Changes already is map, no conversion needed from DB
		}
		if log.Metadata != nil && len(log.Metadata) > 0 {
			// Metadata already is map, no conversion needed from DB
		}
	}

	return logs, total, nil
}

// GetByID retrieves audit log details by ID
func (r *PostgresRepository) GetByID(ctx context.Context, auditID string) (*AuditLog, error) {
	query := `
		SELECT 
			audit_id, timestamp, event_type, event_category, severity,
			user_id, tenant_id, session_id, ip_address, user_agent,
			resource_type, resource_id, action, operation, request_id, trace_id,
			before_state, after_state, changes, result, error_message,
			duration_ms, metadata, created_at
		FROM audit_logs
		WHERE audit_id = $1
	`

	var log AuditLog
	err := r.db.GetContext(ctx, &log, query, auditID)
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if log.BeforeState != nil && len(log.BeforeState) > 0 {
		// BeforeState already is map, no conversion needed from DB
	}
	if log.AfterState != nil && len(log.AfterState) > 0 {
		// AfterState already is map, no conversion needed from DB
	}
	if log.Changes != nil && len(log.Changes) > 0 {
		// Changes already is map, no conversion needed from DB
	}
	if log.Metadata != nil && len(log.Metadata) > 0 {
		// Metadata already is map, no conversion needed from DB
	}

	return &log, nil
}

// DeleteOld deletes old audit logs based on retention days
func (r *PostgresRepository) DeleteOld(ctx context.Context, retentionDays int) error {
	query := `
		DELETE FROM audit_logs
		WHERE timestamp < NOW() - INTERVAL '1 day' * $1
	`

	result, err := r.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		r.logger.Warn("Failed to get rows affected", zap.Error(err))
	}
	r.logger.Info("Deleted old audit logs",
		zap.Int64("deleted_count", deleted),
		zap.Int("retention_days", retentionDays),
	)

	return nil
}

// ============================================
// Audit Log Statistics
// ============================================

// Statistics represents audit log statistics
type Statistics struct {
	TotalLogs       int64            `json:"total_logs"`
	EventTypes      map[string]int64 `json:"event_types"`
	Categories      map[string]int64 `json:"categories"`
	TopUsers        []UserStats      `json:"top_users"`
	TopResources    []ResourceStats  `json:"top_resources"`
	FailureRate     float64          `json:"failure_rate"`
	AverageDuration float64          `json:"average_duration"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID string `json:"user_id"`
	Count  int64  `json:"count"`
}

// ResourceStats represents resource statistics
type ResourceStats struct {
	ResourceType string `json:"resource_type"`
	Count        int64  `json:"count"`
}

// GetStatistics retrieves audit log statistics
func (r *PostgresRepository) GetStatistics(ctx context.Context, startTime, endTime time.Time) (*Statistics, error) {
	stats := &Statistics{
		EventTypes:   make(map[string]int64),
		Categories:   make(map[string]int64),
		TopUsers:     []UserStats{},
		TopResources: []ResourceStats{},
	}

	// Total count
	totalQuery := `
		SELECT COUNT(*) FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	err := r.db.GetContext(ctx, &stats.TotalLogs, totalQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Event type statistics
	eventTypeQuery := `
		SELECT event_type, COUNT(*) as count
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY event_type
		ORDER BY count DESC
	`
	eventTypeRows, err := r.db.QueryContext(ctx, eventTypeQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer eventTypeRows.Close()

	for eventTypeRows.Next() {
		var eventType string
		var count int64
		if err := eventTypeRows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("scan event type row: %w", err)
		}
		stats.EventTypes[eventType] = count
	}
	// Check for errors during rows iteration
	if err = eventTypeRows.Err(); err != nil {
		return nil, err
	}

	// Category statistics
	categoryQuery := `
		SELECT event_category, COUNT(*) as count
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY event_category
		ORDER BY count DESC
	`
	categoryRows, err := r.db.QueryContext(ctx, categoryQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var category string
		var count int64
		if err := categoryRows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		stats.Categories[category] = count
	}
	// Check for errors during rows iteration
	if err = categoryRows.Err(); err != nil {
		return nil, err
	}

	// Top users
	topUsersQuery := `
		SELECT user_id, COUNT(*) as count
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY user_id
		ORDER BY count DESC
		LIMIT 10
	`
	err = r.db.SelectContext(ctx, &stats.TopUsers, topUsersQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Top resources
	topResourcesQuery := `
		SELECT resource_type, COUNT(*) as count
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY resource_type
		ORDER BY count DESC
		LIMIT 10
	`
	err = r.db.SelectContext(ctx, &stats.TopResources, topResourcesQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Failure rate
	failureQuery := `
		SELECT 
			COUNT(CASE WHEN result = 'failure' THEN 1 END) * 100.0 / COUNT(*)
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	err = r.db.GetContext(ctx, &stats.FailureRate, failureQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Average duration
	avgDurationQuery := `
		SELECT AVG(duration_ms)
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	err = r.db.GetContext(ctx, &stats.AverageDuration, avgDurationQuery, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
