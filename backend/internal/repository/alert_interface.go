package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
)

// AlertRepositoryInterface defines the interface for alert repository
type AlertRepositoryInterface interface {
	Create(ctx context.Context, alert *model.Alert) error
	List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error)
	CountActive(ctx context.Context) (int, error)
	Resolve(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status string) error
	GetRecentByDevice(ctx context.Context, deviceID string, ruleID int, cooldownSec int) (*model.Alert, error)
	// FIX-P1-01: N+1 查询优化 - 新增批量查询方法
	GetRecentAlertsByDeviceBatch(ctx context.Context, deviceID string, ruleIDs []int, cooldownSec int) (map[int]*model.Alert, error)
}
