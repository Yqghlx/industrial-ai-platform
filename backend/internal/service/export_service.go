package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/errors"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// ExportServiceInterface defines the interface for export service
type ExportServiceInterface interface {
	Export(ctx context.Context, req *ExportRequest) (*ExportResult, error)
}

// ExportService handles report export functionality
type ExportService struct {
	deviceRepo    *repository.DeviceRepository
	telemetryRepo *repository.TelemetryRepository
	alertRepo     *repository.AlertRepository
	workOrderRepo *repository.WorkOrderRepository
	reportSvc     *ReportService
}

// NewExportService creates a new export service
func NewExportService(
	deviceRepo *repository.DeviceRepository,
	telemetryRepo *repository.TelemetryRepository,
	alertRepo *repository.AlertRepository,
	workOrderRepo *repository.WorkOrderRepository,
	reportSvc *ReportService,
) *ExportService {
	return &ExportService{
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		alertRepo:     alertRepo,
		workOrderRepo: workOrderRepo,
		reportSvc:     reportSvc,
	}
}

// ExportFormat represents the export format type
type ExportFormat string

const (
	FormatPDF  ExportFormat = "pdf"
	FormatXLSX ExportFormat = "xlsx"
)

// ExportRequest represents an export request
type ExportRequest struct {
	ReportType string // "devices", "alerts", "roi"
	Format     ExportFormat
	StartDate  time.Time
	EndDate    time.Time
	DeviceID   string // Optional device filter
}

// ExportResult represents the export result
type ExportResult struct {
	Data     []byte
	Filename string
	MimeType string
	Size     int64
}

// Export exports a report in the specified format
func (s *ExportService) Export(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	// Set default date range if not provided
	if req.StartDate.IsZero() {
		req.StartDate = time.Now().Add(-24 * time.Hour)
	}
	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	// Generate report data based on type
	var data interface{}
	var filename string

	switch req.ReportType {
	case "devices":
		data = s.generateDeviceReportData(ctx, req)
		filename = fmt.Sprintf("设备状态报告_%s", time.Now().Format("20060102"))
	case "alerts":
		data = s.generateAlertReportData(ctx, req)
		filename = fmt.Sprintf("告警统计报告_%s", time.Now().Format("20060102"))
	case "roi":
		data = s.generateROIReportData(ctx)
		filename = fmt.Sprintf("ROI分析报告_%s", time.Now().Format("20060102"))
	default:
		return nil, errors.NewAppError(errors.ErrCodeInvalidInput, "Unsupported report type", req.ReportType)
	}

	// Export based on format
	switch req.Format {
	case FormatPDF:
		return s.exportPDF(data, req.ReportType, filename)
	case FormatXLSX:
		return s.exportXLSX(data, req.ReportType, filename)
	default:
		return nil, errors.NewAppError(errors.ErrCodeInvalidInput, "Unsupported export format", string(req.Format))
	}
}

// DeviceReportData represents device report data
type DeviceReportData struct {
	GeneratedAt time.Time
	Devices     []model.Device
	DeviceStats []model.DeviceStats
	Summary     DeviceSummary
}

type DeviceSummary struct {
	TotalDevices   int
	OnlineDevices  int
	OfflineDevices int
	WarningDevices int
	FaultDevices   int
	AvgTemperature float64
	AvgVibration   float64
}

// AlertReportData represents alert report data
type AlertReportData struct {
	GeneratedAt   time.Time
	Alerts        []model.Alert
	AlertStats    AlertStats
	TopAlertRules []AlertRuleCount
}

type AlertStats struct {
	TotalAlerts    int
	CriticalAlerts int
	HighAlerts     int
	MediumAlerts   int
	LowAlerts      int
	ActiveAlerts   int
	ResolvedAlerts int
}

type AlertRuleCount struct {
	RuleName string
	Count    int
}

// ROIReportData represents ROI report data
type ROIReportData struct {
	GeneratedAt   time.Time
	ROIStats      model.ROIStats
	DeviceMetrics []DeviceMetric
	MonthlyTrend  []MonthlyMetric
}

type DeviceMetric struct {
	DeviceID         string
	DeviceName       string
	UptimeHours      float64
	FaultCount       int
	MaintenanceCount int
	Savings          float64
}

type MonthlyMetric struct {
	Month         string
	TotalSavings  float64
	UptimePercent float64
	AlertCount    int
}

// generateDeviceReportData generates device report data
func (s *ExportService) generateDeviceReportData(ctx context.Context, req *ExportRequest) *DeviceReportData {
	devices, total, err := s.deviceRepo.List(ctx, 1, 100)
	if err != nil {
		logger.L().Warn("Failed to get devices for report",
			zap.Error(err),
		)
		devices = []model.Device{}
	}

	// Calculate summary
	summary := DeviceSummary{
		TotalDevices: total,
	}

	for _, d := range devices {
		switch d.Status {
		case "online":
			summary.OnlineDevices++
		case "offline":
			summary.OfflineDevices++
		case "warning":
			summary.WarningDevices++
		case "fault":
			summary.FaultDevices++
		}
	}

	// Get device stats
	deviceStats := []model.DeviceStats{}
	for _, d := range devices {
		stats, err := s.telemetryRepo.GetStats(ctx, d.ID, req.StartDate, req.EndDate)
		if err == nil {
			deviceStats = append(deviceStats, *stats)
			summary.AvgTemperature += stats.AvgTemperature
			summary.AvgVibration += stats.AvgVibration
		}
	}

	if len(deviceStats) > 0 {
		summary.AvgTemperature /= float64(len(deviceStats))
		summary.AvgVibration /= float64(len(deviceStats))
	}

	return &DeviceReportData{
		GeneratedAt: time.Now(),
		Devices:     devices,
		DeviceStats: deviceStats,
		Summary:     summary,
	}
}

// generateAlertReportData generates alert report data
func (s *ExportService) generateAlertReportData(ctx context.Context, req *ExportRequest) *AlertReportData {
	alerts, total, err := s.alertRepo.List(ctx, "", 1, 100)
	if err != nil {
		logger.L().Warn("Failed to get alerts for report",
			zap.Error(err),
		)
		alerts = []model.Alert{}
	}

	// Calculate stats
	stats := AlertStats{
		TotalAlerts: total,
	}

	ruleCountMap := make(map[string]int)

	for _, a := range alerts {
		switch a.Severity {
		case "critical":
			stats.CriticalAlerts++
		case "high":
			stats.HighAlerts++
		case "medium":
			stats.MediumAlerts++
		case "low":
			stats.LowAlerts++
		}

		switch a.Status {
		case "active":
			stats.ActiveAlerts++
		case "resolved":
			stats.ResolvedAlerts++
		}

		// Count by rule (would need rule name, using rule_id for now)
		ruleCountMap[fmt.Sprintf("规则%d", a.RuleID)]++
	}

	// Convert to top rules
	topRules := []AlertRuleCount{}
	for name, count := range ruleCountMap {
		topRules = append(topRules, AlertRuleCount{
			RuleName: name,
			Count:    count,
		})
	}

	// Sort by count (simple sort)
	for i := 0; i < len(topRules); i++ {
		for j := i + 1; j < len(topRules); j++ {
			if topRules[j].Count > topRules[i].Count {
				topRules[i], topRules[j] = topRules[j], topRules[i]
			}
		}
	}

	// Limit to top 10
	if len(topRules) > 10 {
		topRules = topRules[:10]
	}

	return &AlertReportData{
		GeneratedAt:   time.Now(),
		Alerts:        alerts,
		AlertStats:    stats,
		TopAlertRules: topRules,
	}
}

// generateROIReportData generates ROI report data
func (s *ExportService) generateROIReportData(ctx context.Context) *ROIReportData {
	roiStats, err := s.reportSvc.GetROIStats(ctx)
	if err != nil {
		logger.L().Warn("Failed to get ROI stats for report",
			zap.Error(err),
		)
		roiStats = &model.ROIStats{}
	}

	// Generate device metrics (mock data for demonstration)
	deviceMetrics := []DeviceMetric{}
	devices, _, err := s.deviceRepo.List(ctx, 1, 10)
	if err == nil {
		for _, d := range devices {
			deviceMetrics = append(deviceMetrics, DeviceMetric{
				DeviceID:         d.ID,
				DeviceName:       d.Name,
				UptimeHours:      720.0 * roiStats.UptimePercentage / 100, // Monthly hours
				FaultCount:       2,
				MaintenanceCount: 5,
				Savings:          5000.0,
			})
		}
	}

	// Generate monthly trend (mock data)
	monthlyTrend := []MonthlyMetric{}
	for i := 5; i >= 0; i-- {
		month := time.Now().AddDate(0, -i, 0).Format("2006-01")
		monthlyTrend = append(monthlyTrend, MonthlyMetric{
			Month:         month,
			TotalSavings:  roiStats.PredictedSavings * float64(6-i+1) / 6,
			UptimePercent: 95 + float64(i)*0.5,
			AlertCount:    10 + i*2,
		})
	}

	return &ROIReportData{
		GeneratedAt:   time.Now(),
		ROIStats:      *roiStats,
		DeviceMetrics: deviceMetrics,
		MonthlyTrend:  monthlyTrend,
	}
}

// exportPDF exports data to PDF format
func (s *ExportService) exportPDF(data interface{}, reportType string, filename string) (*ExportResult, error) {
	// PDF generation using a simple text-based approach
	// For full PDF support, use go-pdf/fpdf library
	var buf bytes.Buffer

	// Generate PDF content
	switch reportType {
	case "devices":
		buf.WriteString(s.generateDevicePDFContent(data.(*DeviceReportData)))
	case "alerts":
		buf.WriteString(s.generateAlertPDFContent(data.(*AlertReportData)))
	case "roi":
		buf.WriteString(s.generateROIPDFContent(data.(*ROIReportData)))
	}

	// For now, generate a formatted text report that can be converted to PDF
	// In production, use go-pdf/fpdf for proper PDF generation
	result := &ExportResult{
		Data:     buf.Bytes(),
		Filename: filename + ".txt", // Change to .pdf when using proper PDF library
		MimeType: "text/plain",      // Change to "application/pdf" when using proper PDF library
		Size:     int64(buf.Len()),
	}

	return result, nil
}

// generateDevicePDFContent generates device report PDF content
func (s *ExportService) generateDevicePDFContent(data *DeviceReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          设备状态报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Summary section
	content.WriteString("【概览统计】\n")
	content.WriteString("----------------------------------------\n")
	fmt.Fprintf(&content, "总设备数: %d\n", data.Summary.TotalDevices)
	fmt.Fprintf(&content, "在线设备: %d\n", data.Summary.OnlineDevices)
	fmt.Fprintf(&content, "离线设备: %d\n", data.Summary.OfflineDevices)
	fmt.Fprintf(&content, "告警设备: %d\n", data.Summary.WarningDevices)
	fmt.Fprintf(&content, "故障设备: %d\n", data.Summary.FaultDevices)
	fmt.Fprintf(&content, "平均温度: %.2f°C\n", data.Summary.AvgTemperature)
	fmt.Fprintf(&content, "平均振动: %.2f mm/s\n\n", data.Summary.AvgVibration)

	// Device list
	content.WriteString("【设备列表】\n")
	content.WriteString("----------------------------------------\n")
	content.WriteString("设备ID         名称           类型        位置        状态\n")
	content.WriteString("----------------------------------------\n")

	for _, d := range data.Devices {
		fmt.Fprintf(&content, "%-14s %-14s %-10s %-10s %s\n",
			d.ID, d.Name, d.Type, d.Location, d.Status)
	}
	content.WriteString("\n")

	// Device statistics
	if len(data.DeviceStats) > 0 {
		content.WriteString("【设备运行数据】\n")
		content.WriteString("----------------------------------------\n")
		content.WriteString("设备ID         平均温度   最高温度   平均振动   最高振动   数据点\n")
		content.WriteString("----------------------------------------\n")

		for _, s := range data.DeviceStats {
			fmt.Fprintf(&content, "%-14s %.2f°C   %.2f°C   %.2f     %.2f     %d\n",
				s.DeviceID, s.AvgTemperature, s.MaxTemperature,
				s.AvgVibration, s.MaxVibration, s.DataPoints)
		}
	}

	content.WriteString("\n========================================\n")
	content.WriteString("              报告结束\n")
	content.WriteString("========================================\n")

	return content.String()
}

// generateAlertPDFContent generates alert report PDF content
func (s *ExportService) generateAlertPDFContent(data *AlertReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          告警统计报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Statistics
	content.WriteString("【告警统计】\n")
	content.WriteString("----------------------------------------\n")
	fmt.Fprintf(&content, "总告警数: %d\n", data.AlertStats.TotalAlerts)
	fmt.Fprintf(&content, "活跃告警: %d\n", data.AlertStats.ActiveAlerts)
	fmt.Fprintf(&content, "已解决: %d\n", data.AlertStats.ResolvedAlerts)
	content.WriteString("\n")

	content.WriteString("【严重等级分布】\n")
	content.WriteString("----------------------------------------\n")
	fmt.Fprintf(&content, "严重(Critical): %d\n", data.AlertStats.CriticalAlerts)
	fmt.Fprintf(&content, "高(High): %d\n", data.AlertStats.HighAlerts)
	fmt.Fprintf(&content, "中(Medium): %d\n", data.AlertStats.MediumAlerts)
	fmt.Fprintf(&content, "低(Low): %d\n\n", data.AlertStats.LowAlerts)

	// Top alert rules
	if len(data.TopAlertRules) > 0 {
		content.WriteString("【高频告警规则 TOP 10】\n")
		content.WriteString("----------------------------------------\n")
		for _, r := range data.TopAlertRules {
			fmt.Fprintf(&content, "%-20s: %d 次\n", r.RuleName, r.Count)
		}
		content.WriteString("\n")
	}

	// Alert list
	content.WriteString("【告警列表】\n")
	content.WriteString("----------------------------------------\n")
	content.WriteString("ID    设备ID         严重等级   状态      触发时间              消息\n")
	content.WriteString("----------------------------------------\n")

	for _, a := range data.Alerts {
		fmt.Fprintf(&content, "%-5d %-14s %-10s %-8s %s %s\n",
			a.ID, a.DeviceID, a.Severity, a.Status,
			a.TriggeredAt.Format("2006-01-02 15:04"), truncate(a.Message, 30))
	}

	content.WriteString("\n========================================\n")
	content.WriteString("              报告结束\n")
	content.WriteString("========================================\n")

	return content.String()
}

// generateROIPDFContent generates ROI report PDF content
func (s *ExportService) generateROIPDFContent(data *ROIReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          ROI分析报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// ROI Statistics
	content.WriteString("【核心指标】\n")
	content.WriteString("----------------------------------------\n")
	fmt.Fprintf(&content, "监控设备数: %d\n", data.ROIStats.TotalDevices)
	fmt.Fprintf(&content, "活跃告警: %d\n", data.ROIStats.ActiveAlerts)
	fmt.Fprintf(&content, "待处理工单: %d\n", data.ROIStats.OpenWorkOrders)
	fmt.Fprintf(&content, "已解决问题: %d\n", data.ROIStats.ResolvedIssues)
	fmt.Fprintf(&content, "预测节省金额: ¥%.2f\n", data.ROIStats.PredictedSavings)
	fmt.Fprintf(&content, "设备可用率: %.2f%%\n", data.ROIStats.UptimePercentage)
	fmt.Fprintf(&content, "平均响应时间: %.2f 小时\n\n", data.ROIStats.AvgResponseTime)

	// Monthly trend
	content.WriteString("【月度趋势】\n")
	content.WriteString("----------------------------------------\n")
	content.WriteString("月份      节省金额    可用率    告警数\n")
	content.WriteString("----------------------------------------\n")

	for _, m := range data.MonthlyTrend {
		fmt.Fprintf(&content, "%-10s ¥%.2f    %.2f%%    %d\n",
			m.Month, m.TotalSavings, m.UptimePercent, m.AlertCount)
	}
	content.WriteString("\n")

	// Device metrics
	if len(data.DeviceMetrics) > 0 {
		content.WriteString("【设备绩效】\n")
		content.WriteString("----------------------------------------\n")
		content.WriteString("设备ID         运行时间    故障次数   维护次数   节省金额\n")
		content.WriteString("----------------------------------------\n")

		for _, d := range data.DeviceMetrics {
			fmt.Fprintf(&content, "%-14s %.2f h    %d         %d         ¥%.2f\n",
				d.DeviceID, d.UptimeHours, d.FaultCount, d.MaintenanceCount, d.Savings)
		}
	}

	content.WriteString("\n========================================\n")
	content.WriteString("              报告结束\n")
	content.WriteString("========================================\n")

	return content.String()
}

// exportXLSX exports data to Excel format
func (s *ExportService) exportXLSX(data interface{}, reportType string, filename string) (*ExportResult, error) {
	// For full Excel support, use excelize library
	// This implementation provides a CSV-like format that Excel can open
	var buf bytes.Buffer

	switch reportType {
	case "devices":
		buf.WriteString(s.generateDeviceExcelContent(data.(*DeviceReportData)))
	case "alerts":
		buf.WriteString(s.generateAlertExcelContent(data.(*AlertReportData)))
	case "roi":
		buf.WriteString(s.generateROIExcelContent(data.(*ROIReportData)))
	}

	result := &ExportResult{
		Data:     buf.Bytes(),
		Filename: filename + ".csv", // Change to .xlsx when using excelize
		MimeType: "text/csv",        // Change to "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		Size:     int64(buf.Len()),
	}

	return result, nil
}

// generateDeviceExcelContent generates device Excel content (CSV format)
func (s *ExportService) generateDeviceExcelContent(data *DeviceReportData) string {
	var content strings.Builder

	// Header
	content.WriteString("设备状态报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Summary
	content.WriteString("概览统计\n")
	fmt.Fprintf(&content, "总设备数,%d\n", data.Summary.TotalDevices)
	fmt.Fprintf(&content, "在线设备,%d\n", data.Summary.OnlineDevices)
	fmt.Fprintf(&content, "离线设备,%d\n", data.Summary.OfflineDevices)
	fmt.Fprintf(&content, "告警设备,%d\n", data.Summary.WarningDevices)
	fmt.Fprintf(&content, "故障设备,%d\n", data.Summary.FaultDevices)
	fmt.Fprintf(&content, "平均温度,%.2f\n", data.Summary.AvgTemperature)
	fmt.Fprintf(&content, "平均振动,%.2f\n\n", data.Summary.AvgVibration)

	// Device list
	content.WriteString("设备列表\n")
	content.WriteString("设备ID,名称,类型,位置,状态\n")
	for _, d := range data.Devices {
		fmt.Fprintf(&content, "%s,%s,%s,%s,%s\n",
			d.ID, d.Name, d.Type, d.Location, d.Status)
	}
	content.WriteString("\n")

	// Device statistics
	if len(data.DeviceStats) > 0 {
		content.WriteString("设备运行数据\n")
		content.WriteString("设备ID,平均温度,最高温度,平均振动,最高振动,数据点数\n")
		for _, s := range data.DeviceStats {
			fmt.Fprintf(&content, "%s,%.2f,%.2f,%.2f,%.2f,%d\n",
				s.DeviceID, s.AvgTemperature, s.MaxTemperature,
				s.AvgVibration, s.MaxVibration, s.DataPoints)
		}
	}

	return content.String()
}

// generateAlertExcelContent generates alert Excel content (CSV format)
func (s *ExportService) generateAlertExcelContent(data *AlertReportData) string {
	var content strings.Builder

	// Header
	content.WriteString("告警统计报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Statistics
	content.WriteString("告警统计\n")
	fmt.Fprintf(&content, "总告警数,%d\n", data.AlertStats.TotalAlerts)
	fmt.Fprintf(&content, "活跃告警,%d\n", data.AlertStats.ActiveAlerts)
	fmt.Fprintf(&content, "已解决,%d\n\n", data.AlertStats.ResolvedAlerts)

	// Severity distribution
	content.WriteString("严重等级分布\n")
	content.WriteString("等级,数量\n")
	fmt.Fprintf(&content, "严重(Critical),%d\n", data.AlertStats.CriticalAlerts)
	fmt.Fprintf(&content, "高(High),%d\n", data.AlertStats.HighAlerts)
	fmt.Fprintf(&content, "中(Medium),%d\n", data.AlertStats.MediumAlerts)
	fmt.Fprintf(&content, "低(Low),%d\n\n", data.AlertStats.LowAlerts)

	// Top alert rules
	if len(data.TopAlertRules) > 0 {
		content.WriteString("高频告警规则 TOP 10\n")
		content.WriteString("规则名称,触发次数\n")
		for _, r := range data.TopAlertRules {
			fmt.Fprintf(&content, "%s,%d\n", r.RuleName, r.Count)
		}
		content.WriteString("\n")
	}

	// Alert list
	content.WriteString("告警列表\n")
	content.WriteString("ID,设备ID,严重等级,状态,触发时间,消息\n")
	for _, a := range data.Alerts {
		fmt.Fprintf(&content, "%d,%s,%s,%s,%s,%s\n",
			a.ID, a.DeviceID, a.Severity, a.Status,
			a.TriggeredAt.Format("2006-01-02 15:04:05"), a.Message)
	}

	return content.String()
}

// generateROIExcelContent generates ROI Excel content (CSV format)
func (s *ExportService) generateROIExcelContent(data *ROIReportData) string {
	var content strings.Builder

	// Header
	content.WriteString("ROI分析报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	// ROI Statistics
	content.WriteString("核心指标\n")
	content.WriteString("指标,数值\n")
	fmt.Fprintf(&content, "监控设备数,%d\n", data.ROIStats.TotalDevices)
	fmt.Fprintf(&content, "活跃告警,%d\n", data.ROIStats.ActiveAlerts)
	fmt.Fprintf(&content, "待处理工单,%d\n", data.ROIStats.OpenWorkOrders)
	fmt.Fprintf(&content, "已解决问题,%d\n", data.ROIStats.ResolvedIssues)
	fmt.Fprintf(&content, "预测节省金额,%.2f\n", data.ROIStats.PredictedSavings)
	fmt.Fprintf(&content, "设备可用率,%.2f%%\n", data.ROIStats.UptimePercentage)
	fmt.Fprintf(&content, "平均响应时间,%.2f\n\n", data.ROIStats.AvgResponseTime)

	// Monthly trend
	content.WriteString("月度趋势\n")
	content.WriteString("月份,节省金额,可用率,告警数\n")
	for _, m := range data.MonthlyTrend {
		fmt.Fprintf(&content, "%s,%.2f,%.2f,%d\n",
			m.Month, m.TotalSavings, m.UptimePercent, m.AlertCount)
	}
	content.WriteString("\n")

	// Device metrics
	if len(data.DeviceMetrics) > 0 {
		content.WriteString("设备绩效\n")
		content.WriteString("设备ID,设备名称,运行时间,故障次数,维护次数,节省金额\n")
		for _, d := range data.DeviceMetrics {
			fmt.Fprintf(&content, "%s,%s,%.2f,%d,%d,%.2f\n",
				d.DeviceID, d.DeviceName, d.UptimeHours, d.FaultCount, d.MaintenanceCount, d.Savings)
		}
	}

	return content.String()
}

// truncate truncates a string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetExportFilename returns the filename for export
func GetExportFilename(reportType string, format ExportFormat) string {
	prefix := ""
	switch reportType {
	case "devices":
		prefix = "设备状态报告"
	case "alerts":
		prefix = "告警统计报告"
	case "roi":
		prefix = "ROI分析报告"
	default:
		prefix = "报告"
	}

	ext := ""
	switch format {
	case FormatPDF:
		ext = ".pdf"
	case FormatXLSX:
		ext = ".xlsx"
	}

	return fmt.Sprintf("%s_%s%s", prefix, time.Now().Format("20060102"), ext)
}

// GetMimeType returns the MIME type for export format
func GetMimeType(format ExportFormat) string {
	switch format {
	case FormatPDF:
		return "application/pdf"
	case FormatXLSX:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
