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

