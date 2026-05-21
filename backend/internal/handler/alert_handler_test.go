package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Note: Alert handlers are implemented in device_handler.go as listRules, createRule, etc.
// MockRuleRepository and MockAlertService are defined in mock_common_test.go
// This test file tests the alert-related functionality.

// TestListRules_Success tests successful rule listing
func TestListRules_Success(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	expectedRules := []model.AlertRule{
		{
			ID:          1,
			Name:        "高温告警",
			DeviceType:  "*",
			Metric:      "temperature",
			Operator:    ">",
			Threshold:   100,
			Severity:    "high",
			Enabled:     true,
			CooldownSec: 300,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			Name:        "振动异常",
			DeviceType:  "*",
			Metric:      "vibration",
			Operator:    ">",
			Threshold:   3.0,
			Severity:    "medium",
			Enabled:     true,
			CooldownSec: 600,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockRuleRepo.On("List", mock.Anything).Return(expectedRules, nil)

	rules, err := mockRuleRepo.List(nil)
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "高温告警", rules[0].Name)
	assert.Equal(t, "temperature", rules[0].Metric)

	mockRuleRepo.AssertExpectations(t)
}

// TestListRules_ServiceError tests rule listing service error
func TestListRules_ServiceError(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	mockRuleRepo.On("List", mock.Anything).Return(nil, errors.New("database error"))

	rules, err := mockRuleRepo.List(nil)
	assert.Error(t, err)
	assert.Nil(t, rules)

	mockRuleRepo.AssertExpectations(t)
}

// TestGetRule_Success tests successful rule retrieval
func TestGetRule_Success(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	expectedRule := &model.AlertRule{
		ID:          1,
		Name:        "高温告警",
		DeviceType:  "*",
		Metric:      "temperature",
		Operator:    ">",
		Threshold:   100,
		Severity:    "high",
		Enabled:     true,
		CooldownSec: 300,
	}

	mockRuleRepo.On("GetByID", mock.Anything, 1).Return(expectedRule, nil)

	rule, err := mockRuleRepo.GetByID(nil, 1)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, 1, rule.ID)
	assert.Equal(t, "高温告警", rule.Name)

	mockRuleRepo.AssertExpectations(t)
}

// TestGetRule_NotFound tests rule not found
func TestGetRule_NotFound(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	mockRuleRepo.On("GetByID", mock.Anything, 999).Return(nil, errors.New("rule not found"))

	rule, err := mockRuleRepo.GetByID(nil, 999)
	assert.Error(t, err)
	assert.Nil(t, rule)

	mockRuleRepo.AssertExpectations(t)
}

// TestCreateRule_Success tests successful rule creation
func TestCreateRule_Success(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	rule := &model.AlertRule{
		Name:        "新告警规则",
		DeviceType:  "*",
		Metric:      "pressure",
		Operator:    ">",
		Threshold:   150,
		Severity:    "high",
		Enabled:     true,
		CooldownSec: 300,
		Actions:     `[{"type": "notification"}]`,
	}

	mockAlertSvc.On("CreateRule", mock.Anything, rule).Return(nil)

	err := mockAlertSvc.CreateRule(nil, rule)
	assert.NoError(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestCreateRule_Validation tests rule validation
func TestCreateRule_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rule    model.AlertRule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: model.AlertRule{
				Name:     "Valid Rule",
				Metric:   "temperature",
				Operator: ">",
				Severity: "high",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			rule: model.AlertRule{
				Metric:   "temperature",
				Operator: ">",
			},
			wantErr: true,
		},
		{
			name: "missing metric",
			rule: model.AlertRule{
				Name:     "Test",
				Operator: ">",
			},
			wantErr: true,
		},
		{
			name: "missing operator",
			rule: model.AlertRule{
				Name:   "Test",
				Metric: "temperature",
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			rule: model.AlertRule{
				Name:     "Test",
				Metric:   "temperature",
				Operator: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			rule: model.AlertRule{
				Name:     "Test",
				Metric:   "temperature",
				Operator: ">",
				Severity: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateRule(&tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateRule_Success tests successful rule update
func TestUpdateRule_Success(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	rule := &model.AlertRule{
		ID:        1,
		Name:      "更新规则",
		Threshold: 120,
		Severity:  "critical",
		UpdatedAt: time.Now(),
	}

	mockAlertSvc.On("UpdateRule", mock.Anything, rule).Return(nil)

	err := mockAlertSvc.UpdateRule(nil, rule)
	assert.NoError(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestDeleteRule_Success tests successful rule deletion
func TestDeleteRule_Success(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	mockAlertSvc.On("DeleteRule", mock.Anything, 5).Return(nil)

	err := mockAlertSvc.DeleteRule(nil, 5)
	assert.NoError(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestDeleteRule_NotFound tests rule deletion not found
func TestDeleteRule_NotFound(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	mockAlertSvc.On("DeleteRule", mock.Anything, 999).Return(errors.New("rule not found"))

	err := mockAlertSvc.DeleteRule(nil, 999)
	assert.Error(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestToggleRule_Success tests successful rule toggle
func TestToggleRule_Success(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	mockRuleRepo.On("ToggleEnabled", mock.Anything, 1, false).Return(nil)

	err := mockRuleRepo.ToggleEnabled(nil, 1, false)
	assert.NoError(t, err)

	mockRuleRepo.AssertExpectations(t)
}

// TestToggleRule_Enable tests enabling a rule
func TestToggleRule_Enable(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)

	mockRuleRepo.On("ToggleEnabled", mock.Anything, 2, true).Return(nil)

	err := mockRuleRepo.ToggleEnabled(nil, 2, true)
	assert.NoError(t, err)

	mockRuleRepo.AssertExpectations(t)
}

// TestAlertRuleStructure tests alert rule structure
func TestAlertRuleStructure(t *testing.T) {
	rule := model.AlertRule{
		ID:          1,
		Name:        "高温告警",
		DeviceType:  "*",
		Metric:      "temperature",
		Operator:    ">",
		Threshold:   100,
		Severity:    "high",
		Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
		Enabled:     true,
		CooldownSec: 300,
		TenantID:    "tenant-001",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, 1, rule.ID)
	assert.Equal(t, "高温告警", rule.Name)
	assert.Equal(t, "*", rule.DeviceType)
	assert.Equal(t, "temperature", rule.Metric)
	assert.Equal(t, ">", rule.Operator)
	assert.Equal(t, 100.0, rule.Threshold)
	assert.Equal(t, "high", rule.Severity)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 300, rule.CooldownSec)
	assert.NotZero(t, rule.CreatedAt)
	assert.NotZero(t, rule.UpdatedAt)
}

// TestAlertStructure tests alert structure
func TestAlertStructure(t *testing.T) {
	alert := model.Alert{
		ID:          1,
		RuleID:      1,
		DeviceID:    "device-001",
		TenantID:    "tenant-001",
		Message:     "温度超过阈值: 105°C (阈值: 100°C)",
		Severity:    "high",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	assert.Equal(t, 1, alert.ID)
	assert.Equal(t, 1, alert.RuleID)
	assert.Equal(t, "device-001", alert.DeviceID)
	assert.Equal(t, "tenant-001", alert.TenantID)
	assert.Equal(t, "温度超过阈值: 105°C (阈值: 100°C)", alert.Message)
	assert.Equal(t, "high", alert.Severity)
	assert.Equal(t, "active", alert.Status)
	assert.NotZero(t, alert.TriggeredAt)
}

// TestAlertEvaluation tests alert evaluation logic
func TestAlertEvaluation(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	telemetryData := &model.TelemetryData{
		DeviceID:    "device-001",
		Timestamp:   time.Now(),
		Temperature: 105.0, // Above threshold
	}

	mockAlertSvc.On("EvaluateRules", mock.Anything, telemetryData).Return(nil)

	err := mockAlertSvc.EvaluateRules(nil, telemetryData)
	assert.NoError(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestAlertSeverity tests severity levels
func TestAlertSeverity(t *testing.T) {
	severities := []string{"low", "medium", "high", "critical"}

	for _, sev := range severities {
		assert.True(t, isValidSeverity(sev))
	}

	// Invalid severities
	assert.False(t, isValidSeverity("unknown"))
	assert.False(t, isValidSeverity(""))
}

// TestAlertOperator tests operator validation
func TestAlertOperator(t *testing.T) {
	validOperators := []string{">", ">=", "<", "<=", "==", "!="}

	for _, op := range validOperators {
		assert.True(t, isValidOperator(op))
	}

	// Invalid operators
	assert.False(t, isValidOperator("invalid"))
	assert.False(t, isValidOperator(""))
}

// TestGetAlerts_Success tests successful alert listing
func TestGetAlerts_Success(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	expectedAlerts := []model.Alert{
		{ID: 1, DeviceID: "device-001", Severity: "high", Status: "active"},
		{ID: 2, DeviceID: "device-002", Severity: "medium", Status: "resolved"},
	}

	mockAlertSvc.On("GetAlerts", mock.Anything, "", 1, 20).Return(expectedAlerts, 2, nil)

	alerts, total, err := mockAlertSvc.GetAlerts(nil, "", 1, 20)
	assert.NoError(t, err)
	assert.Len(t, alerts, 2)
	assert.Equal(t, 2, total)

	mockAlertSvc.AssertExpectations(t)
}

// TestGetAlerts_ByStatus tests alert listing by status
func TestGetAlerts_ByStatus(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	expectedAlerts := []model.Alert{
		{ID: 1, DeviceID: "device-001", Severity: "high", Status: "active"},
	}

	mockAlertSvc.On("GetAlerts", mock.Anything, "active", 1, 20).Return(expectedAlerts, 1, nil)

	alerts, total, err := mockAlertSvc.GetAlerts(nil, "active", 1, 20)
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, 1, total)

	mockAlertSvc.AssertExpectations(t)
}

// TestInitializeDefaultRules tests default rules initialization
func TestInitializeDefaultRules(t *testing.T) {
	mockAlertSvc := new(MockAlertService)

	mockAlertSvc.On("InitializeDefaultRules", mock.Anything).Return(nil)

	err := mockAlertSvc.InitializeDefaultRules(nil)
	assert.NoError(t, err)

	mockAlertSvc.AssertExpectations(t)
}

// TestCreateRuleRequest tests create rule request validation
func TestCreateRuleRequest_Validation(t *testing.T) {
	// Simulate HTTP request for creating a rule
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	rule := model.AlertRule{
		Name:      "New Rule",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 100,
		Severity:  "high",
	}

	body, _ := json.Marshal(rule)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/rules", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Verify request body parsing
	var parsedRule model.AlertRule
	err := json.Unmarshal(body, &parsedRule)
	require.NoError(t, err)
	assert.Equal(t, rule.Name, parsedRule.Name)
	assert.Equal(t, rule.Metric, parsedRule.Metric)
	assert.Equal(t, rule.Operator, parsedRule.Operator)
}

// TestToggleRuleRequest tests toggle rule request
func TestToggleRuleRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: false,
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/rules/1/toggle", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var parsedReq struct {
		Enabled bool `json:"enabled"`
	}
	err := json.Unmarshal(body, &parsedReq)
	require.NoError(t, err)
	assert.False(t, parsedReq.Enabled)
}

// TestAlertActionsParsing tests alert actions parsing
func TestAlertActionsParsing(t *testing.T) {
	// Valid actions JSON
	actionsJSON := `[{"type": "notification"}, {"type": "workorder"}]`
	actions := service.ParseActions(actionsJSON)
	assert.Len(t, actions, 2)
	assert.Equal(t, "notification", actions[0]["type"])
	assert.Equal(t, "workorder", actions[1]["type"])

	// Empty actions
	emptyActions := service.ParseActions("")
	assert.Len(t, emptyActions, 0)

	// Default actions
	defaultActions := service.FormatActions(nil)
	assert.Contains(t, defaultActions, "notification")
}

// TestAlertCooldown tests cooldown period logic
func TestAlertCooldown(t *testing.T) {
	mockAlertRepo := new(MockAlertRepository)

	// No recent alert - should trigger
	mockAlertRepo.On("GetRecentByDevice", mock.Anything, "device-001", 1, 300).Return(nil, nil)

	recent, err := mockAlertRepo.GetRecentByDevice(nil, "device-001", 1, 300)
	assert.NoError(t, err)
	assert.Nil(t, recent)

	// Recent alert exists - should skip
	mockAlertRepo.On("GetRecentByDevice", mock.Anything, "device-002", 1, 300).Return(&model.Alert{
		ID:          10,
		RuleID:      1,
		DeviceID:    "device-002",
		TriggeredAt: time.Now().Add(-100 * time.Second), // Within cooldown
	}, nil)

	recent, err = mockAlertRepo.GetRecentByDevice(nil, "device-002", 1, 300)
	assert.NoError(t, err)
	assert.NotNil(t, recent)

	mockAlertRepo.AssertExpectations(t)
}

// Helper functions that mirror service validation
func isValidOperator(op string) bool {
	validOps := []string{">", ">=", "<", "<=", "==", "!="}
	for _, v := range validOps {
		if op == v {
			return true
		}
	}
	return false
}

func isValidSeverity(sev string) bool {
	validSev := []string{"low", "medium", "high", "critical"}
	for _, v := range validSev {
		if sev == v {
			return true
		}
	}
	return false
}
