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
}
