package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// TelemetryRepositoryInterface defines the interface for telemetry repository
type TelemetryRepositoryInterface interface {
	Insert(ctx context.Context, data *model.TelemetryData) error
	GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error)
	GetLatest(ctx context.Context) ([]model.TelemetryData, error)
	GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error)
	// GetStatsBatch 批量获取多个设备的统计数据，优化 N+1 查询问题
	// Performance optimization: replaces N individual GetStats calls with a single batch query
	GetStatsBatch(ctx context.Context, deviceIDs []string, start, end time.Time) (map[string]*model.DeviceStats, error)
}

// TelemetryRepository handles telemetry data access
type TelemetryRepository struct {
	db database.QueryExecutor
}

// NewTelemetryRepository creates a new telemetry repository
func NewTelemetryRepository(db database.QueryExecutor) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *TelemetryRepository) WithTx(tx database.TransactionInterface) *TelemetryRepository {
	return &TelemetryRepository{db: tx}
}

// Insert inserts telemetry data
func (r *TelemetryRepository) Insert(ctx context.Context, data *model.TelemetryData) error {
	query := `
		INSERT INTO device_telemetry (device_id, time, temperature, pressure, vibration, humidity, power, status, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		data.DeviceID, data.Timestamp, data.Temperature, data.Pressure,
		data.Vibration, data.Humidity, data.Power, data.Status, data.Message,
	).Scan(&data.ID)
}

// GetByDeviceID retrieves telemetry for a device within time range
func (r *TelemetryRepository) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	query := `
		SELECT id, device_id, time, temperature, pressure, vibration, humidity, power, status, message
		FROM device_telemetry
		WHERE device_id = $1 AND time >= $2 AND time <= $3
		ORDER BY time DESC
		LIMIT $4
	`
	rows, err := r.db.Query(ctx, query, deviceID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice, not nil
	data := make([]model.TelemetryData, 0)
	for rows.Next() {
		var d model.TelemetryData
		var temp, pressure, vibration, humidity, power sql.NullFloat64
		var message sql.NullString
		if err := rows.Scan(
			&d.ID, &d.DeviceID, &d.Timestamp, &temp, &pressure,
			&vibration, &humidity, &power, &d.Status, &message,
		); err != nil {
			return nil, err
		}
		d.Temperature = temp.Float64
		d.Pressure = pressure.Float64
		d.Vibration = vibration.Float64
		d.Humidity = humidity.Float64
		d.Power = power.Float64
		d.Message = message.String
		data = append(data, d)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return data, nil
}

// GetLatest retrieves latest telemetry for all devices
func (r *TelemetryRepository) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	query := `
		SELECT DISTINCT ON (device_id) id, device_id, time, temperature, pressure, vibration, humidity, power, status, message
		FROM device_telemetry
		ORDER BY device_id, time DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice, not nil
	data := make([]model.TelemetryData, 0)
	for rows.Next() {
		var d model.TelemetryData
		var temp, pressure, vibration, humidity, power sql.NullFloat64
		var message sql.NullString
		if err := rows.Scan(
			&d.ID, &d.DeviceID, &d.Timestamp, &temp, &pressure,
			&vibration, &humidity, &power, &d.Status, &message,
		); err != nil {
			return nil, err
		}
		d.Temperature = temp.Float64
		d.Pressure = pressure.Float64
		d.Vibration = vibration.Float64
		d.Humidity = humidity.Float64
		d.Power = power.Float64
		d.Message = message.String
		data = append(data, d)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return data, nil
}

// GetStats retrieves aggregated statistics for a device
func (r *TelemetryRepository) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	query := `
		SELECT 
			COALESCE(AVG(temperature), 0) as avg_temp,
			COALESCE(AVG(pressure), 0) as avg_pressure,
			COALESCE(AVG(vibration), 0) as avg_vibration,
			COALESCE(MAX(temperature), 0) as max_temp,
			COALESCE(MAX(pressure), 0) as max_pressure,
			COALESCE(MAX(vibration), 0) as max_vibration,
			COUNT(*) as count
		FROM device_telemetry
		WHERE device_id = $1 AND time >= $2 AND time <= $3
	`
	stats := &model.DeviceStats{DeviceID: deviceID}
	err := r.db.QueryRow(ctx, query, deviceID, start, end).Scan(
		&stats.AvgTemperature, &stats.AvgPressure, &stats.AvgVibration,
		&stats.MaxTemperature, &stats.MaxPressure, &stats.MaxVibration,
		&stats.DataPoints,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetStatsBatch 批量获取多个设备的统计数据
// Performance optimization: batch query replaces N individual GetStats calls with a single query
// to solve the N+1 query problem in export_service.go
func (r *TelemetryRepository) GetStatsBatch(ctx context.Context, deviceIDs []string, start, end time.Time) (map[string]*model.DeviceStats, error) {
	// 返回空 map 而不是 nil，避免调用方需要检查 nil
	result := make(map[string]*model.DeviceStats)
	if len(deviceIDs) == 0 {
		return result, nil
	}

	// 使用 GROUP BY 一次性查询所有设备的统计数据
	// Uses GROUP BY to fetch all device stats in a single query instead of N queries
	query := `
		SELECT 
			device_id,
			COALESCE(AVG(temperature), 0) as avg_temp,
			COALESCE(AVG(pressure), 0) as avg_pressure,
			COALESCE(AVG(vibration), 0) as avg_vibration,
			COALESCE(MAX(temperature), 0) as max_temp,
			COALESCE(MAX(pressure), 0) as max_pressure,
			COALESCE(MAX(vibration), 0) as max_vibration,
			COUNT(*) as count
		FROM device_telemetry
		WHERE device_id = ANY($1) AND time >= $2 AND time <= $3
		GROUP BY device_id
	`

	rows, err := r.db.Query(ctx, query, deviceIDs, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		stats := &model.DeviceStats{}
		if err := rows.Scan(
			&stats.DeviceID,
			&stats.AvgTemperature, &stats.AvgPressure, &stats.AvgVibration,
			&stats.MaxTemperature, &stats.MaxPressure, &stats.MaxVibration,
			&stats.DataPoints,
		); err != nil {
			return nil, err
		}
		result[stats.DeviceID] = stats
	}

	return result, rows.Err()
}

// WorkOrderRepositoryInterface defines the interface for work order repository
type WorkOrderRepositoryInterface interface {
	Create(ctx context.Context, wo *model.WorkOrder) error
	GetByID(ctx context.Context, id int) (*model.WorkOrder, error)
	List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	CountOpen(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status string) (int, error)
}

// WorkOrderRepository handles work order data access
type WorkOrderRepository struct {
	db database.QueryExecutor
}

// NewWorkOrderRepository creates a new work order repository
func NewWorkOrderRepository(db database.QueryExecutor) *WorkOrderRepository {
	return &WorkOrderRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *WorkOrderRepository) WithTx(tx database.TransactionInterface) *WorkOrderRepository {
	return &WorkOrderRepository{db: tx}
}

// Create creates a new work order
func (r *WorkOrderRepository) Create(ctx context.Context, wo *model.WorkOrder) error {
	query := `
		INSERT INTO work_orders (title, description, device_id, priority, status, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		wo.Title, wo.Description, wo.DeviceID, wo.Priority, wo.Status,
		wo.AssignedTo, wo.CreatedAt, wo.UpdatedAt,
	).Scan(&wo.ID)
}

// GetByID retrieves a work order by ID
func (r *WorkOrderRepository) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
	query := `
		SELECT id, title, description, device_id, priority, status, assigned_to, created_at, updated_at
		FROM work_orders WHERE id = $1
	`
	wo := &model.WorkOrder{}
	var assignedTo sql.NullInt64
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wo.ID, &wo.Title, &wo.Description, &wo.DeviceID, &wo.Priority,
		&wo.Status, &assignedTo, &wo.CreatedAt, &wo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assignedTo.Valid {
		id := int(assignedTo.Int64)
		wo.AssignedTo = &id
	}
	return wo, nil
}

// List retrieves work orders with filters
func (r *WorkOrderRepository) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if deviceID != "" {
		whereClause += fmt.Sprintf(" AND device_id = $%d", argIdx)
		args = append(args, deviceID)
		argIdx++
	}

	// Count total
	// Security: SQL动态拼接安全性说明 - whereClause由参数化查询条件构建，所有变量值通过$N占位符传递，
	// 非用户直接输入。字段名和操作符均为硬编码常量，不存在SQL注入风险。
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM work_orders %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, title, description, device_id, priority, status, assigned_to, created_at, updated_at
		FROM work_orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.WorkOrder
	for rows.Next() {
		var wo model.WorkOrder
		var assignedTo sql.NullInt64
		if err := rows.Scan(
			&wo.ID, &wo.Title, &wo.Description, &wo.DeviceID, &wo.Priority,
			&wo.Status, &assignedTo, &wo.CreatedAt, &wo.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if assignedTo.Valid {
			id := int(assignedTo.Int64)
			wo.AssignedTo = &id
		}
		orders = append(orders, wo)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return orders, total, nil
}

// UpdateStatus updates work order status
func (r *WorkOrderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE work_orders SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id,
	)
	return err
}

// CountOpen counts open work orders (status: open, in_progress, pending)
func (r *WorkOrderRepository) CountOpen(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM work_orders WHERE status IN ('open', 'in_progress', 'pending')",
	).Scan(&count)
	return count, err
}

// CountByStatus counts work orders by specific status
func (r *WorkOrderRepository) CountByStatus(ctx context.Context, status string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM work_orders WHERE status = $1",
		status,
	).Scan(&count)
	return count, err
}

// NotificationRepositoryInterface defines the interface for notification repository
type NotificationRepositoryInterface interface {
	Create(ctx context.Context, n *model.Notification) error
	List(ctx context.Context, notifType string, unreadOnly bool, page, pageSize int) ([]model.Notification, int, error)
	MarkRead(ctx context.Context, id int) error
}

// NotificationRepository handles notification data access
type NotificationRepository struct {
	db database.QueryExecutor
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db database.QueryExecutor) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *NotificationRepository) WithTx(tx database.TransactionInterface) *NotificationRepository {
	return &NotificationRepository{db: tx}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, n *model.Notification) error {
	query := `
		INSERT INTO notifications (type, title, message, device_id, read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var deviceID interface{}
	if n.DeviceID != nil {
		deviceID = *n.DeviceID
	}
	return r.db.QueryRow(ctx, query,
		n.Type, n.Title, n.Message, deviceID, n.Read, n.CreatedAt,
	).Scan(&n.ID)
}

// List retrieves notifications with filters
func (r *NotificationRepository) List(ctx context.Context, notifType string, unreadOnly bool, page, pageSize int) ([]model.Notification, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if notifType != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, notifType)
		argIdx++
	}
	if unreadOnly {
		whereClause += " AND read = false"
	}

	// Security: SQL动态拼接安全性说明 - whereClause由参数化查询条件构建，所有变量值通过$N占位符传递，
	// 非用户直接输入。字段名和操作符均为硬编码常量，不存在SQL注入风险。
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, type, title, message, device_id, read, created_at
		FROM notifications %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		var deviceID sql.NullString
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Title, &n.Message, &deviceID, &n.Read, &n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if deviceID.Valid {
			n.DeviceID = &deviceID.String
		}
		notifications = append(notifications, n)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return notifications, total, nil
}

// MarkRead marks a notification as read
func (r *NotificationRepository) MarkRead(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "UPDATE notifications SET read = true WHERE id = $1", id)
	return err
}

// BlackBoxRepositoryInterface defines the interface for black box repository
type BlackBoxRepositoryInterface interface {
	Create(ctx context.Context, record *model.BlackBoxRecord) error
	List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error)
}

// BlackBoxRepository handles black box record data access
type BlackBoxRepository struct {
	db database.QueryExecutor
}

// NewBlackBoxRepository creates a new black box repository
func NewBlackBoxRepository(db database.QueryExecutor) *BlackBoxRepository {
	return &BlackBoxRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *BlackBoxRepository) WithTx(tx database.TransactionInterface) *BlackBoxRepository {
	return &BlackBoxRepository{db: tx}
}

// Create creates a new black box record
func (r *BlackBoxRepository) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	query := `
		INSERT INTO blackbox_records (device_id, trigger_type, start_time, end_time, summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		record.DeviceID, record.TriggerType, record.StartTime, record.EndTime,
		record.Summary, record.CreatedAt,
	).Scan(&record.ID)
}

// List retrieves black box records
func (r *BlackBoxRepository) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if deviceID != "" {
		whereClause = fmt.Sprintf("WHERE device_id = $%d", argIdx)
		args = append(args, deviceID)
		argIdx++
	}

	// Security: SQL动态拼接安全性说明 - whereClause由参数化查询条件构建，所有变量值通过$N占位符传递，
	// 非用户直接输入。字段名和操作符均为硬编码常量，不存在SQL注入风险。
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM blackbox_records %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, device_id, trigger_type, start_time, end_time, summary, created_at
		FROM blackbox_records %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []model.BlackBoxRecord
	for rows.Next() {
		var r model.BlackBoxRecord
		var summary sql.NullString
		if err := rows.Scan(
			&r.ID, &r.DeviceID, &r.TriggerType, &r.StartTime,
			&r.EndTime, &summary, &r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		r.Summary = summary.String
		records = append(records, r)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return records, total, nil
}

// ReportRepositoryInterface defines the interface for report repository
type ReportRepositoryInterface interface {
	Create(ctx context.Context, report *model.Report) error
	List(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error)
	GetByID(ctx context.Context, id int) (*model.Report, error)
	Delete(ctx context.Context, id int) error
}

// ReportRepository handles report data access
type ReportRepository struct {
	db database.QueryExecutor
}

// NewReportRepository creates a new report repository
func NewReportRepository(db database.QueryExecutor) *ReportRepository {
	return &ReportRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *ReportRepository) WithTx(tx database.TransactionInterface) *ReportRepository {
	return &ReportRepository{db: tx}
}

// Create creates a new report
func (r *ReportRepository) Create(ctx context.Context, report *model.Report) error {
	query := `
		INSERT INTO reports (title, type, device_id, content, generated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var deviceID interface{}
	if report.DeviceID != nil {
		deviceID = *report.DeviceID
	}
	return r.db.QueryRow(ctx, query,
		report.Title, report.Type, deviceID, report.Content, report.GeneratedAt,
	).Scan(&report.ID)
}

// List retrieves reports
func (r *ReportRepository) List(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error) {
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if reportType != "" {
		whereClause = fmt.Sprintf("WHERE type = $%d", argIdx)
		args = append(args, reportType)
		argIdx++
	}

	// Security: SQL动态拼接安全性说明 - whereClause由参数化查询条件构建，所有变量值通过$N占位符传递，
	// 非用户直接输入。字段名和操作符均为硬编码常量，不存在SQL注入风险。
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reports %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, title, type, device_id, content, generated_at
		FROM reports %s ORDER BY generated_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []model.Report
	for rows.Next() {
		var r model.Report
		var deviceID sql.NullString
		if err := rows.Scan(
			&r.ID, &r.Title, &r.Type, &deviceID, &r.Content, &r.GeneratedAt,
		); err != nil {
			return nil, 0, err
		}
		if deviceID.Valid {
			r.DeviceID = &deviceID.String
		}
		reports = append(reports, r)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return reports, total, nil
}

// GetByID retrieves a report by ID
func (r *ReportRepository) GetByID(ctx context.Context, id int) (*model.Report, error) {
	query := `SELECT id, title, type, device_id, content, generated_at FROM reports WHERE id = $1`
	report := &model.Report{}
	var deviceID sql.NullString
	err := r.db.QueryRow(ctx, query, id).Scan(
		&report.ID, &report.Title, &report.Type, &deviceID, &report.Content, &report.GeneratedAt,
	)
	if err != nil {
		return nil, err
	}
	if deviceID.Valid {
		report.DeviceID = &deviceID.String
	}
	return report, nil
}

// Delete deletes a report by ID
func (r *ReportRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM reports WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// AgentTaskLogRepositoryInterface defines the interface for agent task log repository
type AgentTaskLogRepositoryInterface interface {
	Create(ctx context.Context, log *model.AgentTaskLog) error
	List(ctx context.Context, limit int) ([]model.AgentTaskLog, error)
	WithTx(tx database.TransactionInterface) AgentTaskLogRepositoryInterface
}

// AgentTaskLogRepository handles AI agent task log data access
type AgentTaskLogRepository struct {
	db database.QueryExecutor
}

// NewAgentTaskLogRepository creates a new agent task log repository
func NewAgentTaskLogRepository(db database.QueryExecutor) *AgentTaskLogRepository {
	return &AgentTaskLogRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *AgentTaskLogRepository) WithTx(tx database.TransactionInterface) AgentTaskLogRepositoryInterface {
	return &AgentTaskLogRepository{db: tx}
}

// Create creates a new task log
func (r *AgentTaskLogRepository) Create(ctx context.Context, log *model.AgentTaskLog) error {
	query := `
		INSERT INTO agent_task_logs (session_id, query, response, agent, executed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		log.SessionID, log.Query, log.Response, log.Agent, log.ExecutedAt,
	).Scan(&log.ID)
}

// List retrieves task logs
func (r *AgentTaskLogRepository) List(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	query := `
		SELECT id, session_id, query, response, agent, executed_at
		FROM agent_task_logs
		ORDER BY executed_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice, not nil
	logs := make([]model.AgentTaskLog, 0)
	for rows.Next() {
		var l model.AgentTaskLog
		if err := rows.Scan(
			&l.ID, &l.SessionID, &l.Query, &l.Response, &l.Agent, &l.ExecutedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return logs, nil
}

// Helper to unmarshal JSON
// FIX-P0-04: 添加json.Unmarshal错误处理
func unmarshalActions(actionsJSON string) []map[string]interface{} {
	var actions []map[string]interface{}
	if actionsJSON != "" {
		if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
			// 解析失败时返回nil，符合测试预期
			logger.L().Warn("failed to unmarshal actions", zap.Error(err), zap.String("actionsJSON", actionsJSON))
			return nil
		}
	}
	return actions
}
