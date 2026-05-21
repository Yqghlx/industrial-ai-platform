package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// AlertService Tests using MockAlertService
// ============================================

func TestMockAlertService_GetAlertByID(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetAlertByID", ctx, 1).Return(&model.Alert{ID: 1, Message: "Test"}, nil)

	alert, err := mockSvc.GetAlertByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, alert.ID)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetAlertByID_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetAlertByID", ctx, 1).Return(nil, errors.New("not found"))

	alert, err := mockSvc.GetAlertByID(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, alert)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_AcknowledgeAlert(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("AcknowledgeAlert", ctx, 1).Return(nil)

	err := mockSvc.AcknowledgeAlert(ctx, 1)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_AcknowledgeAlert_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("AcknowledgeAlert", ctx, 1).Return(errors.New("db error"))

	err := mockSvc.AcknowledgeAlert(ctx, 1)
	require.Error(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_DeleteRule(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("DeleteRule", ctx, 1).Return(nil)

	err := mockSvc.DeleteRule(ctx, 1)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_DeleteRule_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("DeleteRule", ctx, 1).Return(errors.New("db error"))

	err := mockSvc.DeleteRule(ctx, 1)
	require.Error(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetRules(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	rules := []model.AlertRule{{ID: 1, Name: "Rule 1"}, {ID: 2, Name: "Rule 2"}}
	mockSvc.On("GetRules", ctx).Return(rules, nil)

	result, err := mockSvc.GetRules(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetRules_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetRules", ctx).Return(nil, errors.New("db error"))

	result, err := mockSvc.GetRules(ctx)
	require.Error(t, err)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_CreateRule(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	rule := &model.AlertRule{Name: "New Rule"}
	mockSvc.On("CreateRule", ctx, rule).Return(nil)

	err := mockSvc.CreateRule(ctx, rule)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_CreateRule_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	rule := &model.AlertRule{Name: "New Rule"}
	mockSvc.On("CreateRule", ctx, rule).Return(errors.New("db error"))

	err := mockSvc.CreateRule(ctx, rule)
	require.Error(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_UpdateRule(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	rule := &model.AlertRule{ID: 1, Name: "Updated Rule"}
	mockSvc.On("UpdateRule", ctx, rule).Return(nil)

	err := mockSvc.UpdateRule(ctx, rule)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetTrendReport(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetTrendReport", ctx, "7d").Return(nil, nil)

	_, err := mockSvc.GetTrendReport(ctx, "7d")
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetDeviceRanking(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetDeviceRanking", ctx, 10).Return(nil, nil)

	_, err := mockSvc.GetDeviceRanking(ctx, 10)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockAlertService_GetEfficiencyReport(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAlertService)

	mockSvc.On("GetEfficiencyReport", ctx).Return(nil, nil)

	_, err := mockSvc.GetEfficiencyReport(ctx)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}
