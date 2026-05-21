package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// Phase 4: TelemetryHandlerNew重构 - 仅依赖Service层
// ============================================

// TelemetryHandlerNew 遥测处理器（新架构）
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "device_id": deviceID})
}

// GetSystemStatus 获取系统状态
func (h *TelemetryHandlerNew) GetSystemStatus(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.telemetrySvc.GetSystemStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAIStatus 获取AI状态
func (h *TelemetryHandlerNew) GetAIStatus(c *gin.Context) {
	ctx := c.Request.Context()

	taskLogs, err := h.agentSvc.GetTaskLogs(ctx, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "active",
		"recent_tasks": taskLogs,
	})
}

// IngestTelemetry 接收遥测数据
func (h *TelemetryHandlerNew) IngestTelemetry(c *gin.Context) {
	var data model.TelemetryData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 占位实现 - 实际需要 TelemetryService.IngestTelemetry 方法
	c.JSON(http.StatusOK, gin.H{"message": "Telemetry ingested (placeholder)"})
}

// AgentQuery AI代理查询
func (h *TelemetryHandlerNew) AgentQuery(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.AgentQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.agentSvc.Query(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
