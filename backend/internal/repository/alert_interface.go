package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
)

// AlertFilter 告警过滤条件
// P0-03: 将过滤条件传递到数据库层，避免内存过滤
type AlertFilter struct {
	Status   string
	Severity string
	DeviceID string
}

// AlertRepositoryInterface defines the interface for alert repository
type AlertRepositoryInterface interface {
	Create(ctx context.Context, alert *model.Alert) error
	List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error)
	// ListWithFilter 支持更多过滤条件的列表查询
	ListWithFilter(ctx context.Context, filter AlertFilter, page, pageSize int) ([]model.Alert, int, error)
	CountActive(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status string) (int, error)
	Resolve(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status string) error
	GetRecentByDevice(ctx context.Context, deviceID string, ruleID int, cooldownSec int) (*model.Alert, error)
	// FIX-P1-01: N+1 查询优化 - 新增批量查询方法
	GetRecentAlertsByDeviceBatch(ctx context.Context, deviceID string, ruleIDs []int, cooldownSec int) (map[int]*model.Alert, error)

	// P2-002: 告警历史管理 - 新增归档方法
	// ArchiveOldAlerts 归档超过指定天数的已解决告警
	ArchiveOldAlerts(ctx context.Context, daysOld int) (int, error)
	// GetArchivedAlerts 查询已归档告警
	GetArchivedAlerts(ctx context.Context, deviceID string, page, pageSize int) ([]model.Alert, int, error)
	// DeleteArchivedAlerts 删除超过指定天数的归档告警
	DeleteArchivedAlerts(ctx context.Context, daysOld int) (int, error)
	// GetAlertStatistics 获取告警统计信息
	GetAlertStatistics(ctx context.Context) (*AlertStatistics, error)
	// GetTrendData 按日期分组获取告警趋势数据
	GetTrendData(ctx context.Context, days int) ([]map[string]interface{}, error)
	// GetDeviceRankingData 按设备分组获取告警数量排名
	GetDeviceRankingData(ctx context.Context, limit int) ([]map[string]interface{}, error)
	// GetEfficiencyData 获取告警处理效率数据（平均解决时间和确认率）
	GetEfficiencyData(ctx context.Context) (*AlertEfficiencyData, error)
}

// AlertStatistics 告警统计信息
type AlertStatistics struct {
	TotalActive    int `json:"total_active"`
	TotalResolved  int `json:"total_resolved"`
	TotalArchived  int `json:"total_archived"`
	TodayTriggered int `json:"today_triggered"`
	TodayResolved  int `json:"today_resolved"`
	WeekTriggered  int `json:"week_triggered"`
	WeekResolved   int `json:"week_resolved"`
	AvgResolveTime int `json:"avg_resolve_time_seconds"` // 平均解决时间（秒）
	CriticalCount  int `json:"critical_count"`
	WarningCount   int `json:"warning_count"`
}

// AlertEfficiencyData 告警处理效率数据
type AlertEfficiencyData struct {
	AvgResolveTime float64 `json:"avg_resolve_time"` // 平均解决时间（秒）
	TotalAlerts    int     `json:"total_alerts"`     // 告警总数
	ResolvedAlerts int     `json:"resolved_alerts"`  // 已确认/解决的告警数
	AckRate        float64 `json:"ack_rate"`          // 确认率（0-1）
}
