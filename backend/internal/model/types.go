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
// BE-P2-03: 添加 Gin binding 验证标签
type Device struct {
	ID          string    `json:"id" db:"id" binding:"required,max=100"`
	Name        string    `json:"name" db:"name" binding:"required,min=1,max=200"`
	Type        string    `json:"type" db:"type" binding:"required,oneof=CNC InjectionMolder AssemblyRobot Conveyor Sensor PLC robot motor pump valve heater cooler"`
	Location    string    `json:"location" db:"location" binding:"max=200"`
	Status      string    `json:"status" db:"status" binding:"omitempty,oneof=online offline maintenance error warning fault"`
	Description string    `json:"description" db:"description" binding:"max=1000"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" binding:"max=100"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceCreateRequest 设备创建请求
type DeviceCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Type        string `json:"type" binding:"required,oneof=CNC InjectionMolder AssemblyRobot Conveyor Sensor PLC robot motor pump valve heater cooler"`
	Location    string `json:"location" binding:"omitempty,max=200"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// DeviceUpdateRequest 设备更新请求
type DeviceUpdateRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=200"`
	Type        string `json:"type" binding:"omitempty,oneof=CNC InjectionMolder AssemblyRobot Conveyor Sensor PLC robot motor pump valve heater cooler"`
	Location    string `json:"location" binding:"omitempty,max=200"`
	Status      string `json:"status" binding:"omitempty,oneof=online offline maintenance error warning fault"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// TelemetryData represents sensor data from a device
// BE-P2-03: 添加 Gin binding 验证标签
type TelemetryData struct {
	ID          int64     `json:"id,omitempty" db:"id"`
	DeviceID    string    `json:"device_id" db:"device_id" binding:"required,max=100"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" binding:"max=100"`
	Timestamp   time.Time `json:"timestamp" db:"time"`
	Temperature float64   `json:"temperature,omitempty" db:"temperature" binding:"omitempty"`
	Pressure    float64   `json:"pressure,omitempty" db:"pressure" binding:"omitempty"`
	Vibration   float64   `json:"vibration,omitempty" db:"vibration" binding:"omitempty"`
	Humidity    float64   `json:"humidity,omitempty" db:"humidity" binding:"omitempty"`
	Power       float64   `json:"power,omitempty" db:"power" binding:"omitempty"`
	Status      string    `json:"status" db:"status" binding:"omitempty,oneof=normal warning fault"`
	Message     string    `json:"message,omitempty" db:"message" binding:"max=500"`
}

// AlertRule defines conditions for triggering alerts
// BE-P2-03: 添加 Gin binding 验证标签
type AlertRule struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required,min=1,max=200"`
	DeviceType  string    `json:"device_type" db:"device_type" binding:"omitempty"`
	Metric      string    `json:"metric" db:"metric" binding:"required,oneof=temperature pressure vibration humidity power"`
	Operator    string    `json:"operator" db:"operator" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64   `json:"threshold" db:"threshold" binding:"required"`
	Severity    string    `json:"severity" db:"severity" binding:"required,oneof=low medium high critical"`
	Actions     string    `json:"actions" db:"actions"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	CooldownSec int       `json:"cooldown_sec" db:"cooldown_sec" binding:"omitempty,min=60,max=3600"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" binding:"max=100"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AlertRuleCreateRequest 告警规则创建请求
type AlertRuleCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	DeviceType  string  `json:"device_type" binding:"omitempty"`
	Metric      string  `json:"metric" binding:"required,oneof=temperature pressure vibration humidity power"`
	Operator    string  `json:"operator" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64 `json:"threshold" binding:"required"`
	Severity    string  `json:"severity" binding:"required,oneof=low medium high critical"`
	Actions     string  `json:"actions" binding:"omitempty"`
	Enabled     bool    `json:"enabled"`
	CooldownSec int     `json:"cooldown_sec" binding:"omitempty,min=60,max=3600"`
}

// AlertRuleUpdateRequest 告警规则更新请求
type AlertRuleUpdateRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=1,max=200"`
	DeviceType  string  `json:"device_type" binding:"omitempty"`
	Metric      string  `json:"metric" binding:"omitempty,oneof=temperature pressure vibration humidity power"`
	Operator    string  `json:"operator" binding:"omitempty,oneof=> >= < <= == !="`
	Threshold   float64 `json:"threshold" binding:"omitempty"`
	Severity    string  `json:"severity" binding:"omitempty,oneof=low medium high critical"`
	Actions     string  `json:"actions" binding:"omitempty"`
	Enabled     bool    `json:"enabled"`
	CooldownSec int     `json:"cooldown_sec" binding:"omitempty,min=60,max=3600"`
}

// Alert represents a triggered alert
// BE-P2-03: 添加 Gin binding 验证标签
type Alert struct {
	ID             int        `json:"id" db:"id"`
	RuleID         int        `json:"rule_id" db:"rule_id" binding:"required"`
	DeviceID       string     `json:"device_id" db:"device_id" binding:"required,max=100"`
	TenantID       string     `json:"tenant_id" db:"tenant_id" binding:"max=100"`
	Message        string     `json:"message" db:"message" binding:"max=500"`
	Severity       string     `json:"severity" db:"severity" binding:"oneof=low medium high critical"`
	Status         string     `json:"status" db:"status" binding:"oneof=active acknowledged resolved"`
	Value          *float64   `json:"value,omitempty" db:"value"`
	Threshold      *float64   `json:"threshold,omitempty" db:"threshold"`
	TriggeredAt    time.Time  `json:"triggered_at" db:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	AcknowledgedBy *string    `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
}

// WorkOrder represents a maintenance work order
// BE-P2-03: 添加 Gin binding 验证标签
type WorkOrder struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" db:"description" binding:"max=1000"`
	DeviceID    string    `json:"device_id" db:"device_id" binding:"required,max=100"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" binding:"max=100"`
	Priority    string    `json:"priority" db:"priority" binding:"oneof=low medium high urgent"`
	Status      string    `json:"status" db:"status" binding:"oneof=pending in_progress completed cancelled"`
	AssignedTo  *int      `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Notification represents a system notification
// BE-P2-03: 添加 Gin binding 验证标签
type Notification struct {
	ID        int       `json:"id" db:"id"`
	Type      string    `json:"type" db:"type" binding:"required,oneof=alert system info"`
	Title     string    `json:"title" db:"title" binding:"required,max=200"`
	Message   string    `json:"message" db:"message" binding:"required,max=500"`
	DeviceID  *string   `json:"device_id,omitempty" db:"device_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id" binding:"max=100"`
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

// DeviceGraph 设备关系图
type DeviceGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

// DeviceStatsDetail 设备详细统计
type DeviceStatsDetail struct {
	DeviceID      string  `json:"device_id"`
	UptimeDays    int     `json:"uptime_days"`
	FaultCount    int     `json:"fault_count"`
	AvgResponseMs float64 `json:"avg_response_ms"`
}

// TrendReport 告警趋势报告
type TrendReport struct {
	Period string       `json:"period"`
	Trend  []TrendEntry `json:"trend"`
}

type TrendEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DeviceRankingEntry 设备告警排行
type DeviceRankingEntry struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	AlertCount int    `json:"alert_count"`
}

// EfficiencyReport 效率报告
type EfficiencyReport struct {
	AvgResolveTime float64 `json:"avg_resolve_time"`
	AckRate        float64 `json:"ack_rate"`
	TotalAlerts    int     `json:"total_alerts"`
	ResolvedAlerts int     `json:"resolved_alerts"`
}
