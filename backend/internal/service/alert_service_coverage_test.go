package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// AlertService Core Method Tests (using repository mocks)
// ============================================

func TestAlertService_EvaluateRules_Coverage(t *testing.T) {
	// Setup mocks from repository package
	mockRuleRepo := &repository.MockRuleRepository{}
	mockAlertRepo := &repository.MockAlertRepository{}
	mockNotifRepo := &repository.MockNotificationRepository{}
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}

	service := NewAlertService(
		mockRuleRepo,
		mockAlertRepo,
		mockNotifRepo,
		mockWorkOrderRepo,
		mockBlackBoxRepo,
		mockTelemetryRepo,
		mockDeviceRepo,
		AlertServiceConfig{NotifyEnabled: false},
	)

	ctx := context.Background()
	data := &model.TelemetryData{
		DeviceID:    "device-1",
		Temperature: 110.0,
		Pressure:    100.0,
		Vibration:   2.0,
		Timestamp:   time.Now(),
	}

	device := &model.Device{
		ID:   "device-1",
		Name: "Test Device",
		Type: "sensor",
	}

	rules := []model.AlertRule{
		{
			ID:          1,
			Name:        "High Temperature",
			DeviceType:  "*",
			Metric:      "temperature",
			Operator:    ">",
			Threshold:   100.0,
			Severity:    "high",
			Actions:     `[{"type":"notification"}]`,
			Enabled:     true,
			CooldownSec: 300,
		},
	}

	// Test case 1: Rule triggers alert
	t.Run("TriggersAlert", func(t *testing.T) {
		mockDeviceRepo.On("GetByID", ctx, "device-1").Return(device, nil).Once()
		mockRuleRepo.On("ListEnabled", ctx).Return(rules, nil).Once()
		mockAlertRepo.On("GetRecentAlertsByDeviceBatch", ctx, "device-1", mock.AnythingOfType("[]int"), 300).Return(map[int]*model.Alert{}, nil).Once()
		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()

		err := service.EvaluateRules(ctx, data)
		assert.NoError(t, err)
	})

	// Test case 2: Device not found - use default device
	t.Run("DeviceNotFound", func(t *testing.T) {
		mockDeviceRepo.On("GetByID", ctx, "device-1").Return(nil, errors.New("not found")).Once()
		mockRuleRepo.On("ListEnabled", ctx).Return(rules, nil).Once()
		mockAlertRepo.On("GetRecentAlertsByDeviceBatch", ctx, "device-1", mock.AnythingOfType("[]int"), 300).Return(map[int]*model.Alert{}, nil).Once()
		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()

		err := service.EvaluateRules(ctx, data)
		assert.NoError(t, err)
	})

	// Test case 3: In cooldown period
	t.Run("InCooldown", func(t *testing.T) {
		recentAlert := &model.Alert{
			ID:          100,
			RuleID:      1,
			DeviceID:    "device-1",
			TriggeredAt: time.Now().Add(-1 * time.Minute),
		}

		mockDeviceRepo.On("GetByID", ctx, "device-1").Return(device, nil).Once()
		mockRuleRepo.On("ListEnabled", ctx).Return(rules, nil).Once()
		mockAlertRepo.On("GetRecentAlertsByDeviceBatch", ctx, "device-1", mock.AnythingOfType("[]int"), 300).Return(map[int]*model.Alert{1: recentAlert}, nil).Once()

		err := service.EvaluateRules(ctx, data)
		assert.NoError(t, err)
	})

	// Test case 4: Rule does not apply to device type
	t.Run("DeviceTypeMismatch", func(t *testing.T) {
		ruleDeviceType := []model.AlertRule{
			{
				ID:          3,
				Name:        "Motor Vibration",
				DeviceType:  "motor",
				Metric:      "vibration",
				Operator:    ">",
				Threshold:   4.0,
				Severity:    "medium",
				Actions:     `[{"type":"notification"}]`,
				Enabled:     true,
				CooldownSec: 300,
			},
		}

		mockDeviceRepo.On("GetByID", ctx, "device-1").Return(device, nil).Once()
		mockRuleRepo.On("ListEnabled", ctx).Return(ruleDeviceType, nil).Once()

		err := service.EvaluateRules(ctx, data)
		assert.NoError(t, err)
	})
}

func TestAlertService_evaluateCondition_Coverage(t *testing.T) {
	service := &AlertService{}

	tests := []struct {
		name     string
		data     *model.TelemetryData
		rule     model.AlertRule
		expected bool
	}{
		{
			name:     "TemperatureGreaterThanThreshold",
			data:     &model.TelemetryData{Temperature: 90.0},
			rule:     model.AlertRule{Metric: "temperature", Operator: ">", Threshold: 80.0},
			expected: true,
		},
		{
			name:     "TemperatureLessThanThreshold",
			data:     &model.TelemetryData{Temperature: 70.0},
			rule:     model.AlertRule{Metric: "temperature", Operator: "<", Threshold: 80.0},
			expected: true,
		},
		{
			name:     "PressureGreaterOrEqual",
			data:     &model.TelemetryData{Pressure: 100.0},
			rule:     model.AlertRule{Metric: "pressure", Operator: ">=", Threshold: 100.0},
			expected: true,
		},
		{
			name:     "VibrationLessOrEqual",
			data:     &model.TelemetryData{Vibration: 3.0},
			rule:     model.AlertRule{Metric: "vibration", Operator: "<=", Threshold: 3.0},
			expected: true,
		},
		{
			name:     "HumidityEqual",
			data:     &model.TelemetryData{Humidity: 50.0},
			rule:     model.AlertRule{Metric: "humidity", Operator: "==", Threshold: 50.0},
			expected: true,
		},
		{
			name:     "PowerNotEqual",
			data:     &model.TelemetryData{Power: 200.0},
			rule:     model.AlertRule{Metric: "power", Operator: "!=", Threshold: 100.0},
			expected: true,
		},
		{
			name:     "ConditionNotMet",
			data:     &model.TelemetryData{Temperature: 70.0},
			rule:     model.AlertRule{Metric: "temperature", Operator: ">", Threshold: 80.0},
			expected: false,
		},
		{
			name:     "UnknownMetric",
			data:     &model.TelemetryData{Temperature: 90.0},
			rule:     model.AlertRule{Metric: "unknown", Operator: ">", Threshold: 80.0},
			expected: false,
		},
		{
			name:     "InvalidOperator",
			data:     &model.TelemetryData{Temperature: 90.0},
			rule:     model.AlertRule{Metric: "temperature", Operator: "invalid", Threshold: 80.0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.evaluateCondition(tt.data, tt.rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertService_triggerActions_Coverage(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	mockNotifRepo := &repository.MockNotificationRepository{}
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}

	service := &AlertService{
		alertRepo:        mockAlertRepo,
		notificationRepo: mockNotifRepo,
		workOrderRepo:    mockWorkOrderRepo,
		blackBoxRepo:     mockBlackBoxRepo,
		telemetryRepo:    mockTelemetryRepo,
		config:           AlertServiceConfig{NotifyEnabled: false},
	}

	ctx := context.Background()
	data := &model.TelemetryData{
		DeviceID:    "device-1",
		Temperature: 110.0,
		Timestamp:   time.Now(),
	}
	device := &model.Device{
		ID:   "device-1",
		Name: "Test Device",
		Type: "sensor",
	}

	t.Run("NotificationAction", func(t *testing.T) {
		rule := model.AlertRule{
			ID:        1,
			Name:      "High Temp",
			Metric:    "temperature",
			Threshold: 100.0,
			Severity:  "high",
			Actions:   `[{"type":"notification"}]`,
		}

		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()

		err := service.triggerActions(ctx, data, device, rule)
		assert.NoError(t, err)
	})

	t.Run("WorkOrderAction", func(t *testing.T) {
		rule := model.AlertRule{
			ID:        2,
			Name:      "Critical Temp",
			Metric:    "temperature",
			Threshold: 120.0,
			Severity:  "critical",
			Actions:   `[{"type":"workorder"}]`,
		}

		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockWorkOrderRepo.On("Create", ctx, mock.AnythingOfType("*model.WorkOrder")).Return(nil).Once()

		err := service.triggerActions(ctx, data, device, rule)
		assert.NoError(t, err)
	})

	t.Run("BlackBoxAction", func(t *testing.T) {
		rule := model.AlertRule{
			ID:        3,
			Name:      "BlackBox Test",
			Metric:    "temperature",
			Threshold: 110.0,
			Severity:  "critical",
			Actions:   `[{"type":"blackbox"}]`,
		}

		telemetryData := []model.TelemetryData{
			{DeviceID: "device-1", Temperature: 105.0, Timestamp: time.Now().Add(-2 * time.Minute)},
			{DeviceID: "device-1", Temperature: 110.0, Timestamp: time.Now()},
		}

		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockTelemetryRepo.On("GetByDeviceID", ctx, "device-1", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 100).Return(telemetryData, nil).Once()
		mockBlackBoxRepo.On("Create", ctx, mock.AnythingOfType("*model.BlackBoxRecord")).Return(nil).Once()

		err := service.triggerActions(ctx, data, device, rule)
		assert.NoError(t, err)
	})

	t.Run("MultipleActions", func(t *testing.T) {
		rule := model.AlertRule{
			ID:        4,
			Name:      "Multiple Actions",
			Metric:    "temperature",
			Threshold: 100.0,
			Severity:  "high",
			Actions:   `[{"type":"notification"},{"type":"workorder"},{"type":"blackbox"}]`,
		}

		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()
		mockWorkOrderRepo.On("Create", ctx, mock.AnythingOfType("*model.WorkOrder")).Return(nil).Once()
		mockTelemetryRepo.On("GetByDeviceID", ctx, "device-1", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 100).Return([]model.TelemetryData{}, nil).Once()
		mockBlackBoxRepo.On("Create", ctx, mock.AnythingOfType("*model.BlackBoxRecord")).Return(nil).Once()

		err := service.triggerActions(ctx, data, device, rule)
		assert.NoError(t, err)
	})

	t.Run("InvalidActionsJSON", func(t *testing.T) {
		rule := model.AlertRule{
			ID:        5,
			Name:      "Invalid Actions",
			Metric:    "temperature",
			Threshold: 100.0,
			Severity:  "medium",
			Actions:   `invalid-json`,
		}

		mockAlertRepo.On("Create", ctx, mock.AnythingOfType("*model.Alert")).Return(nil).Once()
		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()

		err := service.triggerActions(ctx, data, device, rule)
		assert.NoError(t, err)
	})
}

func TestAlertService_getMetricValue_Coverage(t *testing.T) {
	service := &AlertService{}
	data := &model.TelemetryData{
		Temperature: 75.5,
		Pressure:    120.3,
		Vibration:   3.2,
		Humidity:    45.0,
		Power:       200.0,
	}

	tests := []struct {
		metric   string
		expected float64
	}{
		{"temperature", 75.5},
		{"pressure", 120.3},
		{"vibration", 3.2},
		{"humidity", 45.0},
		{"power", 200.0},
		{"unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			result := service.getMetricValue(data, tt.metric)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertService_severityToPriority_Coverage(t *testing.T) {
	service := &AlertService{}

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "urgent"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"unknown", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := service.severityToPriority(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertService_UpdateRule_Coverage(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	service := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{
		ID:        1,
		Name:      "Updated Rule",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 90.0,
	}

	mockRuleRepo.On("Update", ctx, rule).Return(nil).Once()

	err := service.UpdateRule(ctx, rule)
	assert.NoError(t, err)
}

func TestAlertService_ToggleRule_Coverage(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	service := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{
		ID:      1,
		Name:    "Test Rule",
		Enabled: false,
	}

	t.Run("EnableRule", func(t *testing.T) {
		mockRuleRepo.On("GetByID", ctx, 1).Return(rule, nil).Once()
		mockRuleRepo.On("ToggleEnabled", ctx, 1, true).Return(nil).Once()

		err := service.ToggleRule(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("DisableRule", func(t *testing.T) {
		rule.Enabled = true
		mockRuleRepo.On("GetByID", ctx, 1).Return(rule, nil).Once()
		mockRuleRepo.On("ToggleEnabled", ctx, 1, false).Return(nil).Once()

		err := service.ToggleRule(ctx, 1)
		assert.NoError(t, err)
	})
}

func TestAlertService_ParseRuleActions_Coverage(t *testing.T) {
	// Additional coverage tests for ParseRuleActions
	t.Run("NilInput", func(t *testing.T) {
		result := ParseRuleActions(nil)
		assert.Contains(t, result, "notification")
	})

	t.Run("ArrayType", func(t *testing.T) {
		input := []interface{}{map[string]interface{}{"type": "notification"}}
		result := ParseRuleActions(input)
		var resultActions []map[string]interface{}
		err := json.Unmarshal([]byte(result), &resultActions)
		require.NoError(t, err)
		assert.Len(t, resultActions, 1)
	})

	t.Run("MapType", func(t *testing.T) {
		input := map[string]interface{}{"type": "workorder"}
		result := ParseRuleActions(input)
		var resultActions []map[string]interface{}
		err := json.Unmarshal([]byte(result), &resultActions)
		require.NoError(t, err)
		assert.Len(t, resultActions, 1)
	})

	t.Run("StringInput", func(t *testing.T) {
		input := `[{"type":"notification"}]`
		result := ParseRuleActions(input)
		assert.JSONEq(t, input, result)
	})
}

func TestAlertService_CreateNotification_Coverage(t *testing.T) {
	mockNotifRepo := &repository.MockNotificationRepository{}
	service := &AlertService{
		notificationRepo: mockNotifRepo,
		config:           AlertServiceConfig{NotifyEnabled: false},
	}

	ctx := context.Background()
	data := &model.TelemetryData{DeviceID: "device-1", Temperature: 100.0}
	device := &model.Device{ID: "device-1", Name: "Test Device"}
	rule := model.AlertRule{ID: 1, Name: "High Temp", Severity: "high"}
	alert := &model.Alert{ID: 100, Message: "Temperature exceeded"}

	mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Once()

	err := service.createNotification(ctx, data, device, rule, alert)
	assert.NoError(t, err)
}

func TestAlertService_CreateWorkOrder_Coverage(t *testing.T) {
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	service := &AlertService{
		workOrderRepo: mockWorkOrderRepo,
	}

	ctx := context.Background()
	data := &model.TelemetryData{DeviceID: "device-1"}
	device := &model.Device{ID: "device-1", Name: "Test Device"}
	rule := model.AlertRule{Name: "Critical Alert", Severity: "critical"}

	mockWorkOrderRepo.On("Create", ctx, mock.AnythingOfType("*model.WorkOrder")).Return(nil).Once()

	err := service.createWorkOrder(ctx, data, device, rule)
	assert.NoError(t, err)
}

func TestAlertService_CaptureBlackBox_Coverage(t *testing.T) {
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	service := &AlertService{
		telemetryRepo: mockTelemetryRepo,
		blackBoxRepo:  mockBlackBoxRepo,
	}

	ctx := context.Background()
	data := &model.TelemetryData{DeviceID: "device-1", Temperature: 110.0}
	device := &model.Device{ID: "device-1", Name: "Test Device"}
	rule := model.AlertRule{Name: "BlackBox Trigger", Severity: "critical"}

	telemetryData := []model.TelemetryData{
		{DeviceID: "device-1", Temperature: 105.0, Timestamp: time.Now()},
	}

	mockTelemetryRepo.On("GetByDeviceID", ctx, "device-1", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 100).Return(telemetryData, nil).Once()
	mockBlackBoxRepo.On("Create", ctx, mock.AnythingOfType("*model.BlackBoxRecord")).Return(nil).Once()

	err := service.captureBlackBox(ctx, data, device, rule)
	assert.NoError(t, err)
}

func TestAlertService_CaptureBlackBox_TelemetryError(t *testing.T) {
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	service := &AlertService{
		telemetryRepo: mockTelemetryRepo,
		blackBoxRepo:  mockBlackBoxRepo,
	}

	ctx := context.Background()
	data := &model.TelemetryData{DeviceID: "device-1"}
	device := &model.Device{ID: "device-1", Name: "Test Device"}
	rule := model.AlertRule{Name: "Test", Severity: "high"}

	// Telemetry fails, but blackbox should still be created
	mockTelemetryRepo.On("GetByDeviceID", ctx, "device-1", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 100).Return([]model.TelemetryData{}, errors.New("telemetry error")).Once()
	mockBlackBoxRepo.On("Create", ctx, mock.AnythingOfType("*model.BlackBoxRecord")).Return(nil).Once()

	err := service.captureBlackBox(ctx, data, device, rule)
	assert.NoError(t, err) // Should succeed even with telemetry error
}
