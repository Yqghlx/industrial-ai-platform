package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/constants"
	"github.com/industrial-ai/platform/pkg/errors"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// TelemetryService handles telemetry data
type TelemetryService struct {
	telemetryRepo *repository.TelemetryRepository
	deviceRepo    *repository.DeviceRepository
	alertRepo     *repository.AlertRepository
	workOrderRepo *repository.WorkOrderRepository
	alertSvc      *AlertService
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(
	telemetryRepo *repository.TelemetryRepository,
	deviceRepo *repository.DeviceRepository,
	alertRepo *repository.AlertRepository,
	workOrderRepo *repository.WorkOrderRepository,
	alertSvc *AlertService,
) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertRepo:     alertRepo,
		workOrderRepo: workOrderRepo,
		alertSvc:      alertSvc,
	}
}

// Ingest stores telemetry data and triggers alert evaluation
// BE-P2-02: 使用常量替换魔法数字
func (s *TelemetryService) Ingest(ctx context.Context, data *model.TelemetryData) error {
	// Set timestamp if not provided
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// Set status based on data
	if data.Status == "" {
		data.Status = "normal"
		if data.Temperature > constants.HighTemperatureThreshold || data.Vibration > constants.AbnormalVibrationThreshold {
			data.Status = "warning"
		}
		if data.Temperature > constants.CriticalTemperatureThreshold || data.Vibration > constants.CriticalVibrationThreshold {
			data.Status = "fault"
		}
	}

	// Store telemetry
	if err := s.telemetryRepo.Insert(ctx, data); err != nil {
		return errors.NewDatabaseError(err.Error())
	}

	// Update device status
	status := "online"
	switch data.Status {
	case "warning":
		status = "warning"
	case "fault":
		status = "fault"
	}
	if err := s.deviceRepo.UpdateStatus(ctx, data.DeviceID, status); err != nil {
		logger.L().Error("Failed to update device status", zap.Error(err), zap.String("device_id", data.DeviceID))
	}

	// Broadcast via WebSocket
	Broadcast(model.WSMessage{
		Type:      "telemetry",
		Payload:   data,
		Timestamp: time.Now(),
	})

	// Trigger alert evaluation asynchronously
	// BE-P1-FIX: 使用 context.WithTimeout 替代 context.Background()，防止 goroutine 无限等待
	go func() {
		// Check if alertSvc is available before calling
		if s.alertSvc != nil {
			// 创建带超时的 context，防止异步调用无限等待
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Duration(constants.AlertEvaluationTimeoutSec)*time.Second)
			defer cancel()

			if err := s.alertSvc.EvaluateRules(ctx, data); err != nil {
				// 区分超时错误和其他错误，记录更详细的日志
				if ctx.Err() == context.DeadlineExceeded {
					logger.L().Error("Alert evaluation timeout",
						zap.String("device_id", data.DeviceID),
						zap.Int("timeout_sec", constants.AlertEvaluationTimeoutSec),
						zap.Error(err),
					)
				} else {
					logger.L().Error("Alert evaluation error",
						zap.String("device_id", data.DeviceID),
						zap.Error(err),
					)
				}
			}
		}
	}()

	return nil
}

// GetByDeviceID retrieves telemetry history for a device
// BE-P2-02: 使用常量替换魔法数字
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	if limit <= 0 {
		limit = constants.MaxTelemetryLimit
	}
	data, err := s.telemetryRepo.GetByDeviceID(ctx, deviceID, start, end, limit)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return data, nil
}

// GetLatest retrieves latest telemetry for all devices
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	data, err := s.telemetryRepo.GetLatest(ctx)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return data, nil
}

// GetStats retrieves statistics for a device
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	stats, err := s.telemetryRepo.GetStats(ctx, deviceID, start, end)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return stats, nil
}

// WebSocket connection manager
// BE-P2-02: 使用常量替换魔法数字
// FIX-059: 移除 init() goroutine
// 使用 sync.Once 确保 broadcaster 只启动一次，避免在 init() 中启动 goroutine
// broadcaster 应在需要时显式调用 StartWSBroadcaster()，而不是在 init() 中自动启动

var (
	wsClients   = make(map[*websocket.Conn]bool)
	wsClientsMu sync.RWMutex
	wsBroadcast = make(chan model.WSMessage, constants.WSBroadcastChannelSize)

	// FIX-059: 使用 sync.Once 确保 broadcaster 只启动一次
	broadcasterStarted sync.Once
	broadcasterStopped chan struct{} // 用于停止 broadcaster
)

// AddWSClient adds a WebSocket client
func AddWSClient(conn *websocket.Conn) {
	wsClientsMu.Lock()
	wsClients[conn] = true
	wsClientsMu.Unlock()
}

// RemoveWSClient removes a WebSocket client
func RemoveWSClient(conn *websocket.Conn) {
	wsClientsMu.Lock()
	delete(wsClients, conn)
	wsClientsMu.Unlock()
	conn.Close()
}

// Broadcast sends a message to all WebSocket clients
func Broadcast(msg model.WSMessage) {
	select {
	case wsBroadcast <- msg:
	default:
		// Channel full, drop message
	}
}

// StartWSBroadcaster starts the WebSocket broadcaster
// FIX-059: 使用 sync.Once 确保 broadcaster 只启动一次
// 不再在 init() 中启动 goroutine，改为显式调用
func StartWSBroadcaster() {
	broadcasterStarted.Do(func() {
		broadcasterStopped = make(chan struct{})
		go func() {
			for {
				select {
				case msg := <-wsBroadcast:
					wsClientsMu.RLock()
					for conn := range wsClients {
						if err := conn.WriteJSON(msg); err != nil {
							go RemoveWSClient(conn)
						}
					}
					wsClientsMu.RUnlock()
				case <-broadcasterStopped:
					// Broadcaster stopped
					return
				}
			}
		}()
	})
}

// StopWSBroadcaster stops the WebSocket broadcaster
// FIX-059: 提供停止 broadcaster 的方法
func StopWSBroadcaster() {
	if broadcasterStopped != nil {
		close(broadcasterStopped)
	}
}

// GetWSClientCount returns the number of connected WebSocket clients
func GetWSClientCount() int {
	wsClientsMu.RLock()
	defer wsClientsMu.RUnlock()
	return len(wsClients)
}

// ParseTimeRange parses time range from string
func ParseTimeRange(rangeStr string) (start, end time.Time) {
	end = time.Now()
	switch rangeStr {
	case "1h":
		start = end.Add(-1 * time.Hour)
	case "6h":
		start = end.Add(-6 * time.Hour)
	case "24h":
		start = end.Add(-24 * time.Hour)
	case "7d":
		start = end.Add(-7 * 24 * time.Hour)
	case "30d":
		start = end.Add(-30 * 24 * time.Hour)
	default:
		start = end.Add(-1 * time.Hour)
	}
	return
}

// GetTimeRanges returns predefined time ranges
func GetTimeRanges() []map[string]string {
	return []map[string]string{
		{"value": "1h", "label": "最近 1 小时"},
		{"value": "6h", "label": "最近 6 小时"},
		{"value": "24h", "label": "最近 24 小时"},
		{"value": "7d", "label": "最近 7 天"},
		{"value": "30d", "label": "最近 30 天"},
	}
}

// FIX-059: 移除 init() 中的 goroutine 启动
// init() 函数应该同步完成，不应该在其中启动 goroutine
// 原来的 init() 调用 StartWSBroadcaster() 启动了 goroutine，这是不安全的做法
// 现在改为显式调用 StartWSBroadcaster()，使用 sync.Once 确保只启动一次
//
// 已删除的 init() 函数内容:
// func init() {
//     StartWSBroadcaster()  // 这会启动 goroutine，不安全
// }
//
// 如需使用 WebSocket broadcaster，请在服务器初始化时显式调用 service.StartWSBroadcaster()

// InitTelemetryService initializes the telemetry service
func InitTelemetryService(alertSvc *AlertService, telemetryRepo *repository.TelemetryRepository, deviceRepo *repository.DeviceRepository) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      alertSvc,
	}
}

// GetROIStats calculates ROI statistics from real database data
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetROIStats(ctx context.Context) (*model.ROIStats, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Get device count
	deviceCount, err := s.deviceRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get active alerts count
	activeAlerts, err := s.alertRepo.CountActive(ctx)
	if err != nil {
		logger.L().Error("Failed to count active alerts", zap.Error(err))
		activeAlerts = 0 // Continue with 0 if error
	}

	// Get open work orders count
	openWorkOrders, err := s.workOrderRepo.CountOpen(ctx)
	if err != nil {
		logger.L().Error("Failed to count open work orders", zap.Error(err))
		openWorkOrders = 0 // Continue with 0 if error
	}

	// Get resolved issues count
	resolvedAlerts, err := s.alertRepo.CountByStatus(ctx, "resolved")
	if err != nil {
		logger.L().Error("Failed to count resolved alerts", zap.Error(err))
		resolvedAlerts = 0 // Continue with 0 if error
	}

	// Calculate uptime percentage based on alerts vs total devices
	// Assumption: If no devices, uptime is 100%. Otherwise, calculate based on active alerts
	uptimePercentage := 100.0
	if deviceCount > 0 {
		// Each active alert reduces uptime by a factor
		// Simple formula: uptime = 100 - (activeAlerts * 0.5) with min 95% floor
		alertFactor := float64(activeAlerts) / float64(deviceCount) * 10
		uptimePercentage = 100.0 - alertFactor
		if uptimePercentage < 95.0 {
			uptimePercentage = 95.0
		}
	}

	// Calculate predicted savings
	// Base savings: $1000 per device per month for monitoring
	// Bonus savings: $500 per resolved issue (preventive maintenance value)
	// Penalty: $100 per active alert (operational disruption cost)
	baseSavings := float64(deviceCount) * 1000.0
	resolvedSavings := float64(resolvedAlerts) * 500.0
	alertCost := float64(activeAlerts) * 100.0
	savings := baseSavings + resolvedSavings - alertCost
	if savings < 0 {
		savings = 0
	}

	// Calculate average response time (hours)
	// Estimated based on resolved vs active alerts ratio
	avgResponseTime := 2.5 // Default 2.5 hours
	if resolvedAlerts > 0 {
		// If we have resolved alerts, estimate better response time
		avgResponseTime = 1.5 + (float64(activeAlerts) / float64(resolvedAlerts+1) * 2)
	}

	return &model.ROIStats{
		TotalDevices:     deviceCount,
		ActiveAlerts:     activeAlerts,
		OpenWorkOrders:   openWorkOrders,
		ResolvedIssues:   resolvedAlerts,
		PredictedSavings: savings,
		UptimePercentage: uptimePercentage,
		AvgResponseTime:  avgResponseTime,
	}, nil
}

// GetSystemStatus returns system status
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetSystemStatus(ctx context.Context) (*model.SystemStatus, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	start := time.Now()

	// Simple DB ping
	_, err := s.deviceRepo.Count(ctx)
	dbLatency := time.Since(start).Milliseconds()

	dbStatus := "healthy"
	if err != nil {
		dbStatus = "unhealthy"
	}

	deviceCount, err := s.deviceRepo.Count(ctx)
	if err != nil {
		logger.L().Error("Failed to count devices", zap.Error(err))
	}

	return &model.SystemStatus{
		Database:    dbStatus,
		DBLatency:   dbLatency,
		Uptime:      time.Since(start).String(),
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		DeviceCount: deviceCount,
		UserCount:   0,
	}, nil
}

// GetHistoricalData retrieves historical telemetry with time range
// BE-P2-02: 使用常量替换魔法数字
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetHistoricalData(ctx context.Context, deviceID string, timeRange string, limit int) ([]model.TelemetryData, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	start, end := ParseTimeRange(timeRange)
	if limit <= 0 {
		limit = constants.MaxTelemetryLimit
	}
	data, err := s.telemetryRepo.GetByDeviceID(ctx, deviceID, start, end, limit)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return data, nil
}

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ValidateTelemetryData validates incoming telemetry data
// SEC-MED-04: Enhanced validation including device_id format check
func ValidateTelemetryData(data *model.TelemetryData) error {
	if data.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}

	// SEC-MED-04: Validate device_id format
	// Device ID must be UUID format or safe alphanumeric ID
	if len(data.DeviceID) > 100 {
		return fmt.Errorf("device_id too long (max 100 characters)")
	}

	// Basic format validation - alphanumeric, dash, underscore allowed
	for _, c := range data.DeviceID {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("device_id contains invalid characters (only alphanumeric, dash, underscore allowed)")
		}
	}

	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// Validate numerical ranges for sensor data
	if data.Temperature < -100 || data.Temperature > 1000 {
		return fmt.Errorf("temperature value out of valid range")
	}
	if data.Pressure < 0 || data.Pressure > 1000 {
		return fmt.Errorf("pressure value out of valid range")
	}
	if data.Vibration < 0 || data.Vibration > 100 {
		return fmt.Errorf("vibration value out of valid range")
	}
	if data.Humidity < 0 || data.Humidity > 100 {
		return fmt.Errorf("humidity value out of valid range (0-100)")
	}
	if data.Power < 0 || data.Power > 10000 {
		return fmt.Errorf("power value out of valid range")
	}

	return nil
}
