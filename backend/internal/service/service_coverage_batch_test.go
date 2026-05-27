package service

import (
	"context"
	"testing"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================
// AlertService Coverage Tests
// ============================================

func TestAlertService_NewAlertService(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	mockAlertRepo := &repository.MockAlertRepository{}
	mockNotificationRepo := &repository.MockNotificationRepository{}
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}

	svc := NewAlertService(
		mockRuleRepo,
		mockAlertRepo,
		mockNotificationRepo,
		mockWorkOrderRepo,
		mockBlackBoxRepo,
		mockTelemetryRepo,
		mockDeviceRepo,
		AlertServiceConfig{},
	)

	assert.NotNil(t, svc)
}

func TestAlertService_InitializeDefaultRules(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()

	for i := 0; i < 6; i++ {
		mockRuleRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.Anything).Return(nil)
	}

	err := svc.InitializeDefaultRules(ctx)
	assert.NoError(t, err)
}

func TestAlertService_GetTrendReport(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	now := time.Now()

	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", 1, 100).Return([]model.Alert{
		{ID: 1, Severity: "high", TriggeredAt: now},
		{ID: 2, Severity: "medium", TriggeredAt: now.Add(-24 * time.Hour)},
	}, 2, nil)

	report, err := svc.GetTrendReport(ctx, "7d")
	assert.NoError(t, err)
	assert.NotNil(t, report)
}

func TestAlertService_GetDeviceRanking(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", 1, 100).Return([]model.Alert{
		{ID: 1, DeviceID: "CNC-001", Severity: "high"},
		{ID: 2, DeviceID: "CNC-002", Severity: "low"},
	}, 2, nil)

	ranking, err := svc.GetDeviceRanking(ctx, 10)
	assert.NoError(t, err)
	assert.NotNil(t, ranking)
}

func TestAlertService_GetEfficiencyReport(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("CountActive", mock.MatchedBy(func(ctx context.Context) bool { return true })).Return(5, nil)
	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", 1, 100).Return([]model.Alert{}, 0, nil)

	report, err := svc.GetEfficiencyReport(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
}

func TestAlertService_CreateRule(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{
		Name:     "温度告警",
		Severity: "high",
	}

	mockRuleRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), rule).Return(nil)

	err := svc.CreateRule(ctx, rule)
	assert.NoError(t, err)
}

func TestAlertService_GetRules(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rules := []model.AlertRule{
		{ID: 1, Name: "温度告警"},
		{ID: 2, Name: "压力告警"},
	}

	mockRuleRepo.On("List", ctx).Return(rules, nil)

	result, err := svc.GetRules(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestAlertService_GetAlerts(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	alerts := []model.Alert{
		{ID: 1, Severity: "high"},
		{ID: 2, Severity: "medium"},
	}

	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "active", 1, 10).Return(alerts, 2, nil)

	result, total, err := svc.GetAlerts(ctx, "active", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))
}

func TestAlertService_GetAlertByID(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	alert := &model.Alert{ID: 1, Severity: "high"}

	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "all", 1, 1000).Return([]model.Alert{*alert}, 1, nil)

	result, err := svc.GetAlertByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, alert.ID, result.ID)
}

func TestAlertService_ResolveAlert(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("Resolve", ctx, 1).Return(nil)

	err := svc.ResolveAlert(ctx, 1)
	assert.NoError(t, err)
}

func TestAlertService_AcknowledgeAlert(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &AlertService{alertRepo: mockAlertRepo}

	ctx := context.Background()

	mockAlertRepo.On("UpdateStatus", ctx, 1, "acknowledged").Return(nil)

	err := svc.AcknowledgeAlert(ctx, 1)
	assert.NoError(t, err)
}

func TestAlertService_GetRuleByID(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()
	rule := &model.AlertRule{ID: 1, Name: "温度告警"}

	mockRuleRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(rule, nil)

	result, err := svc.GetRuleByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, rule.ID, result.ID)
}

func TestAlertService_DeleteRule(t *testing.T) {
	mockRuleRepo := &repository.MockRuleRepository{}
	svc := &AlertService{ruleRepo: mockRuleRepo}

	ctx := context.Background()

	mockRuleRepo.On("Delete", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(nil)

	err := svc.DeleteRule(ctx, 1)
	assert.NoError(t, err)
}

// ============================================
// DeviceService Coverage Tests
// =============================================

func TestDeviceService_Delete(t *testing.T) {
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewDeviceService(mockDeviceRepo, mockUserRepo)

	ctx := context.Background()

	// Use mock.MatchedBy to handle context wrapping by ensureContextTimeout
	mockDeviceRepo.On("Delete", mock.MatchedBy(func(ctx context.Context) bool { return true }), "CNC-001").Return(nil)

	err := svc.Delete(ctx, "CNC-001")
	assert.NoError(t, err)
}

func TestDeviceService_UpdateStatus(t *testing.T) {
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewDeviceService(mockDeviceRepo, mockUserRepo)

	ctx := context.Background()

	// Use mock.MatchedBy to handle context wrapping by ensureContextTimeout
	mockDeviceRepo.On("UpdateStatus", mock.MatchedBy(func(ctx context.Context) bool { return true }), "CNC-001", "offline").Return(nil)

	err := svc.UpdateStatus(ctx, "CNC-001", "offline")
	assert.NoError(t, err)
}

func TestDeviceService_GetDeviceTypeFromID(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"CNC-001", "数控机床"},
		{"INJ-001", "注塑机"},
		{"ROB-001", "工业机器人"},
		{"CNV-001", "传送带"},
		{"ASM-001", "装配线"},
		{"UNKNOWN-001", "未知设备"},
	}

	for _, tt := range tests {
		result := GetDeviceTypeFromID(tt.id)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDeviceService_GetDeviceNameFromType(t *testing.T) {
	tests := []struct {
		typeStr  string
		expected string
	}{
		{"数控机床", "数控机床"},
		{"CNC", "数控机床"},
		{"注塑机", "注塑机"},
		{"INJ", "注塑机"},
		{"工业机器人", "工业机器人"},
		{"ROB", "工业机器人"},
		{"装配线", "装配线"},
		{"其他", "工业设备"},
	}

	for _, tt := range tests {
		result := GetDeviceNameFromType(tt.typeStr)
		assert.Equal(t, tt.expected, result)
	}
}

// ============================================
// AuthService Coverage Tests
// ============================================

func TestAuthService_NewAuthService(t *testing.T) {
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewAuthService(mockUserRepo)
	assert.NotNil(t, svc)
}

// ============================================
// ExportService Coverage Tests
// ============================================

func TestExportService_NewExportService(t *testing.T) {
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockAlertRepo := &repository.MockAlertRepository{}
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}

	svc := NewExportService(
		mockDeviceRepo,
		mockTelemetryRepo,
		mockAlertRepo,
		mockWorkOrderRepo,
		nil,
	)
	assert.NotNil(t, svc)
}

func TestExportService_GenerateDeviceReportData(t *testing.T) {
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	svc := &ExportService{
		deviceRepo:    mockDeviceRepo,
		telemetryRepo: mockTelemetryRepo,
	}

	ctx := context.Background()
	req := &ExportRequest{
		ReportType: "devices",
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
	}

	devices := []model.Device{
		{ID: "CNC-001", Name: "数控机床1", Status: "online"},
		{ID: "CNC-002", Name: "数控机床2", Status: "offline"},
	}

	mockDeviceRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1, 100).Return(devices, 2, nil)
	// Performance optimization: use batch query mock instead of individual GetStats calls
	mockTelemetryRepo.On("GetStatsBatch", mock.MatchedBy(func(ctx context.Context) bool { return true }), []string{"CNC-001", "CNC-002"}, mock.Anything, mock.Anything).
		Return(map[string]*model.DeviceStats{
			"CNC-001": &model.DeviceStats{DeviceID: "CNC-001"},
			"CNC-002": &model.DeviceStats{DeviceID: "CNC-002"},
		}, nil)

	data := svc.generateDeviceReportData(ctx, req)
	assert.NotNil(t, data)
	assert.Equal(t, 2, data.Summary.TotalDevices)
}

func TestExportService_GenerateAlertReportData(t *testing.T) {
	mockAlertRepo := &repository.MockAlertRepository{}
	svc := &ExportService{alertRepo: mockAlertRepo}

	ctx := context.Background()
	req := &ExportRequest{ReportType: "alerts"}

	alerts := []model.Alert{
		{ID: 1, DeviceID: "CNC-001", Severity: "high", Status: "active"},
		{ID: 2, DeviceID: "CNC-002", Severity: "low", Status: "resolved"},
	}

	// Use mock.MatchedBy to handle context wrapping by ensureContextTimeout
	mockAlertRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", 1, 100).Return(alerts, 2, nil)

	data := svc.generateAlertReportData(ctx, req)
	assert.NotNil(t, data)
	assert.Equal(t, 2, data.AlertStats.TotalAlerts)
}

// ============================================
// WorkOrderService Coverage Tests
// ============================================

func TestWorkOrderService_Create(t *testing.T) {
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewWorkOrderService(mockWorkOrderRepo)

	ctx := context.Background()
	order := &model.WorkOrder{
		Title:    "维修工单",
		DeviceID: "CNC-001",
		Priority: "high",
	}

	mockWorkOrderRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), order).Return(nil)

	err := svc.Create(ctx, order)
	assert.NoError(t, err)
}

func TestWorkOrderService_GetByID(t *testing.T) {
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewWorkOrderService(mockWorkOrderRepo)

	ctx := context.Background()
	order := &model.WorkOrder{ID: 1, Title: "维修工单"}

	mockWorkOrderRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(order, nil)

	result, err := svc.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, order, result)
}

func TestWorkOrderService_UpdateStatus(t *testing.T) {
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewWorkOrderService(mockWorkOrderRepo)

	ctx := context.Background()

	mockWorkOrderRepo.On("UpdateStatus", ctx, 1, "completed").Return(nil)

	err := svc.UpdateStatus(ctx, 1, "completed")
	assert.NoError(t, err)
}

func TestWorkOrderService_List(t *testing.T) {
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewWorkOrderService(mockWorkOrderRepo)

	ctx := context.Background()
	orders := []model.WorkOrder{
		{ID: 1, Title: "工单1"},
		{ID: 2, Title: "工单2"},
	}

	mockWorkOrderRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", "", 1, 10).Return(orders, 2, nil)

	result, total, err := svc.List(ctx, "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))
}

// ============================================
// NotificationService Coverage Tests
// ============================================

func TestNotificationService_Create(t *testing.T) {
	mockNotificationRepo := &repository.MockNotificationRepository{}
	svc := NewNotificationService(mockNotificationRepo)

	ctx := context.Background()
	notification := &model.Notification{
		Title:   "告警通知",
		Message: "设备温度过高",
	}

	mockNotificationRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), notification).Return(nil)

	err := svc.Create(ctx, notification)
	assert.NoError(t, err)
}

func TestNotificationService_MarkRead(t *testing.T) {
	mockNotificationRepo := &repository.MockNotificationRepository{}
	svc := NewNotificationService(mockNotificationRepo)

	ctx := context.Background()

	mockNotificationRepo.On("MarkRead", ctx, 1).Return(nil)

	err := svc.MarkRead(ctx, 1)
	assert.NoError(t, err)
}

func TestNotificationService_List(t *testing.T) {
	mockNotificationRepo := &repository.MockNotificationRepository{}
	svc := NewNotificationService(mockNotificationRepo)

	ctx := context.Background()
	notifications := []model.Notification{
		{ID: 1, Title: "通知1"},
		{ID: 2, Title: "通知2"},
	}

	mockNotificationRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "", true, 1, 10).Return(notifications, 2, nil)

	result, total, err := svc.List(ctx, "unread", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))
}

// ============================================
// BlackBoxService Coverage Tests
// ============================================

func TestBlackBoxService_Create(t *testing.T) {
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	svc := NewBlackBoxService(mockBlackBoxRepo)

	ctx := context.Background()
	record := &model.BlackBoxRecord{
		DeviceID: "CNC-001",
	}

	mockBlackBoxRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), record).Return(nil)

	err := svc.Create(ctx, record)
	assert.NoError(t, err)
}

func TestBlackBoxService_List(t *testing.T) {
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	svc := NewBlackBoxService(mockBlackBoxRepo)

	ctx := context.Background()
	records := []model.BlackBoxRecord{
		{ID: 1, DeviceID: "CNC-001"},
		{ID: 2, DeviceID: "CNC-002"},
	}

	mockBlackBoxRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "CNC-001", 1, 10).Return(records, 2, nil)

	result, total, err := svc.List(ctx, "CNC-001", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))
}

// ============================================
// ReportService Coverage Tests
// ============================================

func TestReportService_NewReportService(t *testing.T) {
	mockReportRepo := &repository.MockReportRepository{}
	svc := NewReportService(mockReportRepo, nil, nil, nil, nil)
	assert.NotNil(t, svc)
}

func TestReportService_ListReports(t *testing.T) {
	mockReportRepo := &repository.MockReportRepository{}
	svc := NewReportService(mockReportRepo, nil, nil, nil, nil)

	ctx := context.Background()
	reports := []model.Report{
		{ID: 1, Title: "报告1"},
		{ID: 2, Title: "报告2"},
	}

	mockReportRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), "daily", 1, 10).Return(reports, 2, nil)

	result, total, err := svc.ListReports(ctx, "daily", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))
}

func TestReportService_GetReportByID(t *testing.T) {
	mockReportRepo := &repository.MockReportRepository{}
	svc := NewReportService(mockReportRepo, nil, nil, nil, nil)

	ctx := context.Background()
	report := &model.Report{ID: 1, Title: "测试报告"}

	mockReportRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(report, nil)

	result, err := svc.GetReportByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, report, result)
}

func TestReportService_DeleteReport_Coverage(t *testing.T) {
	mockReportRepo := &repository.MockReportRepository{}
	svc := NewReportService(mockReportRepo, nil, nil, nil, nil)

	ctx := context.Background()

	mockReportRepo.On("Delete", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(nil)

	err := svc.DeleteReport(ctx, 1)
	assert.NoError(t, err)
}

// ============================================
// AgentService Coverage Tests
// ============================================

func TestAgentService_NewAgentService(t *testing.T) {
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}

	svc := NewAgentService(
		mockTaskLogRepo,
		mockDeviceRepo,
		mockTelemetryRepo,
		nil, // OPT-002: No cache for test
	)
	assert.NotNil(t, svc)
}

func TestAgentService_GetTaskLogs(t *testing.T) {
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	svc := &AgentService{taskLogRepo: mockTaskLogRepo}

	ctx := context.Background()
	logs := []model.AgentTaskLog{
		{SessionID: "session-1", Query: "温度查询"},
		{SessionID: "session-2", Query: "故障预测"},
	}

	mockTaskLogRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool { return true }), 10).Return(logs, nil)

	result, err := svc.GetTaskLogs(ctx, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

// ============================================
// UserService Coverage Tests
// ============================================

func TestUserService_NewUserService(t *testing.T) {
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewUserService(mockUserRepo)
	assert.NotNil(t, svc)
}

func TestUserService_GetByID_Simple(t *testing.T) {
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewUserService(mockUserRepo)

	user := &model.User{ID: 1, Username: "testuser"}

	mockUserRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool { return true }), 1).Return(user, nil)

	result, err := svc.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

// ============================================
// TenantService Coverage Tests
// =============================================

func TestTenantService_NewTenantService(t *testing.T) {
	mockTenantRepo := &repository.MockTenantRepository{}
	svc := NewTenantService(mockTenantRepo)
	assert.NotNil(t, svc)
}