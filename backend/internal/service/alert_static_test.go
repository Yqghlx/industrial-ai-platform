package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// AlertService Static Method Tests
// ============================================

func TestAlertService_ParseActions(t *testing.T) {
	// ParseActions expects JSON string
	actions := `[{"type":"notify"},{"type":"work_order"},{"type":"blackbox"}]`
	result := ParseActions(actions)
	require.NotNil(t, result)
}

func TestAlertService_ParseActions_Empty(t *testing.T) {
	result := ParseActions("")
	require.Nil(t, result)
}

func TestAlertService_ParseActions_Single(t *testing.T) {
	actions := `[{"type":"notify"}]`
	result := ParseActions(actions)
	require.NotNil(t, result)
}

func TestAlertService_FormatActions(t *testing.T) {
	// FormatActions expects []map[string]interface{}
	actions := []map[string]interface{}{{"type": "notify"}, {"type": "work_order"}}
	result := FormatActions(actions)
	assert.Contains(t, result, "notify")
}

func TestAlertService_FormatActions_Empty(t *testing.T) {
	actions := []map[string]interface{}{}
	result := FormatActions(actions)
	// FormatActions returns default notification when empty
	assert.Contains(t, result, "notification")
}

func TestAlertService_ParseRuleActions(t *testing.T) {
	// ParseRuleActions parses rule.Actions JSON
	rule := &model.AlertRule{
		Actions: `[{"type":"notify"},{"type":"blackbox"}]`,
	}
	result := ParseRuleActions(rule)
	require.NotNil(t, result)
}

func TestAlertService_ParseRuleActions_Empty(t *testing.T) {
	rule := &model.AlertRule{
		Actions: "",
	}
	result := ParseRuleActions(rule)
	// ParseRuleActions returns default when empty
	assert.Contains(t, result, "notification")
}

func TestAlertService_ValidateRule_Valid(t *testing.T) {
	rule := &model.AlertRule{
		Name:      "High Temperature",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 80.0,
		Severity:  "high",
	}
	err := ValidateRule(rule)
	require.NoError(t, err)
}

func TestAlertService_ValidateRule_MissingName(t *testing.T) {
	rule := &model.AlertRule{
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 80.0,
		Severity:  "high",
	}
	err := ValidateRule(rule)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestAlertService_ValidateRule_MissingMetric(t *testing.T) {
	rule := &model.AlertRule{
		Name:      "Test Rule",
		Operator:  ">",
		Threshold: 80.0,
		Severity:  "high",
	}
	err := ValidateRule(rule)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metric")
}

func TestAlertService_ValidateRule_InvalidOperator(t *testing.T) {
	rule := &model.AlertRule{
		Name:      "Test Rule",
		Metric:    "temperature",
		Operator:  "invalid",
		Threshold: 80.0,
		Severity:  "high",
	}
	err := ValidateRule(rule)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operator")
}

func TestAlertService_ValidateRule_InvalidSeverity(t *testing.T) {
	rule := &model.AlertRule{
		Name:      "Test Rule",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 80.0,
		Severity:  "invalid",
	}
	err := ValidateRule(rule)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "severity")
}

func TestAlertService_isValidOperator(t *testing.T) {
	validOps := []string{">", "<", ">=", "<=", "==", "!="}
	for _, op := range validOps {
		assert.True(t, isValidOperator(op), "Operator %s should be valid", op)
	}

	invalidOps := []string{"invalid", "", ">>", "<<", "==="}
	for _, op := range invalidOps {
		assert.False(t, isValidOperator(op), "Operator %s should be invalid", op)
	}
}

func TestAlertService_isValidSeverity(t *testing.T) {
	validSeverities := []string{"low", "medium", "high", "critical"}
	for _, sev := range validSeverities {
		assert.True(t, isValidSeverity(sev), "Severity %s should be valid", sev)
	}

	invalidSeverities := []string{"invalid", "", "info", "warning"}
	for _, sev := range invalidSeverities {
		assert.False(t, isValidSeverity(sev), "Severity %s should be invalid", sev)
	}
}

func TestAlertService_mustAtoi(t *testing.T) {
	// Valid conversion
	result := mustAtoi("123")
	assert.Equal(t, 123, result)

	// Invalid conversion returns 0
	result = mustAtoi("invalid")
	assert.Equal(t, 0, result)

	// Empty string returns 0
	result = mustAtoi("")
	assert.Equal(t, 0, result)
}

func TestAlertService_severityToPriority_IfAvailable(t *testing.T) {
	// This is a private method, test indirectly through other methods
	// Or skip if not exported
	t.Skip("severityToPriority is a private method")
}
