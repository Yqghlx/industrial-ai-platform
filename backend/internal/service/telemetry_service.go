package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// TelemetryService handles telemetry data
type TelemetryService struct {
	telemetryRepo *repository.TelemetryRepository
	deviceRepo    *repository.DeviceRepository
	alertSvc      *AlertService
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(
	telemetryRepo *repository.TelemetryRepository,
	deviceRepo *repository.DeviceRepository,
	alertSvc *AlertService,
) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
		alertSvc:      alertSvc,
	}
}

// Ingest stores telemetry data and triggers alert evaluation
func (s *TelemetryService) Ingest(ctx context.Context, data *model.TelemetryData) error {
	// Set timestamp if not provided
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// Set status based on data
	if data.Status == "" {
		data.Status = "normal"
		if data.Temperature > 100 || data.Vibration > 3.0 {
			data.Status = "warning"
		}
		if data.Temperature > 120 || data.Vibration > 5.0 {
			data.Status = "fault"
		}
	}

	// Store telemetry
	if err := s.telemetryRepo.Insert(ctx, data); err != nil {
		return err
	}

	// Update device status
	status := "online"
	switch data.Status {
	case "warning":
		status = "warning"
	case "fault":
		status = "fault"
	}
	s.deviceRepo.UpdateStatus(ctx, data.DeviceID, status)

	// Broadcast via WebSocket
	Broadcast(model.WSMessage{
		Type:      "telemetry",
		Payload:   data,
		Timestamp: time.Now(),
	})

	// Trigger alert evaluation asynchronously
	go func() {
		// Check if alertSvc is available before calling
		if s.alertSvc != nil {
			if err := s.alertSvc.EvaluateRules(context.Background(), data); err != nil {
				logger.L().Error("Alert evaluation error",
					zap.String("device_id", data.DeviceID),
					zap.Error(err),
				)
			}
		}
	}()

	return nil
}

// GetByDeviceID retrieves telemetry history for a device
func (s *TelemetryService) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	if limit <= 0 {
		limit = 1000
	}
	return s.telemetryRepo.GetByDeviceID(ctx, deviceID, start, end, limit)
}

// GetLatest retrieves latest telemetry for all devices
func (s *TelemetryService) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	return s.telemetryRepo.GetLatest(ctx)
}

// GetStats retrieves statistics for a device
func (s *TelemetryService) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	return s.telemetryRepo.GetStats(ctx, deviceID, start, end)
}

// WebSocket connection manager
// FIX-059: 移除 init() goroutine
// 使用 sync.Once 确保 broadcaster 只启动一次，避免在 init() 中启动 goroutine
// broadcaster 应在需要时显式调用 StartWSBroadcaster()，而不是在 init() 中自动启动

var (
	wsClients   = make(map[*websocket.Conn]bool)
	wsClientsMu sync.RWMutex
	wsBroadcast = make(chan model.WSMessage, 100)

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

// GetROIStats calculates ROI statistics
func (s *TelemetryService) GetROIStats(ctx context.Context) (*model.ROIStats, error) {
	// Get device count
	deviceCount, err := s.deviceRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate estimated savings (mock calculation)
	savings := float64(deviceCount) * 5000.0 // $5000 per device per month

	return &model.ROIStats{
		TotalDevices:     deviceCount,
		ActiveAlerts:     0, // Would be calculated from alerts
		OpenWorkOrders:   0, // Would be calculated from work orders
		ResolvedIssues:   0,
		PredictedSavings: savings,
		UptimePercentage: 99.5,
		AvgResponseTime:  2.5,
	}, nil
}

// GetSystemStatus returns system status
func (s *TelemetryService) GetSystemStatus(ctx context.Context) (*model.SystemStatus, error) {
	start := time.Now()

	// Simple DB ping
	_, err := s.deviceRepo.Count(ctx)
	dbLatency := time.Since(start).Milliseconds()

	dbStatus := "healthy"
	if err != nil {
		dbStatus = "unhealthy"
	}

	deviceCount, _ := s.deviceRepo.Count(ctx)

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
func (s *TelemetryService) GetHistoricalData(ctx context.Context, deviceID string, timeRange string, limit int) ([]model.TelemetryData, error) {
	start, end := ParseTimeRange(timeRange)
	if limit <= 0 {
		limit = 1000
	}
	return s.telemetryRepo.GetByDeviceID(ctx, deviceID, start, end, limit)
}

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ValidateTelemetryData validates incoming telemetry data
func ValidateTelemetryData(data *model.TelemetryData) error {
	if data.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}
	return nil
}
