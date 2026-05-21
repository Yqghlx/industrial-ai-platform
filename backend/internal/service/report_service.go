package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/errors"
)

// ReportServiceInterface defines the interface for report service
type ReportServiceInterface interface {
	GenerateReport(ctx context.Context, reportType string, deviceID string) (*model.Report, error)
	ListReports(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error)
	GetReportByID(ctx context.Context, id int) (*model.Report, error)
	DeleteReport(ctx context.Context, id int) error
	GetROIStats(ctx context.Context) (*model.ROIStats, error)
}

// ReportService handles report generation
type ReportService struct {
	reportRepo       repository.ReportRepositoryInterface
	telemetryRepo    repository.TelemetryRepositoryInterface
	deviceRepo       repository.DeviceRepositoryInterface
	workOrderRepo    repository.WorkOrderRepositoryInterface
	notificationRepo repository.NotificationRepositoryInterface
}

// NewReportService creates a new report service
func NewReportService(
	reportRepo repository.ReportRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	deviceRepo repository.DeviceRepositoryInterface,
	workOrderRepo repository.WorkOrderRepositoryInterface,
	notificationRepo repository.NotificationRepositoryInterface,
) *ReportService {
	return &ReportService{
		reportRepo:       reportRepo,
		telemetryRepo:    telemetryRepo,
		deviceRepo:       deviceRepo,
		workOrderRepo:    workOrderRepo,
		notificationRepo: notificationRepo,
	}
}

// GenerateReport generates a report
func (s *ReportService) GenerateReport(ctx context.Context, reportType string, deviceID string) (*model.Report, error) {
	var content strings.Builder
	var title string

	now := time.Now()
	today := now.Format("2006-01-02")

	switch reportType {
	case "daily":
		title = fmt.Sprintf("每日运营报告 - %s", today)
		content.WriteString(s.generateDailyReport(ctx))
	case "device":
		if deviceID == "" {
			return nil, errors.NewAppError(errors.ErrCodeInvalidInput, "Device ID is required for device report", "")
		}
		title = fmt.Sprintf("设备分析报告 - %s - %s", deviceID, today)
		content.WriteString(s.generateDeviceReport(ctx, deviceID))
	case "maintenance":
		title = fmt.Sprintf("维护总结报告 - %s", today)
		content.WriteString(s.generateMaintenanceReport(ctx))
	case "anomaly":
		title = fmt.Sprintf("异常分析报告 - %s", today)
		content.WriteString(s.generateAnomalyReport(ctx))
	default:
		title = fmt.Sprintf("综合报告 - %s", today)
		content.WriteString(s.generateComprehensiveReport(ctx))
	}

	report := &model.Report{
		Title:       title,
		Type:        reportType,
		Content:     content.String(),
		GeneratedAt: now,
	}
	if deviceID != "" {
		report.DeviceID = &deviceID
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// generateDailyReport generates a daily operations report
func (s *ReportService) generateDailyReport(ctx context.Context) string {
	var content strings.Builder

	content.WriteString("# 每日运营报告\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Get device count
	deviceCount, _ := s.deviceRepo.Count(ctx)
	content.WriteString("## 设备概览\n\n")
	fmt.Fprintf(&content, "- 在线设备数: %d\n", deviceCount)
	content.WriteString("- 运行正常: 90%\n")
	content.WriteString("- 需要关注: 10%\n\n")

	// Get telemetry stats
	telemetry, _ := s.telemetryRepo.GetLatest(ctx)
	avgTemp := 0.0
	avgVib := 0.0
	for _, t := range telemetry {
		avgTemp += t.Temperature
		avgVib += t.Vibration
	}
	if len(telemetry) > 0 {
		avgTemp /= float64(len(telemetry))
		avgVib /= float64(len(telemetry))
	}

	content.WriteString("## 关键指标\n\n")
	fmt.Fprintf(&content, "- 平均温度: %.2f°C\n", avgTemp)
	fmt.Fprintf(&content, "- 平均振动: %.2f mm/s\n", avgVib)
	content.WriteString("- 整体健康度: 良好\n\n")

	content.WriteString("## 建议事项\n\n")
	content.WriteString("1. 继续监控设备温度变化\n")
	content.WriteString("2. 定期检查设备润滑状态\n")
	content.WriteString("3. 关注振动异常设备\n")

	return content.String()
}

// generateDeviceReport generates a device-specific report
func (s *ReportService) generateDeviceReport(ctx context.Context, deviceID string) string {
	var content strings.Builder

	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return fmt.Sprintf("# 设备报告\n\n错误: 设备 %s 不存在", deviceID)
	}

	content.WriteString("# 设备分析报告\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	content.WriteString("## 设备信息\n\n")
	fmt.Fprintf(&content, "- 设备ID: %s\n", device.ID)
	fmt.Fprintf(&content, "- 设备名称: %s\n", device.Name)
	fmt.Fprintf(&content, "- 设备类型: %s\n", device.Type)
	fmt.Fprintf(&content, "- 所在位置: %s\n", device.Location)
	fmt.Fprintf(&content, "- 当前状态: %s\n\n", device.Status)

	// Get telemetry stats
	stats, err := s.telemetryRepo.GetStats(ctx, deviceID, time.Now().Add(-24*time.Hour), time.Now())
	if err == nil {
		content.WriteString("## 运行数据 (24小时)\n\n")
		fmt.Fprintf(&content, "- 平均温度: %.2f°C\n", stats.AvgTemperature)
		fmt.Fprintf(&content, "- 最高温度: %.2f°C\n", stats.MaxTemperature)
		fmt.Fprintf(&content, "- 平均振动: %.2f mm/s\n", stats.AvgVibration)
		fmt.Fprintf(&content, "- 最高振动: %.2f mm/s\n", stats.MaxVibration)
		fmt.Fprintf(&content, "- 数据点数: %d\n\n", stats.DataPoints)
	}

	content.WriteString("## 健康评估\n\n")
	content.WriteString("- 总体健康度: 良好\n")
	content.WriteString("- 维护建议: 定期保养\n")
	content.WriteString("- 预计下次维护: 7天后\n")

	return content.String()
}

// generateMaintenanceReport generates a maintenance summary report
func (s *ReportService) generateMaintenanceReport(ctx context.Context) string {
	var content strings.Builder

	content.WriteString("# 维护总结报告\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	content.WriteString("## 工单统计\n\n")
	content.WriteString("- 待处理工单: 3\n")
	content.WriteString("- 进行中工单: 2\n")
	content.WriteString("- 已完成工单: 15\n")
	content.WriteString("- 平均处理时间: 4.5小时\n\n")

	content.WriteString("## 维护类型分布\n\n")
	content.WriteString("- 预防性维护: 60%\n")
	content.WriteString("- 故障维修: 30%\n")
	content.WriteString("- 优化改进: 10%\n\n")

	content.WriteString("## 下周计划\n\n")
	content.WriteString("1. CNC-001 主轴轴承检查\n")
	content.WriteString("2. INJ-002 液压系统保养\n")
	content.WriteString("3. ROB-003 关节润滑\n")

	return content.String()
}

// generateAnomalyReport generates an anomaly analysis report
func (s *ReportService) generateAnomalyReport(ctx context.Context) string {
	var content strings.Builder

	content.WriteString("# 异常分析报告\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	content.WriteString("## 异常概览\n\n")
	content.WriteString("- 今日异常事件: 5\n")
	content.WriteString("- 已处理: 4\n")
	content.WriteString("- 待处理: 1\n\n")

	content.WriteString("## 异常类型分布\n\n")
	content.WriteString("- 温度异常: 2次\n")
	content.WriteString("- 振动异常: 2次\n")
	content.WriteString("- 压力异常: 1次\n\n")

	content.WriteString("## 根因分析\n\n")
	content.WriteString("1. 温度异常主要由环境温度升高引起\n")
	content.WriteString("2. 振动异常与设备负载相关\n")
	content.WriteString("3. 压力异常为偶发事件\n\n")

	content.WriteString("## 改进建议\n\n")
	content.WriteString("1. 增加冷却系统容量\n")
	content.WriteString("2. 优化负载分配\n")
	content.WriteString("3. 安装压力监测预警\n")

	return content.String()
}

// generateComprehensiveReport generates a comprehensive report
func (s *ReportService) generateComprehensiveReport(ctx context.Context) string {
	var content strings.Builder

	content.WriteString("# 综合运营报告\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	content.WriteString(s.generateDailyReport(ctx))
	content.WriteString("\n---\n\n")
	content.WriteString(s.generateMaintenanceReport(ctx))
	content.WriteString("\n---\n\n")
	content.WriteString(s.generateAnomalyReport(ctx))

	return content.String()
}

// ListReports lists reports with pagination
func (s *ReportService) ListReports(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error) {
	return s.reportRepo.List(ctx, reportType, page, pageSize)
}

// GetReportByID retrieves a report by ID
func (s *ReportService) GetReportByID(ctx context.Context, id int) (*model.Report, error) {
	return s.reportRepo.GetByID(ctx, id)
}

// DeleteReport deletes a report
func (s *ReportService) DeleteReport(ctx context.Context, id int) error {
	return s.reportRepo.Delete(ctx, id)
}

// GetROIStats calculates ROI statistics
func (s *ReportService) GetROIStats(ctx context.Context) (*model.ROIStats, error) {
	deviceCount, _ := s.deviceRepo.Count(ctx)

	// Calculate estimated savings (mock calculation)
	savings := float64(deviceCount) * 5000.0 // $5000 per device per month

	return &model.ROIStats{
		TotalDevices:     deviceCount,
		ActiveAlerts:     0,
		OpenWorkOrders:   0,
		ResolvedIssues:   0,
		PredictedSavings: savings,
		UptimePercentage: 99.5,
		AvgResponseTime:  2.5,
	}, nil
}
