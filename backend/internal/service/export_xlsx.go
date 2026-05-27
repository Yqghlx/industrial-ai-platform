package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// exportXLSX exports data to Excel format
func (s *ExportService) exportXLSX(data interface{}, reportType string, filename string) (*ExportResult, error) {
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
		Filename: filename + ".csv",
		MimeType: "text/csv",
		Size:     int64(buf.Len()),
	}

	return result, nil
}

func (s *ExportService) generateDeviceExcelContent(data *DeviceReportData) string {
	var content strings.Builder

	content.WriteString("设备状态报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("概览统计\n")
	fmt.Fprintf(&content, "总设备数,%d\n", data.Summary.TotalDevices)
	fmt.Fprintf(&content, "在线设备,%d\n", data.Summary.OnlineDevices)
	fmt.Fprintf(&content, "平均温度,%.2f\n\n", data.Summary.AvgTemperature)

	content.WriteString("设备列表\n")
	content.WriteString("设备ID,名称,类型,位置,状态\n")
	for _, d := range data.Devices {
		fmt.Fprintf(&content, "%s,%s,%s,%s,%s\n",
			d.ID, d.Name, d.Type, d.Location, d.Status)
	}

	return content.String()
}

func (s *ExportService) generateAlertExcelContent(data *AlertReportData) string {
	var content strings.Builder

	content.WriteString("告警统计报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("告警统计\n")
	fmt.Fprintf(&content, "总告警数,%d\n", data.AlertStats.TotalAlerts)
	fmt.Fprintf(&content, "活跃告警,%d\n\n", data.AlertStats.ActiveAlerts)

	content.WriteString("严重等级分布\n")
	content.WriteString("等级,数量\n")
	fmt.Fprintf(&content, "严重,%d\n", data.AlertStats.CriticalAlerts)
	fmt.Fprintf(&content, "高,%d\n", data.AlertStats.HighAlerts)
	fmt.Fprintf(&content, "中,%d\n", data.AlertStats.MediumAlerts)
	fmt.Fprintf(&content, "低,%d\n\n", data.AlertStats.LowAlerts)

	content.WriteString("告警列表\n")
	content.WriteString("ID,设备ID,严重等级,状态,消息\n")
	for _, a := range data.Alerts {
		fmt.Fprintf(&content, "%d,%s,%s,%s,%s\n",
			a.ID, a.DeviceID, a.Severity, a.Status, truncate(a.Message, 50))
	}

	return content.String()
}

func (s *ExportService) generateROIExcelContent(data *ROIReportData) string {
	var content strings.Builder

	content.WriteString("ROI分析报告\n")
	fmt.Fprintf(&content, "生成时间,%s\n\n", data.GeneratedAt.Format("2006-01-02 15:04:05"))

	content.WriteString("核心指标\n")
	fmt.Fprintf(&content, "监控设备数,%d\n", data.ROIStats.TotalDevices)
	fmt.Fprintf(&content, "预测节省金额,%.2f\n", data.ROIStats.PredictedSavings)
	fmt.Fprintf(&content, "设备可用率,%.2f%%\n\n", data.ROIStats.UptimePercentage)

	content.WriteString("月度趋势\n")
	content.WriteString("月份,节省金额,可用率\n")
	for _, m := range data.MonthlyTrend {
		fmt.Fprintf(&content, "%s,%.2f,%.2f%%\n",
			m.Month, m.TotalSavings, m.UptimePercent)
	}

	return content.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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
