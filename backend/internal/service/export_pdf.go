package service

import (
	"bytes"
	"fmt"
	"strings"
)

// exportPDF exports data to PDF format
func (s *ExportService) exportPDF(data interface{}, reportType string, filename string) (*ExportResult, error) {
	var buf bytes.Buffer

	switch reportType {
	case "devices":
		buf.WriteString(s.generateDevicePDFContent(data.(*DeviceReportData)))
	case "alerts":
		buf.WriteString(s.generateAlertPDFContent(data.(*AlertReportData)))
	case "roi":
		buf.WriteString(s.generateROIPDFContent(data.(*ROIReportData)))
	}

	result := &ExportResult{
		Data:     buf.Bytes(),
		Filename: filename + ".txt",
		MimeType: "text/plain",
		Size:     int64(buf.Len()),
	}

	return result, nil
}

func (s *ExportService) generateDevicePDFContent(data *DeviceReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          设备状态报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("【概览统计】\n")
	content.WriteString("----------------------------------------\n")
	fmt.Fprintf(&content, "总设备数: %d\n", data.Summary.TotalDevices)
	fmt.Fprintf(&content, "在线设备: %d\n", data.Summary.OnlineDevices)
	fmt.Fprintf(&content, "离线设备: %d\n", data.Summary.OfflineDevices)
	fmt.Fprintf(&content, "告警设备: %d\n", data.Summary.WarningDevices)
	fmt.Fprintf(&content, "故障设备: %d\n", data.Summary.FaultDevices)
	fmt.Fprintf(&content, "平均温度: %.2f°C\n", data.Summary.AvgTemperature)
	fmt.Fprintf(&content, "平均振动: %.2f mm/s\n\n", data.Summary.AvgVibration)

	content.WriteString("【设备列表】\n")
	content.WriteString("----------------------------------------\n")
	for _, d := range data.Devices {
		fmt.Fprintf(&content, "%-14s %-14s %-10s %-10s %s\n",
			d.ID, d.Name, d.Type, d.Location, d.Status)
	}

	if len(data.DeviceStats) > 0 {
		content.WriteString("\n【设备运行数据】\n")
		for _, s := range data.DeviceStats {
			fmt.Fprintf(&content, "%-14s %.2f°C %.2f mm/s\n",
				s.DeviceID, s.AvgTemperature, s.AvgVibration)
		}
	}

	content.WriteString("\n========================================\n")
	return content.String()
}

func (s *ExportService) generateAlertPDFContent(data *AlertReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          告警统计报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("【告警统计】\n")
	fmt.Fprintf(&content, "总告警数: %d\n", data.AlertStats.TotalAlerts)
	fmt.Fprintf(&content, "活跃告警: %d\n", data.AlertStats.ActiveAlerts)
	fmt.Fprintf(&content, "已解决: %d\n\n", data.AlertStats.ResolvedAlerts)

	content.WriteString("【严重等级分布】\n")
	fmt.Fprintf(&content, "严重: %d 高: %d 中: %d 低: %d\n\n",
		data.AlertStats.CriticalAlerts, data.AlertStats.HighAlerts,
		data.AlertStats.MediumAlerts, data.AlertStats.LowAlerts)

	if len(data.TopAlertRules) > 0 {
		content.WriteString("【高频告警规则】\n")
		for _, r := range data.TopAlertRules {
			fmt.Fprintf(&content, "%s: %d 次\n", r.RuleName, r.Count)
		}
	}

	content.WriteString("\n========================================\n")
	return content.String()
}

func (s *ExportService) generateROIPDFContent(data *ROIReportData) string {
	var content strings.Builder

	content.WriteString("========================================\n")
	content.WriteString("          ROI分析报告\n")
	content.WriteString("========================================\n\n")
	fmt.Fprintf(&content, "生成时间: %s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("【核心指标】\n")
	fmt.Fprintf(&content, "监控设备数: %d\n", data.ROIStats.TotalDevices)
	fmt.Fprintf(&content, "预测节省: ¥%.2f\n", data.ROIStats.PredictedSavings)
	fmt.Fprintf(&content, "设备可用率: %.2f%%\n\n", data.ROIStats.UptimePercentage)

	if len(data.MonthlyTrend) > 0 {
		content.WriteString("【月度趋势】\n")
		for _, m := range data.MonthlyTrend {
			fmt.Fprintf(&content, "%s: ¥%.2f %.2f%%\n",
				m.Month, m.TotalSavings, m.UptimePercent)
		}
	}

	content.WriteString("\n========================================\n")
	return content.String()
}