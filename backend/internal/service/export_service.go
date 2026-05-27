package service

import (
	"context"
	"fmt"
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
	deviceRepo    repository.DeviceRepositoryInterface
	telemetryRepo repository.TelemetryRepositoryInterface
	alertRepo     repository.AlertRepositoryInterface
	workOrderRepo repository.WorkOrderRepositoryInterface
	reportSvc     *ReportService
}

// NewExportService creates a new export service
func NewExportService(
	deviceRepo repository.DeviceRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	alertRepo repository.AlertRepositoryInterface,
	workOrderRepo repository.WorkOrderRepositoryInterface,
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

	// Get device stats - Performance optimization: batch query replaces N+1 problem
	// 使用批量查询替代循环内单查询，将 N 次数据库查询减少为 1 次
	deviceStats := []model.DeviceStats{}
	if len(devices) > 0 {
		// 收集所有设备ID用于批量查询
		deviceIDs := make([]string, len(devices))
		for i, d := range devices {
			deviceIDs[i] = d.ID
		}

		// 执行单次批量查询获取所有设备的统计数据
		statsMap, err := s.telemetryRepo.GetStatsBatch(ctx, deviceIDs, req.StartDate, req.EndDate)
		if err == nil {
			// 从批量查询结果中提取数据
			for _, d := range devices {
				if stats, ok := statsMap[d.ID]; ok {
					deviceStats = append(deviceStats, *stats)
					summary.AvgTemperature += stats.AvgTemperature
					summary.AvgVibration += stats.AvgVibration
				}
			}
		} else {
			logger.L().Warn("Failed to get device stats batch",
				zap.Error(err),
			)
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

// generateROIReportData generates ROI report data from real data
func (s *ExportService) generateROIReportData(ctx context.Context) *ROIReportData {
	roiStats, err := s.reportSvc.GetROIStats(ctx)
	if err != nil {
		logger.L().Warn("Failed to get ROI stats for report",
			zap.Error(err),
		)
		roiStats = &model.ROIStats{}
	}

	// Calculate date range for metrics (last 30 days)
	now := time.Now()
	startDate := now.AddDate(0, 0, -30)

	// Generate device metrics from real data
	deviceMetrics := []DeviceMetric{}
	devices, _, err := s.deviceRepo.List(ctx, 1, 100) // Get more devices for ROI report
	if err != nil {
		logger.L().Warn("Failed to get devices for ROI report",
			zap.Error(err),
		)
		devices = []model.Device{}
	}

	if len(devices) > 0 {
		// Batch query telemetry stats for all devices
		deviceIDs := make([]string, len(devices))
		for i, d := range devices {
			deviceIDs[i] = d.ID
		}

		statsMap, err := s.telemetryRepo.GetStatsBatch(ctx, deviceIDs, startDate, now)
		if err != nil {
			logger.L().Warn("Failed to get telemetry stats batch for ROI report",
				zap.Error(err),
			)
			statsMap = make(map[string]*model.DeviceStats)
		}

		// Get work orders for each device
		for _, d := range devices {
			// Get telemetry stats
			var uptimeHours float64
			var uptimePercent float64
			if stats, ok := statsMap[d.ID]; ok && stats.DataPoints > 0 {
				// Calculate uptime based on data points (assuming 1 point per minute)
				// 30 days = 43200 minutes, each data point = 1 minute of uptime
				uptimeHours = float64(stats.DataPoints) / 60.0
				uptimePercent = (float64(stats.DataPoints) / 43200.0) * 100
				if uptimePercent > 100 {
					uptimePercent = 100
				}
			}

			// Get work orders for this device
			workOrders, _, err := s.workOrderRepo.List(ctx, "", d.ID, 1, 1000)
			if err != nil {
				logger.L().Warn("Failed to get work orders for device",
					zap.Error(err),
					zap.String("device_id", d.ID),
				)
				workOrders = []model.WorkOrder{}
			}

			// Count fault (urgent/high priority) and maintenance work orders
			var faultCount, maintenanceCount int
			for _, wo := range workOrders {
				// Filter work orders within the time range
				if wo.CreatedAt.Before(startDate) || wo.CreatedAt.After(now) {
					continue
				}
				if wo.Priority == "urgent" || wo.Priority == "high" {
					faultCount++
				} else {
					maintenanceCount++
				}
			}

			// Calculate savings based on uptime and reduced downtime
			// Base savings: $50/hour of uptime, minus cost of faults ($500 each) and maintenance ($100 each)
			savings := uptimeHours*50.0 - float64(faultCount)*500.0 - float64(maintenanceCount)*100.0
			if savings < 0 {
				savings = 0
			}

			deviceMetrics = append(deviceMetrics, DeviceMetric{
				DeviceID:         d.ID,
				DeviceName:       d.Name,
				UptimeHours:      uptimeHours,
				FaultCount:       faultCount,
				MaintenanceCount: maintenanceCount,
				Savings:          savings,
			})
		}
	}

	// Generate monthly trend from real data (last 6 months)
	monthlyTrend := []MonthlyMetric{}
	for i := 5; i >= 0; i-- {
		monthStart := time.Now().AddDate(0, -i, 0)
		monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		monthStr := monthStart.Format("2006-01")

		// Get device stats for this month
		monthDevices, _, err := s.deviceRepo.List(ctx, 1, 100)
		if err != nil {
			logger.L().Warn("Failed to get devices for monthly trend",
				zap.Error(err),
				zap.String("month", monthStr),
			)
			continue
		}

		var totalSavings float64
		var totalUptimePercent float64
		var validDevices int

		if len(monthDevices) > 0 {
			deviceIDs := make([]string, len(monthDevices))
			for i, d := range monthDevices {
				deviceIDs[i] = d.ID
			}

			statsMap, err := s.telemetryRepo.GetStatsBatch(ctx, deviceIDs, monthStart, monthEnd)
			if err != nil {
				logger.L().Warn("Failed to get telemetry stats for monthly trend",
					zap.Error(err),
					zap.String("month", monthStr),
				)
				statsMap = make(map[string]*model.DeviceStats)
			}

			// Calculate uptime and savings for each device
			daysInMonth := monthEnd.Sub(monthStart).Hours() / 24.0
			minutesInMonth := daysInMonth * 24.0 * 60.0

			for _, d := range monthDevices {
				if stats, ok := statsMap[d.ID]; ok && stats.DataPoints > 0 {
					validDevices++
					uptimePercent := (float64(stats.DataPoints) / minutesInMonth) * 100
					if uptimePercent > 100 {
						uptimePercent = 100
					}
					totalUptimePercent += uptimePercent

					uptimeHours := float64(stats.DataPoints) / 60.0
					savings := uptimeHours * 50.0 // $50/hour of uptime
					totalSavings += savings
				}
			}
		}

		// Get alert count for this month
		alerts, _, err := s.alertRepo.List(ctx, "", 1, 10000)
		if err != nil {
			logger.L().Warn("Failed to get alerts for monthly trend",
				zap.Error(err),
				zap.String("month", monthStr),
			)
			alerts = []model.Alert{}
		}

		var alertCount int
		for _, a := range alerts {
			if a.TriggeredAt.After(monthStart) && a.TriggeredAt.Before(monthEnd) {
				alertCount++
			}
		}

		avgUptimePercent := 0.0
		if validDevices > 0 {
			avgUptimePercent = totalUptimePercent / float64(validDevices)
		}

		monthlyTrend = append(monthlyTrend, MonthlyMetric{
			Month:         monthStr,
			TotalSavings:  totalSavings,
			UptimePercent: avgUptimePercent,
			AlertCount:    alertCount,
		})
	}

	return &ROIReportData{
		GeneratedAt:   time.Now(),
		ROIStats:      *roiStats,
		DeviceMetrics: deviceMetrics,
		MonthlyTrend:  monthlyTrend,
	}
}

