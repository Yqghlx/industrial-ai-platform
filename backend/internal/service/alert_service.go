// Package service provides business logic services
// BE-P2-02: 使用常量替换魔法数字
// BE-P2-05: 统一错误处理
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/constants"
	"github.com/industrial-ai/platform/pkg/errors"
	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/industrial-ai/platform/pkg/notify"
	"go.uber.org/zap"
)

// AlertServiceConfig holds configuration for AlertService
type AlertServiceConfig struct {
	FeishuWebhook string
	NotifyEnabled bool
}

// AlertService handles alert evaluation and actions
type AlertService struct {
	ruleRepo         repository.RuleRepositoryInterface
	alertRepo        repository.AlertRepositoryInterface
	notificationRepo repository.NotificationRepositoryInterface
	workOrderRepo    repository.WorkOrderRepositoryInterface
	blackBoxRepo     repository.BlackBoxRepositoryInterface
	telemetryRepo    repository.TelemetryRepositoryInterface
	deviceRepo       repository.DeviceRepositoryInterface
	notifyManager    *notify.NotifyManager
	config           AlertServiceConfig
	// broadcastFn WebSocket 广播函数，由 handler 层注入
	// 如果为 nil，广播消息将被静默丢弃（兼容旧调用方式）
	broadcastFn func(msg model.WSMessage)
}

// NewAlertService creates a new alert service
func NewAlertService(
	ruleRepo repository.RuleRepositoryInterface,
	alertRepo repository.AlertRepositoryInterface,
	notificationRepo repository.NotificationRepositoryInterface,
	workOrderRepo repository.WorkOrderRepositoryInterface,
	blackBoxRepo repository.BlackBoxRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	deviceRepo repository.DeviceRepositoryInterface,
	config AlertServiceConfig,
) *AlertService {
	return &AlertService{
		ruleRepo:         ruleRepo,
		alertRepo:        alertRepo,
		notificationRepo: notificationRepo,
		workOrderRepo:    workOrderRepo,
		blackBoxRepo:     blackBoxRepo,
		telemetryRepo:    telemetryRepo,
		deviceRepo:       deviceRepo,
		config:           config,
		notifyManager:    notify.NewNotifyManager(config.FeishuWebhook, config.NotifyEnabled),
	}
}

// SetBroadcastFn 设置 WebSocket 广播函数
// 由 handler 层在创建 AlertService 后调用，注入实际的广播实现
func (s *AlertService) SetBroadcastFn(fn func(msg model.WSMessage)) {
	s.broadcastFn = fn
}

// broadcast 安全地执行 WebSocket 广播
// 如果 broadcastFn 未设置，静默丢弃消息（不阻塞业务逻辑）
func (s *AlertService) broadcast(msg model.WSMessage) {
	if s.broadcastFn != nil {
		s.broadcastFn(msg)
	}
}

// EvaluateRules evaluates all enabled rules against telemetry data
func (s *AlertService) EvaluateRules(ctx context.Context, data *model.TelemetryData) error {
	// Get device info
	device, err := s.deviceRepo.GetByID(ctx, data.DeviceID)
	if err != nil {
		logger.L().Warn("Device not found for alert evaluation",
			zap.String("device_id", data.DeviceID),
		)
		device = &model.Device{ID: data.DeviceID, Type: "unknown"}
	}
	// Get enabled rules
	rules, err := s.ruleRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rules: %w", err)
	}
	// FIX-P1-01: N+1 查询优化 - 使用批量查询检查 cooldown
	// 1. 过滤出适用于该设备类型的规则
	applicableRules := make([]model.AlertRule, 0)
	ruleIDs := make([]int, 0)
	maxCooldownSec := 0
	for _, rule := range rules {
		if rule.DeviceType == "*" || rule.DeviceType == "" || rule.DeviceType == device.Type {
			applicableRules = append(applicableRules, rule)
			ruleIDs = append(ruleIDs, rule.ID)
			if rule.CooldownSec > maxCooldownSec {
				maxCooldownSec = rule.CooldownSec
			}
		}
	}

	// 2. 批量查询最近的告警（使用最大 cooldown 时间）
	recentAlertsMap := make(map[int]*model.Alert)
	if len(ruleIDs) > 0 {
		recentAlerts, err := s.alertRepo.GetRecentAlertsByDeviceBatch(ctx, data.DeviceID, ruleIDs, maxCooldownSec)
		if err != nil {
			logger.L().Warn("Error batch checking cooldown",
				zap.String("device_id", data.DeviceID),
				zap.Error(err),
			)
			// 继续处理，但不跳过检查 - 使用空 map 作为 fallback
		} else {
			recentAlertsMap = recentAlerts
		}
	}

	// 3. 遍历规则并检查
	for _, rule := range applicableRules {
		// Evaluate rule condition
		if s.evaluateCondition(data, rule) {
			// Check cooldown using pre-fetched data
			if recentAlert, exists := recentAlertsMap[rule.ID]; exists {
				// 检查是否仍在该规则的 cooldown 时间内
				cooldownTime := recentAlert.TriggeredAt.Add(time.Duration(rule.CooldownSec) * time.Second)
				if time.Now().Before(cooldownTime) {
					continue // Still in cooldown period
				}
			}

			// Trigger alert actions
			if err := s.triggerActions(ctx, data, device, rule); err != nil {
				logger.L().Error("Error triggering alert actions",
					zap.String("device_id", data.DeviceID),
					zap.Int("rule_id", rule.ID),
					zap.Error(err),
				)
			}
		}
	}
	return nil
}

// evaluateCondition checks if a rule condition is met
func (s *AlertService) evaluateCondition(data *model.TelemetryData, rule model.AlertRule) bool {
	var value float64
	switch rule.Metric {
	case "temperature":
		value = data.Temperature
	case "pressure":
		value = data.Pressure
	case "vibration":
		value = data.Vibration
	case "humidity":
		value = data.Humidity
	case "power":
		value = data.Power
	default:
		return false
	}

	switch rule.Operator {
	case ">":
		return value > rule.Threshold
	case ">=":
		return value >= rule.Threshold
	case "<":
		return value < rule.Threshold
	case "<=":
		return value <= rule.Threshold
	case "==":
		return value == rule.Threshold
	case "!=":
		return value != rule.Threshold
	default:
		return false
	}
}

// triggerActions executes actions for a triggered rule
func (s *AlertService) triggerActions(ctx context.Context, data *model.TelemetryData, device *model.Device, rule model.AlertRule) error {
	// Create alert record
	alert := &model.Alert{
		RuleID:      rule.ID,
		DeviceID:    data.DeviceID,
		Message:     fmt.Sprintf("%s: %s %.2f (阈值: %.2f)", rule.Name, rule.Metric, s.getMetricValue(data, rule.Metric), rule.Threshold),
		Severity:    rule.Severity,
		Status:      "active",
		TriggeredAt: time.Now(),
	}
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		logger.L().Error("Failed to create alert",
			zap.String("device_id", data.DeviceID),
			zap.Int("rule_id", rule.ID),
			zap.Error(err),
		)
	}

	// Parse actions
	var actions []map[string]interface{}
	if err := json.Unmarshal([]byte(rule.Actions), &actions); err != nil {
		// If Actions is a simple string, create default actions
		actions = []map[string]interface{}{
			{"type": "notification"},
		}
	}

	// Execute each action
	for _, action := range actions {
		actionType, _ := action["type"].(string)
		switch actionType {
		case "notification":
			s.createNotification(ctx, data, device, rule, alert)
			// Also send to Feishu if enabled
			if s.notifyManager != nil {
				if err := s.notifyManager.NotifyAlert(ctx, alert, device, &rule); err != nil {
					logger.L().Warn("Failed to send Feishu notification",
						zap.Error(err),
						zap.Int("alert_id", alert.ID),
					)
				}
			}
		case "workorder":
			s.createWorkOrder(ctx, data, device, rule)
		case "blackbox":
			s.captureBlackBox(ctx, data, device, rule)
		}
	}

	// Broadcast alert via WebSocket
	s.broadcast(model.WSMessage{
		Type: "alert",
		Payload: map[string]interface{}{
			"id":        alert.ID,
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"device_id": data.DeviceID,
			"device":    device.Name,
			"message":   alert.Message,
			"severity":  rule.Severity,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	return nil
}

// getMetricValue returns the metric value from telemetry data
func (s *AlertService) getMetricValue(data *model.TelemetryData, metric string) float64 {
	switch metric {
	case "temperature":
		return data.Temperature
	case "pressure":
		return data.Pressure
	case "vibration":
		return data.Vibration
	case "humidity":
		return data.Humidity
	case "power":
		return data.Power
	default:
		return 0
	}
}

// createNotification creates a notification for an alert
func (s *AlertService) createNotification(ctx context.Context, data *model.TelemetryData, device *model.Device, rule model.AlertRule, alert *model.Alert) error {
	notification := &model.Notification{
		Type:      "alert",
		Title:     fmt.Sprintf("告警: %s", rule.Name),
		Message:   fmt.Sprintf("设备 %s 触发告警: %s", device.Name, alert.Message),
		DeviceID:  &data.DeviceID,
		Read:      false,
		CreatedAt: time.Now(),
	}
	return s.notificationRepo.Create(ctx, notification)
}

// createWorkOrder creates a work order for an alert
func (s *AlertService) createWorkOrder(ctx context.Context, data *model.TelemetryData, device *model.Device, rule model.AlertRule) error {
	wo := &model.WorkOrder{
		Title:       fmt.Sprintf("自动工单: %s - %s", device.Name, rule.Name),
		Description: fmt.Sprintf("设备 %s 触发告警规则 %s，需检查处理。", device.Name, rule.Name),
		DeviceID:    data.DeviceID,
		Priority:    s.severityToPriority(rule.Severity),
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.workOrderRepo.Create(ctx, wo)
}

// captureBlackBox captures telemetry data around the alert
// BE-P2-02: 使用常量替换魔法数字
func (s *AlertService) captureBlackBox(ctx context.Context, data *model.TelemetryData, device *model.Device, rule model.AlertRule) error {
	// Get recent telemetry
	start := time.Now().Add(-time.Duration(constants.BlackBoxSnapshotMinutes) * time.Minute)
	end := time.Now()
	snapshot, err := s.telemetryRepo.GetByDeviceID(ctx, data.DeviceID, start, end, constants.BlackBoxSnapshotLimit)
	if err != nil {
		logger.L().Warn("Failed to get telemetry for blackbox",
			zap.String("device_id", data.DeviceID),
			zap.Error(err),
		)
	}

	summary := fmt.Sprintf("设备 %s 在 %s 触发告警 %s", device.Name, end.Format("2006-01-02 15:04:05"), rule.Name)

	record := &model.BlackBoxRecord{
		DeviceID:    data.DeviceID,
		TriggerType: rule.Name,
		StartTime:   start,
		EndTime:     end,
		Snapshot:    snapshot,
		Summary:     summary,
		CreatedAt:   time.Now(),
	}
	if err := s.blackBoxRepo.Create(ctx, record); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// severityToPriority converts severity to work order priority
func (s *AlertService) severityToPriority(severity string) string {
	switch severity {
	case "critical":
		return "urgent"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

// CreateRule creates a new alert rule
// BE-P2-02: 使用常量替换魔法数字
func (s *AlertService) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	if rule.Actions == "" {
		rule.Actions = `[{"type": "notification"}]`
	}
	if rule.CooldownSec == 0 {
		rule.CooldownSec = constants.DefaultAlertCooldownSec
	}
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// UpdateRule updates an alert rule
// BE-P2-05: 统一错误处理
func (s *AlertService) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// DeleteRule deletes an alert rule
// BE-P2-05: 统一错误处理
func (s *AlertService) DeleteRule(ctx context.Context, id int) error {
	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// GetRules retrieves all alert rules
func (s *AlertService) GetRules(ctx context.Context) ([]model.AlertRule, error) {
	return s.ruleRepo.List(ctx)
}

// GetAlerts retrieves alerts with filters
func (s *AlertService) GetAlerts(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	return s.alertRepo.List(ctx, status, page, pageSize)
}

// GetAlertsWithFilter retrieves alerts with more filters (severity, deviceID)
// P0-03: 将过滤条件传递到数据库层，避免内存过滤
func (s *AlertService) GetAlertsWithFilter(ctx context.Context, status, severity, deviceID string, page, pageSize int) ([]model.Alert, int, error) {
	filter := repository.AlertFilter{
		Status:   status,
		Severity: severity,
		DeviceID: deviceID,
	}
	return s.alertRepo.ListWithFilter(ctx, filter, page, pageSize)
}

// InitializeDefaultRules creates default alert rules
// BE-P2-02: 使用常量替换魔法数字
// P2-002: 优化告警规则，添加分层阈值和更多指标
func (s *AlertService) InitializeDefaultRules(ctx context.Context) error {
	defaultRules := []model.AlertRule{
		// 温度告警 - 分层预警
		{
			Name:        "温度预警",
			DeviceType:  "*",
			Metric:      "temperature",
			Operator:    ">",
			Threshold:   constants.WarningTemperatureThreshold,
			Severity:    "low",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "高温告警",
			DeviceType:  "*",
			Metric:      "temperature",
			Operator:    ">",
			Threshold:   constants.HighTemperatureThreshold,
			Severity:    "high",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
			Enabled:     true,
			CooldownSec: constants.DefaultAlertCooldownSec,
		},
		{
			Name:        "严重高温告警",
			DeviceType:  "*",
			Metric:      "temperature",
			Operator:    ">",
			Threshold:   constants.CriticalTemperatureThreshold,
			Severity:    "critical",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}, {"type": "blackbox"}]`,
			Enabled:     true,
			CooldownSec: constants.ShortAlertCooldownSec,
		},

		// 振动告警 - ISO 10816标准分层
		{
			Name:        "振动预警",
			DeviceType:  "*",
			Metric:      "vibration",
			Operator:    ">",
			Threshold:   constants.WarningVibrationThreshold,
			Severity:    "low",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "振动异常告警",
			DeviceType:  "*",
			Metric:      "vibration",
			Operator:    ">",
			Threshold:   constants.AbnormalVibrationThreshold,
			Severity:    "medium",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "严重振动告警",
			DeviceType:  "*",
			Metric:      "vibration",
			Operator:    ">",
			Threshold:   constants.CriticalVibrationThreshold,
			Severity:    "critical",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}, {"type": "blackbox"}]`,
			Enabled:     true,
			CooldownSec: constants.ShortAlertCooldownSec,
		},

		// 压力告警 - 分层预警
		{
			Name:        "压力预警",
			DeviceType:  "*",
			Metric:      "pressure",
			Operator:    ">",
			Threshold:   constants.WarningPressureThreshold,
			Severity:    "low",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "压力异常告警",
			DeviceType:  "*",
			Metric:      "pressure",
			Operator:    ">",
			Threshold:   constants.AbnormalPressureThreshold,
			Severity:    "high",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
			Enabled:     true,
			CooldownSec: constants.DefaultAlertCooldownSec,
		},
		{
			Name:        "严重压力告警",
			DeviceType:  "*",
			Metric:      "pressure",
			Operator:    ">",
			Threshold:   constants.CriticalPressureThreshold,
			Severity:    "critical",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}, {"type": "blackbox"}]`,
			Enabled:     true,
			CooldownSec: constants.ShortAlertCooldownSec,
		},

		// 湿度告警 - P2-002新增
		{
			Name:        "低湿度警告",
			DeviceType:  "*",
			Metric:      "humidity",
			Operator:    "<",
			Threshold:   constants.LowHumidityThreshold,
			Severity:    "medium",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "高湿度警告",
			DeviceType:  "*",
			Metric:      "humidity",
			Operator:    ">",
			Threshold:   constants.HighHumidityThreshold,
			Severity:    "high",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
			Enabled:     true,
			CooldownSec: constants.DefaultAlertCooldownSec,
		},
		{
			Name:        "严重湿度告警",
			DeviceType:  "*",
			Metric:      "humidity",
			Operator:    ">",
			Threshold:   constants.CriticalHumidityThreshold,
			Severity:    "critical",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
			Enabled:     true,
			CooldownSec: constants.ShortAlertCooldownSec,
		},

		// 功率告警 - P2-002新增
		{
			Name:        "低功率警告",
			DeviceType:  "*",
			Metric:      "power",
			Operator:    "<",
			Threshold:   constants.LowPowerThreshold,
			Severity:    "medium",
			Actions:     `[{"type": "notification"}]`,
			Enabled:     true,
			CooldownSec: constants.LongAlertCooldownSec,
		},
		{
			Name:        "高功率警告",
			DeviceType:  "*",
			Metric:      "power",
			Operator:    ">",
			Threshold:   constants.HighPowerThreshold,
			Severity:    "high",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}]`,
			Enabled:     true,
			CooldownSec: constants.DefaultAlertCooldownSec,
		},
		{
			Name:        "严重功率告警",
			DeviceType:  "*",
			Metric:      "power",
			Operator:    ">",
			Threshold:   constants.CriticalPowerThreshold,
			Severity:    "critical",
			Actions:     `[{"type": "notification"}, {"type": "workorder"}, {"type": "blackbox"}]`,
			Enabled:     true,
			CooldownSec: constants.ShortAlertCooldownSec,
		},
	}

	for i := range defaultRules {
		defaultRules[i].CreatedAt = time.Now()
		defaultRules[i].UpdatedAt = time.Now()
		if err := s.ruleRepo.Create(ctx, &defaultRules[i]); err != nil {
			// Ignore duplicate key errors
			if !strings.Contains(err.Error(), "duplicate") {
				return errors.NewDatabaseError(err.Error())
			}
		}
	}

	return nil
}

// ParseActions parses actions JSON string
func ParseActions(actionsJSON string) []map[string]interface{} {
	var actions []map[string]interface{}
	if actionsJSON != "" {
		if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
			// 解析失败时返回默认值
			return []map[string]interface{}{{"type": "notification"}}
		}
	}
	return actions
}

// FormatActions formats actions to JSON string
func FormatActions(actions []map[string]interface{}) string {
	if len(actions) == 0 {
		return `[{"type": "notification"}]`
	}
	b, err := json.Marshal(actions)
	if err != nil {
		// 序列化失败时返回默认值，并记录错误
		return `[{"type": "notification"}]`
	}
	return string(b)
}

// ParseRuleActions parses rule actions from request
func ParseRuleActions(actions interface{}) string {
	switch v := actions.(type) {
	case string:
		return v
	case []interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return `[{"type": "notification"}]`
		}
		return string(b)
	case map[string]interface{}:
		b, err := json.Marshal([]map[string]interface{}{v})
		if err != nil {
			return `[{"type": "notification"}]`
		}
		return string(b)
	default:
		return `[{"type": "notification"}]`
	}
}

// ValidateRule validates an alert rule
func ValidateRule(rule *model.AlertRule) error {
	if rule.Name == "" {
		return fmt.Errorf("name is required")
	}
	if rule.Metric == "" {
		return fmt.Errorf("metric is required")
	}
	if rule.Operator == "" {
		return fmt.Errorf("operator is required")
	}
	if !isValidOperator(rule.Operator) {
		return fmt.Errorf("invalid operator: %s", rule.Operator)
	}
	if rule.Severity == "" {
		rule.Severity = "medium"
	}
	if !isValidSeverity(rule.Severity) {
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}
	return nil
}

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

// String to int helper
func mustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// ============================================
// Phase 1: 补充AlertService方法 - Handler重构支持
// ============================================

// GetAlertByID retrieves a single alert by ID
func (s *AlertService) GetAlertByID(ctx context.Context, id int) (*model.Alert, error) {
	// 从alertRepo获取所有告警，然后找到匹配的
	alerts, _, err := s.alertRepo.List(ctx, "all", 1, 1000)
	if err != nil {
		return nil, err
	}

	for _, a := range alerts {
		if a.ID == id {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("alert not found")
}

// ResolveAlert marks an alert as resolved
func (s *AlertService) ResolveAlert(ctx context.Context, id int) error {
	return s.alertRepo.Resolve(ctx, id)
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(ctx context.Context, id int) error {
	return s.alertRepo.UpdateStatus(ctx, id, "acknowledged")
}

// GetRuleByID retrieves a single rule by ID
func (s *AlertService) GetRuleByID(ctx context.Context, id int) (*model.AlertRule, error) {
	return s.ruleRepo.GetByID(ctx, id)
}

// ToggleRule toggles a rule's enabled status
func (s *AlertService) ToggleRule(ctx context.Context, id int) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.ruleRepo.ToggleEnabled(ctx, id, !rule.Enabled)
}

// GetTrendReport returns alert trend report (placeholder)
func (s *AlertService) GetTrendReport(ctx context.Context, period string) (map[string]interface{}, error) {
	// 占位实现 - 后续可实现真实统计
	return map[string]interface{}{
		"period":  period,
		"trend":   []interface{}{},
		"message": "Trend report requires full implementation",
	}, nil
}

// GetDeviceRanking returns device ranking by alert count (placeholder)
func (s *AlertService) GetDeviceRanking(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// 占位实现 - 后续可实现真实统计
	return []map[string]interface{}{}, nil
}

// GetEfficiencyReport returns alert handling efficiency (placeholder)
func (s *AlertService) GetEfficiencyReport(ctx context.Context) (map[string]interface{}, error) {
	// 占位实现 - 后续可实现真实统计
	return map[string]interface{}{
		"avg_resolve_time": 0,
		"ack_rate":         0,
		"message":          "Efficiency report requires full implementation",
	}, nil
}
