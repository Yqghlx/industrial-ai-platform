package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/middleware"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/industrial-ai/platform/pkg/response"
	"github.com/industrial-ai/platform/pkg/validation"
	"go.uber.org/zap"
)

// ============================================
// Phase 4: TelemetryHandlerNew重构 - 仅依赖Service层
// ============================================

// TelemetryHandlerNew 遥测处理器（新架构）
// SEC-MAJOR-01: Includes logger for device authentication logging
type TelemetryHandlerNew struct {
	telemetrySvc service.TelemetryServiceInterface
	agentSvc     service.AgentServiceInterface
}

// NewTelemetryHandlerNew 创建遥测处理器
func NewTelemetryHandlerNew(
	telemetrySvc service.TelemetryServiceInterface,
	agentSvc service.AgentServiceInterface,
) *TelemetryHandlerNew {
	return &TelemetryHandlerNew{
		telemetrySvc: telemetrySvc,
		agentSvc:     agentSvc,
	}
}

// GetLatestTelemetry 获取最新遥测数据
func (h *TelemetryHandlerNew) GetLatestTelemetry(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.telemetrySvc.GetLatest(ctx)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetDeviceTelemetry 获取设备遥测数据
func (h *TelemetryHandlerNew) GetDeviceTelemetry(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("id")

	limit := 100
	if l := c.Query("limit"); l != "" {
		var lInt int
		_, _ = fmt.Sscanf(l, "%d", &lInt)
		if lInt > 0 {
			limit = lInt
		}
	}

	data, err := h.telemetrySvc.GetLatestByDevice(ctx, deviceID, limit)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "device_id": deviceID})
}

// GetSystemStatus 获取系统状态
func (h *TelemetryHandlerNew) GetSystemStatus(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.telemetrySvc.GetSystemStatus(ctx)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAIStatus 获取AI状态
func (h *TelemetryHandlerNew) GetAIStatus(c *gin.Context) {
	ctx := c.Request.Context()

	taskLogs, err := h.agentSvc.GetTaskLogs(ctx, 50)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "active",
		"recent_tasks": taskLogs,
	})
}

// IngestTelemetry 接收遥测数据
// SEC-MAJOR-01: Endpoint now requires DeviceAuthRequired middleware for device authentication
// SEC-MED-04: Device ID format validation is required
// Devices must provide valid API key via X-Device-Key header or device_key query parameter
func (h *TelemetryHandlerNew) IngestTelemetry(c *gin.Context) {
	ctx := c.Request.Context()
	var data model.TelemetryData
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// SEC-MED-04: Validate device_id format (UUID or safe alphanumeric ID)
	if data.DeviceID == "" {
		response.BadRequest(c, "device_id is required")
		return
	}

	// Validate device_id format using the validation package
	// This ensures device_id follows UUID format or safe ID pattern
	if err := validation.ValidateDeviceID(data.DeviceID); err != nil {
		response.BadRequest(c, "invalid device_id format: "+err.Error())
		return
	}

	// SEC-MED-05: Check for SQL injection in message field
	if data.Message != "" {
		// Import security package for enhanced SQL injection detection
		// For now, use basic check - enhanced detection in security package
		if strings.Contains(strings.ToLower(data.Message), "' or ") ||
			strings.Contains(strings.ToLower(data.Message), "' and ") ||
			strings.Contains(strings.ToLower(data.Message), "--") ||
			strings.Contains(strings.ToLower(data.Message), ";drop") {
			response.BadRequest(c, "invalid message content")
			return
		}
	}

	// Set timestamp if not provided
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// Set default status if not provided
	if data.Status == "" {
		data.Status = "normal"
	}

	// SEC-MAJOR-01: Log device authentication status if available
	// Device authentication is handled by DeviceAuthRequired middleware
	authDeviceID := middleware.GetDeviceID(c)
	if authDeviceID != "" {
		logger.L().Info("Telemetry from authenticated device",
			zap.String("auth_device_id", authDeviceID),
			zap.String("telemetry_device_id", data.DeviceID),
			zap.Bool("device_authenticated", middleware.IsDeviceAuthenticated(c)))
	}

	// 实际存储遥测数据
	if err := h.telemetrySvc.Ingest(ctx, &data); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Telemetry ingested successfully",
		"device_id":          data.DeviceID,
		"timestamp":          data.Timestamp.Format(time.RFC3339),
		"device_authenticated": middleware.IsDeviceAuthenticated(c),
	})
}

// AgentQuery AI代理查询
func (h *TelemetryHandlerNew) AgentQuery(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.AgentQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.agentSvc.Query(ctx, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
