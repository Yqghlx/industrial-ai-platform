package service

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/mock"
)

// ============================================
// Mock AlertService for Testing
// ============================================

// MockAlertService implements AlertServiceInterface for testing
type MockAlertService struct {
	mock.Mock
}

func (m *MockAlertService) EvaluateRules(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockAlertService) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertService) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertService) DeleteRule(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetRules(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AlertRule), args.Error(1)
}

func (m *MockAlertService) GetAlerts(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
}

func (m *MockAlertService) InitializeDefaultRules(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAlertService) GetAlertByID(ctx context.Context, id int) (*model.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

func (m *MockAlertService) ResolveAlert(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) AcknowledgeAlert(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetRuleByID(ctx context.Context, id int) (*model.AlertRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRule), args.Error(1)
}

func (m *MockAlertService) ToggleRule(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetTrendReport(ctx context.Context, period string) (map[string]interface{}, error) {
	args := m.Called(ctx, period)
	data, _ := args.Get(0).(map[string]interface{})
	return data, args.Error(1)
}

func (m *MockAlertService) GetDeviceRanking(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	args := m.Called(ctx, limit)
	data, _ := args.Get(0).([]map[string]interface{})
	return data, args.Error(1)
}

func (m *MockAlertService) GetEfficiencyReport(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).(map[string]interface{})
	return data, args.Error(1)
}

// ============================================
// Mock AgentService for Testing
// ============================================

// MockAgentService implements AgentServiceInterface for testing
type MockAgentService struct {
	mock.Mock
}

func (m *MockAgentService) Query(ctx context.Context, query model.AgentQuery) (*model.AgentResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentResponse), args.Error(1)
}

func (m *MockAgentService) GetDeviceContext(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAgentService) GetTaskLogs(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AgentTaskLog), args.Error(1)
}
