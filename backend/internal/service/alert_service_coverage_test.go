package service

import (
	"context"
	"testing"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test severityToPriority
func TestAlertService_severityToPriority(t *testing.T) {
	svc := &AlertService{}

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "high"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"unknown", "low"},
	}

	for _, tt := range tests {
		result := svc.severityToPriority(tt.severity)
		assert.Equal(t, tt.expected, result, "severity=%s", tt.severity)
	}
}

// Test CreateRule
func TestAlertService_CreateRule(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{
		Name:      "test-rule",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 80,
		Severity:  "high",
		Actions:   "notify",
		Enabled:   true,
	}

	mockRuleRepo.On("Create", ctx, rule).Return(nil)

	err := svc.CreateRule(ctx, rule)
	assert.NoError(t, err)
	mockRuleRepo.AssertExpectations(t)
}

// Test UpdateRule
func TestAlertService_UpdateRule(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{ID: 1, Name: "updated-rule"}

	mockRuleRepo.On("Update", ctx, rule).Return(nil)

	err := svc.UpdateRule(ctx, rule)
	assert.NoError(t, err)
	mockRuleRepo.AssertExpectations(t)
}

// Test DeleteRule
func TestAlertService_DeleteRule(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()

	mockRuleRepo.On("Delete", ctx, 1).Return(nil)

	err := svc.DeleteRule(ctx, 1)
	assert.NoError(t, err)
	mockRuleRepo.AssertExpectations(t)
}

// Test GetRules
func TestAlertService_GetRules(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	expectedRules := []model.AlertRule{
		{ID: 1, Name: "rule-1"},
		{ID: 2, Name: "rule-2"},
	}

	mockRuleRepo.On("List", ctx).Return(expectedRules, nil)

	rules, err := svc.GetRules(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rules))
	mockRuleRepo.AssertExpectations(t)
}

// Test GetRuleByID
func TestAlertService_GetRuleByID(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	expectedRule := &model.AlertRule{ID: 1, Name: "test-rule"}

	mockRuleRepo.On("GetByID", ctx, 1).Return(expectedRule, nil)

	rule, err := svc.GetRuleByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedRule, rule)
	mockRuleRepo.AssertExpectations(t)
}

// MockAlertRepo for alert tests
type MockAlertRepo struct {
	mock.Mock
}

func (m *MockAlertRepo) GetByID(ctx context.Context, id int) (*model.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

func (m *MockAlertRepo) Resolve(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepo) Acknowledge(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepo) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	return args.Get(0).([]model.Alert), args.Get(1).(int), args.Error(2)
}

func (m *MockAlertRepo) Create(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepo) Update(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

// Test ResolveAlert (需要 AlertRepositoryInterface 实现)
func TestAlertService_ResolveAlert_NoRepo(t *testing.T) {
	// AlertService.ResolveAlert 直接调用 alertRepo.Resolve
	// 如果 alertRepo 为 nil，会 panic
	// 这里跳过测试，等待 AlertRepositoryInterface 实现
}

// Test AcknowledgeAlert_NoRepo
func TestAlertService_AcknowledgeAlert_NoRepo(t *testing.T) {
	// 同上，等待 AlertRepositoryInterface 实现
}