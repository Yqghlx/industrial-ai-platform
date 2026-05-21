package service

import (
	"context"
	"testing"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
)

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

// Test GetAlerts
func TestAlertService_GetAlerts(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	expectedAlerts := []model.Alert{
		{ID: 1, DeviceID: "CNC-001"},
		{ID: 2, DeviceID: "CNC-002"},
	}

	mockAlertRepo.On("List", ctx, "", 1, 20).Return(expectedAlerts, 2, nil)

	alerts, total, err := svc.GetAlerts(ctx, "", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(alerts))
	assert.Equal(t, 2, total)
	mockAlertRepo.AssertExpectations(t)
}

// Test ResolveAlert
func TestAlertService_ResolveAlert(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("Resolve", ctx, 1).Return(nil)

	err := svc.ResolveAlert(ctx, 1)
	assert.NoError(t, err)
	mockAlertRepo.AssertExpectations(t)
}

// Test AcknowledgeAlert
func TestAlertService_AcknowledgeAlert(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("UpdateStatus", ctx, 1, "acknowledged").Return(nil)

	err := svc.AcknowledgeAlert(ctx, 1)
	assert.NoError(t, err)
	mockAlertRepo.AssertExpectations(t)
}

// Test GetAlertByID
func TestAlertService_GetAlertByID(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	expectedAlerts := []model.Alert{
		{ID: 1, DeviceID: "CNC-001"},
	}

	mockAlertRepo.On("List", ctx, "all", 1, 1000).Return(expectedAlerts, 1, nil)

	alert, err := svc.GetAlertByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, &expectedAlerts[0], alert)
	mockAlertRepo.AssertExpectations(t)
}

// Test GetAlertByID_NotFound
func TestAlertService_GetAlertByID_NotFound(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	expectedAlerts := []model.Alert{}

	mockAlertRepo.On("List", ctx, "all", 1, 1000).Return(expectedAlerts, 0, nil)

	alert, err := svc.GetAlertByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, alert)
	mockAlertRepo.AssertExpectations(t)
}

// Test CountActiveAlerts (via Repository)
func TestAlertService_CountActiveAlerts(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	ctx := context.Background()

	mockAlertRepo.On("CountActive", ctx).Return(5, nil)

	count, err := mockAlertRepo.CountActive(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	mockAlertRepo.AssertExpectations(t)
}