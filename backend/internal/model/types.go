package model

import (
	"time"
)

// User represents a system user
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Password     string    `json:"-" db:"password_hash"`
	PasswordHash string    `json:"-" db:"password_hash"` // 别名，用于兼容
	Email        string    `json:"email" db:"email"`
	Role         string    `json:"role" db:"role"`
	Roles        []string  `json:"roles" db:"-"` // 用户角色列表（RBAC）
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	TokenVersion int       `json:"-" db:"token_version"` // Token 版本号，用于撤销所有 Token
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Device represents an industrial device
type Device struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`
	Location    string    `json:"location" db:"location"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description" db:"description"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TelemetryData represents sensor data from a device
type TelemetryData struct {
	ID          int64     `json:"id,omitempty" db:"id"`
	DeviceID    string    `json:"device_id" db:"device_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Timestamp   time.Time `json:"timestamp" db:"time"`
	Temperature float64   `json:"temperature,omitempty" db:"temperature"`
	Pressure    float64   `json:"pressure,omitempty" db:"pressure"`
	Vibration   float64   `json:"vibration,omitempty" db:"vibration"`
	Humidity    float64   `json:"humidity,omitempty" db:"humidity"`
	Power       float64   `json:"power,omitempty" db:"power"`
	Status      string    `json:"status" db:"status"`
	Message     string    `json:"message,omitempty" db:"message"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DeviceType  string    `json:"device_type" db:"device_type"`
	Metric      string    `json:"metric" db:"metric"`
	Operator    string    `json:"operator" db:"operator"`
	Threshold   float64   `json:"threshold" db:"threshold"`
	Severity    string    `json:"severity" db:"severity"`
	Actions     string    `json:"actions" db:"actions"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	CooldownSec int       `json:"cooldown_sec" db:"cooldown_sec"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          int        `json:"id" db:"id"`
	RuleID      int        `json:"rule_id" db:"rule_id"`
	DeviceID    string     `json:"device_id" db:"device_id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Message     string     `json:"message" db:"message"`
	Severity    string     `json:"severity" db:"severity"`
	Status      string     `json:"status" db:"status"`
	TriggeredAt time.Time  `json:"triggered_at" db:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// WorkOrder represents a maintenance work order
type WorkOrder struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	DeviceID    string    `json:"device_id" db:"device_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Priority    string    `json:"priority" db:"priority"`
	Status      string    `json:"status" db:"status"`
	AssignedTo  *int      `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Notification represents a system notification
type Notification struct {
	ID        int       `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	DeviceID  *string   `json:"device_id,omitempty" db:"device_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Read      bool      `json:"read" db:"read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BlackBox Record captures device state during anomalies
type BlackBoxRecord struct {
	ID          int64           `json:"id" db:"id"`
	DeviceID    string          `json:"device_id" db:"device_id"`
	TenantID    string          `json:"tenant_id" db:"tenant_id"`
	TriggerType string          `json:"trigger_type" db:"trigger_type"`
	StartTime   time.Time       `json:"start_time" db:"start_time"`
	EndTime     time.Time       `json:"end_time" db:"end_time"`
	Snapshot    []TelemetryData `json:"snapshot" db:"-"`
	Summary     string          `json:"summary" db:"summary"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Report represents an AI-generated report
type Report struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Type        string    `json:"type" db:"type"`
	DeviceID    *string   `json:"device_id,omitempty" db:"device_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Content     string    `json:"content" db:"content"`
	GeneratedAt time.Time `json:"generated_at" db:"generated_at"`
}

// AI Agent types
type AgentQuery struct {
	Query     string                 `json:"query"`
	Context   map[string]interface{} `json:"context,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
}

type AgentResponse struct {
	SessionID string                   `json:"session_id"`
	Response  string                   `json:"response"`
	Agent     string                   `json:"agent"`
	Actions   []map[string]interface{} `json:"actions,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
}

type AgentTaskLog struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	Query      string    `json:"query" db:"query"`
	Response   string    `json:"response" db:"response"`
	Agent      string    `json:"agent" db:"agent"`
	ExecutedAt time.Time `json:"executed_at" db:"executed_at"`
}

// Statistics types
type DeviceStats struct {
	DeviceID       string  `json:"device_id"`
	AvgTemperature float64 `json:"avg_temperature"`
	AvgPressure    float64 `json:"avg_pressure"`
	AvgVibration   float64 `json:"avg_vibration"`
	MaxTemperature float64 `json:"max_temperature"`
	MaxPressure    float64 `json:"max_pressure"`
	MaxVibration   float64 `json:"max_vibration"`
	DataPoints     int64   `json:"data_points"`
}

type ROIStats struct {
	TotalDevices     int     `json:"total_devices"`
	ActiveAlerts     int     `json:"active_alerts"`
	OpenWorkOrders   int     `json:"open_work_orders"`
	ResolvedIssues   int     `json:"resolved_issues"`
	PredictedSavings float64 `json:"predicted_savings"`
	UptimePercentage float64 `json:"uptime_percentage"`
	AvgResponseTime  float64 `json:"avg_response_time_hours"`
}

// System status
type SystemStatus struct {
	Database    string    `json:"database"`
	DBLatency   int64     `json:"db_latency_ms"`
	Uptime      string    `json:"uptime"`
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	DeviceCount int       `json:"device_count"`
	UserCount   int       `json:"user_count"`
}

// WebSocket message types
type WSMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// Pagination
type PaginationParams struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	Order    string `form:"order" json:"order"`
}

func (p *PaginationParams) Defaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.Order == "" {
		p.Order = "desc"
	}
}
