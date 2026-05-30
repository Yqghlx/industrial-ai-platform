package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	apperrors "github.com/industrial-ai/platform/pkg/errors"
)

// ============================================
// 测试辅助函数
// ============================================

// newMockReportService 创建基于 testify/mock 的 ReportService 测试实例
func newMockReportService() (*ReportService, *repository.MockReportRepository, *repository.MockTelemetryRepository, *repository.MockDeviceRepository, *repository.MockWorkOrderRepository, *repository.MockNotificationRepository) {
	reportRepo := new(repository.MockReportRepository)
	telemetryRepo := new(repository.MockTelemetryRepository)
	deviceRepo := new(repository.MockDeviceRepository)
	workOrderRepo := new(repository.MockWorkOrderRepository)
	notificationRepo := new(repository.MockNotificationRepository)

	svc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, notificationRepo)

	return svc, reportRepo, telemetryRepo, deviceRepo, workOrderRepo, notificationRepo
}

// ============================================
// NewReportService 测试
// ============================================

func TestNewReportService(t *testing.T) {
	reportRepo := new(repository.MockReportRepository)
	telemetryRepo := new(repository.MockTelemetryRepository)
	deviceRepo := new(repository.MockDeviceRepository)
	workOrderRepo := new(repository.MockWorkOrderRepository)
	notificationRepo := new(repository.MockNotificationRepository)

	svc := NewReportService(reportRepo, telemetryRepo, deviceRepo, workOrderRepo, notificationRepo)

	assert.NotNil(t, svc)
	assert.Equal(t, reportRepo, svc.reportRepo)
	assert.Equal(t, telemetryRepo, svc.telemetryRepo)
	assert.Equal(t, deviceRepo, svc.deviceRepo)
	assert.Equal(t, workOrderRepo, svc.workOrderRepo)
	assert.Equal(t, notificationRepo, svc.notificationRepo)
}

// ============================================
// GenerateReport 测试
// ============================================

// ---------- daily 类型 ----------

func TestReportService_GenerateReport_Daily_Success(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// GenerateReport 内部调用 ensureContextTimeout 会创建新的 timerCtx
	// 因此 mock 需要使用 mock.Anything 匹配任意 context
	deviceRepo.On("Count", mock.Anything).Return(10, nil)
	telemetryRepo.On("GetLatest", mock.Anything).Return([]model.TelemetryData{}, nil)
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		assert.Contains(t, report.Title, "每日运营报告")
		assert.Equal(t, "daily", report.Type)
		assert.Contains(t, report.Content, "设备概览")
		assert.Nil(t, report.DeviceID)
	})

	report, err := svc.GenerateReport(ctx, "daily", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Contains(t, report.Title, "每日运营报告")
	assert.Equal(t, "daily", report.Type)
	assert.NotEmpty(t, report.Content)
	assert.Nil(t, report.DeviceID)

	reportRepo.AssertExpectations(t)
	deviceRepo.AssertExpectations(t)
	telemetryRepo.AssertExpectations(t)
}

func TestReportService_GenerateReport_Daily_WithTelemetry(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("Count", mock.Anything).Return(5, nil)
	telemetryRepo.On("GetLatest", mock.Anything).Return([]model.TelemetryData{
		{DeviceID: "dev-1", Temperature: 45.5, Vibration: 2.3},
		{DeviceID: "dev-2", Temperature: 50.0, Vibration: 3.0},
	}, nil)
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil)

	report, err := svc.GenerateReport(ctx, "daily", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	// 验证平均温度和振动计算：(45.5 + 50.0) / 2 = 47.75, (2.3 + 3.0) / 2 = 2.65
	assert.Contains(t, report.Content, "47.75")
	assert.Contains(t, report.Content, "2.65")
}

func TestReportService_GenerateReport_Daily_DeviceCountError(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// Count 返回错误，源码中用 _ 忽略错误，deviceCount 为 0
	deviceRepo.On("Count", mock.Anything).Return(0, errors.New("db error"))
	telemetryRepo.On("GetLatest", mock.Anything).Return([]model.TelemetryData{}, nil)
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil)

	report, err := svc.GenerateReport(ctx, "daily", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	// 设备数为 0，报告仍然正常生成
	assert.Contains(t, report.Content, "在线设备数: 0")
}

// ---------- device 类型 ----------

func TestReportService_GenerateReport_Device_Success(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("GetByID", mock.Anything, "device-001").Return(&model.Device{
		ID:       "device-001",
		Name:     "CNC机床-001",
		Type:     "CNC",
		Location: "车间A",
		Status:   "online",
	}, nil)

	telemetryRepo.On("GetStats", mock.Anything, "device-001", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(&model.DeviceStats{
		DeviceID:       "device-001",
		AvgTemperature: 55.2,
		MaxTemperature: 68.1,
		AvgVibration:   3.5,
		MaxVibration:   5.8,
		DataPoints:     1440,
	}, nil)

	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		assert.Contains(t, report.Title, "设备分析报告")
		assert.Equal(t, "device", report.Type)
		require.NotNil(t, report.DeviceID)
		assert.Equal(t, "device-001", *report.DeviceID)
	})

	report, err := svc.GenerateReport(ctx, "device", "device-001")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Contains(t, report.Content, "CNC机床-001")
	assert.Contains(t, report.Content, "车间A")
	assert.Contains(t, report.Content, "55.20")
	assert.Contains(t, report.Content, "68.10")
	assert.Contains(t, report.Content, "1440")
}

func TestReportService_GenerateReport_Device_EmptyDeviceID(t *testing.T) {
	svc, _, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	report, err := svc.GenerateReport(ctx, "device", "")

	assert.Error(t, err)
	assert.Nil(t, report)
	// 验证返回的是 AppError
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeInvalidInput, appErr.Code)
	assert.Contains(t, err.Error(), "Device ID is required")
}

func TestReportService_GenerateReport_Device_DeviceNotFound(t *testing.T) {
	svc, reportRepo, _, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// 设备不存在时，generateDeviceReport 返回错误信息字符串，不返回 error
	deviceRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("device not found"))
	// 仍然会调用 Create 保存报告（内容包含错误信息）
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil)

	report, err := svc.GenerateReport(ctx, "device", "nonexistent")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Contains(t, report.Content, "错误")
	assert.Contains(t, report.Content, "不存在")
}

func TestReportService_GenerateReport_Device_GetStatsError(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("GetByID", mock.Anything, "dev-1").Return(&model.Device{
		ID:       "dev-1",
		Name:     "测试设备",
		Type:     "CNC",
		Location: "车间B",
		Status:   "online",
	}, nil)

	// GetStats 返回错误，源码中判断 err == nil 才输出运行数据
	telemetryRepo.On("GetStats", mock.Anything, "dev-1", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(nil, errors.New("stats unavailable"))
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil)

	report, err := svc.GenerateReport(ctx, "device", "dev-1")

	require.NoError(t, err)
	require.NotNil(t, report)
	// 设备信息正常输出
	assert.Contains(t, report.Content, "测试设备")
	assert.Contains(t, report.Content, "车间B")
	// GetStats 失败时不输出运行数据部分
	assert.NotContains(t, report.Content, "运行数据")
	// 但健康评估仍然输出
	assert.Contains(t, report.Content, "健康评估")
}

// ---------- maintenance 类型 ----------

func TestReportService_GenerateReport_Maintenance_Success(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		assert.Contains(t, report.Title, "维护总结报告")
		assert.Equal(t, "maintenance", report.Type)
	})

	report, err := svc.GenerateReport(ctx, "maintenance", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Contains(t, report.Content, "工单统计")
	assert.Contains(t, report.Content, "维护类型分布")
	assert.Contains(t, report.Content, "下周计划")
}

// ---------- anomaly 类型 ----------

func TestReportService_GenerateReport_Anomaly_Success(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		assert.Contains(t, report.Title, "异常分析报告")
		assert.Equal(t, "anomaly", report.Type)
	})

	report, err := svc.GenerateReport(ctx, "anomaly", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Contains(t, report.Content, "异常概览")
	assert.Contains(t, report.Content, "异常类型分布")
	assert.Contains(t, report.Content, "根因分析")
	assert.Contains(t, report.Content, "改进建议")
}

// ---------- 默认类型（comprehensive） ----------

func TestReportService_GenerateReport_Default_Comprehensive(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// comprehensive 内部调用 generateDailyReport
	deviceRepo.On("Count", mock.Anything).Return(8, nil)
	telemetryRepo.On("GetLatest", mock.Anything).Return([]model.TelemetryData{}, nil)
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		assert.Contains(t, report.Title, "综合报告")
		assert.Equal(t, "unknown_type", report.Type)
	})

	report, err := svc.GenerateReport(ctx, "unknown_type", "")

	require.NoError(t, err)
	require.NotNil(t, report)
	// 综合报告包含每日、维护和异常三部分
	assert.Contains(t, report.Content, "每日运营报告")
	assert.Contains(t, report.Content, "维护总结报告")
	assert.Contains(t, report.Content, "异常分析报告")
}

func TestReportService_GenerateReport_Comprehensive_WithDeviceID(t *testing.T) {
	svc, reportRepo, telemetryRepo, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// comprehensive 内部调用 generateDailyReport
	deviceRepo.On("Count", mock.Anything).Return(3, nil)
	telemetryRepo.On("GetLatest", mock.Anything).Return([]model.TelemetryData{}, nil)
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(nil).Run(func(args mock.Arguments) {
		report := args.Get(1).(*model.Report)
		require.NotNil(t, report.DeviceID)
		assert.Equal(t, "device-999", *report.DeviceID)
	})

	report, err := svc.GenerateReport(ctx, "comprehensive", "device-999")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.NotNil(t, report.DeviceID)
	assert.Equal(t, "device-999", *report.DeviceID)
}

// ---------- Create 失败 ----------

func TestReportService_GenerateReport_RepoCreateError(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	// maintenance 类型不需要其他 repo 调用
	reportRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Report")).Return(errors.New("database error"))

	report, err := svc.GenerateReport(ctx, "maintenance", "")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "database error")
}

// ============================================
// ListReports 测试
// ============================================

func TestReportService_ListReports_Success(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	expectedReports := []model.Report{
		{ID: 1, Title: "每日运营报告", Type: "daily", Content: "content1"},
		{ID: 2, Title: "每日运营报告", Type: "daily", Content: "content2"},
	}

	reportRepo.On("List", mock.Anything, "daily", 1, 10).Return(expectedReports, 10, nil)

	reports, total, err := svc.ListReports(ctx, "daily", 1, 10)

	require.NoError(t, err)
	assert.Equal(t, expectedReports, reports)
	assert.Equal(t, 10, total)
	assert.Len(t, reports, 2)

	reportRepo.AssertExpectations(t)
}

func TestReportService_ListReports_Empty(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("List", mock.Anything, "daily", 1, 10).Return([]model.Report{}, 0, nil)

	reports, total, err := svc.ListReports(ctx, "daily", 1, 10)

	require.NoError(t, err)
	assert.Empty(t, reports)
	assert.Equal(t, 0, total)
}

func TestReportService_ListReports_AllTypes(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	allReports := []model.Report{
		{ID: 1, Title: "报告A", Type: "daily"},
		{ID: 2, Title: "报告B", Type: "maintenance"},
		{ID: 3, Title: "报告C", Type: "anomaly"},
	}
	reportRepo.On("List", mock.Anything, "", 1, 20).Return(allReports, 3, nil)

	reports, total, err := svc.ListReports(ctx, "", 1, 20)

	require.NoError(t, err)
	assert.Len(t, reports, 3)
	assert.Equal(t, 3, total)
}

func TestReportService_ListReports_Error(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("List", mock.Anything, "daily", 1, 10).Return([]model.Report(nil), 0, errors.New("db connection lost"))

	reports, total, err := svc.ListReports(ctx, "daily", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, reports)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "db connection lost")
}

// ============================================
// GetReportByID 测试
// ============================================

func TestReportService_GetReportByID_Success(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	expectedReport := &model.Report{
		ID:          1,
		Title:       "每日运营报告 - 2025-01-15",
		Type:        "daily",
		Content:     "# 每日运营报告\n\n设备概览...",
		GeneratedAt: time.Now(),
	}

	reportRepo.On("GetByID", mock.Anything, 1).Return(expectedReport, nil)

	report, err := svc.GetReportByID(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, 1, report.ID)
	assert.Equal(t, "每日运营报告 - 2025-01-15", report.Title)
	assert.Equal(t, "daily", report.Type)

	reportRepo.AssertExpectations(t)
}

func TestReportService_GetReportByID_NotFound(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("GetByID", mock.Anything, 999).Return(nil, errors.New("report not found"))

	report, err := svc.GetReportByID(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "report not found")
}

func TestReportService_GetReportByID_DBError(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("GetByID", mock.Anything, 1).Return(nil, errors.New("connection refused"))

	report, err := svc.GetReportByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, report)
}

// ============================================
// DeleteReport 测试
// ============================================

func TestReportService_DeleteReport_Success(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("Delete", mock.Anything, 1).Return(nil)

	err := svc.DeleteReport(ctx, 1)

	assert.NoError(t, err)
	reportRepo.AssertExpectations(t)
}

func TestReportService_DeleteReport_NotFound(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("Delete", mock.Anything, 999).Return(errors.New("report not found"))

	err := svc.DeleteReport(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report not found")
}

func TestReportService_DeleteReport_DBError(t *testing.T) {
	svc, reportRepo, _, _, _, _ := newMockReportService()
	ctx := context.Background()

	reportRepo.On("Delete", mock.Anything, 5).Return(errors.New("connection refused"))

	err := svc.DeleteReport(ctx, 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// ============================================
// GetROIStats 测试
// ============================================

func TestReportService_GetROIStats_Success(t *testing.T) {
	svc, _, _, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("Count", mock.Anything).Return(20, nil)

	roiStats, err := svc.GetROIStats(ctx)

	require.NoError(t, err)
	require.NotNil(t, roiStats)
	assert.Equal(t, 20, roiStats.TotalDevices)
	// 每个设备 $5000/月
	assert.Equal(t, 100000.0, roiStats.PredictedSavings)
	assert.Equal(t, 99.5, roiStats.UptimePercentage)
	assert.Equal(t, 2.5, roiStats.AvgResponseTime)
	assert.Equal(t, 0, roiStats.ActiveAlerts)
	assert.Equal(t, 0, roiStats.OpenWorkOrders)
	assert.Equal(t, 0, roiStats.ResolvedIssues)

	deviceRepo.AssertExpectations(t)
}

func TestReportService_GetROIStats_ZeroDevices(t *testing.T) {
	svc, _, _, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("Count", mock.Anything).Return(0, nil)

	roiStats, err := svc.GetROIStats(ctx)

	require.NoError(t, err)
	require.NotNil(t, roiStats)
	assert.Equal(t, 0, roiStats.TotalDevices)
	assert.Equal(t, 0.0, roiStats.PredictedSavings)
}

func TestReportService_GetROIStats_CountError(t *testing.T) {
	svc, _, _, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	// Count 返回错误，源码用 _ 忽略错误，deviceCount 为 0
	deviceRepo.On("Count", mock.Anything).Return(0, errors.New("db error"))

	roiStats, err := svc.GetROIStats(ctx)

	// 源码中 Count 错误被忽略，不返回 error
	require.NoError(t, err)
	require.NotNil(t, roiStats)
	assert.Equal(t, 0, roiStats.TotalDevices)
}

func TestReportService_GetROIStats_ManyDevices(t *testing.T) {
	svc, _, _, deviceRepo, _, _ := newMockReportService()
	ctx := context.Background()

	deviceRepo.On("Count", mock.Anything).Return(100, nil)

	roiStats, err := svc.GetROIStats(ctx)

	require.NoError(t, err)
	require.NotNil(t, roiStats)
	assert.Equal(t, 100, roiStats.TotalDevices)
	assert.Equal(t, 500000.0, roiStats.PredictedSavings) // 100 * 5000
}
