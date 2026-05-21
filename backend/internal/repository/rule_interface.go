package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
)

// RuleRepositoryInterface defines the interface for alert rule repository
type RuleRepositoryInterface interface {
	Create(ctx context.Context, rule *model.AlertRule) error
	List(ctx context.Context) ([]model.AlertRule, error)
	ListEnabled(ctx context.Context) ([]model.AlertRule, error)
	ToggleEnabled(ctx context.Context, id int, enabled bool) error
}
